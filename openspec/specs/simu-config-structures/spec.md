# simu-config-structures Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
### Requirement: Server reuses Luban TbPlayer config

`match.Service` SHALL 通过 `windypath.com/cs2match/config` 子模块读取 `TbPlayer` 选手属性，并显式构造为 `matchengine.Combatant` 传入引擎；`matchengine` 不直接依赖 `config` 包，只消费自包含的 `MatchInput`。

#### Scenario: Service loads player attributes from config

- **GIVEN** `TbPlayer` 已随 `cfg.Init()` 加载到内存
- **WHEN** `match.Service` 构造引擎队伍时传入选手 ID
- **THEN** `match.Service` 从 `cfg.Global.TbPlayer.Get(id)` 读取 `Entry`、`Aim`、`Firepower` 等属性
- **AND** 将其填充到 `matchengine.Combatant` 后作为 `MatchInput` 的一部分传给引擎
- **AND** 若选手 ID 不存在，返回 `INVALID_LINEUP` 错误

### Requirement: Player config structure is shared between server and client

`TbPlayer` 导出的字段 SHALL 与 `client/src/config/schema.ts` 中自动生成的 TypeScript 类型保持一致，确保前后端对选手属性的理解相同。

#### Scenario: Client displays real player names

- **GIVEN** 服务端返回的战报中包含 `attacker_name: "NiKo"`
- **WHEN** 前端用 `attacker_id` 查询本地 Luban 配置
- **THEN** 名称与 `TbPlayer` 中的 `Name` 字段一致

### Requirement: Route names are consistent between server and client

服务端返回的路线 ID 与显示名称 SHALL 与前端的临时路线常量一致，避免展示歧义。

#### Scenario: Client renders route name

- **GIVEN** 战报中 `RoundResult.RouteMain` 为 `"A_Long"`
- **WHEN** 前端展示回合开始事件
- **THEN** 界面上显示“A大”或“A_Long”文本

### Requirement: Combat constants are centralized

MVP 阶段虽然采用简化战斗，但回合时间上限、一回合固定长度等常量 SHALL 集中在 `matchengine` 的常量文件中，便于后续替换为 Luban 配表。

#### Scenario: Constants are not scattered in engine logic

- **WHEN** 查看 `internal/framework/matchengine` 源码
- **THEN** 存在 `const.go` 或 `model.go` 中统一定义的常量（如 `RoundTimeLimit`、`DefaultBombExplodeTime`）
- **AND** 引擎逻辑中不出现未命名的魔法数字

### Requirement: Config loader exposes player lookup helper

`server/config` 包 SHALL 保持并暴露 `GetPlayer(id string) *cfg.Player` 辅助函数，供 `match.Service` 安全读取选手数据。

#### Scenario: Service retrieves player by ID

- **GIVEN** 服务端硬编码阵容中包含 `"player_niko"`
- **WHEN** `match.Service` 调用 `cfg.GetPlayer("player_niko")`
- **THEN** 返回非 nil 的 `*cfg.Player`
- **AND** 其 `Name` 字段为 `"NiKo"`

### Requirement: 战报 Console 调试开关来自 CombatConst 配表

`configs/Datas/#CombatConst.xlsx` SHALL 包含 Bool 类型的 `BattleReportDebugLog`。该值 SHALL 通过 Luban 导出，不得手工修改生成 JSON；`server/internal/match` SHALL 读取该值并映射到 RPC 响应，`matchengine` SHALL 不读取该业务调试开关。

#### Scenario: 开启战报调试日志
- **GIVEN** `BattleReportDebugLog` 的值为 `true`
- **WHEN** `DebugSimuMatch` 返回完整比赛战报
- **THEN** 响应顶层 `debug_enabled` 为 `true`
- **AND** 框架的 `MatchInput` 和 `MatchResult` 不增加业务调试字段

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

### Requirement: 选手卡面和头像裁切数据随比赛快照传递

服务端 SHALL 从 Luban Player 配置读取完整卡面及头像裁切数据，并经领域模型和比赛快照传递给客户端。新增字段 SHALL 为可选字段，旧 `portrait` 数据 SHALL 继续可用。

#### Scenario: Luban 为服务端和客户端生成一致字段

- **GIVEN** `#Player.xlsx` 包含 `cardImage`、`avatarCropX`、`avatarCropY`、`avatarCropWidth`、`avatarCropHeight`
- **WHEN** 项目运行 Luban 导表
- **THEN** Server Go 和 Client TypeScript 生成结构包含语义一致的字段
- **AND** 生成的 JSON 数据保留相同的资源路径和归一化数值

#### Scenario: Match 服务构建选手视觉资料

- **GIVEN** 一条 Player 配置包含合法的完整卡面和裁切参数
- **WHEN** Match 服务从 `TbPlayer` 构建 `PlayerProfile`
- **THEN** `PlayerProfile` 包含 `CardImage`
- **AND** 四个扁平配置字段被组装为一个可选的 `AvatarCrop` 值对象
- **AND** 回合投影将 `CardImage` 和 `AvatarCrop` 复制到客户端可见的 `PlayerState`

#### Scenario: 非法或缺失新字段使用旧头像

- **GIVEN** Player 未配置完整卡面，或裁切矩形不合法
- **WHEN** Match 服务构建比赛快照
- **THEN** 服务端不输出可用的 `AvatarCrop`
- **AND** 保留现有 `Portrait` 路径供客户端回退
- **AND** 比赛模拟仍可正常创建和推进

#### Scenario: 视觉配置不改变模拟过程

- **GIVEN** 两份 Player 配置只有卡面和头像裁切参数不同
- **WHEN** MatchLoop 使用相同种子和相同比赛输入运行
- **THEN** 两场模拟产生相同的比赛状态与事件结果
- **AND** MatchLoop 不直接读取静态配置文件或图片资源

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

