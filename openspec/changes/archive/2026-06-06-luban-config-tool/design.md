## Context

当前项目已搭建 Nakama Go Plugin 后端和 React TypeScript 前端骨架，但缺少策划配表工具链。根据项目技术架构文档（`doc/cs2SimuProject.md`），已确定使用 Luban 作为配表工具（决策 #15-16）。策划将维护几十张 Excel 配置表，需要标准化的导表流程生成前后端代码和数据文件。

Luban 是 focus-creative-games 开源的 C# (.NET) 工具，需要 dotnet 运行时。本设计需要解决：如何集成到项目中、目录结构、导表脚本、前后端代码加载方案。

### 约束
- Go module 路径前缀: `windypath.com/cs2match/<module>`
- 前端包名前缀: `@windypath/<package>`
- 配置表不需要热更新（决策 #15），随服务器重启即可
- 开发环境使用 Docker Compose，导表也通过 Docker 容器执行（不要求开发者安装 .NET SDK）

## Goals / Non-Goals

**Goals:**
- 建立统一的策划配表目录结构（`configs/`），策划只需编辑 Excel 文件
- 提供跨平台一键导表脚本，生成 Go 和 TypeScript 代码及 JSON 数据文件
- Server 端能从导出的 JSON 数据加载配置，Go 代码提供类型安全的访问接口
- Client 端能从导出的 JSON 数据加载配置，TypeScript 代码提供类型定义
- 包含一份示例配置表（如 `item.xlsx`），验证完整流程
- 导表产出纳入 `.gitignore`，仅在构建时生成

**Non-Goals:**
- 不实现配置热更新（本版本不需要）
- 不实现配置表在线编辑/管理后台
- 不引入 Luban 的 binary 导出（项目规模下 JSON 足够）
- 不修改现有 Nakama RPC、Match Handler 逻辑

## Decisions

### Decision 1: Luban 运行方式 — Docker 容器

**选择**: 构建一个轻量 Docker 镜像（基于 `mcr.microsoft.com/dotnet/sdk:8.0`），在其中安装 Luban CLI 工具，导表脚本通过 `docker run` 调用。

**Dockerfile**（`tools/luban/Dockerfile`）:
```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0
RUN dotnet tool install -g Luban.Tool
ENV PATH="$PATH:/root/.dotnet/tools"
WORKDIR /workspace
ENTRYPOINT ["luban"]
```

**替代方案**:
- ❌ **dotnet tool 全局安装**: 要求每位开发者安装 .NET SDK 8.0+（~300MB），增加环境搭建成本
- ❌ **预编译二进制放在 `tools/luban/`**: Luban 不发布自包含独立 exe，需要自己编译；自包含产物 ~70MB+ 且平台绑定
- ✅ **Docker 容器**: 项目已依赖 Docker（Docker Compose 开发环境），不增加新依赖；版本通过 Dockerfile 中 Luban 版本号锁定；跨平台一致（Windows/macOS/Linux 均可用）；新成员 clone 后无需额外安装

**理由**: 开发者本地不需要安装 .NET SDK。Docker 镜像构建一次后缓存，导表时直接 `docker run`，约 1-2 秒完成。CI/CD 中也复用同一 Dockerfile。

### Decision 2: 导出格式 — JSON（code + data 双输出模式）

**选择**: Server 和 Client 都使用 Luban 4.x 的 `code + data` 模式，数据格式选择 JSON

- **Server (Go)**: Luban 生成 Go 结构体代码 + JSON 数据文件，模板名称 `go-json`
  - 需通过 `-x go-json.lubanGoModule=windypath.com/cs2match/config` 指定 Go module path
  - 代码输出: `server/config/` (Go package: `config`, module: `windypath.com/cs2match/config`)
  - 数据输出: `server/data/config/` (JSON 文件，Go 编译时通过 `embed` 嵌入)
- **Client (TypeScript)**: Luban 生成 TypeScript 类型定义 + JSON 数据文件，模板名称 `typescript-json`
  - 代码输出: `client/src/config/` (TypeScript 类型)
  - 数据输出: `client/public/data/config/` (JSON 文件，运行时 fetch 或打包时内联)

**替代方案**:
- ❌ **Binary 格式**: 性能优势在本项目规模（几十张表，单表最多几千行）下不显著；JSON 调试方便
- ❌ **仅生成代码不分离数据**: 数据硬编码在代码中，策划修改配置需重新编译；JSON 分离后数据可独立替换

**理由**: JSON 可读性强，便于调试；前后端共用同一份 JSON 数据保证一致性；Luban 生成的代码提供类型安全的访问接口。

### Decision 3: 目录结构设计

采用 Luban 4.x 的 Excel-based schema 格式（`__tables__`、`__beans__`、`__enums__` 均为 Excel 文件）：

