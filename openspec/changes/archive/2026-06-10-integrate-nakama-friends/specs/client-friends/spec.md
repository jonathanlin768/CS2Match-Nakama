# Client Friends

Nakama 好友 API 封装与状态管理 — `api/friends.ts` 纯函数层 + `useFriends` React Hook。

## Purpose

封装 Nakama 内置 Friend API（listFriends、addFriends、deleteFriends），提供类型安全的纯函数 API 和 React 状态管理 Hook，供 ProfilePage 好友部分接入使用。

## ADDED Requirements

### Requirement: listFriends API 封装

系统 SHALL 提供一个纯函数 `listFriends(session)`，调用 Nakama `client.listFriends()` 获取当前用户的所有好友列表，返回类型为 `Promise<Friend[]>`。

#### Scenario: 成功获取好友列表

- **GIVEN** 用户已通过 Session 认证
- **WHEN** 调用 `listFriends(session)`
- **THEN** 函数调用 `client.listFriends(session)` 并返回 `Friends` 对象
- **AND** 返回结果包含 `friends: Friend[]` 数组和可选的 `cursor` 分页游标

#### Scenario: Session 无效

- **GIVEN** Session 已过期或无效
- **WHEN** 调用 `listFriends(session)`
- **THEN** 函数不会静默失败，调用方捕获 Nakama SDK 抛出的错误
- **AND** 错误信息包含"未授权"或"Session 无效"相关描述

#### Scenario: 好友列表为空

- **GIVEN** 用户尚未添加任何好友
- **WHEN** 调用 `listFriends(session)`
- **THEN** 返回 `friends: []` 空数组，不抛出错误

### Requirement: addFriends API 封装

系统 SHALL 提供一个纯函数 `addFriendsByUsername(session, username)`，调用 Nakama `client.addFriends()` 通过用户名添加好友，返回 `Promise<boolean>`。

#### Scenario: 成功添加好友

- **GIVEN** 用户输入一个存在的用户名
- **WHEN** 调用 `addFriendsByUsername(session, username)`
- **THEN** 函数调用 `client.addFriends(session, undefined, [username])`
- **AND** 返回 `true` 表示操作成功
- **AND** 对方用户将收到一条 `INVITE_RECEIVED` 状态的好友记录

#### Scenario: 用户名不存在

- **GIVEN** 用户输入一个不存在的用户名
- **WHEN** 调用 `addFriendsByUsername(session, username)`
- **THEN** Nakama SDK 抛出错误"用户不存在"（User not found）
- **AND** 调用方可以捕获并展示错误提示

#### Scenario: 重复添加

- **GIVEN** 用户尝试添加一个已经是好友的用户名
- **WHEN** 调用 `addFriendsByUsername(session, username)`
- **THEN** Nakama SDK 抛出错误"好友关系已存在"
- **AND** 调用方可以捕获并展示相应提示

### Requirement: deleteFriends API 封装

系统 SHALL 提供一个纯函数 `deleteFriend(session, userId)`，调用 Nakama `client.deleteFriends()` 删除指定好友，返回 `Promise<boolean>`。

#### Scenario: 成功删除好友

- **GIVEN** 用户有一个现有好友（state=FRIEND）
- **WHEN** 调用 `deleteFriend(session, userId)`
- **THEN** 函数调用 `client.deleteFriends(session, [userId])`
- **AND** 返回 `true`
- **AND** 该好友记录从好友列表中移除

#### Scenario: 删除不存在的用户

- **GIVEN** 传入一个不在好友列表中的 userId
- **WHEN** 调用 `deleteFriend(session, userId)`
- **THEN** Nakama SDK 抛出错误，调用方可以捕获

### Requirement: useFriends Hook

系统 SHALL 提供一个 React Hook `useFriends(session)`，管理好友列表的加载、状态、操作和分组。

#### Scenario: 初始化加载好友列表

- **WHEN** 组件挂载且 `session` 有效
- **THEN** hook 自动调用 `listFriends(session)` 获取好友列表
- **AND** 返回 `status: "loading"` 直到请求完成
- **AND** 请求成功后返回 `status: "success"` 和 `friends: Friend[]`
- **AND** 请求失败后返回 `status: "error"` 和 `error: string`

#### Scenario: 好友按状态分组

- **GIVEN** 好友列表包含不同 state 的记录
- **WHEN** 调用 hook 后数据加载完成
- **THEN** `friendsByState` 按 state 分组返回：
  - `friendsByState[0]` → "我的好友"（state=FRIEND）
  - `friendsByState[1]` → "已发送请求"（state=INVITE_SENT）
  - `friendsByState[2]` → "收到的请求"（state=INVITE_RECEIVED）

#### Scenario: 添加好友后刷新列表

- **GIVEN** 用户通过 hook 的 `addFriend(username)` 方法成功添加了好友
- **THEN** hook 自动重新调用 `listFriends(session)` 以获取最新列表
- **AND** 新好友出现在对应的分组中

#### Scenario: 删除好友后更新列表

- **GIVEN** 用户通过 hook 的 `removeFriend(userId)` 方法成功删除了好友
- **THEN** hook 自动重新调用 `listFriends(session)` 以获取最新列表
- **AND** 被删除的好友不再出现在列表中

#### Scenario: 搜索过滤

- **GIVEN** 好友列表已加载且包含多个好友
- **WHEN** 用户通过 `searchQuery` 传入搜索关键词（用户名或 displayName 部分匹配）
- **THEN** `filteredFriends` 返回匹配的好友列表（前端过滤，大小写不敏感）
- **AND** 没有匹配项时返回空数组

#### Scenario: Session 为 null 时跳过加载

- **GIVEN** 用户未登录（session 为 null）
- **WHEN** 组件挂载
- **THEN** hook 不调用任何 API
- **AND** 返回 `status: "guest"` 和空数据

#### Scenario: 接受好友请求

- **GIVEN** 用户有一条 state=INVITE_RECEIVED 的好友记录
- **WHEN** 调用 `addFriend(senderUsername)`
- **THEN** hook 调用 `addFriendsByUsername(session, senderUsername)` 与对方建立双向好友关系
- **AND** 成功后双方状态变为 FRIEND(0)
- **AND** 自动刷新好友列表

#### Scenario: 拒绝好友请求

- **GIVEN** 用户有一条 state=INVITE_RECEIVED 的好友记录
- **WHEN** 调用 `removeFriend(senderUserId)`
- **THEN** hook 调用 `deleteFriend(session, senderUserId)` 删除待处理请求
- **AND** 该记录从好友列表中移除
- **AND** 自动刷新好友列表

#### Scenario: 取消已发送的好友请求

- **GIVEN** 用户有一条 state=INVITE_SENT 的好友记录
- **WHEN** 调用 `removeFriend(recipientUserId)`
- **THEN** hook 调用 `deleteFriend(session, recipientUserId)` 取消待处理请求
- **AND** 该记录从好友列表中移除
- **AND** 自动刷新好友列表

