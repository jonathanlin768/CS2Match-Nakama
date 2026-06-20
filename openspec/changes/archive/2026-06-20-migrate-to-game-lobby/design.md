## Context

当前 `client/src` 下的前端是一个响应式 Web 仪表盘：

- 入口 `main.tsx` 使用 `createBrowserRouter` 配置 `LoginPage`、`ProtectedRoute`、`AppLayout` 及多个子页面（`/home`、`/match`、`/gacha`、`/ranking`、`/profile/*`）。
- 页面依赖 `components/ui/` 下的 shadcn/ui 组件和 `components/cs2/` 下的业务组件。
- 全局样式使用 Tailwind CSS v4 的 `@import 'tailwindcss'` + `@theme inline`，并定义了 shadcn/ui 的完整语义变量（`--card`、`--primary`、`--border` 等）。
- Nakama 认证逻辑封装在 `context/AuthContext.tsx`、`hooks/useNakamaAuth.ts`、`api/auth.ts` 中，行为稳定。

新的游戏大厅设计来自 `D:\Project\nba-game\zip\`：固定 1600×900 画幅、暗色面板风格、手游式横版布局，源码基于 Tailwind CSS v3。需要把它迁移到当前技术栈并清理命名。

## Goals / Non-Goals

**Goals:**
- 用固定 1600×900 游戏大厅替换现有 `/home` 页面。
- 重绘 `LoginPage` 使其视觉与大厅一致，同时保留 Nakama 邮箱登录/注册/Session 恢复逻辑。
- 将旧页面/组件完整归档到 `src/legacy/`，不暴露为运行时路由，供后续重构参考。
- 运行时只保留一套 CSS 主题，基于 nba-game 变量系统并迁移到 Tailwind CSS v4。
- 源码中不出现 "nba" / "basketball" 等篮球相关命名或文案。

**Non-Goals:**
- 不修改 Nakama 后端、RPC、Match Handler、Storage 或 Luban 配置表。
- 本次不实现除登录/首页外的其他页面（对战、抽卡、排行榜、个人中心等后续逐步改造）。
- 不做响应式适配；1600×900 画幅在较小视口中直接居中裁剪。
- 不完全删除 shadcn/ui 依赖，但运行时不使用其主题变量。

## Decisions

### 1. 旧代码归档到 `src/legacy/`，不挂路由
- **Rationale**: 用户希望保留旧代码作为重构素材，但不需要运行时访问。归档后新代码不会意外 import 旧组件，避免两套 UI 系统互相污染。
- **Alternative considered**: 挂 `/before/*` 路由作为可运行对照。Rejected：会增加 bundle、造成样式冲突、且用户选择仅源码归档。

### 2. 新组件目录命名为 `components/lobby/`
- **Rationale**: "lobby" 是游戏大厅的中性命名，不暴露原项目主题，也不与后端/玩法概念冲突。
- **Alternative considered**: `components/nba/`（违反命名清理要求）、`components/home/`（与 `pages/Home` 重复）、`components/game/`（过于宽泛）。

### 3. CSS 主题完全替换为 nba-game 变量系统并迁移到 Tailwind 4
- **Rationale**: 用户要求运行时只用一套 CSS。移除 shadcn/ui 语义变量（`--card`、`--primary` 等），保留 Tailwind 4 的 `@import 'tailwindcss'` + `@theme inline` 机制，新增 `--panel`、`--panel-light`、`--gold`、`--font-display` 等变量。
- **Alternative considered**: 保留 shadyn 变量与 nba-game 变量共存。Rejected：违背"一套 CSS"要求。

### 4. `CenterStage` 替代原 `LineupStage`
- **Rationale**: 原组件名暗示"阵容"（篮球 lineup），改为 `CenterStage` 更中性，符合 CS2 主题。

### 5. 第一版保留占位人物图片但移除 NBA badge 与篮球文案
- **Rationale**: 用户允许第一版保留篮球运动员素材，但要求代码/文案中无 NBA 痕迹。NBA badge 需要替换为中性标识或移除。

### 6. 登录页保留 `useAuth()` 和表单状态逻辑，仅替换 UI
- **Rationale**: `useNakamaAuth` 和 `api/auth.ts` 已经过 OpenSpec 规约验证，直接复用可减少回归风险。

### 7. 路由入口 `/home` 指向新 `HomePage`，其他旧页面路由暂时移除
- **Rationale**: 旧页面已归档，新对战/抽卡/排行等页面后续再建。`main.tsx` 仅保留 `LoginPage`、受保护的 `HomePage`、以及 `*` 回退到 `/home`。

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| 移除 shadcn/ui 变量后，从 `src/legacy/` 复制出来的旧组件无法直接运行 | 重构时同步迁移样式；`src/legacy/` 明确为只读参考，不直接 import。 |
| nba-game 源码基于 React 18 / Tailwind 3 / RR6，在当前 React 19 / Tailwind 4 / RR7 下可能出现兼容问题 | 逐个组件验证类型与渲染；`lucide-react` 版本差异通过当前已安装版本 import 修正。 |
| 1600×900 直接裁剪在小屏幕上体验差 | 按用户要求不做自适应；后续若需要再评估缩放方案。 |
| 新 LoginPage 视觉重构可能引入表单状态 bug | 保留原有表单状态与 `useAuth` 调用逻辑，仅替换 JSX 与 className。 |
| 静态资源（`star-player.png`）较大 | 第一版保留；后续替换为 CS2 素材并优化体积。 |

## Migration Plan

1. 将旧页面/组件移动到 `src/legacy/`。
2. 替换 `src/index.css` 为新主题。
3. 创建 `src/components/lobby/*` 与 `src/pages/HomePage.tsx`。
4. 重写 `src/pages/LoginPage.tsx`。
5. 更新 `src/main.tsx` 路由。
6. 复制 nba-game 静态资源到 `client/public/`。
7. 运行 `npm run build` 与类型检查，清理剩余 "nba" 引用。

## Open Questions

- 新对战/抽卡/排行页面是否沿用同一 `lobby` 组件目录，还是后续按页面拆分？（建议后续按页面拆分）
- 登录页是否需要记住我功能？当前旧 LoginPage 的复选框为只读占位，本变更保持现状。
