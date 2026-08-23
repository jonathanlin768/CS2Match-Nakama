## 1. 服务端资料版本与 Storage 模型

- [x] 1.1 扩展 `ContactProfile`、`ExchangeRequest` 和响应 DTO，加入完整 profile revision、请求时 revision、接受时双方 revision 及 stale/revoked 状态，并兼容读取旧 JSON。
- [x] 1.2 修改 profile repository/service：标准化后无变化不递增 revision，任一字段实际修改/新增/删除则整体 revision 递增，并允许清空一个或全部渠道。
- [x] 1.3 实现仅绑定 Session 本人的 `SocialGetContactProfile`，返回完整本人资料、revision 和脱敏摘要，不接受目标 user ID。
- [x] 1.4 为 repository 增加按当前好友分页、批量读取确定性 pair 记录及 CAS 撤销交换所需的读写能力，保持交换与资料 Storage `PermissionRead=0`、`PermissionWrite=0`。

## 2. 服务端授权、收件箱与好友删除

- [x] 2.1 修改 `SocialRequestContactExchange`，记录请求者 profile revision，并对同 revision 重试保持幂等。
- [x] 2.2 修改 `SocialRespondContactExchange`，在接受前复核好友关系、双方渠道和请求者 revision，以 Storage version 原子记录双方接受时 revision。
- [x] 2.3 修改 `SocialGetContactExchange`，只有 accepted、仍为好友且双方 revision 均匹配时返回授权渠道正文；其他状态、旧版 accepted 和版本变化均不返回任何正文。
- [x] 2.4 实现并注册 `SocialListContactExchangeInbox`，分页返回 received、sent、需要重新授权摘要和 `incoming_pending_count`，确保响应不包含联系方式正文。
- [x] 2.5 注册 `RegisterBeforeDeleteFriends` hook，解析 ID/username 目标并在放行删除前把 pair 记录持久转为 revoked；撤销失败时拒绝删除。
- [x] 2.6 保留 DM 结构化卡片作为尽力而为提示，记录不含 PII 的投递失败诊断，同时保证 Storage 成功不因卡片失败而回滚。

## 3. 服务端测试与插件验证

- [x] 3.1 增加 profile 单元测试，覆盖旧资料初始 revision、相同值保存、任一字段修改、单渠道/全部清空、格式验证和本人读取身份绑定。
- [x] 3.2 增加交换状态测试，覆盖请求期间资料修改、双向接受、任一方修改导致整体 stale、旧 accepted 安全降级、并发响应和所有非 accepted 状态不泄露正文。
- [x] 3.3 增加好友删除 hook 测试，覆盖 ID、username、撤销失败阻止删除，以及删除后重新添加不能恢复旧授权。
- [x] 3.4 增加收件箱测试，覆盖 received/sent 分类、pending 计数、stale/revoked 摘要、非好友过滤、分页及响应序列化不含 QQ/微信正文。
- [x] 3.5 执行 Go 格式化、`go test ./...`、权限检查，并按项目构建流程把 Nakama Go 插件成功编译为 `.so`。

## 4. 前端 Social API 与权威状态层

- [x] 4.1 扩展 `client/src/api/social.ts` 类型与调用，接入本人资料、inbox、stale/revoked 和 profile/exchange revision 字段。
- [x] 4.2 实现 `SocialApiError` 转换，解析 Nakama/HTTP Response、结构化 code 和网络异常，提供不包含敏感 payload 的中文错误映射。
- [x] 4.3 将联系方式 inbox 状态提升到 App Shell 可复用层，在登录、显式刷新、Socket 重连、有效卡片和响应操作后拉取权威 inbox。
- [x] 4.4 调整现有卡片 Hook：卡片只触发权威刷新，频道加入/历史失败显式暴露可重试状态，不再静默决定待办或授权。
- [x] 4.5 在好友页加载 `SocialGetContactProfile` 并安全回填表单；保存成功后使用服务端规范化结果更新输入与摘要，加载失败不覆盖本地编辑。

