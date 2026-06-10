## 1. API 层 — 创建 `api/friends.ts`

- [x] 1.1 新建 `client/src/api/friends.ts`，实现 `listFriends(session, limit?): Promise<Friends>` 函数，调用 `client.listFriends(session, undefined, limit ?? 100)` 获取当前用户的所有好友列表，默认 limit=100 避免分页截断
- [x] 1.2 实现 `addFriendsByUsername(session, username): Promise<boolean>` 函数，调用 `client.addFriends(session, undefined, [username])` 通过用户名添加好友
- [x] 1.3 实现 `deleteFriend(session, userId): Promise<boolean>` 函数，调用 `client.deleteFriends(session, [userId])` 删除指定好友
- [x] 1.4 为所有 API 函数添加 TypeScript 类型标注和 JSDoc 注释

## 2. Hook 层 — 创建 `hooks/useFriends.ts`

- [x] 2.1 新建 `client/src/hooks/useFriends.ts`，定义 `FriendState` 状态机：`idle | loading | success | error`，管理 `friends`、`error`、`status` 状态
- [x] 2.2 实现挂载时自动调用 `listFriends(session)` 的数据加载逻辑（session 为 null 时跳过）
- [x] 2.3 实现 `friendsByState` 分组：按 Nakama Friend State（0=好友/1=已发送/2=已接收）分组，返回 `Record<number, Friend[]>` 结构和分组名映射
- [x] 2.4 实现 `addFriend(username): Promise<boolean>` 方法，调用 `addFriendsByUsername` 并自动刷新列表
- [x] 2.5 实现 `removeFriend(userId): Promise<boolean>` 方法，调用 `deleteFriend` 并自动刷新列表
- [x] 2.6 实现 `searchQuery` 状态和 `filteredFriends` 派生数据（前端用户名/displayName 模糊匹配，大小写不敏感）
- [x] 2.7 实现 `retry()` 方法，在 error 状态下重新加载好友列表

## 3. 改造 `ProfilePage.tsx` 好友部分

- [x] 3.1 删除 `mockFriends` 和 `friendGroups` 硬编码数据，删除 `Friend` interface（改用 Nakama 类型）
- [x] 3.2 引入 `useAuth()` 获取 `session`，引入 `useFriends(session)` 获取好友数据和操作方法
- [x] 3.3 好友列表左侧面板：用 `friendsByState` 渲染分组（"我的好友"/"已发送请求"/"收到的请求"），保留折叠/展开状态
- [x] 3.4 每个好友项使用 Nakama `Friend.user` 数据渲染：头像（`avatar_url` + fallback 首字母）、用户名（`username`）、签名（`display_name` 或 `-`）、在线状态（`online` 字段映射为 green/gray 指示器）
- [x] 3.5 实现搜索框实时过滤：输入关键词时更新 `searchQuery`，好友列表根据 `filteredFriends` 重新渲染
- [x] 3.6 添加加载状态：`status === "loading"` 时显示 Spinner 组件
- [x] 3.7 添加空状态：`status === "success"` 且 `friends.length === 0` 时显示 Empty 组件"暂无好友"
- [x] 3.8 添加错误状态：`status === "error"` 时显示错误提示 + "重试"按钮
- [x] 3.9 好友详情右侧面板：映射 Nakama User 字段（`username`、`id`、`location`、`avatar_url`、`online`、`update_time` → 格式化为"最近活跃: X天前"），不存在的字段（`signature`、`level`、`age`、`birthday`、`likes`、`gender`）显示 `-` 或隐藏
- [x] 3.10 移除 mock 的"精选战绩"（Photo Wall）和"勋章"（Medals）区域 — 这些数据 Nakama 不提供
- [x] 3.11 "收到的请求"分组标题旁显示红色 `Badge` 徽章（数字 = 未处理请求数），徽章为 0 时自动隐藏该分组
- [x] 3.12 没有成员的其他分组也自动隐藏（如"已发送请求"为空时不显示）

