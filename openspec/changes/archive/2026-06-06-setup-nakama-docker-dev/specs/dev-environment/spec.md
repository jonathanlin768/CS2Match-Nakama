## ADDED Requirements

### Requirement: Docker Compose 一键启动全部服务

系统 SHALL 提供 `docker-compose.yml` 编排文件，开发者执行 `docker compose up -d` 后，Nakama、PostgreSQL 和前端（Nginx）三个服务自动启动并就绪。

#### Scenario: 首次启动环境

- **WHEN** 开发者在项目根目录执行 `docker compose up -d`
- **THEN** Docker 自动拉取 Nakama 3.30+ 和 PostgreSQL 15+ 镜像
- **AND** PostgreSQL 容器先启动并完成初始化
- **AND** Nakama 容器启动后自动执行 `nakama migrate up` 完成数据库迁移
- **AND** Nakama 进程启动，监听 7350 (HTTP/WebSocket)、7351 (Console) 端口
- **AND** 前端 Nginx 容器启动，监听 3000 端口，可通过 `http://localhost:3000` 访问
- **AND** 执行 `docker compose ps` 显示三个服务状态均为 healthy

#### Scenario: 宿主机访问 Nakama API

- **WHEN** Nakama 容器已启动
- **THEN** 开发者可通过 `http://localhost:7350` 访问 Nakama REST API
- **AND** 可通过 `http://localhost:7351` 访问 Nakama Web Console

#### Scenario: PostgreSQL 仅内部可访问

- **WHEN** 开发环境已运行
- **THEN** PostgreSQL 端口（5432）不暴露到宿主机
- **AND** 仅 Nakama 容器通过 Docker 内部网络连接数据库

### Requirement: 数据库数据持久化

系统 SHALL 使用 Docker named volume 持久化 PostgreSQL 数据，确保容器重启后数据不丢失。

#### Scenario: 容器重启后数据保留

- **WHEN** 开发者重启 Nakama 和数据库容器（`docker compose restart`）
- **THEN** 之前创建的用户、存储的数据仍然存在
- **AND** 无需重新执行数据库迁移

#### Scenario: 显式清理数据

- **WHEN** 开发者执行 `docker compose down -v`
- **THEN** 所有容器和 named volume 被删除
- **AND** 下次启动时重新执行迁移，数据库为空白状态

### Requirement: 环境变量集中管理

系统 SHALL 提供 `.env` 文件模板（`.env.example`），集中管理开发期配置变量，包括数据库密码、Nakama server key、Console 管理员账号密码。

#### Scenario: 从模板创建配置

- **WHEN** 开发者复制 `.env.example` 为 `.env`
- **AND** 执行 `docker compose up -d`
- **THEN** Nakama 使用 `.env` 中定义的数据库密码和 server key
- **AND** Console 管理员账号密码与 `.env` 中配置一致

### Requirement: 前端 Docker 化服务（新人快速上手）

系统 SHALL 在 `docker-compose.yml` 中提供 `frontend` 服务，使用 Nginx 托管前端生产构建产物，使新人无需安装 Node.js 即可一键启动全部项目。

#### Scenario: 一键启动含前端的完整环境

- **WHEN** 开发者 clone 项目后仅安装 Docker
- **AND** 复制 `.env.example` 为 `.env`
- **AND** 执行 `docker compose up -d`
- **THEN** `http://localhost:3000` 可访问前端页面
- **AND** 前端页面自动连接 Nakama 并完成设备认证
- **AND** 页面显示"服务器连接成功"

#### Scenario: 前端独立于宿主机 Node.js

- **WHEN** 宿主机未安装 Node.js 或 npm
- **AND** 执行 `docker compose up -d`
- **THEN** 前端服务仍然正常启动并可访问
- **AND** 前端功能与其他服务一致（认证、RPC 调用等均可用）

#### Scenario: 仅需体验后端时可选择性启动

- **WHEN** 开发者仅需后端服务（Nakama + PostgreSQL）
- **AND** 执行 `docker compose up -d db nakama`
- **THEN** 前端容器不启动
- **AND** Nakama API 和 Console 仍然可用
- **AND** 开发者可使用 Option A（Vite dev server）在宿主机开发前端

### Requirement: Go 插件自动加载

Nakama 容器 SHALL 挂载宿主机编译产物目录 `server/build/` 到容器内 `/nakama/data/modules/`，Nakama 启动时自动加载其中的 `.so` 插件文件。

#### Scenario: 插件文件已存在于挂载目录

- **WHEN** 宿主机 `server/build/` 目录中存在编译好的 `.so` 文件
- **AND** Nakama 容器启动
- **THEN** Nakama 自动加载该插件
- **AND** 插件注册的 RPC 和 Match Handler 可用

#### Scenario: 无插件文件时启动

- **WHEN** 宿主机 `server/build/` 目录为空
- **THEN** Nakama 正常启动，仅加载内置功能
- **AND** 无错误日志输出
