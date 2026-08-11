## Why

当前 `#Player.xlsx` 只有单一 `portrait` 路径，无法同时表达完整 `2:3` 选手卡面和比赛过程中使用的 `5:7` 头像视图。策划需要在配置工作台中上传完整卡面、可视化确定头像裁切区域，并让同一份裁切参数经 Luban 和比赛快照稳定传递到客户端，避免手工制作和维护两套容易失配的图片。

## What Changes

- 扩展 `#Player.xlsx`，增加完整卡面资源路径和归一化头像裁切矩形；保留现有 `portrait` 作为旧数据兼容与缺失配置 fallback。
- 将用户提供的首批选手卡面复制到 `client/public/player-cards/`，为可可靠匹配的现有 Player 预置卡面路径和合法的 `5:7` 默认裁切，使页面首次打开即可看到卡面。
- 在配置工作台的 Player 专属页增加完整卡面上传、固定 `5:7` 裁切框、拖动/缩放、重置和实时头像预览，并把原始卡面复制到受控的项目资源目录。
- 在 Go 配置适配和比赛结果快照中传递卡面路径与裁切数据，使客户端表现只消费服务器提供的静态选手档案，不自行推导策划参数。
- 在现有“15 元教学选人页”使用完整 `2:3` 卡面展示候选选手；在比赛回放的 `TeamRoster` / `PlayerCard` 中使用同一原图和 `5:7` 裁切参数显示头像。
- 对缺失图片、越界裁切、非有限数值、空裁切和不符合 `5:7` 的裁切配置提供校验与兼容回退。
- 不预生成第二张头像文件；客户端根据原图和裁切数据进行无损展示，避免衍生图片与 Excel 参数不同步。

## Capabilities

### New Capabilities

无。本变更扩展现有配置编辑、Player 配置契约和客户端表现能力。

### Modified Capabilities

- `luban-config-visual-editor`: Player 专属页增加完整卡面上传、固定比例裁切工具、实时预览和裁切参数校验。
- `simu-config-structures`: `TbPlayer`、`PlayerProfile` 和回合 `PlayerState` 增加卡面与头像裁切快照字段，并保留旧 `portrait` fallback。
- `simu-client-replay`: 战斗回放中的选手头像按服务端快照提供的 `5:7` 裁切数据渲染。
- `client-pages`: “15 元教学选人页”由文字候选项升级为完整 `2:3` 选手卡面，并保留预算、费用档和选择行为。

## Impact

- **Luban 配置表**：需要修改 `configs/Datas/#Player.xlsx`，增加 `cardImage`、`avatarCropX`、`avatarCropY`、`avatarCropWidth`、`avatarCropHeight` 等字段，并重新生成 Server Go、Client TypeScript 和 JSON 产物。
- **配置编辑器**：影响 `tools/map-semantic-editor` 的图片上传白名单、资源目录、Player 页面、校验、Excel 写回和组件测试；可能引入一个支持固定比例裁切且兼容 React 19 的前端裁切组件依赖。
- **React 前端**：影响 `TutorialPage`、`BattlePage`、`PlayerCard`、选手表现模型、样式和相关测试；新增 `client/public/player-cards/` 受控资源目录。
- **Nakama 后端**：影响 `server/internal/match` 的 Luban 适配和 `server/internal/framework/matchengine` 的静态选手档案/公开投影字段，但不改变模拟计算。
- **状态同步**：只在已有比赛结果和回放快照中增加静态视觉字段，不改变服务器权威状态、不改变 MatchLoop、tick 频率、事件顺序或客户端插值逻辑。
- **RPC / Match Handler / Storage / 数据库**：不新增 RPC，不新增或修改 Match Handler 生命周期，不新增 Storage 操作，不需要数据库迁移。现有响应 JSON 仅增加可选视觉字段。
- **部署**：不增加独立服务或容器；需重新运行 Luban 导表、React 构建和 Nakama Go Plugin Docker 编译。
