## ADDED Requirements

### Requirement: 单节点生产拓扑与开发环境隔离

系统 SHALL 提供独立于开发用 `docker-compose.yml` 的生产部署定义，在单台 AWS Lightsail Linux 实例上以 Docker Compose 运行 PostgreSQL 15、Nakama 3.30 和版本化 Go 插件，并通过 Cloudflare Tunnel 发布 Nakama HTTP/WebSocket 入口。生产部署 SHALL 不改变本地开发 Compose 的端口、默认凭据和一键启动行为。

#### Scenario: 生产 Compose 首次启动

- **GIVEN** 运维者已在支持 Docker Compose 的 Lightsail 实例准备生产环境文件和所需密钥
- **WHEN** 运维者执行文档规定的生产启动命令
- **THEN** PostgreSQL 健康后 Nakama 完成迁移并加载与 Nakama 3.30.0 ABI 匹配的 Go 插件
- **AND** PostgreSQL 健康条件验证正式 TCP 服务，不会被首次初始化期间仅监听 Unix Socket 的临时服务提前满足
- **AND** Cloudflare Tunnel 将指定 API 域名转发到 Nakama 7350
- **AND** 开发用 `docker-compose.yml` 未被生产配置覆盖

#### Scenario: 生产容器自动恢复

- **GIVEN** 生产 Compose 已成功运行
- **WHEN** Lightsail 主机重启或某个生产容器异常退出
- **THEN** PostgreSQL、Nakama 和 Tunnel 按依赖顺序自动恢复
- **AND** PostgreSQL named volume 中的数据保持不变

### Requirement: 生产源站不接受公网直连

生产环境 SHALL 以 Cloudflare Tunnel 作为公开的 Nakama HTTP/WebSocket 入口，以 Tailscale 私网作为 SSH、Nakama Console 和自动部署管理通道。PostgreSQL 5432、Nakama 7350/7351 和 SSH 22 SHALL NOT 对互联网开放，生产 Compose SHALL NOT 将 PostgreSQL 或 Nakama Console 发布到公共宿主机地址。

#### Scenario: 外部用户访问业务 API

- **GIVEN** Cloudflare Tunnel 正常运行且 API 自定义域名已生效
- **WHEN** 用户通过 HTTPS 调用 Nakama REST/RPC 或建立 Nakama WebSocket
- **THEN** 请求经 Cloudflare 转发到 Nakama 7350
- **AND** 用户无需知道或连接 Lightsail 源站 IP

#### Scenario: 管理员访问 Console

- **GIVEN** 管理员设备已通过授权身份加入 Tailscale 网络
- **WHEN** 管理员访问 Nakama Console 或 SSH
- **THEN** 连接仅通过 Tailscale 私网到达实例
- **AND** 未授权公网客户端无法连接 22 或 7351

### Requirement: 生产配置和凭据必须 fail closed

生产部署 SHALL 提供不含真实值的环境变量模板，并要求数据库密码、Console 密码、session encryption key、refresh encryption key、runtime HTTP key、Tunnel token、备份凭据和部署凭据从仓库外注入。生产启动 SHALL 在必需变量缺失或仍为已知开发默认值时失败，不得通过 Compose 默认表达式回退到 `nakama`、`password`、`defaultkey` 或 Nakama 默认加密密钥。客户端 server key SHALL 与服务端一致，但文档 MUST 明确它会嵌入浏览器产物，不能作为保密边界。

#### Scenario: 缺少生产密钥

- **GIVEN** 一个必需生产密钥未设置或仍为禁止的开发默认值
- **WHEN** 运维者验证或启动生产 Compose
- **THEN** 配置检查以非零状态失败并指出缺失的变量名
- **AND** PostgreSQL、Nakama 和 Tunnel 不以弱默认值启动

#### Scenario: 仓库凭据扫描

- **WHEN** CI 检查提交内容和生产环境模板
- **THEN** 仓库中不存在真实数据库密码、session/runtime 密钥、Cloudflare token、Tailscale 凭据或对象存储凭据
- **AND** 模板只包含说明性占位符和非敏感配置