## 4. 添加好友对话框

- [x] 4.1 在 ProfilePage 中实现添加好友 Dialog（Modal），包含用户名输入框、"添加"和"取消"按钮
- [x] 4.2 实现前端校验：空输入提示"请输入用户名"、输入自己的用户名提示"无法添加自己为好友"
- [x] 4.3 实现"添加"按钮点击：调用 `useFriends.addFriend(username)`，loading 状态显示 spinner
- [x] 4.4 处理添加结果：成功 → 关闭对话框 + Toast 提示"好友请求已发送"；失败 → 对话框内显示错误信息（"用户不存在"/"已是好友"）
- [x] 4.5 对话框关闭后清空输入和错误状态

## 5. 动态操作按钮（根据 Friend State 切换）

- [x] 5.1 去除 mock 面板中的"编辑资料"和"发消息"按钮（它们不是真实功能）
- [x] 5.2 重构详情面板底部按钮区域为动态渲染，根据 `selectedFriend.state` 显示不同按钮组合
- [x] 5.3 state=FRIEND(0)：显示"删除好友"按钮（危险样式）
- [x] 5.4 "删除好友"：点击弹出确认对话框 → 确认后调用 `removeFriend(userId)` → 成功后 Toast 提示 + 列表刷新 + 详情面板重置
- [x] 5.5 state=INVITE_SENT(1)：显示"取消请求"按钮（outline 样式）
- [x] 5.6 "取消请求"：点击弹出确认对话框 → 确认后调用 `removeFriend(userId)` → 成功后 Toast 提示"已取消好友请求" + 列表刷新 + 详情面板重置
- [x] 5.7 state=INVITE_RECEIVED(2)：显示"接受"按钮（primary 强调样式）+"拒绝"按钮（outline 样式）
- [x] 5.8 "接受"：点击后调用 `addFriend(senderUsername)` → 成功后 Toast 提示"已接受好友请求" + 好友移至"我的好友"分组 + 按钮更新为 FRIEND 状态按钮
- [x] 5.9 "拒绝"：点击弹出确认对话框 → 确认后调用 `removeFriend(userId)` → 成功后 Toast 提示"已拒绝好友请求" + 列表刷新 + 详情面板重置

## 6. 验证

- [x] 6.1 TypeScript 类型检查通过：`cd client && npx tsc --noEmit`
- [x] 6.2 启动 Docker 环境（`docker compose up -d`），确保 Nakama 在 `localhost:7350` 可用
- [x] 6.3 验证好友列表加载：注册 2 个测试账号 → 互相添加好友 → 登录后进入 ProfilePage → 好友列表正确展示
- [x] 6.4 验证添加好友：对话框输入对方用户名 → 点击添加 → 对方出现在"已发送请求"分组 → Toast 提示正确
- [x] 6.5 验证接受好友请求：对方登录后"收到的请求"有红色徽章 → 选中该请求 → 点击"接受" → 好友移至"我的好友" → Toast 正确
- [x] 6.6 验证拒绝好友请求：对方登录后"收到的请求"有红色徽章 → 选中该请求 → 点击"拒绝" → 确认 → 好友从列表移除 → Toast 正确
- [x] 6.7 验证取消已发送请求：选中"已发送请求"中的好友 → 点击"取消请求" → 确认 → 好友从列表移除 → Toast 正确
- [x] 6.8 验证删除好友：选中"我的好友"中的好友 → 点击删除 → 确认 → 好友从列表中移除 → Toast 正确
- [x] 6.9 验证搜索过滤：在搜索框中输入关键词 → 好友列表实时过滤
- [x] 6.10 验证空分组自动隐藏（如无"已发送请求"时该分组不显示）
- [x] 6.11 验证空状态（无任何好友时显示"暂无好友"）和错误状态（Nakama 不可达时显示错误提示 + 重试按钮）
