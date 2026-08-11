## ADDED Requirements

### Requirement: 静默设备认证提供访客 Session
客户端 SHALL 在没有可恢复 Session 时生成并持久化合法设备 ID，通过 Nakama device authentication 创建或恢复 Guest Session。该过程 SHALL 不要求用户先填写邮箱、密码或昵称。

#### Scenario: 首次打开游戏
- **GIVEN** 浏览器没有 Session token 和设备 ID
- **WHEN** 应用初始化
- **THEN** 客户端生成 10 至 128 字节的设备 ID 并完成 device authentication
- **AND** 将返回 Session 标记为访客身份

### Requirement: 系统生成八位唯一 username
新账号的 Nakama username SHALL 由项目认证逻辑生成为恰好 8 位大写英文字母与数字 code。客户端 SHALL NOT 接受玩家自定义 username 或 display name；创建冲突时 SHALL 重新生成并有限次重试。

#### Scenario: 创建访客账号
- **GIVEN** device ID 尚未关联 Nakama 用户
- **WHEN** 服务器创建账号
- **THEN** 账号 username 符合 `^[A-Z0-9]{8}$`
- **AND** UI 将其格式化为 `玩家#CODE`

#### Scenario: 客户端尝试改名
- **GIVEN** 玩家已经拥有系统 username
- **WHEN** 修改客户端提交不符合规范的 account update
- **THEN** before-hook 拒绝请求
- **AND** username 保持不变

### Requirement: 正式登录不保存教学阵容
邮箱注册、账号链接或登录已有账号 SHALL 只迁移身份与 Session，不得把客户端临时 15 元教学阵容写入账号 Storage。

#### Scenario: 教学过程中登录
- **GIVEN** Guest Session 已选择临时教学阵容
- **WHEN** 用户完成邮箱注册或登录
- **THEN** 客户端切换到正式身份并恢复目标页面
- **AND** 服务端不创建正式阵容记录
