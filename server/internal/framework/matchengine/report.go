package matchengine

// MatchResult 表示整场比赛完成后的权威结果和完整回合战报。
type MatchResult struct {
	MatchInfo       *MatchInfo         `json:"match_info"`         // 比赛元信息和最终摘要。
	Rounds          []*RoundResult     `json:"rounds"`             // 按回合号排序的完整回合结果。
	FinalStats      *FinalStats        `json:"final_stats"`        // 整场聚合统计。
	Winner          string             `json:"winner"`             // 获胜队伍在比赛结束时的阵营。
	WinnerTeamID    string             `json:"winner_team_id"`     // 获胜队伍唯一标识。
	FinalScoreTeamA int                `json:"final_score_team_a"` // Team A 最终得分。
	FinalScoreTeamB int                `json:"final_score_team_b"` // Team B 最终得分。
	TotalRounds     int                `json:"total_rounds"`       // 实际完成的总回合数。
	Report          *ExplainableReport `json:"report,omitempty"`
}

// MatchInfo 表示客户端展示和复现比赛所需的元信息。
type MatchInfo struct {
	MatchID         string `json:"match_id"`           // 本场比赛唯一标识。
	MapID           string `json:"map_id"`             // 地图配置 ID。
	MapName         string `json:"map_name"`           // 地图显示名称。
	MapVersion      string `json:"map_version"`        // 地图语义快照版本。
	RuleSetID       string `json:"rule_set_id"`        // 本场使用的规则集 ID。
	Seed            int64  `json:"seed"`               // 用于复现整场模拟的基础随机种子。
	TeamAID         string `json:"team_a_id"`          // Team A 唯一标识。
	TeamBID         string `json:"team_b_id"`          // Team B 唯一标识。
	TeamAName       string `json:"team_a_name"`        // Team A 显示名称。
	TeamBName       string `json:"team_b_name"`        // Team B 显示名称。
	StartTime       int64  `json:"start_time"`         // 比赛开始时间，Unix 毫秒。
	TotalRounds     int    `json:"total_rounds"`       // 实际完成的总回合数。
	FinalScoreTeamA int    `json:"final_score_team_a"` // Team A 最终得分。
	FinalScoreTeamB int    `json:"final_score_team_b"` // Team B 最终得分。
	WinnerTeamID    string `json:"winner_team_id"`     // 获胜队伍唯一标识。
}

// RoundResult 表示单个回合结束后的权威状态和事件列表。
type RoundResult struct {
	RoundNumber          int                 `json:"round_number"`                      // 全场连续回合号，从 1 开始。
	Phase                string              `json:"phase"`                             // 比赛阶段，例如 regulation 或 overtime。
	Half                 int                 `json:"half"`                              // 当前半场或加时阶段编号。
	OvertimeBlock        int                 `json:"overtime_block,omitempty"`          // 加时 block 编号，常规赛为零。
	OvertimeRoundInBlock int                 `json:"overtime_round_in_block,omitempty"` // 当前加时 block 内的回合序号。
	IsSideSwitch         bool                `json:"is_side_switch,omitempty"`          // 本回合开始前是否发生换边。
	Seed                 int64               `json:"seed"`                              // 本回合派生随机种子。
	SideAttacking        string              `json:"side_attacking"`                    // 当前进攻阵营，CS 规则下为 T。
	TeamTID              string              `json:"team_t_id"`                         // 当前作为 T 方的队伍 ID。
	TeamCTID             string              `json:"team_ct_id"`                        // 当前作为 CT 方的队伍 ID。
	Winner               string              `json:"winner"`                            // 本回合获胜阵营。
	WinnerTeamID         string              `json:"winner_team_id"`                    // 本回合获胜队伍 ID。
	WinReason            string              `json:"win_reason"`                        // 淘汰、超时、拆包或爆炸等终局原因。
	ScoreTeamA           int                 `json:"score_team_a"`                      // 本回合结束后的 Team A 总比分。
	ScoreTeamB           int                 `json:"score_team_b"`                      // 本回合结束后的 Team B 总比分。
	ScoreT               int                 `json:"score_t"`                           // 本回合结束后当前 T 方队伍的比分。
	ScoreCT              int                 `json:"score_ct"`                          // 本回合结束后当前 CT 方队伍的比分。
	RouteMain            string              `json:"route_main"`                        // 本回合选择的主进攻路线 ID。
	RouteSub             string              `json:"route_sub"`                         // 预留的辅助路线 ID。
	StrategyTemplateID   string              `json:"strategy_template_id"`              // 本回合选择的战术模板 ID。
	CTSetupTemplateID    string              `json:"ct_setup_template_id,omitempty"`
	Events               []*GameEvent        `json:"events"`                   // 按时间戳排序的公开事件列表。
	PlayerStates         []*PlayerState      `json:"player_states"`            // 回合结束时的选手公共状态。
	Bomb                 *BombPublicState    `json:"bomb,omitempty"`           // 回合结束时的炸弹公共状态。
	FinalControls        []*NodeControlState `json:"final_controls,omitempty"` // 回合结束时的关键节点控制权快照。
	Report               *ExplainableReport  `json:"report,omitempty"`
}

