# Client Pages

CS2 模拟器前端页面规格 — 登录页、首页仪表盘、对战页面、抽卡系统、排行榜、个人中心。

## Purpose

定义核心页面的 UI 结构和行为。LoginPage 和 ProfilePage 好友部分已接入 Nakama 真实 API，其余页面仍使用硬编码 mock。

## Requirements

### Requirement: LoginPage 登录/注册页面

系统 SHALL 提供 LoginPage，包含登录/注册页签切换、邮箱/密码输入框、社交登录入口。页面加载时先尝试从 localStorage 恢复 Nakama Session，成功则自动跳转到 `/home`。恢复失败则根据页签展示登录或注册表单。登录调用 `loginWithEmail`（`create=false`），注册调用 `registerWithEmail`（`create=true`）。认证失败时显示错误提示，不跳转。

#### Scenario: Session 恢复成功自动跳转

- **WHEN** 用户访问 `/` 且 localStorage 中存在有效 token
- **AND** `restoreSession()` 返回有效 Session
- **THEN** 页面自动跳转到 `/home`（不显示登录/注册表单）

#### Scenario: Session 恢复中显示加载状态

- **WHEN** 用户访问 `/` 且系统正在尝试恢复 Session
- **THEN** 页面显示加载指示器（spinner + "正在恢复登录..."）
- **AND** 不显示登录/注册表单

#### Scenario: 渲染登录表单

- **WHEN** 用户访问 `/` 且无法从 localStorage 恢复有效 Session
- **AND** 当前页签为"登录"
- **THEN** 页面显示 CS2 SIMULATOR 品牌标识
- **AND** 显示邮箱输入框和密码输入框
- **AND** 显示"登录"按钮、"记住我"复选框和"忘记密码"链接

#### Scenario: 点击登录按钮

- **WHEN** 用户在登录页签下点击"登录"按钮
- **THEN** 按钮显示加载状态（spinner + "登录中..."）
- **AND** 前端调用 `loginWithEmail(email, password)` 发起 Nakama 认证
- **AND** 认证成功后导航到 `/home`

#### Scenario: 登录失败显示错误

- **WHEN** 用户点击"登录"按钮后 Nakama 返回认证错误
- **THEN** 按钮恢复为正常状态
- **AND** 表单上方显示错误提示信息（区分"邮箱或密码错误"、"无法连接服务器"）
- **AND** 页面不跳转，用户可重新输入

#### Scenario: 渲染注册表单

- **WHEN** 用户切换到"注册"页签
- **THEN** 显示注册表单（邮箱、密码、确认密码）
- **AND** 显示"注册"按钮

#### Scenario: 点击注册按钮

- **WHEN** 用户在注册页签下点击"注册"按钮
- **THEN** 前端校验：两次密码一致 + 密码 ≥ 8 位
- **AND** 校验通过后调用 `registerWithEmail(email, password)`
- **AND** `created=true` → 认证成功，跳转 `/home`
- **AND** `created=false` → 显示"该邮箱已注册，请直接登录"

#### Scenario: 注册校验失败

- **WHEN** 用户点击"注册"按钮且两次密码不一致
- **THEN** 显示"两次输入的密码不一致"
- **WHEN** 密码长度不足 8 位
- **THEN** 显示"密码长度至少需要8位"

#### Scenario: 页签切换清除错误

- **WHEN** 用户在登录/注册页签间切换
- **THEN** 表单错误提示被清除

#### Scenario: 密码显示切换

- **WHEN** 用户点击密码输入框右侧的眼睛图标
- **THEN** 密码从隐藏状态切换为明文显示
- **AND** 图标从 Eye 切换为 EyeOff

### Requirement: Home 首页

Home 页面 SHALL 展示 CS2 模拟器的核心仪表盘布局，包含玩家信息、赛季横幅、今日任务、快速开始、活动中心、战队排行、最近比赛和卡包推荐等模块。

#### Scenario: 首页渲染

