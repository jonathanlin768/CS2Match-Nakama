## 1. 建立失败基线和禁止项

- [x] 1.1 为当前 `resolveRoundWinner -> resolveRoundEvents` 反向因果链增加失败测试，证明预定胜方正在影响击杀方向、目标存活人数和下包/拆包分支。
- [x] 1.2 增加生产源码静态检查，禁止 Encounter、Decision、Bomb、Scheduler 和 Terminal API 接收 `winner`、预定 `win_reason`、`target_survivors` 或等价输入。
- [x] 1.3 增加回合不变量测试骨架，覆盖事件先于终局、伤害先于死亡、RoundEnd 后无比赛内事件、Bomb/Carrier/Player 状态一致。
- [x] 1.4 把 MR12、换边和加时比分规则测试从 `MatchInput.ForcedRoundWinners` 解耦，先建立独立 MatchRules/score state 测试或明确的 Match 层 fake 接口。

## 2. 完成 Dust2 第一版 Luban 配置

- [x] 2.1 盘点当前 `#RouteTemplate/#Scenario/#MapTag/#EncounterModifier/#MapNode/#MapEdge/#Visibility/#Route/#CombatConst` 与 `simuMatchDesign.md` 的差距，形成稳定 ID 和引用清单；RouteTemplate 需补齐 `side/route_ids/route_allocations/common_ct_setup_ids`。
- [x] 2.2 在 Luban 表源中补齐六个 `side=T` 战术模板：A_Long_Rush、A_Short_Split、B_Tunnel_Explode、Mid_To_B、Default_Pick、Fake_A_Go_B，以及目标包点、节奏、人数、角色、场景、标签、路线分配和 fallback。
- [x] 2.3 补齐多个 `side=CT` 开局 setup 模板以及 T 进攻、CT 开局防守/补防/回防路线，覆盖 A 大、A 小、B 二楼、中路、A/B PostPlant 所需语义路径；每个 CT setup 的 route_allocations 总人数闭合并满足 Route min/max。
- [x] 2.4 补齐核心节点 T_SPAWN、LONG_DOOR、A_LONG、PIT、A_SITE、MID、CATWALK、SHORT、B_UPPER、B_TUNNEL、B_SITE、CT_SPAWN、CT_MID、B_DOOR、CAR 及必要风险热点。
- [x] 2.5 补齐 MapEdge、主要 intercept_nodes/risk_points 和关键 Visibility，保证双方开局路线、转点、救包和回防在语义图上可达。
- [x] 2.6 补齐 OpeningDuel、MidControl、SiteEntry、Retake、BombResolution 的 A/B/中路 Scenario、MapTag 和 EncounterModifier。
- [x] 2.7 按 `simuMatchDesign.md` 的 `tb_combat_const` 完整必配库存逐键补齐 `#CombatConst.xlsx`：Round/Bomb/Plant/Defuse/Pickup 时间及 Pickup/Plant/Defuse/Move 边界、Decision/ForceExecute、Pulse/CombatDuration、DamagePotential/Exposure、Noise/Close/Decisive、DefaultStrategyTemplateID、DefaultCTSetupTemplateID、StrategyMemory、Intel confidence/TTL、UtilityBudget、全部 Scheduler/NoOp 上限及 Attribute/HP/Stamina/Focus/Hit/Kill clamp；十属性场景族权重和具体修正可位于 EncounterModifier。不得少配后以隐藏默认值补齐，也不得新增 TargetHPFactor。
- [x] 2.8 使用 Luban 正式导出流程更新生成数据和代码，确认没有手工修改 `server/config/data/*.json` 或客户端生成 JSON。
- [x] 2.9 扩展 `config_adapter.go` 和 MapConfig 模型映射所有新增字段；把 TbPlayer 完整十属性/角色/档案映射为 PlayerProfile，保持 `matchengine` 不导入配置包且不再依赖旧三属性 Combatant 结算模型。
- [x] 2.10 扩展配置校验/fixture，逐项覆盖 `simuMatchDesign.md` 的 RouteTemplate side、route_ids/route_allocations、T/CT 模板阵营与覆盖、Scenario、MapTag、EncounterModifier、MapNode 几何、MapEdge、Visibility、Route、Plant 节点、CombatConst 和不可达错误矩阵；模板阵营/引用错误使用 `CONFIG_BAD_TEMPLATE_SCENARIO`，模板人数闭合错误使用 `CONFIG_BAD_TEMPLATE_LIMIT`，正式覆盖缺失精确返回 `CONFIG_INCOMPLETE_DUST2_COVERAGE`，十属性权重缺项/不归一精确返回 `CONFIG_BAD_SCENARIO_WEIGHT`。唯一可降级的 `CONFIG_MISSING_KILL_SAMPLE` 仅用于具有合法 x/y、被选作 KillSample 来源但缺少可用区域几何的节点并回退其 x/y；缺少坐标仍返回 `CONFIG_BAD_NODE_COORD`。
- [x] 2.11 实现 `CommunicationDelay` 第一版只接受 0，非零返回 `CONFIG_UNSUPPORTED_COMMUNICATION_DELAY`；T OpeningPlan 使用 `OpeningPlanSeed`、CT setup 使用独立 `CTSetupSeed`，两者保持相同 MatchInput/root seed，初次 AttemptOrdinal=0、非法时仅以 AttemptOrdinal=1 重试一次，再分别只使用 DefaultStrategyTemplateID / DefaultCTSetupTemplateID；CT 选择不得读取本回合 T 隐藏计划，默认仍失败则报错且不得硬编码模板/站位。

