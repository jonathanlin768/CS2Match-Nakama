## 1. 引擎模型与规则

- [x] 1.1 扩展 `server/internal/framework/matchengine` 模型类型，覆盖队伍身份、阵营分配、`RuleSet`、地图配置快照、选手档案、武器装配/规格、炸弹状态、控制权状态、事件原因和整场统计。
- [x] 1.2 实现 `RuleSet` 校验，覆盖 MR12 常规赛、换边、加时 block、回合计时、炸弹计时、决策限制和测试专用短模拟。
- [x] 1.3 将仅按阵营计分的状态替换为队伍权威比赛状态，同时保留现有 JSON 消费方需要的兼容字段。
- [x] 1.4 实现完整比赛推进：第 1-12 回合、半场换边、第 13-24 回合、常规赛先到 13 分获胜、12-12 进入加时、OT1 前 3 局延续下半场阵营、OT1 后 3 局换边、完整 6 小局 MR3 加时 block、加时 block 平局续打，以及仅在加时 block 完成后判定最终胜者。
- [x] 1.5 为比赛、回合、决策、遭遇战、炸弹、记忆和事件位置随机源实现确定性 seed 派生。
- [x] 1.6 跨所有模拟回合聚合最终选手统计，包括击杀、死亡、伤害、ADR、首杀、多杀、下包和拆包。

## 2. 地图配置运行时

- [x] 2.1 增加引擎自有的地图语义快照类型，覆盖路线模板、场景、地图标签、遭遇战修正、地图节点、地图路径、视野、路线和战斗常量。
- [x] 2.2 实现地图配置校验，覆盖重复 ID、缺失引用、枚举合法性、节点几何、路线连通性、风险/拦截引用、必需战斗常量和非法常量范围，并在服务初始化时缓存每张地图的成功快照或配置错误状态。
- [x] 2.3 用地图配置快照中的路线/模板/节点/位置数据替代 `matchengine` 中硬编码的 Dust2 路线和路线位置；必需配置失败时返回结构化错误，而不是隐藏 fallback 行为。
- [x] 2.4 将引擎需要的战斗常量写入 `configs/Datas/#CombatConst.xlsx`，通过 Luban 生成 `tbcombatconst.json`，并实现类型化访问器，使回合时间、炸弹时间、下包/拆包时间、clamp 值、决策限制和战斗缩放来自配置；禁止手工维护生成 JSON。
- [x] 2.5 使用派生 seed，从节点几何、边/运行时位置、路线 fallback、风险/拦截候选中实现事件位置采样。
- [x] 2.6 增加 import 边界测试或静态检查，证明 `server/internal/framework/matchengine` 不导入 `windypath.com/cs2match/config`。

## 3. Match 业务适配与 RPC

- [x] 3.1 构建 `server/internal/match` 适配器，将 `server/config` 中 Luban 生成表转换为 `matchengine.MapConfig`。
- [x] 3.2 在 `server/internal/match` 中直接从 Luban `TbPlayer.GetDataList()` 按 `team` 字段筛选，构造 Team A `Falcons` 与 Team B `Vitality`（每队恰好 5 人，队内保持稳定表顺序），传递 ID、名称、角色标签、头像路径和属性，保证数据唯一/有效，并确保 `matchengine` 内部没有运行时选手表依赖且 `tbplayer.json` 不被手工修改。
- [x] 3.3 增加第一版业务层 AK-47 和 M4A1-S 武器规格，并按回合当前阵营动态分配武器装配；`PlayerProfile` 不绑定固定枪械，半场和加时换边后回合内武器同步切换，同时将策划可编辑武器表留给后续提案。
- [x] 3.4 更新 `DebugSimuMatch`，构造完整 `MatchInput`：地图配置快照、默认队伍、武器规格、默认 MR12/加时 `RuleSet`、地图版本和有效 seed；当请求提供可选 `seed` 时优先使用该值。
- [x] 3.5 扩展 `server/internal/match` DTO 和 JSON 响应字段，覆盖完整比赛元信息、队伍比分、胜者队伍 ID、回合阵营分配、加时元数据、炸弹/控制权快照和最终统计。
- [x] 3.6 保留 `DebugSimuMatch` RPC 注册，并确认未引入新的 Nakama Storage 操作、数据库迁移或客户端可写权限。
- [x] 3.7 为非法地图、非法阵容、配置校验失败、非法比赛输入和模拟失败返回结构化 RPC 错误。
- [x] 3.8 更新 MatchPage 开始匹配 loading 体验，展示“匹配中”“正在模拟比赛”“进入战场”等阶段性反馈。
- [x] 3.9 从 `#CombatConst.xlsx` 读取 `BattleReportDebugLog`，在 `server/internal/match` 响应顶层返回 `debug_enabled`，且不将业务调试开关传入 `matchengine`。

