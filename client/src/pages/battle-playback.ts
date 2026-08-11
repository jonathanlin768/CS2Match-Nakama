import type { BombPublicState, GameEvent, MatchReport, RoundReport } from "../types/match-report"

export interface TeamScore {
  teamA: number
  teamB: number
}

export interface CumulativePlayerStats {
  kills: number
  deaths: number
  assists: number
}

export interface PlaybackPlayerVitals {
  alive: boolean
  hp: number
}

export interface BattleEventContext {
  teamAID: string
  teamAName: string
  teamBID: string
  teamBName: string
  teamTID: string
  teamCTID: string
  winnerTeamID: string
  winReason: string
  strategyTemplateID: string
  ctSetupTemplateID?: string
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
    const bomb = events[index].bomb ?? events[index].state?.bomb
    if (bomb) return bomb
  }
  return round.bomb
}

export function selectedRoundInitialEventCount(round: RoundReport): number {
	if (round.events.length === 0) return 0
	const roundStartIndex = round.events.findIndex((event) => event.event_type === "ROUND_START")
	return roundStartIndex >= 0 ? roundStartIndex + 1 : 1
}

export function cumulativePlayerStatsAtPlayback(
  report: MatchReport,
  roundIndex: number,
  visibleEvents: GameEvent[],
): Record<string, CumulativePlayerStats> {
  const stats: Record<string, CumulativePlayerStats> = {}
  const ensure = (playerID: string) => {
    if (!stats[playerID]) stats[playerID] = { kills: 0, deaths: 0, assists: 0 }
    return stats[playerID]
  }
  const apply = (event: GameEvent) => {
    if (event.event_type !== "KILL") return
    if (event.attacker_id) ensure(event.attacker_id).kills++
    if (event.victim_id) ensure(event.victim_id).deaths++
    for (const assistantID of eventAssistIDs(event)) {
      if (assistantID !== event.attacker_id && assistantID !== event.victim_id) ensure(assistantID).assists++
    }
  }
  for (let index = 0; index < roundIndex; index++) {
    report.rounds[index]?.events.forEach(apply)
  }
  visibleEvents.forEach(apply)
  return stats
}

export function playerVitalsAtPlayback(
  round: RoundReport,
  visibleEvents: GameEvent[],
): Record<string, PlaybackPlayerVitals> {
  const vitals: Record<string, PlaybackPlayerVitals> = {}
  round.player_states.forEach((player) => {
    vitals[player.player_id] = { alive: true, hp: 100 }
  })

  visibleEvents.forEach((event) => {
    if (event.event_type === "DAMAGE" && event.victim_id) {
      const current = vitals[event.victim_id] ?? { alive: true, hp: 100 }
      const damage = numericExtra(event, "damage")
      const snapshot = event.state?.players?.find((player) => player.player_id === event.victim_id)
      const hp = damage === undefined
        ? Math.max(0, Math.min(100, snapshot?.hp ?? current.hp))
        : Math.max(0, current.hp - damage)
      vitals[event.victim_id] = { alive: hp > 0, hp }
    }
    if (event.event_type === "KILL" && event.victim_id) {
      vitals[event.victim_id] = { alive: false, hp: 0 }
    }
  })

  return vitals
}

