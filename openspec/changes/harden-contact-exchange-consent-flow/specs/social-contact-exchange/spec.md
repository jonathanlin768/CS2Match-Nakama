## MODIFIED Requirements

### Requirement: 玩家联系方式由私有服务端 Storage 管理
系统 SHALL 通过 `SocialSetContactProfile` 保存本人 QQ 和/或微信号，并通过不接受目标 user ID 的 `SocialGetContactProfile` 向本人返回完整资料。记录 SHALL 使用 `Collection=social_contact_profile`、`Key=profile`、`UserID=本人 UUID`、`PermissionRead=0`、`PermissionWrite=0`，并维护覆盖整个资料的单调递增 `revision`；联系方式 Storage MUST 禁止客户端直接读写，服务端日志 MUST NOT 输出联系方式字段正文。

#### Scenario: 设置本人联系方式
- **GIVEN** 正式登录玩家提交格式和长度合法的联系方式字段
- **WHEN** `SocialSetContactProfile` 完成标准化和验证
- **THEN** 服务端将资料写入该玩家私有 `social_contact_profile` 记录
- **AND** 响应返回服务端规范化后的本人资料、脱敏摘要和 revision

#### Scenario: 只设置一种联系方式
- **GIVEN** 正式玩家只填写合法 QQ 或只填写合法微信号，另一字段为空
- **WHEN** `SocialSetContactProfile` 保存资料
- **THEN** 服务端成功保存已填写的单一渠道，不要求同时填写另一渠道
- **AND** 客户端发起交换时提交本人当前已保存的渠道

#### Scenario: 刷新后读取本人联系方式
- **GIVEN** 正式玩家此前已经保存 QQ 和/或微信号
- **WHEN** 本人调用不包含目标 user ID 的 `SocialGetContactProfile`
- **THEN** 服务端只按 Session 中的本人 UUID 返回完整资料和 revision
- **AND** 调用者不能用该 RPC 查询其他用户

#### Scenario: 保存相同资料
- **GIVEN** 玩家提交的 QQ 和微信经标准化后与已保存值完全相同
- **WHEN** `SocialSetContactProfile` 保存资料
- **THEN** 服务端保持原 revision
- **AND** 既有联系方式交换授权不因此失效

#### Scenario: 修改或清空任一字段
- **GIVEN** 玩家已经保存联系方式
- **WHEN** 玩家实际修改、新增或清空 QQ 或微信中的任一字段
- **THEN** 服务端将整个 profile revision 增加一次
- **AND** 允许玩家清空一个或全部渠道

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
`SocialRequestContactExchange` SHALL 验证请求者与接收者当前互为 Nakama 好友、请求者至少已配置一种联系方式、目标不是本人，并对重复请求实施幂等和频率限制。请求 SHALL 由服务端从请求者资料派生当前全部非空渠道并记录请求者当前完整 profile revision；交换记录 SHALL 使用 `Collection=social_contact_exchange`、系统 owner、确定性 pair key、`PermissionRead=0`、`PermissionWrite=0`。

#### Scenario: 向好友发起请求
- **GIVEN** 双方当前为好友且请求者至少已填写一种联系方式
- **WHEN** 请求者发起双向联系方式交换
- **THEN** 服务端以双方排序 user ID 生成 pair key 并写入 pending 状态
- **AND** pending 渠道为请求者资料中当前全部非空渠道，不信任客户端省略已保存渠道
- **AND** 记录请求者当前 profile revision
- **AND** 返回稳定 request ID 和权威 version

#### Scenario: 向非好友发起请求
- **GIVEN** 目标用户不是请求者当前好友
- **WHEN** 请求者调用交换 RPC
- **THEN** 服务端返回 `NOT_FRIENDS`
- **AND** 不写入交换记录或频道卡片

#### Scenario: 重复提交同一请求
- **GIVEN** 同一对好友已经存在相同渠道和请求者 revision 的 pending 请求
- **WHEN** 请求者使用同一幂等键再次提交
- **THEN** 服务端返回现有请求状态
- **AND** 不创建第二条逻辑请求或重复待办

#### Scenario: 请求期间请求者修改资料
- **GIVEN** A 发起请求后修改了 QQ 或微信中的任一字段
- **WHEN** B 尝试处理原请求
- **THEN** 服务端拒绝接受并将请求标记为 stale
- **AND** A 必须基于新 profile revision 重新发起请求

### Requirement: 接收者可以接受或拒绝交换
`SocialRespondContactExchange` SHALL 只允许 pending 请求的 recipient 执行 accept 或 decline，并使用 Storage version 防止并发覆盖。接受表示双方互相授权各自当前资料中的全部非空联系方式，不要求双方拥有相同渠道；服务端 SHALL 在接受事务中验证好友关系、请求者 revision 和双方各自至少一个非空渠道，把双方渠道并集写入交换记录，并记录双方接受时的完整 profile revision。

