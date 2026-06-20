# CS2Match-Nakama CI/CD 部署指南

> **方案: Plan A — Lightsail $10/月 + Cloudflare Pages $0 + GitHub Actions 手动部署**
>
> 目标: 在 GitHub 网页上点一下按钮，自动构建 → 后端部署到 Lightsail，前端部署到 Cloudflare Pages。
> 核心安全原则: **零 AWS 长期凭据暴露，零个人信息泄露。**

---

## 目录

- [一、架构总览](#一架构总览)
- [二、前置条件检查清单](#二前置条件检查清单)
- [三、AWS 端设置（一次性，~15 分钟）](#三aws-端设置一次性15-分钟)
  - [3.1 创建 Lightsail 实例](#31-创建-lightsail-实例)
  - [3.2 配置 Lightsail 防火墙](#32-配置-lightsail-防火墙)
  - [3.3 安装 Docker 环境](#33-安装-docker-环境)
  - [3.4 创建项目目录结构](#34-创建项目目录结构)
  - [3.5 首次启动服务](#35-首次启动服务)
- [四、Cloudflare 端设置（一次性，~20 分钟）](#四cloudflare-端设置一次性20-分钟)
  - [4.1 注册与域名](#41-注册与域名)
  - [4.2 Cloudflare Tunnel（HTTPS API）](#42-cloudflare-tunnelhttps-api)
  - [4.3 Cloudflare Pages（前端托管）](#43-cloudflare-pages前端托管)
  - [4.4 获取 API Token 和 Account ID](#44-获取-api-token-和-account-id)
- [五、AWS IAM OIDC 设置（可选）](#五aws-iam-oidc-设置可选)
- [六、GitHub 端设置](#六github-端设置)
  - [6.1 添加 GitHub Secrets](#61-添加-github-secrets)
  - [6.2 创建 Workflow 文件](#62-创建-workflow-文件)
- [七、日常部署操作](#七日常部署操作)
  - [7.1 部署步骤](#71-部署步骤)
  - [7.2 验证部署结果](#72-验证部署结果)
  - [7.3 部署失败回滚](#73-部署失败回滚)
- [八、凭据与密钥清单](#八凭据与密钥清单)
- [九、故障排查](#九故障排查)

---

## 一、架构总览

```
你 git push 代码到 GitHub
        │
        ▼
   GitHub 仓库（代码干净地待着，不触发任何部署）
        │
        │  你去 Actions 页 → 点 "Run workflow" 按钮
        ▼
┌──────────────────────────────────────────────────────────┐
│  GitHub Actions (Ubuntu runner, 免费)                      │
│                                                          │
│  Phase 1 — Build:                                        │
│    ① Docker run pluginbuilder → backend.so                │
│    ② npm ci && vite build → client/dist/ （注入生产 env）  │
│                                                          │
│  Phase 2 — Deploy Backend (Lightsail):                   │
│    ③ SCP backend.so → Lightsail                          │
│    ④ SSH → docker compose up -d nakama                   │
│                                                          │
│  Phase 3 — Deploy Frontend (Cloudflare):                 │
│    ⑤ wrangler pages deploy client/dist/ → Cloudflare      │
└───────────────┬──────────────────────┬───────────────────┘
                │                      │
                ▼                      ▼
┌──────────────────────┐    ┌──────────────────────────────┐
│  AWS Lightsail $10    │    │  Cloudflare $0                │
│                      │    │                              │
│  Docker:             │    │  Pages (前端 CDN)             │
│  ┌────────┐ ┌──────┐ │    │  yourdomain.com              │
│  │ Nakama │ │  PG  │ │    │  HTTPS ✅ 全球加速             │
│  │ :7350  │ │:5432 │ │    │                              │
│  └───┬────┘ └──────┘ │    │  Tunnel (API HTTPS)           │
│      │               │    │  api.yourdomain.com          │
│  ┌───┴───────────┐   │    │  ──→ Lightsail:7350          │
│  │  cloudflared   │───┼────┘  HTTPS ✅                    │
│  │  Tunnel 客户端  │   │                                  │
│  └───────────────┘   │                                  │
└──────────────────────┘    └──────────────────────────────┘
         │                            │
         └──────── 用户浏览器 ─────────┘
              • 加载前端: Cloudflare Pages (HTTPS, 全球 CDN)
              • 调用 API: Cloudflare Tunnel → Lightsail (HTTPS)
```

**关键设计**:
- **前端走 Cloudflare Pages** — 不上 Lightsail，不消耗 Lightsail 流量
- **API 走 Cloudflare Tunnel** — 免费 HTTPS，浏览器不会报 Mixed Content
- **GitHub Actions 零长期凭据** — 没有 AWS Access Key，没有 Cloudflare Global Key，只用最小权限 Token
- **手动触发（workflow_dispatch）** — push 代码不会自动部署

---

## 二、前置条件检查清单

开始之前，确认你已有：

- [ ] AWS 账号（Lightsail）
- [ ] GitHub 仓库（代码已上传）
- [ ] Cloudflare 账号（免费注册 [dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up)）
- [ ] （推荐）一个域名托管在 Cloudflare DNS，如 `yourdomain.com`
- [ ] Docker Desktop 本地已安装（用于首次构建 backend.so）

---

## 三、AWS 端设置（一次性，~15 分钟）

### 3.1 创建 Lightsail 实例

1. 打开 [AWS Lightsail Console](https://lightsail.aws.amazon.com/)
2. 点击 **Create instance**
3. 配置：

| 选项 | 值 |
|------|------|
| **Region** | 选离你玩家最近的（亚太: `ap-southeast-1` 新加坡 或 `ap-northeast-1` 东京） |
| **Platform** | Linux/Unix |
| **Blueprint** | OS Only → **Amazon Linux 2023** |
| **Instance plan** | **$10 USD** (1 vCPU, 2 GB RAM, 60 GB SSD, 3 TB 流量) |
| **Instance name** | `cs2match-server` |

4. 点击 **Create instance**，等待 **Running**（约 1-2 分钟）
5. **记录公网 IP**（Lightsail 首页直接可见）

### 3.2 配置 Lightsail 防火墙

Lightsail Console → 选中 `cs2match-server` → **Networking** → **IPv4 Firewall**

删除全部默认规则，添加：

| 应用 | 协议 | 端口 | 来源 | 说明 |
|------|------|------|------|------|
| SSH | TCP | 22 | **仅你自己的 IP** | 管理入口 |
| Nakama Console | TCP | 7351 | **仅你自己的 IP** | 管理后台 |
| Nakama API | TCP | 7350 | **仅你自己的 IP** | 初始测试，Tunnel 配好后可选关闭 |

> ⚠️ **SSH 和 Console 端口必须限制 IP。** 等 Cloudflare Tunnel 验证通过后，7350 端口可以完全关闭——所有 API 流量走隧道。

### 3.3 安装 Docker 环境

在 Lightsail Console 中点击 **Connect using SSH**，逐条执行：

```bash
# 1. 更新系统
sudo yum update -y

# 2. 安装 Docker
sudo yum install -y docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ec2-user

# 3. 退出重新登录使 docker 组生效
exit
```

重新连接 SSH：

```bash
# 4. 安装 Docker Compose
sudo curl -L \
  "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 5. 验证
docker --version
docker-compose --version
```

### 3.4 创建项目目录结构

```bash
# 在 Lightsail 上
mkdir -p /home/ec2-user/cs2match/server/build
mkdir -p /home/ec2-user/cs2match/logs

# 目录结构:
# /home/ec2-user/cs2match/
#   ├── docker-compose.yml      ← 只含 db + nakama（不含前端）
#   ├── .env                    ← 数据库密码等
#   ├── nakama-config.yml       ← Nakama 配置
#   ├── server/
#   │   └── build/
#   │       └── backend.so      ← CI 自动更新
#   └── logs/
```

### 3.5 首次启动服务

在**本地项目根目录**创建生产环境 compose 文件 `docker-compose.prod.yml`（内容见 [aws-deployment.md 第四步](./aws-deployment.md#第四步上传配置文件并首次启动)）。

然后用 scp 上传文件到 Lightsail：

```bash
# 在本地项目根目录执行
scp docker-compose.prod.yml ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/docker-compose.yml
scp nakama-config.yml ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/
```

回到 Lightsail SSH：

```bash
cd /home/ec2-user/cs2match

# 创建 .env
cat > .env << 'EOF'
DB_PASSWORD=<生成随机密码，16+ 位>
NAKAMA_SERVER_KEY=<生成随机 key，32+ 位>
CONSOLE_USERNAME=admin
CONSOLE_PASSWORD=<生成随机密码>
EOF

# 本地先构建一个 backend.so 并上传
# （在本地项目根目录执行: bash server/build.sh）
# scp server/build/backend.so ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/server/build/

# 启动
docker compose up -d

# 验证
docker compose ps
# cs2match-db       Up
# cs2match-nakama   Up

curl http://localhost:7350/healthcheck
# {"status":"ok",...}
```

---

## 四、Cloudflare 端设置（一次性，~20 分钟）

### 4.1 注册与域名

1. 注册 [Cloudflare 账号](https://dash.cloudflare.com/sign-up)（免费）
2. （推荐）准备一个域名，将 DNS 托管到 Cloudflare：
   - Cloudflare Dashboard → **Websites** → **Add a site**
   - 输入你的域名，按向导操作
   - 在你的域名注册商处把 nameserver 改成 Cloudflare 提供的两个地址
   - 等待 DNS 生效（通常几分钟）
3. 如果没有域名，可以用 `*.trycloudflare.com` 临时隧道（但 URL 每次重启会变，不推荐用于实际使用）

> **域名成本**: `.xyz` ~$1/年、`.cfd` ~$2/年、`.click` ~$3/年。建议花这几块钱，一劳永逸。

### 4.2 Cloudflare Tunnel（HTTPS API）

在 Lightsail SSH 中执行：

```bash
# 1. 下载 cloudflared
sudo curl -L \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
  -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared --version

# 2. 登录 Cloudflare
cloudflared tunnel login
# 会输出一个 URL，在浏览器打开 → 选你的域名 → 授权
# 证书自动下载到 ~/.cloudflared/cert.pem

# 3. 创建隧道
cloudflared tunnel create cs2match-api
# 输出: Created tunnel cs2match-api with id <TUNNEL-UUID>
# 记下这个 UUID！
# 凭据文件: /home/ec2-user/.cloudflared/<TUNNEL-UUID>.json

# 4. 绑定域名
cloudflared tunnel route dns cs2match-api api.yourdomain.com
# 把 api.yourdomain.com 的流量路由到这个隧道

# 5. 写配置文件
mkdir -p /home/ec2-user/.cloudflared
cat > /home/ec2-user/.cloudflared/config.yml << 'EOF'
tunnel: <TUNNEL-UUID>
credentials-file: /home/ec2-user/.cloudflared/<TUNNEL-UUID>.json

ingress:
  # Nakama HTTP API + WebSocket
  - hostname: api.yourdomain.com
    service: http://localhost:7350
    originRequest:
      httpHostHeader: localhost

  # 兜底：拒绝其他请求
  - service: http_status:404
EOF
```

> ⚠️ 把 `<TUNNEL-UUID>` 替换为第 3 步输出的实际 UUID。

```bash
# 6. 测试隧道（前台运行）
cloudflared tunnel run cs2match-api
# 看到 Registered tunnel connection 即表示成功
# Ctrl+C 退出

# 7. 安装为 systemd 服务（开机自启）
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
sudo systemctl status cloudflared
# 应显示 active (running)
```

**验证隧道是否生效**：

```bash
# 从你的本地电脑执行
curl https://api.yourDomain.com/healthcheck
# 应返回 {"status":"ok","version":"0.1.0",...}
```

如果成功，意味着：
- 浏览器通过 `https://api.yourdomain.com` 访问你的 Lightsail API
- 全程 HTTPS 加密
- 没有 Mixed Content 问题
- Lightsail 上的 7350 端口可以对外关闭（流量走隧道）

### 4.3 Cloudflare Pages（前端托管）

1. Cloudflare Dashboard → 左侧 **Workers & Pages** → **Create** → **Pages**
2. 选择 **Upload assets**（⭐ 重要：不要连 GitHub，我们用 GitHub Actions 手动部署）
3. Project name: `cs2match`
4. 点击 **Create project**
5. 暂时不传文件（CI 会自动部署），先创建空项目
6. 记录项目名称 `cs2match`，后面 CI 配置会用到

### 4.4 获取 API Token 和 Account ID

**Account ID**：
- Cloudflare Dashboard 首页右侧就能看到
- 或者打开 Workers & Pages 页面，URL 里就有: `dash.cloudflare.com/<ACCOUNT_ID>/workers-and-pages`

**API Token**：
1. Cloudflare Dashboard 右上角头像 → **My Profile** → **API Tokens**
2. 点击 **Create Token** → **Custom Token**
3. 配置：

| 字段 | 值 |
|------|------|
| Token name | `cs2match-deploy` |
| Permissions | `Account` — `Cloudflare Pages` — `Edit` |
| Account Resources | Include — 你的账户 |
| TTL | 设一个遥远的时间（或不设过期） |

4. 点击 **Continue to summary** → **Create Token**
5. ⚠️ **立刻复制保存！这个 token 只显示一次。**

---

## 五、AWS IAM OIDC 设置（可选）

当前方案用 SSH 做部署，不依赖 AWS API，所以 IAM 配置不是必需的。但为了将来扩展（S3 备份、SSM 替代 SSH 等），可以提前配置。

<details>
<summary>展开查看 OIDC 配置步骤（可选）</summary>

1. [AWS IAM Console](https://console.aws.amazon.com/iam/) → **Identity providers** → **Add provider**
2. 选择 **OpenID Connect**
   - Provider URL: `https://token.actions.githubusercontent.com`
   - Audience: `sts.amazonaws.com`
3. 点击 **Get thumbprint** → **Add provider**
4. IAM → **Roles** → **Create role**
   - Trusted entity: Web identity → 选上面的 provider
   - Audience: `sts.amazonaws.com`
   - 暂不附加策略 → Role name: `github-actions-cs2match`
5. 编辑 Trust Policy，限定只有你的仓库能使用：

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::<AWS_ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
      },
      "StringLike": {
        "token.actions.githubusercontent.com:sub": "repo:<GITHUB_USER>/<REPO_NAME>:*"
      }
    }
  }]
}
```

</details>

---

## 六、GitHub 端设置

### 6.1 添加 GitHub Secrets

打开你的 GitHub 仓库 → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

逐个添加：

| Secret 名称 | 值 | 来源 |
|------------|------|------|
| `LIGHTSAIL_IP` | Lightsail 公网 IP | AWS Lightsail 首页 |
| `LIGHTSAIL_SSH_USER` | `ec2-user` | Amazon Linux 2023 默认 |
| `LIGHTSAIL_SSH_KEY` | SSH 私钥完整内容 | 👇 见下方获取方式 |
| `NAKAMA_SERVER_KEY` | 和 Lightsail `.env` 中一致 | 第 3.5 节 |
| `DB_PASSWORD` | 和 Lightsail `.env` 中一致 | 第 3.5 节 |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API Token | 第 4.4 节 |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare Account ID | 第 4.4 节 |
| `CLOUDFLARE_PROJECT` | `cs2match` | Pages 项目名称 |

**获取 Lightsail SSH 私钥**：

创建 Lightsail 实例时 AWS 会让你下载 `.pem` 文件。如果没有或丢失：

1. Lightsail Console → 选中实例 → **Connect** 标签
2. 页面底部 → **Download default key**（或 Create new key）
3. 下载得到 `LightsailDefaultKey-<region>.pem`
4. 用文本编辑器打开 `.pem`，复制**全部内容**（从 `-----BEGIN RSA PRIVATE KEY-----` 到 `-----END RSA PRIVATE KEY-----`，包括这两行）
5. 粘贴到 `LIGHTSAIL_SSH_KEY` Secret 中（支持多行）

### 6.2 创建 Workflow 文件

在本地项目中创建 `.github/workflows/deploy.yml`：

```yaml
# ==============================================
# CS2Match — 部署到 Lightsail + Cloudflare Pages
# ==============================================
# 手动触发: GitHub Actions → Run workflow
# 不会在 git push 时自动运行
# ==============================================

name: Deploy to Lightsail + Cloudflare

on:
  workflow_dispatch:
    inputs:
      skip_build:
        description: 'Skip build, only redeploy existing artifacts'
        type: boolean
        default: false
        required: false
      skip_frontend:
        description: 'Skip frontend deploy (only deploy backend)'
        type: boolean
        default: false
        required: false
      skip_backend:
        description: 'Skip backend deploy (only deploy frontend)'
        type: boolean
        default: false
        required: false

jobs:
  deploy:
    name: Build & Deploy
    runs-on: ubuntu-latest
    timeout-minutes: 20

    steps:
      # ==========================================
      # Phase 0: 准备
      # ==========================================
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: 'npm'
          cache-dependency-path: client/package-lock.json

      # ==========================================
      # Phase 1: 构建 Go 插件
      # ==========================================
      - name: Build Go plugin (backend.so)
        if: ${{ !inputs.skip_build && !inputs.skip_backend }}
        run: |
          echo "=== Pulling Nakama pluginbuilder 3.30.0 ==="
          docker pull registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0

          echo "=== Building backend.so ==="
          docker run --rm \
            -v ${{ github.workspace }}/server:/app \
            -w /app \
            registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0 \
            go build -v -buildmode=plugin -trimpath -o build/backend.so .

          echo "=== Build complete ==="
          ls -lh server/build/backend.so

      # ==========================================
      # Phase 2: 构建前端
      # ==========================================
      - name: Build frontend
        if: ${{ !inputs.skip_build && !inputs.skip_frontend }}
        run: |
          cd client
          npm ci
          npx vite build
        env:
          VITE_NAKAMA_HOST: api.${{ vars.DOMAIN || 'yourdomain.com' }}
          VITE_NAKAMA_PORT: '443'
          VITE_NAKAMA_SERVER_KEY: ${{ secrets.NAKAMA_SERVER_KEY }}
          VITE_NAKAMA_USE_SSL: 'true'

      # ==========================================
      # Phase 3: 部署后端 → Lightsail
      # ==========================================
      - name: Upload backend.so to Lightsail
        if: ${{ !inputs.skip_backend }}
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.LIGHTSAIL_IP }}
          username: ${{ secrets.LIGHTSAIL_SSH_USER }}
          key: ${{ secrets.LIGHTSAIL_SSH_KEY }}
          source: "server/build/backend.so"
          target: "/home/ec2-user/cs2match/"
          strip_components: 2
          overwrite: true

      - name: Restart Nakama on Lightsail
        if: ${{ !inputs.skip_backend }}
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.LIGHTSAIL_IP }}
          username: ${{ secrets.LIGHTSAIL_SSH_USER }}
          key: ${{ secrets.LIGHTSAIL_SSH_KEY }}
          script: |
            set -e
            cd /home/ec2-user/cs2match

            echo "=== Restarting Nakama ==="
            docker compose up -d --force-recreate nakama

            echo "=== Cleanup ==="
            docker image prune -f

            echo "=== Status ==="
            docker compose ps

            echo "=== Backend deploy complete! ==="

      # ==========================================
      # Phase 4: 部署前端 → Cloudflare Pages
      # ==========================================
      - name: Deploy frontend to Cloudflare Pages
        if: ${{ !inputs.skip_frontend }}
        id: deploy
        uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          command: >
            pages deploy client/dist
            --project-name=${{ secrets.CLOUDFLARE_PROJECT || 'cs2match' }}
            --commit-dirty=true
            --branch=main

      - name: Print deployment URLs
        if: ${{ !inputs.skip_frontend }}
        run: |
          echo "🌐 Frontend:  https://${{ vars.DOMAIN || 'cs2match.pages.dev' }}"
          echo "🔌 API:       https://api.${{ vars.DOMAIN || 'yourdomain.com' }}"
```

> ⚠️ 注意 `vite build` 那步的环境变量。`VITE_NAKAMA_HOST` 需要是你 Cloudflare Tunnel 的域名（如 `api.yourdomain.com`）。这个值通过 GitHub Variables (不是 Secrets) 来配，或者直接写死在 workflow 中。

**额外配置 GitHub Variable**：

Settings → Secrets and variables → Actions → **Variables** 标签 → New variable：

| Variable | 值 | 说明 |
|----------|------|------|
| `DOMAIN` | `yourdomain.com`（不含 https） | 域名，用于拼接 `api.yourdomain.com` |

> Variables 和 Secrets 的区别：Variables 不会在日志中脱敏，适合存储非机密配置（如域名）。Secret 自动脱敏，适合存储密码/密钥。

创建完 `deploy.yml` 后，commit 并 push：

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: add deploy workflow for Lightsail + Cloudflare Pages"
git push origin master
```

---

## 七、日常部署操作

### 7.1 部署步骤

```
1. 本地开发、测试完成
   ↓
2. git add . && git commit && git push origin master
   （只是推送代码，不会触发部署）
   ↓
3. 打开 GitHub → Actions → Deploy to Lightsail + Cloudflare
   ↓
4. 点击 "Run workflow"
   ↓
   ┌─────────────────────────────────────┐
   │  Branch: master                     │
   │  ☐ Skip build                       │
   │  ☐ Skip frontend deploy             │
   │  ☐ Skip backend deploy              │
   │                                     │
   │  [Run workflow]                     │
   └─────────────────────────────────────┘
   ↓
5. 等待 3-5 分钟（首次约 8-10 分钟）
   ↓
6. 看到绿色 ✅ → 部署完成
```

**部分部署**（只更新一端）：
- 只改了后端代码 → 勾选 `Skip frontend deploy` → 只更新 Lightsail
- 只改了前端代码 → 勾选 `Skip backend deploy` → 只更新 Cloudflare Pages
- 只改配置，不需要重新构建 → 勾选 `Skip build` + 选择性地 skip

**CLI 部署**（不用开网页）：

```bash
gh workflow run "Deploy to Lightsail + Cloudflare"
# 或指定跳过项
gh workflow run deploy.yml -f skip_frontend=true
```

### 7.2 验证部署结果

```bash
# 1. API 是否正常
curl https://api.yourDomain.com/healthcheck
# → {"status":"ok","version":"0.1.0","timestamp":"..."}

# 2. 前端是否可访问
# 浏览器打开 https://yourdomain.com（如果有域名）
# 或 https://cs2match.pages.dev（Cloudflare 默认域名）
# 打开 F12 → Console → 看有没有报错

# 3. WebSocket 是否连通
# 前端登录后，F12 → Network → WS → 看是否有 101 Switching Protocols

# 4. 数据库是否正常
ssh ec2-user@<LIGHTSAIL_IP>
docker exec cs2match-db psql -U nakama -d nakama -c "SELECT count(*) FROM users;"
```

### 7.3 部署失败回滚

**后端回滚**（如果 .so 文件有问题）：

```bash
# SSH 进 Lightsail
ssh ec2-user@<LIGHTSAIL_IP>
cd /home/ec2-user/cs2match

# 如果有备份（建议部署脚本里加自动备份上版本）
cp server/build/backend.so.backup server/build/backend.so

# 重启
docker compose restart nakama
```

**前端回滚**：

```bash
# Cloudflare Pages 自带版本管理
# Cloudflare Dashboard → Workers & Pages → cs2match
# → Deployments → 选择上一个成功的版本 → ... → Rollback
```

**全部重来**：

在本地修复代码后，重新 `git push`，再触发一次 workflow。

---

## 八、凭据与密钥清单

### 🔴 绝对不能泄露 / 不能提交到 Git

| 凭据 | 存储位置 | 用途 |
|------|---------|------|
| **Lightsail SSH 私钥** | GitHub Secret `LIGHTSAIL_SSH_KEY` | CI 连接服务器 |
| **Cloudflare API Token** | GitHub Secret `CLOUDFLARE_API_TOKEN` | CI 部署到 Pages |
| **NAKAMA_SERVER_KEY** | GitHub Secret + Lightsail `.env` | 客户端-服务器签名 |
| **DB_PASSWORD** | GitHub Secret + Lightsail `.env` | PostgreSQL 密码 |
| **CONSOLE_PASSWORD** | Lightsail `.env` | Nakama 管理后台 |
| **Cloudflare Tunnel 凭据 JSON** | Lightsail `~/.cloudflared/<UUID>.json` | Tunnel 认证 |

### 🟡 不宜随意公开

| 信息 | 存储位置 | 说明 |
|------|---------|------|
| **Lightsail IP** | GitHub Secret `LIGHTSAIL_IP` | 暴露会增加扫描风险 |
| **Cloudflare Account ID** | GitHub Secret `CLOUDFLARE_ACCOUNT_ID` | 配合 Token 使用 |
| **Tunnel UUID** | Lightsail 配置文件 | 知道 UUID 也不能直接操控隧道 |

### 🟢 公开也没事

| 信息 | 说明 |
|------|------|
| GitHub 仓库名 / 用户名 | 公开仓库的话本来就能看到 |
| Cloudflare Pages URL (`*.pages.dev`) | 对外提供服务的地址 |
| 域名 | 你的应用域名，本来就是公开的 |
| AWS 区域 | 公开信息 |

### 不应当出现在任何地方的

| 东西 | 为什么 |
|------|--------|
| **AWS Access Key / Secret Key** | 不需要！SSH 部署不需要 AWS API |
| **AWS 账号根用户邮箱/密码** | 永远别在任何 CI/CD 里用 |
| **个人 GitHub 密码** | `GITHUB_TOKEN` 是自动注入的临时 token |
| **信用卡信息** | 只在 AWS Console 里填，和 CI 完全无关 |

### 你的本地备份清单

建议在项目目录**之外**的地方创建：

```
# ~/Desktop/cs2match-credentials.txt（不提交到 Git！）

=== CS2Match 凭据 ===
创建: 2026-06-14

-- AWS --
Lightsail IP:       ___
区域:               ___
SSH 密钥:           ~/Downloads/LightsailDefaultKey-xxx.pem

-- Cloudflare --
账号邮箱:           ___
Account ID:         ___
API Token:          ___（只显示一次！）
Pages 项目名:       cs2match
域名:               yourdomain.com

-- 数据库 --
DB_USER:            nakama
DB_PASSWORD:        ___

-- Nakama --
SERVER_KEY:         ___
CONSOLE_USERNAME:   admin
CONSOLE_PASSWORD:   ___
```

---

## 九、故障排查

| 症状 | 可能原因 | 解决 |
|------|---------|------|
| Actions 报 `Host key verification failed` | SSH 首次连线需要信任主机 | 在 workflow 中加 `fingerprint`，或先手动 SSH 一次 |
| Actions 报 `Permission denied (publickey)` | SSH 私钥不匹配 | 检查 Secret 里的私钥内容是否完整（包括头尾标记） |
| 前端部署成功但访问 404 | Pages 项目名不对 | 检查 `CLOUDFLARE_PROJECT` Secret |
| `wrangler-action` 报 authentication error | API Token 权限不足 | Token 需要 `Cloudflare Pages — Edit` 权限 |
| 前端能打开，API 调用报 Mixed Content | VITE_NAKAMA_USE_SSL 不是 true | 检查构建时注入的 `VITE_NAKAMA_USE_SSL=true` |
| 前端能打开，API 报 CORS 错误 | Nakama 没允许跨域 | 检查 `nakama-config.yml` 的 `runtime.http_key` |
| API curl 正常但前端连接失败 | Tunnel 只配了 HTTP 没配 WebSocket | Cloudflare Tunnel 默认支持 WebSocket，但检查 ingress 配置 |
| 部署后 Nakama 启动失败 | backend.so 版本不匹配 | 确认用了 `nakama-pluginbuilder:3.30.0` 编译 |
| `docker compose up` 报 port already in use | 旧容器没清理 | `docker compose down && docker compose up -d` |
| `vite build` 失败 | Node 版本不对或依赖缺失 | GitHub Actions 环境用 `node-version: 22`，本地用 `nvm use 22` |

---

> **最后复查清单**:
>
> - [ ] Lightsail 实例运行中，IP 已记录
> - [ ] Docker + Docker Compose 已安装
> - [ ] 首次服务启动成功，curl 验证通过
> - [ ] 防火墙：SSH 和 7351 限 IP
> - [ ] Cloudflare 账号已注册，域名已托管
> - [ ] Cloudflare Tunnel 已安装并验证（`curl https://api.yourdomain.com` 成功）
> - [ ] Cloudflare Pages 项目已创建（空项目即可，CI 会部署）
> - [ ] Cloudflare API Token 已生成并记录
> - [ ] GitHub Secrets 全部填入（8 个）
> - [ ] GitHub Variable `DOMAIN` 已设置
> - [ ] `.github/workflows/deploy.yml` 已创建并 push
> - [ ] 首次 workflow 手动触发成功
> - [ ] 浏览器打开前端 + 登录测试通过
> - [ ] 数据库备份 cron job 已配置
> - [ ] 本地凭据备份文件已创建