// Location 表示战报事件在地图雷达上的归一化位置。
type Location struct {
	Name       string  `json:"name"` // 点位显示名称。
	X          float64 `json:"x"`    // 雷达横轴归一化坐标，范围为 0 到 1。
	Y          float64 `json:"y"`    // 雷达纵轴归一化坐标，范围为 0 到 1。
	SourceType string  `json:"source_type,omitempty"`
	SourceID   string  `json:"source_id,omitempty"`
	Floor      string  `json:"floor,omitempty"`
	Seed       int64   `json:"seed,omitempty"`
}

type ReasonModifier struct {
	Code   string  `json:"code"`
	Value  float64 `json:"value"`
	Detail string  `json:"detail,omitempty"`
}

type ReasonValue struct {
	Kind   string   `json:"kind"`
	Number *float64 `json:"number,omitempty"`
	String *string  `json:"string,omitempty"`
	Bool   *bool    `json:"bool,omitempty"`
}

type ReasonStateChange struct {
	Field  string      `json:"field"`
	Before ReasonValue `json:"before"`
	After  ReasonValue `json:"after"`
}

// EventReason 描述事件产生的可解释原因和评分影响。
type EventReason struct {
	Code           string              `json:"code"`        // 稳定的原因代码。
	MainFactor     string              `json:"main_factor"` // 对该事件影响最大的阶段、属性或规则。
	Modifiers      []ReasonModifier    `json:"modifiers,omitempty"`
	ScoreDelta     float64             `json:"score_delta"` // 相关评分差值，用于调试和解释。
	Probability    *float64            `json:"probability,omitempty"`
	Formula        string              `json:"formula,omitempty"`
	Inputs         map[string]float64  `json:"inputs,omitempty"`
	StateChanges   []ReasonStateChange `json:"state_changes,omitempty"`
	SourceActionID string              `json:"source_action_id,omitempty"`
	SourceEffectID string              `json:"source_effect_id,omitempty"`
	Detail         string              `json:"detail,omitempty"` // 可选的补充说明。
}

type EventStateSnapshot struct {
	ScoreTeamA int                 `json:"score_team_a"`
	ScoreTeamB int                 `json:"score_team_b"`
	ScoreT     int                 `json:"score_t"`
	ScoreCT    int                 `json:"score_ct"`
	Players    []*PlayerState      `json:"players,omitempty"`
	Bomb       *BombPublicState    `json:"bomb,omitempty"`
	Controls   []*NodeControlState `json:"controls,omitempty"`
}

type ExplainableReport struct {
	KeyEvents       []*GameEvent   `json:"key_events"`
	StrategySummary string         `json:"strategy_summary"`
	LossReasons     []*EventReason `json:"loss_reasons"`
	WinFactors      []*EventReason `json:"win_factors"`
}

