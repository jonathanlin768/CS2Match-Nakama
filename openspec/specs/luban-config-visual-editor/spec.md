# luban-config-visual-editor Specification

## Purpose
TBD - created by archiving change add-luban-config-visual-editor. Update Purpose after archive.
## Requirements
### Requirement: 配置工作台提供多页面 Luban 可视化编辑入口

本地配置工作台 SHALL 在同一个 Web 工具中提供地图配置、选手配置、战队配置、新手战斗配置和全部表格等页面入口。地图配置页 SHALL 继续承载现有雷达画布编辑能力；非地图配置页 SHALL 使用表格、列表、表单和引用选择器维护 Luban 表。

#### Scenario: 在配置页面之间切换

- **GIVEN** 用户已启动本地配置工作台
- **WHEN** 用户点击 `地图配置`、`选手配置`、`战队配置`、`新手战斗` 或 `全部表格`
- **THEN** 系统切换到对应页面
- **AND** 地图配置页保留现有雷达画布、图层、属性面板和底部日志布局
- **AND** 非地图配置页显示适合表格和表单编辑的列表、筛选、详情和写入入口

#### Scenario: 非地图页不启用地图画布快捷键

- **GIVEN** 用户处于选手配置、战队配置、新手战斗或全部表格页面
- **WHEN** 用户按下 `S`、`N`、`E`、`V`、`L` 或 `R`
- **THEN** 系统不切换地图画布工具
- **AND** 输入框、表格单元格和多选控件保留正常文本编辑行为

### Requirement: 工作台读取 Luban Excel 表元数据和数据行

本地写入服务 SHALL 支持读取 `configs/Datas/#*.xlsx` 的 Luban 表头和数据行。读取结果 SHALL 包含文件名、推导表名、字段名、字段类型、导出分组、字段注释、数据行和解析警告。

#### Scenario: 读取 Player 表结构

- **GIVEN** `configs/Datas/#Player.xlsx` 存在且包含标准 4 行 Luban 表头
- **WHEN** 前端请求读取该表
- **THEN** 服务返回 `id`、`name`、`teamId`、`entry`、`aim`、`positions` 等字段元数据
- **AND** `teamId` 字段被识别为引用 `TbTeam` 的字段
- **AND** `positions` 字段被识别为字符串列表字段
- **AND** 数据行从第 5 行开始返回

#### Scenario: 读取 TutorialBattle 表结构

- **GIVEN** `configs/Datas/#TutorialBattle.xlsx` 存在
- **WHEN** 前端请求读取该表
- **THEN** 服务返回 `tier5PlayerIds`、`tier4PlayerIds`、`tier3PlayerIds`、`tier2PlayerIds`、`tier1PlayerIds` 和 `opponentPlayerIds` 字段
- **AND** 这些字段被识别为引用 `TbPlayer` 的多选列表字段
- **AND** `opponentTeamId` 被识别为引用 `TbTeam` 的单选字段

### Requirement: 工作台提供通用 Luban 表格编辑能力

工作台 SHALL 为未定制的非地图 `#*.xlsx` 表提供通用表格编辑 fallback。通用表格 SHALL 支持查看字段注释和类型、编辑单元格、新增行、复制行、删除行、保存当前表和查看校验结果。第一版 SHALL 允许通用表格页编辑 `#item.xlsx` 等已有非地图表，而不是仅开放查看。地图语义表以及 `#CombatConst.xlsx` SHALL 继续由地图配置页单一维护，不得通过通用表格形成第二个编辑入口。

#### Scenario: 编辑通用表格单元格

- **GIVEN** 用户打开 `全部表格` 页面并选择一张 `#*.xlsx` 表
- **WHEN** 用户修改某个数据单元格
- **THEN** 系统将该表标记为有未保存修改
- **AND** 单元格编辑器显示对应字段的 `##type` 和 `##` 注释

#### Scenario: 新增和删除通用表格行

- **GIVEN** 用户打开一张通用 Luban 表
- **WHEN** 用户点击 `新增行`
- **THEN** 系统追加一条空数据行
- **WHEN** 用户选择一条数据行并点击 `删除行`
- **THEN** 系统从当前表数据中移除该行
- **AND** 该修改在用户保存前不写入 Excel 文件

#### Scenario: 编辑 item 等非地图表

