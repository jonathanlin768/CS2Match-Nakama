openspec创建提案，需求如下：
预期目标：点击“开始匹配“后，执行完整的一场模拟比赛，包含T与CT进行12小局的游戏，然后换边执行12小局的游戏。先得到13分的一方获胜。如果出现12-12的情况，则进入加时。具体逻辑与目前主流CS2比赛一致。
模拟比赛逻辑：读取 模拟对局设计文档 simuMatchDesign.md，严格按照文档内逻辑执行。
选手：你自己生成10个选手放到两队中即可。
配表：严格按照 模拟对局设计文档 的内容，进行配表读取。目前已有配表工具在 tools/map-semantic-editor 。后续由我人工配置各个地图点位和进攻路线等。你按照 模拟对局设计文档 进行读表和逻辑编写。
每回合枪械及金钱：目前不涉及金钱，枪械为T方固定AK47，CT方固定M4A1-S。具体数值你从网上找一份合理的即可。
**模拟比赛的代码务必全部写在 framework\matchengine 下，作为整个项目的一个基础能力，提供给其他模块进行使用。**
所以你需要为了实现这个需求而进行优雅的拆分，分清楚什么数据是属于framework的，什么数据应该放在上层业务。

另外，本次的修改你会涉及到客户端的改动。请你一并对客户端进行同步修改。针对于这次 模拟对局设计文档 中提到的新内容，相应的请你修改客户端的代码，务必确保表现优良。

服务器代码：
- 模拟对局框架：D:\Project\CS2Match-Nakama\server\internal\framework\matchengine
- 发起模拟业务：D:\Project\CS2Match-Nakama\server\internal\match

客户端代码：
- D:\Project\CS2Match-Nakama\client\src\pages\BattlePage.tsx (路由为 "\battle")


# review问题
1. server/internal/framework/matchengine/model.go:232 PlayerState 这里不用兼容旧战报，没有必要
```
IsAlive     bool          `json:"is_alive"`               // 兼容旧战报消费者的存活标记。
```

2. 这个engine.go 有1000多行，客观上讲，它的逻辑确实相对复杂。是否有更好的设计模式或者代码结构来实现这个功能？
你先介绍一下engine
请你基于这个世界上最先进的软件工程实践，将engine.go这个游戏模拟引擎，判断目前是否已经到了可以已进行重构的时候，或者考虑是否有一些内容可以提前重构？
先给出你的回答，是重构还是不重构，重构哪一部分？
