# CS2 Simu Project - 技术架构清单

> 本文档定义了CS2 Simu项目的完整技术栈、架构选型及部署方案。
> 版本: v0.1 | 日期: 2026-05-29

---

## 一、技术栈总览

| 层级 | 技术选型 | 版本建议 | 说明 |
|------|---------|---------|------|
| **后端游戏服务器** | Nakama | 3.30+ | 开源游戏后端服务器，支持多人联机、匹配、聊天、社交等 |
| **后端运行时语言** | Go | 1.24.5 | 编写Nakama服务器插件（RPC、Match Handler、Hook等） |
| **前端框架** | React | 18+ | 用户界面与交互层 |
| **前端构建工具** | Vite | 5+ | 快速开发与构建，支持现代ESM |
| **Nakama客户端SDK** | @heroiclabs/nakama-js | 最新版 | 官方JavaScript/TypeScript客户端SDK |
| **前端状态管理** | Zustand / React Query | 最新版 | 轻量级状态管理，推荐Zustand（配合Nakama实时数据） |
| **数据库** | PostgreSQL | 15+ | Nakama要求的数据库（兼容Postgres wire protocol） |
| **容器化** | Docker + Docker Compose | 最新版 | 开发环境与生产部署 |
| **策划配表工具** | Luban | 最新版 | 游戏配置表解决方案（Excel→JSON/Binary→代码生成） |
| **AI开发工作流** | OpenSpec | 最新版 | 规范驱动开发(SDD)框架，用于AI辅助编程前的需求对齐与变更管理 |

---

## 二、架构分层详解

### 2.1 后端架构：Nakama + Go Plugin

```
┌─────────────────────────────────────────────────────────────┐
│                      Nakama 游戏服务器                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Authentication │  │  Matchmaking  │  │  Realtime Match  │  │
│  │  (认证模块)      │  │  (匹配模块)    │  │  (实时对战)       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Leaderboard   │  │  Storage      │  │  Chat / Social   │  │
│  │  (排行榜)       │  │  (存储引擎)    │  │  (聊天社交)       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Go Plugin (自定义逻辑层)                  │  │
│  │  • RPC Functions        (HTTP/gRPC调用)               │  │
│  │  • Match Handlers       (权威状态同步)          │  │
│  │  • Before/After Hooks   (请求拦截与扩展)              │  │
│  │  • Scheduler/Timers     (定时任务)                    │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   PostgreSQL 数据库                          │
│  • 用户账户 / 好友关系 / 聊天记录 / 排行榜数据                  │
│  • 游戏存储对象 / 匹配数据 / 运行时配置                        │
└─────────────────────────────────────────────────────────────┘
```

**核心职责划分：**

| 模块 | Nakama内置功能 | Go Plugin自定义开发 |
|------|---------------|-------------------|
| 用户认证 | ✅ Device/Email/Steam等认证 | 自定义登录逻辑、封禁检查 |
| 匹配系统 | ✅ 基础Matchmaker | 自定义匹配规则、房间管理 |
| 实时对战 | ✅ WebSocket传输层 | Match Handler（游戏逻辑权威判定） |
| 排行榜 | ✅ 基础Leaderboard | 赛季逻辑、奖励分发 |
| 数据存储 | ✅ Storage Engine | 业务数据校验、敏感操作保护 |
| 聊天社交 | ✅ Chat/Group/Friends | 敏感词过滤、消息持久化策略 |
| 游戏配置 | - | Luban配置表加载与管理 |

**Go Plugin 开发要点：**
- 使用 `nakama-common` 包开发，编译为 `.so` 动态链接库
- 入口函数 `InitModule` 注册所有RPC、Match Handler、Hooks
- 每个Match Handler包含7个生命周期函数：`MatchInit`, `MatchJoinAttempt`, `MatchJoin`, `MatchLeave`, `MatchLoop`, `MatchTerminate`, `MatchSignal`
- RPC用于客户端请求-响应式调用（如获取房间列表、领取奖励）
- Match Handler用于实时多人游戏逻辑（如CS2模拟器的对战状态同步）

### 2.2 前端架构：React + Nakama JS SDK

