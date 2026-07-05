## Context

- `doc/CS2Match_SimuMatch_Module_Design.md` 已给出完整的模拟比赛设计，但项目当前只实现了 Nakama 登录、HealthCheck RPC 和游戏大厅 UI；`client/src/hooks/useSimStream.ts` 仍是 stub。
- `doc/project-structure.md` 明确了服务器应采用子系统 + 框架能力层的分层：`internal/match` 作为业务子系统暴露 RPC，`internal/framework/matchengine` 作为纯净推演引擎被 `match.Service` 调用。
- 当前服务器模块 `windypath.com/cs2match/server` 下还没有 `internal/` 目录，所有业务代码都在 `main.go` 和 `config/` 中。本次 change 需要首次引入 `internal/` 分层。
- 客户端已有 `MatchPage`（含「开始匹配」按钮）和 `BattlePage`（基于 mock 数据 `battleState` 的对战 UI），可以临时复用作调试入口。
- 现有 Luban 配置已导出 `TbPlayer`（选手属性），但还没有路线/地图配置表。本次 change 先把路线结构定义在引擎代码中，后续再迁移到 Luban。

## Goals / Non-Goals

**Goals:**

- 实现 `DebugSimuMatch` RPC，完成「客户端按钮触发 → 服务器推演 → 返回完整战报」的最小链路。
- 定义并落地 `internal/match` 与 `internal/framework/matchengine` 的包结构和依赖方向。
- 定义前后端共用的战报数据结构（`GameEvent`、`RoundResult`、`MatchResult`、`PlayerState`、`PlayerMatchStats`）。
- 复用 `TbPlayer` 读取选手属性，验证 Luban 配置在战斗链路中的可用性。
- 实现最简“比大小”战斗逻辑，能稳定输出首杀、击杀序列、回合胜负和赛后统计。
- 客户端通过「开始匹配」按钮调用 RPC，并以实时回放形式展示战报（事件播报、比分牌、阵容、地图击杀点、数据统计弹窗）。

**Non-Goals:**

- 不实现完整战斗公式（命中率、护甲减伤、补枪、残局加成、位置标签加成）。
- 不实现多回合循环、半场换边、加时、真实玩家阵容输入。
- 不使用 Nakama Match Handler 或 WebSocket 实时同步。
- 不把战报保存到 Storage 或数据库。
- 不实现战后持久化统计界面（本阶段 `EventFeed` 内的「数据统计」Dialog 仅展示当前回合 K/D/A）。

## Decisions

### 1. 服务器分层：`internal/match` + `internal/framework/matchengine`

**Decision:** 严格按照 `doc/project-structure.md` 组织代码：

- `internal/match` 是业务子系统，包含 RPC DTO、Service、Repository（本阶段 repository 仅留接口位置，不持久化）。
- `internal/framework/matchengine` 是框架能力层，不包含 `api_rpc.go`，只暴露 `Service.Simulate(input) (*MatchResult, error)`。

**Rationale:** 与项目架构文档保持一致，避免后续把战斗逻辑和 Nakama RPC 混在一起，也方便未来把 `matchengine` 接入 Match Handler 的 `MatchLoop`。

### 2. 一请求一实例的 `MatchEngine`

**Decision:** `matchengine.Service.Simulate()` 内部创建 `NewMatchEngine(input)`，完成推演后由 GC 回收。

**Rationale:** 状态隔离、可测试、可回放；MVP 阶段没有并发共享状态问题。

### 3. 简化战斗：比大小

**Decision:** 单回合中：

- 随机选一条进攻路线；
- 将 T 方 5 人与 CT 方 5 人按某种简单规则（如 `Entry` 降序）配对；
- 每对比较“综合战力”或单一属性，高者击杀低者；
- 若 T 方全灭则 CT 胜，若 CT 方全灭则 T 胜，否则按剩余人数决定（T 剩余多→T 胜，否则 CT 胜），保证必有结果。

**Rationale:** 本次 change 的核心目标是验证协议和结构，不是战斗平衡性。比大小足够产生稳定的击杀事件、比分和统计。

### 4. 路线配置先用代码常量，后续迁移到 Luban

**Decision:** MVP 阶段在 `matchengine` 中定义 `RouteConfig` Go 结构和一张内建 `dust2AttackRoutes` 切片，包含 6 条进攻路线的 ID、名称、目标包点、基础推进时间。

