# Health Check RPC

最小可用的 RPC 端点，用于验证前后端全链路连通性。

## Requirements

### Requirement: HealthCheck RPC 端点

Go 插件 SHALL 注册一个名为 `HealthCheck` 的 RPC 函数，接受空 payload，返回包含服务器时间戳和状态的 JSON 响应。

#### Scenario: 通过 http_key 调用 HealthCheck（无需认证）

- **WHEN** 通过 HTTP POST `http://localhost:7350/v2/rpc/HealthCheck?http_key=defaultkey` 发送空 payload `""`
- **THEN** 返回 JSON：`{"status": "ok", "timestamp": <ISO8601 时间字符串>, "version": "0.1.0"}`
- **AND** HTTP 状态码为 200

#### Scenario: 前端通过 SDK 调用 HealthCheck（已认证）

- **WHEN** 前端已获取有效 Session
- **AND** 前端调用 `client.rpc(session, "HealthCheck", {})`
- **THEN** 返回的 JSON payload 包含 `status: "ok"`
- **AND** 前端在控制台输出连接成功信息

### Requirement: 前后端全链路验证

系统 SHALL 提供一条完整的验证路径：前端启动 → 设备认证获取 Session → 调用 HealthCheck RPC → 显示响应结果。

#### Scenario: 全链路成功

- **WHEN** Docker 环境已启动（Nakama + PostgreSQL）
- **AND** Go 插件已编译并挂载
- **AND** 前端开发服务器已启动
- **AND** 开发者打开浏览器访问前端
- **THEN** 前端自动完成设备认证
- **AND** 前端自动调用 HealthCheck RPC
- **AND** 页面显示"服务器连接成功"（客户端表现）
- **AND** 浏览器控制台输出服务器返回的状态信息
