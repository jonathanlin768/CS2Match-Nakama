// Package matchengine 提供纯净的模拟比赛推演能力。
// 它不依赖任何业务子系统，只接收 MatchInput 并输出 MatchResult。
// 业务子系统（如 internal/match）负责把 Nakama RPC 请求转换为 MatchInput。
package matchengine

import "math"

const (
	SideT  = "T"
	SideCT = "CT"

	EventMatchStart       = "MATCH_START"
	EventRoundStart       = "ROUND_START"
	EventHalfTime         = "HALF_TIME"
	EventSideSwitch       = "SIDE_SWITCH"
	EventOvertime         = "OVERTIME_START"
	EventDamage           = "DAMAGE"
	EventKill             = "KILL"
	EventStrategyAdjusted = "STRATEGY_ADJUSTED"
	EventRotate           = "ROTATE"
	EventReinforce        = "REINFORCE"
	EventControlGained    = "CONTROL_GAINED"
	EventBombDrop         = "BOMB_DROP"
	EventBombPickup       = "BOMB_PICKUP"
	EventPlantStart       = "BOMB_PLANT_START"
	EventPlantInterrupt   = "BOMB_PLANT_INTERRUPT"
	EventDefuseStart      = "DEFUSE_START"
	EventDefuseInterrupt  = "DEFUSE_INTERRUPT"
	EventBombPlant        = "BOMB_PLANT"
	EventBombDefuse       = "BOMB_DEFUSE"
	EventBombExplode      = "BOMB_EXPLODE"
	EventRoundEnd         = "ROUND_END"
	EventMatchEnd         = "MATCH_END"
	BombStatusCarried     = "Carried"
	BombStatusPlanted     = "Planted"
	BombStatusDefused     = "Defused"
	BombStatusExplode     = "Exploded"
	BombStatusDropped     = "Dropped"
)

// EngineError 是引擎层返回的结构化错误。
type EngineError struct {
	Code    string `json:"code"`    // 稳定的错误码，供调用方分类处理。
	Message string `json:"message"` // 面向日志或调用方的错误说明。
}

func (e *EngineError) Error() string {
	return e.Code + ": " + e.Message
}

// TeamInput 是传入引擎的自包含队伍快照。
type TeamInput struct {
	TeamID   string          `json:"team_id"`             // 队伍唯一标识。
	Name     string          `json:"name"`                // 队伍显示名称。
	Players  []PlayerProfile `json:"players"`             // 本场比赛使用的选手档案快照。
	TeamTags []string        `json:"team_tags,omitempty"` // 队伍风格、地区等可选标签。
}

// PlayerProfile 表示选手静态档案。武器不属于该档案，需按回合阵营派生。
type PlayerProfile struct {
	PlayerID       string           `json:"player_id"`             // 整场比赛内唯一且不透明的实例标识。
	ConfigPlayerID string           `json:"config_player_id"`      // 对应 TbPlayer.id 的策划选手标识。
	DisplayName    string           `json:"display_name"`          // 战报和客户端使用的显示名称。
	Portrait       string           `json:"portrait,omitempty"`    // 头像资源路径或 URL。
	CardImage      string           `json:"card_image,omitempty"`  // 完整选手卡面资源路径或 URL。
	AvatarCrop     *ImageCrop       `json:"avatar_crop,omitempty"` // 从完整卡面裁切战斗头像的归一化矩形。
	RoleTags       []string         `json:"role_tags,omitempty"`   // 选手位置、职责等角色标签。
	Attributes     PlayerAttributes `json:"attributes"`            // 参与模拟计算的静态能力值。
}

