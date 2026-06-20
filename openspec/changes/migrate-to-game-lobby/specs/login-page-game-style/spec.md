# Login Page Game Style

登录/注册页的游戏风格重绘规格。保持 Nakama 认证行为不变，仅变更 UI 表现。

## ADDED Requirements

### Requirement: LoginPage 使用游戏风格布局

LoginPage SHALL 使用与游戏大厅一致的 1600×900 固定画幅暗色面板风格，替代现有的响应式 Web 表单布局。

#### Scenario: 登录页渲染

- **WHEN** 用户访问 `/`
- **THEN** 页面渲染居中 1600×900 画幅
- **AND** 画幅内显示游戏品牌标识、登录/注册表单区、操作按钮
- **AND** 整体视觉与 `/home` 游戏大厅保持一致

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
