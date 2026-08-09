## Context

`server/internal/framework/matchengine/engine.go` 当前把整场 MR12/换边/加时编排、回合模拟、事件生成、炸弹叙事和统计聚合集中在一个文件中。每回合先调用 `resolveRoundWinner`，再把 `winnerSide` 传给 `resolveRoundEvents`、`pickKillSide`、目标存活人数、下包概率和 `inferWinReason`。因此事件只与预定结果相容，不是结果的原因。

本设计以 `doc/simuMatchDesign.md` 为模拟语义的权威来源。现有 MR12、换边、加时、输入快照、地图配置边界、RPC 战报和 seed 派生可以保留，但回合内部必须替换为真正的离散事件状态机。归档后的主规格仍包含早期 MVP 的“综合属性比大小”“引擎内建路线”和“单回合输出”等冲突要求，本 change 的 delta specs 显式移除或改写这些要求。

当前 Dust2 生成数据也是占位深度：实际只有一条主要 T 路线、一个 RouteTemplate、一个 Scenario、一个 MapTag 和一个 EncounterModifier，且缺少设计文档列出的多条主路线、关键节点、CT 部署/回防路径和完整战斗参数。本变更必须同时完成第一版所需的配置数据，不能让新引擎退回隐藏 fallback 或只模拟 A 大单一路线。

### 文档歧义的工程化解释

以下内容可以实现，但必须先明确边界：

1. `3.3` 要求 `ScheduledAction`、打断、并发和同秒优先级，`4.11` 又把“类实时 Action 调度器”列为非目标。解释为：实现只服务于一次性 RPC 的轻量确定性事件队列、行动版本和互斥占用；不实现固定 tick、网络同步、逐帧动画、复杂行为树或通用实时 ECS。
2. `EncounterResolver.Start` 只确定 `CombatDuration` 并安排若干未来 `CombatPulse` 和 `CombatEnd` action；`ResolvePulse` 到时才读取最新不可变快照并返回 Effect。其他行动可以按绝对时间穿插，禁止在 Encounter 启动时预生成 `CombatEvents` 或固定伤亡表。
3. 文档同时要求 `KillChance` 和实际 HP/伤害。实现把 `KillChance` 定义为一次射击窗口形成“致命伤害包”的概率；无论致命还是非致命，都先生成实际 DamageEffect，只有应用后 HP 归零才产生死亡/KILL，禁止直接设置死亡结果。
4. `4.11` 不要求完整情报传播，但 `2.6` 禁止 AI 读取真实全图。实现最小但真实的 `TeamIntel`：直接视野、交火、声音、死亡和控制权记录，带 confidence/TTL，并在队内立即共享。保留 `CommunicationDelay` 配置字段，但第一版只接受 `0`，非零值返回配置错误；声音只生成真实来源的区域级低/中置信度记录，不生成虚假位置或虚假人数。
5. `4.11` 不要求完整 pathfinding，但炸弹拾取、回防和拦截需要可达性。实现仅在配置的语义 `MapNode/MapEdge/Route` 图上做确定性的有界路径选择，不构建物理导航网格或连续空间寻路。

这些解释没有改变文档的因果预期，也没有发现需要用户降低目标的阻塞项。

## Goals / Non-Goals

**Goals:**

- 严格实现“先模拟状态变化，后判定胜负”：战术、角色、移动、遭遇、伤害、中期决策、包点、下包、回防、拆包和爆炸逐步推进权威状态。
- 使用轻量离散事件队列表达不均匀时间、多行动并发、选手互斥占用、中途拦截、行动打断、同秒优先级和计时器 deadline。
- 让参战选手、属性、角色、武器、姿态、协同、UtilityBudget、HP、Stamina、Focus、Momentum、Scenario、MapTag、Visibility、控制权和 seed 共同决定每个 CombatPulse。
- 让 T/CT 决策只读取当前局势和本方情报，伤亡、剩余时间、控制权和 Bomb 状态必须实质改变后续行动。
- 完整实现第一版炸弹实体生命周期和所有设计边界，由纯终局函数从事件批次后的状态推导 Winner/WinReason。
- 保持 `MatchInput + MapVersion + RuleSet + Seed` 的逐事件确定性、稳定位置采样和可解释 Reason。
- 完成 Dust2 第一版主数据和可重复批量标定，使公式通过配置调整而不是通过目标胜方调整。

**Non-Goals:**

- 不模拟固定 tick、逐帧移动、碰撞、真实弹道、弹道穿透、完整物理视野或烟闪实体。
- 不实现客户端回合内指令、实时观战同步、Match Handler 循环或客户端预测；RPC 仍一次性计算完整比赛。
- 不实现完整经济系统、购买、掉枪、多地图、复杂换弹或独立守包小游戏。
- 不实现通用物理 pathfinding；只在配置语义图上选择已定义路线/边。
- 不把 Action 系统扩展成通用游戏框架；它只解决本模拟器的离散行动、占用、打断和排序。
- 不通过预抽胜方、预抽终局、目标存活人数、强制下包率或事后修补事件来满足统计目标。

## Decisions

### 1. 保留整场规则编排，替换回合模拟内核

`MatchEngine.StartMatch` 继续负责：输入验证、MatchState 初始化、常规赛回合循环、半场换边、加时 block、比分推进、比赛结束、最终统计与 MatchResult 汇总。它只消费 `RoundSimulationResult.Terminal` 给出的自然回合结果。内部 `MatchState.ScoreByTeam[TeamID]` 是唯一权威比分，RoundInput 获得该 map 的只读快照；公开 DTO 再固定投影为 `ScoreTeamA/ScoreTeamB`。SideSwitch 只修改 `SideByTeam`，不得交换或重置队伍比分。`ScoreT/ScoreCT` 仅在构造回合/兼容 DTO 时按当前 `TeamTID/TeamCTID` 投影，比赛早停、加时和 WinnerTeamID 均不得读取阵营累计分。

新的内部模块建议如下：

```text
matchengine/
  engine.go              整场 MR12/加时编排和结果汇总
  match_state.go         队伍比分、阵营映射、StrategyMemory
  round_engine.go        单回合离散事件循环
  round_state.go         权威 Round/Player/Bomb/Control/Intel 状态
  phase.go               阶段转换和重评估触发
  action.go              Intent、ActionInstance、Effect、BusyInterval
  scheduler.go           确定性优先队列、deadline、版本校验
  planner.go             按当前状态生成候选行动
  strategy.go            战术选择、角色分配、开局计划
  movement.go            语义图移动、边上位置、拦截候选
  decision.go            T/CT 中期决策
  intel.go               TeamIntel、KnownControl、TTL/置信度
  encounter.go           Encounter 生命周期和参战者锁定
  combat.go              CombatPulse、评分、目标、命中与伤害
  utility.go             回合 UtilityBudget 和局部消耗
  bomb.go                携带、掉落、拾取、下包、拆包、爆炸
  terminal.go            纯终局判定与不变量
  event.go               从已应用 Effect 生成事件和快照
  explainer.go           整理实际 ReasonRecord
  calibration.go         批量模拟和指标聚合
  model.go               对外输入/输出模型与共享配置快照
```

这些文件仍属于 `windypath.com/cs2match/server/internal/framework/matchengine`，不新增 Go module 或外部依赖。

**备选方案：**继续扩大现有 `engine.go`。拒绝原因是调度、状态、战斗和炸弹需要独立不变量与单测；继续集中会让胜方参数再次穿透多个阶段。

### 2. 使用两级状态机

#### 2.1 Match 状态机

Match 状态机只关心回合边界：

```text
ValidateInput
  -> InitializeMatchState
  -> BuildRoundInput
  -> SimulateRoundToTerminal
  -> ApplyNaturalRoundWinnerToTeamScore
  -> UpdateStrategyMemoryAndStats
  -> SideSwitch / Overtime / MatchEnd
```

Round simulator 返回前，Winner 字段不存在。Match 层不得把比分压力转换成直接胜率；比分只可以作为 StrategyScore 的可解释输入。

