## ADDED Requirements

### Requirement: Luban 提供独立队伍表并由选手引用
系统 SHALL 新增自动导入的 `#Team.xlsx`/`TbTeam`，至少配置队伍 ID、正式名称、简称、昵称和 Logo。`#Player.xlsx` SHALL 使用引用 `TbTeam` 的 `teamId` 替换自由文本 team 字段；未来客户端图鉴和当前服务端比赛阵容构建 SHALL 按 teamId 关联选手。本变更 SHALL NOT 在 Team 表预置首发领取或阵容调整规则。

#### Scenario: 按队伍查询选手
- **GIVEN** `TbTeam` 存在某个 team ID 且多名 `TbPlayer.teamId` 引用它
- **WHEN** 图鉴按该队伍筛选
- **THEN** 客户端返回所有引用该 teamId 的选手
- **AND** 不依赖队伍展示名称进行字符串匹配

#### Scenario: Player 引用不存在队伍
- **GIVEN** `#Player.xlsx` 某行的 teamId 不存在于 `TbTeam`
- **WHEN** 执行 Luban 导出或配置测试
- **THEN** 导出/测试失败并指出非法引用

### Requirement: Luban 提供版本化教学战表
系统 SHALL 新增自动导入的 `#TutorialBattle.xlsx`/`TbTutorialBattle`。每个启用方案 SHALL 配置 ID、版本、预算、阵容人数、地图、5/4/3/2/1 元 Player ID 列表、机器人 team ID 和机器人五个 Player ID；Player 与 Team 字段 SHALL 使用表引用约束。

#### Scenario: 加载有效教学方案
- **GIVEN** 教学表包含一个启用且引用完整的方案
- **WHEN** 服务端配置初始化
- **THEN** 服务端缓存各价格档、预算、地图和机器人阵容
- **AND** 客户端可加载相同导出数据用于展示

#### Scenario: 教学档位重复选手
- **GIVEN** 同一 Player ID 同时出现在两个价格档
- **WHEN** 配置测试执行
- **THEN** 测试失败并标识重复 Player ID

#### Scenario: 教学方案无预算可解阵容
- **GIVEN** 配置的价格池无法选择指定人数且不超过预算
- **WHEN** 配置测试执行
- **THEN** 测试失败且该方案不可启用

### Requirement: Team 与 TutorialBattle 同步导出到前后端
Luban 导表 SHALL 生成 Go、TypeScript 与客户端/服务端 JSON 产物，并更新 `server/config/loader.go` 与 `client/src/config/index.ts` 以加载 `TbTeam` 和 `TbTutorialBattle`。生成文件不得手工维护。

#### Scenario: 导表后验证
- **GIVEN** Team、Player 和 TutorialBattle Excel 已保存
- **WHEN** 执行 `scripts/gen-config.ps1`
- **THEN** 前后端生成 schema 与 JSON 包含新表及 Player.teamId
- **AND** `server/config` Go 测试和客户端 TypeScript 检查通过

## REMOVED Requirements

### Requirement: Match 服务从 TbPlayer 构建默认队伍
**Reason**: 旧要求按 `TbPlayer.team` 自由文本筛选并硬编码 Falcons/Vitality，无法支持正式 Team 图鉴、可配置首发队和教学对手。
**Migration**: 调试入口从 `TbTeam` 与 Player.teamId 构建默认队；生产 `SimuMatch` 从玩家阵容及 `TbTutorialBattle` 对手配置构建双方队伍。
