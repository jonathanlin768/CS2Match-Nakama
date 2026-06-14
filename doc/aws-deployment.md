# CS2Match-Nakama → AWS 部署方案与成本估算

> **目标：50 人同时在线。按实际用量计价，不依赖 12 个月免费层。**

---

## 一、先算账：50 人同时在线的资源需求

### 负载估算

| 指标 | 值 | 说明 |
|------|------|------|
| WebSocket 并发连接 | 50 | 每人一条长连接，Nakama 轻松应付 |
| 活跃对局数 (峰值) | ~15 局 | 50 人中约 30 人在对局 = 15 局 1v1 |
| MatchLoop 频率 | 10 Hz | 每秒 150 个 tick（15 局 × 10Hz） |
| 每 tick 广播数据 | ~500 bytes | JSON 状态快照（位置/比分/计时） |
| **出站流量/月** | **~200 GB** | 10 matches 平均 × 2人 × 10Hz × 500B |

### 内存预算

| 进程 | 内存占用 |
|------|------|
| Nakama + Go plugin | 150-200 MB |
| PostgreSQL 15 | 200-400 MB（可调参压缩到 200） |
| OS (Amazon Linux) | ~250 MB |
| **cloudflared 隧道** | ~30 MB（新增，替代原 Nginx） |
| 余量 | ~120-370 MB |
| **最低要求** | **1 GB**（紧张）/ **2 GB**（舒适） |

> 结论：2 GB 从容。前端迁到 Cloudflare Pages 后，Lightsail 只跑 Nakama + DB + Tunnel，反而更宽裕。

---

## 二、方案对比（每月成本）

### Plan A: Lightsail $10 + Cloudflare Pages $0 — 🏆 推荐

```
┌──────────────────────────────────────┐
│  Lightsail $10/月                     │
│  1 vCPU / 2 GB RAM / 60 GB SSD        │
│  3 TB 出站流量（已含）                  │
│                                      │
│  ┌────────────┐ ┌──────────────────┐ │
│  │ Nakama     │ │ PostgreSQL 15     │ │
│  │ :7350      │ │ :5432             │ │
│  │ + plugin   │ │ (Docker volume)   │ │
│  └─────┬──────┘ └──────────────────┘ │
│        │                              │
│  ┌─────┴──────┐                       │
│  │ cloudflared│ ← 出站隧道，无需开端口  │
│  │ Tunnel     │                       │
│  └────────────┘                       │
└──────────────────────────────────────┘
         │
         │  Cloudflare Tunnel（加密出站连接）
         ▼
┌──────────────────────────────────────┐
│  Cloudflare（免费）                    │
│                                      │
│  Pages ($0)          Tunnel ($0)     │
│  前端 HTML/JS/CSS      HTTPS API     │
│  yourdomain.com       api.yourdomain │
│  全球 CDN              .com          │
└──────────────────────────────────────┘
         │                    │
         ▼                    ▼
      用户浏览器 ←────── JS API 调用（HTTPS）
```

| 项目 | 月费 |
|------|------|
| Lightsail 实例 (2 GB) | **$10.00** |
| 出站流量 (3 TB 含) | $0 |
| 静态 IP | $0（Lightsail 免费附带） |
| Cloudflare Pages | **$0**（无限站点/流量/请求） |
| Cloudflare Tunnel | **$0**（无限隧道/流量） |
| 域名 (可选但强烈推荐) | ~$0.10/月 ($1-5/年) |
| **合计** | **$10.00/月** |

✅ 优点：一口价，前端 CDN 全球加速，API 自带 HTTPS，无需管理证书
✅ Lightsail 上少跑一个 Nginx 容器，内存更宽裕
⚠️ 需要 Cloudflare 账号（免费注册）

### Plan B: Lightsail $5 — 最小可行

| 项目 | 月费 |
|------|------|
| Lightsail 实例 (1 GB) | **$5.00** |
| 出站流量 (2 TB 含) | $0 |
| Cloudflare 前端 + 隧道 | $0 |
| **合计** | **$5.00/月** |

1 GB 内存跑 PostgreSQL + Nakama + cloudflared 需要调参：

```ini
# postgresql.conf 关键调优（压到 ~200 MB）
shared_buffers = 128MB
effective_cache_size = 256MB
work_mem = 2MB
maintenance_work_mem = 32MB
max_connections = 30          # 50 用户 + 内部连接
```

