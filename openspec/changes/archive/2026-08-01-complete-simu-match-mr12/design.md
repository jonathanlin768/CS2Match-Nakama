## 背景

当前模拟路径偏向调试型 MVP。`server/internal/framework/matchengine` 暴露了 `Service.Simulate(ctx, *MatchInput)`，但 `MatchEngine.StartMatch` 只模拟一个回合，使用硬编码的 Dust2 进攻路线和路线坐标，并返回 T/CT 阵营比分，而不是持久的队伍比分。`server/internal/match` 从 `TbPlayer` 构造两套硬编码的五人阵容，并通过 `DebugSimuMatch` 调用引擎。`client/src/pages/BattlePage.tsx` 已有回放壳子，但只读取 `report.rounds[0]`，并把 `MAX_ROUNDS` 当作 24。

`doc/simuMatchDesign.md` 定义了目标边界：`matchengine` 必须是可复用的模拟框架，只消费阵容、武器、规则、地图配置版本和随机种子等显式快照。业务代码可以读取 Luban/选手默认数据来构造这些输入，但引擎运行时逻辑不得直接查询 `server/config`。该设计文档也明确第一版模拟仍是一次性 RPC 计算，而不是 Nakama Match Handler 循环。

本提案使用的外部规则与武器参考如下：CS2 公开比赛格式已转向 MR12 常规赛制；这里将主流整场比赛加时建模为 12-12 后的 6 小局 MR3 block。第一版固定武器默认值采用合理的 CS2 风格数值：AK-47 伤害 36、600 RPM、30 发弹匣、穿甲 0.775；M4A1-S 伤害 38、600 RPM、20 发弹匣、穿甲 0.70。这些数值是标定输入，不是隐藏公式常量。

## 目标 / 非目标

**目标：**

- 用户点击“开始匹配”后模拟完整比赛：上半场 12 小局、换边、下半场最多 12 小局、先到 13 分获胜，12-12 后进入完整 6 小局 MR3 加时 block；OT1 前 3 局延续下半场阵营，后 3 局再换边。
- 将所有可复用模拟代码保留在 `server/internal/framework/matchengine` 下。
- 将业务默认数据保留在 `server/internal/match` 下：从 Luban `TbPlayer.GetDataList()` 按 `team` 字段筛选，构造 Team A `Falcons` 与 Team B `Vitality`（每队恰好 5 人，队内保持稳定表顺序）、提供按阵营动态分配的武器默认值、处理 RPC 请求默认值，并转换生成配置/选手数据。`#Player.xlsx` 是选手数据源，不手工维护 `tbplayer.json`。
- 通过框架自有的输入/配置快照消费 Luban 生成的地图语义和常量，且该快照匹配 `doc/simuMatchDesign.md`。
- 扩展战报契约，让服务端和客户端能够展示完整常规赛/加时推进、队伍身份、每回合阵营分配、事件、炸弹/控制权快照和聚合统计。
- 更新 BattlePage，让它成为表现优秀的整场比赛回放页面，同时保持纯表现层职责。

**非目标：**

- 本次不实现经济系统。
- 不实现实时玩家输入、Match Handler、MatchLoop/tick 模拟或 WebSocket 广播。
- 不做数据库迁移或 Nakama Storage 变更。
- 除 `de_dust2` 外，不新增多地图生产内容。
- 不做完整空间寻路、弹道、连续视野传播或真实 CS 移动模拟。
- 不新增 Go module、npm 包或外部运行时服务。
- 本次不引入策划可编辑的武器表；如果武器数值需要迁出业务层默认值，后续提案应新增 `tb_weapon` 或等价表。

## 技术决策

### 决策 1：保持 `matchengine` 由快照驱动，且不依赖生成配置包

`matchengine` 将定义框架类型，例如 `MatchInput`、`RuleSet`、`TeamInput`、`PlayerProfile`、`WeaponLoadout`、`MapConfig`、`RouteTemplate`、`Scenario`、`MapNode`、`MapEdge`、`Visibility` 和 `CombatConstants`。它不会导入 `windypath.com/cs2match/config`。

`server/internal/match` 将通过读取生成配置/选手表来构造 `MatchInput`，并把这些数据映射成引擎自有快照。这样可以保留用户要求的框架/业务边界：`matchengine` 是基础能力，而用户/选手默认数据属于调用方。

备选方案：让 `matchengine.NewService` 接收 `cfg.Global` 或直接读取 `server/config`。拒绝原因：这会让框架依赖生成业务代码，并妨碍独立单元测试。

### 决策 2：以队伍比分作为比赛权威比分，阵营比分仅作为每回合派生信息

引擎将追踪 `TeamAScore` 和 `TeamBScore` 作为权威比赛比分。每个 `RoundResult` 会包含 `TeamTID`、`TeamCTID`、`WinnerTeamID`、`ScoreTeamA`、`ScoreTeamB`，并保留向后兼容所需的 `ScoreT` 和 `ScoreCT`，用于表示该回合/分段状态下的阵营比分。

