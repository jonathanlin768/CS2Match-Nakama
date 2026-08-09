## ADDED Requirements

### Requirement: 引擎消费地图配置快照

`matchengine` SHALL 消费自包含的地图配置快照，该快照包含路线模板、场景、地图标签、遭遇战修正、地图节点、地图路径、视野行、路线和战斗常量。

#### Scenario: 有效 Dust2 地图配置初始化引擎
- **GIVEN** `server/internal/match` 将 Luban 生成的 Dust2 表转换为 `matchengine.MapConfig`
- **WHEN** `matchengine.NewService` 或 `Service.Simulate` 收到该快照
- **THEN** 引擎可以在不读取 `server/config` 的情况下选择路线模板和场景
- **AND** 路线、节点、视野或战斗常量不从硬编码 Dust2 全局变量读取
- **AND** 必需地图配置缺失时返回错误，而不是用内建 fallback 数据替换

### Requirement: 地图配置校验语义引用

地图配置校验层 SHALL 在模拟使用配置前拒绝重复 ID、缺失引用、非法枚举值、非法节点几何、断裂路线节点、缺失必需战斗常量，以及非法风险/拦截引用。

#### Scenario: 路线引用缺失节点
- **GIVEN** `tb_route.nodes` 包含 `A_LONG` 和 `MISSING_NODE`
- **WHEN** 校验配置快照
- **THEN** 校验失败并返回 `CONFIG_BAD_ROUTE_NODE`
- **AND** 不会为该地图配置启动模拟

#### Scenario: 缺失必需战斗常量
- **GIVEN** 配置快照不包含 `RoundTimeLimit`
- **WHEN** 校验配置快照
- **THEN** 校验失败并返回 `CONFIG_MISSING_COMBAT_CONST`

#### Scenario: 必需配置不使用 fallback
- **GIVEN** 配置快照缺失 `de_dust2` 的全部 `tb_route_template` 行
- **WHEN** `Service.Simulate` 校验输入
- **THEN** 校验失败并返回结构化配置错误
- **AND** 引擎不会使用任何硬编码 Dust2 路线模板 fallback

### Requirement: 服务初始化时缓存地图配置校验状态

`server/internal/match` SHALL 在服务初始化时构建并校验地图配置快照，并缓存每张地图的成功快照或配置错误状态。配置错误 SHALL NOT 要求整个 Nakama 插件启动失败；调用对应地图模拟 RPC 时 SHALL 返回缓存的结构化配置错误。

#### Scenario: 配置错误不阻止插件启动但阻止模拟
- **GIVEN** `de_dust2` 配置缺失必需 `tb_route_template` 行
- **WHEN** Nakama 加载后端插件并初始化 `match.Service`
- **THEN** 插件初始化可以完成
- **AND** `match.Service` 缓存 `de_dust2` 的配置错误状态
- **WHEN** 客户端调用 `DebugSimuMatch`
- **THEN** RPC 返回对应 `CONFIG_*` JSON 错误
- **AND** 不启动该地图模拟

### Requirement: 事件位置使用地图节点和路线语义

KILL 和炸弹相关事件的位置 SHALL 根据 `doc/simuMatchDesign.md`，从当前路线、场景、地图节点几何、边位置或配置的风险/拦截候选中生成，并使用派生 seed 保证稳定采样。

#### Scenario: 击杀从节点几何中采样
- **GIVEN** KILL 事件来源节点的 `area_usages` 包含 `KillSample`，且 `shape = Circle`
- **WHEN** 生成事件位置
- **THEN** `GameEvent.Location.X` 和 `GameEvent.Location.Y` 位于配置圆形范围内
- **AND** 使用相同 seed 重复模拟时返回相同采样坐标

### Requirement: 战斗常量通过类型化访问器读取

来自 `tb_combat_const` 的战斗常量 SHALL 被解析为类型化值，并通过命名的引擎方法或字段访问，而不是以魔法数字散落在逻辑中。

#### Scenario: 回合时间来自配置
- **GIVEN** `tb_combat_const.RoundTimeLimit` 配置为 115 秒
- **WHEN** 模拟一个回合
- **THEN** 回合计时器使用配置快照中的 115 秒
- **AND** 引擎逻辑不依赖无关的硬编码 `RoundTimeLimit` 常量

### Requirement: 框架不依赖生成配置包

`server/internal/framework/matchengine` SHALL NOT 直接导入 `windypath.com/cs2match/config` 或任何 Luban 生成表包。

#### Scenario: Import 边界检查
- **WHEN** 后端测试检查 `server/internal/framework/matchengine` 下的 imports
- **THEN** 没有任何文件导入 `windypath.com/cs2match/config`
- **AND** 配置转换代码位于框架包之外
