# 结构重构验证记录

日期：2026-10-01。

## 变更范围

- 保留单个 `matchengine` package 和同目录测试，按职责拆分输入、战报、地图配置与错误模型，补充 `doc.go` 和维护索引。
- 将包内运行时类型、resolver 和常量改为未导出名称；保留业务入口、配置校验与现有离线标定接口。公开 Go 声明从 377 个收敛到 80 个，297 个内部声明不再暴露给导入方。
- 队列查询、清空与拦截去重归调度器管理；回合时钟与计数方法归位 `round_state.go`；移除无用途的回合到整场对象引用。
- 更新当前教程和源码追踪引用；将因果引擎的 core / config / map delta 合并到主规格，保留后续视觉、教学阵容和比赛实例身份要求。历史 Review 与原始设计变更保留历史语境。
- 没有接入新的玩法路径。情报、包点决策、移动拦截/进度、恢复完成、开局重试与同秒跨行动缺口记录在引擎 README 中，不能用内部测试通过宣称这些生产路径完成。

内部符号的收紧是 Go 源码层面的 API 变更；仓库内现有业务与标定调用继续编译。公开 JSON 字段和枚举字符串保持原值。

## 通过的检查

| 检查 | 结果 |
|---|---|
| 重构前 `go test ./internal/framework/matchengine ./internal/match -count=1` | 两个包通过 |
| 重构后同一组包测试 | 两个包通过 |
| 重构后 `go test ./... -count=1`（server module） | 所有包通过 |
| `go test ./internal/framework/matchengine -run '^TestServicePublicContract' -count=1` | 公开契约通过 |
| 三个固定 seed：11、42、20261001 | 完整 `MatchResult` JSON SHA-256 与重构前完全一致；不是重构后重建预期值 |
| 公开契约的附加检查 | 输入快照保持只读、完整比赛输出、无效输入结构化错误、Context 取消 |
| `node --experimental-strip-types --test src/pages/battle-playback.test.mjs`（client） | 11/11 通过 |
| `npm test`（client，恢复锁定依赖后） | 29/29 通过 |
| 当前教程、设计文档和追踪表的 `.go#symbol` 引用 | 149 处，未解析引用 0 处 |
| 本次 change 与四份修改后的主规格严格校验 | 均通过 |
| `git diff --check` | 无差异格式错误 |
| 原有用户改动 | `round_engine.go` 的内联注释保留；`internal/match/api_rpc.go`、`internal/match/service.go`、`main.go` 的原有注释未修改 |

## Docker 恢复后的构建与运行验证

### Linux Nakama plugin 编译

2026-10-01 14:59（Asia/Shanghai），按 `cs2match-local-update` skill 从仓库根目录运行：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File server/build.ps1
docker compose up -d --build db nakama
docker compose restart nakama
```

构建脚本使用 `heroiclabs/nakama-pluginbuilder:3.30.0`、Go 1.24.5 linux/amd64，以及 `-buildmode=plugin -trimpath -gcflags='all=-N -l'`。先生成 `backend.so.tmp`，成功后替换 `server/build/backend.so`。产物大小 21,880,760 bytes。

宿主文件与容器 `/nakama/data/modules/backend.so` 的 SHA-256 一致：

```text
c1a00d42739576bad6a853a8da0bf690138a59f764681664038fe0bd7762022d
```

Nakama 重启后日志显示 `CS2Match Go plugin loaded successfully`、`Go runtime modules loaded`，HealthCheck / DebugSimuMatch / SimuMatch 注册成功。容器状态 healthy，实际调用后最新 100 行日志没有 error / fatal / panic 或插件加载失败。

### race 检查

在相同固定 Linux builder 中，只读挂载 `server/`，设置 `CGO_ENABLED=1`，执行：

```sh
go test -mod=readonly -race ./internal/framework/matchengine ./internal/match -count=1
```

两个包均通过（38.841s / 10.304s），没有 race 报告。引擎外部契约测试也在该 Linux 测试中执行，三个重构前完整战报基线继续通过。

### 真实服务调用

使用隔离的本地测试设备账号 `codex-matchengine-refactor-smoke` 完成认证，并以 Bearer session 调用实际 HTTP RPC。凭据与 session token 不写入记录。

| 检查 | 结果 |
|---|---|
| HealthCheck RPC | `status=ok` |
| 设备认证 | 成功取得 session |
| DebugSimuMatch：`map_id=de_dust2`、`seed=42` | 24 回合，11:13，10 名选手，530 个事件，134 次击杀 |
| SimuMatch：`mode=computer` | 19 回合，6:13，10 名选手，384 个事件，98 次击杀 |
| Go Remote 端口 2345 | TCP 连接通过 |

两场比赛都验证了完整回合数、连续回合编号、每回合 10 人状态、队伍胜者身份、总比分与回合数一致、终局事件与原因存在、事件时间有序、最后一回合比分与最终摘要一致、击杀事件总数与选手统计一致。

本次后端构建没有改动 `go.mod`、`go.sum` 或生成配置；前端未重建。GoLand 若已连接旧进程，需要在 Nakama 重启后重新连接 Go Remote。

### 历史环境阻塞（已解决）

首次验证时 Docker Linux daemon 不可用，启动日志显示旧的 `Docker/run/dockerInference` socket 无法访问。可恢复的移动尝试失败，未改动该文件或 Docker 配置，结束了失败的启动尝试。本机当时 `CGO_ENABLED=0`，无法运行 race。

用户启动 Docker 后，上述 Linux 编译、race 和真实服务验证均已完成，任务 4.3 / 4.5 不再阻塞。