export function formatBattleEvent(event: GameEvent, context: BattleEventContext): string {
  const attacker = event.attacker_name || event.attacker_id || "未知选手"
  const victim = event.victim_name || event.victim_id || "未知选手"
  const location = event.location?.name || event.bomb?.node_id || event.state?.bomb?.node_id || "未知位置"
  const damage = numericExtra(event, "damage")
  const actorTeam = teamName(event.attacker_team_id, context)
  const decisionType = stringExtra(event, "decision_type")
  const targetNode = stringExtra(event, "target_node") || location

	switch (event.event_type) {
		case "MATCH_START":
			return `${context.teamAName} 对阵 ${context.teamBName} 的比赛开始。`
		case "ROUND_START":
      return `${teamName(context.teamTID, context)} 作为 T 执行 ${context.strategyTemplateID} 战术；${teamName(context.teamCTID, context)} 作为 CT 使用 ${context.ctSetupTemplateID || "默认"} 防守。`
    case "STRATEGY_ADJUSTED":
      return `${actorTeam} 根据此前回合调整了战术评分。`
    case "DAMAGE":
      return `${attacker} 在 ${location} 攻击了 ${victim}，造成 ${damage ?? 0} 点伤害。`
    case "KILL": {
      const suffix = event.is_trade ? "（补枪）" : event.is_first_kill ? "（首杀）" : ""
      return `${attacker} 在 ${location} 使用 ${event.weapon || "武器"} 击杀了 ${victim}${suffix}。`
    }
    case "ROTATE":
      return `${actorTeam} 执行 ${decisionType || "转点"}，向 ${targetNode} 推进。`
    case "REINFORCE":
      return `${actorTeam} 执行 ${decisionType || "补防"}，前往 ${targetNode}。`
    case "CONTROL_GAINED":
      return `${actorTeam} 取得了 ${location} 的控制权。`
    case "BOMB_DROP":
      return `${event.victim_name || event.victim_id || attacker} 在 ${location} 丢掉了炸弹。`
    case "BOMB_PICKUP":
      return `${attacker} 在 ${location} 拾取了炸弹。`
    case "BOMB_PLANT_START":
      return `${attacker} 开始在 ${event.bomb?.site || location} 安放炸弹。`
    case "BOMB_PLANT_INTERRUPT":
      return `${event.victim_name || event.victim_id || attacker} 的安包被打断。`
    case "BOMB_PLANT":
      return `${attacker} 在 ${event.bomb?.site || location} 安放了炸弹。`
    case "DEFUSE_START":
      return `${attacker} 开始拆除炸弹。`
    case "DEFUSE_INTERRUPT":
      return `${event.victim_name || event.victim_id || attacker} 的拆包被打断。`
    case "BOMB_DEFUSE":
      return `炸弹被 ${attacker} 拆除。`
    case "BOMB_EXPLODE":
      return `炸弹在 ${event.bomb?.site || location} 爆炸。`
		case "ROUND_END": {
      const winner = teamName(context.winnerTeamID, context)
      const side = context.winnerTeamID === context.teamTID ? "T" : "CT"
			return `${winner} 作为 ${side} 以${winReasonLabel(context.winReason)}赢得本回合。`
		}
		case "HALF_TIME":
			return `上半场结束，双方准备换边。`
		case "SIDE_SWITCH":
			return `${teamName(context.teamTID, context)} 转为 T，${teamName(context.teamCTID, context)} 转为 CT。`
		case "OVERTIME_START":
			return `常规时间战平，比赛进入加时。`
		case "MATCH_END":
			return `${teamName(context.winnerTeamID, context)} 赢得整场比赛。`
		default:
      return event.message
  }
}

function eventAssistIDs(event: GameEvent): string[] {
  const values = event.extra?.assist_ids
  return Array.isArray(values) ? values.filter((value): value is string => typeof value === "string" && value.length > 0) : []
}

function numericExtra(event: GameEvent, key: string): number | undefined {
  const value = event.extra?.[key]
  return typeof value === "number" && Number.isFinite(value) ? value : undefined
}

function stringExtra(event: GameEvent, key: string): string {
  const value = event.extra?.[key]
  return typeof value === "string" ? value : ""
}

function teamName(teamID: string | undefined, context: BattleEventContext): string {
  if (teamID === context.teamAID) return context.teamAName
  if (teamID === context.teamBID) return context.teamBName
  return teamID || "未知队伍"
}

function winReasonLabel(reason: string): string {
  const labels: Record<string, string> = {
    elimination: "消灭对手的方式",
    timeout: "拖延至超时的方式",
    bomb_defused: "拆除炸弹的方式",
    bomb_exploded: "引爆炸弹的方式",
    bomb_secured: "守住炸弹的方式",
    no_progress_timeout: "迫使对手无法推进的方式",
  }
  return labels[reason] ?? `${reason || "有效终局"}的方式`
}
