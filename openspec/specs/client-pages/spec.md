# Client Pages

CS2 模拟器前端页面规格 — 登录页、首页仪表盘、对战页面、抽卡系统、排行榜。

## Purpose

定义 5 个核心页面的 UI 结构和行为。当前阶段所有数据使用硬编码 mock（不接 Nakama 后端），后续逐步接入真实服务器逻辑。

## Requirements

### Requirement: LoginPage 登录页面

系统 SHALL 提供 LoginPage，包含邮箱/密码输入框、登录按钮、社交登录入口和注册链接。点击登录后直接导航到 `/home`，不做真实认证。

#### Scenario: 渲染登录表单

- **WHEN** 用户访问 `/`
- **THEN** 页面显示 CS2 SIMULATOR 品牌标识
- **AND** 显示邮箱输入框和密码输入框
- **AND** 显示"登录"按钮、"记住我"复选框和"忘记密码"链接

#### Scenario: 点击登录按钮

- **WHEN** 用户点击"登录"按钮
- **THEN** 按钮显示加载状态（spinner + "登录中..."）约 1 秒
- **AND** 页面导航到 `/home`

#### Scenario: 密码显示切换

- **WHEN** 用户点击密码输入框右侧的眼睛图标
- **THEN** 密码从隐藏状态切换为明文显示
- **AND** 图标从 Eye 切换为 EyeOff

### Requirement: Home 首页

Home 页面 SHALL 展示 CS2 模拟器的核心仪表盘布局，包含玩家信息、赛季横幅、今日任务、快速开始、活动中心、战队排行、最近比赛和卡包推荐等模块。

#### Scenario: 首页渲染

- **WHEN** 用户访问 `/home`
- **THEN** 页面渲染 AppLayout 布局（Header + 内容 + Footer）
- **AND** 内容区显示 PlayerSidebar、SeasonBanner、DailyTasks、ChampionshipCenter、QuickStart、ActivityCenter、TeamRankings、RecentMatches、CardPack 共 9 个业务组件
- **AND** 所有组件使用硬编码 mock 数据正常渲染

#### Scenario: 响应式布局

- **WHEN** 用户在不同屏幕尺寸下访问首页
- **THEN** 移动端（<1024px）组件垂直堆叠
- **AND** 桌面端组件按水平网格布局排列

### Requirement: MatchPage 对战页面

MatchPage SHALL 展示模拟对战的比赛界面，包含比分栏、双方阵容、战术地图、回合事件、玩家数据和战术部署。当前阶段使用 stub hook（不连接 WebSocket），显示空状态界面。

#### Scenario: 对战页面渲染（未连接状态）

- **WHEN** 用户访问 `/match`
- **THEN** 页面顶部显示红色"未连接"状态指示器
- **AND** MatchScoreBar 显示双方队名、Logo、比分和倒计时
- **AND** 左右两侧显示 TeamRoster（T 方和 CT 方，空阵容）
- **AND** 中间显示 TacticalMap（Dust2 地图）
- **AND** 底部显示 RoundEvents（暂无事件）、PlayerStats（空表格）、RoundTactics（暂无战术）

#### Scenario: 回合事件展示

- **WHEN** 没有模拟事件发生（stub 返回空数组）
- **THEN** RoundEvents 显示"暂无事件"
- **AND** RoundTactics 显示"暂无战术信息"

### Requirement: GachaPage 抽卡页面

GachaPage SHALL 展示抽卡系统界面，包含卡包选择侧边栏、卡牌轮播、保底面板和卡池预览。

#### Scenario: 抽卡页面渲染

- **WHEN** 用户访问 `/gacha`
- **THEN** 左侧显示 PackSidebar（默认选中"传奇之路"）
- **AND** 中间显示 CardCarousel（玩家卡牌轮播）
- **AND** 右侧显示 PityPanel（幸运值进度条、保底机制、掉落概率）
- **AND** 底部显示 PoolPreview（卡池可选卡牌一览）

#### Scenario: 切换卡包类别

- **WHEN** 用户在 PackSidebar 点击不同的卡包类别
- **THEN** 选中项高亮显示（`bg-primary/10 border-l-primary`）
- **AND** `selectedPack` 状态更新

### Requirement: RankingPage 排行页面

RankingPage SHALL 展示排行榜系统界面，包含排行榜类别侧边栏、排行表格和右侧信息面板。

#### Scenario: 排行页面渲染

- **WHEN** 用户访问 `/ranking`
- **THEN** 左侧显示 RankingSidebar（默认选中"玩家排行榜"）
- **AND** 中间显示 RankingTable（玩家排行列表，含排名、头像、ELO、综合分等列）
- **AND** 右侧显示 RankingInfoPanel（排行榜说明、奖励信息、个人详情）

#### Scenario: 移动端表格适配

- **WHEN** 用户在移动端访问排行页面
- **THEN** RankingTable 表格列缩减为排名、玩家名和综合分
- **AND** 其余列（ELO、胜场、胜率、K/D、资产、阵容）隐藏

### Requirement: 全局暗金主题

所有页面 SHALL 使用统一的 dark gold 主题，CSS 变量定义在全局样式表中，通过 Tailwind CSS 4 的 `@theme inline` 映射为语义类名。

#### Scenario: 主题显示

- **WHEN** 任意页面渲染
- **THEN** 背景色为 `#0a0e14`（`bg-background`）
- **AND** 卡片背景为 `#141a24`（`bg-card`）
- **AND** 主色调为金色 `#c9a227`（`text-primary`）
- **AND** 边框色为 `#2a3444`（`border-border`）
