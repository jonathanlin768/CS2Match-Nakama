## ADDED Requirements

### Requirement: 战斗阵容使用策划配置的 5:7 头像裁切

BattlePage 的队伍阵容 SHALL 优先使用比赛快照中的完整卡面和裁切参数，在固定 `5:7` 视口中动态显示选手头像。裁切属于客户端展示，不得修改回放状态或比赛结果。

#### Scenario: 在战斗选手卡中还原头像裁切

- **GIVEN** `PlayerState` 包含可读取的 `cardImage` 和合法的 `avatarCrop`
- **WHEN** TeamRoster 渲染 PlayerCard
- **THEN** PlayerCard 使用完整卡面作为源图
- **AND** 依据源图自然尺寸及归一化矩形在 `5:7` 视口内准确显示裁切区域
- **AND** 不依赖预先生成的第二张头像图片

#### Scenario: 回放状态变化不改变裁切区域

- **GIVEN** 战斗回放正在逐事件更新选手状态
- **WHEN** 选手生命值、存活状态或装备发生变化
- **THEN** 头像继续使用快照中同一组裁切参数
- **AND** 死亡灰化等状态样式叠加在裁切结果之上

#### Scenario: 卡面或裁切不可用时逐级回退

- **GIVEN** `cardImage` 加载失败或 `avatarCrop` 不合法
- **WHEN** PlayerCard 渲染头像
- **THEN** 客户端首先回退到现有 `portrait`
- **AND** `portrait` 也不可用时回退到站点默认选手图
- **AND** 阵容布局和回放流程不因图片错误中断
