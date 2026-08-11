## ADDED Requirements

### Requirement: TutorialPage 使用完整选手卡面进行阵容选择

TutorialPage SHALL 在候选选手区域展示未裁切的完整 `2:3` 卡面，同时保留当前费用、阵容名额和选择规则。

#### Scenario: 展示候选选手完整卡面

- **GIVEN** 候选 Player 配置了可读取的 `cardImage`
- **WHEN** 用户进入新手战斗选人页面
- **THEN** 每个候选项以稳定的 `2:3` 画幅完整展示卡面
- **AND** 卡面不使用战斗头像的 `5:7` 裁切参数
- **AND** 页面仍清晰显示选手名、战队、价格和选中状态

#### Scenario: 使用卡面选择选手

- **GIVEN** 用户正在浏览某个费用档的候选选手
- **WHEN** 用户点击完整卡面或其选择控件
- **THEN** 页面沿用现有预算、人数、费用档和重复选择约束
- **AND** 卡面展示不会改变 TutorialBattle 的业务校验规则

#### Scenario: 完整卡面不可用时显示回退资源

- **GIVEN** 候选 Player 未配置有效 `cardImage`
- **WHEN** TutorialPage 渲染该候选项
- **THEN** 页面回退显示现有 `portrait`
- **AND** `portrait` 也不可用时显示默认选手图
- **AND** 用户仍可识别并选择该候选选手

#### Scenario: 不同视口下保持完整卡面可用

- **GIVEN** TutorialPage 在桌面端或移动端显示
- **WHEN** 候选卡片根据可用宽度重新排布
- **THEN** 卡面容器始终保持 `2:3` 比例
- **AND** 桌面端候选区每行固定显示 5 名选手，移动端每行显示 2 名选手
- **AND** 每张卡片通过价格角标表达所属费用档，不依赖独占整行的费用档标题
- **AND** 页面内容超过可用高度时由主内容区显示纵向滚动条并允许滚动到底部
- **AND** 图片、名称、价格和选择状态不互相遮挡
