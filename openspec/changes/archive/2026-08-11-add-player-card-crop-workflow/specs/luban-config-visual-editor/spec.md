## ADDED Requirements

### Requirement: Player 专属页配置完整卡面与 5:7 头像裁切

配置工作台 SHALL 允许策划为选手上传完整卡面，并在原图上使用固定 `5:7` 宽高比的裁切框配置比赛头像。工作台 SHALL 将完整卡面路径和归一化裁切矩形写入 `#Player.xlsx`，而不是生成第二张裁切图片。

#### Scenario: 上传并预览完整卡面

- **GIVEN** 用户正在编辑一条 Player 记录
- **WHEN** 用户选择电脑上的 PNG、JPEG 或 WebP 卡面图片
- **THEN** 本地服务将图片复制到 `client/public/player-cards/`
- **AND** 系统将 `player-cards/<文件名>` 项目相对路径写入 `cardImage`
- **AND** 页面同时显示完整 `2:3` 卡面和 `5:7` 头像预览

#### Scenario: 使用固定比例裁切头像

- **GIVEN** 当前 Player 已配置可读取的完整卡面
- **WHEN** 用户拖动或缩放头像裁切框
- **THEN** 裁切框宽高比始终保持为 `5:7`
- **AND** 头像预览实时显示裁切后的结果
- **AND** 系统以原图自然尺寸为基准更新 `avatarCropX`、`avatarCropY`、`avatarCropWidth`、`avatarCropHeight`
- **AND** 四个值均使用 `0..1` 归一化坐标

#### Scenario: 重置头像裁切区域

- **GIVEN** 当前 Player 已配置完整卡面
- **WHEN** 用户点击“重置裁切”
- **THEN** 系统在原图范围内生成居中的最大 `5:7` 裁切框
- **AND** 页面立即更新头像预览和归一化参数

#### Scenario: 阻止保存非法裁切参数

- **GIVEN** Player 已配置 `cardImage`
- **WHEN** 裁切参数包含非有限数值、非正宽高、越界矩形或不符合 `5:7` 比例的矩形
- **THEN** 校验结果显示 `ERROR`
- **AND** 错误项可以定位到该 Player 的卡面裁切区域
- **AND** 工作台阻止保存和写入 Luban

#### Scenario: 完整卡面缺失时保留旧头像回退

- **GIVEN** Player 未配置 `cardImage` 或卡面资源无法读取
- **WHEN** 工作台加载该记录
- **THEN** 工作台继续允许使用现有 `portrait` 预览
- **AND** 系统提示完整卡面或裁切配置缺失
- **AND** 不因旧数据缺少新增字段而损坏工作簿

#### Scenario: 保存时只持久化源图和裁切数据

- **GIVEN** 用户已完成卡面上传和头像裁切
- **WHEN** 用户保存 Player 表
- **THEN** 系统保存完整卡面文件、`cardImage` 和四个裁切参数
- **AND** 系统不生成或维护独立的裁切头像图片

#### Scenario: 初始素材迁移后直接显示默认卡面

- **GIVEN** 项目获得一批带可识别选手文件名的完整卡面素材
- **WHEN** 执行本变更的初始数据迁移
- **THEN** 所有素材被复制到 `client/public/player-cards/` 受控目录
- **AND** 能可靠匹配现有 Player 的素材被写入对应 `cardImage`
- **AND** 对应 Player 获得合法的 `5:7` 归一化默认裁切参数
- **AND** 无法可靠匹配的 Player 保留 `portrait` 回退且迁移结果明确列出
