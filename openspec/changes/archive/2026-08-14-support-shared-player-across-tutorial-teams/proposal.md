## Why

教学战目前禁止玩家候选池与固定对手阵容引用同一个 `TbPlayer.id`，导致策划无法设计“双方使用同一名选手配置”的对局。模拟引擎又把 `PlayerID` 当作整场比赛内的唯一键，因此仅删除交叉校验会使状态、事件、炸弹、战术分配、最终统计、回放关联和确定性随机派生发生覆盖或串线。

## What Changes

- 允许教学战玩家阵容与固定对手阵容引用同一个策划选手 ID，同时继续禁止任一单侧阵容内部重复选手，并保留费用档之间的重复校验。
- 在 Match 服务构建双方 `TeamInput` 时，为每个战斗单位生成由 `team_id` 与策划选手 ID 组成的比赛实例 ID，例如 `tutorial_players/player_zywoo`；引擎继续以这个比赛内唯一 ID 关联状态、事件、随机派生和统计。
- **BREAKING**：扩展比赛快照与战报中的选手身份协议。`player_id` 明确表示比赛实例 ID，新增 `config_player_id` 表示对应的 `TbPlayer.id`，并同时保留 `team_id` 与展示名称。
- 扩展 `PlayerProfile`、`PlayerState` 和最终统计的身份字段，确保事件输出比赛实例 ID，而卡面、头像裁切、能力值和后续策划数据查询使用 `config_player_id`。
- 更新 React 回放逻辑：生命值、存活状态、炸弹和 KDA 按比赛实例 ID 关联，卡面与策划配置按 `config_player_id` 关联，并覆盖双方选择同一名选手的回归场景。
- 取消 Luban 配置校验、服务端配置校验和可视化编辑器中“候选池与对手阵容不得重叠”的限制；编辑器不再跨双方禁用选项或报告该重叠错误。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `simu-match-rpc`: 教学模式接受双方共享策划选手 ID，由服务层分配比赛实例 ID，并返回包含双重身份的权威战报。
- `simu-engine-core`: 模拟状态、事件、确定性派生和最终统计以比赛实例 ID 唯一关联，同时携带原始策划选手 ID。
- `simu-config-structures`: 明确 `PlayerProfile`、`PlayerState`、最终统计及教学配置校验中的比赛实例 ID 与策划选手 ID 语义。
- `luban-config-visual-editor`: 允许候选池与固定对手阵容交叉选择同一选手，保留单侧和跨费用档重复限制。
- `simu-client-replay`: 回放状态按比赛实例 ID 驱动，视觉与策划资料按 `config_player_id` 解析。
- `guest-battle-onboarding`: 教学战允许双方共享策划选手配置，并把服务端生成的两个独立比赛实例传入模拟与战报。

## Impact

- **Nakama 后端**：影响 `internal/match` 教学阵容校验和 `TeamInput` 构建、`matchengine` 领域模型、事件与最终统计投影，以及相关 Go 测试；不新增 RPC，不新增或修改 Match Handler，不新增 Storage 操作或权限。
- **React 前端**：影响 `MatchReport` TypeScript 类型、Battle 回放状态索引、KDA 与选手视觉资料解析，以及教学选人/回放测试。
- **Luban 与配置工具**：需要更新 `#TutorialBattle.xlsx` 的编辑器和导出后服务端校验规则，可能调整测试数据；不新增配置表或字段，不改变 `TbPlayer.id` 与现有表引用结构。
- **状态同步**：本变更作用于 `SimuMatch` RPC 生成的服务器权威战报，不改变实时 MatchLoop、tick rate 或客户端插值机制；客户端仍只消费权威快照和事件。
- **数据库与部署**：无数据库 schema、Nakama Storage、Docker 编排或新增依赖变更；需要现有 Go/TypeScript 测试、Luban 导表校验和 Docker 联调。
- **兼容性**：`player_id` 的含义从策划选手 ID 收窄为比赛实例 ID，所有依赖其读取 `TbPlayer` 的消费者必须迁移到 `config_player_id`。预计涉及约 10–18 个代码文件、8–15 个测试，推荐方案总工期约 2–3 个工作日。
