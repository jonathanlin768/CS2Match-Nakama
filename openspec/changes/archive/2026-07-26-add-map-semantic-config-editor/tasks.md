## 1. 基础结构与依赖

- [x] 1.1 在 `tools/map-semantic-editor/` 创建独立 Vite + React 19 + TypeScript 工具应用结构
- [x] 1.2 增加本地 Node 写入服务，用于读取和写入当前项目 `configs/Datas/` 下的 Luban Excel 文件
- [x] 1.3 在根目录或工作区脚本中增加 `npm run map-editor`，同时启动编辑器 UI 和本地写入服务
- [x] 1.4 为编辑器增加并锁定 npm 依赖：Konva/React-Konva、Zustand、Zod、ExcelJS 以及测试所需依赖
- [x] 1.5 配置编辑器和本地服务的 TypeScript、Vite、ESLint、测试和构建脚本
- [x] 1.6 接入 Dust2 雷达底图路径 `client/public/csmaps/de_dust2_radar_trans.webp`，并提供可替换底图配置

## 2. Luban 表结构与文件模板

- [x] 2.1 读取当前 `configs/Datas/__tables__.xlsx`、`__beans__.xlsx`、`__enums__.xlsx`、`#Player.xlsx` 和 `#item.xlsx`，整理项目内 Luban Excel 写入模板
- [x] 2.2 确认地图语义表采用 `#*.xlsx` auto-import，不在 `__tables__.xlsx` 中重复注册 `tb_route_template`、`tb_scenario`、`tb_map_tag`、`tb_encounter_modifier`、`tb_map_node`、`tb_map_edge`、`tb_visibility`、`tb_route`、`tb_combat_const`
- [x] 2.3 确认第一版不新增 `__beans__.xlsx` 复合类型，地图语义表先使用 primitive、string 和 list 字段
- [x] 2.4 确认第一版不新增 `__enums__.xlsx` 枚举定义，站点、阵营、路线、阶段、距离、节点类型、节点范围用途、节点范围形状、楼层、高低关系等字段先按 `string` 导出
- [x] 2.5 创建或更新 `configs/Datas/#RouteTemplate.xlsx`、`#Scenario.xlsx`、`#MapTag.xlsx`、`#EncounterModifier.xlsx`、`#MapNode.xlsx`、`#MapEdge.xlsx`、`#Visibility.xlsx`、`#Route.xlsx`、`#CombatConst.xlsx`
- [x] 2.6 运行 `scripts/gen-config.ps1`，确认新增表结构能被 Luban 正确解析

## 3. 数据模型与工程文件

- [x] 3.1 定义地图工程模型，覆盖地图元数据、底图、坐标归一化、图层状态、画布视图、选中对象和草稿对象
- [x] 3.2 定义语义对象模型，覆盖地图节点、路径、视野关系、路线、风险热点和高低层级关系
- [x] 3.3 定义合并后的 `MapNode` 模型，覆盖 `id`、`map_id`、`name`、`zone`、`site`、`node_type`、`default_side`、`x`、`y`、`floor`、`area_usages`、`shape`、`radius`、`points`
- [x] 3.4 定义导出模型，对齐 `doc/simuMatchDesign.md` 2.2 的地图语义配表字段
- [x] 3.5 使用 Zod 实现工程文件导入、保存和版本兼容校验
- [x] 3.6 实现新建、打开、保存地图工程文件的本地服务文件操作

## 4. 本地写入服务

- [x] 4.1 实现本地服务 API：读取工程文件、保存工程文件、读取 Luban 表摘要、写入 Luban 表、运行导表验证
- [x] 4.2 对所有文件路径执行 resolve 和越界检查，只允许访问项目根目录内的明确允许路径
- [x] 4.3 实现 Excel 读取和写入工具，保留现有 `##var`、`##type`、`##group`、`##` 表头风格
- [x] 4.4 实现写入前备份或原子写入，确保失败时可恢复原 Excel 文件
- [x] 4.5 实现写入日志，记录文件路径、表名、记录数量、备份路径、警告和错误
- [x] 4.6 实现导表执行接口，在项目根目录运行 `scripts/gen-config.ps1`，返回状态、退出码、耗时、stdout 和 stderr
- [x] 4.7 为导表执行接口增加并发保护、超时处理和错误回显
- [x] 4.8 为越界路径、非法扩展名、写入失败、恢复流程和导表执行失败编写服务端单元测试

## 5. 编辑器状态与画布交互

- [x] 5.1 使用 Zustand 实现工程数据、工具模式、选中对象、图层状态、校验结果和预览状态管理
- [x] 5.2 实现撤销/重做历史，覆盖对象创建、删除、移动和属性修改
- [x] 5.3 确认中央 Dust2 雷达画布现有固定视图、网格、吸附状态和坐标显示满足第一版；第一版不实现额外缩放、平移、居中、重置视图或辅助线
- [x] 5.4 实现地图节点创建、选择、拖动、复制、删除和基础属性编辑
- [x] 5.5 实现地图节点几何范围编辑，支持 `None`、圆形、多边形、范围用途、楼层和采样预览
- [x] 5.6 实现路径连接、方向性、耗时、风险、暴露/噪音、风险热点和拦截点编辑；第一版暂不配置移动方式
- [x] 5.7 实现视野关系编辑，覆盖单向/双向、距离、角度优势、遮挡、暴露修正和高低层级语义
- [x] 5.8 实现路线编辑，覆盖节点序列、阵营、目标区域、人数范围、战术用途和断裂提示

