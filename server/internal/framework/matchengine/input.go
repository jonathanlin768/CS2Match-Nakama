package matchengine

import "math"

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
