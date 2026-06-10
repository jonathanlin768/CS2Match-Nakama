# Client Pages (Delta)

ProfilePage 好友部分从 mock 数据切换到 Nakama 真实好友 API。

## ADDED Requirements

### Requirement: ProfilePage 好友列表渲染

ProfilePage 的好友部分 SHALL 使用 Nakama `listFriends` API 返回的真实好友数据渲染好友列表，替代硬编码的 `mockFriends` 数据。

#### Scenario: 好友列表加载中

- **WHEN** 用户切换到"我的好友"菜单
- **AND** `useFriends` hook 正在从 Nakama 获取好友数据（status="loading"）
- **THEN** 好友列表区域显示加载指示器（spinner）
- **AND** 好友搜索框和添加按钮仍然可见

#### Scenario: 好友列表渲染（有好友）

- **WHEN** 用户切换到"我的好友"菜单
- **AND** `useFriends` hook 成功返回好友列表（status="success"）
- **THEN** 左侧好友列表按分组展示（我的好友、已发送请求、收到的请求）
- **AND** 每个分组显示可折叠标题（含在线/总数统计）
- **AND** "收到的请求"分组标题旁显示红色数字徽章（未处理请求数量）
- **AND** 没有成员的分组自动隐藏（如"已发送请求"为空时不显示该分组）
- **AND** 每个好友项显示用户头像（avatar_url 或首字母 Fallback）、用户名、个性签名（display_name 或 `-`）
- **AND** 在线状态指示器根据 `online` 字段显示绿色（在线）或灰色（离线）

#### Scenario: 好友列表为空

- **WHEN** 用户切换到"我的好友"菜单
- **AND** `useFriends` hook 返回空列表（status="success", friends=[]）
- **THEN** 左侧好友列表显示 Empty 组件："暂无好友"
- **AND** 搜索框和添加按钮仍然可见

#### Scenario: 好友列表加载失败

- **WHEN** 用户切换到"我的好友"菜单
- **AND** Nakama API 返回错误（status="error"）
- **THEN** 好友列表区域显示错误提示："加载失败，请稍后重试"
- **AND** 提供"重试"按钮重新加载

#### Scenario: 点击好友查看详情

- **GIVEN** 好友列表已加载
- **WHEN** 用户点击某个好友
- **THEN** `selectedFriendId` 更新为该好友的 `id`
- **AND** 右侧详情面板显示该好友的详细信息：
  - 头像（avatar_url）、用户名（username）、在线状态（online）
  - 用户 ID（id）、所在地（location 或 `-`）
  - 不可用字段（signature、level、age、birthday、likes、gender）显示为 `-` 或隐藏
- **AND** 选中的好友项高亮（`bg-primary/10`）

#### Scenario: 搜索过滤好友

- **GIVEN** 好友列表已加载
- **WHEN** 用户在搜索框中输入关键词
- **THEN** 好友列表实时过滤，仅显示用户名或 displayName 匹配关键词的好友（大小写不敏感）
- **AND** 分组标题在过滤状态下根据匹配结果动态调整

### Requirement: ProfilePage 添加好友

ProfilePage SHALL 提供添加好友功能，用户通过输入用户名发送好友请求。

#### Scenario: 点击添加按钮

- **WHEN** 用户点击好友列表顶部的 "+"(Plus) 按钮
- **THEN** 弹出一个添加好友对话框（Dialog 组件）
- **AND** 对话框包含：标题"添加好友"、用户名输入框、"添加"按钮、"取消"按钮

#### Scenario: 成功发送好友请求

- **GIVEN** 添加好友对话框已打开
- **WHEN** 用户输入一个存在的用户名并点击"添加"
- **THEN** 调用 `useFriends.addFriend(username)`
- **AND** 按钮显示加载状态（spinner + "发送中..."）
- **AND** 成功后对话框关闭
- **AND** 好友列表自动刷新，显示新的"已发送请求"条目

#### Scenario: 用户不存在

- **GIVEN** 添加好友对话框已打开
- **WHEN** 用户输入一个不存在的用户名并点击"添加"
- **THEN** 对话框内显示错误提示："用户不存在，请检查用户名"
- **AND** 对话框保持打开，允许用户修改并重试

#### Scenario: 添加自己为好友

- **GIVEN** 添加好友对话框已打开
- **WHEN** 用户输入自己的用户名并点击"添加"
- **THEN** 对话框内显示错误提示："无法添加自己为好友"
- **AND** 不发送 API 请求（前端校验）

#### Scenario: 重复添加

