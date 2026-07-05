## ADDED Requirements

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

### Requirement: Engine simulates exactly one round in MVP

MVP 阶段 `MatchEngine` SHALL 只推演一回合，回合结束后立即生成结果。

#### Scenario: Single round output

- **WHEN** `Simulate` 返回成功
- **THEN** `MatchResult.Rounds` 长度为 1
- **AND** `MatchResult.TotalRounds` 为 1

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