第 1-12 回合使用输入中的 `InitialSideByTeam`。第 13 回合换边。常规赛 12-12 后进入加时，OT1 前 3 局延续下半场阵营分配，OT1 后 3 局再换边。后续每个加时 block 继续沿用“上一个半段后换边”的节奏：每个 block 的前 3 局使用当前阵营，后 3 局换边。加时始终打完整个 6 小局 block；只有 block 完成后某队领先才判定获胜。如果 block 结束后仍平，则开始下一个完整 block。

备选方案：继续将仅按阵营统计的 `ScoreT`/`ScoreCT` 作为最终权威比分。拒绝原因：换边后 T/CT 分数会误导半场和加时后的队伍比分展示。

### 决策 3：使用 `RuleSet` 快照表达 MR12 和加时

业务代码会向 `MatchInput` 传入完整默认 `RuleSet`：

- `RegulationHalfRounds = 12`
- `RegulationWinRounds = 13`
- `RegulationMaxRounds = 24`
- `OvertimeEnabled = true`
- `OvertimeHalfRounds = 3`
- `OvertimeBlockRounds = 6`
- `OvertimeWinBy = 1 after a completed six-round block`
- 回合/炸弹/下包/拆包时间来自 `tb_combat_const`

引擎校验会拒绝缺失、为 0 或互相矛盾的规则字段。这样未来 MR15、自定义训练赛或测试用单回合运行都可以在不修改引擎公式的情况下支持。`DebugSimuMatch` 也接受可选 seed，方便 QA 和调参复现某一整场比赛。

备选方案：在 `const.go` 中硬编码 MR12 常量。拒绝原因：设计文档要求规则和常量必须是显式输入或从配置派生的快照。

### 决策 4：在业务边界把 Luban 生成数据转换为引擎地图配置

`server/config` 已经包含路线模板、场景、地图标签、遭遇战修正、地图节点、地图路径、视野、路线和战斗常量等生成表。实现时应在 `server/internal/match` 或一个窄适配包中添加 mapper，将这些生成结构转换为 `matchengine.MapConfig`。

`configs/Datas/#CombatConst.xlsx` 是战斗常量的唯一可编辑源。实现所需常量及其类型、值和上下限必须写入该工作簿，再通过 Luban 同步生成服务端与客户端 `tbcombatconst.json`；生成 JSON 不接受手工修改。

`matchengine` 拥有对转换后快照的校验逻辑：重复 ID、缺失引用、非法节点几何、路线连通性、非法战斗常量，以及缺失 Dust2 关键对象等。服务初始化时应构建并校验地图配置快照，将成功快照或失败状态缓存到业务服务中。运行时模拟只能使用该快照。必需配置失败是硬错误，不做隐藏 fallback；但该错误不要求整个 Nakama 插件启动失败，对应地图的 RPC 应返回结构化配置错误，例如 `CONFIG_BAD_ROUTE_NODE` 或 `CONFIG_MISSING_COMBAT_CONST`。

备选方案：在 `matchengine` 下直接加载原始 JSON 文件。拒绝原因：生成配置已经是项目惯例，直接加载 JSON 会绕过 Luban 的类型化输出。

### 决策 5：固定武器是装配数据，不是公式

本次变更中，`PlayerProfile` 表示选手基础属性、角色标签和头像路径，不绑定固定枪械。头像路径原样来自 `TbPlayer.Portrait` 并随回合选手快照返回；客户端把相对路径解析到 Vite `public` 根目录，资源缺失时使用现有默认头像。`server/internal/match` 提供业务层 `WeaponSpec` 快照，包含伤害、射速、弹匣容量、穿甲、距离修正和显示名称。模拟派生 `RoundInput` 或回合状态时根据当前阵营动态分配武器：T 方固定 `WeaponLoadout{Primary: "AK47"}`，CT 方固定 `WeaponLoadout{Primary: "M4A1S"}`；半场和加时换边后同步变化。本次刻意不引入 Luban 武器表；当武器数值成为策划可编辑内容时，后续提案应新增专门武器表。

第一版战斗公式可以通过回合内 `PlayerState`/`RoundInput` 的当前武器，以及 `MapTag`/`Scenario` 上下文读取武器派生值。武器常量不得作为魔法数字散落在遭遇战逻辑中。

备选方案：继续只把武器保留为字符串标签。拒绝原因：本需求需要武器专属数值，并且未来需要做平衡调参。

### 决策 6：扩展事件模型，但不要求实时同步

`RoundEvent` 将包含比赛/回合时间戳、事件类型、参与者、武器、位置、原因、比分快照、炸弹快照，以及可选公共状态增量。实现时至少应支持现有 `ROUND_START`、`KILL`、`ROUND_END`，并在 resolver 生成时支持 `MATCH_START`、`HALF_TIME`、`OVERTIME_START`、`SIDE_SWITCH`、`BOMB_PLANT`、`BOMB_DEFUSE`、`BOMB_EXPLODE` 和 `MATCH_END`。

