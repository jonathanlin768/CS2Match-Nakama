import { BarChart3, Radio, Users } from "lucide-react"
import { useState } from "react"
import { formatBattleEvent, formatCompactBattleEvent } from "../../pages/battle-playback"
import type { BattleTeam } from "./data/battle"
import BattleReplayControls from "./BattleReplayControls"
import BattleStatsTables from "./BattleStatsTables"
import BombStatusStrip from "./BombStatusStrip"
import MapView from "./MapView"
import Scoreboard from "./Scoreboard"
import type { BattleViewModel } from "./battle-view-model"

type MobilePanel = "live" | "roster" | "stats"

function AlivePlayersStrip({ teamA, teamB }: { teamA: BattleTeam; teamB: BattleTeam }) {
  const sideName = (team: BattleTeam) => team.side.toUpperCase()
  return (
    <div className="compact-alive-strip" aria-label="双方存活状态">
      {[teamA, teamB].map((team) => (
        <div key={team.name} className={`compact-alive-team side-${team.side}`}>
          <strong>{sideName(team)}</strong>
          <div>{team.players.map((player) => (
            <span
              key={player.instanceId ?? player.id}
              className={player.alive ? "is-alive" : "is-dead"}
              title={`${player.id}：${player.alive ? `${player.health} HP` : "已阵亡"}`}
              aria-label={`${player.id} ${player.alive ? "存活" : "已阵亡"}`}
              data-testid="alive-player"
            />
          ))}</div>
        </div>
      ))}
    </div>
  )
}

function CompactEventFeed({ model }: { model: BattleViewModel }) {
  const [expanded, setExpanded] = useState(false)
  const events = expanded ? model.visibleEvents : model.visibleEvents.slice(-5)
  return (
    <section className="compact-event-feed" aria-label="实时播报">
      <header>
        <span><Radio size={15} />实时播报</span>
        {model.visibleEvents.length > 5 && (
          <button type="button" onClick={() => setExpanded((value) => !value)}>{expanded ? "最近事件" : "全部播报"}</button>
        )}
      </header>
      <div className="compact-event-list" data-testid="compact-event-list">
        {events.length === 0 && <p>等待本回合事件...</p>}
        {events.map((event, index) => (
          <div key={event.event_id ?? `${event.timestamp}-${event.event_type}-${index}`} className={`compact-event-row type-${event.event_type.toLowerCase()}`}>
            <span>{formatCompactBattleEvent(event, model.eventContext)}</span>
            {expanded && <small>{formatBattleEvent(event, model.eventContext)}</small>}
          </div>
        ))}
      </div>
    </section>
  )
}

function CompactRoster({ model }: { model: BattleViewModel }) {
  const [selected, setSelected] = useState<"a" | "b">("a")
  const team = selected === "a" ? model.teamA : model.teamB
  return (
    <section className="compact-roster-panel" aria-label="双方阵容">
      <div className="compact-team-switch" role="tablist" aria-label="选择队伍">
        {[{ id: "a" as const, team: model.teamA }, { id: "b" as const, team: model.teamB }].map((item) => (
          <button key={item.id} type="button" role="tab" aria-selected={selected === item.id} onClick={() => setSelected(item.id)}>
            <span>{item.team.side.toUpperCase()}</span>{item.team.name}
          </button>
        ))}
      </div>
      <div className="compact-player-list">
        {team.players.map((player) => (
          <div key={player.instanceId ?? player.id} className={player.alive ? "compact-player-row" : "compact-player-row is-dead"}>
            <span className={`compact-player-side side-${team.side}`}>{team.side.toUpperCase()}</span>
            <strong>{player.id}</strong>
            <span>{player.alive ? `${player.health} HP` : "已阵亡"}</span>
            <span>{player.weapon || "无主武器"}</span>
            <small>{player.kills} / {player.deaths} / {player.assists}</small>
          </div>
        ))}
      </div>
    </section>
  )
}

export default function CompactBattleLayout({ model }: { model: BattleViewModel }) {
  const [panel, setPanel] = useState<MobilePanel>("live")
  return (
    <main className="battle-stage compact-battle-layout" data-battle-layout="compact">
      <header className="compact-battle-heading">
        <h1>{model.fullMatchRevealed ? "比赛结束" : "比赛进行中"}</h1>
        <span>第 {model.currentRound.round_number} 回合 · {model.phaseLabel}</span>
      </header>
      <Scoreboard
        teamA={model.teamA}
        teamB={model.teamB}
        round={model.currentRound.round_number}
        maxRounds={model.fullMatchRevealed ? model.report.total_rounds : undefined}
        phaseLabel={model.phaseLabel}
        winnerName={model.winnerName}
      />
      <BattleReplayControls model={model} compact />

      <nav className="compact-battle-tabs" aria-label="比赛视图">
        <button type="button" className={panel === "live" ? "is-active" : ""} onClick={() => setPanel("live")}><Radio size={16} />实时战况</button>
        <button type="button" className={panel === "roster" ? "is-active" : ""} onClick={() => setPanel("roster")}><Users size={16} />阵容</button>
        <button type="button" className={panel === "stats" ? "is-active" : ""} onClick={() => setPanel("stats")}><BarChart3 size={16} />数据统计</button>
      </nav>

      <div className="compact-battle-panel">
        {panel === "live" && (
          <section className="compact-live-stage" data-testid="compact-live-stage">
            <BombStatusStrip bomb={model.bomb} compact />
            <MapView events={model.visibleEvents} className="compact-battle-map" />
            <AlivePlayersStrip teamA={model.teamA} teamB={model.teamB} />
            <CompactEventFeed model={model} />
          </section>
        )}
        {panel === "roster" && <CompactRoster model={model} />}
        {panel === "stats" && <BattleStatsTables teamA={model.teamA} teamB={model.teamB} />}
      </div>
    </main>
  )
}
