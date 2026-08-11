## ADDED Requirements

### Requirement: 访客无需显式登录即可进入游戏主界面
客户端 SHALL 在没有正式登录 Session 时静默创建或恢复 Nakama 设备 Session，并将该身份标记为访客；访客 SHALL 能访问主界面和机器人教学战，但 SHALL NOT 被要求先填写邮箱、密码或公开昵称。

#### Scenario: 新设备首次访问
- **GIVEN** 浏览器没有正式 Session 或设备 Session
- **WHEN** 用户访问 `/`
- **THEN** 客户端创建并持久化设备标识，完成 Nakama device authentication
- **AND** 主界面在身份初始化完成后显示，不弹出登录表单

#### Scenario: 访客打开好友社交功能
- **GIVEN** 当前身份是访客
- **WHEN** 用户从辅助好友入口进入好友列表或联系方式交换
- **THEN** 客户端显示登录/注册 Modal
- **AND** 不丢失用户原本准备访问的目标路由

### Requirement: 首次访问提供可拒绝的战斗体验 Modal
主界面 SHALL 使用版本化本地状态判断是否展示首次体验 Modal。Modal SHALL 询问用户是否体验战斗，并提供明确的确认与拒绝操作；拒绝 SHALL 留在可交互的主界面。

#### Scenario: 首次访问接受体验
- **GIVEN** 当前浏览器没有 `onboarding.battleOffer.v1` 完成记录
- **WHEN** 主界面身份初始化完成且用户点击“体验一下”
- **THEN** 客户端记录该版本提示已经处理
- **AND** 导航到 15 元临时组队流程

#### Scenario: 首次访问拒绝体验
- **GIVEN** 首次体验 Modal 正在显示
- **WHEN** 用户点击“暂不体验”
- **THEN** Modal 关闭并记录该版本提示已经处理
- **AND** 用户留在三入口主界面

### Requirement: 15 元临时组队遵守配置预算与人数限制
教学组队页 SHALL 从 `TbTutorialBattle` 读取启用方案的预算、人数和 5、4、3、2、1 元选手池，并要求用户选择恰好五名不重复选手且总价不超过 15 元后才能开始教学战。该阵容 SHALL 是访客临时状态，不得覆盖正式玩家阵容；客户端展示价格不得成为服务端校验依据。

#### Scenario: 合法阵容开始教学战
- **GIVEN** 访客已选择五名不重复选手且总价等于或小于 15 元
- **WHEN** 用户点击“开始战斗”
- **THEN** 客户端允许进入机器人模拟请求阶段
- **AND** 页面清楚标识这是教学体验

#### Scenario: 超出预算
- **GIVEN** 当前选择加入下一名选手后总价将超过 15 元
- **WHEN** 用户尝试加入该选手
- **THEN** 客户端拒绝该选择并显示剩余预算提示
- **AND** 已选阵容保持不变

#### Scenario: 阵容未满
- **GIVEN** 当前只选择了四名选手
- **WHEN** 用户查看开始按钮
- **THEN** 开始按钮不可用并说明仍需选择一名选手

#### Scenario: 教学配置无效
- **GIVEN** `TbTutorialBattle` 的 Player 引用不存在、价格档重复或无法组成预算内五人阵容
- **WHEN** 客户端加载该教学方案
- **THEN** 页面不允许开始比赛并显示配置暂不可用
- **AND** 服务端同样拒绝该 config ID 或 version

### Requirement: 教学双方真实阵容进入模拟引擎
开始教学战时客户端 SHALL 只提交 tutorial config ID、config version 与已选五个 Player ID。服务端 SHALL 重新校验玩家阵容，并从 `TbTutorialBattle` 读取机器人五人阵容和地图，再从 `TbPlayer` 构造双方完整 `TeamInput` 传入 `matchengine`。

#### Scenario: 合法教学阵容开始真实模拟
- **GIVEN** 玩家五人阵容满足当前教学配置且机器人配置包含五名有效选手
- **WHEN** 客户端调用 `SimuMatch` 的 tutorial 模式
- **THEN** 服务端用这十名选手的真实 `TbPlayer` 属性构造 `MatchInput`
- **AND** 返回战报中的双方阵容与提交/配置的 Player ID 一致

#### Scenario: 客户端伪造价格或属性
- **GIVEN** 修改客户端提交了价格、属性或不属于当前档位的 Player ID
- **WHEN** 服务端处理教学模拟请求
- **THEN** 服务端忽略客户端价格/属性并按 Luban 配置重新校验
- **AND** 非法 Player ID 使请求失败且不进入 `matchengine`

### Requirement: 教学阵容不得保存为正式阵容
客户端 SHALL 支持将邮箱凭据链接到当前设备账号或登录已有账号，但无论采用哪种认证路径，15 元教学阵容 MUST NOT 写入账号阵容或成为正式首发阵容。比赛结束、刷新、退出教学流程或切换账号后，系统 MAY 丢弃该临时选择。

#### Scenario: 教学战结束
- **GIVEN** 访客或正式账号使用 15 元阵容完成教学战
- **WHEN** 用户离开 Battle 页面
- **THEN** 服务端不存在由该教学选择创建的玩家阵容记录
- **AND** 主界面的调整阵容入口仍显示未开放
