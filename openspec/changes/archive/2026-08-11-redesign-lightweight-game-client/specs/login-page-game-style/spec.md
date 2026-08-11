## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: 认证成功返回原目标
登录/注册 Modal SHALL 记录触发它的目标位置，并在认证成功、Session 刷新完成后返回该目标；取消 SHALL 回到原页面且不破坏 Guest Session。

#### Scenario: 登录后返回好友页
- **GIVEN** 访客从好友辅助入口打开登录 Modal
- **WHEN** 邮箱登录成功
- **THEN** Modal 关闭并导航到好友页
- **AND** 新 Session username 已反映系统八位 code
