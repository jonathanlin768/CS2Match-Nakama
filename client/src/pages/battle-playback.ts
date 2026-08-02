import type { BombPublicState, GameEvent, MatchReport, RoundReport } from "../types/match-report"

export interface TeamScore {
  teamA: number
  teamB: number
}

export function previousTeamScore(report: MatchReport, roundIndex: number): TeamScore {
  if (roundIndex <= 0) return { teamA: 0, teamB: 0 }
  const previous = report.rounds[roundIndex - 1]
  return { teamA: previous.score_team_a, teamB: previous.score_team_b }
}

export function authoritativeScoreAtPlayback(report: MatchReport, round: RoundReport, roundIndex: number, events: GameEvent[]): TeamScore {
  const hasRoundEnd = events.some((event) => event.event_type === "ROUND_END")
  if (hasRoundEnd) return { teamA: round.score_team_a, teamB: round.score_team_b }
  return previousTeamScore(report, roundIndex)
}

export function latestBombAtPlayback(round: RoundReport, events: GameEvent[]): BombPublicState | undefined {
  for (let index = events.length - 1; index >= 0; index--) {
    if (events[index].bomb) return events[index].bomb
  }
  return round.bomb
}
