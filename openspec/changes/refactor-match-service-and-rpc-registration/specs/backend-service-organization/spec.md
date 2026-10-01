## ADDED Requirements

### Requirement: 模块采用有类型的有序 RPC 注册清单

后端 SHALL 通过静态清单关联现有 RPC 名称与 Nakama handler。Match SHALL 提供模块注册入口，Social SHALL 使用有序清单。通用助手 SHALL 按清单顺序注册并记录成功日志，失败时 SHALL 返回携带 RPC 名称且保留原错误的包装错误，并停止后续注册。

#### Scenario: 中途注册失败
- **GIVEN** 清单包含三个 RPC，第二个 RegisterRpc 返回错误
- **WHEN** 执行注册助手
- **THEN** 只尝试第一个和第二个注册
- **AND** 返回错误包含第二个 RPC 名称且 errors.Is 可以匹配原错误

#### Scenario: 既有模块注册成功
- **GIVEN** 插件初始化且 Social 交换功能启用
- **WHEN** 执行 HealthCheck、Match 和 Social 注册入口
- **THEN** 现有 RPC 名称与 handler 均注册成功
- **AND** 现有好友删除和拒绝客户端聊天 hooks 保留

### Requirement: 模式准备与共同模拟流程分离

SimuMatch SHALL 使用字符串 MatchMode 常量，按模式将配置校验与组队交给独立准备函数，返回包含地图与双方 TeamInput 的统一内部结果。共同流程 SHALL 保留原有模拟、错误转换与响应构造。未知模式 SHALL 保持 INVALID_MODE，不得进入模拟。

#### Scenario: 模式选择保持 JSON 契约
- **GIVEN** 客户端仍提交 mode=computer 或 mode=tutorial 的 JSON 字符串
- **WHEN** 解析请求并准备比赛
- **THEN** 原有默认队伍或教学配置校验参与组队
- **AND** 返回现有完整战报，前端请求/响应类型不变

#### Scenario: 未知模式与缺失认证
- **GIVEN** 请求模式无效或 RPC 上下文无用户身份
- **WHEN** 服务或 RPC 处理请求
- **THEN** 分别返回原有 INVALID_MODE 或 UNAUTHORIZED
- **AND** 不调用模拟引擎

### Requirement: 引擎服务依赖名称表达生命周期

Match 服务的引擎入口依赖 SHALL 命名为 engineService，构造参数 SHALL 使用相同名称。该依赖 SHALL 继续引用 matchengine.Service，由其在每次模拟调用中创建单场运行时对象。

#### Scenario: 调用引擎入口
- **GIVEN** Match 服务通过构造函数收到引擎服务
- **WHEN** 完成模式准备并执行模拟
- **THEN** 调用该 engineService 的 Simulate
- **AND** 引擎 JSON 输出与原有比赛规则不受命名调整影响
