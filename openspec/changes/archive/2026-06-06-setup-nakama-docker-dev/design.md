## Context

项目处于初始化阶段，代码仓库中尚无任何服务代码或 Docker 配置。技术栈已在 `doc/cs2SimuProject.md` 中确定：Nakama 3.30+ 作为游戏服务器、PostgreSQL 15+ 作为数据库、Go 1.22+ 编写服务器插件、React 18 + Vite 5 + TypeScript 构建前端。需要从零搭建一套本地 Docker 开发环境，使开发者执行一条命令即可启动全部服务并开始开发。

## Goals / Non-Goals

**Goals:**
- 通过 `docker compose up` 一键启动 Nakama + PostgreSQL + 前端（Docker Nginx 托管）完整开发环境
- Go 插件项目骨架，支持编译 `.so` 并挂载到 Nakama 容器
- React + Vite + TypeScript 前端项目骨架，集成 Nakama JS SDK 完成设备认证
- 一个 `HealthCheck` RPC 验证前后端全链路连通
- 开发期环境变量集中管理（`.env` 文件）

**Non-Goals:**
- 生产部署配置（生产环境 AWS 部署在后续 change 中处理）
- CockroachDB 支持（已决策使用 PostgreSQL）
- Luban 配置表模块搭建（后续 change）
- 游戏逻辑 / Match Handler / 状态同步实现（后续 change）
- CI/CD 流水线（后续 change）
- HTTPS/TLS 配置（开发环境使用 HTTP）
- 前端 UI 组件库选择和搭建（仅脚手架）

## Decisions

### D1: Docker Compose 服务拓扑

**选择**: 三个核心服务 — `db`（PostgreSQL）、`nakama`（Nakama 3.30+）、`frontend`（Nginx 托管 React 静态资源），通过 Docker 内部网络通信。`frontend` 服务包含两种运行模式（见 D3）。

**理由**:
- Nakama 官方推荐的开发环境部署方式（db + nakama）
- PostgreSQL 容器不暴露端口到宿主机（仅 Nakama 通过内部网络访问），减少端口冲突和安全风险
- Nakama 暴露 7350 (HTTP/WS) 和 7351 (Console) 到宿主机
- 新增 `frontend` 服务默认为新人快速上手设计（Option B），暴露 3000 端口到宿主机；活跃开发时可停止该服务改用 Vite dev server（Option A）

**备选**: 使用 Nakama 官方 Docker 镜像自带的 CockroachDB 旁路。**不采用** — 项目已决策使用 PostgreSQL。

### D2: Go 插件编译与挂载策略

**选择**: 使用多阶段构建 — 宿主机编译 Go 插件为 `.so`，通过 Docker volume 挂载到 Nakama 容器的 `/nakama/data/modules/` 目录。Nakama 启动时自动加载该目录下的所有 `.so` 文件。

**理由**:
- 与 Nakama 官方推荐的 Go Runtime 加载方式一致，兼容性无问题
- 开发迭代时只需重新编译 `.so` 并重启 Nakama 容器（`docker compose restart nakama`），无需重新构建镜像
- Go 编译在宿主机进行，利用本地 Go 工具链和 IDE 支持，开发体验更好

**备选**: 在 Nakama 容器内编译 Go 插件。**不采用** — 需要在容器内安装 Go 工具链，增加镜像体积和复杂度，IDE 调试支持差。

### D3: 前端开发模式（双模式）

**提供两种前端运行模式，开发者按场景选择：**

| 维度 | Option A: Vite Dev Server（宿主机） | Option B: Docker Nginx（容器化） |
|------|-------------------------------------|----------------------------------|
| **启动命令** | `cd client && npm run dev` | `docker compose up -d`（自动启动） |
| **访问地址** | `http://localhost:5173` | `http://localhost:3000` |
| **HMR 热更新** | ✅ 原生支持，保存即刷新 | ❌ 需手动 `npm run build` 后重建镜像 |
| **适用场景** | 活跃前端开发，需要即时反馈 | 新人快速体验项目、后端开发不需改前端 |
| **Nakama 连接** | SDK 直连 `localhost:7350` | SDK 直连 `localhost:7350`（浏览器通过宿主机端口映射） |
| **前置依赖** | 需安装 Node.js 18+ | 仅需 Docker |

