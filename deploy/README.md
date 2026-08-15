# CS2Match 生产部署清单

第一次配置 AWS、Cloudflare、Tailscale、GitHub 和生产密钥时，请按 [`doc/aws-deployment-step-by-step.md`](../doc/aws-deployment-step-by-step.md) 执行。本文件是完成首次配置后的运维速查表。

仓库自动化只管理容器、备份和发布，不会创建或删除 AWS、Cloudflare、Tailscale、R2 或 GitHub 资源，也不会修改开发用 `docker-compose.yml` / `.env.example`。

## 主机和目录

- Ubuntu 24.04 LTS x86_64，Docker Engine + Compose plugin，建议 Lightsail 2 GB RAM 套餐。
- 管理账户为 `ubuntu`，已加入 `docker` 组；SSH 和 Nakama Console 仅经 Tailscale 使用。
- 仓库检出到 `/opt/cs2match`，归 `ubuntu:ubuntu`；真实配置是 `/opt/cs2match/.env.production`，权限必须为 `0600`。
- `deploy/state` 保存锁、上一健康镜像和可能存在的 pending 发布；PostgreSQL 数据在 `cs2match-postgres-data` named volume；二者不得提交。
- 后端、cloudflared 使用 digest；PostgreSQL、restic 使用明确版本，升级必须单独验证。

## 仓库与外部资源盘点

生产新增根目录 `.dockerignore`、`.env.production.example`、`docker-compose.prod.yml`、`nakama-config.prod.yml`，后端镜像文件 `server/Dockerfile.prod`，前端生产校验/测试/smoke/负载脚本，`deploy/` 下的预检、备份、恢复、发布、停服、监控、Tunnel/Tailnet 模板，以及手动 GitHub workflow。现有开发 `docker-compose.yml`、`.env.example`、`client/Dockerfile` 和 `client/nginx.conf` 保持原用途，不由生产脚本覆盖。

外部人工资源包括：Lightsail 实例与防火墙、AWS Budget、Cloudflare DNS/Pages/Tunnel/R2、Tailscale 主机 tag/ACL/OIDC trust、GitHub `production` Environment/GHCR package。真实 secrets 只分布在主机 `.env.production` 和 GitHub Environment；完整清单见 `.env.production.example` 与 `doc/ci-cd-github-actions.md`。

## 首次准备

```bash
sudo mkdir -p /opt/cs2match
sudo chown ubuntu:ubuntu /opt/cs2match
git clone YOUR_REPOSITORY /opt/cs2match
cd /opt/cs2match
cp .env.production.example .env.production
chmod 600 .env.production
```

用 `openssl rand -hex 32` 分别生成数据库、Console 密码、Console signing、session、refresh、runtime 和 restic 密码；每次命令的输出填入对应一行，这些服务端密钥必须不同。Server key 会进入浏览器，不应被视作真正秘密。

除 `BACKEND_IMAGE_REF` 外替换所有占位符。首次发布前该行保留 `REPLACE_ME`，然后从 GitHub Actions 手动运行 **Deploy production**；工作流会把真实镜像 digest 传给预检并写回环境文件。首次健康启动后它还会创建初始异地备份。部署成功后再执行：

```bash
deploy/scripts/preflight.sh
deploy/scripts/restore-verify.sh latest
```

不要在第一次发布前直接运行 `docker compose up`，否则占位镜像无法拉取，也会绕过工作流的测试、健康检查和首次备份流程。

本机部署探针使用服务器专属的 `RUNTIME_HTTP_KEY` 调用真实 `healthcheck` RPC；`NAKAMA_SERVER_KEY` 是浏览器可见值，不能代替该认证。首次发布尚无旧健康镜像时，如果探针失败，脚本会执行不带 `-v` 的 Compose `down`：公开 Tunnel 和所有半部署容器关闭，但 PostgreSQL named volume 保留，前端不会发布。

GitHub 工作流使用两阶段后端确认：本机健康后先保存 `pending-backend-deployment`，公网认证/RPC/WebSocket smoke 成功后才 finalize；失败则调用 `rollback-backend.sh`。若工作流被取消导致 pending 遗留，下一次发布会拒绝覆盖现场；查看该文件第一行，并用 `bash deploy/scripts/rollback-backend.sh '完整pending digest'` 恢复。

按 [Tailscale 官方 Linux 安装说明](https://tailscale.com/kb/1031/install-linux) 安装后，在控制台先应用 `deploy/tailscale/policy.hujson.example`，再运行 `sudo tailscale up --ssh --advertise-tags=tag:prod`。用 `tailscale ip -4` 的结果填写 `TAILSCALE_IP`；确认管理员可经 Tailnet SSH 且 Console 可通过 `http://TAILSCALE_IP:7351` 访问后，关闭 Lightsail 公网 22。CI trust 必须限定到本仓库和 `production` Environment。

脚本依赖 `bash`、`curl`、`flock`、Docker。生产文件采用简单的 `KEY=value`，随机值使用 hex，避免空格、引号和 shell 特殊字符。

## 人工控制台资源

- AWS Lightsail：完成 Tailscale 后关闭公网 22/5432/7350/7351；配置多级 Budget 告警。Tailscale 失效时先在控制台把 22 临时限制为你当前公网 IP，修复后立即关闭。
- Cloudflare：Pages 绑定前端域名；Tunnel 只绑定 API 域名，Token 写入服务器环境文件；R2 bucket 为私有并创建最小权限 API token。
- Tailscale：按 `deploy/tailscale/policy.hujson.example` 配置主机 tag、管理员和 CI workload identity。
- GitHub：创建受保护的 `production` Environment，并配置部署文档列出的 variables/secrets 和审批人。

`deploy/scripts/service-control.sh close` 只停 Nakama 和 Tunnel，保留数据库卷；实例仍计费。永久停服必须先成功运行备份与隔离恢复，再由用户在 AWS 控制台删除实例，并检查静态 IP、snapshot 和附加磁盘等残留资源。
