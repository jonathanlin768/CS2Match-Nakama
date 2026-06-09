# Nakama 架构 & 项目自定义代码说明

## 1. Nakama 是什么

Nakama 是一个**开源游戏后端**，你把它跑起来后，它自带：

```
Nakama 裸服务（镜像 pull 下来就有）
├── 用户注册/登录        (Email, OAuth, Steam, 设备ID...)
├── PostgreSQL 数据库     (存玩家数据)
├── WebSocket 实时通信    (客户端长连接)
├── 好友/群组/聊天        (社交功能)
├── 匹配系统              (自动组队/配对)
├── 排行榜/成就           (竞技功能)
└── Console 管理后台       (localhost:7351)
```

但光有这些不够——你的 CS2 对战逻辑（教练选人、地图模拟、回合结算）是**你自己的业务逻辑**，Nakama 不认识。

---

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Compose                        │
│                                                          │
│  ┌──────────┐   ┌──────────────────┐   ┌─────────────┐ │
│  │PostgreSQL │◄──│    Nakama 3.30   │◄──│   React     │ │
│  │  :5432    │   │  :7350 (HTTP/WS) │   │  :3000→:80  │ │
│  │           │   │  :7351 (Console) │   │  (Nginx)    │ │
│  └──────────┘   │  :7349 (gRPC)    │   └─────────────┘ │
│                 │                  │                     │
│                 │  加载你的 .so ──►│                     │
│                 │  ┌────────────┐ │                     │
│                 │  │ backend.so │ │                     │
│                 │  └────────────┘ │                     │
│                 └──────────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

### 端口说明

| 端口 | 协议 | 用途 |
|------|------|------|
| 7350 | HTTP + WebSocket | 客户端 API 调用和长连接 |
| 7349 | gRPC | 内部服务间通信 |
| 7351 | HTTP | Nakama Console 管理后台 |
| 5432 | TCP | PostgreSQL 数据库 |

---

## 3. 你的自定义代码在哪里

Nakama 通过 **Go Plugin（.so 文件）** 来加载你的自定义代码。

所有自定义代码在 `server/` 目录下：

```
server/
├── main.go              ← 自定义插件入口（你写的）
│   ├── InitModule()         Nakama 启动时自动调用
│   │   ├── cfg.Init()                  初始化配置表
│   │   └── RegisterRpc("HealthCheck")  注册 RPC 端点
│   │
│   └── healthCheckRPC()    一个简单的健康检查
│
├── config/               ← Luban 配置表子模块
│   ├── go.mod                module windypath.com/cs2match/config
│   ├── loader.go             从 embed JSON 加载配置数据
│   ├── item.go               手工写的 Item 结构体
│   ├── Tables.go             Luban 自动生成的表结构
│   ├── Tbitem.go             Luban 自动生成的 Item 行
│   └── data/tbitem.json      编译时嵌入的配置数据
│
├── go.mod                ← module windypath.com/cs2match/server
├── build.sh / build.bat / build.ps1
└── build/
    └── backend.so        ← 编译产物（被挂载进 Nakama 容器）
```

### Go Module 结构

项目采用双模块设计：

```
server/                           ← 主模块
├── go.mod    (windypath.com/cs2match/server)
│   └── require github.com/heroiclabs/nakama-common v1.40.0
│   └── require windypath.com/cs2match/config v0.0.0
│   └── replace windypath.com/cs2match/config => ./config   ← 本地引用
│
└── config/                       ← 子模块
    └── go.mod  (windypath.com/cs2match/config)
```

`replace` 指令让主模块引用本地的 config 子模块，不需要发布到远程。

---

## 4. InitModule 入口

Nakama 启动时会扫描 `runtime.path` 目录（`/nakama/data/modules`）下的所有 `.so` 文件，调用里面的 `InitModule()`。

**函数签名必须完全匹配：**

```go
func InitModule(
    ctx context.Context,            // Go 标准 context
    logger runtime.Logger,          // Nakama 日志接口
    db *sql.DB,                     // Nakama 的 PostgreSQL 连接（可直接用）
    nk runtime.NakamaModule,        // Nakama 内置 API
    initializer runtime.Initializer, // 用它注册 RPC / Hooks / Match Handler
) error
```