- **GIVEN** 添加好友对话框已打开
- **WHEN** 用户输入一个已是好友的用户名并点击"添加"
- **THEN** 对话框内显示错误提示："该用户已是您的好友"
- **AND** 对话框保持打开

### Requirement: ProfilePage 动态操作按钮

ProfilePage 的好友详情面板底部操作按钮 SHALL 根据选中好友的 `state` 动态切换，替代当前固定显示的"编辑资料"和"发消息"按钮。

#### Scenario: FRIEND 状态按钮

- **GIVEN** 用户选中了一个 state=FRIEND(0) 的好友
- **WHEN** 右侧详情面板渲染
- **THEN** 底部显示"删除好友"按钮（危险/红色样式）
- **AND** 不显示 mock 的"编辑资料"和"发消息"按钮

#### Scenario: INVITE_SENT 状态按钮

- **GIVEN** 用户选中了一个 state=INVITE_SENT(1) 的好友（自己发出的请求）
- **WHEN** 右侧详情面板渲染
- **THEN** 底部显示"取消请求"按钮（次要/outline 样式）
- **AND** 不显示其他操作按钮

#### Scenario: INVITE_RECEIVED 状态按钮

- **GIVEN** 用户选中了一个 state=INVITE_RECEIVED(2) 的好友（收到的请求）
- **WHEN** 右侧详情面板渲染
- **THEN** 底部显示"接受"按钮（主色/强调样式）和"拒绝"按钮（次要/outline 样式）
- **AND** 不显示其他操作按钮

### Requirement: ProfilePage 好友请求处理（接受/拒绝/取消）

ProfilePage SHALL 支持对好友请求的接受、拒绝和取消操作。

#### Scenario: 接受好友请求

- **GIVEN** 用户选中了一个 state=INVITE_RECEIVED 的好友
- **WHEN** 用户点击"接受"按钮
- **THEN** 调用 `useFriends.addFriend(senderUsername)` 建立双向好友关系
- **AND** 按钮显示加载状态
- **AND** 成功后该好友移至"我的好友"分组（state 变为 FRIEND）
- **AND** 详情面板按钮更新为 FRIEND 状态的按钮
- **AND** 显示 Toast："已接受 [username] 的好友请求"

#### Scenario: 拒绝好友请求

- **GIVEN** 用户选中了一个 state=INVITE_RECEIVED 的好友
- **WHEN** 用户点击"拒绝"按钮
- **THEN** 弹出确认对话框："确定要拒绝 [username] 的好友请求吗？"
- **AND** 用户确认后调用 `useFriends.removeFriend(userId)` 删除待处理请求
- **AND** 成功后该好友从列表中移除
- **AND** 右侧详情面板重置为空状态
- **AND** 显示 Toast："已拒绝 [username] 的好友请求"

#### Scenario: 取消已发送请求

- **GIVEN** 用户选中了一个 state=INVITE_SENT 的好友
- **WHEN** 用户点击"取消请求"按钮
- **THEN** 弹出确认对话框："确定要取消发送给 [username] 的好友请求吗？"
- **AND** 用户确认后调用 `useFriends.removeFriend(userId)` 取消请求
- **AND** 成功后该好友从列表中移除
- **AND** 右侧详情面板重置为空状态

#### Scenario: 删除好友

- **GIVEN** 用户选中了一个好友（state=FRIEND）
- **WHEN** 用户点击"删除好友"按钮
- **THEN** 弹出确认对话框："确定要删除好友 [username] 吗？"
- **AND** 用户确认后调用 `useFriends.removeFriend(userId)`
- **AND** 成功后好友列表自动刷新
- **AND** 右侧详情面板重置为空状态
- **AND** 显示 Toast："已删除好友 [username]"

### Requirement: ProfilePage 请求通知徽章

"收到的请求"分组标题 SHALL 显示红色数字徽章表示待处理请求数量。

#### Scenario: 有待处理请求时显示徽章

- **GIVEN** 好友列表中存在 state=INVITE_RECEIVED 的记录
- **WHEN** 好友列表渲染
- **THEN** "收到的请求"分组标题旁显示红色 `Badge` 组件，数字为请求数量
- **AND** 当请求数量为 0 时该分组自动隐藏

#### Scenario: 接受/拒绝后徽章更新

- **GIVEN** "收到的请求"徽章显示为 3
- **WHEN** 用户接受或拒绝其中 1 个请求
- **THEN** 好友列表刷新后徽章数字更新为 2

## MODIFIED Requirements

_（无现有 ProfilePage 需求需要修改 — ProfilePage 此前未被 spec 覆盖）_

## REMOVED Requirements

_（无废弃需求 — `mockFriends` 和 `friendGroups` 是硬编码数据，不是 spec 级功能）_
