# 项目结构
```plaintext
esports-manager-go/
├── main.go                     # Nakama Module 的唯一入口
├── internal/                   # 内部代码，防止被其他 Go 项目错误引用
│   ├── user/               # 👤 子系统 1：玩家信息与账号
│   │   ├── api_rpc.go          # RPC 接口 (如：获取玩家资料)
│   │   ├── service.go          # 玩家信息查询与初始化
│   │   ├── repository.go       # 玩家数据 Storage 读写
│   │   └── model.go            # 玩家相关数据结构
│   │
│   ├── club/               # 🏀 子系统 2：俱乐部与阵容
│   │   ├── api_rpc.go          # RPC 接口 (如：选手上下阵、管理阵容)
│   │   ├── service.go          # 俱乐部核心逻辑 (如：默认阵容、选手管理)
│   │   ├── repository.go       # 阵容数据 Storage 读写
│   │   └── model.go            # 俱乐部相关数据结构
│   │
│   ├── player/             # 🧑‍🤝‍🧑 子系统 3：选手图鉴/数据库（可选）
│   │   ├── api_rpc.go          # RPC 接口 (如：查看选手详情)
│   │   ├── service.go          # 选手图鉴逻辑
│   │   ├── repository.go       # 选手数据读写
│   │   └── model.go
│   │
│   ├── gacha/              # 🎲 子系统 4：抽卡
│   │   ├── api_rpc.go          # RPC 接口 (如：抽卡)
│   │   ├── service.go          # 抽卡逻辑、概率、保底
│   │   ├── repository.go       # 抽卡记录读写
│   │   └── model.go
│   │
│   ├── shop/               # 🏪 子系统 5：商店
│   │   ├── api_rpc.go          # RPC 接口 (如：购买商品)
│   │   ├── service.go          # 商店逻辑、限购、折扣
│   │   ├── repository.go       # 购买记录读写
│   │   └── model.go
│   │
│   ├── quest/              # 📜 子系统 6：任务与成就
│   │   ├── api_rpc.go          # RPC 接口 (如：领取任务奖励)
│   │   ├── service.go          # 任务进度与领奖逻辑
│   │   ├── config.go           # 任务配置加载
│   │   ├── scheduler.go        # 重置/过期检查
│   │   ├── reward.go           # 奖励发放
│   │   ├── repository.go       # 任务进度 Storage 读写
│   │   └── model.go            # 任务相关数据结构
│   │
│   ├── activity/           # 🎉 子系统 7：限时/节日活动 orchestration
│   │   ├── api_rpc.go          # RPC 接口 (如：获取活动详情)
│   │   ├── service.go          # 活动包装与生命周期编排
│   │   ├── repository.go       # 活动参与状态读写
│   │   └── model.go
│   │
│   ├── onboarding/         # 🎁 子系统 8：新手流程编排（可选）
│   │   ├── api_rpc.go          # RPC 接口 (如：初始化新玩家)
│   │   ├── service.go          # 新手奖励/任务/阵容初始化编排
│   │   └── model.go
│   │
│   ├── match/              # ⚔️ 子系统 9：匹配对战
│   │   ├── api_rpc.go          # RPC 接口：开始人机对战、PVP匹配、新手引导比赛等
│   │   ├── service.go          # 比赛业务编排：查阵容、构造Bot、调用matchengine、保存记录、触发事件
│   │   ├── bot.go              # Bot队伍解析与构造
│   │   ├── repository.go       # 对战记录 Storage 读写
│   │   └── model.go            # 比赛领域模型、RPC DTO、Storage结构
│   │
│   ├── ranking/            # 🏆 子系统 10：排行榜与段位
│   │   ├── api_rpc.go          # RPC 接口 (如：获取排行榜)
│   │   ├── service.go          # 排行榜与段位逻辑
│   │   └── model.go
│   │
│   ├── framework/                # 🔧 框架能力层
│   │   ├── questkit/             # 任务条件引擎
│   │   │   ├── dispatcher.go     # 事件分发
│   │   │   ├── condition.go      # 条件接口
│   │   │   ├── registry.go       # 条件注册表
│   │   │   ├── tracker.go        # 进度追踪
│   │   │   └── events.go         # 标准事件定义
│   │   │
│   │   ├── activitykit/          # 活动生命周期框架
│   │   │   ├── scheduler.go      # 开启/结束/重置调度
│   │   │   ├── eligibility.go    # 可见性与参与资格
│   │   │   └── lifecycle.go      # 活动实例生命周期
│   │   │
│   │   ├── economy/              # 货币与背包
│   │   │   ├── wallet.go         # 钱包/货币
│   │   │   ├── inventory.go      # 背包/道具
│   │   │   └── transaction.go    # 交易日志
│   │   │
│   │   ├── messaging/            # 邮件与通知
│   │   │   ├── mail.go           # 邮件系统
│   │   │   └── notification.go   # 通知/公告
│   │   │
│   │   └── matchengine/          # 模拟比赛引擎（框架能力层，无RPC）
│   │       ├── service.go        # 引擎入口：每场比赛创建并执行 MatchEngine
│   │       ├── engine.go         # MatchEngine：每场比赛一个实例，推演完由GC回收
│   │       ├── fsm.go            # 比赛状态机
│   │       ├── combat.go         # 对战推演
│   │       └── model.go          # MatchInput / MatchResult / Team / Round 等引擎数据结构
│   │
│   └── shared/                 # ⚙️ 公共依赖模块（跨子系统复用的纯工具）
│       ├── constants.go        # 全局枚举
│       ├── errors.go           # 统一错误码
│       └── utils/              # 工具函数
├── go.mod
└── go.sum
```

