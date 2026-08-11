# CS2 配置可视化编辑器

本工具是本地 Web 配置工作台，用于可视化编辑 CS2 模拟器的地图语义配置、选手、战队、新手战斗和其他 Luban auto-import 表，并直接写入当前项目的 Excel 配表。

## 启动

在项目根目录运行：

```powershell
npm run map-editor
```

也可以使用新名称启动同一个工具：

```powershell
npm run config-editor
```

启动后访问：

```text
http://127.0.0.1:5177
```

该命令会同时启动：

- Vite Web UI：`127.0.0.1:5177`
- 本地写入服务：`127.0.0.1:5178`

`map-editor` 旧命令会继续保留，页面名称统一显示为 `CS2 Config Editor`。

## 配置工作台

顶部导航提供五个页面：

- `地图配置`：维护地图节点、路径、视野、路线、标签、场景和战斗常量。
- `选手配置`：维护 `#Player.xlsx`，支持筛选、能力值、位置标签、战队引用、完整卡面上传和战斗头像裁切。
- `战队配置`：维护 `#Team.xlsx`，支持引用反查、Logo 和显式 ID 同步。
- `新手战斗`：维护 `#TutorialBattle.xlsx` 的费用档候选池、预算和对手阵容。
- `全部表格`：查看所有 `configs/Datas/#*.xlsx`，并通用编辑 `#item.xlsx` 等非地图表。

地图语义表和 `#CombatConst.xlsx` 只允许从 `地图配置` 页面修改；`全部表格` 页面只显示其归属和跳转按钮，不提供第二个编辑入口。

非地图页面的 `保存当前表` 只写入当前已修改的 Excel，`保存全部` 会以一个批次写入全部已修改表。切换页面不会丢失内存中的修改；关闭或刷新有未保存数据的页面时，浏览器会提示确认。

每次覆盖 Excel 前，服务会在以下目录保留备份：

```text
configs/Datas/.bak/config-editor/<时间戳>/
```

保存会保留原 workbook、sheet、四行 Luban 表头、列宽和样式，只重写第 5 行开始的数据区。文件已被外部修改、缺失、损坏或无法解析时，服务会阻止覆盖，不会自动创建替代工作簿。

## 专属配置规则

- Player 的 15 项能力值必须是 `0..100` 范围内的整数。
- Team 和 Player 被其他表引用时禁止删除。修改被引用的 ID 必须点击 `应用 ID`，确认引用清单后批量同步。
- TutorialBattle 每个费用档至少需要一个候选选手，整张表最多启用一套方案。
- 启用的 TutorialBattle 必须满足预算、5 人阵容、跨档不重复、预算可组成完整阵容，以及对手 5 人均属于所选战队等运行时规则。
- ERROR 会阻止保存；WARNING 会显示在底部校验面板，但允许保存。

战队 Logo 和选手头像可从外部选择。服务会校验 PNG、JPEG 或 WebP 内容并分别复制到：

```text
client/public/teams/
client/public/portraits/
```

表格中保存的是 `teams/<文件名>` 或 `portraits/<文件名>`。同名文件必须明确确认覆盖；缺失图片会显示告警，但不会阻止保存。

### 选手卡面与战斗头像

在 `选手配置` 中选中一名 Player 后，可在“选手视觉资源”区域完成以下流程：

1. 点击 `上传完整卡面`，选择 PNG、JPEG 或 WebP 图片。服务会把源图复制到 `client/public/player-cards/`，表格中保存 `player-cards/<文件名>`。
2. 在裁切画布中拖动或缩放固定 `5:7` 的选框；右侧预览会实时显示战斗中使用的小头像。
3. 点击 `重置裁切` 可恢复为源图中央最大 `5:7` 区域。
4. 点击 `保存当前表` 或 `保存全部`，把卡面路径和裁切参数写入 `configs/Datas/#Player.xlsx`。
5. 点击 `运行导表`，生成 Server Go、Client TypeScript 和两端 JSON。重新打开编辑器或点击页面的重新读取操作后，裁切框应与保存前一致。

Player 新字段约定：

| 字段 | 取值与含义 |
| --- | --- |
| `cardImage` | 相对 `client/public/` 的完整卡面路径；空值时回退到旧 `portrait` |
| `avatarCropX` | 裁切左边界相对源图宽度的比例，范围 `0..1` |
| `avatarCropY` | 裁切上边界相对源图高度的比例，范围 `0..1` |
| `avatarCropWidth` | 裁切宽度相对源图宽度的比例，范围 `(0..1]` |
| `avatarCropHeight` | 裁切高度相对源图高度的比例，范围 `(0..1]` |

裁切必须满足 `x + width <= 1`、`y + height <= 1`。若源图自然尺寸为 `W x H`，还必须满足 `(width * W) / (height * H) = 5 / 7`（允许编辑器校验所定义的微小浮点误差）。系统只保存原始卡面和归一化裁切参数，不生成额外的裁切图片；客户端在显示时依据参数动态裁切。

现有 Player 的兼容迁移顺序是：先复制卡面源图，再填写 `cardImage` 和合法裁切参数，保存 `#Player.xlsx`，最后运行导表。没有卡面或裁切无效时，客户端按 `cardImage + crop -> portrait -> 默认图` 回退；完整卡面场景按 `cardImage -> portrait -> 默认图` 回退。

## 本地工程文件

