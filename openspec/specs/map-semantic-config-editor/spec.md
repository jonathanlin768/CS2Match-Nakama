# map-semantic-config-editor Specification

## Purpose
TBD - created by archiving change add-map-semantic-config-editor. Update Purpose after archive.
## Requirements
### Requirement: 编辑器提供本地地图工程管理

地图语义配置编辑器 SHALL 支持创建、打开、保存地图工程文件。工程文件 SHALL 保存地图 ID、地图名称、地图版本、雷达底图路径、`0..1` 坐标归一化规则、图层状态、选中对象、草稿对象和画布状态等编辑器状态。

#### Scenario: 新建并保存地图工程

- **GIVEN** 用户打开地图语义配置编辑器
- **WHEN** 用户点击 `新建`
- **THEN** 系统弹出标题为 `新建项目` 的确认对话框
- **AND** 对话框正文显示 `即将新建项目，当前的修改将全部丢失，是否继续？`
- **AND** 用户确认后系统才新建 `de_dust2` 地图工程
- **WHEN** 用户保存该工程
- **THEN** 系统生成一个地图工程文件
- **AND** 工程文件包含 `map_id`、`name`、`version`、`radar_image`、`coordinate_normalization` 和图层状态
- **AND** 当同名工程文件已存在时，系统在覆盖前创建本地备份

#### Scenario: 打开已有地图工程

- **GIVEN** 用户已有一个地图工程文件
- **WHEN** 用户在编辑器中打开该文件
- **THEN** 系统恢复雷达底图、地图节点、路径、视野、路线、图层状态、选中对象和画布视图状态

#### Scenario: 从发布目录读回地图工程

- **GIVEN** `configs/Datas/de_dust2.json` 存在且是合法地图工程快照
- **WHEN** 用户点击 `从项目中读取`
- **THEN** 系统从该发布快照恢复地图节点、路径、视野、路线、图层状态和画布视图状态
- **AND** 系统不需要从 Luban Excel 反向推导编辑器工程状态

### Requirement: 编辑器显示可交互雷达画布

编辑器 SHALL 在画布中显示中央 Dust2 地图雷达底图，并支持对象选中、拖动、复制、删除、撤销、重做、网格、吸附和 `0..1` 归一化坐标显示。第一版 SHALL 不要求额外实现缩放、平移、居中、重置视图或辅助线交互。

编辑器 SHALL 允许用户在当前会话中调整 `bottomPanel` 高度、左侧 `sidePanel leftPanel` 宽度和右侧 `sidePanel propertyPanel` 宽度。该尺寸状态 SHALL 只保存在浏览器运行时内存中，不写入工程文件、`configs/Datas/<map_id>.json` 或浏览器持久化存储；用户刷新页面后 SHALL 恢复默认尺寸。除这三个面板尺寸外，编辑器 SHALL 不提供其他布局区域的尺寸调整能力。

#### Scenario: 在雷达图上移动地图节点

- **GIVEN** 用户已加载 `client/public/csmaps/de_dust2_radar_trans.webp`
- **AND** 画布中存在一个地图节点
- **WHEN** 用户拖动该节点锚点
- **THEN** 节点在画布中的显示位置更新
- **AND** 右侧属性面板显示更新后的 `0..1` 归一化 `x` / `y` 坐标

#### Scenario: 撤销对象移动

- **GIVEN** 用户刚刚移动了一个地图节点
- **WHEN** 用户执行撤销
- **THEN** 节点回到移动前的位置
- **AND** 工程数据中的节点坐标同步恢复

#### Scenario: 使用单键快捷键切换工具

- **GIVEN** 用户焦点不在输入框、文本域、下拉框或可编辑文本中
- **WHEN** 用户按下 `S`、`N`、`E`、`V`、`L` 或 `R`
- **THEN** 编辑器分别切换到 `选择`、`点位`、`路径`、`视野`、`路线` 或 `风险热点` 工具
- **AND** `Ctrl+Z`、`Ctrl+Y`、`Enter`、`Esc` 和 `Delete` 等已有快捷键语义保持可用

