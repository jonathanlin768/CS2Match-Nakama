## ADDED Requirements

### Requirement: 比赛快照分离实例身份与策划身份

Match 服务构建的 `PlayerProfile` SHALL 同时包含比赛内唯一的 `player_id` 和对应 `TbPlayer.id` 的 `config_player_id`。公开 `PlayerState` 与最终 `PlayerMatchStats` SHALL 复制这两个字段，并通过 `team_id` 表达本场队伍归属；能力值、角色标签、卡面和头像裁切 SHALL 来自 `config_player_id` 对应的同一份策划配置快照。

#### Scenario: 同一策划选手生成两个比赛实例
- **GIVEN** 两支队伍都引用 `player_zywoo`
- **WHEN** Match 服务构建双方 `TeamInput`
- **THEN** 两份 `PlayerProfile` 的 `config_player_id` 都为 `player_zywoo`
- **AND** 两份 `PlayerProfile` 的 `player_id` 按各自 `team_id` 生成且互不相同
- **AND** 两个实例读取相同的策划能力值与视觉配置

#### Scenario: 身份字段传递到公开投影
- **GIVEN** 引擎收到同时包含 `player_id` 和 `config_player_id` 的 `PlayerProfile`
- **WHEN** 引擎生成回合 `PlayerState` 与最终 `PlayerMatchStats`
- **THEN** 输出保留同一个比赛实例 `player_id`
- **AND** 输出保留对应的 `config_player_id` 和 `team_id`

### Requirement: 教学配置允许双方共享策划选手

Luban 导出校验和服务端配置初始化 SHALL 允许某个 Player ID 同时存在于一个教学费用档与固定对手阵容。系统 SHALL 继续拒绝同一个 Player ID 跨多个费用档出现、固定对手阵容内部重复以及无效 Player/Team 引用。

#### Scenario: 有效的跨双方重叠配置
- **GIVEN** `player_zywoo` 同时存在于一个费用档和固定对手阵容，且两侧各自的唯一性、人数、预算与引用约束均满足
- **WHEN** 执行 Luban 导出或服务端配置测试
- **THEN** 配置通过校验
- **AND** 导出结构继续保存两处相同的 `TbPlayer.id` 引用

#### Scenario: 费用档内部规则保持不变
- **GIVEN** 同一 Player ID 同时出现在两个价格档
- **WHEN** 配置测试执行
- **THEN** 测试失败并标识重复 Player ID

#### Scenario: 对手阵容内部规则保持不变
- **GIVEN** 固定对手五个位置内重复同一 Player ID
- **WHEN** 配置测试执行
- **THEN** 测试失败并标识对手阵容内部重复