#### 2.2 Round 高层阶段

阶段用于解释战报、帮助 planner 选择当前主导活动和限定局部行动资格，不是各调用一次的流水线，也不是排他性的全局行动锁：

| Phase | 进入动作 | 允许的主要 action | 退出条件 |
|---|---|---|---|
| `OpeningDeploy` | 选择模板、角色和双方部署 | 生成 Intent、Move、Hold | 开局行动全部排入队列；不消耗 RoundTimer |
| `Advance` | T 推进、CT 持点/补位 | Move、Hold、InterceptCheck、Decision | 触发威胁进入 Clash；到达包点进入 SiteContest；关键变化进入 RotateDecision |
| `Clash` | 一个或多个互不重叠 Encounter 活跃 | CombatPulse、CombatEnd，以及不共享 Actor/区域的 Move、Decision、SiteContest、Plant 等合法行动 | 当前主导活动被更高投影优先级取代，或最后一个 Encounter 结束后重评估 |
| `RotateDecision` | 因伤亡、控制、情报、Bomb 或时间触发 | Decision、Move、Reinforce、Hold | 行动计划落地后回 Advance/SiteContest/PostPlant |
| `SiteContest` | T 到达目标包点或执行入口 | Site Encounter、Utility、Hold、PlantDecision | 包点威胁未清进入 Clash；允许下包进入 Planting；撤回进入 RotateDecision |
| `Planting` | 合法持包者启动下包 | CombatPulse、PlantComplete、BombDrop | 成功进入 PostPlant；打断/死亡进入 Clash 或 SiteContest |
| `PostPlant` | Bomb 已 Planted | Retake Move、Encounter、Defuse、BombExplode、Hold/Save | Defused、Exploded 或 BombSecured 后进入 RoundEnd |
| `RoundEnd` | `EvaluateRoundTerminal` 返回终局 | 只生成 ROUND_END/快照 | 禁止再调度比赛内 action |

同一时刻可能有多个局部行动；`RoundState.Phase` 表示根据当前 Bomb/Action/Encounter 状态派生的主导活动投影，`ActiveEngagements`、局部 ActivityStage 和每名选手的 ActionState 表达真实并发。Phase 不能作为“其他区域不得生成新行动”的白名单：例如 A 大参与者进行 CombatPulse 时，B 二楼未参战者不仅可以完成 Move，还可以在满足位置、Bomb、威胁和时间前置条件后进入 SiteContest、开始 Plant 或触发新的 Decision。某个 Encounter 未结束不得阻塞不共享 Actor/语义区域的合法行动。

主导 Phase 在每个时间戳批次后由权威状态重新投影，推荐优先级为 `RoundEnd > PostPlant > Planting > SiteContest > Clash > RotateDecision > Advance > OpeningDeploy`；该优先级只影响输出标签和默认 planner 上下文，不改变已经通过局部资格校验的 action。

### 3. 权威运行时状态与公开投影分离

公开 DTO 不直接承担模拟。内部状态至少如下：

```go
type RoundPlan struct {
    TStrategyTemplateID string
    CTSetupTemplateID   string
    RoleAssignments     []RoleAssignment
    OpeningRoutes       map[string]string // PlayerID -> RouteID；投影/哈希前稳定排序
    BombCarrierID       string
}

type NoProgressRecoveryState struct {
    CycleID          string
    Status           RecoveryStatus // NotAttempted/Running/Failed/Succeeded
    RecoveryActionID string
    StartedAt        int
    CompletedAt      int
    ResultCode       string
}
```

Encounter 不属于静态 RoundPlan；它由运行时实际位置、可见性、行动和拦截条件动态建立。

```go
type RoundState struct {
    RoundNumber int
    Phase       RoundPhase
    Timeline    int

    RoundDeadline int // 绝对时间；下包前有效
    BombDeadline  int // 绝对时间；下包后有效，未下包为 0

    TeamTID  string
    TeamCTID string
    Plan     RoundPlan

    Players map[string]*RoundPlayerState
    Bomb    BombState
    Nodes   map[string]*NodeRuntimeState
    Intel   map[string]*TeamIntel // side -> intel

    ActiveEngagements map[string]*EncounterState
    Scheduler         *ActionScheduler

    MomentumT  int
    MomentumCT int
    Utility    map[string]*TeamUtilityState

    DecisionCount  int
    RotationCount  map[string]int
    TransitionCount int
    NoOpCount       int
    NoProgressEligible bool
    RecoveryOrdinal int // 单回合单调递增，reset 不回退
    RecoveryAttempt NoProgressRecoveryState
    Events          []*GameEvent
    Terminal        *RoundTerminal
}
```

`NoProgressRecoveryState` 形成显式生命周期 `NotAttempted -> Running -> Failed/Succeeded`，用于证明一次确定性的 ForceExecute/RecoverBomb 恢复已经实际尝试。进入 Running 后冻结当前 NoOp 计数；由 `RecoveryActionID` 自身产生的入队、Timeline 推进、移动、事件和完成 bookkeeping 只更新当前 Cycle，不触发通用 reset。完成后统一重算可达性：若恢复合法下包路径或普通可执行 action，则记 Succeeded 后清空 `NoProgressEligible/NoOpCount/RecoveryAttempt` 并回到正常规划；若仍不可恢复，则记 Failed 并保留证明到紧随其后的 NoProgressCheck。来自其他 action/effect 的真实进展，或外部 Effect 打断恢复 action，才立即清空当前 Cycle、资格和计数。普通 planner 空转不得直接设置资格标记。

计时器在内部保存绝对 deadline，避免每个行动手工减时间：

```text
RoundTimer = max(0, RoundDeadline - Timeline)  // pre-plant
BombTimer  = max(0, BombDeadline - Timeline)  // post-plant
```

下包成功时冻结 RoundDeadline 的业务作用并设置 `BombDeadline = Timeline + BombExplodeTime`。公开 `RoundTimer/BombTimer` 如需输出，由当前 Timeline 投影。

```go
type RoundPlayerState struct {
    Profile PlayerProfile
    TeamID  string
    Side    string
    Weapon  WeaponLoadout

    Alive      bool
    HP         int
    Stamina    int
    Focus      int
    Suppressed bool
    Momentum   int

    Location PlayerLocation // Node XOR Edge
    Posture  CombatPosture
    Intent   Intent
    Action   PlayerActionState

    HasBomb bool
    Kills   int
    Deaths  int
    Damage  int
}
```

```go
type PlayerLocation struct {
    NodeID string
    Edge   *OnEdgeLocation
}
```

必须维护不变量：NodeID 与 Edge 只能一个有效；死亡玩家没有 Running action；Bomb.CarrierID 与玩家 HasBomb 唯一一致；RoundEnd 时队列不得再应用比赛内 action。

现有 `RoundResult`、`PlayerState`、`BombPublicState`、`FinalControls` 和关键事件快照只由内部状态投影，不能反向写回。

### 4. Intent、ActionInstance、BusyInterval 和版本失效

```go
type Intent struct {
    ID       string
    Type     IntentType
    TargetID string
    Priority int
    CreatedAt int
}

type ScheduledAction struct {
    ID          string
    IntentID    string
    Type        ActionType
    ActorIDs    []string
    From        PlayerLocation
    ToNodeID    string
    StartAt     int
    ResolveAt   int
    Priority    int
    VersionByActor map[string]int
    MinRequiredActors int
    Payload     ActionPayload
}

type PlayerActionState struct {
    CurrentActionID string
    Version         int
    Status          ActionStatus // Idle/Moving/Holding/Engaged/Planting/Defusing
    BusyUntil       int
}
```

Action 类型至少包括：

```text
MoveStart / MovementArrive / InterceptCheck
HoldStart
EncounterStart / CombatPulse / CombatEnd
DecisionResolve
PlantStart / PlantComplete
PickupComplete
DefuseStart / DefuseComplete
BombExplode
RoundExpire
IntelDecay / ControlDecay
```