你的 `InitModule` 目前做了两件事：
1. **初始化配置表** — `cfg.Init()` 从嵌入的 JSON 加载游戏配置
2. **注册 HealthCheck RPC** — 一个简单的健康检查端点，返回 `{"status":"ok"}`

### 完整代码

```go
func InitModule(
    ctx context.Context,
    logger runtime.Logger,
    db *sql.DB,
    nk runtime.NakamaModule,
    initializer runtime.Initializer,
) error {
    logger.Info("CS2Match Go plugin loaded successfully")

    // 初始化配置表
    if err := cfg.Init(); err != nil {
        logger.Error("Failed to init config: %v", err)
        return err
    }

    // 注册 RPC
    if err := initializer.RegisterRpc("HealthCheck", healthCheckRPC); err != nil {
        logger.Error("Failed to register HealthCheck RPC: %v", err)
        return err
    }

    return nil
}
```

---

## 5. 编译流程

### 关键：必须用 Nakama 官方 pluginbuilder 镜像编译

Go plugin 的 `.so` 文件要求**编译时的 Go 版本、依赖版本和运行时的 Nakama 容器完全一致**。不能用本机 Go 编译。

### 编译命令（build.sh 核心代码）

```bash
docker run --rm \
  -v "${SCRIPT_DIR}:/app" \
  -w /app \
  registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0 \
  go build -v -mod=mod -buildmode=plugin -trimpath -o build/backend.so .
```

**流程：**

```
你改代码                          发布上线
   │                                ▲
   ▼                                │
server/main.go              docker compose up
server/config/*.go              │
   │                            ▼
   ▼                  Nakama 容器启动，执行：
bash build.sh             │
   │                     ├─ migrate up（自动建表）
   ▼                     ├─ 加载 backend.so
Docker 拉取                 │   └─ 调用 InitModule()
pluginbuilder:3.30.0       │       ├─ cfg.Init()  ← 配置表就绪
   │                       │       └─ RegisterRpc("HealthCheck", ...)
   ▼                       │
编译 backend.so            ├─ 监听 :7350 等待客户端连接
   │                       │
   ▼                       └─ React 客户端通过
输出到                          @heroiclabs/nakama-js SDK
server/build/backend.so        调用 Nakama API
```

### Docker 如何加载 .so

`docker-compose.yml` 中的关键配置：

```yaml
volumes:
  - ./server/build:/nakama/data/modules:ro  # 把 build 目录挂进容器（只读）

entrypoint:
  - ...
  - --runtime.path /nakama/data/modules     # 告诉 Nakama 去这个目录找 .so
```

`nakama-config.yml` 中的配置：

```yaml
runtime:
  path: "/nakama/data/modules"   # 和 docker-compose 的 volume 挂载路径一致
```

---

## 6. 当前状态一览

| 层 | 当前有什么 | 还没有什么 |
|----|-----------|-----------|
| Nakama 自带 | 注册/登录、数据库、WebSocket、匹配框架 | — |
| 你的 Go 插件 | `HealthCheck` RPC、Luban 配置表加载 | Match Handler、游戏逻辑、Storage 操作 |
| React 前端 | 登录页、5 个页面骨架、UI 组件库 | 真正的 API 调用、WebSocket 实时通信 |
| PostgreSQL | Nakama 自动建表（用户/好友/排行榜等） | 你的自定义业务表 |

---

## 7. 相关文件索引

| 文件 | 作用 |
|------|------|
| `server/main.go` | 插件入口，InitModule |
| `server/go.mod` | 主模块定义 |
| `server/config/` | Luban 配置表子模块 |
| `server/build.sh` | 编译脚本（Linux/Mac/WSL） |
| `server/build.bat` | 编译脚本（Windows） |
| `nakama-config.yml` | Nakama 运行时配置 |
| `docker-compose.yml` | Docker 服务编排 |
| `openspec/specs/` | 各模块规范定义 |
