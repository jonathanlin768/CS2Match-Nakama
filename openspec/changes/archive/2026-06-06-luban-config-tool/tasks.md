## 1. Docker 基础设施

- [x] 1.1 创建 `tools/luban/Dockerfile`，基于 `mcr.microsoft.com/dotnet/sdk:8.0` 安装 Luban CLI 工具
- [x] 1.2 验证 Docker 镜像构建成功：`docker build -t luban-runner tools/luban/`

## 2. 基础目录结构搭建

- [x] 2.1 创建 `configs/` 目录及其子目录 (`configs/Defines/`, `configs/Datas/`)
- [x] 2.2 编写 `configs/luban.conf` Luban 4.x 主配置（JSON 格式：groups/schemaFiles/dataDir/targets），定义 Server（go-json）和 Client（typescript-json）两个导出目标
- [x] 2.3 更新项目根目录 `.gitignore`，排除导表产物目录 (`server/config/`, `server/data/config/`, `client/src/config/`, `client/public/data/config/`)

## 3. 示例配置表（Luban 4.x Excel Schema 格式）

- [x] 3.1 创建 `configs/Datas/__tables__.xlsx`（表定义）、`configs/Datas/__beans__.xlsx`（结构定义）、`configs/Datas/__enums__.xlsx`（枚举定义），定义 Item 结构（id:int, name:string, rarity:int）和 TbItem 表
- [x] 3.2 创建 `configs/Datas/#item.xlsx`（业务数据表），包含至少 3 条示例道具数据，使用 Luban 4.x 三行表头格式（##var / ##type / ##）
- [x] 3.3 通过 Docker 运行 Luban 命令验证导表流程：确认 Go 代码和 TS 类型均生成成功

## 4. 导表脚本

- [x] 4.1 创建 `scripts/gen-config.sh` (Unix/macOS)，包含 Docker 可用性检查、Luban 镜像构建、清理旧输出、生成 Server 和 Client 配置
- [x] 4.2 创建 `scripts/gen-config.ps1` (Windows)，功能与 `.sh` 版本一致
- [ ] 4.3 在 Windows 环境下测试脚本可执行且输出正确

## 5. Server 端 Go 配置加载模块

- [x] 5.1 创建 `server/config/go.mod`，module path 为 `windypath.com/cs2match/config`，Go 版本 1.24.5
- [x] 5.2 创建 `server/config/loader.go`，使用 Go `embed` 包加载 `server/config/data/` 下的 JSON 数据文件
- [x] 5.3 编写配置初始化函数，解析 JSON 到 Luban 生成的 Go 结构体中，暴露类型安全的访问接口
- [x] 5.4 在 `server/main.go` 的 `InitModule` 中调用配置初始化，日志输出加载的表数量，并示例性地打印一两个道具的详细信息（如 id、name、desc、count）
- [x] 5.5 编译 Go Plugin（`build.sh`），验证 `.so` 文件生成成功且包含配置数据

## 6. Client 端 TypeScript 配置加载

- [x] 6.1 创建 `client/src/config/index.ts`，导出 Luban 生成的 TypeScript 类型和 JSON 数据加载函数
- [x] 6.2 实现配置加载函数：开发环境从 `/data/config/` fetch JSON
- [x] 6.3 在 `client/src/App.tsx` 中添加配置加载示例代码（如加载道具列表并在控制台输出），验证前后端链路

## 7. 集成验证与文档

- [x] 7.1 更新 `docker-compose.yml`（无需改动，现有挂载和构建流程已兼容）
- [x] 7.2 更新 `server/build.sh`，在编译 Go Plugin 前自动检测并调用导表脚本
- [x] 7.3 更新 `README.md`，添加 Luban 导表说明（Docker 运行方式、导表命令、目录结构说明），明确标注开发者无需安装 .NET SDK
- [x] 7.4 端到端验证：Go Plugin 编译成功（.so 包含配置数据），前端可从 /data/config/ 加载 JSON
