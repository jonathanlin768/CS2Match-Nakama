## Context

当前仓库的 `docker-compose.yml` 明确服务于本地开发：Nakama 7350/7351 和前端 3000 映射到宿主机，凭据带开发默认值，Go 插件以宿主机 `server/build` bind mount 加载，前端 Docker 构建默认连接浏览器侧 `localhost:7350`。归档设计也明确将 AWS 生产部署留给后续 change。现有两份部署文档提出 Lightsail、Cloudflare Pages/Tunnel 和 GitHub Actions，但没有对应仓库文件，且存在价格与负载模型过时、公网 SSH 与隐藏源站目标冲突、生产密钥缺失、构建跳过逻辑不可用、备份不异地及回滚未实现等问题。

目标用户是最多约 50 人的小圈子，月度目标约 70 元、可接受上限约 100 元。当前战斗由 `SimuMatch` 在单次 RPC 内完成整场模拟并返回战报，不使用实时 Match Handler；认证、社交和聊天仍使用 Nakama HTTP/WebSocket。部署需要优先满足可预测成本、低维护、源站隐藏、可恢复和一次点击更新，而不是高可用集群。

OpenSpec 全局上下文仍写 RDS PostgreSQL，但在当前预算下 RDS 会显著超限。本变更选择同一 Lightsail 内自管 PostgreSQL，并要求实现时同步更正全局部署上下文，记录这是成本约束下的有意识决策。

## Goals / Non-Goals

**Goals:**

- 为 AWS Lightsail 单节点提供可重复的生产 Compose、生产配置和验证脚本，同时完全保留开发 Compose。
- 用 Cloudflare Pages 承载静态前端，用 Cloudflare Tunnel 发布 Nakama 7350，用 Tailscale 提供无公网 SSH/Console 的管理和 CI 通道。
- 通过 GitHub Actions 手动按钮完成测试、构建、版本化发布、备份、更新、健康验证和自动回滚。
- 对 PostgreSQL 提供加密异地备份、隔离恢复演练和长期停服恢复路径。
- 明确单实例重启窗口、真实成本边界和当前 `SimuMatch` 容量模型。

**Non-Goals:**

- 不在仓库自动购买、创建或删除 AWS、Cloudflare、Tailscale、GitHub 资源，不自动操作用户账单。
- 不提供 Kubernetes、RDS、Nakama 多节点、负载均衡、跨区域灾备或零停机发布。
- 不实现 Go 插件进程内热加载；Go 插件更新需要重启单个 Nakama 实例。
- 不新增游戏 RPC、Match Handler、Storage 操作、数据库 schema 或 Luban 表。
- 不承诺 50 个用户同时触发完整模拟时的性能；实现阶段提供可复现的压测方法和基线，实测后再决定是否扩容。

## Decisions

### 1. 单台 Lightsail 承载 Nakama 与 PostgreSQL，前端使用 Cloudflare Pages

生产 Compose 运行 PostgreSQL、Nakama 和 `cloudflared`。React 由 CI 构建为静态产物并发布到 Cloudflare Pages，不在 2 GB 实例常驻前端 Nginx。这样减少常驻内存和静态资源出站，并保留 Cloudflare Pages 的版本回滚。

选择自管 PostgreSQL 而不是 RDS/Lightsail 托管数据库，是因为后者会突破 100 元月度上限。代价是项目必须承担备份、恢复、磁盘监控和数据库升级责任。开发 `docker-compose.yml` 保持不变，生产使用独立文件，避免本地便利性与生产安全互相污染。

PostgreSQL 官方镜像在空数据目录初始化时会先启动仅监听 Unix Socket 的临时服务，再关闭它并启动正式 TCP 服务。Compose 健康检查、镜像集成测试、发布前备份等待和隔离恢复 SHALL 使用 `127.0.0.1:5432` TCP 探针，避免把临时服务误判为可供 Nakama 使用的数据库；集成测试还需以真实 SQL 查询确认测试凭据可用，并在超时后输出数据库容器日志。

替代方案：全部容器部署到 Lightsail 更直观，但失去 Pages CDN/独立回滚并增加服务器构建和静态流量；RDS 运维更轻但超预算；Kubernetes 与当前规模不匹配。

