# CS2Match 前端架构 & 代码说明

## 1. 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| React | 19.2 | UI 框架 |
| TypeScript | 6.0 | 类型安全 |
| Vite | 8.0 | 开发服务器 + 构建打包 |
| React Router | 7.14 | SPA 客户端路由 |
| Tailwind CSS | 4.2 | 原子化 CSS |
| shadcn/ui | — | Headless UI 组件库（基于 Radix UI） |
| @heroiclabs/nakama-js | 2.8 | Nakama 客户端 SDK |
| next-themes | 0.4 | 深色/浅色主题切换 |
| lucide-react | 0.564 | 图标库 |
| sonner | 1.7 | Toast 通知 |

---

## 2. 整体定位

前端在整个系统中的位置：

```
浏览器 (用户)
    │
    │  HTTP/WebSocket :7350
    ▼
Nakama 3.30 (Docker)
    │
    │  静态资源 :3000→:80
    ▼
React SPA (Nginx Docker)
    └── 通过 @heroiclabs/nakama-js SDK 与 Nakama 通信
```

前端是**纯静态 SPA**，无服务端渲染。所有业务逻辑通过 Nakama SDK 调用后端 API。

---

## 3. 源码结构

```
client/
├── index.html                  ← SPA 入口 HTML
├── vite.config.ts              ← Vite 构建配置（@/ 路径别名）
├── tsconfig.json               ← TypeScript 总配置
├── tsconfig.app.json           ← 应用代码 TS 配置
├── tsconfig.node.json          ← Vite/Node 端 TS 配置
├── package.json                ← 依赖和脚本
├── postcss.config.mjs          ← PostCSS 配置（Tailwind）
├── eslint.config.js            ← ESLint 规则
├── Dockerfile                  ← 多阶段 Docker 构建
├── nginx.conf                  ← Nginx 静态托管 + SPA 回退
├── .env.example                ← 环境变量模板
│
├── public/
│   ├── favicon.svg
│   ├── icons.svg
│   └── data/config/
│       └── tbitem.json         ← Luban 配置表数据（与 server 共享）
│
└── src/
    ├── main.tsx                ← 应用入口：创建 Router + AuthProvider
    ├── index.css               ← Tailwind 入口 + CSS 变量（主题色）
    ├── nakama.ts               ← Nakama 客户端单例
    │
    ├── api/                    ← API 调用层
    │   └── auth.ts             ← 邮箱登录/注册/Session 恢复
    │
    ├── hooks/                  ← 自定义 Hooks
    │   ├── useNakamaAuth.ts    ← 认证状态管理（核心 Hook）
    │   ├── useSimStream.ts     ← 模拟对战 WebSocket（当前为 stub）
    │   ├── use-mobile.ts       ← 响应式断点检测
    │   └── use-toast.ts        ← Toast 通知 Hook
    │
    ├── context/                ← React Context
    │   └── AuthContext.tsx      ← 认证上下文（Provider + useAuth）
    │
    ├── components/             ← UI 组件
    │   ├── ProtectedRoute.tsx   ← 路由守卫（未登录重定向）
    │   ├── theme-provider.tsx   ← 主题 Provider（封装 next-themes）
    │   │
    │   ├── ui/                  ← shadcn/ui 基础组件（~50 个）
    │   │   ├── button.tsx
    │   │   ├── card.tsx
    │   │   ├── dialog.tsx
    │   │   ├── dropdown-menu.tsx
    │   │   ├── sidebar.tsx
    │   │   ├── table.tsx
    │   │   ├── tabs.tsx
    │   │   └── ...              ← 共 40+ 个 UI 原子组件
    │   │
    │   └── cs2/                 ← 业务组件
    │       ├── Header.tsx       ← 全局顶栏（Logo + 导航 + 货币显示）
    │       ├── Footer.tsx       ← 全局底栏（链接导航）
    │       ├── PlayerSidebar.tsx
    │       ├── SeasonBanner.tsx
    │       ├── DailyTasks.tsx
    │       ├── ChampionshipCenter.tsx
    │       ├── QuickStart.tsx
    │       ├── ActivityCenter.tsx
    │       ├── TeamRankings.tsx
    │       ├── RecentMatches.tsx
    │       ├── CardPack.tsx
    │       │
    │       ├── match/           ← 对战页面子组件
    │       │   ├── MatchScoreBar.tsx
    │       │   ├── TeamRoster.tsx
    │       │   ├── TacticalMap.tsx
    │       │   ├── RoundEvents.tsx
    │       │   ├── PlayerStats.tsx
    │       │   └── RoundTactics.tsx
    │       │
    │       ├── gacha/           ← 抽卡页面子组件
    │       │   ├── PackSidebar.tsx
    │       │   ├── CardCarousel.tsx
    │       │   ├── PityPanel.tsx
    │       │   └── PoolPreview.tsx
    │       │
    │       └── ranking/         ← 排行页面子组件
    │           ├── RankingSidebar.tsx
    │           ├── RankingTable.tsx
    │           └── RankingInfoPanel.tsx
    │
    ├── pages/                   ← 路由页面
    │   ├── LoginPage.tsx        ← 登录/注册（含表单验证）
    │   ├── Home.tsx             ← 首页仪表盘
    │   ├── MatchPage.tsx        ← 对战模拟回放
    │   ├── GachaPage.tsx        ← 抽卡系统
    │   └── RankingPage.tsx      ← 排行榜
    │
    ├── layout/
    │   └── AppLayout.tsx        ← 受保护页面布局（Header + Outlet + Footer）
    │
    ├── config/                  ← Luban 配置表加载
    │   ├── index.ts             ← loadConfig() 异步加载 JSON
    │   └── schema.ts           ← Luban 自动生成的 Table 类
    │
    ├── types/
    │   └── sim.ts               ← 模拟对战 TypeScript 类型定义
    │
    └── lib/
        └── utils.ts             ← cn() 工具函数（Tailwind 类名合并）
```

