# simu-engine-maintainability Specification

## Purpose

定义比赛引擎的公开 API、内部封装、职责索引、状态所有权、行为兼容验证与文档维护约定，为结构重构提供可验证的兼容性约束，并要求维护文档准确记录生产调用路径和已有功能缺口。

## Requirements

### Requirement: 引擎明确公开与内部调用边界

`matchengine` SHALL 保留 `Service.Simulate` 作为业务发起整场比赛的入口，以及输入、战报、地图配置、错误、配置助手和离线标定的必要 API。仅供包内使用的运行时类型、resolver 与常量 SHALL 不导出。

#### Scenario: 业务与标定调用继续编译
- **GIVEN** 现有业务层及标定测试使用引擎公开契约
- **WHEN** 编译整个 server module
- **THEN** 原业务和标定调用继续可用
- **AND** 外部包无法通过导出名称构造内部回合状态或调用内部战斗 resolver

### Requirement: 引擎按职责组织模型与状态访问

引擎 SHALL 提供包说明和文件职责索引；输入、战报、地图配置与错误 SHALL 具有各自的模型文件。生产代码 SHALL 在调度器实现中封装队列与拦截去重记录的直接访问。

#### Scenario: 回合主循环查询队列
- **GIVEN** 回合运行逻辑需要判断指定类型行动是否已经排队
- **WHEN** 查询行动队列
- **THEN** 使用调度器提供的方法
- **AND** 回合运行文件不直接读取或修改调度器队列字段

### Requirement: 结构整理保持公开模拟行为

本次重构 SHALL 保持公开类型、JSON 字段、错误码、随机派生、事件排序和比赛计算行为。外部包契约测试 SHALL 使用重构前生成的固定输入和完整战报摘要验证兼容性，不得使用内部 resolver 或生产测试钩子。

#### Scenario: 固定输入产生相同完整战报
- **GIVEN** 重构前记录的输入、StartTime 和三个 seed
- **WHEN** 外部测试通过 `Service.Simulate` 执行比赛
- **THEN** 每个完整 MatchResult 的 JSON摘要与重构前一致
- **AND** 输入快照不被修改

#### Scenario: 无效输入与取消继续返回错误
- **GIVEN** 空输入或已取消的 Context
- **WHEN** 外部测试调用 `Service.Simulate`
- **THEN** 返回相应结构化错误或 Context 取消错误
- **AND** 不返回部分比赛结果

### Requirement: 当前文档区分要求与实现缺口

当前维护文档 SHALL 使用现存文件与符号，并标明已知生产接线缺口。主规格 SHALL 去除被因果引擎正式替换的旧 MVP 要求，保留后续实例身份要求与尚有效场景。

#### Scenario: 阅读维护入口
- **GIVEN** 开发者阅读引擎 README、教程或主规格
- **WHEN** 查找公开入口、内部类型与已知缺口
- **THEN** 可以定位当前源码
- **AND** 不把情报、拦截等设计要求误认为本次结构调整已经修复