### 2. 后端以不可变 OCI 镜像发布，不 SCP 裸 `.so`

新增后端生产 Dockerfile：构建阶段使用与 Nakama 3.30.0 精确匹配的 pluginbuilder 生成 `backend.so`，运行阶段基于相同 Nakama 版本并复制插件。CI 将镜像推送到 GHCR，以完整 commit SHA 作为不可变 tag；生产 Compose 接收明确的 `BACKEND_IMAGE` 或 `BACKEND_TAG`，不得隐式使用 `latest`。

GitHub 仓库显示名可以包含大写字母，但 OCI/GHCR 镜像路径必须为小写。工作流仅在生成镜像路径时将完整 `owner/repository` 转为小写，不重命名 GitHub 仓库，也不额外追加会造成重复的 `-nakama` 后缀。Docker 构建仅对 429、超时、连接重置、临时 DNS 故障和 HTTP 502/503/504 等瞬时网络错误重试；非法 tag、Dockerfile 语法、Go 编译和文件缺失等确定性错误立即失败。

不可变镜像将 Nakama 版本、Go ABI 和插件绑定在一起，解决旧文档中路径剥离、Runner 缺少旧 artifact 和裸文件半上传问题，也使回滚只需恢复旧 tag。服务器拉取私有 GHCR 镜像时使用只读、最小范围凭据；CI 发布使用仓库临时 `GITHUB_TOKEN`。

替代方案：SCP `backend.so` 更少文件，但难以证明 ABI、完整性和版本对应关系，失败时容易留下半更新状态。

### 3. 公网与管理平面分离

公开 DNS 只包含前端和 API 域名。API 域名经 Cloudflare Tunnel 映射到 Nakama 7350，一条 HTTP ingress 同时承载 REST 和 WebSocket。Nakama Console 不配置公开 hostname。Lightsail 防火墙不向互联网开放 22、7350、7351 或 5432；Tailscale 安装在宿主机，管理员和 GitHub Runner 经 ACL 限定的私网 SSH 连接。

GitHub Actions 使用 Tailscale 的短生命周期 CI 身份，优先采用 workload identity federation，而不是长期 SSH 私钥和公网 IP。Tunnel token、Tailscale 身份和 GHCR 只读凭据相互独立，任何一个泄露都不应赋予完整云账户权限。

替代方案：限制公网 SSH 到个人 IP 无法兼容 GitHub 托管 Runner 的动态地址；Cloudflare Access SSH 可行，但非交互 CI 配置更复杂；直接暴露 7350 再依赖安全组无法隐藏源站。

### 4. 生产配置独立且在启动前验证

新增生产环境模板与预检脚本。Compose 使用 `${VAR:?message}` 或等价预检确保必需变量存在，不允许生产默认值。Nakama 生产配置使用 INFO 日志、Docker 日志轮转、独立的 server key、runtime HTTP key、session encryption key、refresh encryption key 和 Console 凭据。

`socket.server_key` 会嵌入浏览器，因此被视为应用标识而不是秘密；它仍不得使用 `defaultkey`。数据库、session、runtime、Console、Tunnel、备份和管理凭据不进入 Git。生产 `.env` 位于服务器受限目录并设为仅服务账户可读；CI environment secrets 只在部署 job 获得授权后可用。

不引入新的 Go 或 npm 运行依赖。`cloudflared`、Tailscale、容器工具和备份工具属于部署依赖，必须锁定受支持版本或镜像 digest。

### 5. 部署顺序以后端兼容性和可回滚为中心

`workflow_dispatch` 工作流分为两个权限阶段：

1. 无生产凭据的验证阶段：checkout、后端测试、前端测试、插件编译、前端构建。
2. 生产发布阶段：构建并推送 SHA 镜像、加入 Tailscale、创建发布前备份、记录当前 tag、拉取并切换后端、检查本地 `/healthcheck`、验证公开 Tunnel、执行最小认证/RPC/WebSocket smoke、发布 Pages、记录部署结果。

工作流使用 `environment: production` 和固定 `concurrency`。所有第三方 Action 固定到完整 commit SHA。前端在后端健康后发布，降低新前端调用尚未就绪后端的风险。前后端协议发生破坏性变化时仍要求代码本身提供兼容发布顺序；本部署系统不自动解决协议兼容。

