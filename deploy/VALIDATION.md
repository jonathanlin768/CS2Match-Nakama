# 部署提案验证记录

验证日期：2026-08-15。以下均在无真实云账号、无生产凭据的本地环境完成。

| 范围 | 命令/方法 | 结果 |
|---|---|---|
| Go | `cd server && go test ./...`（Go 1.24.5） | 全部通过；没有新增 RPC、Match Handler、Storage 权限或 Luban 协议变更，`server/main.go` 仅增加注册完成日志 |
| 前端单测 | `npm test` | 20/20 通过，含生产 host/443/SSL/server key fail-closed 测试 |
| 前端生产构建 | `npm run build:production`，使用 `.invalid` 测试域名和客户端可见测试 key | 成功；`dist` 59 个文件，最大文件 2,974,765 bytes，低于 Pages 的 20,000 文件与 25 MiB/文件限制 |
| npm 依赖审计 | `npm audit` | 生产与开发依赖均为 0 个已知漏洞；React Router 已升级至 7.18.2 |
| 前端 secret 扫描 | 在 `dist` 搜索 DB/session/refresh/runtime/Tunnel/Tailscale/R2 secret 名称 | 未发现 |
| 部署静态测试 | `node --test deploy/tests/*.test.mjs` | 9/9 通过：内部数据库、端口绑定、Tunnel 单 ingress、镜像 digest、Action SHA、回滚/备份、SPA 回退 |
| 脚本失败演练 | `bash deploy/tests/script-failures.sh` | 通过：缺省/重复 key、权限（Linux CI）、dump/对象存储失败、并发锁、错误解密/损坏快照均 fail closed |
| Workflow | actionlint 1.7.12（官方 SHA256 校验） | 零报错；workflow 内也固定运行相同版本 |
| Compose | 使用无真实凭据的临时环境执行 `docker compose -f docker-compose.prod.yml config --quiet` | 通过 |
| 后端镜像 | `docker build -f server/Dockerfile.prod` | 成功；Nakama/pluginbuilder 均为 3.30.0，最终上下文约 284 KB，`.so` 路径正确，镜像环境层无项目 secrets |
| Nakama 集成 | 临时 PostgreSQL 15 + 生产镜像迁移/启动 | 16 个 migration 成功；`backend.so` 加载；HealthCheck、DebugSimuMatch、SimuMatch、4 个 social RPC 与身份/聊天 hooks 均注册；临时容器和网络已清理 |
| 备份恢复 | mock 演练 | 控制流已验证；真实 R2 加密上传与隔离恢复必须在外部环境验收 |

本机尝试重新构建既有 `client/Dockerfile` 时，Docker 内的 `npm ci` 长时间无输出后被终止；其 Dockerfile 未改变，新增 `.dockerignore` 只排除本地依赖/产物。Nginx `try_files` 与 Pages `_redirects` 的 SPA 回退已由静态测试覆盖，Linux CI 会继续执行前端生产构建。

## 真实环境待验收

- Lightsail 防火墙从公网验证 22/5432/7350/7351 不可达，API 只能经 Tunnel，SSH/Console 只能经 Tailscale。
- Pages 域名加载、设备/邮箱认证、Session 恢复、HealthCheck、SimuMatch、好友/联系方式卡片、WebSocket 与 Nakama 短暂重启恢复。
- named volume 重建保留、真实 R2 备份、错误密码失败、隔离恢复与恢复日期记录。
- 20/50 并发负载 JSON 和同时间段 CPU/内存/IO CSV。
- 日志轮转、磁盘余量、容器健康、备份失败通知、AWS Budget 多级告警；Budget 仅告警，不是硬消费上限。
- GHCR public pull、Tailscale OIDC、后端失败回滚、Pages 发布/回滚各完成一次受控演练。
