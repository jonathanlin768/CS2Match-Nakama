# simu-match-rpc Specification

## Purpose
TBD - created by archiving change simumatch-mvp-mvp. Update Purpose after archive.
## Requirements
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

`DebugSimuMatchResponse` SHALL 包含完整比赛元信息、所有常规赛/加时回合战报、最终队伍比分、最终选手统计和胜者队伍身份。字段命名 SHALL 与 `internal/match/model.go` DTO 和前端 `MatchReport` 类型保持一致。

#### Scenario: 响应结构校验
- **WHEN** 客户端成功调用 `DebugSimuMatch`
- **THEN** 响应 JSON 包含 `match_info`，其中有 `match_id`、`map_id`、`map_name`、`team_a_name`、`team_b_name`、`start_time` 和 `total_rounds`
- **AND** 响应 JSON 包含所有 `rounds[]`，每个回合包含 `round_number`、阵营/队伍身份、胜者阵营、胜者队伍 ID、获胜原因、队伍比分、路线/模板 ID、`events[]`、`player_states[]`、炸弹状态，以及可用时的最终控制权
- **AND** 响应 JSON 包含 `final_stats`，其中有最终队伍比分和 `player_stats[]`
- **AND** 响应 JSON 包含顶层 `winner_team_id`
- **AND** 响应 JSON 包含由业务配置派生的顶层布尔值 `debug_enabled`

### Requirement: DebugSimuMatch uses hard-coded lineups in MVP

`DebugSimuMatch` SHALL 在客户端未传入阵容数据时继续可用，但默认阵容 SHALL 由 `server/internal/match` 从 Luban `TbPlayer` 加载，并作为自包含输入快照传给 `matchengine`。

#### Scenario: 不传阵容的请求仍返回战报
- **WHEN** 客户端发送空 JSON 对象 `{}`
- **THEN** RPC 使用业务层默认队伍执行完整比赛模拟
- **AND** 响应包含两支队伍，且每队正好 5 名选手
- **AND** 响应包含完整比赛战报，而不是单回合战报

### Requirement: DebugSimuMatch returns structured errors

`DebugSimuMatch` SHALL 在失败时返回前端友好的 JSON 错误结构，包含 `code` 和 `message` 字段。

#### Scenario: Engine internal error

- **GIVEN** 引擎推演过程中发生异常
- **WHEN** 客户端调用 `DebugSimuMatch`
- **THEN** 返回错误码 `SIMULATION_ERROR`
- **AND** `message` 包含可读的错误描述

### Requirement: DebugSimuMatch 使用默认完整比赛规则

当请求未提供显式调试规则覆盖时，`DebugSimuMatch` SHALL 使用默认 MR12/加时规则集。

#### Scenario: 默认请求可以产生加时
- **GIVEN** 生成的模拟在常规赛后达到 12-12
- **WHEN** `DebugSimuMatch` 使用默认规则集
- **THEN** 返回战报包含加时回合
- **AND** 比赛胜者在加时后被判定

### Requirement: DebugSimuMatch 接受可选 seed

`DebugSimuMatch` SHALL 接受可选数值型 `seed` 请求字段。提供该字段时，RPC SHALL 将它传入 `matchengine.MatchInput.Seed`；未提供时，`match.Service` SHALL 生成 seed，并将有效 seed 包含在响应元信息中。

#### Scenario: 请求 seed 可复现比赛
- **GIVEN** 客户端发送 `{"map_id": "de_dust2", "seed": 123456}`
- **WHEN** 客户端连续两次发送同一请求
- **THEN** 两个响应包含相同的有效 seed
- **AND** 两个响应包含相同的回合胜者、事件、比分和最终胜者

#### Scenario: 缺失 seed 时生成 seed
- **GIVEN** 客户端发送 `{"map_id": "de_dust2"}`
- **WHEN** `DebugSimuMatch` 开始模拟
- **THEN** 服务生成一个非零 seed
- **AND** 响应元信息包含该 seed，供后续复现使用

### Requirement: DebugSimuMatch 返回地图配置错误

当地图语义配置无法加载或校验失败时，`DebugSimuMatch` SHALL 返回结构化错误。

#### Scenario: 非法地图配置
- **GIVEN** Dust2 配置包含一条引用缺失节点的路线
- **WHEN** 客户端调用 `DebugSimuMatch`
- **THEN** RPC 返回 code 为 `CONFIG_BAD_ROUTE_NODE` 的 JSON 错误
- **AND** 不返回部分比赛战报

### Requirement: MatchPage 展示匹配与模拟阶段反馈

当用户点击“开始匹配”并等待 `DebugSimuMatch` 返回时，客户端 SHALL 展示清晰的阶段性 loading 文案，使用户理解系统正在匹配并一次性模拟完整比赛。

#### Scenario: 开始匹配显示阶段文案
- **GIVEN** 用户已登录并访问 `/match`
- **WHEN** 用户点击“开始匹配”
- **THEN** 按钮或页面状态先显示“匹配中”
- **AND** RPC 等待期间显示“正在模拟比赛”或等价文案
- **AND** RPC 成功后进入 `/battle` 前显示“进入战场”或等价文案

