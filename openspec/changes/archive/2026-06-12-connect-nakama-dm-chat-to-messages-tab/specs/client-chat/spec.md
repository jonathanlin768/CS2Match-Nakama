## ADDED Requirements

### Requirement: useSocket — Shared WebSocket Connection

系统 SHALL 提供一个 `useSocket(session)` React Hook，管理共享的 Nakama WebSocket 连接，返回连接状态和 Socket 实例。Socket 实例为模块级单例，所有 `useSocket` 调用共享同一连接。自动重连采用指数退避策略。

#### Scenario: 首次连接

- **GIVEN** 用户已认证（session 有效）且首次调用 `useSocket(session)`
- **WHEN** hook 初始化
- **THEN** 调用 `client.createSocket()` 创建新 WebSocket 连接
- **AND** 返回 `status: "connecting"`
- **AND** 连接成功后返回 `status: "connected"` 和 `socket` 实例
- **AND** 连接成功后调用 `socket.connect(session, true)` 完成鉴权握手

#### Scenario: 共享已存在的 Socket

- **GIVEN** 模块级 socket 已连接
- **WHEN** 另一个组件/调用 `useSocket(session)`
- **THEN** 返回相同的 `socket` 实例和当前 `status`
- **AND** 不创建新的 WebSocket 连接

#### Scenario: Socket 断连自动重连

- **GIVEN** socket 已连接且正常运行
- **WHEN** WebSocket 连接意外断开
- **THEN** `status` 更新为 `"reconnecting"`
- **AND** 自动尝试重连，间隔按指数退避（1s → 2s → 4s → ... → 最大 30s）
- **AND** 重连成功后 `status` 恢复为 `"connected"`
- **AND** 重连成功后调用者（useFriendDM）需要重新加入所有 DM 频道（`joinChat` 状态随连接断开而丢失）
- **AND** 重连后重新加载当前选中会话的最近消息历史，以补回断连期间错过的消息
- **AND** 原有所有事件监听器（onnotification, onchannelmessage, onchannelpresence）在重连后仍然有效

#### Scenario: Session 为 null 时跳过连接

- **GIVEN** session 为 `null`
- **WHEN** 调用 `useSocket(null)`
- **THEN** 不创建 socket，不发起连接
- **AND** 返回 `status: "disconnected"` 和 `socket: null`

#### Scenario: 组件卸载时 socket 保持

- **GIVEN** socket 已连接且有多个组件通过 `useSocket` 订阅
- **WHEN** 其中一个组件卸载
- **THEN** socket 保持连接状态（其他组件仍依赖）
- **AND** 该组件的 socket 事件监听器（如 onchannelmessage）被正确清理

### Requirement: writeChatMessage API 封装

系统 SHALL 提供一个纯函数 `writeChatMessage(socket, channelId, content)`，通过 WebSocket 向指定频道发送文本消息。

#### Scenario: 成功发送消息

- **GIVEN** socket 已连接且用户已加入 DM 频道
- **WHEN** 调用 `writeChatMessage(socket, channelId, "Hello!")`
- **THEN** 函数调用 `socket.writeChatMessage(channelId, { content: "Hello!" })`
- **AND** 返回一个 ChannelMessageAck（包含 message_id、channel_id、create_time 等）
- **AND** 服务器端消息持久化到 PostgreSQL

#### Scenario: Socket 未连接时发送

- **GIVEN** socket 尚未连接或已断开
- **WHEN** 调用 `writeChatMessage(socket, channelId, "Hello!")`
- **THEN** Nakama SDK 抛出错误
- **AND** 调用方可捕获并提示"消息发送失败，请检查网络连接"

### Requirement: listChannelMessages API 封装

系统 SHALL 提供一个纯函数 `listChannelMessages(session, channelId, options?)`，获取指定频道的消息历史，支持分页。

#### Scenario: 加载消息历史

