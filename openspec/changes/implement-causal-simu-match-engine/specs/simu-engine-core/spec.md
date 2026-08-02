## ADDED Requirements

### Requirement: 正式模拟禁止预先选择胜方或终局

`matchengine` 的正式回合模拟 SHALL NOT 在伤害、击杀、时间、控制权和炸弹事件发生前选择、抽取、传入或缓存回合胜方或终局类型。任何遭遇战、决策或炸弹 resolver 的输入 SHALL NOT 包含预定胜方、目标存活人数或用于反向生成事件的预定终局。

#### Scenario: 回合从无胜方状态开始

- **GIVEN** 双方阵容、地图配置、规则集和 seed 均有效
- **WHEN** 引擎开始模拟新回合
- **THEN** 初始 `RoundState` 不包含 Winner、WinnerTeamID 或预定 WinReason
- **AND** 开局遭遇战 resolver 不接收任何胜方参数

#### Scenario: 不允许按预定胜方生成击杀

- **GIVEN** 一次双方均有存活参战者的 Encounter
- **WHEN** 引擎生成战斗脉冲
- **THEN** 每次伤害和击杀由当时的参战者、属性、武器、场景修正、资源和派生随机采样决定
- **AND** 引擎不存在根据回合胜方提高某一方击杀概率或设置目标存活人数的生产分支

#### Scenario: 不允许预抽合法终局

- **GIVEN** 回合尚未发生决定性的伤害、计时或炸弹事件
- **WHEN** 引擎推进任一阶段
- **THEN** 引擎不预先选择 elimination、timeout、bomb_defused、bomb_exploded 或其他终局类型
- **AND** 后续事件序列不受任何预定终局约束

### Requirement: 回合按设计阶段循环推进权威状态

`matchengine` SHALL 使用 `OpeningDeploy`、`Advance`、`Clash`、`RotateDecision`、`SiteContest`、`Planting`、`PostPlant` 和 `RoundEnd` 高层阶段表达回合主导活动。`SelectStrategyTemplate` 与 `AssignRoles` 生成开局计划；Opening Encounter、MidRound Decision、Site Encounter 和 Bomb action SHALL 由事件队列按当前状态反复触发，而不是各执行一次的线性流水线。每次状态变化 SHALL 成为后续 planner/resolver 的权威输入。全局 `RoundState.Phase` SHALL 是从 Bomb/Action/Encounter 状态派生的解释性投影，SHALL NOT 阻止其他区域、不共享 Actor 的合法局部行动。

#### Scenario: 开局伤亡进入中期决策

- **GIVEN** 开局 Encounter 使 T 方一名选手死亡且另一名选手受伤
- **WHEN** 引擎进入中期决策
- **THEN** 决策输入中的存活人数、HP、Focus、Stamina、时间和控制权反映开局 Encounter 的实际结果
- **AND** 决策不得使用开局前保存的完整阵容状态代替当前状态

#### Scenario: 阶段输出真实改变后续模拟

- **GIVEN** 两个回合输入只在开局 Encounter 的已落地伤亡结果上不同
- **WHEN** 两个回合继续进入中期和包点阶段
- **THEN** 后续参战者、决策评分、可用行动或炸弹资格至少有一项由该状态差异重新计算
- **AND** 引擎不生成完全由预定胜方固定的相同后续过程

#### Scenario: 一个回合允许多次冲突和决策

- **GIVEN** T 方开局在 A 大受阻后仍有时间和可用转点路线
- **WHEN** 第一个 Encounter 结束并触发中期决策
- **THEN** 状态机可以生成转点行动、第二个 Encounter 和新的决策窗口
- **AND** 回合不被固定为一次 Opening、一次 MidRound 和一次 Site 调用

#### Scenario: RoundEnd 后不再推进

- **GIVEN** 事件批次已经产生合法 RoundTerminal
- **WHEN** 状态机进入 `RoundEnd`
- **THEN** 所有未完成比赛内 action 被失效或清空
- **AND** 后续只允许输出 ROUND_END 或比赛边界事件

#### Scenario: A 区交火不阻止 B 区开始下包

- **GIVEN** A 大存在仍在推进的 Encounter，且另一组不共享 Actor 的 T 已合法到达 B_SITE、持有炸弹并满足 Plant 前置条件
- **WHEN** B 区 action 在 A 大 Encounter 结束前到达自己的 ResolveAt
- **THEN** B 区可以进入 SiteContest 并启动 Plant action
- **AND** 主导 Phase 为 Clash 不得把该独立行动延迟到所有 ActiveEncounter 清空之后

### Requirement: 回合使用统一离散时间模型

`RoundState` SHALL 维护单调递增的 `Timeline`、绝对 `RoundDeadline` 和下包后建立的绝对 `BombDeadline`。resolver/planner MAY 计算 action `Duration`，但 SHALL NOT 直接推进时间或以整个阶段/Encounter 的 `TimeCost` 阻塞其他 action；scheduler SHALL 冻结 `ResolveAt = StartAt + Duration`，比较全部有效 action 与当前 deadline 后只推进到最早 `nextAt`。公开 `RoundTimer/BombTimer` SHALL 由 deadline 与 Timeline 投影，二者 SHALL NOT 同时作为胜负计时条件。

#### Scenario: 下包前推进消耗回合时间

- **GIVEN** 炸弹尚未下包，一次 Move 从第 4 秒开始、Duration 为 6 秒并形成 `ResolveAt=10`，且期间没有更早 action/deadline
- **WHEN** scheduler 选择下一个结算点
- **THEN** `Timeline` 推进到第 10 秒
- **AND** `RoundTimer = max(0, RoundDeadline - 10)`
- **AND** `BombTimer` 不参与终局判断

#### Scenario: 局部 Encounter 不吞掉中间 action

- **GIVEN** A 区 Encounter 的 CombatEnd.ResolveAt 为 15，但其中 CombatPulse 分别位于 9、12、15，B 区 Move.ResolveAt 为 10
- **WHEN** scheduler 按绝对时间选择 action
- **THEN** 先结算第 9 秒 pulse，再结算第 10 秒 Move
- **AND** 不得将 Timeline 从 EncounterStart 直接增加整个 CombatDuration 跳到第 15 秒

#### Scenario: 下包后切换炸弹计时

