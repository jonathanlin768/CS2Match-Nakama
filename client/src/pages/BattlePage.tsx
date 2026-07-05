import SubPageHeader from "../components/SubPageHeader"
import Scoreboard from "../components/battle/Scoreboard"
import TeamRoster from "../components/battle/TeamRoster"
import MapView from "../components/battle/MapView"
import KillFeed from "../components/battle/KillFeed"
import { battleState } from "../components/battle/data/battle"

export default function BattlePage() {
  const { title, round, maxRounds, teamA, teamB, killFeed } = battleState

  return (
    // Full viewport wrapper that centers the fixed 1920x900 game frame
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      <main className="flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
        {/* 标题：比赛进行中，隐藏返回（比赛中不允许退出） */}
        <SubPageHeader title={title} hideBack />

        {/* 两队名称 + 比分 */}
        <Scoreboard teamA={teamA} teamB={teamB} round={round} maxRounds={maxRounds} />

        {/* 三栏布局：左阵容 / 中间地图+实时播报 / 右阵容 */}
        <div className="flex min-h-0 flex-1 gap-4 px-[40px] pb-[24px] pt-2">
          <TeamRoster team={teamA} align="left" />

          <div className="flex min-h-0 flex-1 flex-col gap-3">
            <MapView />
            <KillFeed events={killFeed} />
          </div>

          <TeamRoster team={teamB} align="right" />
        </div>
      </main>
    </div>
  )
}