// ImageCrop 表示基于原图自然尺寸的归一化裁切矩形。完整卡面固定为 2:3，裁切输出固定为 5:7。
type ImageCrop struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Valid 判断裁切矩形是否有界，并在 2:3 源图上保持 5:7 像素比例。
func (c ImageCrop) Valid() bool {
	values := []float64{c.X, c.Y, c.Width, c.Height}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	if c.X < 0 || c.Y < 0 || c.Width <= 0 || c.Height <= 0 || c.X+c.Width > 1.000001 || c.Y+c.Height > 1.000001 {
		return false
	}
	pixelAspect := (c.Width * 2) / (c.Height * 3)
	return math.Abs(pixelAspect-5.0/7.0) <= 0.005
}

// PlayerAttributes 表示选手的静态能力评分，默认有效范围由战斗常量约束。
type PlayerAttributes struct {
	Entry       int `json:"entry"`       // 突破与首轮交火能力。
	Aim         int `json:"aim"`         // 瞄准与枪法精度。
	Trade       int `json:"trade"`       // 补枪和协同换人能力。
	Clutch      int `json:"clutch"`      // 残局处理能力。
	Firepower   int `json:"firepower"`   // 正面火力输出能力。
	Gamesense   int `json:"gamesense"`   // 局势阅读与决策意识。
	Reaction    int `json:"reaction"`    // 遭遇目标后的反应速度。
	Positioning int `json:"positioning"` // 站位和空间利用能力。
	Awareness   int `json:"awareness"`   // 对敌情、声音和队友信息的感知能力。
	Teamplay    int `json:"teamplay"`    // 团队配合和战术执行能力。
	Utility     int `json:"utility"`     // 投掷物使用能力。
	Composure   int `json:"composure"`   // 高压场景下的稳定性。
	Mobility    int `json:"mobility"`    // 移动与转点能力。
	Endurance   int `json:"endurance"`   // 持续作战能力。
	Discipline  int `json:"discipline"`  // 战术纪律和风险控制能力。
}

// WeaponLoadout 表示选手在某一回合内按阵营派生的装备快照。
type WeaponLoadout struct {
	Primary   string   `json:"primary"`             // 主武器配置 ID。
	Secondary string   `json:"secondary,omitempty"` // 副武器配置 ID。
	Armor     bool     `json:"armor"`               // 是否装备护甲。
	Helmet    bool     `json:"helmet"`              // 是否装备头盔。
	HasKit    bool     `json:"has_kit,omitempty"`   // 是否携带拆弹器，仅 CT 有效。
	Grenades  []string `json:"grenades,omitempty"`  // 携带的投掷物配置 ID 列表。
}

// WeaponSpec 表示由调用方传入的武器数值快照。
type WeaponSpec struct {
	ID               string  `json:"id"`                // 武器唯一标识。
	DisplayName      string  `json:"display_name"`      // 战报使用的武器名称。
	Damage           int     `json:"damage"`            // 单发基础伤害。
	RoundsPerMinute  int     `json:"rounds_per_minute"` // 理论射速，单位为发/分钟。
	MagazineSize     int     `json:"magazine_size"`     // 单个弹匣容量。
	ArmorPenetration float64 `json:"armor_penetration"` // 护甲穿透比例，范围为 0 到 1。
	RangeModifier    float64 `json:"range_modifier"`    // 距离伤害修正系数。
}

// RuleSet 表示完整比赛规则快照。
type RuleSet struct {
	RuleSetID            string `json:"rule_set_id"`            // 规则集唯一标识。
	RegulationHalfRounds int    `json:"regulation_half_rounds"` // 常规赛每半场回合数。
	RegulationWinRounds  int    `json:"regulation_win_rounds"`  // 常规赛获胜所需队伍分数。
	RegulationMaxRounds  int    `json:"regulation_max_rounds"`  // 常规赛最多回合数。
	OvertimeEnabled      bool   `json:"overtime_enabled"`       // 常规赛平局时是否进入加时。
	OvertimeHalfRounds   int    `json:"overtime_half_rounds"`   // 每个加时半段的回合数。
	OvertimeBlockRounds  int    `json:"overtime_block_rounds"`  // 每个完整加时 block 的回合数。
}