**Option A（Vite Dev Server）— 活跃前端开发首选**:
- Vite HMR 极快，保存即刷新，开发体验最佳
- `@heroiclabs/nakama-js` SDK 直连 `localhost:7350` 上的 Nakama 容器
- 无需为前端额外配置 Docker 服务

**Option B（Docker Nginx）— 新人上手 / 后端开发首选**:
- 前端静态资源由 Nginx 容器托管，通过 `docker compose up -d` 一键启动全部服务
- 前端构建时通过 `VITE_NAKAMA_HOST=nakama` 将 API 指向容器内网服务名
- Nginx 暴露 3000 端口到宿主机
- 陌生人 clone 项目后仅需 Docker，零 Node.js 依赖即可看到完整应用

**实施策略**:
- `docker-compose.yml` 中 `frontend` 服务默认启用（Option B）
- 前端开发者可选择性停止 `frontend` 容器改用 Option A
- 两种模式共享同一套源代码（`client/src/`），无代码分叉

### D4: Nakama 配置管理

**选择**: 使用独立 `nakama-config.yml` 文件 + 环境变量插值。`nakama-config.yml` 通过 volume 挂载到容器内 `/nakama/data/nakama-config.yml`。

**理由**:
- Nakama 官方支持 YAML 配置文件和环境变量两种方式，YAML 更结构化
- 敏感值（数据库密码、server key）通过 `{{env.NX_VAR_NAME}}` 语法引用 Docker 环境变量
- 配置文件可版本控制，环境差异通过 `.env` 隔离

### D5: Go 模块路径

**选择**: Go module 路径为 `windypath.com/cs2match/server`。Nakama 导入路径使用 `github.com/heroiclabs/nakama-common`。前端 npm 包名遵循 `@windypath/` scope（后续按需引入）。

**理由**:
- 项目统一使用 `windypath.com` 作为基域名，所有包名以该域名为前缀，如 `windypath.com/cs2match/server`、`windypath.com/cs2match/luban` 等
- Go 插件必须使用 `nakama-common` 包的接口定义（`runtime.Initializer`、`runtime.NakamaModule` 等）
- 本地开发使用组织域名作为模块路径，明确代码归属，避免与 GitHub 等其他仓库地址混淆

### D6: npm 包选型

**选择**: 前端仅引入 `@heroiclabs/nakama-js` 作为核心依赖，以及 `react`、`react-dom`、`vite`、`typescript` 作为脚手架基础依赖。状态管理（Zustand）和路由（React Router）在后续 UI 开发 change 中引入。

**理由**:
- 本 change 目标是连通性验证，最小依赖原则
- `@heroiclabs/nakama-js` 是 Nakama 官方 JS SDK，支持 WebSocket、RPC、认证等全部 API
- 无替代方案 — 官方 SDK 是唯一成熟的 Nakama JS 客户端库

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| **Docker Desktop 资源消耗**: Windows/Mac 上 Docker 占用内存较大（建议 4GB+） | 在 README 中标注系统要求和 WSL2 配置建议 |
| **Go 插件跨平台编译**: 宿主机的 Go 编译器必须输出 Linux amd64 格式的 `.so`（因为 Nakama 容器是 Linux） | 编译脚本中使用 `GOOS=linux GOARCH=amd64` 环境变量；Windows 用户通过 WSL2 编译 |
| **Nakama 版本锁定**: 使用 `3.30.0` 固定镜像 tag 而非 `latest`，避免意外升级导致不兼容 | 后续按需升级并在 change 中记录 |
| **PostgreSQL 数据持久化**: 容器删除时数据丢失 | docker-compose 配置 named volume 持久化数据库文件；提供 `docker compose down -v` 命令说明（清理数据时使用） |
| **端口冲突**: 7350/7351 可能与本地其他服务冲突 | 在 README 中记录端口占用及如何修改 `.env` |
| **前端 Docker 模式无 HMR**: Option B 下修改前端代码需重建镜像才能看到变更 | 默认启动 Option B 供快速体验；活跃前端开发时推荐 Option A（Vite dev server），README 中明确两种模式的切换方式 |
