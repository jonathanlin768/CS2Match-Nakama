## 实现备注

- 当前 `de_dust2` 语义配置仍是最小可运行数据：只有 `D2_A_LONG` 路线和 `TPL_A_LONG_DEFAULT` 模板。`matchengine` 已按 `MapConfig` 快照消费路线、节点、场景、地图标签、路径、视野和战斗常量；后续可直接通过 `tools/map-semantic-editor` / Luban 补齐更多点位、进攻路线、风险点和拦截候选。
- 本次固定武器数值仍位于 `server/internal/match` 的业务层默认快照中：T 方 `AK47`，CT 方 `M4A1S`。`PlayerProfile` 没有固定武器字段。后续如果要策划可编辑武器数值，需要新增 `tb_weapon` 或等价提案。
- 默认调试阵容不再依赖人工创建的 `debug_*` 选手。`configs/Datas/#Player.xlsx` 是唯一选手数据源，业务层按 `team` 字段构建 Team A `Falcons` 与 Team B `Vitality`，每队必须恰好 5 人且队内沿用 Luban 表顺序；`tbplayer.json` 只能由 Luban 生成。
- `configs/Datas/#CombatConst.xlsx` 是战斗常量的唯一数据源，必需的 39 项常量及合法上下限均在 Excel 中维护，并由 Luban 同步导出服务端和客户端 JSON。
- `TbPlayer.Portrait` 以站点根目录相对路径传递。形如 `portraits/player_niko.jpg` 的资源应放在 `client/public/portraits/player_niko.jpg`；客户端加载失败时回退到 `/images/star-player.png`。
- 第一版回合内逻辑表达了模板选择、场景参与评分、开局/中期击杀、炸弹阶段、事件原因、地图坐标采样和全场记忆边界，但仍将连续移动、完整寻路、细粒度 action 生命周期和完整情报传播保持为抽象逻辑。
- 当前前端项目没有 Vitest/RTL/Playwright 等测试设施。本次已通过 `npx tsc --noEmit` 和 `npm run build` 验证类型与生产构建；若后续要完成组件级回放测试，应另行引入前端测试基础设施或在既有测试框架出现后补测。
- 本地 `server/vendor/` 被 `.gitignore` 忽略且可能存在旧副本。项目构建脚本已使用 `-mod=mod`，本次验证也使用 `go test -mod=mod ./...`；若本机普通 `go test ./...` 命中旧 vendor，可删除本地 ignored vendor 或使用 `-mod=mod`。
