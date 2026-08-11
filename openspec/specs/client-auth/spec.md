# Client Auth

Nakama 邮箱认证与 Session 管理的 API 封装 — `loginWithEmail`、`registerWithEmail`、`restoreSession`、`clearSession`。

## Purpose

封装 Nakama SDK 的认证 API，提供登录、注册、Session 恢复和清除功能。登录与注册分离：`loginWithEmail` 仅对已有账号认证（`create=false`），`registerWithEmail` 自动创建新账号（`create=true`）并返回 `created` 标志。成功后持久化 token/refresh_token 到 localStorage。
## Requirements
### Requirement: 邮箱登录 `loginWithEmail`

系统 SHALL 提供 `loginWithEmail(email, password)` 函数，调用 Nakama SDK 的 `client.authenticateEmail(email, password, false)` 进行邮箱认证。`create=false` 表示仅登录，不自动创建账号。

#### Scenario: 已有账号登录

- **WHEN** 用户使用已注册的邮箱和正确密码调用 `loginWithEmail`
- **THEN** Nakama 返回有效 Session
- **AND** Session 的 `token` 和 `refresh_token` 被保存到 localStorage（key 为 `nakama_token` 和 `nakama_refresh`）

#### Scenario: 密码错误

- **WHEN** 用户使用已注册的邮箱但错误的密码调用 `loginWithEmail`
- **THEN** 函数抛出包含认证失败信息的 Error
- **AND** 函数不修改 localStorage 中的 token

#### Scenario: 服务器不可达

- **WHEN** Nakama 服务器不可达（网络断开或容器未启动）
- **AND** 用户调用 `loginWithEmail`
- **THEN** 函数抛出包含网络错误信息的 Error

### Requirement: 邮箱注册 `registerWithEmail`

系统 SHALL 提供 `registerWithEmail(email, password)` 函数，调用 Nakama SDK 的 `client.authenticateEmail(email, password, true)` 进行注册。`create=true` 表示账号不存在时自动创建。返回 `{ session, created }`，其中 `created: boolean` 表示是否为新创建账号。

#### Scenario: 首次注册

- **WHEN** 用户使用未注册的邮箱和密码调用 `registerWithEmail`
- **THEN** Nakama 自动创建新账号并返回有效 Session
- **AND** `created` 为 `true`
- **AND** Session 的 `token` 和 `refresh_token` 被保存到 localStorage

#### Scenario: 重复注册

- **WHEN** 用户使用已注册的邮箱调用 `registerWithEmail`
- **THEN** Nakama 返回该账号的 Session（不创建新账号）
- **AND** `created` 为 `false`
- **AND** Session 信息被持久化到 localStorage

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

### Requirement: 清除 Session `clearSession`

系统 SHALL 提供 `clearSession()` 函数，清除 localStorage 中的 `nakama_token` 和 `nakama_refresh`。

#### Scenario: 用户退出当前账号

- **GIVEN** localStorage 中保存了 Nakama token 或 refresh token
- **WHEN** 客户端调用 `clearSession()`
- **THEN** `nakama_token` 和 `nakama_refresh` 均被移除

### Requirement: 静默设备认证提供访客 Session
客户端 SHALL 在没有可恢复 Session 时生成并持久化合法设备 ID，通过 Nakama device authentication 创建或恢复 Guest Session。该过程 SHALL 不要求用户先填写邮箱、密码或昵称。

#### Scenario: 首次打开游戏
- **GIVEN** 浏览器没有 Session token 和设备 ID
- **WHEN** 应用初始化
- **THEN** 客户端生成 10 至 128 字节的设备 ID 并完成 device authentication
- **AND** 将返回 Session 标记为访客身份

### Requirement: 系统生成八位唯一 username
新账号的 Nakama username SHALL 由项目认证逻辑生成为恰好 8 位大写英文字母与数字 code。客户端 SHALL NOT 接受玩家自定义 username 或 display name；创建冲突时 SHALL 重新生成并有限次重试。

#### Scenario: 创建访客账号
- **GIVEN** device ID 尚未关联 Nakama 用户
- **WHEN** 服务器创建账号
- **THEN** 账号 username 符合 `^[A-Z0-9]{8}$`
- **AND** UI 将其格式化为 `玩家#CODE`

#### Scenario: 客户端尝试改名
- **GIVEN** 玩家已经拥有系统 username
- **WHEN** 修改客户端提交不符合规范的 account update
- **THEN** before-hook 拒绝请求
- **AND** username 保持不变

### Requirement: 正式登录不保存教学阵容
邮箱注册、账号链接或登录已有账号 SHALL 只迁移身份与 Session，不得把客户端临时 15 元教学阵容写入账号 Storage。

#### Scenario: 教学过程中登录
- **GIVEN** Guest Session 已选择临时教学阵容
- **WHEN** 用户完成邮箱注册或登录
- **THEN** 客户端切换到正式身份并恢复目标页面
- **AND** 服务端不创建正式阵容记录