⚠️ 风险：内存吃紧，流量高峰期可能 OOM。建议加 swap（2 GB）。

### Plan C: Lightsail $5 + Lightsail 托管数据库 $15

分离数据库到 Lightsail 托管 PostgreSQL：

| 项目 | 月费 |
|------|------|
| Lightsail 实例 (1 GB, 跑 Nakama + Tunnel) | **$5.00** |
| Lightsail 托管 PostgreSQL (1 GB, 40 GB SSD) | **$15.00** |
| 出站流量 (2 TB 含) | $0 |
| Cloudflare 前端 + 隧道 | $0 |
| **合计** | **$20.00/月** |

✅ 数据库自动备份、7 天时间点恢复
✅ Nakama 独占 1 GB 内存，更稳

### Plan D: EC2 t3.small + RDS — 标准云架构

| 项目 | 月费 |
|------|------|
| EC2 t3.small (2 vCPU, 2 GB) | **$15.18** |
| EBS gp3 30 GB | **$2.40** |
| RDS db.t3.micro (1 vCPU, 1 GB, 20 GB SSD, 单AZ) | **$12.41** |
| 出站流量 100 GB 免费后 ~100 GB | **~$9.00** |
| **合计** | **~$39.00/月** |

❌ 不推荐：比 Lightsail 贵 4 倍，还没有包含流量。

---

## 三、费用明细拆解

### 出站流量 — 最容易忽略的开销

```
每次 MatchLoop tick 广播状态给 2 个客户端:
  500 bytes × 2 人 = 1 KB/tick
  1 KB × 10 Hz × 3600s = 36 MB/小时/局
  
平均 10 局活跃 × 36 MB × 720 小时/月 = 259 GB/月

AWS EC2 免费额度: 100 GB
超出: 159 GB × $0.09/GB = $14.31/月（仅 WebSocket 数据流量）

前端静态资源走 Cloudflare Pages CDN:
  HTML/JS/CSS/图片 — 完全不走 Lightsail，$0
  
Lightsail 只处理 WebSocket 实时数据:
  3 TB 包含量 vs 200 GB 实际用量 = 绰绰有余
```

> Lightsail 的 2-3 TB 流量 + Cloudflare Pages 免费 CDN = 流量完全不愁。

### 数据库存储增长

| 数据 | 每人 | 50 人 |
|------|------|------|
| Nakama 内置表（用户/好友/聊天） | ~1 MB | ~50 MB |
| 对局记录（每局 ~10 KB × 1000 局） | — | ~10 MB |
| 玩家数据（Storage Engine） | ~5 MB | ~250 MB |
| **合计** | | **~300 MB** |

20 GB 够用很久。

---

## 四、最终推荐

| 方案 | 月费 | 内存 | 数据库 | 前端 | 适合 |
|------|------|------|------|------|------|
| **Plan A** | **$10** | 2 GB | 自管 Docker | Cloudflare Pages | 🏆 **推荐** |
| Plan B | $5 | 1 GB | 自管 Docker | Cloudflare Pages | 极致省钱（有风险） |
| Plan C | $20 | 1 GB | Lightsail 托管 | Cloudflare Pages | 不想管数据库 |
| Plan D | ~$39 | 2 GB | RDS 托管 | EC2 自管 | 标准 EC2 路线 |

### Plan A ($10/月) 部署步骤

---

#### 第一步：AWS Lightsail 实例

```bash
# 1. AWS Console → Lightsail → Create Instance
#    Platform: Linux/Unix
#    Blueprint: OS Only → Amazon Linux 2023
#    Instance plan: $10 USD (1 vCPU, 2 GB, 60 GB, 3 TB)
#    Name: cs2match-server

# 2. SSH 进去装 Docker
ssh ec2-user@<lightsail-ip>
sudo yum update -y
sudo yum install -y docker
sudo systemctl enable docker && sudo systemctl start docker
sudo usermod -aG docker ec2-user
# 退出重新登录

# 3. 装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 第二步：配置 Lightsail 防火墙

Lightsail Console → 选中 `cs2match-server` → **Networking** → **IPv4 Firewall**：

| 应用 | 协议 | 端口 | 来源 | 说明 |
|------|------|------|------|------|
| SSH | TCP | 22 | **仅你自己的 IP** | 管理用 |
| Nakama Console | TCP | 7351 | **仅你自己的 IP** | 管理后台 |
| Nakama API | TCP | 7350 | **仅你自己的 IP** | 初始测试用，Cloudflare Tunnel 配置好后可关闭 |

> ⚠️ 7350 等 Cloudflare Tunnel 配置完毕并验证通过后，可以完全关闭公网访问。API 流量全部走隧道。

#### 第三步：创建项目目录结构

```bash
# 在 Lightsail 上
mkdir -p /home/ec2-user/cs2match/server/build
mkdir -p /home/ec2-user/cs2match/logs

