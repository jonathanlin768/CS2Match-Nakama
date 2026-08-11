# game-lobby-ui Specification

## Purpose
TBD - created by archiving change migrate-to-game-lobby. Update Purpose after archive.
## Requirements
### Requirement: Home 页面渲染游戏大厅

Home 页面 SHALL 在认证后展示一个固定 1920×900 的游戏大厅界面，替代原有的响应式仪表盘布局。

#### Scenario: 首页渲染

- **WHEN** 已认证用户访问 `/home`
- **THEN** 页面渲染一个水平垂直居中的 1920×900 固定画幅
- **AND** 画幅内自上而下显示 TopBar、PromoBar、三列内容区、BottomNav
- **AND** 三列内容区从左到右显示 LeftPanel、CenterStage、RightPanel

### Requirement: 源码无篮球主题命名

所有游戏大厅组件与资源路径 SHALL 不包含 "nba"、"basketball" 等篮球相关命名或文案。

#### Scenario: 源码检查

- **WHEN** 对 `src/components/lobby/`、`src/pages/HomePage.tsx`、`src/index.css` 执行大小写不敏感搜索
- **THEN** 不存在 "nba"、"basketball"、"lineup"（组件名与文件名除外需已重命名）等字符串
- **AND** 第一版允许的占位图片路径 `/images/star-player.png` 可保留

### Requirement: 主界面只突出三个核心操作
Home 主界面 SHALL 将“调整阵容”“好友对战”“匹配对战”作为唯一一级玩法入口。账号、好友通知和设置 MAY 作为低层级辅助入口，但 SHALL NOT 与三个核心操作竞争主视觉层级。本变更中“调整阵容”和“好友对战” SHALL 明确显示未开放且不可进入功能流程，“匹配对战” SHALL 提供可用的电脑模拟入口。

#### Scenario: 正式玩家查看主界面
- **GIVEN** 玩家已经正式登录
- **WHEN** `/` 主界面完成渲染
- **THEN** 页面可直接识别三个入口，并能操作匹配对战
- **AND** 调整阵容和好友对战显示一致的未开放状态
- **AND** 不显示活动、任务、背包、邮件、抽卡、商城或多货币入口

#### Scenario: 访客查看主界面
- **GIVEN** 当前身份为访客
- **WHEN** 主界面完成渲染
- **THEN** 三个核心操作仍然可见
- **AND** 未开放入口不因访客或正式登录状态而改变为可用

### Requirement: 主界面采用真实响应式布局
主界面 SHALL 使用流式布局适应手机、平板和桌面视口，不得通过整体缩放或裁剪固定画布实现适配。桌面内容区 SHALL 有可读的最大宽度，移动端 SHALL 支持纵向滚动、安全区和至少 44px 的主要触控目标。

#### Scenario: 桌面端布局
- **GIVEN** 视口宽度不小于 1024px
- **WHEN** 主界面渲染
- **THEN** 核心内容在居中的有限宽度容器中组织
- **AND** 三个主操作无需横向滚动即可访问

#### Scenario: 手机端布局
- **GIVEN** 视口宽度约为 390px
- **WHEN** 主界面渲染
- **THEN** 核心操作以适合单手触控的纵向或紧凑布局展示
- **AND** 页面无由固定 1920×900 画布造成的横向溢出或内容裁剪

### Requirement: 新界面保持统一比赛品牌而非管理手游仪表盘
新 App Shell SHALL 使用统一的比赛品牌、清晰的文字层级和少量强调色组织页面，并优先表现阵容与对战。客户端表现 SHALL 不暗示不存在的奖励、赛季、经济或实时玩家数量等服务器状态。

#### Scenario: 数据尚未加载
- **GIVEN** 阵容或匹配 RPC 仍在加载
- **WHEN** 页面展示骨架或等待状态
- **THEN** UI 只表达正在读取或模拟
- **AND** 不使用伪造资产、奖励或在线人数填充界面

