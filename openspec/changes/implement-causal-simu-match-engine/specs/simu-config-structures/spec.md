## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Server reuses Luban TbPlayer config

`match.Service` SHALL 通过 `windypath.com/cs2match/config` 读取 `TbPlayer`，将每名选手完整映射为自包含的 `matchengine.PlayerProfile` 后放入 `MatchInput`；正式引擎输入 SHALL NOT 依赖旧的三属性 `Combatant` 结算模型。`matchengine` 不直接依赖 `config` 包，也不得在模拟期间回查或修改 `TbPlayer`。

#### Scenario: Service maps the complete player profile

- **GIVEN** `TbPlayer` 已随 cfg.Init 加载且包含有效选手
- **WHEN** match.Service 构造 MatchInput
- **THEN** PlayerProfile 包含稳定 ID、名称、角色标签、头像以及 Aim、Reaction、Positioning、Awareness、Teamplay、Utility、Composure、Mobility、Endurance、Discipline 十项基础属性
- **AND** 当前阵营武器装配在回合输入/状态中独立派生，不写回 PlayerProfile
- **AND** 选手不存在或必需属性非法时返回 INVALID_LINEUP

### Requirement: Combat constants are centralized

正式模拟使用的回合时间、炸弹时间、资源边界、概率 clamp、战斗尺度、场景属性权重、决策阈值和标定参数 SHALL 统一来自 `MapConfig.CombatConstants` 及关联的配置快照。`matchengine` 可以定义稳定键名和纯公式顺序，但 SHALL NOT 在 resolver 中散落未命名数值或以代码默认值掩盖正式输入缺失。

#### Scenario: Constants come from the validated snapshot

- **GIVEN** `server/internal/match` 已从 Luban 生成表构建有效 `MapConfig`
- **WHEN** Encounter、Decision 或 Bomb resolver 读取平衡参数
- **THEN** 参数通过类型化 CombatConstants/EncounterModifier 访问层获得
- **AND** 必配键缺失时配置校验失败
- **AND** `matchengine` 不直接导入 `windypath.com/cs2match/config`

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
