## MODIFIED Requirements

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

## ADDED Requirements

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