## 3. 拆分整场规则与单回合模拟接口

- [x] 3.1 提取 `MatchState.ScoreByTeam[TeamID]`、阵营映射、常规赛/加时结束和 SideSwitch 逻辑，使其只消费自然 `RoundTerminal`；RoundInput 只读携带 ScoreByTeam，公开 DTO 固定投影 ScoreTeamA/ScoreTeamB；SideSwitch 只修改 SideByTeam，`ScoreT/ScoreCT` 只按当前 TeamTID/TeamCTID 投影，不得交换或重置队伍比分。
- [x] 3.2 定义内部 `roundSimulator` 接口；生产 Service 固定装配 causal RoundEngine，MatchRules 测试可注入不生成战斗事件的 fake。
- [x] 3.3 删除正式 MatchInput 中的 `ForcedRoundWinners` 以及可用固定 RoundCount/MaxRounds 截断整场的字段；删除/拒绝 DebugSimuMatch 的任何单回合、RoundCount 或 MaxRounds 兼容字段，更新 MR12/加时测试 fixture，确认 RPC JSON 无法提交胜方脚本或截断整场，且业务 RPC 不获得内部 roundSimulator。
- [x] 3.4 保留 `Simulate(ctx, *MatchInput)` 唯一正式入口和 MatchResult 契约，使其始终按 RuleSet 模拟到比赛终局；增加整场规则层单测验证早停、换边、完整 OT block 和多 OT。
- [x] 3.5 冻结自包含 MatchInput：包含 MapConfigSnapshot、WeaponSpecs、只描述 MR12/换边/加时的 RuleSet、两队 PlayerProfile 和 root Seed；PlayerProfile 映射头像/角色/十属性但不绑定武器，RoundInput 按当前阵营装配 AK-47/M4A1-S 并在换边时重新派生；RoundEngine 的时间/脉冲/决策/Scheduler 参数只读 CombatConstants，兼容 RuleSet 同名旧字段在适配层验证后丢弃。

## 4. 实现权威 RoundState 与状态投影

