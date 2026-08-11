## Context

当前 React 客户端以 `AppLayout` 和固定 1920×900 画幅组织页面，根路由进入登录页，核心页面均受 Session 守卫保护。`MatchPage` 以短暂客户端等待模拟“匹配”，随后调用 `DebugSimuMatch`，服务端使用配置表阵容生成完整 `MatchReport`；`BattlePage` 已能按服务端战报逐局回放。Legacy 前端已经接入 Nakama 好友 API、单例 Socket、好友 DM 频道、历史消息和实时消息，但 DM 允许任意 `{text}` 内容。

本变更横跨 React 页面、身份生命周期、Nakama Storage/RPC 与实时消息安全。目标规模不超过 100 人同时在线，优先选择简单、可审计、可渐进上线的设计，不引入新的服务或数据库迁移。

## Goals / Non-Goals

**Goals:**

- 提供桌面和移动浏览器均自然可用的轻量三入口主界面。
- 让用户无需显式注册或登录即可完成一次 15 元组队和机器人战斗。
- 让临时教学阵容真实触发一次模拟战，并让新 UI 继续调用现有模拟与回放链路，但不保存正式阵容。
- 复用 Nakama 好友和 Socket 能力，但将 DM 收敛为服务端权威的联系方式交换卡片。
- 将个人联系方式保持为私有数据，只在双方仍为好友且交换已接受时通过 RPC 返回。

**Non-Goals:**

- 不实现 Nakama Matchmaker、路人实时对战、好友比赛房间或新的 Match Handler。
- 不实现抽卡、背包、邮件、任务、活动、货币商城或英雄经验成长。
- 不实现首发队领取、正式阵容 Storage、阵容调整、好友挑战或路人匹配；对应主入口显示未开放。
- 不修改 `matchengine` 内部比赛规则、战报 DTO 或 MatchLoop tick；本变更会新增/修改 Luban 业务配置表和 `internal/match` 输入适配层。
- 不开放玩家自由文本昵称、私聊、图片、语音或外链消息。

## Decisions

### 1. 使用响应式 Web Shell，而不是虚拟 19:9 画布

新 `AppShell` 使用 CSS Grid/Flex、流式宽度和断点布局。桌面内容区设置约 1200px 的最大宽度，移动端以约 390px 视口为基准，但不将任一宽高写成必须裁剪的画布。三项主操作在桌面可并列或主次分区，在窄屏纵向堆叠；安全区、触控尺寸和文本换行属于基础布局约束。

备选方案是继续缩放 1920×900 画布。它能复用旧视觉，但会在手机浏览器产生不可访问的小字号、裁剪和空白，因此拒绝。

### 2. “不登录可玩”采用静默设备身份，产品层仍视为访客

首次进入时客户端生成并持久化设备 ID，通过 Nakama device authentication 创建或恢复 Guest Session，不展示登录表单。这样 `DebugSimuMatch`、限流和审计仍有稳定 user ID。访客的首次体验标记和临时 15 元阵容保存在本机；临时阵容不写入正式玩家阵容 Storage。

用户选择登录/注册时，优先把邮箱凭据 link 到当前设备账号；若产品选择登录到已有账号，则切换 Session，访客临时阵容不自动覆盖已有阵容。该边界避免无提示的数据合并。

备选方案是开放完全未认证 RPC。它减少一次静默认证，却失去可靠限流和用户关联，并扩大滥用面，因此拒绝。

### 3. 路由按“可游玩”和“需账号能力”分层

`/` 为公开主界面；`/onboarding/lineup`、`/match/computer` 与 `/battle` 接受 Guest Session；`/friends` 和账号资料要求正式登录。`/lineup` 与 `/match/friend` 统一显示未开放状态，不因登录而解锁。访客点击真正需要账号的好友社交入口时显示轻量登录 Modal，并在认证成功后返回原目标。旧 `/home` 等路径提供临时重定向，避免收藏链接直接失效。

首次体验 Modal 由版本化本地键控制，例如 `onboarding.battleOffer.v1`。Modal 的“以后再说”只关闭当前版本提示，不阻止用户从界面再次主动进入体验。

### 4. 教学阵容是一次性战斗输入，不是玩家资产

15 元体验阵容的选择草稿只保存在当前客户端教学流程中，价格档、候选范围、机器人对手与地图来自 `TbTutorialBattle`。开始比赛时客户端只提交 tutorial config ID、config version 与五个 Player ID；服务端重新读取配置，校验人数、唯一性、每名选手所在价格档、总价不超过预算，并从配置取得机器人五人阵容。校验通过后，`internal/match` 将双方 `TbPlayer` 档案转换为 `matchengine.TeamInput`，构造真实 `MatchInput`。客户端不能提交属性、价格或比赛结果。