# 游戏内服务器模块
第 1 步：穷举功能（按玩家旅程，不要按系统）

- 玩家信息
  - 基础信息（玩家昵称、头像等 Nakama 自带或自定义字段）

- 俱乐部（我的阵容）
  - 新手默认阵容
  - 选手上下阵
  - 管理选手删改查
  - 调整选手倾向（2.0做）

- 选手图鉴/数据库（可选）
  - 全局选手信息展示

- 抽卡
  - 常驻抽卡
  - 活动抽卡

- 商店
  - 常驻商店
  - 活动商店

- 任务与成就
  - 日常任务
  - 周常任务
  - 新手任务
  - 成就
  - 活动任务

- 限时/节日活动
  - 活动包装与生命周期
  - 关联任务/商店/抽卡

- 排行榜与段位
  - 当前段位
  - 排行榜
  - 历史赛季段位（作为读模型）

- 匹配对战
  - 人机对战（对决历年冠军）
  - 匹配对战（接入nakama匹配模块）

- 比赛推演
  - 模拟比赛游戏引擎

第 2 步：归类成子系统或框架能力

## 子系统

- **user / 玩家信息子系统**
  - 玩家昵称、头像、自定义资料
  - 玩家历史信息读模型聚合（可选）

- **club / 俱乐部与阵容子系统**
  - 新手默认阵容
  - 选手上下阵
  - 选手管理删改查
  - 调整选手倾向

- **player / 选手图鉴子系统（可选）**
  - 全局选手数据库/图鉴展示
  - 注：如果选手只是 Luban 配表、没有运行时详情界面，可不作为子系统

- **gacha / 抽卡子系统**
  - 常驻抽卡
  - 活动抽卡配置

- **shop / 商店子系统**
  - 常驻商店
  - 活动商店配置

- **quest / 任务与成就子系统**
  - 日常、周常、新手、成就、活动任务
  - 任务配置、进度追踪、领奖
  - 注：活动任务本身归这里，`activity` 只做包装和引用

- **activity / 限时活动子系统**
  - 节日/限时活动 orchestration
  - 活动时间、可见性、关联内容包装
  - 注：不自己实现任务/商店/抽卡，只引用 `quest` / `shop` / `gacha`

- **onboarding / 新手流程子系统（可选）**
  - 新玩家初始化编排（默认阵容、初始货币、激活新手任务等）
  - 注：项目早期可合并到 `user`，复杂后再独立

- **match / 匹配对战子系统**
  - 人机对战
  - 匹配对战
  - 触发 `match.completed` 事件

- **ranking / 排行榜与段子系统**
  - 排行榜接入
  - 当前段位
  - 历史赛季段位

## 框架能力

- **questkit / 任务条件框架**
  - 条件注册
  - 事件分发
  - 进度追踪

- **activitykit / 活动生命周期框架**
  - 活动开启/结束/重置调度
  - 可见性与参与资格

- **economy / 经济与背包框架**
  - 钱包（货币）
  - 背包（道具）
  - 交易日志

- **messaging / 邮件通知框架**
  - 邮件
  - 通知/公告

- **matchengine / 模拟比赛引擎**
  - 纯净的比赛状态机推演

## 共享模块

