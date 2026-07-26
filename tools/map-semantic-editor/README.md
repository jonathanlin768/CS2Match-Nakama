# 地图语义配置编辑器

本工具是本地 Web 编辑器，用于编辑 CS2 模拟器地图语义配置，并直接写入当前项目的 Luban Excel 配表。

## 启动

在项目根目录运行：

```powershell
npm run map-editor
```

启动后访问：

```text
http://127.0.0.1:5177
```

该命令会同时启动：

- Vite Web UI：`127.0.0.1:5177`
- 本地写入服务：`127.0.0.1:5178`

## 本地工程文件

编辑器默认读取或创建：

```text
tools/map-semantic-editor/data/de_dust2.json
```

工程文件保存编辑器状态和地图语义对象，包括地图节点、路径、视野、路线、风险热点、图层和画布视图。

顶部 `保存` 只写入这个本地工程文件，不会修改 Luban Excel。覆盖保存前，服务会在 `tools/map-semantic-editor/data/.bak/` 下备份旧工程文件。顶部 `从项目中读取` 会从发布目录 `configs/Datas/<map_id>.json` 读取最近一次写入 Luban 时保存的工程快照，并加载回编辑器内存。

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
