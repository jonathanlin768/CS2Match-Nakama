## ADDED Requirements

### Requirement: 服务端提供权威联系方式交换收件箱
系统 SHALL 提供 `SocialListContactExchangeInbox`，基于当前正式好友及 canonical pair 交换记录返回收到的 pending、发出的 pending 和需要重新授权的摘要。响应 SHALL 提供 `incoming_pending_count`，且 MUST NOT 包含任何 QQ 或微信正文。

#### Scenario: 查询收到的待处理申请
- **GIVEN** A 向 B 发起了仍有效的 pending 交换请求
- **WHEN** B 调用收件箱 RPC
- **THEN** 响应的 received 列表包含 A、request ID、渠道、version、申请与过期时间
- **AND** `incoming_pending_count` 增加一
- **AND** 响应不包含双方联系方式正文

#### Scenario: 查询发出的待处理申请
- **GIVEN** A 向 B 发起了仍有效的 pending 交换请求
- **WHEN** A 调用收件箱 RPC
- **THEN** 响应的 sent 列表包含 B 和等待处理状态
- **AND** 该请求不计入 A 的 `incoming_pending_count`

#### Scenario: 查询需要重新授权的好友
- **GIVEN** A 与 B 的交换因资料修改变为 stale 或因安全撤销变为 revoked
- **WHEN** 任一方调用收件箱 RPC
- **THEN** 响应返回不含正文的重新授权状态摘要
- **AND** 客户端可以从该摘要进入重新申请流程

#### Scenario: 非好友记录不进入可操作收件箱
- **GIVEN** pair Storage 中存在与当前用户相关但好友关系已失效的旧记录
- **WHEN** 当前用户查询收件箱
- **THEN** 该记录不得作为可接受的 pending 返回
- **AND** 任何返回的历史摘要均不得包含联系方式正文

### Requirement: 收件箱不依赖 DM 卡片恢复
交换请求的成功和可处理性 SHALL 以 canonical Storage 和收件箱 RPC 为准。DM 卡片 SHALL 只触发低延迟刷新；卡片发送失败、Socket 断线或频道历史缺失不得丢失待办。

#### Scenario: 接收者离线时创建请求
- **GIVEN** B 未连接 Socket 或未打开好友页
- **WHEN** A 成功创建交换请求且 DM 卡片未实时送达
- **THEN** 请求仍保存在 canonical Storage
- **AND** B 下次打开好友页时通过收件箱 RPC 看到申请

#### Scenario: DM 卡片投递失败
- **GIVEN** 交换请求 Storage 写入成功但 `ChannelMessageSend` 失败
- **WHEN** 发起者收到 RPC 响应且接收者稍后刷新
- **THEN** 发起者仍看到请求已创建
- **AND** 接收者通过收件箱 RPC 恢复同一请求

#### Scenario: Socket 重连
- **GIVEN** 好友页 Socket 曾断线
- **WHEN** Socket 重连
- **THEN** 客户端重新查询权威收件箱
- **AND** 待处理计数与服务器状态一致

### Requirement: 收件箱支持分页和有界读取
`SocialListContactExchangeInbox` SHALL 对当前正式好友分页，并使用确定性 pair key 批量读取交换记录；接口 SHALL 返回 cursor，避免好友数量增长时执行无界扫描。

#### Scenario: 好友超过单页限制
- **GIVEN** 当前用户的正式好友数量超过请求 limit
- **WHEN** 客户端查询第一页收件箱
- **THEN** 服务端只处理本页好友并返回下一页 cursor
- **AND** 客户端可以继续分页合并待办
