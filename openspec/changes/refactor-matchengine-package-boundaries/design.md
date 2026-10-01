## Context

`matchengine` 的 28 个生产文件共享权威回合状态、行动和确定性排序。直接按战斗/移动/炸弹拆包会引入状态与调度器双向依赖。业务主要使用 `Service.Simulate`、配置模型与助手；`internal/match/calibration_test.go` 还使用 `RoundInput`、`CalibrateRounds` 和 `CalibrationSummary`。

## Goals / Non-Goals

**Goals:**
- 明确业务入口、离线标定 API 和包内实现的边界。
- 拆分不同语义的模型文件、收回队列访问、更新可跟踪的源码引用。
- 保留当前比赛计算、事件排序、JSON、错误码和固定输入战报。

**Non-Goals:**
- 本次结构整理不改变情报、拦截、包点决策、恢复或同秒批次的玩法实现；将缺口记录在当前维护文档中。
- 不新增子包、依赖、配置表或网络生命周期。

## Decisions

1. 保持一个 `matchengine` 包，使用 `doc.go` 和 `README.md` 展示调用链、文件职责、状态所有权及拆包条件。按运行时对象逐个拆子包需要先解开共享状态依赖，留待独立能力接口稳定后实施。
2. 将 `model.go` 拆为 `input.go`、`report.go`、`map_config.go`，把协议枚举归入 `const.go`，错误归入 `errors.go`。保持公开类型名、字段和 JSON 标签；不建立全局 `types/common` 包。
3. 依据当前仓库调用点和公开类型的传递依赖保留 API，其余顶层运行时类型、函数及常量改为未导出。用 Go 类型信息定位同一对象的定义和引用，避免改到同名字段、JSON 键、字符串或测试名称。`RoundInput` 和 `StrategyMemory` 作为离线标定契约继续公开。
4. 回合运行代码通过调度器查询方法获取排队信息，把 `scheduler.actions` 和拦截去重记录留在 `scheduler.go` 内。把 `RoundState` 的时间/计数方法移至 `round_state.go`，清理 `causalRoundEngine.owner` 的无用反向引用。
5. 重构前通过现有测试输入生成独立 JSON fixture，并固定 StartTime 和三个 seed，记录完整 MatchResult 的 SHA-256。外部 `matchengine_test` 包使用公开 API验证完整战报、错误结构、输入只读及取消行为；不添加生产测试钩子。
6. 维护文档和主规格区分设计要求与当前缺口。旧因果变更的要求合并到主规格时保留后续新增的实例身份要求与未被替换的场景；历史 Review 保留原提交上下文。

## Risks / Trade-offs

- [符号改名影响调用或静态因果守卫] → 全仓库调用审查、类型感知改名、全量 Go 测试；守卫继续按大小写无关名称检查。
- [固定战报误受当前时间影响] → 显式设置 StartTime，使用完整 JSON摘要，fixture 在修改生产逻辑之前生成。
- [主规格同步被误解为缺口已修复] → 在当前维护文档中单列尚未接入的生产路径，不把规格视为实现完成证明。
- [教程引用被改名打断] → 更新当前教程与需求文档中的符号和模型文件；历史审查资料注明映射入口。
- [Windows 无法直接构建 Go plugin] → 用项目固定的 Nakama 3.30.0 pluginbuilder 编译到临时容器路径，不替换本地服务已加载的产物。

## Migration Plan

保存基线 → 结构调整 → 外部契约与全量回归 → 文档/规格一致性 → 前端回放检查和 Linux plugin 编译。无需数据迁移；回滚使用本次变更的反向补丁并保留用户原有未提交改动。
