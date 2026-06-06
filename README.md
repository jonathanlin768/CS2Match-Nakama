# CS2 Simu Project

基于 Nakama 游戏服务器的 CS2 教练模拟对战项目。

## 技术栈

| 层级 | 技术 |
|------|------|
| 游戏服务器 | Nakama 3.30 |
| 服务器插件 | Go 1.24.5 (编译为 .so) |
| 前端 | React 18 + TypeScript + Vite 5 |
| 数据库 | PostgreSQL 15 |
| 容器化 | Docker + Docker Compose |

## 系统要求

- **Docker Desktop** 4.x+ (含 Docker Compose)
- 建议 4GB+ 内存分配给 Docker
- Windows 用户需启用 WSL2

**可选**（仅 Option A 开发模式需要）：
- Node.js 22+（Vite 8 要求）
- Go 1.24+ (含 CGO 环境)

---

## 快速启动

### 🚀 Option B: Docker 一键启动（推荐新人）

无需安装 Node.js 或 Go，只需 Docker：

```bash
# 1. 克隆项目
git clone <repo-url> && cd cs2-simu-project

# 2. 创建环境变量（使用默认值）
cp .env.example .env

# 3. 一键启动全部服务
docker compose up -d

# 4. 访问
#    前端:   http://localhost:3000
#    Console: http://localhost:7351  (admin / password)
#    API:    http://localhost:7350
```

三条命令即可看到完整应用。

### 💻 Option A: 本地开发模式（活跃前端开发）

适合需要即时热更新（HMR）的前端开发：

```bash
# 1. 启动后端服务
cp .env.example .env
docker compose up -d db nakama

# 2. 编译 Go 插件
cd server
bash build.sh        # Linux/Mac
# 或 Windows WSL2: wsl bash build.sh
cd ..

# 3. 重启 Nakama 加载插件
docker compose restart nakama

# 4. 启动前端开发服务器
cd client
cp .env.example .env
npm install
npm run dev          # http://localhost:5173
```

---

## 目录结构

```
cs2-simu-project/
├── docker-compose.yml       # Docker Compose 编排
├── nakama-config.yml        # Nakama 服务器配置
├── .env.example             # 环境变量模板
├── README.md
│
├── server/                  # Go 插件项目
│   ├── go.mod               # 模块: windypath.com/cs2match/server
│   ├── main.go              # InitModule 入口 + RPC 注册
│   ├── build.sh             # 编译脚本 (Linux/Mac)
│   ├── build.ps1            # 编译脚本 (Windows WSL2)
│   └── build/               # 编译产物 (.so)
│
├── configs/                 # 策划配表 (Excel) + Luban 配置
│   ├── luban.conf            # Luban 主配置 (groups/schemaFiles/targets)
│   ├── Defines/              # 内置类型 XML 扩展
│   └── Datas/                # 配表数据
│       ├── __tables__.xlsx   # 表定义
│       ├── __beans__.xlsx    # 结构定义
│       ├── __enums__.xlsx    # 枚举定义
│       └── *.xlsx            # 业务数据表
│
├── scripts/                  # 工具脚本
│   ├── gen-config.sh         # 导表脚本 (Linux/Mac/Git Bash)
│   └── gen-config.ps1        # 导表脚本 (Windows PowerShell)
│
├── tools/luban/              # Luban Docker 镜像
│   └── Dockerfile
│
├── client/                  # React 前端项目
│   ├── Dockerfile           # 前端 Docker 多阶段构建
│   ├── nginx.conf           # Nginx SPA 配置
│   ├── .env.example         # 前端环境变量模板
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx          # 主页面 (连接状态 + HealthCheck)
│       ├── App.css
│       ├── main.tsx
│       ├── index.css
│       ├── nakama.ts        # Nakama 客户端单例
│       └── hooks/
│           └── useNakamaAuth.ts  # 认证 Hook
│
├── doc/                     # 项目文档
│   └── cs2SimuProject.md
│
└── openspec/                # OpenSpec 变更管理
    ├── config.yaml
    └── changes/
```

---

## 常用命令

### Docker 服务管理

```bash
docker compose up -d           # 启动全部服务
docker compose up -d db nakama # 仅启动后端
docker compose ps              # 查看服务状态
docker compose logs -f nakama  # 查看 Nakama 日志
docker compose restart nakama  # 重启 Nakama（重新加载 .so 插件）
docker compose down            # 停止服务
docker compose down -v         # 停止并删除数据卷（清空数据库）
```

### Go 插件开发

```bash
cd server
bash build.sh                  # 编译插件
docker compose restart nakama  # 重新加载
docker compose logs nakama     # 查看加载日志
```

### 前端开发

```bash
cd client
npm run dev                    # 启动 Vite 开发服务器
npm run build                  # 生产构建（Option B 使用）
npx tsc --noEmit               # TypeScript 类型检查
```

### API 验证

```bash
# HealthCheck RPC (无需认证，通过 http_key)
curl "http://localhost:7350/v2/rpc/HealthCheck?http_key=defaultkey" \
  -X POST -H "Content-Type: application/json" -d '""'

# 带认证的调用（先获取 token）
curl -s -X POST "http://localhost:7350/v2/account/authenticate/device?create=true" \
  -u "defaultkey:" -H "Content-Type: application/json" \
  -d '{"id":"your-device-id-12345"}'
```

---

