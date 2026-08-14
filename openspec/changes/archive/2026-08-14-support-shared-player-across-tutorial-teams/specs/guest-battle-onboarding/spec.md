## MODIFIED Requirements

### Requirement: 教学双方真实阵容进入模拟引擎
开始教学战时客户端 SHALL 只提交 tutorial config ID、config version 与已选五个策划 Player ID。服务端 SHALL 重新校验玩家阵容，并从 `TbTutorialBattle` 读取机器人五人阵容和地图，再从 `TbPlayer` 构造双方完整 `TeamInput` 传入 `matchengine`。双方 MAY 引用同一个策划 Player ID；服务端 SHALL 为每个阵容位置生成按队伍区分且全场唯一的比赛实例 Player ID。

#### Scenario: 合法教学阵容开始真实模拟
- **GIVEN** 玩家五人阵容满足当前教学配置且机器人配置包含五名有效选手
- **WHEN** 客户端调用 `SimuMatch` 的 tutorial 模式
- **THEN** 服务端用双方十个阵容位置对应的真实 `TbPlayer` 属性构造 `MatchInput`
- **AND** 返回战报中的 `config_player_id` 与提交/配置的策划 Player ID 一致
- **AND** 返回战报中的 `player_id` 在整场比赛内唯一

#### Scenario: 玩家选择与对手相同的选手
- **GIVEN** 玩家合法选择了固定对手阵容中也存在的一名选手
- **WHEN** 客户端调用 `SimuMatch` 的 tutorial 模式
- **THEN** 服务端接受该阵容并生成两个独立的比赛实例
- **AND** 客户端回放将两个实例显示在各自队伍中

#### Scenario: 客户端伪造价格或属性
- **GIVEN** 修改客户端提交了价格、属性或不属于当前档位的 Player ID
- **WHEN** 服务端处理教学模拟请求
- **THEN** 服务端忽略客户端价格/属性并按 Luban 配置重新校验
- **AND** 非法 Player ID 使请求失败且不进入 `matchengine`