- **WHEN** 用户访问 `/home`
- **THEN** 页面渲染 AppLayout 布局（Header + 内容 + Footer）
- **AND** 内容区显示 PlayerSidebar、SeasonBanner、DailyTasks、ChampionshipCenter、QuickStart、ActivityCenter、TeamRankings、RecentMatches、CardPack 共 9 个业务组件
- **AND** 所有组件使用硬编码 mock 数据正常渲染

#### Scenario: 响应式布局

- **WHEN** 用户在不同屏幕尺寸下访问首页
- **THEN** 移动端（<1024px）组件垂直堆叠
- **AND** 桌面端组件按水平网格布局排列

### Requirement: MatchPage 对战页面

MatchPage SHALL 展示模拟对战的比赛界面，包含比分栏、双方阵容、战术地图、回合事件、玩家数据和战术部署。当前阶段使用 stub hook（不连接 WebSocket），显示空状态界面。

#### Scenario: 对战页面渲染（未连接状态）

- **WHEN** 用户访问 `/match`
- **THEN** 页面顶部显示红色"未连接"状态指示器
- **AND** MatchScoreBar 显示双方队名、Logo、比分和倒计时
- **AND** 左右两侧显示 TeamRoster（T 方和 CT 方，空阵容）
- **AND** 中间显示 TacticalMap（Dust2 地图）
- **AND** 底部显示 RoundEvents（暂无事件）、PlayerStats（空表格）、RoundTactics（暂无战术）

#### Scenario: 回合事件展示

- **WHEN** 没有模拟事件发生（stub 返回空数组）
- **THEN** RoundEvents 显示"暂无事件"
- **AND** RoundTactics 显示"暂无战术信息"

### Requirement: GachaPage 抽卡页面

GachaPage SHALL 展示抽卡系统界面，包含卡包选择侧边栏、卡牌轮播、保底面板和卡池预览。

#### Scenario: 抽卡页面渲染

- **WHEN** 用户访问 `/gacha`
- **THEN** 左侧显示 PackSidebar（默认选中"传奇之路"）
- **AND** 中间显示 CardCarousel（玩家卡牌轮播）
- **AND** 右侧显示 PityPanel（幸运值进度条、保底机制、掉落概率）
- **AND** 底部显示 PoolPreview（卡池可选卡牌一览）

#### Scenario: 切换卡包类别

- **WHEN** 用户在 PackSidebar 点击不同的卡包类别
- **THEN** 选中项高亮显示（`bg-primary/10 border-l-primary`）
- **AND** `selectedPack` 状态更新

### Requirement: RankingPage 排行页面

RankingPage SHALL 展示排行榜系统界面，包含排行榜类别侧边栏、排行表格和右侧信息面板。

#### Scenario: 排行页面渲染

- **WHEN** 用户访问 `/ranking`
- **THEN** 左侧显示 RankingSidebar（默认选中"玩家排行榜"）
- **AND** 中间显示 RankingTable（玩家排行列表，含排名、头像、ELO、综合分等列）
- **AND** 右侧显示 RankingInfoPanel（排行榜说明、奖励信息、个人详情）

#### Scenario: 移动端表格适配

- **WHEN** 用户在移动端访问排行页面
- **THEN** RankingTable 表格列缩减为排名、玩家名和综合分
- **AND** 其余列（ELO、胜场、胜率、K/D、资产、阵容）隐藏

### Requirement: 全局暗金主题

所有页面 SHALL 使用统一的 dark gold 主题，CSS 变量定义在全局样式表中，通过 Tailwind CSS 4 的 `@theme inline` 映射为语义类名。

#### Scenario: 主题显示

- **WHEN** 任意页面渲染
- **THEN** 背景色为 `#0a0e14`（`bg-background`）
- **AND** 卡片背景为 `#141a24`（`bg-card`）
- **AND** 主色调为金色 `#c9a227`（`text-primary`）
- **AND** 边框色为 `#2a3444`（`border-border`）

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