- **GIVEN** `configs/Datas/#item.xlsx` 存在
- **WHEN** 用户在 `全部表格` 页面打开该表并修改基础类型字段
- **THEN** 系统允许用户编辑该字段并保存当前表
- **AND** 保存后仍保留原 Luban 表头结构

#### Scenario: 地图语义表只使用地图配置入口

- **GIVEN** `#MapNode.xlsx`、`#MapEdge.xlsx`、`#Route.xlsx` 或 `#CombatConst.xlsx` 等地图语义表存在
- **WHEN** 用户在 `全部表格` 页面查看表列表并选择其中一张
- **THEN** 系统标明该表归属 `地图配置`
- **AND** 系统不提供通用单元格编辑或通用保存入口
- **AND** 系统提供前往 `地图配置` 页面的操作

### Requirement: 工作台提供 Player 专属编辑页

工作台 SHALL 为 `#Player.xlsx` 提供专属选手配置页。该页面 SHALL 支持按战队、稀有度和位置筛选选手，编辑基础资料、战队引用、能力值、位置标签、稀有度和头像路径。所有选手能力值字段 SHALL 统一限定为 `0..100` 的整数。

#### Scenario: 编辑选手战队引用

- **GIVEN** 用户打开 `选手配置` 页面
- **AND** `#Team.xlsx` 已加载并包含多个战队
- **WHEN** 用户选中一个选手并修改 `teamId`
- **THEN** 系统通过战队下拉或搜索选择器写入有效的 `TbTeam` 引用 ID
- **AND** 若用户选择不存在的战队 ID，校验结果显示引用错误

#### Scenario: 编辑选手能力值

- **GIVEN** 用户选中一个选手
- **WHEN** 用户修改 `entry`、`aim`、`trade`、`clutch`、`firepower`、`gamesense`、`reaction`、`positioning`、`awareness`、`teamplay`、`utility`、`composure`、`mobility`、`endurance` 或 `discipline`
- **THEN** 系统以数值输入控件保存该字段
- **AND** 字段值不符合整数格式时显示校验错误
- **AND** 字段值小于 0 或大于 100 时显示阻止写入的校验错误

#### Scenario: 编辑选手位置标签

- **GIVEN** 用户选中一个选手
- **WHEN** 用户新增或移除 `positions` 标签
- **THEN** 系统将标签列表转换为 Luban 可读的列表单元格值
- **AND** 保存后 `#Player.xlsx` 中该字段保持 `(list#sep=,),string` 语义

#### Scenario: 选择外部文件并预览选手头像

- **GIVEN** 用户选中一个选手
- **WHEN** 用户为 `portrait` 选择电脑上的外部图片文件
- **THEN** 本地服务将图片复制到 `client/public/portraits/`
- **AND** 系统将 `portraits/<文件名>` 项目相对路径写入 `portrait`
- **AND** 页面显示该头像图片预览
- **AND** 若图片文件缺失或不可读取，页面显示非阻塞告警

### Requirement: 工作台提供 Team 专属编辑页

工作台 SHALL 为 `#Team.xlsx` 提供专属战队配置页。该页面 SHALL 支持编辑战队 ID、官方名称、简称、中文昵称、Logo 路径，并显示被选手和新手战斗配置引用的情况。Logo SHALL 支持项目内文件选择与图片预览。

#### Scenario: 编辑战队资料

- **GIVEN** 用户打开 `战队配置` 页面
- **WHEN** 用户选中一个战队并修改 `name`、`shortName`、`nickname` 或 `logo`
- **THEN** 系统更新当前战队记录
- **AND** 若 `logo` 路径可在项目中解析，页面显示 Logo 预览
- **AND** 若 `logo` 路径不可解析，页面显示非阻塞提示

#### Scenario: 选择外部文件并预览战队 Logo

- **GIVEN** 用户选中一个战队
- **WHEN** 用户为 `logo` 选择电脑上的外部图片文件
- **THEN** 本地服务将图片复制到 `client/public/teams/`
- **AND** 系统将 `teams/<文件名>` 项目相对路径写入 `logo`
- **AND** 页面显示该 Logo 图片预览
- **AND** 若图片文件缺失或不可读取，页面显示非阻塞告警

#### Scenario: 提示战队引用数量

- **GIVEN** `#Player.xlsx` 中存在多个选手引用 `team_vitality`
- **WHEN** 用户在战队配置页选中 `team_vitality`
- **THEN** 系统显示该战队被选手表引用的数量
- **AND** 若 TutorialBattle 的 `opponentTeamId` 引用该战队，系统也显示该引用

