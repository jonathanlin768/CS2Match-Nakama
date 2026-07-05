/** 战报事件类型 */
export type GameEventType = "ROUND_START" | "KILL" | "ROUND_END" | "BOMB_PLANT" | "BOMB_DEFUSE" | "BOMB_EXPLODE" | "MATCH_START" | "MATCH_END" | "HALF_TIME"

/** 事件在地图上的位置（比例 0.0 ~ 1.0） */
export interface EventLocation {
  name: string
  x: number
  y: number
}

/** 单条游戏事件（HLTV 播报原始数据） */
export interface GameEvent {
  timestamp: number
  event_type: GameEventType
  attacker_id?: string
  attacker_name?: string
  victim_id?: string
  victim_name?: string
  weapon?: string
  location?: EventLocation
  is_first_kill?: boolean
  is_trade?: boolean
  message: string
  extra?: Record<string, unknown>
}

/** 回合结束时选手状态 */
export interface PlayerState {
  player_id: string
  player_name: string
  side: "T" | "CT"
  is_alive: boolean
  kills: number
  deaths: number
  damage: number
}

/** 单回合战报 */
export interface RoundReport {
  round_number: number
  side_attacking: "T" | "CT"
  winner: "T" | "CT"
  win_reason: string
  score_t: number
  score_ct: number
  route_main: string
  route_sub: string
  events: GameEvent[]
  player_states: PlayerState[]
}

/** 比赛元信息 */
export interface MatchInfo {
  match_id: string
  map_id: string
  map_name: string
  team_a_name: string
  team_b_name: string
  start_time: number
  total_rounds: number
}

/** 选手整场比赛统计 */
export interface PlayerMatchStats {
  player_id: string
  player_name: string
  side: "T" | "CT"
  kills: number
  deaths: number
  adr: number
  fk: number
  mk: number
}

/** 最终统计 */
export interface FinalStats {
  score_t: number
  score_ct: number
  player_stats: PlayerMatchStats[]
}

/** 完整战报响应（DebugSimuMatch 返回体） */
export interface MatchReport {
  match_info: MatchInfo
  rounds: RoundReport[]
  final_stats: FinalStats
  winner: "T" | "CT"
}
