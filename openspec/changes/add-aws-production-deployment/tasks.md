## 1. 部署基线与全局约束

- [x] 1.1 更新 `openspec/config.yaml` 的部署上下文：记录预算内生产方案使用单台 AWS Lightsail、Docker Compose 和同机 PostgreSQL，RDS 仅作为未来扩容选项，并保持单区域、无 Kubernetes 决策。
- [x] 1.2 盘点并记录生产部署将新增/修改的仓库文件、外部控制台资源和 secrets，确认不会覆盖用户现有开发 `docker-compose.yml`、`.env.example` 或未提交改动。
- [x] 1.3 定义支持的生产主机前置条件、目录布局、服务账户、文件权限、域名变量和版本锁定策略，并提供无真实凭据的示例值。

## 2. 不可变 Nakama 后端镜像

- [x] 2.1 新增后端生产 Dockerfile，使用 Nakama 3.30.0 对应 pluginbuilder 编译 `server/build/backend.so`，再复制到相同版本 Nakama 运行镜像；确保构建包含已提交的生成配置且不携带源码 secrets。
- [x] 2.2 增加镜像元数据和 commit SHA 版本注入，生产部署禁止隐式使用 `latest`，并记录 Nakama/Go plugin ABI 对齐信息。
- [x] 2.3 新增 Nakama 生产配置，设置 INFO 日志、日志轮转所需约束、独立 runtime/session/refresh/Console 配置，并避免开发默认密钥。
- [x] 2.4 添加后端镜像构建检查，验证 `.so` 存在、镜像内路径正确、Nakama `check`/启动可加载插件且 `HealthCheck`、`SimuMatch` 和社交 RPC 注册成功。
- [x] 2.5 运行 Go 单元测试及官方 pluginbuilder 编译验证，确认本变更不新增或修改 RPC、Match Handler、Storage 权限和 Luban 生成协议。

## 3. 生产前端构建

- [x] 3.1 调整前端构建配置和示例，使开发构建继续默认 `localhost:7350`，生产构建必须显式提供 API host、443、server key 和 SSL，并在缺失/使用禁止默认值时失败。
- [x] 3.2 增加生产前端配置测试，验证 Nakama Client 使用 HTTPS 自定义域名且构建产物不包含数据库、session/runtime、Cloudflare、Tailscale 或对象存储 secrets。
- [x] 3.3 验证 Cloudflare Pages 静态构建、SPA 路由、认证、RPC 与 WebSocket 所需资源路径，并保留本地 Nginx Docker 构建和页面回退测试。
- [x] 3.4 运行前端 TypeScript 检查、单元测试和生产构建，记录构建输出目录、文件数量及单文件大小是否满足 Pages 限制。

## 4. 生产 Compose、入口与管理平面

- [x] 4.1 新增独立生产 Compose，编排 PostgreSQL named volume、SHA 标记的 Nakama 后端镜像和固定版本/digest 的 `cloudflared`，配置健康检查、依赖、重启策略、资源/日志限制且不发布 5432、7350、7351 到公网地址。
- [x] 4.2 新增生产环境变量模板和 fail-closed 预检脚本，拒绝缺失变量、开发默认值、可疑占位符及权限过宽的生产 secrets 文件。
- [x] 4.3 提供 Cloudflare Tunnel 配置模板，使用一条 API hostname HTTP ingress 同时代理 REST/WebSocket 并以 404 catch-all 结束；不得为 Console 创建公开路由。
- [x] 4.4 提供 Tailscale 主机初始化和 ACL/CI 身份配置说明，使管理员、Nakama Console 和 GitHub Runner 仅通过私网访问，并保留 Lightsail 浏览器控制台作为紧急通道。
- [x] 4.5 添加生产 Compose 静态验证和本地集成测试，确认数据库仅内部可达、Nakama/Cloudflared 启动顺序正确、主机重启后自动恢复且开发 Compose 行为不变。
- [x] 4.6 将生产/开发 Compose、镜像集成测试、发布前备份和隔离恢复的 PostgreSQL 就绪检查统一为正式 TCP 探针；集成测试增加真实 SQL 认证检查、超时日志和防止临时 Unix Socket 误判的回归覆盖。

## 5. 备份、恢复与停服保护

- [x] 5.1 实现可锁定并发的 PostgreSQL 逻辑备份脚本，支持每日和发布前标签，使用固定版本的客户端加密备份工具上传私有 S3 兼容对象存储，输出脱敏状态并按策略保留。
- [x] 5.2 实现默认恢复到独立临时 PostgreSQL 容器的验证脚本，校验备份完整性、Nakama 关键表和基础记录后清理临时资源；覆盖生产目标必须额外显式确认。
- [x] 5.3 为备份失败、对象存储不可达、错误解密密钥、并发备份和损坏快照编写脚本级测试或容器化演练，确认失败不会删除最后一个有效备份或修改生产库。
- [x] 5.4 实现或记录临时关闭公开入口/应用的安全操作，确保保留数据库 volume，并明确该操作不停止 Lightsail 实例费用。
- [x] 5.5 编写长期停服清单：验证最新异地备份后由用户显式删除实例，并检查静态 IP、snapshot、附加磁盘等残留计费资源；仓库脚本不得无确认销毁资源。

## 6. GitHub Actions 一键发布与回滚

