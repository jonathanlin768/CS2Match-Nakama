## Context

WebClientSimu 已有 5 个完整界面的 CS2 模拟器前端项目（React 18 + Vite + Tailwind CSS 4 + shadcn/ui + react-router-dom v7）。当前 Nakama 项目的 `client/` 只有 Vite 骨架。需要将界面迁移过来，仅做 UI 呈现，后续再逐步接入 Nakama 后端。

源项目使用 `@/` 路径别名指向项目根目录，导入路径如 `@/components/cs2/Header`。目标项目的路径别名需要对齐，否则所有导入都要改动。

## Goals / Non-Goals

**Goals:**
- 5 个页面能在浏览器中正常渲染，无控制台报错
- 路由导航正常工作（LoginPage → Home → 其他页面）
- 全局 dark gold 主题正常显示
- 所有 CS2 业务组件正常渲染，使用硬编码 mock 数据
- MatchPage 不依赖 WebSocket 连接，stub hook 返回未连接状态

**Non-Goals:**
- 不接入 Nakama 服务器认证
- 不实现真实的 WebSocket 模拟对战
- 不实现真实的抽卡/排行 API
- 不修改 Docker 构建配置（后续再处理）
- 不修改现有 Go 插件代码

## Decisions

### 1. 路由方案：react-router-dom v7

**选择**: `createBrowserRouter` + `RouterProvider`（和源项目一致）

**理由**: 源项目所有组件都使用 react-router API（`Link`、`NavLink`、`useNavigate`、`useLocation`、`Outlet`），换库成本高。v7 是当前最新稳定版。

**路由结构**:
```
/          → LoginPage（无 AppLayout 包裹，独立登录页）
/home      → Home（AppLayout 包裹）
/match     → MatchPage（AppLayout 包裹）
/gacha     → GachaPage（AppLayout 包裹）
/ranking   → RankingPage（AppLayout 包裹）
```

### 2. CSS 方案：Tailwind CSS 4 + CSS 变量主题

**选择**: 使用源项目的 `app/globals.css`（dark gold 主题 CSS 变量），替换当前 `App.css` 和 `index.css`

**理由**: 源项目所有组件使用 `bg-card`、`text-primary`、`border-border` 等 Tailwind 语义类名，依赖这些 CSS 变量。`@theme inline` 映射了 `--color-*` 变量到 Tailwind 类名。

### 3. shadcn/ui 组件

**选择**: 完整复制源项目 `components/ui/` 下被引用的 shadcn 组件

**源项目有 40+ 个 UI 组件，但实际被 CS2 业务组件引用的不多。需要先分析依赖关系，只复制必要的 UI 组件。** 初步判断需要：`button`、`card`、`input`、`table`、`avatar`、`badge`、`progress`、`separator`、`tabs`、`tooltip`、`dropdown-menu`、`scroll-area`、`select` 等。

### 4. 主题方案：next-themes

**选择**: 继续使用 `next-themes` 的 `ThemeProvider`，默认 dark 主题

**理由**: 源项目 `AppLayout` 包裹 `<ThemeProvider defaultTheme="dark">`，全局 CSS 变量在 `:root` 下定义（默认为 dark）。保持一致性。

### 5. MatchPage 的 useSimStream Stub

**选择**: 创建 stub 版本的 `useSimStream`，始终返回 `connected: false`、`simStarted: false`、空数组

**理由**: 目标是不接真实 WebSocket，但组件渲染流程依赖 hook 返回值。Stub 保证 `MatchPage` 渲染不崩溃，显示"未连接"状态。

### 6. 路径别名

**选择**: 在 `vite.config.ts` 和 `tsconfig.app.json` 中配置 `@` → `./src` 别名

**理由**: 源项目 `@` 指向项目根目录（因为组件在 `components/`、hooks 在 `hooks/`）。目标项目将内容放在 `src/` 下，`@` 指向 `src/` 更符合 Vite 惯例，且只需要批量替换一次导入路径前缀。

### 7. npm 包选型

| 包名 | 理由 |
|------|------|
| `tailwindcss` + `@tailwindcss/postcss` | 源项目使用 Tailwind CSS 4，postcss 插件模式 |
| `react-router-dom` ^7 | 源项目使用的路由库，保持 API 兼容 |
| `lucide-react` | 源项目图标库，被大量组件引用 |
| `next-themes` | 主题切换（dark mode），AppLayout 依赖 |
| `clsx` + `tailwind-merge` | `cn()` 工具函数依赖 |
| radix-ui 系列 | shadcn/ui 组件的底层依赖 |

## Risks / Trade-offs

- **[风险] 复制大量文件后 TypeScript 类型检查失败** → 逐步修复导入路径和类型错误，必要时放宽 `noUnusedLocals` 等 lint 规则
- **[风险] shadcn/ui 组件版本不兼容** → 保持和源项目相同的 radix-ui 版本范围
- **[风险] Vite 代理配置冲突** → 当前 stub 阶段不配置 `/ws` 代理，等接入真实 WebSocket 时再加
- **[权衡] 复制全部 UI 组件 vs 按需引入** → 选择先全部复制（源项目有完整集合），后续清理未使用的。避免了逐个排查依赖的耗时

## Open Questions

- 是否需要在 Docker 构建中验证前端编译？（后续 tasks 中决定）
- `client/src/index.css` 是否保留还是全部替换为 globals.css？（替换）