- **shared**
  - 常量
  - 错误码
  - 工具函数


## 预期各文件结构及开发范式

在 Nakama 框架下，每个业务子系统通常由以下四层文件组成。核心原则：**`api_rpc.go` 只负责协议转换，业务逻辑写在 `service.go`，数据访问写在 `repository.go`，数据结构定义在 `model.go`。**

### model.go

定义本模块的领域模型、RPC 请求/响应 DTO、以及 Nakama Storage 序列化结构。

```go
package club

import "time"

// --- 领域模型 ---
type Club struct {
    UserID    string
    Name      string
    Level     int
    CreatedAt time.Time
    Roster    []RosterSlot
}

type RosterSlot struct {
    PlayerID string
    Position int
}

// --- RPC 请求/响应 DTO ---
type UpdateRosterRequest struct {
    PlayerID string `json:"player_id"`
    Position int    `json:"position"`
}

type UpdateRosterResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}

// --- Storage 序列化结构 ---
type ClubStorage struct {
    Name   string       `json:"name"`
    Level  int          `json:"level"`
    Roster []RosterSlot `json:"roster"`
}
```

### repository.go

封装 Nakama Storage 读写。每个子系统只操作自己的 collection，不跨模块读写。

```go
package club

import (
    "context"
    "database/sql"
    "encoding/json"

    "github.com/heroiclabs/nakama-common/runtime"
)

const (
    StorageCollection = "club"
    StorageKeyClub    = "club"
)

type Repository struct {
    db *sql.DB
    nk runtime.NakamaModule
}

func NewRepository(db *sql.DB, nk runtime.NakamaModule) *Repository {
    return &Repository{db: db, nk: nk}
}

func (r *Repository) GetClub(ctx context.Context, userID string) (*Club, error) {
    reads, err := r.nk.StorageRead(ctx, []*runtime.StorageRead{
        {
            Collection: StorageCollection,
            Key:        StorageKeyClub,
            UserID:     userID,
        },
    })
    if err != nil {
        return nil, err
    }
    if len(reads) == 0 {
        return nil, ErrClubNotFound
    }

    var club Club
    if err := json.Unmarshal([]byte(reads[0].Value), &club); err != nil {
        return nil, err
    }
    return &club, nil
}

func (r *Repository) SaveClub(ctx context.Context, userID string, club *Club) error {
    value, err := json.Marshal(club)
    if err != nil {
        return err
    }

    _, err = r.nk.StorageWrite(ctx, []*runtime.StorageWrite{
        {
            Collection: StorageCollection,
            Key:        StorageKeyClub,
            UserID:     userID,
            Value:      string(value),
        },
    })
    return err
}
```

### service.go

核心业务逻辑。持有 repository 和依赖的其他 service/framework，处理规则校验、状态变更、跨模块调用。

```go
package club

import (
    "context"
    "fmt"

    "github.com/heroiclabs/nakama-common/runtime"

    "project/internal/framework/economy"
)

type Service struct {
    repo    *Repository
    economy *economy.Service
    logger  runtime.Logger
}

func NewService(repo *Repository, economy *economy.Service, logger runtime.Logger) *Service {
    return &Service{
        repo:    repo,
        economy: economy,
        logger:  logger,
    }
}

// InitializeDefaultClub 新玩家初始化默认俱乐部
func (s *Service) InitializeDefaultClub(ctx context.Context, userID string) error {
    club := &Club{
        UserID: userID,
        Name:   "Default Club",
        Level:  1,
        Roster: defaultRoster(),
    }
    return s.repo.SaveClub(ctx, userID, club)
}

// UpdateRoster 调整阵容
func (s *Service) UpdateRoster(ctx context.Context, userID string, req UpdateRosterRequest) (*UpdateRosterResponse, error) {
    club, err := s.repo.GetClub(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("get club: %w", err)
    }

    // 业务校验
    if req.Position < 0 || req.Position >= len(club.Roster) {
        return nil, ErrInvalidPosition
    }

    // 业务操作
    club.Roster[req.Position].PlayerID = req.PlayerID

    if err := s.repo.SaveClub(ctx, userID, club); err != nil {
        return nil, err
    }

    return &UpdateRosterResponse{Success: true}, nil
}
```

### api_rpc.go

Nakama RPC 入口。只做三件事：解析 payload → 调用 service → 返回 JSON。

