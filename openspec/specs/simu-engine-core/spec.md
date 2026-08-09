# simu-engine-core Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
### Requirement: matchengine exposes Simulate entrypoint

`internal/framework/matchengine.Service` SHALL 提供 `Simulate(ctx, *MatchInput) (*MatchResult, error)` 方法，作为战斗推演的唯一入口。

#### Scenario: Simulate returns a result

- **GIVEN** 调用方构造了有效的 `MatchInput`（含 TeamA、TeamB、MapID、Seed）
- **WHEN** 调用 `s.engine.Simulate(ctx, input)`
- **THEN** 方法返回非 nil 的 `*MatchResult`
- **AND** 返回结果包含一个回合的战报和最终统计

### Requirement: MatchEngine uses per-match seed for reproducibility

每个 `MatchEngine` 实例 SHALL 使用独立的随机源，种子来自 `MatchInput.Seed`。

#### Scenario: Same seed produces same route selection

- **GIVEN** 使用相同阵容和相同 Seed 调用两次 `Simulate`
- **THEN** 两次返回的 `RoundResult.RouteMain` 相同
- **AND** 两次返回的击杀事件顺序相同

### Requirement: Combat resolution uses simple comparison

MVP 阶段对决 SHALL 采用简化“比大小”规则：双方选手比较综合属性（如 `Entry + Aim + Firepower` 加权和），高者击杀低者，败者标记为死亡。

#### Scenario: Higher stat player wins duel

- **GIVEN** T 方选手综合属性为 250，CT 方选手综合属性为 200
- **WHEN** 引擎执行该对决
- **THEN** T 方选手存活，CT 方选手死亡
- **AND** 生成一个 `KILL` 事件，攻击者为 T 方选手，被击杀者为 CT 方选手

### Requirement: Engine produces valid round result

`RoundResult` SHALL 包含回合编号、进攻方、胜者、获胜原因、当前比分、主路线、事件列表和选手状态。

#### Scenario: Round result fields

- **WHEN** 一回合推演完成
- **THEN** `RoundResult.RoundNumber` 为 1
- **AND** `RoundResult.Winner` 为 `"T"` 或 `"CT"`
- **AND** `RoundResult.WinReason` 为 `"elimination"` 或 `"timeout"`
- **AND** `RoundResult.ScoreT` 和 `RoundResult.ScoreCT` 为 0 或 1
- **AND** `RoundResult.Events` 按发生顺序排列

### Requirement: Engine produces HLTV-style events

`GameEvent` SHALL 支持 `ROUND_START`、`KILL`、`ROUND_END` 三种事件类型，并携带足够的前端展示字段。

#### Scenario: Round start event

- **WHEN** 回合开始
- **THEN** 生成 `EventType = "ROUND_START"` 的事件
- **AND** `Message` 包含回合编号和进攻路线信息

#### Scenario: Kill event

- **WHEN** 对决产生击杀
- **THEN** 生成 `EventType = "KILL"` 的事件
- **AND** 字段包含 `attacker_id`、`attacker_name`、`victim_id`、`victim_name`、`weapon`、`location`
- **AND** `location` 为结构化对象 `{ name: string, x: number, y: number }`，`x`/`y` 为地图比例坐标（0.0 ~ 1.0）

### Requirement: Engine generates map coordinates for kill locations

`GameEvent.Location` SHALL 使用结构化的 `{ name, x, y }` 坐标对象，便于前端在雷达图上定位击杀标记。

#### Scenario: Kill location has normalized coordinates

- **GIVEN** 当前回合主路线为 `A_Long`
- **WHEN** 生成击杀事件
- **THEN** `Location.Name` 为路线显示名（如 "A大"）
- **AND** `Location.X` / `Location.Y` 取自该路线预定义的地图坐标，并带有小幅随机偏移

### Requirement: Route configuration includes map positions

MVP 阶段 `matchengine` 内建的每条 `RouteConfig` SHALL 关联一个雷达图比例坐标；引擎提供 `RandomRouteLocation(routeID, rng)` 用于生成带随机偏移的击杀位置。

#### Scenario: Route has map position

- **GIVEN** 路线 ID 为 `A_Long`
- **WHEN** 调用 `RandomRouteLocation("A_Long", rng)`
- **THEN** 返回非 nil 的 `*Location`
- **AND** 坐标落在 Dust2 雷达图的可视区域内

#### Scenario: Round end event

- **WHEN** 回合胜负已分
- **THEN** 生成 `EventType = "ROUND_END"` 的事件
- **AND** `Message` 包含获胜方和当前比分

### Requirement: Engine computes final player stats

`MatchResult.FinalStats.PlayerStats` SHALL 包含每名选手整场比赛的击杀、死亡、ADR、首杀数和多杀次数。