- **GIVEN** 用户已加入 DM 频道且频道包含历史消息
- **WHEN** 调用 `listChannelMessages(session, channelId)`
- **THEN** 调用 `client.listChannelMessages(session, channelId, 100, undefined, false)`
- **AND** 返回 `ChannelMessageList`（包含 `messages: ChannelMessage[]` 和 `next_cursor`）
- **AND** 消息按时间升序排列（旧的在前，新的在后）

#### Scenario: 向前翻页加载更早消息

- **GIVEN** 频道有大量历史消息（超过一页）
- **WHEN** 调用 `listChannelMessages(session, channelId, { cursor, forward: true })`
- **THEN** 使用 `cursor` 游标返回更早的消息
- **AND** 返回结果包含 `next_cursor` 用于继续翻页

#### Scenario: 频道无历史消息

- **GIVEN** 用户刚加入 DM 频道且无历史消息
- **WHEN** 调用 `listChannelMessages(session, channelId)`
- **THEN** 返回空的 `messages: []` 和 `next_cursor: undefined`
- **AND** 不抛出错误

#### Scenario: Session 无效

- **GIVEN** session 已过期
- **WHEN** 调用 `listChannelMessages(session, channelId)`
- **THEN** Nakama SDK 抛出鉴权错误
- **AND** 调用方可捕获错误

### Requirement: joinChat API 封装

系统 SHALL 提供一个纯函数 `joinDMChannel(socket, friendUserId)`，加入与指定好友的 1v1 DirectMessage 频道。

#### Scenario: 成功加入 DM 频道

- **GIVEN** socket 已连接
- **WHEN** 调用 `joinDMChannel(socket, friendUserId)`
- **THEN** 调用 `socket.joinChat(friendUserId, 2, true, false)`（type=2 DirectMessage, persistence=true, hidden=false）
- **AND** 返回 `Channel` 对象（包含 `id`, `presences`, `self`, `type`, `room_name` 等字段）
- **AND** 如果频道已存在（双方之前聊过），返回相同频道（channel_id 不变）

#### Scenario: 加入频道失败

- **GIVEN** socket 未连接或目标用户 ID 无效
- **WHEN** 调用 `joinDMChannel(socket, friendUserId)`
- **THEN** Nakama SDK 抛出错误
- **AND** 调用方可捕获并处理

### Requirement: useFriendDM Hook

系统 SHALL 提供一个 `useFriendDM(session, friends: Friend[])` React Hook，管理好友私聊的完整状态：自动加入 DM 频道、加载历史消息、发送新消息、实时接收推送。仅处理 `state=FRIEND(0)` 的好友。

#### Scenario: 初始化 — 自动加入频道

- **GIVEN** 用户已认证且有 3 个好友（state=FRIEND）
- **WHEN** `useFriendDM` hook 首次执行且 socket 已连接
- **THEN** 并行调用 `joinDMChannel` 加入每个好友的 DM 频道
- **AND** 返回 `conversations: Conversation[]`（按最近消息时间降序排列）
- **AND** 返回 `status: "loading"` 直到所有频道加入完成

#### Scenario: 加入频道后初始化在线状态与历史

- **GIVEN** 所有频道加入完成
- **WHEN** `status` 变为 `"success"`
- **THEN** 每个 `conversation` 的 `isFriendOnline` 根据 `joinChat` 返回的 `Channel.presences` 初始化（好友的 Presence 在列表中 → `true`，否则 → `false`）
- **AND** 每个 `conversation` 包含 `lastMessage`（最新一条消息摘要）
- **AND** 每个 `conversation` 包含 `unreadCount`（本地计算的未读计数）
- **AND** 当前选中会话 `selectedConversation` 的完整消息列表已加载

#### Scenario: 发送消息

- **GIVEN** 用户选中了好友 A 的会话并输入消息 "Hi!"
- **WHEN** 用户点击发送（或按 Enter）
- **THEN** 调用 `writeChatMessage(socket, channelId, "Hi!")`
- **AND** 消息立即追加到本地消息列表（乐观更新，isSelf=true）
- **AND** 发送成功后消息更新为已确认状态（message_id 已分配）
- **AND** 会话列表的 `lastMessage` 更新为新消息摘要
- **AND** 输入框清空

