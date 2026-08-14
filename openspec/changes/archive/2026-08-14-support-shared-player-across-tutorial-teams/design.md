## Context

当前 `internal/match` 从 `TbPlayer.id` 直接填充 `matchengine.PlayerProfile.PlayerID`。引擎随后把它作为整场比赛内的 Actor 主键：`RoundState.Players`、角色和路线分配、炸弹携带者、事件攻击者/受害者、最终统计以及确定性随机派生都按该值索引。若教学玩家阵容和固定对手阵容都包含 `player_zywoo`，后创建的 Actor 会覆盖或拒绝先创建的 Actor；因此配置层交叉校验不能在身份模型适配前单独删除。

这项变更横跨 Luban 配置校验、本地可视化编辑器、Nakama `SimuMatch` 服务、`matchengine` 数据模型和 React 回放。现有 RPC 名称、请求结构和服务器权威模拟流程保持不变，客户端仍只提交策划选手 ID，由服务端读取权威配置并构建比赛快照。

## Goals / Non-Goals

**Goals:**

- 允许教学战双方引用同一个 `TbPlayer.id`，并在模拟内表现为两个完全独立的战斗单位。
- 明确区分比赛实例 ID 与策划选手 ID，避免状态、事件、统计和回放关联发生覆盖。
- 复用同一份 `TbPlayer` 属性、卡面和头像裁切数据，不复制策划记录。
- 保持相同服务端输入与 seed 的确定性，并维持现有服务器权威边界。
- 将实现控制在服务适配层和身份字段扩展内，避免全面把引擎容器重构为复合键。

**Non-Goals:**

- 不允许同一支比赛阵容内部重复选择同一个策划选手。
- 不允许同一个策划选手同时出现在多个费用档。
- 不新增正式阵容资产、持久化规则、RPC、Match Handler、Storage 操作或 Luban 表字段。
- 不重构全部引擎 API 为显式的 `(PlayerID, InstanceID)` 复合键。
- 不承诺旧客户端能够正确消费新战报；这是一次协调发布的协议变更。

## Decisions

### 1. 服务层生成不透明且确定的比赛实例 ID

`internal/match` 在把 `TbPlayer` 转换为 `PlayerProfile` 时生成比赛实例 ID：

```text
<escaped-team-id>/<escaped-config-player-id>
```

例如：

```text
tutorial_players/player_zywoo
team_vitality/player_zywoo
```

两个组成部分使用稳定的路径段转义，消费者 MUST 把结果视为不透明字符串，不通过切割字符串恢复字段。构建器仍显式保存 `TeamID` 和 `ConfigPlayerID`，并校验双方比赛 `TeamID` 不同、每队内部 `ConfigPlayerID` 唯一、生成后的十个 `PlayerID` 唯一。

选择服务层适配，是因为它在调用 `matchengine.Service.Simulate` 前即可满足引擎现有的全局唯一键不变量。替代方案是让所有引擎 map 使用复合键或嵌套 team/player map；它会触及几乎所有调度、战斗、炸弹和统计代码，预计扩大到 4–7 个工作日，本期不采用。

比赛实例 ID 对所有由 Match 服务构建的队伍使用同一规则，而不是只在检测到交叉重复时启用。这样同一字段不会因阵容内容不同而改变语义，测试与下游消费者也无需处理两种格式。

### 2. `PlayerID` 固定表示比赛实例，新增 `ConfigPlayerID`

身份字段语义统一为：

| 字段 | 语义 | 主要用途 |
|---|---|---|
| `player_id` | 整场比赛内唯一的实例 ID | 引擎 map、状态、事件、炸弹、角色/路径、随机派生、KDA/回放关联 |
| `config_player_id` | 对应 `TbPlayer.id` 的策划选手 ID | 属性来源追踪、卡面/头像裁切、客户端策划数据查询、未来持久化映射 |
| `team_id` | 本场比赛队伍 ID | 队伍归属和实例 ID 命名空间 |
| `player_name` / `display_name` | 展示快照 | 战报和 UI 展示 |

`PlayerProfile`、客户端可见的 `PlayerState` 和 `PlayerMatchStats` 增加 `config_player_id`。队伍快照中的 `PlayerProfile` 由外层 `TeamInput.TeamID` 提供队伍上下文；公开状态和最终统计继续显式输出 `team_id`。击杀、伤害、助攻、炸弹和其他 Actor 事件字段仍只输出比赛实例 ID，不增加重复的 config attacker/victim 字段；客户端可通过同一战报内的选手快照解析策划身份。

替代方案是保留 `player_id=TbPlayer.id` 并新增 `instance_id`。这会要求修改引擎中所有现有 `PlayerID` 引用，且更容易遗漏 map 和随机种子路径，因此不采用。

### 3. 引擎继续使用单一字符串主键