- **GIVEN** 下包者存活到下包完成时间
- **WHEN** `BOMB_PLANT` 状态变化生效
- **THEN** 炸弹进入 `Planted` 状态并设置 `BombDeadline = Timeline + BombExplodeTime`
- **AND** `BombTimer` 只投影为 `max(0, BombDeadline - Timeline)`，后续 scheduler 比较 Defuse/Combat action 与 BombDeadline
- **AND** `RoundTimer` 归零不再直接使 CT 获胜

### Requirement: EncounterResolver 结算真实伤害和击杀

每个 Encounter SHALL 从当前存活且可参战的选手中构造参与者，并通过有限 `CombatPulse` 计算命中、伤害、击杀、资源变化、Momentum、控制权和耗时。计算 SHALL 使用 `simuMatchDesign.md` 定义的选手属性、角色、武器、姿态、协同、Scenario、MapTag、EncounterModifier、UtilityBudget、HP、Stamina、Focus 和概率 clamp。系统 SHALL 记录团队级 `EncounterScore`，但该分数只影响主动权、姿态、Pulse 时长/数量、撤退倾向和解释，不得直接指定 Encounter 或回合胜者。

第一版 SHALL 只使用一套冻结的攻击/生存分公式：`CombatScore = PlayerCombatScore = WeightedPlayerScore + RoleTagModifier + WeaponModifier + PostureModifier + VisibilityModifier + TeamSupportModifier - StaminaPenalty - DamagePenalty - SuppressionPenalty`；`TargetSurvivalScore = Positioning + Reaction + CoverModifier + TeamSupportModifier + Focus - MovementExposure - DamagePenalty`。Utility、Momentum 和 TimePressure SHALL 只在团队级 EncounterScore 各计算一次，不得又直接加入 PlayerCombatScore；TargetSurvivalScore SHALL NOT 引入 `WeightedSurvivalAttributes` 或 `SuppressionPenalty`。

`KillChance` SHALL 使用 `RawKillChance = HitChance * DamagePotential * ExposureModifier` 以及 `clamp(min(RawKillChance, HitChance), 0, MaxKillChance)`；公式 SHALL NOT 加入 `TargetHPFactor`。低 HP 的死亡风险 SHALL 由实际 DamageEffect 与剩余 HP 自然体现。

#### Scenario: 武器和属性改变脉冲结果概率

- **GIVEN** 两个固定 seed 的 Encounter 具有相同场景和参与者身份，但攻击者武器伤害/距离修正或相关场景属性不同
- **WHEN** 计算该攻击者的 `CombatScore`、`HitChance` 和伤害潜力
- **THEN** 公式输入和计算结果反映这些武器与属性差异
- **AND** 差异记录在可测试的计算结果或 `EventReason` 中

#### Scenario: 命中未击杀仍更新状态

- **GIVEN** 一个战斗脉冲命中目标但造成的伤害小于目标当前 HP
- **WHEN** 应用脉冲结果
- **THEN** 目标 HP 降低但仍存活
- **AND** 攻击者伤害统计增加实际伤害值
- **AND** 目标 Focus、Stamina 或压制状态按配置规则更新
- **AND** 下一脉冲使用更新后的状态重新计算

#### Scenario: HP 归零自然产生击杀

- **GIVEN** 一个战斗脉冲的实际伤害使目标 HP 降至 0
- **WHEN** 应用脉冲结果
- **THEN** 目标先被标记死亡并退出后续可参战集合
- **AND** 引擎生成与攻击者、目标、武器、伤害和位置一致的 `KILL` 事件
- **AND** 该击杀不是通过目标所属方是否为预定败方决定

#### Scenario: 概率和资源严格 clamp

- **GIVEN** 极端高分差、低分差、受伤和资源输入
- **WHEN** 计算 HitChance、KillChance、HP、Stamina、Focus 和 Momentum
- **THEN** 所有值落在配置的最小/最大边界内
- **AND** 未配置为确定性的普通交火不会产生 0% 或 100% 的隐式必然结果

#### Scenario: 攻击分和生存分遵循唯一公式

- **GIVEN** 一个固定十属性、Role、Weapon、Posture、Visibility、TeamSupport、Stamina、Damage、Suppression、Focus、Cover 和 MovementExposure 的 golden case
- **WHEN** 计算 PlayerCombatScore 与 TargetSurvivalScore
- **THEN** 输出精确等于本 Requirement 冻结公式的结果
- **AND** Utility、Momentum、TimePressure 不在 PlayerCombatScore 中重复计分
- **AND** TargetSurvivalScore 不读取 WeightedSurvivalAttributes 或 SuppressionPenalty

#### Scenario: DecisiveScoreGap 只保护评分趋势

- **GIVEN** 双方不含 RandomNoise 的 DeterministicEncounterScore 分差达到 DecisiveScoreGap
- **WHEN** 引擎加入受 Discipline、IGL 和 Awareness 约束的有界 RandomNoise
- **THEN** RandomNoise 不得翻转双方 EncounterScore 的高低顺序
- **AND** 后续 HitRoll、DamageRoll 和 CombatPulse 仍正常采样并允许小概率爆冷
- **AND** 引擎不因明显分差预设 EncounterWinner、伤亡或回合 Winner

### Requirement: 中期决策由当前局势驱动

中期决策 SHALL 只使用当前存活与资源、剩余时间、本方已知控制权/情报、炸弹位置、路线受阻、角色匹配和比赛记忆，评分选择继续推进、转点、强攻、补防或包点执行。决策结果 SHALL 改变实际后续阶段输入或时间成本，而不只是生成叙事文本。

#### Scenario: 伤亡和时间改变 T 方选择

- **GIVEN** T 方开局路线受阻、持包者存活且 `RoundTimer` 仍充足
- **WHEN** AI 进行中期决策
- **THEN** 继续推进、转点和强攻评分读取实际伤亡、资源、控制权和时间
- **AND** 选中的行动更新路线/参战者/时间成本中的至少一项

#### Scenario: 决策不读取敌方隐藏真值

- **GIVEN** CT 的真实补防计划尚未被 T 方视野、交火、声音或控制权情报暴露
- **WHEN** T 方进行中期决策
- **THEN** 决策只读取 T 方已知信息及其置信度
- **AND** 不读取敌方完整真实计划或未来包点遭遇结果

### Requirement: 炸弹阶段由炸弹实体状态推进

