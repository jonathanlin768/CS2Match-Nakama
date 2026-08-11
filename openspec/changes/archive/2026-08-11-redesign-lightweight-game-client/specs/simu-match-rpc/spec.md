## ADDED Requirements

### Requirement: 生产用 SimuMatch RPC 接受权威可校验阵容引用
系统 SHALL 在 `InitModule` 注册 `SimuMatch` RPC。请求 SHALL 包含 `mode`、地图或配置引用以及必要的 Player ID，不得接受客户端提供的选手属性、价格、胜率或比赛结果。现有 `DebugSimuMatch` SHALL 保留为调试兼容入口。

#### Scenario: 教学模式请求
- **GIVEN** 客户端具有有效 Guest 或正式 Session
- **WHEN** 提交 `mode=tutorial`、tutorial config ID、config version 和五个 Player ID
- **THEN** RPC 将请求交给 `internal/match` 服务端校验与构建
- **AND** 不使用客户端提交的任何属性或价格字段

#### Scenario: 未认证请求
- **GIVEN** 请求上下文没有 Nakama user ID
- **WHEN** 调用 `SimuMatch`
- **THEN** RPC 返回结构化 `UNAUTHORIZED` 错误

### Requirement: 教学模式按 Luban 配置校验双方阵容
`SimuMatch` 的 tutorial 模式 SHALL 从 `TbTutorialBattle` 读取预算、人数、价格池、地图和机器人 Player ID，并从 `TbPlayer` 读取十名选手属性。服务端 SHALL 校验玩家恰好五人、ID 唯一、属于配置池、总价不超预算，对手恰好五人且所有引用有效。

#### Scenario: 合法教学请求进入引擎
- **GIVEN** 请求匹配启用且版本一致的教学配置
- **WHEN** 服务端完成双方阵容校验
- **THEN** `internal/match` 构造包含真实双方 `TeamInput` 的 `MatchInput`
- **AND** 调用现有 `matchengine.Service.Simulate`

#### Scenario: 过期配置版本
- **GIVEN** 客户端提交的 config version 与当前服务端配置不一致
- **WHEN** 服务端校验请求
- **THEN** 返回 `CONFIG_VERSION_MISMATCH`
- **AND** 不进入模拟引擎

#### Scenario: 超预算或越池选手
- **GIVEN** 玩家阵容总价超过配置预算或包含未列入任一价格档的选手
- **WHEN** 服务端校验阵容
- **THEN** 返回 `INVALID_TUTORIAL_LINEUP`
- **AND** 不生成战报

### Requirement: SimuMatch 返回现有标准 MatchReport
`SimuMatch` SHALL 返回与现有模拟回放兼容的完整 `MatchReport`，并使双方队名、Player ID、选手属性、比分、事件和胜者来自本次服务端 `MatchInput` 与模拟结果。

#### Scenario: 前端播放教学战
- **GIVEN** `SimuMatch` 成功模拟教学双方真实阵容
- **WHEN** 客户端接收响应
- **THEN** 响应可由现有 `MatchReport` TypeScript 类型解析
- **AND** BattlePage 无需客户端重新计算比赛结果

### Requirement: 电脑模式不依赖玩家正式阵容
`SimuMatch` 的 computer 模式 SHALL 在本期复用服务端配置的默认双方队伍，从 `TbTeam` 与 `TbPlayer.teamId` 构造比赛输入并直接触发战斗。该模式 SHALL NOT 读取或创建玩家正式阵容，也不得把默认队伍表述为玩家拥有资产。

#### Scenario: 从主界面开始电脑战
- **GIVEN** 任意 Guest 或正式 Session 从匹配对战选择电脑模式
- **WHEN** 客户端调用 `SimuMatch(mode=computer)`
- **THEN** 服务端用默认配置双方构造真实 `MatchInput` 并返回战报
- **AND** 不读写玩家阵容 Storage

### Requirement: SimuMatch 不引入 Match Handler 状态同步
`SimuMatch` SHALL 保持一次 RPC 内完成整场离线模拟并返回战报，不注册 Match Handler，也不影响 MatchLoop 的 10Hz/20Hz tick 行为。

#### Scenario: 插件初始化
- **GIVEN** 本变更已经部署
- **WHEN** Nakama 执行 `InitModule`
- **THEN** 注册 `SimuMatch` RPC 但不新增该功能的 Match Handler
- **AND** 现有 MatchLoop tick 行为不变
