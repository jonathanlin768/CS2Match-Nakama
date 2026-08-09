## ADDED Requirements

### Requirement: 运行时地图语义驱动可打断移动

RoundEngine SHALL 使用 Route、MapNode、MapEdge、Visibility、risk_points 和 intercept_nodes 生成 Move、MovementArrive 和 InterceptCheck，而不是把路线仅用于回合标题或击杀坐标。移动 SHALL 消耗配置时间/体能并维护 Node XOR OnEdgeLocation；地图语义只提供风险和接触条件，不直接产生死亡。

#### Scenario: MapEdge 决定移动时间和到达

- **GIVEN** 当前路线包含 from/to 节点有效且 BaseTime 为 8 秒的 MapEdge
- **WHEN** planner 为选手生成该 edge 的 Move
- **THEN** ResolveAt 使用 BaseTime 以及 Mobility、Stamina、Tempo 和队形配置修正
- **AND** MovementArrive 成功前选手处于 OnEdgeLocation
- **AND** 到达后 CurrentNode 才切换为目标节点

#### Scenario: Visibility 和风险产生拦截候选而非结果

- **GIVEN** 移动 edge 存在合法 risk_point/intercept_node，且敌方存在有效 Visibility 威胁
- **WHEN** scheduler 安排本次 traversal
- **THEN** 可以基于配置权重和派生 seed 生成一个 InterceptCheck
- **AND** 只有 InterceptCheck 形成合法 Encounter 且 CombatPulse 造成致命伤害时才生成 KILL
- **AND** 配置热点本身不能直接把选手设为死亡或让炸弹掉落

### Requirement: 地图语义图支持双方开局、转点、救包和回防

正式 de_dust2 MapConfig SHALL 使 T 和 CT 都能通过配置 Route/MapEdge 执行 OpeningDeploy，并支持设计文档主战术所需的转点、Bomb pickup、A/B retake 和 defuse 到达检查。路径选择 SHALL 是语义图上的确定性有界选择，不要求物理导航网格。

#### Scenario: CT 使用配置路线回防

- **GIVEN** Bomb 已在 A_SITE 下包且存活 CT 位于 B_SITE/CT_MID 路线节点
- **WHEN** CT planner 生成 Retake
- **THEN** 使用配置 MapEdge/Route 计算到达 A_SITE 的实际时间
- **AND** 到达时间进入 CanDefuse/DefuseFinishAt 判定
- **AND** 引擎不以固定 CT 回防时间或预定胜方代替路径结果

#### Scenario: Bomb 不可达不会被瞬移拾取

- **GIVEN** Bomb 掉落节点与所有存活 T 之间没有合法语义路径
- **WHEN** T planner 评估 RecoverBomb
- **THEN** 不生成 PickupComplete
- **AND** 状态机先评估其他合法强攻或可恢复路线
- **AND** 若配置/图合法但没有任何存活 T 能到达炸弹后再到包点，且没有其他持包/下包路径，则只在连续 NoOp 达阈值、一次确定性恢复尝试仍失败并通过显式 NoProgressCheck 后形成正式 `no_progress_timeout`

## MODIFIED Requirements

### Requirement: 引擎消费地图配置快照

`matchengine` SHALL 消费自包含的地图配置快照，该快照包含完整第一版所需的带 side/route_ids/route_allocations 的 T 战术与 CT setup RouteTemplate、场景、地图标签、遭遇战修正、地图节点、地图路径、视野行、双方路线和战斗常量。引擎 SHALL 使用这些数据进行独立战术/setup 选择、双方部署、语义移动、接触/拦截、Encounter 修正、控制权、炸弹可达性和事件位置，而不只是选择一个回合主路线。

#### Scenario: 有效 Dust2 地图配置初始化因果引擎

- **GIVEN** `server/internal/match` 将 Luban 生成的完整 Dust2 表转换为 `matchengine.MapConfig`
- **WHEN** `matchengine.NewService` 或 `Service.Simulate` 收到该快照
- **THEN** 引擎可以在不读取 `server/config` 的情况下选择战术/场景并生成双方 Move/Hold/Encounter/Bomb action
- **AND** 路线、节点、路径、视野、风险、场景修正或战斗常量不从硬编码 Dust2 全局变量读取
- **AND** 必需地图配置缺失时返回错误，而不是用内建 fallback 数据替换

### Requirement: 地图配置校验语义引用