---

## 4. 入口与路由

### 4.1 启动流程

```
index.html
    └── <script type="module" src="/src/main.tsx">
            │
            ├── createBrowserRouter([...])   ← React Router 路由树
            ├── createRoot(#root)            ← React 18 渲染入口
            └── <AuthProvider>               ← 包裹整个应用
                └── <RouterProvider>
```

### 4.2 路由表

```
/                         → LoginPage          (公开 — 所有人可访问)
 ├── <ProtectedRoute>     → 路由守卫            (检查认证状态)
 │   └── <AppLayout>      → Header + Footer 壳
 │       ├── /home        → Home               (首页仪表盘)
 │       ├── /match       → MatchPage          (对战回放)
 │       ├── /gacha       → GachaPage          (抽卡)
 │       └── /ranking     → RankingPage        (排行榜)
```

### 4.3 路由守卫（ProtectedRoute）

```
用户访问受保护页面
    │
    ├── status === "restoring"
    │   └── 显示 "正在恢复登录..." 加载动画
    │
    ├── status === "guest"
    │   └── <Navigate to="/" replace />
    │
    └── status === "authenticated"
        └── <Outlet />  ← 渲染子路由
```

---

## 5. 认证系统（三层设计）

认证系统是前端最核心的基础设施，采用三层解耦设计：

