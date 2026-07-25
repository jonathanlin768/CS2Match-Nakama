# simu-config-structures Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
### Requirement: Server reuses Luban TbPlayer config

`match.Service` SHALL 通过 `windypath.com/cs2match/config` 子模块读取 `TbPlayer` 选手属性，并显式构造为 `matchengine.Combatant` 传入引擎；`matchengine` 不直接依赖 `config` 包，只消费自包含的 `MatchInput`。

#### Scenario: Service loads player attributes from config

- **GIVEN** `TbPlayer` 已随 `cfg.Init()` 加载到内存
- **WHEN** `match.Service` 构造引擎队伍时传入选手 ID
- **THEN** `match.Service` 从 `cfg.Global.TbPlayer.Get(id)` 读取 `Entry`、`Aim`、`Firepower` 等属性
- **AND** 将其填充到 `matchengine.Combatant` 后作为 `MatchInput` 的一部分传给引擎
- **AND** 若选手 ID 不存在，返回 `INVALID_LINEUP` 错误

### Requirement: Player config structure is shared between server and client

`TbPlayer` 导出的字段 SHALL 与 `client/src/config/schema.ts` 中自动生成的 TypeScript 类型保持一致，确保前后端对选手属性的理解相同。

#### Scenario: Client displays real player names

- **GIVEN** 服务端返回的战报中包含 `attacker_name: "NiKo"`
- **WHEN** 前端用 `attacker_id` 查询本地 Luban 配置
- **THEN** 名称与 `TbPlayer` 中的 `Name` 字段一致

### Requirement: Route configuration is defined in engine code for MVP

MVP 阶段 `internal/framework/matchengine` SHALL 内建 Dust2 的 6 条进攻路线配置，包含路线 ID、显示名称、目标包点、基础推进时间以及雷达图比例坐标。

#### Scenario: Route config available in engine

- **WHEN** 引擎初始化
- **THEN** 可通过路线 ID（如 `"A_Long"`）查询到对应 `RouteConfig`
- **AND** `RouteConfig.TargetSite` 为 `"A"` 或 `"B"`
- **AND** 存在该路线对应的雷达图比例坐标 `Location { x, y }`

### Requirement: Route names are consistent between server and client

服务端返回的路线 ID 与显示名称 SHALL 与前端的临时路线常量一致，避免展示歧义。

#### Scenario: Client renders route name

- **GIVEN** 战报中 `RoundResult.RouteMain` 为 `"A_Long"`
- **WHEN** 前端展示回合开始事件
- **THEN** 界面上显示“A大”或“A_Long”文本

### Requirement: Combat constants are centralized

MVP 阶段虽然采用简化战斗，但回合时间上限、一回合固定长度等常量 SHALL 集中在 `matchengine` 的常量文件中，便于后续替换为 Luban 配表。

#### Scenario: Constants are not scattered in engine logic

- **WHEN** 查看 `internal/framework/matchengine` 源码
- **THEN** 存在 `const.go` 或 `model.go` 中统一定义的常量（如 `RoundTimeLimit`、`DefaultBombExplodeTime`）
- **AND** 引擎逻辑中不出现未命名的魔法数字

### Requirement: Config loader exposes player lookup helper

`server/config` 包 SHALL 保持并暴露 `GetPlayer(id string) *cfg.Player` 辅助函数，供 `match.Service` 安全读取选手数据。

#### Scenario: Service retrieves player by ID

- **GIVEN** 服务端硬编码阵容中包含 `"player_niko"`
- **WHEN** `match.Service` 调用 `cfg.GetPlayer("player_niko")`
- **THEN** 返回非 nil 的 `*cfg.Player`
- **AND** 其 `Name` 字段为 `"NiKo"`

