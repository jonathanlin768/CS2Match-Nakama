import type { BattleEventContext } from "../../pages/battle-playback"
import type { BombPublicState, GameEvent, MatchReport, RoundReport } from "../../types/match-report"
import type { BattleTeam } from "./data/battle"

export interface BattleViewModel {
  report: MatchReport
  currentRound: RoundReport
  visibleRounds: RoundReport[]
  visibleEvents: GameEvent[]
  teamA: BattleTeam
  teamB: BattleTeam
  bomb?: BombPublicState
  eventContext: BattleEventContext
  roundIndex: number
  playing: boolean
  fullMatchRevealed: boolean
  phaseLabel: string
  winnerName?: string
  onTogglePlaying: () => void
  onSkipMatch: () => void
  onSelectRound: (roundIndex: number) => void
}
