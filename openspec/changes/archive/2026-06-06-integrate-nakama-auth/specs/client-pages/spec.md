## MODIFIED Requirements

### Requirement: LoginPage 登录页面

系统 SHALL 提供 LoginPage，包含邮箱/密码输入框、登录按钮、社交登录入口和注册链接。页面加载时先尝试从 localStorage 恢复 Nakama Session，成功则自动跳转到 `/home`。恢复失败则展示登录表单，用户提交邮箱密码后调用 Nakama `authenticateEmail` API 进行真实认证。登录失败时显示错误提示，不跳转。

#### Scenario: 渲染登录表单

- **WHEN** 用户访问 `/` 且无法从 localStorage 恢复有效 Session
- **THEN** 页面显示 CS2 SIMULATOR 品牌标识
- **AND** 显示邮箱输入框和密码输入框
- **AND** 显示"登录"按钮、"记住我"复选框和"忘记密码"链接

#### Scenario: Session 恢复成功自动跳转

- **WHEN** 用户访问 `/` 且 localStorage 中存在有效 token
- **AND** `restoreSession()` 返回有效 Session
- **THEN** 页面自动跳转到 `/home`（不显示登录表单）

#### Scenario: Session 恢复中显示加载状态

- **WHEN** 用户访问 `/` 且系统正在尝试恢复 Session（检查 token 有效期或刷新 token）
- **THEN** 页面显示加载指示器（spinner + "正在恢复登录..."）
- **AND** 不显示登录表单

#### Scenario: 点击登录按钮

- **WHEN** 用户点击"登录"按钮
- **THEN** 按钮显示加载状态（spinner + "登录中..."）
- **AND** 前端调用 `loginWithEmail(email, password)` 发起 Nakama 认证
- **AND** 认证成功后导航到 `/home`

#### Scenario: 登录失败显示错误

- **WHEN** 用户点击"登录"按钮后 Nakama 返回认证错误（密码错误或网络异常）
- **THEN** 按钮恢复为正常状态
- **AND** 登录表单上方显示错误提示信息（如"邮箱或密码错误"）
- **AND** 页面不跳转，用户可重新输入

#### Scenario: 密码显示切换

- **WHEN** 用户点击密码输入框右侧的眼睛图标
- **THEN** 密码从隐藏状态切换为明文显示
- **AND** 图标从 Eye 切换为 EyeOff
