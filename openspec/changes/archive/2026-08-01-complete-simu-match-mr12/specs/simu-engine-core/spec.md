## ADDED Requirements

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

## REMOVED Requirements

### Requirement: Engine simulates exactly one round in MVP

**Reason**: 模拟不再是单回合 MVP；它必须产出完整 MR12/加时比赛战报。

**Migration**: 如单元测试需要单回合行为，可通过 `RuleSet` 和测试专用输入请求。生产 `DebugSimuMatch` 使用默认完整比赛规则集。
