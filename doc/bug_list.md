# bug
- [x] 前端实时播报内容不正确
```
round started from authoritative opening plan
damage applied
damage applied
player killed
bomb dropped
node control resolved from encounter
damage applied
damage applied
damage applied
player killed
player killed
node control resolved from encounter
CT wins the round by timeout
```
这里预期显示的是
 TeamA 执行了xxx战术
 选手A 在 哪里 击杀了 选手B
 选手A 在 哪里 攻击了 选手B，造成38点伤害
 选手A 丢掉/安放了 炸弹
 炸弹被 选手B 拆除
 炸弹爆炸
 TeamA 作为 T/CT 赢得本局胜利

 等等

- [x] 前端数据统计不正确。
击杀/死亡/助攻 K D A 应该是多回合累计的。不是某一局的数据

- [x] 跳过比赛后，点击某个回合，直接显示的是回合的结果，并且点击播放时，播放下一回合。我希望改成，点击某一回合，显示回合开始，然后点击播放的时候，播放这个回合。

- [x] 分析一下这个战报为什么会一直刷decision resolved into real actions

```
round started from authoritative opening plan
strategy score adjusted by completed-round memory
strategy score adjusted by completed-round memory
damage applied
player killed
node control resolved from encounter
damage applied
player killed
damage applied
damage applied
player killed
node control resolved from encounter
decision resolved into real actions
decision resolved into real actions
decision resolved into real actions
decision resolved into real actions
decision resolved into real actions
decision resolved into real actions
damage applied
player killed
decision resolved into real actions
damage applied
decision resolved into real actions
damage applied
player killed
bomb dropped
node control resolved from encounter
decision resolved into real actions
CT wins the round by timeout
```
并且之前的控制台日志也没有了。

- [x] 实时播报是否能够自动拖进度条到底
- [x] 实时播报显示`京介 在 A Site 攻击了 豆豆，造成 59 点伤害。 ` 但是右侧 豆豆还是100滴血，应该同步变化