`BombResolver` SHALL 根据当前存活、位置、交火、控制权、时间和炸弹实体状态结算下包、打断、掉落、拾取、拆包和爆炸。下包概率、拆包结果和爆炸结果 SHALL NOT 根据预定胜方修正。

#### Scenario: 下包者死亡会打断下包并掉包

- **GIVEN** 存活持包者已开始下包
- **WHEN** 同一结算时间内战斗伤害使下包者死亡
- **THEN** 击杀先于 `BombPlantComplete` 生效
- **AND** 下包失败并生成 `BOMB_DROP`
- **AND** 炸弹不会进入 `Planted` 状态

#### Scenario: 下包完成恰好等于回合 deadline

- **GIVEN** 存活下包者的 PlantCompleteAt 等于 RoundDeadline
- **WHEN** 状态机应用该时间戳完整批次
- **THEN** Combat/Death/BombDrop 先处理，合法 PlantComplete 随后处理
- **AND** 若下包者仍存活且 action 有效，Bomb 进入 Planted 并使 timeout 条件失效
- **AND** 若下包者同秒死亡，PlantComplete 失效并由最终 PrePlant 状态判定

#### Scenario: T 全灭但炸弹已下仍继续结算

- **GIVEN** 炸弹已经下包且所有 T 均死亡
- **WHEN** CT 仍有存活选手
- **THEN** 回合继续检查 CT 到达时间、拆包时间和 BombTimer
- **AND** 只有实际完成拆包或炸弹爆炸后才确定对应胜方

#### Scenario: 拆包完成与爆炸同秒

- **GIVEN** 存活拆包者的 `DefuseFinishAt` 等于 `Bomb.ExplodeAt`
- **WHEN** 状态机应用该同秒事件批次
- **THEN** `BombDefuseComplete` 先于 `BombExplode` 生效
- **AND** 炸弹状态为 `Defused`
- **AND** CT 由终局判定获得回合胜利

#### Scenario: 已下包回合可由 CT 全灭锁定

- **GIVEN** 炸弹已下包且最后一名 CT 被实际击杀
- **WHEN** 同秒不存在更高优先级的有效拆包完成事件
- **THEN** 终局状态可判定 T 以 `bomb_secured` 获胜
- **AND** 该结果来自 CT 存活状态与炸弹状态而非预定胜方

### Requirement: 同秒事件批次先更新状态再判胜

状态机 SHALL 对相同 `ResolveAt` 的行动按 `CombatPulseCommit`（同 pulse 全部 DamageEffect 原子提交后再派生 Death/KILL）、`CombatAftermath`、`BombDrop`、`BombPlantComplete`、`BombDefuseComplete`、`BombExplode`、`ControlChange`、`MovementArrive`、`Decision`、`IntelDecay` 的优先级以及稳定排序键完成整批应用，然后才执行全局终局判定。

#### Scenario: 最后一名 T 死亡与下包完成同秒

- **GIVEN** 最后一名 T 是下包者且其死亡与下包完成具有相同时间戳
- **WHEN** 应用同秒事件批次
- **THEN** 死亡和掉包先于下包完成
- **AND** 下包完成行动因选手死亡而失效
- **AND** CT 由未下包且 T 全灭的最终状态获胜

#### Scenario: 事件排序可复现

- **GIVEN** 多个事件具有相同时间戳和优先级
- **WHEN** 使用相同输入和 seed 模拟两次
- **THEN** ScheduledAction 按 `ResolveAt/Priority/ActionType/MinActorID/ActionID` 保持相同顺序
- **AND** 同一 action/Effect 派生的公开事件按 `Timestamp/Priority/ActionType/MinActorID/SourceActionID/SourceEffectID/EventID` 保持相同顺序

### Requirement: EventReason 解释实际计算过程

关键战斗、决策、炸弹和回合结束事件的 `EventReason` SHALL 来自 resolver 执行时记录的实际输入、主要因素、修正项、分差或概率以及状态变化。公开 DTO SHALL 向后兼容扩展现有 `GameEvent/EventReason`：`GameEvent` 保留稳定 `EventID/SourceActionID/SourceEffectID`；`EventReason` 保留 `Code/MainFactor/ScoreDelta(float64)/Detail` 并增加结构化 `Modifiers`、可空 `Probability`、`Formula`、`Inputs`、仅包含公开字段的 `StateChanges` 和 source IDs。StateChanges 的 Before/After SHALL 使用 Kind 标记且只设置一个对应的 Number/String/Bool/Null 标量，禁止不受控 map/object。解释器 SHALL NOT 截断 ScoreDelta、把实际 0 概率当成缺失、泄漏隐藏状态，或仅根据最终胜方反向猜测原因。

#### Scenario: KILL 原因对应实际脉冲

- **GIVEN** 某次脉冲在 CT 架点、长距离武器和补枪覆盖修正下产生击杀
- **WHEN** 输出 `KILL` 事件
- **THEN** `Reason` 包含实际参与评分的主要因素和修正项
- **AND** `ScoreDelta` 或公式输入与该脉冲计算记录一致
- **AND** `SourceActionID/SourceEffectID` 与产生致命 DamageEffect 的 AppliedBatch 一致

#### Scenario: 无概率事件与零概率值可以区分

- **GIVEN** 一个规则型终局 Reason 不适用概率，而另一个审计 fixture 记录实际 Probability=0
- **WHEN** 两者投影为公开 EventReason
- **THEN** 前者 Probability 为缺失/null，后者精确保留 0
- **AND** float64 ScoreDelta、结构化 Modifiers、Formula、Inputs 和公开 StateChanges 均不丢失

#### Scenario: ROUND_END 原因对应终局触发

- **GIVEN** BombTimer 在事件批次后归零且炸弹未被拆除
- **WHEN** 输出 `ROUND_END`
- **THEN** `Reason.MainFactor` 指向实际 `BOMB_EXPLODED` 状态变化
- **AND** 不包含未发生的淘汰、拆包或预定胜方叙事

### Requirement: 测试控制能力与生产因果路径隔离

对比分推进、换边和加时的测试 SHALL 使用独立规则状态机测试或测试注入的 round simulator；生产 `Service` SHALL 始终装配因果 round simulator。固定测试结果 SHALL 通过固定 seed、固定随机 roll、resolver fixture 或测试 fake 产生，且 SHALL NOT 使生产 Encounter/Bomb resolver 接收强制胜方。

#### Scenario: RPC 无法提交强制胜方

