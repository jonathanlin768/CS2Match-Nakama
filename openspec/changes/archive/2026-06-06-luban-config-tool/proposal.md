## Why

游戏策划需要维护几十张 Excel 配置表（道具、角色、技能、地图等），这些表需要转换成前后端可直接使用的代码和数据文件。当前项目缺少配置表管理工具链，策划无法独立维护配表，开发者也缺少标准化的导表流程。引入 Luban 作为配表工具，建立从 Excel → 代码生成 → 前后端同步的完整工作流。

## What Changes

- 在项目根目录新建 `configs/` 目录，存放策划维护的 Excel 配表源文件
- 集成 Luban 导表工具（通过 dotnet 运行）到项目中
- 提供跨平台导表脚本（Windows `.ps1` + Unix `.sh`），一键导出所有配置表
- 导表输出分别放到 `server/config/`（Go 代码 + JSON 数据）和 `client/src/config/`（TypeScript 代码 + JSON 数据）
- Go 后端通过 Go module `windypath.com/cs2match/config` 加载配置数据
- React 前端通过 TypeScript 模块 `@windypath/config` 加载配置数据
- 提供示例 Excel 配置表（如 `item.xlsx`），演示完整的配表→导表→加载流程
- 更新 `.gitignore` 和 Docker 配置，确保导表产出不被提交但随构建生成

## Capabilities

### New Capabilities

- `luban-config-tool`: Luban 配表工具链集成，包括 Excel 源文件管理、导表脚本、前后端代码生成、配置数据加载

### Modified Capabilities

<!-- 本次变更不修改已有 spec，均为新增能力 -->

## Impact

- **Nakama 后端**: 新增 `server/config/` 目录（Go 配置加载模块，module path: `windypath.com/cs2match/config`），不影响现有 RPC/Match Handler
- **React 前端**: 新增 `client/src/config/` 目录（TypeScript 配置类型定义，package: `@windypath/config`），不影响现有组件
- **开发环境**: 新增 `configs/` 策划配表目录、`tools/luban/` 工具目录、`scripts/` 导表脚本目录
- **部署构建**: Docker Compose 和 CI/CD 构建流程需增加导表步骤（不影响运行时，仅构建时执行）
- **不需要新 RPC / Match Handler / Storage 操作**: 配置表通过代码直接内嵌到 Go Plugin 和前端 bundle 中，无需运行时动态加载
- **不需要更新 Luban 配置表**: 本次变更是工具链本身的搭建，不涉及游戏配置内容变更
