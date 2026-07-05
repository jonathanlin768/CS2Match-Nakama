## 1. 后端：建立 matchengine 框架能力层

- [x] 1.1 在 `server/internal/framework/matchengine/` 创建 `model.go`，定义 `MatchInput`、`MatchResult`、`RoundResult`、`GameEvent`、`PlayerState`、`PlayerMatchStats`、`RouteConfig`、`Team`、`Combatant` 等结构。
- [x] 1.2 在 `server/internal/framework/matchengine/` 创建 `const.go`，定义 Dust2 进攻路线数组、回合时间上限、默认武器等常量。
- [x] 1.3 在 `server/internal/framework/matchengine/` 创建 `engine.go`，实现 `MatchEngine` 结构、`NewMatchEngine(input)` 和 `StartMatch()`，完成随机路线选择、简化“比大小”对决、事件生成、回合胜负判定。
- [x] 1.4 在 `server/internal/framework/matchengine/` 创建 `service.go`，实现 `Service` 及 `Simulate(ctx, *MatchInput) (*MatchResult, error)` 工厂方法。
- [x] 1.5 编写 `server/internal/framework/matchengine/engine_test.go`，验证相同 seed 下路线选择一致、对决高属性选手获胜、一回合输出结构正确。

## 2. 后端：建立 match 业务子系统

- [x] 2.1 在 `server/internal/match/` 创建 `model.go`，定义 `DebugSimuMatchRequest`、`DebugSimuMatchResponse`、错误响应结构以及子系统内部模型。
- [x] 2.2 在 `server/internal/match/` 创建 `service.go`，实现 `Service`：硬编码两套 5 人阵容、调用 `cfg.GetPlayer` 读取属性、构造 `matchengine.MatchInput`、调用引擎、返回战报。
- [x] 2.3 在 `server/internal/match/` 创建 `api_rpc.go`，实现 `RPCDebugSimuMatch(service)`：解析 payload、校验 `map_id`、从 context 读取 userID 做认证检查、调用 service、返回 JSON。
- [x] 2.4 在 `server/internal/match/` 创建 `repository.go`（占位），定义 `Repository` 结构和构造函数，本阶段不实现 Storage 写入。
- [x] 2.5 更新 `server/main.go`：实例化 `match.Service`，注册 `DebugSimuMatch` RPC。

## 3. 后端：编译与联调验证

- [x] 3.1 运行 `bash server/build.sh` 编译 `backend.so`，确认无编译错误。
- [x] 3.2 启动本地 `docker compose up -d`，确认 Nakama 加载插件成功且日志输出 `DebugSimuMatch RPC registered`。
- [x] 3.3 使用已认证 Session 调用 `POST /v2/rpc/DebugSimuMatch`，验证返回 JSON 包含 `match_info`、`rounds[0].events`、`final_stats` 和 `winner`。
- [x] 3.4 验证错误场景：传入不支持的 `map_id` 时返回 `INVALID_MAP`；引擎异常时返回 `SIMULATION_ERROR`。

## 4. 前端：类型与 API 层

- [x] 4.1 在 `client/src/types/sim.ts`（或新增 `client/src/types/match-report.ts`）定义 `MatchReport`、`GameEvent`、`RoundReport`、`PlayerState`、`PlayerMatchStats` 等 TypeScript 类型，与服务端字段对齐。
- [x] 4.2 在 `client/src/api/` 新增 `simu.ts`，封装 `debugSimuMatch(session, mapId)` 函数，调用 `client.rpc(session, "DebugSimuMatch", { map_id })`。
- [x] 4.3 运行 `npx tsc --noEmit` 确认新增类型无错误。

## 5. 前端：调试入口与战报展示

