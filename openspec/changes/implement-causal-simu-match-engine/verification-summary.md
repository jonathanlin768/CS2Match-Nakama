# 验收摘要

## 生成数据

- 所有表源修改位于 `configs/Datas/#*.xlsx`。
- 已运行 `powershell -File scripts/gen-config.ps1`，同时更新服务端 Go/JSON 与客户端 TypeScript/JSON。
- 已运行 `go mod vendor` 同步嵌套 `config` 模块生成数据；没有手工修改生成 JSON。

## 自动化测试与构建

- `go test ./internal/framework/matchengine ./internal/match`：通过。
- `go test ./...`（`server/config`）：通过。
- `npm test`：2 个权威比分/炸弹回放测试通过。
- `npm run build`：通过；仅有 Vite 大 chunk 警告。
- `server/build.bat`：通过，生成 `server/build/backend.so`。
- `SIMU_LONG_CALIBRATION=1 go test ./internal/match -run TestLubanCalibrationLongSample -count=1 -v`：10,000 基准 + 2,000 受控强弱回合通过。

## 静态范围复核

- 生产 `matchengine`/`match` 源码未发现 `TargetWinner`、`ForcedRoundWinners`、目标幸存者、结果重采样或测试钩子。
- 没有新增 Go module、npm package、RPC、Match Handler、Storage 操作、数据库迁移或外部运行时依赖。
- `DebugSimuMatch` 仍只构造正式 `MatchInput` 并调用 `Simulate`；内部 round simulator 只用于 Match 层、测试和离线标定。
