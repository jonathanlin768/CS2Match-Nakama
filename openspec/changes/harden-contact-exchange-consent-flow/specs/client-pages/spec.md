## MODIFIED Requirements

### Requirement: ProfilePage 好友列表渲染
好友页面 SHALL 复用 Nakama `listFriends` 的真实好友、已发送申请和收到申请状态，并在桌面和移动端以新 App Shell 展示。好友详情 SHALL 提供好友对战状态入口和联系方式交换入口，但只有正式好友可发起交换；页面 SHALL 在普通好友列表之前显示权威“联系方式申请”分组。

#### Scenario: 查看正式好友
- **GIVEN** `useFriends` 已加载 state=FRIEND 的好友
- **WHEN** 用户选择该好友
- **THEN** 详情显示系统玩家标识、在线状态、好友对战入口状态和权威联系方式交换状态
- **AND** 不显示自由文本聊天输入框

#### Scenario: 查看待处理好友请求
- **GIVEN** 选中记录是 INVITE_SENT 或 INVITE_RECEIVED
- **WHEN** 详情面板渲染
- **THEN** 页面显示对应接受、拒绝或取消好友关系操作
- **AND** 不允许发起联系方式交换

#### Scenario: 查看收到的联系方式申请
- **GIVEN** 权威收件箱包含收到的 pending 交换申请
- **WHEN** 好友页面渲染
- **THEN** “联系方式申请”分组醒目显示申请者、渠道、时间和过期状态
- **AND** 当前用户可以直接接受或拒绝

#### Scenario: 查看需要重新授权的交换
- **GIVEN** 权威交换状态为 stale 或 revoked
- **WHEN** 用户打开对应好友的联系方式详情
- **THEN** 页面不显示此前缓存的 QQ 或微信正文
- **AND** 显示“联系方式已变化，需要重新授权”及重新申请入口

### Requirement: MessagesTab 聊天 UI 结构
原 MessagesTab SHALL 迁移为好友事件与联系方式交换卡片视图。桌面端 MAY 使用联系方式申请/好友列表与详情双栏，移动端 SHALL 使用列表到详情的层级导航；权威待办 SHALL 优先于 DM 事件历史呈现，任何视口都 SHALL NOT 渲染自由文本、多媒体或链接消息输入工具。

#### Scenario: 桌面端交换视图
- **GIVEN** 视口宽度不小于 1024px
- **WHEN** 交换视图渲染
- **THEN** 左侧优先显示权威联系方式申请和好友状态，右侧显示选中交换的权威状态、结构化事件和操作
- **AND** 不显示文本输入框

#### Scenario: 移动端交换视图
- **GIVEN** 视口宽度小于 1024px
- **WHEN** 用户从联系方式申请或好友列表打开交换详情
- **THEN** 页面以单栏详情呈现并提供返回列表操作
- **AND** 接受、拒绝和重新申请入口无需依赖卡片历史才能访问

## ADDED Requirements

### Requirement: 本人联系方式在刷新后安全回填
好友页 SHALL 通过 `SocialGetContactProfile` 获取当前登录用户的完整资料并回填 QQ/微信输入框，不得从 localStorage 或直接 Storage 读取联系方式正文。有效修改前 SHALL 告知用户整次既有授权会失效。

#### Scenario: 刷新好友页
- **GIVEN** 当前正式玩家已保存 QQ 和微信
- **WHEN** 玩家刷新或重新进入好友页
- **THEN** 输入框回填服务端返回的本人 QQ 和微信
- **AND** 页面同时显示脱敏保存摘要

#### Scenario: 保存实际修改
- **GIVEN** 玩家编辑后的标准化资料与服务端现值不同
- **WHEN** 玩家点击保存
- **THEN** 页面提示修改会使所有既有联系方式交换授权失效
- **AND** 保存成功后使用服务端规范化结果更新输入框和摘要

#### Scenario: 读取本人资料失败
- **GIVEN** `SocialGetContactProfile` 请求失败
- **WHEN** 好友页加载资料编辑区
- **THEN** 页面显示可重试错误且不以空表单暗示资料不存在
- **AND** 不覆盖用户尚未提交的本地输入

### Requirement: 好友入口显示权威待处理计数
App Shell 的好友入口 SHALL 显示 `SocialListContactExchangeInbox` 返回的收到 pending 数量。计数 SHALL 在登录、页面刷新、Socket 重连、有效卡片到达和处理申请后刷新。

#### Scenario: 在其他页面收到交换申请
- **GIVEN** 用户已正式登录且不在好友页
- **WHEN** 权威收件箱出现一条收到的 pending 申请并触发刷新
- **THEN** App Shell 好友入口显示待处理数量
- **AND** 点击好友入口进入可直接处理该申请的分组

#### Scenario: 接受或拒绝后更新计数
- **GIVEN** 好友入口显示一条待处理申请
- **WHEN** 用户成功接受或拒绝该申请
- **THEN** 客户端重新查询权威收件箱
- **AND** 待处理计数立即减少

### Requirement: Social RPC 错误使用可理解提示
客户端 SHALL 将 Social RPC 的结构化 code、HTTP/SDK Response 和网络异常转换为可操作的安全文案，不得直接渲染对象字符串或联系方式 payload。

#### Scenario: 参与者未配置任何渠道
- **GIVEN** 发起方或接受方的 QQ 与微信均为空
- **WHEN** 用户发起或接受交换且服务端返回 `PROFILE_INCOMPLETE`
- **THEN** 页面提示 QQ 或微信至少保存一种即可
- **AND** 不显示 `[object Response]`

#### Scenario: 接受方只保存一个不同渠道
- **GIVEN** 发起方保存 QQ 和微信，接受方只保存 QQ；或双方分别只保存不同渠道
- **WHEN** 接受方点击“接受并互相授权”
- **THEN** 页面允许完成交换，不提示补齐另一渠道
- **AND** 详情将对方未提供的渠道显示为“未提供”

#### Scenario: 并发状态冲突
- **GIVEN** 用户处理的申请已在另一设备发生变化
- **WHEN** 服务端返回 `CONFLICT` 或 `INVALID_STATE`
- **THEN** 页面提示状态已变化并刷新权威详情
- **AND** 不重复提交响应