比赛结束、退出流程、刷新或切换账号均不把这五人写入玩家阵容 Storage。本变更不创建 `internal/playerlineup`，也不注册 `PlayerGetLineup`、`PlayerChooseStarterLineup` 或 `PlayerUpdateLineup`。主界面的“调整阵容”入口显示未开放。

### 5. 生产模拟 RPC 复用现有 matchengine 与一次性战报链路

当前 `matchengine` 已能消费两支五人 `TeamInput`，但 `DebugSimuMatchRequest` 目前只有 `map_id` 和 `seed`，`internal/match.Service` 仍固定构造 Falcons/Vitality。新增生产用 `SimuMatch` RPC，按 `mode=tutorial|computer` 接收配置 ID 与玩家阵容引用，并在服务端校验后调用现有 `engine.Simulate`。`DebugSimuMatch` 保留默认阵容、seed 和调试日志兼容能力，两者共享构建/模拟服务，避免复制比赛逻辑。

新“匹配对战”页面先呈现“电脑对战”；点击后调用 `SimuMatch`，收到完整 `MatchReport` 后进入新 Battle UI。客户端的等待动画只表达请求阶段，不宣称已经接入真实玩家池。服务端战报仍是比分、事件、阵亡状态、胜者和统计的唯一权威来源；响应式界面只派生当前播放位置的可见状态。

由于本期没有玩家正式阵容，`mode=computer` SHALL 复用当前默认模拟对局：服务端从 `TbTeam`/`TbPlayer.teamId` 解析配置的两支默认队伍并直接生成战报，不宣称其中一方是玩家持有阵容。`mode=tutorial` 才使用本次 15 元临时选择作为 Team A。两种模式均只触发战斗，不产生阵容资产。

不在本变更中把 `SimuMatch` 包装成 Nakama Matchmaker，因为真实玩家匹配需要队列、房间、断线恢复和权威对局生命周期，是独立架构变更。

### 6. Team 与教学战使用独立 Luban 表

新增自动导入的 `#Team.xlsx`，生成 `TbTeam`。本期字段为 `id`、`name`、`shortName`、`nickname`、`logo`。`#Player.xlsx` 将现有自由文本 `team` 替换为带 `TbTeam` 引用校验的 `teamId`，未来图鉴和当前比赛展示均按 teamId 查询，不以展示文案关联数据。首发候选和展示顺序等阵容功能字段留给后续阵容提案，不在本期预埋。

新增自动导入的 `#TutorialBattle.xlsx`，生成 `TbTutorialBattle`。一行表示一个可版本化教学方案，建议字段为 `id`、`enabled`、`budget`、`rosterSize`、`mapId`、`tier5PlayerIds` 至 `tier1PlayerIds`、`opponentTeamId`、`opponentPlayerIds`。所有 Player ID 引用 `TbPlayer`；服务初始化时强校验价格档不重复、对手恰好五人、ID 存在、预算可组成合法五人阵容。

备选方案是把价格写回 Player 表。该价格只属于特定教学玩法，未来不同活动可能使用不同分档，放入独立 TutorialBattle 表更符合配置归属。

### 7. Social 子系统以 RPC + 私有 Storage 管理联系方式授权

新增 `server/internal/social`，采用 `model.go`、`repository.go`、`service.go`、`api_rpc.go`、`hooks.go`。建议 Storage：

- `Collection=social_contact_profile, Key=profile, UserID=<本人 UUID>`：保存本人 QQ、微信及更新时间，`PermissionRead=0`、`PermissionWrite=0`，因此即使 owner 本人也只能通过 RPC 访问。
- `Collection=social_contact_exchange, Key=<排序后的 pair key>, UserID=00000000-0000-0000-0000-000000000000`：系统 owner 记录 requester、recipient、requested channels、status、version 与时间戳，同样设置 `PermissionRead=0`、`PermissionWrite=0`。

这些对象物理上由 Nakama Storage Engine 持久化到 Nakama 使用的 PostgreSQL 数据库，但项目不建立自定义 SQL 表，也不让客户端直接读取 JSON。Go runtime 作为服务器权威方可绕过对象权限，通过 `StorageRead/StorageWrite` 访问；客户端只能经 Social RPC 获得经过好友关系与授权状态过滤的最小响应。

建议 RPC：

- `SocialSetContactProfile`
- `SocialGetContactExchange`
- `SocialRequestContactExchange`
- `SocialRespondContactExchange`

服务端在每次请求和读取时验证 Session、双方当前仍为好友、状态迁移、请求频率和 Storage version。只有 `accepted` 状态才返回双方已授权渠道的实际值；频道消息永不包含 QQ/微信正文。

### 8. DM 只承载服务器生成的结构化卡片事件

