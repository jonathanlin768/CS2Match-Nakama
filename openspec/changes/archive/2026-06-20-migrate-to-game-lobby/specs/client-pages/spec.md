# Client Pages

CS2 模拟器前端页面规格变更 — 首页从响应式仪表盘改为固定画幅游戏大厅。

## MODIFIED Requirements

### Requirement: Home 首页

Home 页面 SHALL 在认证后展示固定 1600×900 的游戏大厅界面，替代原有的响应式仪表盘布局。

#### Scenario: 首页渲染

- **WHEN** 已认证用户访问 `/home`
- **THEN** 页面渲染一个水平垂直居中、固定 1600×900 的游戏大厅画幅
- **AND** 画幅内自上而下显示 TopBar、PromoBar、三列内容区、BottomNav
- **AND** 三列内容区从左到右显示 LeftPanel、CenterStage、RightPanel
- **AND** 所有数据使用硬编码 mock 数据正常渲染

#### Scenario: 画幅裁剪策略

- **WHEN** 浏览器视口小于 1600×900
- **THEN** 大厅画幅保持 1600×900 不变
- **AND** 画幅在视口中居中显示
- **AND** 超出视口的部分被裁剪，不可滚动缩放

## REMOVED Requirements

无。