回合 resolver 必须先确定合法的终局类型，再生成与该终局相容的击杀和炸弹事件。击杀阶段只有在双方都达到各自目标存活人数后才结束；`elimination` 要求败方存活人数为 0，且全灭后不得继续下包；`timeout` 要求 CT 获胜、双方仍有存活且没有下包；发生 `BOMB_PLANT` 后必须继续产生与胜方一致的 `BOMB_DEFUSE` 或 `BOMB_EXPLODE`。`win_reason` 由最终存活状态和炸弹状态共同推导，不能只由预定胜方推断。

这仍然是一个 RPC 响应。由于模拟采用离散事件，且战报时间戳是回合内秒数，因此不需要 MatchLoop 级别的精度要求。

备选方案：从 Match Handler 流式推送回放事件。拒绝原因：设计文档明确第一版模拟保持 RPC 驱动。

### 决策 7：BattlePage 成为整场比赛回放控制器

客户端将服务端战报视为权威状态。它不计算胜者、换边、加时规则或战斗结果。它可以根据已播放事件和回合快照派生展示状态。

BattlePage 应支持：

- 跨全部回合的整场时间线
- 当前回合指示器，以及常规赛/加时标签
- 约每秒显示一条事件的自动播放
- 正常播放时仅逐局揭示已推进的回合，不提前显示完整局数或未来回合导航
- “跳过比赛”按钮，立即将展示推进到最终结果
- 自然播放结束或跳过后才揭示完整局数与完整回合时间线，可点击回看任意回合
- 以队伍为单位的比分，而不仅是 T/CT 阵营比分
- 选中/当前回合中每支队伍的可见阵营分配
- 按回合分组的事件流
- 来自当前回合/事件快照的选手阵容状态
- 最终统计弹窗或表格
- 当前回放位置中所有可见击杀/炸弹事件的地图标记
- 炸弹状态表现：至少在事件流和地图中展示下包/拆包/爆炸位置；若实现成本允许，提供炸弹状态小组件展示未掉落、掉落、已下包、拆包中、爆炸等状态
- 直接进入 `/battle` 且没有战报时的 fallback UI

`BattleReportDebugLog` 作为 `#CombatConst.xlsx` 中的业务调试开关，由 `server/internal/match` 读取并通过顶层 `debug_enabled` 返回。该开关不进入 `MatchInput`，也不由 `matchengine` 读取。开启时，BattlePage 在 Console 中输出一次完整原始战报、逐回合摘要与事件/选手表，并随一秒一条的播放进度输出当前事件；关闭时不产生这些调试日志。

备选方案：保留现有单回合动画，只展示最终结果。拒绝原因：这会隐藏新的整场比赛能力。

## 风险 / 取舍

- [响应体过大] 完整比赛战报可能包含数百个事件。-> 保持事件 payload 结构化但紧凑，避免泄漏内部状态；如有需要，未来增加 RPC 裁剪选项。
- [配置内容不完整] 编辑器可能还没有覆盖每个场景所需的全部 Dust2 语义行。-> 服务初始化时校验并缓存地图配置状态；使用严格校验和清晰错误码，并为 `de_dust2` 提供最小种子数据。
- [框架/业务边界漂移] 实现时很容易在引擎里读取 `cfg.Global`。-> 增加测试或 import 检查，防止 `matchengine` 导入 `server/config`。
- [加时规则差异] 不同赛事的加时金钱/回合规则可能不同。-> 用 `RuleSet` 编码规则；默认使用 MR3 block，并保持经济系统关闭。
- [确定性回归] 多回合模拟会增加随机调用顺序风险。-> 按回合/子系统派生 seed，并在随机选择前稳定排序所有候选。
- [客户端复杂度] 整场比赛回放可能变得视觉噪音过多。-> 构建清晰的回放控制器，提供紧凑回合导航、当前回合聚焦和汇总最终统计。

## 迁移计划

1. 增加引擎自有模型类型，同时尽量保留现有 JSON 字段。
2. 在业务层添加 config-to-engine 快照映射，在服务初始化时校验并缓存成功快照或配置错误状态。
3. 在现有 `Simulate` 入口后面，将单回合循环替换为整场比赛推进。
4. 更新 `DebugSimuMatch`，构造默认队伍、固定武器装配、默认 MR12 规则和地图配置快照。
5. 更新 MatchPage loading 文案，使“开始匹配”后展示匹配中、正在模拟比赛、进入战场等阶段反馈。
6. 更新 TypeScript 战报类型和 BattlePage 回放行为。
7. 运行后端单元测试、Go 插件编译验证、前端 typecheck/build，以及目标组件测试。
8. 由于本功能仍位于 `DebugSimuMatch` 后面，回滚较直接：回退变更，或临时在业务代码中使用单回合 `RuleSet`。

## 待后续提案确认

- 后续提案：武器数值是否迁入专门的 `tb_weapon` 表，以及该表应如何与经济/装配系统关联？
- 后续提案：`simuMatchDesign.md` 中 action 生命周期、边上移动和情报传播要在多大程度上从第一版抽象升级为细粒度模拟？
