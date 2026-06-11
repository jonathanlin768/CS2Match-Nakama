## 1. Chat API 封装 (`api/chat.ts`)

- [x] 1.1 创建 `client/src/api/chat.ts`，实现 `joinDMChannel(socket, targetUserId)` — 调用 `socket.joinChat(targetUserId, 2, true, false)` 加入 DirectMessage 持久化频道，返回 `Channel` 对象
- [x] 1.2 实现 `writeChatMessage(socket, channelId, content)` — 调用 `socket.writeChatMessage(channelId, { content })` 发送文本消息，返回 `ChannelMessageAck`
- [x] 1.3 实现 `listChannelMessages(session, channelId, limit?, cursor?, forward?)` — 调用 `client.listChannelMessages(session, channelId, limit, cursor, forward)` 获取频道消息历史，支持分页
- [x] 1.4 实现 `leaveDMChannel(socket, channelId)` — 调用 `socket.leaveChat(channelId)` 退出指定 DM 频道（好友删除时调用）

## 2. Shared Socket Hook (`hooks/useSocket.ts`)

- [x] 2.1 创建 `client/src/hooks/useSocket.ts`，实现模块级 Socket 单例工厂 `getOrCreateSocket(session)` — 首次调用时创建并连接，后续调用返回同一实例
- [x] 2.2 实现自动重连逻辑 — 监听 socket `onclose` 事件，指数退避重连（1s → 2s → 4s → ... → max 30s），重连后自动重新鉴权
- [x] 2.3 实现 `useSocket(session)` React Hook — 返回 `{ socket, status: "connecting" | "connected" | "disconnected" | "reconnecting" | "guest" }`
- [x] 2.4 实现事件监听器管理 — 提供 `addListener`/`removeListener` 机制，确保组件卸载时自动清理该组件注册的监听器。支持的事件类型包含：`onnotification`、`onchannelmessage`、`onchannelpresence`

## 3. Friend DM Hook (`hooks/useFriendDM.ts`)

- [x] 3.1 创建 `client/src/hooks/useFriendDM.ts`，定义类型接口 — `Conversation`（friendUserId, channelId, lastMessage, unreadCount, lastMessageTime, isFriendOnline）、`ChatMessage`（messageId, channelId, senderId, content, createTime, isSelf, sendStatus）
- [x] 3.2 实现 `useFriendDM(session, friends: Friend[])` hook 骨架 — 使用 `useSocket` 获取共享 socket，过滤 `state=FRIEND(0)` 的好友列表
- [x] 3.3 实现自动加入 DM 频道 — 在 `useEffect` 中并行调用 `joinDMChannel` 加入所有好友的 DM 频道（`Promise.allSettled`），将 channelId 存入 conversations map
- [x] 3.4 实现消息历史加载 — 当选中某个会话时，调用 `listChannelMessages` 加载最近 100 条消息，存入本地 `Map<channelId, Message[]>`
- [x] 3.5 实现发送消息 — `sendMessage(friendUserId, content)` 函数：乐观更新本地消息列表（状态 `sending`），调用 `writeChatMessage`，成功后更新为 `sent`，失败后标记为 `failed`
- [x] 3.6 实现实时消息接收（含 self-message 去重）— 向共享 socket 注册 `onchannelmessage` 监听器；收到消息时先按 `message_id` 查重：如果本地已有相同 `message_id`（自己发送后被广播回来），不追加重复消息仅更新服务器时间戳；如果是新消息则追加到会话消息列表，更新 `lastMessage` 和 `unreadCount`
- [x] 3.7 实现分页翻页 — `loadMoreMessages(channelId, cursor)` 函数，加载更早的消息并插入到消息列表头部
- [x] 3.8 实现会话排序和未读管理 — conversations 按 `lastMessageTime` 降序排列；切换会话时将 `unreadCount` 重置为 0；未选中的会话在收到新消息时递增 `unreadCount`
- [x] 3.9 实现 onchannelpresence 监听器 — 注册到共享 socket，追踪每个 DM 频道中对方用户的 Presence；过滤掉自身 Presence 事件（自身从另一设备加入/离开）；根据 `joins`/`leaves` 更新 `conversation.isFriendOnline`；组件卸载时移除监听器
- [x] 3.10 实现好友删除时的 leaveChat 清理 — 在 `useEffect` 中对比 `friends` prop 变化，检测被删除的好友；对被删除好友调用 `leaveDMChannel` 退出频道并清理本地缓存
- [x] 3.11 实现重连后重新加入频道 — 监听 `useSocket` 的 `status` 从 `"reconnecting"` 变为 `"connected"`；重连后重新调用 `joinDMChannel` 加入所有好友的 DM 频道（`joinChat` 状态随连接断开丢失）；重新加载当前选中会话的最近历史消息以补回断连期间错过的消息