- [x] 4.1 实现内部 `RoundState`，包含 Phase、Timeline、RoundDeadline、BombDeadline、Plan、Players、Bomb、Nodes、Intel、ActiveEngagements、Scheduler、Momentum、Utility、NoOpCount、NoProgressEligible、恢复尝试状态、Events 和 Terminal。
- [x] 4.2 实现 `RoundPlayerState`，从 PlayerProfile/当前阵营装配初始化 HP、Stamina、Focus、Suppressed、Posture、Location、Intent、ActionState、Bomb 和统计。
- [x] 4.3 实现 NodeRuntimeState 的 ActualControl/KnownControl、UpdatedAt/ObservedBy/TTL，以及 Contested 结束后的合法归并。
- [x] 4.4 实现内部 BombState 的 Carried、Dropped、Planting、Planted、Defusing、Defused、Exploded 及唯一 Carrier/Action/绝对时间字段。
- [x] 4.5 为 HP、Stamina、Focus、Momentum、概率和时间实现集中 clamp 与状态不变量验证。
- [x] 4.6 实现内部状态到现有 RoundResult、PlayerState、BombPublicState、FinalControls 的单向投影测试，禁止公开 DTO 反向驱动模拟。

## 5. 实现 Action/Effect 与确定性 Scheduler 内核

- [x] 5.1 定义 Intent、ScheduledAction、PlayerActionState、BusyInterval、Effect、AppliedBatch 和稳定 Action/Effect/Event ID 规则；公开 GameEvent 保留 EventID/SourceActionID/SourceEffectID，action 生命周期事件允许空 SourceEffectID 但不得缺 SourceActionID。
- [x] 5.2 实现按 `ResolveAt/Priority/ActionType/MinActorID/ActionID` 排序的确定性 heap，并测试任意插入顺序得到相同弹出序列；事件投影独立按 `Timestamp/Priority/ActionType/MinActorID/SourceActionID/SourceEffectID/EventID` 排序。
- [x] 5.3 实现 Actor 互斥占用、group action、ActionVersion 增长和 ScheduledAction.VersionByActor；完成前逐 Actor 校验版本/存活/CurrentActionID/位置。group action 冻结由来源 Route/模板人数约束派生的 MinRequiredActors，成员失效后按剩余合法人数确定性继续或整体取消，队形完整度只作为 Reason 输入，不使用隐藏阈值或临时随机；过期 action 静默丢弃。
- [x] 5.4 实现 RoundDeadline、BombDeadline、Intel/Control TTL 与队首 action 的统一 next-time 选择，确保普通 action 不能越过更早 deadline；resolver 只计算 Duration/ResolveAt，不得用阶段 TimeCost 直接推进或跳过 Encounter 中间 pulse/action。
- [x] 5.5 实现同时间戳 priority 分层事务，并把整个 CombatPulse 作为单一 priority-100 `CombatPulseCommit`：同 pulse 全部攻击基于一个 snapshot 计算、全部 DamageEffect 原子应用后再派生 Death/KILL/BombDrop；完成后才重新校验 Plant/Defuse/Movement 等外部低带 action。
- [x] 5.6 实现 MaxStateTransitions、MaxScheduledActions、MaxEffectsPerTimestamp、MaxNoOpTransitions、MaxRotationsPerTeam 和 MaxRoundTimeline guard，超限返回结构化错误。
- [x] 5.7 添加 Scheduler 单测：同 pulse 互换、致命攻击不取消同 pulse 非致命反击、多攻击者伤害/击杀稳定归属、逐 Actor VersionByActor 失效、group action 剩余人数达到/低于 MinRequiredActors 的继续/取消、死亡取消移动/Plant/Defuse、Defuse 等于 Explode、拆包者同秒死亡、Plant 等于 RoundExpire、空队列推进 deadline。

## 6. 实现语义移动、OnEdgeLocation 和拦截

- [x] 6.1 实现基于 Route/MapEdge 的 MoveTime 计算，使用 BaseTime、Tempo、Mobility、Stamina 和 Formation 修正并 clamp。
- [x] 6.2 实现 Node XOR Edge 的 PlayerLocation、移动 Progress、From/To 插值和派生 seed 的 OnEdgeLocation 显示位置。
- [x] 6.3 实现配置语义图上的确定性有界路径选择，用于开局、转点、救包和回防；不可达时返回决策反馈而不是传送。
- [x] 6.4 为每次 edge traversal 实现最多一个抽象 InterceptCheck，使用 Risk、Noise、Visibility、姿态、情报和派生 seed 决定是否建立 Encounter。
- [x] 6.5 测试未参战者在其他 Encounter 期间继续移动、边上死亡/掉包位置、拦截打断 MovementArrive、风险热点不直接制造 KILL。