开始新互斥 action 时，对每个 Actor 增加 `Version` 并记录当前 ActionID。行动结算前逐 Actor 校验：

```text
Alive
AND CurrentActionID == Action.ID
AND Version == Action.VersionByActor[ActorID]
AND Timeline == Action.ResolveAt
```

单人 action 校验失败时静默丢弃，不生成完成事件，不退还已消耗的时间/体能。group action 构造时冻结稳定排序的 ActorIDs、逐 Actor 版本，以及从来源 Route/模板人数约束派生的 `MinRequiredActors`（单人 action 为 1），并验证 `1 <= MinRequiredActors <= len(ActorIDs)`；违反该不变量的候选不得入队。成员失效后先移除无资格 Actor，合法成员数仍不少于该值才继续，否则整体取消并让幸存成员重新规划。队形完整度只记录 `合法成员数 / 原始 ActorIDs 数` 作为 Reason 输入，不再引入未配置阈值，也不得临时采样。死亡、Encounter 打断、转点重规划、Plant/Defuse 打断都通过增加对应 Actor 版本让旧资格失效。

`Hold` 是持续 Intent，不安排虚假的完成时间；它在被替换、死亡、移动或 Encounter 锁定时结束。Encounter 完成后只保留 Intent，planner 重新评分是否恢复，不自动恢复旧 ActionInstance。

**备选方案：**删除 action 版本，结算时只检查 Alive。拒绝原因是移动、下包、拆包和多人同步行动会在打断后残留幽灵完成事件。

### 5. 确定性 Scheduler、deadline 和主循环

使用标准库 `container/heap` 或等价稳定最小堆，不引入第三方依赖。队列排序键固定为：

```text
ResolveAt ASC
Priority DESC
ActionType ASC
MinActorID ASC
ActionID ASC
```

所有 ActorIDs 在构造 action 时排序。ActionID 由稳定上下文生成，不使用全局自增随机顺序，例如：

```text
Hash(RoundSeed, Type, IntentID, StartAt, ResolveAt, SortedActorIDs, LocalOrdinal)
```

Effect/Event ID 同样不使用全局自增值：resolver 先按 `EffectType/ActorID/TargetID/semantic payload` 稳定排序候选，再分配局部 ordinal。

```text
EffectID = Hash(ActionID, EffectType, ActorID, TargetID, EffectOrdinal)
EventID  = Hash(RoundSeed, "event", SourceActionID, SourceEffectID, EventType, EventOrdinal)
```

真实 action 生命周期事件的 SourceEffectID 为空，但 EventOrdinal 仍在该 SourceActionID 内按稳定语义排序分配。任何 map 遍历顺序都不得进入 ID。

每轮循环先比较队首 action 与当前有效 deadline：

```text
PrePlant:  RoundDeadline
PostPlant: BombDeadline
Intel:     earliest ExpiresAt
Control:   earliest TTL deadline
```

如果 deadline 更早或同秒，生成对应系统 action 放入同一批次。这样一个 8 秒 Move 不会跨过只剩 3 秒的 RoundDeadline 后才判超时。

```go
for state.Phase != PhaseRoundEnd {
    if err := guardLimits(state); err != nil { return nil, err }

    ensureRunnableActions(state) // 仅为空闲 Actor/必要系统 deadline 生成
    nextAt := scheduler.NextResolveAtWithDeadlines(state)
    if nextAt < state.Timeline { return SIMULATION_ERROR }

    state.AdvanceTime(nextAt - state.Timeline)
    batch := scheduler.PopAllAt(nextAt)
    batch = append(batch, dueDeadlineActions(state, nextAt)...)

    applied := executeTimestampBatch(state, batch)
    state.Events = append(state.Events, eventBuilder.FromApplied(applied)...)
    updateIntelControlAndDecisionTriggers(state, applied)

    if terminal := EvaluateRoundTerminal(state, applied); terminal != nil {
        enterRoundEnd(state, terminal)
        break
    }

    transition.Reevaluate(state, applied)
}
```

`AdvanceTime` 只移动 Timeline、更新边上行动的可查询 Progress、计算资源/情报到期候选；resolver 不得直接改计时器。

### 6. 同时间戳使用优先级分层事务

同秒不能简单逐 action 修改状态，否则 ActorID 排序会决定谁能开枪。批次按优先级带执行：

1. 取当前最高 priority 的所有 action。
2. 对这些 action 使用该 priority 带开始时的不可变 snapshot 做合法性检查与随机计算。
3. 产生 `Effect` 集合；同带 Effect 以稳定键排序后原子应用。
4. 应用派生的同秒 Effect（如死亡派生 BombDrop、打断、控制变化）。
5. 增加受影响 Actor 的 action version，然后重新校验下一低 priority 带。

```go
type Effect struct {
    ID        string
    Type      EffectType
    Priority  int
    ActorID   string
    TargetID  string
    Damage    int
    StateDelta StateDelta
    Reason    ReasonRecord
}
```

CombatPulse 是一个不可拆分的 priority-100 复合事务：同一 pulse 中所有有资格开火的选手先基于 pulse snapshot 选择目标并完成命中、致命窗口和伤害采样；随后一次性提交该 pulse 的全部 DamageEffect，再从提交后的 HP 派生 Death/KILL。已经在 pulse snapshot 中完成攻击资格判定的选手，即使被同 pulse 另一结果击杀，其已发生攻击仍有效；不得因为某个致命 Effect 先排序就取消另一个非致命 Effect。

若同一目标在同 pulse 被多个攻击者命中，目标 HP 只减少到 0，团队伤害统计只累计实际损失 HP。实际损失按各 DamageEffect 的原始伤害占比分配，余数按 `RawDamage DESC / EffectID ASC` 稳定分配；KILL 归属实际伤害贡献最高者，同贡献按 `EffectID ASC`。该排序只决定统计和事件归属，不改变目标是否死亡。

不同带严格执行文档优先级：

| Priority | Effect/Action |
|---:|---|
| 100 | CombatPulseCommit（同 pulse 全部 lethal/non-lethal Damage）/ derived Death / Kill |
| 90 | CombatAftermath（Suppression / Focus / Stamina / Momentum / Interrupt）；不得撤销已提交的同 pulse Damage |
| 80 | BombDrop |
| 70 | PlantComplete |
| 60 | DefuseComplete |
| 50 | BombExplode |
| 40 | ControlChange |
| 30 | MovementArrive |
| 20 | DecisionResolve |
| 10 | Intel/Control decay |

CombatPulseCommit 完成后增加受影响 Actor 的 action version，再重新校验 Plant/Defuse/Movement 等外部低优先级 action。Plant/Defuse complete effect 在应用时再次检查 Actor 存活、action version、Bomb 状态和位置；高优先级死亡或伤害打断后自然失效。`DefuseComplete == BombExplode` 时 Defuse 先应用并让爆炸 action 失效。若拆包者在 `DefuseFinishAt` 同秒死亡，CombatPulseCommit 先使其死亡，DefuseComplete 随后失效；只有拆包者仍存活时等号才判给 CT。

`RoundExpire` 是同时间戳终局检查 marker，不在 PlantComplete 之前修改 BombState。完整批次应用后才运行 Terminal：因此 `PlantCompleteAt == RoundDeadline` 时，合法 PlantComplete 先把 Bomb 置为 Planted，timeout 条件随即失效；若下包者同秒死亡，则高优先级死亡/掉包让 PlantComplete 失效，Terminal 再以 T 全灭或 timeout 判定。

### 7. 并发、Actor 锁定和 Engagement 仲裁

互不共享 Actor 的 Move、Hold、Decision、Encounter 可以并发存在。一个 Actor 在任一时刻只能有一个互斥 ActionInstance。多人同步拉枪使用一个 group action，ActorIDs 稳定排序。

`ActiveEngagements` 可以同时保存地图不同区域、Actor 不重叠的 Encounter。若候选 Encounter 共享 Actor 或语义区域重叠，则按以下稳定顺序选一个：

```text
ThreatScore DESC
Scenario.BaseWeight DESC
EarliestContactAt ASC
EncounterID ASC
```

