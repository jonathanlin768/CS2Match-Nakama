# simu-battle-full-replay Specification

## Purpose
TBD - created by archiving change complete-simu-match-mr12. Update Purpose after archive.
## Requirements
### Requirement: BattlePage 展示整场比赛时间线

`client/src/pages/BattlePage.tsx` SHALL 逐局播放服务端战报返回的回合，包括常规赛、换边、加时 block 和比赛结束。默认情况下，它 SHALL 约每秒自动显示一条事件。正常播放期间，页面 SHALL 只揭示已经推进到的回合，不得提前显示完整局数或未来回合导航；自然播放结束或用户点击“跳过比赛”后，页面 SHALL 展示完整局数和全部回合导航。

#### Scenario: 正常播放逐局揭示回合
- **GIVEN** 服务端状态包含一个有 30 个回合的 `MatchReport`
- **WHEN** `/battle` 正在播放第 4 回合且用户没有点击“跳过比赛”
- **THEN** 计分板显示当前为第 4 回合，但不显示完整局数 30
- **AND** 回合导航只包含已经推进到的第 1 至第 4 回合
- **AND** 第 5 至第 30 回合不会提前出现在导航中

#### Scenario: 自然播放结束后显示完整局数
- **GIVEN** 服务端状态包含一个有 30 个回合的 `MatchReport`
- **WHEN** 最后一回合的最后一条事件播放完成
- **THEN** 计分板显示完整局数 30
- **AND** 回合导航显示全部 30 个回合
- **AND** 最终比分、胜者和最终统计可见

### Requirement: BattlePage 可以跳到最终结果

BattlePage SHALL 提供“跳过比赛”操作，使客户端表现立即推进到最终战报状态，且不改变服务端状态。

#### Scenario: 跳过比赛显示最终结果
- **GIVEN** 服务端状态包含完整 `MatchReport`
- **WHEN** 用户点击“跳过比赛”
- **THEN** 客户端表现停止自动事件播放
- **AND** 展示上所有回合都被视为已完成
- **AND** 最终比分、胜者和最终统计可见
- **AND** 完整回合时间线仍然可用，用户可以点击任意回合回看

#### Scenario: 跳过后回看指定回合
- **GIVEN** 用户已经点击“跳过比赛”
- **WHEN** 用户在时间线上选择第 8 回合
- **THEN** 页面显示第 8 回合的事件、地图标记和回合快照
- **AND** 最终比分入口仍然可访问

### Requirement: 移动端回放控制保持当前上下文

BattlePage SHALL 在移动端提供暂停或继续、跳过比赛和选择已揭示回合的操作。移动端布局切换、视口尺寸变化或辅助视图切换 SHALL 保留当前回合、已显示事件数量、播放状态和已揭示回合边界，不得重新开始回放或提前揭示未来回合。

#### Scenario: 调整手机方向不重置回放
- **GIVEN** 用户正在第 8 回合播放到第 4 个可见事件
- **WHEN** 设备从竖屏切换为横屏后又返回竖屏
- **THEN** BattlePage 仍显示第 8 回合和相同的已播放事件范围
- **AND** 播放状态不会因布局切换被意外重置

#### Scenario: 手机选择已揭示回合
- **GIVEN** 正常播放已经揭示第 1 至第 6 回合
- **WHEN** 用户在移动端回合导航中选择第 3 回合
- **THEN** 客户端表现暂停自动播放并显示第 3 回合事件和地图标记
- **AND** 第 7 回合及未来回合仍不出现在可选范围

### Requirement: 移动端辅助视图不丢失阵容与统计

BattlePage SHALL 在移动端提供实时战况、阵容和数据统计的明确切换入口。阵容视图 SHALL 显示双方五名选手及当前回放位置的存活状态；统计视图 SHALL 使用服务器统计或由已播放服务器事件派生的当前统计，并 SHALL NOT 因桌面阵容栏被隐藏而丢失功能。