#### Scenario: 临时调整面板尺寸

- **GIVEN** 用户打开地图语义配置编辑器
- **WHEN** 用户拖动左侧面板、右侧属性面板或底部面板边界
- **THEN** 系统只调整对应的 `leftPanel` 宽度、`propertyPanel` 宽度或 `bottomPanel` 高度
- **AND** 中央画布跟随剩余空间重新布局
- **AND** 用户刷新页面后这些尺寸恢复默认值

### Requirement: 编辑器提供语义图层管理

编辑器 SHALL 至少提供地图节点层、节点范围层、路径层、视野层和路线层。每个图层 SHALL 支持显示、隐藏、锁定和颜色区分。节点范围层 SHALL 渲染 `tb_map_node.shape`、`radius` 和 `points` 定义的可选几何范围。`area_usages` 包含 `Risk` 的风险热点节点 SHALL 随地图节点层管理，不提供独立风险热点图层开关。

#### Scenario: 隐藏节点范围图层

- **GIVEN** 画布中存在多个带几何范围的地图节点
- **WHEN** 用户隐藏节点范围层
- **THEN** 画布不再显示圆形或多边形范围
- **AND** 地图节点锚点仍可显示
- **AND** 节点范围数据仍保留在地图工程文件中

#### Scenario: 锁定路径图层

- **GIVEN** 画布中存在路径对象
- **WHEN** 用户锁定路径层
- **THEN** 用户不能在画布上拖动、删除或编辑路径对象
- **AND** 其他未锁定图层仍可编辑

### Requirement: 编辑器支持地图节点编辑

编辑器 SHALL 使用 `tb_map_node` 作为唯一的地图地点表。地图节点 SHALL 支持位置锚点和可选几何范围，字段覆盖 `id`、`map_id`、`name`、`zone`、`site`、`node_type`、`default_side`、`x`、`y`、`floor`、`area_usages`、`shape`、`radius` 和 `points`。`x/y`、`radius` 和 `points` 坐标 SHALL 统一使用 `0..1` 归一化雷达图坐标。

#### Scenario: 编辑地图节点基础属性

- **GIVEN** 用户选中一个地图节点
- **WHEN** 用户修改节点名称、类型、宏观区域、阵营倾向、层级和 `x/y`
- **THEN** 工程数据中的节点属性更新
- **AND** 画布中的节点标记使用对应图层颜色和节点符号展示

#### Scenario: 为地图节点绘制多边形范围

- **GIVEN** 用户选中地图节点 `A_SITE`
- **WHEN** 用户点击右侧属性面板的 `划定多边形范围`，在雷达图上逐点绘制多边形，并点击 `完成范围` 或按 `Enter`
- **THEN** 系统将多边形顶点写入该节点的 `points`
- **AND** 系统将该节点的 `shape` 设为 `Polygon`
- **AND** 用户按 `Esc` 或点击 `取消范围` 时，未完成的顶点不会写入 `points`
- **AND** 该节点仍使用同一条 `tb_map_node` 记录导出

#### Scenario: 为地图节点划定圆形范围

- **GIVEN** 用户选中地图节点 `A_SITE`
- **WHEN** 用户点击右侧属性面板的 `划定圆形范围`，并在雷达图上点击圆形边界点
- **THEN** 系统根据 `A_SITE.x/y` 与边界点距离计算并写入 `radius`
- **AND** 系统将该节点的 `shape` 设为 `Circle`
- **AND** 系统清空该节点的 `points`

#### Scenario: 为地图节点配置范围用途

- **GIVEN** 地图节点 `A_SITE` 已配置有效多边形范围
- **WHEN** 用户将 `area_usages` 设置为 `KillSample` 和 `Plant`
- **THEN** 击杀点预览使用该节点自身范围采样
- **AND** 导出数据中 `tb_map_node.area_usages` 保留这两个用途

