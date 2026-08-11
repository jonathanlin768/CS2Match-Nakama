# simu-client-replay Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
### Requirement: MatchPage "开始匹配" button triggers DebugSimuMatch
新匹配页面中的“电脑对战”和教学战操作 SHALL 调用生产用 `SimuMatch` RPC，收到有效完整 `MatchReport` 后导航到 `/battle`。该操作 SHALL 同时支持 Guest Session 和正式 Session，并 SHALL 明确描述为电脑模拟而不是真实玩家 Matchmaker；`DebugSimuMatch` 只保留给调试和兼容测试入口。

#### Scenario: 正式玩家开始电脑模拟
- **GIVEN** 正式玩家访问匹配对战页面
- **WHEN** 玩家点击“电脑对战”
- **THEN** 前端使用当前 Session 调用 `SimuMatch`
- **AND** 调用成功后携带或缓存战报并导航到 `/battle`

#### Scenario: 访客开始教学模拟
- **GIVEN** 访客已完成合法的 15 元教学阵容
- **WHEN** 访客点击“开始战斗”
- **THEN** 前端使用 Guest Session 调用 `SimuMatch`，提交 tutorial config ID、version 与五个 Player ID
- **AND** 调用成功后导航到响应式 Battle 页面

### Requirement: BattlePage logs the match report to console

`BattlePage` SHALL 仅在 `DebugSimuMatch` 或正式 `SimuMatch` 响应的 `debug_enabled` 为 `true` 时，将完整战报及回放流程输出到浏览器 Console；当该值为 `false` 时 SHALL 不输出模拟战报调试日志。

#### Scenario: 调试开关开启
- **GIVEN** 战报包含 `debug_enabled: true`
- **WHEN** BattlePage 挂载并开始逐事件回放
- **THEN** Console 包含完整原始战报、逐回合摘要、逐事件/选手状态数据
- **AND** 每次可见事件推进时输出对应回合和事件序号

#### Scenario: 调试开关关闭
- **GIVEN** 战报包含 `debug_enabled: false`
- **WHEN** BattlePage 挂载并完成回放
- **THEN** Console 不包含由模拟战报调试器生成的日志

### Requirement: BattlePage renders a minimal HLTV-style kill feed

`BattlePage` SHALL 根据事件结构化字段渲染本地化实时播报，覆盖战术、伤害、击杀、移动决策、控制权、炸弹生命周期和回合终局；不得直接把 `damage applied`、`decision resolved into real actions` 等后端通用英文 `message` 作为面向玩家的主要内容。

#### Scenario: Kill feed visible

- **GIVEN** 战报中包含 3 个击杀事件和 1 个回合结束事件
- **WHEN** `BattlePage` 渲染
- **THEN** 页面某处（如右下角或底部）显示这些事件的文本列表
- **AND** 每个 `KILL` 事件显示攻击者、被击杀者和位置
- **AND** `DAMAGE` 显示攻击者、受击者、位置和实际伤害
- **AND** `ROUND_END` 显示获胜队伍及其当局 T/CT 阵营

### Requirement: Frontend defines TypeScript types for match report

`client/src/types/sim.ts` 或新增 `client/src/types/match-report.ts` SHALL 定义与服务端 `MatchResult` 对应的 TypeScript 类型。

#### Scenario: Type checking passes

- **GIVEN** 前端代码使用 `MatchReport`、`GameEvent`、`PlayerMatchStats` 等类型
- **WHEN** 执行 `npx tsc --noEmit`
- **THEN** 无类型错误

### Requirement: RPC call handles errors gracefully

前端调用 `DebugSimuMatch` 失败时 SHALL 显示错误提示，并允许用户重试。

#### Scenario: Server returns error

- **GIVEN** 网络断开或服务端返回 `SIMULATION_ERROR`
- **WHEN** 用户点击「开始匹配」
- **THEN** 按钮恢复可点击状态
- **AND** 页面显示 Toast 或提示文本：“模拟战斗失败，请重试”

### Requirement: Mock UI remains as visual placeholder
BattlePage SHALL 不再以旧固定画幅 mock `battleState` 作为正常回放数据。存在有效 `MatchReport` 时，比分、阵容状态、事件流、地图标记、胜者和统计 SHALL 来自服务端战报；直接访问且没有战报时 SHALL 显示可恢复的空状态并提供返回匹配入口，不得展示一场伪造比赛。