### Requirement: MessagesTab 渲染好友驱动的会话列表

MessagesTab SHALL 使用 `useFriendDM` hook 获取真实 DM 会话列表，不再使用硬编码 mock 数据。会话列表仅包含互为好友（state=FRIEND）的用户。

#### Scenario: 会话列表加载中

- **WHEN** 用户打开 MessagesTab（或切换到"站内信"标签）
- **AND** `useFriendDM` 正在加入频道并加载数据（status="loading"）
- **THEN** 左侧会话列表显示加载指示器（spinner 或骨架屏）
- **AND** 搜索框仍然可见

#### Scenario: 会话列表渲染（有好友对话）

- **WHEN** `useFriendDM` 返回 `status="success"` 且 `conversations` 非空
- **THEN** 左侧会话列表显示每个好友的会话条目
- **AND** 每个条目显示：好友在线状态圆点（绿色=在线/在频道中、灰色=离线）、好友头像（avatar_url 或首字母 Fallback）、好友用户名、最后一条消息摘要（截断超长文本）、最后消息时间（格式化显示）
- **AND** 有未读消息的会话显示未读计数徽章（`Badge` 组件）
- **AND** 会话按最近消息时间降序排列
- **AND** 默认选中第一个会话（`selectedId` 为第一个 conversation 的 friendUserId）

#### Scenario: 会话列表为空（无好友）

- **WHEN** `useFriendDM` 返回 `status="success"` 且 `conversations` 为空
- **THEN** 左侧会话列表显示 Empty 组件："暂无好友会话，请先添加好友"
- **AND** 右侧聊天区域显示空状态提示："选择好友开始聊天"

#### Scenario: 会话列表加载失败

- **WHEN** `useFriendDM` 返回 `status="error"`
- **THEN** 左侧会话列表显示错误提示："加载失败，请稍后重试"
- **AND** 提供"重试"按钮重新加载

#### Scenario: 点击会话切换聊天对象

- **GIVEN** 会话列表已加载且有多个会话
- **WHEN** 用户点击某个会话条目
- **THEN** `selectedId` 更新为该会话的好友 userId
- **AND** 选中条目高亮显示（`bg-primary/10`）
- **AND** 右侧聊天区域切换为该好友的消息历史
- **AND** 该会话的未读计数重置为 0

### Requirement: MessagesTab 发送文本消息

MessagesTab SHALL 支持通过输入框发送文本消息到当前选中的 DM 频道。

#### Scenario: 发送消息

- **GIVEN** 用户选中了好友 A 的会话
- **AND** 输入框中输入了文本 "Hello World"
- **WHEN** 用户点击"发送"按钮（或按 Enter 键，Shift+Enter 换行）
- **THEN** 调用 `useFriendDM.sendMessage(selectedId, "Hello World")`
- **AND** 消息乐观更新到当前聊天消息列表（isSelf=true，状态为发送中）
- **AND** 输入框清空
- **AND** 发送成功后消息状态更新为已发送（显示时间戳）

#### Scenario: 空消息阻止发送

- **GIVEN** 输入框内容为空或仅包含空白字符
- **WHEN** 用户点击"发送"按钮
- **THEN** 消息不发送
- **AND** 输入框保持原状

#### Scenario: 发送失败显示错误状态

- **GIVEN** 发送消息时发生网络错误
- **WHEN** `sendMessage` 抛出异常
- **THEN** 乐观更新的消息气泡显示错误状态（红色边框或错误图标）
- **AND** 用户可点击该消息的"重发"按钮重新尝试发送

### Requirement: MessagesTab 实时接收新消息

MessagesTab SHALL 通过 WebSocket 实时接收新消息，并自动更新 UI。

#### Scenario: 当前会话收到新消息

