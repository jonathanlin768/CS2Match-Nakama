# CS2Match 从零到上线：AWS 逐步部署手册

> 适用版本：本仓库当前生产部署实现；核对日期：2026-08-15。本文假设 AWS Lightsail 已创建，Ubuntu 24.04 上的 Docker Engine 与 Docker Compose 已安装并验证成功。

这份手册的目标是：第一次人工完成基础设施配置，以后只需在 GitHub Actions 中点击 **Run workflow**，即可测试、备份、更新后端、验证公网 API 并发布前端。

## 0. 先理解最终结构

旧方案中的“服务器编译 `backend.so`、服务器运行前端容器”已经被当前实现替代：

```text
浏览器 ──HTTPS──> Cloudflare Pages（React 前端）
   └────HTTPS/WSS──> api.example.com ──Cloudflare Tunnel──> Nakama:7350

管理员 / GitHub Actions ──Tailscale──> Ubuntu SSH:22 / Nakama Console:7351
Nakama ──Docker 内网──> PostgreSQL:5432
备份脚本 ──restic 客户端加密──> Cloudflare R2 私有 bucket
```

AWS 主机只运行 `db`、`nakama` 和 `cloudflared` 三个容器。前端放在 Pages，构建后端镜像的工作交给 GitHub Actions。这样主机不需要暴露 5432、7350、7351 或前端端口，也不需要在每次部署时现场编译 Go 插件。

建议准备两个二级域名：

```text
play.你的域名.com    # 前端
api.你的域名.com     # Nakama HTTP 和 WebSocket 共用入口
```

## 1. 确认 AWS Lightsail 主机