#### Scenario: 风险热点随地图节点图层管理

- **GIVEN** 地图节点 `A_LONG_CROSS` 的 `area_usages` 包含 `Risk`
- **WHEN** 用户隐藏 `地图节点` 图层
- **THEN** 画布不再显示该风险热点标记
- **AND** 用户隐藏 `节点范围` 图层时不再显示该风险热点的几何范围
- **AND** 左侧对象列表在 MapNode 分组中显示该节点，并通过 `area_usages=Risk` 识别其风险热点用途
- **AND** 当地图节点层可见时，该风险热点使用区别于普通地图节点的高亮颜色或外环符号展示
- **AND** 导出数据仍将该节点写入 `tb_map_node`

#### Scenario: 点击节点范围选中地图节点

- **GIVEN** 用户处于选择工具，且节点范围图层可见且未锁定
- **AND** 地图节点 `A_SITE` 配置了圆形或多边形范围
- **WHEN** 用户点击该范围内部
- **THEN** 系统选中 `A_SITE` 对应的 `tb_map_node`
- **AND** 右侧属性面板显示该地图节点字段

### Requirement: 编辑器支持地图节点范围采样

编辑器 SHALL 支持在 `tb_map_node` 中定义可选采样范围。`shape = Circle` 时使用 `x/y` 和 `radius` 采样；`shape = Polygon` 时使用 `points` 采样；`shape = None` 时使用 `x/y` 和极小偏移预览。

#### Scenario: 圆形节点范围采样

- **GIVEN** 一个 `shape` 为 `Circle` 且 `area_usages` 包含 `KillSample` 的地图节点
- **WHEN** 用户请求预览随机采样点
- **THEN** 所有采样点落在 `x/y` 和 `radius` 定义的圆形范围内
- **AND** 采样点使用 `0..1` 归一化雷达图坐标展示

#### Scenario: 多边形节点范围采样

- **GIVEN** 一个 `shape` 为 `Polygon` 且 `area_usages` 包含 `KillSample` 的地图节点
- **WHEN** 用户请求预览随机采样点
- **THEN** 所有采样点落在 `points` 定义的多边形内部
- **AND** 若多边形非法，系统显示可定位到该节点对象的校验错误

#### Scenario: 无范围节点采样回退

- **GIVEN** 地图节点 `LONG_DOOR` 的 `shape` 为 `None`
- **WHEN** 用户请求击杀点预览
- **THEN** 画布在 `LONG_DOOR.x/y` 附近显示极小偏移采样点
- **AND** 系统在导出摘要中提示该节点没有专用 `KillSample` 范围

### Requirement: 编辑器支持路径、视野、路线和风险引用地图节点

路径、视野、路线和风险热点 SHALL 统一引用 `tb_map_node.id`。`tb_map_edge.from_node/to_node`、`tb_visibility.from_node/to_node`、`tb_route.nodes`、`tb_map_edge.risk_points` 和 `tb_map_edge.intercept_nodes` SHALL 不引用任何独立区域表。

#### Scenario: 路径连接两个地图节点

- **GIVEN** 画布中存在 `T_SPAWN` 和 `LONG_DOOR` 两个地图节点
- **WHEN** 用户使用路径工具连接它们
- **THEN** 系统创建一条 `tb_map_edge` 记录
- **AND** `from_node` 和 `to_node` 字段引用对应的 `tb_map_node.id`

#### Scenario: 路线按多个地图节点创建

- **GIVEN** 画布中存在 `T_SPAWN`、`LONG_DOOR`、`A_LONG` 和 `A_SITE` 多个地图节点
- **WHEN** 用户使用路线工具依次点击这些地图节点，并按 `Enter` 或点击 `完成路线`
- **THEN** 系统创建一条 `tb_route` 记录
- **AND** `nodes` 字段按点击顺序写入 `T_SPAWN,LONG_DOOR,A_LONG,A_SITE`
- **AND** 用户按 `Esc` 或点击 `取消路线` 时，未完成的路线草稿不会写入 `tb_route`