### Requirement: 工作台提供 TutorialBattle 专属编辑页

工作台 SHALL 为 `#TutorialBattle.xlsx` 提供专属新手战斗配置页。该页面 SHALL 支持编辑启用状态、版本、预算、阵容人数、地图 ID、各费用档候选选手、对手战队和对手阵容。

#### Scenario: 配置费用档候选选手

- **GIVEN** 用户打开 `新手战斗` 页面
- **AND** `#Player.xlsx` 已加载
- **WHEN** 用户在 `tier5PlayerIds`、`tier4PlayerIds`、`tier3PlayerIds`、`tier2PlayerIds` 或 `tier1PlayerIds` 中选择多个选手
- **THEN** 系统将选手 ID 写入对应列表字段
- **AND** 选择器显示选手名称、战队和稀有度等辅助信息
- **AND** 列表中出现不存在的选手 ID 时显示引用错误

#### Scenario: 阻止保存空费用档

- **GIVEN** 用户打开 `新手战斗` 页面
- **AND** `tier1PlayerIds`、`tier2PlayerIds`、`tier3PlayerIds`、`tier4PlayerIds` 或 `tier5PlayerIds` 中任意一个档位为空
- **WHEN** 用户点击 `保存当前表`
- **THEN** 系统显示该费用档缺少候选选手的校验错误
- **AND** 系统不写入 `#TutorialBattle.xlsx`

#### Scenario: 配置对手阵容

- **GIVEN** 用户在 `新手战斗` 页面选择了 `opponentTeamId`
- **WHEN** 用户编辑 `opponentPlayerIds`
- **THEN** 系统提供引用 `TbPlayer` 的多选控件
- **AND** 系统优先显示属于 `opponentTeamId` 的选手
- **AND** 若已选选手不属于该战队，系统显示阻止保存的校验错误且不自动删除该选手

#### Scenario: 切换唯一启用的新手战斗配置

- **GIVEN** `#TutorialBattle.xlsx` 已有一条 `enabled=true` 的配置
- **WHEN** 用户尝试启用另一条配置
- **THEN** 系统提示启用后将关闭当前已启用配置
- **AND** 用户确认后系统将原配置改为 `enabled=false` 并启用新配置
- **AND** 用户取消时两条配置均保持原状态

#### Scenario: 阻止保存不满足运行时约束的启用配置

- **GIVEN** 用户正在编辑一条启用的 TutorialBattle 配置
- **WHEN** `budget` 不大于 0、`rosterSize` 不等于 5、选手跨费用档重复、候选池无法在预算内组成完整阵容、对手阵容人数不等于 `rosterSize`、对手选手重复或对手选手不属于 `opponentTeamId`
- **THEN** 系统显示与当前运行时校验一致的 ERROR
- **AND** 系统不写入 `#TutorialBattle.xlsx`

### Requirement: 工作台保护被引用记录并显式同步 ID 修改

工作台 SHALL 根据 Luban `##type` 引用元数据查找 Team、Player 及其他记录的引用位置。被引用记录 SHALL NOT 被直接删除；修改被引用 ID 时，系统 SHALL 显示引用范围，并且只在用户明确执行 `修改 ID 并同步引用` 后更新源记录及全部引用单元格。跨表同步写入 SHALL 作为一个批次执行，失败时不得留下部分更新状态。

#### Scenario: 阻止删除被引用的战队

- **GIVEN** `#Player.xlsx.teamId` 或 `#TutorialBattle.xlsx.opponentTeamId` 引用了当前战队
- **WHEN** 用户尝试删除该战队
- **THEN** 系统阻止删除
- **AND** 系统列出引用表、字段和记录 ID

#### Scenario: 显式修改 Player ID 并同步引用

- **GIVEN** 当前 Player ID 被 TutorialBattle 的费用档或对手阵容引用
- **WHEN** 用户输入新 ID 并点击 `修改 ID 并同步引用`
- **THEN** 系统展示所有将被更新的引用位置并要求确认
- **AND** 用户确认后系统更新 Player ID 及全部单选和列表引用
- **AND** 系统在写入前为全部受影响 Excel 创建备份
- **AND** 任一文件校验或写入失败时系统恢复本批次已写文件