```go
package club

import (
    "context"
    "database/sql"
    "encoding/json"

    "github.com/heroiclabs/nakama-common/runtime"
)

func RPCUpdateRoster(service *Service) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
    return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
        var req UpdateRosterRequest
        if err := json.Unmarshal([]byte(payload), &req); err != nil {
            return "", err
        }

        // 从 context 获取当前用户 ID
        userID, ok := ctx.Value(runtime.USER_ID).(string)
        if !ok {
            return "", ErrUnauthorized
        }

        resp, err := service.UpdateRoster(ctx, userID, req)
        if err != nil {
            errResp := UpdateRosterResponse{Success: false, Error: err.Error()}
            bytes, _ := json.Marshal(errResp)
            return string(bytes), nil
        }

        bytes, err := json.Marshal(resp)
        if err != nil {
            return "", err
        }
        return string(bytes), nil
    }
}
```

### main.go / InitModule 中的注册

所有依赖注入和 RPC 注册统一在 `main.go` 的 `InitModule` 中完成。

```go
package main

import (
    "context"
    "database/sql"

    "github.com/heroiclabs/nakama-common/runtime"

    "project/internal/club"
    "project/internal/framework/economy"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
    // 初始化 economy
    economyRepo := economy.NewRepository(db, nk)
    economyService := economy.NewService(economyRepo)

    // 初始化 club
    clubRepo := club.NewRepository(db, nk)
    clubService := club.NewService(clubRepo, economyService, logger)

    // 注册 RPC
    if err := initializer.RegisterRpc("club_update_roster", club.RPCUpdateRoster(clubService)); err != nil {
        return err
    }

    return nil
}
```

### 关键约定

1. **`api_rpc.go` 不直接调用 repository**：必须通过 service。
2. **`service.go` 不直接调用 Nakama API**：通过 repository 抽象 Storage/Leaderboard/Wallet 等操作。
3. **`repository.go` 只负责本模块的 collection**：不要跨 collection 读写其他模块数据。
4. **`model.go` 区分三类模型**：领域模型、RPC DTO、Storage 结构，视情况可合并。
5. **错误统一处理**：service 返回 Go error，RPC 层转换为前端友好的响应结构。
6. **依赖注入在 InitModule**：避免全局变量，方便单元测试。

## 跨子系统调用约定

不同业务子系统之间需要互相调用时，**不应通过 `api_rpc.go`**。`api_rpc.go` 只负责把客户端的 Nakama RPC 请求转给 `service.go`，子系统内部调用应直接走 `service.go` 暴露的公共方法。

### 推荐调用路径

```plaintext
子系统 A 的 service.go
    ↓ 直接方法调用
子系统 B 的 service.go
    ↓
子系统 B 的 repository.go
```

例如 `club` 子系统需要扣款时，应持有 `economy.Service` 的引用并直接调用：

```go
// internal/club/service.go
type Service struct {
    repo    *Repository
    economy *economy.Service
    logger  runtime.Logger
}

func (s *Service) DoSomething(ctx context.Context, userID string) error {
    if err := s.economy.Deduct(ctx, userID, 100); err != nil {
        return err
    }
    // ...
}
```

### 不应做的事

- ❌ 在 `service.go` 里通过 `nk.Rpc(...)` 调用同进程内的其他 RPC。
- ❌ 一个子系统的 `api_rpc.go` 直接调用另一个子系统的 `api_rpc.go`。
- ❌ 子系统之间直接读写对方的 Storage collection（应通过对方的 `repository` 抽象，最好是通过对方的 `service`）。

## 框架能力层与业务子系统的关系（以 matchengine 为例）

`framework/matchengine` 是**框架能力层**，不是业务子系统。它的定位是：

> 纯净的比赛状态机推演。

因此：

- **框架能力层不提供 `api_rpc.go`**。
- 框架能力层只暴露干净的 Go API，供业务子系统的 `service.go` 调用。
- 客户端永远不应该直接调用 `matchengine`。

### 分层示例

```plaintext
客户端
  ↓ RPC
internal/match/api_rpc.go        # 业务 RPC 入口：StartMatch、StartBotMatch 等
  ↓
internal/match/service.go        # 业务编排：查阵容、构造 Bot、保存记录、触发事件
  ↓ 方法调用
internal/framework/matchengine   # 纯净推演：Simulate(MatchInput) -> MatchResult
```

### `match` 子系统的 service 持有 `matchengine.Service`

