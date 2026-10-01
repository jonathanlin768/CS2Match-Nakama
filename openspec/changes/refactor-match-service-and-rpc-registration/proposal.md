## Why

RPC 注册重复了日志与错误处理，比赛服务在一个 switch 中混合模式选择、配置校验和组队。三个已确认的 TODO 需要通过明确命名、职责提取和静态注册清单改善维护性。

## What Changes

- 将比赛服务的引擎服务依赖与构造参数统一命名为 `engineService`。
- 使用保持 JSON 字符串值的 `MatchMode` 常量，将 computer / tutorial 的准备过程提取成函数，返回统一内部结果。
- 增加类型明确的 RPC 注册清单和小型公共注册助手；Match 拥有自己的注册入口，Social 改为有序清单，统一日志与错误包装。
- 保留现有 RPC 名称、请求/响应、模式校验、权限检查、教学阵容规则和模拟调用。

## Capabilities

### New Capabilities
- `backend-service-organization`：后端静态 RPC 注册、模式准备边界与命名约定。

### Modified Capabilities
无既有业务要求变更。

## Impact

- Nakama 后端：`main.go`、`internal/match`、`internal/social` 和新的 `internal/framework/rpcregistry` 包。
- React 前端：HTTP JSON 保持兼容，使用现有客户端检查；不改变页面或重建前端。
- 数据库 / 部署：无迁移，使用项目构建脚本编译并更新本地后端。
- 不新增 RPC、Match Handler、Storage 操作、Go module 或依赖，不更新 Luban 配表，不改变玩法或状态同步。
