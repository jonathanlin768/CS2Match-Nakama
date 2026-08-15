# CS2Match 从零到上线：AWS 逐步部署手册

> 适用版本：本仓库当前生产部署实现；核对日期：2026-08-15。本文假设 AWS Lightsail 已创建，Ubuntu 24.04 上的 Docker Engine 与 Docker Compose 已安装并验证成功。

这份手册的目标是：第一次人工完成基础设施配置，以后只需在 GitHub Actions 中点击 **Run workflow**，即可测试、备份、更新后端、验证公网 API 并发布前端。

## 部署速记：未来先看这一节

这套方案不是“把前后端都扔进一台 AWS”。AWS 只常驻 PostgreSQL、Nakama 和 Cloudflare Tunnel；前端在 Cloudflare Pages，镜像在 GHCR，发布由 GitHub Actions 执行，管理链路走 Tailscale，数据库备份加密后放 R2。

### 一共用了多少外部服务

按供应商/账号关系算，需要记住 **5 个外部平台**：域名注册商、AWS、GitHub、Cloudflare、Tailscale。按实际承担的功能拆分，是 **10 项核心在线能力**：

| 平台 | 核心能力 | 本项目用途 |
|---|---:|---|
| 域名注册商 | 1 | 持有并续费主域名；DNS 托管到 Cloudflare 后，注册商不承载业务流量 |
| AWS | 1 | Lightsail 运行 `db`、`nakama`、`cloudflared`；AWS Budget 只是附加告警，不是硬限额 |
| GitHub | 3 | 仓库使用 `master`；Actions/production Environment 验证和部署；GHCR 保存不可变后端镜像 |
| Cloudflare | 4 | DNS 管域名；Tunnel 发布 API；Pages 托管前端；R2 保存加密数据库备份 |
| Tailscale | 1 | 管理员 SSH/Console 私网和 GitHub Runner 的短生命周期部署通道 |

Docker Hub、npm registry、Go/Node 官方下载和 GitHub Releases 是构建期公共依赖，不计入上面 10 项，但它们暂时不可用时 CI 也可能失败。

### 只需要记住两种流程

**第一次冷启动，只做一次：** 创建 Lightsail → 安装 Docker/Tailscale → clone 到 `/opt/cs2match` → 创建 Cloudflare Pages/Tunnel/R2 → 填写 `.env.production` → 配置 GitHub production Environment 与 Tailscale OIDC → 手动运行工作流。

**以后发布：** push 到 `master` → GitHub Actions 手动 **Run workflow** → validate → 推送 GHCR digest → 经 Tailscale 传输准确 commit → 发布前备份 → 更新 Nakama → 本机 Runtime RPC 健康检查 → 公网认证/RPC/WebSocket smoke → 发布 Pages。

重要边界：

- `git push` **不会自动部署**；本工作流是 `workflow_dispatch`，仍要点击 **Run workflow**。
- 服务器上的第一次 clone 是冷启动。后续不用手工 `git pull`，工作流会传输准确 commit，并保留服务器上的 `.env.production`。
- Go 插件不是进程内热更新；更新 Nakama 会有几秒短暂断连，前端 Pages 发布则不重启 AWS。
- `BACKEND_IMAGE_REF` 只有第一次可保留占位符；成功发布后必须是 `ghcr.io/...@sha256:...`。

### 凭据边界速记