地图配置校验层 SHALL 在模拟使用配置前拒绝重复 ID、缺失引用、非法枚举值、模板/路线阵营不匹配、route_allocations 人数不闭合、非法节点几何、断裂/不可达路线、缺失必需战斗常量、非法风险/拦截/视野引用、缺少双方开局/回防能力、缺少设计文档主模板/阶段覆盖以及属性权重不归一。校验 SHALL 保留 `simuMatchDesign.md` 配置矩阵中的稳定错误码；模板引用/阵营错误使用 `CONFIG_BAD_TEMPLATE_SCENARIO`，模板人数/route_allocations 错误使用 `CONFIG_BAD_TEMPLATE_LIMIT`；除主设计明确允许的 `CONFIG_MISSING_KILL_SAMPLE` 告警外，不得把具体错误折叠成不可诊断的通用成功/fallback。

稳定错误矩阵 SHALL 包括：`CONFIG_DUP_ROUTE_TEMPLATE`、`CONFIG_BAD_TEMPLATE_SCENARIO`、`CONFIG_BAD_TEMPLATE_LIMIT`、`CONFIG_DUP_SCENARIO`、`CONFIG_BAD_SCENARIO_TAG`、`CONFIG_BAD_SCENARIO_MAP_TAG`、`CONFIG_BAD_SCENARIO_WEIGHT`、`CONFIG_INCOMPLETE_DUST2_COVERAGE`、`CONFIG_DUP_MAP_TAG`、`CONFIG_BAD_MAP_TAG`、`CONFIG_BAD_ENCOUNTER_MODIFIER`、`CONFIG_BAD_REASON_CODE`、`CONFIG_DUP_NODE`、`CONFIG_BAD_NODE_ENUM`、`CONFIG_BAD_NODE_COORD`、`CONFIG_BAD_NODE_SHAPE`、`CONFIG_BAD_NODE_CIRCLE`、`CONFIG_BAD_NODE_POLYGON`、`CONFIG_BAD_EDGE_NODE`、`CONFIG_BAD_RISK_POINT`、`CONFIG_BAD_INTERCEPT_NODE`、`CONFIG_BAD_VISIBILITY_NODE`、`CONFIG_BAD_ROUTE_NODE`、`CONFIG_ROUTE_NOT_CONNECTED`、`CONFIG_BAD_ROUTE_LIMIT`、`CONFIG_NO_PLANT_SITE`、`CONFIG_DUP_COMBAT_CONST`、`CONFIG_BAD_COMBAT_CONST_TYPE`、`CONFIG_BAD_COMBAT_CONST_RANGE`、`CONFIG_MISSING_COMBAT_CONST`、`CONFIG_BAD_BOMB_CONST`、`CONFIG_UNSUPPORTED_COMMUNICATION_DELAY` 和 `CONFIG_UNREACHABLE_NODE`；运行时开局计划错误使用 `INVALID_OPENING_PLAN`。

#### Scenario: 路线引用缺失节点

- **GIVEN** `tb_route.nodes` 包含 `A_LONG` 和 `MISSING_NODE`
- **WHEN** 校验配置快照
- **THEN** 校验失败并返回 `CONFIG_BAD_ROUTE_NODE`
- **AND** 不会为该地图配置启动模拟

#### Scenario: 缺失必需战斗常量

- **GIVEN** 配置快照不包含 `RoundTimeLimit` 或必需 Scheduler/Combat 参数
- **WHEN** 校验配置快照
- **THEN** 校验失败并返回 `CONFIG_MISSING_COMBAT_CONST` 或更具体结构化错误

#### Scenario: 必需配置不使用 fallback

- **GIVEN** 配置快照只有单条 T 路线、一个 RouteTemplate 或缺少 CT 开局/回防路线
- **WHEN** `Service.Simulate` 校验正式 de_dust2 输入
- **THEN** 校验失败并返回 `CONFIG_INCOMPLETE_DUST2_COVERAGE`
- **AND** 引擎不会使用任何硬编码 Dust2 模板、节点、路线或固定移动时间 fallback

#### Scenario: 模板场景和修正引用错误稳定分类

- **GIVEN** 配置分别存在重复 RouteTemplate、缺失 Scenario、非法 Scenario tag、缺失 MapTag 或非法 EncounterModifier reason_code
- **WHEN** ValidateMapConfig 校验对应 fixture
- **THEN** 分别返回 `CONFIG_DUP_ROUTE_TEMPLATE`、`CONFIG_BAD_TEMPLATE_SCENARIO`、`CONFIG_BAD_SCENARIO_TAG`、`CONFIG_BAD_SCENARIO_MAP_TAG` 或 `CONFIG_BAD_REASON_CODE`
- **AND** 不启动该地图模拟