未选候选不会制造事件；相关 Actor 保持原 Intent，下一次状态变化后重评估。这样避免通过 Visibility 链把半张地图合并，同时允许 A 大交火期间 B 区推进继续。

### 8. 语义移动、边上位置和拦截

开局计划为每名选手或小队选择配置 Route/节点序列。移动时间：

```text
MoveTime = clamp(
    Edge.BaseTime
  * TempoFactor
  + FormationPenalty
  + StaminaPenalty
  - MobilityModifier,
  MinMoveTime,
  MaxMoveTime)
```

MoveStart 后 Actor 进入 Moving/Busy，保存边、StartAt、ResolveAt。任意 Timeline 可计算：

```text
Progress = clamp((Timeline - StartAt) / (ResolveAt - StartAt), 0, 1)
```

运行时 `OnEdgeLocation.X/Y` 由 From/To 锚点插值，再用 `LocationSeed` 在关联 risk_points/intercept_nodes 的合法区域内做有界偏移。配置热点只改变暴露/拦截权重和位置候选，不直接制造死亡。

每条 edge 至多为本次 traversal 安排一个抽象 `InterceptCheck`：其时间和概率来自 Edge.Risk、Noise、Visibility、双方姿态、KnownIntel 和派生 seed。触发后取消/暂停相关 Move，建立 Encounter；未触发则 MovementArrive 正常完成。第一版不逐厘米扫描路径。

转点在语义图上选择配置路线或确定性的最短合法 Edge 序列；仅以节点/边权重进行有界搜索，不做物理导航。不可达 action 被拒绝并触发新 Decision；Bomb/包点全部不可达时进入真实 NoOp/timeout 路径。

### 9. 战术、角色、比赛记忆和开局部署

T StrategyScore：

```text
TemplateBaseWeight
+ LineupFitScore
+ CurrentScorePressure
+ PreviousSuccessBonus
- RepeatPenalty
- CounterReadRisk
+ BoundedRandomNoise
```

候选模板必须按 ID 稳定排序后采样。明显分差超过 `DecisiveScoreGap` 时随机只能改变细节，不能覆盖主要排序趋势。

这里的“明显分差保护”只约束 `RandomNoise`：先计算不含噪声的确定性候选分；若领先候选与次选候选的差达到 `DecisiveScoreGap`，则把噪声裁剪到不会翻转两者顺序的范围。它不锁定后续 Encounter、命中、伤害或回合胜方，实际 CombatPulse 仍完整采样并允许小概率爆冷。近似同分由 `CloseScoreGap` 判定。

角色分配使用模板 RequiredRoles、角色标签、场景属性适配和稳定最大评分匹配；不满足角色时仍可生成计划，但输出缺口 penalty/Reason。Bomb carrier 默认从非 Entry、Composure/Discipline 合适且路线可达的 T 中选择，选择结果是计划的一部分，不由回合终局决定。

`RouteTemplate` 同时承载 `side=T` 战术模板和 `side=CT` setup 模板，并显式保存同阵营 `route_ids/route_allocations`。T 模板可用 `common_ct_setup_ids` 表达赛前地图先验，但它不指定本回合 CT 实际 setup。CT planner 只根据本方阵容/角色/属性、配置权重、历史 StrategyMemory 和独立 `CTSetupSeed` 选择 setup/hold routes；输入中不得出现本回合 T StrategyTemplateID、T Route/Intent 或其他隐藏状态。模板内 route_allocations 总人数必须等于队伍人数，且逐路线满足 min/max 和从 CT_SPAWN 的可达性。

T OpeningPlan 与 CT setup 各自以 AttemptOrdinal 0 生成、最多以 AttemptOrdinal 1 重试一次。T 失败只回退 `DefaultStrategyTemplateID`，CT 失败只回退 `DefaultCTSetupTemplateID`；默认模板必须存在、阵营正确、引用闭合且可达，否则模拟返回结构化错误。不能从 T 的 `common_ct_setup_ids` 强制选择实际 CT setup，也不能硬编码默认站位。开局部署不消耗 RoundTimer，部署完成后双方 Move/Hold action 从 Timeline 0 排入队列。

StrategyMemory 只进入下一局模板权重、CT setup 和初始补防倾向；不得修改 Aim/Reaction 等基础属性，也不得直接写入 Encounter 胜率。半场换边衰减阵营相关记忆，保留队伍风格。

### 10. TeamIntel、控制权与 AI 读取边界

内部同时保存：

```text
ActualControl
KnownControlT
KnownControlCT
TeamIntelT
TeamIntelCT
```

AI DecisionContext 只能暴露本方 KnownControl/TeamIntel、公共计时器、本方真实状态、已公开死亡和合法 BombIntel；不暴露敌方 Intent、ActionQueue、未观察位置或 ActualControl。

最小情报来源：

- 直接 Visibility 接触：90-100 confidence；
- Encounter/受伤来源：71-89；死亡瞬间来源和大致人数不超过 70；
- 真实声音、移动暴露、下包/掉包：30-70；
- 空点假设和历史控制：1-29。

记录保存 `ObservedBy/LastSeenAt/ExpiresAt`。队内第一版立即共享；`CommunicationDelay` 作为预留配置字段存在但必须等于 0，非零值由配置校验拒绝。声音只根据实际发生的声音来源生成区域级记录，不伪造位置或人数；不确定性只通过 30-70 confidence 和 TTL 表达。死亡观察者使对应记录 confidence 下降；到期后在 priority 10 衰减/删除。低 confidence 只能修正 DecisionScore，不能直接触发确定性全员补防。

控制权由实际局部存活、撤退和观察结果更新；Contested 在 Encounter 结束后必须转为某方控制或 Unknown。KnownControl 在无人观察超过 `ControlIntelTTL` 后退化，不能永久提供全图信息。

### 11. 状态驱动的中期 Decision

Decision 不固定只发生一次。以下事件设置 `DecisionTrigger`：

- Encounter 结束；
- 人数或关键 HP/Focus/Stamina 档位变化；
- 关键点控制变化或发现空点；
- Bomb 掉落/拾取/下包；
- Route blocked/不可达；
- RoundTimer 首次跨过 ForceExecuteThreshold；
- PostPlant 回防到达或 Defuse 被打断。

planner 对空闲或可被替换 Intent 的 Actor 生成：Continue、Rotate、Reinforce、Hold、ForceExecute、RecoverBomb、Plant、Retake、Defuse、DenyDefuse、Save，以及由 Lurker 执行的 GatherIntel、HoldFlank、InterceptRotate。CT 可以根据本方情报和配置路径生成 InterceptRotate，但不能读取 T 的真实转点 action。每个候选输出 DecisionScore 和 ReasonRecord。

中期 `DecisionScore` 的 RandomNoise 同时受队伍 IGL、执行者/相关队员的 Discipline 和 Awareness 约束，并使用 `CloseScoreGap/DecisiveScoreGap` 的同一趋势保护规则；不能只读取 Discipline，也不能把趋势保护变成预定行动结果。

Decision 有 `DecisionDelay`，不是瞬时；选择后生成实际 Move/Hold/Plant 等 action。`MaxDecisionCount` 和 `MaxRotationsPerTeam` 限制循环，但达到上限时只限制继续转点并强制基于当前可达路线执行，不能直接指定胜负。

### 12. Encounter 生命周期和 CombatPulse

#### 12.1 EncounterStart

Encounter 只能在双方存在存活、可参战且未被其他 Engagement 锁定的 Actor，并且 Scenario/当前接触允许时启动。启动后：

1. 锁定参与 Actor，打断冲突的 Move/Plant/Defuse action；
2. 根据姿态、Visibility、Scenario、MapTag、UtilityContext 和角色确定初始 Posture；
3. 计算 CombatDuration 和 `1..MaxEncounterPulses` 的 pulse 时间；
4. 安排绝对时间的 CombatPulse 与 CombatEnd；
5. 未参战 Actor 保持原 action。

