## Why

ProfilePage 的好友功能当前使用硬编码的 `mockFriends` 数据 — 好友列表、添加好友、搜索、分组展开都是前端静态模拟，无法反映真实的好友关系和社交状态。Nakama 内置了完整的好友系统（添加/删除/列表查询），但前端尚未接入。本次变更将 ProfilePage 的好友部分从 mock 数据切换到 Nakama 真实好友 API，让用户可以看到真实的好友列表并进行添加、删除等社交操作。

## What Changes

- **新增** `client/src/api/friends.ts`，封装 Nakama 好友 API（`listFriends`、`addFriends`、`deleteFriends`）
- **新增** `client/src/hooks/useFriends.ts`，管理好友列表状态、加载/错误状态、搜索过滤和分组逻辑
- **改造** `client/src/pages/ProfilePage.tsx`，好友部分从 `mockFriends` 替换为 `useFriends` hook 返回的真实数据，删除 `mockFriends` 和 `friendGroups` 硬编码数据
- **新增** `AddFriendDialog` 组件（在 ProfilePage 内或独立组件），提供按用户名搜索并添加好友的交互
- 好友分组基于 Nakama Friend State：Friend（已添加）、Invite Sent（已发送请求）、Invite Received（收到的请求）
- 好友上线状态由 Nakama User 的 `online` 字段提供（需 WebSocket 连接以获取实时状态）
- 搜索通过 Nakama `listFriends` 返回数据做前端过滤

## Capabilities

### New Capabilities

- `client-friends`: Nakama 好友 API 封装（列表查询、添加、删除）、好友状态管理 hook、添加好友交互

### Modified Capabilities

- `client-pages`: ProfilePage 好友部分从 mock 数据改为调用真实 Nakama Friend API，增加加载/空状态/错误处理，增加添加好友对话框

## Impact

| 模块 | 影响说明 |
|------|---------|
| **React 前端** | `client/src/api/friends.ts` 新增；`client/src/hooks/useFriends.ts` 新增；`client/src/pages/ProfilePage.tsx` 改造（移除 mock 数据，接入真实 API） |
| **Nakama 后端** | 无影响 — 使用 Nakama 内置 Friend API（`listFriends`、`addFriends`、`deleteFriends`），无需新增 Go 代码或 RPC |
| **数据库** | 无 schema 变更 — Nakama 自动管理 `user_edge` 表（好友关系） |
| **部署** | 无影响 |
| **Luban 配置** | 无影响 |