```
┌──────────────────────────────────────────────────────┐
│  第 3 层：React Context                               │
│  context/AuthContext.tsx                              │
│  ┌──────────────────────────────────────────────┐    │
│  │ AuthProvider — 在 <RouterProvider> 外层包裹    │    │
│  │ useAuth()    — 任何子组件读取认证状态          │    │
│  │                                                  │    │
│  │ 暴露: status, session, error, login(),          │    │
│  │       register(), logout()                      │    │
│  └──────────────┬───────────────────────────────┘    │
│                 │ 调用                                │
│  ┌──────────────▼───────────────────────────────┐    │
│  │ 第 2 层：状态管理 Hook                          │    │
│  │ hooks/useNakamaAuth.ts                         │    │
│  │                                                  │    │
│  │ 三态状态机:                                      │    │
│  │   restoring → authenticated | guest              │    │
│  │                                                  │    │
│  │ 挂载时自动 restoreSession()                      │    │
│  │ login/register 调用 API 层并更新状态             │    │
│  └──────────────┬───────────────────────────────┘    │
│                 │ 调用                                │
│  ┌──────────────▼───────────────────────────────┐    │
│  │ 第 1 层：API 函数                              │    │
│  │ api/auth.ts                                     │    │
│  │                                                  │    │
│  │ loginWithEmail()    → authenticateEmail(false)  │    │
│  │ registerWithEmail() → authenticateEmail(true)   │    │
│  │ restoreSession()    → Session.restore() + 刷新   │    │
│  │ clearSession()      → 清除 localStorage          │    │
│  └──────────────┬───────────────────────────────┘    │
│                 │ 调用                                │
│  ┌──────────────▼───────────────────────────────┐    │
│  │ Nakama SDK                                    │    │
│  │ nakama.ts — Client 单例                       │    │
│  │                                                  │    │
│  │ new Client(serverKey, host, port, useSSL)       │    │
│  │ 配置来源: import.meta.env.VITE_NAKAMA_*         │    │
│  └──────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────┘
```

### 5.1 Nakama Client 单例 (`nakama.ts`)

```typescript
const client = new Client(
  NAKAMA_SERVER_KEY,  // default: "defaultkey"
  NAKAMA_HOST,        // default: "localhost"
  NAKAMA_PORT,        // default: "7350"
  NAKAMA_USE_SSL      // default: false
);
```

- **宿主机开发**：环境变量 `VITE_NAKAMA_HOST=localhost`，前端直连宿主机端口映射
- **Docker 部署**：环境变量 `VITE_NAKAMA_HOST=nakama`，前端通过 Docker 网络通信

### 5.2 Session 持久化与恢复

```
登录/注册成功
    └── session.token + session.refresh_token
        └── 存入 localStorage
            │
            ▼
        下次打开页面
            └── restoreSession()
                ├── 从 localStorage 读取 token + refresh_token
                ├── Session.restore() — 纯本地 JWT 解析
                ├── token 未过期 → 直接返回（零网络请求）
                ├── token 已过期 → sessionRefresh() 换新 token
                │   └── 成功 → 更新 localStorage
                └── 失败 → clearSession() → 返回 guest 状态
```

### 5.3 邮箱注册流程

```
registerWithEmail(email, password)
    └── client.authenticateEmail(email, password, true)
        ├── created=true  → 新账号，注册成功
        └── created=false → 该邮箱已存在
            └── Hook 层返回友好提示："该邮箱已注册，请直接登录"
```

---

## 6. 页面详解

### 6.1 LoginPage（登录/注册）

- **路径**：`/`
- **组件**：`pages/LoginPage.tsx`
- **状态**：`"use client"`；自管理表单状态
- **功能**：
  - Tab 切换：登录 / 注册
  - 登录表单：邮箱 + 密码 → `login(email, password)`
  - 注册表单：邮箱 + 密码 + 确认密码 → `register(email, password)`
  - 客户端校验：密码一致性、最小长度 8 位
  - 服务端错误中文化（`errorToMessage()`）
  - 认证成功后自动跳转到 `/home`
  - 社交登录按钮（Google / GitHub — UI 已就绪，后端待接入）
- **UI 特性**：渐变背景、密码显隐切换、加载状态

### 6.2 Home（首页仪表盘）

- **路径**：`/home`
- **组件**：`pages/Home.tsx` + 9 个业务子组件
- **布局**：响应式三行网格

```
┌─────────────────────────────────────────────────────┐
│ PlayerSidebar │ SeasonBanner │ DailyTasks │ Champ.  │
├─────────────────────────────────────────────────────┤
│ QuickStart (4 个游戏模式卡片)          │ Activity    │
├─────────────────────────────────────────────────────┤
│ TeamRankings (宽)  │ RecentMatches │ CardPack      │
└─────────────────────────────────────────────────────┘
```

- **响应式**：移动端堆叠、平板双列、桌面完整布局
- **当前状态**：纯静态 UI，数据均为硬编码占位

### 6.3 MatchPage（对战回放）