// GameEvent 表示客户端可见的一条离散比赛事件。
type GameEvent struct {
	EventID        string                 `json:"event_id,omitempty"`         // 稳定事件 ID。
	SourceActionID string                 `json:"source_action_id,omitempty"` // 产生事件的真实行动 ID。
	SourceEffectID string                 `json:"source_effect_id,omitempty"` // 产生事件的已应用效果 ID。
	Timestamp      int64                  `json:"timestamp"`                  // 回合内事件时间，单位为秒。
	EventType      string                 `json:"event_type"`                 // 事件类型常量，例如 KILL 或 BOMB_PLANT。
	AttackerID     string                 `json:"attacker_id,omitempty"`      // 进攻行为发起者的选手 ID。
	AttackerName   string                 `json:"attacker_name,omitempty"`    // 进攻行为发起者的显示名称。
	AttackerTeamID string                 `json:"attacker_team_id,omitempty"` // 进攻行为发起者所属队伍 ID。
	VictimID       string                 `json:"victim_id,omitempty"`        // 受击或死亡选手 ID。
	VictimName     string                 `json:"victim_name,omitempty"`      // 受击或死亡选手显示名称。
	VictimTeamID   string                 `json:"victim_team_id,omitempty"`   // 受击或死亡选手所属队伍 ID。
	Weapon         string                 `json:"weapon,omitempty"`           // 事件使用的武器显示名称或配置 ID。
	Location       *Location              `json:"location,omitempty"`         // 事件发生的公开地图位置。
	IsFirstKill    bool                   `json:"is_first_kill,omitempty"`    // 是否为本回合首杀。
	IsTrade        bool                   `json:"is_trade,omitempty"`         // 是否为短时间内发生的补枪。
	Message        string                 `json:"message"`                    // 面向客户端直接展示的战报文本。
	Reason         *EventReason           `json:"reason,omitempty"`           // 事件的可解释原因。
	Bomb           *BombPublicState       `json:"bomb,omitempty"`             // 事件发生后的炸弹状态快照。
	State          *EventStateSnapshot    `json:"state,omitempty"`
	ScoreTeamA     int                    `json:"score_team_a,omitempty"` // 事件发生后的 Team A 比分。
	ScoreTeamB     int                    `json:"score_team_b,omitempty"` // 事件发生后的 Team B 比分。
	Extra          map[string]interface{} `json:"extra,omitempty"`        // 事件类型特有的扩展公开数据。
	sortPriority   int                    `json:"-"`
	sortActionType string                 `json:"-"`
	sortMinActorID string                 `json:"-"`
}

// PlayerState 表示单个选手在回合结束时的公开状态。
type PlayerState struct {
	PlayerID       string        `json:"player_id"`              // 整场比赛内唯一且不透明的实例标识。
	ConfigPlayerID string        `json:"config_player_id"`       // 对应 TbPlayer.id 的策划选手标识。
	PlayerName     string        `json:"player_name"`            // 兼容旧战报消费者的选手名称。
	DisplayName    string        `json:"display_name"`           // 客户端优先使用的选手显示名称。
	Portrait       string        `json:"portrait,omitempty"`     // 头像资源路径或 URL。
	CardImage      string        `json:"card_image,omitempty"`   // 完整选手卡面资源路径或 URL。
	AvatarCrop     *ImageCrop    `json:"avatar_crop,omitempty"`  // 归一化 5:7 头像裁切矩形。
	TeamID         string        `json:"team_id"`                // 本回合所属队伍 ID。
	Side           string        `json:"side"`                   // 本回合所属 T/CT 阵营。
	IsAlive        bool          `json:"is_alive"`               // 兼容旧战报消费者的存活标记。
	Alive          bool          `json:"alive"`                  // 当前标准存活标记。
	HP             int           `json:"hp"`                     // 回合结束时生命值。
	Stamina        int           `json:"stamina"`                // 回合结束时体力值。
	Focus          int           `json:"focus"`                  // 回合结束时专注值。
	CurrentNode    string        `json:"current_node,omitempty"` // 当前所在地图节点 ID。
	HasBomb        bool          `json:"has_bomb,omitempty"`     // 是否携带炸弹。
	Kills          int           `json:"kills"`                  // 本回合击杀数。
	Deaths         int           `json:"deaths"`                 // 本回合死亡数。
	Damage         int           `json:"damage"`                 // 本回合造成的总伤害。
	RoleTags       []string      `json:"role_tags,omitempty"`    // 选手角色标签快照。
	Weapon         WeaponLoadout `json:"weapon"`                 // 本回合按阵营派生的装备。
}

