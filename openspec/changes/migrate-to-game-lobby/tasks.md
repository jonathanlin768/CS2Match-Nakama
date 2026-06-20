## 1. 归档旧前端代码

- [x] 1.1 将 `client/src/pages/` 下的旧页面（除后续要重写的 `LoginPage` 外）移动到 `client/src/legacy/pages/`
- [x] 1.2 将 `client/src/components/cs2/`、`client/src/components/ui/`、`client/src/components/theme-provider.tsx` 移动到 `client/src/legacy/components/`
- [x] 1.3 将 `client/src/layout/AppLayout.tsx` 移动到 `client/src/legacy/layout/`
- [x] 1.4 将旧 hooks（`useSimStream.ts`、`use-mobile.ts`、`use-toast.ts`）与 `client/src/types/sim.ts` 移动到 `client/src/legacy/`
- [x] 1.5 确认 `client/src/context/AuthContext.tsx`、`client/src/hooks/useNakamaAuth.ts`、`client/src/api/`、`client/src/nakama.ts`、`client/src/lib/utils.ts`、`client/src/config/` 保留在原位作为共享代码

## 2. 统一 CSS 主题

- [x] 2.1 重写 `client/src/index.css`：保留 Tailwind CSS v4 入口，引入 Noto Sans SC / Oswald 字体
- [x] 2.2 在 `:root` 与 `@theme inline` 中定义新主题变量：`--background`、`--foreground`、`--panel`、`--panel-light`、`--accent`、`--accent-foreground`、`--gold`、`--muted`、`--font-display`
- [x] 2.3 移除 shadcn/ui 运行时变量依赖，添加 `.court-bg`、`.clip-slant`、`.font-display` 等工具类
- [x] 2.4 更新 `client/index.html`：`lang="zh-CN"`、viewport `user-scalable=no`、theme-color、apple-icon

## 3. 创建游戏大厅组件与首页

- [x] 3.1 创建 `client/src/components/lobby/data/lobby.ts`，替换篮球相关文案为中性/CS2 文案
- [x] 3.2 创建 `client/src/components/lobby/TopBar.tsx`
- [x] 3.3 创建 `client/src/components/lobby/PromoBar.tsx`
- [x] 3.4 创建 `client/src/components/lobby/LeftPanel.tsx`
- [x] 3.5 创建 `client/src/components/lobby/CenterStage.tsx`（原 LineupStage 重命名）
- [x] 3.6 创建 `client/src/components/lobby/RightPanel.tsx`
- [x] 3.7 创建 `client/src/components/lobby/BottomNav.tsx`
- [x] 3.8 创建 `client/src/pages/HomePage.tsx`，组合上述组件为 1600×900 固定画幅

## 4. 重写登录页

- [x] 4.1 将旧 `client/src/pages/LoginPage.tsx` 移动到 `client/src/legacy/pages/LoginPage.tsx` 作为参考
- [x] 4.2 新建 `client/src/pages/LoginPage.tsx`，保留 `useAuth()`、表单状态、`errorToMessage`、登录/注册校验逻辑
- [x] 4.3 用 1600×900 固定画幅暗色面板风格替换登录页 UI，确保与大厅视觉一致
- [x] 4.4 保留 Session 恢复加载状态、密码显示切换、错误提示行为

## 5. 更新路由与应用入口

- [x] 5.1 更新 `client/src/main.tsx`：保留 `LoginPage` 在 `/`，受保护路由下仅保留 `/home` 指向新 `HomePage`
- [x] 5.2 移除 `/match`、`/gacha`、`/ranking`、`/profile/*` 运行时路由（这些页面已归档到 `src/legacy/`）
- [x] 5.3 确认 `AuthProvider` 与 `ProtectedRoute` 仍然包裹受保护路由

## 6. 复制静态资源

- [x] 6.1 将 `D:\Project\nba-game\zip\public\images\star-player.png` 复制到 `client/public/images/`
- [x] 6.2 将 nba-game 的图标资源（`icon.svg`、`apple-icon.png`、`icon-dark-32x32.png`、`icon-light-32x32.png`）复制到 `client/public/`
- [x] 6.3 确认 `client/public/` 中旧 favicon 不会与新图标冲突，必要时更新 `index.html` 引用

## 7. 验证与清理

- [x] 7.1 在 `client/` 目录执行 `npm run build`，确保 TypeScript 类型检查与 Vite 构建通过
- [x] 7.2 使用 `npx tsc --noEmit` 验证无类型错误
- [x] 7.3 对 `client/src/` 执行大小写不敏感搜索，确认不存在 "nba"、"basketball" 字符串
- [x] 7.4 启动开发服务器，验证 `/` 登录页与 `/home` 大厅页正常渲染，1600×900 画幅居中裁剪
- [x] 7.5 验证登录/注册/Session 恢复流程仍可正常工作
