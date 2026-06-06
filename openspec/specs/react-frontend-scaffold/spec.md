# React Frontend Scaffold

React + Vite + TypeScript 前端项目骨架规格 — 项目初始化、Nakama SDK 集成、Docker 托管。

## Requirements

### Requirement: Vite + React + TypeScript 项目骨架

前端项目 SHALL 位于 `client/` 目录，使用 Vite 作为构建工具，React 18+ 作为 UI 框架，TypeScript 作为开发语言。项目 SHALL 引入 Tailwind CSS 4 作为样式框架，react-router-dom v7 管理路由，shadcn/ui 作为组件库。

#### Scenario: 项目初始化

- **WHEN** 开发者在 `client/` 目录执行 `npm install` 然后 `npm run dev`
- **THEN** Vite 开发服务器在 `http://localhost:5173` 启动
- **AND** 浏览器打开后显示 React 应用页面

#### Scenario: TypeScript 类型检查

- **WHEN** 开发者执行 `npx tsc --noEmit`
- **THEN** 项目通过类型检查，无类型错误

#### Scenario: 前端项目目录结构

- **WHEN** 开发者查看 `client/src/` 目录
- **THEN** 目录包含 `pages/`（5 个页面）、`components/cs2/`（CS2 业务组件）、`components/ui/`（shadcn/ui 组件）、`layout/`（AppLayout）、`hooks/`（自定义 hooks）、`lib/`（工具函数）、`types/`（类型定义）
- **AND** `src/` 下存在 `main.tsx`（入口，含路由配置）和 `index.css`（全局样式）

### Requirement: Nakama JS SDK 集成

前端项目 SHALL 依赖 `@heroiclabs/nakama-js`，提供 Nakama 客户端单例模块，连接地址从 `VITE_NAKAMA_HOST` 环境变量读取，默认 `localhost:7350`，以及邮箱认证和 Session 恢复的 API 封装。

#### Scenario: 创建 Nakama 客户端

- **WHEN** 前端代码导入 Nakama 客户端模块
- **AND** 调用客户端创建函数，传入 server key `"defaultkey"` 和 host `"localhost"`、port `"7350"`、useSSL `false`
- **THEN** 返回已配置的 `Client` 实例

#### Scenario: 邮箱注册（Email Registration）

- **WHEN** 用户提交邮箱和密码进行注册
- **AND** 前端调用 `client.authenticateEmail(email, password, true)`
- **THEN** Nakama 服务器返回有效的 `Session` 对象（账号不存在时自动创建）
- **AND** `session.created` 为 `true`（新账号）或 `false`（已存在）
- **AND** Session 信息持久化到 `localStorage`

#### Scenario: 邮箱登录（Email Authentication）

- **WHEN** 用户提交已注册的邮箱和密码进行登录
- **AND** 前端调用 `client.authenticateEmail(email, password, false)`
- **THEN** Nakama 服务器验证凭证并返回有效的 `Session` 对象
- **AND** Session 信息持久化到 `localStorage`

#### Scenario: 认证失败的错误处理

- **WHEN** Nakama 服务器不可达或认证信息无效
- **AND** 前端调用 `authenticateEmail`
- **THEN** 前端捕获错误并返回错误信息给调用方
- **AND** 不阻塞页面渲染（LoginPage 显示错误提示）

#### Scenario: Session 恢复

- **WHEN** 前端应用初始化时 localStorage 中存在 token 和 refresh_token
- **AND** 调用 Session 恢复逻辑
- **THEN** token 未过期则本地恢复，零网络请求
- **AND** token 过期但 refresh_token 有效则刷新 Session
- **AND** 恢复成功则设置当前 session 并跳转受保护页面
- **AND** 恢复失败则清除过期 token，展示登录页

### Requirement: 前端 Docker 构建与 Nginx 托管

前端项目 SHALL 提供 `Dockerfile`（node:22-alpine 多阶段构建）和 `nginx.conf`（SPA 路由回退），构建产物由 Nginx 容器托管。`VITE_NAKAMA_HOST` 设为 `localhost`（前端 JS 在浏览器中执行，通过宿主机端口映射连接 Nakama）。

#### Scenario: 构建前端 Docker 镜像

- **WHEN** 执行 `docker compose up -d --build`
- **THEN** 多阶段 Docker 构建完成：Stage 1 使用 `node:22-alpine` 执行 `npm ci && npm run build`
- **AND** Stage 2 使用 `nginx:alpine` 托管构建产物
- **AND** 自定义 `nginx.conf` 配置 SPA 路由回退（`try_files $uri /index.html`）

#### Scenario: 前端页面路由回退

- **WHEN** 用户直接访问 `http://localhost:3000/game/123` 等 SPA 子路由
- **THEN** Nginx 返回 `index.html`（而非 404）

### Requirement: 前端开发环境连接 Nakama

前端 JS SDK 在浏览器中执行，连接地址统一使用 `localhost`（浏览器通过宿主机端口映射连接 Nakama 容器）。Vite 开发服务器（Option A）和 Docker Nginx（Option B）均适用。

#### Scenario: SDK 直连 Nakama（Option A 宿主机开发）

- **WHEN** Vite 开发服务器运行在 `localhost:5173`
- **AND** Nakama 容器运行在 `localhost:7350`
- **THEN** 前端通过 `@heroiclabs/nakama-js` 直接连接 Nakama
- **AND** 设备认证和 RPC 调用正常完成
