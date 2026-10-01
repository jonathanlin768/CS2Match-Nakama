# 比赛模拟引擎

业务层通过 `Service.Simulate(ctx, *MatchInput)` 发起完整比赛。引擎消费自包含快照，输出 `MatchResult`；RPC、阵容/配表读取、持久化和奖励属于 `server/internal/match` 等业务层。

## 从哪里开始读

```text
Service.Simulate                         service.go
  → newProductionMatchEngine            engine.go
  → matchEngine.simulateMatch           整场规则、回合输入、比分、统计
    → roundSimulator.SimulateRound      round_contract.go
    → runCausalRound                    round_engine.go
      → 开局计划、行动队列、推进时间、结算、决策、终局
      → projectRoundResult              round_projection.go
```

## 文件职责

| 范围 | 文件 | 职责 |
|---|---|---|
| 公开入口与契约 | `service.go`、`doc.go` | 业务入口、包说明 |
| 公开模型 | `input.go`、`report.go`、`map_config.go`、`errors.go`、`const.go` | 队伍/规则输入、战报、配置、结构化错误、协议常量 |
| 校验 | `validation.go` | 输入/配置校验和配置数值读取 |
| 整场编排 | `engine.go`、`match_rules.go`、`match_memory.go` | MR12、换边、加时、统计与战术记忆 |
| 回合运行 | `round_contract.go`、`round_engine.go`、`round_state.go` | 内部回合契约、主循环、权威状态与时钟 |
| 行动调度 | `action.go`、`scheduler.go` | 行动/Effect、占用、优先队列、版本有效性 |
| 战术与决策 | `strategy.go`、`decision.go`、`intel.go`、`opening_plan.go` | 战术/角色、局势决策、本方情报、尚未接线的开局重试路径 |
| 地图与战斗 | `movement.go`、`encounter.go`、`combat.go`、`effect_apply.go`、`utility.go` | 语义移动、遭遇战、脉冲计算、伤害提交、道具预算 |
| 炸弹与终局 | `bomb.go`、`terminal.go`、`noop.go` | 炸弹生命周期、纯终局判定、无进展恢复 |
| 投影与解释 | `round_projection.go`、`event_projection.go` | 公共状态、事件位置和实际计算原因 |
| 确定性与工具 | `seed.go`、`calibration.go` | seed 派生、批量标定 |

`*_test.go` 与实现放在同一目录，普通构建不包含它们。内部测试验证不变量；`service_contract_test.go` 使用外部 `matchengine_test` 包验证公开调用和重构前的完整战报基线。

## API 与状态所有权

- 业务 API：`NewService`、`Service.Simulate`，以及公开输入/输出、错误和配置模型。
- 配置助手：`DefaultMR12RuleSet`、`IsSupportedMap`、`ValidateRuleSet`、`ValidateMapConfig`、`CombatConstants.Int/Float`。
- 离线标定 API：`CalibrateRounds`、`RoundInput`、`StrategyMemory`、`CalibrationSummary`。`internal/match/calibration_test.go` 是现有调用者，业务 RPC 不用它们截断整场比赛。
- `matchEngine` 拥有单场比分、阵营、统计和记忆；`roundState` 拥有单回合玩家、炸弹、控制权、情报和时钟。
- `actionScheduler` 拥有行动堆及拦截去重记录。其他文件通过查询、快照、调度和清空方法访问；快照复制队列切片，其嵌套 payload 按只读约定使用。
- resolver 读取当前状态；战斗伤害由 `applyCombatPulseCommit` 提交并派生死亡、打断和掉包；公开 DTO 由投影产生，不写回内部状态。

Go 的未导出名称对整个包可见，因此上述包内所有权仍需代码审查维持。当前没有 `state → scheduler → state` 的跨包依赖，也没有为每个模拟阶段建立子包。

## 当前功能缺口

以下缺口在结构重构前已存在，本次固定战报基线保留当前行为。对应 OpenSpec 描述设计要求，不等于这些路径已经完成生产接线。

| 缺口 | 当前观察与后续方向 |
|---|---|
| 情报输入 | `recordIntel`、`degradeIntelForObserverDeath`、`intelScoreModifier` 尚未接入生产调用；需定义观测来源并验证 AI 读取边界 |
| 包点争夺 | `planSiteContest` 尚未接入正式回合规划；需验证下包资格与争夺决策 |
| 移动拦截/进度 | 有内部算法与测试，生产调度接线仍需按移动阶段复核 |
| 无进展恢复 | `completeNoProgressRecovery` 尚未接入生产；需验证恢复成功与失败分类 |
| 开局重试 | `opening_plan.go` 的 `planOpening` 尚未接入正式开局选择 |
| 同秒跨行动 | 当前 `resolveTimestamp` 逐行动应用，同 pulse 原子性已有实现；跨行动同优先级快照事务需要独立功能修复 |

历史细节见 `doc/reviews/ebc8906-review-guide.md` 和 `doc/cs2match-tutorial/15-验收与差距.md`。历史 Review 保留原提交的名称；当前包内类型/函数改用小写开头，例如 `RoundState → roundState`、`ResolveCombatPulse → resolveCombatPulse`。

## 何时考虑拆包

地图图算法、独立标定工具或整场/回合边界具备稳定输入输出、独立调用者和单向依赖后，可选择性提取。先明确状态所有权与只读视图，避免新包继续直接操作完整的 `roundState`。文件数量用于触发维护检查，不能单独决定包边界。

## 验证

在 `server/` 执行 `go test ./...`；公开契约基线见 `testdata/README.md`。Linux Go plugin 使用项目固定的 `heroiclabs/nakama-pluginbuilder:3.30.0` 编译。刻意修改玩法时，应审查完整战报差异再更新基线。
