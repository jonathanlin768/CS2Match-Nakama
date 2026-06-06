## Why

LoginPage 当前使用 `setTimeout` 模拟登录（1 秒延迟后直接跳转），未接入 Nakama 的认证系统。前端已有 `@heroiclabs/nakama-js` SDK 和设备认证基础设施，但缺少邮箱登录和 Session 恢复能力。本次变更让用户通过真实账号密码登录，为后续的玩家数据持久化和社交功能奠定基础。

## What Changes

- **新增** `client/src/api/auth.ts`，封装邮箱登录（`loginWithEmail`）和 Session 恢复（`restoreSession`）两个核心函数
- **改造** `client/src/nakama.ts`，保留 Nakama 客户端单例，移除未使用的设备认证逻辑
- **改造** `client/src/hooks/useNakamaAuth.ts`，从纯设备认证 hook 改为 Session 恢复 + 邮箱登录状态管理的 hook
- **改造** `client/src/pages/LoginPage.tsx`，将 `handleLogin` 从假登录替换为真实 Nakama email auth 调用，增加错误状态显示
- **改造** `client/src/main.tsx`，添加 AuthProvider 包裹路由，实现"已登录自动跳转 /home，未登录停留在 /"的守卫逻辑

## Capabilities

### New Capabilities

- `client-auth`: Nakama 邮箱认证 API 封装、Session 恢复、登录状态管理，前端认证流程

### Modified Capabilities

- `client-pages`: LoginPage 从 mock 登录改为真实 Nakama 认证，增加加载/错误状态和交互反馈
- `react-frontend-scaffold`: 认证 hook 从纯设备认证改为 Session 恢复优先 + 邮箱登录的模式

## Impact

| 模块 | 影响说明 |
|------|---------|
| **React 前端** | `client/src/api/auth.ts` 新增；`LoginPage.tsx`、`useNakamaAuth.ts`、`nakama.ts`、`main.tsx` 改造 |
| **Nakama 后端** | 无影响 — 使用 Nakama 内置 email auth API，无需新增 Go 代码或 RPC |
| **数据库** | 无 schema 变更 — Nakama 自动管理 `users` 表（邮箱、密码哈希）和 `storage` 表（Session） |
| **部署** | 无影响 |
| **Luban 配置** | 无影响 |