引擎主体保留 `map[string]...` 和现有 `PlayerID` 字段。初始化时要求 `PlayerProfile.PlayerID` 全局唯一，并把 `ConfigPlayerID` 作为静态元数据复制到回合投影和最终统计。所有 Actor 相关事件、炸弹携带者、角色/路径分配以及 stable seed/hash 输入继续读取 `PlayerID`，从而自动使用比赛实例 ID。

同一策划选手在双方拥有相同能力与视觉快照，但 HP、存活状态、装备、位置、击杀和随机分支彼此独立。因为实例 ID 参与稳定排序和随机派生，引入命名空间后旧战报的具体模拟结果可能变化；变更后的相同完整输入与 seed 仍必须可复现。

### 4. 配置只放宽跨双方重叠

`#TutorialBattle.xlsx` 结构不变。Luban 导出/服务端初始化和可视化编辑器删除“费用档候选池与固定对手阵容不得重叠”的错误，编辑器也不再在另一侧禁用该选手。

以下约束保持：每个策划选手最多属于一个费用档、玩家提交阵容内部 ID 唯一、对手阵容内部 ID 唯一、人数/预算/引用/对手战队归属合法。这样放宽只作用于两个阵容集合的交集，不会允许单队出现两个同源 Actor。

### 5. 客户端按身份用途选择键

React 回放 reducer、生命值/存活状态、事件目标、炸弹携带者和 KDA 行使用 `player_id`。卡面、头像裁切或本地 `TbPlayer` 查找优先使用 `config_player_id`；战报已含视觉快照时仍优先使用权威快照。UI 不合并相同 `config_player_id` 的两行，并以 `team_id + player_id` 所属关系分别渲染。

服务器状态与客户端表现保持分离：服务器输出两个独立实例及其事件，客户端只据此显示，不推导或合并战斗状态。TypeScript 协议与 Go JSON 字段协调发布。

### 6. 保持 Nakama 插件与同步架构不变

现有 `SimuMatch` RPC 的请求仍携带 tutorial config ID、version 和玩家选择的五个 `TbPlayer.id`；响应扩展身份字段。无需注册新 RPC、修改 `InitModule`、Match Handler 生命周期或 Nakama Storage 权限。

该模拟路径不由实时 `MatchLoop` 驱动，因此不会改变 10/20Hz tick、插值或状态广播。若未来实时模式复用这些模型，也必须维持 `player_id` 的比赛内唯一语义。实现不引入新的 Go 依赖或 npm 包。

## Risks / Trade-offs

- [旧客户端把 `player_id` 当作 `TbPlayer.id`] → 同步更新 TypeScript 类型和所有策划数据查找点，并增加双方同选手回放测试；服务端与客户端协调发布。
- [遗漏某个事件或统计仍使用策划 ID] → 增加跨双方重复配置的端到端测试，遍历玩家状态、Actor 事件、炸弹携带者和最终统计，断言只出现两个唯一实例且均映射到同一 `config_player_id`。
- [实例 ID 拼接歧义或碰撞] → 路径段转义、禁止空 ID、要求双方 `TeamID` 不同，并在进入引擎前执行最终全局唯一性校验。
- [实例 ID 改变确定性结果] → 接受一次性的基线变化；校准测试改为验证新格式下同输入同 seed 可复现，而不保证与旧版本逐事件一致。
- [客户端错误合并同源选手] → 所有回放状态容器以 `player_id` 为键，`config_player_id` 只用于只读配置/视觉查找。
- [范围扩大到全面身份重构] → 保留引擎单字符串键，只扩展边界模型和投影；发现内部仍需复合键时单独提案。

## Migration Plan

1. 先扩展 Go/TypeScript 身份模型和序列化契约，并为旧的非空 `PlayerProfile` 构建路径补齐 `ConfigPlayerID`。
2. 更新 Match 服务，统一生成实例 ID；在服务与引擎边界增加唯一性测试。
3. 更新事件投影、状态和最终统计，并运行引擎确定性与整场模拟测试。
4. 更新客户端回放和视觉查找，再加入双方共享 `config_player_id` 的回归测试。
5. 最后放宽 Luban、服务端配置校验和编辑器限制，避免任何可生成的新配置先进入未适配的运行时。
6. 运行 Luban 导表校验、Go 单元测试与插件编译、React/编辑器测试和 Docker 教学战联调后协调发布。

回滚时应同时回滚运行时、客户端和配置校验，并在旧版本重新上线前恢复“双方不得重叠”校验；已保存的交叉重叠教学配置必须先停用或修正。没有数据库或 Storage 数据需要迁移。

## Open Questions

- 当前没有阻塞实施的问题。未来战报持久化若需要按策划选手聚合，应单独定义聚合维度，不得把本场 `player_id` 直接写回为 `TbPlayer.id`。
