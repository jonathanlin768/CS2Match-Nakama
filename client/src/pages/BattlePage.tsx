import { useEffect, useMemo, useRef, useState } from "react"
import { useLocation } from "react-router-dom"
import { Swords } from "lucide-react"
import CompactBattleLayout from "../components/battle/CompactBattleLayout"
import DesktopBattleLayout from "../components/battle/DesktopBattleLayout"
import type { BattleViewModel } from "../components/battle/battle-view-model"
import type { BattlePlayer, BattleTeam } from "../components/battle/data/battle"
import { useMediaQuery } from "../hooks/useMediaQuery"
import type { MatchReport, PlayerState, RoundReport } from "../types/match-report"
import {
  authoritativeScoreAtPlayback,
  cumulativePlayerStatsAtPlayback,
  latestBombAtPlayback,
  playerVitalsAtPlayback,
  selectedRoundInitialEventCount,
  teamSideAtRound,
  type CumulativePlayerStats,
  type PlaybackPlayerVitals,
} from "./battle-playback"

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

function phaseLabel(round: RoundReport) {
  if (round.phase === "overtime") {
    return `OT${round.overtime_block ?? 1} · ${round.overtime_round_in_block ?? 1}/6`
  }
  return round.is_side_switch ? "半场换边" : `半场 ${round.half}`
}

function buildBattlePlayer(
  state: PlayerState,
  vitals: PlaybackPlayerVitals,
  stats: CumulativePlayerStats,
): BattlePlayer {
  return {
    instanceId: state.player_id,
    configPlayerId: state.config_player_id,
    id: state.player_name || state.display_name,
    avatar: portraitUrl(state.portrait),
    portrait: state.portrait,
    cardImage: state.card_image,
    avatarCrop: state.avatar_crop,
    alive: vitals.alive,
    health: vitals.alive ? vitals.hp : 0,
    armor: state.weapon.armor ? 100 : 0,
    helmet: state.weapon.helmet,
    money: 0,
    kills: stats.kills,
    deaths: stats.deaths,
    assists: stats.assists,
    weapon: state.weapon.primary === "AK47" ? "AK-47" : state.weapon.primary === "M4A1S" ? "M4A1-S" : state.weapon.primary,
    defuseKit: state.weapon.has_kit,
    grenades: [],
  }
}

function buildTeam(name: string, tag: string, side: "t" | "ct", score: number, players: BattlePlayer[]): BattleTeam {
  return { name, tag, side, score, players }
}

function winnerName(report: MatchReport) {
  if (report.winner_team_id === report.match_info.team_a_id) return report.match_info.team_a_name
  if (report.winner_team_id === report.match_info.team_b_id) return report.match_info.team_b_name
  return ""
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
  const isCompact = useMediaQuery("(max-width: 1023px)")

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
    const score = authoritativeScoreAtPlayback(report, currentRound, roundIndex, visibleEvents)
    const cumulativeStats = cumulativePlayerStatsAtPlayback(report, roundIndex, visibleEvents)
    const playerVitals = playerVitalsAtPlayback(currentRound, visibleEvents)
    const playerStats = (playerID: string) => cumulativeStats[playerID] ?? { kills: 0, deaths: 0, assists: 0 }
    const vitals = (playerID: string) => playerVitals[playerID] ?? { alive: true, hp: 100 }
    const teamAPlayers = currentRound.player_states
      .filter((p) => p.team_id === report.match_info.team_a_id)
      .map((p) => buildBattlePlayer(p, vitals(p.player_id), playerStats(p.player_id)))
    const teamBPlayers = currentRound.player_states
      .filter((p) => p.team_id === report.match_info.team_b_id)
      .map((p) => buildBattlePlayer(p, vitals(p.player_id), playerStats(p.player_id)))
    const sideA = teamSideAtRound(currentRound, report.match_info.team_a_id)
    const sideB = teamSideAtRound(currentRound, report.match_info.team_b_id)
    return {
      teamA: buildTeam(report.match_info.team_a_name, sideA.toUpperCase(), sideA, score.teamA, teamAPlayers),
      teamB: buildTeam(report.match_info.team_b_name, sideB.toUpperCase(), sideB, score.teamB, teamBPlayers),
      bomb: latestBombAtPlayback(currentRound, visibleEvents),
    }
  }, [report, currentRound, roundIndex, visibleEvents])

  if (!report || !currentRound || !playback) {
    return (
      <div className="battle-empty">
        <Swords size={38} /><h1>还没有比赛战报</h1>
        <p>请从匹配对战或 15 元教学进入。回放页不会使用虚构比分或假选手数据。</p>
        <button className="primary-button" onClick={() => window.history.back()}>返回选择比赛</button>
      </div>
    )
  }

  const { teamA, teamB, bomb } = playback
  const visibleRounds = fullMatchRevealed ? report.rounds : report.rounds.slice(0, furthestRoundIndex + 1)
  const eventContext = {
    teamAID: report.match_info.team_a_id,
    teamAName: report.match_info.team_a_name,
    teamBID: report.match_info.team_b_id,
    teamBName: report.match_info.team_b_name,
    teamTID: currentRound.team_t_id,
    teamCTID: currentRound.team_ct_id,
    winnerTeamID: currentRound.winner_team_id,
    winReason: currentRound.win_reason,
    strategyTemplateID: currentRound.strategy_template_id,
    ctSetupTemplateID: currentRound.ct_setup_template_id,
  }

  const handleSkipMatch = () => {
    const finalRoundIndex = report.rounds.length - 1
    if (report.debug_enabled) {
      console.log(`[SimuMatch] 玩家跳过比赛，直接显示最终结果 ${report.final_score_team_a}:${report.final_score_team_b}`)
    }
    setFullMatchRevealed(true)
    setPlaying(false)
    setRoundIndex(finalRoundIndex)
    setFurthestRoundIndex(finalRoundIndex)
    setVisibleEventCount(report.rounds[finalRoundIndex].events.length)
  }

  const handleSelectRound = (index: number) => {
    const round = report.rounds[index]
    if (!round || index > furthestRoundIndex && !fullMatchRevealed) return
    setRoundIndex(index)
    setVisibleEventCount(selectedRoundInitialEventCount(round))
    setPlaying(false)
  }

  const model: BattleViewModel = {
    report,
    currentRound,
    visibleRounds,
    visibleEvents,
    teamA,
    teamB,
    bomb,
    eventContext,
    roundIndex,
    playing,
    fullMatchRevealed,
    phaseLabel: phaseLabel(currentRound),
    winnerName: fullMatchRevealed ? winnerName(report) : undefined,
    onTogglePlaying: () => setPlaying((value) => !value),
    onSkipMatch: handleSkipMatch,
    onSelectRound: handleSelectRound,
  }

  return (
    <div className={`battle-shell ${isCompact ? "is-compact" : ""}`}>
      {isCompact ? <CompactBattleLayout model={model} /> : <DesktopBattleLayout model={model} />}
    </div>
  )
}