#### Scenario: 有效战报渲染
- **GIVEN** 用户从电脑模拟获得有效 `MatchReport`
- **WHEN** BattlePage 渲染
- **THEN** 页面从战报建立初始回放状态并开始播放
- **AND** 不使用 mock 比分或 mock 胜者覆盖服务器数据

#### Scenario: 无战报直接访问
- **GIVEN** 路由状态和客户端缓存中都没有有效战报
- **WHEN** 用户直接访问 `/battle`
- **THEN** 页面显示“暂无可播放比赛”等空状态
- **AND** 提供返回匹配对战的操作

### Requirement: BattlePage replays events at 1-second intervals

`BattlePage` SHALL 使用回放控制器推进完整比赛战报。它 SHALL 按当前回合内的时间戳顺序逐条展示事件，然后根据选中的回放模式进入下一回合。默认自动节奏 SHALL 约为每秒显示一条可见事件。

#### Scenario: 事件跨多个回合出现
- **GIVEN** 比赛战报包含第 1 回合和第 2 回合事件
- **WHEN** 用户进入 `/battle` 并让回放继续
- **THEN** 第 1 回合事件按顺序出现
- **AND** 第 1 回合结束后，第 2 回合成为当前回合
- **AND** 第 2 回合开始前不提前显示后续未播放回合或完整局数
- **AND** 比分、阵容、事件流和地图标记根据当前回放位置的服务端状态更新

#### Scenario: 跳过会停止逐事件播放
- **GIVEN** 比赛回放正在第 4 回合
- **WHEN** 用户点击“跳过比赛”
- **THEN** 基于 interval 的播放停止
- **AND** 页面立即显示最终战报状态
- **AND** 用户仍可通过回合时间线回看任意回合

#### Scenario: 跳过后选择历史回合从头播放
- **GIVEN** 用户已跳过比赛并展开全部回合
- **WHEN** 用户点击任意历史回合
- **THEN** 页面只显示该回合的 `ROUND_START` 初始状态
- **AND** 用户再次点击播放时从该回合继续，而不是立即进入下一回合

### Requirement: Scoreboard displays server-returned team names and live score

`Scoreboard` SHALL 读取服务端返回的队伍名称和队伍比分，并使用 `RoundResult.TeamTID`、`RoundResult.TeamCTID` 或等价元数据显示当前回合中每支队伍的当前阵营。

#### Scenario: 计分板反映模拟结果
- **GIVEN** 服务端返回队伍名称 `Falcons` 和 `Vitality`
- **WHEN** 第 13 回合的 `ROUND_END` 事件被回放
- **THEN** 计分板显示服务端提供的队伍名称
- **AND** 计分板显示第 13 回合后的队伍比分
- **AND** 阵营徽标展示半场后的 T/CT 分配

### Requirement: Team rosters reflect player alive/dead state

`TeamRoster` / `PlayerCard` SHALL 读取服务端选手快照和当前回合已可见事件增量。选手阵营、武器、存活/死亡状态和 K/D 展示 SHALL 根据服务端状态和回放位置更新。

#### Scenario: 死亡选手变灰
- **GIVEN** 某选手在当前回合中被一个已可见 `KILL` 事件击杀
- **WHEN** 该选手的 `PlayerCard` 渲染
- **THEN** 卡片显示为死亡状态
- **AND** 该选手死亡数反映可见服务端事件/状态

### Requirement: MapView renders Dust2 radar and kill markers

`MapView` SHALL 渲染 `client/public/csmaps/de_dust2_radar_trans.webp`，并为当前回合中包含服务端位置的可见事件显示标记。

#### Scenario: 地图上出现击杀标记
- **GIVEN** 当前回合包含一个 `location: { name, x, y }` 的 `KILL` 事件
- **WHEN** 该事件在回放中变为可见
- **THEN** 地图在对应归一化坐标显示标记
- **AND** 尚未显示的未来回合标记不会出现在当前回合视图中

### Requirement: EventFeed shows real-time stats dialog