```
┌─────────────────────────────────────────────────────────────┐
│                      React 前端应用                          │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  UI Components │  │  Pages / Views │  │  Game Renderer   │  │
│  │  (界面组件)      │  │  (页面路由)    │  │  (游戏渲染器)     │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Nakama Client Layer                      │  │
│  │  ┌──────────────┐  ┌──────────────────────────────┐  │  │
│  │  │ Nakama Client │  │ Nakama Socket (WebSocket)    │  │  │
│  │  │ • authenticate │  │ • joinMatch                  │  │  │
│  │  │ • rpc         │  │ • sendMatchData              │  │  │
│  │  │ • getAccount  │  │ • onMatchData (实时消息监听)  │  │  │
│  │  │ • readStorage │  │ • chat/channel               │  │  │
│  │  └──────────────┘  └──────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              State Management (Zustand)               │  │
│  │  • UserStore      (用户状态)                          │  │
│  │  • MatchStore     (对战状态)                          │  │
│  │  • GameStore      (游戏内状态)                        │  │
│  │  • UIStore        (界面状态)                          │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**前端技术选型说明：**

| 技术 | 选型建议 | 理由 |
|------|---------|------|
| 构建工具 | **Vite** | 比CRA更快，支持现代ESM，配置简洁 |
| 语言 | **TypeScript** | 类型安全，配合Nakama SDK有更好的IDE体验 |
| 样式方案 | **Tailwind CSS** 或 **CSS Modules** | 快速开发，避免样式冲突 |
| 路由 | **React Router v6** | 标准方案 |
| 状态管理 | **Zustand** | 比Redux轻量，适合游戏状态；配合不可变更新 |
| HTTP客户端 | **Nakama SDK内置** | 无需额外axios，SDK已封装REST/gRPC调用 |
| 实时通信 | **Nakama Socket** | 官方WebSocket封装，支持自动重连 |

**Nakama 前端连接规范：**

```typescript
// 1. 创建Client（单例，每个服务器一个）
const client = new Client("defaultkey", "nakama.example.com", "7350", true);

// 2. 认证获取Session
const session = await client.authenticateDevice(deviceId, true, "username");
// 或邮箱认证：await client.authenticateEmail(email, password, true);

// 3. 创建Socket连接（实时功能必需）
const socket = client.createSocket();
await socket.connect(session, true); // appearOnline = true

// 4. 监听实时消息
socket.onmatchdata = (matchData) => {
  // 处理服务器推送的游戏状态
};

// 5. 加入Match
const match = await socket.joinMatch(matchId);

// 6. 发送操作
await socket.sendMatchState(matchId, opcode, JSON.stringify(payload));
```

### 2.3 前后端分离部署架构

```
                              玩家浏览器
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                        CDN / Nginx                          │
│              (静态资源托管 / HTTPS / 反向代理)                 │
└─────────────────────────────────────────────────────────────┘
                                  │
              ┌───────────────────┴───────────────────┐
              │                                       │
              ▼                                       ▼
┌─────────────────────────┐               ┌─────────────────────────┐
│      React 前端应用       │               │      Nakama 服务器       │
│   (Docker / Vercel等)   │◄─────────────►│   (Docker / 云服务器)    │
│                         │   WebSocket   │   • API Port: 7350      │
│                         │   REST API    │   • Console: 7351       │
│                         │               │   • gRPC: 7349          │
└─────────────────────────┘               └─────────────────────────┘
                                                    │
                                                    ▼
                                          ┌─────────────────────────┐
                                          │     PostgreSQL 数据库    │
                                          │  (Docker / 托管数据库)   │
                                          └─────────────────────────┘