- **WHEN** 客户端调用现有模拟 RPC
- **THEN** 请求 DTO 和公开 `MatchInput` JSON 不接受逐回合胜方脚本
- **AND** 生产模拟只根据输入快照和 seed 运行因果管线

#### Scenario: 规则测试不生成伪造战斗事件

- **GIVEN** MR12 或加时测试需要固定一串回合得分
- **WHEN** 测试使用规则状态机或 fake round simulator
- **THEN** 该 fake 只验证比分、换边和比赛结束规则
- **AND** 因果战斗事件与终局不变量由生产 round simulator 的独立测试覆盖

### Requirement: 因果模拟支持批量标定

项目 SHALL 提供可重复的批量模拟测试或标定入口，基于固定配置版本和 seed 集合统计 T 回合胜率、平均击杀数、下包率、拆包率、爆炸率、首杀方胜率和平均回合时长。标定 SHALL 调整配置输入而不是写入目标胜方概率。

#### Scenario: 固定 seed 集合产生稳定指标

- **GIVEN** 固定阵容、配置版本、规则集和 seed 集合
- **WHEN** 连续运行两次批量标定
- **THEN** 两次聚合指标和逐 seed 终局完全一致
- **AND** 标定代码不向引擎传入期望胜方或目标终局分布

### Requirement: 离散事件调度器按绝对时间推进

RoundEngine SHALL 使用确定性的 `ScheduledAction` 队列推进 Move、Hold、Intercept、CombatPulse、Decision、Plant、Pickup、Defuse、RoundExpire、BombExplode 和情报/控制权衰减。队列 SHALL 按 `ResolveAt ASC / Priority DESC / ActionType ASC / MinActorID ASC / ActionID ASC` 排序，并在前进到下一时间点前比较当前有效 deadline。ActionID、EffectID 和 EventID SHALL 分别从稳定语义 tuple、稳定排序后分配的局部 ordinal 与 RoundSeed 派生，禁止使用全局自增或 map 遍历顺序；EventID SHALL 包含 SourceActionID、可空 SourceEffectID、EventType 和 EventOrdinal。

#### Scenario: 未参战者在局部交火期间继续行动

- **GIVEN** A 大 Encounter 从第 8 秒持续到第 15 秒，同时一名未参战 T 的 B 区 Move 在第 9 秒完成
- **WHEN** scheduler 按绝对时间推进
- **THEN** B 区 MovementArrive 在第 9 秒应用
- **AND** A 大 CombatPulse 继续在其各自时间点结算
- **AND** 引擎不等待整个 A 大 Encounter 完成后才处理 B 区行动

#### Scenario: 回合 deadline 早于普通行动

- **GIVEN** PrePlant 阶段 Timeline 为 100、RoundDeadline 为 103，队首 Move.ResolveAt 为 108
- **WHEN** scheduler 选择下一个结算点
- **THEN** 第 103 秒的 RoundExpire 先于 Move 结算
- **AND** 该 Move 不会越过回合时间后再到达

#### Scenario: 炸弹 deadline 不能遗漏

- **GIVEN** Bomb 已下且 BombDeadline 早于所有 CT 行动完成时间
- **WHEN** scheduler 选择下一批次
- **THEN** BombExplode 在 BombDeadline 进入队列并结算
- **AND** 不会因为 action queue 暂时为空或 CT 路径过长而跳过爆炸

### Requirement: 行动生命周期支持互斥、打断和版本失效

每名选手 SHALL 持有持续 Intent、至多一个互斥 ActionInstance、ActionVersion 和 BusyUntil。ScheduledAction SHALL 以 `VersionByActor` 冻结每个 Actor 安排时的版本。新 Move、Encounter、Plant、Defuse、Death 或重新决策 SHALL 按冲突规则失效旧 action；action 完成前 SHALL 逐 Actor 重新校验存活、CurrentActionID、版本、位置和实体状态。group action SHALL 同时冻结由来源 Route/模板人数约束派生的 `MinRequiredActors` 并验证 `1 <= MinRequiredActors <= len(ActorIDs)`；成员失效后，合法人数达到该值才由剩余成员继续，否则整体取消并让幸存者重新规划，不得使用隐藏阈值或临时随机。

#### Scenario: 移动途中进入 Encounter

- **GIVEN** 一名选手正在执行第 1 版 Move action
- **WHEN** 移动途中触发 Intercept 并进入 Encounter
- **THEN** 选手 ActionVersion 增加且状态变为 Engaged
- **AND** 原 MovementArrive 到时因版本不匹配被静默丢弃
- **AND** 不产生幽灵到达事件

#### Scenario: 下包被打断后旧完成事件失效

- **GIVEN** 持包者已安排 PlantComplete
- **WHEN** Planting 期间伤害或 Encounter 打断该 action
- **THEN** 持包者 action version 改变
- **AND** 原 PlantComplete 即使仍留在 heap 中也不能改变 BombState 或生成 BOMB_PLANT

#### Scenario: 同一选手不能并行完成互斥行动

- **GIVEN** 一名选手处于 Planting 或 Defusing BusyInterval
- **WHEN** planner 生成 Move、Rotate 或另一个 Plant/Defuse 候选
- **THEN** 候选被拒绝或显式替换旧 action
- **AND** 两个互斥完成事件不会同时生效

#### Scenario: group action 一名成员失效后满足人数约束

- **GIVEN** 一个稳定排序 ActorIDs、VersionByActor 和 MinRequiredActors=2 的三人同步 action
- **WHEN** 一名成员死亡或版本失效
- **THEN** 其余两名合法成员可以继续，失效成员不再产生完成效果

#### Scenario: group action 低于最小人数后整体取消

- **GIVEN** 一个 MinRequiredActors=2 的 group action 当前只剩两名合法成员
- **WHEN** 其中一名成员死亡或版本失效
- **THEN** 整个 group action 取消，唯一幸存成员回到 planner
- **AND** 重复模拟的继续/取消结果不受 map 遍历顺序影响

### Requirement: 同优先级战斗使用快照原子结算

同一时间戳、同一 CombatPulse 中所有合格攻击 SHALL 使用 pulse 开始时的不可变状态快照计算目标、命中和伤害，并作为一个不可拆分的 `CombatPulseCommit` 原子应用全部 lethal/non-lethal DamageEffect，再派生 Death/KILL。系统 SHALL NOT 在同一 pulse 内因为一个致命 Effect 先应用而取消另一个已经形成的非致命或致命攻击。CombatPulseCommit 完成后，Plant/Defuse/Movement 等外部低优先级 action SHALL 重新校验。

