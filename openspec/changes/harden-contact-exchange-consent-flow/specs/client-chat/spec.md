## MODIFIED Requirements

### Requirement: 客户端以结构化类型解析社交卡片
客户端 SHALL 只将受支持版本的服务端结构化消息解析为联系方式交换事件提示。卡片内容 SHALL 只用于触发 `SocialListContactExchangeInbox` 或交换详情 RPC 刷新；未知、畸形或客户端自造消息 SHALL NOT 被渲染为已授权状态，也 SHALL NOT 成为待办存在或授权披露的唯一依据。

#### Scenario: 收到有效请求卡片
- **GIVEN** Socket 收到包含受支持 `type`、`request_id`、`action` 和 `version` 的服务端消息
- **WHEN** 社交事件 hook 处理该消息
- **THEN** 客户端刷新权威收件箱和相关好友交换状态
- **AND** 根据 RPC 结果渲染待接受、已接受、已拒绝、已失效或已撤销状态

#### Scenario: 收到未知消息类型
- **GIVEN** DM 历史或实时事件包含未知 `type`
- **WHEN** 客户端解析消息
- **THEN** 客户端忽略该消息或显示不可操作的兼容提示
- **AND** 不把其正文当作聊天文本或授权内容展示

#### Scenario: 卡片与 RPC 状态冲突
- **GIVEN** 本地卡片显示 requested，但权威 RPC 已返回 stale、declined 或 revoked
- **WHEN** 客户端刷新该交换
- **THEN** 客户端只展示 RPC 状态
- **AND** 清除任何已缓存联系方式正文

### Requirement: 交换卡片历史和实时事件复用共享 Socket
客户端 SHALL 继续复用现有单例 Socket、DM 加入、历史分页、重连和多监听器分发能力，将受支持卡片用于事件历史和低延迟刷新。权威会话状态和全局待处理计数 SHALL 来自收件箱/RPC；重连后 SHALL 重新查询收件箱，卡片加载失败 SHALL 显示可诊断且可重试的非敏感错误而不是静默为空。

#### Scenario: 断线后恢复
- **GIVEN** 用户查看好友交换卡片时 Socket 断线
- **WHEN** 共享 Socket 重连并重新加入 DM 频道
- **THEN** 客户端恢复历史与实时监听
- **AND** 重新查询权威收件箱和可见好友的交换状态

#### Scenario: 非当前好友收到卡片
- **GIVEN** 用户正在查看好友 A
- **WHEN** 好友 B 的频道收到有效交换卡片
- **THEN** 客户端刷新权威收件箱并更新好友 B 的摘要和待处理计数
- **AND** 当前好友 A 的详情不被错误替换

#### Scenario: 卡片历史加载失败
- **GIVEN** 加入 DM 或读取频道历史失败
- **WHEN** 好友页仍能访问 Social RPC
- **THEN** 客户端继续使用权威收件箱呈现申请
- **AND** 对卡片历史显示可重试的非敏感错误状态