```

**部署要点：**
- 前端通过Nakama JavaScript SDK直接与Nakama服务器通信（**无需中间Node.js BFF层**）
- 前端静态资源可部署到Vercel/Netlify/自有CDN
- Nakama服务器需要暴露：
  - `7350` - HTTP REST API / WebSocket（前端主要通信端口）
  - `7349` - gRPC（可选，前端一般不用）
  - `7351` - Web Console（管理后台，生产环境需限制访问）
- 数据库建议使用托管PostgreSQL（如AWS RDS、阿里云RDS），或Docker部署
- 生产环境Nakama与数据库应部署在同一私有网络，不直接暴露数据库端口

### 2.4 数据库架构：PostgreSQL

**Nakama数据库设计特点：**
- Nakama自行管理所有表结构，通过 `nakama migrate up` 执行迁移
- **不要手动修改Nakama核心表**，只通过Nakama SDK/Runtime操作数据
- 如需自定义业务表，建议：
  - 方式1：使用Nakama Storage Engine（推荐，自动处理权限、缓存）
  - 方式2：在Go Plugin中直接操作同一PostgreSQL实例的不同Schema

**配置建议：**
```yaml
# docker-compose.yml 关键配置
db:
  image: postgres:15-alpine
  environment:
    POSTGRES_DB: nakama
    POSTGRES_PASSWORD: your_secure_password
  volumes:
    - postgres-data:/var/lib/postgresql/data

nakama:
  image: registry.heroiclabs.com/heroiclabs/nakama:3.30.0
  command: >
    sh -cx "nakama migrate up --database.address postgres://nakama:password@db:5432/nakama &&
            exec nakama --database.address postgres://nakama:password@db:5432/nakama"
  ports:
    - "7350:7350"
    - "7351:7351"
  depends_on:
    - db
```

### 2.5 策划配表模块：Luban

**工作流程：**

```
策划Excel配表 ──► Luban导表工具 ──► 生成Go代码 + JSON/Binary数据 ──► 前端/后端加载
```

**项目集成方案：**

| 目标 | 输出格式 | 使用场景 |
|------|---------|---------|
| Nakama后端 (Go) | Go代码 + JSON | 游戏配置热加载、RPC返回配置数据 |
| React前端 | TypeScript代码 + JSON | 客户端配置预览、本地化文本 |

**目录结构建议：**
```
project/
├── configs/                    # 策划配表源文件 (Excel/JSON)
│   ├── __tables__.xlsx         # 表定义
│   ├── datas/                  # 数据目录
│   │   ├── item.xlsx
│   │   ├── character.xlsx
│   │   └── skill.xlsx
│   └── defines/                # 结构定义
│       └── __root__.xml
├── tools/
│   └── luban/                  # Luban工具链
├── server/
│   └── data/                   # 导出到后端的配置数据
└── client/public/data/         # 导出到前端的配置数据
```

**Luban导表示例：**
```bash
# 生成后端Go代码 + JSON数据
dotnet Luban.dll \
  -t server \
  -c go-json \
  -d json \
  --schemaPath configs/defines/__root__.xml \
  -x inputDataDir=configs/datas \
  -x outputCodeDir=server/config/ \
  -x outputDataDir=server/data/

# 生成前端TypeScript代码 + JSON数据
dotnet Luban.dll \
  -t client \
  -c ts-json \
  -d json \
  --schemaPath configs/defines/__root__.xml \
  -x inputDataDir=configs/datas \
  -x outputCodeDir=client/src/config/ \
  -x outputDataDir=client/public/data/
```

### 2.6 AI开发工作流：OpenSpec

**OpenSpec** 是由 Fission-AI 开源的规范驱动开发（Spec-driven Development, SDD）框架。它在项目中的定位是**开发流程管理工具**，而非运行时技术组件。

**核心用途：**
- 在AI辅助编程（如使用Cursor、GitHub Copilot、Claude Code等）时，建立结构化的需求规格层
- 通过 `/opsx:propose`、`/opsx:apply`、`/opsx:archive` 等工作流命令，管理功能提案、技术设计和实现任务
- 每个变更对应一个独立目录（`openspec/changes/<feature-name>/`），包含 `proposal.md`、`design.md`、`specs/`、`tasks.md`
- 避免需求散落在聊天历史中，确保人类开发者与AI助手在编码前对齐意图

**在项目中的使用场景：**

| 场景 | 示例命令 | 产出 |
|------|---------|------|
| 提案新功能 | `/opsx:propose "实现5v5匹配系统"` | proposal.md + design.md + tasks.md |
| 继续实现 | `/opsx:apply` | 按tasks.md逐项完成代码 |
| 归档完成的功能 | `/opsx:archive` | 归档到 `openspec/changes/archive/` |
| 管理多个并行变更 | `openspec dashboard` | 查看进行中的变更列表 |

**技术栈定位说明：**

OpenSpec **不属于项目运行时技术栈**，它是团队协作和AI辅助开发的**工作流工具**。类比来说：
- 就像 Git 是版本控制工具、Notion 是文档协作工具一样
- OpenSpec 是"AI时代的需求规格管理工具"
- 它不直接参与游戏服务器的运行、不部署到生产环境、不引入运行时依赖

**项目集成方式：**
```bash
# 全局安装
npm install -g @fission-ai/openspec@latest

