## 上下文

`doc/simuMatchDesign.md` 2.2 已经把第一版地图语义配置拆成 `tb_route_template`、`tb_scenario`、`tb_map_tag`、`tb_encounter_modifier`、`tb_map_node`、`tb_map_edge`、`tb_visibility`、`tb_route` 和 `tb_combat_const` 等表。合并后的 `tb_map_node` 同时承载位置锚点和可选几何范围。现有 Luban 配置使用 `configs/luban.conf` 指向 `configs/Datas/`，数据表文件使用 `configs/Datas/#Player.xlsx`、`#item.xlsx` 这种 auto-import 表头风格，而不是 `configs/Defines/*.xml` 风格。

本变更新增的是本地设计工具，不是在线对局功能。编辑器服务于玩法设计者、程序开发者和前端开发者：设计者在 Dust2 雷达图上编辑语义对象，程序开发者检查配置引用和模型约束，前端开发者预览击杀点、路线和视野表现。编辑器导出的数据必须 1:1 写入当前项目的 Luban Excel 文件结构，并能直接被现有导表脚本消费。

与 Nakama 插件架构的兼容性：本变更不新增 Go module，不修改 `windypath.com/cs2match/<module>` 相关 Go 运行时代码，不注册 RPC、Hook 或 Match Handler，不读写 Nakama Storage。编辑器仅生成离线配置输入，因此不会影响 `InitModule`、MatchLoop、服务器权威状态同步或客户端运行时状态。

## 目标 / 非目标

**目标：**

- 提供一个位于 `tools/map-semantic-editor/` 的本地 Web 编辑器。
- 使用 React 19、TypeScript、Vite 8 复用项目已有前端技术上下文。
- 提供一个随编辑器启动的本地 Node 写入服务，负责安全读取和写入当前项目 `configs/Datas/` 下的 Excel 文件。
- 在雷达底图上编辑地图节点、路径、视野、路线、风险热点和高低层级语义。
- 支持合并后的 `tb_map_node` 配置：节点既有 `x/y` 位置锚点，也可以配置 `shape/radius/points/area_usages` 表达击杀采样、包点、控制区、交火区、声音区和风险区。
- 保存地图工程文件，保留编辑器状态和未导出的草稿信息。
- 按当前项目 Luban 风格直接写入 Excel：更新或生成 `#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx`、`#CombatConst.xlsx`。这些 `#*.xlsx` 文件 SHALL 依赖 Luban auto-import，不重复写入 `__tables__.xlsx`。
- 在导出前执行强校验，并能把错误定位到具体对象和画布位置。

**非目标：**

- 不做真实 3D 地图编辑器。
- 不自动识别地图结构，不自动生成完整寻路网络。
- 不模拟弹道、投掷物轨迹或连续视野传播。
- 不做线上多人协作编辑。
- 不替代 Luban，不修改现有导表器职责。
- 不让用户通过浏览器下载 Excel 后再手动导入 `configs/`。
- 不在本变更中让模拟器运行时直接消费新表；运行时接入应由后续变更处理。
- 不新增 Nakama RPC、Match Handler、Storage 或数据库迁移。

## 决策

### 决策 1：编辑器作为独立本地 Web 工具

编辑器 SHALL 放在 `tools/map-semantic-editor/`，并通过根目录命令 `npm run map-editor` 或等价脚本同时启动 Vite UI 和本地 Node 写入服务。

理由：它是策划和开发期工具，不属于玩家主客户端首屏体验；独立目录可以隔离依赖、构建配置和测试，同时复用项目的 React/Vite/TypeScript 技术栈。

替代方案：放入 `client/src/pages`。该方案会把设计工具混入玩家客户端路由和生产构建，容易扩大包体和权限边界，因此不作为第一版方案。

### 决策 2：直接写入 `configs/Datas/` 需要本地 Node 服务

编辑器 UI SHALL 通过本地 Node 服务完成导出写入。普通浏览器页面不能可靠、无交互地静默写入仓库目录；因此第一版 SHALL 不使用“浏览器下载文件再手动导入”的流程，而是由 Node 服务把导出结果直接写入当前项目对应位置。写入 Luban 成功时，服务 SHALL 同步把完整编辑器工程快照保存为 `configs/Datas/<map_id>.json`，用于后续从发布目录读回编辑器状态。