#### Scenario: 路线按节点序列预览

- **GIVEN** 用户配置了一条包含多个地图节点的路线
- **WHEN** 用户打开路线连通预览
- **THEN** 画布按 `tb_route.nodes` 顺序高亮节点和路径
- **AND** 若路线存在断裂，系统在断裂位置显示预览警告

#### Scenario: 选择重叠的路径和路线

- **GIVEN** 用户处于选择工具，且同一段画布线段同时命中 `tb_map_edge` 和 `tb_route`
- **WHEN** 用户重复点击该线段
- **THEN** 系统在可见且未锁定图层中的候选路径和路线对象之间循环切换选中状态
- **AND** 隐藏或锁定的路径层、路线层不参与画布线段命中

#### Scenario: 风险热点引用地图节点

- **GIVEN** 地图节点 `A_LONG_CROSS` 的 `area_usages` 包含 `Risk`
- **WHEN** 用户将它加入某条路径的 `risk_points`
- **THEN** 导出的 `tb_map_edge.risk_points` 引用 `A_LONG_CROSS`
- **AND** 该节点作为赛前风险热点影响模拟概率和运行时事件位置采样候选，不直接强制决定最终事件坐标

### Requirement: 编辑器提供轻量预览能力

编辑器 SHALL 提供击杀点随机采样、路线连通、视野线、高低关系和导出数据摘要预览。预览 SHALL 只使用编辑器工程数据，不运行完整模拟器。

#### Scenario: 预览击杀采样点

- **GIVEN** 一个地图节点配置了 `KillSample` 用途和有效几何范围
- **WHEN** 用户请求预览 20 个随机采样点
- **THEN** 画布在该节点范围内显示 20 个采样点
- **AND** 采样点坐标均落在节点几何范围内部

#### Scenario: 清除击杀采样点预览

- **GIVEN** 画布已经显示击杀采样点预览
- **WHEN** 用户点击切换后的 `清除预览` 按钮
- **THEN** 画布移除所有采样点预览
- **AND** 该按钮恢复显示为 `预览`

#### Scenario: 预览视野高低关系

- **GIVEN** 用户创建了从 `CATWALK` 到 `A_SITE` 的视野关系
- **WHEN** 用户将高低关系设置为 `HighToLow` 并打开预览
- **THEN** 画布使用可区分的视觉样式展示该关系
- **AND** 属性面板记录该视野关系的距离、角度优势、遮挡和暴露修正

### Requirement: 编辑器执行导出前强校验

编辑器 SHALL 在写入 Luban 配表前执行强校验。校验 SHALL 覆盖引用对象不存在、ID 重复、路线断裂、关键地图节点缺少坐标、节点几何非法、路径方向配置冲突、视野关系缺少必要语义、高低层级关系配置不完整，以及导出数据无法满足 `doc/simuMatchDesign.md` 2.2 模型约束。

#### Scenario: 校验发现缺失引用

- **GIVEN** 一条路线引用了不存在的地图节点 ID
- **WHEN** 用户运行导出前校验
- **THEN** 校验结果包含缺失引用错误
- **AND** 错误项显示对象 ID、字段名和缺失的引用 ID

#### Scenario: 校验发现非法节点多边形

- **GIVEN** 地图节点 `A_SITE` 的 `shape` 为 `Polygon` 但 `points` 少于 3 个顶点
- **WHEN** 用户运行导出前校验
- **THEN** 校验结果包含非法节点几何错误
- **AND** 系统阻止写入 Luban 配表

#### Scenario: 校验发现路线断裂或路径方向冲突

