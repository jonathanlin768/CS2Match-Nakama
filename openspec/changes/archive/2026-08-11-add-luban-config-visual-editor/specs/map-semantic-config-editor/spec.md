## ADDED Requirements

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
