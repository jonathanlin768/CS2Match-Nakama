# Game Lobby UI

固定 1920×900 游戏大厅首页的 UI 规格。

## ADDED Requirements

### Requirement: Home 页面渲染游戏大厅

Home 页面 SHALL 在认证后展示一个固定 1920×900 的游戏大厅界面，替代原有的响应式仪表盘布局。

#### Scenario: 首页渲染

- **WHEN** 已认证用户访问 `/home`
- **THEN** 页面渲染一个水平垂直居中的 1920×900 固定画幅
- **AND** 画幅内自上而下显示 TopBar、PromoBar、三列内容区、BottomNav
- **AND** 三列内容区从左到右显示 LeftPanel、CenterStage、RightPanel

### Requirement: 1920×900 画幅居中裁剪

系统 SHALL 将游戏大厅画幅固定在 1920×900，并在视口不足时居中显示，不缩放也不响应式适配。

#### Scenario: 大视口显示

- **WHEN** 浏览器视口大于等于 1920×900
- **THEN** 大厅画幅在视口中完全可见且居中

#### Scenario: 小视口裁剪

- **WHEN** 浏览器视口小于 1920×900
- **THEN** 大厅画幅保持 1920×900 不变
- **AND** 画幅在视口中居中显示
- **AND** 超出视口的部分被裁剪，不可滚动缩放

### Requirement: 主画幅内边距

`HomePage` 主容器 SHALL 使用非对称内边距：左右 80px，上下 15px。

#### Scenario: 内边距检查

- **WHEN** 检查 `client/src/pages/HomePage.tsx`
- **THEN** `main` 元素使用 `px-[80px] py-[15px]`
- **AND** 内容区不贴顶/底边框，左右保留握持边距

### Requirement: 顶部信息栏 TopBar

TopBar SHALL 高度固定为 80px，左侧展示玩家信息，右侧展示货币与系统入口图标。

#### Scenario: 顶部栏渲染

- **WHEN** 首页渲染
- **THEN** `TopBar` 高度为 80px
- **AND** 左侧显示玩家头像、等级、名称、战力、手册进度与胜率
- **AND** 右侧显示两种货币 pill、好友、邮件、个人信息、排行榜入口图标
- **AND** 货币 pill 与系统图标处于同一组，靠右对齐

### Requirement: 活动快捷栏 PromoBar

PromoBar SHALL 高度固定为 100px，活动图标靠右对齐。

#### Scenario: 活动入口渲染

- **WHEN** 首页渲染
- **THEN** `PromoBar` 高度为 100px
- **AND** `PromoBar` 使用 `justify-end` 使图标靠右
- **AND** 显示 7 个快捷入口，每个入口包含图标、标签与可选角标

### Requirement: 中间内容区尺寸

中间三列内容区 SHALL 高度为 500px，TopBar、PromoBar、BottomNav 高度相应调整，使总高度精确等于 900px。

#### Scenario: 内容区尺寸验证

- **WHEN** 检查 `client/src/pages/HomePage.tsx`
- **THEN** 三列容器高度为 500px
- **AND** TopBar 高度为 80px，PromoBar 高度为 100px，BottomNav 高度为 130px
- **AND** 总高度满足：15 + 80(TopBar) + 20 + 100(PromoBar) + 20 + 500(Content) + 20 + 130(BottomNav) + 15 = 900px

### Requirement: 三列宽度分配

LeftPanel、CenterStage、RightPanel SHALL 按固定宽度横向排布，总宽度等于 1920px 减去左右内边距与列间隙。

#### Scenario: 宽度验证

- **WHEN** 检查 `client/src/pages/HomePage.tsx`
- **THEN** `LeftPanel` 宽度为 480px
- **AND** `RightPanel` 宽度为 450px
- **AND** `CenterStage` 宽度为 790px
- **AND** 总宽度满足：80 + 480 + 20 + 790 + 20 + 450 + 80 = 1920px

