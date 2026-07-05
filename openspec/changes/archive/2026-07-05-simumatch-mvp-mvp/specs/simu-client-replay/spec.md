## ADDED Requirements

### Requirement: MatchPage "开始匹配" button triggers DebugSimuMatch

`client/src/pages/MatchPage.tsx` 中的「开始匹配」按钮 SHALL 在 MVP 阶段先调用 `DebugSimuMatch` RPC，待收到响应后再导航到 `/battle`。

#### Scenario: Click starts debug simulation

- **GIVEN** 用户已登录并访问 `/match`
- **WHEN** 用户点击「开始匹配」按钮
- **THEN** 前端调用 `client.rpc(session, "DebugSimuMatch", { map_id: "de_dust2" })`
- **AND** 按钮显示加载状态（如 disabled 或 spinner）
- **AND** 调用成功后导航到 `/battle`

### Requirement: BattlePage logs the match report to console

`BattlePage` SHALL 在挂载后将 `DebugSimuMatch` 返回的战报对象打印到浏览器控制台，便于开发者直接查看协议结构。

#### Scenario: Report logged on mount

- **GIVEN** 用户从 `MatchPage` 点击「开始匹配」并进入 `/battle`
- **WHEN** `BattlePage` 挂载
- **THEN** 浏览器控制台输出完整 `MatchReport` JSON

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

`BattlePage` SHALL 使用 `useEffect` + `setInterval` 每 1 秒推进一条事件，模拟「正在实时在打」的播报节奏。

#### Scenario: Events appear one by one

- **GIVEN** 战报包含 `ROUND_START`、3 个 `KILL` 和 1 个 `ROUND_END` 事件
- **WHEN** 用户进入 `/battle`
- **THEN** 事件按时间顺序逐条出现，相邻事件间隔约 1 秒
- **AND** 比分、阵容、地图标记随事件推进同步更新

### Requirement: Scoreboard displays server-returned team names and live score

`Scoreboard` SHALL 读取 `MatchInfo.team_a_name` / `team_b_name` 作为队名，并在 `ROUND_END` 事件出现后更新为当前回合比分。

#### Scenario: Scoreboard reflects simulation result

- **GIVEN** 服务器返回队名「测试T队」和「测试CT队」
- **WHEN** `ROUND_END` 事件播放完毕
- **THEN** 上方比分牌显示正确队名和 T/CT 比分

### Requirement: Team rosters reflect player alive/dead state

`TeamRoster` / `PlayerCard` SHALL 读取服务器阵容，并根据已播放的击杀事件动态将死亡选手置灰、存活选手保持彩色。

#### Scenario: Dead players turn gray

- **GIVEN** 某选手在已播放事件中被击杀
- **WHEN** 其对应的 `PlayerCard` 渲染
- **THEN** 该卡片显示为灰色/低饱和度样式
- **AND** 其 K/D 统计同步更新

### Requirement: MapView renders Dust2 radar and kill markers

`MapView` SHALL 渲染 `client/public/csmaps/de_dust2_radar_trans.webp` 雷达图，并为每个已播放且带 `location` 的 `KILL` 事件在对应坐标上显示红色 ×。

#### Scenario: Kill markers appear on map

- **GIVEN** 战报中的 `KILL` 事件包含 `location: { name, x, y }`
- **WHEN** 该事件在回放中到达
- **THEN** 地图对应比例位置出现红色 ×
- **AND** 标记随事件推进逐条出现，不提前显示

### Requirement: EventFeed shows real-time stats dialog

`EventFeed` 标题栏 SHALL 提供一个靠右的「数据统计」按钮；点击后弹出 shadcn 风格 Dialog，展示 A/B 两队选手的 K/D/A 表格，并随比赛进行实时更新。

#### Scenario: Stats dialog updates live

- **GIVEN** 比赛进行中某选手完成 1 次击杀并死亡
- **WHEN** 用户点击「数据统计」
- **THEN** 弹窗中该选手的 K 列显示 1、D 列显示 1
- **AND** 死亡选手行以灰色显示