// BombPublicState 表示某一事件或回合结束时的公开炸弹状态。
type BombPublicState struct {
	Status    string `json:"status"`               // 携带、掉落、已下包、已拆除或已爆炸。
	CarrierID string `json:"carrier_id,omitempty"` // 当前或最近一次炸弹携带者的选手 ID。
	NodeID    string `json:"node_id,omitempty"`    // 炸弹当前所在地图节点 ID。
	Site      string `json:"site,omitempty"`       // 下包所在包点，例如 A 或 B。
	PlantedAt int    `json:"planted_at,omitempty"` // 下包完成的回合内秒数。
	ExplodeAt int    `json:"explode_at,omitempty"` // 预计爆炸的回合内秒数。
	DroppedAt int    `json:"dropped_at,omitempty"` // 炸弹掉落的回合内秒数。
}

// NodeControlState 表示回合结束时某个地图节点的公开控制权。
type NodeControlState struct {
	NodeID    string `json:"node_id"`     // 地图节点 ID。
	Status    string `json:"status"`      // T 控制、CT 控制或争夺中等状态。
	KnownByT  bool   `json:"known_by_t"`  // T 方是否已知该控制权信息。
	KnownByCT bool   `json:"known_by_ct"` // CT 方是否已知该控制权信息。
	UpdatedAt int    `json:"updated_at"`  // 控制权最后更新时间，单位为回合内秒数。
}

// PlayerMatchStats 表示单个选手的整场聚合统计。
type PlayerMatchStats struct {
	PlayerID       string  `json:"player_id"`        // 整场比赛内唯一且不透明的实例标识。
	ConfigPlayerID string  `json:"config_player_id"` // 对应 TbPlayer.id 的策划选手标识。
	PlayerName     string  `json:"player_name"`      // 选手显示名称。
	TeamID         string  `json:"team_id"`          // 选手所属队伍 ID。
	Side           string  `json:"side"`             // 选手开局时的初始阵营，仅作摘要展示。
	Kills          int     `json:"kills"`            // 整场击杀数。
	Deaths         int     `json:"deaths"`           // 整场死亡数。
	Assists        int     `json:"assists"`          // 整场助攻数。
	Damage         int     `json:"damage"`           // 整场造成的总伤害。
	ADR            float64 `json:"adr"`              // 场均回合伤害，即总伤害除以总回合数。
	FK             int     `json:"fk"`               // 整场首杀次数。
	MK             int     `json:"mk"`               // 整场多杀回合次数。
	Plants         int     `json:"plants"`           // 整场下包次数。
	Defuses        int     `json:"defuses"`          // 整场拆包次数。
}

// FinalStats 表示比赛结束后的比分与选手聚合统计。
type FinalStats struct {
	ScoreT       int                 `json:"score_t"`        // 比赛结束时当前 T 方队伍的比分。
	ScoreCT      int                 `json:"score_ct"`       // 比赛结束时当前 CT 方队伍的比分。
	ScoreTeamA   int                 `json:"score_team_a"`   // Team A 最终得分。
	ScoreTeamB   int                 `json:"score_team_b"`   // Team B 最终得分。
	WinnerTeamID string              `json:"winner_team_id"` // 获胜队伍唯一标识。
	PlayerStats  []*PlayerMatchStats `json:"player_stats"`   // 全部选手的整场聚合统计。
}