#### Scenario: 实时接收新消息

- **GIVEN** socket 已连接且已加入好友 A 的 DM 频道
- **WHEN** 服务器通过 WebSocket 推送好友 A 发来的新消息（onchannelmessage）
- **THEN** 如果当前正在查看好友 A 的聊天，消息立即追加到消息列表
- **AND** 如果正在查看其他会话或未选中会话，好友 A 的 `unreadCount` 增加
- **AND** 好友 A 的会话在列表中的位置更新（按最新消息时间排序）
- **AND** 好友 A 的 `lastMessage` 更新

#### Scenario: 切换会话

- **GIVEN** 会话列表中有好友 A 和好友 B
- **WHEN** 用户点击好友 B 的会话
- **THEN** `selectedConversation` 更新为好友 B
- **AND** 右侧聊天区域显示好友 B 的完整消息历史（从本地缓存或加载）
- **AND** 好友 B 的 `unreadCount` 重置为 0

#### Scenario: 好友列表为空

- **GIVEN** 用户没有好友（friends 为空或没有 state=FRIEND 的好友）
- **WHEN** `useFriendDM` hook 执行
- **THEN** `conversations` 为空数组
- **AND** `status` 为 `"success"`（不发起任何 API 调用）
- **AND** 不创建任何 socket 频道连接

#### Scenario: Session 为 null 时挂起

- **GIVEN** session 为 null
- **WHEN** `useFriendDM(null, friends)` 执行
- **THEN** `status` 为 `"guest"`
- **AND** `conversations` 为空数组
- **AND** 不发起任何 API 调用

#### Scenario: 发送消息失败

- **GIVEN** 用户发送消息时网络断开
- **WHEN** `writeChatMessage` 抛出错误
- **THEN** 本地乐观更新的消息显示错误状态（红色感叹号或"发送失败"标记）
- **AND** 提供"重发"按钮

#### Scenario: 频道加入失败容错

- **GIVEN** 用户有 3 个好友，其中 1 个加入频道失败
- **WHEN** 所有 `joinDMChannel` 完成（`Promise.allSettled`）
- **THEN** 成功的 2 个好友出现在会话列表中
- **AND** 失败的 1 个好友不出现，但也不阻塞其他功能
- **AND** 无全局错误提示（静默跳过）

#### Scenario: 好友删除后退出频道

- **GIVEN** 用户有好友 A 的 DM 频道已 join
- **WHEN** 好友 A 从好友列表中被删除（`friends` prop 中不再包含好友 A）
- **THEN** `useFriendDM` 在 `useEffect` 中检测到好友 A 被移除
- **AND** 调用 `leaveDMChannel(socket, channelId)` 退出好友 A 的 DM 频道
- **AND** 好友 A 的会话从 `conversations` 中移除
- **AND** 好友 A 的消息缓存从 `messages` map 中清除

#### Scenario: onchannelpresence 更新在线状态

- **GIVEN** socket 已连接且已注册 onchannelpresence 监听器
- **WHEN** 收到好友 A 的 `joins` Presence 事件
- **THEN** 好友 A 的 `conversation.isFriendOnline` 更新为 `true`
- **WHEN** 收到好友 A 的 `leaves` Presence 事件
- **THEN** 好友 A 的 `conversation.isFriendOnline` 更新为 `false`

### Requirement: leaveChat API 封装

系统 SHALL 提供一个纯函数 `leaveDMChannel(socket, channelId)`，退出指定 DM 频道。

#### Scenario: 成功退出频道

- **GIVEN** socket 已连接且用户已加入 DM 频道
- **WHEN** 调用 `leaveDMChannel(socket, channelId)`
- **THEN** 调用 `socket.leaveChat(channelId)`
- **AND** 返回 Promise（void）
- **AND** 退出后不再收到该频道的 `onchannelmessage` 推送

#### Scenario: 退出未加入的频道

- **GIVEN** 用户未加入某个频道
- **WHEN** 调用 `leaveDMChannel(socket, channelId)`
- **THEN** Nakama SDK 可能静默忽略或抛出错误
- **AND** 调用方不依赖退出结果做关键逻辑判断

