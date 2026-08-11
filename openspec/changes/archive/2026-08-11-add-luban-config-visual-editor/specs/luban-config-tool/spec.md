## ADDED Requirements

### Requirement: Luban 配表支持本地可视化编辑入口

项目 SHALL 提供本地 Web 可视化编辑入口，用于编辑当前项目 `configs/Datas/#*.xlsx` 中的 Luban 配表。该入口 SHALL 以 Excel 文件作为权威输入和输出，不绕过 `scripts/gen-config.ps1` 导表链路，不要求用户从浏览器下载文件后手动导入。

#### Scenario: 通过 Web 工具编辑 Player 表并导表

- **GIVEN** `configs/Datas/#Player.xlsx` 存在
- **WHEN** 用户在本地 Web 配置工作台中修改选手数据并保存
- **THEN** 系统更新 `configs/Datas/#Player.xlsx`
- **AND** 用户运行 `scripts/gen-config.ps1` 或点击 Web 中的 `运行导表` 后，Server 和 Client 生成配置产物反映该 Excel 修改

#### Scenario: 通过 Web 工具编辑 Team 表并保持 Luban auto-import

- **GIVEN** `configs/Datas/#Team.xlsx` 使用 `#*.xlsx` auto-import 风格
- **WHEN** 用户在本地 Web 配置工作台中保存战队数据
- **THEN** 系统不向 `configs/Datas/__tables__.xlsx` 重复注册该表
- **AND** 该表仍由 Luban auto-import 规则导出

#### Scenario: 可视化编辑不新增运行时接口

- **GIVEN** 用户通过本地 Web 配置工作台保存任意 `#*.xlsx` 表
- **WHEN** 保存完成
- **THEN** 项目不新增 Nakama RPC、Match Handler、Storage 操作或数据库迁移
- **AND** 在线对局状态同步逻辑不因保存动作发生变化