- **GIVEN** 路线 `D2_A_LONG` 的相邻节点为 `A_LONG -> A_SITE`
- **WHEN** 两个节点之间不存在可用 `tb_map_edge`，或只存在 `A_SITE -> A_LONG` 的单向路径
- **THEN** 校验结果包含路线断裂或路径方向冲突错误
- **AND** 系统阻止写入 Luban 配表

#### Scenario: 校验发现孤立关键地图节点

- **GIVEN** 地图节点 `A_SITE` 的 `node_type` 为 `Site` 但没有任何路径连接
- **WHEN** 用户运行导出前校验
- **THEN** 校验结果包含孤立关键节点告警
- **AND** 用户可以根据该告警决定是否补充路径

#### Scenario: 从错误列表定位对象

- **GIVEN** 校验结果中存在一个地图节点几何错误
- **WHEN** 用户点击该错误项
- **THEN** 画布高亮对应地图节点
- **AND** 属性面板显示该节点的可编辑字段

### Requirement: 编辑器直接写入当前 Luban 配表目录

编辑器 SHALL 将工程数据转换为当前项目 Luban 可消费的 Excel 配表，并通过本地写入服务直接写入 `configs/Datas/`。导出 SHALL 遵循当前项目的 `#*.xlsx` auto-import 文件风格，不要求用户在浏览器下载后手动导入。

#### Scenario: 写入通过校验的 Excel 配表

- **GIVEN** 地图工程已通过导出前强校验
- **WHEN** 用户执行写入 Luban 配表
- **THEN** 本地写入服务更新或生成 `configs/Datas/#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx` 和 `#Route.xlsx` 等地图语义表
- **AND** 本地写入服务不把这些 `#*.xlsx` 表重复注册到 `configs/Datas/__tables__.xlsx`
- **AND** 导出日志列出每个写入文件的路径和记录数量
- **AND** 本地写入服务同步保存 `configs/Datas/de_dust2.json` 工程快照，作为后续读回编辑器的来源

#### Scenario: 未通过校验时禁止写入

- **GIVEN** 地图工程存在 ID 重复错误
- **WHEN** 用户执行写入 Luban 配表
- **THEN** 系统阻止写入任何 Excel 文件
- **AND** 校验结果列表显示阻止写入的错误

#### Scenario: 写入后从编辑器运行 Luban 导表

- **GIVEN** 编辑器已经将地图语义配表写入 `configs/Datas/`
- **WHEN** 用户在 Web 编辑器点击 `运行导表`
- **THEN** 本地写入服务在项目根目录执行 `scripts/gen-config.ps1`
- **AND** Web 编辑器自动切换到底部 `导表输出` 面板并显示执行中状态
- **AND** Web 编辑器显示导表状态、退出码、耗时、stdout 和 stderr
- **AND** 导表脚本成功生成 Server 和 Client 配置产物
- **AND** 地图语义表没有产生 Luban 结构或数据错误

#### Scenario: 导表脚本执行失败时回显错误

- **GIVEN** 编辑器已经将地图语义配表写入 `configs/Datas/`
- **AND** `scripts/gen-config.ps1` 因 Luban 配置错误以非零退出码结束
- **WHEN** 用户查看 `导表输出`
- **THEN** Web 编辑器显示非零退出码
- **AND** Web 编辑器显示 stderr 或 stdout 中的失败信息
- **AND** 系统不隐藏导表失败状态

### Requirement: 本地写入服务限制文件权限边界

本地写入服务 SHALL 只允许读写当前项目允许范围内的地图工程文件和 `configs/Datas/` 下的 Luban Excel 文件。服务 SHALL 拒绝任意越界路径、绝对路径注入和非预期扩展名写入。

#### Scenario: 拒绝越界写入

- **GIVEN** 前端请求写入 `..\\outside.xlsx`
- **WHEN** 本地写入服务解析该路径
- **THEN** 服务拒绝该请求
- **AND** `configs/Datas/` 外部文件不被创建或修改