## 4. 改造 useFriends 使用共享 Socket

- [x] 4.1 修改 `client/src/hooks/useFriends.ts` — 移除内部 `client.createSocket()` 和独立的 socket 生命周期管理
- [x] 4.2 引入 `useSocket(session)` 替代 — 向共享 socket 注册 `onnotification` 监听器
- [ ] 4.3 验证兼容性 — 确保好友列表加载、好友通知刷新、add/remove 操作均正常工作

## 5. MessagesTab UI 改造

- [x] 5.1 重写 `client/src/pages/profile/MessagesTab.tsx` — 移除所有硬编码 mock 数据（`conversations`、`chatMessages`），引入 `useFriendDM` 和 `useAuth`
- [x] 5.2 实现会话列表渲染 — 从 `conversations` map 渲染左侧会话列表：在线状态圆点（绿色/灰色，来自 `isFriendOnline`）、头像、用户名、lastMessage 摘要、时间、unreadCount 徽章；高亮选中项；按时间排序
- [x] 5.3 实现聊天头部渲染 — 显示当前会话的好友用户名、头像和在线状态指示器（绿色圆点+"在线"/灰色圆点+"离线"，来自 `isFriendOnline`）
- [x] 5.4 实现聊天消息列表渲染 — 从 `messages` map 渲染当前会话的消息：isSelf 消息靠右蓝色气泡、非 isSelf 消息靠左灰色气泡；显示发送者用户名（多用户群聊预留）和时间戳
- [x] 5.5 实现消息输入与发送 — 输入框绑定 Enter 发送（Shift+Enter 换行）、点击发送按钮；空消息阻止发送；发送后清空输入框；发送失败消息显示错误状态和重发按钮
- [x] 5.6 实现历史消息向上翻页 — 监听消息列表滚动到顶部时触发 `loadMoreMessages`；显示加载指示器；保持滚动位置稳定
- [x] 5.7 实现空状态和加载状态 — 无好友时显示"暂无好友会话，请先添加好友"；未选中会话时显示"选择好友开始聊天"；加载中显示 spinner；错误状态显示重试按钮
- [x] 5.8 实现新消息自动滚动 — 当前会话收到新消息时，如用户已在底部则自动滚动到底部；如用户在看历史消息则不滚动，显示"新消息"浮动按钮
- [x] 5.9 移动端适配 — <1024px 时默认显示会话列表（全宽），点击后进入聊天视图（全宽），顶部返回按钮

## 6. 验证

- [x] 6.1 使用 Docker Compose 启动 Nakama 开发环境，验证服务正常运行
- [x] 6.2 注册两个测试账号 A 和 B，互相添加好友
- [x] 6.3 验证 DM 聊天流程 — A 在 MessagesTab 向 B 发送消息 → B 实时收到推送并显示在 MessagesTab → B 回复消息 → A 实时收到
- [x] 6.4 验证消息历史持久化 — 刷新页面后重新进入 MessagesTab，历史消息正确加载
- [x] 6.5 验证边界情况 — 无好友时的空状态、网络断开时的错误提示、重连后的消息恢复
- [x] 6.6 TypeScript 编译验证 — 运行 `npx tsc --noEmit` 确保无类型错误