```go
// internal/match/service.go
package match

import "windypath.com/cs2match/server/internal/framework/matchengine"

type Service struct {
    repo        *Repository
    engine      *matchengine.Service
    clubService *club.Service
    logger      runtime.Logger
}

func (s *Service) StartBotMatch(
    ctx context.Context,
    userID string,
    req StartBotMatchRequest,
) (*StartMatchResponse, error) {
    roster, err := s.clubService.GetRoster(ctx, userID)
    if err != nil {
        return nil, err
    }

    input := &matchengine.MatchInput{
        TeamA: roster.ToEngineTeam(),
        TeamB: s.botResolver.Resolve(req.BotID),
        MapID: req.MapID,
        Seed:  s.generateSeed(),
    }

    result, err := s.engine.Simulate(ctx, input)
    if err != nil {
        return nil, err
    }

    // 保存对战记录、触发 match.completed 事件、发放奖励等
    // ...

    return &StartMatchResponse{MatchID: result.MatchID}, nil
}
```

### 每场比赛一个 `MatchEngine` 实例

`matchengine` 内部推荐采用**每场比赛新建一个 `MatchEngine` 对象**的模式：接收到推演请求时创建实例，调用 `StartMatch` 完成整局推演，返回结果后对象自然由 Go GC 回收。

这种设计的好处：

- **状态隔离**：每场比赛拥有独立的 FSM、随机源、tick 计数器，不存在共享可变状态。
- **并发安全**：多个比赛同时推演时互不干扰，不需要额外加锁。
- **可回放**：通过固定的 `Seed` 可以复现同一场比赛。
- **可测试性**：单个 `MatchEngine` 可以独立构造、独立推演、独立断言。

```go
// internal/framework/matchengine/engine.go
package matchengine

import (
    "context"
    "math/rand"
)

type MatchEngine struct {
    input *MatchInput
    state *MatchState
    rng   *rand.Rand
    tick  int
}

func NewMatchEngine(input *MatchInput) *MatchEngine {
    return &MatchEngine{
        input: input,
        state: initialState(input),
        rng:   rand.New(rand.NewSource(input.Seed)),
        tick:  0,
    }
}

// StartMatch 执行完整推演，返回结果
func (e *MatchEngine) StartMatch(ctx context.Context) (*MatchResult, error) {
    for !e.isFinished() {
        if err := e.tickOnce(); err != nil {
            return nil, err
        }
    }
    return e.buildResult(), nil
}
```

`matchengine.Service` 作为工厂和调用入口：

```go
// internal/framework/matchengine/service.go
package matchengine

type Service struct {
    logger runtime.Logger
}

func NewService(logger runtime.Logger) *Service {
    return &Service{logger: logger}
}

func (s *Service) Simulate(ctx context.Context, input *MatchInput) (*MatchResult, error) {
    engine := NewMatchEngine(input)
    return engine.StartMatch(ctx)
}
```

`match.Service` 仍然只依赖 `matchengine.Service`，不直接构造 `MatchEngine`：

```go
// internal/match/service.go
result, err := s.engine.Simulate(ctx, input) // ✅
// engine := matchengine.NewMatchEngine(input) // ❌ 不推荐
```

### `MatchEngine` 的注意事项

1. **不要依赖业务子系统**：`MatchEngine` 只操作 `MatchInput` 和 `MatchResult`，不查询玩家阵容、不发奖励、不写 Storage。
2. **随机源必须基于 `Seed`**：每个实例使用独立种子，避免共享全局随机源导致并发问题和不可回放。
3. **支持两种生命周期**：
   - 人机对战/离线推演：`Service.Simulate()` 内部创建 `MatchEngine`，一次调用完成整局推演。
   - Nakama Match Handler 实时 1v1：Match Handler 持有同一个 `MatchEngine` 实例，在 `MatchLoop` 的每个 tick 中调用 `TickOnce()`，直到比赛结束。

### 依赖方向

业务子系统可以依赖框架能力层，**框架能力层不能反向依赖业务子系统**。

```plaintext
match → matchengine   ✅
matchengine → match   ❌
```

## RPC 请求/响应 DTO 的放置

**RPC 请求/响应 DTO 应放在发起该 RPC 的子系统的 `model.go` 中**，不要全局统一注册。

### 为什么不分全局注册

- 每个 DTO 只属于暴露该 RPC 的子系统。
- 全局注册会增加无关模块之间的耦合。
- 违反“本模块的 `model.go` 只定义本模块关心的结构”的约定。

### 示例

`match` 子系统暴露 `StartBotMatch` RPC，则其 DTO 放在 `internal/match/model.go`：