#### 12.2 EncounterScore 与趋势保护

每个 pulse 先记录可解释的双方团队级 EncounterScore：

```text
EncounterScore =
    Sum(PlayerCombatScore * RoleWeight)
  + TeamModifier
  + ScenarioModifier
  + UtilityModifier
  + MomentumModifier
  + TimePressureModifier
  + BoundedRandomNoise
```

不含 `BoundedRandomNoise` 的分数称为 DeterministicEncounterScore。若双方确定性分差达到 `DecisiveScoreGap`，只裁剪 RandomNoise 使其不能翻转 EncounterScore 的高低关系。EncounterScore 影响初始姿态、主动权、Pulse 数/时长、撤退倾向及 Reason，但不直接设置局部胜者、伤亡或回合终局；每一次命中和实际伤害仍由 CombatPulse 采样。因此明显优势方更稳定，但仍允许由实际概率链产生爆冷。

#### 12.3 PlayerCombatScore

```text
WeightedPlayerScore = Σ(Attribute_i * ScenarioFamilyWeight_i)

PlayerCombatScore =
    WeightedPlayerScore
  + RoleTagModifier
  + WeaponModifier
  + PostureModifier
  + VisibilityModifier
  + TeamSupportModifier
  - StaminaPenalty
  - DamagePenalty
  - SuppressionPenalty
```

`CombatScore = PlayerCombatScore`，这是第一版唯一的 pulse 攻击分公式。Utility、Momentum 和 TimePressure 只在团队级 EncounterScore 各计算一次，用于主动权、姿态、脉冲数量/时长和撤退倾向；不得再直接加入 PlayerCombatScore。具体道具或节奏造成的可观察姿态、视野、协同变化可以通过对应单项 modifier 间接影响 pulse，但同一原因不得跨层重复计分。团队修正不直接产生击杀；TradeCoverage、Crossfire、Spacing、SyncPeek、Isolation 均从本 Encounter 的实际参与者、位置和姿态计算。

```text
TargetSurvivalScore =
    Positioning
  + Reaction
  + CoverModifier
  + TeamSupportModifier
  + Focus
  - MovementExposure
  - DamagePenalty
```

这是第一版唯一冻结的生存分公式。Suppression 已在攻击方 PlayerCombatScore 中生效，不得又从 TargetSurvivalScore 扣除。

#### 12.4 目标选择

每个 eligible Actor 每个 pulse 最多形成一次攻击窗口。候选 Target 必须存活、处于当前 Engagement、存在 ActiveThreat/Scenario 接触且不受不可见规则排除。目标排序：

```text
ThreatScore DESC
Exposure DESC
Distance ASC
TargetHP ASC
PlayerID ASC
```

近似同分时用 Actor/Pulse/Target 派生 seed 做有界选择；不得遍历 Go map 后直接随机。

#### 12.5 命中、致命窗口和实际伤害

```text
RawHitChance = sigmoid((CombatScore - TargetSurvivalScore) / CombatScale)
HitChance = clamp(RawHitChance, MinHitChance, MaxHitChance)

BaseShotDamage = Weapon.Damage
    * RangeModifierForScenario
    * ArmorFactor(ArmorPenetration, target armor)

BurstCapacity = clamp(
    floor(Weapon.RoundsPerMinute / 60 * PulseFireWindow),
    1,
    Weapon.MagazineSize)

DamagePotential = clamp(
    BaseShotDamage * BurstPotential(BurstCapacity) / MaxHP,
    MinDamagePotential,
    MaxDamagePotential)

RawKillChance = HitChance * DamagePotential * ExposureModifier
KillChance = clamp(min(RawKillChance, HitChance), 0, MaxKillChance)
```

使用独立派生 roll，避免额外随机调用改变其他 Actor：

```text
HitRoll   = HashRand(PulseSeed, ActorID, TargetID, "hit")
LethalRoll= HashRand(PulseSeed, ActorID, TargetID, "lethal")
DamageRoll= HashRand(PulseSeed, ActorID, TargetID, "damage")
```

流程：

1. `HitRoll >= HitChance`：MissEffect，消耗少量 Focus/Stamina；
2. 命中后用 `KillChance / HitChance` 作为条件致命窗口概率；
3. 非致命窗口根据 BaseShotDamage、BurstCapacity 和 DamageRoll 生成实际 damage；
4. 致命窗口生成足以覆盖目标当前 HP 的实际 damage package，但统计伤害只记录目标实际损失 HP，Reason 保留武器/暴露/致命概率；`KillChance` 不乘 `TargetHPFactor`，低 HP 的易死亡效果由实际伤害自然产生；
5. 统一应用 DamageEffect，`HP = clamp(HP-damage, 0, MaxHP)`；只有 HP 变 0 才派生 Death/KILL、BombDrop、Momentum 和控制变化。

这仍是“少模拟、多结算”：KillChance 表示一个 pulse 内形成致命射击/爆发的概率，不模拟每颗子弹；但每个死亡都有实际伤害、武器、Actor、Target、位置和 pulse 作为原因。

#### 12.6 Pulse 后更新与结束

应用 pulse 后更新 Focus、Stamina、Suppression、Momentum、TradeCoverage 和 Posture。首杀只是本回合第一条自然 KILL。低 HP/低 Focus/低 Stamina 会进入下一次 Decision 和 CombatScore。

Encounter 在局部一方全灭、撤退完成、失去有效接触、达到 MaxEncounterPulses 或 CombatDuration 后结束。CombatEnd 更新实际/已知控制权、释放 Actor 锁并触发 Decision；它不设置回合 Winner。

### 13. 有限 UtilityBudget

每队每回合初始化有限预算；预算可按队伍平均 Utility、Support 标签、模板和携带 Grenades 计算，但所有上限来自配置。使用动作显式从预算扣除，并记录作用域：

```text
VisibilitySuppression
OpeningInitiative
ExposureReduction
SyncPeekQuality
PlantCover
DefuseCover
```

Utility modifier 只对指定 Encounter/action 和持续窗口有效；不能跨所有 Encounter 重复使用，也不能全局直接加 TeamPower。预算不足时 modifier 为 0，并在 Reason 中标明 LOW_UTILITY。

### 14. Bomb 实体和动作状态机

```text
Carried
  -> Dropped -> Pickup -> Carried
  -> Planting -> Planted -> Defusing -> Defused
                         -> Exploded
```

内部额外保存 Planting/Defusing 的 Actor/ActionID/StartAt/FinishAt；公开状态可以保持现有枚举兼容并通过事件表达进行中状态。

#### 14.1 Plant

CanAttemptPlant 必须检查：存活 T、唯一 Carrier、位于 Site、非 Engaged、Action 可替换、Bomb=Carried、Timeline 不晚于 RoundDeadline。PlantScore 决定 AI 是否承担风险，不绕过资格。

PlantStart 设置 Actor=Planting、Bomb=Planting 并安排 PlantComplete。可见 CT 威胁或拦截 action 可在完成前启动 Encounter。死亡/伤害打断增加 action version；持包者死亡派生 BombDrop。

`PlantCompleteAt == RoundDeadline` 时两者同批：Combat Kill/Damage/BombDrop 先处理，存活且 action 仍有效的 PlantComplete priority 70 再处理，然后 RoundExpire 终局检查发现 Bomb 已 Planted 而失效。

#### 14.2 Drop/Pickup

持包者死亡后 Bomb 保存其 Node/OnEdgeLocation。Pickup 只能由存活 T 通过实际 Move 到达后开始。CTControlled/Contested 点先安排 Encounter；不可达时 RecoverBomb 候选失败并进入强攻/timeout，而不是瞬移拾取。

#### 14.3 PostPlant/Retake/Defuse

PlantComplete 设置 PlantedAt、BombDeadline/ExplodeAt 并安排不可遗漏的 BombExplode action。CT 根据 Known BombIntel 和语义路径安排 Retake；到达后若仍有 T 可在 DefuseFinishAt 前形成威胁，先安排 DenyDefuse Encounter。