# 在项目根目录初始化
cd cs2-simu-project
openspec init

# 根据团队习惯选择工作流配置
openspec config profile
```

**与项目CI/CD的关系：**
- OpenSpec 生成的规格文件（`.md`）应提交到代码仓库，作为活文档
- 可在 CI 中增加轻量检查：确保涉及核心模块的PR关联了对应的OpenSpec规格目录
- 不参与编译、打包、部署流程

**注意：** 如果项目需要**游戏内Bot/NPC AI运行时**（如CS2模拟器的电脑玩家行为决策），这属于另一个独立的技术领域，需要专门的行为树、状态机或决策框架来实现，与OpenSpec无关。

---

## 三、Nakama 最佳实践 vs 当前架构对比

### 3.1 Nakama 官方推荐架构特点

基于Nakama官方文档和社区实践，其推荐架构模式如下：

| 方面 | Nakama最佳实践 | 本架构 | 评估 |
|------|---------------|--------|------|
| **数据库** | **CockroachDB**（官方首选，分布式强一致） | PostgreSQL | ⚠️ PostgreSQL兼容但非最优，CockroachDB在分布式/高可用场景更有优势 |
| **前后端通信** | 官方SDK直连，REST认证 + WebSocket实时 | 一致 | ✅ 符合最佳实践 |
| **游戏逻辑位置** | **Server Authoritative**（服务器权威），客户端只发输入 | 计划使用 | ✅ 正确方向，防止作弊 |
| **部署方式** | Docker Compose（开发）/ Kubernetes（生产集群） | Docker Compose | ⚠️ 需规划生产K8s方案 |
| **服务器运行时** | Go Plugin（.so）或 Lua / TypeScript | Go Plugin | ✅ 推荐Go Plugin，性能最好 |
| **静态资源** | Nakama内置Web服务器可托管，但前后端分离也是可行方案 | 前后端分离 | ✅ 现代Web部署标准 |
| **多节点扩展** | Nakama Enterprise支持集群（$600+/月），开源版支持单节点 | 单节点起步 | ⚠️ 需评估未来是否需要Enterprise |

### 3.2 关键差异与建议

#### 🔴 差异1：PostgreSQL vs CockroachDB

**Nakama官方推荐CockroachDB的原因：**
- CockroachDB是分布式SQL数据库，原生支持水平扩展
- Nakama的某些功能（如分布式锁、强一致性存储）在CockroachDB上表现更好
- 官方文档和测试以CockroachDB为主

**PostgreSQL的适用场景：**
- 项目初期/中小型规模（<10万DAU）完全够用
- 团队更熟悉PostgreSQL运维
- 托管PostgreSQL服务（RDS等）更易获取

**建议：**
- **开发/测试阶段**：使用PostgreSQL，更简单
- **生产评估**：如果预计DAU > 50万或需要多区域部署，考虑迁移到CockroachDB
- **迁移成本**：Nakama使用标准SQL，两者迁移成本不高

#### 🟡 差异2：前后端分离的部署细节

**当前架构优点：**
- 前端可独立部署到CDN，加载速度快
- 前后端可以独立扩展
- 开发团队可并行工作

**需要注意的风险：**
- **CORS配置**：Nakama需要正确配置允许的前端域名
- **WebSocket跨域**：确保Nakama的WebSocket端点允许跨域连接
- **SSL/TLS**：生产环境必须全链路HTTPS/WSS，Nakama支持TLS证书配置

**Nakama配置建议：**
```yaml
# nakama.yml
socket:
  server_key: "defaultkey"
  
