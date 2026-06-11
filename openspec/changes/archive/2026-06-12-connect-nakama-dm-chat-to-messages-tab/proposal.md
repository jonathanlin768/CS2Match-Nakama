## Why

MessagesTab 目前完全使用硬编码 mock 数据，无法让玩家之间进行真正的私聊通信。好友系统已通过 Nakama Friends API 实现，现在需要接通 Nakama 内置的 Channel/Chat 功能，让玩家能与已添加的好友进行一对一实时私聊。Nakama 3.30+ 内置支持 DirectMessage 频道、消息持久化和 WebSocket 实时推送，无需编写 Go 插件即可满足需求。

## What Changes

- 新增 `api/chat.ts`：封装 Nakama `writeChatMessage`、`listChannelMessages` 等纯函数 API
- 新增 `hooks/useSocket.ts`：管理共享的 Nakama WebSocket 连接，统一处理 `onchannelmessage`、`onchannelpresence` 事件推送
- 新增 `hooks/useFriendDM.ts`：基于好友列表驱动 DM 会话列表，管理频道加入、消息发送、历史加载和实时接收
- 改造 `MessagesTab.tsx`：从纯 mock UI 替换为接入真实 Nakama DM 数据的交互组件，仅显示互为好友的会话
- 前端强制"仅好友可聊天"规则：会话列表由好友列表（state=FRIEND）驱动，不暴露非好友的聊天入口

## Capabilities

### New Capabilities

- `client-chat`: Nakama DM 聊天功能 — 包含 chat API 封装（`api/chat.ts`）、共享 WebSocket 管理（`hooks/useSocket.ts`）、好友私聊 Hook（`hooks/useFriendDM.ts`）、MessagesTab 接入

### Modified Capabilities

- `client-pages`: MessagesTab 模块 — 从 mock 占位变为接入 Nakama 真实 DM 的聊天页面（新增 MessagesTab 相关 Scenarios）

## Impact

- **模块影响**: React 前端（仅前端变更，不涉及 Nakama Go 插件、数据库、部署）
- **新增依赖**: 无（`@heroiclabs/nakama-js` 已安装，所有用到的 API 均来自现有 SDK）
- **需要新 RPC**: 否（使用 Nakama 内置 Channel API）
- **需要 Luban 配置变更**: 否
- **需要 Match Handler**: 否
- **Breaking Changes**: 无（MessagesTab 原本就是 mock 占位，无外部依赖）
- **影响现有代码**:
  - `client/src/pages/profile/MessagesTab.tsx` — 完全重写
  - `client/src/hooks/useFriends.ts` — 可能需要提取共享 socket（当前 hook 内部创建独立 socket）