Node 服务 SHALL：

- 只监听本地地址。
- 将所有写入路径限制在当前项目根目录下的 `configs/Datas/` 和编辑器工程文件允许目录内。
- 在写入前做路径规范化和越界检查。
- 对 Excel 写入提供备份或原子写入策略，避免半写入损坏配表。
- 提供从 `configs/Datas/<map_id>.json` 读取工程快照的接口，便于导出后重新进入编辑器继续维护。

替代方案：浏览器端 ExcelJS 生成 `.xlsx` 并下载。该方案技术可行，但会产生手动导入步骤，不符合当前需求。

### 决策 3：导出遵循当前 Luban Excel 风格

导出 SHALL 使用当前项目的 Luban 文件约定：

- `configs/Datas/#*.xlsx` 保存具体数据，表头使用现有 `##var`、`##type`、`##group`、`##` 风格。
- `configs/Datas/__tables__.xlsx` 只用于非 `#` 前缀文件；地图语义表第一版不写入这里。
- `configs/Datas/__beans__.xlsx` 和 `__enums__.xlsx` 第一版不新增定义；枚举字段先导出为 `string`，后续运行时接入需要强类型枚举时再升级。

地图语义表建议映射：

| 设计表 | 数据文件 |
|---|---|
| `tb_route_template` | `configs/Datas/#RouteTemplate.xlsx` |
| `tb_scenario` | `configs/Datas/#Scenario.xlsx` |
| `tb_map_tag` | `configs/Datas/#MapTag.xlsx` |
| `tb_encounter_modifier` | `configs/Datas/#EncounterModifier.xlsx` |
| `tb_map_node` | `configs/Datas/#MapNode.xlsx` |
| `tb_map_edge` | `configs/Datas/#MapEdge.xlsx` |
| `tb_visibility` | `configs/Datas/#Visibility.xlsx` |
| `tb_route` | `configs/Datas/#Route.xlsx` |
| `tb_combat_const` | `configs/Datas/#CombatConst.xlsx` |

实现前 SHALL 用现有 Luban 表定义最终确认 `full_name`、`value_type`、`input`、`index`、`mode`、`group` 和 `output` 的具体写法。

### 决策 4：画布层使用 Konva / React-Konva

第一版实现 SHALL 优先使用 Konva / React-Konva 处理雷达图、拖拽地图节点、编辑圆形/多边形节点范围、连线、图层和选中框。根据当前编辑器使用反馈，第一版中央 Dust2 雷达画布保持固定自适应展示，不额外提供缩放、平移或辅助线交互。

理由：编辑器核心是 2D 标注和几何交互，Konva 比手写 Canvas 命中测试、拖拽和层级管理成本更低，也比 SVG 更适合中等复杂度画布交互。

替代方案：直接使用 SVG。SVG 对简单点线面足够，但复杂拖拽、对象命中、性能和图层控制会更快变重。

### 决策 5：工程文件与导出模型分离

编辑器 SHALL 定义两层数据模型：

- 工程模型：保存编辑器私有状态，例如图层可见性、锁定状态、颜色、选区、草稿对象、`0..1` 坐标归一化规则和最近预览配置。
- 导出模型：仅包含模拟器和 Luban 需要消费的数据，对齐 `doc/simuMatchDesign.md` 2.2 的表结构。

理由：草稿和 UI 状态不应污染配表；导出模型保持稳定后，后续模拟器接入可以只依赖 Luban 生成结果。

### 决策 6：`tb_map_node` 同时承载位置锚点和可选几何范围

第一版 SHALL 只使用 `tb_map_node` 表达地图地点语义。每个节点都有 `0..1` 归一化 `x/y` 锚点坐标，并可选配置自身几何范围：

