## Context

现有 Social 模块将联系方式正文保存在客户端不可直接读写的 `social_contact_profile/profile` 中，将每对好友的交换状态保存在系统 owner 下的确定性 pair key 中。接受后，`SocialGetContactExchange` 每次读取双方当前资料，因此资料变更会把新值自动暴露给旧授权。好友删除只在读取时临时返回 expired，没有持久撤销；如果好友关系删除后重建，旧 accepted 记录可能复活。

客户端目前用 `useContactInbox` 为每名好友加入 persistent DM，依靠结构化卡片历史生成好友行摘要和本地未读数。实测中实时卡片能够到达，但只有一段很小的好友行文字和数字徽章，没有独立待办或全局提示；请求失败会显示 `[object Response]`，刷新后本人资料输入框为空。卡片投递和频道历史错误又被静默忽略，因此不能作为权威工作流。

本变更同时影响 Nakama Go 插件、React 客户端和 Nakama 私有 Storage JSON。系统规模上限约 100 个在线用户，可优先选择简单、可审计的服务端批量读取，而不是引入额外索引服务。

## Goals / Non-Goals

**Goals:**

- 实现双方对接受时完整联系方式版本的双向授权。
- 任意一方对 QQ 或微信作出有效修改或删除时，使整次授权失效且不返回任何联系方式正文。
- 删除好友时持久撤销请求和授权，重新添加好友不恢复旧状态。
- 提供服务端权威交换收件箱，在离线、刷新、换设备和 DM 卡片失败时恢复待办。
- 允许正式账号通过专用 RPC 读取自己的完整联系方式并回填表单。
- 提供明确的待办入口、状态、计数和结构化错误反馈。
- 保持私有 Storage 权限、日志脱敏和服务端授权判断。

**Non-Goals:**

- 不提供自由文本、图片、文件或链接聊天。
- 不发送邮件、短信或浏览器系统通知。
- 不解决同一 Chrome Profile 内多个标签页共享登录凭据的问题；双账号 E2E 使用隔离浏览器上下文。
- 不引入新的 PostgreSQL schema、外部队列、Go/npm 依赖或 Luban 配置。
- 不改动 Match Handler、MatchLoop、战斗状态同步或玩法。

## Decisions

### 1. 用单一资料修订号绑定整次授权

`ContactProfile` 增加单调递增的 `revision`。`SocialSetContactProfile` 先做 trim 和格式标准化，再与已保存 QQ、微信比较：标准化后无变化时保持 revision；任一字段实际变化、增加或删除时 revision 加一。允许用户清空一个或全部渠道，清空也视为变化。

请求创建时记录 `requester_profile_revision`，并由服务端从请求者资料派生其当前全部非空渠道。接受时服务端重新读取双方资料：请求者 revision 必须仍与请求一致，双方各自至少具备一个非空渠道，但不要求渠道相同；随后把双方当前非空渠道的并集写入交换记录，并原子记录 `accepted_requester_profile_revision` 和 `accepted_recipient_profile_revision`。读取 accepted 交换时再次读取双方资料，只有两个 revision 都一致才按渠道并集过滤并返回双方各自实际存在的联系方式，否则返回 `stale` 且不返回双方任何正文。

选择单一资料 revision，而不是每渠道 revision，是因为产品已明确任一联系方式变化都使整次交换失效。接受代表双方互相授权各自当时已保存的整份联系方式资料，而不是要求双方拥有相同渠道；例如 A 仅有 QQ、B 仅有微信时，A 只看到 B 的微信，B 只看到 A 的 QQ。交换记录只保存 revision 和渠道并集，不复制联系方式正文，减少旧敏感数据副本。

### 2. 服务端状态是唯一授权依据

扩展状态为 `none`、`pending`、`accepted`、`declined`、`expired`、`cancelled`、`stale`、`revoked`。客户端卡片、localStorage 和未读标记不得决定授权。`SocialGetContactExchange` 和收件箱 RPC 都执行参与者、好友关系、状态和 revision 检查。

已有但缺少接受时 revision 的 accepted 记录按 `stale` 返回。旧 profile 没有 revision 时，在读取中按初始 revision 1 解释；下一次实际修改再递增。这样不会自动继承旧授权，同时避免为了 profile 做离线数据库迁移。

### 3. 删除好友前持久撤销

Social 注册 `RegisterBeforeDeleteFriends`。hook 从请求中的 IDs 或 usernames 解析目标 UUID，并在放行 Nakama 删除前，将相关 pair 记录原子转为 `revoked`、递增 version、清除任何有效 revision 绑定。撤销写入失败则拒绝本次好友删除，避免出现已删好友但授权记录未撤销、随后重新添加导致复活的窗口。

