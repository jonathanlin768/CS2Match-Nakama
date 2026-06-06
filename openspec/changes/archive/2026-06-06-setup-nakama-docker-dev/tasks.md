## 1. Docker Compose 开发环境配置

- [x] 1.1 创建 `docker-compose.yml`，定义 `db`（PostgreSQL 15）、`nakama`（3.30.0）、`frontend`（Nginx）三个服务，配置 healthcheck
- [x] 1.2 创建 `nakama-config.yml`，配置 server key、Console 管理账号、数据库连接字符串、socket 端口
- [x] 1.3 创建 `.env.example` 模板文件，包含 DB_PASSWORD、NAKAMA_SERVER_KEY、CONSOLE_USERNAME、CONSOLE_PASSWORD、FRONTEND_PORT 等变量
- [x] 1.4 `.gitignore` 已存在且覆盖所需排除模式（.env、build/、node_modules/、dist/），无需额外修改
- [x] 1.5 验证：`docker compose up -d`，Nakama healthy，Console `http://localhost:7351` 可访问
- [x] 1.6 验证：执行 `docker compose down -v` 确认清理（容器/网络/volume 全部删除），再启动日志显示 "Successfully applied migration count=16"，数据库全新迁移

## 2. Go 插件项目骨架

- [x] 2.1 创建 `server/go.mod`，模块路径 `windypath.com/cs2match/server`，依赖 `nakama-common v1.40.0`（与 Nakama 3.30.0 精确匹配）
- [x] 2.2 创建 `server/main.go`，定义 `InitModule` 入口函数，注册 HealthCheck RPC
- [x] 2.3 创建 `server/build.sh`（Linux/Mac）编译脚本，使用 `nakama-pluginbuilder:3.30.0` 官方镜像编译
- [x] 2.4 创建 `server/build.ps1`（Windows WSL2）编译脚本
- [x] 2.5 配置 `docker-compose.yml` 中的 volume 挂载，将 `./server/build/` 挂载到 Nakama 容器的 `/nakama/data/modules/`
- [x] 2.6 编译验证：使用官方 `nakama-pluginbuilder:3.30.0` Docker 镜像编译成功，生成 `server/build/backend.so`（11MB, ELF Linux x86-64）
- [x] 2.7 加载验证：Nakama 日志确认插件加载 — "CS2Match Go plugin loaded successfully" + "HealthCheck RPC registered"

## 3. 前端 Docker 化（新人一键启动）

- [x] 3.1 创建 `client/Dockerfile`，多阶段构建：Stage 1 `node:22-alpine` 编译，Stage 2 `nginx:alpine` 托管
- [x] 3.2 创建 `client/nginx.conf`，配置 SPA 路由回退（`try_files $uri /index.html`）
- [x] 3.3 在 `docker-compose.yml` 中添加 `frontend` 服务，端口 `3000:80`，依赖 `nakama` 服务
- [x] 3.4 创建 `client/.env.example`，`VITE_NAKAMA_HOST` 统一用 `localhost`（浏览器通过宿主机端口映射连接 Nakama）
- [x] 3.5 Nakama 客户端从 `import.meta.env.VITE_NAKAMA_HOST` 读取 host，默认 `"localhost"`
- [x] 3.6 验证：Docker 前端全链路 — `docker compose up -d` + `localhost:3000`，设备认证成功 + HealthCheck 返回正确 JSON

## 4. React 前端项目骨架（Option A：宿主机开发）

- [x] 4.1 使用 Vite 官方模板初始化 `client/`（`npm create vite@latest client -- --template react-ts`）
- [x] 4.2 安装 `@heroiclabs/nakama-js` 依赖
- [x] 4.3 创建 `client/src/nakama.ts`，导出 Nakama 客户端单例
- [x] 4.4 创建 `client/src/hooks/useNakamaAuth.ts`，封装设备认证逻辑
- [x] 4.5 修改 `client/src/App.tsx`，自动设备认证 + 显示连接状态
- [x] 4.6 配置 Vite 环境变量（`client/.env.example`）
- [x] 4.7 验证（Option A）：`npm run dev` 启动成功，`localhost:5173` 可访问
- [x] 4.8 验证（Option A）：前端设备认证成功，Nakama Console 可看到新用户

## 5. HealthCheck RPC 实现

- [x] 5.1 在 `server/main.go` 中注册 `HealthCheck` RPC，返回 `{"status":"ok","timestamp":"<ISO8601>","version":"0.1.0"}`
- [x] 5.2 编译并重启 Nakama，确认 HealthCheck RPC 加载
- [x] 5.3 无认证访问：`curl "http://localhost:7350/v2/rpc/HealthCheck?http_key=defaultkey"` 返回 HTTP 200 + 正确 JSON
- [x] 5.4 有认证访问：设备认证后通过 Bearer token 调用 HealthCheck，返回正确 JSON

## 6. 全链路验证与文档

- [x] 6.1 在前端 `App.tsx` 中添加 HealthCheck RPC 调用（认证成功后自动执行）
- [x] 6.2 全链路验证（Option B Docker）：`docker compose up -d` → `localhost:3000` 显示"已连接" + HealthCheck 响应
- [x] 6.3 全链路验证（Option A 宿主机）：Docker Compose + Go 编译 + `npm run dev` → `localhost:5173` 显示"已连接" + HealthCheck 响应
- [x] 6.4 创建 `README.md`，包含系统要求、快速启动（双模式）、目录结构、常用命令
- [x] 6.5 验证：clean setup 按 README 从头执行 — `down -v` 清理后 `docker compose up -d`，Option B `localhost:3000` 显示"已连接" + HealthCheck `{"status":"ok"}`

---

### 关键发现

- **Go 依赖版本精确匹配**: Nakama 3.30.0 使用 `nakama-common v1.40.0`（非 v1.30.0）。必须通过 `go get github.com/heroiclabs/nakama/v3@v3.30.0` 获取精确依赖
- **编译镜像**: 必须使用官方 `registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0` 编译插件
- **编译命令**: `go build -buildmode=plugin -trimpath -o build/backend.so .`
- **Go 版本**: Nakama 3.30.0 使用 Go 1.24.5