#### Scenario: Stats after one round

- **GIVEN** 某选手在一回合中击杀 2 人并死亡
- **WHEN** 推演结束
- **THEN** 该选手的 `PlayerMatchStats.Kills` 为 2
- **AND** `Deaths` 为 1
- **AND** `ADR` 为造成伤害除以回合数（1 回合）
- **AND** `FK` 和 `MK` 根据事件标记正确计算

### Requirement: 引擎从 MatchInput 模拟完整比赛

`internal/framework/matchengine.Service` SHALL 使用 `MatchInput`，根据其中的 `RuleSet`、地图配置快照、队伍快照、阵营分配、武器、seed 和请求模式模拟完整比赛。

#### Scenario: 完整比赛输出
- **GIVEN** 一个有效 `MatchInput`，包含 `de_dust2` 上的两支五人队伍
- **WHEN** `s.engine.Simulate(ctx, input)` 成功返回
- **THEN** `MatchResult.Rounds` 包含判定胜者所需的全部常规赛和加时回合
- **AND** `MatchResult.TotalRounds` 等于 `len(MatchResult.Rounds)`
- **AND** `MatchResult.WinnerTeamID` 标识获胜队伍

### Requirement: 引擎派生稳定的回合 seed

`matchengine` SHALL 基于 `MatchInput.Seed`、地图版本、规则集 ID、回合号和子系统 key，派生逐回合与逐子系统 seed。

#### Scenario: 回合 seed 隔离事件位置采样
- **GIVEN** 两个相同且 seed 相同的比赛输入
- **WHEN** 代码变更在事件位置采样前新增了决策分支，但没有改变事件来源数据
- **THEN** 由 `LocationSeed` 生成的事件位置对相同事件 ID 保持稳定

### Requirement: 引擎返回扩展后的回合公共状态

每个 `RoundResult` SHALL 包含回合号、阶段、队伍阵营 ID、获胜阵营、获胜队伍 ID、获胜原因、队伍比分、路线/模板 ID、事件列表、选手公共状态、炸弹公共状态和最终控制权快照。

#### Scenario: 回合结果包含阵营和队伍身份
- **WHEN** 一个换边后的回合完成
- **THEN** `RoundResult.TeamTID` 标识当前作为 T 方的队伍
- **AND** `RoundResult.TeamCTID` 标识当前作为 CT 方的队伍
- **AND** `RoundResult.WinnerTeamID` 匹配赢下该回合的队伍

### Requirement: 回合事件、公共状态与获胜原因保持一致

`matchengine` SHALL 只生成能够由回合事件和最终公共状态共同证明的获胜原因，不得在一方全灭后继续产生下包事件。

#### Scenario: 淘汰终局
- **WHEN** 回合以 `elimination` 结束
- **THEN** 败方最终存活人数为 0
- **AND** 该回合不包含败方全灭后发生的 `BOMB_PLANT`

#### Scenario: 超时终局
- **WHEN** 回合以 `timeout` 结束
- **THEN** 获胜阵营为 CT
- **AND** 双方最终均有存活选手
- **AND** 回合中没有发生 `BOMB_PLANT`
- **AND** `ROUND_END` 的时间戳等于回合时间上限

#### Scenario: 炸弹终局
- **WHEN** 回合中发生 `BOMB_PLANT`
- **THEN** 回合继续产生与胜方一致的 `BOMB_DEFUSE` 或 `BOMB_EXPLODE`
- **AND** `RoundResult.WinReason` 分别为 `bomb_defused` 或 `bomb_exploded`

### Requirement: 引擎聚合全回合选手统计

`MatchResult.FinalStats` SHALL 跨所有模拟回合聚合击杀、死亡、伤害、ADR、首杀、多杀、下包和拆包。

#### Scenario: ADR 除以总回合数
- **GIVEN** 某选手在 30 回合比赛中造成 2400 点伤害
- **WHEN** 构建最终统计
- **THEN** 该选手的 ADR 为 80

### Requirement: 第一版实现遵循设计分层但限制细节深度

第一版完整比赛实现 SHALL 遵循 `doc/simuMatchDesign.md` 的数据边界和主要模拟阶段，包括战术模板选择、角色分配、开局遭遇战、中期决策、包点遭遇战、炸弹阶段、可解释战报和比赛记忆。它 MAY 将详细 action 生命周期、连续边上移动、完整寻路和完整情报传播保留为简化或抽象的第一版逻辑。

#### Scenario: 主要阶段被表达
- **WHEN** 模拟一个回合
- **THEN** 生成的战报或调试解释标识选中的战术/模板以及关键战斗/炸弹结果
- **AND** 引擎使用地图/场景/配置输入参与评分和事件位置生成
- **AND** 本次变更不要求实现每条路径段上的连续移动