推荐 Ubuntu 24.04 LTS x86_64、2 vCPU、2 GB RAM、60 GB SSD 的 Lightsail 2 GB 套餐。当前公开套餐价格约为 12 美元/月；它比 1 GB 套餐更接近你的 100 元人民币上限，但能明显降低 Nakama 与 PostgreSQL 同机时换页、卡顿的风险。[AWS Lightsail 套餐](https://docs.aws.amazon.com/lightsail/latest/userguide/amazon-lightsail-bundles.html)

在 Tailscale 尚未配置前，Lightsail 防火墙只保留：

- TCP 22，仅允许你当前的公网 IP；
- 不开放 5432、7350、7351、3000 和 80/443 源站端口。

先登录主机：

```bash
ssh ubuntu@你的Lightsail公网IP
```

你已经完成 Docker 安装时，可直接验证：

```bash
docker version
docker compose version
docker run --rm hello-world
```

如果执行 Docker 仍提示权限不足，运行：

```bash
sudo usermod -aG docker ubuntu
exit
```

然后重新执行 `ssh ubuntu@你的Lightsail公网IP`。关闭当前 SSH 会话再连接即可，不需要重启服务器。

另外建议立即在 AWS Billing 创建 50%、80%、100% 等多级 Budget 邮件告警。Budget 只负责通知，不是消费硬上限。[AWS Budget 指南](https://docs.aws.amazon.com/pdfs/hands-on/latest/control-your-costs-free-tier-budgets/control-your-costs-free-tier-budgets.pdf)

## 2. 安装并验证 Tailscale

在服务器执行官方安装方式：

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --ssh --advertise-tags=tag:prod
```

浏览器会给出登录链接。完成授权后执行：

```bash
tailscale status
tailscale ip -4
```

记下 `tailscale ip -4` 返回的 `100.x.x.x` 地址，稍后填入 `TAILSCALE_IP`。

在 Tailscale 管理后台的 Access controls 中，把仓库的 [`deploy/tailscale/policy.hujson.example`](../deploy/tailscale/policy.hujson.example) 合并进现有策略。若 Tailnet 已有规则，不要直接覆盖整个文件。该模板允许管理员访问生产机的 SSH/Console，并只允许 `tag:ci` 通过 SSH 部署。

在你自己的电脑加入同一 Tailnet 后，先测试：

```bash
tailscale ssh ubuntu@服务器的Tailscale主机名或100.x.x.x
```

此时先不要关闭 Lightsail 公网 22；等第一次自动部署成功后再关闭，避免把自己锁在服务器外。

## 3. 把仓库放到服务器

公开仓库可直接执行：

```bash
sudo mkdir -p /opt/cs2match
sudo chown ubuntu:ubuntu /opt/cs2match
git clone 你的GitHub仓库地址 /opt/cs2match
cd /opt/cs2match
chmod +x deploy/scripts/*.sh
```

私有仓库建议使用只读 GitHub Deploy Key 完成第一次 clone，不要把 GitHub 密码或个人主 SSH 私钥放到服务器。后续工作流会把准确的 commit 内容经 Tailscale 传到 `/opt/cs2match`，服务器不需要保存 GHCR PAT。

## 4. 在 Cloudflare 创建三项资源

### 4.1 Pages 前端项目

在 Cloudflare 控制台创建名为 `cs2match`（也可自定义）的 Pages Direct Upload 项目，并绑定 `play.你的域名.com`。不要再配置另一个 Git 自动发布流程；本仓库由 GitHub Actions 使用 Wrangler 上传同一 commit 的 `client/dist`。[Pages Direct Upload](https://developers.cloudflare.com/pages/get-started/direct-upload/)

### 4.2 API Tunnel

在 Zero Trust 控制台创建 remotely-managed Cloudflare Tunnel，复制 Tunnel token，并添加一条 Public hostname：

```text
Hostname: api.你的域名.com
Service:  http://nakama:7350
```

不要添加 Console 的 7351 路由，也不用为 WebSocket 单独建第二条规则；Nakama HTTP 与 WebSocket 共用这一个 hostname。

### 4.3 私有 R2 备份桶

创建一个私有 R2 bucket，例如 `cs2match-backups`，再创建仅能读写该 bucket 的最小权限 S3 API token。保存以下值：

```text
Cloudflare Account ID
R2 Access Key ID
R2 Secret Access Key
bucket 名称
```

R2 Standard 当前每月包含一定免费存储与请求额度，但这是免费额度，不是消费硬上限；仍需定期查看用量。[R2 定价](https://developers.cloudflare.com/r2/pricing/)

## 5. 在服务器创建真实生产配置

生产环境使用 `.env.production`，不是开发用 `.env` 或 `.env.example`：

```bash
cd /opt/cs2match
cp .env.production.example .env.production
chmod 600 .env.production
nano .env.production
```

先分别运行多次下列命令；每次输出只使用一次：

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

按下面完整模板填写：

```dotenv
COMPOSE_PROJECT_NAME=cs2match-prod

# 第一次部署前暂时保留占位符；工作流发布镜像后会自动写入真实 digest。
BACKEND_IMAGE_REF=ghcr.io/OWNER/REPOSITORY-nakama@sha256:REPLACE_ME
POSTGRES_IMAGE=postgres:15.18-alpine
CLOUDFLARED_IMAGE_REF=cloudflare/cloudflared:2026.7.2@sha256:4f6655284ab3d252b7f28fedb19fe6c8fc82ee5b1295c20ac74d475e5398a52d
RESTIC_IMAGE=restic/restic:0.18.0

DB_NAME=nakama
DB_USER=nakama
DB_PASSWORD=第一次openssl输出

NAKAMA_SERVER_KEY=单独生成的非默认随机值
CONSOLE_USERNAME=cs2operator
CONSOLE_PASSWORD=另一次openssl输出
CONSOLE_SIGNING_KEY=另一次openssl输出
SESSION_ENCRYPTION_KEY=另一次openssl输出
SESSION_REFRESH_ENCRYPTION_KEY=另一次openssl输出
RUNTIME_HTTP_KEY=另一次openssl输出

API_HOST=api.你的域名.com
FRONTEND_ORIGIN=https://play.你的域名.com
TAILSCALE_IP=100.x.x.x
CLOUDFLARE_TUNNEL_TOKEN=Cloudflare给出的TunnelToken

RESTIC_REPOSITORY=s3:https://你的ACCOUNT_ID.r2.cloudflarestorage.com/cs2match-backups
RESTIC_PASSWORD=另一次openssl输出
AWS_ACCESS_KEY_ID=你的R2AccessKeyID
AWS_SECRET_ACCESS_KEY=你的R2SecretAccessKey
AWS_DEFAULT_REGION=auto
BACKUP_KEEP_DAILY=7
BACKUP_KEEP_WEEKLY=4
```

保存 nano：按 `Ctrl+O`、回车，再按 `Ctrl+X`。随后检查权限和 Compose 解析：

```bash
stat -c '%a %n' .env.production
docker compose --env-file .env.production -f docker-compose.prod.yml config --quiet
```

第一条应显示 `600 .env.production`。

注意以下规则：

- `DB_PASSWORD`、Console、Session、Runtime、Tunnel、R2 和 restic 凭据必须保密且不能互相复用；
- `NAKAMA_SERVER_KEY` 会被编译进浏览器，它不是可靠的秘密边界，但仍要使用非默认随机值；
- 第一次部署前 `BACKEND_IMAGE_REF` 可以暂时保留 `REPLACE_ME`。自动部署会把当次构建得到的真实 GHCR digest 传给预检并写回该行；此时单独运行不带参数的 `deploy/scripts/preflight.sh` 会因占位符而失败，这是预期保护；
- `VITE_NAKAMA_*` 不写进服务器文件。GitHub 工作流根据 `API_HOST`、`NAKAMA_SERVER_KEY` 自动构建前端，端口固定为 443，SSL 固定为 true。

## 6. 配置 GitHub production Environment

在 GitHub 仓库进入 **Settings → Environments → New environment**，创建 `production`。建议限制只允许 `master` 部署，并按你的 GitHub 套餐能力添加人工审批。

添加 Environment variables：

| 名称 | 值示例 |
|---|---|
| `API_HOST` | `api.你的域名.com` |
| `FRONTEND_HOST` | `play.你的域名.com` |
| `PRODUCTION_TAILSCALE_HOST` | 生产机的 Tailscale MagicDNS 名称或 `100.x.x.x` |
| `CLOUDFLARE_PAGES_PROJECT` | `cs2match` |

添加 Environment secrets：

| 名称 | 值 |
|---|---|
| `TS_OAUTH_CLIENT_ID` | Tailscale workload identity client ID |
| `TS_AUDIENCE` | Tailscale trust credential audience |
| `NAKAMA_SERVER_KEY` | 必须和服务器 `.env.production` 完全一致 |
| `CLOUDFLARE_API_TOKEN` | 仅允许发布目标 Pages project 的 token |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare Account ID |

在 Tailscale 管理后台创建 GitHub Actions workload identity/trust credential：issuer 使用 GitHub Actions OIDC，限制到当前仓库和 `production` Environment，只授予创建 `tag:ci` 临时节点所需的 `auth_keys`/tag 权限。把生成的 client ID 和 audience 填入上述 secrets。本工作流使用 `id-token: write`，节点为短生命周期节点，不保存长期 Tailscale auth key。[Tailscale Workload Identity Federation](https://tailscale.com/docs/features/workload-identity-federation)、[Tailscale GitHub Action](https://tailscale.com/docs/integrations/github/github-action)

Cloudflare API token 只放 GitHub；R2、Tunnel、数据库等 secrets 只放服务器。GHCR package 可以保持 private：工作流用当次短期 `GITHUB_TOKEN` 让服务器临时 `docker login`，结束时自动 logout，不需要服务器 PAT。

## 7. 第一次一键部署

先把本仓库生产部署文件提交并推送到 GitHub 的 `master`。然后打开：

```text
GitHub 仓库 → Actions → Deploy production → Run workflow
```

`ref` 保持 `master`，点击运行；如果设置了 Environment reviewer，在进入 deploy job 时批准。

工作流依次执行：

1. Go、前端与部署测试；
2. 构建包含 `backend.so` 的 Nakama 镜像并推送到 GHCR；
3. 通过 Tailscale 把准确 commit 传到服务器；
4. 临时登录私有 GHCR，按 digest 拉取后端；
5. 第一次部署因尚无数据库 volume，会跳过“发布前备份”；服务健康后立即创建第一份 R2 异地备份；
6. 从公网测试认证、RPC 与 WebSocket；
7. 测试通过后构建并发布 Cloudflare Pages；
8. 从服务器移除临时 GHCR 登录信息。

后续部署已有数据库 volume，会在更新 Nakama 前强制完成异地备份；新镜像健康检查失败时自动恢复上一 digest。

## 8. 第一次部署后验证

通过 Tailscale 登录服务器：

```bash
cd /opt/cs2match
deploy/scripts/service-control.sh status
docker compose --env-file .env.production -f docker-compose.prod.yml ps
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=100 nakama
```

在服务器本机检查 RPC 健康状态：

```bash
set -a
source .env.production
set +a
curl -u "${NAKAMA_SERVER_KEY}:" \
  -H 'Content-Type: application/json' \
  -d '{}' \
  'http://127.0.0.1:7350/v2/rpc/HealthCheck?unwrap'
```

然后在自己的电脑检查：

```text
https://api.你的域名.com
https://play.你的域名.com
http://100.x.x.x:7351    # 仅连接 Tailnet 时打开 Nakama Console
```

API 根路径不一定返回普通网页；最终应以 Actions 的公网 smoke 成功及下列业务验证为准：

- 页面和静态资源正常；
- 注册/登录和 Session 恢复正常；
- `HealthCheck`、`SimuMatch` RPC 正常；
- 聊天 WebSocket 正常；
- 好友、联系方式卡片等实际功能正常；
- 浏览器控制台没有 CORS、Mixed Content 或 WebSocket 错误。

确认 Tailscale SSH 与 Console 均可用后，在 Lightsail 控制台关闭公网 TCP 22。最终不需要任何公网入站端口；Cloudflare Tunnel 是由服务器主动向外建立连接。若 Tailscale 故障，可临时把 22 仅开放给你当时的公网 IP，修复后立即关闭。

## 9. 日常更新：以后只点一次

正常更新不再在服务器执行 `git pull`、`server/build.sh` 或手工替换 `backend.so`。代码推送到 `master` 后：

```text
Actions → Deploy production → Run workflow → ref: master
```

这就是项目的一键更新。Go 插件不能在 Nakama 进程内真正热替换；发布后端时 Nakama 会短暂重启，通常会有几秒中断。前端由 Pages 原子发布，不需要重启 AWS 容器。

查看状态或日志：

```bash
cd /opt/cs2match
deploy/scripts/service-control.sh status
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=200 nakama
docker stats --no-stream
docker system df
```

## 10. 每日备份与恢复演练

手工执行一次并验证可恢复：

```bash
cd /opt/cs2match
deploy/scripts/backup-db.sh daily
deploy/scripts/restore-verify.sh latest
```

确认两条命令均成功后，为 `ubuntu` 用户添加每天凌晨 03:15 的任务：

```bash
crontab -e
```

加入一行：

```cron
15 3 * * * cd /opt/cs2match && /opt/cs2match/deploy/scripts/backup-db.sh daily >> /opt/cs2match/deploy/state/backup-cron.log 2>&1
```

至少每月运行一次 `restore-verify.sh latest` 并记录结果。restic 在上传前加密，但仍不要公开 R2 bucket 或泄露 `RESTIC_PASSWORD`。

## 11. 临时停服、重新开放与永久停机

临时关闭公网服务：

```bash
cd /opt/cs2match
deploy/scripts/service-control.sh close
```

它会停止 Nakama 和 Tunnel，保留 PostgreSQL 与 named volume。Lightsail 实例仍然计费。

重新开放：

```bash
deploy/scripts/service-control.sh open
```

永久停机前：

```bash
deploy/scripts/backup-db.sh final
deploy/scripts/restore-verify.sh latest
```

确认异地备份能恢复后，再到 AWS 控制台删除实例，并检查是否遗留 static IP、snapshot、附加磁盘等收费资源。删除实例和 volume 是不可逆操作，仓库脚本不会自动执行。

## 12. 负载验收与故障处理

在自己的电脑从仓库执行 20 并发测试：

```bash
cd client
VITE_NAKAMA_HOST=api.你的域名.com \
VITE_NAKAMA_PORT=443 \
VITE_NAKAMA_SERVER_KEY=与生产一致的ServerKey \
VITE_NAKAMA_USE_SSL=true \
npm run load:production -- 20
```

同时在服务器采样：

```bash
cd /opt/cs2match
deploy/scripts/capture-runtime-stats.sh 120 2
```

低峰期再测试 50 并发，观察成功率、p95、内存、CPU、磁盘 IO 与是否出现 swap。若新版本失败，工作流会自动回滚；进一步排查：

```bash
cd /opt/cs2match
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=200 nakama
cat deploy/state/last-known-good-image
```

需要手工重发已确认的旧版本时：

```bash
deploy/scripts/deploy-backend.sh 'ghcr.io/OWNER/IMAGE@sha256:已确认的64位digest'
```

不要把 `.env.production`、完整日志中的凭据、Tunnel token 或 R2 密钥粘贴到 GitHub issue 或聊天中。

## 最终验收清单

- [ ] Lightsail 为 2 GB Ubuntu，AWS Budget 告警已启用；
- [ ] 服务器 Docker、Compose、Tailscale 均正常；
- [ ] 公网入站端口已全部关闭，SSH/Console 仅 Tailnet 可达；
- [ ] `play.*` 绑定 Pages，`api.*` 只有一条 Tunnel 到 `nakama:7350`；
- [ ] `.env.production` 权限为 0600，所有真实 secrets 均不同且未提交；
- [ ] GitHub `production` Environment variables/secrets 完整；
- [ ] 第一次 Actions 发布、前端、RPC、WebSocket 均验证成功；
- [ ] R2 初始备份成功，每日 cron 和每月恢复演练已安排；
- [ ] 20/50 并发测试与主机资源采样达到可接受结果。

CI/CD 的字段说明和回滚细节见 [`doc/ci-cd-github-actions.md`](ci-cd-github-actions.md)，架构、价格和安全取舍见 [`doc/aws-deployment.md`](aws-deployment.md)。
