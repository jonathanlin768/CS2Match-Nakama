## ADDED Requirements

### Requirement: 玩家联系方式由私有服务端 Storage 管理
系统 SHALL 通过 `SocialSetContactProfile` 保存本人 QQ 和/或微信号。记录 SHALL 使用 `Collection=social_contact_profile`、`Key=profile`、`UserID=本人 UUID`、`PermissionRead=0`、`PermissionWrite=0`；联系方式 Storage MUST 禁止客户端直接读写，服务端日志 MUST NOT 输出联系方式字段正文。

#### Scenario: 设置本人联系方式
- **GIVEN** 正式登录玩家提交至少一个格式和长度合法的联系方式字段
- **WHEN** `SocialSetContactProfile` 完成验证
- **THEN** 服务端将资料写入该玩家私有 `social_contact_profile` 记录
- **AND** 响应只返回渠道是否已配置及脱敏摘要

#### Scenario: 访客尝试设置联系方式
- **GIVEN** 当前 Session 被标记为访客
- **WHEN** 客户端调用 `SocialSetContactProfile`
- **THEN** 服务端拒绝请求并返回需要正式登录的结构化错误

#### Scenario: 客户端直接读取自己的联系方式 Storage
- **GIVEN** 正式玩家是 `social_contact_profile/profile` 的 owner
- **WHEN** 客户端绕过 RPC 调用 Storage read 或 write
- **THEN** Nakama 权限拒绝该直接访问
- **AND** Social RPC 仍可由服务端 runtime 读取该记录

### Requirement: 只有当前好友可以发起交换请求
`SocialRequestContactExchange` SHALL 验证请求者与接收者当前互为 Nakama 好友、请求者已配置所请求渠道、目标不是本人，并对重复请求实施幂等和频率限制。交换记录 SHALL 使用 `Collection=social_contact_exchange`、系统 owner、确定性 pair key、`PermissionRead=0`、`PermissionWrite=0`。

#### Scenario: 向好友发起请求
- **GIVEN** 双方当前为好友且请求者已填写微信号
- **WHEN** 请求者发起交换微信请求
- **THEN** 服务端以双方排序 user ID 生成 pair key 并写入 pending 状态
- **AND** 返回稳定 request ID 和权威 version

#### Scenario: 向非好友发起请求
- **GIVEN** 目标用户不是请求者当前好友
- **WHEN** 请求者调用交换 RPC
- **THEN** 服务端返回 `NOT_FRIENDS`
- **AND** 不写入交换记录或频道卡片

#### Scenario: 重复提交同一请求
- **GIVEN** 同一对好友已经存在相同渠道的 pending 请求
- **WHEN** 请求者使用同一幂等键再次提交
- **THEN** 服务端返回现有请求状态
- **AND** 不创建第二条逻辑请求

### Requirement: 接收者可以接受或拒绝交换
`SocialRespondContactExchange` SHALL 只允许 pending 请求的 recipient 执行 accept 或 decline，并使用 Storage version 防止并发覆盖。接受表示双方对请求渠道互相授权，而不是单方向泄露。

#### Scenario: 接受有效请求
- **GIVEN** 当前用户是 pending 请求的 recipient 且双方仍为好友
- **WHEN** 当前用户选择接受
- **THEN** 服务端将状态原子更新为 accepted
- **AND** 双方后续均可通过读取 RPC 获取已授权渠道

#### Scenario: 非接收者响应
- **GIVEN** 当前用户不是该 pending 请求的 recipient
- **WHEN** 当前用户尝试接受或拒绝
- **THEN** 服务端返回 `FORBIDDEN`
- **AND** 交换状态保持不变

### Requirement: 联系方式只在授权且仍为好友时披露
`SocialGetContactExchange` SHALL 在每次读取时重新验证双方好友关系。只有 accepted 状态下才能返回双方对已授权渠道的实际联系方式；pending、declined、cancelled、失效或已删除好友状态 SHALL NOT 返回实际值。

#### Scenario: 已接受且仍为好友
- **GIVEN** 交换状态为 accepted 且双方仍互为好友
- **WHEN** 任一方读取交换详情
- **THEN** 响应包含双方已授权渠道的实际值
- **AND** 不包含未授权渠道的值

#### Scenario: 接受后删除好友
- **GIVEN** 交换曾被接受但双方已不再是好友
- **WHEN** 任一方再次读取交换详情
- **THEN** 服务端不返回 QQ 或微信正文
- **AND** 返回交换已失效的状态

### Requirement: 私聊频道只允许服务端结构化卡片
服务器 SHALL 注册 realtime before-hook，拒绝客户端 `ChannelMessageSend`、`ChannelMessageUpdate` 和 `ChannelMessageRemove`。交换状态改变后，服务器 MAY 向双方既有 persistent DM 频道发送只包含类型、request ID、动作和 version 的卡片；卡片 MUST NOT 包含 QQ 或微信正文。

#### Scenario: 修改客户端尝试发送文本
- **GIVEN** 用户已建立 Socket 并加入好友 DM 频道
- **WHEN** 客户端直接发送 `{text:"hello"}` 或其他自定义频道内容
- **THEN** before-hook 拒绝 realtime 操作
- **AND** 频道历史中不出现该消息

#### Scenario: 服务端发送请求卡片
- **GIVEN** 交换请求状态已经成功写入 Storage
- **WHEN** Social 服务调用服务端频道消息 API
- **THEN** 好友通过现有 `onchannelmessage` 收到结构化请求卡片
- **AND** 客户端使用 request ID 调用 RPC 获取权威状态

#### Scenario: 卡片投递失败
- **GIVEN** Storage 已写入交换状态但频道发送暂时失败
- **WHEN** 任一方刷新好友详情
- **THEN** 客户端仍能通过 RPC 恢复权威交换状态
- **AND** 不将频道消息是否成功作为授权依据

### Requirement: 玩家公共标识不接受自由文本昵称
系统 SHALL 将恰好 8 位大写英文字母与数字组成的 code 保存为 Nakama 唯一 username，并以 `玩家#<CODE>` 展示。客户端 SHALL NOT 提供修改公开昵称的自由文本入口；好友查找 SHALL 直接以该 code 查询唯一 Nakama username。完整 UUID user ID SHALL 继续作为内部权限与 Storage owner 键。

#### Scenario: 新账号获得公共标识
- **GIVEN** 新设备账号或正式账号尚无规范公共代码
- **WHEN** 服务端初始化玩家社交资料
- **THEN** 服务端生成 8 位大写字母数字唯一代码并在 username 冲突时重试
- **AND** UI 以 `玩家#CODE` 展示该账号

#### Scenario: 修改客户端提交自定义 username
- **GIVEN** 客户端直接提交不符合 8 位大写字母数字格式的 username 更新
- **WHEN** Nakama update-account before-hook 校验请求
- **THEN** 服务端拒绝修改
- **AND** 原 username 保持不变