## 6. 工具型 UI

- [x] 6.1 实现顶部工具栏，覆盖选择、地图节点、路径、视野、路线、风险热点、预览、校验、写入 Luban 和运行导表命令
- [x] 6.2 实现左侧面板，覆盖地图工程信息、图层列表和对象列表
- [x] 6.3 实现中间画布布局，确保雷达底图与语义对象在桌面和较窄视口下不重叠失真
- [x] 6.4 实现右侧属性面板，选中不同对象时显示对应字段编辑器
- [x] 6.5 实现底部面板，覆盖校验结果、预览输出、写入日志和导表输出
- [x] 6.6 为图层和对象类型配置稳定颜色、图标或符号，确保高低层级关系在视觉上可区分

## 7. 校验与预览

- [x] 7.1 实现 ID 唯一性和引用存在性校验
- [x] 7.2 实现路线节点、路径端点、视野端点、风险热点和拦截点等 `tb_map_node.id` 引用校验
- [x] 7.3 实现路线连通、路径方向冲突和孤立关键节点校验
- [x] 7.4 实现地图节点几何校验，覆盖圆形半径、锚点坐标、多边形顶点格式、多边形合法性、范围用途和采样边界
- [x] 7.5 实现视野关系必要语义和高低层级完整性校验
- [x] 7.6 实现导出模型约束校验，确保输出满足 `doc/simuMatchDesign.md` 2.2
- [x] 7.7 实现校验错误列表到画布对象的定位、高亮和属性面板定位
- [x] 7.8 实现击杀点随机采样预览、路线连通预览、视野线预览、高低关系预览和写入摘要预览

## 8. Luban 写入与导表

- [x] 8.1 实现 `toExportTables(project)` 转换层，将工程模型转换为地图语义配表数据
- [x] 8.2 实现 `tb_route_template`、`tb_scenario`、`tb_map_tag`、`tb_encounter_modifier`、`tb_map_node`、`tb_map_edge`、`tb_visibility`、`tb_route` 和 `tb_combat_const` 的导出映射
- [x] 8.3 实现本地服务直接写入 `configs/Datas/#*.xlsx` 数据表文件
- [x] 8.4 实现本地服务按 `#*.xlsx` auto-import 规则写入地图语义表，且不重复修改 `__tables__.xlsx`、`__beans__.xlsx`、`__enums__.xlsx`
- [x] 8.5 在写入日志中展示每张表的文件路径、记录数、备份路径、警告和阻止写入的错误
- [x] 8.6 写入完成后支持在 Web 编辑器点击按钮运行 `scripts/gen-config.ps1` 验证导表
- [x] 8.7 在导表输出面板显示执行中、成功、失败、退出码、耗时、stdout 和 stderr

## 9. 前端测试与验证

- [x] 9.1 为工程文件 schema、导入保存、模型转换和校验器编写单元测试
- [x] 9.2 为 MapNode 圆形采样、多边形采样、无范围回退采样、路线连通和视野/高低关系预览编写单元测试
- [x] 9.3 为 Excel 写入映射编写样例测试，验证工作簿、表头、行数和关键字段
- [x] 9.4 使用组件或端到端测试覆盖新建工程、编辑节点、校验错误定位和写入 Luban 流程
- [x] 9.5 运行编辑器 lint、TypeScript 检查、测试和生产构建
- [x] 9.6 使用 Playwright 或等价方式检查桌面和窄视口画布、面板和按钮文字无重叠

## 10. Luban、后端与权限回归验证

- [x] 10.1 使用编辑器样例工程直接写入 `configs/Datas/`，并确认相关 Excel 文件发生预期变更
- [x] 10.2 运行 `scripts/gen-config.ps1`，验证导表后 Server 和 Client 配置产物成功生成
- [x] 10.3 运行 Go 后端测试，确认新增配表定义未破坏现有配置加载和模拟器测试
- [x] 10.4 编译 Nakama Go Plugin，确认 `.so` 构建仍成功
- [x] 10.5 检查本变更未新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移
- [x] 10.6 检查 MatchLoop tick 逻辑和在线对局状态同步代码未被修改
- [x] 10.7 验证本地写入服务无法写入项目根目录外的任意路径

## 11. 文档与交付

- [x] 11.1 编写编辑器本地启动、工程文件保存、直接写入 Luban 配表和导表验证说明
- [x] 11.2 在文档中说明第一版不走浏览器下载导入流程，而是由本地写入服务直接写入 `configs/Datas/`
- [x] 11.3 记录第一版已知限制：仅 Dust2、无多人协作、无真实 3D、无自动识图、无运行时模拟器接入
- [x] 11.4 根据实现结果更新 OpenSpec spec 中的验收说明或后续变更建议
