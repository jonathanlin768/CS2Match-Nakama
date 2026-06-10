## Context

ProfilePage 的 `case "friends"` 分支目前使用 `mockFriends`（9 条硬编码数据）和 `friendGroups`（基于 `friend.group` 字段分组），完全不与 Nakama 通信。用户登录后看到的好友列表是假的，搜索和添加好友按钮也无实际功能。

Nakama 内置了完整的好友系统，通过 `@heroiclabs/nakama-js` SDK 可调用：
- `client.listFriends(session, state?, limit?, cursor?)` — 列出好友，支持按状态过滤和分页
- `client.addFriends(session, ids?, usernames?)` — 通过 ID 或用户名添加好友
- `client.deleteFriends(session, ids?, usernames?)` — 删除好友
- `client.listFriendsOfFriends(session, limit?, cursor?)` — 列出好友的好友

好友状态定义：`Friend_FRIEND(0)`、`Friend_INVITE_SENT(1)`、`Friend_INVITE_RECEIVED(2)`。`BLOCKED(3)` 因 Nakama 不支持解除屏蔽后恢复好友关系，本变更不启用。

本次变更仅涉及前端代码，使用 Nakama 内置 API，不需要修改 Go 插件。

## Goals / Non-Goals

**Goals:**
- 好友列表从 mock 数据切换为 Nakama `listFriends` 真实数据
- 实现添加好友功能（通过用户名搜索并发送好友请求）
- 实现**完整的好友请求生命周期**：发送请求、接受请求、拒绝请求、取消已发送请求
- 实现删除好友功能
- 好友按 Nakama Friend State 分组展示：已添加(0)、已发送(1)、已接收(2)
- 支持搜索过滤（前端基于用户名/displayName 过滤）
- 处理加载状态（loading）、空状态（无好友）和错误状态（网络异常/Nakama 不可达）
- **详情面板的操作按钮根据 Friend State 动态切换**（而非当前始终显示"编辑资料"+"发消息"）
- **"收到的请求"分组显示未处理请求数量徽章**

**Non-Goals:**
- 不实现导入 Facebook/Steam 好友的功能（Nakama 内置 API 已支持，但按钮不接入）
- 不实现好友在线状态实时推送（需要 WebSocket，后续版本通过 `online` 字段 + socket 连接实现）
- 不实现"好友的好友"页面（`listFriendsOfFriends` API 封装到位但不在 UI 中使用）
- 不修改 Go 插件代码
- 不改动好友详细资料面板的信息结构（Nakama User 没有 signature、birthday、level、likes 等字段，这些保持 mock 或显示为 N/A）
- 不支持分组管理（Nakama 原生不支持自定义好友分组，后续可通过 Storage 实现）

## Decisions

### 1. 文件组织：沿用 `api/auth.ts` 模式

**选择**:
```
client/src/
  api/friends.ts          ← 新建：封装 Nakama Friend API 纯函数
  hooks/useFriends.ts     ← 新建：React hook，管理好友状态
  pages/ProfilePage.tsx   ← 改造：删除 mock 数据，使用 useFriends
```

**理由**:
- `api/friends.ts` 是纯函数模块，与 React 解耦，方便测试和复用
- `useFriends.ts` hook 负责 React 状态管理（loading/error/data）
- 前端 API 层遵循 `api/auth.ts` 的命名和导出模式

**备选方案**: 直接在 hook 中调用 `client.xxx()` — 混杂了 API 调用和状态管理，违背关注点分离原则

### 2. 好友分组：基于 Nakama Friend State 而非自定义分组

**选择**: 将 Nakama Friend State 映射为分组：
- State 0 (`FRIEND`) → "我的好友"
- State 1 (`INVITE_SENT`) → "已发送请求"
- State 2 (`INVITE_RECEIVED`) → "收到的请求"
**理由**:
- Nakama 原生支持这些状态（`user_edge` 表内置）
- 无需自定义 Storage 或额外 RPC
- v0 实验阶段，自定义分组（特别关心、朋友、最近组队）价值不大
- **注意**：Nakama 的 `BLOCKED`(3) 状态被移除。因为 Nakama 不支持"解除屏蔽后恢复好友"，`blockFriends` → `deleteFriends` 会永久删除好友关系，不符合国内用户对"屏蔽"的理解

**备选方案**: 使用 Nakama Storage 存储自定义好友分组 — 需要在 Go 端新建 RPC 来管理分组，前端 v0 阶段过度设计

### 3. 添加好友：使用 `addFriends` + `usernames` 参数

**选择**: 弹出对话框让用户输入好友用户名，调用 `client.addFriends(session, undefined, [username])`

**理由**:
- 用户名比 UUID 更友好（用户不会记得 UUID）
- Nakama `addFriends` 同时支持 `ids` 和 `usernames`，传递用户名即可
- Nakama 会自动发送好友请求通知（`INVITE_SENT` → 对方看到 `INVITE_RECEIVED`）

**备选方案**: 通过 Nakama Storage 搜索用户 — Nakama 没有全局用户搜索 API，无法实现"搜索所有用户"功能。因此只能让用户手动输入精确的用户名

### 4. 在线状态：使用 Nakama User `online` 字段

**选择**: 从 `listFriends` 返回的 `Friend.user.online` 字段读取在线状态（boolean），映射为 UI 状态：`online={true}` → "在线"、`online={false}` → "离线"、`state=3`（ingame 不可知）

**理由**:
- Nakama REST API 返回的 User 包含 `online: boolean`，注意此字段仅在通过 WebSocket 连接时准确
- v1 不使用 WebSocket，`online` 字段可能不准确，但仍比完全伪造更有价值
- 后续可通过 Nakama Socket 连接实时更新在线状态

