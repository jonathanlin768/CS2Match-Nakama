# Go Plugin Scaffold

Go 插件项目骨架规格 — 模块路径、编译工具链、入口函数。

## Requirements

### Requirement: Go 项目目录结构

Go 插件项目 SHALL 位于 `server/` 目录，包含 `go.mod`、`main.go`（`InitModule` 入口函数）、编译脚本 `build.sh`/`build.ps1`。

#### Scenario: 目录结构完整性

- **WHEN** 查看 `server/` 目录
- **THEN** 存在 `go.mod` 文件，模块名为 `windypath.com/cs2match/server`
- **AND** 存在 `main.go`，包含 `InitModule` 函数并注册至少一个 RPC
- **AND** 存在 `build.sh`（Linux/Mac）和 `build.ps1`（Windows）编译脚本

### Requirement: Go 依赖管理

`go.mod` SHALL 引入 `github.com/heroiclabs/nakama-common` 依赖，版本号必须与 Nakama 3.30.0 编译时使用的版本一致（当前为 v1.40.0，通过 `go get github.com/heroiclabs/nakama/v3@v3.30.0` 可查询精确依赖）。

#### Scenario: 依赖下载

- **WHEN** 开发者在 `server/` 目录执行 `go mod download`
- **THEN** `nakama-common` 及其传递依赖被成功下载
- **AND** `go mod verify` 验证通过

### Requirement: 插件编译

编译使用官方 `heroiclabs/nakama-pluginbuilder:3.30.0` Docker 镜像（Go 1.24.5，依赖与 Nakama 二进制精确匹配）。编译输出为 `server/build/backend.so`。

#### Scenario: 编译插件

- **WHEN** 开发者执行 `bash server/build.sh` 或等效 Docker 命令
- **THEN** `server/build/backend.so` 文件被生成
- **AND** 文件格式为 Linux ELF 64-bit shared object
- **AND** 编译命令中设置了 `-buildmode=plugin -trimpath`

### Requirement: InitModule 入口函数

`main.go` SHALL 定义 `InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error` 函数，在该函数中注册 RPC 和 Hook。

#### Scenario: InitModule 执行成功

- **WHEN** Nakama 加载编译好的 `.so` 插件
- **THEN** `InitModule` 函数被调用
- **AND** 函数返回 `nil`（无错误）
- **AND** Nakama 日志输出插件初始化成功信息
