## Why

当前 `matchengine` 在每回合开始时先通过双方整体评分抽取胜方，再根据该预定胜方生成伤亡、炸弹事件和获胜原因。这条因果链与 `doc/simuMatchDesign.md` 定义的模拟模型相反，使选手属性、武器、遭遇战状态、中期决策和炸弹阶段无法真正决定比赛结果，因此现有战报只能证明“结果一致”，不能证明“比赛被真实模拟”。

## What Changes

- **BREAKING**：移除正式回合流程中的 `resolveRoundWinner` 预抽胜方，以及所有读取预定胜方后再控制击杀方向、存活人数、下包概率或终局类型的逻辑。
- 按 `simuMatchDesign.md` 构建轻量离散事件状态机，而不是把回合实现成各执行一次的线性阶段流水线：战术模板和角色分配生成开局 Intent/Action，调度器按绝对时间推进移动、架点、拦截、CombatPulse、决策、下包、拆包和计时器 deadline，并在关键状态变化后循环重评估。
- 实现 `Intent -> ActionInstance -> BusyInterval` 行动生命周期、逐 Actor 行动版本失效、按选手互斥占用、可打断移动/下包/拆包，以及无共同 Actor 的行动并发；group action 冻结 `VersionByActor` 和来源人数约束，成员失效后按剩余合法人数确定性继续或整体取消，同一选手不得同时完成互斥行动。
- 实现确定性的绝对时间 `ScheduledAction` 优先队列与同时间戳事务批次：resolver 只计算 Duration/ResolveAt，scheduler 比较 action 和 deadline 后推进到最早时间点，不允许阶段 TimeCost 吞掉并发行动；同优先级战斗使用批次开始快照计算以允许自然补枪/互换，高优先级效果落地后重新校验并取消低优先级过期行动。
- 将整个 `CombatPulse` 作为单一原子战斗事务：同一 pulse 的全部合法攻击先基于快照完成采样并统一应用伤害，再派生死亡、击杀、掉包和打断；同 pulse 中被击杀者已经形成的攻击不得被取消。
- 引入明确的 `RoundState`、`PlayerState`、`BombState`、局部控制权和统一时间推进，使伤害、死亡、资源、剩余时间、控制权与炸弹状态成为后续阶段的权威输入。
- 实现由实际参战选手、场景、地图语义、武器、姿态、协同、资源和身份派生 seed 驱动的 `EncounterResolver`/战斗脉冲；使用主设计唯一冻结且不重复计分的 `PlayerCombatScore`、`TargetSurvivalScore`、`EncounterScore`、HitChance、DamagePotential 和 `KillChance <= HitChance` 公式，先应用实际伤害，再重新评估局势。`DecisiveScoreGap` 只阻止有界噪声翻转评分排序，不预定 Encounter 胜者。
- 实现由当前公开/已知状态驱动的中期继续推进、转点、强攻、自由人、截回防和补防决策；禁止决策读取尚未产生的终局结果。`RoundState.Phase` 只是主导活动投影，不能阻塞其他区域互不共享 Actor 的合法行动。
- 实现 ActualControl 与双方 KnownControl/TeamIntel 分离、情报置信度与 TTL；AI 只能读取本方情报，不能读取敌方隐藏行动、未观察炸弹位置或全图真实控制权。第一版保留 `CommunicationDelay` 字段但要求为 `0`、队内即时共享，声音线索不生成虚假位置或人数。
- 实现下包、打断、掉包、拾取、拆包、爆炸及同秒事件优先级，由事件批次后的最终状态统一判定回合胜方和 `WinReason`。
- 将连续 NoOp 达到阈值、确定性 ForceExecute/恢复尝试仍失败后确认的合法 T 方不可达状态定义为正式 `NO_PROGRESS_TIMEOUT` 终局，并公开独立 `RoundResult.WinReason = no_progress_timeout`；恢复过程使用 `CycleID + NotAttempted/Running/Failed/Succeeded` 生命周期，恢复 action 自身的 bookkeeping 不会提前清除失败证明，恢复成功或其他来源真实进展才重置周期。普通批次仅满足不可达条件不能提前终局，配置、调度或状态错误必须返回模拟错误，不能伪装成该终局。
- 完成 `simuMatchDesign.md` 列出的 Dust2 第一版主数据和完整必配常量库存，而不是继续使用当前仅含一条路线/一个模板/一个场景的占位配置：RouteTemplate 同时配置带 side/路线人数分配的六个 T 战术和多个 CT setup，CT 使用独立 seed 与本方信息选择且不得读取本回合 T 隐藏计划；同时覆盖必要的双方进攻/防守/回防路线、阶段场景、地图节点、风险热点、关键视线和战斗参数；缺少正式覆盖或场景权重非法时以稳定错误码拒绝加载。
- 将现有 `GameEvent/EventReason` 向后兼容扩展为可审计 DTO：保留 `float64 ScoreDelta`，增加 Event/Action/Effect 稳定 ID、实际概率、公式输入、结构化修正和公开状态变化，不再为预定结果生成事后叙事；后端完整输出下包/拆包开始、打断和完成事件、`STRATEGY_ADJUSTED`，并生成包含 KeyEvents、StrategySummary、LossReasons 和 WinFactors 的 `ExplainableReport`，前端仍可只展示 KILL。
- 保留 MR12、换边、加时、固定 seed、稳定排序、完整战报和聚合统计；队伍身份比分是唯一权威比分，`ScoreT/ScoreCT` 仅按当前阵营映射投影，换边不得交换或重置队伍累计分。测试专用强制能力必须与正式模拟路径隔离，且不得直接指定胜方来生成生产事件。
- 扩充单元测试与批量标定测试，验证因果不变量、公式/clamp 边界、同秒优先级、确定性回放以及目标分布指标。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `simu-engine-core`：将“事件与预定终局相容”的弱约束升级为严格的因果模拟要求；删除过时的 MVP 比大小/硬编码路线要求，并规定离散事件调度、行动生命周期、并发/打断、遭遇战伤害结算、状态驱动决策、情报边界、炸弹状态机及事后终局判定。
- `simu-config-structures`：要求完整 Dust2 第一版 T 战术/CT setup/路线/场景数据以及战斗公式、场景权重、概率边界、资源消耗、调度上限和决策阈值由自包含快照及 Luban 配置提供；TbPlayer 完整映射为 PlayerProfile，禁止沿用旧三属性 Combatant、占位配置或散落隐藏常量完成正式模拟。
- `simu-map-config-runtime`：要求运行时真正使用 Route/MapEdge/Visibility/风险热点推进语义移动、OnEdgeLocation、拦截、双方可达性和事件位置，并将完整 Dust2 数据覆盖、主设计配置错误矩阵和开局计划恢复规则纳入地图可用性校验。