- `shape = None` 时，节点只作为路线、路径、视野和默认事件展示锚点使用。
- `shape = Circle` 时，节点范围是以 `x/y` 为中心、`radius` 为半径的圆；`radius` 同样使用 `0..1` 归一化单位。
- `shape = Polygon` 时，节点范围由 `points` 定义，`x/y` 仍作为默认展示中心。
- `area_usages` 包含 `KillSample` 时，击杀显示坐标优先从该节点自身范围内采样。
- `area_usages` 可包含 `Plant`、`Control`、`Encounter`、`Sound`、`Risk` 等用途，用于后续模拟器或编辑器预览消费。
- 编辑器 UI SHALL 将 `area_usages` 包含 `Risk` 的节点作为普通 `MapNode` 管理：它随“地图节点”图层显示、隐藏和锁定，在对象列表中也归入 MapNode 分组。为了避免和普通节点混淆，画布 SHALL 使用独立的高亮颜色或外环符号标识风险热点；导出时仍然写入同一张 `tb_map_node`。风险热点是赛前配置的概率和采样权重输入，不是运行时死亡、掉包或拦截坐标的硬编码来源。
- 节点几何范围不参与真实空间仿真，只表达雷达图展示和地图语义。

导出校验 SHALL 覆盖节点 ID 唯一性、`x/y` 坐标范围、`shape` 字段合法性、圆形半径范围、多边形顶点格式、`area_usages` 合法性，以及关键节点缺少 `KillSample` 范围时的告警。

### 决策 7：校验使用 Zod 和领域规则两层

工程文件导入、导出摘要和表结构输出 SHALL 先通过 Zod schema 做基础类型校验，再执行领域规则校验，例如引用存在性、ID 唯一、路线连通、区域多边形合法性、视野语义完整性和高低层级完整性。

理由：Zod 适合前端运行时解析和错误聚合；领域规则需要访问对象图，不能只靠静态类型完成。

### 决策 8：不接入 Nakama 和 MatchLoop

本变更 SHALL 不新增 RPC、Storage、Match Handler，也不改变 MatchLoop tick 行为。

理由：编辑器是离线配表生产工具，输出的是导表输入。运行时如何用新地图表替换 MVP 内建路线配置，应由后续模拟器配置接入变更处理。

## 界面线框与交互设计

编辑器是高频使用的工作型工具，第一屏 SHALL 直接进入地图编辑工作台，不做营销页或说明页。界面采用「顶部工具栏 + 左侧工程/图层/对象 + 中央雷达画布 + 右侧属性面板 + 底部校验/预览/写入日志」结构。

左侧面板、右侧属性面板和底部面板之间提供拖拽分隔条，分别用于临时调整 `leftPanel` 宽度、`propertyPanel` 宽度和 `bottomPanel` 高度。尺寸只保存在当前页面内存中，刷新页面后恢复默认值；第一版不持久化这些布局尺寸，也不允许用户调整顶部工具栏或中央画布之外的其他布局区域。

### 主界面线框

```text
--------------------------------------------------------------------------------------------------+
| CS2 Map Semantic Editor [新建] [打开] [从项目中读取] [保存] | de_dust2 v1 | 服务: 已连接 [写入 Luban] [运行导表] |
+--------------------------------------------------------------------------------------------------+
| 工具: [选择] [点位 (tb_map_node)] [路径 (tb_map_edge)] [视野 (tb_visibility)] [路线 (tb_route)] |
|       [风险热点 (risk_points)]                        | [撤销] [重做] | [校验] [预览] |
+-----------------------------+------------------------------------------------------+-------------+
| 左侧面板                  ║ | 中央雷达画布                                      ║ | 右侧属性    |
|                             |                                                      |             |
| 项目                        |  +------------------------------------------------+  | 当前对象    |
|  map_id: de_dust2           |  |                                                |  | 类型: 地图节点 |
|  radar: ...dust2.webp       |  |       Dust2 radar image                         |  | id          |
|  坐标: normalized 0..1      |  |                                                |  | name        |
|                             |  |   o T_SPAWN          ---- path ----  o LONG    |  | area        |
| 图层                        |  |        \                              /         |  | site        |
|  [✓] 地图节点 [锁] [色]     |  |         \                            /          |  | node_type   |
|  [✓] 节点范围 [锁] [色]     |  |       polygon: A_SITE sample range              |  | floor       |
|  [✓] 路径    [锁] [色]      |  |             [sample points preview]             |  | x / y       |
|  [✓] 视野    [锁] [色]      |  |                                                |  | shape       |
|  [✓] 路线    [锁] [色]      |  |   visibility: CATWALK -> A_SITE  HighToLow      |  |             |
|                             |  |                                                |  | area_usages |
|                             |  +------------------------------------------------+  | [复制]      |
| 对象列表                    |  鼠标坐标 x=0.623 y=0.411 | snap on | 工具 select          | [删除]      |
|  搜索: [A_SITE________]     |                                                      |             |
|  MapNode                    |                                                      |             |
|   - A_SITE                  |                                                      |             |
|   - CATWALK                 |                                                      |             |
|   - A_SITE                  |                                                      |             |
+-----------------------------+------------------------------------------------------+-------------+
| ↑ 拖拽调整 bottomPanel 高度                                                               |
| 底部面板: [校验结果] [采样预览] [导出摘要] [写入日志]                                      |
| P1 ERROR Route D2_A_SPLIT: nodes[2] 引用缺失 -> [定位]                                      |
| INFO  #MapNode.xlsx 15 rows, auto-import table written                                       |
+--------------------------------------------------------------------------------------------------+
```

