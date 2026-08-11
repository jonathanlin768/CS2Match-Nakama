## Context

当前 `tools/map-semantic-editor` 已经具备本地 Web UI、本地 Node 写入服务、ExcelJS 读写、`configs/Datas/.bak/` 备份、`scripts/gen-config.ps1` 运行和导表结果展示能力。它的现有页面是为 Dust2 地图语义表服务的专业画布：顶部工具栏、左侧图层和对象列表、中间雷达画布、右侧属性面板、底部日志。

非地图配置表与地图语义表的编辑形态不同。`#Player.xlsx` 有大量能力值、`teamId: string#ref=TbTeam` 和 `positions: list,string`；`#Team.xlsx` 更像实体资料；`#TutorialBattle.xlsx` 的 `tier1PlayerIds` 至 `tier5PlayerIds` 和 `opponentPlayerIds` 都是 `list,string#ref=TbPlayer`。如果只把这些表塞进现有地图画布 UI，会让地图工具变得混乱；如果另建工具，又会重复本地写入、备份、导表和路径安全能力。

因此本变更采用“同一工具外壳，多页面配置工作台”的方案：保留地图配置页，新增通用表格页和若干专属配置页。

## Goals / Non-Goals

**Goals:**

- 将当前本地工具升级为 `CS2 Config Editor`，通过顶层页签或路由承载地图配置和非地图 Luban 表配置。
- 提供通用 Luban workbook 读取与写回能力，支持 `#*.xlsx` auto-import 表的表头解析、行数据编辑、备份和导表验证。
- 第一版重点支持 `#Player.xlsx`、`#Team.xlsx`、`#TutorialBattle.xlsx` 的可视化编辑。
- 第一版通用表格页允许编辑 `#item.xlsx` 等其他已有非地图 `#*.xlsx` 表；地图语义表及 `#CombatConst.xlsx` 继续由地图配置页单一维护。
- 为引用字段提供下拉、搜索、多选或拖拽选择体验，减少手写 ID 错误。
- 为未定制的 `#*.xlsx` 表提供通用表格 fallback，保证工具可以扩展到更多配置表。
- 保持现有地图配置页行为、快捷键、画布和 Luban 写入流程可用。

**Non-Goals:**

- 不在第一版实现多人协作、权限系统、远程 Web 部署或浏览器下载导入流程。
- 不直接编辑 `__beans__.xlsx`、`__enums__.xlsx` 或 `__tables__.xlsx`；第一版只消费其含义或保持现有表结构。
- 不新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移。
- 不让通用表格编辑器替代所有专属体验；复杂表仍允许后续逐张做定制页面。
- 不在运行时热更新线上比赛配置；配置仍通过 Luban 导表和正常构建/加载链路进入运行时。

## Decisions

### 决策 1：把工具升级为配置工作台，而不是创建第二个工具

采用当前 `tools/map-semantic-editor` 作为基础，逐步重命名 UI 文案为 `CS2 Config Editor` 或相近名称。第一版 SHALL 保留 `npm run map-editor` 旧命令，并新增更语义化的 `npm run config-editor` 别名；旧命令和新 UI 名称并存。

理由：

- 现有本地服务已经具备写项目文件的安全边界、备份和导表能力。
- 地图配置和非地图配置最终都写入 `configs/Datas/#*.xlsx`，底层能力一致。
- 策划只需要启动一个本地工具，不需要在多个 UI 之间切换。

替代方案：

- 新建 `tools/luban-config-editor`。优点是概念干净；缺点是重复服务端写入、导表和测试，且地图页之后仍需要集成。
- 直接在 Excel 中继续维护。实现成本为零；但无法解决引用错误、列表字段难维护和专属配置页体验问题。

### 决策 2：顶层导航使用页面级切换，地图页不嵌入通用表格布局

建议 UI：

```text
+--------------------------------------------------------------------------------+
| CS2 Config Editor  [地图配置] [选手配置] [战队配置] [新手战斗] [全部表格]        |
| 服务: 已连接                                      [保存当前表] [运行导表]        |
+--------------------------------------------------------------------------------+
|                                                                                |
|  当前页面内容区域                                                               |
|                                                                                |
+--------------------------------------------------------------------------------+
| [校验结果] [写入日志] [导表输出]                                                |
+--------------------------------------------------------------------------------+
```

`地图配置` 页保留现有雷达画布结构：

