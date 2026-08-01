import { useEffect, useMemo, useRef, useState } from "react"
import { useLocation } from "react-router-dom"
import { Pause, Play, SkipForward } from "lucide-react"
import SubPageHeader from "../components/SubPageHeader"
import Scoreboard from "../components/battle/Scoreboard"
import TeamRoster from "../components/battle/TeamRoster"
import MapView from "../components/battle/MapView"
import KillFeed from "../components/battle/KillFeed"
import EventFeed from "../components/battle/EventFeed"
import { battleState, type BattlePlayer, type BattleTeam } from "../components/battle/data/battle"
import type { BombPublicState, GameEvent, MatchReport, PlayerState, RoundReport } from "../types/match-report"

interface BattleLocationState {
  report?: MatchReport
}

const AVATAR = "/images/star-player.png"
const loggedMatchReports = new Set<string>()

function logFullMatchReport(report: MatchReport) {
  if (!report.debug_enabled || loggedMatchReports.has(report.match_info.match_id)) return
  loggedMatchReports.add(report.match_info.match_id)

  console.groupCollapsed(
    `[SimuMatch] ${report.match_info.team_a_name} vs ${report.match_info.team_b_name} · seed ${report.match_info.seed}`
  )
  console.log("[SimuMatch] 完整原始战报", report)
  console.table(
    report.rounds.map((round) => ({
      round: round.round_number,
      phase: round.phase,
      half: round.half,
      teamT: round.team_t_id,
      teamCT: round.team_ct_id,
      winner: round.winner_team_id,
      reason: round.win_reason,
      score: `${round.score_team_a}:${round.score_team_b}`,
      events: round.events.length,
      bomb: round.bomb?.status ?? "-",
    }))
  )
  report.rounds.forEach((round) => {
    console.groupCollapsed(
      `[SimuMatch][R${round.round_number}] ${round.team_t_id}(T) vs ${round.team_ct_id}(CT) · ${round.win_reason}`
    )
    console.log("回合快照", round)
    console.table(
      round.events.map((event, index) => ({
        index,
        time: event.timestamp,
        type: event.event_type,
        attacker: event.attacker_name ?? "",
        victim: event.victim_name ?? "",
        weapon: event.weapon ?? "",
        message: event.message,
        reason: event.reason?.code ?? "",
        bomb: event.bomb?.status ?? "",
      }))
    )
    console.table(
      round.player_states.map((player) => ({
        player: player.display_name || player.player_name,
        team: player.team_id,
        side: player.side,
        alive: player.alive,
        kills: player.kills,
        deaths: player.deaths,
        weapon: player.weapon.primary,
        bomb: player.has_bomb,
      }))
    )
    console.groupEnd()
  })
  console.log("[SimuMatch] 最终统计", report.final_stats)
  console.groupEnd()
}

function portraitUrl(portrait?: string) {
  if (!portrait) return AVATAR
  if (portrait.startsWith("/") || portrait.startsWith("http://") || portrait.startsWith("https://") || portrait.startsWith("data:")) {
    return portrait
  }
  return `/${portrait.replace(/^\.\//, "")}`
}

function sideToUi(side: "T" | "CT") {
  return side.toLowerCase() as "t" | "ct"
}

function teamSide(round: RoundReport, teamID: string): "T" | "CT" {
  return round.team_t_id === teamID ? "T" : "CT"
}

function phaseLabel(round: RoundReport) {
  if (round.phase === "overtime") {
    return `OT${round.overtime_block ?? 1} · ${round.overtime_round_in_block ?? 1}/6`
  }
  return round.is_side_switch ? "半场换边" : `半场 ${round.half}`
}

function latestBomb(round: RoundReport, events: GameEvent[]): BombPublicState | undefined {
  for (let i = events.length - 1; i >= 0; i--) {
    if (events[i].bomb) return events[i].bomb
  }
  return round.bomb
}

function buildBattlePlayer(
  state: PlayerState,
  aliveMap: Map<string, boolean>,
  killsMap: Map<string, number>,
  deathsMap: Map<string, number>
): BattlePlayer {
  const alive = aliveMap.get(state.player_id) ?? true
  return {
    id: state.player_name || state.display_name,
    avatar: portraitUrl(state.portrait),
    alive,
    health: alive ? state.hp || 100 : 0,
    armor: state.weapon.armor ? 100 : 0,
    helmet: state.weapon.helmet,
    money: 0,
    kills: killsMap.get(state.player_id) ?? 0,
    deaths: deathsMap.get(state.player_id) ?? 0,
    assists: 0,
    weapon: state.weapon.primary === "AK47" ? "AK-47" : state.weapon.primary === "M4A1S" ? "M4A1-S" : state.weapon.primary,
    defuseKit: state.weapon.has_kit,
    grenades: [],
  }
}

function buildTeam(name: string, tag: string, side: "t" | "ct", score: number, players: BattlePlayer[]): BattleTeam {
  return { name, tag, side, score, players }
}

