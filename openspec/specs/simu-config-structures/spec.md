# simu-config-structures Specification

## Purpose

定义比赛输入、Luban 配置、武器与地图语义、选手身份和因果模拟参数的数据契约。

本规格包含已完成因果引擎变更的设计要求；当前实现缺口见 `server/internal/framework/matchengine/README.md`，规格同步不表示接线缺口已修复。

## Requirements

### Requirement: Server reuses Luban TbPlayer config

`match.Service` SHALL 通过 `windypath.com/cs2match/config` 读取 `TbPlayer`，将每名选手完整映射为自包含的 `matchengine.PlayerProfile` 后放入 `MatchInput`；正式引擎输入 SHALL NOT 依赖旧的三属性 `Combatant` 结算模型。`matchengine` 不直接依赖 `config` 包，也不得在模拟期间回查或修改 `TbPlayer`。

#### Scenario: Service maps the complete player profile

- **GIVEN** `TbPlayer` 已随 cfg.Init 加载且包含有效选手
- **WHEN** match.Service 构造 MatchInput
- **THEN** PlayerProfile 包含稳定 ID、名称、角色标签、头像以及 Aim、Reaction、Positioning、Awareness、Teamplay、Utility、Composure、Mobility、Endurance、Discipline 十项基础属性
- **AND** 当前阵营武器装配在回合输入/状态中独立派生，不写回 PlayerProfile
- **AND** 选手不存在或必需属性非法时返回 INVALID_LINEUP

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

正式模拟使用的回合时间、炸弹时间、资源边界、概率 clamp、战斗尺度、场景属性权重、决策阈值和标定参数 SHALL 统一来自 `MapConfig.CombatConstants` 及关联的配置快照。`matchengine` 可以定义稳定键名和纯公式顺序，但 SHALL NOT 在 resolver 中散落未命名数值或以代码默认值掩盖正式输入缺失。

#### Scenario: Constants come from the validated snapshot

- **GIVEN** `server/internal/match` 已从 Luban 生成表构建有效 `MapConfig`
- **WHEN** Encounter、Decision 或 Bomb resolver 读取平衡参数
- **THEN** 参数通过类型化 CombatConstants/EncounterModifier 访问层获得
- **AND** 必配键缺失时配置校验失败
- **AND** `matchengine` 不直接导入 `windypath.com/cs2match/config`

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
- **THEN** Team A 选手的回合装配以 AK-47 作为主武器
- **AND** Team B 选手的回合装配以 M4A1-S 作为主武器

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

武器数值 SHALL 表达为显式 `WeaponSpec` 输入快照，至少包含显示名称、伤害、射速、弹匣容量、穿甲和距离修正。Encounter 的伤害潜力、距离表现或单位时间火力 SHALL 实际读取这些字段；仅在战报中显示武器而不参与结算不满足本要求。

#### Scenario: AK-47 和 M4A1-S 数值参与伤害结算

- **WHEN** `match.Service` 构造默认比赛输入并模拟一次 Encounter
- **THEN** 武器规格快照包含 AK-47 和 M4A1-S
- **AND** 攻击者当前阵营装配决定其使用的 WeaponSpec
- **AND** Damage、RoundsPerMinute、MagazineSize、ArmorPenetration 和 RangeModifier 中与当前脉冲有关的字段进入可测试的伤害潜力或射击容量计算
- **AND** `EventReason` 或调试计算记录可证明所使用的武器修正

#### Scenario: 武器表仍是后续工作

- **WHEN** 实现本次变更
- **THEN** 不要求新增 Luban `tb_weapon` 表
- **AND** 武器规格继续由调用方作为自包含输入快照提供
- **AND** `matchengine` 不使用硬编码 AK-47/M4A1-S 数值覆盖该快照

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

### Requirement: 因果战斗配置覆盖公式和状态边界

`configs/Datas/#CombatConst.xlsx`、Scenario、MapTag 和 EncounterModifier SHALL 为因果模拟提供场景属性权重、姿态/协同/资源修正、HitChance/KillChance clamp、伤害与时间尺度、UtilityBudget、CloseScoreGap/DecisiveScoreGap、决策阈值及状态上限。所有必配值 SHALL 经 Luban 导出并由 `server/internal/match` 映射为自包含引擎快照；不得手工修改生成 JSON。配置与实现 SHALL 使用 `simuMatchDesign.md` 冻结的唯一 PlayerCombatScore/TargetSurvivalScore/EncounterScore 分层，Utility/Momentum/TimePressure 不得跨层重复计分；`KillChance` SHALL 使用主设计冻结公式且不配置或读取 `TargetHPFactor`。

#### Scenario: 配置快照包含 Encounter 必需参数

- **GIVEN** Luban 已成功导出 de_dust2 的战斗配置
- **WHEN** `server/internal/match` 构建 `matchengine.MapConfig`
- **THEN** Encounter 评分、战斗脉冲、资源消耗、中期决策和炸弹结算所需的全部必配键均可通过类型化访问器读取
- **AND** 场景属性权重及相关修正可追溯到 CombatConst 或 EncounterModifier
- **AND** 引擎公式中不存在替代这些配置的隐藏平衡常量