## 7. 实现控制权与最小 TeamIntel

- [x] 7.1 实现 TeamIntel/IntelRecord，覆盖直接视野、Encounter、真实声音、死亡、空点假设和 BombIntel 的 confidence/ObservedBy/LastSeenAt/ExpiresAt；空点假设 confidence 使用 1-29（golden 用 20），死亡瞬间情报不超过 70，声音仅来自真实来源且为区域级 30-70 confidence，不生成虚假位置/人数。
- [x] 7.2 实现 CommunicationDelay=0 的队内即时共享、观察者死亡降级、Intel/Control priority 10 衰减和 KnownControl TTL；确认没有延迟投递队列或假声音分支。
- [x] 7.3 构造按稳定 ID 排序的只读 DecisionView，暴露本方状态、KnownControl、TeamIntel 和合法 BombIntel，排除完整 RoundState 引用、敌方 Intent/ActionQueue/未观察位置/ActualControl；Decision/Encounter 只接收身份派生 seed 或 RollSource，不共享可变 `*rand.Rand`。
- [x] 7.4 添加信息防作弊测试：隐藏补防不可见、低置信度只修正评分、情报到期失效、未观察 Bomb 不泄露精确节点。

## 8. 实现战术选择、角色分配和开局行动

- [x] 8.1 实现 StrategyScore：模板基础、LineupFit、比分压力、PreviousSuccessBonus、RepeatPenalty、CounterReadRisk，以及同时受 Discipline/IGL/Awareness 约束的噪声；CloseScoreGap/DecisiveScoreGap 只保护评分排序，不预定行动或战斗结果。
- [x] 8.2 实现稳定候选排序/采样和明显优势不被噪声反转，输出完整 Strategy ReasonRecord。
- [x] 8.3 实现 RequiredRoles/RoleTags/属性适配的稳定角色分配、Bomb carrier 选择和角色缺口 penalty。
- [x] 8.4 实现独立 CT setup 选择：只读取本方阵容/角色/属性、CT RouteTemplate、StrategyMemory 和 CTSetupSeed，不读取当前 T 模板/Route/Intent；T/CT 双方从 Timeline 0 生成 Move/Hold Intent/Action，OpeningDeploy 不消耗 RoundTimer。
- [x] 8.5 测试六个 T 模板和多个 CT setup 可达、双方共十名选手均有合法开局 Intent、T/CT Attempt 0/1 各自稳定且派生流不同、两个阵营默认模板恢复、CT 无 T 当前计划泄漏、半场记忆衰减和角色缺口解释。

## 9. 实现有限 UtilityBudget

- [x] 9.1 实现每队回合 UtilityBudget 初始化，使用选手 Utility、Support、装配 Grenades、模板和配置上限。
- [x] 9.2 实现局部 VisibilitySuppression、OpeningInitiative、ExposureReduction、SyncPeek、PlantCover 和 DefuseCover 消耗/持续窗口。
- [x] 9.3 保证 Utility modifier 只作用于指定 action/Encounter，预算耗尽后不再生效，不直接全局增加 TeamPower。
- [x] 9.4 添加预算消耗、Support 效率、重复 Encounter 不复用、LOW_UTILITY Reason 和同 seed 测试。

## 10. 实现 Encounter 生命周期与并发仲裁

- [x] 10.1 实现从当前 Scenario、接触、存活/Busy Actor 生成 Encounter 候选及稳定 ThreatScore。
- [x] 10.2 实现共享 Actor/重叠区域仲裁、Actor 锁定和不重叠 Encounter 并发。
- [x] 10.3 将 EncounterResolver 明确拆为 `Start -> EncounterSchedule` 与 `ResolvePulse -> CombatPulseResult`：Start 只计算 Posture、CombatDuration、Pulse 数/绝对时间并安排 CombatPulse/CombatEnd，不预生成未来 CombatEvents/伤亡，也不阻塞其他 action。
- [x] 10.4 CombatEnd 释放 Actor、更新实际/已知控制、处理撤退/失去接触并触发 Decision，不直接设置 Winner。
- [x] 10.5 测试共享 Actor 不重复参战、A/B 独立行动并发、局部全灭只结束 Encounter、脉冲上限/时长退出。
- [x] 10.6 测试 A 大 Encounter 尚未结束时，不共享 Actor 的 B 区选手可继续进入 SiteContest、启动 Plant 或触发 Decision；主导 RoundPhase 只作为投影，不能充当全局行动锁。

