## MODIFIED Requirements

### Requirement: 教学模式按 Luban 配置校验双方阵容
`SimuMatch` 的 tutorial 模式 SHALL 从 `TbTutorialBattle` 读取预算、人数、价格池、地图和机器人策划 Player ID，并从 `TbPlayer` 读取双方十个阵容位置对应的选手属性。服务端 SHALL 校验玩家恰好五人、玩家阵容内部 ID 唯一、属于配置池、总价不超预算，并校验对手恰好五人、对手阵容内部 ID 唯一且所有引用有效。

费用档候选池与固定对手阵容 MAY 包含相同的策划 Player ID。服务端 SHALL NOT 因双方集合相交而返回 `INVALID_LINEUP`；服务端 SHALL 为双方每个阵容位置生成全场唯一的比赛实例 `player_id`，并保留原始 `config_player_id` 后再进入模拟引擎。

#### Scenario: 合法教学请求进入引擎
- **GIVEN** 请求匹配启用且版本一致的教学配置
- **WHEN** 服务端完成双方阵容校验
- **THEN** `internal/match` 构造包含真实双方 `TeamInput` 的 `MatchInput`
- **AND** 调用现有 `matchengine.Service.Simulate`

#### Scenario: 双方引用同一策划选手
- **GIVEN** 玩家合法阵容和固定对手阵容都包含 `player_zywoo`，且各自阵容内部没有重复 ID
- **WHEN** 服务端构建教学比赛输入
- **THEN** 请求通过阵容校验并进入模拟引擎
- **AND** 两个 `PlayerProfile` 的 `config_player_id` 都是 `player_zywoo`
- **AND** 两个 `PlayerProfile` 获得不同且全场唯一的 `player_id`

#### Scenario: 过期配置版本
- **GIVEN** 客户端提交的 config version 与当前服务端配置不一致
- **WHEN** 服务端校验请求
- **THEN** 返回 `CONFIG_VERSION_MISMATCH`
- **AND** 不进入模拟引擎

#### Scenario: 超预算或越池选手
- **GIVEN** 玩家阵容总价超过配置预算或包含未列入任一价格档的选手
- **WHEN** 服务端校验阵容
- **THEN** 返回 `INVALID_TUTORIAL_LINEUP`
- **AND** 不生成战报

#### Scenario: 单侧阵容仍禁止重复
- **GIVEN** 玩家阵容或固定对手阵容在自身五个位置中重复同一策划 Player ID
- **WHEN** 服务端校验阵容
- **THEN** 返回结构化非法阵容错误
- **AND** 不进入模拟引擎

### Requirement: SimuMatch 返回现有标准 MatchReport
`SimuMatch` SHALL 返回可由当前版本模拟回放消费的完整 `MatchReport`，并使双方队名、比赛实例 Player ID、策划 Player ID、选手属性、比分、事件和胜者来自本次服务端 `MatchInput` 与模拟结果。战报中的 `player_id` SHALL 表示整场唯一的比赛实例 ID；公开选手状态和最终统计 SHALL 同时包含对应 `TbPlayer.id` 的 `config_player_id` 与所属 `team_id`。所有 Actor 事件和炸弹携带者字段 SHALL 使用比赛实例 ID。

#### Scenario: 前端播放教学战
- **GIVEN** `SimuMatch` 成功模拟教学双方真实阵容
- **WHEN** 客户端接收响应
- **THEN** 响应可由更新后的 `MatchReport` TypeScript 类型解析
- **AND** BattlePage 无需客户端重新计算比赛结果

#### Scenario: 战报区分同源选手实例
- **GIVEN** 双方阵容都引用同一个 `TbPlayer.id`
- **WHEN** 服务端返回玩家状态、Actor 事件和最终统计
- **THEN** 两个单位拥有不同的 `player_id` 和各自的 `team_id`
- **AND** 两个单位拥有相同的 `config_player_id` 与对应策划资料快照
- **AND** 事件、炸弹状态和统计均可按各自 `player_id` 无歧义关联
