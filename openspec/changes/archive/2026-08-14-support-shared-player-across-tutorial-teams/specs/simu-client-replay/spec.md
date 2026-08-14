## ADDED Requirements

### Requirement: 回放按比赛实例身份关联状态与策划资料

客户端 SHALL 把服务端战报中的 `player_id` 视为比赛实例 ID，并使用它关联玩家状态、事件、生命值、存活状态、炸弹携带者和最终 KDA。客户端 SHALL 使用 `config_player_id` 查询 `TbPlayer` 策划资料或视觉资源，不得按 `config_player_id` 合并两个比赛实例。服务器战报是战斗状态的唯一权威来源，客户端表现层 SHALL NOT 根据策划身份重建或合并状态。

#### Scenario: 双方同源选手分别回放
- **GIVEN** 战报双方各有一个 `config_player_id=player_zywoo` 且 `player_id` 不同的选手
- **WHEN** BattlePage 逐事件回放比赛
- **THEN** 两个选手分别显示各自的生命值、存活状态、事件和 KDA
- **AND** 一侧的击杀或死亡不会更新另一侧选手状态
- **AND** 两侧均可显示同一份 ZyWOo 卡面、头像裁切和名称资料

#### Scenario: Actor 事件定位到正确实例
- **GIVEN** 击杀事件的攻击者和受害者字段包含比赛实例 ID
- **WHEN** 客户端应用该事件并更新可见状态
- **THEN** 客户端只更新与这些 `player_id` 匹配的实例
- **AND** 客户端不使用 `config_player_id` 猜测事件归属

#### Scenario: TypeScript 类型表达双重身份
- **GIVEN** 前端代码导入更新后的 `PlayerProfile`、`PlayerState` 和 `PlayerMatchStats` 类型
- **WHEN** 在 `client` 中执行 TypeScript 类型检查
- **THEN** 类型明确包含比赛实例 `player_id` 与策划 `config_player_id`
- **AND** 回放状态键接受完整的比赛实例 ID 字符串