CanDefuse 检查存活、位置、Focus/Composure、HasKit 修正、`DefuseFinishAt <= BombDeadline` 和当前威胁。DefuseStart/Complete 可被 Kill/Damage/Encounter 打断。`DefuseComplete == BombDeadline` 时 Defuse priority 60 先于 BombExplode priority 50。

T 全灭但 Bomb 已 Planted 不结束回合；CT 必须实际到达并拆除，否则爆炸。CT 全灭且没有同秒有效 DefuseComplete 后可判 BombSecured。

### 15. 纯终局判定和阶段不变量

`EvaluateRoundTerminal` 不接受 winner、scoreDelta、目标终局或 RNG，只读取事件批次应用后的状态：

```go
func EvaluateRoundTerminal(state *RoundState, applied AppliedBatch) (*RoundTerminal, error)
```

优先级：

1. Bomb=Defused → CT / `bomb_defused`；
2. Bomb=Exploded → T / `bomb_exploded`；
3. PrePlant 且 TAlive=0 → CT / `elimination`（内部 code `T_ELIMINATED` 或 `T_ELIMINATED_BOMB_DROPPED`）；
4. PrePlant 且 CTAlive=0 → T / `elimination`（内部 code `CT_ELIMINATED`）；
5. PrePlant 且 Timeline >= RoundDeadline → CT / `timeout`；
6. PostPlant 且 CTAlive=0 且无同秒有效 DefuseComplete → T / `bomb_secured`；
7. PrePlant、`NoProgressEligible=true` 且满足 ValidNoProgress → CT / `no_progress_timeout`；
8. 调度上限、配置或状态不变量错误返回结构化 simulation error，不产生回合 Winner。

```text
ValidNoProgress =
    ConfigAndGraphValid
AND Bomb.Status != Planted
AND TAlive > 0
AND (
       (Bomb is Carried AND carrier cannot reach any plant site)
    OR (Bomb is Dropped AND no alive T can reach bomb then any plant site)
    )
AND no pending valid action can restore reachability
```

`ValidNoProgress` 只是业务状态条件，不能在普通批次后自动终局。只有连续 NoOp 达到 `MaxNoOpTransitions`，并且同一 `RecoveryAttempt.CycleID` 已记录一次确定性的 ForceExecute/恢复行动为 Failed 或确认不存在合法恢复 action 后，才设置一次性的 `NoProgressEligible=true` 并执行显式 `NoProgressCheck`。`EvaluateRoundTerminal` 仍是纯函数，只读取资格标记和当前状态。恢复 action 自身的生命周期变化不会提前销毁失败证明；恢复成功、其他 action/effect 的真实进展或外部打断才重置该 Cycle、资格和连续 NoOp 计数。

`no_progress_timeout` 是正式公开 `RoundResult.WinReason`，不是普通 `timeout` 的内部别名，也不能用于掩盖未实现 planner、缺表、断图或 scheduler bug。

PostPlant 的 TAlive=0 不是终局。双方同秒死亡时依炸弹阶段和上述优先级决定。每个 terminal 执行不变量验证；不一致返回 `SIMULATION_INVARIANT_ERROR`，禁止修补事件或重新采样。

RoundEnd 后清空/失效所有 action，生成一次 ROUND_END；之后只能由 Match 层追加 MATCH_END 等边界事件。

### 16. NoOp 和硬终止限制

使用以下上限，全部来自已校验 CombatConstants；RuleSet 只处理整场赛制，不覆盖回合参数：

```text
MaxStateTransitions
MaxScheduledActions
MaxEffectsPerTimestamp
MaxNoOpTransitions
MaxRotationsPerTeam
MaxDecisionCount
MaxEncounterPulses
MaxRoundTimeline
```

NoOp 必须以状态指纹判断：Timeline、位置、HP/Stamina/Focus、Bomb、控制、Intel、action queue、恢复状态和事件均无变化才计数。队列为空但存在未来有效 deadline 不算死锁，调度 deadline 即可。通用 reset 不能仅比较前后指纹，还必须读取 AppliedBatch 的 SourceActionID/SourceEffectID：当前 RecoveryActionID 自身的变更只推进恢复生命周期；其他来源的真实进展才结束当前恢复周期。

连续 NoOp 达到 `MaxNoOpTransitions` 后：以 `Hash(RoundSeed, "no_progress_recovery", RecoveryOrdinal)` 创建唯一 CycleID 并递增不回退的单回合 RecoveryOrdinal、冻结 NoOpCount，并在 Status=Running 期间至多生成一次确定性 ForceExecute/RecoverBomb；Bomb 已下则推进最近 CT 行动或 BombDeadline，不进入 PrePlant NoProgress 终局。恢复 action 完成后统一重算可达性：成功则清空 Cycle 回归正常规划；失败或确认不存在合法恢复 action 则保留 Failed/ResultCode，设置一次性资格并立即执行 `NoProgressCheck`。仅当配置和图合法、`NoProgressEligible=true` 且 ValidNoProgress 仍为真时形成正式 `no_progress_timeout`。配置损坏、图引用异常、planner/scheduler 超限或状态不变量错误返回结构化 simulation error。达到 guard 不能“按当前局势猜胜方”。

### 17. EventReason、事件与位置均来自已应用 Effect

resolver 生成内部 `ReasonRecord`：

```go
type ReasonModifier struct {
    Code   string
    Value  float64
    Detail string
}

type ReasonValue struct {
    Kind   string // Number/String/Bool/Null
    Number *float64
    String *string
    Bool   *bool
}

type ReasonStateChange struct {
    Field  string
    Before ReasonValue // 只允许公开标量状态
    After  ReasonValue // 只允许公开标量状态
}

type ReasonRecord struct {
    Code       string
    MainFactor string
    Modifiers  []ReasonModifier
    ScoreDelta float64
    Probability *float64 // 不适用为 nil；实际 0 不能与缺失混淆
    Formula    string
    Inputs     map[string]float64
    StateChanges []ReasonStateChange
    SourceActionID string
    SourceEffectID string
}
```

公开 DTO 采用现有 `GameEvent/EventReason` 的向后兼容扩展，而不是另建一套 `RoundEvent`：`GameEvent` 新增稳定 `EventID/SourceActionID/SourceEffectID`；`EventReason` 保留现有 `Code/MainFactor/ScoreDelta(float64)/Detail`，并新增结构化 `Modifiers`、可空 `Probability`、`Formula`、`Inputs`、只含公开字段的 `StateChanges` 以及 source IDs。ReasonRecord 投影不得截断 ScoreDelta、丢失真实 0 概率或把隐藏 Intent/ActionQueue/ActualControl 写入 StateChanges；ReasonValue 必须且只能设置与 Kind 匹配的一个标量字段，不接受 map/object。来自 action 生命周期但没有 Effect 的事件必须有 SourceActionID，SourceEffectID 可以为空；来自已应用 Effect 的事件二者都必须与 AppliedBatch 一致。

ScheduledAction 队列的稳定键固定为 `ResolveAt/Priority/ActionType/MinActorID/ActionID`。公开事件在成功应用 action/effect 后，独立按 `Timestamp/Priority/ActionType/MinActorID/SourceActionID/SourceEffectID/EventID` 排序；`EventID` 不替代 scheduler 的 `ActionID`，两层排序不能混用。

只有成功应用的 Effect 或真实 Action 生命周期变化才能生成公开事件。被版本校验取消的 PlantComplete/DefuseComplete/MovementArrive 不得产生完成事件。后端完整事件集合至少包含 `ROUND_START`、`STRATEGY_ADJUSTED`、`DAMAGE`、`KILL`、`ROTATE`、`REINFORCE`、`CONTROL_GAINED`、`BOMB_DROP`、`BOMB_PICKUP`、`BOMB_PLANT_START`、`BOMB_PLANT_INTERRUPT`、`BOMB_PLANT`、`DEFUSE_START`、`DEFUSE_INTERRUPT`、`BOMB_DEFUSE`、`BOMB_EXPLODE` 和 `ROUND_END`。Start/Interrupt 事件只能来自实际 action 开始或失效原因，不能从最终结果反推。Explainer 只排序、裁剪和聚合 ReasonRecord，不根据最终 Winner 补写原因。

