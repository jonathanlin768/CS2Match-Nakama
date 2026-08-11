## 1. 应用外壳与导航

- [x] 1.1 梳理现有 `tools/map-semantic-editor` 页面结构，确认地图页组件、状态和服务 API 的可复用边界
- [x] 1.2 新增配置工作台顶层外壳，提供 `地图配置`、`选手配置`、`战队配置`、`新手战斗`、`全部表格` 页面切换
- [x] 1.3 将现有地图编辑布局封装为 `地图配置` 页面，保持雷达画布、图层、属性面板、底部日志和现有工具栏行为
- [x] 1.4 限制地图快捷键只在 `地图配置` 页面生效，非地图页面保留输入框和表格文本编辑行为
- [x] 1.5 保留根目录 `npm run map-editor` 旧命令，并新增 `npm run config-editor` 别名，UI 使用配置工作台新名称

## 2. 通用 Luban 表模型

- [x] 2.1 定义 `LubanTableDocument`、`LubanField`、`LubanRow`、`LubanCellValue` 和表级 dirty 状态模型
- [x] 2.2 实现 `##var`、`##type`、`##group`、`##` 四行表头解析，生成字段名、类型、分组和注释
- [x] 2.3 实现字段类型解析，覆盖 `string`、`int`、`float`、`bool`、`list`、`(list#sep=,),string`、`string#ref=TbX` 和 `(list#sep=,),string#ref=TbX`
- [x] 2.4 建立 `TbX -> #X.xlsx` 的引用表索引，并定义候选项显示名称优先级：`name`、`shortName`、`nickname`、`id`
- [x] 2.5 为未知或复杂 Luban 类型提供文本 fallback，并保留原始 `##type`

## 3. 本地服务 API

- [x] 3.1 新增读取表列表接口，返回 `configs/Datas/#*.xlsx` 文件、推导表名、行数、解析状态和编辑入口归属；地图语义表及 `#CombatConst.xlsx` 标记为仅由地图配置页编辑
- [x] 3.2 新增读取单表接口，返回表元数据、字段定义、数据行和解析警告
- [x] 3.3 新增保存单表接口，校验路径必须在 `configs/Datas/` 下且文件名必须为允许的 `.xlsx`
- [x] 3.4 实现保存时保留原 workbook、sheet、表头、列宽和样式，只清理并重写第 5 行开始的数据区；缺失、无法解析或损坏的 workbook 直接报错且不自动重建
- [x] 3.5 保存前在 `configs/Datas/.bak/` 下创建备份，并在写入失败时保留原文件
- [x] 3.6 新增批量保存已修改表接口，按表依赖和引用关系处理保存顺序；跨表 ID 同步批次任一写入失败时恢复已写文件
- [x] 3.7 复用现有 `scripts/gen-config.ps1` 运行接口，确保导表状态、退出码、耗时、stdout 和 stderr 可被非地图页消费
- [x] 3.8 保存前比较读取时记录的修改时间或内容摘要，检测外部修改并阻止覆盖
- [x] 3.9 新增受控图片上传接口，仅允许外部图片分别复制到 `client/public/portraits/` 和 `client/public/teams/`，并处理格式校验、路径越界与同名冲突
- [x] 3.10 为越界路径、非法扩展名、缺失 workbook、损坏 workbook、外部修改冲突、图片路径安全和保存失败编写服务端测试

## 4. 通用表格编辑页

- [x] 4.1 实现 `全部表格` 页面，展示 `#*.xlsx` 表列表、行数、脏状态、解析警告和编辑入口归属
- [x] 4.2 实现通用表格视图，支持查看字段注释、字段类型、编辑单元格、新增行、复制行和删除行
- [x] 4.3 根据字段类型选择基础控件：文本、数字、布尔开关、标签列表、单选引用、多选引用和未知类型文本框
- [x] 4.4 允许通用表格页编辑并保存 `#item.xlsx` 等已有非地图表；地图语义表及 `#CombatConst.xlsx` 只显示归属和“前往地图配置”，不提供第二个编辑入口
- [x] 4.5 实现表级 `保存当前表`、`保存全部已修改表` 和未保存修改提示
- [x] 4.6 在底部面板显示通用表校验结果、写入日志和导表输出

## 5. 战队配置页

- [x] 5.1 实现 `战队配置` 页面，读取和编辑 `configs/Datas/#Team.xlsx`
- [x] 5.2 提供战队列表、搜索、当前战队详情表单和新增/复制/删除战队操作
- [x] 5.3 支持编辑 `id`、`name`、`shortName`、`nickname` 和 `logo`
- [x] 5.4 对 `logo` 提供外部图片文件选择，将文件复制到 `client/public/teams/`，写入 `teams/<文件名>`，并提供图片预览、同名处理和缺失告警
- [x] 5.5 显示该战队被 `#Player.xlsx.teamId` 和 `#TutorialBattle.xlsx.opponentTeamId` 引用的数量
- [x] 5.6 被引用的战队禁止直接删除；修改被引用 ID 时显示引用位置，并提供显式的“修改 ID 并同步引用”批次操作

## 6. 选手配置页

