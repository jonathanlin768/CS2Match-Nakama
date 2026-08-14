## ADDED Requirements

### Requirement: 引擎以比赛实例 ID 唯一关联战斗单位

`matchengine` SHALL 把 `PlayerProfile.player_id` 作为整场比赛内唯一且不透明的战斗实例主键。回合玩家状态、角色与路径分配、炸弹携带者、击杀/伤害/助攻事件、最终统计、稳定排序以及确定性随机派生 SHALL 使用该比赛实例 ID；`config_player_id` SHALL 只作为策划来源元数据，不得用作可变战斗状态的 map 键。

#### Scenario: 同源的双方选手状态相互独立
- **GIVEN** 两个 `PlayerProfile` 拥有相同 `config_player_id` 和不同 `player_id`
- **WHEN** 引擎创建并推进回合状态
- **THEN** `RoundState` 保存两个独立玩家条目
- **AND** 其中一个实例的 HP、位置、装备或存活变化不修改另一个实例

#### Scenario: Actor 输出使用比赛实例 ID
- **GIVEN** 两支队伍包含相同 `config_player_id` 的不同比赛实例
- **WHEN** 引擎生成击杀、伤害、助攻或炸弹相关事件并聚合最终统计
- **THEN** 攻击者、受害者、助攻者、携带者和统计条目均使用对应比赛实例 `player_id`
- **AND** 任何结果都不因相同 `config_player_id` 覆盖另一侧实例

#### Scenario: 新身份模型保持确定性
- **GIVEN** 两次模拟具有完全相同的双方比赛实例 ID、策划快照、地图、规则和 seed
- **WHEN** 分别执行两次 `Simulate`
- **THEN** 两次模拟生成相同顺序的事件、比分和最终统计

#### Scenario: 引擎拒绝重复比赛实例 ID
- **GIVEN** 两个 `PlayerProfile` 使用相同 `player_id`，无论其 `config_player_id` 是否相同
- **WHEN** 引擎初始化比赛或回合
- **THEN** 引擎返回非法阵容错误
- **AND** 不开始模拟
