## 1. 身份模型与服务适配

- [x] 1.1 扩展 Go `PlayerProfile`、公开 `PlayerState` 和 `PlayerMatchStats`，加入 JSON 字段 `config_player_id`，并把现有 `player_id` 注释和校验明确为比赛实例 ID。
- [x] 1.2 在 `internal/match` 实现稳定、转义安全的比赛实例 ID 构造函数，并校验非空 Team/Player ID、双方 Team ID 不同和生成结果全场唯一。
- [x] 1.3 修改所有 Match 服务队伍构建路径，使 `TbPlayer.id` 写入 `ConfigPlayerID`，以 `<team-id>/<config-player-id>` 生成 `PlayerID`，同时继续从同一 `TbPlayer` 复制属性、角色和视觉快照。
- [x] 1.4 调整 `validateTutorialLineup` 和默认队伍构建校验：允许双方共享策划选手 ID，保留玩家阵容内部、对手阵容内部以及其他单队内部的重复拒绝。
- [x] 1.5 添加 Match 服务测试，覆盖双方同选手成功进入引擎、两个实例 ID 唯一、相同 `config_player_id`、单侧重复仍失败和实例 ID 碰撞保护。

## 2. 模拟引擎投影与确定性

- [x] 2.1 审计回合状态、角色/路线、炸弹、击杀/伤害/助攻事件、最终统计及 stable seed/hash 路径，确保所有可变状态和 Actor 引用继续使用比赛实例 `PlayerID`，不以 `ConfigPlayerID` 建键。
- [x] 2.2 修改回合状态和事件投影，将 `ConfigPlayerID` 从 `PlayerProfile` 复制到公开 `PlayerState`，并在整场统计初始化与聚合中保留 `ConfigPlayerID` 和 `TeamID`。
- [x] 2.3 添加引擎单元测试，验证相同 `ConfigPlayerID` 的双方实例状态互不覆盖、Actor/炸弹/最终统计引用正确、重复实例 ID 被拒绝。
- [x] 2.4 添加确定性回归测试，验证包含双方同源选手的相同完整输入与 seed 产生完全一致的事件、比分和最终统计。

## 3. Luban 配置与可视化编辑器

- [x] 3.1 删除 Luban 导出测试和服务端配置初始化中“候选池与固定对手阵容不得重叠”的错误，保留跨费用档、对手阵容内部、人数、预算、引用和战队归属校验。
- [x] 3.2 修改 TutorialBattle 可视化编辑器的候选选手和对手阵容选择逻辑，取消跨双方禁用与重叠 ERROR，不改变费用档之间和对手阵容内部的禁用规则。
- [x] 3.3 更新编辑器与配置测试，覆盖跨双方重叠可保存/可加载、跨费用档重复仍报错、对手阵容内部重复仍报错。
- [x] 3.4 运行 Luban 导表与前后端生成数据校验，确认无需新增配置字段且交叉重叠的 `TbPlayer.id` 引用在导出结果中保留。

## 4. React 战报协议与回放

- [x] 4.1 更新客户端 `MatchReport` 相关 TypeScript 类型，为选手档案、状态和最终统计加入 `config_player_id`，明确 `player_id` 是不透明比赛实例 ID。
- [x] 4.2 审计并修改 Battle 回放 reducer、生命值/存活状态、事件目标、炸弹携带者和 KDA 索引，使其只用 `player_id` 关联可变战斗状态。
- [x] 4.3 修改卡面、头像裁切及其他 `TbPlayer` 查询，使其使用 `config_player_id`，并继续优先展示服务器战报中的视觉快照。
- [x] 4.4 添加客户端回归测试：双方均为同一策划选手时渲染两行独立实例、事件只更新目标实例、KDA 不合并且双方共用正确卡面。
- [x] 4.5 运行客户端 TypeScript 类型检查、单元测试和生产构建。

## 5. 集成验证与发布保护

- [x] 5.1 运行 Go 格式化和后端相关单元测试，并编译 Nakama Go Plugin `.so`，确认现有 `SimuMatch` RPC 注册、认证与权限边界未改变且未新增 Storage 操作。
- [x] 5.2 使用一份双方包含相同 `TbPlayer.id` 的 TutorialBattle 配置执行 Docker 联调，核对请求成功、十个比赛实例唯一、战报双重身份、回放生命值/炸弹/KDA 和视觉展示。
- [x] 5.3 验证旧的不重叠教学配置、computer 模式和 DebugSimuMatch 回归通过，并记录 `player_id` 协议变更需要服务端与客户端协调发布。
- [x] 5.4 验证回滚保护：旧运行时恢复前必须重新启用交叉重叠校验或停用所有重叠教学配置，且不需要数据库或 Nakama Storage 迁移。
