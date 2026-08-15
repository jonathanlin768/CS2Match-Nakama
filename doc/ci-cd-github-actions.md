# GitHub Actions 一键生产发布

如果尚未完成服务器和外部服务的第一次配置，请先按 [`aws-deployment-step-by-step.md`](aws-deployment-step-by-step.md) 操作；本文集中说明工作流行为与回滚。

工作流是 [`.github/workflows/deploy-production.yml`](../.github/workflows/deploy-production.yml)。在 GitHub 的 Actions 页面选择 **Deploy production**，填写 branch/tag/commit 并点击运行，即可按“无凭据验证 → GHCR 镜像 → 私网后端部署 → 公网 smoke → Pages 前端”执行。生产 Environment 可要求人工批准；GitHub 文档说明 Environment secrets 在保护规则通过前不会提供给 job。[GitHub Environments](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments)

## 自动化做什么

1. 固定 Go 1.24.5 与 Node 版本，执行 Go、前端和部署静态测试，构建 production bundle 与后端镜像；这个 job 不读取生产 secrets。
2. 使用 GitHub 临时 token 推送 `commit SHA` 标记镜像到 GHCR，输出不可变 `image@sha256:digest`。所有 `uses:` Action 固定到完整 commit SHA。
3. 进入受保护 `production` Environment，通过 Tailscale OIDC workload identity 创建短生命周期 CI 节点；不使用公网 Lightsail IP或长期 SSH 私钥。
4. 把精确 commit 传到 `/opt/cs2match`，先备份，再拉取 digest、更新 Compose、检查本地 RPC。失败会恢复旧 digest 并再次检查；二次失败停止循环并保留日志。
5. 从 GitHub Runner 经公网域名验证设备认证、`HealthCheck`、`SimuMatch` 和 WebSocket，成功后才发布同一提交的 Pages 产物。

后端构建直接使用 Docker Hub 上的 Heroic Labs 官方 Nakama 3.30.0/pluginbuilder 3.30.0 镜像，避开 `registry.heroiclabs.com` Scarf Gateway 的独立限流。验证与发布 job 都通过 `deploy/scripts/docker-build-retry.sh` 构建；只有 429、超时、连接重置、临时 DNS 故障和 HTTP 502/503/504 等瞬时网络错误最多尝试三次，非法 tag、Dockerfile 语法、Go 编译或文件缺失会立即失败。版本仍必须严格一致，不能用通用 Go 镜像替代 pluginbuilder。

GitHub 仓库显示名可以继续使用 `CS2Match-Nakama`。工作流只在生成 GHCR 路径时将完整 `owner/repository` 转为小写，因此本仓库发布为 `ghcr.io/jonathanlin768/cs2match-nakama:<commit SHA>`，不会重复追加 `-nakama`。

工作流不会创建 AWS 实例、DNS、Tunnel、Tailnet、R2 bucket、Pages project、Budget，也不会删除资源。这些是首次部署的人工步骤。

## GitHub 首次配置

创建名为 `production` 的 Environment，限制允许部署的 branch/tag，并按你的 GitHub 套餐能力设置 required reviewer。GitHub 的 concurrency 只保证工作流组内最多一个部署在运行，它不代替 Environment 审批。[GitHub 部署与 concurrency](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/control-deployments)

Environment variables：

| 名称 | 示例 | 用途 |
|---|---|---|
| `API_HOST` | `api.example.com` | Tunnel API 域名 |
| `FRONTEND_HOST` | `play.example.com` | Environment deployment URL |
| `PRODUCTION_TAILSCALE_HOST` | `cs2match-prod.tailnet.ts.net` | 私网部署目标，不是公网 IP |
| `CLOUDFLARE_PAGES_PROJECT` | `cs2match` | 已存在的 Pages project |

Environment secrets：

| 名称 | 用途 |
|---|---|
| `TS_OAUTH_CLIENT_ID`、`TS_AUDIENCE` | Tailscale GitHub OIDC federation；credential 只授予 `tag:ci` 的 auth_keys/tag 权限 |
| `NAKAMA_SERVER_KEY` | smoke 与前端构建；它是客户端可见值，仍不要使用默认值 |
| `CLOUDFLARE_API_TOKEN`、`CLOUDFLARE_ACCOUNT_ID` | 仅能部署目标 Pages project 的最小权限 token |

