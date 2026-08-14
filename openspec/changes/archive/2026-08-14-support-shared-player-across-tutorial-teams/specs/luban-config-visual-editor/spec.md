## MODIFIED Requirements

### Requirement: 工作台提供 TutorialBattle 专属编辑页

工作台 SHALL 为 `#TutorialBattle.xlsx` 提供专属新手战斗配置页。该页面 SHALL 支持编辑启用状态、版本、预算、阵容人数、地图 ID、各费用档候选选手、对手战队和对手阵容。费用档候选池与固定对手阵容 SHALL 允许引用同一个 Player ID，但费用档之间和对手阵容内部的唯一性约束 SHALL 保持生效。

#### Scenario: 配置费用档候选选手

- **GIVEN** 用户打开 `新手战斗` 页面
- **AND** `#Player.xlsx` 已加载
- **WHEN** 用户在 `tier5PlayerIds`、`tier4PlayerIds`、`tier3PlayerIds`、`tier2PlayerIds` 或 `tier1PlayerIds` 中选择多个选手
- **THEN** 系统将选手 ID 写入对应列表字段
- **AND** 选择器显示选手名称、战队和稀有度等辅助信息
- **AND** 选手一旦被某个费用档选中，其他费用档中的该选手选项立即禁用并提示其已被占用；从原费用档取消后恢复可选
- **AND** 固定对手阵容中的同一选手仍可在一个费用档中选择，系统不跨双方禁用或报告重叠 ERROR
- **AND** 列表中出现不存在的选手 ID 时显示引用错误

#### Scenario: 阻止保存空费用档

- **GIVEN** 用户打开 `新手战斗` 页面
- **AND** `tier1PlayerIds`、`tier2PlayerIds`、`tier3PlayerIds`、`tier4PlayerIds` 或 `tier5PlayerIds` 中任意一个档位为空
- **WHEN** 用户点击 `保存当前表`
- **THEN** 系统显示该费用档缺少候选选手的校验错误
- **AND** 系统不写入 `#TutorialBattle.xlsx`

#### Scenario: 配置对手阵容

- **GIVEN** 用户在 `新手战斗` 页面选择了 `opponentTeamId`
- **WHEN** 用户编辑 `opponentPlayerIds`
- **THEN** 系统提供引用 `TbPlayer` 的多选控件
- **AND** 系统优先显示属于 `opponentTeamId` 的选手
- **AND** 候选池中已存在的同一选手仍可加入对手阵容
- **AND** 若已选选手不属于该战队，系统显示阻止保存的校验错误且不自动删除该选手

#### Scenario: 切换唯一启用的新手战斗配置

- **GIVEN** `#TutorialBattle.xlsx` 已有一条 `enabled=true` 的配置
- **WHEN** 用户尝试启用另一条配置
- **THEN** 系统提示启用后将关闭当前已启用配置
- **AND** 用户确认后系统将原配置改为 `enabled=false` 并启用新配置
- **AND** 用户取消时两条配置均保持原状态

#### Scenario: 阻止保存不满足运行时约束的启用配置

- **GIVEN** 用户正在编辑一条启用的 TutorialBattle 配置
- **WHEN** `budget` 不大于 0、`rosterSize` 不等于 5、选手跨费用档重复、候选池无法在预算内组成完整阵容、对手阵容人数不等于 `rosterSize`、对手选手重复或对手选手不属于 `opponentTeamId`
- **THEN** 系统显示与当前运行时校验一致的 ERROR
- **AND** 系统不写入 `#TutorialBattle.xlsx`

#### Scenario: 候选池与对手重叠仍可保存

- **GIVEN** 一条启用配置只有候选池与对手阵容交叉引用同一个有效 Player ID，其他运行时约束均满足
- **WHEN** 用户执行校验并保存当前表
- **THEN** 工作台不产生交叉重叠 ERROR
- **AND** 系统将配置写入 `#TutorialBattle.xlsx`