- **路径**：`/match`
- **组件**：`pages/MatchPage.tsx` + 6 个子组件
- **数据源**：`useSimStream()` — **当前为 stub，返回空数据**
- **布局**：经典电竞观战界面

```
┌───────────────────────────────────────────────┐
│ 连接状态栏 (connected: green/red)              │
├───────────────────────────────────────────────┤
│ MatchScoreBar (队伍比分 + 回合计时器)          │
├─────────┬──────────────────────┬──────────────┤
│ T Roster│    TacticalMap       │ CT Roster    │
│ (左)    │    (中，地图+点位)    │ (右)         │
├─────────┴──────────────────────┴──────────────┤
│ RoundEvents │ PlayerStats │ RoundTactics      │
└───────────────────────────────────────────────┘
```

- **数据转换**：`SimPlayerState` → `TeamRoster` 格式、`SimKillEvent` → `RoundEvents` 格式、像素坐标 → 百分比坐标
- **地图**：Dust2（MAP_W=1024, MAP_H=984）

### 6.4 GachaPage（抽卡）

- **路径**：`/gacha`
- **组件**：`pages/GachaPage.tsx` + 4 个子组件
- **状态**：`selectedPack`（当前选中卡包类型）
- **布局**：三栏 + 底部

```
┌───────────────────────────────────────────────┐
│ PackSidebar │ CardCarousel │ PityPanel        │
├───────────────────────────────────────────────┤
│ PoolPreview（当前卡池内容一览）                │
└───────────────────────────────────────────────┘
```

- **当前状态**：纯静态 UI，无后端交互

### 6.5 RankingPage（排行榜）

- **路径**：`/ranking`
- **组件**：`pages/RankingPage.tsx` + 3 个子组件
- **状态**：`selectedCategory`（当前排行类别）
- **布局**：三栏

```
┌───────────────────────────────────────────────┐
│ RankingSidebar │ RankingTable │ RankingInfo   │
└───────────────────────────────────────────────┘
```

- **当前状态**：纯静态 UI

---

## 7. 组件体系

### 7.1 UI 基础组件 (`components/ui/`)

基于 shadcn/ui（Radix UI 原语 + Tailwind 样式），共 40+ 个：

| 分类 | 组件 |
|------|------|
| 表单 | `button`, `input`, `textarea`, `checkbox`, `radio-group`, `select`, `switch`, `slider`, `toggle`, `label`, `field` |
| 布局 | `card`, `separator`, `sidebar`, `scroll-area`, `aspect-ratio`, `collapsible`, `accordion` |
| 浮层 | `dialog`, `alert-dialog`, `popover`, `hover-card`, `tooltip`, `dropdown-menu`, `context-menu`, `menubar`, `sheet` |
| 导航 | `breadcrumb`, `navigation-menu`, `pagination`, `tabs` |
| 数据展示 | `table`, `badge`, `avatar`, `progress`, `skeleton`, `kbd`, `item` |
| 反馈 | `alert`, `toast`, `toaster`, `sonner`, `spinner`, `empty` |
| 移动端 | `use-mobile`（响应式检测 Hook） |

### 7.2 路径别名

```typescript
// vite.config.ts
resolve: {
  alias: {
    '@': path.resolve(__dirname, './src'),
  },
}
```

所有 import 使用 `@/` 前缀：`import { useAuth } from "@/context/AuthContext"`

---

## 8. 配置表加载（Luban 共享）

前端与 server 共享同一套 Luban 配置表数据：

```
server/config/data/tbitem.json   ← Luban 导出（Go embed 编译进 .so）
client/public/data/config/tbitem.json  ← 同份数据（Vite 作为静态资源托管）
client/src/config/schema.ts      ← Luban 自动生成的 TypeScript 类
```

### 加载方式

```typescript
// config/index.ts
export async function loadConfig(): Promise<InstanceType<typeof Tables>> {
  // 并行 fetch 所有 JSON 配置表
  await Promise.all(TABLE_NAMES.map(async (name) => {
    const resp = await fetch(`/data/config/${name}.json`);
    dataCache[name] = await resp.json();
  }));
  return new Tables(loader);
}
```