- 浏览器可见：`NAKAMA_SERVER_KEY`。它必须非默认且前后端一致，但不是秘密。
- 仅服务器 `.env.production`：数据库、Session/Refresh、`RUNTIME_HTTP_KEY`、Console、Tunnel、R2 S3 凭据和 `RESTIC_PASSWORD`。
- 仅 GitHub production Environment：Tailscale client ID/audience、Pages API token、Cloudflare Account ID，以及用于前端构建/smoke 的同一 `NAKAMA_SERVER_KEY`。
- Cloudflare Tunnel token、R2 S3 token、Pages API token 是三套不同凭据，不能混用。
- `RESTIC_PASSWORD` 是你自己生成的备份加密密码，不是 R2 提供的密码；丢失后即使 R2 对象仍在也无法恢复。
- 任何真实 secret 一旦进入 Git 历史，都应立即在对应平台轮换/撤销；只从文件删除还不够。

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
```

先在 Tailscale 管理后台的 **Access controls** 中，把仓库的 [`deploy/tailscale/policy.hujson.example`](../deploy/tailscale/policy.hujson.example) 合并进现有策略，特别是先定义 `tag:prod` 的 `tagOwners`。若 Tailnet 已有规则，不要直接覆盖整个文件。未先授权 tag 就运行下一条命令，会得到 `requested tags [tag:prod] are invalid or not permitted`。

策略保存成功后，在服务器执行：

```bash
sudo tailscale up --ssh --advertise-tags=tag:prod
```

浏览器会给出登录链接。完成授权后执行：

```bash
tailscale status
tailscale ip -4
```

记下 `tailscale ip -4` 返回的 `100.x.x.x` 地址，稍后填入 `TAILSCALE_IP`。

该策略允许管理员访问生产机的 SSH/Console，并只允许 `tag:ci` 通过 SSH 部署。

在自己的 Windows 电脑安装 Tailscale，用与服务器授权相同的账号/组织登录并确认设备出现在同一 Tailnet，然后测试：

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

这里的 `CLOUDFLARE_TUNNEL_TOKEN` 是创建 Tunnel 后安装命令 `cloudflared ... --token <值>` 中的 connector token。生产环境由 AWS 上的 `cloudflared` Docker 容器使用它，不要在自己的 Windows 电脑执行 `cloudflared service install`。它也不是 Cloudflare API token，更不是 R2 Access Key。

### 4.3 私有 R2 备份桶

创建一个私有 R2 bucket，例如 `cs2match-backups`，再创建“对象读和写”、且仅应用于该 bucket 的最小权限 S3 API token；不需要 R2 管理员权限。除非 Lightsail 已有稳定出口 IP，否则先不要设置客户端 IP 过滤。保存以下值：

```text
Cloudflare Account ID
R2 Access Key ID
R2 Secret Access Key
bucket 名称
```

Cloudflare Account ID 是账号标识，不是“令牌值”。R2 S3 endpoint 通常形如 `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`，其中子域部分就是这里要用的 Account ID。创建 token 后页面还会显示 Access Key ID 与 Secret Access Key；Secret 通常只显示一次。

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
BACKEND_IMAGE_REF=ghcr.io/owner/cs2match-nakama@sha256:REPLACE_ME
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

第二条成功时通常**没有任何输出**，退出码为 0 就表示 Compose 能解析；它不会启动容器。若仍想确认退出码，可紧接着运行 `echo $?`。第一次部署的镜像占位符能通过 Compose 语法解析，但不带 `--backend-image` 的严格预检仍会拒绝它，这是两层不同检查。

注意以下规则：

- `DB_PASSWORD`、Console、Session、Runtime、Tunnel、R2 和 restic 凭据必须保密且不能互相复用；
- `NAKAMA_SERVER_KEY` 会被编译进浏览器，它不是可靠的秘密边界，但仍要使用非默认随机值；
- 第一次部署前 `BACKEND_IMAGE_REF` 可以暂时保留 `REPLACE_ME`。自动部署会把当次构建得到的真实 GHCR digest 传给预检并写回该行；此时单独运行不带参数的 `deploy/scripts/preflight.sh` 会因占位符而失败，这是预期保护；
- GitHub 的 `FRONTEND_HOST` 只填 `play.你的域名.com`，不加协议；服务器的 `FRONTEND_ORIGIN` 必须填完整的 `https://play.你的域名.com`，因为前者是 hostname，后者是浏览器 Origin；
- `VITE_NAKAMA_*` 不写进服务器文件。GitHub 工作流根据 `API_HOST`、`NAKAMA_SERVER_KEY` 自动构建前端，端口固定为 443，SSL 固定为 true。

## 6. 配置 GitHub production Environment

在 GitHub 仓库进入 **Settings → Environments → New environment**，创建 `production`。建议限制只允许 `master` 部署，并按你的 GitHub 套餐能力添加人工审批。

Deployment branches and tags 中出现 `1 branch ... master` 就是正确结果；不要再配置 `main`。个人仓库能否使用 required reviewer 取决于仓库可见性和当前 GitHub 套餐，它是可选审批层，不影响 branch 限制和 workflow concurrency。

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

Cloudflare Pages API token 只放 GitHub；它应只给目标账号的 **Cloudflare Pages: Edit**，不要把 R2 Account Token 填到这里。R2、Tunnel、数据库等 secrets 只放服务器。