## 5. 好友与联系方式交互

- [x] 5.1 在好友页普通好友列表之前实现响应式“联系方式申请”分组，展示收到/发出的申请、渠道、时间、过期状态及接受/拒绝操作。
- [x] 5.2 在 App Shell 好友入口显示权威 `incoming_pending_count` 徽章，并确保从其他页面可以进入对应待办。
- [x] 5.3 更新交换详情状态 UI，支持 pending、accepted、declined、expired、stale、revoked；失效时先清除已展示正文并提供重新申请入口。
- [x] 5.4 在实际修改本人联系方式前显示整次既有授权失效提示，并支持清空单个或全部联系方式。
- [x] 5.5 将 `PROFILE_INCOMPLETE`、`NOT_FRIENDS`、`CONFLICT`、`INVALID_STATE`、`RATE_LIMITED` 等错误映射为可操作提示，消除 `[object Response]`。

## 6. 前端测试与构建验证

- [x] 6.1 增加 API/错误转换单元测试，覆盖字符串 payload、对象 payload、Fetch/SDK Response、未知错误及敏感数据不进入文案。
- [x] 6.2 增加权威 inbox 与卡片协调测试，覆盖离线恢复、卡片投递/历史失败、重连刷新、非当前好友事件和待处理计数。
- [x] 6.3 增加好友页组件测试，覆盖本人资料刷新回填、申请分组、接受/拒绝、stale/revoked 清空正文、修改确认和响应式入口。
- [x] 6.4 执行前端单元测试、类型检查、lint（如项目已配置）和生产构建，修复所有与本变更相关的失败。

## 7. 本地集成、安全与验收

- [x] 7.1 重新构建并加载本地 Nakama 后端和 React 前端，确认新增 RPC 与好友删除 hook 已注册且功能开关开启。
- [x] 7.2 使用两个隔离 Chrome Profile/浏览器上下文执行 A 发起、B 实时看到、B 接受、双方刷新后查看的端到端流程。
- [x] 7.3 验证 B 离线或 DM 卡片投递不可用时仍能从权威收件箱恢复申请，且 App Shell 待处理计数正确更新。
- [x] 7.4 验证任一方修改或清空 QQ/微信后整次交换变 stale、双方均不再获得任何正文，重新申请并接受后才恢复。
- [x] 7.5 验证删除好友立即撤销、重新添加不复活旧授权，并验证客户端直接读写私有 Storage 和伪造 request ID 均被拒绝。
- [x] 7.6 检查服务端日志、DM 卡片、inbox 响应、Toast 和浏览器持久存储均不包含不应披露的联系方式正文。
- [x] 7.7 更新 `doc/bug_list.md` 对应问题、必要的运行配置说明和手工验收记录。

## 8. 单渠道填写澄清

- [x] 8.1 明确 QQ 与微信均为选填，只填写任一合法渠道即可保存并仅申请该渠道。
- [x] 8.2 修正容易暗示两项必填的界面/错误文案，增加单 QQ、单微信回归测试并重新验证构建。

## 9. 不对称单渠道交换修正

- [x] 9.1 补充双方渠道可不对称的授权语义：各自至少一种，接受时授权各自全部非空联系方式，任一资料变更仍使整次授权失效。
- [x] 9.2 修改请求与接受服务逻辑，由服务端派生发起方渠道，并在接受时验证双方各自至少一种联系方式、记录双方渠道并集。
- [x] 9.3 增加 QQ+微信/单 QQ、单 QQ/单微信等回归测试，并修正 pending 与 `PROFILE_INCOMPLETE` 文案避免暗示双方必须同渠道。
- [x] 9.4 完成后端、前端及 OpenSpec 严格验证，重新构建并加载本地 Docker 前后端，执行双账号真实端到端验收。
