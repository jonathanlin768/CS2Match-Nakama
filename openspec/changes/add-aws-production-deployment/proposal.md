## Why

仓库目前只有面向本地开发的 Docker Compose 规约与配置，已有 AWS/CI-CD 文档包含过时价格、与当前一次性 `SimuMatch` RPC 不符的负载模型，以及公网 SSH、密钥、备份和回滚方面的缺口，无法安全、可重复地将项目部署给小规模真实用户。现在需要建立一套预算可控、隐藏源站、可手动一键发布并可恢复的单节点生产部署能力，同时保留现有本地开发体验。

## What Changes

- 新增 AWS Lightsail 单节点生产部署规约，以 Docker Compose 运行 Nakama 3.30、Go 插件与 PostgreSQL 15；前端构建产物发布到 Cloudflare Pages，API/WebSocket 通过 Cloudflare Tunnel 暴露。
- 新增独立的生产 Compose、生产环境变量模板和 Nakama 生产配置；开发用 `docker-compose.yml`、本地端口与默认开发流程保持不变。
- 生产前端构建 SHALL 显式注入 HTTPS API 域名、443 端口与 SSL 开关，不再依赖开发默认值 `localhost:7350`。
- 生产环境 SHALL 不向公网开放 PostgreSQL、Nakama API、Nakama Console 或 SSH；Cloudflare Tunnel 作为公开业务入口，Tailscale 私网作为人工管理和 GitHub Actions 部署通道。
- 新增 GitHub Actions 手动发布流程，执行测试、Go 插件编译、前端构建、版本化产物发布、数据库发布前备份、后端更新、健康检查、前端发布和失败回滚，并禁止同一生产环境并发部署。
- 新增 PostgreSQL 每日异地备份、保留、恢复验证和长期停服说明；明确 Lightsail 停止实例不停止计费，长期停服需在备份后删除实例或其他计费资源。
- 修订 `doc/aws-deployment.md` 与 `doc/ci-cd-github-actions.md`，使价格、当前 RPC 负载模型、Cloudflare/Tailscale 拓扑、凭据边界、部署与回滚步骤和仓库实际文件一致。
- 生产部署只承诺单实例短时重启更新；不声称 Go 插件支持无中断热加载，并要求客户端能够处理 Nakama 短暂断连。
- 不自动购买、创建、删除或收费使用 AWS、Cloudflare、Tailscale、GitHub 外部资源；仓库提供配置、脚本、工作流和可验证的人工控制台步骤。

## Capabilities

### New Capabilities

- `aws-production-deployment`: 覆盖 Lightsail 单节点生产拓扑、生产 Compose、源站隐藏、生产凭据、手动一键部署、健康检查与回滚、数据库异地备份/恢复、停服与成本保护以及部署文档一致性。

### Modified Capabilities

- `react-frontend-scaffold`: 将仅面向 `localhost:7350` 的 Docker 构建要求扩展为区分开发默认连接与生产构建时显式注入的 HTTPS Nakama API 配置。

## Impact

- **Nakama 后端**：增加生产构建/装载方式和生产配置覆盖；不新增 RPC、Match Handler 或 Nakama Storage 操作，不改变 `SimuMatch` 一次 RPC 完成整场模拟的行为，也不改变未来 MatchLoop 状态同步语义。
- **React 前端**：生产构建显式注入 API host、port、server key 与 SSL 配置；需要验证认证、RPC 与 WebSocket 在自定义域名下工作。
- **数据库**：仍使用 PostgreSQL 15 和 Docker named volume，不新增 schema migration；新增一致性备份、异地保留与恢复流程。
- **部署**：新增生产 Compose、生产环境模板、Nakama 生产配置、Cloudflare Tunnel/Tailscale 接入说明、GitHub Actions、部署/备份/恢复脚本及两份部署文档修订。
- **依赖与外部系统**：运行环境增加 `cloudflared` 与 Tailscale；使用 AWS Lightsail、Cloudflare Pages/Tunnel/R2、GitHub Actions/GHCR 和 Tailscale Personal。所有第三方 Actions 必须最小权限并固定到不可变版本。
- **Luban 配置**：不需要修改 Luban 表或导表协议；发布构建继续消费仓库中已生成并提交的配置产物。