// MatchInput 是引擎的整场推演输入。
type MatchInput struct {
	MatchID           string                   `json:"match_id"`             // 本场比赛唯一标识。
	MapID             string                   `json:"map_id"`               // 地图配置 ID。
	MapName           string                   `json:"map_name"`             // 地图显示名称。
	MapVersion        string                   `json:"map_version"`          // 地图语义快照版本，用于确定性 seed 派生。
	Seed              int64                    `json:"seed"`                 // 整场比赛的基础随机种子，必须非零。
	RuleSet           RuleSet                  `json:"rule_set"`             // 本场比赛使用的完整规则快照。
	TeamA             TeamInput                `json:"team_a"`               // Team A 的自包含输入快照。
	TeamB             TeamInput                `json:"team_b"`               // Team B 的自包含输入快照。
	InitialSideByTeam map[string]string        `json:"initial_side_by_team"` // 队伍 ID 到初始 T/CT 阵营的映射。
	MapConfig         *MapConfig               `json:"map_config"`           // 调用方构造并传入的地图语义快照。
	WeaponSpecs       map[string]WeaponSpec    `json:"weapon_specs"`         // 武器 ID 到武器数值的映射。
	SideLoadouts      map[string]WeaponLoadout `json:"side_loadouts"`        // T/CT 阵营到默认回合装备的映射。

	StartTime int64 `json:"start_time,omitempty"` // 比赛开始时间，Unix 毫秒；为零时由引擎生成。
}

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

// MapConfig 表示调用方构造的地图语义快照，matchengine 不直接读取 Luban 或业务配置。
type MapConfig struct {
	MapID              string                       `json:"map_id"`              // 地图唯一标识。
	MapName            string                       `json:"map_name"`            // 地图显示名称。
	Version            string                       `json:"version"`             // 地图语义快照版本。
	RouteTemplates     map[string]RouteTemplate     `json:"route_templates"`     // 战术模板 ID 到模板数据的映射。
	Scenarios          map[string]Scenario          `json:"scenarios"`           // 场景 ID 到场景数据的映射。
	MapTags            map[string]MapTag            `json:"map_tags"`            // 地图标签 ID 到标签数据的映射。
	EncounterModifiers map[string]EncounterModifier `json:"encounter_modifiers"` // 遭遇修正 ID 到修正数据的映射。
	Nodes              map[string]MapNode           `json:"nodes"`               // 地图节点 ID 到节点数据的映射。
	Edges              map[string]MapEdge           `json:"edges"`               // 地图路径边 ID 到边数据的映射。
	Visibility         map[string]Visibility        `json:"visibility"`          // 视野关系 ID 到视野数据的映射。
	Routes             map[string]Route             `json:"routes"`              // 实际路线 ID 到路线数据的映射。
	CombatConstants    CombatConstants              `json:"combat_constants"`    // 引擎运行所需的战斗常量快照。
	Warnings           []EngineError                `json:"warnings,omitempty"`  // 允许降级的配置诊断，例如 KillSample 几何回退。
}

// RouteTemplate 表示可供回合决策选择的战术路线模板。
type RouteTemplate struct {
	ID               string             `json:"id"`                            // 战术模板唯一标识。
	MapID            string             `json:"map_id"`                        // 所属地图 ID。
	Side             string             `json:"side"`                          // 模板所属阵营，T 或 CT。
	TargetSite       string             `json:"target_site"`                   // 目标包点或 None。
	Tempo            string             `json:"tempo"`                         // 战术节奏，例如 Fast、Default 或 Slow。
	RecommendedMin   int                `json:"recommended_min"`               // 推荐参与人数下限。
	RecommendedMax   int                `json:"recommended_max"`               // 推荐参与人数上限。
	RequiredRoles    []string           `json:"required_roles,omitempty"`      // 战术要求的选手角色。
	KeyAttributes    map[string]float64 `json:"key_attributes,omitempty"`      // 关键能力名称及其评分权重。
	RouteIDs         []string           `json:"route_ids,omitempty"`           // 模板允许使用的语义路线 ID。
	RouteAllocations map[string]int     `json:"route_allocations,omitempty"`   // 路线 ID 到分配人数的闭合映射。
	ScenarioIDs      []string           `json:"scenario_ids,omitempty"`        // 可进入的场景 ID 列表。
	MapTagIDs        []string           `json:"map_tag_ids,omitempty"`         // 参与战术评分的地图标签 ID。
	CommonCTSetupIDs []string           `json:"common_ct_setup_ids,omitempty"` // T 模板可见的常见 CT 配置先验。
	SuccessNextPhase string             `json:"success_next_phase"`            // 执行成功后的下一阶段名称。
	FailureFallbacks []string           `json:"failure_fallbacks,omitempty"`   // 执行失败时可选择的后备模板 ID。
}