**Rationale:** 项目当前 Luban 只有 `TbPlayer`，新增 Luban 表需要改 `__tables__.xlsx`、定义 XML、导表脚本等，超出本次“先跑通链路”的范围。路线结构一旦稳定，下一 change 即可迁移到 `TbRoute` 配表并同步生成前端类型。

### 5. 前端临时入口：MatchPage「开始匹配」按钮

**Decision:** 保留 `MatchPage` 的 UI，点击「开始匹配」时先调用 `DebugSimuMatch`，拿到响应后导航到 `/battle`；`BattlePage` 用 `useEffect` + `setInterval` 每 1 秒推进一条事件，驱动 `Scoreboard`、`TeamRoster`、`MapView`、`EventFeed` 等组件实时回放。无战报时仍回退到 `battleState` mock UI。

**Rationale:** 复用现有页面，避免为临时调试新建页面；真实战报直接替换视觉占位，能更直观地验证协议字段和回放节奏。

### 6. 请求 / 响应 DTO 放在 `internal/match/model.go`

**Decision:** `DebugSimuMatchRequest` / `DebugSimuMatchResponse` 放在 `internal/match/model.go`；引擎输入输出 `MatchInput` / `MatchResult` 放在 `internal/framework/matchengine/model.go`。

**Rationale:** 符合“RPC DTO 属于暴露 RPC 的子系统，引擎输入输出属于引擎自身”的约定。

### 7. 击杀位置使用结构化比例坐标

**Decision:** `GameEvent.Location` 由字符串改为 `{ name, x, y }` 对象，`x`/`y` 为 0.0 ~ 1.0 的雷达图比例坐标；引擎在 `const.go` 中为每条 Dust2 进攻路线预定义坐标，并通过 `RandomRouteLocation` 添加小幅随机偏移。

**Rationale:** 前端 `MapView` 需要精确坐标才能叠加击杀标记；比例坐标不受容器大小影响，随机偏移可避免所有标记重叠。

### 8. EventFeed 内嵌实时数据统计弹窗

**Decision:** `EventFeed` 标题栏右侧提供「数据统计」按钮，点击弹出 shadcn Dialog，展示 A/B 两队选手的 K/D/A 表格；数据直接复用 `BattlePage` 已构建的 `BattleTeam` 对象，因此会随击杀事件实时更新。

**Rationale:** 在 MVP 阶段以最小成本提供赛后/局中统计视角，无需新增独立页面。

## Risks / Trade-offs

- **[Risk] `internal/` 目录首次引入，需要调整 Go 模块编译路径** → Mitigation：保持 `server/go.mod` 模块名为 `windypath.com/cs2match/server`，内部包路径使用 `windypath.com/cs2match/server/internal/match`，`build.sh` 编译命令不变。
- **[Risk] 简化战斗逻辑无法验证真实平衡性** → Mitigation：本 change 明确只验证结构，战斗平衡在后续 change 中通过完整公式和大量模拟验证。
- **[Risk] 路线配置硬编码导致前后端不一致** → Mitigation：前端本次也使用一份临时 TypeScript 常量，并在设计中记录迁移到 Luban 的步骤。
- **[Risk] BattlePage 同时存在 mock UI 和真实战报，易造成混淆** → Mitigation：`BattlePage` 优先使用真实战报驱动完整回放 UI；仅在无战报时回退到 `battleState` mock UI，避免两者同时生效。
- **[Risk] 一次性返回完整战报体积未知** → Mitigation：一回合战报很小（<10KB），本阶段不存在性能问题；后续多回合时评估是否需要分页/流式。

## Migration Plan

- 开发阶段：本地 `docker compose up -d` 启动 Nakama，编译 `.so` 后调用 `DebugSimuMatch`。
- 验证方式：前端点击「开始匹配」→ 浏览器控制台打印战报 → 确认 JSON 字段与 `TbPlayer` 数据一致；观察 `BattlePage` 比分、阵容、地图击杀标记、事件播报与数据统计弹窗随回放同步更新。
- 回滚：删除 `internal/match` 和 `internal/framework/matchengine` 目录，从 `main.go` 取消注册 RPC 即可恢复。
- 后续演进：
  1. 把 `RouteConfig` 迁移到 Luban `TbRoute` 表；
  2. 在 `matchengine` 中替换为完整战斗公式；
  3. 增加多回合状态机；
  4. 新增正式 `StartMatch` RPC 并接入 Nakama Matchmaker。

## Open Questions

1. 简化“比大小”规则是否只比较 `Entry`，还是使用 `Entry + Aim + Firepower` 的加权和？建议先用加权和，使不同选手有区分度。
