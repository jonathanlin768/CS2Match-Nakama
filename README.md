# CS2 Simu Project

基于 Nakama 游戏服务器的 CS2 教练模拟对战项目。

## 技术栈

| 层级 | 技术 |
|------|------|
| 游戏服务器 | Nakama 3.30 |
| 服务器插件 | Go 1.24.5 (编译为 .so) |
| 前端 | React 19 + TypeScript + Vite 8 |
| 数据库 | PostgreSQL 15 |
| 容器化 | Docker + Docker Compose |

## 系统要求

- **Docker Desktop** 4.x+ (含 Docker Compose)
- 建议 4GB+ 内存分配给 Docker
- Windows Docker Desktop 使用 WSL2 后端；构建脚本无需手动进入 WSL 或在 WSL 中安装 Go

**可选**（仅 Option A 开发模式需要）：
- Node.js 22+（Vite 8 要求）
- Go 1.24+ (含 CGO 环境)

### 安装 OpenSpec CLI

项目使用 [OpenSpec](https://github.com/anthropics/openspec) 进行规范驱动开发（SDD），AI 辅助操作依赖 `openspec` 命令行工具。

```bash
npm install -g openspec-extensions
openspec --version   # 确认安装成功（当前版本: 1.4.1+）
```

---

## 快速启动

### 🚀 Option B: Docker 本地调试（推荐）

两台 Windows 开发设备都可以在克隆后让 Codex 执行 `$cs2match-local-update`，并说“启动本地调试”。技能会调用仓库脚本，自动创建缺失的 `.env`、编译 Go 插件、启动并检查 Docker 服务。也可直接在项目根目录运行：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/start-local-debug.ps1
```

脚本不会覆盖现有 `.env`，也不依赖宿主机 Go 或 Node.js。成功后网页在 `http://localhost:3000`，Nakama Console 在 `http://localhost:7351`，Go Remote 调试器在 `127.0.0.1:2345`。Linux/macOS 可先运行 `bash server/build.sh`，再运行 `docker compose up -d --build`（首次运行先从 `.env.example` 创建 `.env`）。

本地 Docker 构建默认从 `goproxy.cn` 下载 Go 模块，并由 Go 校验模块校验和；如设备使用其他可达代理，可设置 `CS2MATCH_GOPROXY` 环境变量覆盖。生产镜像的构建流程不使用此本地设置。

#### GoLand 断点调试

1. 在 GoLand 2026.2.3 打开仓库根目录；首次打开时为 `server/go.mod` 选择本机 Go 1.24.5 SDK。
2. 在运行配置下拉框选择项目共享的 **Nakama (Go Remote)**，点击 **Debug**。它连接 `127.0.0.1:2345`；停止调试连接时选择 **Leave it running**，容器里的 Delve 继续运行，之后可以重连。
3. 在 `server/internal/identity/hooks.go` 的 `BeforeAuthenticateDevice` 设置行断点，再用**新的浏览器无痕窗口**打开 `http://localhost:3000`。网页没有既有 Session 时会自动进行访客设备认证，触发该 Hook；普通刷新可能只恢复 Session。
4. 在 `server/internal/match/api_rpc.go` 的 `RPCSimuMatch` **返回的请求处理函数内**设置行断点，在网页进入“选择对战方式”，点击“电脑对战／开始模拟”，触发 `SimuMatch` RPC。

修改 Go 代码后，请让 Codex 执行 `$cs2match-local-update` 的“更新后端”，或手动重新运行 `server/build.ps1` 并重启 Nakama。Go 插件需要**重新编译和重启**；重启后在 GoLand 重新连接。断开调试前若程序停在断点，请先继续运行。前端需要保存即生效时，可按 Option A 使用 Vite HMR。

本地 `nakama-config.yml` 将 HTTP 写入超时设置为 10 分钟，方便暂停 Hook/RPC 请求并单步检查；它不用于生产镜像。调试暂停期间，当前请求及其他 Goroutine 可能暂时无法响应。

### 💻 Option A: 本地开发模式（活跃前端开发）

适合需要即时热更新（HMR）的前端开发：

```bash
# 1. 先按 Option B 启动本地调试环境；如使用 Vite，可停止 Docker 前端
docker compose stop frontend

# 2. 启动前端开发服务器
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
├── .agents/skills/cs2match-local-update/ # AI 本地更新技能
│
├── server/                  # Go 插件项目
│   ├── go.mod               # 模块: windypath.com/cs2match/server
│   ├── main.go              # InitModule 入口 + RPC 注册
│   ├── build.sh             # 编译脚本 (Linux/Mac)
│   ├── build.ps1            # 编译脚本 (Windows Docker)
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
│   ├── gen-config.ps1        # 导表脚本 (Windows PowerShell)
│   ├── start-local-debug.ps1 # 本地调试首次启动与就绪检查
│   └── update-local-config.ps1 # 导表并更新本地前后端
│
├── tools/luban/              # Luban Docker 镜像
│   └── Dockerfile
├── tools/map-semantic-editor/ # CS2 配置可视化编辑器
│   └── README.md              # 编辑器完整使用说明
│
├── client/                  # React 前端项目
│   ├── Dockerfile           # 前端 Docker 多阶段构建
│   ├── nginx.conf           # Nginx SPA 配置
│   ├── .env.example         # 前端环境变量模板
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.tsx
│       ├── index.css
│       ├── nakama.ts        # Nakama 客户端单例
│       ├── api/auth.ts      # 访客设备认证
│       ├── api/simu.ts      # SimuMatch RPC 客户端
│       ├── pages/MatchPage.tsx # 电脑对战入口
│       └── hooks/useNakamaAuth.ts # 认证状态 Hook
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
docker compose up -d           # 已完成首次构建后启动全部服务
docker compose up -d db nakama # 仅启动后端
docker compose ps              # 查看服务状态
docker compose logs -f nakama  # 查看 Nakama 日志
docker compose restart nakama  # 重启 Nakama（重新加载 .so 插件）
docker compose down            # 停止服务
docker compose down -v         # 停止并删除数据卷（清空数据库）
```

### Go 插件开发

从项目根目录执行：

```powershell
# Windows PowerShell / CMD（build.bat 也调用同一个 PowerShell 构建脚本）
.\server\build.ps1
docker compose up -d --build db nakama
docker compose restart nakama
docker compose logs --tail 100 nakama
```

```bash
# Linux / macOS / WSL2
bash server/build.sh
docker compose up -d --build db nakama
docker compose restart nakama
docker compose logs --tail 100 nakama
```

### 前端开发

```bash
cd client
npm run dev                    # 启动 Vite 开发服务器
npm run build                  # 生产构建（Option B 使用）
npx tsc --noEmit               # TypeScript 类型检查
```

### 让 AI 更新本地服务

项目内置 `.agents/skills/cs2match-local-update/SKILL.md`。支持直接对可发现项目技能的 AI 说：

- `帮我更新前端`
- `帮我更新后端`
- `帮我更新前后端`

这里的“更新”指重新构建、重新加载对应的本地 Docker 服务并验证状态，不代表让 AI 修改业务源码。前后端同时更新时会先更新后端，再更新前端；任一步失败都会停止并报告。

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

## CS2 配置可视化编辑器

`tools/map-semantic-editor` 是本地配置工作台，可编辑地图语义、选手、战队、新手战斗和其他 Luban 表。它会同时启动 Web UI 和仅监听本机的写入服务，不依赖 Nakama 也能编辑。

首次使用先安装依赖，然后从项目根目录启动：

```powershell
npm --prefix tools/map-semantic-editor ci
npm run config-editor
```

旧命令 `npm run map-editor` 等价。启动后访问 `http://127.0.0.1:5177`；本地写入服务使用 `127.0.0.1:5178`。

常用操作：

- `保存当前表` / `保存全部`：写入 `configs/Datas/#*.xlsx`，覆盖前自动备份。
- `运行导表`：运行 `scripts/gen-config.ps1`，生成 Go、TypeScript 和前后端 JSON 配置。
- `更新本地前后端`：依次导表、编译 `server/build/backend.so`、重载 Nakama、无缓存重建并重新创建 Docker 前端容器。
- 地图工程文件位于 `tools/map-semantic-editor/data/`；写入 Luban 后的工程快照位于 `configs/Datas/`。

验证编辑器：

```powershell
npm run map-editor:test
npm run config-editor:build
```

详细的数据规则、备份路径和地图编辑操作见 [`tools/map-semantic-editor/README.md`](tools/map-semantic-editor/README.md)。

---

## FAQ

### Q: `docker compose up -d` 后前端显示"连接中…"？

A: 前端 JS 在浏览器里执行，浏览器不认识 Docker 内部主机名（如 `nakama`）。`VITE_NAKAMA_HOST` 必须设置为 `localhost`，浏览器通过宿主机端口映射 `localhost:7350` 连接 Nakama。

### Q: 改了前端代码后需要 `docker compose build` 吗？

A: 看你怎么跑：
- **Option A**（`npm run dev`）：不需要，Vite HMR 自动热更新，保存即时生效。
- **Option B**（Docker 前端）：需要重新构建镜像并重启容器。

**Docker 前端更新命令**：

```bash
# 方式 1：完全重建（最可靠，确保所有层重新构建）
docker compose build --no-cache frontend
docker compose up -d --no-deps --force-recreate frontend

# 方式 2：增量重建（Docker 自动检测文件变更）
docker compose build frontend
docker compose up -d --no-deps --force-recreate frontend
```

> ⚠️ `docker compose up -d --build` 有时会用缓存层跳过重新编译（`COPY . .` 显示 CACHED）。如果改了前端代码但没生效，用 **方式 1** `--no-cache` 强制重建。

**为什么不用 `docker compose up -d --build` 一条命令？** Docker 的构建缓存可能误判文件未变更，导致修改不生效。分两步执行：先 build 确保新镜像，再 up 重启容器。

日常开发推荐 Option A（`npm run dev`），保存即时生效，无需等 Docker 构建。

### Q: 改了 Go 后端代码怎么重新部署？

A: 必须同时完成“生成新的 Linux `.so`”和“重启 Nakama 加载新插件”。请在**项目根目录**执行与当前终端匹配的命令。

Windows PowerShell / CMD：

```powershell
.\server\build.ps1
docker compose up -d --build db nakama
docker compose restart nakama
docker compose logs --tail 100 nakama
```

Linux / macOS / WSL2：

```bash
bash server/build.sh
docker compose up -d --build db nakama
docker compose restart nakama
docker compose logs --tail 100 nakama
```

`server/build.ps1`、`server/build.bat` 和 `server/build.sh` 均使用 `heroiclabs/nakama-pluginbuilder:3.30.0`，以 `-mod=mod -buildmode=plugin -trimpath -gcflags='all=-N -l'` 编译本地调试插件。生产镜像使用独立的 `server/Dockerfile.prod`，不会加入 Delve 或调试编译参数。

如果只想确认编译产物已刷新，可查看 `server/build/backend.so` 的修改时间；若日志显示插件加载失败，以 `docker compose logs --tail 100 nakama` 的错误为准。

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
Nakama 与 pluginbuilder 使用 Docker Hub 上的 Heroic Labs 官方 `heroiclabs/*` 镜像；两者必须保持相同的 Nakama 版本。若 Docker Hub 返回限流，请稍后重试或登录 Docker Hub，不要随意替换 pluginbuilder 版本。