#### Scenario: 节点几何错误与可降级告警分离

- **GIVEN** MapNode 分别存在非法/缺失坐标、Circle 缺 radius、Polygon 顶点不足，或一个具有合法 x/y 且被事件位置解析器选作 KillSample 来源的节点缺少可用 KillSample 区域几何
- **WHEN** ValidateMapConfig 校验
- **THEN** 前三类分别返回 `CONFIG_BAD_NODE_COORD`、`CONFIG_BAD_NODE_CIRCLE`、`CONFIG_BAD_NODE_POLYGON`
- **AND** 最后一类产生 `CONFIG_MISSING_KILL_SAMPLE` 告警并回退该节点 x/y；缺少 x/y 始终属于 `CONFIG_BAD_NODE_COORD`

#### Scenario: 路径视野和风险引用错误稳定分类

- **GIVEN** fixture 分别包含不存在的 Edge 节点、risk_point、intercept_node、Visibility 节点或不连通 Route
- **WHEN** ValidateMapConfig 校验
- **THEN** 分别返回 `CONFIG_BAD_EDGE_NODE`、`CONFIG_BAD_RISK_POINT`、`CONFIG_BAD_INTERCEPT_NODE`、`CONFIG_BAD_VISIBILITY_NODE` 或 `CONFIG_ROUTE_NOT_CONNECTED`

#### Scenario: CombatConst 错误稳定分类

- **GIVEN** fixture 分别包含重复 key、类型不匹配、越界值、必配键缺失、非法炸弹参数或非零 CommunicationDelay
- **WHEN** ValidateMapConfig 校验
- **THEN** 分别返回 `CONFIG_DUP_COMBAT_CONST`、`CONFIG_BAD_COMBAT_CONST_TYPE`、`CONFIG_BAD_COMBAT_CONST_RANGE`、`CONFIG_MISSING_COMBAT_CONST`、`CONFIG_BAD_BOMB_CONST` 或 `CONFIG_UNSUPPORTED_COMMUNICATION_DELAY`

#### Scenario: 非法开局计划只使用各自阵营配置恢复

- **GIVEN** T OpeningPlan 或 CT setup 初次生成超出 Route min/max、模板阵营错误或包含不可达分配
- **WHEN** planner 返回 `INVALID_OPENING_PLAN`
- **THEN** 保持相同 MatchInput/root seed；T 使用 `OpeningPlanSeed`，CT 使用独立 `CTSetupSeed`，两者初次均为 AttemptOrdinal=0 且只以 AttemptOrdinal=1 重试一次
- **AND** T 重试仍失败时只使用合法 `DefaultStrategyTemplateID`，CT 重试仍失败时只使用合法 `DefaultCTSetupTemplateID`
- **AND** CT setup 选择不得读取本回合 T 模板、T Route/Intent 或其他隐藏状态
- **AND** 对应默认模板缺失、阵营错误或仍非法时返回结构化错误，而不是硬编码模板继续模拟

### Requirement: 事件位置使用地图节点和路线语义

KILL、DAMAGE 和炸弹相关事件的位置 SHALL 在对应 action/Encounter 已经确定语义来源后，从当前节点区域、运行时 OnEdgeLocation 或配置风险/拦截候选中使用 LocationSeed 稳定生成。位置采样 SHALL NOT 反向改变命中、伤害、死亡或 Bomb 状态。

#### Scenario: 击杀从节点几何中采样

- **GIVEN** KILL 事件来源节点的 `area_usages` 包含 `KillSample`，且 `shape = Circle`
- **WHEN** 已应用致命 DamageEffect 后生成事件位置
- **THEN** `GameEvent.Location.X` 和 `GameEvent.Location.Y` 位于配置圆形范围内
- **AND** 使用相同 EventID、SourceObjectID 和 seed 重复模拟时返回相同坐标

#### Scenario: 边上死亡使用运行时位置

- **GIVEN** 选手在 MapEdge Progress 0.5 处的 Intercept Encounter 中死亡
- **WHEN** 生成 KILL 和可选 BOMB_DROP 位置
- **THEN** 位置来自该选手当时 OnEdgeLocation
- **AND** 不使用目标路线终点坐标冒充死亡位置
