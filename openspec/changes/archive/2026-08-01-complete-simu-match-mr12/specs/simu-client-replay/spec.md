## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: 战报类型包含整场比赛字段

前端 TypeScript 类型 SHALL 包含整场比赛字段：队伍 ID、胜者队伍 ID、队伍比分、阵营分配、回合阶段、加时元数据、炸弹状态、控制权快照、回合内当前武器装配、有效 seed、事件原因和最终统计。

#### Scenario: 类型检查完整比赛战报
- **GIVEN** 前端代码导入 `MatchReport`、`RoundReport` 和 `GameEvent`
- **WHEN** 在 `client` 中执行 `npx tsc --noEmit`
- **THEN** 完整比赛战报用法通过类型检查
