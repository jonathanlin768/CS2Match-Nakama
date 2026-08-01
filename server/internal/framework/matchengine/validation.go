package matchengine

import (
	"fmt"
	"strconv"
)

var requiredCombatConstKeys = []string{
	"RoundTimeLimit",
	"BombExplodeTime",
	"BasePlantTime",
	"BaseDefuseTime",
	"BasePickupTime",
	"ForceExecuteThreshold",
	"MaxDecisionCount",
	"MaxEncounterPulses",
	"CombatScale",
	"BaseNoise",
	"MaxRandomNoise",
	"CloseScoreGap",
	"DecisiveScoreGap",
	"StrategyRepeatWindow",
	"RepeatFreeCount",
	"StrategyRepeatPenalty",
	"SuccessBonusPerRound",
	"MaxPreviousSuccessBonus",
	"CounterMemoryBonus",
	"MinStrategyWeight",
	"MaxStrategyWeight",
	"ControlIntelTTL",
	"MinAttribute",
	"MaxAttribute",
	"MinHP",
	"MaxHP",
	"MinStamina",
	"MaxStamina",
	"MinFocus",
	"MaxFocus",
	"MinHitChance",
	"MaxHitChance",
	"MaxKillChance",
	"MinPlantTime",
	"MaxPlantTime",
	"MinDefuseTime",
	"MaxDefuseTime",
	"MinMoveTime",
	"MaxMoveTime",
}