#### Scenario: 缺失或非法因果参数拒绝模拟

- **GIVEN** 某个场景属性权重和不为 1、概率最小值大于最大值或必要时间/资源参数缺失
- **WHEN** 服务构建或校验地图配置快照
- **THEN** 返回稳定的结构化配置错误
- **AND** 引擎不使用隐藏默认值继续运行该地图的正式模拟

### Requirement: 配置不得表达目标胜方或目标终局

战斗与标定配置 SHALL 只描述能力、权重、概率边界、资源、耗时、场景和决策倾向，SHALL NOT 提供按回合指定胜方、指定终局、目标存活人数或根据预定胜方调整事件的生产参数。

#### Scenario: 标定参数不绕过模拟

- **WHEN** 策划调整 T/CT 基准胜率、下包率或拆包率
- **THEN** 调整通过 Scenario、EncounterModifier、CombatConst、路线或资源参数影响实际阶段结算
- **AND** 不新增 `ForcedWinner`、`TargetWinner`、`TargetSurvivors` 或等价配置项

### Requirement: Dust2 配置覆盖第一版主战术空间

正式 `de_dust2` 配置 SHALL NOT 只包含单条 A 大占位路线。Luban 表源 SHALL 使 RouteTemplate 同时表达 `side=T` 战术模板与 `side=CT` 开局 setup 模板，并覆盖 `simuMatchDesign.md` 第一版列出的六个 T 战术模板、多个合法 CT setup、双方相应路线、CT 补防/回防路线、Opening/Mid/Site/Retake/Bomb 场景族、核心节点、必要 MapEdge/Visibility/风险热点、MapTag 和 EncounterModifier。每个模板 SHALL 通过 `route_ids/route_allocations` 冻结同阵营路线与人数约束；T 模板的 `common_ct_setup_ids` 只表达赛前先验，不指定本回合实际 CT setup。

#### Scenario: 六个 T 战术模板全部可选

- **WHEN** 适配器构建正式 de_dust2 MapConfig
- **THEN** RouteTemplates 至少包含 A_Long_Rush、A_Short_Split、B_Tunnel_Explode、Mid_To_B、Default_Pick 和 Fake_A_Go_B 的稳定配置 ID
- **AND** 每个模板的 side 为 T，并引用存在的 Scenario、MapTag、T Route/目标包点和合法人数/角色约束

#### Scenario: CT setup 模板独立且完整

- **WHEN** 校验正式 de_dust2 RouteTemplates
- **THEN** 至少存在多个 `side=CT` setup 模板，能够表达 A/B/Mid 初始覆盖
- **AND** 每个 setup 的 route_ids 全部引用 `Route.side=CT`，route_allocations 总人数等于队伍人数并满足各 Route min/max
- **AND** CT setup 的选择输入不包含本回合 T StrategyTemplateID、T Route、T Intent 或敌方隐藏位置

#### Scenario: 双方都拥有可执行地图行动

- **GIVEN** 任一正式开局模板被选择
- **WHEN** planner 为 T 和 CT 生成 OpeningDeploy
- **THEN** T 进攻 Actor 有从 T_SPAWN 出发的合法路线
- **AND** CT 防守 Actor 有从 CT_SPAWN 到 A/B/中路默认站位的合法路线或 Hold 部署
- **AND** A/B 回防和必要转点存在可达语义边

#### Scenario: 核心节点与阶段场景完整

- **WHEN** 校验正式 de_dust2 配置
- **THEN** 核心节点至少包含 T_SPAWN、LONG_DOOR、A_LONG、PIT、A_SITE、MID、CATWALK、SHORT、B_UPPER、B_TUNNEL、B_SITE、CT_SPAWN、CT_MID、B_DOOR 和 CAR
- **AND** Scenario 覆盖 OpeningDuel、MidControl、SiteEntry、Retake 和 BombResolution
- **AND** A、B 和中路/转点主路径均有相应场景或明确配置关联

#### Scenario: 占位深度配置被拒绝

- **GIVEN** MapConfig 只有一个 RouteTemplate、一个 Scenario 和一条 T 路线
- **WHEN** 作为正式 de_dust2 配置执行完整校验
- **THEN** 返回 `CONFIG_INCOMPLETE_DUST2_COVERAGE`
- **AND** 引擎不以该配置继续生成看似完整的 MR12 战报

### Requirement: 调度与行动生命周期参数来自配置快照

