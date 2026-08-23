# 轻量客户端发布与回滚

## 功能开关

- 服务端：`nakama-config.yml` 的 `runtime.env` 中设置 `SOCIAL_CONTACT_EXCHANGE_ENABLED=false`，重启 Nakama 后停止注册联系方式 RPC；客户端文本频道写入拦截仍保持启用。
- 客户端：构建参数 `VITE_SOCIAL_CONTACT_EXCHANGE_ENABLED=false` 会隐藏联系方式配置与交换入口，原生好友列表继续可用。

两个开关应同时关闭。Storage 数据无需删除，恢复开关后仍可读取已授权状态。

联系方式授权绑定双方整份资料的 revision。任一方实际修改、新增或清空 QQ/微信都会让整次授权变为 `stale`，客户端不得再显示任何渠道正文，双方重新申请并接受后才能恢复。删除好友会通过 `BeforeDeleteFriends` hook 把授权持久标记为 `revoked`；重新添加好友不会复活旧授权。

当前本地与生产配置已经分别在 `nakama-config.yml` / `nakama-config.prod.yml` 和 `client/Dockerfile` 中启用对应开关，本次变更不需要新增环境变量。上线时必须同时更新 Nakama Go 插件和前端镜像。

## 回滚

发布前为当前稳定前端镜像打标签：

```powershell
docker image tag cs2match-nakama-frontend:latest cs2match-nakama-frontend:rollback-lightweight-client
```

如新界面出现严重回归，使用稳定标签重新创建 frontend；服务端新增 Storage 与 RPC 可保留，不涉及数据库 schema 回滚。Nakama 插件回滚时恢复上一份 `server/build/backend.so` 并重启 `nakama`。

## 监控项

- device/email 认证失败率与 8 位 username 唯一冲突。
- `SimuMatch` 的 `INVALID_LINEUP`、`BUDGET_EXCEEDED`、`CONFIG_VERSION_MISMATCH` 和模拟错误。
- `social_contact_profile` / `social_contact_exchange` Storage version 冲突。
- 服务端卡片投递失败；授权以 Storage 为准，因此失败后客户端仍可通过 RPC 恢复。
- `NOT_FRIENDS`、`FORBIDDEN` 等越权读取/响应拒绝数量。
- 客户端生产构建错误、首页到战斗转化和 `/battle` 空状态异常进入量。