console:
  username: "admin"
  password: "secure_password"
  
# 生产环境关键配置
runtime:
  env:
    - "CORS_ORIGIN=https://your-frontend-domain.com"
```

#### 🟢 差异3：Go Runtime vs TypeScript/Lua Runtime

**选择Go的优势（与最佳实践一致）：**
- **性能**：Go编译为原生代码，比Lua/TS解释执行快10-100倍
- **生态**：可使用完整的Go生态（如AI推理库、数值计算等）
- **类型安全**：编译期检查，减少线上错误

**选择Go的注意事项：**
- **热更新困难**：Go插件需要重新编译并重启Nakama（Lua/TS支持热重载）
- **编译流程**：需要设置Go编译环境，CI/CD中增加编译步骤
- **调试复杂度**：不如TS/Lua直观

**建议：**
- 保持Go方案，但需要建立完善的CI/CD自动编译流程
- 开发环境可考虑用TypeScript快速原型验证，再迁移到Go

### 3.3 Nakama 核心最佳实践清单

以下实践强烈建议在本项目中采用：

#### ✅ 认证与Session管理
```typescript
// 前端：保存session，自动刷新
const session = await client.authenticateDevice(deviceId, true);
localStorage.setItem("nakama_token", session.token);
localStorage.setItem("nakama_refresh", session.refresh_token);

// 启动时恢复session
const authToken = localStorage.getItem("nakama_token");
const refreshToken = localStorage.getItem("nakama_refresh");
if (authToken && refreshToken) {
  const session = Session.restore(authToken, refreshToken);
  if (session.isexpired || session.isexpired(Date.now() + 300000)) {
    session = await client.sessionRefresh(session);
  }
}
```

#### ✅ 服务器权威模式（Server Authoritative）
对于CS2模拟器这类竞技游戏，**必须**采用服务器权威：
- 所有游戏状态变更由Match Handler的`MatchLoop`驱动
- 客户端只发送"输入"（如移动指令、射击指令）
- 服务器校验输入合法性，更新状态后广播给所有客户端
- 客户端做预测+回滚（Client-side Prediction）来隐藏延迟

#### ✅ 使用Storage Engine存储玩家数据
```go
// Go Plugin中权威写入
data := map[string]interface{}{"kills": 100, "deaths": 50}
jsonData, _ := json.Marshal(data)