编辑器默认读取或创建：

```text
tools/map-semantic-editor/data/de_dust2.json
```

工程文件保存编辑器状态和地图语义对象，包括地图节点、路径、视野、路线、风险热点、图层和画布视图。

顶部 `保存` 只写入这个本地工程文件，不会修改 Luban Excel。覆盖保存前，服务会在 `tools/map-semantic-editor/data/.bak/` 下备份旧工程文件。顶部 `从项目中读取` 会从发布目录 `configs/Datas/<map_id>.json` 读取最近一次写入 Luban 时保存的工程快照，并加载回编辑器内存。

顶部 `读取当前Excel配置` 会直接解析 `configs/Datas/#*.xlsx`（9 张 Luban 表）的当前内容并载入编辑器，替换节点、路径、视野、路线、模板、场景、标签、修正和常量数据；工程元数据（地图名、雷达图、图层、视图）保持不变。每张表的行数与解析警告显示在底部 `读取日志` 页签。

## 常用编辑

- 创建点位：点击顶部 `点位 (tb_map_node)`，再点击雷达图空白位置。
- 划定点位范围：先选中点位，再在右侧属性面板点击 `划定圆形范围` 或 `划定多边形范围`。
- 圆形范围：点击圆形边界点，编辑器会根据点位 `x/y` 自动计算并写入 `radius`。
- 多边形范围：在雷达图上逐点点击顶点，点击 `完成范围` 或按 `Enter` 写入 `points`；按 `Esc` 或点击 `取消范围` 放弃草稿。
- 选择点位范围：点击顶部 `选择`，再点击圆形或多边形范围内部，编辑器会选中该范围所属的 `点位 (tb_map_node)`。
- 创建路线：点击顶部 `路线 (tb_route)`，按顺序点击多个点位，再按 `Enter` 或点击 `完成路线` 写入 `tb_route.nodes`；按 `Esc` 或点击 `取消路线` 放弃草稿。
- 选择路径/路线：点击顶部 `选择`，再点击画布上的线段；如果同一段上同时命中 `路径 (tb_map_edge)` 和 `路线 (tb_route)`，重复点击会在候选对象之间循环切换。隐藏或锁定的图层不会参与画布命中。
- 图层过滤：`地图节点` 图层控制所有 `tb_map_node` 锚点；`节点范围` 图层控制所有节点的圆形或多边形范围。`area_usages` 包含 `Risk` 的风险热点也是 `tb_map_node`，不再提供独立图层勾选框。
- 风险热点显示：画布上风险热点使用玫红色节点、金色外环和金色范围，与普通地图节点区分。
- 风险热点：这是赛前配置的高冲突、高暴露或容易被拦截的位置，用于影响模拟概率和事件位置采样候选；它不是运行时死亡、掉包或拦截坐标的硬编码来源。
- 删除对象：选中点位、路径、视野或路线后按 `Delete`，也可以点击右侧 `删除`。
- 撤销/重做：按 `Ctrl+Z` 撤销，按 `Ctrl+Y` 重做；输入框、下拉框和多行文本中保留浏览器自己的文本编辑行为。
- 采样预览：点击 `预览` 后会显示黄色采样点，同一个按钮会切换为 `清除预览`；再次点击即可移除这些临时预览点。

## 写入 Luban

点击 Web 顶部的 `写入 Luban` 后，本地服务会把当前工程转换并写入：

```text
configs/Datas/#RouteTemplate.xlsx
configs/Datas/#Scenario.xlsx
configs/Datas/#MapTag.xlsx
configs/Datas/#EncounterModifier.xlsx
configs/Datas/#MapNode.xlsx
configs/Datas/#MapEdge.xlsx
configs/Datas/#Visibility.xlsx
configs/Datas/#Route.xlsx
configs/Datas/#CombatConst.xlsx
```

这些表使用当前项目的 `#*.xlsx` auto-import 风格，不需要在 `__tables__.xlsx` 中重复注册。

覆盖已有文件前，服务会在 `configs/Datas/.bak/map-semantic-editor/` 下创建备份。

写入成功后，服务还会额外保存一份编辑器工程快照：

```text
configs/Datas/de_dust2.json
```

这个 JSON 不参与 Luban 导表，只用于后续从编辑器点击 `从项目中读取` 恢复完整编辑状态。

## 运行导表

点击 Web 顶部的 `运行导表` 后，本地服务会在项目根目录执行：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/gen-config.ps1
```

执行结果会显示在底部 `导表输出` 面板中，包括状态、退出码、耗时、stdout 和 stderr。

## 第一版限制

- 只内置 Dust2 样例工程。
- 不做多人协作。
- 不做真实 3D 地图编辑。
- 不自动识别地图结构。
- 只负责生产 Luban 输入，不接入模拟器运行时消费。
- 不编辑 `__beans__.xlsx`、`__enums__.xlsx` 和 `__tables__.xlsx`。
- 地图语义表及 `#CombatConst.xlsx` 不开放通用表格编辑。
- 未识别的复杂 Luban 类型使用文本输入 fallback，并保留原始 `##type`。
- 图片只复制到项目受控资源目录，不转换源图格式；选手战斗头像使用参数化动态裁切，不生成裁切副本。

需要回滚工作台入口时，可以隐藏非地图导航和对应路由；原 `地图配置` 页面、`npm run map-editor` 命令及地图写入流程可以继续独立使用。