失败回滚恢复上一个明确 tag 并再次健康检查。若回滚也失败，工作流停止自动重试并输出人工恢复命令，避免循环破坏。Pages 保留平台部署版本，文档提供回滚到上一次成功 production deployment 的步骤。

Lightsail 本机的 RPC 就绪探针使用 server-only 的 `runtime.http_key` 通过 HTTP Basic Auth 调用规范化的小写 `healthcheck` RPC；不得使用会进入浏览器的 `socket.server_key`，也不把 runtime key 放入 URL。镜像集成测试必须实际请求同一 RPC 并校验成功 JSON，而不只检查“RPC registered”启动日志。首次部署没有上一健康镜像可回滚时，探针失败后使用不带 `-v` 的 Compose `down` 关闭 Tunnel、Nakama 和数据库容器，保留 PostgreSQL named volume，并恢复环境文件中的首次部署占位状态，避免工作流失败但半部署 API 仍公开在线。

部署脚本加载生产环境文件后，Shell 中已导出的变量优先级高于 Docker Compose 的 `--env-file`。因此首次部署用工作流传入的真实镜像 digest 替换模板占位符时，必须同时原子更新环境文件和当前部署进程的 `BACKEND_IMAGE_REF`；回滚时也必须同步两者，避免 Compose 继续解析旧占位符或失败镜像。

部署入口调用同目录的预检和备份脚本时显式使用 Bash，不依赖 Git checkout 是否保留可执行位。这样 Windows/WSL 本地验证与 GitHub 原生 Linux Runner 的行为一致；服务器初始化和工作流仍设置脚本可执行位，支持文档中的人工直接运行命令。

### 6. 将更新定义为可恢复的短中断发布

Go plugin 只在 Nakama 启动时加载，单节点更新使用 `docker compose up -d` 重新创建 Nakama。当前一次性 `SimuMatch` RPC 在切换时可能失败，WebSocket 会断开；前端现有 session 恢复和 socket 重连能力必须通过 smoke 验证。部署输出记录中断开始、恢复和版本信息。

本变更不引入双实例蓝绿发布，因为 2 GB 主机同时运行两个 Nakama 实例会压缩数据库余量，且未来有状态 Match Handler 不能仅靠 HTTP 负载均衡无损迁移。将来引入实时比赛后，应单独提案设计排空、会话保持和 Match 状态策略。

### 7. 使用加密逻辑备份和隔离恢复验证

备份采用 `pg_dump` 一致性逻辑导出，并由支持 S3 兼容对象存储及客户端加密的备份工具写入私有 Cloudflare R2 bucket。实现可选用固定版本 restic 容器，以避免在宿主机散装安装工具；restic repository password 与 R2 凭据分离保存。每日计划备份与发布前备份复用同一脚本，使用锁避免并发，并按快照数量/日期保留。

恢复脚本默认只允许目标为专用临时容器/数据库；覆盖生产库需要额外显式参数和交互确认，不在 CI 自动执行。恢复演练至少校验 Nakama 关键表存在、可读及记录数量合理，然后清理临时资源。

替代方案：Lightsail 本机 cron dump 不能抵御实例/磁盘丢失；整机 snapshot 恢复粒度粗且持续计费；未加密 R2 dump 会暴露账户与业务数据。

### 8. 停服与成本控制分层处理

临时关闭服务通过禁用 Tunnel/维护页和停止 Nakama 处理，不删除数据库 volume；这不停止 Lightsail 费用。长期停服先执行并验证异地备份，再由用户在 AWS 控制台显式删除实例，并检查静态 IP、snapshot、附加磁盘等残留资源。仓库文档只提供清单和可审查命令，不默认提供无确认的销毁 workflow。

AWS Budget 用于多级预警而非硬上限。文档中的价格和配额标记核对日期、币种和官方链接，不把促销或免费层视为长期保证。

### 9. 文档与实际配置共同验收

