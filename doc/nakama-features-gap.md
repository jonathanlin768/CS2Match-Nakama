# Nakama 功能接入状态对比分析

> 基于 Nakama 官方文档与项目实际代码的对比分析
> 生成日期: 2026-06-08

---

## 项目当前已接入的 Nakama 功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 邮箱认证 (`authenticateEmail`) | ✅ 已接入 | 登录/注册/Session 恢复 |
| Session 管理 | ✅ 已接入 | token + refresh_token 持久化与自动刷新 |
| Go Plugin 运行时 | ✅ 已接入 | `server/main.go` 编译为 `.so` 加载 |
| RPC 框架 | ✅ 已接入 | 仅注册了 `HealthCheck` |
| Nakama Console | ✅ 已接入 | `:7351` 管理后台随 Docker 启动 |
| 配置表加载 (Luban) | ⚠️ 非 Nakama 功能 | 通过 Go embed 实现，与 Nakama 无关 |

---

## 尚未接入但项目需要的 Nakama 功能

按**优先级/核心程度**排列：

### 🔴 核心玩法必需（阻塞上线）

| Nakama 功能 | 项目哪里需要 | 当前缺失 |
|-------------|-----------|---------|
| **Match Handler（实时对战）** | `MatchPage` 的 `useSimStream` 需要 WebSocket 状态同步 | Hook 目前是 stub，`connected` 永远为 `false`；后端未注册任何 Match Handler |
| **Matchmaker（匹配系统）** | 1v1 教练对战需要把两个玩家匹配到一起 | 文档提到需要，但前后端均未接入 |
| **Storage Engine（存储引擎）** | 用户金币、钻石、等级、抽卡记录、选手阵容、道具库存 | 所有页面数据均为硬编码；后端未调用 `nk.StorageWrite/Read` |
| **Leaderboards（排行榜）** | `RankingPage` 需要真实排行数据 | 前端 UI 就绪，但未接入 `nk.Leaderboard` API |

### 🟡 重要功能（影响体验）

| Nakama 功能 | 项目哪里需要 | 当前缺失 |
|-------------|-----------|---------|
| **User Account / Profile** | `Header` 显示金币/`PlayerSidebar` 显示用户信息；需要拉取真实玩家资料 | 未调用 `client.getAccount()`，未使用 Nakama 的 user metadata |
| **In-App Notifications（应用内通知）** | 任务完成、赛季奖励、抽卡结果等提示 | 未接入 `nk.NotificationsSend` / `socket.onnotification` |
| **WebSocket Socket 连接** | 实时对战数据流、通知推送 | 前端 `nakama.ts` 只有 Client 单例，未创建 Socket |
| **Hooks（Before/After）** | 注册时校验、敏感操作拦截、防刷 | 文档提到需要封禁检查，但 `InitModule` 中未注册任何 Hook |

### 🟢 已规划但未实现

| Nakama 功能 | 项目哪里需要 | 当前缺失 |
|-------------|-----------|---------|
| **OAuth 认证** | `LoginPage` 有 Google/GitHub 按钮但无响应 | 未接入 `authenticateGoogle` / `authenticateSteam` 等 |
| **Friends（好友系统）** | 社交社区、1v1 邀请对战 | 前端无对应页面，未使用 Friends API |
| **Groups/Clans（群组/战队）** | 战队排行、战队系统 | 前端有 `TeamRankings` 组件但无后端支撑 |
| **Chat（实时聊天）** | 对战内教练交流、全局聊天室 | 未接入 Chat API |

### ⚪ 可选/暂时不需要

| Nakama 功能 | 项目是否需要 | 说明 |
|-------------|-----------|------|
| **Tournaments（锦标赛）** | 未来可能需要 | 当前实验项目阶段，1v1 匹配即可满足 |
| **Parties（组队/小队）** | 暂时不需要 | 1v1 教练对战，无组队需求 |
| **Purchase Validation（内购验证）** | 暂时不需要 | 实验项目，无真实支付 |
| **Relayed Multiplayer** | 不需要 | 项目采用 Server-Authoritative（状态同步），不需要 Relay 模式 |

---

## 一句话总结

> 你的项目目前**只搭好了 Nakama 的壳**（认证 + RPC 框架），核心游戏功能所需的 **Match Handler、Matchmaker、Storage、Leaderboards、Socket 连接** 全部尚未接入。下一步最优先的是实现 `Match Handler` + `Socket`（让对战跑起来），其次是 `Storage`（让玩家数据持久化）。
