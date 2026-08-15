# CS2Match 低成本 AWS 生产部署

> 价格与配额核对日期：2026-08-15。购买前再次查看官方页面；汇率和税费会改变人民币金额。

如果你正准备在新服务器上实际操作，请直接按 [`aws-deployment-step-by-step.md`](aws-deployment-step-by-step.md) 逐步执行；本文主要解释架构、价格和安全取舍。

生产数据库固定为 PostgreSQL 15.18；该版本修复了 15.17 及更早版本的多项安全问题。[PostgreSQL 15.18 发布说明](https://www.postgresql.org/about/news/postgresql-184-1710-1614-1518-and-1423-released-3297/)

## 结论

推荐使用一台 AWS Lightsail 2 GB Linux 实例承载 Nakama 与 PostgreSQL，Cloudflare Pages 托管 React 静态站，Cloudflare Tunnel 暴露唯一 API 域名，Tailscale 承担 SSH/Console 私网管理，R2 保存 restic 加密备份。

Lightsail 2 GB 公网 IPv4 套餐当前是每月 12 美元、2 vCPU、2 GB 内存、60 GB SSD 和 3 TB 流量，约 86 元人民币，通常能把整套方案控制在 100 元以内。1 GB/7 美元虽然更贴近 70 元目标，但 Nakama、PostgreSQL、Docker 构建与整场战报 RPC 共用 1 GB 容易换页卡顿，不作为生产推荐。[AWS Lightsail 套餐](https://docs.aws.amazon.com/lightsail/latest/userguide/amazon-lightsail-bundles.html)

Cloudflare Pages Free 当前每月 500 次构建、每站最多 20,000 文件、单文件 25 MiB；R2 Standard 每月包含 10 GB、100 万 Class A、1,000 万 Class B 且不收公网出口费。这个小圈子规模预计都落在免费额度内，但它们是配额而非消费硬上限。[Pages 限制](https://developers.cloudflare.com/pages/platform/limits/)，[R2 价格](https://developers.cloudflare.com/r2/pricing/)

## 拓扑和安全边界

```text
浏览器 ──HTTPS──> Cloudflare Pages（前端）
   └────HTTPS/WSS──> api.example.com ──Cloudflare Tunnel──> Nakama:7350

管理员/GitHub Runner ──Tailscale──> SSH:22 / Console:7351
Nakama ──内部 Docker 网络──> PostgreSQL:5432
备份脚本 ──restic 客户端加密──> 私有 R2 bucket
```

源站不需要为 API 暴露公网 IP。Lightsail 防火墙最终关闭 22、5432、7350、7351 的公网入站；Compose 不发布数据库端口，7350 只绑定主机 loopback，7351 只绑定主机的 Tailscale IP。Tunnel 只有一条 API hostname ingress，REST 和 WebSocket 共用它，末尾必须是 404；Console 不进入 Tunnel。

Cloudflare 能隐藏 API 源站并吸收常见边缘攻击，但域名本身当然公开。限制账单风险还需要 AWS Budget 多级告警、R2 用量观察和定期磁盘检查；AWS Budget 只告警，不会自动阻止消费。不要配置自动创建大实例、snapshot 或无限日志的权限。

## 为什么不是 RDS、ECS 或 Kubernetes

50 人以内、当前一次 `SimuMatch` 返回完整战报的负载，单机 Compose 足够简单。RDS 的固定成本会明显挤压预算；ECS/Fargate 加数据库也更贵；Kubernetes 增加控制面、网络与运维复杂度。数据库同机意味着实例故障会同时影响应用和数据，所以必须依靠每日异地加密备份与恢复演练补偿。以后 CPU/内存基线持续超标，再把 PostgreSQL 移到托管数据库。

## 区域与性能

先在东京、新加坡等候选区域各创建同规格临时实例，从实际玩家所在地连续测晚高峰延迟、丢包和 `SimuMatch` p95；不要只看地理距离。选择多数玩家稳定且跨境线路较好的区域。测试实例当天删除，并检查 static IP、snapshot 和磁盘是否残留计费。

上线后运行：

```bash
cd /opt/cs2match/client
VITE_NAKAMA_HOST=api.example.com VITE_NAKAMA_PORT=443 \
VITE_NAKAMA_SERVER_KEY=你的客户端可见ServerKey VITE_NAKAMA_USE_SSL=true \
npm run load:production -- 20
# 低峰期再运行 50
```

记录成功率、p50、p95、平均响应字节，同时在主机观察 `docker stats`、磁盘和 PostgreSQL。目标不是承诺某个虚构 TPS，而是确认你的真实战报大小和阵容配置在 20/50 并发下没有明显换页、超时或错误。

主机可在负载期间运行 `deploy/scripts/capture-runtime-stats.sh 120 2`，把 Nakama/PostgreSQL 的 CPU、内存、网络与块 IO 采样写到 `deploy/state/*.csv`，再与负载 JSON 一起留档。

## 部署前置与首次上线

主机使用 Ubuntu 24.04 LTS x86_64，安装 Docker Engine 与 Compose plugin，再把 `ubuntu` 加入 docker 组并重新登录。仓库目录、变量生成和首次命令见 [`deploy/README.md`](../deploy/README.md)。生产使用：

- [`server/Dockerfile.prod`](../server/Dockerfile.prod)：Nakama 3.30.0 pluginbuilder 与运行镜像 ABI 对齐。
- [`docker-compose.prod.yml`](../docker-compose.prod.yml)：独立于开发 Compose 的生产栈。
- [`.env.production.example`](../.env.production.example)：不含真实值的模板，真实文件权限必须是 0600。
- [`deploy/cloudflared/config.yml.example`](../deploy/cloudflared/config.yml.example) 与 [`deploy/tailscale/policy.hujson.example`](../deploy/tailscale/policy.hujson.example)：人工控制台配置参考。

`NAKAMA_SERVER_KEY` 会被编译进浏览器，只是应用标识/基础门槛；数据库、session、refresh、runtime、Console、Tunnel、Tailscale、R2 与 restic 凭据才是真正 secrets。`SESSION_ENCRYPTION_KEY`、`SESSION_REFRESH_ENCRYPTION_KEY`、`RUNTIME_HTTP_KEY` 和 `CONSOLE_SIGNING_KEY` 各使用一次独立的 `openssl rand -hex 32` 输出，不得复用。

## 停服与恢复

临时关闭：

```bash
cd /opt/cs2match
deploy/scripts/service-control.sh close
```

这会停 Nakama 与 Tunnel，保留 PostgreSQL 容器/volume，但 Lightsail 仍按实例存在时间计费。重新开放执行 `service-control.sh open`。

长期停服先执行 `backup-db.sh final` 与 `restore-verify.sh latest`，记录成功结果，然后由你在 AWS 控制台显式删除实例，并逐项检查 static IP、snapshot、附加磁盘。仓库没有任何自动删除云资源的脚本。

恢复时创建新主机、恢复仓库与 `.env.production`，运行预检，再用 `restore-verify.sh` 在隔离库验证；覆盖生产库属于高风险操作，本提案刻意不自动完成，确认维护窗口与目标后再人工导入。
