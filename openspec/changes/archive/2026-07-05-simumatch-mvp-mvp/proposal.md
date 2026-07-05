## Why

`doc/CS2Match_SimuMatch_Module_Design.md` 已定义模拟比赛系统的完整 MVP 机制，`doc/CS2Match_SimuMatch_Module_Tasks.md` 把实现拆成了三步。当前项目 Nakama 插件只注册了 `HealthCheck`，前端 `useSimStream` 还是 stub，尚未验证**开战请求 → 服务器推演 → 战报返回 → 客户端消费**这条核心链路。

本 change 聚焦第一步「MVP 的 MVP」：先不追求完整战斗公式，而是用最小化的方式把链路跑通，重点固化四类结构：

1. 前后端通讯协议（请求 / 响应 / 战报事件）。
2. 选手与地图路线的配置结构（复用 Luban `TbPlayer`，引入路线配置）。
3. 模拟器引擎接收外部请求和返回结果的接口结构（`MatchInput` / `MatchResult`）。
4. 客户端消费战报所需的数据结构。

具体的胜负计算先用“比大小”的简化逻辑实现，确保能稳定输出战报和统计信息即可。完整的战斗公式、多回合、半场换边、实时 Match Handler 留到后续 change。

## What Changes

- 在 Nakama Go Plugin 中新增一个测试 RPC `DebugSimuMatch`（放在 `internal/match` 子系统，由 `InitModule` 注册），接收客户端请求后，使用硬编码的两套阵容执行一回合简化模拟对战。
- 按照 `doc/project-structure.md` 的服务器结构组织代码：
  - `server/internal/match/api_rpc.go`：仅做协议转换，调用 `match.Service`。
  - `server/internal/match/service.go`：业务编排，构造 `matchengine.MatchInput`、保存/打印战报。
  - `server/internal/match/model.go`：定义 RPC DTO（`DebugSimuMatchRequest` / `DebugSimuMatchResponse`）和子系统模型。
  - `server/internal/framework/matchengine/service.go`：引擎入口工厂。
  - `server/internal/framework/matchengine/engine.go`：单场比赛实例，执行一回合推演。
  - `server/internal/framework/matchengine/model.go`：引擎输入输出结构（`MatchInput` / `MatchResult` / `RoundResult` / `GameEvent` / `PlayerState` 等）。
- 复用现有 Luban 配表 `TbPlayer` 作为选手属性来源；若 `tb_route` 配表尚未定义，则在 `server/internal/framework/matchengine` 中用常量/内建结构先定义 Dust2 的 6 条进攻路线，并同步生成前端 TypeScript 类型与数据（或先写死在前端代码中），后续再迁移到 Luban。
- 最简版模拟器引擎：
  - 服务端收到请求后随机选择一条进攻路线；
  - 对决采用简化“比大小”规则（例如比较双方选手的 `Entry` 或综合属性），胜者存活、败者死亡；
  - 生成 `ROUND_START`、`KILL`、`ROUND_END` 等事件即可；
  - 输出一回合的比分、事件序列、选手状态、赛后统计。
- 前端临时复用 `client/src/pages/MatchPage.tsx` 中的「开始匹配」按钮（当前点击后 `navigate("/battle")`）：点击时先调用 `DebugSimuMatch` RPC，拿到战报后进入 `BattlePage`；`BattlePage` 以 1 秒/条的节奏回放事件，驱动 `Scoreboard`、`TeamRoster`、`MapView`、`EventFeed` 展示真实队名、比分、存活/死亡状态、地图击杀标记与事件播报。无战报时仍回退到 `battleState` mock UI。
- 明确不实现：多回合循环、半场换边、真实玩家阵容输入、完整战斗公式（命中率 / 护甲 / 补枪概率 / 残局加成）、HLTV 特殊前缀、Match Handler 实时同步、Storage 持久化。

## Capabilities

### New Capabilities

- `simu-match-rpc`：`internal/match` 子系统新增 `DebugSimuMatch` RPC，负责请求解析、阵容校验、调用引擎、返回标准化 JSON 战报。
- `simu-engine-core`：`internal/framework/matchengine` 框架能力层新增一回合推演能力，定义 `MatchInput` / `MatchResult` 及事件结构，采用简化“比大小”战斗计算。
- `simu-config-structures`：复用 Luban `TbPlayer` 选手属性，新增（或先常量定义）Dust2 进攻路线配置结构，明确前后端共用的选手 / 路线 / 战斗常数数据结构。
- `simu-client-replay`：前端新增调试入口，发起 RPC 并渲染 / 输出 HLTV 风格战报，验证事件序列与数据结构。

### Modified Capabilities

- 无现有 spec 的需求变更，仅新增能力。

## Impact

- **Nakama 后端**：
  - 新增 `server/internal/match/` 子系统（`api_rpc.go`、`service.go`、`model.go`）。
  - 新增 `server/internal/framework/matchengine/` 框架能力层（`service.go`、`engine.go`、`model.go`）。
  - `server/main.go` 的 `InitModule` 中实例化 `match.Service` 并注册 `DebugSimuMatch` RPC。
  - 复用 `windypath.com/cs2match/config` 子模块读取 `TbPlayer`；Go module 路径保持 `windypath.com/cs2match/server`。
- **React 前端**：
  - 临时复用 `client/src/pages/MatchPage.tsx` 的「开始匹配」按钮作为调试入口，点击后调用 `DebugSimuMatch` 再进入 `BattlePage`。
  - 在 `client/src/api/` 或 `client/src/simu/` 新增 `debugSimuMatch` 调用函数。
  - `BattlePage` 以 1 秒/条节奏回放事件，驱动 `Scoreboard`、`TeamRoster`、`MapView`、`EventFeed` 等组件展示真实队名、比分、存活/死亡状态、地图击杀标记与事件播报；无战报时回退到 `battleState` mock UI。
  - 新增或复用 `client/src/types/sim.ts` 中的战报类型。
- **数据库 / Storage**：无变更，MVP 阶段阵容硬编码在 `match.Service` 或请求体中，不写入 Storage。
- **Luban 配表**：复用 `configs/Datas/#Player.xlsx` 导出的 `TbPlayer`；路线配置若未在 Luban 中定义，则先用代码内常量承载，并在设计中说明迁移路径。
- **部署 / CI/CD**：无变更，新增代码随现有 `build.sh` / GitHub Actions 流程编译为 `.so`。
- **状态同步**：仍采用 RPC 请求 / 响应模式，不引入 Match Handler 实时推送。
