## ADDED Requirements

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
