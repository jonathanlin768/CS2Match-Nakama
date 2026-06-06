## Why

WebClientSimu 项目已有 5 个完整的 CS2 模拟器前端界面（登录、首页、对战、抽卡、排行），当前 Nakama 项目的 `client/` 只有 Vite 骨架。将这些界面迁移到 client 目录下，可以快速建立可演示的前端原型，为后续接入 Nakama 后端铺路。本次仅迁移 UI 样式效果，不接服务器逻辑。

## What Changes

- 安装 Tailwind CSS 4 + shadcn/ui 主题系统（覆盖原 client 的 App.css）
- 安装 react-router-dom v7，建立 `/`（LoginPage）→ `/home`、`/match`、`/gacha`、`/ranking` 的路由结构
- 迁移 5 个页面组件：LoginPage、Home、MatchPage、GachaPage、RankingPage
- 迁移 ~20 个 CS2 业务组件（Header、Footer、PlayerSidebar、SeasonBanner 等）
- 迁移 shadcn/ui 基础组件库（button、card、input、table 等）
- 迁移 AppLayout 布局（Header + Outlet + Footer + ThemeProvider）
- 迁移全局 CSS 变量（dark gold 主题）和 Tailwind 配置
- MatchPage 的 `useSimStream` hook 用 mock 数据 stub 掉，返回空数据/未连接状态
- LoginPage 点击登录后 navigate 到 `/home`（不做真实认证）

## Capabilities

### New Capabilities

- `client-routing`: react-router-dom v7 路由系统，LoginPage 为首页 `/`，AppLayout 包裹的受保护页面为 `/home`、`/match`、`/gacha`、`/ranking`
- `client-pages`: 5 个页面及其依赖的 CS2 业务组件、shadcn/ui 组件库、全局主题样式

### Modified Capabilities

- `react-frontend-scaffold`: 在原有 Vite + React + Nakama SDK 骨架基础上，新增 Tailwind CSS 4 + 路由 + 页面体系的前端依赖和目录结构

## Impact

- **React 前端**: `client/` 目录大幅扩展 — 新增 `src/pages/`、`src/components/`、`src/layout/`、`src/hooks/`、`src/lib/`、`src/types/`
- **依赖**: 新增 tailwindcss、postcss、react-router-dom、lucide-react、next-themes、clsx、tailwind-merge 及 shadcn/ui 所需的 radix-ui 系列包
- **Nakama 后端**: 无影响
- **数据库**: 无影响
- **Luban 配置**: 无影响