```go
// internal/match/model.go
package match

type StartBotMatchRequest struct {
    BotID int64 `json:"bot_id"`
    MapID int64 `json:"map_id"`
}

type StartBotMatchResponse struct {
    MatchID string `json:"match_id"`
}
```

框架能力层 `matchengine` 的入参/出参也放在自己的 `model.go` 中：

```go
// internal/framework/matchengine/model.go
package matchengine

type MatchInput struct {
    TeamA *Team
    TeamB *Team
    MapID int64
    Seed  int64
}

type MatchResult struct {
    Winner    int
    Rounds    []Round
    Stats     *MatchStats
    ReplayKey string
}
```

### 不要把业务 RPC DTO 放到框架能力层

`StartBotMatchRequest` 属于 `match` 子系统的 RPC 契约，不应放到 `matchengine/model.go`。引擎不应该知道“BotID 是哪位冠军”、“玩家从大厅选择了什么”等业务概念。

## 多入口业务发起与调用路径

当其他业务需要“临时打一场比赛”时，入口选择取决于这场比赛是否需要走完整业务生命周期。

### 需要产生真实对局记录 → 走 `match.Service`

如果这场比赛需要：

- 查询玩家真实阵容
- 保存对战记录到 Storage
- 触发 `match.completed` 事件
- 影响任务、成就、排行榜、活动
- 发放奖励

入口应统一在 `match.Service`：

```go
// internal/match/service.go
func (s *Service) StartBotMatch(ctx context.Context, userID string, req StartBotMatchRequest) (*StartMatchResponse, error)
func (s *Service) StartTutorialMatch(ctx context.Context, userID string, req StartTutorialMatchRequest) (*StartMatchResponse, error)
func (s *Service) StartPvPMatch(ctx context.Context, userID string, req StartPvPMatchRequest) (*StartMatchResponse, error)
```

这样 `quest`、`activity`、`ranking` 等子系统只需监听 `match.completed` 事件，无需关心比赛如何发起。

### 只需要纯推演结果 → 可直接调 `matchengine.Service`

如果某个场景只是需要引擎算一下结果，不产生业务副作用，例如：

- 后台平衡性测试工具
- AI 训练/蒙特卡洛模拟
- 管理后台的“模拟对阵”预览

可以直接调用 `matchengine.Service`：

```go
// internal/admin/balancetool/service.go
input := &matchengine.MatchInput{TeamA: teamA, TeamB: teamB, MapID: 1}
result, err := s.engine.Simulate(ctx, input)
```

但注意：这种调用方通常不是面向玩家的业务子系统，而是工具/管理/批处理模块。

### 不应该做的事

❌ 不要让 `club`、`quest`、`gacha` 等业务子系统直接持有 `matchengine.Service` 来发起玩家相关的比赛。否则会产生“幽灵比赛”——任务系统认为比赛已完成，但 `match` 子系统中没有任何记录，排行榜、活动、回放都会出问题。

## 业务请求与引擎输入的分层

`match` 子系统的 RPC 请求和 `matchengine` 的引擎输入是两套结构，不要合并。

| 结构 | 所在位置 | 含义 |
|---|---|---|
| `StartBotMatchRequest` | `internal/match/model.go` | 客户端业务请求：打哪个 Bot、哪张图 |
| `MatchInput` | `internal/framework/matchengine/model.go` | 引擎推演输入：两支队伍、地图、随机种子 |

### 字段对比示例

| 字段 | `match.StartBotMatchRequest` | `matchengine.MatchInput` |
|---|---|---|
| `TeamA` | ❌ 没有 | ✅ `*Team`（玩家阵容） |
| `TeamB` | ❌ 没有 | ✅ `*Team`（Bot 队伍） |
| `BotID` | ✅ `int64` | ❌ 引擎不需要 |
| `MapID` | ✅ `int64` | ✅ `int64` |
| `Seed` | ❌ 没有 | ✅ `int64`（服务端生成） |

`match.Service` 负责把业务请求翻译成引擎输入：

```go
input := &matchengine.MatchInput{
    TeamA: roster.ToEngineTeam(),
    TeamB: s.botResolver.Resolve(req.BotID),
    MapID: req.MapID,
    Seed:  s.generateSeed(),
}
```

这不是重复定义，而是不同抽象层的数据结构，类似于 Web 开发中的 HTTP Request DTO、Service Input DTO 和 Database Entity 之分。
