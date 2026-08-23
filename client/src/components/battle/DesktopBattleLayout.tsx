import SubPageHeader from "../SubPageHeader"
import BattleReplayControls from "./BattleReplayControls"
import BombStatusStrip from "./BombStatusStrip"
import EventFeed from "./EventFeed"
import MapView from "./MapView"
import Scoreboard from "./Scoreboard"
import TeamRoster from "./TeamRoster"
import type { BattleViewModel } from "./battle-view-model"

export default function DesktopBattleLayout({ model }: { model: BattleViewModel }) {
  return (
    <main className="battle-stage desktop-battle-layout" data-battle-layout="desktop">
      <SubPageHeader title={model.fullMatchRevealed ? "比赛结束" : "比赛进行中"} hideBack />
      <Scoreboard
        teamA={model.teamA}
        teamB={model.teamB}
        round={model.currentRound.round_number}
        maxRounds={model.fullMatchRevealed ? model.report.total_rounds : undefined}
        phaseLabel={model.phaseLabel}
        winnerName={model.winnerName}
      />
      <BattleReplayControls model={model} />
      <div className="desktop-battle-content">
        <TeamRoster team={model.teamA} align="left" />
        <div className="desktop-battle-center">
          <BombStatusStrip bomb={model.bomb} />
          <MapView events={model.visibleEvents} />
          <EventFeed events={model.visibleEvents} teamA={model.teamA} teamB={model.teamB} eventContext={model.eventContext} />
        </div>
        <TeamRoster team={model.teamB} align="right" />
      </div>
    </main>
  )
}