#### Scenario: 手机查看双方阵容
- **GIVEN** 当前移动端默认显示实时战况
- **WHEN** 用户选择“阵容”
- **THEN** 页面显示双方各五名选手及当前回放位置的存活状态
- **AND** 用户可以切换队伍或纵向浏览而不产生页面级横向滚动

#### Scenario: 从统计返回比赛舞台
- **GIVEN** 用户在移动端打开数据统计
- **WHEN** 用户切换回“实时战况”
- **THEN** 比赛舞台恢复显示同一当前回合和已播放事件位置
- **AND** 服务器战报和回放控制器状态保持不变

### Requirement: BattlePage 区分服务器状态与客户端表现

客户端 SHALL 将 `MatchReport`、`RoundResult`、事件快照、比分、胜者、阵营分配和选手统计视为服务器状态。客户端表现层 MAY 根据已播放事件派生可见回放状态，但 SHALL NOT 重新计算比赛规则或胜者。

#### Scenario: 客户端显示服务端胜者
- **GIVEN** 服务器状态包含 `winner_team_id = "team_b"`
- **WHEN** 回放到比赛结束
- **THEN** 客户端表现将 Team B 显示为胜者
- **AND** 不从本地 T/CT 阵营总分推断胜者

### Requirement: BattlePage 显示换边和加时标签

BattlePage SHALL 使用战报元数据清晰展示半场换边和加时回合。

#### Scenario: 出现半场标记
- **GIVEN** 服务器状态包含第 12 回合后的换边
- **WHEN** 回放时间线到达第 13 回合
- **THEN** 客户端表现标记换边
- **AND** Team A 和 Team B 的阵营徽标反映服务端提供的第 13 回合 T/CT 分配

#### Scenario: 出现加时标记
- **GIVEN** 服务器状态包含 12-12 后的加时回合
- **WHEN** 回放时间线进入第一个加时回合
- **THEN** 客户端表现将该回合标记为加时
- **AND** 常规赛和加时比分保持可读

### Requirement: BattlePage 显示当前回合事件和地图标记

BattlePage SHALL 使用服务端事件位置，为选中或当前正在回放的回合显示事件流条目和地图标记。

#### Scenario: 地图标记跟随选中回合
- **GIVEN** 第 5 回合和第 18 回合都包含带位置的 KILL 事件
- **WHEN** 客户端表现选择第 18 回合
- **THEN** 地图显示第 18 回合可见事件的标记
- **AND** 第 5 回合的标记不会作为当前回合标记显示

### Requirement: BattlePage 展示炸弹状态

BattlePage SHALL 在事件流和地图中展示下包、拆包和爆炸事件；如果战报提供炸弹状态快照，BattlePage SHOULD 提供炸弹状态小组件，用于展示未掉落、掉落、已下包、拆包中、爆炸等状态。

#### Scenario: 炸弹事件显示在地图和事件流
- **GIVEN** 当前回合包含带位置的 `BOMB_PLANT` 和 `BOMB_EXPLODE` 事件
- **WHEN** 这些事件在回放中变为可见
- **THEN** 事件流显示对应炸弹事件文本
- **AND** 地图在服务端提供的位置显示炸弹事件标记

#### Scenario: 炸弹状态小组件跟随快照更新
- **GIVEN** 战报事件包含炸弹状态快照
- **WHEN** 回放推进到 `BOMB_PLANT`
- **THEN** 炸弹状态小组件显示“已下包”或等价状态
- **AND** 回放推进到 `BOMB_DEFUSE` 或 `BOMB_EXPLODE` 后显示对应最终状态

### Requirement: BattlePage 显示聚合统计

BattlePage SHALL 基于 `MatchResult.FinalStats` 提供易读的聚合统计视图。

#### Scenario: 最终统计表使用服务器状态
- **GIVEN** 服务器状态包含 10 名选手的 `FinalStats.PlayerStats`
- **WHEN** 用户打开统计视图或回放到比赛结束
- **THEN** 客户端表现显示来自服务器状态的击杀、死亡、ADR、首杀和多杀次数