```
project/
├── configs/                          # 策划配表源文件（提交到 Git）
│   ├── luban.conf                    # Luban 主配置（JSON 格式，定义 groups/schemaFiles/dataDir/targets）
│   ├── Defines/                      # XML 类型定义（如 vector 等内置类型扩展）
│   └── Datas/                        # 数据目录（dataDir）
│       ├── __tables__.xlsx           # 表定义（table schema）
│       ├── __beans__.xlsx            # 结构定义（bean schema）
│       ├── __enums__.xlsx            # 枚举定义（enum schema）
│       ├── #item.xlsx                # 业务数据表（# 前缀 = dataDir 内文件）
│       ├── #character.xlsx
│       └── ...
├── tools/
│   └── luban/
│       └── Dockerfile                # Luban 运行容器镜像
├── scripts/
│   ├── gen-config.ps1                # Windows 导表脚本（docker run）
│   └── gen-config.sh                 # Unix/macOS 导表脚本（docker run）
├── server/
│   ├── config/                       # Luban 生成的 Go 代码（gitignore）
│   └── data/
│       └── config/                   # Luban 生成的 JSON 数据（gitignore）
└── client/
    ├── src/
    │   └── config/                   # Luban 生成的 TS 类型（gitignore）
    └── public/
        └── data/
            └── config/               # Luban 生成的 JSON 数据（gitignore）
```

> **注意**: Luban 4.x 的表/结构/枚举定义使用 Excel 文件（`__tables__.xlsx` 等），而非旧版的 `__root__.xml`。`Defines/` 目录仅存放少量内置类型扩展的 XML（按需）。

### Decision 4: 导表脚本设计

**选择**: 提供 PowerShell（Windows）和 Bash（Unix/macOS）两套脚本，均通过 `docker run` 调用 Luban 容器。

脚本负责:
1. 检查 Docker 是否可用
2. 构建/确保 Luban Docker 镜像存在（首次运行 `docker build -t luban-runner tools/luban/`）
3. 清理旧的生成输出
4. 通过 `docker run --rm -v` 挂载项目目录到容器，执行 Luban 导出
5. 输出导表结果摘要

```bash
# gen-config.sh 核心逻辑
IMAGE="luban-runner"
docker build -t "$IMAGE" tools/luban/ > /dev/null

# 生成 Server (Go) 代码 + JSON 数据
docker run --rm -v "$(pwd):/workspace" "$IMAGE" \
  -t server -c go-json -d json \
  --conf configs/luban.conf \
  -x outputCodeDir=server/config \
  -x outputDataDir=server/data/config \
  -x go-json.lubanGoModule=windypath.com/cs2match/config

# 生成 Client (TypeScript) 代码 + JSON 数据
docker run --rm -v "$(pwd):/workspace" "$IMAGE" \
  -t client -c typescript-json -d json \
  --conf configs/luban.conf \
  -x outputCodeDir=client/src/config \
  -x outputDataDir=client/public/data/config
```

**理由**: Docker 镜像内已包含 .NET SDK 和 Luban CLI，脚本只需调用容器；镜像构建缓存机制保证首次之后几乎无开销。

### Decision 5: Server 端加载方案 — Go embed

**选择**: 使用 Go 1.16+ 的 `embed` 包将 JSON 数据文件嵌入到编译产物中

```go
//go:embed data/config/*.json
var configData embed.FS
```

**理由**: 配置数据随 .so 文件一起分发，无需运行时文件路径依赖；Nakama 加载 .so 插件时自动可用，不依赖工作目录。

### Decision 6: Client 端加载方案 — 构建时内联

**选择**: Vite 构建时将 `public/data/config/` 下的 JSON 通过动态 import 或 fetch 在运行时加载；开发阶段直接 fetch `public/` 路径的文件。

**理由**: Vite 的 `public/` 目录文件在开发和生产构建中均可直接访问；不需要额外配置。

## Risks / Trade-offs

- **[风险] Luban 版本升级可能改变生成代码格式** → 在 `configs/luban.conf` 中锁定 target 版本；导表脚本可在 CI 中固定 Luban 版本
- **[风险] Docker 镜像首次构建需下载 .NET SDK 基础镜像 (~200MB)** → 一次性成本，后续构建使用 Docker 缓存；团队成员首次 clone 后运行导表脚本会自动构建
- **[权衡] JSON 格式比 Binary 体积大** → 项目规模小（几十张表），JSON 体积影响可忽略；若未来表数据量大可切换到 binary 格式
- **[权衡] 导表与构建流程解耦** → 开发者需手动运行导表脚本后再编译；可后续在 CI/CD 中自动化

## Open Questions

- **Q1**: 是否需要支持 Luban 的 validate 功能（自定义数据校验规则）？→ 暂不需要，后续需要再补
- **Q2**: Luban 的 `luban.conf` 配置文件使用 XML 还是 YAML 格式？→ 使用 Luban 默认的 XML 格式，与 `__root__.xml` 风格一致
