package matchengine

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