- **GIVEN** 用户正在查看好友 A 的聊天窗口
- **WHEN** 服务器推送好友 A 发来的新消息（onchannelmessage）
- **THEN** 新消息自动追加到消息列表底部
- **AND** 消息气泡样式为接收方样式（isSelf=false，左侧灰色气泡）
- **AND** 消息列表自动滚动到底部（如用户已在底部）

#### Scenario: 非当前会话收到新消息

- **GIVEN** 用户正在查看好友 A 的聊天窗口
- **WHEN** 服务器推送好友 B 发来的新消息
- **THEN** 当前聊天窗口保持不变
- **AND** 好友 B 的会话条目更新最后消息摘要和时间
- **AND** 好友 B 的会话条目移动到列表顶部（最新消息优先）
- **AND** 好友 B 的未读计数增加

#### Scenario: 用户在其他标签页时收到新消息

- **GIVEN** 用户打开 MessagesTab 但未选中该标签页（在浏览器后台）
- **WHEN** 服务器推送新消息
- **THEN** 消息正常接收并更新本地状态
- **AND** 当用户切换回 MessagesTab 时，未读计数已正确更新

### Requirement: MessagesTab 消息历史加载

MessagesTab SHALL 在选中会话时加载该频道的消息历史。

#### Scenario: 首次选中会话加载历史

- **GIVEN** 用户点击好友 A 的会话
- **WHEN** 该会话的消息尚未本地缓存
- **THEN** 右侧聊天区域显示加载指示器
- **AND** 调用 `listChannelMessages` 加载最近 100 条消息
- **AND** 加载完成后显示消息列表（自己发送的靠右蓝色气泡，对方发送的靠左灰色气泡）
- **AND** 每条消息显示发送者用户名和时间戳

#### Scenario: 再次选中会话使用缓存

- **GIVEN** 用户之前已查看过好友 A 的聊天（消息已缓存）
- **WHEN** 用户从好友 B 切回好友 A
- **THEN** 立即显示缓存的消息列表（无需重新加载）
- **AND** 如果在查看其他会话期间收到好友 A 的新消息，这些消息也已追加到缓存中

#### Scenario: 向上滚动加载更早消息

- **GIVEN** 好友 A 的聊天有大量历史消息（超过一页）
- **WHEN** 用户滚动到消息列表顶部
- **THEN** 触发向前翻页加载更早的历史消息
- **AND** 加载过程中顶部显示加载指示器
- **AND** 加载完成后更早的消息插入列表顶部
- **AND** 滚动位置保持稳定（不跳动）

#### Scenario: 无历史消息

- **GIVEN** 用户与好友 A 之前没有聊过天
- **WHEN** 用户选中好友 A 的会话
- **THEN** 消息列表为空，显示"开始新对话"的友好提示
- **AND** 不显示错误或异常状态

### Requirement: MessagesTab 聊天 UI 结构

MessagesTab SHALL 保持现有的左右分栏布局结构，但对接真实数据。

#### Scenario: 桌面端布局

- **WHEN** 用户在桌面端（≥1024px）访问 MessagesTab
- **THEN** 左右分栏显示：左侧会话列表（w-72） + 右侧聊天区域（flex-1）
- **AND** 左侧顶部显示搜索框和新建按钮（Plus 按钮预留，本阶段可隐藏）
- **AND** 右侧顶部显示当前会话的好友用户名、头像和在线状态指示器（绿色圆点+文字"在线"/灰色圆点+文字"离线"）
- **AND** 右侧中间显示消息列表（flex-1，overflow-y-auto）
- **AND** 右侧底部显示输入工具栏（表情/附件/语音等按钮 + 文本输入框 + 发送按钮）
- **AND** 输入工具栏中仅"发送"按钮实际可用，其余按钮显示但 disabled/不响应

#### Scenario: 移动端布局

- **WHEN** 用户在移动端（<1024px）访问 MessagesTab
- **THEN** 默认显示会话列表（全宽）
- **AND** 点击会话后切换为聊天区域（全宽）
- **AND** 聊天区域顶部显示"返回"按钮回到会话列表
