## Why

当前项目已经有 `tools/map-semantic-editor` 可以可视化维护 Dust2 地图语义表，并能直接写入 `configs/Datas/#*.xlsx` 与运行 Luban 导表。但 `#Player.xlsx`、`#Team.xlsx`、`#TutorialBattle.xlsx` 等非地图配置仍需要直接编辑 Excel，引用字段、列表字段和数值字段容易填错，也缺少面向策划的预览、筛选和校验体验。

现在需要把现有本地 Web 工具扩展为一个 Luban 可视化配置工作台：地图配置保留专业雷达画布，非地图配置通过表格、表单、引用选择器和专属视图维护，并继续直接写回当前项目的 Luban Excel 文件。

## What Changes

- 将 `tools/map-semantic-editor` 从单一地图语义编辑器升级为本地配置工作台，增加顶层页面导航或路由：地图配置、通用表格配置、选手配置、战队配置、新手战斗配置、导表输出。
- 新增通用 Luban 表读取能力，解析 `configs/Datas/#*.xlsx` 的 `##var`、`##type`、`##group` 和 `##` 表头，识别 `string`、`int`、`float`、`bool`、`list`、`ref` 等字段语义。
- 新增通用 Luban 表写回能力，在保留表头和尽量保留 workbook 样式的前提下，只替换数据区，覆盖前继续写入 `configs/Datas/.bak/` 备份；检测文件被外部修改时阻止覆盖，文件缺失或损坏时不自动重建。
- 为 `#Player.xlsx` 提供专属可视化编辑页：选手列表、战队引用下拉、能力值编辑、位置标签、稀有度、头像文件选择与预览等；外部头像复制到 `client/public/portraits/`。
- 为 `#Team.xlsx` 提供专属可视化编辑页：战队列表、名称、简称、中文昵称、Logo 文件选择与预览；外部 Logo 复制到 `client/public/teams/`。
- 为 `#TutorialBattle.xlsx` 提供专属可视化编辑页：基础开关、预算、阵容人数、地图 ID、1-5 费选手池、对手战队和对手阵容引用选择。
- 对未定制的非地图 `#*.xlsx` 表提供通用表格编辑 fallback，至少支持查看、增删行、编辑单元格、保存、写入和导表；地图语义表（包括 `#CombatConst.xlsx`）保持由地图配置页单一维护，不提供通用表格第二入口。
- 为被其他表引用的 Team/Player ID 提供引用位置提示、删除保护和显式的“修改 ID 并同步引用”操作，不执行静默级联修改。
- 保存 `#TutorialBattle.xlsx` 前严格执行现有运行时业务约束，并限制最多一条配置处于启用状态。
- 复用现有本地写入服务、路径越界保护、备份机制和 `scripts/gen-config.ps1` 运行入口。
- 不新增 Nakama RPC、Match Handler、Storage 操作、数据库迁移或在线对局状态同步逻辑。

## Capabilities

### New Capabilities

- `luban-config-visual-editor`: 覆盖本地 Web 工具中的通用 Luban Excel 表读取、可视化编辑、引用字段选择、专属配置页、写回 `configs/Datas/#*.xlsx` 和导表验证能力。

### Modified Capabilities

- `map-semantic-config-editor`: 将现有地图语义编辑器作为配置工作台中的 `地图配置` 子页保留，并要求其现有画布、写入和导表行为在新增非地图配置页后继续可用。
- `luban-config-tool`: 现有 Luban 导表链路增加本地 Web 可视化编辑入口，但仍以 `configs/Datas/#*.xlsx` 和 `scripts/gen-config.ps1` 作为权威输入与导出流程。

## Impact

- React 前端：主要影响 `tools/map-semantic-editor/src/`，需要增加顶层导航、通用表格编辑视图、Player/Team/TutorialBattle 专属页面、共享状态和测试。
- 本地 Node 服务：影响 `tools/map-semantic-editor/server/`，需要增加通用 `#*.xlsx` 元数据读取、数据区写回、引用表查询、表级校验和备份。
- Luban 配置表：需要读取并写回现有 `configs/Datas/#Player.xlsx`、`#Team.xlsx`、`#TutorialBattle.xlsx`，不新增必需表；后续可扩展到其他非地图 `#*.xlsx`，地图语义表继续由地图配置页维护。
- 客户端静态资源：选择外部头像或 Logo 时，本地服务分别写入 `client/public/portraits/` 和 `client/public/teams/`，Excel 只保存对应的项目相对路径。
- Nakama 后端：不新增 RPC、Match Handler、Storage 操作，不修改 MatchLoop tick；只有运行导表后生成的配置代码和 JSON 可能随 Excel 内容更新。
- 数据库与部署：不需要数据库迁移，不改变 Docker Compose / Nakama 部署；本工具仍是本地离线开发/策划工具。
- 游戏状态同步：无直接影响；配置变更只通过 Luban 导表进入后续运行时读取流程。
