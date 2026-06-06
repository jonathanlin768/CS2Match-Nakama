# CLAUDE.md

## 工作规范

- **任何指令都先查 OpenSpec**：收到用户需求后，先用 Grep/Glob 查阅 `openspec/specs/` 和 `openspec/config.yaml`，定位相关模块和已有规约，再行动
- **优先使用 Grep 工具**进行内容搜索，不要为了找某个字符串而 Read 整个文件
- 同样优先使用 Glob 进行文件名匹配，不要用 `ls` 或 `find` 命令
- 仅在 Grep/Glob 定位到目标行/文件后，才用 Read 查看上下文

## 技术栈

- Go 1.24.5（Nakama 插件编译为 .so）
- Nakama 3.30.0
- React 18 + TypeScript + Vite 5
- Docker + Docker Compose
- Luban 4.5.0（策划配表工具）