- [x] 6.1 实现 `选手配置` 页面，读取和编辑 `configs/Datas/#Player.xlsx`
- [x] 6.2 提供按战队、稀有度和位置标签筛选选手的列表视图
- [x] 6.3 支持编辑选手基础字段：`id`、`name`、`nationality`、`rarity`、`portrait`
- [x] 6.4 使用 `TbTeam` 引用选择器编辑 `teamId`
- [x] 6.5 使用数字输入编辑 `entry`、`aim`、`trade`、`clutch`、`firepower`、`gamesense`、`reaction`、`positioning`、`awareness`、`teamplay`、`utility`、`composure`、`mobility`、`endurance` 和 `discipline`，并限制为 `0..100`
- [x] 6.6 使用标签输入编辑 `positions`，保存为 Luban 列表单元格
- [x] 6.7 对 `portrait` 提供外部图片文件选择，将文件复制到 `client/public/portraits/`，写入 `portraits/<文件名>`，并提供图片预览、同名处理和缺失告警
- [x] 6.8 被引用的选手禁止直接删除；修改被引用 ID 时显示费用档和对手阵容引用，并提供显式的“修改 ID 并同步引用”批次操作

## 7. 新手战斗配置页

- [x] 7.1 实现 `新手战斗` 页面，读取和编辑 `configs/Datas/#TutorialBattle.xlsx`
- [x] 7.2 支持编辑 `id`、`version`、`enabled`、`budget`、`rosterSize` 和 `mapId`
- [x] 7.3 为 `tier5PlayerIds`、`tier4PlayerIds`、`tier3PlayerIds`、`tier2PlayerIds` 和 `tier1PlayerIds` 提供 `TbPlayer` 多选控件
- [x] 7.4 多选候选项显示选手名称、战队、稀有度和位置标签，并支持搜索和筛选
- [x] 7.5 为 `opponentTeamId` 提供 `TbTeam` 单选控件
- [x] 7.6 为 `opponentPlayerIds` 提供 `TbPlayer` 多选控件，并优先显示属于 `opponentTeamId` 的选手
- [x] 7.7 对任一费用档候选池为空显示 ERROR，并阻止保存 `#TutorialBattle.xlsx`
- [x] 7.8 限制整张 `#TutorialBattle.xlsx` 最多一条记录启用；启用新记录时确认并自动关闭原记录
- [x] 7.9 对启用配置严格执行运行时约束：预算大于 0、阵容人数等于 5、费用档选手不重复、候选池可在预算内组成完整阵容、对手阵容人数正确且选手唯一并属于对手战队

## 8. 校验与写入规则

- [x] 8.1 实现通用表校验：空 ID、重复 ID、整数/浮点/布尔类型错误、单选引用不存在、多选引用不存在
- [x] 8.2 实现 Player 表校验：能力值必须为 `0..100` 整数，战队引用必须存在，列表字段必须可序列化
- [x] 8.3 实现 Team 表校验：ID 唯一，Logo 缺失作为 WARNING，不阻止写入
- [x] 8.4 实现 TutorialBattle 表校验：每个费用档至少一个候选选手、最多一条启用配置；启用配置严格对齐 `server/config.Validate()`，所有运行时不合法状态均为 ERROR
- [x] 8.5 ERROR 级别问题阻止保存当前表；WARNING 级别问题允许保存但必须出现在校验结果中
- [x] 8.6 运行导表前检测未保存修改，提示用户保存或确认继续，未确认时不启动导表
- [x] 8.7 实现通用引用反查、被引用记录删除保护、ID 修改范围预览和显式跨表同步校验

## 9. 测试与验证

- [x] 9.1 为 Luban 表头解析、类型解析、引用解析和列表序列化编写单元测试
- [x] 9.2 为 workbook 数据区写回、表头保留、外部修改冲突和损坏文件拒绝重建编写服务端测试，覆盖 `#Player.xlsx`、`#Team.xlsx` 和 `#TutorialBattle.xlsx`
- [x] 9.3 为通用表格页编写组件测试，覆盖编辑 `#item.xlsx`、新增行、删除行、保存按钮状态，以及地图语义表只能跳转到地图配置页
- [x] 9.4 为 Team 专属页编写组件测试，覆盖 Logo 外部文件复制、图片预览、同名处理、Logo 告警、引用数量、删除保护和显式 ID 同步
- [x] 9.5 为 Player 专属页编写组件测试，覆盖 `teamId` 引用选择、能力值 `0..100` 校验、`positions` 标签编辑、头像外部文件复制和显式 ID 同步
- [x] 9.6 为 TutorialBattle 专属页编写组件测试，覆盖费用档多选、唯一启用切换、空费用档阻止保存、预算可行性、跨档重复、对手阵容人数和战队归属错误
- [x] 9.7 为跨表 ID 同步编写服务端测试，覆盖批次备份、引用列表更新和中途失败回滚
- [x] 9.8 运行 `npm --prefix tools/map-semantic-editor test` 和 `npm --prefix tools/map-semantic-editor run build`
- [x] 9.9 保存三张表后运行 `scripts/gen-config.ps1`，确认 Server 和 Client 配置产物成功生成
- [x] 9.10 运行 `go test ./...` 于 `server/config`，确认生成配置加载仍通过
- [x] 9.11 通过项目 Docker 构建流程编译 Nakama Go Plugin，确认 `.so` 构建仍成功
- [x] 9.12 检查本变更未新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移

## 10. 文档与交付

- [x] 10.1 更新工具 README，说明配置工作台页面、启动命令、保存表、备份、导表和第一版限制
- [x] 10.2 更新 OpenSpec 验收说明，记录地图配置页和非地图配置页的边界
- [x] 10.3 记录第一版已知限制：不编辑 `__beans__.xlsx`、`__enums__.xlsx`、`__tables__.xlsx`，地图语义表及 `#CombatConst.xlsx` 不开放通用编辑，未知复杂类型使用文本 fallback
- [x] 10.4 记录回滚方式：隐藏非地图配置导航入口时，现有地图配置页仍可继续使用
