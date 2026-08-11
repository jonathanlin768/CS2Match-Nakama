## ADDED Requirements

### Requirement: 选手卡面和头像裁切数据随比赛快照传递

服务端 SHALL 从 Luban Player 配置读取完整卡面及头像裁切数据，并经领域模型和比赛快照传递给客户端。新增字段 SHALL 为可选字段，旧 `portrait` 数据 SHALL 继续可用。

#### Scenario: Luban 为服务端和客户端生成一致字段

- **GIVEN** `#Player.xlsx` 包含 `cardImage`、`avatarCropX`、`avatarCropY`、`avatarCropWidth`、`avatarCropHeight`
- **WHEN** 项目运行 Luban 导表
- **THEN** Server Go 和 Client TypeScript 生成结构包含语义一致的字段
- **AND** 生成的 JSON 数据保留相同的资源路径和归一化数值

#### Scenario: Match 服务构建选手视觉资料

- **GIVEN** 一条 Player 配置包含合法的完整卡面和裁切参数
- **WHEN** Match 服务从 `TbPlayer` 构建 `PlayerProfile`
- **THEN** `PlayerProfile` 包含 `CardImage`
- **AND** 四个扁平配置字段被组装为一个可选的 `AvatarCrop` 值对象
- **AND** 回合投影将 `CardImage` 和 `AvatarCrop` 复制到客户端可见的 `PlayerState`

#### Scenario: 非法或缺失新字段使用旧头像

- **GIVEN** Player 未配置完整卡面，或裁切矩形不合法
- **WHEN** Match 服务构建比赛快照
- **THEN** 服务端不输出可用的 `AvatarCrop`
- **AND** 保留现有 `Portrait` 路径供客户端回退
- **AND** 比赛模拟仍可正常创建和推进

#### Scenario: 视觉配置不改变模拟过程

- **GIVEN** 两份 Player 配置只有卡面和头像裁切参数不同
- **WHEN** MatchLoop 使用相同种子和相同比赛输入运行
- **THEN** 两场模拟产生相同的比赛状态与事件结果
- **AND** MatchLoop 不直接读取静态配置文件或图片资源