Tailscale 官方 Action 的 workload identity 模式需要 `id-token: write`、client ID、audience 和 tag，生成的节点是 ephemeral；本工作流仅在 deploy job 开这个权限。[Tailscale GitHub Action](https://github.com/tailscale/github-action)

GHCR package 可以保持 private。部署 job 用当次运行的短期 `GITHUB_TOKEN` 通过 Tailscale 给主机执行 `docker login`，部署结束无论成功失败都会执行 `docker logout`；不需要在服务器保存 PAT 或长期 registry 密码。若任务被强制终止而来不及 logout，该 token 也会随工作流结束失效。

## 外部服务首次配置

1. AWS：创建 Lightsail 2 GB Ubuntu 实例，先临时把 SSH 22 限制为你的 IP；安装 Docker/Tailscale并完成 `/opt/cs2match` 初始化后关闭公网 22。5432/7350/7351 从不开放。
2. Tailscale：给主机 `tag:prod`，应用 [`deploy/tailscale/policy.hujson.example`](../deploy/tailscale/policy.hujson.example)，配置 GitHub OIDC trust 只允许仓库的 `production` Environment 与 `tag:ci`。
3. Cloudflare Tunnel：创建 remotely-managed Tunnel，API hostname 指向 `http://nakama:7350`；对应容器使用 token。不要再创建第二条 WebSocket ingress，也不要公开 Console。
4. Cloudflare Pages：预先创建 Direct Upload project、绑定前端二级域名；部署由已锁定在 `client/package-lock.json` 的 Wrangler 完成。由于 GitHub Actions checkout 指定 SHA 后处于 detached HEAD，省略 `--branch` 会被 Wrangler 推断成字面值 `HEAD` 并创建 preview。工作流使用现有 Pages Edit token 通过官方 API 将 project 的生产分支验证或校准为 `master`，再显式传 `--branch master`；发布后从 `FRONTEND_HOST` 验证 React 入口。SPA 回退由 [`client/public/_redirects`](../client/public/_redirects) 提供。
5. R2：创建私有 bucket 和仅该 bucket 的读写 token；restic 在客户端加密。把凭据只放服务器 `/opt/cs2match/.env.production`，不放 GitHub。
6. AWS Billing：设置 50%、80%、100% 等多级 Budget 邮件告警；Budget 不是硬限额。

## 日常发布和回滚

在 Actions 页面手动运行并选定 ref。验证 job 失败不会触达生产；进入 deploy job 前按 Environment 提示批准。后端健康后前端才发布，所以前端失败不会破坏已经健康的 API；在 Cloudflare Pages 的 Deployments 页面把上一成功 deployment 重新提升/回滚即可。

本机 RPC 就绪检查使用 `.env.production` 中的 `RUNTIME_HTTP_KEY`，不会使用浏览器可见的 `NAKAMA_SERVER_KEY`。第一次部署若尚无旧健康 digest 且检查失败，脚本会执行不带 `-v` 的 Compose `down`，关闭 Tunnel 和半部署容器但保留 PostgreSQL named volume；Pages 步骤不会执行。下一次运行会把已存在的数据库卷视为需要先备份的数据，不要手工删除该 volume。

后端部署脚本记录当前镜像并执行发布前备份。本机健康后，新 digest 先保持 pending；只有公网认证、RPC、WebSocket smoke 全部成功才写入 `last-known-good-image`。本机或公网检查失败都会恢复旧 digest；首次部署没有旧版本时则关闭半部署栈并保留数据库卷。

若工作流被取消，或日志明确显示自动回滚失败，登录 Tailscale 后先检查 pending 状态：

```bash
cd /opt/cs2match
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=200 nakama
cat deploy/state/last-known-good-image
cat deploy/state/pending-backend-deployment
```

`pending-backend-deployment` 第一行是待确认/失败 digest。确认要撤销后运行：

```bash
bash deploy/scripts/rollback-backend.sh '第一行的完整digest'
```

若没有 pending 文件，需要手工重发已确认的旧版本时，再运行 `bash deploy/scripts/deploy-backend.sh 'ghcr.io/owner/cs2match-nakama@sha256:已确认的digest'`。

不要把 `.env.production`、Tunnel token、restic 密码或日志全文粘贴到 issue。脚本日志只输出 tag/digest/状态，不打印秘密。

## 运维与验收

```bash
cd /opt/cs2match
deploy/scripts/preflight.sh
deploy/scripts/backup-db.sh daily
deploy/scripts/restore-verify.sh latest
deploy/scripts/service-control.sh status
docker stats --no-stream
docker system df
```

每月至少一次隔离恢复并记录日期、snapshot、结果；检查 json-file 日志轮转、磁盘余量、容器 health、R2 最近快照和 Budget 通知。真实环境还需人工验证：公网端口扫描只看到预期边缘服务、Console/SSH 仅 Tailnet 可达、前端认证/Session 恢复/好友与联系方式卡片/WebSocket 正常、Nakama 短暂重启后客户端能恢复。

仓库提供 `npm run load:production -- 20|50` 记录当前一次性 `SimuMatch` 的成功率、p50/p95 与响应大小；同时在主机运行 `deploy/scripts/capture-runtime-stats.sh 120 2` 留存 CPU/内存/IO CSV。它不是自动扩容承诺。