### Requirement: 工作台校验通用类型、ID 和引用关系

工作台 SHALL 在保存 Luban 表前执行表级校验。校验 SHALL 覆盖空 ID、重复 ID、类型不匹配、单选引用不存在、多选引用不存在和必要字段格式错误。ERROR 级别问题 SHALL 阻止写入当前表；WARNING 级别问题 SHALL 允许写入但必须显示在校验结果中。

#### Scenario: 阻止写入重复 ID

- **GIVEN** 当前表存在两行相同 `id`
- **WHEN** 用户点击 `保存当前表`
- **THEN** 系统显示重复 ID 错误
- **AND** 系统不写入该 Excel 文件

#### Scenario: 阻止写入不存在引用

- **GIVEN** `#Player.xlsx` 中某个选手的 `teamId` 指向不存在的 `TbTeam`
- **WHEN** 用户点击 `保存当前表`
- **THEN** 系统显示 `teamId` 引用不存在错误
- **AND** 系统不写入 `#Player.xlsx`

#### Scenario: 允许带告警保存

- **GIVEN** `#Team.xlsx` 中某个 `logo` 路径无法预览
- **WHEN** 用户点击 `保存当前表`
- **THEN** 系统显示 Logo 路径告警
- **AND** 若其他字段没有 ERROR，系统允许写入 `#Team.xlsx`

### Requirement: 工作台写回 Luban Excel 时保留表头并创建备份

本地写入服务 SHALL 在保存非地图 Luban 表时保留原 workbook 的 Luban 表头、字段类型、导出分组和注释。服务 SHALL 只替换数据区，覆盖前在 `configs/Datas/.bak/` 下创建备份，并拒绝写入项目允许范围外的路径。服务 SHALL 检测读取后发生的外部文件修改并阻止覆盖；原文件缺失、无法解析或损坏时 SHALL 报错，不得自动重建并覆盖原路径。

#### Scenario: 保存 Player 表

- **GIVEN** 用户修改了 `#Player.xlsx` 的选手数据
- **WHEN** 用户点击 `保存当前表`
- **THEN** 服务在 `configs/Datas/.bak/` 下创建原文件备份
- **AND** 服务写回 `configs/Datas/#Player.xlsx`
- **AND** 写回后的 workbook 保留 `##var`、`##type`、`##group` 和 `##` 表头
- **AND** 数据行从第 5 行开始写入

#### Scenario: 拒绝越界文件写入

- **GIVEN** 前端请求保存 `../outside.xlsx`
- **WHEN** 本地写入服务解析文件路径
- **THEN** 服务拒绝该请求
- **AND** 项目根目录外的文件不被创建或修改

#### Scenario: 阻止覆盖外部修改的 workbook

- **GIVEN** 用户打开 `#Player.xlsx` 后，该文件又被 Excel 或其他进程修改
- **WHEN** 用户在工作台点击 `保存当前表`
- **THEN** 服务根据修改时间或内容摘要识别冲突
- **AND** 服务不覆盖文件并提示用户从项目重新读取

#### Scenario: 损坏的 workbook 不被自动重建

- **GIVEN** 目标 `.xlsx` 文件不存在、无法解析或已损坏
- **WHEN** 用户尝试读取或保存该表
- **THEN** 服务返回明确错误
- **AND** 服务不在目标路径创建替代 workbook

### Requirement: 工作台复用 Luban 导表验证入口

工作台 SHALL 继续提供 Web 按钮运行 `scripts/gen-config.ps1`。导表输出 SHALL 显示执行中、成功、失败、退出码、耗时、stdout 和 stderr。导表动作 SHALL 不隐式保存未确认的脏表，除非用户在确认提示中明确同意。

#### Scenario: 保存后运行导表

- **GIVEN** 用户已经保存 `#Player.xlsx`、`#Team.xlsx` 或 `#TutorialBattle.xlsx`
- **WHEN** 用户点击 `运行导表`
- **THEN** 本地服务在项目根目录执行 `scripts/gen-config.ps1`
- **AND** 页面显示导表状态、退出码、耗时、stdout 和 stderr

#### Scenario: 存在未保存修改时运行导表

- **GIVEN** 当前页面存在未保存的表格修改
- **WHEN** 用户点击 `运行导表`
- **THEN** 系统提示存在未保存修改
- **AND** 用户未确认保存或继续时，系统不开始导表

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
