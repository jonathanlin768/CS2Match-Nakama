## Why

当前 CS2Match-Nakama 前端是一个响应式 Web 仪表盘风格（Header + 内容 + Footer），更接近管理后台而非游戏客户端。为了提升游戏沉浸感并与手游式的横版体验对齐，需要把首页改造成固定 1600×900 画幅的游戏大厅界面，同时保留 Nakama 认证体系与旧代码作为重构素材。

## What Changes

- **React 前端首页重构**：将现有 `Home` 页面替换为固定 1600×900 的游戏大厅布局（顶部信息栏、活动快捷栏、左侧面板、中央展示区、右侧面板、底部导航）。
- **登录页风格统一**：保留 `LoginPage` 的 Nakama 邮箱登录/注册/Session 恢复逻辑，但 UI 重构成与游戏大厅一致的手游横版风格。
- **旧前端源码归档**：将现有 `pages/`、`components/cs2/`、`components/ui/`、`layout/` 等页面与组件完整移动到 `src/legacy/`，**不暴露为运行时路由**，仅作为后续重构好友/聊天等功能的代码参考。
- **CSS 主题统一**：运行时只保留一套 CSS 主题，基于 nba-game 的暗色面板风格迁移到 Tailwind CSS v4 语法；移除对 shadcn/ui 主题变量的运行时依赖。
- **命名清理**：源码中不出现任何 "nba"、"basketball" 等篮球相关命名或文案，第一版允许保留占位人物图片。
- **路由与行为保留**：登录成功后仍跳转 `/home`；1600×900 画幅在较小视口中直接居中裁剪，不做自适应。

## Capabilities

### New Capabilities
- `game-lobby-ui`: 固定 1600×900 游戏大厅首页，包含 TopBar、PromoBar、LeftPanel、CenterStage、RightPanel、BottomNav 组件。
- `login-page-game-style`: 与游戏大厅视觉一致的手游横版登录/注册页，保留 Nakama 认证行为。

### Modified Capabilities
- `client-pages`: Home 页面的 UI 结构与布局要求发生变化（从响应式仪表盘改为固定画幅大厅），登录/注册的功能行为不变，仅实现层样式变更。

## Impact

- **React 前端**：`client/src/index.css`、`client/src/main.tsx`、`client/src/pages/LoginPage.tsx`、`client/src/pages/Home.tsx` 等文件会新增或重写；旧页面/组件归档到 `client/src/legacy/`。
- **Nakama 后端 / 数据库 / 部署**：无影响。认证 API、Session 管理、Dockerfile、nginx 配置保持不变。
- **RPC / Match Handler / Storage**：无需新增。
- **Luban 配置表**：无需更新。
- **构建产物**：移除旧 shadcn/ui 运行时依赖后 bundle 体积可能减小；`src/legacy/` 中的归档代码不参与运行时构建（未被 import）。