```text
+--------------------------------------------------------------------------------+
| [地图配置]                                                                      |
+------------------------+-----------------------------+-------------------------+
| 工程/图层/对象列表     | Dust2 雷达画布               | 地图对象属性            |
|                        |                             |                         |
+------------------------+-----------------------------+-------------------------+
| 校验结果 / 预览 / 写入日志 / 导表输出                                            |
+--------------------------------------------------------------------------------+
```

非地图表页使用工作台布局：

```text
+--------------------------------------------------------------------------------+
| [选手配置] 表: #Player.xlsx  行: 20  状态: 有未保存修改                          |
+----------------------+--------------------------------+------------------------+
| 筛选/分组             | 选手列表或表格                  | 当前选手属性            |
| - 战队                | id / name / team / rarity       | teamId [下拉: TbTeam]  |
| - 稀有度              | 能力值摘要                      | aim [ 92 ]             |
| - 位置                |                                | positions [Entry,AWPer]|
+----------------------+--------------------------------+------------------------+
| 校验结果 / 写入日志 / 导表输出                                                   |
+--------------------------------------------------------------------------------+
```

理由：

- 地图页需要画布、图层和空间选择；通用表页需要密集列表、筛选和属性表单。
- 页面级切换避免把工具栏按钮堆成不可扫描的一长串。
- 后续可按路由拆分 bundle 或测试。

### 决策 3：底层新增通用 Luban table model

新增一个与地图工程模型分离的通用模型：

```ts
type LubanTableDocument = {
  fileName: string
  tableName: string
  fields: LubanField[]
  rows: LubanRow[]
  originalMeta: WorkbookMeta
  dirty: boolean
}

type LubanField = {
  key: string
  type: string
  group: string
  comment: string
  kind: 'string' | 'int' | 'float' | 'bool' | 'list' | 'ref' | 'unknown'
  refTable?: string
  list?: boolean
}
```

读取时解析第 1-4 行表头，从第 5 行开始读取数据。写回时保留原 workbook、sheet、表头、列宽和样式，只清理并重写数据区。读取时记录文件修改时间或内容摘要，保存前再次比较；若文件已被 Excel 或其他进程修改，系统阻止覆盖并要求用户重新读取。原文件不存在、无法解析或损坏时直接报错，不自动重建 workbook。

理由：

- 通用表不能像地图表一样总是重建 workbook，否则容易丢失策划在 Excel 中调整的列宽、注释和样式。
- `#Player/#Team/#TutorialBattle` 的字段结构已经能通过 `##type` 推导出基础控件。

### 决策 4：先做三张表的专属页面，其他表使用通用表格 fallback

第一版专属页面：

```text
#Player.xlsx
- 选手列表、按战队筛选、按稀有度筛选
- teamId 使用 TbTeam 引用下拉
- entry/aim/trade/... 使用数值输入，统一强制限制在 0..100，超出范围为 ERROR
- positions 使用标签输入
- portrait 支持选择电脑上的外部图片，由本地服务复制到 `client/public/portraits/`，字段写入 `portraits/<文件名>`，并提供路径文本、图片预览和缺失告警

#Team.xlsx
- 战队列表
- name/shortName/nickname/logo 表单
- logo 支持选择电脑上的外部图片，由本地服务复制到 `client/public/teams/`，字段写入 `teams/<文件名>`，并提供路径文本、图片预览和缺失告警
- 被 Player/TutorialBattle 引用时提示引用数量

#TutorialBattle.xlsx
- enabled/budget/rosterSize/mapId 基础配置
- tier1PlayerIds 至 tier5PlayerIds 使用 TbPlayer 多选
- opponentTeamId 使用 TbTeam 下拉
- opponentPlayerIds 使用 TbPlayer 多选，并可按 opponentTeamId 过滤
- 同一张表最多允许一条记录 enabled=true；启用另一条记录时先确认，再自动关闭原记录
```

通用表格 fallback：

```text
+--------------------------------------------------------------------------------+
| 表: #Any.xlsx  [新增行] [复制行] [删除行] [保存当前表]                           |
+--------------------------------------------------------------------------------+
| id | name | ...                                                                 |
|----|------|---------------------------------------------------------------------|
|    |      |                                                                     |
+--------------------------------------------------------------------------------+
| 右侧或弹窗显示当前单元格类型、注释、引用目标和校验信息                           |
+--------------------------------------------------------------------------------+
```

通用表格页第一版 SHALL 允许编辑所有已支持基础类型的非地图 `#*.xlsx`，包括 `#item.xlsx` 等已有非地图表；未知复杂类型字段使用文本 fallback，不因没有专属页而变成只读。