### Requirement: Socket onchannelpresence 事件监听

`useSocket` SHALL 支持向共享 socket 注册 `onchannelpresence` 监听器，用于感知频道内用户的 Presence 变化（加入/离开频道）。

#### Scenario: 对方加入频道

- **GIVEN** socket 已连接且 `useFriendDM` 注册了 onchannelpresence 监听器
- **WHEN** 好友 B 加入 DM 频道（上线或打开聊天）
- **THEN** `onchannelpresence` 触发，`joins` 数组包含好友 B 的 `Presence`（含 user_id、username、session_id）
- **AND** `useFriendDM` 更新该会话的 `isFriendOnline` 为 `true`

#### Scenario: 对方离开频道

- **GIVEN** socket 已连接且好友 B 当前在频道中在线
- **WHEN** 好友 B 离开频道（下线或断开连接）
- **THEN** `onchannelpresence` 触发，`leaves` 数组包含好友 B 的 `Presence`
- **AND** `useFriendDM` 更新该会话的 `isFriendOnline` 为 `false`

#### Scenario: 监听器在组件卸载时清理

- **GIVEN** `useFriendDM` 注册了 onchannelpresence 监听器
- **WHEN** 组件卸载
- **THEN** 该监听器从 socket 事件列表中移除
- **AND** 不影响其他组件注册的监听器（如 onnotification、onchannelmessage）

#### Scenario: 重连后监听器仍然有效

- **GIVEN** socket 断连后自动重连成功
- **WHEN** 服务器推送 onchannelpresence 事件
- **THEN** 所有已注册的 onchannelpresence 监听器正常工作

### Requirement: Socket onchannelmessage 事件监听

`useSocket` SHALL 提供向共享 socket 注册 `onchannelmessage` 监听器的机制，并确保监听器在组件卸载时正确移除。

#### Scenario: 注册和移除监听器

- **GIVEN** socket 已连接且 `useFriendDM` 注册了 onchannelmessage 监听器
- **WHEN** `useFriendDM` 所在的组件卸载
- **THEN** 该监听器从 socket 事件列表中移除
- **AND** 不影响其他组件注册的监听器

#### Scenario: 多个监听器共存

- **GIVEN** socket 已连接
- **WHEN** `useFriends` 注册了 onnotification 监听器，`useFriendDM` 注册了 onchannelmessage 监听器
- **THEN** 两个监听器互不干扰
- **AND** onnotification 事件触发时仅调用 useFriends 的回调
- **AND** onchannelmessage 事件触发时仅调用 useFriendDM 的回调

#### Scenario: 自己发送的消息广播回来时去重

- **GIVEN** 用户通过 `writeChatMessage` 发送了一条消息，且该消息已乐观添加到本地消息列表（带临时 ID，status="sending"）
- **WHEN** `writeChatMessage` 的 Promise resolve 返回 `ChannelMessageAck`（含真实 `message_id`）
- **THEN** 本地乐观消息的临时 ID 更新为真实 `message_id`，status 更新为 "sent"
- **WHEN** 随后 Nakama 通过 `onchannelmessage` 将同一消息广播回来（`message_id` 相同）
- **THEN** `useFriendDM` 的 onchannelmessage 处理函数检测到该 `message_id` 已存在于本地消息列表中
- **AND** 不追加重复消息，仅用服务器时间戳（`create_time`）更新本地消息
- **AND** 如果广播在 `ChannelMessageAck` 之前到达（乱序），仍需通过 `message_id` 去重而非仅依赖时序

#### Scenario: 自身 Presence 事件过滤

- **GIVEN** socket 已连接且注册了 onchannelpresence 监听器
- **WHEN** 收到 `onchannelpresence` 事件
- **THEN** 处理函数过滤掉 `user_id === 当前用户ID` 的 Presence（自身从另一设备加入/离开）
- **AND** 仅对好友的 Presence 变化更新 `isFriendOnline`