- [x] 6.1 新增手动生产部署 workflow，设置 production environment、允许的分支/提交输入、固定 concurrency、最小 `contents`/`packages`/`id-token` 权限和部署 URL。
- [x] 6.2 实现无生产凭据的验证 job：运行 Go 测试、前端测试、pluginbuilder 编译、后端镜像构建和生产前端构建；任何失败均阻止进入部署阶段。
- [x] 6.3 使用 GitHub 临时 token 将 SHA 标记的后端镜像推送 GHCR，并生成可追踪的镜像 digest/构建摘要；第三方 Actions 固定到完整 commit SHA。
- [x] 6.4 使用 Tailscale workload identity 将 GitHub Runner 作为短生命周期 CI 节点接入私网，按 ACL 只允许访问生产 SSH，并验证工作流不依赖公网 `LIGHTSAIL_IP` 或长期 SSH 私钥。
- [x] 6.5 实现发布前备份、记录当前镜像、拉取指定 digest、原子更新 Compose、容器/本地 `/healthcheck`、公开 Tunnel 和最小认证/RPC/WebSocket smoke 的后端部署步骤。
- [x] 6.6 实现后端健康检查失败时恢复上一已知健康镜像并再次验证；回滚失败时停止自动循环、保留日志并输出脱敏的人工恢复命令。
- [x] 6.7 后端验证成功后用 Cloudflare 官方发布方式部署同一提交的 Pages 产物，记录 Pages deployment；补充前端失败和手工回滚上一成功版本的处理。
- [x] 6.8 对 workflow 做 actionlint/语法检查和 dry-run 可测试部分，覆盖并发触发、构建失败、无凭据阶段、后端失败回滚、前端失败及日志密钥脱敏。
- [x] 6.9 修正混合大小写 GitHub 仓库对应的 GHCR 镜像路径，移除重复后缀，并将 Docker 构建重试限定为瞬时网络错误；增加镜像命名、429、非法 tag 和 Go 编译失败的回归测试并同步文档示例。
- [x] 6.10 修复首次部署写入真实镜像 digest 后当前 Shell 仍覆盖 Compose env 文件的问题；镜像切换与回滚同步更新文件和进程环境，并增加首次部署占位符回归测试。
- [x] 6.11 修复 GitHub Linux Runner 因部署子脚本缺少可执行位导致的验证失败；入口显式使用 Bash 调用预检/备份脚本并增加跨平台文件模式回归断言。
- [x] 6.12 隔离首次部署、备份失败和恢复失败脚本测试的 marker/调用记录，并在本地 WSL2/Linux 容器中按 GitHub `validate` 顺序完成提交前验证。

## 7. 安全、运行时与性能验证

- [ ] 7.1 验证生产 Compose 和 Lightsail 防火墙验收清单中 22、5432、7350、7351 均不接受公网直连，API 仅通过 Tunnel 可用，Console/SSH 仅通过 Tailscale 可用。
- [x] 7.2 验证生产 server key 不为默认值但按客户端可见信息处理，数据库、session、refresh、runtime、Console、Tunnel、Tailscale、GHCR 读取和备份凭据均未进入 Git、镜像层、前端 bundle 或 CI 日志。
- [ ] 7.3 运行端到端生产拓扑 smoke：前端加载、设备/邮箱认证、Session 恢复、`HealthCheck`、`SimuMatch`、好友/聊天 RPC、WebSocket 连接及 Nakama 短暂重启后的客户端恢复。
- [x] 7.4 提供容器化负载基线工具或脚本，对当前一次性 `SimuMatch` RPC 记录 20/50 并发成功率、p50/p95、CPU、内存和响应大小，避免沿用 10Hz MatchLoop 流量估算。
- [ ] 7.5 验证 PostgreSQL named volume 在容器重启/重建后保留，异地备份在隔离环境成功恢复，并记录最近一次恢复演练日期和结果。
- [ ] 7.6 验证日志轮转、磁盘余量、容器健康、备份失败和 AWS Budget 多级告警的人工配置清单；明确 Budget 为告警而非硬消费上限。

## 8. 部署文档重写与交付

- [x] 8.1 重写 `doc/aws-deployment.md`，以当前官方价格/配额核对日期、预算、安全拓扑、当前 RPC 负载模型、区域实测方法、临时/长期停服及恢复策略替换过时内容。
- [x] 8.2 重写 `doc/ci-cd-github-actions.md`，引用仓库真实 Compose、Dockerfile、脚本和 workflow，说明外部账户首次配置、GitHub environment/secrets、首次部署、日常发布、短中断、健康检查、回滚和故障排查。
- [x] 8.3 在文档中明确区分仓库自动化与 AWS/Cloudflare/Tailscale/GitHub 控制台人工步骤，删除无效 Console URL、重复 Tunnel ingress、`runtime.http_key`/CORS 混淆、不可用 `skip_build` 和不存在 `.backup` 等旧说明。
- [ ] 8.4 从空白目录按文档执行所有无需真实云账号的命令和静态检查，并由真实外部环境验收人逐项记录 AWS、Cloudflare、Tailscale、Pages、R2 与 GitHub 配置结果。
- [x] 8.5 汇总 Go、前端、Docker、备份恢复、workflow、安全端口、端到端和负载验证证据，确认所有 OpenSpec 场景有对应测试或明确人工验收项后准备实施交付。
