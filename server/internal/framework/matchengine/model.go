// Package matchengine 提供纯净的模拟比赛推演能力。
// 它不依赖任何业务子系统，只接收 MatchInput 并输出 MatchResult。
// 业务子系统（如 internal/match）负责把 Nakama RPC 请求转换为 MatchInput。
package matchengine

// Team 表示一支队伍（T 或 CT）。
type Team struct {
	ID      string
	Name    string
	Side    string // "T" or "CT"
	Players []*Combatant
}

// Combatant 是单名选手在一场推演中的运行时状态。
type Combatant struct {
	PlayerID   string
	Name       string
	Side       string // "T" or "CT"
	Entry      int32
	Aim        int32
	Trade      int32
	Clutch     int32
	Firepower  int32
	Gamesense  int32
	Alive      bool
	Kills      int
	Deaths     int
	Damage     int // 本回合造成伤害
	FirstKills int
	MultiKills int // 一回合 3 杀以上次数
}

// MatchInput 是引擎的推演输入。
type MatchInput struct {
	MatchID string
	TeamA   *Team // 默认 T 方
	TeamB   *Team // 默认 CT 方
	MapID   string
	Seed    int64
}

// MatchResult 是整场比赛的推演结果。
type MatchResult struct {
	MatchInfo   *MatchInfo     `json:"match_info"`
	Rounds      []*RoundResult `json:"rounds"`
	FinalStats  *FinalStats    `json:"final_stats"`
	Winner      string         `json:"winner"`
	TotalRounds int            `json:"total_rounds"`
}

// MatchInfo 记录比赛元信息。
type MatchInfo struct {
	MatchID     string `json:"match_id"`
	MapID       string `json:"map_id"`
	MapName     string `json:"map_name"`
	TeamAName   string `json:"team_a_name"`
	TeamBName   string `json:"team_b_name"`
	StartTime   int64  `json:"start_time"`
	TotalRounds int    `json:"total_rounds"`
}

// RoundResult 记录单回合战报。
type RoundResult struct {
	RoundNumber   int            `json:"round_number"`
	SideAttacking string         `json:"side_attacking"`
	Winner        string         `json:"winner"`
	WinReason     string         `json:"win_reason"`
	ScoreT        int            `json:"score_t"`
	ScoreCT       int            `json:"score_ct"`
	RouteMain     string         `json:"route_main"`
	RouteSub      string         `json:"route_sub"`
	Events        []*GameEvent   `json:"events"`
	PlayerStates  []*PlayerState `json:"player_states"`
}

// Location 表示事件在地图上的位置。
// X、Y 为相对于地图宽高的比例（0.0 ~ 1.0），便于前端在不同尺寸下定位。
type Location struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// GameEvent 是 HLTV 风格播报的原始事件。
type GameEvent struct {
	Timestamp    int64                  `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	AttackerID   string                 `json:"attacker_id,omitempty"`
	AttackerName string                 `json:"attacker_name,omitempty"`
	VictimID     string                 `json:"victim_id,omitempty"`
	VictimName   string                 `json:"victim_name,omitempty"`
	Weapon       string                 `json:"weapon,omitempty"`
	Location     *Location              `json:"location,omitempty"`
	IsFirstKill  bool                   `json:"is_first_kill,omitempty"`
	IsTrade      bool                   `json:"is_trade,omitempty"`
	Message      string                 `json:"message"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// PlayerState 记录回合结束时选手状态。
type PlayerState struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Side       string `json:"side"`
	IsAlive    bool   `json:"is_alive"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	Damage     int    `json:"damage"`
}

// PlayerMatchStats 记录选手整场比赛统计。
type PlayerMatchStats struct {
	PlayerID   string  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Side       string  `json:"side"`
	Kills      int     `json:"kills"`
	Deaths     int     `json:"deaths"`
	ADR        float64 `json:"adr"`
	FK         int     `json:"fk"`
	MK         int     `json:"mk"`
}

// FinalStats 记录最终队伍和选手统计。
type FinalStats struct {
	ScoreT      int                 `json:"score_t"`
	ScoreCT     int                 `json:"score_ct"`
	PlayerStats []*PlayerMatchStats `json:"player_stats"`
}

// RouteConfig 表示一条进攻路线配置。
type RouteConfig struct {
	ID         string
	Name       string
	TargetSite string // "A" / "B"
	BaseTime   int    // 基础推进时间（秒）
	MinPlayers int
	MaxPlayers int
}

// MatchState 是引擎内部状态。
type MatchState struct {
	ScoreT int
	ScoreCT int
}
