## Context

当前 `useFriends` hook 在组件内部创建独立 WebSocket 连接用于好友通知监听。`MessagesTab.tsx` 完全使用硬编码 mock 数据，不连接任何 Nakama API。本次变更需要引入 Nakama Channel/DirectMessage 聊天能力，需要一个共享的 WebSocket 连接来同时承载好友通知和聊天消息推送。

Nakama 内置 Chat API 完全满足需求：`joinChat`（type=2 DirectMessage, persistence=true）加入持久化私聊频道、`writeChatMessage` 发送消息、`listChannelMessages` 分页加载历史、`socket.onchannelmessage` 实时推送、`socket.onchannelpresence` 频道在线状态感知、`socket.leaveChat` 频道退出。不需要 Go 插件改动。`updateChatMessage`（消息编辑）和 `removeChatMessage`（消息删除）为本阶段明确的 Non-Goals。

## Goals / Non-Goals

**Goals:**
- 玩家可以与已添加的好友（state=FRIEND）进行一对一实时私聊
- 历史消息持久化到 Nakama PostgreSQL，支持向前翻页
- 新消息通过 WebSocket 实时推送到客户端
- MessagesTab 展示真实的会话列表（按最近消息时间排序）和聊天界面
- 共享 WebSocket 连接，同时服务好友通知和聊天推送
- 通过 `onchannelpresence` 感知频道中对方用户的在线/离线状态

**Non-Goals:**
- 不支持群聊（本阶段仅 1v1 DM）
- 不支持非好友私聊（UI 层不暴露入口，后端不额外限制）
- 不支持消息的已读/未读状态同步到服务器（仅本地 UI 展示）
- 不修改 Nakama 后端 Go 插件
- 不支持图片/文件/表情等富媒体消息（UI 按钮保留但仅实现文本）
- 不支持消息删除/撤回（`removeChatMessage`）
- 不支持消息编辑（`updateChatMessage`）

## Decisions

### Decision 1: Shared Socket Singleton

**选择**：模块级 Socket 单例（`nakama.ts` 中导出 `getSocket()` 工厂），`useSocket` hook 仅包装 React 响应式状态。

**为什么**：
- 当前 `useFriends` 创建独立 socket，聊天需要另一个 socket → 浪费连接
- Socket 单例模式与现有 `client` 单例一致，风格统一
- 模块级单例避免了 React Context 的额外嵌套和性能开销
- 100 人同时在线的小项目不需要复杂的连接池

**替代方案考虑**：
- React Context (`SocketProvider`)：对项目来说过度设计，且 socket 生命周期绑定到 Provider 而非真正需要时创建
- 每个 hook 独立 socket：浪费 Nakama 连接资源，且通知重复

### Decision 2: DM Channel ID 约定

**选择**：使用 Nakama 内置的 `joinChat(targetUserId, ChannelType.DirectMessage, true, false)` — Nakama 自动为每对用户生成确定性 DM 频道 ID。

**为什么**：
- Nakama 内部使用排序后的两个用户 ID 生成 DM channel_id，保证双方 join 同一个频道
- `persistence=true` 使消息持久化到 PostgreSQL
- `hidden=false` 使频道在用户列表中可见（未来可扩展）

### Decision 3: 仅好友可聊天的前端强制

**选择**：会话列表完全由好友列表（state=FRIEND）派生，MessagesTab 不提供"搜索用户发起聊天"功能。

**为什么**：
- 符合产品需求：私聊仅限好友之间
- Nakama DM 本身不限制非好友通信 — 前端作为唯一入口点来控制
- 如果未来需要服务端校验，可以加 RPC Hook，但当前阶段不需要

### Decision 4: useFriendDM Hook 设计

**选择**：单一 `useFriendDM(session, friends)` hook 管理所有 DM 聊天状态，而非拆分为多个小 hook。

**为什么**：
- DM 聊天涉及频道加入、消息发送、历史加载、实时推送，这些操作共享同一状态（`Map<friendUserId, Message[]>`）
- 单个 hook 内部管理状态一致性更简单，避免跨 hook 状态同步问题
- `useFriends` 已经证明了这种"胖 hook"模式在项目中运作良好

### Decision 5: 改造 useFriends 的 Socket 使用

**选择**：`useFriends` 改为使用共享 socket（`useSocket`），而非内部创建独立 socket。

**为什么**：
- 共享 socket 意味着只剩一个 WebSocket 连接
- `useFriends` 的 `onnotification` 监听器注册到共享 socket
- 向后兼容：`useFriends` 的公开 API 不变

### Decision 6: onchannelpresence — 频道在线状态感知

**选择**：在 `useFriendDM` 中注册 `socket.onchannelpresence` 监听器，追踪每个 DM 频道中对方用户的 Presence 状态（joined/left）。

**为什么**：
- Nakama 的 `ChannelPresenceEvent` 提供 `joins` 和 `leaves` 数组，包含 `user_id`、`username`、`session_id`
- 可用于显示"对方在线"/"对方正在输入"等状态指示器
- 在 MessagesTab 的会话列表和聊天头部显示在线状态（绿色/灰色圆点）
- 相比 `followUsers`（需要单独 follow 每个用户），`onchannelpresence` 更轻量且自动覆盖所有已加入频道

**替代方案考虑**：
- `socket.followUsers` + `onstatuspresence`：需要显式 follow/unfollow 每个好友，增加复杂度，且与频道绑定关系不直接

### Decision 7: leaveChat — 频道退出时机

**选择**：仅在好友被删除时调用 `leaveChat` 退出 DM 频道；MessagesTab 组件卸载时不移出频道（保持频道在线以接收后台推送）。

**为什么**：
- 用户需要在未打开 MessagesTab 时也能收到新消息推送
- 仅在社交关系解除（删除好友）时才需要退出频道
- `useFriendDM` 在检测到 friends 列表缩减时（有好友被删除），对该好友调用 `leaveChat`

**替代方案考虑**：
- 组件卸载时 leaveChat：会导致用户在其他页面时无法收到消息推送

## Risks / Trade-offs

- **[风险] 共享 socket 单点故障**：如果 socket 断连，好友通知和聊天推送同时中断 → **缓解**：`useSocket` 实现自动重连（指数退避，1s→2s→4s→...，最大 30s）
- **[风险] Nakama DM 频道无服务端权限控制**：非好友理论上也能发消息到 DM 频道 → **缓解**：前端强制过滤，不在 UI 中暴露非好友入口；如未来需要服务端校验，可通过 Go RPC Hook 拦截 `channel_message_send`
- **[权衡] 消息数据存于 hook 本地 state**：切换好友/页面会丢失消息状态 → **缓解**：每次选中好友时重新 `listChannelMessages` 加载历史，实时消息在 socket 监听器中追加；这避免了全局状态管理开销，且符合"消息持久化在服务器"的设计
- **[风险] 首次加入频道可能延迟**：好友多时需要逐个 joinChat → **缓解**：并行 join（`Promise.allSettled`），UI 非阻塞显示

## Open Questions

- 是否需要在好友详情面板（FriendsTab）增加"发消息"按钮，快捷跳转到 MessagesTab 并选中该好友？（当前设计文档中已移除该按钮，可后续需求再定）
- 未读消息计数是否需要跨会话持久化到 localStorage？（当前阶段仅内存中计数，页面刷新后丢失）