function previousScore(report: MatchReport, roundIndex: number) {
  if (roundIndex <= 0) return { teamA: 0, teamB: 0 }
  const previous = report.rounds[roundIndex - 1]
  return { teamA: previous.score_team_a, teamB: previous.score_team_b }
}

function scoreAtPlayback(report: MatchReport, round: RoundReport, roundIndex: number, events: GameEvent[]) {
  const hasRoundEnd = events.some((ev) => ev.event_type === "ROUND_END")
  if (hasRoundEnd) return { teamA: round.score_team_a, teamB: round.score_team_b }
  return previousScore(report, roundIndex)
}

function winnerName(report: MatchReport) {
  if (report.winner_team_id === report.match_info.team_a_id) return report.match_info.team_a_name
  if (report.winner_team_id === report.match_info.team_b_id) return report.match_info.team_b_name
  return ""
}

function BombWidget({ bomb }: { bomb?: BombPublicState }) {
  const labelMap: Record<string, string> = {
    Carried: "携带中",
    Dropped: "已掉落",
    Planted: "已下包",
    Defused: "已拆除",
    Exploded: "已爆炸",
  }
  const status = bomb?.status ?? "Carried"
  return (
    <div className="flex h-12 items-center justify-between rounded-md bg-panel/70 px-4 text-sm ring-1 ring-white/10">
      <span className="font-semibold text-foreground/80">炸弹</span>
      <span className="rounded-md bg-amber-400/15 px-3 py-1 font-bold text-amber-200 ring-1 ring-amber-300/40">
        {labelMap[status] ?? status}
      </span>
      <span className="w-32 truncate text-right text-muted">{bomb?.site ? `${bomb.site} 区` : bomb?.node_id ?? "未下包"}</span>
    </div>
  )
}