#### Scenario: 同脉冲自然互换

- **GIVEN** 两名互为目标的选手在同一 CombatPulse 开始时均存活且具备攻击资格
- **WHEN** 两个攻击结果都产生致命 DamageEffect
- **THEN** 两个 DamageEffect 都被应用
- **AND** 两名选手均由 HP 归零产生 KILL/Death
- **AND** ActorID 排序不会取消其中一人的已发生攻击

#### Scenario: 同脉冲致命攻击不取消非致命反击

- **GIVEN** A 和 B 在同一 CombatPulse 开始时都有攻击资格，A 对 B 产生致命伤害，B 对 A 产生非致命伤害
- **WHEN** 提交 CombatPulseCommit
- **THEN** A 和 B 的 DamageEffect 都被应用
- **AND** B 随后死亡但 A 仍承受 B 已经形成的非致命伤害
- **AND** lethal/non-lethal 事件分类或 EffectID 排序不得改变这一状态结果

#### Scenario: 多攻击者同脉冲伤害归属稳定

- **GIVEN** 两个攻击者在同 pulse 命中同一目标且总原始伤害超过目标当前 HP
- **WHEN** 原子应用这些 DamageEffect
- **THEN** 目标 HP 最低为 0，团队统计的实际伤害总和等于目标本次实际损失 HP
- **AND** 实际伤害按原始伤害占比及稳定余数规则分配
- **AND** KILL 归属最高实际伤害贡献者，同贡献按 EffectID 稳定决胜

#### Scenario: 高优先级死亡取消低优先级下包完成

- **GIVEN** 下包者死亡和 PlantComplete 具有相同 ResolveAt
- **WHEN** 状态机按 priority 带执行该批次
- **THEN** lethal combat、Death 和 BombDrop 先应用
- **AND** PlantComplete 在低优先级带重新校验时失效

### Requirement: Encounter 以可调度 CombatPulse 推进

EncounterStart SHALL 锁定实际参战 Actor，确定 Posture、CombatDuration 和 `1..MaxEncounterPulses`，并把 CombatPulse/CombatEnd 安排为未来 action。共享 Actor 或重叠区域的 Encounter 候选 SHALL 通过稳定 ThreatScore 仲裁；Actor 不重叠且区域不冲突的 Encounter SHALL 能够并行存在于调度队列。

#### Scenario: CombatPulse 之间状态反馈

- **GIVEN** Encounter 安排了三个 CombatPulse
- **WHEN** 第一个 pulse 造成受伤、Focus/Stamina 变化和 Momentum 变化
- **THEN** 第二个 pulse 使用更新后的 HP、Focus、Stamina、Suppression、Momentum 和存活参与者重新计算
- **AND** Encounter 不在开始时一次性预生成固定伤亡表

#### Scenario: 共享 Actor 的重叠 Encounter 不并发

- **GIVEN** 两个 Encounter 候选共享同一名选手
- **WHEN** planner 仲裁候选
- **THEN** 只安排 ThreatScore/稳定排序更高的 Encounter
- **AND** 该选手不会同时被两个 ActiveEngagement 锁定

### Requirement: 语义移动产生可打断的边上状态

Move SHALL 根据配置 MapEdge、Mobility、Stamina、Tempo 和队形计算耗时，并在执行期间维护唯一 `OnEdgeLocation`。每次 edge traversal 最多安排一个抽象 InterceptCheck；risk_points/intercept_nodes 只影响风险、时机和位置候选，不直接制造击杀。

#### Scenario: 边上拦截位置可复现

- **GIVEN** 一名选手从 LONG_DOOR 移向 A_LONG 且 InterceptCheck 在移动完成前触发
- **WHEN** 相同 RoundSeed、ActionID 和 edge 输入重复模拟
- **THEN** OnEdgeLocation.Progress、X、Y 和 DisplayName 完全一致
- **AND** 选手进入 Encounter 而不是先到达 A_LONG

#### Scenario: 风险热点不能直接杀人

- **GIVEN** MapEdge 配置了高权重 risk_point
- **WHEN** 本次 InterceptCheck 未形成合法双方 Encounter 或战斗采样没有致命伤害
- **THEN** 不生成 KILL 或 BOMB_DROP
- **AND** risk_point 只影响候选权重和 Reason

### Requirement: AI 决策只能读取本方情报

RoundState SHALL 分离 ActualControl、KnownControlT、KnownControlCT、TeamIntelT 和 TeamIntelCT。DecisionContext SHALL 只暴露本方合法公开状态、KnownControl、BombIntel 和带 confidence/TTL 的情报记录，不得暴露敌方隐藏 Intent/ActionQueue、未观察位置或全图 ActualControl。第一版队内情报 SHALL 即时共享，`CommunicationDelay` 配置 SHALL 存在且必须为 0；声音情报 SHALL 只来自实际声音事件并以区域、confidence 和 TTL 表达不确定性，不得生成虚假位置或虚假人数。死亡瞬间情报 confidence SHALL NOT 大于 70。

#### Scenario: 未观察补防不能被 T 方读取

- **GIVEN** CT 已真实安排两人转向 B，但没有视野、声音、交火或其他情报来源暴露该行动
- **WHEN** T 方进行中期决策
- **THEN** T 方 DecisionContext 不包含该真实补防 action
- **AND** 决策只能使用已有 KnownControl/TeamIntel 和随机扰动评分

#### Scenario: 低置信度情报只修正评分

- **GIVEN** CT 只有 confidence 20 的空点假设
- **WHEN** CT 评估全员补防
- **THEN** 该记录只作为有界 DecisionScore modifier
- **AND** 不直接触发确定性全员转点

#### Scenario: 情报到期后失效

- **GIVEN** 一条 KnownControl/Intel 记录的 ExpiresAt 已到
- **WHEN** 同秒更高优先级事件处理完成并执行 decay
- **THEN** 记录降级或删除
- **AND** 后续决策不再把它当作当前确认信息

#### Scenario: 声音情报不制造假敌人