// Scenario 表示路线模板内的一种具体战斗场景。
type Scenario struct {
	ID             string   `json:"id"`                    // 场景唯一标识。
	Route          string   `json:"route"`                 // 场景关联的路线或路线类型。
	Phase          string   `json:"phase"`                 // 场景所处的回合阶段。
	Range          string   `json:"range"`                 // 主要交战距离。
	Site           string   `json:"site"`                  // 关联包点或区域。
	Tempo          string   `json:"tempo"`                 // 场景节奏。
	Posture        string   `json:"posture"`               // 队伍姿态，例如进攻、架枪或回防。
	UtilityContext string   `json:"utility_context"`       // 投掷物使用上下文。
	MapTagIDs      []string `json:"map_tag_ids,omitempty"` // 影响该场景的地图标签 ID。
	BaseTimeCost   int      `json:"base_time_cost"`        // 场景基础耗时，单位为秒。
	BaseWeight     int      `json:"base_weight"`           // 场景被选择时的基础权重。
}

// MapTag 表示参与决策和评分的地图语义标签。
type MapTag struct {
	ID          string `json:"id"`                    // 地图标签唯一标识。
	MapID       string `json:"map_id"`                // 所属地图 ID。
	Category    string `json:"category"`              // 标签分类。
	Value       string `json:"value"`                 // 标签的语义值。
	Side        string `json:"side"`                  // 标签影响的阵营，空值表示双方。
	Weight      int    `json:"weight"`                // 标签参与评分时的权重。
	ReasonCode  string `json:"reason_code"`           // 用于战报解释的原因代码。
	Description string `json:"description,omitempty"` // 面向配置人员的说明。
}

// EncounterModifier 表示特定场景下对阵营或属性的遭遇战修正。
type EncounterModifier struct {
	ID         string `json:"id"`          // 遭遇修正唯一标识。
	ScenarioID string `json:"scenario_id"` // 关联的场景 ID。
	Factor     string `json:"factor"`      // 修正因素名称。
	Side       string `json:"side"`        // 受到修正的阵营。
	Attribute  string `json:"attribute"`   // 受到修正的选手属性。
	Weight     int    `json:"weight"`      // 加入遭遇评分的修正权重。
	ReasonCode string `json:"reason_code"` // 用于战报解释的原因代码。
}

// MapNode 表示地图语义图中的一个区域或关键点位。
type MapNode struct {
	ID          string   `json:"id"`                    // 地图节点唯一标识。
	MapID       string   `json:"map_id"`                // 所属地图 ID。
	Name        string   `json:"name"`                  // 点位显示名称。
	Zone        string   `json:"zone"`                  // 所属地图分区。
	Site        string   `json:"site"`                  // 关联包点，例如 A、B 或 None。
	NodeType    string   `json:"node_type"`             // 节点类型，例如 Spawn、Lane 或 Site。
	DefaultSide string   `json:"default_side"`          // 默认占优或出生阵营。
	X           float64  `json:"x"`                     // 雷达横轴归一化坐标，范围为 0 到 1。
	Y           float64  `json:"y"`                     // 雷达纵轴归一化坐标，范围为 0 到 1。
	Floor       string   `json:"floor"`                 // 楼层或高度层级标识。
	AreaUsages  []string `json:"area_usages,omitempty"` // 下包、击杀采样等区域用途。
	Shape       string   `json:"shape"`                 // 采样区域形状，例如 None、Circle 或 Polygon。
	Radius      float64  `json:"radius,omitempty"`      // 圆形区域半径，使用雷达归一化单位。
	Points      string   `json:"points,omitempty"`      // 多边形顶点串，格式由地图配置约定。
}