两份现有部署文档改为分别承担：架构/成本/安全决策，以及从外部账户准备到首次部署、日常更新、回滚、备份恢复和停服的操作手册。文档示例必须引用仓库真实路径和工作流输入，不能内嵌一份与实际文件独立演化的大段 Compose/Workflow 副本；必要片段只用于解释并指向源文件。

容量章节按当前 `SimuMatch` 同步 RPC 重写，并提供容器化压测步骤，记录 20/50 并发下成功率、p50/p95、CPU、RSS 和响应字节。测试结果用于选择 1 GB/2 GB/4 GB，而不是把纸面估计写成保证。

## Risks / Trade-offs

- [单机故障导致全站不可用] → 每日异地备份、可复现 Compose、恢复演练；接受分钟级人工恢复，不承诺 HA。
- [2 GB 同时运行 Nakama 和 PostgreSQL 内存紧张] → Pages 分离前端、生产日志降级与轮转、容器资源观测、50 并发基线；实测不合格则升级而非过度调参。
- [Cloudflare 免费网络在中国大陆的路径波动] → 首次部署前对香港/东京/新加坡按小时实测真实 RPC 与 WebSocket，文档不承诺特定区域必然最低延迟。
- [Cloudflare 或 Tailscale 第三方故障导致入口/管理不可用] → 保留 Lightsail 浏览器控制台作为紧急恢复路径，但日常不开放公网管理端口。
- [生产 secrets 或 CI 供应链泄露] → GitHub environment、OIDC/短生命周期身份、最小权限、Action SHA 固定、日志脱敏和定期轮换。
- [自动回滚遇到数据库不兼容] → 本变更不引入应用 schema；发布前备份。未来数据库迁移必须在独立 change 中设计前后兼容和回滚。
- [单实例更新中断 RPC/WebSocket] → 后端先发布、健康检查、客户端重试验证、记录维护窗口；不误称无中断热更新。
- [R2 免费额度或供应商价格变化] → 文档标记核对日期并监控用量；备份格式保持可迁移到其他 S3 兼容存储。
- [外部控制台步骤无法在仓库 CI 自动验证] → 提供逐项验收命令和截图/结果清单，仓库测试覆盖配置结构和脚本 dry-run。

## Migration Plan

1. 更新 OpenSpec 全局部署上下文，将当前低成本生产数据库决策从 RDS 改为同机 PostgreSQL，并说明未来可迁移到托管数据库。
2. 新增并本地验证后端不可变镜像、生产 Nakama 配置、生产 env 模板和生产 Compose；确认开发 Compose 测试保持通过。
3. 新增配置预检、部署、备份、恢复和 smoke 脚本，全部先在本地临时容器和假外部端点验证。
4. 新增 GitHub Actions，先以无生产 secrets 的构建/测试模式验证，再配置 production environment、GHCR、Cloudflare 和 Tailscale。
5. 用户手动创建 Lightsail、Cloudflare Pages/Tunnel/R2 和 Tailscale 外部资源，按文档记录非敏感 ID 并保存 secrets。
6. 首次部署前完成数据库空库备份/恢复演练；部署后验证前端、认证、RPC、WebSocket、Console 私网访问和公网端口关闭。
7. 用测试账户运行并记录 20/50 并发基线，依据结果保留 2 GB 或升级。
8. 旧文档内容在新配置验收后替换；旧开发 Compose 不迁移、不删除。

回滚仓库变更时，恢复旧文档并停止使用生产 workflow/Compose，不删除现有 PostgreSQL volume。回滚线上版本时恢复上一 SHA 镜像并回滚 Pages；若需要重建实例，从最后已验证的异地备份恢复。任何生产数据销毁均保持人工确认。

## Open Questions

- 最终 Lightsail 区域应由目标玩家在香港、东京和新加坡的实测结果决定，不能仅按地理距离选择。
- GitHub 仓库是否为 private 会影响 GHCR 拉取凭据和 GitHub environment 保护功能的可用范围，实施时需按实际仓库权限核对。
- 用户域名是否已托管到 Cloudflare；若不希望迁移整个 DNS zone，需要在实施前确认可接受的子域/独立域方案。
- R2 是否作为最终备份目标；若用户不希望为 R2 配置支付资料，应选择另一 S3 兼容私有存储并更新成本说明。