## 4. 客户端类型与 BattlePage

- [x] 4.1 扩展 `client/src/types/match-report.ts`，增加整场比赛字段：队伍 ID、胜者队伍 ID、队伍比分、阵营分配、回合阶段、加时元数据、炸弹状态、控制权快照、武器装配、事件原因和最终统计。
- [x] 4.2 将 `client/src/pages/BattlePage.tsx` 从 `rounds[0]` 回放重构为整场比赛回放控制器，支持当前回合、最远已播放回合、可见事件索引、每秒一条事件自动播放、暂停/继续、仅回看已揭示回合，以及自然结束或“跳过比赛”后才显示完整局数和全部时间线。
- [x] 4.3 更新计分板表现，展示服务端队伍名称、队伍比分、当前 T/CT 阵营徽标、半场换边、加时标签和最终胜者。
- [x] 4.4 更新队伍阵容/选手卡派生逻辑，使用当前回合的服务端快照、已可见事件增量、当前阵营武器装配、存活/死亡状态和 K/D 总数。
- [x] 4.5 更新地图/事件流表现，展示当前回合事件标记、炸弹事件、炸弹状态小组件（若战报提供快照）、按回合分组的时间线，以及仅来自服务端返回状态的聚合统计。
- [x] 4.6 保留直接进入 `/battle` 且没有战报时的精致 fallback，并避免在客户端运行比赛规则逻辑。
- [x] 4.7 BattlePage 仅在 `debug_enabled=true` 时向 Console 输出完整原始战报、逐回合/逐事件表和实时播放事件日志，并记录“跳过比赛”操作。

## 5. 后端测试与构建验证

- [x] 5.1 增加 matchengine 单元测试，覆盖 MR12 提前获胜、半场换边、12-12 进入加时、OT1 延续下半场阵营开局、完整加时 block 结算、加时不提前结束、加时平局续打、队伍比分正确性，以及跨 seed 校验淘汰/超时/拆包/爆炸的事件、存活状态和获胜原因一致性。
- [x] 5.2 增加确定性模拟测试，证明相同输入和 seed 会生成相同的回合胜者、事件、位置、比分和最终统计。
- [x] 5.3 增加地图配置校验测试，覆盖错误路线节点、缺失场景/地图标签、非法节点几何、缺失战斗常量、错误风险/拦截引用，以及配置错误不阻止插件启动但阻止对应地图模拟。
- [x] 5.4 增加适配器/RPC 测试，覆盖默认按 `TbPlayer.team` 构建 `Falcons vs Vitality`、队内稳定表顺序、选手头像路径传递与客户端 fallback、`PlayerProfile` 不绑定固定枪械、按回合阵营分配 AK-47/M4A1-S、可选 seed 复现、结构化错误和完整 `DebugSimuMatch` 响应结构。
- [x] 5.5 运行受影响后端包测试，包括 `server/internal/framework/matchengine`、`server/internal/match` 和 `server/config`。
- [x] 5.6 使用项目 Nakama 插件构建流程运行 Go 插件编译验证，并确认 `server/build/backend.so` 成功编译。
- [x] 5.7 验证 Luban 导出包含调试开关、RPC JSON 顶层包含 `debug_enabled`，并运行相关 Go 测试、客户端类型检查和生产构建。

## 6. 前端测试与构建验证

- [ ] 6.1 增加或更新前端测试，覆盖整场战报类型、每秒一条事件播放、播放中隐藏完整局数与未来回合、自然结束/跳过后揭示完整局数、跳过到最终结果后仍可回看、回合导航、半场/加时标签、计分板队伍比分、阵容存活/死亡状态、炸弹事件/状态表现、统计视图和地图标记过滤。
- [x] 6.2 在 `client` 中运行 `npx tsc --noEmit` 并修复所有类型错误。
- [ ] 6.3 运行受影响的 BattlePage/battle 组件客户端测试。
- [x] 6.4 运行客户端生产构建，并验证 `/battle` 在直接 fallback 和战报驱动状态下渲染无布局重叠。

## 7. OpenSpec 与最终验证

- [x] 7.1 运行 `openspec validate complete-simu-match-mr12 --strict`。
- [x] 7.2 检查生成 artifacts 和实现 diff 中是否存在框架/业务边界违规，尤其是 `matchengine` 内的 config import。
- [x] 7.3 记录实现期间发现的任何未解决标定值或配置编写缺口。
