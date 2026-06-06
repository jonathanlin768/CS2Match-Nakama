## ADDED Requirements

### Requirement: Vite + React + TypeScript 项目骨架

前端项目 SHALL 位于 `client/` 目录，使用 Vite 5 作为构建工具，React 18 作为 UI 框架，TypeScript 作为开发语言。

#### Scenario: 项目初始化

- **WHEN** 开发者在 `client/` 目录执行 `npm install` 然后 `npm run dev`
- **THEN** Vite 开发服务器在 `http://localhost:5173` 启动
- **AND** 浏览器打开后显示默认 React 欢迎页面

#### Scenario: TypeScript 类型检查

- **WHEN** 开发者执行 `npx tsc --noEmit`
- **THEN** 项目通过类型检查，无类型错误

### Requirement: Nakama JS SDK 集成

前端项目 SHALL 依赖 `@heroiclabs/nakama-js`，并提供 Nakama 客户端单例模块，默认连接 `http://localhost:7350`。

#### Scenario: 创建 Nakama 客户端

- **WHEN** 前端代码导入 Nakama 客户端模块
- **AND** 调用客户端创建函数，传入 server key `"defaultkey"` 和 host `"localhost"`、port `"7350"`、useSSL `false`
- **THEN** 返回已配置的 `Client` 实例

#### Scenario: 设备认证（Device Authentication）

- **WHEN** 用户首次访问前端应用
- **AND** 前端调用 `client.authenticateDevice(deviceId, true, "username")`
- **THEN** Nakama 服务器返回有效的 `Session` 对象
- **AND** Session 包含 `token` 和 `refresh_token`
- **AND** Session 信息持久化到 `localStorage`

#### Scenario: 认证失败的降级处理

- **WHEN** Nakama 服务器不可达
- **AND** 前端调用 `client.authenticateDevice`
- **THEN** 前端捕获错误并在控制台输出连接失败信息
- **AND** 不阻塞页面渲染（显示"连接服务器中..."状态）

### Requirement: 前端 Docker 构建与 Nginx 托管

前端项目 SHALL 提供 `Dockerfile` 和 `nginx.conf`，将 Vite 生产构建产物通过 Nginx 容器托管，支持通过 Docker Compose 一键启动。

#### Scenario: 构建前端 Docker 镜像

- **WHEN** 执行 `docker compose build frontend`（或 `docker build -t cs2match-frontend ./client`）
- **THEN** 多阶段 Docker 构建完成：Stage 1 使用 `node:18-alpine` 执行 `npm install && npm run build`
- **AND** Stage 2 使用 `nginx:alpine` 复制构建产物到 `/usr/share/nginx/html/`
- **AND** 自定义 `nginx.conf` 配置正确的 SPA 路由回退和 API 代理

#### Scenario: 前端在 Docker 内连接 Nakama

- **WHEN** 前端在 Docker 容器内运行
- **AND** Vite 构建时 `VITE_NAKAMA_HOST=nakama`（容器间通过 Docker 内网通信）
- **THEN** 前端 JS SDK 连接 `nakama:7350`（Docker 服务名，而非 `localhost`）
- **AND** 设备认证和 RPC 调用正常完成

#### Scenario: 前端页面路由回退

- **WHEN** 用户直接访问 `http://localhost:3000/game/123` 等 SPA 子路由
- **THEN** Nginx 返回 `index.html`（而非 404）
- **AND** React Router 处理后正常渲染对应页面

### Requirement: 前端开发环境连接 Nakama

Vite 开发服务器 SHALL 支持两种连接模式：Option B 通过 Docker 内网连接 Nakama；Option A 在宿主机开发时 SDK 直连 `localhost:7350`。

#### Scenario: SDK 直连 Nakama（Option A 宿主机开发）

- **WHEN** Vite 开发服务器运行在 `localhost:5173`
- **AND** Nakama 容器运行在 `localhost:7350`
- **THEN** 前端通过 `@heroiclabs/nakama-js` 可直接连接 Nakama
- **AND** 无需额外代理配置（Nakama SDK 自带连接能力）