func newError(code, format string, args ...interface{}) *EngineError {
	return &EngineError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func ValidateRuleSet(rule RuleSet) error {
	if rule.RuleSetID == "" {
		return newError("INVALID_RULE_SET", "missing rule_set_id")
	}
	if rule.RegulationHalfRounds <= 0 || rule.RegulationWinRounds <= 0 || rule.RegulationMaxRounds <= 0 {
		return newError("INVALID_RULE_SET", "regulation rounds must be positive")
	}
	if rule.RegulationMaxRounds != rule.RegulationHalfRounds*2 {
		return newError("INVALID_RULE_SET", "regulation max rounds must equal two halves")
	}
	if rule.RegulationWinRounds <= rule.RegulationHalfRounds {
		return newError("INVALID_RULE_SET", "win rounds must exceed one half")
	}
	if rule.RegulationWinRounds > rule.RegulationMaxRounds {
		return newError("INVALID_RULE_SET", "win rounds cannot exceed regulation max rounds")
	}
	if rule.OvertimeEnabled {
		if rule.OvertimeHalfRounds <= 0 || rule.OvertimeBlockRounds <= 0 {
			return newError("INVALID_RULE_SET", "overtime rounds must be positive")
		}
		if rule.OvertimeBlockRounds != rule.OvertimeHalfRounds*2 {
			return newError("INVALID_RULE_SET", "overtime block must equal two overtime halves")
		}
	}
	if rule.RoundTimeLimit <= 0 || rule.BombExplodeTime <= 0 || rule.BasePlantTime <= 0 || rule.BaseDefuseTime <= 0 {
		return newError("INVALID_RULE_SET", "timing constants must be positive")
	}
	if rule.MaxDecisionCount <= 0 || rule.MaxEncounterPulses <= 0 {
		return newError("INVALID_RULE_SET", "decision and encounter limits must be positive")
	}
	return nil
}

func ValidateMapConfig(cfg *MapConfig) error {
	if cfg == nil {
		return newError("CONFIG_MISSING_MAP", "missing map config")
	}
	if cfg.MapID == "" {
		return newError("CONFIG_MISSING_MAP", "missing map id")
	}
	if len(cfg.RouteTemplates) == 0 {
		return newError("CONFIG_MISSING_ROUTE_TEMPLATE", "map %s has no route templates", cfg.MapID)
	}
	if len(cfg.Routes) == 0 {
		return newError("CONFIG_MISSING_ROUTE", "map %s has no routes", cfg.MapID)
	}
	if len(cfg.Nodes) == 0 {
		return newError("CONFIG_MISSING_NODE", "map %s has no map nodes", cfg.MapID)
	}
	if len(cfg.Scenarios) == 0 {
		return newError("CONFIG_MISSING_SCENARIO", "map %s has no scenarios", cfg.MapID)
	}
	if len(cfg.MapTags) == 0 {
		return newError("CONFIG_MISSING_MAP_TAG", "map %s has no map tags", cfg.MapID)
	}
	for _, key := range requiredCombatConstKeys {
		if _, ok := cfg.CombatConstants.Values[key]; !ok {
			return newError("CONFIG_MISSING_COMBAT_CONST", "missing combat const: %s", key)
		}
	}
	for id, node := range cfg.Nodes {
		if id == "" || node.ID == "" {
			return newError("CONFIG_BAD_NODE", "node id is empty")
		}
		if node.MapID != cfg.MapID {
			return newError("CONFIG_BAD_NODE_ENUM", "node %s has map %s", id, node.MapID)
		}
		if node.X < 0 || node.X > 1 || node.Y < 0 || node.Y > 1 {
			return newError("CONFIG_BAD_NODE_COORD", "node %s coordinate out of range", id)
		}
		switch node.Shape {
		case "", "None":
		case "Circle":
			if node.Radius <= 0 {
				return newError("CONFIG_BAD_NODE_CIRCLE", "node %s has bad circle radius", id)
			}
		case "Polygon":
			if !validPolygonPoints(node.Points) {
				return newError("CONFIG_BAD_NODE_POLYGON", "node %s has bad polygon points", id)
			}
		default:
			return newError("CONFIG_BAD_NODE_SHAPE", "node %s has shape %s", id, node.Shape)
		}
	}
	for id, scenario := range cfg.Scenarios {
		if scenario.ID == "" {
			return newError("CONFIG_DUP_SCENARIO", "scenario id is empty")
		}
		for _, tagID := range scenario.MapTagIDs {
			if _, ok := cfg.MapTags[tagID]; !ok {
				return newError("CONFIG_BAD_SCENARIO_MAP_TAG", "scenario %s references missing map tag %s", id, tagID)
			}
		}
	}
	for id, tpl := range cfg.RouteTemplates {
		if tpl.ID == "" {
			return newError("CONFIG_DUP_ROUTE_TEMPLATE", "route template id is empty")
		}
		if tpl.MapID != cfg.MapID {
			return newError("CONFIG_BAD_ROUTE_TEMPLATE", "template %s has map %s", id, tpl.MapID)
		}
		if tpl.RecommendedMin > tpl.RecommendedMax {
			return newError("CONFIG_BAD_TEMPLATE_LIMIT", "template %s recommended min > max", id)
		}
		for _, scenarioID := range tpl.ScenarioIDs {
			if _, ok := cfg.Scenarios[scenarioID]; !ok {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references missing scenario %s", id, scenarioID)
			}
		}
		for _, tagID := range tpl.MapTagIDs {
			if _, ok := cfg.MapTags[tagID]; !ok {
				return newError("CONFIG_BAD_TEMPLATE_MAP_TAG", "template %s references missing map tag %s", id, tagID)
			}
		}
	}
	for id, modifier := range cfg.EncounterModifiers {
		if _, ok := cfg.Scenarios[modifier.ScenarioID]; !ok {
			return newError("CONFIG_BAD_ENCOUNTER_MODIFIER", "modifier %s references missing scenario %s", id, modifier.ScenarioID)
		}
	}
	for id, edge := range cfg.Edges {
		if _, ok := cfg.Nodes[edge.FromNode]; !ok {
			return newError("CONFIG_BAD_EDGE_NODE", "edge %s references missing from node %s", id, edge.FromNode)
		}
		if _, ok := cfg.Nodes[edge.ToNode]; !ok {
			return newError("CONFIG_BAD_EDGE_NODE", "edge %s references missing to node %s", id, edge.ToNode)
		}
		for _, nodeID := range edge.RiskPoints {
			if _, ok := cfg.Nodes[nodeID]; !ok {
				return newError("CONFIG_BAD_RISK_POINT", "edge %s references missing risk point %s", id, nodeID)
			}
		}
		for _, nodeID := range edge.InterceptNodes {
			if _, ok := cfg.Nodes[nodeID]; !ok {
				return newError("CONFIG_BAD_INTERCEPT_NODE", "edge %s references missing intercept node %s", id, nodeID)
			}
		}
	}
	for id, vis := range cfg.Visibility {
		if _, ok := cfg.Nodes[vis.FromNode]; !ok {
			return newError("CONFIG_BAD_VISIBILITY_NODE", "visibility %s references missing from node %s", id, vis.FromNode)
		}
		if _, ok := cfg.Nodes[vis.ToNode]; !ok {
			return newError("CONFIG_BAD_VISIBILITY_NODE", "visibility %s references missing to node %s", id, vis.ToNode)
		}
	}
	edgeSet := make(map[string]struct{}, len(cfg.Edges)*2)
	for _, edge := range cfg.Edges {
		edgeSet[edge.FromNode+"->"+edge.ToNode] = struct{}{}
		if edge.Bidirectional {
			edgeSet[edge.ToNode+"->"+edge.FromNode] = struct{}{}
		}
	}
	for id, route := range cfg.Routes {
		if route.MinPlayers > route.MaxPlayers {
			return newError("CONFIG_BAD_ROUTE_LIMIT", "route %s min players > max players", id)
		}
		if len(route.Nodes) == 0 {
			return newError("CONFIG_BAD_ROUTE_NODE", "route %s has no nodes", id)
		}
		for _, nodeID := range route.Nodes {
			if _, ok := cfg.Nodes[nodeID]; !ok {
				return newError("CONFIG_BAD_ROUTE_NODE", "route %s references missing node %s", id, nodeID)
			}
		}
		for i := 1; i < len(route.Nodes); i++ {
			if _, ok := edgeSet[route.Nodes[i-1]+"->"+route.Nodes[i]]; !ok {
				return newError("CONFIG_ROUTE_NOT_CONNECTED", "route %s has disconnected segment %s -> %s", id, route.Nodes[i-1], route.Nodes[i])
			}
		}
	}
	if !hasPlantSite(cfg, "A") || !hasPlantSite(cfg, "B") {
		return newError("CONFIG_NO_PLANT_SITE", "map %s must include plant nodes for A and B", cfg.MapID)
	}
	return nil
}

func (c CombatConstants) Int(key string, fallback int) int {
	if c.Values == nil {
		return fallback
	}
	raw, ok := c.Values[key]
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(raw.Value)
	if err != nil {
		return fallback
	}
	return n
}

func (c CombatConstants) Float(key string, fallback float64) float64 {
	if c.Values == nil {
		return fallback
	}
	raw, ok := c.Values[key]
	if !ok {
		return fallback
	}
	n, err := strconv.ParseFloat(raw.Value, 64)
	if err != nil {
		return fallback
	}
	return n
}

func hasPlantSite(cfg *MapConfig, site string) bool {
	for _, node := range cfg.Nodes {
		if node.Site != site {
			continue
		}
		for _, usage := range node.AreaUsages {
			if usage == "Plant" {
				return true
			}
		}
	}
	return false
}

func validPolygonPoints(points string) bool {
	if points == "" {
		return false
	}
	count := 0
	for _, pair := range splitNonEmpty(points, ";") {
		xy := splitNonEmpty(pair, ",")
		if len(xy) != 2 {
			return false
		}
		if _, err := strconv.ParseFloat(xy[0], 64); err != nil {
			return false
		}
		if _, err := strconv.ParseFloat(xy[1], 64); err != nil {
			return false
		}
		count++
	}
	return count >= 3
}

func splitNonEmpty(s, sep string) []string {
	var out []string
	for _, item := range fastSplit(s, sep) {
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func fastSplit(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	var out []string
	start := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			out = append(out, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	out = append(out, s[start:])
	return out
}