#### Scenario: 写入前创建备份

- **GIVEN** `configs/Datas/#MapNode.xlsx` 已存在
- **WHEN** 编辑器执行写入 Luban 配表
- **THEN** 本地写入服务在覆盖前创建备份或使用原子写入策略
- **AND** 写入失败时原文件可恢复

### Requirement: 编辑器与 Nakama 运行时保持隔离

地图语义配置编辑器 SHALL 作为本地离线工具运行，不新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移。编辑器 SHALL 不参与在线对局状态同步，也不改变 MatchLoop tick 行为。

#### Scenario: 启动编辑器不需要 Nakama 服务

- **GIVEN** 本地 Nakama 服务未启动
- **WHEN** 开发者运行 `npm run map-editor`
- **THEN** 地图语义配置编辑器可以启动并编辑本地工程文件
- **AND** 系统不会请求 Nakama RPC 或 Storage API

#### Scenario: 写入配置不影响在线对局同步

- **GIVEN** 用户在编辑器中写入 Luban 配表文件
- **WHEN** 写入完成
- **THEN** Nakama MatchLoop tick 逻辑没有被修改
- **AND** 在线对局的服务器状态和客户端表现不因编辑器写入动作发生变化

### Requirement: 编辑器提供本地运行与验证命令

项目 SHALL 提供本地启动编辑器和写入服务的命令，并在实现中提供前端测试、构建验证、写入样例验证和 Luban 导表验证方式。

#### Scenario: 开发者启动编辑器

- **GIVEN** 开发者已安装项目所需 Node 依赖
- **WHEN** 开发者在项目根目录运行 `npm run map-editor`
- **THEN** 系统启动地图语义配置编辑器开发服务器和本地写入服务
- **AND** 终端输出可访问的本地 URL

#### Scenario: 导出结果进入 Luban 流程

- **GIVEN** 编辑器已经写入地图语义配表文件
- **WHEN** 开发者运行 `scripts/gen-config.ps1` 或在编辑器中点击 `运行导表`
- **THEN** 导表脚本成功生成 Server 和 Client 配置产物
- **AND** 写入的地图语义表没有产生 Luban 结构或数据错误

### Requirement: 地图配置页嵌入配置工作台

地图语义配置编辑器 SHALL 作为本地配置工作台中的 `地图配置` 页面继续可用。工作台新增非地图 Luban 配置页面后，地图配置页 SHALL 保持现有雷达画布、图层管理、地图对象属性编辑、校验、写入 Luban、从项目中读取和运行导表能力。

#### Scenario: 从工作台进入地图配置页

- **GIVEN** 用户已启动本地配置工作台
- **WHEN** 用户点击 `地图配置`
- **THEN** 系统显示现有地图语义编辑界面
- **AND** 页面包含 Dust2 雷达画布、地图节点、路径、视野、路线、左侧对象列表、右侧属性面板和底部日志
- **AND** 现有地图配置数据不因切换到非地图配置页再切回而丢失

#### Scenario: 地图页继续写入地图语义表

- **GIVEN** 用户处于 `地图配置` 页面
- **AND** 地图工程已通过导出前强校验
- **WHEN** 用户执行地图配置写入
- **THEN** 系统继续写入 `#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx` 和 `#CombatConst.xlsx`
- **AND** 系统继续保存 `configs/Datas/de_dust2.json` 工程快照
- **AND** 非地图配置页的脏表不会被隐式写入

#### Scenario: 地图工具快捷键仅在地图页生效

- **GIVEN** 用户处于 `地图配置` 页面
- **WHEN** 用户按下 `S`、`N`、`E`、`V`、`L` 或 `R`
- **THEN** 系统继续切换到对应地图编辑工具
- **WHEN** 用户切换到非地图配置页
- **THEN** 这些按键不再触发地图工具切换