用户发起或响应交换时调用 Social RPC。事务状态写入成功后，服务端使用 Nakama `ChannelIdBuild` 与 `ChannelMessageSend` 向现有 persistent direct-message channel 写入结构化消息，例如 `{type, request_id, action, version}`。客户端通过现有 `onchannelmessage` 和历史加载呈现卡片，并通过 RPC 获取权威状态。

`InitModule` 注册 `RegisterBeforeRt` 钩子，拒绝客户端的 ChannelMessageSend、ChannelMessageUpdate 和 ChannelMessageRemove。服务端模块直接发送的卡片不经过客户端 before-hook。仅移除文本输入框不构成安全措施，因此必须同时实施该钩子。

好友删除后，任何 Social read RPC 都重新校验好友关系并拒绝返回联系方式。历史卡片可以保留为事件记录，但卡片显示“已失效”，不缓存或展示此前返回的联系方式。

### 9. 公共玩家标识采用 Nakama 唯一 username 短码

Nakama 的内部 `user.id` 是 UUID，不适合作为短展示 ID；未传 username 时，Nakama 3.30 默认生成 10 位大小写字母 username。项目 SHALL 覆盖默认行为：在创建 device/email 账号时生成 8 位大写英文字母与数字组成的唯一 Nakama username，数据库唯一约束冲突时重新生成并重试。数据库保存裸 code（例如 `7K2M9QXA`），UI 统一展示 `玩家#7K2M9QXA`，好友搜索直接使用该唯一 username。

客户端不提供自定义公开昵称；服务端认证/update-account hook 只允许项目格式，阻止修改客户端绕过 UI 提交敏感名称。由于产品尚未上线，现有测试账号可一次性清理或批量规范化，无需惰性兼容迁移。内部关联、Storage owner 和权限判断始终使用完整 UUID user ID，8 位 code 只用于公开展示和查找。

### 10. 不增加第三方依赖

布局、Modal、状态和 RPC 封装使用项目已有 React、路由、状态工具和 Nakama SDK；Go 侧仅使用当前 Nakama runtime 与标准库。无需新增 Go module 或 npm 包。

## Risks / Trade-offs

- [访客实际上拥有静默 Nakama 设备账号，可能与用户对“未登录”的理解不一致] → UI 不称其为注册账号，不收集联系方式；在隐私说明中披露本地设备标识用途。
- [设备丢失或清理浏览器数据会失去访客状态] → 本期只提供一次性教学，不承诺保存教学阵容或访客进度。
- [8 位随机 code 理论上可能碰撞] → 使用 Nakama username 唯一约束作为最终仲裁，认证创建遇到 AlreadyExists 时生成新 code 并有限次重试；不从 UUID 截断后假设唯一。
- [教学表配置可能无法组成五人且 15 元以内的阵容] → 导表测试和服务初始化强校验所有引用、档位互斥、预算可解性和对手五人完整性。
- [联系方式属于敏感个人数据] → Storage 客户端零直读、RPC 最小披露、好友关系每次重验、拒绝任意客户端频道写入，并避免日志打印字段值。
- [服务端写交换状态成功但卡片发送失败] → Storage 是权威来源；RPC 返回成功，客户端刷新状态即可恢复，消息发送记录错误并允许幂等补发。
- [双方并发发起或响应导致状态覆盖] → 使用确定性 pair key、Storage version 与幂等 request ID；冲突时重读后返回当前权威状态。
- [禁止全部客户端频道写入会影响未来其他聊天功能] → hook 仅按本产品允许的 realtime op 建立明确策略；若未来需要其他频道类型，按 channel type/metadata 建白名单，不恢复任意 DM 文本。
- [旧页面规格与新页面并存导致测试冲突] → delta specs 明确移除或替换旧大厅、抽卡入口和文本消息契约；迁移期通过路由重定向而非同时维护两套首页。

## Migration Plan

1. 先增加 Guest Session、正式身份判定、阵容与 Social 服务端能力及自动化测试，不改变旧页面入口。
2. 在测试环境启用频道写入 hook，验证旧文本发送被拒绝、服务端卡片可送达、好友删除后联系方式不可读。
3. 构建新响应式 App Shell、主界面、阵容、匹配、好友与 Battle 表现，使用旧路由重定向做并行验收。
4. 切换根路由到新主界面，移除旧大厅和文本聊天入口；保留可回滚的前端构建产物。
5. 监控 RPC 错误、Storage 冲突和 Socket 卡片投递。若 Social 出现问题，可关闭联系方式交换入口并保留好友列表；若新 UI 出现严重回归，可回滚前端而不删除新 Storage 数据。

## Open Questions

- `#Team.xlsx` 的正式 team ID、中文昵称和 Logo 仍需策划填表，但不阻塞页面与数据结构实施；本期不配置首发队顺序。