### Requirement: 手动一键生产部署

系统 SHALL 提供由 GitHub Actions `workflow_dispatch` 触发的生产部署工作流。工作流 SHALL 在获得生产环境凭据前完成后端测试、前端测试、Go 插件 ABI 对齐编译和前端构建；随后发布以 Git commit SHA 标识的不可变后端镜像，通过短生命周期 Tailscale CI 节点连接源站，执行部署前备份、后端更新、健康检查和 Cloudflare Pages 前端发布。工作流 SHALL 使用生产 environment、最小权限和单一并发组，且同一时刻最多一个生产部署运行。

#### Scenario: 成功部署指定提交

- **GIVEN** `master` 分支上的指定提交通过全部构建和测试且外部服务凭据有效
- **WHEN** 授权用户在 GitHub Actions 点击生产部署按钮
- **THEN** GHCR 中存在以该 commit SHA 标识的后端镜像
- **AND** 镜像路径由 GitHub 的 `owner/repository` 转为全小写，仓库显示名无需重命名且不追加重复组件名
- **AND** Lightsail 运行该不可变镜像并通过本地及公开健康检查
- **AND** Cloudflare Pages 发布由同一提交构建的前端
- **AND** 工作流输出提交 SHA、部署结果和公开 URL，但不输出任何密钥

#### Scenario: 首次部署替换镜像占位符

- **GIVEN** 服务器生产环境文件中的 `BACKEND_IMAGE_REF` 仍是首次部署占位符
- **AND** 工作流已发布后端镜像并向部署脚本传入有效的不可变 digest
- **WHEN** 部署脚本验证配置并启动生产 Compose
- **THEN** 环境文件和当前部署进程使用同一个真实 digest
- **AND** Docker Compose 不会因 Shell 环境变量优先级继续解析旧占位符

#### Scenario: 发布构建发生失败

- **GIVEN** 工作流正在构建待发布的后端镜像
- **WHEN** Docker 返回 429、超时、连接重置、临时 DNS 故障或 HTTP 502/503/504
- **THEN** 构建按有界退避策略重试
- **AND** 非法镜像名、Dockerfile 语法错误、Go 编译错误或文件缺失立即失败且不重试

#### Scenario: 并发触发生产部署

- **GIVEN** 一个生产部署仍在运行
- **WHEN** 授权用户再次触发生产部署
- **THEN** GitHub Actions 按生产并发策略等待或取消旧的待处理运行
- **AND** 两个部署不会同时修改生产 Compose 或备份状态

#### Scenario: 构建或测试失败

- **WHEN** 后端测试、前端测试、插件编译或前端构建任一步失败
- **THEN** 工作流停止且不读取部署阶段专用凭据
- **AND** 生产服务器、数据库和 Cloudflare Pages 当前版本保持不变

### Requirement: 部署健康检查和自动回滚

生产发布 SHALL 在切换前记录当前后端镜像版本，并在更新后验证容器健康、Nakama `/healthcheck`、关键认证/RPC 可达性和公开 Tunnel 路径。任何后端发布后检查失败时 SHALL 自动恢复上一已知健康镜像并再次验证；前端发布 SHALL 在后端健康后进行，并记录 Cloudflare Pages 可回滚版本。单实例 Go 插件更新 MUST 被描述为有短暂断连的重启更新，而非无中断热加载。

#### Scenario: 新后端镜像启动失败

- **GIVEN** 当前生产版本健康且新镜像无法加载 Go 插件或无法通过健康检查
- **WHEN** 部署工作流尝试切换到新镜像
- **THEN** 工作流恢复上一已知健康镜像
- **AND** 回滚后的 Nakama 再次通过健康检查
- **AND** 前端新版本不被发布

#### Scenario: 单实例更新发生短暂断连

- **WHEN** Nakama 容器因 Go 插件版本更新被重新创建
- **THEN** 部署记录明确标记短暂服务中断窗口
- **AND** 验证客户端能够在服务恢复后重新认证或重连
- **AND** 规范不承诺保留进程内 Match Handler 状态

