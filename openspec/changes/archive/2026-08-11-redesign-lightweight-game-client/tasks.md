## 1. 实施前基线与产品数据确认

- [x] 1.1 记录当前客户端路由、Session、Home/Match/Battle、好友和 DM 的测试基线，确认现有 `DebugSimuMatch` 与完整战报回放通过
- [x] 1.2 新建 `configs/Datas/#Team.xlsx`，配置 team ID、正式名、简称、昵称和 Logo，并录入教学战及现有默认模拟所需队伍数据
- [x] 1.3 将 `#Player.xlsx` 的自由文本 `team` 迁移为引用 `TbTeam` 的 `teamId`，补齐现有选手引用并增加按 teamId 查询测试
- [x] 1.4 新建 `configs/Datas/#TutorialBattle.xlsx`，按版本配置 15 元预算、五人、5/4/3/2/1 元 Player 列表、地图和机器人五人阵容
- [x] 1.5 执行 Luban 导出，更新服务端/客户端 loader 与生成产物，并验证非法 Team/Player 引用、重复价格档、无预算可解阵容和非法对手配置会失败
- [x] 1.6 将全部测试账号清理或规范化为 8 位大写字母数字 Nakama username，不保留旧自由昵称兼容分支

## 2. 访客身份与账号迁移基础

- [x] 2.1 扩展客户端认证层，生成/持久化设备 ID 并静默创建或恢复 Guest Session，区分 guest、正式账号和初始化状态
- [x] 2.2 实现当前设备账号链接邮箱注册的流程，以及登录已有账号时安全切换 Session 且不迁移临时阵容的流程
- [x] 2.3 为 device/email 创建流程生成恰好 8 位大写字母数字 Nakama username，以唯一约束处理碰撞并有限重试，UI 统一添加 `玩家#` 展示前缀
- [x] 2.4 注册认证与 update-account hooks，拒绝任意自定义 username/displayName，并在 username 变化后正确刷新 Session
- [x] 2.5 添加认证单元/集成测试，覆盖首次设备认证、恢复、链接账号、切换已有账号、格式拒绝、碰撞重试和访客访问限制

## 4. 服务端 Social 联系方式交换子系统

- [x] 4.1 创建 `server/internal/social` 的 model、repository、service、RPC 和 hook 文件，定义联系方式 profile、pair exchange、状态机、版本与幂等字段
- [x] 4.2 实现 owner 为玩家 UUID、key 为 `profile`、读写权限均为 0 的 `social_contact_profile` Storage，并实现 `SocialSetContactProfile` 的身份、渠道、长度/格式校验和脱敏响应
- [x] 4.3 实现 system owner、读写权限均为 0、确定性 pair key 的 `social_contact_exchange` Storage，以及请求、接受、拒绝、失效和并发 version 控制
- [x] 4.4 实现 `SocialGetContactExchange`、`SocialRequestContactExchange`、`SocialRespondContactExchange`，在每次请求与读取时通过 Nakama 好友列表重验关系
- [x] 4.5 实现请求频率限制与幂等重试，确保联系方式正文不进入频道 payload、错误日志或普通响应
- [x] 4.6 使用服务端 `ChannelIdBuild`/`ChannelMessageSend` 发送结构化卡片，并处理“Storage 成功、卡片投递失败”的可恢复路径
- [x] 4.7 注册 before realtime hooks，拒绝客户端 DM 的 ChannelMessageSend、ChannelMessageUpdate 和 ChannelMessageRemove，同时验证服务端卡片仍可投递
- [x] 4.8 在 `InitModule` 注册 Social RPC/hooks，并编写 Go 测试覆盖 Storage 权限、非好友访问、越权响应、并发请求、删除好友后撤销披露、客户端消息拦截和卡片投递失败恢复

## 5. 新响应式应用外壳与路由

- [x] 5.1 建立新的响应式设计 token、AppShell 和基础组件，覆盖约 390px 手机、平板与约 1200px 最大桌面内容宽度、安全区和 44px 触控目标
- [x] 5.2 重构 React Router：`/` 主界面、访客教学/电脑战路由、正式账号好友路由、阵容/好友对战未开放页及登录 Modal 返回目标逻辑
- [x] 5.3 为 `/home` 提供到 `/` 的兼容重定向，并移除或下线 `/gacha`、`/ranking` 等非本期导航入口
- [x] 5.4 实现只突出“调整阵容、好友对战、匹配对战”的 Home 页面；前两项显示未开放，匹配对战可进入电脑模拟
- [x] 5.5 添加路由与布局测试，覆盖 Guest/正式身份守卫、认证后返回目标、390px/1024px 断点、无固定画布横向溢出

