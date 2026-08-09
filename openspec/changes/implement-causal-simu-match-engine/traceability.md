# 设计追踪表

| 主设计章节 | OpenSpec 落点 | Task 落点 | 实现落点 | 测试落点 | 结论 |
|---|---|---|---|---|---|
| 1.1–1.3 目标、边界、核心抽象 | `simu-engine-core`：正式模拟禁止预先选择胜方；`Simulate` 入口 | 1.1–1.4、14.1–14.4 | `engine.go`、`round_engine.go` | `production_path_test.go`、`round_engine_test.go` | 生产 Match 只走 causal RoundEngine；无旧反向结算入口 |
| 2.1–2.2 Dust2 地图与配置表 | `simu-map-config-runtime`、`simu-config-structures` | 2.1–2.9、5.1–5.5 | `config_adapter.go`、`validation.go`、Luban `#*.xlsx` | `validation_test.go`、`service_test.go`、`movement_test.go` | 路线、节点、边、视野、场景和常量均来自校验快照 |
| 2.3–2.4 资源、边界、十属性、Utility | 配置覆盖公式和状态边界；调度参数来自快照 | 4.1–4.6、9.1–9.4 | `round_state.go`、`utility.go` | `round_state_test.go`、`utility_test.go` | Utility 有有限预算、局部 scope、生产消耗和 LOW_UTILITY 证据 |
| 2.5 团战、唯一公式、CombatPulse | `EncounterResolver` 真实伤害；攻击/生存唯一公式；同优先级快照原子结算 | 10.1–11.7 | `encounter.go`、`combat.go`、`effect_apply.go` | `encounter_test.go`、`combat_test.go`、`effect_apply_test.go` | `PlayerCombatScore` 和 `TargetSurvivalScore` 无额外加数；HP 归零才派生死亡 |
| 2.6 AI 决策与情报边界 | 中期决策由当前局势驱动；AI 只读本方情报 | 8.1–8.5、12.1–12.6 | `strategy.go`、`decision.go`、`intel.go` | `strategy_test.go`、`decision_test.go`、`intel_test.go` | CT 不读取 T 隐藏计划；低置信度、TTL、死亡情报均受边界限制 |
| 2.7 炸弹状态机 | 炸弹实体推进；同秒事件批次后判胜 | 13.1–13.7 | `bomb_state.go`、`bomb.go`、`terminal.go` | `bomb_test.go`、`terminal_test.go`、`effect_apply_test.go` | Plant/Drop/Pickup/Defuse/Explode 可打断；同秒优先级按最终状态判定 |
| 3.1–3.4 流程、离散时间、阶段并发 | 阶段循环；统一离散时间；局部 Encounter 不阻塞其他 action | 3.1–3.8、6.1–7.6、15.1–15.5 | `round_engine.go`、`scheduler.go`、`action.go`、`movement.go` | `scheduler_test.go`、`round_engine_test.go`、`movement_test.go` | 绝对时间队列、多次 Clash/Decision、A 区交火与 B 区行动并发均有测试 |
| 3.3 group action 版本/互斥/打断 | 行动生命周期支持版本失效；group action 人数约束 | 6.4–6.7 | `action.go`、`scheduler.go`、`effect_apply.go` | `scheduler_test.go`、`effect_apply_test.go` | ActorVersion、MinRequiredActors 和批次失效规则一致 |
| 3.3 Recovery Cycle / NoOp | NoOp 不伪造普通胜负；合法不可达形成独立终局 | 7.1–7.6、14.3 | `noop.go`、`terminal.go`、`round_engine.go` | `noop_test.go`、`terminal_test.go` | 自身 bookkeeping 保留失败证明；外部进展重置；断图返回配置/调度错误 |
| 3.5–3.6 数据结构、事件 DTO、位置和解释 | EventReason 解释实际过程；完整因果事件和 ExplainableReport | 16.1–16.6 | `model.go`、`event_projection.go`、`round_projection.go` | `event_projection_test.go`、`match_memory_test.go` | Event/Action/Effect source IDs、Reason、StateChanges、Bomb、位置 seed 均可审计 |
| 4.2 输入/输出、MR12、比分投影 | 完整 causal result；换边只改变阵营投影 | 14.1–14.4、15.1–15.5、17.6 | `engine.go`、`match_score.go`、前端 `battle-playback.ts` | `engine_test.go`、`match_score_test.go`、`battle-playback.test.mjs` | `ScoreTeamA/B` 是唯一累计比分；`ScoreT/CT` 仅为当局阵营投影 |
| 4.3–4.8 加载、校验、随机性、算法规则 | 稳定配置错误分类；相同完整输入 deep-equal | 2.8–2.9、17.1–17.2 | `validation.go`、`random.go`、稳定排序/ID helpers | `validation_test.go`、`round_engine_test.go`、`scheduler_test.go` | map 反向插入不改变完整 MatchResult；配置错误矩阵稳定 |
| 4.9 标定 | 因果模拟支持固定 seed 批量标定 | 17.3–17.4 | `calibration.go`、Luban 标定参数 | `calibration_test.go`、`internal/match/calibration_test.go` | 短 CI + 显式 10k 基准 + 2k 强弱样本；偏差如实保留 |
| 4.10–4.11 实现顺序和非目标 | 测试能力与生产路径隔离；第一版限制细节深度 | 17.5–17.9 | 现有 `DebugSimuMatch` RPC；无新 Handler/Storage/外部依赖 | `production_path_test.go`、构建/静态扫描 | 未扩展到 tick、弹道、经济、实时同步或目标结果接口 |

## 一致性复核

- `SHALL`：唯一公式、因果终局、权威队伍比分、同秒批次、配置校验、情报边界、source IDs、Recovery Cycle 均有实现和测试落点。
- `MAY`：前端可不展示全部 Intel/Report、场景权重可存 EncounterModifier 等可选项，没有被实现或任务提升为相互矛盾的强制条件。
- 未发现 design/spec/task 之间互相矛盾的 MAY/SHALL；旧的 simple-comparison 兼容段由更具体的 causal CombatPulse 要求约束，生产实现以冻结公式和实际 DamageEffect 为准。