### 面板职责

| 区域 | 职责 | 主要操作 |
|---|---|---|
| 顶部标题栏 | 工程级操作和写入状态 | 新建、打开、保存、查看本地服务状态、写入 Luban、运行导表 |
| 顶部工具栏 | 切换编辑模式和画布辅助操作 | 选择、地图节点、路径、视野、路线、风险热点、撤销、重做、校验、预览 |
| 左侧项目区 | 管理地图工程元数据 | 修改地图 ID、版本、底图路径、`0..1` 坐标归一化规则 |
| 左侧图层区 | 控制语义图层显示和编辑权限 | 显示/隐藏、锁定/解锁、图层颜色 |
| 左侧对象列表 | 快速查找和选中对象 | 按类型分组、搜索、点击定位 |
| 中央画布 | 编辑雷达图上的语义对象 | 拖动地图节点、编辑节点范围、连接路径、配置视野、预览采样 |
| 右侧属性面板 | 编辑当前选中对象字段 | 修改 ID、名称、枚举、引用、数值、应用/复制/删除 |
| 底部面板 | 展示校验、预览和写入反馈 | 点击错误定位对象、查看导出摘要和写入日志 |

### 工具栏状态

```text
[选择]    点击点位、节点范围或线段 -> 选中；重叠线段重复点击 -> 在路径/路线候选对象间循环切换；Delete -> 删除；Ctrl+Z -> 撤销；Ctrl+Y -> 重做
[地图节点] 点击画布 -> 创建 MapNode；右侧填写 name/site/node_type/floor/x/y，并可切换 None/Circle/Polygon 范围
[路径]    点击起点，再点击终点 -> 创建 MapEdge；右侧配置 bidirectional/base_time/stamina_cost/risk/noise/risk_points/intercept_nodes；第一版不配置移动方式
[视野]    点击观察点，再点击被观察点 -> 创建 Visibility；右侧配置 range/angle/cover/exposure/elevation
[路线]    按顺序点击多个地图节点 -> 追加到路线草稿；Enter 或 [完成路线] 创建 Route.nodes；Esc 或 [取消路线] 放弃
[风险热点] 在路径或地图节点范围上点击 -> 创建或选择 Risk 用途地图节点；可关联到 MapEdge.risk_points，用于赛前定义高冲突区域并影响模拟概率/权重
快捷键     S=选择，N=点位，E=路径，V=视野，L=路线，R=风险热点；焦点在输入控件内时不触发工具切换
[校验]    运行强校验；结果进入底部面板；阻塞错误会禁用写入 Luban
[预览]    按当前选中对象展示采样点、连通路径、视野线或导出摘要；生成采样点后按钮切换为 [清除预览]
[写入 Luban]  校验通过后调用本地服务写入 configs/Datas，并展示写入日志
[从项目中读取] 从 configs/Datas/<map_id>.json 读取最近一次写入 Luban 时发布的工程快照
[运行导表] 点击后调用本地服务在项目根目录执行 scripts/gen-config.ps1；按钮切换为 [导表中] 并自动打开底部 [导表输出]；底部面板显示 exit code、stdout、stderr 和耗时
```

### 对象属性面板线框

地图节点属性：