关键事件包含当时 Bomb/比分/受影响玩家快照；ROUND_START、BOMB_PLANT、BOMB_DEFUSE、BOMB_EXPLODE、ROUND_END 保留完整公开快照。内部敌方隐藏 Intent/ActionQueue/ActualControl 不进入客户端 DTO。

引擎为每回合和整场生成 `ExplainableReport`：

```go
type ExplainableReport struct {
    KeyEvents       []GameEvent
    StrategySummary string
    LossReasons     []EventReason
    WinFactors      []EventReason
}
```

`KeyEvents` 只能引用已输出事件；StrategySummary 来自实际模板、角色和 Decision；LossReasons/WinFactors 由已应用 ReasonRecord 聚合。最终 Winner 只用于选择“胜方/败方视角”的分组，不能生成不存在的原因。前端第一版仍可只展示 KILL，但 RPC/后端结果保留完整事件和 Report。

位置先由实际行动/Encounter 决定语义来源，再使用 `LocationSeed` 采样。风险热点只提供候选权重。OnEdge 死亡/掉包使用运行时 Progress 位置。相同 EventID、SourceObjectID 和 RoundSeed 必须得到相同位置。

### 18. Seed 派生和可复现性

禁止共享可变全局 RNG。每个决定使用身份派生 seed：

```text
RoundSeed      = Hash(MatchSeed, MapVersion, RuleSetID, RoundNumber)
StrategySeed   = Hash(RoundSeed, "strategy", TeamID)
OpeningPlanSeed= Hash(RoundSeed, "opening_plan", AttemptOrdinal)
CTSetupSeed    = Hash(RoundSeed, "ct_setup", AttemptOrdinal)
DecisionSeed   = Hash(RoundSeed, "decision", Phase, Timeline, DecisionOrdinal)
ActionSeed     = Hash(RoundSeed, "action", ActionID)
EncounterSeed  = Hash(RoundSeed, "encounter", EncounterID)
PulseSeed      = Hash(EncounterSeed, PulseIndex, ResolveAt)
ActorRollSeed  = Hash(PulseSeed, ActorID, TargetID, RollKind)
BombSeed       = Hash(RoundSeed, "bomb", BombActionID)
MemorySeed     = Hash(RoundSeed, "memory", StrategyTemplateID)
LocationSeed   = Hash(RoundSeed, "event_location", EventID, SourceObjectID)
```

T OpeningPlan 与 CT setup 分别使用 `OpeningPlanSeed` 和 `CTSetupSeed`，初次生成固定使用 `AttemptOrdinal=0`，仅一次重试固定使用 `AttemptOrdinal=1`。两次保持相同 MatchInput/root seed，但不会重复消费完全相同的随机流；CT seed 上下文不得混入当前 T 隐藏计划。新增不相关 action 或 Reason 字段不应漂移既有 Actor 的随机结果。所有 map 候选转 slice 后稳定排序。相同输入、配置、规则和 seed 要求 RoundResult 深度相等（允许 StartTime 由业务层固定后比较）。

### 19. Dust2 第一版配置完成标准

正式模拟不能以当前单一路线占位数据通过验收。配置必须覆盖：

- 文档列出的六个 T 战术模板：`A_Long_Rush`、`A_Short_Split`、`B_Tunnel_Explode`、`Mid_To_B`、`Default_Pick`、`Fake_A_Go_B`；
- 多个带稳定 ID 的 CT setup RouteTemplate，使用 side/route_ids/route_allocations 覆盖 A/B/Mid 初始部署，且实际选择独立于本回合 T 隐藏计划；
- 对应的 T 语义 Route，以及 CT A/B/中路开局部署、补防和回防 Route；
- `OpeningDuel`、`MidControl`、`SiteEntry`、`Retake`、`BombResolution` 场景族，A/B/中路和 PostPlant 均有合法组合；
- 文档核心节点：T_SPAWN、LONG_DOOR、A_LONG、PIT、A_SITE、MID、CATWALK、SHORT、B_UPPER、B_TUNNEL、B_SITE、CT_SPAWN、CT_MID、B_DOOR、CAR，以及必要风险热点；
- 路线所需 MapEdge、主要拦截候选和关键 Visibility；
- 长/近距离、CT 架点、包点压力、转点风险、Timing/Posture 等 MapTag 与对应 EncounterModifier；
- 场景族十项属性权重、姿态/协同/资源修正、UtilityBudget、Pulse/Duration、Hit/Kill/Damage clamp、移动/决策、Intel/Control、调度上限和标定常量。

正式第一版 CombatConst 必配库存冻结为：`RoundTimeLimit`、`BombExplodeTime`、`BasePlantTime`、`BaseDefuseTime`、`BasePickupTime`、`MinPickupTime`、`MaxPickupTime`、`ForceExecuteThreshold`、`DecisionDelay`、`MaxDecisionCount`、`MaxEncounterPulses`、`MinCombatDuration`、`MaxCombatDuration`、`PulseFireWindow`、`CombatScale`、`MinDamagePotential`、`MaxDamagePotential`、`MinExposureModifier`、`MaxExposureModifier`、`BaseNoise`、`MaxRandomNoise`、`CloseScoreGap`、`DecisiveScoreGap`、`DefaultStrategyTemplateID`、`DefaultCTSetupTemplateID`、StrategyMemory 相关键、`ControlIntelTTL`、`CommunicationDelay`、`SoundIntelMinConfidence`、`SoundIntelMaxConfidence`、`DeathIntelMaxConfidence`、`MinIntelTTL`、`MaxIntelTTL`、`UtilityBudget`、`MaxStateTransitions`、`MaxScheduledActions`、`MaxEffectsPerTimestamp`、`MaxNoOpTransitions`、`MaxRotationsPerTeam`、`MaxRoundTimeline`，以及主设计表中列明的 Attribute/HP/Stamina/Focus/HitChance/KillChance/Plant/Defuse/Move clamp。十属性场景族权重和场景修正可位于 EncounterModifier；除此之外，引擎不得依赖未命名默认值。

`ValidateMapConfig` 除引用/几何/连通性外，还验证模板 side、route_ids/route_allocations 的同阵营引用与人数闭合、六个 T 模板和多个 CT setup 覆盖、场景阶段覆盖、A/B Plant 节点、双方开局可达路线、两个默认模板阵营正确、属性权重和为 1、概率/时间上下界、Scheduler 上限正值以及 `CommunicationDelay == 0`。缺失时对应地图 RPC 返回结构化配置错误，不使用隐藏 fallback。

校验错误码必须覆盖主设计矩阵，而不是只返回一个泛化错误：