## Impact

- **Nakama 后端**：主要修改 `server/internal/framework/matchengine` 的模型、回合编排、决策、遭遇战、炸弹、事件解释、统计与测试；`server/internal/match` 仅在输入快照或 DTO 映射确有新增字段时同步调整。
- **React 前端**：不改变现有战报 RPC 的主要 JSON 契约；若事件类型或原因字段扩展，只做兼容性类型补充，不改回放交互目标。
- **数据库 / Storage**：无数据库迁移，不新增 Nakama Storage 操作。
- **RPC / Match Handler / 状态同步**：继续使用现有一次性 `DebugSimuMatch` RPC，不新增 RPC，不引入 Match Handler，也不改变实时状态同步机制；`DebugSimuMatch` 只构造正式 `MatchInput` 并调用 `Simulate`，始终推演到 RuleSet 终局，不接受或保留单回合、RoundCount、MaxRounds 等截断字段。内部 `roundSimulator` 只供测试、离线调试和标定使用，业务 RPC 不获得该接口。
- **Luban 配置**：必须扩充 `#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx` 和 `#CombatConst.xlsx`，使 Dust2 数据覆盖设计文档的第一版主路径、调度/战斗参数与校验要求；生成 JSON 必须由 Luban 导出，不得手工维护。
- **部署 / 依赖**：不新增外部运行时依赖，不改变 Docker、Nakama 或 PostgreSQL 部署结构。