# 目录结构:
# /home/ec2-user/cs2match/
#   ├── docker-compose.yml        # 手动上传（只含 db + nakama）
#   ├── .env                       # 环境变量
#   ├── nakama-config.yml          # 手动上传
#   ├── server/
#   │   └── build/
#   │       └── backend.so         # CI 自动更新
#   └── logs/
```

#### 第四步：上传配置文件并首次启动

在**本地**先用 scp 上传配置文件：

```bash
# 在本地项目根目录执行
scp docker-compose.prod.yml ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/docker-compose.yml
scp nakama-config.yml ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/
scp server/build/backend.so ec2-user@<LIGHTSAIL_IP>:/home/ec2-user/cs2match/server/build/
```

`docker-compose.prod.yml`（生产环境，不含前端，新增本地 `docker-compose.prod.yml` 文件）：

```yaml
# docker-compose.prod.yml — Lightsail 生产环境
# 只有 db + nakama，前端由 Cloudflare Pages 托管
services:
  db:
    image: postgres:15-alpine
    container_name: cs2match-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: nakama
      POSTGRES_USER: nakama
      POSTGRES_PASSWORD: ${DB_PASSWORD:-nakama}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - nakama-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U nakama -d nakama"]
      interval: 5s
      timeout: 5s
      retries: 5

  nakama:
    image: registry.heroiclabs.com/heroiclabs/nakama:3.30.0
    container_name: cs2match-nakama
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    entrypoint:
      - "/bin/sh"
      - "-ecx"
      - |
        /nakama/nakama migrate up \
          --database.address "nakama:${DB_PASSWORD:-nakama}@db:5432/nakama?sslmode=disable" \
          --logger.level DEBUG && \
        exec /nakama/nakama \
          --config /nakama/data/nakama-config.yml \
          --database.address "nakama:${DB_PASSWORD:-nakama}@db:5432/nakama?sslmode=disable" \
          --logger.level DEBUG \
          --runtime.path /nakama/data/modules \
          --console.username "${CONSOLE_USERNAME:-admin}" \
          --console.password "${CONSOLE_PASSWORD:-password}" \
          --socket.server_key "${NAKAMA_SERVER_KEY:-defaultkey}" \
          --runtime.http_key "${NAKAMA_SERVER_KEY:-defaultkey}" \
          2>&1 | tee /nakama/data/logs/nakama.log
    ports:
      - "7350:7350"
      - "7351:7351"
    environment:
      NAKAMA_SERVER_KEY: ${NAKAMA_SERVER_KEY:-defaultkey}
      CONSOLE_USERNAME: ${CONSOLE_USERNAME:-admin}
      CONSOLE_PASSWORD: ${CONSOLE_PASSWORD:-password}
    volumes:
      - ./nakama-config.yml:/nakama/data/nakama-config.yml:ro
      - ./server/build:/nakama/data/modules:ro
      - ./logs:/nakama/data/logs
    networks:
      - nakama-network
    healthcheck:
      test: ["CMD-SHELL", "kill -0 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

networks:
  nakama-network:
    driver: bridge

volumes:
  postgres_data:
    name: cs2match-postgres-data
```

在 Lightsail 上创建 `.env` 并启动：

```bash
cd /home/ec2-user/cs2match

cat > .env << 'EOF'
DB_PASSWORD=<生成一个随机密码，16+ 位>
NAKAMA_SERVER_KEY=<生成一个随机 key，32+ 位>
CONSOLE_USERNAME=admin
CONSOLE_PASSWORD=<生成一个随机密码>
EOF

# 启动
docker compose up -d

# 验证
docker compose ps
# cs2match-db       Up
# cs2match-nakama   Up

# 本地测试 API 是否可访问
curl http://<LIGHTSAIL_IP>:7350/healthcheck
# 应返回 {"status":"ok",...}
```

#### 第五步：Cloudflare Tunnel（API 走 HTTPS）

> **为什么需要**：Cloudflare Pages 强制 HTTPS，浏览器会拦截 HTTPS 页面里发出的 HTTP API 请求（Mixed Content）。Cloudflare Tunnel 免费解决这个矛盾。

**5.1 注册 Cloudflare 账号**

打开 [cloudflare.com](https://dash.cloudflare.com/sign-up)，用邮箱免费注册。

**5.2 （推荐）准备一个域名**

把域名托管到 Cloudflare DNS（免费）。便宜域名参考：`.xyz` ~$1/年、`.cfd` ~$2/年、`.click` ~$3/年。

也可以不买域名，用 Cloudflare 的临时隧道 URL（`*.trycloudflare.com`），但**每次重启会变**，不推荐用于生产。本指南假设你有域名 `yourdomain.com`。

**5.3 在 Lightsail 上安装 cloudflared**

```bash
# SSH 进 Lightsail
ssh ec2-user@<LIGHTSAIL_IP>

# 下载 cloudflared
sudo curl -L \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 \
  -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared

# 验证
cloudflared --version
```

**5.4 授权并创建隧道**

```bash
# 登录 Cloudflare（会给你一个 URL，在浏览器打开授权）
cloudflared tunnel login

# 创建隧道
cloudflared tunnel create cs2match-api
# 输出: Created tunnel cs2match-api with id <TUNNEL-UUID>
# 凭据文件保存在: /home/ec2-user/.cloudflared/<TUNNEL-UUID>.json
```

**5.5 配置隧道 DNS**

```bash
# 把 api.yourdomain.com 指向这个隧道
cloudflared tunnel route dns cs2match-api api.yourdomain.com
```

**5.6 写隧道配置文件**

```bash
mkdir -p /home/ec2-user/.cloudflared
cat > /home/ec2-user/.cloudflared/config.yml << 'EOF'
tunnel: <TUNNEL-UUID>                            # 替换为上一步输出的 UUID
credentials-file: /home/ec2-user/.cloudflared/<TUNNEL-UUID>.json

ingress:
  # API 流量转发到本地 Nakama
  - hostname: api.yourdomain.com
    service: http://localhost:7350

  # WebSocket 也需要转发（Nakama 的实时通信）
  - hostname: api.yourdomain.com
    service: http://localhost:7350
    originRequest:
      httpHostHeader: localhost

  # 其他请求一律拒绝
  - service: http_status:404
EOF
```

> ⚠️ Nakama 同时使用 HTTP（REST API）和 WebSocket（实时通信）。Cloudflare Tunnel 对 WebSocket 有原生支持，上述配置即可。

**5.7 安装为系统服务（开机自启）**

```bash
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
sudo systemctl status cloudflared
# 应显示 active (running)
```

**5.8 验证**

```bash
# 从本地（或任何外网）测试
curl https://api.yourDomain.com/healthcheck
# 应返回 {"status":"ok",...}

# 浏览器打开
# https://api.yourDomain.com:7351  → Nakama Console（如果有开）
```

#### 第六步：Cloudflare Pages（前端托管）

**6.1 创建 Pages 项目**

1. 打开 [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. 左侧 → **Workers & Pages** → **Create** → **Pages**
3. 选择 **Upload assets**（不要连 GitHub——我们通过 CI 手动部署）
4. Project name: `cs2match`
5. 点击 **Create project**

**6.2 获取 Cloudflare API Token 和 Account ID**

1. Cloudflare Dashboard 右上角 → **My Profile** → **API Tokens**
2. 点击 **Create Token** → 选择 **Custom Token**
3. 配置：
   - Token name: `cs2match-deploy`
   - Permissions: `Account` — `Cloudflare Pages` — `Edit`
   - Account Resources: 选你的账户
4. 点击 **Create Token**，复制保存（只显示一次！）

Account ID 在 Cloudflare Dashboard 首页右侧或 Workers & Pages 页面可以看到。

**6.3 配置前端环境变量**

前端构建时需要的变量，在 **Pages 项目 → Settings → Environment variables** 中添加（也可以在 CI 构建时注入）：

| 变量 | 值 | 说明 |
|------|------|------|
| `VITE_NAKAMA_HOST` | `api.yourdomain.com` | API 域名（Cloudflare Tunnel 给的） |
| `VITE_NAKAMA_PORT` | `443` | HTTPS 默认端口 |
| `VITE_NAKAMA_SERVER_KEY` | 你的 server key | 和第 4 步 `.env` 一致 |
| `VITE_NAKAMA_USE_SSL` | `true` | 走 HTTPS |

> 也可以不在 Cloudflare 上设这些变量，改为在 GitHub Actions 构建时用 Secret 注入——两种方式都行。CI 注入的方式更灵活（可以在构建前改值），所以推荐在 CI 中处理。

---

## 五、数据库备份（必须做）

Lightsail 不自带数据库备份，需要自己加：

```bash
# 在 Lightsail 上创建备份脚本
cat > /home/ec2-user/backup-db.sh << 'SCRIPT'
#!/bin/bash
BACKUP_DIR="/home/ec2-user/db-backups"
RETENTION_DAYS=7
mkdir -p "$BACKUP_DIR"

docker exec cs2match-db pg_dump -U nakama nakama \
  | gzip > "$BACKUP_DIR/nakama-$(date +%Y%m%d-%H%M).sql.gz"

# 删除 7 天前的备份
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: $(date)"
SCRIPT

chmod +x /home/ec2-user/backup-db.sh

# 每日凌晨 3 点备份
(crontab -l 2>/dev/null; echo "0 3 * * * /home/ec2-user/backup-db.sh >> /home/ec2-user/db-backups/backup.log 2>&1") | crontab -
```

> 建议同时把备份同步到 S3 或 Cloudflare R2（每月几美分）。

---

## 六、前端环境变量参考

```bash
# 开发环境 (client/.env)
VITE_NAKAMA_HOST=localhost
VITE_NAKAMA_PORT=7350
VITE_NAKAMA_SERVER_KEY=defaultkey
VITE_NAKAMA_USE_SSL=false

# 生产环境 — 通过 CI 注入（不在文件里存）
VITE_NAKAMA_HOST=api.yourdomain.com
VITE_NAKAMA_PORT=443
VITE_NAKAMA_SERVER_KEY=<your-key>
VITE_NAKAMA_USE_SSL=true
```

---

## 七、与 Oracle Cloud 的诚实对比

| 维度 | AWS Lightsail + Cloudflare | Oracle Cloud Always Free |
|------|------|------|
| 月费 | **$10** (+ $0-0.40 域名) | **$0** |
| 内存 | 2 GB | 24 GB |
| CPU | 1 vCPU | 4 OCPU |
| 存储 | 60 GB | 200 GB |
| 出站流量 | 3 TB + CF CDN 无限 | 10 TB |
| 前端 CDN | ✅ Cloudflare Pages 全球加速 | 需自己搭 |
| HTTPS | ✅ 自动（Tunnel + Pages） | 需自己配 |
| 免费性质 | 不是免费，是廉价 | **真正永久免费** |
| 稳定性 | AWS 商用级 | 偶尔抽风（ARM 资源紧张时难开新实例） |
| 注册难度 | 容易 | 信用卡验证严格，部分人被拒 |

> **一句话：如果你想一分钱不花，Oracle Cloud 是唯一选择。如果你愿意每月花 $10 买个安心，Lightsail + Cloudflare 更省心。**

---

## 八、附加可选服务

| 服务 | 月费 | 用途 |
|------|------|------|
| Cloudflare R2（类似 S3） | $0（10 GB 免费） | 数据库备份异地存储 |
| Cloudflare D1 | $0（5 GB 免费） | 可选轻量分析数据库 |
| Route 53 / Cloudflare DNS | $0 | DNS 托管 |
| 域名 | ~$0.10/月 ($1-5/年) | `yourdomain.com` |

---

## 九、总结

```
┌─────────────────────────────────────────────────────────┐
│  50 人同时在线的 CS2Match                                  │
│                                                          │
│  🏆 AWS Lightsail $10/月 (2 GB)          ← 后端           │
│  🏆 Cloudflare Pages $0                   ← 前端 CDN       │
│  🏆 Cloudflare Tunnel $0                  ← HTTPS API     │
│  💰 总成本: $10/月                                          │
│                                                          │
│  数据流量: 3 TB Lightsail + Cloudflare 无限 CDN             │
│  数据库: Docker volume 持久化，cron 每日备份                 │
└─────────────────────────────────────────────────────────┘
```