## 11. 实现 CombatPulse 评分、目标和伤害

- [x] 11.1 实现场景族十属性 WeightedPlayerScore、唯一 `CombatScore=PlayerCombatScore` 公式和团队级 EncounterScore：PlayerCombatScore 精确为 `WeightedPlayerScore + RoleTag + Weapon + Posture + Visibility + TeamSupport - Stamina - Damage - Suppression`；Utility/Momentum/TimePressure 只在 EncounterScore 各计算一次。EncounterScore 只影响主动权、时长、脉冲和撤退倾向，不设置胜者。
- [x] 11.2 实现唯一 `TargetSurvivalScore = Positioning + Reaction + Cover + TeamSupport + Focus - MovementExposure - DamagePenalty`，不得引入 WeightedSurvivalAttributes 或 SuppressionPenalty；实现 TradeCoverage、Crossfire、Spacing、SyncPeek、Isolation 等实际参与者协同计算。
- [x] 11.3 实现合法目标过滤与 ThreatScore/Exposure/Distance/HP/PlayerID 稳定排序，使用 Actor/Pulse/Target 身份派生 roll。
- [x] 11.4 实现 sigmoid HitChance、DamagePotential、ExposureModifier 和 `KillChance=clamp(min(HitChance*DamagePotential*ExposureModifier, HitChance),0,MaxKillChance)`；禁止 TargetHPFactor，让 Weapon Damage/RPM/Magazine/ArmorPenetration/RangeModifier 实际参与。
- [x] 11.5 实现 Miss、非致命和致命 damage package；所有结果先产生实际 DamageEffect，只有 HP 应用到 0 才派生 Death/KILL/BombDrop。
- [x] 11.6 实现同 pulse snapshot 攻击资格、实际损失 HP 伤害统计、Focus/Stamina/Suppression/Momentum/Posture 更新和自然首杀/补枪标记。
- [x] 11.7 添加 PlayerCombatScore/TargetSurvivalScore 精确 golden 和“Utility/Momentum/TimePressure 不重复计分”断言，并覆盖 EncounterScore、Close/Decisive 趋势保护、噪声后仍执行概率采样、十权重和、KillChance<=HitChance/无 TargetHPFactor、武器差异、Armor/Range、非致命受伤、低 HP、同 pulse 互换、首杀非预抽和 map 顺序无关测试。

## 12. 实现循环中期 Decision 与 Planner

- [x] 12.1 实现 EncounterEnd、人数/资源档位、控制变化、空点、Bomb、路线受阻、ForceExecuteThreshold 和 PostPlant 事件的 DecisionTrigger。
- [x] 12.2 实现 T Continue/Rotate/ForceExecute/RecoverBomb/Plant、Lurker GatherIntel/HoldFlank/InterceptRotate 和 CT Reinforce/Hold/Retake/Defuse/Save/InterceptRotate 候选评分；CT 截回防不得读取 T 的真实 Rotate action。
- [x] 12.3 实现 DecisionDelay，选择结果必须生成实际 Move/Hold/Plant/Defuse action，而不是只产叙事事件。
- [x] 12.4 实现 MaxDecisionCount/MaxRotationsPerTeam：达到上限后限制继续转点并强制可达执行，但不直接指定胜负。
- [x] 12.5 添加开局伤亡、低 HP/Focus/Stamina、Bomb 掉落、控制权、情报置信度和时间压力改变后续实际 action 的测试。

## 13. 实现 SiteContest 与 Bomb 状态机