## 6. 首次体验与 15 元教学阵容

- [x] 6.1 实现版本化首次体验本地状态和可接受/拒绝的战斗体验 Modal，确保拒绝后主界面仍可操作
- [x] 6.2 实现 5/4/3/2/1 元候选池、选手选择/撤销、剩余预算、五人唯一性与总价不超过 15 的校验
- [x] 6.3 实现一次性教学阵容状态，确保访客或正式账号完成比赛、刷新、退出或切换账号时都不创建玩家阵容 Storage
- [x] 6.4 调用 `SimuMatch(mode=tutorial)` 提交 config ID/version 与五个 Player ID，让服务端用教学表对手和双方真实 `TbPlayer` 属性构造 `MatchInput`
- [x] 6.5 添加前端测试，覆盖 Modal 版本状态、接受/拒绝、预算边界、人数/重复校验、刷新恢复策略和访客转正式账号边界

## 8. 匹配与 Battle 界面迁移

- [x] 8.1 重构匹配页面，第一阶段提供可用的电脑对战，并将路人匹配和权威好友对战显示为真实的未开放状态
- [x] 8.2 新增生产用 `SimuMatch` RPC/API 与请求类型，复用 `MatchReport` 响应；保留 `DebugSimuMatch` 调试兼容，并移除 250ms 假匹配和虚构真实玩家文案
- [x] 8.3 在 `internal/match` 共享构建路径中校验 tutorial/computer 请求，从 `TbTeam`、`TbTutorialBattle`、`TbPlayer` 构造双方真实 `MatchInput`，覆盖 config version 和非法阵容错误
- [x] 8.4 将 BattlePage 和比赛组件迁移到响应式布局，保持现有逐局回放、跳过、时间线、地图和统计行为
- [x] 8.5 移除正常流程的 mock 比分/胜者回退；无有效战报直接访问 `/battle` 时显示空状态和返回匹配入口
- [x] 8.6 添加前后端回归测试，验证 Guest/正式 Session、真实双方 Player ID、配置篡改拒绝、服务器战报权威、回放顺序不变及响应式布局

## 9. 好友与联系方式卡片前端

- [x] 9.1 将 legacy `useFriends`、好友 API 和共享 Socket 迁入新页面结构，保留申请、接受、拒绝、取消、删除、通知刷新和在线状态
- [x] 9.2 新增 Social RPC TypeScript 类型/API/hooks，展示本人联系方式配置的脱敏状态及好友交换权威状态
- [x] 9.3 将 MessagesTab 替换为结构化交换卡片视图，支持请求、接受、拒绝、已接受和已失效状态，不渲染任何自由文本或媒体输入
- [x] 9.4 复用 DM join/history/onchannelmessage/重连能力解析卡片；对未知类型安全忽略，并在重连或选中好友时按 request ID 刷新 RPC 状态
- [x] 9.5 删除 UI 对 `writeChatMessage(text)` 的依赖和导出，确认修改客户端直接调用频道写入时由服务器拒绝
- [x] 9.6 添加前端测试，覆盖好友分组、卡片实时/历史/未读、越权失败、删除好友后清除联系方式显示、未知卡片和移动端单栏导航

## 10. 清理、验证与发布

- [x] 10.1 删除未再引用的旧固定画幅大厅组件、mock 首页数据、文本聊天控件和抽卡/排行导航，保留与本期无冲突的底层 API
- [x] 10.2 执行客户端 lint、TypeScript 检查、单元测试和生产构建，并进行桌面与常见手机视口的可访问性/视觉回归检查
- [x] 10.3 执行全部 Go 测试与 `go build`/Nakama plugin `.so` 编译验证，确认没有新增 Go module 或未说明的依赖
- [x] 10.4 在 Docker Compose 测试环境执行端到端验证：新访客教学且阵容不保存、电脑战回放、两个未开放入口、好友申请、交换卡片、客户端消息拦截和删除好友撤权
- [x] 10.5 验证本变更没有注册 Match Handler、没有改变 MatchLoop tick、没有修改 matchengine 规则；Luban 生成产物只由导表脚本产生，且没有数据库 schema migration
- [x] 10.6 准备前端回滚构建与 Social 入口功能开关，记录上线监控项：认证失败、阵容 RPC 错误、Storage 冲突、卡片投递失败和越权读取拒绝
