# Client Chat

CS2 模拟器前端 WebSocket 聊天系统规格 — 共享 Socket 连接、DM 频道管理、实时消息收发。

## Purpose

定义客户端聊天相关的 API 封装、Hook 和 WebSocket 事件管理。useSocket 提供模块级共享 Socket 单例，useFriendDM 管理好友私聊的完整生命周期（加入频道、发送消息、实时接收、在线状态感知）。
## Requirements
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

### Requirement: 客户端以结构化类型解析社交卡片
客户端 SHALL 只将受支持版本的服务端结构化消息解析为联系方式交换卡片。卡片内容 SHALL 以 `request_id` 查询 Social RPC 获得的权威状态为准；未知、畸形或客户端自造消息 SHALL NOT 被渲染为已授权状态。

#### Scenario: 收到有效请求卡片
- **GIVEN** Socket 收到包含受支持 `type`、`request_id`、`action` 和 `version` 的服务端消息
- **WHEN** 社交事件 hook 处理该消息
- **THEN** 客户端按 request ID 获取最新交换状态
- **AND** 渲染待接受、已接受、已拒绝或已失效卡片

#### Scenario: 收到未知消息类型
- **GIVEN** DM 历史或实时事件包含未知 `type`
- **WHEN** 客户端解析消息
- **THEN** 客户端忽略该消息或显示不可操作的兼容提示
- **AND** 不把其正文当作聊天文本展示

### Requirement: 交换卡片历史和实时事件复用共享 Socket
客户端 SHALL 继续复用现有单例 Socket、DM 加入、历史分页、重连和多监听器分发能力，但会话摘要和未读状态 SHALL 从受支持的结构化卡片派生。重连后 SHALL 通过 RPC 刷新卡片状态，不能仅相信本地缓存。

#### Scenario: 断线后恢复
- **GIVEN** 用户查看好友交换卡片时 Socket 断线
- **WHEN** 共享 Socket 重连并重新加入 DM 频道
- **THEN** 客户端恢复历史与实时监听
- **AND** 对可见 request ID 重新读取权威交换状态

#### Scenario: 非当前好友收到卡片
- **GIVEN** 用户正在查看好友 A
- **WHEN** 好友 B 的频道收到有效交换卡片
- **THEN** 好友 B 的事件摘要和未读数更新
- **AND** 当前好友 A 的详情不被错误替换

### Requirement: 客户端不暴露任意频道写入操作
生产客户端 SHALL 不导出或调用面向 UI 的 `writeChatMessage(text)` 能力，也 SHALL NOT 渲染文本、多媒体或自定义 payload 输入控件。发起和响应交换只能调用 Social RPC。

#### Scenario: 用户发起交换
- **GIVEN** 当前选中对象是正式好友
- **WHEN** 用户点击“请求交换联系方式”并确认渠道
- **THEN** 客户端调用 `SocialRequestContactExchange`
- **AND** 不调用 Socket `writeChatMessage`

#### Scenario: 服务端拒绝交换
- **GIVEN** 好友关系已经失效但客户端尚未刷新
- **WHEN** Social RPC 返回 `NOT_FRIENDS`
- **THEN** 卡片显示请求不可用并刷新好友状态
- **AND** 客户端不回退到发送普通频道消息
