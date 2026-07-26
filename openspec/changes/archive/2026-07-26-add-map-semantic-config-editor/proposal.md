## 背景与动机

`doc/simuMatchDesign.md` 2.2 定义了包含战术模板、场景、地图标签、遭遇战修正、地图节点、路径、视野、路线和战斗常量的地图语义配置表。`tb_map_node` 同时承载位置锚点和可选几何范围；直接在 Excel 里维护这些空间关系成本高、容易出错，也无法直观预览雷达图上的节点范围、采样效果、路线断点和高低关系。

现在需要一个本地可视化编辑器，让设计者在 Dust2 雷达底图上配置地图语义，并将编辑结果 1:1 写入项目当前 Luban 配表目录和文件风格中，使数据可以直接进入现有 `scripts/gen-config.ps1` / `scripts/gen-config.sh` 导表链路。

## 变更内容

- 新增一个本地 Web 地图语义配置编辑器，作为项目工具链的一部分。
- 编辑器使用浏览器 UI 进行交互，但导出动作 SHALL 通过本地 Node 写入服务直接写入当前项目的 `configs/Datas/` 目录。
- 导出文件 SHALL 按当前项目 Luban 风格生成或更新，例如 `configs/Datas/#MapNode.xlsx`、`#MapEdge.xlsx` 等自动导入数据表文件；第一版不把 `#*.xlsx` 表重复注册到 `__tables__.xlsx`。
- 支持地图工程文件管理，包括地图 ID、显示名称、雷达底图路径、地图版本、`0..1` 归一化坐标规则、图层状态、选中对象和草稿编辑状态。
- 提供基于雷达图的可编辑画布，覆盖地图节点、路径、视野、路线和风险热点等语义图层。
- 支持对象创建、选中、拖动、编辑、复制、删除、撤销、重做、网格、吸附和 `0..1` 归一化坐标显示；第一版不额外实现缩放、平移或辅助线交互。
- 提供对象属性面板，用于编辑地图节点、节点几何范围、路径、视野关系、路线和击杀采样范围。
- 支持合并后的 `tb_map_node`，用于表达位置锚点、路线/视野图关系，以及 `KillSample`、`Plant`、`Control`、`Encounter`、`Sound`、`Risk` 等可选几何范围用途。
- 提供导出前校验，能定位 ID 重复、引用缺失、路线断裂、非法多边形、路径方向配置冲突、视野或高低关系语义不完整，以及不满足 `doc/simuMatchDesign.md` 2.2 的导出数据。
- 提供轻量预览，包括击杀点随机采样、路线连通关系、视野线、高低关系和导出摘要。
- 增加本地运行脚本和包命令，开发者可通过 `npm run map-editor` 启动编辑器和本地写入服务。
- 本变更不修改 Nakama 运行时、Match Handler、Storage 或实时状态同步。

## 能力范围

### 新增能力

- `map-semantic-config-editor`：覆盖本地地图工程管理、雷达画布编辑、地图节点语义图层、配置校验、效果预览、本地文件写入服务，以及按当前 Luban 风格直接写入 `configs/Datas/` 的地图语义配表导出。

### 修改能力

- 无。现有 Luban 导表脚本、Nakama RPC、Match Handler、Storage 和模拟器运行时规约不在本变更中修改。

## 影响范围

- React 前端：新增本地工具应用，建议位置为 `tools/map-semantic-editor/`，实现时优先复用 React 19、TypeScript 和 Vite，并按设计引入 Konva/React-Konva、Zustand、Zod、ExcelJS 等编辑器依赖。
- Node 本地工具服务：新增仅限开发期运行的本地写入服务，用于读取/写入 `configs/Datas/` 下的 Excel 文件；写入路径必须限制在项目根目录允许范围内。
- Nakama 后端：无运行时代码变更，不新增 RPC，不修改 Match Handler，不新增 Storage 操作。
- 数据库：无表结构或数据迁移。
- 部署：无生产部署影响；编辑器和写入服务都是本地开发和设计工具。
- 状态同步机制：无影响；编辑器不参与在线对局，不进入服务器权威 MatchLoop，也不改变客户端状态同步。
- Luban 配置表：需要新增与 `doc/simuMatchDesign.md` 2.2 对齐的 `#*.xlsx` 地图语义表文件，并按当前 `configs/Datas/*.xlsx` 自动导入风格写入；第一版枚举值先以 `string` 字段承载，避免过早引入 `__enums__.xlsx` 维护成本。
