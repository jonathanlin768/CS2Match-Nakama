## ADDED Requirements

### Requirement: 战报 Console 调试开关来自 CombatConst 配表

`configs/Datas/#CombatConst.xlsx` SHALL 包含 Bool 类型的 `BattleReportDebugLog`。该值 SHALL 通过 Luban 导出，不得手工修改生成 JSON；`server/internal/match` SHALL 读取该值并映射到 RPC 响应，`matchengine` SHALL 不读取该业务调试开关。

#### Scenario: 开启战报调试日志
- **GIVEN** `BattleReportDebugLog` 的值为 `true`
- **WHEN** `DebugSimuMatch` 返回完整比赛战报
- **THEN** 响应顶层 `debug_enabled` 为 `true`
- **AND** 框架的 `MatchInput` 和 `MatchResult` 不增加业务调试字段

### Requirement: Match 服务从 TbPlayer 构建默认队伍

`match.Service` SHALL 提供默认调试阵容：直接读取 `configs/Datas/#Player.xlsx` 经 Luban 导出的 `TbPlayer`，按 `team` 字段筛选，`Falcons` 的 5 名选手组成 Team A，`Vitality` 的 5 名选手组成 Team B，队内顺序沿用 `GetDataList()` 的稳定表顺序，同时将生成选手表访问保留在 `matchengine` 之外。任一队伍不是恰好 5 人或选手 ID 重复时 SHALL 返回 `INVALID_LINEUP`。`tbplayer.json` SHALL 仅由 Luban 导出生成，不得为默认阵容手工改写。

#### Scenario: 默认队伍包含十名唯一选手
- **WHEN** 调用不带阵容字段的 `DebugSimuMatch`
- **THEN** `match.Service` 构建 Team A 和 Team B，且每队 5 名选手
- **AND** 十个 `PlayerID` 全部唯一
- **AND** Team A 名称为 `Falcons`，成员来自 `team=Falcons` 的 5 条记录
- **AND** Team B 名称为 `Vitality`，成员来自 `team=Vitality` 的 5 条记录
- **AND** 两队内部的 `PlayerID` 顺序分别保持 `TbPlayer.GetDataList()` 的稳定表顺序
- **AND** 每个默认选手由 `match.Service` 从 `windypath.com/cs2match/config` 的 `TbPlayer` 读取
- **AND** `matchengine` 仅通过 `MatchInput` 接收这些选手，且不读取 `TbPlayer`

#### Scenario: 选手头像路径随快照传递
- **WHEN** `match.Service` 将 `TbPlayer` 选手转换为 `PlayerProfile`
- **THEN** `TbPlayer.Portrait` 原样写入 `PlayerProfile.Portrait`
- **AND** 回合 `PlayerState` 将该头像路径返回给客户端
- **AND** 客户端将 `portraits/player_niko.jpg` 解析为站点根路径 `/portraits/player_niko.jpg`
- **AND** 图片不存在或加载失败时回退到 `/images/star-player.png`

### Requirement: Match 服务分配第一版固定武器

在这个无经济系统版本中，`match.Service` SHALL 提供按阵营默认的武器装配规则：当前作为 T 方的选手使用 AK-47，当前作为 CT 方的选手使用 M4A1-S。`PlayerProfile` SHALL 只表达选手基础属性、角色标签和头像等静态档案信息，不绑定固定枪械；具体枪械 SHALL 在回合派生输入或回合内选手状态中按当前阵营动态确定。

#### Scenario: 装配跟随当前阵营
- **GIVEN** Team A 以 T 方开局，Team B 以 CT 方开局
- **WHEN** 构建第 1 回合输入
- **THEN** Team A 选手档案以 AK-47 作为主武器
- **AND** Team B 选手档案以 M4A1-S 作为主武器

#### Scenario: PlayerProfile 不绑定固定枪械
- **WHEN** `match.Service` 将 `TbPlayer` 转换为 `matchengine.PlayerProfile`
- **THEN** `PlayerProfile` 包含选手基础属性、角色标签和头像路径
- **AND** `PlayerProfile` 不包含会在换边后失效的固定主武器绑定
- **AND** 当前主武器由回合阵营装配规则派生

#### Scenario: 装配随换边切换
- **GIVEN** Team A 以 T 方开局，Team B 以 CT 方开局
- **WHEN** 半场换边后派生第 13 回合输入
- **THEN** Team A 选手状态在 CT 方战斗中使用 M4A1-S 作为主武器
- **AND** Team B 选手状态在 T 方战斗中使用 AK-47 作为主武器

### Requirement: 武器规格是显式输入数据

武器数值 SHALL 表达为显式的 `WeaponSpec` 或等价输入数据，字段包含显示名称、伤害、射速、弹匣容量、穿甲和距离修正。

#### Scenario: AK-47 和 M4A1-S 数值可用
- **WHEN** `match.Service` 构造默认比赛输入
- **THEN** 武器规格快照包含 AK-47 条目
- **AND** 武器规格快照包含 M4A1-S 条目
- **AND** 遭遇战逻辑可以读取武器数值，而不依赖硬编码武器常量

#### Scenario: 武器表是后续工作
- **WHEN** 实现本次变更
- **THEN** 不要求新增 Luban `tb_weapon` 表
- **AND** 代码结构保持 `WeaponSpec` 独立，使后续提案可以将武器数值迁入配置

### Requirement: 地图语义表输入引擎快照

`match.Service` 或相邻适配器 SHALL 在调用 `Simulate` 前，将 Luban 生成的地图语义表转换为 `matchengine.MapConfig`。

#### Scenario: 适配器转换生成路线模板
- **GIVEN** 生成配置包含 `de_dust2` 的 `TbRouteTemplate` 行
- **WHEN** 适配器构建 `matchengine.MapConfig`
- **THEN** 每个生成路线模板行都以 ID、目标包点、节奏、所需角色、关键属性、场景、地图标签、成功阶段和失败 fallback 形式出现在引擎快照中

## REMOVED Requirements

### Requirement: Route configuration is defined in engine code for MVP

**Reason**: 路线模板、路线、节点、路径、视野和风险热点现在必须来自 Luban 支持的地图语义配置，而不是引擎硬编码。

**Migration**: 将 Dust2 路线和坐标内容迁入 `tools/map-semantic-editor` 生成的 Luban 数据，然后将生成配置转换为 `matchengine.MapConfig`。
