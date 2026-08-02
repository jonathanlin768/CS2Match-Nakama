## 背景与动机

当前模拟对局仍是 MVP 阶段的单回合调试流程：Dust2 路线被硬编码在引擎里，`BattlePage` 也只回放 `rounds[0]`。下一个里程碑需要一套完整的 CS2 风格模拟比赛，让“开始匹配”流程能够产出有价值的 5v5 整场战报，同时保持可复用的模拟引擎与业务默认数据清晰分离。

## 变更内容

- 将 `server/internal/framework/matchengine` 从单回合模拟升级为整场比赛模拟：MR12 常规赛制、12 小局后换边、先到 13 分获胜，常规赛 12-12 时进入加时。加时采用完整 6 小局 MR3 block，block 内不提前结束；OT1 前 3 局延续下半场阵营，后 3 局再换边。
- 在引擎内实现比赛级状态：回合号、队伍到阵营的映射、按队伍与按阵营的比分、换边、加时分段、每回合确定性 seed、最终胜者和聚合选手统计。
- 保证每回合可见战报、最终存活状态与获胜原因一致：淘汰必须以败方全灭结束且全灭后不得下包；超时必须由 CT 在双方仍有存活且未下包时获胜；下包回合必须以拆包或爆炸形成完整终局。
- 用框架自有的地图配置快照替代引擎中硬编码的 Dust2 路线/位置常量。该快照由 Luban 生成表构造，并匹配 `doc/simuMatchDesign.md`：`tb_route_template`、`tb_scenario`、`tb_map_tag`、`tb_encounter_modifier`、`tb_map_node`、`tb_map_edge`、`tb_visibility`、`tb_route` 和 `tb_combat_const`。服务启动时校验配置并缓存校验状态；必需配置缺失或非法时对应地图 RPC 返回结构化错误，不使用隐藏 fallback 数据，也不要求整个 Nakama 插件启动失败。
- 将选手归属和默认阵容保留在 `server/internal/match`：直接读取 `configs/Datas/#Player.xlsx` 经 Luban 导出的 `TbPlayer`，按 `team` 字段筛选，`Falcons` 的 5 名选手组成 Team A、`Vitality` 的 5 名选手组成 Team B，队内顺序沿用 `GetDataList()` 的稳定表顺序，并将包含基础属性、角色标签和头像路径的自包含 `PlayerProfile` 快照传入 `matchengine`；不手工修改生成的 `tbplayer.json`。任一默认队伍不是恰好 5 人或选手 ID 重复时返回 `INVALID_LINEUP`。
- 为本功能加入武器装配数据：`PlayerProfile` 本身不绑定固定枪械；模拟过程中按回合当前阵营动态分配，T 方固定 AK-47，CT 方固定 M4A1-S。合理的第一版固定武器数值保存在业务层输入快照中，而不是隐藏在战斗公式里。如果武器数据后续需要由策划配置，应另起提案引入专门的武器配置表。
- 演进 `DebugSimuMatch`：保留现有 RPC 名称以服务当前调试/开始匹配流程，但返回完整比赛战报，并支持可选请求字段 `seed` 以便确定性复现。
- 更新 React 客户端类型和 `client/src/pages/BattlePage.tsx`，让 `/battle` 展示整场比赛回放：常规赛/加时比分推进、换边感知、每秒自动播放一条事件、播放过程中只逐局揭示已推进的回合且不提前显示完整局数、“跳过比赛”或自然播放结束后才展示完整局数和全部时间线、改进事件流、选手统计、炸弹状态表现，以及已播放回合的地图标记。
- 在 `configs/Datas/#CombatConst.xlsx` 增加 `BattleReportDebugLog` 调试开关。`server/internal/match` 将它映射为战报响应的 `debug_enabled`，客户端仅在该值为 `true` 时向浏览器 Console 输出完整战报摘要、逐回合/逐事件数据和实时播放轨迹；`matchengine` 不感知此业务调试开关。
- 增加后端和前端测试，覆盖确定性、MR12/加时规则、配置加载边界、RPC 响应结构和 BattlePage 渲染行为。
- 为后续提案明确本次实现深度：第一版实现 SHALL 遵循 `simuMatchDesign.md` 的数据边界和主要阶段，同时将 action 生命周期、连续边上移动、完整情报传播保留为第一版的简化/抽象逻辑。

## 能力范围

### 新增能力
- `simu-match-rules`：完整比赛规则，包括常规赛、换边、加时、按队伍计分和确定性比赛推进。
- `simu-map-config-runtime`：运行时地图/战术/常量配置快照，`matchengine` 从 Luban 生成数据消费这些快照，不与业务层耦合。
- `simu-battle-full-replay`：客户端侧整场比赛回放，以及 `/battle` 路由的表现能力。

### 修改能力
- `simu-engine-core`：从 MVP 单回合模拟改为由 `MatchInput`、`RuleSet`、地图配置、逐回合状态和聚合统计驱动的完整比赛模拟。
- `simu-config-structures`：从硬编码路线常量改为 Luban 支持的地图语义配置，并显式传入武器/装配快照。
- `simu-match-rpc`：将 `DebugSimuMatch` 从单回合调试战报改为使用默认队伍和完整比赛规则的整场战报。
- `simu-client-replay`：将 BattlePage 从只回放第一回合改为展示支持换边和加时的完整模拟比赛。

## 影响范围

- Nakama 后端：影响 `server/internal/framework/matchengine`、`server/internal/match`、`server/config` 下的生成配置适配代码，以及相关测试。本次不需要新增 Nakama Match Handler；模拟仍是 RPC 驱动、服务端权威、一次性计算。
- React 前端：影响 `client/src/pages/BattlePage.tsx`、战报 TypeScript 类型、回放界面使用的 battle 组件，以及前端测试。客户端继续只渲染服务端返回的状态，不运行比赛逻辑。
- RPC/API：复用现有 `DebugSimuMatch` RPC 名称，并扩展响应结构。除非实现阶段发现必须通过独立端点维持向后兼容，否则不新增 RPC。
- Storage/数据库：不需要新的 Nakama Storage 操作、数据库迁移或权限变更。
- 状态同步：不改变 MatchLoop/tick 行为。本功能通过 RPC 生成完整确定性战报；未来实时观战可以继续消费同一事件模型。
- Luban/配置：需要按 `doc/simuMatchDesign.md` 更新或生成地图语义和战斗常量数据；战斗常量的唯一源头是 `configs/Datas/#CombatConst.xlsx`，`tbcombatconst.json` 只能由 Luban 导出生成，不允许手工补写；`tools/map-semantic-editor` 仍作为地图节点、路线、路径、视野和风险热点的编辑工具。
- 依赖/部署：预计不新增 Go 或 npm 依赖；继续使用现有 Go 1.24.5、Nakama 3.30.0、React 19、Vite 8 和 Luban 工作流。
