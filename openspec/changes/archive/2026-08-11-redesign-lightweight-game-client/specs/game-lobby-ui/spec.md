## ADDED Requirements

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

## REMOVED Requirements

### Requirement: 1920×900 画幅居中裁剪
**Reason**: 固定画幅与手机浏览器及轻量网页体验目标冲突。
**Migration**: 迁移到真实响应式 App Shell 和流式页面容器。

### Requirement: 主画幅内边距
**Reason**: 固定像素内边距依赖旧 1920×900 画幅。
**Migration**: 使用响应式 spacing token 与安全区内边距。

### Requirement: 顶部信息栏 TopBar
**Reason**: 旧 TopBar 的货币、邮件和排行榜入口不属于本期核心产品。
**Migration**: 仅保留低层级账号、好友通知和设置入口。

### Requirement: 活动快捷栏 PromoBar
**Reason**: 本期明确不建设活动生态。
**Migration**: 无；删除活动入口。

### Requirement: 中间内容区尺寸
**Reason**: 固定内容区尺寸由旧画幅驱动。
**Migration**: 使用响应式容器。

### Requirement: 三列宽度分配
**Reason**: 固定三列无法覆盖新三入口在不同视口的布局。
**Migration**: 使用断点驱动的 Grid/Flex。

### Requirement: 左侧面板 LeftPanel
**Reason**: 旧管理面板信息架构被取消。
**Migration**: 阵容信息迁移到调整阵容页面。

### Requirement: 中央展示区 CenterStage
**Reason**: 旧大厅中央展示不再承担主要导航。
**Migration**: 主界面使用三个动作卡片或按钮。

### Requirement: 右侧面板 RightPanel
**Reason**: 旧任务/活动类侧栏不在本期范围。
**Migration**: 无。

### Requirement: 底部导航 BottomNav
**Reason**: 旧导航包含抽卡等已取消入口。
**Migration**: 页面导航围绕三个核心操作和返回主界面组织。

### Requirement: 招募入口与比赛 START 按钮高度以视觉对齐为准
**Reason**: 招募/抽卡入口被移除。
**Migration**: 新组件使用统一触控与按钮 token。

### Requirement: 面板内部卡片比例固定
**Reason**: 固定卡片高度妨碍响应式文字与移动端访问。
**Migration**: 卡片高度由内容和断点决定。
