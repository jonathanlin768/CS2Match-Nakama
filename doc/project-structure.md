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
│   │   ├── api_rpc.go          # RPC 接口 (如：开始匹配)
│   │   ├── service.go          # 匹配逻辑、人机对战、事件触发
│   │   ├── repository.go       # 对战记录 Storage 读写
│   │   └── model.go
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
│   │   └── matchengine/          # 模拟比赛引擎
│   │       ├── fsm.go            # 状态机
│   │       └── combat.go         # 对战推演
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