```text
+------------------------------+
| 当前对象: MapNode             |
| id            [A_SITE       ] |
| name          [A 包点       ] |
| zone          [A区          ] |
| site          [A v]           |
| node_type     [Site v]        |
| default_side  [CT v]          |
| floor         [Ground v]      |
| x             [0.642]         |
| y             [0.382]         |
| shape         [Polygon v]     |
| radius        [空: polygon]   |
| area_usages   [KillSample, Plant] |
| points        [查看/编辑顶点] |
|                              |
| [划定圆形范围] [划定多边形范围] |
| [完成范围] [取消范围]          |
| [预览采样点] [复制] [删除]    |
+------------------------------+
```

视野属性：

```text
+------------------------------+
| 当前对象: Visibility          |
| from_node     [CATWALK v]     |
| to_node       [A_SITE v]      |
| visible       [是]            |
| direction     [单向 v]        |
| range         [Mid v]         |
| angle_adv     [T v]           |
| elevation     [HighToLow v]   |
| cover_mod     [-8]            |
| exposure_mod  [12]            |
|                              |
| [预览视野线] [应用] [删除]    |
+------------------------------+
```

### 核心操作流

#### 新建 Dust2 工程

```text
点击 [新建]
  -> 弹出确认框
     标题: 新建项目
     正文: 即将新建项目，当前的修改将全部丢失，是否继续？
  -> 点击 [取消] 保留当前编辑状态
  -> 点击 [继续新建] 后才清空当前内存修改
  -> 选择模板 de_dust2
  -> 自动填入 radar_image = client/public/csmaps/de_dust2_radar_trans.webp
  -> 进入画布，默认显示地图节点/节点范围/路径/视野/路线图层
  -> 点击 [保存]，本地服务写入工程文件
```

#### 创建地图节点并配置击杀采样范围

```text
点击 [地图节点]
  -> 在 A 包点默认展示位置点击
  -> 右侧 id 填 A_SITE
  -> node_type 选择 Site
  -> 点击右侧 [划定多边形范围]
  -> 在 A 包点周围逐点绘制节点范围
  -> 点击 [完成范围] 或按 Enter，将顶点写入 points，并自动设置 shape=Polygon
  -> area_usages 选择 KillSample 和 Plant
  -> 点击 [预览采样点]，画布显示该节点范围内随机点
```

#### 创建路径和路线

```text
点击 [路径]
  -> 点击 T_SPAWN
  -> 点击 LONG_DOOR
  -> 右侧配置 base_time/stamina_cost/risk/noise/bidirectional
  -> 重复连接 LONG_DOOR -> A_LONG -> A_SITE

点击 [路线]
  -> 依次点击 T_SPAWN, LONG_DOOR, A_LONG, A_SITE
  -> Enter 或点击 [完成路线] 完成
  -> 右侧配置 side=T, target_site=A, min_players/max_players/style_tags
  -> 点击 [预览]，画布高亮整条路线；断裂处用警告样式标出
```

#### 配置视野和高低关系

```text
点击 [视野]
  -> 点击 CATWALK
  -> 点击 A_SITE
  -> 右侧设置 visible=true, range=Mid, angle_advantage=T
  -> elevation 选择 HighToLow
  -> 点击 [预览视野线]
  -> 画布用带箭头的线表示单向视野，用高度标识区分 HighToLow
```

#### 校验、定位和写入 Luban

```text
点击 [校验]
  -> 底部面板显示 ERROR / WARNING / INFO
  -> 点击任意错误的 [定位]
  -> 画布高亮对应对象
  -> 右侧属性面板显示可修复字段

校验通过后点击 [写入 Luban]
  -> 前端调用本地 Node 写入服务
  -> 服务写入 configs/Datas/#MapNode.xlsx 等文件
  -> 服务同步写入 configs/Datas/de_dust2.json 作为编辑器工程快照
  -> 服务按 auto-import 规则写入 #*.xlsx，不重复注册 __tables__.xlsx
  -> 底部写入日志列出文件、记录数、备份路径
  -> 用户点击 [运行导表]
  -> 本地服务在项目根目录执行 scripts/gen-config.ps1
  -> 底部面板实时或完成后显示导表输出、错误、退出码和耗时
```

### 底部面板线框

