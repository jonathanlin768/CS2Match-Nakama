## 1. API 层 — 创建 `api/auth.ts`

- [x] 1.1 新建 `client/src/api/auth.ts`，实现 `loginWithEmail(email, password): Promise<Session>` 函数，内部调用 `client.authenticateEmail(email, password, false)`（仅登录，不自动创建），成功后持久化 token/refresh_token 到 localStorage
- [x] 1.2 实现 `restoreSession(): Promise<Session | null>` 函数，从 localStorage 读取 token 和 refresh_token → token 未过期直接返回 Session → 过期则调用 `client.sessionRefresh()` 刷新 → 失败返回 `null` 并清除过期数据
- [x] 1.3 实现 `clearSession()` 函数，清除 localStorage 中的 `nakama_token` 和 `nakama_refresh`
- [x] 1.4 实现 `registerWithEmail(email, password): Promise<{ session: Session; created: boolean }>` 函数，内部调用 `client.authenticateEmail(email, password, true)`（账号不存在时自动创建），返回 Session 及 `created` 标志（`true` = 新账号，`false` = 该邮箱已注册），成功后持久化 token/refresh_token 到 localStorage

## 2. 重构 `nakama.ts` 客户端单例

- [x] 2.1 保留 Nakama 客户端单例创建逻辑（`Client` 实例）
- [x] 2.2 移除 `authenticateDevice()` 函数（设备认证）
- [x] 2.3 确认客户端配置仍从 `VITE_NAKAMA_HOST` / `VITE_NAKAMA_PORT` / `VITE_NAKAMA_SERVER_KEY` / `VITE_NAKAMA_USE_SSL` 环境变量读取

## 3. 重构 `useNakamaAuth` Hook

- [x] 3.1 将 `useNakamaAuth` 从设备认证改为 Session 恢复 + 邮箱登录/注册模式
- [x] 3.2 组件挂载时自动调用 `restoreSession()`，返回状态：`restoring`（恢复中）、`authenticated`（已登录，含 Session）、`guest`（未登录，需展示登录表单）
- [x] 3.3 提供 `login(email, password)` 方法，调用 `loginWithEmail()`（`create=false`）并更新状态
- [x] 3.4 提供 `register(email, password)` 方法，调用 `registerWithEmail()`（`create=true`），检查 `created` 字段：`true` → 切到 `authenticated`；`false` → 报错"该邮箱已注册，请直接登录"
- [x] 3.5 提供 `logout()` 方法，调用 `clearSession()` 并重置状态为 `guest`

## 4. 改造 `LoginPage.tsx`

- [x] 4.1 引入 `useAuth`（from AuthContext），读取 `status`、`login`、`register`、`error` 状态
- [x] 4.2 实现登录/注册页签切换（`activeTab: "login" | "register"`），两套独立的表单状态
- [x] 4.3 替换 `handleLogin`：移除 `setTimeout` mock，改为调用 `login(email, password)`
- [x] 4.4 实现 `handleRegister`：客户端校验（两次密码一致 + 密码 ≥ 8 位） → 调用 `register(email, password)` → 成功后自动跳转 `/home`
- [x] 4.5 增加错误状态显示：在表单上方展示错误信息（区分"邮箱或密码错误"、"无法连接服务器"、"该邮箱已注册，请直接登录"、"两次输入的密码不一致"、"密码长度至少需要8位"）
- [x] 4.6 登录/注册成功后自动跳转 `/home`（由 hook 状态变化触发 `useEffect` + `navigate`）
- [x] 4.7 管理一个独立的 `formError` 状态，收到 `error` 则 `formError = errorToMessage(error)`，点击登录/注册、修改邮箱密码时清除

## 5. 改造 `main.tsx` 路由守卫

- [x] 5.1 创建 `AuthProvider` 包装组件 + `ProtectedRoute`，在路由渲染前完成 Session 恢复
- [x] 5.2 路由 `/` 仍渲染 `LoginPage`，但 `LoginPage` 内部根据 `status` 展示不同 UI（`restoring` → 加载画面；`guest` → 登录表单；`authenticated` → 自动跳转 `/home`）
- [x] 5.3 确保非认证用户无法访问 `/home`、`/match`、`/gacha`、`/ranking`（未登录时重定向回 `/`）

## 6. 验证

- [X] 6.1 启动 Docker 环境（`docker compose up -d`），确保 Nakama 在 `localhost:7350` 可用
- [X] 6.2 验证 Session 恢复流程：先注册新账号 → 刷新页面 → 自动跳过登录页进入 `/home`
- [X] 6.3 验证邮箱注册：输入新邮箱+密码+确认密码 → 点击注册 → 自动创建账号并跳转 `/home` → PostgreSQL `users` 表确认新用户记录已写入
- [x] 6.4 验证注册密码校验：两次密码不一致 → 显示"两次输入的密码不一致"；密码 < 8 位 → 显示"密码长度至少需要8位"
- [X] 6.5 验证重复注册：已注册的邮箱再次注册 → 显示"该邮箱已注册，请直接登录"
- [X] 6.6 验证邮箱登录：输入已注册邮箱+正确密码 → 点击登录 → 跳转 `/home`；输入错误密码 → 显示"邮箱或密码错误" → 停留在登录页
- [X] 6.7 验证路由守卫：未登录直接访问 `/home` → 重定向回 `/`
- [X] 6.8 验证页签切换：登录/注册页签切换时，错误提示被清除
- [x] 6.9 TypeScript 类型检查通过：`cd client && npx tsc --noEmit`