| 范围 | 稳定错误码 |
|---|---|
| RouteTemplate | `CONFIG_DUP_ROUTE_TEMPLATE`、`CONFIG_BAD_TEMPLATE_SCENARIO`、`CONFIG_BAD_TEMPLATE_LIMIT` |
| Scenario/MapTag/Modifier | `CONFIG_DUP_SCENARIO`、`CONFIG_BAD_SCENARIO_TAG`、`CONFIG_BAD_SCENARIO_MAP_TAG`、`CONFIG_BAD_SCENARIO_WEIGHT`、`CONFIG_DUP_MAP_TAG`、`CONFIG_BAD_MAP_TAG`、`CONFIG_BAD_ENCOUNTER_MODIFIER`、`CONFIG_BAD_REASON_CODE` |
| 正式 Dust2 覆盖 | `CONFIG_INCOMPLETE_DUST2_COVERAGE` |
| MapNode geometry | `CONFIG_DUP_NODE`、`CONFIG_BAD_NODE_ENUM`、`CONFIG_BAD_NODE_COORD`、`CONFIG_BAD_NODE_SHAPE`、`CONFIG_BAD_NODE_CIRCLE`、`CONFIG_BAD_NODE_POLYGON`；事件位置解析选作 KillSample 来源的合法坐标节点缺少可用区域几何时输出 `CONFIG_MISSING_KILL_SAMPLE` 告警并回退该节点 `x/y`，缺少 `x/y` 仍为 `CONFIG_BAD_NODE_COORD` |
| Edge/Visibility/Route | `CONFIG_BAD_EDGE_NODE`、`CONFIG_BAD_RISK_POINT`、`CONFIG_BAD_INTERCEPT_NODE`、`CONFIG_BAD_VISIBILITY_NODE`、`CONFIG_BAD_ROUTE_NODE`、`CONFIG_ROUTE_NOT_CONNECTED`、`CONFIG_BAD_ROUTE_LIMIT`、`CONFIG_UNREACHABLE_NODE` |
| Bomb/Constants | `CONFIG_NO_PLANT_SITE`、`CONFIG_DUP_COMBAT_CONST`、`CONFIG_BAD_COMBAT_CONST_TYPE`、`CONFIG_BAD_COMBAT_CONST_RANGE`、`CONFIG_MISSING_COMBAT_CONST`、`CONFIG_BAD_BOMB_CONST`、`CONFIG_UNSUPPORTED_COMMUNICATION_DELAY` |

OpeningDeploy 若产生模板阵营错误、超出 Route min/max、route_allocations 不闭合或不可达的分配，返回内部 `INVALID_OPENING_PLAN`。T planner 使用 `OpeningPlanSeed`，CT planner 使用独立 `CTSetupSeed`，均以 AttemptOrdinal 0 初次生成且仅以 AttemptOrdinal 1 重试一次；仍失败时分别只能选择配置快照显式提供的 `DefaultStrategyTemplateID` / `DefaultCTSetupTemplateID`。不能在引擎硬编码 `Default_Pick`、默认 CT 站位，或从 T 本回合模板强制推出 CT setup。默认模板不存在、阵营错误或仍非法时返回结构化配置/计划错误，不进入模拟。

所有表格修改发生在 `configs/Datas/#*.xlsx`，随后用 Luban 导出 `server/config/data` 和生成代码；不得手改生成 JSON。`matchengine` 只读取映射后的 MapConfig。

### 20. 标定不是验收因果正确性的替代品

先通过单元/不变量/确定性测试，再运行固定配置和 seed 集合的 10k+ 单回合标定。输出：T 胜率、平均击杀、下包率、拆包率、爆炸率、首杀方胜率、5v3 胜率、3v5 翻盘率、强弱队胜率和平均时长。

指标未进入目标区间时，优先调整 RouteTemplate、Scenario、EncounterModifier、MapTag、UtilityBudget、CombatScale、时间和 Intel 参数。不得加入 TargetWinner、强制终局或按目标统计重新采样整个回合。标定测试同时验证所有 seed 的 terminal invariant。

### 21. 测试接缝不污染生产因果路径

删除 `MatchInput.ForcedRoundWinners`。MR12/换边/加时使用独立 MatchRules/score state 测试，或在 Match 层注入只返回 `RoundTerminal` 的测试 fake；fake 不声称生成真实战斗事件。

Round engine 测试通过固定 `RollSource`/派生 seed、构造 Scheduler action、resolver fixture 和状态 snapshot 控制边界。生产 Service 只能构造标准 deterministic roll source 和 causal round engine，RPC 不暴露胜方脚本。

静态测试扫描生产源码，禁止出现以 winner/target survivors 为 Encounter/Bomb/Decision 输入的签名，以及 `matchengine` 导入 `windypath.com/cs2match/config`。

### 22. Nakama、RPC 和客户端兼容

现有 `DebugSimuMatch` RPC 继续一次性构造 MatchInput 并调用 `Service.Simulate`，且始终返回按 RuleSet 推演到终局的完整比赛。RPC 不保留、不新增也不解释任何单回合/RoundCount/MaxRounds 截断字段；内部 `roundSimulator` 只供 Match 规则单测、RoundEngine 单测、离线调试命令和批量标定使用，业务 RPC 不获得该接口。不新增 RPC、Match Handler、Storage 操作、数据库迁移、Go module 或 npm 包；MatchLoop 和实时同步机制不受影响。

RPC 主体字段保持兼容。新增事件类型（包括 Plant/Defuse Start/Interrupt、DAMAGE、BOMB_DROP/PICKUP、ROTATE、REINFORCE、STRATEGY_ADJUSTED）、ExplainableReport 和 EventReason 扩展采用向后兼容字段。客户端继续只渲染服务器权威事件；如果当前 UI 不展示这些事件/Intel/Report，它们仍保留在后端战报和统计中。

## Risks / Trade-offs

- [事件队列、打断和并发使实现复杂度显著增加] → 以 Action/Effect/Scheduler/Terminal 四个小内核分层，每层先做纯函数和不变量测试，再接完整回合。
- [文档的 KillChance 与实际伤害容易形成两套死亡机制] → 只允许 DamageEffect 修改 HP；KillChance 只选择致命 damage package，死亡永远由 HP 归零派生。
- [同秒处理顺序可能产生 ActorID 偏置] → 同 priority 使用 snapshot 原子结算，不同 priority 才顺序应用；所有 tie 使用稳定键和身份派生 seed。
- [配置当前深度不足导致实现被迫 fallback] → 把 Dust2 数据完成度设为正式验收项，缺失则结构化失败，绝不隐藏 fallback。
- [语义图路径选择可能被误解为完整 pathfinding] → 仅在 MapNode/MapEdge/Route 上做有界确定性搜索，不模拟物理导航。
- [10k+ 标定会增加测试耗时] → 普通 CI 使用较小固定样本验证稳定性，完整标定作为显式长测试/命令运行并保存汇总。
- [重构后旧 seed 的具体战报改变] → 接受本次语义修复导致的旧结果不兼容；新实现以子 seed 和 golden/invariant 测试建立新的稳定基线。
- [完整 Reason/快照增加响应体] → 引擎保留完整结果；RPC 可裁剪非关键 debug 输入，但不能裁剪决定胜负的事件、Bomb 或最终状态。
- [测试 fake 掩盖生产战斗] → MatchRules fake 只验证比赛规则；生产 RoundEngine 必须另有端到端因果、不变量、同秒和标定测试。

## Migration Plan

1. 先建立静态禁令和失败测试，锁定当前预抽胜方、线性阶段、计时越界、幽灵 action 和同秒偏置等反例。
2. 完成 Dust2 配置与 CombatConstants 清单、Luban 表源和校验 fixture；导出生成数据，不修改生产引擎。
3. 实现 RoundState、Action/Effect、Scheduler、deadline、批次事务和 Terminal 纯函数，使用合成 action 测试所有边界。
4. 实现语义移动、Intel/Control、Strategy/Role/Planner 和 Decision，验证并发推进、打断和 AI 信息边界。
5. 实现 Encounter/CombatPulse/Utility 和 Bomb 状态机，验证伤害到死亡、同 pulse 互换、Plant/Defuse/Explode 同秒规则。
6. 将 `simulateRound` 切换到新 RoundEngine，删除 `resolveRoundWinner`、按 winner 生成事件、目标存活人数、`inferWinReason(winnerSide, ...)` 和 `ForcedRoundWinners`。
7. 接回 MR12/换边/加时、事件投影、统计、Reason 和客户端类型，运行端到端比赛测试。
8. 运行短 CI 标定、完整 10k+ 标定、后端测试、前端构建、Go Plugin 构建和 OpenSpec strict 校验。

回滚不涉及数据库；可以整体回滚代码与配置版本。但旧引擎是已知不符合设计的实现，不能与新 RoundEngine 在同一生产路径按概率混用。

## Open Questions

无阻塞问题。上述方案能够在不引入逐帧物理仿真的前提下实现 `simuMatchDesign.md` 的全部第一版核心因果要求。实现阶段若配置表表达能力与文档字段发生真实冲突，应停止修改代码并先提交具体字段/导表限制供讨论，不能用硬编码绕过。
