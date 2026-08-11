# Client Routing

React Router v7 前端路由系统规格 — 页面导航、路由守卫、Header 联动。

## Purpose

定义前端 SPA 的路由结构、页面组织方式和导航行为。LoginPage 为独立入口（无 AppLayout），其他页面由 AppLayout（Header + Footer）包裹。
## Requirements
### Requirement: React Router 路由系统
前端项目 SHALL 继续使用 react-router-dom v7 的 `createBrowserRouter` 与 `RouterProvider`。`/` SHALL 显示新主界面；`/onboarding/lineup`、`/match/computer`、`/battle` SHALL 允许 Guest Session；`/friends` SHALL 要求正式登录；`/lineup` 与 `/match/friend` SHALL 显示未开放状态。旧 `/home` SHALL 重定向到 `/`，旧非本期玩法路由 SHALL 重定向到主界面或显示已下线状态。

#### Scenario: 访问根路径
- **GIVEN** 身份初始化已经完成
- **WHEN** 用户访问 `/`
- **THEN** 系统显示响应式三入口主界面
- **AND** 不强制先显示独立 LoginPage

#### Scenario: 访客访问教学战
- **GIVEN** 当前身份是有效 Guest Session
- **WHEN** 用户访问 `/onboarding/lineup` 或 `/match/computer`
- **THEN** 路由允许渲染对应页面

#### Scenario: 访客访问好友页
- **GIVEN** 当前身份是 Guest Session
- **WHEN** 用户访问 `/friends`
- **THEN** 路由显示登录/注册 Modal 或认证说明
- **AND** 认证成功后返回 `/friends`

#### Scenario: SPA 子路由回退
- **GIVEN** Nginx 托管生产构建
- **WHEN** 用户直接访问任一有效子路由
- **THEN** Nginx 返回 `index.html`
- **AND** React Router 解析并渲染目标页面

### Requirement: 路由与 Header 导航联动
新 App Shell 的导航 SHALL 反映当前主操作，并提供一致的返回主界面方式；导航项 SHALL 不包含抽卡、排行榜、活动、任务、背包或邮件。窄屏导航 SHALL 不依赖桌面 Header 的固定横向空间。

#### Scenario: 打开未开放阵容入口
- **GIVEN** 任意身份访问 `/lineup`
- **WHEN** 未开放页面渲染
- **THEN** 页面说明调整阵容尚未开放
- **AND** 玩家可在一次明确操作内返回主界面

#### Scenario: 手机端导航
- **GIVEN** 视口宽度小于 640px
- **WHEN** App Shell 渲染
- **THEN** 导航使用适合窄屏的返回按钮、菜单或底部动作
- **AND** 不产生水平溢出
