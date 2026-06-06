## Context

当前 `LoginPage` 使用 `setTimeout` 假登录后直接 `navigate("/home")`。客户端的 `useNakamaAuth` hook 仅实现了设备认证（Device Auth），每次自动用浏览器指纹生成 deviceId 调用 `client.authenticateDevice()`。这有两个问题：1) 用户无法用账号密码登录/跨设备使用；2) 每次刷新页面都重新认证，没有 Session 恢复机制。

Nakama 内置了完整的 email auth 体系（`authenticateEmail`），且 SDK 的 Session 对象提供 `token`、`refresh_token` 和 `is_expired()` 方法，天然支持 Session 恢复。

本设计仅涉及前端改动，不需要修改 Go 插件代码。

## Goals / Non-Goals

**Goals:**
- Session 恢复：用户刷新页面或重新打开浏览器时，自动从 localStorage 恢复 Session，跳过登录页直接进入 `/home`
- 邮箱登录即注册：首次输入邮箱+密码时，Nakama 自动创建账号（`create=true`），用户无需区分"登录"和"注册"
- 登录失败反馈：密码错误、网络异常、Nakama 不可达等场景给出清晰的 UI 错误提示
- 保留现有 LoginPage UI 不变（品牌标识、邮箱/密码表单、记住我、社交登录按钮等）

**Non-Goals:**
- 不实现独立的"注册"页面（`create=true` 已覆盖注册场景）
- 不修改 Go 插件代码（使用 Nakama 内置 API）
- 不接入 Google/GitHub OAuth（UI 已有按钮，但保持 mock 状态）
- 不实现"忘记密码"功能（Nakama 需配合邮件服务，超出当前范围）

## Decisions

### 1. 认证流程：Session 恢复优先 → 邮箱登录

**选择**: 进入应用时先尝试从 token/refresh_token 恢复 Session → 成功则跳转 `/home` → 失败则展示 LoginPage → 用户提交邮箱密码 → `client.authenticateEmail(email, password, true)` → 持久化 token 并跳转

**理由**: 
- Nakama JWT token 默认 60 秒过期，refresh_token 可用于续期
- Session 恢复让已登录用户无需重复输入密码，体验流畅
- 只在恢复失败时才展示登录表单，避免不必要的表单展示

**备选方案**: Cookie-based session（Nakama 也支持 `HTTP Cookies`），但需要服务端配置且前端 fetch 需要 `credentials: 'include'`，与当前 SDK 直连模式不一致

### 2. 注册方式：`authenticateEmail` 的 `create=true` 参数

**选择**: 统一使用 `client.authenticateEmail(email, password, true)` — 用户已存在则登录，不存在则自动创建

**理由**:
- Nakama 原生支持，零后端工作
- 用户无需区分"登录"和"注册"，降低 UI 复杂度
- `create=true` 只在账号不存在时创建，已存在则忽略（等价于登录）

**备选方案**: 前端先调注册 RPC，再调登录 — 需要写 Go 代码，增加复杂度且无额外收益

### 3. Session 恢复实现

**选择**: 从 localStorage 读取 token 和 refresh_token → 尝试用 `client.authenticateToken(token)` 或构造 Session 对象检查 `is_expired()` → 过期则调用 `client.sessionRefresh(session)` → 成功则设置 session → 失败则清除 localStorage 并展示登录页

**理由**:
- `authenticateToken` 是 Nakama SDK 提供的 session restore API
- 如果 token 未过期，直接恢复，零网络请求
- 如果过期，用 refresh_token 换新 token，避免重新登录
- 只在 refresh 也失败时才要求用户重新输入密码

### 4. 文件组织

**选择**:
```
client/src/
  api/auth.ts        ← 新建：loginWithEmail(), restoreSession()
  nakama.ts          ← 改造：移除 authenticateDevice()，保留 client 单例
  hooks/useNakamaAuth.ts ← 改造：从设备认证改为 session 恢复 + 邮箱登录
  pages/LoginPage.tsx    ← 改造：handleLogin 改为真实 auth，增加 error 状态
  main.tsx               ← 改造：包裹 AuthProvider / 路由守卫
```

**理由**: `api/auth.ts` 是纯函数模块，与 React 解耦，方便测试和复用；`useNakamaAuth` hook 负责 React 状态管理；`LoginPage` 仅关注 UI 交互

### 5. 错误处理策略

**选择**: 区分三种错误类型并在 UI 上展示不同提示
- 网络错误（Nakama 不可达）→ "无法连接服务器，请检查网络"
- 认证错误（密码错误）→ "邮箱或密码错误"
- 未知错误 → "登录失败，请稍后重试"

**理由**: Nakama SDK 抛出的 Error 有不同 message，前端不需要解析错误码，用 `error.message` 即可。对用户展示中文友好提示

## Risks / Trade-offs

- **[风险] localStorage 中的 token 被清除或损坏** → Session 恢复失败，回退到登录页，用户重新输入即可
- **[风险] `authenticateEmail` 的 `create=true` 可能被滥用创建垃圾账号** → 当前实验项目（100 人在线），暂不需要验证码或邮箱验证。后续可通过 Go RPC 添加创建限制
- **[权衡] 未实现设备认证降级** → 用户明确要求：登录失败即拒绝，不降级到设备认证。这保证了账号体系的一致性（都通过邮箱注册）
- **[风险] refresh_token 也有过期时间** → Nakama 默认 refresh token 有效期较长（数天），如果两 token 都过期，用户重新登录即可

## Open Questions

- 后续是否需要邮箱验证（Nakama 内置了验证邮件功能，需配置 SMTP）？当前 `create=true` 不验证邮箱
- "记住我"复选框是否需要实际效果？（当前保持 UI，暂不改变 token 过期策略）