与 server 的区别：
- **server**：`//go:embed` 编译时嵌入 → 零网络开销
- **client**：运行时 `fetch()` 加载 → 从 Nginx 静态资源获取

---

## 9. 样式系统

### 9.1 Tailwind CSS 4 + CSS 变量

核心在 `index.css`，使用 Tailwind 4 的 `@theme inline` 语法定义语义化颜色：

```
--primary: #c9a227 (金色)     ← CS2 品牌色
--background: #0a0e14 (深黑蓝)
--card: #141a24
--border: #2a3444
--ring: #c9a227               ← 聚焦环
```

### 9.2 Dark Mode

使用 `next-themes` 实现主题切换，`ThemeProvider` 配置为强制深色模式：

```tsx
<ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme" attribute="class">
```

`@custom-variant dark (&:is(.dark *))` 允许按需添加浅色主题变量。

### 9.3 类名合并

`lib/utils.ts` 提供 `cn()` 函数，合并 `clsx` + `tailwind-merge`：

```typescript
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

---

## 10. 构建与部署

### 10.1 开发模式

```bash
npm run dev       # Vite dev server，HMR 热更新
npm run build     # tsc -b 类型检查 + vite build
npm run preview   # 预览构建产物
```

### 10.2 Docker 多阶段构建

```dockerfile
# Stage 1: 编译 React 应用
FROM node:22-alpine AS builder
ARG VITE_NAKAMA_HOST=localhost    # Docker 网络内用 nakama
ARG VITE_NAKAMA_PORT=7350
ARG VITE_NAKAMA_SERVER_KEY=defaultkey
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build                  # 输出到 /app/dist

# Stage 2: Nginx 托管
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### 10.3 Nginx 配置要点

- **SPA 路由回退**：`try_files $uri $uri/ /index.html` — 所有非文件路径返回 index.html，由 React Router 处理
- **静态资源缓存**：JS/CSS/图片 → `expires 1y; Cache-Control public, immutable`
- **Nakama API 代理**（已注释）：可选配置 `proxy_pass http://nakama:7350`，当前方案是前端 SDK 直连

### 10.4 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_NAKAMA_HOST` | `localhost` | Docker 构建时设为 `nakama` |
| `VITE_NAKAMA_PORT` | `7350` | Nakama HTTP/WS 端口 |
| `VITE_NAKAMA_SERVER_KEY` | `defaultkey` | 与 server key 对应 |
| `VITE_NAKAMA_USE_SSL` | `false` | 生产环境应设为 `true` |

---

## 11. 已实现功能一览

| 模块 | 已完成 | 待实现 |
|------|--------|--------|
| 路由 | 5 个页面路由 + 路由守卫 | — |
| 认证 | 邮箱登录/注册/Session 自动恢复/登出 | OAuth 接入、忘记密码 |
| 首页 | 9 个业务组件 UI（PlayerSidebar、SeasonBanner 等） | 真实数据接入 |
| 对战 | 完整观战 UI（地图、Roster、击杀事件、C4 状态） | WebSocket 实时数据流 |
| 抽卡 | 4 个组件 UI（卡包选择、卡牌轮播、保底、卡池预览） | 后端抽卡逻辑 |
| 排行 | 3 个组件 UI（侧栏、表格、详情） | 后端排行数据 |
| UI 体系 | 40+ shadcn/ui 组件 + 暗色主题 | — |
| 配置表 | Luban schema + fetch 加载 | — |
| 部署 | Docker 多阶段构建 + Nginx | — |

---

## 12. 相关文件索引

