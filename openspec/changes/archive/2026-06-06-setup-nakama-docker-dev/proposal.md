## Why

项目已确定技术栈和架构方案（Nakama 3.30+ + Go Plugin + React 18 + PostgreSQL），需要从零搭建 Docker Compose 本地开发环境，使团队能够在统一、可复现的环境中开发、调试和测试全部组件。当前项目仓库中尚无任何 Docker 配置、后端或前端项目代码，属于基础设施搭建阶段。

## What Changes

- **新增** Docker Compose 编排文件，包含 Nakama (3.30+)、PostgreSQL (15+) 两个核心服务
- **新增** Nakama 配置文件（`nakama-config.yml`），配置端口、数据库连接、Console 管理账号、CORS 等开发期参数
- **新增** Go 项目骨架（`server/`），包含 `go.mod`、`main.go`（`InitModule` 入口）、编译脚本，产出 `.so` 插件供 Nakama 加载
- **新增** React + Vite + TypeScript 前端项目骨架（`client/`），集成 `@heroiclabs/nakama-js` 客户端 SDK
- **新增** 健康检查 RPC（`HealthCheck`），验证前后端连通性
- **新增** 开发工作流文档（启动、迁移、日志查看、插件重新编译等日常操作）
- **新增** `.env` 环境变量文件模板，管理开发期敏感配置（数据库密码、Nakama server key 等）

## Capabilities

### New Capabilities
- `dev-environment`: Docker Compose 开发环境一键启动，包含 Nakama、PostgreSQL 服务编排与健康检查
- `go-plugin-scaffold`: Go 项目骨架与编译工具链，可产出 `.so` 插件并挂载到 Nakama
- `react-frontend-scaffold`: React + Vite + TypeScript 前端项目骨架，集成 Nakama JS SDK 与基础认证流程
- `health-check-rpc`: 最小可用的 RPC 端点，用于验证前后端连通性和开发环境正确性

### Modified Capabilities
<!-- 无现有 capability 需要修改，本次为首次基础设施搭建 -->

## Impact

| 模块 | 影响说明 |
|------|---------|
| **部署 / Docker** | 新增 `docker-compose.yml`、`nakama-config.yml`、`.env` 文件，定义开发环境服务拓扑 |
| **Nakama 后端** | 新增 `server/` Go 项目目录，包含插件编译脚本和 `HealthCheck` RPC |
| **React 前端** | 新增 `client/` 目录，Vite + React + TypeScript 脚手架，集成 nakama-js SDK |
| **数据库** | PostgreSQL 15 容器自动初始化，Nakama 启动时自动执行 `migrate up` |
| **Luban 配置** | 本变更不涉及（配置表模块在后续 change 中搭建） |
| **状态同步 / 玩法** | 本变更不涉及（仅搭建连通性基础，游戏逻辑在后续 change 中开发） |