**备选方案**: 自定义 Go RPC 跟踪在线用户 — 需要维护额外的用户状态表，v1 阶段过度设计

### 5. 好友详情面板：保留 UI 结构但映射真实字段

**选择**: 保留 ProfilePage 右侧好友详情面板的 UI 结构，将可用字段映射到 Nakama User：
- `username` → 昵称
- `id` → 用户 ID
- `display_name` → 显示名（可编辑？v1 不改造）
- `avatar_url` → 头像 URL
- `location` → 所在地
- `online` → 在线状态
- `update_time` → 最近活跃时间

不存在的字段（signature、level、gender、age、birthday、likes）显示为 `-` 或隐藏。

**理由**: 保持 UI 一致性，减少额外的 UI 改动。v1 只聚焦"接通数据"，不做 UI 重构。

### 6. 状态管理：使用 `useState` + `useCallback`（与现有 `useNakamaAuth` 模式一致）

**选择**: 在 `useFriends` 中使用 React 原生 hooks，不引入 Zustand/React Query

**理由**:
- 好友列表是单页面数据，不需要全局状态共享
- 与现有 `useNakamaAuth` 模式一致，保持项目统一性
- 减少外部依赖

### 7. 好友请求生命周期：接受/拒绝/取消

**选择**: 利用 Nakama 的双向添加机制管理请求生命周期：
- **发送请求**: 用户 A 调用 `addFriends(usernames=[B])` → A 侧状态 `INVITE_SENT`，B 侧状态 `INVITE_RECEIVED`
- **接受请求**: 用户 B 对收到的请求调用 `addFriends(usernames=[A])` → 双方状态变为 `FRIEND`（双向确认）
- **拒绝请求**: 用户 B 调用 `deleteFriends(ids=[A])` → 删除待处理的好友关系边
- **取消请求**: 用户 A 调用 `deleteFriends(ids=[B])` → 删除自己发出的待处理请求

**理由**:
- Nakama 好友系统基于"相互添加"：双方都添加对方后才成为真正的 FRIEND
- 单向添加 = INVITE_SENT / INVITE_RECEIVED，双向添加 = FRIEND
- 不需要额外的 Go RPC，`addFriends` 和 `deleteFriends` 两个 API 即可覆盖全部生命周期
- 与主流社交平台的好友逻辑一致

**备选方案**: 通过自定义 Go RPC 在服务端实现 accept/reject 逻辑 — 过度设计，Nakama 内置机制已满足需求

### 8. 动态操作按钮：根据 Friend State 切换详情面板按钮

**选择**: ProfilePage 右侧好友详情面板底部的操作按钮不再固定为"编辑资料"+"发消息"，而是根据选中好友的 `state` 动态渲染：

| Friend State | 按钮 1 | 按钮 2 |
|-------------|--------|--------|
| FRIEND (0) | "删除好友"（危险按钮） | — |
| INVITE_SENT (1) | "取消请求"（次要按钮） | — |
| INVITE_RECEIVED (2) | "接受"（主按钮，强调色） | "拒绝"（次要按钮） |

**理由**:
- 当前 mock 的"编辑资料"和"发消息"按钮不是真实功能，必须移除
- 不同状态下的操作语义不同：FRIEND 需要删除，INVITE_RECEIVED 需要接受/拒绝
- 保持详情面板的 UI 结构不变，仅替换底部的按钮区域

**备选方案**: 将所有操作放在三点菜单（ContextMenu）中 — 会降低操作可见性，不如直接按钮直观

### 9. "收到的请求"分组徽章

**选择**: "收到的请求"分组标题旁边显示红色数字徽章（`Badge`），表示未处理的请求数量。当数量为 0 时整个分组自动隐藏（等同于展开空分组）。

**理由**:
- 提醒用户有待处理的好友请求，提升好友请求的响应率
- 分组为空时自动隐藏，保持界面整洁

## Risks / Trade-offs

- **[风险] `online` 字段在 REST API 中不可靠** → v1 作为参考值展示，标注为"可能不准确"。后续通过 WebSocket 连接校正
- **[风险] 输入错误用户名导致添加失败** → Nakama `addFriends` 会返回错误，前端捕获并展示提示
- **[风险] 好友数量多时（100+）前端过滤搜索可能卡顿** → v1 使用前端过滤，`listFriends` 默认传 `limit=100` 确保一次加载全部好友（无需分页），100 人在线规模下足够。后续如需支持更大规模，可改用后端 state 过滤 + 分页
- **[权衡] 不支持自定义分组** → 牺牲了原 mock UI 中的"特别关心/朋友/最近组队"分组，换取零后端工作。后续可通过 Nakama Storage 实现
- **[权衡] 不搜索用户列表** → 用户必须输入精确用户名才能添加好友，略为不便，但 Nakama 无全局用户搜索 API，添加搜索需要 Go RPC
- **[风险] Session 过期导致 API 调用失败** → `useFriends` 依赖 AuthContext 中的 session，session 过期时会回退到登录页
- **[风险] 好友详情面板的 mock 字段（level、likes、medals、photo wall）全部移除** → 面板内容比 mock 版本精简，但数据都是真实的。后续可通过 Nakama Storage + 自定义 metadata 恢复这些扩展字段

## Open Questions

- 后续是否需要自定义好友分组（通过 Nakama Storage + Go RPC 实现）？
- 后续是否需要 WebSocket 连接以获取实时在线状态和好友上线通知？
- 后续是否需要通过 Nakama User `metadata` 字段存储扩展资料（签名、生日、等级等），以丰富好友详情面板？