离散事件状态机的移动、决策、Encounter、Utility、Intel/Control、NoOp 和安全上限 SHALL 只由已校验 CombatConstants 快照提供；RuleSet SHALL 只描述 MR12/换边/加时等整场赛制，不覆盖回合模拟参数。正式第一版必配库存 SHALL 精确覆盖主设计 `tb_combat_const` 表列出的键：Round/Bomb/Plant/Defuse/Pickup 时间及 Pickup/Plant/Defuse/Move 边界，ForceExecuteThreshold、DecisionDelay、MaxDecisionCount、MaxEncounterPulses、Min/MaxCombatDuration、PulseFireWindow、CombatScale、DamagePotential/ExposureModifier 边界、Noise/CloseScoreGap/DecisiveScoreGap、DefaultStrategyTemplateID、DefaultCTSetupTemplateID、StrategyMemory 相关键、ControlIntelTTL、CommunicationDelay、Sound/Death confidence 与 IntelTTL 边界、UtilityBudget、MaxStateTransitions、MaxScheduledActions、MaxEffectsPerTimestamp、MaxNoOpTransitions、MaxRotationsPerTeam、MaxRoundTimeline，以及 Attribute/HP/Stamina/Focus/HitChance/KillChance clamp。场景族十属性权重和具体修正 MAY 存放于 EncounterModifier；除此之外实现 SHALL NOT 使用隐藏默认值补齐缺项。

#### Scenario: Scheduler 参数完整可读

- **WHEN** `server/internal/match` 构建并校验 MapConfig/RuleSet
- **THEN** scheduler、planner、movement、encounter、intel 和 bomb 所需上限/时间参数均可通过类型化访问器读取
- **AND** 所有最大值为正且 min 不大于 max
- **AND** resolver 不以代码 fallback 掩盖正式配置缺失
- **AND** RuleSet 中不存在会覆盖 RoundTime/Bomb/Plant/Defuse/Pulse/Decision/Scheduler 的第二套权威值；兼容 DTO 副本若保留必须在适配层验证相等后丢弃

#### Scenario: 第一版通信配置固定即时共享

- **GIVEN** MapConfig 包含预留 `CommunicationDelay`
- **WHEN** 校验第一版正式 de_dust2 配置
- **THEN** 该值必须等于 0
- **AND** 非零值返回 `CONFIG_UNSUPPORTED_COMMUNICATION_DELAY`
- **AND** 引擎不静默忽略非零值，也不建立延迟投递或虚假声音情报机制

#### Scenario: 默认战术由配置指定

- **GIVEN** OpeningPlan 初次以 `OpeningPlanSeed = Hash(RoundSeed, "opening_plan", 0)` 生成非法，并以同一 root seed、`AttemptOrdinal=1` 的派生 seed 确定性重试仍失败
- **WHEN** planner 选择恢复模板
- **THEN** 只读取 MapConfig/CombatConstants 中存在且合法的 `DefaultStrategyTemplateID`
- **AND** `matchengine` 不硬编码 `Default_Pick` 或其他模板 ID
- **AND** 默认模板不存在或仍非法时返回结构化错误

#### Scenario: 默认 CT setup 由独立配置指定

- **GIVEN** CT setup 初次以 `CTSetupSeed = Hash(RoundSeed, "ct_setup", 0)` 生成非法，并以同一 root seed、`AttemptOrdinal=1` 的派生 seed 重试仍失败
- **WHEN** CT planner 选择恢复 setup
- **THEN** 只读取存在、合法且 `side=CT` 的 `DefaultCTSetupTemplateID`
- **AND** 不得从 T 当前模板的 common_ct_setup_ids 强制选择实际 setup
- **AND** 默认 setup 缺失、阵营错误或仍非法时返回结构化错误

### Requirement: 场景族属性权重完整且归一化

配置 SHALL 为 OpeningDuel、FastExecute/SiteEntry、SlowDefault/MidControl 和 Retake/PostPlant 场景族提供 Aim、Reaction、Positioning、Awareness、Teamplay、Utility、Composure、Mobility、Endurance 和 Discipline 十项权重，每个场景族的权重和 SHALL 在配置精度容差内等于 1.0。

#### Scenario: 权重和错误拒绝加载

- **GIVEN** 某场景族缺少一项权重或十项权重和超出允许容差
- **WHEN** ValidateMapConfig 校验配置
- **THEN** 返回 `CONFIG_BAD_SCENARIO_WEIGHT`
- **AND** 引擎不使用写死的默认权重替代该场景族

### Requirement: 配置支持情报、控制权和抽象拦截

MapConfig SHALL 提供 ControlIntelTTL、必须为 0 的 CommunicationDelay、情报 confidence/TTL 边界、主要 Visibility、MapEdge risk_points/intercept_nodes 及其合法引用，使 AI 信息边界和移动拦截可以由地图语义驱动；风险热点 SHALL NOT 表达运行时必死或固定掉包结果，声音配置 SHALL NOT 表达虚假位置或虚假人数。

#### Scenario: 拦截引用和情报边界可校验

- **GIVEN** MapEdge 引用了 intercept node、risk point 或 Visibility
- **WHEN** ValidateMapConfig 校验正式地图
- **THEN** 所有引用节点存在且坐标/用途合法
- **AND** confidence/TTL 上下界合法
- **AND** 配置中不存在 `always_kill`、`forced_drop` 或等价运行时结果字段
- **AND** 死亡情报 confidence 上限不超过 70，声音情报范围为 30-70