### Requirement: 左侧面板 LeftPanel

LeftPanel SHALL 宽度 480px、高度 500px，内部卡片按固定高度排列，不得溢出容器。

#### Scenario: 左侧面板渲染

- **WHEN** 首页渲染
- **THEN** `LeftPanel` 尺寸为 480×500px，内边距 12px，卡片间隙 12px
- **AND** 月卡 142px、活动中心/热门榜单行 142px、招募入口 168px，卡片总高度 452px（等于可用高度）
- **AND** 聊天按钮以绝对定位显示在面板左下角，不占用流式高度

### Requirement: 中央展示区 CenterStage

CenterStage SHALL 作为大厅核心视觉区，宽度 790px、高度 500px，展示当前出战角色/选手与空位插槽。

#### Scenario: 中央区渲染

- **WHEN** 首页渲染
- **THEN** `CenterStage` 容器尺寸为 790×500px（含上下 12px 内边距后实际舞台 790×476px）
- **AND** 显示标题、当前出战角色大图、右侧空位插槽、底部轮播指示器
- **AND** 背景使用 court-bg 纹理

### Requirement: 右侧面板 RightPanel

RightPanel SHALL 宽度 450px、高度 500px，内部卡片按固定高度排列，不得溢出容器。

#### Scenario: 右侧面板渲染

- **WHEN** 首页渲染
- **THEN** `RightPanel` 尺寸为 450×500px，内边距 12px，卡片间隙 12px
- **AND** 主线任务/日常任务行 70px、竞技赛 128px、商业赛/赛季联赛行 100px、比赛 START 140px，卡片总高度 438px（可用高度 440px，保留 2px 视觉余量）
- **AND** 底部显示显眼的"比赛 START"按钮

### Requirement: 底部导航 BottomNav

BottomNav SHALL 高度固定为 130px，展示游戏主功能导航。

#### Scenario: 底部导航渲染

- **WHEN** 首页渲染
- **THEN** `BottomNav` 高度为 130px
- **AND** 显示 8 个导航项，包含图标、标签、锁定标记与红点提示
- **AND** 图标与文字大小适配 130px 高度

### Requirement: 招募入口与比赛 START 按钮高度以视觉对齐为准

`LeftPanel` 招募入口卡片高度 SHALL 为 168px，`RightPanel` 底部"比赛 START"按钮高度 SHALL 为 140px。即使该数值与理论均分不一致，也以实际视觉效果为准。

#### Scenario: 关键按钮高度验证

- **WHEN** 检查 `LeftPanel.tsx` 招募卡片
- **THEN** 其高度类名为 `h-[168px]`
- **AND** 检查 `RightPanel.tsx` 比赛 START 按钮
- **AND** 其高度类名为 `h-[140px]`

### Requirement: 面板内部卡片比例固定

LeftPanel 与 RightPanel 内部所有卡片 SHALL 使用固定高度，确保比例固定且不溢出容器。

#### Scenario: 卡片高度验证

- **WHEN** 检查 `LeftPanel.tsx` 与 `RightPanel.tsx`
- **THEN** 所有卡片均有明确的 `h-[...px]` 类名
- **AND** 卡片总高度加上内边距与间隙等于或略小于面板高度（视觉对齐优先于理论均分）
- **AND** 不存在 `flex-1` 或 `min-h-0` 等导致比例不确定的类名

### Requirement: 源码无篮球主题命名

所有游戏大厅组件与资源路径 SHALL 不包含 "nba"、"basketball" 等篮球相关命名或文案。

#### Scenario: 源码检查

- **WHEN** 对 `src/components/lobby/`、`src/pages/HomePage.tsx`、`src/index.css` 执行大小写不敏感搜索
- **THEN** 不存在 "nba"、"basketball"、"lineup"（组件名与文件名除外需已重命名）等字符串
- **AND** 第一版允许的占位图片路径 `/images/star-player.png` 可保留

