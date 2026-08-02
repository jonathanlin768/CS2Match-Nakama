# simu-client-replay Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
### Requirement: MatchPage "开始匹配" button triggers DebugSimuMatch

`client/src/pages/MatchPage.tsx` 中的「开始匹配」按钮 SHALL 在 MVP 阶段先调用 `DebugSimuMatch` RPC，待收到响应后再导航到 `/battle`。

#### Scenario: Click starts debug simulation

- **GIVEN** 用户已登录并访问 `/match`
- **WHEN** 用户点击「开始匹配」按钮
- **THEN** 前端调用 `client.rpc(session, "DebugSimuMatch", { map_id: "de_dust2" })`
- **AND** 按钮显示加载状态（如 disabled 或 spinner）
- **AND** 调用成功后导航到 `/battle`

### Requirement: BattlePage logs the match report to console

`BattlePage` SHALL 仅在 `DebugSimuMatch` 响应的 `debug_enabled` 为 `true` 时，将完整战报及回放流程输出到浏览器 Console；当该值为 `false` 时 SHALL 不输出模拟战报调试日志。

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

`BattlePage` SHALL 在现有 mock UI 基础上，额外渲染一个最简事件列表，展示本回合的 `ROUND_START`、`KILL`、`ROUND_END` 事件文本。

#### Scenario: Kill feed visible

- **GIVEN** 战报中包含 3 个击杀事件和 1 个回合结束事件
- **WHEN** `BattlePage` 渲染
- **THEN** 页面某处（如右下角或底部）显示这些事件的文本列表
- **AND** 每个 `KILL` 事件显示攻击者、被击杀者和位置

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

`BattlePage` 现有的 `battleState` mock 数据和组件（`Scoreboard`、`TeamRoster`、`MapView`、`KillFeed`）在 MVP 阶段 SHALL 继续保留，真实战报只用于控制台和最简列表输出。

#### Scenario: BattlePage still renders mock layout

- **GIVEN** 用户未点击「开始匹配」直接访问 `/battle`
- **WHEN** `BattlePage` 渲染
- **THEN** 仍显示 `battleState` 定义的初始比分、阵容和击杀播报
- **AND** 页面不崩溃

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

`EventFeed` 标题栏 SHALL 提供一个靠右的「数据统计」按钮；点击后弹出 shadcn 风格 Dialog，展示 A/B 两队选手的 K/D/A 表格，并随比赛进行实时更新。

#### Scenario: Stats dialog updates live

- **GIVEN** 比赛进行中某选手完成 1 次击杀并死亡
- **WHEN** 用户点击「数据统计」
- **THEN** 弹窗中该选手的 K 列显示 1、D 列显示 1
- **AND** 死亡选手行以灰色显示

### Requirement: 战报类型包含整场比赛字段

前端 TypeScript 类型 SHALL 包含整场比赛字段：队伍 ID、胜者队伍 ID、队伍比分、阵营分配、回合阶段、加时元数据、炸弹状态、控制权快照、回合内当前武器装配、有效 seed、事件原因和最终统计。

#### Scenario: 类型检查完整比赛战报
- **GIVEN** 前端代码导入 `MatchReport`、`RoundReport` 和 `GameEvent`
- **WHEN** 在 `client` 中执行 `npx tsc --noEmit`
- **THEN** 完整比赛战报用法通过类型检查