这种顺序可能在 Nakama 后续删除失败时提前撤销授权，但隐私上属于安全失败：好友仍存在，双方需要重新授权。相比只在读取时判断好友或 after-hook 尽力撤销，它能可靠观察“曾经删除过”这一事实。

### 4. 收件箱由当前好友和 canonical pair 记录生成

新增 `SocialListContactExchangeInbox`。服务端分页列出当前正式好友，批量读取这些好友对应的确定性 pair 记录，生成：收到的 pending、发出的 pending、stale/revoked 等需要重新操作的摘要，以及 `incoming_pending_count`。响应只包含参与者 ID、玩家代码、渠道、状态、request/version 和时间，不包含联系方式正文。

不新增第二套 inbox Storage 索引，避免 canonical exchange 与索引双写不一致。在当前最多约 100 人规模下，一次好友列表加一次批量 StorageRead 成本可控；接口保留 limit/cursor 以便未来扩展。

全局徽章表示“当前收到且可处理的 pending 数”，而不是仅保存在浏览器中的已读消息数，因此刷新和换设备一致。DM 卡片仍用于低延迟刷新触发和事件历史，但投递失败只记录无敏感正文的诊断信息，不影响请求成功。

### 5. 本人资料通过身份绑定 RPC 读取

新增无 `user_id` 参数的 `SocialGetContactProfile`，只根据 Session context 中的本人 UUID 读取完整资料。客户端继续不能直接读写 Storage，也不能借此查询他人。好友页 mount 和 Session 变化时调用该 RPC 回填；保存成功后使用服务端规范化结果更新表单和摘要。

### 6. 前端以权威 inbox 呈现工作流

好友页增加独立“联系方式申请”分组，收到的 pending 直接展示接受/拒绝，发出的 pending 显示等待状态，stale 显示重新申请。`AppShell` 好友入口展示 `incoming_pending_count`。页面 mount、显式刷新、Socket 重连和有效 DM 卡片到达后重新拉取 inbox；失败时展示可重试错误，不静默降级为空。

好友详情仍通过 `SocialGetContactExchange` 获取可披露正文。任何状态变为 stale/revoked/not-friends 时，客户端先清空已显示的联系方式，再渲染状态提示。

### 7. 统一 Social RPC 错误

前端 RPC 包装层将 Nakama/Fetch Response 的 HTTP 状态、响应 JSON 和 Social `code` 转为 `SocialApiError`，UI 按 code 映射中文文案并保留未知错误的安全兜底。日志和 Toast 不包含联系方式 payload。无需新增 npm 包。

## Risks / Trade-offs

- **[风险] 一次 profile 修改会使所有好友的所有既有授权失效** → 这是已确认的隐私策略；保存前明确提示影响，并只在标准化值实际变化时递增 revision。
- **[风险] 收件箱按好友批量读取在好友规模扩大后变慢** → 当前规模可接受；使用分页和批量 StorageRead，未来再引入一致性索引。
- **[风险] before-delete 撤销成功但 Nakama 删除失败，好友仍在而授权已撤销** → 采用隐私优先的安全失败，UI 显示需要重新授权。
- **[风险] DM 投递失败导致没有即时动画或事件历史** → 权威 inbox 在进入页面、重连和刷新时恢复，服务端记录不含 PII 的失败指标。
- **[风险] accepted 旧数据统一变 stale 会打断已有用户** → 在 UI 明确说明安全升级需要重新授权，不尝试猜测旧同意对应的资料版本。
- **[风险] 并发保存、接受和删除好友发生冲突** → 使用 Storage version 条件写和接受前二次读取；冲突返回结构化 `CONFLICT` 并要求刷新。

## Migration Plan

1. 先发布包含新字段兼容读取、新 RPC、旧 accepted 安全降级和删除好友 hook 的 Go 插件。
2. 验证新旧 profile/exchange JSON 均能读取，旧 accepted 返回 stale 且不含正文。
3. 发布使用新 inbox、profile getter、状态和错误契约的 React 客户端。
4. 通过两个隔离浏览器上下文执行发起、离线恢复、接受、刷新、修改失效、删除好友重加和错误显示 E2E。
5. 观察不含 PII 的 RPC code、DM 投递失败和 Storage conflict 指标。

回滚时先回滚前端，再回滚后端。新 JSON 字段对旧 Go 解码器可忽略，但回滚旧后端会重新产生永久式授权风险，因此若必须回滚，应临时关闭 `SOCIAL_CONTACT_EXCHANGE_ENABLED`，而不是继续开放旧交换读取。

## Open Questions

- 当前无阻塞性产品问题。邮件/系统通知、批量向历史好友重新申请和跨浏览器 Profile 会话管理留待独立变更。