#### Scenario: 接受有效请求
- **GIVEN** 当前用户是 pending 请求的 recipient、双方仍为好友、请求者 revision 未变化且双方各自至少保存一种联系方式
- **WHEN** 当前用户选择接受
- **THEN** 服务端将状态原子更新为 accepted
- **AND** 将双方当前全部非空渠道的并集记录为本次授权渠道
- **AND** 记录双方当前完整 profile revision
- **AND** 双方后续均可通过读取 RPC 获取对方实际提供的已授权渠道

#### Scenario: 双方只保存不同的单一渠道
- **GIVEN** A 只保存 QQ、B 只保存微信且 A 向 B 发起交换
- **WHEN** B 选择接受并互相授权
- **THEN** 服务端接受该请求，不要求 B 也保存 QQ
- **AND** A 只能获得 B 的微信，B 只能获得 A 的 QQ

#### Scenario: 非接收者响应
- **GIVEN** 当前用户不是该 pending 请求的 recipient
- **WHEN** 当前用户尝试接受或拒绝
- **THEN** 服务端返回 `FORBIDDEN`
- **AND** 交换状态保持不变

#### Scenario: 两台设备并发响应
- **GIVEN** B 在两台设备上同时打开同一 pending 请求
- **WHEN** 两台设备分别提交接受和拒绝
- **THEN** 只有第一个满足 Storage version 的响应成功
- **AND** 第二个响应返回 `CONFLICT` 或 `INVALID_STATE` 并刷新权威状态

### Requirement: 联系方式只在授权且仍为好友时披露
`SocialGetContactExchange` SHALL 在每次读取时重新验证双方好友关系，并比较双方当前完整 profile revision 与接受时 revision。只有 accepted 且两个 revision 都一致时才能按接受时双方渠道并集返回双方各自实际存在的联系方式；pending、declined、cancelled、expired、stale、revoked、已删除好友或旧版缺少接受 revision 的记录 SHALL NOT 返回任何联系方式正文。

#### Scenario: 已接受且版本未变
- **GIVEN** 交换状态为 accepted、双方仍互为好友且双方 profile revision 均与接受时一致
- **WHEN** 任一方读取交换详情
- **THEN** 响应包含对方在接受时授权且当前 revision 仍匹配的实际值
- **AND** 对方没有保存的渠道保持为空，不阻止整次交换

#### Scenario: 任一联系方式发生变化
- **GIVEN** 交换曾被接受
- **WHEN** 任一方实际修改、新增或清空 QQ 或微信中的任一字段
- **THEN** 下次读取返回 stale
- **AND** 响应不包含双方任何 QQ 或微信正文
- **AND** 双方必须完成新的请求和接受才能再次读取

#### Scenario: 接受后删除好友
- **GIVEN** 交换曾被接受但双方已不再是好友
- **WHEN** 任一方读取交换详情
- **THEN** 服务端不返回 QQ 或微信正文
- **AND** 返回 revoked 状态

#### Scenario: 旧版 accepted 记录缺少资料版本
- **GIVEN** 部署前 accepted 记录没有双方接受时 profile revision
- **WHEN** 任一方在升级后读取交换详情
- **THEN** 服务端按 stale 处理
- **AND** 不推断或迁移为有效授权

## ADDED Requirements

### Requirement: 删除好友永久撤销交换
服务器 SHALL 在 Nakama 好友删除操作执行前持久撤销该好友对的 pending 请求和 accepted 授权；撤销失败 MUST 阻止好友删除。重新建立好友关系 SHALL NOT 恢复旧交换。

#### Scenario: 删除已有授权的好友
- **GIVEN** A 与 B 有 accepted 交换
- **WHEN** 任一方通过 ID 或 username 删除好友
- **THEN** before-delete hook 将 pair 记录写为 revoked 并清除有效 revision 绑定
- **AND** Nakama 仅在撤销成功后继续删除好友关系

#### Scenario: 删除后重新添加好友
- **GIVEN** A 与 B 的旧交换已因删除好友变为 revoked
- **WHEN** 双方重新成为好友并打开联系方式详情
- **THEN** 旧授权保持不可用且不返回正文
- **AND** 双方必须重新申请和接受

#### Scenario: 撤销写入失败
- **GIVEN** 删除好友前 Storage 撤销写入失败
- **WHEN** before-delete hook 处理删除请求
- **THEN** hook 拒绝删除并返回结构化错误
- **AND** 不产生已删除好友但旧授权未撤销的状态
