## Why

当前前端以固定 1920×900 游戏大厅为中心，并承载活动、货币、抽卡、邮件、排行榜等尚未形成核心价值的重型入口，增加了开发成本，也让玩家难以快速进入阵容选择与比赛。需要将产品收敛为可在桌面和手机浏览器直接游玩的轻量体验，同时复用已经可用的比赛模拟、战报回放和 Nakama 好友能力。

## What Changes

- **BREAKING** 完全重构前端信息架构与视觉外壳，废弃固定 19:9/1920×900 裁剪大厅，改为以桌面和移动端自适应页面承载三个主入口：调整阵容、好友对战、匹配对战。
- 为未登录访客提供可浏览主界面和首次体验入口；首次访问弹出战斗体验 Modal，确认后进入临时“15 元、5/4/3/2/1 档选手池”组队流程并对战机器人，拒绝后留在主界面。
- “调整阵容”入口在本变更中显示为未开放，不实现首发队领取、正式阵容保存或阵容编辑；15 元教学阵容仅用于触发当次教学战，比赛结束后不写入账号。
- 保留现有 `DebugSimuMatch` 作为调试兼容入口，并新增面向游戏流程的 `SimuMatch` RPC；教学战将玩家选择的五名选手和 Luban 配置的机器人五人阵容转换为真实 `MatchInput` 后进入现有 `matchengine`，复用服务端战报和客户端回放数据契约。
- 保留 Nakama 好友列表、好友申请和实时连接基础设施，在新界面中提供好友与联系方式交换；“好友对战”主入口在本阶段只显示未开放，不新增好友挑战、比赛房间或实时 Match Handler。
- **BREAKING** 移除任意文本私聊能力，聊天区域只允许发送服务端定义的“请求交换联系方式”卡片；好友接受后才可按请求授权查看双方填写的 QQ 和/或微信号。
- 新增 Nakama `internal/social` 子系统、联系方式资料/交换状态 Storage、社交 RPC 和实时消息写入拦截。联系方式内容不写入频道消息，客户端不能绕过 RPC 直接发送私聊内容。
- 使用系统生成的 `玩家#XXXXXX` 风格显示标识，避免开放玩家自定义公开昵称及其敏感词审核范围；账号身份仍以 Nakama user ID 为权威键。

## Capabilities

### New Capabilities

- `guest-battle-onboarding`: 未登录访问、首次体验 Modal、15 元临时组队和机器人教学战的产品行为。
- `social-contact-exchange`: 好友间联系方式资料、请求卡片、授权读取、服务端校验和消息安全限制。

### Modified Capabilities

- `game-lobby-ui`: 用响应式三入口主界面替换固定画幅、活动栏、货币、邮件和底部重型导航。
- `client-routing`: 根路径、访客可访问页面、登录后页面及需要认证的社交/账号能力采用新的路由与守卫规则。
- `client-pages`: 首页、登录入口、阵容页、对战入口和好友消息界面调整为新的轻量页面契约，并移除抽卡等非本期页面入口。
- `client-chat`: 取消客户端任意文本发送，将既有 DM 实时通道限制为服务端生成的结构化社交卡片。
- `simu-client-replay`: 从新匹配入口调用现有模拟 RPC，并在不改变权威战报含义的前提下迁移至响应式比赛与回放表现。
- `simu-match-rpc`: 新增生产用阵容模拟请求，并由服务端校验教学模式、双方选手与配置版本后构造真实 MatchInput。
- `simu-config-structures`: 新增队伍与教学战 Luban 配置，将 Player 的自由文本 team 改为 teamId 引用，并从配置构建双方阵容。
- `client-auth`: 增加静默设备访客身份、正式账号切换、8 位系统 username 与 Session 生命周期。
- `login-page-game-style`: 用响应式登录/注册 Modal 替换固定画幅独立登录页，并支持认证后返回原目标。

## Impact

- **React 前端**：重写应用外壳、路由、首页、首次体验、匹配、好友及消息视图；阵容调整与好友对战只提供未开放状态；复用 Nakama Session、好友 API、Socket 和 `MatchReport` 回放状态，废弃旧大厅导航及文本输入聊天 UI。
- **Nakama 后端**：新增 `server/internal/social` 模块并在 `InitModule` 注册联系方式 RPC 与实时频道消息 before-hook；扩展 `internal/match` 适配层与 RPC，但现有模拟引擎和比赛规则无需改变。
- **RPC**：新增生产用 `SimuMatch`，以及设置本人联系方式、读取交换状态、发起交换请求、响应交换请求等 RPC；本变更不新增阵容 RPC，现有 `DebugSimuMatch` 保留为调试兼容入口。
- **Storage**：只新增服务器控制读写的联系方式资料和双人交换状态记录；不新增玩家阵容 Storage，客户端不得直接读取联系方式 Storage。
- **Match Handler / 状态同步**：本变更不新增或修改 Match Handler，不影响 MatchLoop tick、服务器权威比赛状态或战报生成。真实路人匹配和好友权威对局留作后续变更。
- **数据库与部署**：使用 Nakama Storage Engine，不新增 PostgreSQL schema migration；现有 Docker Compose/AWS 单节点部署拓扑不变。
- **Luban**：新增 `#Team.xlsx` 与 `#TutorialBattle.xlsx`，并将 `#Player.xlsx` 的自由文本 `team` 改为引用 `TbTeam` 的 `teamId`。Team 表承载队名、简称、昵称和 Logo；TutorialBattle 表按版本配置预算、人数、5/4/3/2/1 元选手池、机器人对手五人阵容与地图。导表后同步 Go/TypeScript schema、JSON 和 loader。