GitHub 仓库名 `CS2Match-Nakama` 可以保留大写，但 GHCR/OCI 镜像路径必须全小写，因此镜像名是 `ghcr.io/jonathanlin768/cs2match-nakama`。GHCR package 可以保持 private：工作流用当次短期 `GITHUB_TOKEN` 让服务器临时 `docker login`，结束时自动 logout，不需要服务器 PAT。Runner 上“credentials are stored unencrypted”的警告发生在一次性环境，不等于凭据已提交到仓库。

## 7. 第一次一键部署

先把本仓库生产部署文件提交并推送到 GitHub 的 `master`。然后打开：

```text
GitHub 仓库 → Actions → Deploy production → Run workflow
```

`ref` 保持 `master`，点击运行；如果设置了 Environment reviewer，在进入 deploy job 时批准。

再次强调：push 只让 GitHub 拿到代码，不会自动启动这个手动工作流。服务器第一次 clone 的目的只是建立 `/opt/cs2match`、放置 `.env.production` 和脚本；日常发布不需要到 AWS 执行 `git pull`。工作流会把当次准确 commit 打包传到 `/opt/cs2match`，并明确排除 `.env.production`。

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

本机健康检查会用服务器专属的 `RUNTIME_HTTP_KEY` 请求真实 `healthcheck` RPC，不能用浏览器可见的 `NAKAMA_SERVER_KEY` 代替。第一次部署还没有上一健康 digest 时，如果该检查失败，脚本会执行不带 `-v` 的 Compose `down`，关闭 Tunnel 和半部署容器但保留 PostgreSQL named volume；修正错误后直接重新运行工作流，不要删除数据库卷。

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
curl --fail --silent --show-error \
  -u "${RUNTIME_HTTP_KEY}:" \
  -H 'Content-Type: application/json' \
  -d '{}' \
  'http://127.0.0.1:7350/v2/rpc/healthcheck?unwrap'
