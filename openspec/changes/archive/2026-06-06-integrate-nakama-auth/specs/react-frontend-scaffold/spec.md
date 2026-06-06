## MODIFIED Requirements

### Requirement: Nakama JS SDK 集成

前端项目 SHALL 依赖 `@heroiclabs/nakama-js`，提供 Nakama 客户端单例模块（连接地址从 `VITE_NAKAMA_HOST` 环境变量读取，默认 `localhost:7350`），以及邮箱认证和 Session 恢复的 API 封装。

#### Scenario: 创建 Nakama 客户端

- **WHEN** 前端代码导入 Nakama 客户端模块
- **AND** 调用客户端创建函数，传入 server key `"defaultkey"` 和 host `"localhost"`、port `"7350"`、useSSL `false`
- **THEN** 返回已配置的 `Client` 实例

#### Scenario: 邮箱认证（Email Authentication）

- **WHEN** 用户提交邮箱和密码
- **AND** 前端调用 `client.authenticateEmail(email, password, true)`
- **THEN** Nakama 服务器返回有效的 `Session` 对象（账号不存在时自动创建）
- **AND** Session 包含 `token` 和 `refresh_token`
- **AND** Session 信息持久化到 `localStorage`

#### Scenario: 认证失败的降级处理

- **WHEN** Nakama 服务器不可达或认证信息无效
- **AND** 前端调用 `authenticateEmail`
- **THEN** 前端捕获错误并返回错误信息给调用方
- **AND** 不阻塞页面渲染（LoginPage 显示错误提示）

#### Scenario: Session 恢复

- **WHEN** 前端应用初始化时 localStorage 中存在 token 和 refresh_token
- **AND** 调用 Session 恢复逻辑
- **THEN** 尝试用 refresh_token 续期或直接恢复有效 Session
- **AND** 恢复成功则设置当前 session 并跳转受保护页面
- **AND** 恢复失败则清除过期 token，展示登录页

## REMOVED Requirements

### Requirement: 设备认证（Device Authentication）

**Reason**: 设备认证无法支持跨设备登录和账号密码体系，被邮箱认证 + Session 恢复方案取代
**Migration**: 使用 `loginWithEmail()` 函数替代 `authenticateDevice()`；使用 `restoreSession()` 替代基于 deviceId 的自动认证逻辑