export default function BattlePage() {
  const location = useLocation()
  const { report } = (location.state as BattleLocationState) ?? {}
  const [roundIndex, setRoundIndex] = useState(0)
  const [furthestRoundIndex, setFurthestRoundIndex] = useState(0)
  const [visibleEventCount, setVisibleEventCount] = useState(1)
  const [playing, setPlaying] = useState(Boolean(report))
  const [fullMatchRevealed, setFullMatchRevealed] = useState(false)
  const loggedPlaybackEvents = useRef(new Set<string>())

  const currentRound = report?.rounds[roundIndex]

  useEffect(() => {
    if (report) logFullMatchReport(report)
  }, [report])

  useEffect(() => {
    if (!report || !currentRound || !playing) return
    const timer = setInterval(() => {
      setVisibleEventCount((count) => {
        if (count < currentRound.events.length) return count + 1
        if (roundIndex < report.rounds.length - 1) {
          const nextRoundIndex = roundIndex + 1
          setRoundIndex(nextRoundIndex)
          setFurthestRoundIndex((idx) => Math.max(idx, nextRoundIndex))
          return 1
        }
        setPlaying(false)
        setFullMatchRevealed(true)
        return count
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [report, currentRound, playing, roundIndex])

  const visibleEvents = useMemo(
    () => currentRound?.events.slice(0, Math.min(visibleEventCount, currentRound.events.length)) ?? [],
    [currentRound, visibleEventCount]
  )

  useEffect(() => {
    if (!report?.debug_enabled || !currentRound) return
    visibleEvents.forEach((event, eventIndex) => {
      const key = `${currentRound.round_number}:${eventIndex}`
      if (loggedPlaybackEvents.current.has(key)) return
      loggedPlaybackEvents.current.add(key)
      console.log(`[SimuMatch][播放][R${currentRound.round_number}][${eventIndex + 1}/${currentRound.events.length}]`, event)
    })
  }, [report, currentRound, visibleEvents])

  const playback = useMemo(() => {
    if (!report || !currentRound) return null
    const alive = new Map<string, boolean>()
    const kills = new Map<string, number>()
    const deaths = new Map<string, number>()
    currentRound.player_states.forEach((p) => alive.set(p.player_id, true))
    visibleEvents.forEach((ev) => {
      if (ev.event_type !== "KILL") return
      if (ev.victim_id) {
        alive.set(ev.victim_id, false)
        deaths.set(ev.victim_id, (deaths.get(ev.victim_id) ?? 0) + 1)
      }
      if (ev.attacker_id) kills.set(ev.attacker_id, (kills.get(ev.attacker_id) ?? 0) + 1)
    })

    const score = scoreAtPlayback(report, currentRound, roundIndex, visibleEvents)
    const teamAPlayers = currentRound.player_states
      .filter((p) => p.team_id === report.match_info.team_a_id)
      .map((p) => buildBattlePlayer(p, alive, kills, deaths))
    const teamBPlayers = currentRound.player_states
      .filter((p) => p.team_id === report.match_info.team_b_id)
      .map((p) => buildBattlePlayer(p, alive, kills, deaths))
    const sideA = sideToUi(teamSide(currentRound, report.match_info.team_a_id))
    const sideB = sideToUi(teamSide(currentRound, report.match_info.team_b_id))
    return {
      teamA: buildTeam(report.match_info.team_a_name, sideA.toUpperCase(), sideA, score.teamA, teamAPlayers),
      teamB: buildTeam(report.match_info.team_b_name, sideB.toUpperCase(), sideB, score.teamB, teamBPlayers),
      bomb: latestBomb(currentRound, visibleEvents),
    }
  }, [report, currentRound, roundIndex, visibleEvents])

  if (!report || !currentRound || !playback) {
    const { title, round: mockRound, maxRounds, teamA: mockA, teamB: mockB, killFeed } = battleState
    return (
      <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
        <main className="flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
          <SubPageHeader title={title} hideBack />
          <div className="px-[40px] pt-3 text-sm text-muted">暂无模拟战报，显示训练预览。</div>
          <Scoreboard teamA={mockA} teamB={mockB} round={mockRound} maxRounds={maxRounds} />
          <div className="flex min-h-0 flex-1 gap-4 px-[40px] pb-[24px] pt-2">
            <TeamRoster team={mockA} align="left" />
            <div className="flex min-h-0 flex-1 flex-col gap-3">
              <MapView />
              <KillFeed events={killFeed} />
            </div>
            <TeamRoster team={mockB} align="right" />
          </div>
        </main>
      </div>
    )
  }

  const { teamA, teamB, bomb } = playback
  const visibleRounds = fullMatchRevealed ? report.rounds : report.rounds.slice(0, furthestRoundIndex + 1)

  return (
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      <main className="flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
        <SubPageHeader title={fullMatchRevealed ? "比赛结束" : "比赛进行中"} hideBack />

        <Scoreboard
          teamA={teamA}
          teamB={teamB}
          round={currentRound.round_number}
          maxRounds={fullMatchRevealed ? report.total_rounds : undefined}
          phaseLabel={phaseLabel(currentRound)}
          winnerName={fullMatchRevealed ? winnerName(report) : undefined}
        />

        <div className="flex h-16 shrink-0 items-center gap-3 px-[40px]">
          <button
            type="button"
            onClick={() => setPlaying((value) => !value)}
            className="grid h-10 w-10 place-items-center rounded-md bg-panel/80 text-foreground ring-1 ring-white/10 transition hover:bg-panel"
            title={playing ? "暂停" : "继续"}
          >
            {playing ? <Pause size={18} /> : <Play size={18} />}
          </button>
          {!fullMatchRevealed && (
            <button
              type="button"
              onClick={() => {
                const finalRoundIndex = report.rounds.length - 1
                if (report.debug_enabled) {
                  console.log(`[SimuMatch] 玩家跳过比赛，直接显示最终结果 ${report.final_score_team_a}:${report.final_score_team_b}`)
                }
                setFullMatchRevealed(true)
                setPlaying(false)
                setRoundIndex(finalRoundIndex)
                setFurthestRoundIndex(finalRoundIndex)
                setVisibleEventCount(report.rounds[finalRoundIndex].events.length)
              }}
              className="flex h-10 items-center gap-2 rounded-md bg-gold px-4 text-sm font-bold text-background transition hover:bg-gold/90"
            >
              <SkipForward size={17} />
              跳过比赛
            </button>
          )}
          <div className="min-w-0 flex-1 overflow-x-auto">
            <div className="flex gap-1">
              {visibleRounds.map((round, idx) => (
                <button
                  key={round.round_number}
                  type="button"
                  onClick={() => {
                    setRoundIndex(idx)
                    setVisibleEventCount(round.events.length)
                    setPlaying(false)
                  }}
                  className={`h-9 min-w-11 rounded-md px-2 text-xs font-bold tabular-nums ring-1 transition ${
                    idx === roundIndex
                      ? "bg-primary text-primary-foreground ring-primary"
                      : round.phase === "overtime"
                        ? "bg-amber-400/10 text-amber-200 ring-amber-400/30"
                        : "bg-panel/70 text-muted ring-white/10 hover:text-foreground"
                  }`}
                >
                  {round.round_number}
                </button>
              ))}
            </div>
          </div>
          <div className="w-56 text-right text-xs text-muted">
            Seed {report.match_info.seed} · {report.match_info.rule_set_id}
          </div>
        </div>

        <div className="flex min-h-0 flex-1 gap-4 px-[40px] pb-[24px] pt-2">
          <TeamRoster team={teamA} align="left" />

          <div className="flex min-h-0 flex-1 flex-col gap-3">
            <BombWidget bomb={bomb} />
            <MapView events={visibleEvents} />
            <EventFeed
              events={visibleEvents}
              teamA={teamA}
              teamB={teamB}
              finalStats={fullMatchRevealed ? report.final_stats : undefined}
            />
          </div>

          <TeamRoster team={teamB} align="right" />
        </div>
      </main>
    </div>
  )
}
