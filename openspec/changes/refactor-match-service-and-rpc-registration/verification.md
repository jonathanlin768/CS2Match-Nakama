# 验证记录

日期：2026-10-01（Asia/Shanghai）。

## 变更结果

- `match.Service.engineService` 与构造参数命名一致；三个已确认 TODO 已落实并移除。
- `model.go` 定义字符串 `MatchMode`。`mode_prepare.go` 中短 switch 分派至两个准备函数，输出 `preparedMatch` 后进入共同模拟与响应流程。
- `match.RegisterRPCs` 拥有两个比赛 RPC 清单；`rpcregistry.Register` 按切片顺序注册、记录成功日志、包装错误并保留错误原因。
- HealthCheck 与 Social 复用同一助手，Social 六个 RPC 改成有序清单，启用开关和 hooks 保留。
- 更新当前项目结构和教程说明；前端协议、RPC 名称、配表和比赛算法保持原值。

## 自动检查

| 检查 | 结果 |
|---|---|
| 修改前 Match / Social 基线测试 | 通过 |
| `go test ./... -count=1`（server） | 所有包通过，包含引擎原有完整战报基线 |
| 固定 Linux builder 中 `go test -mod=readonly -race ./internal/framework/rpcregistry ./internal/match ./internal/social -count=1` | 三个包通过，无 race 报告 |
| 注册助手边界测试 | 顺序和 handler 绑定正确；第二个注册失败后停止，错误保留名称与 errors.Is 原因 |
| Match RPC 边界测试 | JSON 模式字符串兼容；空/未知/大小写错误模式返回 INVALID_MODE；两个 RPC 无用户身份时返回 UNAUTHORIZED，不进入模拟 |
| 现有 computer / tutorial 集成测试 | 默认队伍、配置版本、阵容重复、预算和双方实例身份继续通过 |
| `npm test`（client） | 29/29 通过 |
| `gofmt -l` 本次修改的 Go 文件 | 无未格式化文件 |
| OpenSpec strict / `git diff --check` | 通过 |

## 编译、加载与运行

按 `cs2match-local-update` skill 从仓库根目录执行：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File server/build.ps1
docker compose up -d --build db nakama
docker compose restart nakama
```

使用固定 `heroiclabs/nakama-pluginbuilder:3.30.0` 和 Go 1.24.5。插件成功生成并更新，15:29:49 写入 `server/build/backend.so`，大小 21,889,912 bytes。宿主与容器挂载插件的 SHA-256 一致：

```text
23be8d02e30d101229a06a0472ced5af8fd02d28ccdcde71a7cd9f497a76479f
```

Nakama 新进程日志确认插件加载成功，九个既有 RPC 全部注册，注册成功日志统一来自 `rpcregistry/registry.go`。db / frontend / nakama 均 healthy；最新 100 行日志没有 error / fatal / panic 或插件加载失败，Go Remote 2345 可连接。

使用既有本地隔离测试设备账号 `codex-matchengine-refactor-smoke`，不记录凭据或 session token：

| 实际 HTTP 调用 | 结果 |
|---|---|
| HealthCheck | status=ok |
| 设备认证 | 成功 |
| DebugSimuMatch：Dust2 / seed=42 | 24 回合，11:13，530 个事件，134 次击杀 |
| SimuMatch：computer | 18 回合，5:13，366 个事件，91 次击杀 |
| SimuMatch：tutorial / tutorial_default v1 / 有效五人阵容 | 20 回合，13:7，507 个事件，130 次击杀 |
| 未知模式 | INVALID_MODE |
| 过期教学版本 | CONFIG_VERSION_MISMATCH |
| 两个比赛 RPC 无用户 Session | UNAUTHORIZED |
| SocialGetContactProfile 使用访客 Session | 按既有规则返回 FORMAL_ACCOUNT_REQUIRED |

三场比赛均验证了完整回合数、连续回合编号、每回合 10 人状态、胜者身份、比分与总回合数一致、终局事件和原因、事件时间顺序、最终比分与最后回合一致、选手击杀统计与击杀事件总数一致。人机和教学模式的 seed 由服务生成，上表记录本次实际调用结果。

Social 最初检查误以为访客可读取 ContactProfile；源码确认 GetProfile 也要求正式账号，随后验证其 FORMAL_ACCOUNT_REQUIRED 拒绝行为正确。没有更改该权限或执行社交写入。四个 Social hooks（好友删除和三种客户端聊天写入拒绝）均由启动日志确认注册。

## 总注册入口修正后的验证

根据用户反馈，`registerRpcFunc` 将直接返回 `match.RegisterRPCs(...)` 改为检查该模块注册错误，全部注册完成后统一返回 `nil`，后续模块可以继续添加注册调用。

- `go test .`：入口包编译通过。
- `server/build.ps1`：Linux Go 插件构建成功，随后按本地更新流程重启 Nakama。
- Nakama healthy，HealthCheck 返回 `status=ok`；最新启动日志确认九个既有 RPC 注册成功，没有 error / fatal / panic 或插件加载失败。
- 本次宿主与容器插件 SHA-256 一致：`eaf5305d9c443510c6d11739655be955c3ca37ba5be1fa49e3cdc9c661f0c78b`。

## 提交前复核

用户要求在确认编译、比赛模拟和业务兼容后，将全部未提交内容提交并推送到 `origin/master`。本次复核结果：

- `go test ./... -count=1`：全部 Go 包通过。
- 固定 Linux pluginbuilder 中对 matchengine、rpcregistry、match、social 执行 `go test -mod=readonly -race ... -count=1`：四个包通过，没有 race 报告。
- seed 11、42、20261001 的完整战报 SHA-256 再次与重构前基线一致。
- `npm test`：29/29 通过。
- 再次运行 `server/build.ps1` 并加载新插件，宿主与容器 SHA-256 仍为上节记录值；Nakama healthy，HealthCheck 成功，九个 RPC 注册成功，无运行或插件加载错误。

| 实际 HTTP 比赛调用 | 回合 | 比分 | 事件 | 击杀 |
|---|---|---|---|---|
| DebugSimuMatch / de_dust2 / seed=42 | 24 | 11:13 | 530 | 134 |
| SimuMatch / computer | 20 | 7:13 | 411 | 104 |
| SimuMatch / tutorial_default v1 | 16 | 13:3 | 391 | 99 |

三场战报再次通过回合连续性、十个独立选手实例、队伍身份、累计比分、终局事件及原因、事件时间顺序和击杀聚合检查。调试场结果与前次相同；正式模式按既有规则生成随机 seed。INVALID_MODE、CONFIG_VERSION_MISMATCH、两个比赛 RPC 的 UNAUTHORIZED 和 Social 的 FORMAL_ACCOUNT_REQUIRED 均符合原有规则。

两项本次 OpenSpec 变更的严格校验通过。全量主规格严格校验为 22/24 通过；health-check-rpc 与 react-frontend-scaffold 因缺少 Purpose 标题失败。与 HEAD 比较确认这两份文件没有改动，属于既有格式问题。

## 差异范围

保留上一次引擎结构重构和用户其他未提交内容。`api_rpc.go` 的未知字段说明注释、`round_engine.go` 的 started 内联注释仍保留。没有修改 Go 依赖、生成配置或前端源码。全部改动按用户要求一并提交，OpenSpec 变更保留在当前目录。重启 Nakama 后，已连接的 GoLand 调试器需要重新连接。
