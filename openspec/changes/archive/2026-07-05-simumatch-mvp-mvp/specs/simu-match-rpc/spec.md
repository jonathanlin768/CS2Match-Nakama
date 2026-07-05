## ADDED Requirements

### Requirement: DebugSimuMatch RPC is registered in InitModule

系统 SHALL 在 `server/main.go` 的 `InitModule` 中注册名为 `DebugSimuMatch` 的 RPC，使其可通过 Nakama REST API 调用。

#### Scenario: RPC 注册成功

- **WHEN** Nakama 加载编译后的 `backend.so`
- **THEN** `InitModule` 成功注册 `DebugSimuMatch` RPC
- **AND** Nakama 日志输出 RPC 注册成功信息
- **AND** 客户端可通过 `POST /v2/rpc/DebugSimuMatch` 调用该接口

### Requirement: DebugSimuMatch RPC parses and validates request

`DebugSimuMatch` RPC SHALL 解析 JSON 请求体，校验 `map_id` 字段，并在当前 MVP 阶段忽略阵容字段（使用服务端硬编码阵容）。

#### Scenario: Valid request with map_id

- **WHEN** 客户端发送 `{"map_id": "de_dust2"}`
- **THEN** RPC 解析成功并返回 HTTP 200
- **AND** 响应体包含完整战报

#### Scenario: Unsupported map_id

- **WHEN** 客户端发送 `{"map_id": "de_inferno"}`
- **THEN** RPC 返回包含错误码 `INVALID_MAP` 的 JSON 响应
- **AND** HTTP 状态码为 400

### Requirement: DebugSimuMatch response contains standardized match report

`DebugSimuMatchResponse` SHALL 包含比赛元信息、单回合战报、最终统计和胜者阵营，字段命名与 `internal/match/model.go` 中定义的 DTO 一致。

#### Scenario: Response structure validation

- **WHEN** 客户端成功调用 `DebugSimuMatch`
- **THEN** 响应 JSON 包含 `match_info`（含 `match_id`、`map_id`、`start_time`、`total_rounds`）
- **AND** 包含 `rounds[]`，其中每个回合包含 `round_number`、`winner`、`win_reason`、`score_t`、`score_ct`、`events[]`、`player_states[]`
- **AND** 包含 `final_stats`（含 `score_t`、`score_ct`、`player_stats[]`）
- **AND** 包含顶层字段 `winner`（值为 `"T"` 或 `"CT"`）

### Requirement: DebugSimuMatch uses hard-coded lineups in MVP

MVP 阶段 `DebugSimuMatch` SHALL 不依赖客户端传入阵容，而是使用服务端预先定义的两套 5 人阵容进行推演。

#### Scenario: Request without lineup still returns report

- **WHEN** 客户端发送空对象 `{}`
- **THEN** RPC 使用服务端硬编码阵容执行推演
- **AND** 返回完整战报

### Requirement: DebugSimuMatch returns structured errors

`DebugSimuMatch` SHALL 在失败时返回前端友好的 JSON 错误结构，包含 `code` 和 `message` 字段。

#### Scenario: Engine internal error

- **GIVEN** 引擎推演过程中发生异常
- **WHEN** 客户端调用 `DebugSimuMatch`
- **THEN** 返回错误码 `SIMULATION_ERROR`
- **AND** `message` 包含可读的错误描述