- **GIVEN** 一次真实移动或炸弹 action 产生声音，但接收方没有直接视野
- **WHEN** 引擎生成声音 IntelRecord
- **THEN** 记录只包含实际来源对应的区域级位置、30-70 confidence 和 TTL
- **AND** 不生成不存在的敌人、虚假精确节点或虚假人数

#### Scenario: 死亡瞬间情报置信度封顶

- **GIVEN** 一名队员被实际 CombatPulse 击杀
- **WHEN** 其队伍获得攻击来源和大致人数情报
- **THEN** 该 IntelRecord.confidence 不大于 70

### Requirement: 中期 AI 覆盖自由人与截回防决策

T 方 planner SHALL 为具备 Lurker 标签且状态允许的选手评分 `GatherIntel`、`HoldFlank` 和 `InterceptRotate`，读取信息缺口、单走风险、已知控制权、路线与截回防价值。CT 方 SHALL 能依据本方 TeamIntel、KnownControl 和配置路径评分补防、重防与拦截转点，但 SHALL NOT 读取 T 方真实 Rotate action。所有中期 Decision RandomNoise SHALL 同时受 Discipline、IGL 和 Awareness 约束，并遵循 CloseScoreGap/DecisiveScoreGap 趋势保护。

#### Scenario: Lurker 行为产生实际行动和情报

- **GIVEN** T 方有存活 Lurker、主队正在另一条路线施压且某侧存在合法信息路线
- **WHEN** planner 选择 GatherIntel 或 HoldFlank
- **THEN** 生成实际 Move/Hold action 并消耗时间/资源
- **AND** 只有真实到达、视野、声音或 Encounter 才能产生相应 TeamIntel/截回防机会

#### Scenario: CT 不读取真实转点进行拦截

- **GIVEN** T 正在执行未被观察的 Rotate action
- **WHEN** CT 评分 InterceptRotate
- **THEN** CT 只能使用本方情报、已知控制、路径风险和有界噪声
- **AND** 未知的 T ActionID、目标节点和精确到达时间不进入 DecisionContext

### Requirement: NoOp 和调度上限不得伪造普通胜负

RoundEngine SHALL 通过包含 Timeline、位置、资源、Bomb、控制、Intel、Intent/ActionQueue、RecoveryAttempt 和事件的状态指纹检测真正 NoOp，并执行 MaxStateTransitions、MaxScheduledActions、MaxEffectsPerTimestamp、MaxNoOpTransitions、MaxRotationsPerTeam 和 MaxRoundTimeline 限制。连续 NoOp 达到 MaxNoOpTransitions 后 SHALL 以 `Hash(RoundSeed, "no_progress_recovery", RecoveryOrdinal)` 创建唯一 Recovery Cycle；RecoveryOrdinal 在单回合内单调递增且周期重置不回退。系统冻结 NoOpCount，并以 `NotAttempted -> Running -> Failed/Succeeded` 记录至多一次确定性 ForceExecute/恢复行动。由当前 RecoveryActionID 自身产生的排队、Timeline、移动、事件和完成 bookkeeping SHALL 保留该 Cycle；完成后恢复可达性则 Succeeded 并重置，仍不可达或不存在合法恢复行动则 Failed，保留失败证明、一次性设置 `NoProgressEligible=true` 并立即执行 NoProgressCheck。其他 action/effect 的真实进展或外部打断 SHALL 清空 Cycle、NoProgressEligible 和连续 NoOp 计数。终局纯函数 SHALL 仅在 `NoProgressEligible=true` 且合法业务状态仍满足 ValidNoProgress 时形成正式 `no_progress_timeout`；配置、图、planner、scheduler 或状态不变量错误 SHALL 返回结构化模拟错误，不得伪造 elimination/bomb/no-progress 事件。

#### Scenario: 空队列但存在 deadline 不是死锁

- **GIVEN** PostPlant 阶段 action queue 暂时为空但 BombDeadline 仍在未来
- **WHEN** 状态机检查可运行 action
- **THEN** scheduler 安排或推进到 BombExplode deadline
- **AND** 不以 NoOp 直接结束回合

#### Scenario: 超过调度硬上限返回错误

- **GIVEN** 错误 planner 持续生成 action 并超过 MaxScheduledActions
- **WHEN** guard 检测到上限
- **THEN** 返回稳定的 `SIMULATION_ERROR` 或更具体错误
- **AND** 不通过随机指定一方胜利来掩盖循环

#### Scenario: 合法不可达形成正式 NoProgress 终局

- **GIVEN** 地图配置和图引用有效、炸弹未下、T 仍有存活者，但任何存活 T 都无法到达掉落炸弹后再到包点，也没有持包者或其他 T 能到达任一可下包点
- **AND** 队列中没有能够恢复可达性的有效 action
- **WHEN** 连续 NoOp 达到 MaxNoOpTransitions、一次确定性 ForceExecute/恢复尝试不能推进，并由 NoProgressCheck 设置 NoProgressEligible 后，EvaluateRoundTerminal 检查 ValidNoProgress
- **THEN** CT 获胜
- **AND** `RoundResult.WinReason` 精确为 `no_progress_timeout`
- **AND** 不生成未发生的 elimination、timeout、plant 或 bomb 事件

#### Scenario: 恢复 action 自身进展不会清除失败证明

- **GIVEN** 连续 NoOp 已创建 Recovery Cycle 且 ForceExecute 进入 Running
- **WHEN** 该 RecoveryActionID 产生入队、Timeline 推进、移动和完成事件，但完成后仍不存在合法下包路径
- **THEN** Cycle 不因这些自身变化被通用 progress reset 清除
- **AND** Recovery 状态变为 Failed，随后一次 NoProgressCheck 可以读取该失败证明

#### Scenario: 外部真实进展重置恢复周期

- **GIVEN** Recovery Cycle 处于 Running
- **WHEN** 其他 SourceActionID/SourceEffectID 产生交火、Bomb、控制权或普通可执行 action 等真实进展
- **THEN** 当前 Cycle、NoProgressEligible 和 NoOpCount 被原子清空
- **AND** 不会沿用旧失败证明形成 no_progress_timeout

#### Scenario: 断图不是 NoProgress 回合结果

- **GIVEN** 不可达来自缺失 MapEdge、非法引用或其他未通过配置校验的问题
- **WHEN** 服务尝试启动模拟或状态机检测到该不变量错误
- **THEN** 返回结构化配置或 `SIMULATION_ERROR`
- **AND** 不返回 `no_progress_timeout` 或任何回合 Winner