| 文件 | 作用 |
|------|------|
| `src/main.tsx` | 应用入口，创建 Router + AuthProvider |
| `src/nakama.ts` | Nakama Client 单例 |
| `src/api/auth.ts` | 认证 API（登录/注册/Session 恢复） |
| `src/hooks/useNakamaAuth.ts` | 认证状态机（核心 Hook） |
| `src/context/AuthContext.tsx` | 认证上下文（Provider + useAuth） |
| `src/components/ProtectedRoute.tsx` | 路由守卫 |
| `src/layout/AppLayout.tsx` | 受保护页面布局壳 |
| `src/components/cs2/Header.tsx` | 全局导航栏 |
| `src/components/cs2/Footer.tsx` | 全局页脚 |
| `src/pages/LoginPage.tsx` | 登录/注册页面 |
| `src/pages/Home.tsx` | 首页仪表盘 |
| `src/pages/MatchPage.tsx` | 对战回放页面 |
| `src/pages/GachaPage.tsx` | 抽卡页面 |
| `src/pages/RankingPage.tsx` | 排行榜页面 |
| `src/hooks/useSimStream.ts` | 模拟对战数据流（stub） |
| `src/types/sim.ts` | 模拟对战类型定义 |
| `src/config/index.ts` | Luban 配置表加载器 |
| `src/config/schema.ts` | Luban 自动生成的 Table 类 |
| `src/lib/utils.ts` | cn() 工具函数 |
| `src/index.css` | Tailwind + CSS 变量主题 |
| `vite.config.ts` | Vite 构建配置（@ 别名） |
| `Dockerfile` | 多阶段 Docker 构建 |
| `nginx.conf` | Nginx SPA 托管配置 |
| `.env.example` | 环境变量模板 |

---

## 13. Review 发现的问题

以下是 2026-06-07 代码 review 中发现的问题，按优先级排列。

### 13.1 🟡 WebSocket 模拟数据流未接入（useSimStream）

**文件**：`src/hooks/useSimStream.ts`

`useSimStream` 当前是一个 stub —— `connected` 永远为 `false`，`startSim()` 和 `reset()` 都是 no-op。`MatchPage` 因此无法获得任何实时对战数据，所有派生状态（`mapPlayers`、`allEvents`、`roundTime` 等）都落在空数组 / 默认值。

```typescript
// 当前：stub 返回空状态
export function useSimStream() {
  const [state] = useState<SimStreamState>({
    connected: false,
    simStarted: false,
    tickMsg: null,
    killEvents: [],
    c4Events: [],
    roundEnd: null,
    players: [],
  });

  const startSim = useCallback((_tactic: string, _setup: string, _seed: number) => {
    // stub: do nothing
  }, []);
  ...
}
```

**影响**：MatchPage 观战 UI 骨架已完成，但无法看到实际对战数据。

**建议**：
1. 在 `useSimStream` 内建立到 Nakama 的 WebSocket 连接（`client.createSocket()`）
2. 注册 Match Handler 对应的消息 channel，解析 `SimTickMsg` / `SimKillEvent` / `SimEndMsg`
3. 在 Page 层或 Hook 内处理 `SimStartMsg` 的 `tactic` + `setup` + `seed` 参数回传

---

### 13.2 🟡 首页/抽卡/排行页面数据全为硬编码

**文件**：`src/pages/Home.tsx`、`src/pages/GachaPage.tsx`、`src/pages/RankingPage.tsx`

所有页面展示的数据（金币、排名、卡包列表、任务列表）均写死在组件 JSX 或本地 state 中，未从 Nakama 拉取。这是 UI 先行策略的必然阶段，但后续需要逐页接入。

**影响**：
- 页面看起来完整，但数据不会随用户/服务端状态变化
- Header 中显示 `23,568` 金币、`1,250` 钻石 — 是无意义的固定值

**建议**：
1. 优先接入认证后的用户 Profile 数据（金币、等级、头像）
2. 创建统一的 API 层（`api/profile.ts`、`api/ranking.ts` 等）
3. 在 Home 页添加 loading / error 状态处理

---

### 13.3 🟡 认证错误信息未结构化

**文件**：`src/hooks/useNakamaAuth.ts`、`src/pages/LoginPage.tsx`

当前 `useNakamaAuth` 只返回一个字符串 `error`，`LoginPage` 通过 `errorToMessage()` 做关键字匹配翻译：

```typescript
// LoginPage.tsx
function errorToMessage(error: string): string {
  if (error.includes("Network") || error.includes("fetch") || error.includes("connect")) {
    return "无法连接服务器，请检查网络";
  }
  if (error.includes("已注册")) {
    return error;
  }
  if (error.includes("credentials") || error.includes("password") || error.includes("email") || error.includes("invalid")) {
    return "邮箱或密码错误";
  }
  return "登录失败，请稍后重试";
}
```

