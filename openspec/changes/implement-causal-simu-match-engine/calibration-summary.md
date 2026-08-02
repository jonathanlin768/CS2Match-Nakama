# 标定摘要

## 最终配置

标定只修改 Luban 源表并通过 `scripts/gen-config.ps1` 重新导出，没有修改冻结公式或加入目标胜方、目标终局、结果重采样和生产测试钩子。

- `RoundTimeLimit=90`、`BombExplodeTime=30`、`BasePlantTime=2`、`MinPlantTime=2`
- `ForceExecuteThreshold=60`、`MaxDecisionCount=3`
- `MaxEncounterPulses=5`、`MaxScheduledActions=1500`、`CombatScale=50`
- T 主进攻路线缩短到 2–4 秒的语义边，CT A/Mid/B 首段为 10/8/11 秒
- T FastTiming/TExecuting 与 T 场景协同修正提高，CT 架点修正降低；所有修正仍落入冻结的 Posture、Visibility、TeamSupport 或团队级 EncounterScore 项

## 10,000 回合基准样本

命令：`SIMU_LONG_CALIBRATION=1 go test ./internal/match -run TestLubanCalibrationLongSample -count=1 -v`

| 指标 | 目标 | 结果 | 状态 |
|---|---:|---:|---|
| T 回合胜率 | 45%–55% | 51.09% | 命中 |
| 平均击杀 | 6.5–8.5 | 5.4617 | 未命中 |
| 下包率 | 30%–45% | 51.27% | 未命中 |
| 拆包率（已下包回合） | 8%–18% | 12.56% | 命中 |
| 爆炸率 | 15%–30% | 14.48% | 低 0.52 个百分点 |
| 首杀方胜率 | 60%–75% | 61.56% | 命中 |
| 5v3 胜率 | 80%–92% | 72.87% | 未命中 |
| 3v5 翻盘率 | 3%–10% | 27.13% | 未命中 |
| 平均回合时长 | 45–85 秒 | 59.7326 秒 | 命中 |

终局覆盖：`bomb_defused=644`、`bomb_exploded=1448`、`bomb_secured=3035`、`elimination=1081`、`timeout=3792`。所有 10,000 个 seed 均形成合法终局，没有调度错误或缺失结果。

## 2,000 回合受控强弱样本

默认两队总属性只相差约 1%，不构成强弱机会；标定器现在要求至少 10% 属性差才计入 `StrongTeamOpportunities`。长标定另用相同配置与生产 RoundEngine，对强队全属性 `+15`、弱队全属性 `-15` 的输入快照运行 2,000 回合，并交替阵营。

- 强队胜率：57.10%（目标 60%–75%，接近但未命中）
- 机会数：2,000
- 平均回合时长：60.9405 秒

## 结论

配置已把阵营平衡、拆包/爆炸、首杀价值和时长拉回合理区间，并让属性差在 `CombatScale=50` 下真实影响概率。剩余偏差集中在击杀密度、下包偏高和人数优势滚雪球不足；本提案保留这些真实标定结果，不通过改写冻结公式或预定结果伪造目标分布。后续可继续仅从 Scenario、RouteTemplate、MapTag、EncounterModifier、UtilityBudget、时间和情报参数迭代。