nk.StorageWrite(ctx, []*runtime.StorageWrite{
    {
        Collection: "player_stats",
        Key:        "season_1",
        UserID:     userId,
        Value:      string(jsonData),
        PermissionRead:  1, // 仅自己和服务器可读
        PermissionWrite: 0, // 仅服务器可写（客户端无法篡改）
    },
})
```

#### ✅ 使用RPC替代直接数据库操作
所有客户端获取数据、触发逻辑的操作，应通过RPC：
- 可在RPC中加入业务校验（如防刷、权限检查）
- 便于后续版本兼容（不改客户端代码，只改服务端RPC实现）

---

## 四、已确认的关键决策

以下问题已与项目决策者确认，答案及影响分析如下：

### 4.1 OpenSpec工作流

| # | 问题 | 决策 | 影响 |
|---|------|------|------|
| 1 | OpenSpec的工作流集成深度？ | **仅个人使用**，为了加速开发 | 不需要团队规范约束，个人在本地使用 `openspec` CLI管理AI辅助编程流程即可 |
| 2 | 是否需要定制化模板？ | **暂时不需要**，后续需要再补 | 使用OpenSpec默认工作流即可，降低上手成本 |

### 4.2 游戏玩法与同步方案 ⭐核心决策

| # | 问题 | 决策 | 影响 |
|---|------|------|------|
| 3 | 核心玩法是什么？ | **1v1教练对战**：两名玩家扮演教练，各自挑选5名选手，在地图中模拟比赛；局中可释放技能/战术 | 服务器需要驱动10个单位（2×5选手）的AI行为 + 处理玩家战术指令 |
| 4 | 单局真实玩家人数？ | **2人**（1v1教练） | Match Handler设计简化，单局仅需维护2个真实玩家的连接状态 |
| 5 | 延迟策略？ | **完全服务器驱动**，客户端听服务器下发的通知（ntf）变更状态 | 采用状态同步方案，不需要客户端预测和回滚（见5.1节） |
| 6 | 观战/回放系统？ | **需要** | 必须在Match Handler中设计状态快照序列化和回放数据存储机制 |

### 4.3 技术栈细节

| # | 问题 | 决策 | 影响 |
|---|------|------|------|
| 7 | 前端框架？ | **必须用React**（AI大模型写React能力强） | 保持React方案，使用Canvas或SVG实现2D战术地图渲染 |
| 8 | 数据库？ | **PostgreSQL** | 不需要评估CockroachDB，简化运维 |
| 9 | 多端支持？ | **仅响应式Web端**，支持PC和移动端的浏览器访问 | Nakama JS SDK完全满足，前端使用响应式布局适配不同屏幕 |
| 10 | 后端运行时语言？ | **Go** | 使用Go Plugin开发Nakama服务器逻辑，不备选TypeScript/Lua |

### 4.4 部署与运维

| # | 问题 | 决策 | 影响 |
|---|------|------|------|
| 11 | 目标部署环境？ | **AWS** | 后续部署脚本和CI/CD基于AWS设计（如EC2 + RDS/自建PostgreSQL） |
| 12 | 同时在线规模？ | **最多100人**（实验项目） | **单节点Nakama完全够用**，不需要Enterprise，不需要K8s集群 |
| 13 | 全球化部署？ | **不需要** | 单区域部署即可，数据库和服务器同区域 |
| 14 | 运维人员？ | **开发者兼任运维** | 部署方案必须极简，优先使用Docker Compose + 托管服务，避免复杂的K8s运维 |

### 4.5 策划配表

| # | 问题 | 决策 | 影响 |
|---|------|------|------|
| 15 | 配置表热更新？ | **本版本不需要** | 简化设计，Luban导表后随服务器重启加载即可 |
| 16 | 配置表数量？ | **几十张** | Luban导表性能无压力，使用常规管理方案 |

---

## 五、基于确认需求的架构细化

### 5.1 同步方案：服务器权威状态同步

经确认，项目采用 **服务器权威的固定时间步状态同步（Server-Authoritative Fixed Timestep State Sync）**，而非帧同步（Deterministic Lockstep）。

**两种方案的本质区别：**

| 维度 | 传统帧同步 (Lockstep) | 服务器权威状态同步 (State Sync) |
|------|----------------------|-------------------------------|
| 逻辑运行 | 所有客户端+服务器各自运行相同游戏逻辑 | **仅服务器运行游戏逻辑** |
| 同步内容 | 只同步玩家**输入**（按键、指令） | 同步**状态快照**（所有单位的位置、血量等） |
| 客户端职责 | 客户端也要计算游戏状态 | **客户端只做表现（渲染+插值）** |
| 确定性要求 | **必须强确定性**（浮点数、随机数需一致） | 不需要确定性 |
| AI单位处理 | 每个客户端各自计算AI → 容易不一致 | **服务器统一计算AI行为** → 天然一致 |
| 观战/回放 | 需要录下所有输入，重新模拟一遍 | **直接录状态快照序列**，回放简单可靠 |
| 网络要求 | 受最慢玩家拖累，需要等待所有玩家输入 | 不受单个玩家影响，服务器自主推进 |
| 适用场景 | RTS（星际争霸）、格斗游戏 | MOBA、战术竞技、本项目的教练模拟 |

**为什么采用状态同步：**

1. **10个AI单位由服务器驱动**：如果帧同步，每个客户端都要跑10个AI的决策逻辑，任何一点随机数/路径计算差异都会导致状态分歧（desync），调试极其痛苦。状态同步下，AI只在服务器跑，客户端只负责"看到什么画什么"。
2. **观战/回放需求**：状态同步天然支持——把每帧状态快照存下来，回放时按时间戳重放即可。帧同步的回放需要保存整场对局的所有输入并按相同逻辑重新模拟，一旦逻辑版本变化就无法回放旧对局。
3. **"一切听服务器下发"**：用户明确表示客户端不预测、不听本地，这正是状态同步的核心特征。帧同步的客户端是有本地逻辑的。
4. **只有2个真实玩家**：状态同步每tick广播状态快照给2人+若干观战者，带宽完全够用。帧同步的优势（省带宽）在本项目不明显。

**状态同步架构：**

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Nakama Match Handler                         │
│                                                                     │
│  Tick Rate: 10Hz (100ms/tick) 或 20Hz (50ms/tick)                  │
│                                                                     │
│  每 Tick 执行：                                                      │
│  1. MatchLoop 被调用                                                 │
│  2. 处理两名教练玩家的战术指令（技能释放、阵型调整等）                    │
│  3. 更新10名选手的AI状态（移动、瞄准、射击等）                          │
│  4. 检测碰撞/伤害/事件                                                │
│  5. 生成 StateSnapshot（状态快照）                                    │
│  6. 广播 StateSnapshot 给所有参与者（2名玩家 + N名观战者）              │
│                                                                     │
│  StateSnapshot 结构示例：                                            │
│  {                                                                  │
│    tick: 1234,                                                      │
│    timestamp: 1716988800000,                                        │
│    players: [                                                       │
│      { id: "p1", pos: {x, y}, hp: 100, action: "moving" },          │
│      { id: "p2", pos: {x, y}, hp: 80,  action: "shooting" },        │
│      ...                                                            │
│    ],                                                               │
│    events: ["p3_killed_p5", "bomb_planted"],                        │
│    coach_commands: [{coachId, commandId, targetId}]                 │
│  }                                                                  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼ WebSocket Broadcast
┌─────────────────────────────────────────────────────────────────────┐
│                           React 客户端                                │
│                                                                     │
│  收到 StateSnapshot 后：                                              │
│  1. 将新状态写入 GameStore（Zustand）                                 │
│  2. React组件监听状态变化，重新渲染Canvas/SVG                          │
│  3. 使用插值（Lerp）让单位移动更平滑（如果tick rate是10Hz）            │
│  4. 客户端不运行任何游戏逻辑，只做纯表现                                 │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 观战与回放系统设计

**观战系统（实时）：**
- 观战者通过Nakama Socket以**只读身份**加入正在进行的Match
- 观战者不发送任何指令，只接收与其他玩家相同的`StateSnapshot`广播
- Match Handler中区分`Player`和`Spectator`角色， spectators不纳入胜负判定
- 可在Match创建时或进行中通过RPC申请观战席位

**回放系统（历史）：**
- 对局开始时，Match Handler初始化一个`ReplayRecorder`
- 每tick的`StateSnapshot`被追加到内存中的`replayBuffer`
- 对局结束（或每N tick）时，将整个`replayBuffer`序列化为JSON/Binary，通过`nk.StorageWrite`存入Nakama Storage Engine，或使用`nk.SqlExec`写入自定义表
- 回放元数据（对局ID、时间、两名教练、比分、时长）单独存储便于检索
- 回放播放时，客户端按固定时间间隔（如10Hz）顺序读取状态快照并渲染，支持暂停、快进、慢放、跳转

**回放数据存储建议：**

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| A. Nakama Storage Engine | 与现有基础设施一致，自带权限控制 | 单对象大小有限制（约1-2MB），超长对局可能需要分片 | ⭐⭐⭐ 推荐 |
| B. PostgreSQL自定义表（Bytea/JSONB） | 可存储大体积数据，方便查询元数据 | 需要绕过Nakama直接操作DB，维护边界模糊 | ⭐⭐ 备选 |
| C. AWS S3 + 预签名URL | 无限容量，成本低 | 需要额外AWS SDK集成，架构更复杂 | ⭐ 不推荐（实验项目过度设计） |

对于100人在线的实验项目，**方案A（Storage Engine）**足够。一场10分钟的对局，10Hz采样，每帧状态约5KB，整场约3MB，分2-3个storage object存储即可。

### 5.3 部署架构（针对实验项目简化版）

基于"100人同时在线、AWS、开发者兼运维、单区域"的决策，推荐以下极简部署方案：

```
┌─────────────────────────────────────────────────────────────┐
│                         AWS 单区域                            │
│                                                             │
│  ┌─────────────────────┐    ┌─────────────────────────────┐│
│  │   EC2 (t3.medium)   │    │      RDS PostgreSQL         ││
│  │   • Docker Engine   │◄──►│      (db.t3.micro)          ││
│  │   • Docker Compose  │    │      • 同VPC内网连接         ││
│  │                     │    │      • 自动备份              ││
│  │  ┌───────────────┐  │    └─────────────────────────────┘│
│  │  │  Nakama容器   │  │                                    │
│  │  │  - Port 7350  │  │◄─── 玩家WebSocket/REST连接        │
│  │  │  - Port 7351  │  │      (Nginx反向代理 + SSL)         │
│  │  └───────────────┘  │                                    │
│  │                     │                                    │
│  │  ┌───────────────┐  │                                    │
│  │  │  Nginx容器    │  │◄─── 静态资源托管（可选）            │
│  │  │  - SSL终止    │  │      或前端部署到Vercel/Netlify     │
│  │  │  - 反向代理   │  │                                    │
│  │  └───────────────┘  │                                    │
│  └─────────────────────┘                                    │
└─────────────────────────────────────────────────────────────┘
```

**为什么这样设计：**
- **单台EC2 + RDS**：100人在线，Nakama单节点轻松支撑（官方数据显示单节点可支撑数万CCU）
- **不需要K8s**：减少运维复杂度，开发者用Docker Compose管理即可
- **RDS托管数据库**：自动备份、补丁管理，开发者不需要自己维护PostgreSQL
- **Nginx做SSL终止**：Nakama本身支持TLS，但用Nginx统一管理证书更方便
- **前端可独立部署**：React静态资源可以放到Vercel/Netlify（免费），通过CDN加速，只把API指向AWS上的Nakama

**月成本预估（AWS us-east-1）：**
- EC2 t3.medium（2vCPU, 4GB）：~$30/月
- RDS db.t3.micro：~$15/月
- 数据传输：~$5/月
- **总计：~$50/月**

---

## 六、推荐后续步骤

1. ✅ **关键决策已确认**（同步方案、部署规模、技术栈）
2. **搭建开发环境**：基于Docker Compose的Nakama + PostgreSQL + React最小可运行原型
3. **技术验证**：
   - 实现一个简单RPC（如"HelloWorld"）验证前后端连通
   - 实现一个**固定时间步的Match Handler**（每1秒广播一次状态），验证状态同步机制
   - 验证Luban导表流程，确认前后端配置加载
4. **核心原型开发**：
   - 实现10个选手在地图上的基础移动（服务器驱动）
   - 实现教练释放一个简单战术指令（如"全员前压"）
   - 实现观战者加入并接收状态广播
5. **架构决策记录**：将关键决策（状态同步方案、PostgreSQL、AWS单节点部署）记录为ADR
6. **CI/CD搭建**：GitHub Actions → 自动编译Go Plugin → Docker Compose部署到AWS EC2

---

## 六、参考资源

| 资源 | 链接 |
|------|------|
| Nakama 官方文档 | https://heroiclabs.com/docs/nakama/ |
| Nakama JavaScript SDK | https://heroiclabs.com/docs/nakama/client-libraries/javascript/ |
| Nakama Go Runtime | https://heroiclabs.com/docs/nakama/server-framework/go-runtime/ |
| Nakama GitHub | https://github.com/heroiclabs/nakama |
| Luban 官方文档 | https://luban.doc.code-philosophy.com/ |
| OpenSpec GitHub | https://github.com/Fission-AI/OpenSpec |
| React + Nakama 示例 | https://github.com/UnfazedHope/tic-tac-toe-multiplayer |
| 狼人杀 (Nakama + React) | https://github.com/majiayu000/werewolf-nakama |
