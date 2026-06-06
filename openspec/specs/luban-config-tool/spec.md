# luban-config-tool Specification

## Purpose
TBD - created by archiving change luban-config-tool. Update Purpose after archive.
## Requirements
### Requirement: 策划配表目录结构

项目 SHALL 提供标准化的策划配表目录 `configs/`，包含以下子目录：
- `configs/Defines/` 存放 Luban 结构定义文件（XML）
- `configs/Tables/` 存放策划维护的 Excel 数据表
- `configs/luban.conf` 作为 Luban 导表配置入口

`configs/` 目录下的所有内容 SHALL 纳入 Git 版本管理。

#### Scenario: 策划新增一张配表

- **GIVEN** 策划在 `configs/Tables/` 下创建 `skill.xlsx`
- **WHEN** 在 `configs/Defines/__root__.xml` 中注册该表的结构定义
- **THEN** 运行导表脚本后，Server 和 Client 均能获取到 `skill` 表的代码类型和 JSON 数据

#### Scenario: 目录结构创建

- **GIVEN** 新克隆的项目仓库
- **WHEN** 开发者查看项目根目录
- **THEN** 存在 `configs/Defines/`、`configs/Tables/` 目录及 `configs/luban.conf` 文件

### Requirement: 一键导表脚本

项目 SHALL 提供跨平台导表脚本，开发者运行单个命令即可完成所有配置表的导出。

脚本 SHALL:
- 位于 `scripts/gen-config.ps1`（Windows）和 `scripts/gen-config.sh`（Unix/macOS）
- 检查 dotnet 和 Luban 是否已安装，若未安装则输出明确的安装指引
- 分别生成 Server（Go）和 Client（TypeScript）的代码和数据文件
- 在导表完成后输出成功/失败的明确信息

#### Scenario: Windows 开发者运行导表

- **GIVEN** 开发者已安装 .NET SDK 8.0+ 和 Luban CLI 工具
- **WHEN** 在项目根目录执行 `./scripts/gen-config.ps1`
- **THEN** 导表脚本成功完成，在 `server/config/code/` 生成 Go 代码，`server/data/config/` 生成 JSON 数据，`client/src/config/` 生成 TypeScript 类型，`client/public/data/config/` 生成 JSON 数据

#### Scenario: macOS/Linux 开发者运行导表

- **GIVEN** 开发者已安装 .NET SDK 8.0+ 和 Luban CLI 工具
- **WHEN** 在项目根目录执行 `./scripts/gen-config.sh`
- **THEN** 导表脚本成功完成，输出文件位置与 Windows 脚本一致

#### Scenario: 未安装 dotnet 时给出清晰提示

- **GIVEN** 开发者未安装 .NET SDK
- **WHEN** 运行导表脚本
- **THEN** 脚本输出错误信息，包含 .NET SDK 下载链接和安装指引，以非零退出码退出

#### Scenario: 未安装 Luban 时给出清晰提示

- **GIVEN** 开发者已安装 .NET SDK 但未安装 Luban CLI 工具
- **WHEN** 运行导表脚本
- **THEN** 脚本输出错误信息，包含 `dotnet tool install -g Luban.Tool` 安装命令，以非零退出码退出

### Requirement: Server 端 Go 配置加载

Server 端 SHALL 提供类型安全的 Go 配置加载模块。

该模块 SHALL:
- 使用 Go module path `windypath.com/cs2match/config`
- 使用 Go `embed` 包将导出的 JSON 数据嵌入到编译产物中
- 提供类型化的配置访问接口（与 Luban 生成的 Go 结构体对应）

#### Scenario: Server 加载道具配置

- **GIVEN** 策划在 `item.xlsx` 中定义了 3 个道具，已运行导表脚本
- **WHEN** Go Plugin 启动并初始化配置模块
- **THEN** 配置模块成功加载全部 3 个道具数据，可通过类型安全接口访问

#### Scenario: 配置数据内嵌到编译产物

- **GIVEN** 已运行导表脚本生成 JSON 数据
- **WHEN** 编译 Go Plugin（`build.sh` / `build.ps1`）
- **THEN** 生成的 `.so` 文件包含所有配置 JSON 数据，Nakama 加载插件后配置立即可用，无需额外的数据文件

### Requirement: Client 端 TypeScript 配置加载

Client 端 SHALL 提供类型安全的 TypeScript 配置访问接口。

该模块 SHALL:
- 位于 `client/src/config/`（Luban 生成的 TypeScript 类型）
- 提供 JSON 数据的类型化加载函数
- 支持开发和生产构建两种模式下的配置加载

#### Scenario: 前端加载配置并渲染道具列表

- **GIVEN** 前端应用启动
- **WHEN** 调用配置加载函数获取道具列表
- **THEN** 返回类型为 `Item[]` 的数组，包含全部道具数据，可在 React 组件中渲染

#### Scenario: 类型安全保证

- **GIVEN** `item.xlsx` 定义了字段 `id: int`, `name: string`, `rarity: int`
- **WHEN** 前端开发者使用生成的 TypeScript 类型 `Item`
- **THEN** TypeScript 编译器强制检查字段类型：`item.id` 为 number，`item.name` 为 string，`item.rarity` 为 number

### Requirement: 示例配置表

项目 SHALL 包含至少一张示例配置表（`configs/Tables/item.xlsx`），用于验证导表流程的完整性。

示例表 SHALL:
- 包含至少 3 行数据
- 覆盖常见字段类型：整数、字符串、浮点数
- 在 `configs/Defines/__root__.xml` 中有对应的结构定义

#### Scenario: 示例表导表验证

- **GIVEN** 项目初始状态
- **WHEN** 运行导表脚本
- **THEN** `item` 表在 Server（Go）和 Client（TypeScript）均生成对应的代码和 JSON 数据，JSON 数据包含与 Excel 一致的 3 行记录

### Requirement: 导表产物管理

导表脚本生成的代码和数据文件 SHALL 被 `.gitignore` 排除，不纳入 Git 版本管理。

导表产物 SHALL:
- 在 CI/CD 构建流程中通过脚本自动生成
- 在开发环境中由开发者手动运行导表脚本生成

#### Scenario: Git 不跟踪导表产物

- **GIVEN** 导表脚本已运行，`server/config/`、`server/data/config/`、`client/src/config/`、`client/public/data/config/` 目录存在
- **WHEN** 执行 `git status`
- **THEN** 上述目录不在未跟踪文件列表中（已被 `.gitignore` 忽略）

#### Scenario: CI 构建自动导表

- **GIVEN** GitHub Actions CI 工作流触发
- **WHEN** 执行构建步骤
- **THEN** CI 自动安装 Luban 并运行导表脚本，生成所有配置代码和数据，然后编译 Server 和 Client