// MapEdge 表示地图语义图中两个节点之间的可通行路径。
type MapEdge struct {
	ID             string   `json:"id"`                        // 路径边唯一标识。
	FromNode       string   `json:"from_node"`                 // 起点节点 ID。
	ToNode         string   `json:"to_node"`                   // 终点节点 ID。
	BaseTime       int      `json:"base_time"`                 // 通过该路径的基础耗时，单位为秒。
	StaminaCost    int      `json:"stamina_cost"`              // 通过该路径消耗的体力值。
	Risk           int      `json:"risk"`                      // 路径基础风险评分。
	Noise          int      `json:"noise"`                     // 移动产生的基础噪声评分。
	RiskPoints     []string `json:"risk_points,omitempty"`     // 路径关联的风险点节点 ID。
	InterceptNodes []string `json:"intercept_nodes,omitempty"` // 对手可能拦截该路径的节点 ID。
	Bidirectional  bool     `json:"bidirectional"`             // 是否允许双向通行。
}

// Visibility 表示两个地图节点之间的可见性和交火条件。
type Visibility struct {
	ID               string `json:"id"`                  // 视野关系唯一标识。
	FromNode         string `json:"from_node"`           // 观察方所在节点 ID。
	ToNode           string `json:"to_node"`             // 被观察方所在节点 ID。
	Visible          bool   `json:"visible"`             // 两节点之间是否存在直接视线。
	Range            string `json:"range"`               // 典型交战距离。
	AngleAdvantage   string `json:"angle_advantage"`     // 架枪或探点的角度优势描述。
	Elevation        string `json:"elevation,omitempty"` // 两点之间的高低差描述。
	CoverModifier    int    `json:"cover_modifier"`      // 掩体对观察或交火评分的修正值。
	ExposureModifier int    `json:"exposure_modifier"`   // 暴露程度对交火评分的修正值。
}

// Route 表示地图语义图上可实际执行的一条有序路线。
type Route struct {
	ID         string   `json:"id"`                   // 路线唯一标识。
	Name       string   `json:"name"`                 // 路线显示名称。
	Side       string   `json:"side"`                 // 可执行该路线的阵营。
	TargetSite string   `json:"target_site"`          // 路线目标包点或 None。
	Nodes      []string `json:"nodes"`                // 按移动顺序排列的地图节点 ID。
	MinPlayers int      `json:"min_players"`          // 执行路线所需的最少人数。
	MaxPlayers int      `json:"max_players"`          // 执行路线允许的最多人数。
	StyleTags  []string `json:"style_tags,omitempty"` // 快攻、默认控图等路线风格标签。
}

// CombatConstValue 表示一项由调用方配表转换得到的原始战斗常量。
type CombatConstValue struct {
	Key       string `json:"key"`                 // 常量唯一键。
	Category  string `json:"category"`            // 常量所属分类。
	ValueType string `json:"value_type"`          // 原始值类型，例如 Int、Float 或 Bool。
	Value     string `json:"value"`               // 等待类型化访问器解析的原始值。
	MinValue  string `json:"min_value,omitempty"` // 配置允许的最小值。
	MaxValue  string `json:"max_value,omitempty"` // 配置允许的最大值。
	Unit      string `json:"unit"`                // 秒、评分、比例等单位。
}

// CombatConstants 保存以常量键索引的完整战斗常量快照。
type CombatConstants struct {
	Values map[string]CombatConstValue `json:"values"` // 常量键到原始常量数据的映射。
}
