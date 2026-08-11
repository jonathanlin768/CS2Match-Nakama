## ADDED Requirements

### Requirement: 客户端以结构化类型解析社交卡片
客户端 SHALL 只将受支持版本的服务端结构化消息解析为联系方式交换卡片。卡片内容 SHALL 以 `request_id` 查询 Social RPC 获得的权威状态为准；未知、畸形或客户端自造消息 SHALL NOT 被渲染为已授权状态。

#### Scenario: 收到有效请求卡片
- **GIVEN** Socket 收到包含受支持 `type`、`request_id`、`action` 和 `version` 的服务端消息
- **WHEN** 社交事件 hook 处理该消息
- **THEN** 客户端按 request ID 获取最新交换状态
- **AND** 渲染待接受、已接受、已拒绝或已失效卡片

#### Scenario: 收到未知消息类型
- **GIVEN** DM 历史或实时事件包含未知 `type`
- **WHEN** 客户端解析消息
- **THEN** 客户端忽略该消息或显示不可操作的兼容提示
- **AND** 不把其正文当作聊天文本展示

### Requirement: 交换卡片历史和实时事件复用共享 Socket
客户端 SHALL 继续复用现有单例 Socket、DM 加入、历史分页、重连和多监听器分发能力，但会话摘要和未读状态 SHALL 从受支持的结构化卡片派生。重连后 SHALL 通过 RPC 刷新卡片状态，不能仅相信本地缓存。

#### Scenario: 断线后恢复
- **GIVEN** 用户查看好友交换卡片时 Socket 断线
- **WHEN** 共享 Socket 重连并重新加入 DM 频道
- **THEN** 客户端恢复历史与实时监听
- **AND** 对可见 request ID 重新读取权威交换状态

#### Scenario: 非当前好友收到卡片
- **GIVEN** 用户正在查看好友 A
- **WHEN** 好友 B 的频道收到有效交换卡片
- **THEN** 好友 B 的事件摘要和未读数更新
- **AND** 当前好友 A 的详情不被错误替换

### Requirement: 客户端不暴露任意频道写入操作
生产客户端 SHALL 不导出或调用面向 UI 的 `writeChatMessage(text)` 能力，也 SHALL NOT 渲染文本、多媒体或自定义 payload 输入控件。发起和响应交换只能调用 Social RPC。

#### Scenario: 用户发起交换
- **GIVEN** 当前选中对象是正式好友
- **WHEN** 用户点击“请求交换联系方式”并确认渠道
- **THEN** 客户端调用 `SocialRequestContactExchange`
- **AND** 不调用 Socket `writeChatMessage`

#### Scenario: 服务端拒绝交换
- **GIVEN** 好友关系已经失效但客户端尚未刷新
- **WHEN** Social RPC 返回 `NOT_FRIENDS`
- **THEN** 卡片显示请求不可用并刷新好友状态
- **AND** 客户端不回退到发送普通频道消息

## REMOVED Requirements

### Requirement: writeChatMessage API 封装
**Reason**: 任意客户端消息无法满足不接入敏感词审核的产品与安全目标。
**Migration**: UI 动作改为调用 Social RPC；频道仅接收服务端生成卡片。

### Requirement: useFriendDM Hook
**Reason**: 现有 hook 的核心契约包含发送和展示任意文本消息。
**Migration**: 拆分/替换为复用 Socket 与频道历史的结构化社交事件 hook，不提供 `sendMessage(text)`。
