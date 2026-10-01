## Why

比赛引擎已有按职责拆分的单包布局，但公开 API 与内部实现混在一起，模型文件承担多种数据契约，回合主循环还直接读取调度器内部队列。需要在保留现有模拟行为的前提下明确调用边界、状态所有权和阅读入口。

## What Changes

- 增加包级说明和文件职责索引，保持 `matchengine` 单包与同目录测试布局。
- 按输入、输出、地图配置和常量拆分公开模型，保留名称及 JSON 契约。
- **BREAKING**：仅在包内使用的运行时类型、resolver 和常量改为未导出名称；保留业务实际使用的配置助手及离线标定 API。
- 把队列查询收回调度器，移除不使用的回合引擎反向引用。
- 增加从外部包调用 `Service.Simulate` 的契约测试，以重构前固定输入的完整战报摘要验证结果保持一致。
- 同步因果引擎主规格与当前源码引用，记录已知功能接线缺口及后续拆包条件。

## Capabilities

### New Capabilities
- `simu-engine-maintainability`：公开 API、内部状态封装、文件职责和行为兼容验证的维护约定。

### Modified Capabilities
无新增玩法要求；已有因果引擎 delta 规格的同步属于文档一致性修复。

## Impact

- Nakama 后端：`server/internal/framework/matchengine` 及对应学习文档、OpenSpec 主规格；业务 RPC、阵容构造和离线标定调用保持兼容。
- React 前端：战报字段和模拟行为保持一致，执行现有战报/回放相关检查。
- 数据库、部署：无迁移或配置修改；用现有 Nakama pluginbuilder 验证 Linux `.so` 编译。
- 不新增 RPC、Match Handler、Storage 操作、Go module、外部依赖或 Luban 配表。保持一次 RPC 完成整场离线模拟，不改变 MatchLoop tick 与网络同步。