### Requirement: 后端输出完整因果事件和 ExplainableReport

RoundEngine SHALL 输出 `ROUND_START`、`STRATEGY_ADJUSTED`、`DAMAGE`、`KILL`、`ROTATE`、`REINFORCE`、`CONTROL_GAINED`、`BOMB_DROP`、`BOMB_PICKUP`、`BOMB_PLANT_START`、`BOMB_PLANT_INTERRUPT`、`BOMB_PLANT`、`DEFUSE_START`、`DEFUSE_INTERRUPT`、`BOMB_DEFUSE`、`BOMB_EXPLODE` 和 `ROUND_END` 中实际发生的事件。MatchResult 和 RoundResult SHALL 包含 `ExplainableReport`，其 KeyEvents、StrategySummary、LossReasons 和 WinFactors SHALL 只由实际 action、已应用 Effect、Decision 和 terminal ReasonRecord 聚合。前端 MAY 只展示 KILL，但 RPC/后端结果 SHALL 保留完整事件和报告。

#### Scenario: 下包被打断保留过程证据

- **GIVEN** 持包者实际开始 Plant，随后在 PlantComplete 前被合法 Encounter 打断
- **WHEN** 构建回合事件和 ExplainableReport
- **THEN** 事件包含带实际 Reason 的 `BOMB_PLANT_START` 和 `BOMB_PLANT_INTERRUPT`
- **AND** 不包含成功 `BOMB_PLANT`
- **AND** Report.KeyEvents 只能引用这些实际输出事件

#### Scenario: 拆包完成事件遵循同秒存活条件

- **GIVEN** DefuseFinishAt 与 CombatPulse ResolveAt 相同
- **WHEN** 拆包者在 CombatPulseCommit 后死亡
- **THEN** 输出 `DEFUSE_INTERRUPT` 而不是 `BOMB_DEFUSE`
- **AND** 若只有其他 CT 同秒死亡而拆包者仍存活，则有效 DefuseComplete 可以输出 `BOMB_DEFUSE`

#### Scenario: 战术记忆调整可解释

- **GIVEN** PreviousSuccess、RepeatPenalty 或 CounterReadRisk 实际改变了模板评分
- **WHEN** 输出 ROUND_START 和 ExplainableReport
- **THEN** 输出 `STRATEGY_ADJUSTED`
- **AND** 在 ROUND_START.Reason 中保留同等完整的实际评分输入
- **AND** StrategySummary 不根据最终胜负反向改写模板选择原因

## MODIFIED Requirements

### Requirement: MatchEngine uses per-match seed for reproducibility

`MatchInput.Seed` SHALL be an immutable root seed, not a mutable random stream owned by `MatchEngine`. RoundEngine SHALL derive `RoundSeed` from root seed, map version, rule set ID and round number, then derive subsystem/action/pulse/actor seeds from stable identities. Production code SHALL NOT share or sequentially consume a mutable `*rand.Rand` across planners, actions, resolvers or events.

#### Scenario: Unrelated action does not shift existing combat rolls

- **GIVEN** two simulations have identical combat identities and root seed
- **WHEN** the second simulation schedules an unrelated non-conflicting action before an existing CombatPulse
- **THEN** existing ActorRollSeed values for that pulse remain unchanged
- **AND** the unrelated action does not shift target, hit, damage or location sampling by consuming a shared RNG

#### Scenario: Same complete input produces a deep-equal result

- **GIVEN** identical MatchInput, map/config snapshot, rule set and seed
- **WHEN** Service.Simulate is executed twice
- **THEN** the complete causal event ordering, RoundResult terminal states and final statistics are identical

### Requirement: Engine generates map coordinates for kill locations

`GameEvent.Location` SHALL remain a normalized structured `{ name, x, y }` value, but its semantic source SHALL come from the actual applied action/Encounter state: current MapNode geometry, runtime `OnEdgeLocation`, risk/intercept source selected for the action, or the documented node-coordinate fallback. RouteMain or a final winner SHALL NOT choose kill coordinates. Sampling SHALL use identity-derived `LocationSeed`.

#### Scenario: Kill location follows actual combat source

- **GIVEN** a player is killed during an Intercept Encounter while located on the edge from LONG_DOOR to A_LONG
- **WHEN** the KILL event is projected from the applied Death effect
- **THEN** Location.Name/X/Y come from that runtime OnEdgeLocation or its actual semantic source
- **AND** the engine does not substitute the round's RouteMain coordinate
- **AND** the same EventID, SourceObjectID and RoundSeed reproduce the same location

### Requirement: matchengine exposes Simulate entrypoint

`internal/framework/matchengine.Service` SHALL 提供 `Simulate(ctx, *MatchInput) (*MatchResult, error)` 作为业务使用的唯一正式入口，并按 `MatchInput.RuleSet` 模拟判定整场胜者所需的全部回合。单回合 resolver 或调试入口 SHALL NOT 绕过正式回合的因果状态推进。

#### Scenario: Simulate returns a complete causal result

- **GIVEN** 调用方构造了有效的 `MatchInput`，包含两支五人队伍、MapConfig、WeaponSpecs、RuleSet 和 Seed
- **WHEN** 调用 `Service.Simulate(ctx, input)`
- **THEN** 方法返回非 nil 的 `MatchResult`
- **AND** 结果包含规则集要求的全部回合、最终队伍比分、胜者队伍和最终统计
- **AND** 每个回合胜方均可由该回合实际事件与终局状态证明

### Requirement: Engine produces valid round result

`RoundResult` SHALL 包含回合编号、阶段、双方队伍与阵营、胜者、获胜原因、权威 `ScoreTeamA/ScoreTeamB`、兼容阵营投影 `ScoreT/ScoreCT`、战术/路线、事件列表、选手最终状态、炸弹最终状态和控制权快照。`Winner`、`WinnerTeamID` 和 `WinReason` SHALL 在终局状态确定后填写；SideSwitch SHALL 只改变阵营映射，比赛规则和 MatchResult 最终比分 SHALL 始终读取队伍身份比分。

#### Scenario: Round result fields follow terminal state

- **WHEN** 一回合因实际淘汰、超时、拆包、爆炸或已下包后 CT 无存活完成模拟
- **THEN** `RoundResult.Winner` 为 `"T"` 或 `"CT"`
- **AND** `WinnerTeamID` 匹配本回合阵营映射
- **AND** `WinReason` 与最终 Player/Bomb/Timer 状态一致
- **AND** `Events` 按时间戳和同秒优先级排序
- **AND** `ROUND_END` 后不再出现比赛内事件

