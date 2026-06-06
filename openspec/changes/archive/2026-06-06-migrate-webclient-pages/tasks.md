## 1. 依赖安装与项目配置

- [x] 1.1 安装 Tailwind CSS 4 相关依赖：`tailwindcss`、`@tailwindcss/postcss`、`postcss`、`tw-animate-css`
- [x] 1.2 安装 shadcn/ui 相关依赖：`clsx`、`tailwind-merge`、`class-variance-authority`、`lucide-react`、`next-themes`、`sonner`
- [x] 1.3 安装 react-router-dom v7：`react-router-dom`
- [x] 1.4 安装 shadcn/ui 所需的 radix-ui 系列包（按需，从源项目 package.json 复制版本号）
- [x] 1.5 创建 `client/postcss.config.mjs`（`@tailwindcss/postcss` 插件）
- [x] 1.6 更新 `client/vite.config.ts`：添加 `@` → `./src` 路径别名
- [x] 1.7 更新 `client/tsconfig.app.json`：添加 `@/*` → `./src/*` 路径映射，调整 `types`、`lib` 等配置以匹配源项目
- [x] 1.8 替换 `client/src/index.css`（原 Vite 默认样式）为源项目的 `app/globals.css`（dark gold 主题 + Tailwind 配置）

## 2. 基础模块（lib、types、hooks）

- [x] 2.1 创建 `client/src/lib/utils.ts`（`cn` 函数，使用 `clsx` + `tailwind-merge`）
- [x] 2.2 创建 `client/src/types/sim.ts`（SimPlayerState、SimKillEvent、SimTickMsg 等类型定义）
- [x] 2.3 创建 `client/src/hooks/useSimStream.ts`，stub 版本，不创建 WebSocket 连接
- [x] 2.4 创建 `client/src/hooks/use-mobile.ts` 和 `client/src/hooks/use-toast.ts`

## 3. shadcn/ui 组件库

- [x] 3.1 复制源项目 `components/theme-provider.tsx` → `client/src/components/theme-provider.tsx`
- [x] 3.2 复制源项目 `components/ui/` 下所有 shadcn 组件到 `client/src/components/ui/`
- [x] 3.3 修复 UI 组件中的 `@/` 导入路径（在新项目中 `@` = `src/`，路径结构对齐，无需改动）

## 4. 布局组件（Layout）

- [x] 4.1 创建 `client/src/components/cs2/Header.tsx`（从源项目复制，硬编码用户数据）
- [x] 4.2 创建 `client/src/components/cs2/Footer.tsx`（从源项目复制）
- [x] 4.3 创建 `client/src/layout/AppLayout.tsx`（ThemeProvider + Header + Outlet + Footer，移除 `@/app/globals.css` 导入）

## 5. Home 页业务组件

- [x] 5.1 创建 `client/src/components/cs2/PlayerSidebar.tsx`
- [x] 5.2 创建 `client/src/components/cs2/SeasonBanner.tsx`
- [x] 5.3 创建 `client/src/components/cs2/DailyTasks.tsx`
- [x] 5.4 创建 `client/src/components/cs2/ChampionshipCenter.tsx`
- [x] 5.5 创建 `client/src/components/cs2/QuickStart.tsx`
- [x] 5.6 创建 `client/src/components/cs2/ActivityCenter.tsx`
- [x] 5.7 创建 `client/src/components/cs2/TeamRankings.tsx`
- [x] 5.8 创建 `client/src/components/cs2/RecentMatches.tsx`
- [x] 5.9 创建 `client/src/components/cs2/CardPack.tsx`

## 6. Gacha 页业务组件

- [x] 6.1 创建 `client/src/components/cs2/gacha/PackSidebar.tsx`
- [x] 6.2 创建 `client/src/components/cs2/gacha/CardCarousel.tsx`
- [x] 6.3 创建 `client/src/components/cs2/gacha/PityPanel.tsx`
- [x] 6.4 创建 `client/src/components/cs2/gacha/PoolPreview.tsx`

## 7. Match 页业务组件

- [x] 7.1 创建 `client/src/components/cs2/match/MatchScoreBar.tsx`
- [x] 7.2 创建 `client/src/components/cs2/match/TeamRoster.tsx`
- [x] 7.3 创建 `client/src/components/cs2/match/TacticalMap.tsx`
- [x] 7.4 创建 `client/src/components/cs2/match/RoundEvents.tsx`
- [x] 7.5 创建 `client/src/components/cs2/match/PlayerStats.tsx`
- [x] 7.6 创建 `client/src/components/cs2/match/RoundTactics.tsx`

## 8. Ranking 页业务组件

- [x] 8.1 创建 `client/src/components/cs2/ranking/RankingSidebar.tsx`
- [x] 8.2 创建 `client/src/components/cs2/ranking/RankingTable.tsx`
- [x] 8.3 创建 `client/src/components/cs2/ranking/RankingInfoPanel.tsx`

## 9. 页面组件创建 & 服务器逻辑剥离

每个页面从源项目复制后，必须做以下清理：

- [x] 9.1 **LoginPage** — 源项目无真实服务器逻辑
  - 保留 `handleLogin` 中的 `setTimeout(1000)` + `navigate("/home")`（模拟登录延迟，不调任何 API）
- [x] 9.2 **Home** — 源项目全是硬编码 mock 数据，无服务器逻辑
  - 直接复制，路径对齐后无需改动
- [x] 9.3 **MatchPage** — 源项目依赖 WebSocket 实时数据流，已全部剥离：
  - [x] 9.3.1 替换为 stub hook（`useSimStream` 返回空数据）
  - [x] 9.3.2 移除 `useEffect` auto-start 调用
  - [x] 9.3.3 保留 `useMemo` 派生逻辑（stub 空数组自然产出空结果）
  - [x] 9.3.4-9.3.9 tickMsg/c4State/players 全为空，组件显示空状态
- [x] 9.4 **GachaPage** — 源项目仅使用 `useState` 本地状态，无服务器逻辑
- [x] 9.5 **RankingPage** — 源项目仅使用 `useState` 本地状态，无服务器逻辑
- [x] 9.6 更新 `client/src/main.tsx`：配置 `createBrowserRouter` 路由表，删除原 `App.tsx`

## 10. 清理与验证

- [x] 10.1 删除不再需要的文件：`client/src/App.tsx`、`client/src/App.css`
- [x] 10.2 运行 `npm install` 确保所有依赖安装成功
- [x] 10.3 运行 `npx tsc --noEmit` 检查 TypeScript 类型错误并逐一修复
- [x] 10.4 运行 `npm run build` 验证 Vite 构建成功（tsc + vite build 通过）
- [X] 10.5 测试路由导航：`/` → 登录 → 点击登录 → `/home` → Header 导航到其他页面
