## MODIFIED Requirements

### Requirement: Vite + React + TypeScript 项目骨架

前端项目 SHALL 位于 `client/` 目录，使用 Vite 作为构建工具，React 18+ 作为 UI 框架，TypeScript 作为开发语言。项目 SHALL 引入 Tailwind CSS 4 作为样式框架，react-router-dom v7 管理路由，shadcn/ui 作为组件库。

#### Scenario: 项目初始化

- **WHEN** 开发者在 `client/` 目录执行 `npm install` 然后 `npm run dev`
- **THEN** Vite 开发服务器在 `http://localhost:5173` 启动
- **AND** 浏览器打开后显示 LoginPage 登录页面

#### Scenario: TypeScript 类型检查

- **WHEN** 开发者执行 `npx tsc --noEmit`
- **THEN** 项目通过类型检查，无类型错误

#### Scenario: 前端项目目录结构

- **WHEN** 开发者查看 `client/src/` 目录
- **THEN** 目录包含 `pages/`（5 个页面）、`components/cs2/`（CS2 业务组件）、`components/ui/`（shadcn/ui 组件）、`layout/`（AppLayout）、`hooks/`（自定义 hooks）、`lib/`（工具函数）、`types/`（类型定义）
- **AND** `src/` 下存在 `main.tsx`（入口，含路由配置）和 `index.css`（全局样式）