### Requirement: 数据库异地备份与可验证恢复

生产环境 SHALL 在每日计划任务和每次后端发布前生成 PostgreSQL 一致性逻辑备份，并以客户端加密方式上传到私有异地对象存储。备份 SHALL 具有明确保留策略、成功/失败日志和完整性检查；仓库 SHALL 提供不会覆盖当前生产库的恢复验证流程。仅保存在 Lightsail 本机的文件不得被视为有效备份。

#### Scenario: 每日备份成功

- **GIVEN** PostgreSQL 健康且对象存储凭据有效
- **WHEN** 每日备份任务运行
- **THEN** 生成可校验的加密备份并上传到私有对象存储
- **AND** 按保留策略清理过期备份
- **AND** 日志不包含数据库密码、备份解密密钥或用户明文数据

#### Scenario: 隔离恢复演练

- **GIVEN** 对象存储中存在一个已验证备份
- **WHEN** 运维者执行恢复验证流程
- **THEN** 数据被恢复到独立的临时 PostgreSQL 容器或数据库
- **AND** 恢复只在临时 PostgreSQL 的正式 TCP 服务就绪后开始
- **AND** 流程校验 Nakama 关键表可读后销毁临时恢复目标
- **AND** 当前生产数据库未被覆盖

### Requirement: 停服、销毁与成本保护必须区分

部署文档和控制脚本 SHALL 区分关闭公开入口、停止应用容器、停止 Lightsail 实例和备份后删除计费资源的效果。文档 MUST 明确 Lightsail 停止实例仍继续产生实例费用、AWS Budget 默认不是硬性消费上限，并提供恢复前置条件和外部资源清理清单。仓库自动化 SHALL NOT 在没有显式人工确认和最新异地备份的情况下删除云资源或数据库 volume。

#### Scenario: 临时关闭公开服务

- **WHEN** 运维者执行经过确认的临时停服操作
- **THEN** Cloudflare 公开 API 不再到达 Nakama且前端显示或指向维护状态
- **AND** PostgreSQL volume 保留
- **AND** 文档说明 Lightsail 实例费用不会因此停止

#### Scenario: 长期停服前置检查

- **GIVEN** 运维者计划删除 Lightsail 实例以停止实例费用
- **WHEN** 执行长期停服清单
- **THEN** 清单要求先完成并验证最新异地数据库备份
- **AND** 清单列出实例、未挂载静态 IP、快照、磁盘和其他可能继续计费的资源
- **AND** 删除云资源仍需在 AWS 控制台或独立的显式授权流程中确认

### Requirement: 生产部署文档必须与仓库和当前架构一致

`doc/aws-deployment.md` 与 `doc/ci-cd-github-actions.md` SHALL 描述仓库实际提供的生产文件、命令、域名、密钥边界、备份/恢复和回滚流程。成本和配额 SHALL 标注核对日期并链接官方来源；容量模型 SHALL 以当前一次 RPC 生成完整战报的 `SimuMatch` 行为为基础，不得继续把当前实现描述为 10Hz MatchLoop 持续广播。

#### Scenario: 按文档从空白主机部署

- **GIVEN** 运维者拥有满足前置条件的空白 Lightsail Linux 实例和已配置的外部账户
- **WHEN** 运维者严格按部署文档执行仓库侧步骤
- **THEN** 所有引用的文件、变量和命令均在仓库中存在或被明确标记为控制台人工步骤
- **AND** 最终前端、认证、RPC、WebSocket、备份和回滚检查均有可执行验证方法

#### Scenario: 文档审查当前比赛负载模型

- **WHEN** 维护者对照 OpenSpec 和 `SimuMatch` 实现审查容量章节
- **THEN** 文档说明当前比赛在单次 RPC 内完成模拟并返回完整战报
- **AND** 未来引入实时 Match Handler 后需要重新测量 CPU、内存、WebSocket 和出站流量
