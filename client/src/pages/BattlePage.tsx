import { useEffect, useMemo, useState } from "react"
import { useLocation } from "react-router-dom"
import SubPageHeader from "../components/SubPageHeader"
import Scoreboard from "../components/battle/Scoreboard"
import TeamRoster from "../components/battle/TeamRoster"
import MapView from "../components/battle/MapView"
import KillFeed from "../components/battle/KillFeed"
import EventFeed from "../components/battle/EventFeed"
import { battleState, type BattlePlayer, type BattleTeam } from "../components/battle/data/battle"
import type { MatchReport, PlayerState } from "../types/match-report"

interface BattleLocationState {
  report?: MatchReport
}

const AVATAR = "/images/star-player.png"
const MAX_ROUNDS = 24

/**
 * 将服务器返回的选手状态转换为 UI 需要的 BattlePlayer。
 * aliveMap 由当前已播放的事件动态决定。
 */
function buildBattlePlayer(
  state: PlayerState,
  aliveMap: Map<string, boolean>,
  killsMap: Map<string, number>,
  deathsMap: Map<string, number>
): BattlePlayer {
  const alive = aliveMap.get(state.player_id) ?? state.is_alive
  const side = state.side.toLowerCase() as "t" | "ct"
  return {
    id: state.player_name,
    avatar: AVATAR,
    alive,
    health: alive ? 100 : 0,
    armor: 100,
    helmet: true,
    money: 0,
    kills: killsMap.get(state.player_id) ?? state.kills,
    deaths: deathsMap.get(state.player_id) ?? state.deaths,
    assists: 0,
    weapon: side === "t" ? "AK-47" : "M4A4",
    defuseKit: side === "ct",
    grenades: [],
  }
}

function buildTeam(name: string, tag: string, side: "t" | "ct", players: BattlePlayer[]): BattleTeam {
  return { name, tag, side, score: 0, players }
}

export default function BattlePage() {
  const location = useLocation()
  const { report } = (location.state as BattleLocationState) ?? {}

  const round = report?.rounds[0]
  const allEvents = round?.events ?? []

  // 控制事件播放：每 1 秒显示下一条
  const [eventIndex, setEventIndex] = useState(0)
  useEffect(() => {
    if (!report || allEvents.length === 0) return
    setEventIndex(0)
    const timer = setInterval(() => {
      setEventIndex((idx) => {
        if (idx >= allEvents.length - 1) {
          clearInterval(timer)
          return idx
        }
        return idx + 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [report, allEvents.length])

  const visibleEvents = useMemo(() => allEvents.slice(0, eventIndex + 1), [allEvents, eventIndex])

  // 根据已播放事件计算存活、击杀、死亡、比分
  const { aliveMap, killsMap, deathsMap, currentScore } = useMemo(() => {
    const alive = new Map<string, boolean>()
    const kills = new Map<string, number>()
    const deaths = new Map<string, number>()
    let scoreT = 0
    let scoreCT = 0

    // 初始全部存活
    round?.player_states.forEach((p) => alive.set(p.player_id, true))

    visibleEvents.forEach((ev) => {
      if (ev.event_type === "KILL") {
        if (ev.victim_id) alive.set(ev.victim_id, false)
        if (ev.attacker_id) kills.set(ev.attacker_id, (kills.get(ev.attacker_id) ?? 0) + 1)
        if (ev.victim_id) deaths.set(ev.victim_id, (deaths.get(ev.victim_id) ?? 0) + 1)
      }
      if (ev.event_type === "ROUND_END") {
        scoreT = round?.score_t ?? 0
        scoreCT = round?.score_ct ?? 0
      }
    })

    return { aliveMap: alive, killsMap: kills, deathsMap: deaths, currentScore: { scoreT, scoreCT } }
  }, [visibleEvents, round])

  // 构建左右两队（T 左，CT 右）
  const teams = useMemo(() => {
    if (!report || !round) return null

    const teamAPlayers: BattlePlayer[] = []
    const teamBPlayers: BattlePlayer[] = []

    round.player_states.forEach((p) => {
      const bp = buildBattlePlayer(p, aliveMap, killsMap, deathsMap)
      if (p.side === "T") {
        teamAPlayers.push(bp)
      } else {
        teamBPlayers.push(bp)
      }
    })

    return {
      teamA: buildTeam(report.match_info.team_a_name, "T", "t", teamAPlayers),
      teamB: buildTeam(report.match_info.team_b_name, "CT", "ct", teamBPlayers),
    }
  }, [report, round, aliveMap, killsMap, deathsMap])

  // 控制台打印完整战报
  useEffect(() => {
    if (report) {
      // eslint-disable-next-line no-console
      console.log("[DebugSimuMatch] 完整战报:", report)
    }
  }, [report])

  // 没有战报时回退到 mock UI
  if (!report || !round || !teams) {
    const { title, round: mockRound, maxRounds, teamA: mockA, teamB: mockB, killFeed } = battleState
    return (
      <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
        <main className="flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
          <SubPageHeader title={title} hideBack />
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

  const { teamA, teamB } = teams

  // 应用当前比分
  teamA.score = currentScore.scoreT
  teamB.score = currentScore.scoreCT

  return (
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      <main className="flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
        <SubPageHeader title="比赛进行中" hideBack />

        <Scoreboard
          teamA={teamA}
          teamB={teamB}
          round={round.round_number}
          maxRounds={MAX_ROUNDS}
        />

        <div className="flex min-h-0 flex-1 gap-4 px-[40px] pb-[24px] pt-2">
          <TeamRoster team={teamA} align="left" />

          <div className="flex min-h-0 flex-1 flex-col gap-3">
            <MapView events={visibleEvents} />
            <EventFeed events={visibleEvents} teamA={teamA} teamB={teamB} />
          </div>

          <TeamRoster team={teamB} align="right" />
        </div>
      </main>
    </div>
  )
}
