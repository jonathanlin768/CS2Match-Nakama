# Client Routing

React Router v7 前端路由系统规格 — 页面导航、路由守卫、Header 联动。

## Purpose

定义前端 SPA 的路由结构、页面组织方式和导航行为。LoginPage 为独立入口（无 AppLayout），其他页面由 AppLayout（Header + Footer）包裹。

## Requirements

### Requirement: React Router 路由系统

前端项目 SHALL 使用 react-router-dom v7 的 `createBrowserRouter` + `RouterProvider` 模式管理页面路由。路由结构包含一个独立的登录页和四个由 AppLayout 包裹的内部页面。

#### Scenario: 用户访问根路径显示登录页

- **WHEN** 用户在浏览器中访问 `/`
- **THEN** 系统显示 LoginPage 登录页面
- **AND** 不显示 Header 和 Footer

#### Scenario: 登录后跳转到首页

- **WHEN** 用户在 LoginPage 点击登录按钮（无需真实认证）
- **THEN** 系统导航到 `/home`
- **AND** 页面显示 AppLayout（包含 Header、Footer 和 Home 页面内容）

#### Scenario: 导航栏路由跳转

- **WHEN** 用户点击 Header 导航栏中的"对战"链接
- **THEN** 系统导航到 `/match` 并显示 MatchPage
- **WHEN** 用户点击"抽卡"链接
- **THEN** 系统导航到 `/gacha` 并显示 GachaPage
- **WHEN** 用户点击"排行"链接
- **THEN** 系统导航到 `/ranking` 并显示 RankingPage

#### Scenario: SPA 路由回退（由 Nginx 处理）

- **WHEN** 用户直接访问 `/home` 等子路由
- **THEN** Nginx 返回 `index.html`
- **AND** React Router 解析 URL 并渲染对应页面

### Requirement: 路由与 Header 导航联动

Header 组件中的导航链接 SHALL 使用 `NavLink` 组件，当前活跃路由对应的导航项 SHALL 高亮显示。

#### Scenario: 导航高亮

- **WHEN** 用户当前在 `/match` 页面
- **THEN** Header 中"对战"链接显示为活跃状态（`text-foreground`）
- **AND** 其他导航链接显示为非活跃状态（`text-muted-foreground`）