- [x] 13.1 实现 SiteContest planner：根据当前到达 Actor、CT 威胁、控制权和 Utility 选择 Site Encounter、撤回或 PlantDecision。
- [x] 13.2 实现 PlantScore/PlantRisk、PlantTime、PlantStart/Complete、可见威胁拦截、伤害/死亡打断和同秒 RoundExpire 边界。
- [x] 13.3 实现持包者死亡的边上/节点 BombDrop、CTControlled/Contested 先争夺、实际 Move+PickupTime 后拾取和不可达失败。
- [x] 13.4 PlantComplete 后切换 BombDeadline、安排不可遗漏 BombExplode，并使 RoundDeadline 失效。
- [x] 13.5 实现 PostPlant Retake、到达时间、DefuseScore/Time/HasKit、T CanDenyDefuse、DefuseStart/Interrupt/Complete。
- [x] 13.6 添加完整炸弹测试：下包者同秒死亡、Plant 等于 RoundExpire、掉包救包、未下包且 Bomb=Dropped 时 T 全灭使用内部 `T_ELIMINATED_BOMB_DROPPED` 而非一般 `T_ELIMINATED`、T 全灭后 CT 拆/爆、CT 全灭 bomb_secured、Defuse 等于 Explode、拆包者同秒死亡失败、其他 CT 同秒死亡但拆包者存活成功。

## 14. 实现纯 Terminal、NoOp 和 RoundEngine 循环

- [x] 14.1 实现只读取 AppliedBatch 后 RoundState 的 `EvaluateRoundTerminal`，按 Defused、Exploded、PrePlant 淘汰、PrePlant timeout、PostPlant bomb_secured、`NoProgressEligible && ValidNoProgress` 的优先级返回 RoundTerminal；ValidNoProgress 只是纯条件，公开独立 WinReason `no_progress_timeout` 只能经显式资格门槛产生。
- [x] 14.2 实现终局不变量，任何事件/Player/Bomb/Timer 不一致返回 `SIMULATION_INVARIANT_ERROR`，不得修补或重新采样。
- [x] 14.3 实现包含 Timeline/位置/资源/Bomb/控制/Intel/Intent/ActionQueue/RecoveryAttempt/事件的状态指纹 NoOp，以及 `CycleID + NotAttempted/Running/Failed/Succeeded` 恢复生命周期：CycleID 由 RoundSeed 和单回合单调不回退的 RecoveryOrdinal 派生；达到阈值后冻结 NoOpCount，每周期至多安排一次确定性 ForceExecute/RecoverBomb；当前 RecoveryActionID 自身的排队、Timeline、移动、事件和完成 bookkeeping 保留周期，完成后成功恢复则重置、失败则保留证明并执行一次 NoProgressCheck，其他来源真实进展或外部打断才清空周期/资格/计数。PostPlant 只推进合法 action/deadline；配置/调度/不变量错误返回结构化错误。
- [x] 14.4 完成 RoundEngine 主循环：ensure actions、选择 action/deadline、AdvanceTime、同秒事务、事件投影、Intel/DecisionTrigger、Terminal、Phase 重评估。
- [x] 14.5 将全局阶段实现为可循环 OpeningDeploy/Advance/Clash/RotateDecision/SiteContest/Planting/PostPlant/RoundEnd，验证多次冲突和多次决策。

## 15. 切换生产管线并删除旧因果倒置逻辑

- [x] 15.1 将 `MatchEngine.simulateRound` 切换为新的 causal RoundEngine，Match 层只在获得 RoundTerminal 后更新 Team A/B 比分。
- [x] 15.2 删除 `resolveRoundWinner`、`resolveRoundEvents(...winner...)`、`pickKillSide`、`survivorTarget`、按胜方下包/拆包分支和 `inferWinReason(winnerSide, ...)`。
- [x] 15.3 删除或重构只服务旧预定胜方路径的 `teamPower/playerPower/buildControls`，保留可复用位置采样/统计辅助并迁入对应模块。
- [x] 15.4 运行静态禁止项测试，确认生产调用图中不存在胜方/终局反向输入且 `matchengine` 不导入配置包。

## 16. 可解释战报、统计和比赛记忆