- [x] 5.1 修改 `client/src/pages/MatchPage.tsx`：点击「开始匹配」按钮时调用 `debugSimuMatch`，进入加载状态，成功后把战报通过 router state 传递到 `/battle` 并导航。
- [x] 5.2 修改 `client/src/pages/BattlePage.tsx`：挂载时从 location state 读取战报并 `console.log` 完整对象；若不存在则回退到现有 `battleState` mock。
- [x] 5.3 在 `BattlePage` 中新增最简 HLTV 事件列表组件（如右下角 `SimuKillFeed`），渲染 `ROUND_START`、`KILL`、`ROUND_END` 事件文本。
- [x] 5.4 在 `BattlePage` 中保留现有 `Scoreboard`、`TeamRoster`、`MapView`、`KillFeed` mock UI，并加 TODO 注释说明后续替换为真实战报。
- [x] 5.5 处理 RPC 错误：调用失败时恢复按钮状态，显示 Toast 或提示文本“模拟战斗失败，请重试”。

## 6. 前端：构建与运行时验证

- [x] 6.1 运行 `npm run build`（在 `client/` 目录），确认无 TypeScript / 构建错误。
- [x] 6.2 启动前端开发服务器，登录后进入 `/match`，点击「开始匹配」，确认浏览器控制台输出战报 JSON。
- [x] 6.3 确认 `BattlePage` 的最简事件列表正确显示击杀和回合结束信息。
- [x] 6.4 验证网络断开或服务端错误时页面显示友好提示且按钮可重试。

## 7. 文档与收尾

- [x] 7.1 在 `server/internal/framework/matchengine/` 和 `server/internal/match/` 中添加包级注释，说明职责和调用关系。
- [x] 7.2 更新 `doc/CS2Match_SimuMatch_Module_Tasks.md` 中本阶段的完成状态。
- [x] 7.3 运行 `openspec status --change simumatch-mvp-mvp` 确认所有 artifact 已完成。

> 注：6.2 ~ 6.4 为浏览器运行时验证，需在前端开发服务器启动后手动点击「开始匹配」按钮完成。代码层面已确保 API 调用、状态传递与错误处理正确。

## 8. 实时战报播放与阵容同步（追加需求）

- [x] 8.1 后端 `MatchInfo` 增加 `team_a_name` / `team_b_name` 并在 `match.Service` 中填写真实队名。
- [x] 8.2 `BattlePage` 使用 `useEffect` + `setInterval` 每 1 秒推进一条事件，实现“正在实时在打”的播报节奏。
- [x] 8.3 新增 `EventFeed` 组件，按时间顺序显示已播放的 `ROUND_START` / `KILL` / `ROUND_END` 事件。
- [x] 8.4 `Scoreboard` 读取服务器返回的队名，并在 `ROUND_END` 事件出现后更新为当前比分。
- [x] 8.5 `TeamRoster` / `PlayerCard` 读取服务器阵容，根据已播放的击杀事件动态将死亡选手置灰、存活选手保持彩色。
- [x] 8.6 前端类型 `MatchInfo` 同步增加队名字段；`npm run build` 通过。
- [x] 8.7 `EventFeed` 标题栏右侧增加「数据统计」按钮，点击弹出 shadcn 风格 Dialog，展示 A/B 两队 K/D/A 表格；数据随击杀事件与回合结束实时更新。

## 9. 地图与击杀点位（追加需求）

- [x] 9.1 `MapView` 组件渲染 `client/public/csmaps/de_dust2_radar_trans.webp`，使用 `object-contain` 保证整张地图可见。
- [x] 9.2 后端 `GameEvent.Location` 由字符串改为 `Location { name, x, y }` 结构，x/y 为地图比例坐标。
- [x] 9.3 后端为每条进攻路线预定义地图坐标，击杀事件使用 `RandomRouteLocation` 生成带随机偏移的阵亡位置。
- [x] 9.4 `MapView` 接收 `events`，为每个带 `location` 的 `KILL` 事件渲染红色 ×，位置随事件推进逐条出现。
- [x] 9.5 `BattlePage` 将 `visibleEvents` 传给 `MapView`。
- [x] 9.6 后端重新编译、Nakama 重启、前端 `npm run build` 通过。

> 注：6.2 ~ 6.4 以及 8.x / 9.x 的浏览器动画效果，仍需在 `npm run dev` 环境下手动点击「开始匹配」验证。