#### Scenario: 换边只改变阵营比分投影

- **GIVEN** Team A 上半场结束时权威比分为 8，随后与 Team B 换边
- **WHEN** 构造下半场首个 RoundInput 和 RoundResult
- **THEN** ScoreTeamA 仍从 8 继续累计且 SideByTeam 已改变
- **AND** ScoreT/ScoreCT 分别投影当前 TeamTID/TeamCTID 的队伍比分
- **AND** SideSwitch 不交换或重置 ScoreTeamA/ScoreTeamB

### Requirement: 回合事件、公共状态与获胜原因保持一致

`matchengine` SHALL 先应用实际事件和状态变化，再从最终公共状态推导回合胜方及获胜原因。事件、选手状态、炸弹状态、计时器和获胜原因 SHALL 构成完整因果证明；仅让事件与预定结果表面相容不满足本要求。

#### Scenario: 淘汰终局由实际死亡触发

- **WHEN** 回合以 `elimination` 结束
- **THEN** 败方每名选手均已通过实际伤害/击杀事件死亡
- **AND** 最后一名败方选手死亡后才产生终局判定
- **AND** 未下包时败方全灭后不得继续产生下包事件

#### Scenario: 掉落炸弹后的 T 全灭保留精确内部原因

- **GIVEN** 炸弹未下、Bomb.Status 为 Dropped 且最后一名存活 T 由实际伤害死亡
- **WHEN** 事件批次完成后执行终局判定
- **THEN** CT 以公开 `elimination` 赢得回合
- **AND** terminal ReasonRecord 的内部 code 精确为 `T_ELIMINATED_BOMB_DROPPED`
- **AND** 不会先被一般 `T_ELIMINATED` 条件吞掉

#### Scenario: 超时终局由计时器触发

- **WHEN** 回合以 `timeout` 结束
- **THEN** 获胜阵营为 CT
- **AND** 炸弹从未成功下包
- **AND** `RoundTimer` 已由实际行动耗时推进至 0
- **AND** `ROUND_END` 时间戳对应统一 Timeline 的时间耗尽点

#### Scenario: 炸弹终局由状态机触发

- **WHEN** 回合以 `bomb_defused` 或 `bomb_exploded` 结束
- **THEN** 回合事件中存在实际成功的 `BOMB_PLANT`
- **AND** `BombState` 随后通过有效拆包行动进入 `Defused` 或通过 BombTimer 归零进入 `Exploded`
- **AND** `RoundResult.WinReason` 由该最终炸弹状态推导

#### Scenario: 终局字段不能反向影响过程

- **GIVEN** 回合尚未达到任何终局条件
- **WHEN** 生成下一批战斗、决策或炸弹事件
- **THEN** 尚未填写的 Winner 和 WinReason 不参与任何评分、概率或候选选择

### Requirement: 第一版实现遵循设计分层但限制细节深度

第一版完整比赛实现 SHALL 严格遵循 `doc/simuMatchDesign.md` 的数据边界和核心模拟阶段，包括战术模板选择、角色分配、开局遭遇战、状态驱动的中期决策、包点遭遇战、炸弹阶段、事后终局判定、可解释战报和比赛记忆。第一版 SHALL 实现本变更规定的轻量 ScheduledAction、ActionVersion、BusyInterval、语义 MapEdge、抽象 OnEdgeLocation 和最小 TeamIntel；不得再以“第一版简化”为由跳过这些要求。第一版不要求逐帧移动、物理导航网格、链式完整视野传播、非零 CommunicationDelay、虚假声音情报、网络同步或通用实时 ECS。

#### Scenario: 主要阶段真实参与结算

- **WHEN** 模拟一个包含开局接触、中期决策和包点阶段的回合
- **THEN** 战术与角色决定 Encounter 参与者和场景修正
- **AND** 开局伤害与伤亡进入中期决策输入
- **AND** 中期决策改变包点阶段的参与者、路线、控制权或时间
- **AND** 炸弹阶段只在当时状态满足条件时发生
- **AND** 最终胜方只在这些阶段完成后由终局状态确定

#### Scenario: 语义抽象保留必需状态

- **GIVEN** 第一版不实现逐帧物理移动和复杂通信网络
- **WHEN** 使用 Route、MapEdge、Duration/ResolveAt、OnEdgeLocation 和即时 TeamIntel 进行语义结算
- **THEN** 移动仍具有可查询进度、实际时间/体能消耗、打断和版本失效
- **AND** AI 仍只能读取本方带 confidence/TTL 的情报
- **AND** 不得用预定胜方、预定终局或目标存活人数替代任何缺失的物理细节

## REMOVED Requirements

### Requirement: Combat resolution uses simple comparison

**Reason**: 综合属性比大小后直接指定击杀属于早期 MVP 逻辑，无法表达 `simuMatchDesign.md` 要求的场景权重、武器、伤害、资源、协同和多脉冲因果结算。

**Migration**: 使用本变更规定的 `EncounterResolver` 与 `CombatPulse`，先计算命中和实际伤害，再由 HP 归零自然产生击杀。

#### Scenario: Higher stat player wins duel

- **GIVEN** 旧实现会让综合属性 250 的选手必然击杀综合属性 200 的选手
- **WHEN** 迁移到因果 Encounter 结算
- **THEN** 单次交火结果改由场景相关属性、武器、状态、协同、概率 clamp 和派生 seed 共同决定

### Requirement: Route configuration includes map positions

**Reason**: `matchengine` 内建路线坐标和 `RandomRouteLocation` 已被调用方提供的 `MapConfig`、MapNode/MapEdge/Scenario 语义及稳定 LocationSeed 取代。

**Migration**: 事件位置从当前 Encounter/行动的语义来源和 MapConfig 节点区域生成；引擎不再维护硬编码 Dust2 RouteConfig。

#### Scenario: Route has map position

- **GIVEN** 旧实现通过内建 `RouteConfig` 查询 A_Long 坐标
- **WHEN** 迁移后的引擎生成事件位置
- **THEN** 坐标来自输入地图快照中的节点、区域、边上位置或明确 fallback
- **AND** 相同 EventID 和 LocationSeed 产生相同位置