```

不要使用 `NAKAMA_SERVER_KEY` 做这个无 Session RPC 探针，否则 Nakama 返回 `401 HTTP key invalid`。从自己的电脑还可以检查不含秘密的公开基础健康地址：

```bash
curl --fail https://api.你的域名.com/healthcheck
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
bash deploy/scripts/deploy-backend.sh 'ghcr.io/owner/cs2match-nakama@sha256:已确认的64位digest'
```

如果 Actions 被取消或提示公网 smoke 的自动回滚也失败，不要直接开始下一次发布。先查看 `deploy/state/pending-backend-deployment`；第一行是失败 digest，确认后执行 `bash deploy/scripts/rollback-backend.sh '第一行的完整digest'`。回滚成功会清除 pending；否则新部署会 fail closed，避免覆盖尚未处理的现场。

不要把 `.env.production`、完整日志中的凭据、Tunnel token 或 R2 密钥粘贴到 GitHub issue 或聊天中。

## 13. 2026-08-15 实战踩坑索引

这次第一次真实部署曾出现“数据库迁移、Nakama、插件、Tunnel 都启动成功，但 Actions 仍回滚”的情况。最终确认不是 Docker 镜像或 AWS 性能问题，而是部署探针拿浏览器可见的 server key 去做 Runtime RPC 认证，得到 `401 HTTP key invalid`。修复基线提交为 `a72e3fd`：探针改用 `RUNTIME_HTTP_KEY`，集成测试真实请求 RPC，首次无旧版本失败时关闭半部署栈但保留数据库卷。

| 现象 | 原因与正确处理 |
|---|---|
| `apt update` 报 Docker `sources` 的 `Malformed entry ... (URI)` | DEB822 文件的 `Types/URIs/Suites/Components/Architectures/Signed-By` 必须属于同一段，不要在每行之间插空行；重写后先 `cat` 检查 |
| Docker 安装成功但普通用户无权限 | `sudo usermod -aG docker ubuntu` 后退出当前 SSH 并重新登录；无需重启主机 |
| Tailscale 授权报 `tag:prod invalid or not permitted` | 先在 Access controls 定义并授权 `tagOwners`，再运行带 `--advertise-tags=tag:prod` 的 `tailscale up` |
| Tailscale 在个人电脑上很慢 | 日常 CI 仍可只用 Tailscale；人工紧急管理可从 Lightsail 浏览器控制台进入，或临时把公网 22 仅限自己的当前 IP，修复后关闭 |
| Cloudflare 页面展示多种 token，不知道填哪个 | Tunnel connector token → 服务器；R2 Access/Secret → 服务器；Pages API token → GitHub；三者绝不能互换 |
| `RESTIC_PASSWORD` 似乎应该由 R2 提供 | 它是本地客户端加密口令，必须自己生成并独立保管；R2 只提供对象存储 Access/Secret |
| `docker compose ... config --quiet` 没反应 | 无输出且退出码 0 就是成功；这只解析配置，不启动容器 |
| `FRONTEND_HOST` 与 `FRONTEND_ORIGIN` 看起来重复 | GitHub `FRONTEND_HOST` 只写 hostname；服务器 `FRONTEND_ORIGIN` 必须包含 `https://` |
| GitHub 仓库有大写，GHCR 报 repository must be lowercase | GitHub 仓库名无需改；只把镜像路径转小写为 `ghcr.io/jonathanlin768/cs2match-nakama`。非法 tag 在 build/push 前就失败，不会产生镜像 |
| Docker build 瞬时失败 | 只对 429、超时、DNS、连接重置、502/503/504 重试；非法 tag、语法和编译错误立即失败，重试无意义 |
| Linux Runner 报脚本 `Permission denied` | Windows Git 常见 `100644`。工作流/入口用 `bash script.sh` 调子脚本，并在服务器初始化时执行 `chmod +x deploy/scripts/*.sh` |
| 测试列表全部 `pass`，job 最后仍是 exit 1 | `pass` 只代表前一个测试命令；继续向下看同一 step 的最后命令和完整日志。本项目曾是后续脚本的 mock 状态未隔离 |
| 直接运行 Nakama 镜像检查时报数据库连接拒绝 | Nakama 启动需要 PostgreSQL；镜像集成检查必须先启动真实测试 DB、确认 TCP/SQL，再启动 Nakama |
| 首次部署仍读取 `BACKEND_IMAGE_REF=...REPLACE_ME` | Shell 导出变量会覆盖 Compose env 文件。部署脚本现在会同时原子更新文件与当前进程环境；不要手工把占位符当真实 digest |
| 容器全部 Healthy，但部署等待 60 秒后回滚 | Docker process health 不等于 RPC 可用；查本机真实 RPC HTTP 状态。Runtime RPC 使用 `RUNTIME_HTTP_KEY`，不是 `NAKAMA_SERVER_KEY` |
| 首次 Actions 失败，但公网 `/healthcheck` 仍返回 200 | 旧脚本可能留下已运行的半部署 Tunnel。当前脚本在无旧健康版本时执行不带 `-v` 的 Compose `down`；数据库 volume 保留，Pages 不发布 |
| 公网 smoke 只显示 `Response {}` | Nakama JS 会直接抛出非 2xx `Response`。本项目还禁止注册时自定义用户名；smoke 必须让服务端生成 8 位玩家码。当前脚本按认证、RPC、WebSocket 分阶段输出 HTTP 状态和错误正文；后端在 smoke 通过前处于 pending，失败会恢复旧 digest，首次部署则停掉半部署栈并保留数据库卷 |
| 认证与 RPC 均通过，但 WebSocket 显示 `[object ErrorEvent]` | Nakama JS 的 `createSocket()` 不继承 Client 的 HTTPS 设置，默认使用 `ws://`；生产 smoke 和浏览器必须显式传入 SSL 配置以使用 `wss://`。错误诊断不能序列化 WebSocket event target，因为连接 URL 可能含 Session token |
| GitGuardian 报 Generic Password | 先定位具体 commit/行；如果是真 secret，立即轮换/撤销并清理历史。即使怀疑是示例值，也不要在确认前忽略告警 |
| WSL 中找不到 Docker | 在 Docker Desktop 启用该发行版的 WSL integration；也可用 Git Bash 调本机 Docker。Git Bash 给容器传 `/nakama/...` 时设置 `MSYS_NO_PATHCONV=1`，避免 MSYS 把容器路径改写成 Windows 路径 |
| 已 push 但 AWS 没更新 | 这是手动工作流；还要到 Actions 点击 **Run workflow**。服务器不需要 `git pull` |

排障顺序固定为：先看失败 job 的第一个非零命令 → 判断是 validate、publish 还是 deploy → deploy 时先查 `docker compose ps` 和 Nakama 日志 → 再查本机 Runtime RPC → 最后查公开 Tunnel/smoke。不要因为容器显示 Healthy 就跳过真实 RPC。

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