地图语义表 `#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx` 和 `#CombatConst.xlsx` 不在通用表格中直接编辑。`全部表格` 可以显示这些表的存在和归属，但点击后引导用户前往 `地图配置`；`#CombatConst.xlsx` 不提供双入口，继续由地图配置页读写。

理由：

- 三张表直接服务当前玩法闭环：战队、选手、新手战斗配置。
- 未定制表仍可编辑，避免每新增一张表都必须先开发页面。

### 决策 5：引用字段从 `##type` 推导，并以已加载表作为候选源

字段解析规则：

- `string#ref=TbTeam` 识别为单选引用，候选来自 `#Team.xlsx`。
- `(list#sep=,),string#ref=TbPlayer` 识别为多选引用，候选来自 `#Player.xlsx`。
- `int`、`float`、`bool`、`string` 使用对应基础控件。
- `list,string` 和 `(list#sep=,),string` 使用标签/多值输入。
- 未识别复杂类型保留文本编辑，并在字段提示中显示原始 `##type`。

引用解析需要建立 `tableName -> document`、`generated table name -> file`、`field id -> display label` 的索引。第一版可以约定候选显示优先级为 `name`、`shortName`、`nickname`、`id`。

### 决策 6：保存表和运行导表解耦

非地图页提供：

- `保存当前表`：只写回当前 `#*.xlsx`。
- `保存全部已修改表`：按依赖顺序写回多个 `#*.xlsx`。
- `运行导表`：执行 `scripts/gen-config.ps1`，不自动隐式保存未确认的脏表，除非 UI 明确提示并由用户确认。

理由：

- 直接导表前自动写入所有修改风险较高。
- 用户可以先保存一张表，再观察写入日志和备份路径。

### 决策 7：引用对象的删除受保护，ID 修改必须显式同步

- Team 或 Player ID 被其他表引用时，普通删除操作 SHALL 被阻止，并列出引用表、字段和记录 ID。
- 直接编辑被引用 ID 时，UI SHALL 提示引用位置，不得静默修改引用方。
- 用户明确执行 `修改 ID 并同步引用` 后，系统同时更新目标 ID 和通过 `##type` 发现的全部引用单元格。
- 该同步操作 SHALL 在所有受影响文件写入前创建备份，并作为一个批次提交；任一校验或写入失败时恢复本批次已写文件，避免出现一半新 ID、一半旧引用的状态。

理由：

- 禁止静默级联可以让用户看清修改范围。
- 显式同步又避免要求用户手工搜索所有单选和列表引用。
- ID 与引用必须作为同一个提交单元，否则 Luban 导表可能在中间状态失败。

### 决策 8：校验分为通用校验和表级业务校验

通用校验：

- 必填 `id` 为空。
- `id` 重复。
- `int/float/bool` 类型不合法。
- `ref` 指向不存在记录。
- `list ref` 中存在不存在记录。

表级校验：

- Player 能力值第一版统一限制在 `0..100`，低于 0、高于 100 或非整数均为 ERROR。
- TutorialBattle 整张表最多允许一条记录 `enabled=true`；启用另一条记录时必须由用户确认关闭原记录。
- TutorialBattle 的 `tier1PlayerIds` 至 `tier5PlayerIds` 每个档位必须至少有一个候选选手；任一档为空为 ERROR。
- 对启用的 TutorialBattle，`budget` 必须大于 0，`rosterSize` 必须等于 5，所有选手引用必须存在，选手不得跨费用档重复。
- 对启用的 TutorialBattle，候选池必须至少能组成 `rosterSize` 人阵容，并且最低可选阵容总费用不得超过 `budget`。
- 对启用的 TutorialBattle，`opponentTeamId` 必须存在，`opponentPlayerIds` 数量必须等于 `rosterSize`，不得重复，且每名对手选手必须存在并属于该战队；不满足时为 ERROR。

校验失败是否阻止写入按严重级别决定：ERROR 阻止，WARNING 允许写入但提示。

### 决策 9：外部图片复制到受控的项目资源目录

- Player 头像只允许复制到 `client/public/portraits/`，Team Logo 只允许复制到 `client/public/teams/`。
- 服务端只接受常见 Web 图片格式，并拒绝路径穿越、非图片内容和写入上述目录之外的请求。
- 默认保留所选文件名；目标文件同名时要求用户确认覆盖或重新命名。
- Excel 中只保存 `portraits/<文件名>` 或 `teams/<文件名>`，不保存本机绝对路径。

### 决策 10：继续保持本地离线工具边界

该工具只读写项目文件，不访问 Nakama，不新增 RPC，不接入 Storage，不改变 MatchLoop。运行时是否读取新配置由已有 Luban 导出和服务器加载逻辑决定。

