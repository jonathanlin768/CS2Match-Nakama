export type Side = "T" | "CT"
export type RoundPhase = "regulation" | "overtime"

export type GameEventType =
  | "MATCH_START"
  | "ROUND_START"
  | "HALF_TIME"
  | "SIDE_SWITCH"
  | "OVERTIME_START"
  | "KILL"
  | "BOMB_PLANT"
  | "BOMB_DEFUSE"
  | "BOMB_EXPLODE"
  | "ROUND_END"
  | "MATCH_END"

export interface EventLocation {
  name: string
  x: number
  y: number
}

export interface EventReason {
  code: string
  main_factor: string
  score_delta: number
  detail?: string
}

export interface WeaponLoadout {
  primary: string
  secondary?: string
  armor: boolean
  helmet: boolean
  has_kit?: boolean
  grenades?: string[]
}

export interface BombPublicState {
  status: "Carried" | "Dropped" | "Planted" | "Defused" | "Exploded"
  carrier_id?: string
  node_id?: string
  site?: string
  planted_at?: number
  explode_at?: number
  dropped_at?: number
}

export interface NodeControlState {
  node_id: string
  status: "Unknown" | "TControlled" | "CTControlled" | "Contested" | "EmptyKnown"
  known_by_t: boolean
  known_by_ct: boolean
  updated_at: number
}

export interface GameEvent {
  timestamp: number
  event_type: GameEventType
  attacker_id?: string
  attacker_name?: string
  attacker_team_id?: string
  victim_id?: string
  victim_name?: string
  victim_team_id?: string
  weapon?: string
  location?: EventLocation
  is_first_kill?: boolean
  is_trade?: boolean
  message: string
  reason?: EventReason
  bomb?: BombPublicState
  score_team_a?: number
  score_team_b?: number
  extra?: Record<string, unknown>
}

export interface PlayerState {
  player_id: string
  player_name: string
  display_name: string
  portrait?: string
  team_id: string
  side: Side
  is_alive: boolean
  alive: boolean
  hp: number
  stamina: number
  focus: number
  current_node?: string
  has_bomb?: boolean
  kills: number
  deaths: number
  damage: number
  role_tags?: string[]
  weapon: WeaponLoadout
}

export interface RoundReport {
  round_number: number
  phase: RoundPhase
  half: number
  overtime_block?: number
  overtime_round_in_block?: number
  is_side_switch?: boolean
  seed: number
  side_attacking: Side
  team_t_id: string
  team_ct_id: string
  winner: Side
  winner_team_id: string
  win_reason: string
  score_team_a: number
  score_team_b: number
  score_t: number
  score_ct: number
  route_main: string
  route_sub?: string
  strategy_template_id: string
  events: GameEvent[]
  player_states: PlayerState[]
  bomb?: BombPublicState
  final_controls?: NodeControlState[]
}

export interface MatchInfo {
  match_id: string
  map_id: string
  map_name: string
  map_version: string
  rule_set_id: string
  seed: number
  team_a_id: string
  team_b_id: string
  team_a_name: string
  team_b_name: string
  start_time: number
  total_rounds: number
  final_score_team_a: number
  final_score_team_b: number
  winner_team_id: string
}

export interface PlayerMatchStats {
  player_id: string
  player_name: string
  team_id: string
  side: Side
  kills: number
  deaths: number
  damage: number
  adr: number
  fk: number
  mk: number
  plants: number
  defuses: number
}

export interface FinalStats {
  score_t: number
  score_ct: number
  score_team_a: number
  score_team_b: number
  winner_team_id: string
  player_stats: PlayerMatchStats[]
}

export interface MatchReport {
	debug_enabled: boolean
	match_info: MatchInfo
  rounds: RoundReport[]
  final_stats: FinalStats
  winner: Side | ""
  winner_team_id: string
  final_score_team_a: number
  final_score_team_b: number
  total_rounds: number
}