`EventFeed` 标题栏 SHALL 提供一个靠右的「数据统计」按钮；点击后弹出 shadcn 风格 Dialog，展示 A/B 两队选手截至当前回放位置的跨回合累计 K/D/A，并随比赛进行实时更新。

#### Scenario: Stats dialog updates live

- **GIVEN** 比赛进行中某选手在此前回合完成 1 次击杀，并在当前回合死亡且获得 1 次助攻
- **WHEN** 用户点击「数据统计」
- **THEN** 弹窗中该选手的 K、D、A 列均显示 1
- **AND** 统计不会在切换到新回合时清零
- **AND** 死亡选手行以灰色显示

### Requirement: 战报类型包含整场比赛字段

前端 TypeScript 类型 SHALL 包含整场比赛字段：队伍 ID、胜者队伍 ID、队伍比分、阵营分配、回合阶段、加时元数据、炸弹状态、控制权快照、回合内当前武器装配、有效 seed、事件原因和最终统计。

#### Scenario: 类型检查完整比赛战报
- **GIVEN** 前端代码导入 `MatchReport`、`RoundReport` 和 `GameEvent`
- **WHEN** 在 `client` 中执行 `npx tsc --noEmit`
- **THEN** 完整比赛战报用法通过类型检查

### Requirement: 战斗阵容使用策划配置的 5:7 头像裁切

BattlePage 的队伍阵容 SHALL 优先使用比赛快照中的完整卡面和裁切参数，在固定 `5:7` 视口中动态显示选手头像。裁切属于客户端展示，不得修改回放状态或比赛结果。

#### Scenario: 在战斗选手卡中还原头像裁切

- **GIVEN** `PlayerState` 包含可读取的 `cardImage` 和合法的 `avatarCrop`
- **WHEN** TeamRoster 渲染 PlayerCard
- **THEN** PlayerCard 使用完整卡面作为源图
- **AND** 依据源图自然尺寸及归一化矩形在 `5:7` 视口内准确显示裁切区域
- **AND** 不依赖预先生成的第二张头像图片

#### Scenario: 回放状态变化不改变裁切区域

- **GIVEN** 战斗回放正在逐事件更新选手状态
- **WHEN** 选手生命值、存活状态或装备发生变化
- **THEN** 头像继续使用快照中同一组裁切参数
- **AND** 死亡灰化等状态样式叠加在裁切结果之上

#### Scenario: 卡面或裁切不可用时逐级回退

- **GIVEN** `cardImage` 加载失败或 `avatarCrop` 不合法
- **WHEN** PlayerCard 渲染头像
- **THEN** 客户端首先回退到现有 `portrait`
- **AND** `portrait` 也不可用时回退到站点默认选手图
- **AND** 阵容布局和回放流程不因图片错误中断

### Requirement: Battle 回放界面响应不同视口
BattlePage SHALL 在不改变 `MatchReport` 权威含义和回放顺序的前提下采用响应式布局。桌面端 SHALL 同时容纳主要赛况与辅助信息；移动端 SHALL 对比分、当前事件和回放控制保持优先，并允许用户切换阵容、地图和统计区域。

#### Scenario: 桌面回放
- **GIVEN** 视口宽度不小于 1024px 且存在有效战报
- **WHEN** 回放进行
- **THEN** 比分、主要事件和比赛场景无需横向滚动即可查看
- **AND** 辅助阵容或统计可在同页合理呈现

#### Scenario: 手机回放
- **GIVEN** 视口宽度约为 390px 且存在有效战报
- **WHEN** 回放进行
- **THEN** 当前比分、回合和播放控制保持可见或易于访问
- **AND** 页面不通过整体缩小固定桌面画布来适配

### Requirement: 匹配请求阶段不伪造真实玩家状态
客户端 SHALL 区分“正在请求电脑模拟”和未来真实 Matchmaker 状态。等待 `SimuMatch` 响应时只能展示模拟请求、准备比赛或生成战报等准确反馈。

#### Scenario: 模拟 RPC 等待中
- **GIVEN** `SimuMatch` 请求尚未返回
- **WHEN** 匹配页面展示 loading
- **THEN** 文案表示正在准备电脑比赛或生成战报
- **AND** 不显示已找到某个真实玩家、实时队列人数或伪造对手 ID