```text
+--------------------------------------------------------------------------------+
| [校验结果] [采样预览] [导出摘要] [写入日志] [导表输出]                         |
+--------------------------------------------------------------------------------+
| ERROR  MapNode A_SITE.points 多边形顶点少于 3 个                      [定位]    |
| WARN   Route D2_A_LONG.nodes 缺少可达路径 A_LONG -> A_SITE             [定位]    |
| INFO   MapNode A_SITE: Polygon 6 points, usages=KillSample,Plant               |
+--------------------------------------------------------------------------------+
```

写入日志：

```text
+--------------------------------------------------------------------------------+
| 写入成功: configs/Datas/                                                        |
| #MapNode.xlsx          15 rows   backup: .bak/20260726-.../#MapNode.xlsx        |
| #MapEdge.xlsx          7 rows    backup: .bak/20260726-.../#MapEdge.xlsx        |
| #Visibility.xlsx       1 rows    backup: .bak/20260726-.../#Visibility.xlsx     |
| #Route.xlsx            1 rows    backup: .bak/20260726-.../#Route.xlsx          |
| de_dust2.json          project snapshot, not consumed by Luban                  |
| 下一步: 点击 [运行导表] 执行 scripts/gen-config.ps1                             |
+--------------------------------------------------------------------------------+
```

导表输出：

```text
+--------------------------------------------------------------------------------+
| scripts/gen-config.ps1                                                          |
| 状态: 成功  exit=0  duration=12.4s                                               |
| stdout                                                                          |
|   Generating server config...                                                   |
|   Generating client config...                                                   |
| stderr                                                                          |
|   <empty>                                                                       |
+--------------------------------------------------------------------------------+
```

## 风险 / 取舍

- [风险] 把 `#*.xlsx` 自动导入表重复注册进 `__tables__.xlsx` 会导致 Luban 类型重复或解析失败 → [缓解] 第一版只写 `#*.xlsx`，不写 `__tables__.xlsx`；用现有 `#Player.xlsx` / `#item.xlsx` 抽出写入模板，导出后必须运行 `scripts/gen-config.ps1` 验证。
- [风险] 直接写入项目 Excel 文件可能覆盖人工编辑内容 → [缓解] 写入前保留备份，导出日志展示变更文件列表；只更新地图语义相关表，不改无关表。
- [风险] 本地写入服务带来路径安全风险 → [缓解] 服务只绑定 localhost，所有路径必须 resolve 到项目根和允许目录内，禁止接收任意绝对路径写入。
- [风险] Web 按钮触发脚本可能导致重复导表或长时间占用进程 → [缓解] 本地服务同一时间只允许一个导表进程，执行中禁用按钮，支持超时、退出码、stdout/stderr 回显和失败提示。
- [风险] 工程模型和导出模型双模型导致转换逻辑增加 → [缓解] 使用明确的 `toExportTables(project)` 转换层，并为每张表写单元测试。
- [风险] 高低关系没有真实 3D 数据，语义可能被误解 → [缓解] UI 使用明确枚举和稳定符号表达 `HighToLow`、`LowToHigh`、`SameLevel`、`HeightBlocked`，并在导出摘要中列出关系数量。
- [风险] 第一版直接写 Luban 表后，运行时仍未消费这些新表 → [缓解] 在文档和任务中明确本变更只完成生产和导表闭环，运行时接入另起变更。

## 迁移计划

1. 新增 `tools/map-semantic-editor/` 工具目录、Vite UI、本地 Node 写入服务和根目录启动命令。
2. 基于当前 `configs/Datas` auto-import 风格补齐地图语义 `#*.xlsx` 数据表模板。
3. 实现编辑器工程模型、导出模型、校验器、预览和直接写入 Excel 的导出流程。
4. 使用样例 Dust2 工程导出到 `configs/Datas/`，运行 `scripts/gen-config.ps1` 验证。
5. 本变更不需要生产迁移或回滚；若工具不可用，可移除工具目录和根目录脚本，并恢复导出前备份的 Excel 文件，不影响 Nakama 服务和玩家客户端。

## 待确认问题

- 第一版已选择 `#*.xlsx` auto-import，不在 `__tables__.xlsx` 中手写 `full_name`；后续若改成非 `#` 文件再重新确认命名规则。
- `tb_map_node.points` 的字符串格式需要在实现时固定，例如 `x1,y1;x2,y2;x3,y3`，并确保 Luban 生成后的前后端解析层一致。