## Risks / Trade-offs

- [风险] 通用写回破坏 Excel 样式或表头 → [缓解] 读取原 workbook 后只替换数据区；增加回归测试验证表头、字段、行数和已有列宽尽量保留。
- [风险] Luban 类型语法复杂，第一版解析不完所有类型 → [缓解] 第一版覆盖 primitive、list、ref；未知类型回退文本编辑并保留原始值。
- [风险] 引用选择器加载多个表后状态复杂 → [缓解] 服务端提供表元数据和行数据 API，前端建立只读引用索引，保存时按表独立提交。
- [风险] 专属页面和通用页面产生两套编辑逻辑 → [缓解] 专属页面只使用通用 table document 的读写 API，控件层定制，不复制 Excel 写入逻辑。
- [风险] `#TutorialBattle.xlsx` 多选字段较长，普通下拉体验差 → [缓解] 使用搜索式多选、按战队/稀有度过滤，并提供已选列表。
- [风险] 地图页和新配置页快捷键冲突 → [缓解] 仅地图页启用 `S/N/E/V/L/R` 画布工具快捷键；表格页保留浏览器文本编辑快捷键。
- [风险] 通用表格绕过地图强校验并造成工程 JSON 与 Excel 不一致 → [缓解] 地图语义表及 `#CombatConst.xlsx` 只允许从地图配置页编辑。
- [风险] Excel 在页面打开后被外部修改 → [缓解] 保存时比较修改时间或内容摘要，冲突时阻止覆盖并要求重新读取。
- [风险] 跨表 ID 同步只写入部分文件 → [缓解] 写入前逐表备份，批次失败时恢复已写文件。
- [风险] 全量 `openspec validate --all --strict` 可能暴露既有旧 spec 问题 → [缓解] 本变更至少保证新增/修改 capability 的严格校验通过，并记录既有失败项。

## Migration Plan

1. 保留现有 `npm run map-editor` 命令，新增别名 `npm run config-editor`，UI 使用配置工作台新名称。
2. 抽出当前地图页为 `MapConfigPage`，把现有 `Toolbar/LeftPanel/RadarCanvas/PropertiesPanel/BottomPanel` 组合迁入该页。
3. 新增应用外壳 `ConfigEditorShell`，提供顶层导航和共享服务状态。
4. 新增通用 Luban table server API，再接入通用表格页。
5. 按顺序实现 `Team`、`Player`、`TutorialBattle` 专属页，因为 Player 和 TutorialBattle 都依赖 Team/Player 引用索引。
6. 复用现有导表输出面板，必要时把底部日志拆成全局日志和页面日志。
7. 更新 README 和 OpenSpec 主规约。

回滚策略：保留地图页和现有 `/api/luban/write` 地图写入接口；如果非地图配置页出问题，可以隐藏新导航入口，不影响现有地图配置流程。

## 验收说明

- 配置工作台已在同一个本地 Web 应用中提供地图、Player、Team、TutorialBattle 和全部表格五个页面；原地图状态、画布、快捷键、保存和导表行为保持在 `地图配置` 页面内。
- `#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx` 和 `#CombatConst.xlsx` 被明确标记为地图页所有，通用表格页只能跳转，不能编辑。
- `#Player.xlsx`、`#Team.xlsx` 和 `#TutorialBattle.xlsx` 使用专属页面；`#item.xlsx` 等其他非地图 auto-import 表使用通用表格页面。所有页面共享同一套 workbook 读取、校验、备份、并发冲突检测和写回服务。
- 实现未新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移；运行时边界保持不变。
- 验收已运行编辑器 ESLint、49 个 Vitest 测试、Vite 生产构建、Docker Luban 导表、`server/config` Go 测试，以及 Nakama 3.30.0 pluginbuilder Docker 插件编译。

## Open Questions

当前第一版范围内无阻塞开放问题。已确认决策：

- 通用表格页允许编辑 `#item.xlsx` 等已有非地图表；地图语义表及 `#CombatConst.xlsx` 只在地图配置页编辑，不使用双入口。
- Player 能力值统一限定 `0..100`。
- `portrait` 和 `logo` 第一版支持选择外部文件、复制到项目资源目录并预览。
- 保存 `#TutorialBattle.xlsx` 时严格对齐当前运行时约束，并且整张表最多启用一条配置。
- 被引用的 Team/Player 不允许直接删除；ID 修改通过显式的跨表同步操作完成。
- 第一版保留旧命令 `npm run map-editor`，同时新增 `npm run config-editor`；UI 使用新配置工作台名称。
