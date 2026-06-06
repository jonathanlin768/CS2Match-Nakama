## ADDED Requirements

### Requirement: 邮箱登录 `loginWithEmail`

系统 SHALL 提供 `loginWithEmail(email, password)` 函数，调用 Nakama SDK 的 `client.authenticateEmail(email, password, true)` 进行邮箱认证。`create=true` 表示账号不存在时自动创建，实现"登录即注册"。

#### Scenario: 首次登录自动注册

- **WHEN** 用户使用未注册的邮箱和密码调用 `loginWithEmail`
- **THEN** Nakama 自动创建新账号并返回有效 Session
- **AND** Session 的 `token` 和 `refresh_token` 被保存到 localStorage（key 为 `nakama_token` 和 `nakama_refresh`）

#### Scenario: 已有账号登录

- **WHEN** 用户使用已注册的邮箱和正确密码调用 `loginWithEmail`
- **THEN** Nakama 返回有效 Session
- **AND** Session 信息被持久化到 localStorage

#### Scenario: 密码错误

- **WHEN** 用户使用已注册的邮箱但错误的密码调用 `loginWithEmail`
- **THEN** 函数抛出包含认证失败信息的 Error
- **AND** 函数不修改 localStorage 中的 token

#### Scenario: 服务器不可达

- **WHEN** Nakama 服务器不可达（网络断开或容器未启动）
- **AND** 用户调用 `loginWithEmail`
- **THEN** 函数抛出包含网络错误信息的 Error

### Requirement: Session 恢复 `restoreSession`

系统 SHALL 提供 `restoreSession()` 函数，从 localStorage 读取已持久化的 token 和 refresh_token，尝试恢复有效的 Nakama Session。

#### Scenario: 从有效 token 恢复

- **WHEN** localStorage 中存在未过期的 `nakama_token`
- **AND** 调用 `restoreSession()`
- **THEN** 返回有效的 Session 对象
- **AND** 不发起网络请求

#### Scenario: 从过期 token + 有效 refresh_token 恢复

- **WHEN** localStorage 中的 token 已过期但 refresh_token 仍然有效
- **AND** 调用 `restoreSession()`
- **THEN** 系统使用 refresh_token 向 Nakama 请求新的 token
- **AND** 返回有效的 Session 对象
- **AND** 更新 localStorage 中的 token 和 refresh_token

#### Scenario: token 和 refresh_token 均失效

- **WHEN** localStorage 中的 token 和 refresh_token 均已失效
- **AND** 调用 `restoreSession()`
- **THEN** 返回 `null`
- **AND** 清除 localStorage 中的 `nakama_token` 和 `nakama_refresh`

#### Scenario: localStorage 中无 token

- **WHEN** localStorage 中不存在 `nakama_token`
- **AND** 调用 `restoreSession()`
- **THEN** 返回 `null`