- [x] 16.1 实现 ReasonRecord 与向后兼容 GameEvent/EventReason 投影：保留 Code/MainFactor/float64 ScoreDelta/Detail，新增结构化 Modifiers、可空 Probability、Formula、Inputs、仅公开 StateChanges、EventID/SourceActionID/SourceEffectID；StateChanges 使用 Kind + 单一 Number/String/Bool/Null 标量，区分 nil 概率与实际 0，禁止不受控 map/object、截断分差或泄漏隐藏状态。仅已应用 Effect 或真实 action 生命周期变化能生成 EventReason。
- [x] 16.2 实现 ROUND_START、STRATEGY_ADJUSTED、DAMAGE、KILL、ROTATE、REINFORCE、CONTROL_GAINED、BOMB_DROP/PICKUP、BOMB_PLANT_START/INTERRUPT/PLANT、DEFUSE_START/INTERRUPT、BOMB_DEFUSE/EXPLODE、ROUND_END 事件和关键状态快照；Start/Interrupt 必须来自真实 action 生命周期。
- [x] 16.3 实现 Node/Area/RiskHotspot/OnEdge 位置来源与 LocationSeed，确认位置不反向参与已完成战斗结果。
- [x] 16.4 从实际 Damage/KILL/Bomb 事件聚合 kills、deaths、damage、ADR、FK、MK、plants、defuses，防止 overkill 伤害超记。
- [x] 16.5 仅从完成 RoundResult 更新 StrategyMemory，验证 PreviousSuccess、RepeatPenalty、CounterReadRisk 和半场衰减边界。
- [x] 16.6 实现 RoundResult/MatchResult.ExplainableReport，聚合实际 KeyEvents、StrategySummary、LossReasons、WinFactors；最终 Winner 只能选择报告视角，不能补造 Reason。

## 17. 端到端测试、标定和兼容验证

- [x] 17.1 为生产 RoundEngine 增加跨 seed 端到端 invariant 测试，覆盖 elimination、timeout、bomb_defused、bomb_exploded、bomb_secured、no_progress_timeout（含 Recovery Cycle 自身进展保留失败证明、外部进展重置）和同秒边界。
- [x] 17.2 增加确定性深度比较：相同 MatchInput/MapVersion/RuleSet/Seed 的 action/event/reason/location/terminal/比分/统计一致，map 插入顺序不影响结果。
- [x] 17.3 实现短样本 CI 标定和显式 10k+ 长标定，输出 T 胜率、平均击杀、下包/拆包/爆炸率、首杀方胜率、5v3、3v5、强弱队和平均时长。
- [x] 17.4 依据标定只调整 Luban 配置参数并重新导出；确认不存在 TargetWinner、目标终局、结果重采样或测试钩子进入生产路径。
- [x] 17.5 运行 `go test` 覆盖 `server/internal/framework/matchengine`、`server/internal/match` 和 `server/config`。
- [x] 17.6 检查 React MatchReport 类型对新增可选 EventID/SourceActionID/SourceEffectID、Reason/StateChanges/Bomb 字段以及 ScoreTeamA/ScoreTeamB 权威比分的兼容性，补前端类型/回放测试并运行 `npm run build`，确认客户端只消费服务器权威状态且不把 ScoreT/ScoreCT 当作队伍累计分。
- [x] 17.7 使用 `go build -buildmode=plugin -trimpath -o build/backend.so .` 或 `server/build.bat` 验证 Nakama Go Plugin 编译；确认未新增 RPC、Match Handler、Storage 权限和外部依赖。
- [x] 17.8 运行 `openspec validate implement-causal-simu-match-engine --strict` 并复核生成数据、实现 diff、测试报告和标定摘要。
- [x] 17.9 对照 `doc/simuMatchDesign.md` 建立章节级追踪表，确认公式、阶段并发、情报边界、事件 DTO/source IDs、队伍比分/阵营投影、group action 版本、Recovery Cycle、同秒炸弹规则和配置错误矩阵均有 design/spec/task/test 落点且不存在互相矛盾的 MAY/SHALL。
