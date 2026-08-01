## ADDED Requirements

### Requirement: 引擎模拟 MR12 常规赛

`matchengine` SHALL 支持完整 MR12 常规赛：每半场 12 回合，第 12 回合后换边，任一队伍在常规赛中先到 13 分时比赛结束。

#### Scenario: 队伍在 24 个常规回合前获胜
- **GIVEN** 一个有效 `MatchInput`，其中包含两支五人队伍、`RegulationHalfRounds = 12`、`RegulationWinRounds = 13`，且开启加时
- **WHEN** Team A 在 Team B 达到 12 分前先达到 13 分
- **THEN** `MatchResult.WinnerTeamID` 为 Team A
- **AND** `MatchResult.TotalRounds` 等于实际模拟的回合数
- **AND** 获胜回合之后不再模拟额外常规赛回合

#### Scenario: 上半场结束后换边
- **GIVEN** Team A 以 T 方开局，Team B 以 CT 方开局
- **WHEN** 模拟第 13 回合
- **THEN** `RoundResult.TeamTID` 为 Team B
- **AND** `RoundResult.TeamCTID` 为 Team A
- **AND** 队伍比分保留第 1 到第 12 回合累积结果

### Requirement: 引擎在 12-12 时进入加时

常规赛以 Team A 和 Team B 12-12 结束时，`matchengine` SHALL 进入加时。

#### Scenario: 常规赛平局开始加时
- **GIVEN** 第 24 回合结束后常规赛比分为 Team A 12、Team B 12
- **WHEN** `RuleSet` 中开启加时
- **THEN** 下一个模拟回合被标记为加时
- **AND** `RoundResult.Phase` 或等价元数据将该回合识别为加时
- **AND** 比赛不会在 12-12 时结束

### Requirement: 引擎按 MR3 block 结算加时

加时 SHALL 以完整 6 小局 MR3 block 模拟。OT1 前 3 局 SHALL 延续下半场阵营分配，OT1 后 3 局 SHALL 换边。后续每个加时 block 的前 3 局使用当前阵营，后 3 局换边。引擎 SHALL NOT 在 block 内提前结束加时；只有完整加时 block 结束后才判定比赛胜者。如果一个 block 结束后仍然平局，比赛 SHALL 继续进入下一个加时 block。

#### Scenario: 一个加时 block 后产生胜者
- **GIVEN** 常规赛以 12-12 结束
- **WHEN** Team B 赢下接下来 6 个加时回合中的 4 个
- **THEN** `MatchResult.WinnerTeamID` 为 Team B
- **AND** `MatchResult.FinalScoreTeamB` 大于 `MatchResult.FinalScoreTeamA`
- **AND** `MatchResult.Rounds` 中包含该加时 block 的全部 6 个回合
- **AND** 不会模拟第二个加时 block

#### Scenario: OT1 延续下半场阵营开局
- **GIVEN** Team A 上半场为 T，下半场为 CT，且常规赛以 12-12 结束
- **WHEN** 模拟 OT1 第 1 到第 3 个回合
- **THEN** Team A 继续作为 CT
- **AND** Team B 继续作为 T
- **WHEN** 模拟 OT1 第 4 到第 6 个回合
- **THEN** Team A 切换为 T
- **AND** Team B 切换为 CT

#### Scenario: 加时 block 不提前结束
- **GIVEN** 常规赛以 12-12 结束
- **WHEN** Team A 赢下第一个加时 block 的前 4 个回合
- **THEN** 该加时 block 的第 5 和第 6 回合仍然会被模拟
- **AND** 胜者只会在该 block 第 6 回合结束后最终确定

#### Scenario: 平局后进入额外加时 block
- **GIVEN** 常规赛以 12-12 结束
- **WHEN** 两队在第一个加时 block 中各赢 3 回合
- **THEN** 系统模拟第二个加时 block
- **AND** 回合编号保持严格递增

### Requirement: 比赛结果使用队伍比分作为权威比分

`MatchResult` 和 `RoundResult` SHALL 暴露队伍比分字段，使比分在换边和加时后仍能表达真实队伍身份。

#### Scenario: 换边后比分保持正确
- **GIVEN** Team A 以 T 方开局并赢下上半场 8 个回合
- **WHEN** Team A 在半场后以 CT 方继续比赛
- **THEN** Team A 的 `ScoreTeamA` 从 8 继续累加
- **AND** 如果为了兼容仍存在 `ScoreT` 和 `ScoreCT`，客户端不将其作为最终队伍身份比分

### Requirement: 整场比赛模拟按 seed 保持确定性

对于相同的输入快照、地图配置版本、规则集和 seed，`matchengine` SHALL 返回完全一致的回合、事件、阵营分配、比分和最终统计。

#### Scenario: 相同 seed 产生相同整场比赛
- **GIVEN** 两个完全相同且 `Seed` 相同的 `MatchInput`
- **WHEN** 分别调用 `Service.Simulate`
- **THEN** 两个 `MatchResult` 包含相同的 `WinnerTeamID`
- **AND** 两个结果包含相同的回合胜者序列
- **AND** 所有事件时间戳、参与者、武器和位置完全一致
