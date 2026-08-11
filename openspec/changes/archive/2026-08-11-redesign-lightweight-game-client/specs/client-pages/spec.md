## MODIFIED Requirements

### Requirement: Home 首页
系统 SHALL 提供响应式 Home 主界面，一级内容只包含调整阵容、好友对战、匹配对战。调整阵容与好友对战 SHALL 显示未开放，匹配对战 SHALL 可进入电脑模拟。Home SHALL 承载首次体验 Modal，但主界面本身不得依赖 Modal 才可操作。

#### Scenario: 首页渲染
- **GIVEN** 身份初始化成功
- **WHEN** 用户访问 `/`
- **THEN** 页面显示三个核心入口及低层级账号状态
- **AND** 调整阵容与好友对战标识为未开放，匹配对战可操作
- **AND** 不显示旧管理手游仪表盘、资产面板或活动入口

#### Scenario: 窄屏渲染
- **GIVEN** 用户使用手机视口访问 `/`
- **WHEN** 页面渲染
- **THEN** 三个入口无需缩放页面文字即可阅读和触控

### Requirement: MatchPage 对战页面
Match 页面 SHALL 作为对战模式选择与请求状态页面。第一阶段匹配对战 SHALL 明确提供“电脑对战”并调用现有模拟 RPC；“路人匹配”和权威“好友对战”若尚未实现 SHALL 显示真实的未开放状态，不得使用假倒计时冒充真实匹配。

#### Scenario: 选择电脑对战
- **GIVEN** 用户拥有有效 Guest 或正式 Session
- **WHEN** 用户选择电脑对战并确认开始
- **THEN** 页面显示 RPC 请求阶段反馈
- **AND** 成功收到战报后导航到 Battle 页面

#### Scenario: 选择未实现的路人匹配
- **GIVEN** 本期未注册真实 Matchmaker
- **WHEN** 用户查看路人匹配选项
- **THEN** 页面将其标识为尚未开放或后续能力
- **AND** 不显示伪造的匹配玩家结果

### Requirement: ProfilePage 好友列表渲染
好友页面 SHALL 复用 Nakama `listFriends` 的真实好友、已发送申请和收到申请状态，并在桌面和移动端以新 App Shell 展示。好友详情 SHALL 提供好友对战状态入口和联系方式交换入口，但只有正式好友可发起交换。

#### Scenario: 查看正式好友
- **GIVEN** `useFriends` 已加载 state=FRIEND 的好友
- **WHEN** 用户选择该好友
- **THEN** 详情显示系统玩家标识、在线状态、好友对战入口状态和联系方式交换状态
- **AND** 不显示自由文本聊天输入框

#### Scenario: 查看待处理好友请求
- **GIVEN** 选中记录是 INVITE_SENT 或 INVITE_RECEIVED
- **WHEN** 详情面板渲染
- **THEN** 页面显示对应接受、拒绝或取消操作
- **AND** 不允许发起联系方式交换

### Requirement: MessagesTab 聊天 UI 结构
原 MessagesTab SHALL 迁移为好友事件与联系方式交换卡片视图。桌面端 MAY 使用好友列表与详情双栏，移动端 SHALL 使用列表到详情的层级导航；任何视口都 SHALL NOT 渲染自由文本、多媒体或链接消息输入工具。

#### Scenario: 桌面端交换视图
- **GIVEN** 视口宽度不小于 1024px 且选中一名正式好友
- **WHEN** 交换视图渲染
- **THEN** 左侧显示好友/事件列表，右侧显示结构化交换卡片和操作
- **AND** 不显示文本输入框

#### Scenario: 移动端交换视图
- **GIVEN** 视口宽度小于 1024px
- **WHEN** 用户从好友列表打开交换详情
- **THEN** 页面以单栏详情呈现并提供返回好友列表操作

## REMOVED Requirements

### Requirement: GachaPage 抽卡页面
**Reason**: 本期不建设抽卡和重型奖励生态。
**Migration**: 旧入口重定向到主界面，不迁移抽卡状态。

### Requirement: RankingPage 排行页面
**Reason**: 排行榜不属于三入口 MVP。
**Migration**: 从导航移除，未来按独立变更恢复。

### Requirement: MessagesTab 发送文本消息
**Reason**: 任意文本需要内容审核并可能泄露不受控信息。
**Migration**: 使用服务端权威的联系方式交换请求卡片。
