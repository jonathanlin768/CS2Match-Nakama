# 验收记录

日期：2026-08-23（Asia/Shanghai）

## 自动化验证

- Go：`gofmt`、`go test ./... -count=1`、`go vet ./...` 和 `server/build.bat` 通过，生成 `server/build/backend.so`。
- 前端：`npm test`、`npm run lint`、`npm run build` 通过。
- Mock 浏览器组件验收：权威申请分组、全局徽章、接受/拒绝、资料回填与失败重试、修改确认、`stale` / `revoked` 不披露正文以及 390×844 响应式入口通过。

## 本地 Docker 集成

- 后端插件重载后 `nakama` 容器健康；日志确认新增本人资料 RPC、权威 inbox RPC 和 `deletefriends` before hook 已注册，功能开关为 `true`。
- 前端使用 `docker compose build --no-cache frontend` 重建并强制替换，`http://localhost:3000` 返回 HTTP 200。
- 使用两个全新、互相隔离的 Chromium BrowserContext 创建 A/B 正式账号并验证：加好友；B 离线时 A 发起；B 返回后仅靠权威 inbox 恢复申请；B 接受后双方刷新均能查看对方授权渠道。
- B 修改 QQ 后，双方整个交换立即变为 `stale`，旧 QQ/微信正文均不再返回；限流窗口结束后重新申请并接受，双方才重新获得当前资料。
- A 删除 B 后授权持久变为 `revoked`；重新添加好友未恢复旧正文。
- 客户端直接读取本人私有 profile Storage 返回空对象，直接写入被拒绝，伪造 request ID 的响应 RPC 被拒绝。

## PII 检查

- 权威 inbox 响应、错误响应、Toast 和浏览器 localStorage 未出现对方联系方式正文。
- 验收时间窗口内 Nakama 日志未匹配到已知测试 QQ/微信正文。
- PostgreSQL `message` 表中的联系方式交换卡片仅包含 `type`、`action`、`version` 和 `request_id`，未包含 QQ/微信正文。

## 单渠道补充验收

- 服务端测试确认只填写合法 QQ、微信留空可以保存；只填写合法微信号、QQ 留空也可以保存。
- 客户端只把当前已保存的渠道加入交换请求；界面与 `PROFILE_INCOMPLETE` 文案明确提示“QQ 或微信至少填写一项即可”。