**问题**：
1. **脆弱**：依赖英文字符串匹配（Nakama 返回的错误信息可能随版本变化）
2. **不区分来源**：登录错误和注册错误共用同一个 `error` 字段和翻译函数
3. **字面匹配已注册**：`"该邮箱已注册，请直接登录"` 是 Hook 层写死的中文，`errorToMessage` 靠 `includes("已注册")` 识别——如果未来换英文环境会失效

**建议**：
在 Hook 层返回结构化错误对象，替代裸字符串：

```typescript
interface AuthError {
  code: "NETWORK" | "INVALID_CREDENTIALS" | "ALREADY_REGISTERED" | "SERVER_ERROR";
  message: string;  // 用户可读的中文提示
}
```

---

### 13.4 🟢 "记住我" 和 "忘记密码" 是 UI 占位

**文件**：`src/pages/LoginPage.tsx`

```tsx
<label className="flex items-center gap-2 cursor-pointer">
  <input
    type="checkbox"
    checked={false}
    readOnly
    className="w-4 h-4 rounded ..."
  />
  <span className="text-sm text-muted-foreground">记住我</span>
</label>
<Link to="#" className="text-sm text-primary hover:underline">
  忘记密码？
</Link>
```

- checkbox `checked={false}` + `readOnly` → 永远不会被选中
- "忘记密码" 链接指向 `#`，点击无效

**建议**：两者都是合法的低优先级功能，可在认证系统稳定后再实现。暂时保留 UI 占位以维持设计完整性。

---

### 13.5 🟢 社交登录按钮无响应

**文件**：`src/pages/LoginPage.tsx`

Google 和 GitHub 登录按钮是纯静态 JSX，未绑定 `onClick` 处理：

```tsx
<button className="flex items-center justify-center gap-2 py-2.5 px-4 border ...">
  <svg ... />
  <span className="text-sm">Google</span>
</button>
```

**建议**：Nakama 原生支持 Google OAuth（`authenticateGoogle`）。接入步骤：
1. 在 Google Cloud Console 创建 OAuth 2.0 凭证
2. 在 `nakama-config.yml` 中配置 `social.google` 段
3. 前端调用 `client.authenticateGoogle(token)` 并传入 Google 返回的 access token

---

### 13.6 🟢 Nginx Nakama API 代理未启用

**文件**：`nginx.conf`

Nginx 配置中有一段已注释的反代配置：

```nginx
# location /nakama-api/ {
#     proxy_pass http://nakama:7350/;
#     proxy_http_version 1.1;
#     proxy_set_header Upgrade $http_upgrade;
#     proxy_set_header Connection "upgrade";
#     ...
# }
```

**当前方案**：前端 SDK 直连 Nakama 7350 端口（`VITE_NAKAMA_HOST=nakama`），浏览器必须能路由到 Nakama 容器。

**生产环境建议**：
- 启用 Nginx 反代：前端 SDK 的 `host` 设为空字符串（同源），所有 `/nakama-api/` 请求经 Nginx 转发
- 好处：不暴露 Nakama 端口、统一 HTTPS、简化 CORS
- 注意：WebSocket 的 `Upgrade` / `Connection` header 透传已配置好，无需额外处理

---

### 13.7 问题汇总

| 优先级 | 问题 | 影响范围 | 阻塞上线？ |
|--------|------|----------|-----------|
| 🟡 | useSimStream 未接入 WebSocket | MatchPage 无实时数据 | 是（对战是核心功能） |
| 🟡 | 页面数据硬编码 | Home / Gacha / Ranking | 是（用户看到假数据） |
| 🟡 | 认证错误未结构化 | LoginPage 错误翻译脆弱 | 否（当前能工作，但需重构） |
| 🟢 | "记住我"/"忘记密码" 占位 | LoginPage | 否 |
| 🟢 | 社交登录无响应 | LoginPage | 否 |
| 🟢 | Nginx 反代未启用 | 部署架构 | 否（开发阶段直连可接受） |