## 端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| 3000 | 前端 (Nginx) | Option B Docker 模式，SPA 静态资源 |
| 7350 | Nakama API | HTTP REST + WebSocket |
| 7351 | Nakama Console | Web 管理后台 |
| 5173 | Vite Dev Server | Option A 前端开发模式 |

---

## 环境变量

所有可配置变量见 `.env.example`。主要变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_PASSWORD` | nakama | PostgreSQL 密码 |
| `NAKAMA_SERVER_KEY` | defaultkey | Nakama 服务器密钥 |
| `CONSOLE_USERNAME` | admin | Console 管理账号 |
| `CONSOLE_PASSWORD` | password | Console 管理密码 |
| `FRONTEND_PORT` | 3000 | 前端 Nginx 端口 |
| `VITE_NAKAMA_HOST` | localhost | 前端 JS 连接 Nakama 的地址（浏览器里跑，用 localhost） |

---

## 切换 Option A ↔ Option B

**从 Option B 切换到 Option A：**
```bash
docker compose stop frontend   # 停止 Docker 前端
cd client && npm run dev       # 启动本地 Vite
```

**从 Option A 切换到 Option B：**
```bash
# 停止 npm run dev (Ctrl+C)
docker compose up -d frontend  # 启动 Docker 前端
```

---

## 策划配表（Luban）

项目使用 [Luban 4.x](https://github.com/focus-creative-games/luban) 管理游戏配置表，通过 **Docker** 运行，**无需安装 .NET SDK**。

### 导表命令

```bash
# Windows (PowerShell)
.\scripts\gen-config.ps1

# Linux / macOS / Git Bash
bash scripts/gen-config.sh
```

### 导表流程

```
configs/Datas/*.xlsx  ──►  Luban (Docker)  ──►  server/config/ (Go 代码 + JSON)
                                           ──►  client/src/config/ (TS 类型)
                                           ──►  client/public/data/config/ (JSON)
```

### 添加新配置表

1. 在 `configs/Datas/` 下创建 Excel 数据表（如 `#skill.xlsx`）
2. 在 `configs/Datas/__beans__.xlsx` 中定义数据结构
3. 在 `configs/Datas/__tables__.xlsx` 中注册新表
4. 运行导表脚本
5. 在代码中使用 `cfg.Global`（Go）或 `loadConfig()`（TS）访问

### 目录约定

| 目录 | 说明 | Git |
|------|------|-----|
| `configs/` | 策划 Excel 源文件 + luban.conf | ✅ 提交 |
| `server/config/` | Luban 生成的 Go 代码 | ❌ gitignore |
| `server/config/data/` | Luban 生成的 JSON 数据 (嵌入 .so) | ❌ gitignore |
| `client/src/config/` | Luban 生成的 TS 类型 | ❌ gitignore |
| `client/public/data/config/` | Luban 生成的 JSON 数据 | ❌ gitignore |

---

## FAQ

### Q: `docker compose up -d` 后前端显示"连接中…"？

A: 前端 JS 在浏览器里执行，浏览器不认识 Docker 内部主机名（如 `nakama`）。`VITE_NAKAMA_HOST` 必须设置为 `localhost`，浏览器通过宿主机端口映射 `localhost:7350` 连接 Nakama。

### Q: 改了前端代码后需要 `docker compose build` 吗？

A: 看你怎么跑：
- **Option A**（`npm run dev`）：不需要，Vite HMR 自动热更新，保存即时生效。
- **Option B**（Docker 前端）：需要 `docker compose up -d --build`，前端代码是构建时打进镜像的。

日常开发推荐 Option A。

### Q: 改了 Go 后端代码怎么重新部署？

A:
```bash
# 1. 重新编译
docker run --rm --entrypoint "" \
  -v "$(pwd)/server:/app" -w /app \
  registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0 \
  go build -buildmode=plugin -trimpath -o build/backend.so .

# 2. 重启 Nakama 加载新插件
docker compose restart nakama
```

### Q: `docker compose stop frontend` 为什么能停掉 `cs2match-frontend` 容器？

A: `docker-compose.yml` 里定义的叫**服务名**（`frontend`），`container_name` 是给 `docker ps` 看的别名。`docker compose` 所有子命令都跟服务名交互，`docker`（不加 compose）才用容器名。

### Q: 怎么停掉本项目的所有容器？

A:
| 命令 | 效果 |
|------|------|
| `docker compose stop` | 停，不删容器 |
| `docker compose down` | 停 + 删容器和网络（数据保留） |
| `docker compose down -v` | 停 + 删容器 + 删数据库（⚠️ 数据清空） |

### Q: `docker compose down -v` 后怎么确认清理干净了？

A:
```bash
docker compose ps                    # 应该无输出
docker volume ls | grep cs2match     # volume 已删除，无输出
```

### Q: Docker Hub 拉不到镜像（`node`、`nginx` 等）怎么办？

A: 网络限制时可通过 DaoCloud 镜像站中转：
```bash
docker pull docker.m.daocloud.io/node:22-alpine
docker tag docker.m.daocloud.io/node:22-alpine node:22-alpine

docker pull docker.m.daocloud.io/nginx:alpine
docker tag docker.m.daocloud.io/nginx:alpine nginx:alpine
```
Nakama 相关镜像走 Heroic Labs 官方仓库 `registry.heroiclabs.com`，不受影响。
