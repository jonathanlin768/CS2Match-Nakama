## MODIFIED Requirements

### Requirement: 前端 Docker 构建与 Nginx 托管

前端项目 SHALL 提供 `Dockerfile`（node:22-alpine 多阶段构建）和 `nginx.conf`（SPA 路由回退），构建产物由 Nginx 容器托管。开发构建 SHALL 默认使用 `localhost:7350`；非本地或生产构建 SHALL 显式注入 `VITE_NAKAMA_HOST`、`VITE_NAKAMA_PORT`、`VITE_NAKAMA_SERVER_KEY` 和 `VITE_NAKAMA_USE_SSL`，不得把开发默认地址误发布给远程用户。相同的显式生产变量 SHALL 适用于发布到 Cloudflare Pages 的静态构建。

#### Scenario: 构建开发前端 Docker 镜像

- **WHEN** 开发者执行 `docker compose up -d --build`
- **THEN** 多阶段 Docker 构建完成：Stage 1 使用 `node:22-alpine` 执行 `npm ci && npm run build`
- **AND** Stage 2 使用 `nginx:alpine` 托管构建产物
- **AND** 自定义 `nginx.conf` 配置 SPA 路由回退（`try_files $uri /index.html`）
- **AND** 未提供生产变量时前端继续连接本地 `localhost:7350`

#### Scenario: 构建生产前端产物

- **GIVEN** 构建环境提供生产 API 域名、443 端口、匹配的 Nakama server key 和 SSL 开关
- **WHEN** CI 构建 Cloudflare Pages 静态产物或生产前端容器
- **THEN** 产物中的 Nakama 客户端连接生产 API 域名并启用 TLS
- **AND** 构建验证失败而不是回退到 `localhost:7350` 或 `defaultkey`

#### Scenario: 前端页面路由回退

- **WHEN** 用户直接访问 `http://localhost:3000/game/123` 等 SPA 子路由
- **THEN** Nginx 返回 `index.html`（而非 404）
