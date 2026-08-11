# login-page-game-style Specification

## Purpose
TBD - created by archiving change migrate-to-game-lobby. Update Purpose after archive.
## Requirements
### Requirement: LoginPage 使用游戏风格布局
登录与注册 SHALL 使用新响应式 App Shell 中的 Modal 或窄页表单，不再使用固定 1600×900 画幅。表单 SHALL 在约 390px 手机视口和桌面视口中完整可读、可滚动且不产生固定画布裁剪。

#### Scenario: 访客从好友功能触发登录
- **GIVEN** Guest 用户尝试进入需要正式账号的好友社交页面
- **WHEN** 登录 Modal 打开
- **THEN** Modal 展示游戏品牌一致的登录/注册表单
- **AND** 背景保留当前页面上下文

#### Scenario: 手机端登录
- **GIVEN** 视口宽度约为 390px
- **WHEN** 登录/注册表单显示
- **THEN** 所有字段、错误和主要操作可在视口内阅读与触控
- **AND** 不通过缩放固定桌面画布适配

### Requirement: 保留登录/注册页签切换

LoginPage SHALL 保留登录与注册两种模式，并支持页签切换。

#### Scenario: 切换页签

- **WHEN** 用户点击登录/注册页签
- **THEN** 表单区切换为对应模式
- **AND** 当前页签高亮显示
- **AND** 切换时清除当前表单错误提示

### Requirement: 保留邮箱登录行为

登录表单 SHALL 继续使用 `useAuth().login(email, password)` 进行 Nakama 邮箱认证。

#### Scenario: 登录成功

- **WHEN** 用户输入正确的邮箱和密码并提交
- **THEN** 按钮显示加载状态
- **AND** 认证成功后自动跳转 `/home`

#### Scenario: 登录失败

- **WHEN** 用户提交错误的邮箱或密码
- **THEN** 表单上方显示中文错误提示（"邮箱或密码错误"或"无法连接服务器"）
- **AND** 页面不跳转

### Requirement: 保留邮箱注册行为

注册表单 SHALL 继续使用 `useAuth().register(email, password)` 进行 Nakama 邮箱注册。

#### Scenario: 注册校验

- **WHEN** 用户两次输入的密码不一致
- **THEN** 显示"两次输入的密码不一致"
- **WHEN** 密码长度不足 8 位
- **THEN** 显示"密码长度至少需要8位"

#### Scenario: 注册成功

- **WHEN** 用户输入新邮箱、两次一致且长度足够的密码并提交
- **THEN** 创建新账号并自动登录
- **AND** 跳转 `/home`

#### Scenario: 重复注册

- **WHEN** 用户输入已存在的邮箱并提交
- **THEN** 显示"该邮箱已注册，请直接登录"
- **AND** 不跳转

### Requirement: 保留 Session 恢复加载状态

LoginPage SHALL 在应用初始化尝试恢复 Nakama Session 时显示加载状态。

#### Scenario: Session 恢复中

- **WHEN** `useAuth().status === "restoring"`
- **THEN** 页面显示加载指示与"正在恢复登录..."提示
- **AND** 不显示登录/注册表单

#### Scenario: Session 恢复成功自动跳转

- **WHEN** `useAuth().status` 从 restoring 变为 authenticated
- **THEN** 页面自动跳转 `/home`

### Requirement: 保留密码显示切换

登录/注册表单的密码输入框 SHALL 支持显示/隐藏密码。

#### Scenario: 切换密码可见性

- **WHEN** 用户点击密码输入框右侧的眼睛图标
- **THEN** 密码在隐藏与明文之间切换
- **AND** 图标在 Eye 与 EyeOff 之间切换

### Requirement: 登录页无篮球主题命名

LoginPage 的源码、文案、资源路径 SHALL 不包含 "nba"、"basketball" 等篮球相关命名。

#### Scenario: 源码检查

- **WHEN** 对 `src/pages/LoginPage.tsx` 执行大小写不敏感搜索
- **THEN** 不存在 "nba"、"basketball" 字符串

### Requirement: 认证成功返回原目标
登录/注册 Modal SHALL 记录触发它的目标位置，并在认证成功、Session 刷新完成后返回该目标；取消 SHALL 回到原页面且不破坏 Guest Session。

#### Scenario: 登录后返回好友页
- **GIVEN** 访客从好友辅助入口打开登录 Modal
- **WHEN** 邮箱登录成功
- **THEN** Modal 关闭并导航到好友页
- **AND** 新 Session username 已反映系统八位 code

