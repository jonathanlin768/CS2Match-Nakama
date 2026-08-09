package matchengine

import (
	"fmt"
	"math"
	"strconv"
)

var requiredCombatConstKeys = []string{
	"RoundTimeLimit",
	"BombExplodeTime",
	"BasePlantTime",
	"BaseDefuseTime",
	"BasePickupTime",
	"MinPickupTime",
	"MaxPickupTime",
	"ForceExecuteThreshold",
	"DecisionDelay",
	"MaxDecisionCount",
	"MaxEncounterPulses",
	"MinCombatDuration",
	"MaxCombatDuration",
	"PulseFireWindow",
	"CombatScale",
	"MinDamagePotential",
	"MaxDamagePotential",
	"MinExposureModifier",
	"MaxExposureModifier",
	"BaseNoise",
	"MaxRandomNoise",
	"CloseScoreGap",
	"DecisiveScoreGap",
	"DefaultStrategyTemplateID",
	"DefaultCTSetupTemplateID",
	"StrategyRepeatWindow",
	"RepeatFreeCount",
	"StrategyRepeatPenalty",
	"SuccessBonusPerRound",
	"MaxPreviousSuccessBonus",
	"CounterMemoryBonus",
	"MinStrategyWeight",
	"MaxStrategyWeight",
	"ControlIntelTTL",
	"CommunicationDelay",
	"SoundIntelMinConfidence",
	"SoundIntelMaxConfidence",
	"DeathIntelMaxConfidence",
	"MinIntelTTL",
	"MaxIntelTTL",
	"UtilityBudget",
	"MaxStateTransitions",
	"MaxScheduledActions",
	"MaxEffectsPerTimestamp",
	"MaxNoOpTransitions",
	"MaxRotationsPerTeam",
	"MaxRoundTimeline",
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

var scenarioWeightAttributes = []string{
	"aim", "reaction", "positioning", "awareness", "teamplay",
	"utility", "composure", "mobility", "endurance", "discipline",
}

var requiredDust2Templates = []string{
	"A_Long_Rush", "A_Short_Split", "B_Tunnel_Explode", "Mid_To_B", "Default_Pick", "Fake_A_Go_B",
}

var requiredDust2Routes = []string{
	"D2_T_A_LONG", "D2_T_A_SHORT", "D2_T_B_TUNNEL", "D2_T_MID_B", "D2_T_MID_CONTROL", "D2_T_A_FAKE_PRESSURE",
	"D2_T_ROTATE_A_TO_B", "D2_T_ROTATE_B_TO_A", "D2_CT_A_LONG_HOLD", "D2_CT_A_SHORT_HOLD", "D2_CT_B_HOLD",
	"D2_CT_MID_HOLD", "D2_CT_REINFORCE_A", "D2_CT_REINFORCE_B", "D2_CT_RETAKE_A_FROM_B", "D2_CT_RETAKE_B_FROM_A",
}

var requiredDust2Nodes = []string{
	"T_SPAWN", "LONG_DOOR", "A_LONG", "PIT", "A_SITE", "MID", "CATWALK", "SHORT",
	"B_UPPER", "B_TUNNEL", "B_SITE", "CT_SPAWN", "CT_MID", "B_DOOR", "CAR",
}

var requiredDust2Edges = []string{
	"EDGE_T_LONG", "EDGE_T_MID", "EDGE_T_B_UPPER", "EDGE_CT_A", "EDGE_CT_MID", "EDGE_CT_B_DOOR",
}

var requiredDust2Visibility = []string{"VIS_PIT_A_LONG", "VIS_CAR_MID", "VIS_B_SITE_TUNNEL"}

var expectedCombatConstTypes = func() map[string]string {
	types := make(map[string]string, len(requiredCombatConstKeys))
	for _, key := range requiredCombatConstKeys {
		types[key] = "Int"
	}
	for _, key := range []string{
		"CombatScale", "MinDamagePotential", "MaxDamagePotential", "MinExposureModifier", "MaxExposureModifier",
		"BaseNoise", "MaxRandomNoise", "CloseScoreGap", "DecisiveScoreGap", "StrategyRepeatPenalty",
		"SuccessBonusPerRound", "MaxPreviousSuccessBonus", "CounterMemoryBonus", "MinStrategyWeight", "MaxStrategyWeight",
		"MinHitChance", "MaxHitChance", "MaxKillChance",
	} {
		types[key] = "Float"
	}
	types["DefaultStrategyTemplateID"] = "String"
	types["DefaultCTSetupTemplateID"] = "String"
	return types
}()

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
	return nil
}

func ValidateMapConfig(cfg *MapConfig) error {
	if cfg == nil {
		return newError("CONFIG_MISSING_MAP", "missing map config")
	}
	cfg.Warnings = nil
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
		value, ok := cfg.CombatConstants.Values[key]
		if !ok {
			return newError("CONFIG_MISSING_COMBAT_CONST", "missing combat const: %s", key)
		}
		if value.Key != "" && value.Key != key {
			return newError("CONFIG_DUP_COMBAT_CONST", "combat const map key %s contains row %s", key, value.Key)
		}
	}
	if err := validateCombatConstants(cfg.CombatConstants); err != nil {
		return err
	}
	for id, node := range cfg.Nodes {
		if id == "" || node.ID == "" || node.ID != id {
			return newError("CONFIG_DUP_NODE", "node key/id mismatch: %s/%s", id, node.ID)
		}
		if node.MapID != cfg.MapID || !oneOf(node.Site, "A", "B", "None") ||
			!oneOf(node.NodeType, "Spawn", "Connector", "Lane", "Cover", "Site") ||
			!oneOf(node.DefaultSide, SideT, SideCT, "Both") ||
			!oneOf(node.Floor, "Ground", "Ramp", "Upper") {
			return newError("CONFIG_BAD_NODE_ENUM", "node %s has map %s", id, node.MapID)
		}
		if !finiteInRange(node.X, 0, 1) || !finiteInRange(node.Y, 0, 1) {
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
		if hasString(node.AreaUsages, "KillSample") && !hasUsableNodeGeometry(node) {
			cfg.Warnings = append(cfg.Warnings, *newError("CONFIG_MISSING_KILL_SAMPLE", "node %s falls back to x/y for kill samples", id))
		}
	}
	for id, tag := range cfg.MapTags {
		if id == "" || tag.ID == "" || tag.ID != id {
			return newError("CONFIG_DUP_MAP_TAG", "map tag key/id mismatch: %s/%s", id, tag.ID)
		}
		if tag.MapID != cfg.MapID || tag.Value == "" || tag.ReasonCode == "" ||
			!oneOf(tag.Category, "Range", "Angle", "Site", "Risk", "Posture", "Timing") ||
			!oneOf(tag.Side, SideT, SideCT, "Both") {
			return newError("CONFIG_BAD_MAP_TAG", "map tag %s has invalid fields", id)
		}
	}
	for id, scenario := range cfg.Scenarios {
		if id == "" || scenario.ID == "" || scenario.ID != id {
			return newError("CONFIG_DUP_SCENARIO", "scenario key/id mismatch: %s/%s", id, scenario.ID)
		}
		if !oneOf(scenario.Route, "A_Long", "A_Short", "B_Tunnel", "Mid", "PostPlant_A", "PostPlant_B") ||
			!oneOf(scenario.Site, "A", "B", "None") ||
			!oneOf(scenario.Phase, "OpeningDuel", "MidControl", "SiteEntry", "Retake", "BombResolution") {
			return newError("CONFIG_BAD_SCENARIO_TAG", "scenario %s has invalid route/site/phase", id)
		}
		for _, tagID := range scenario.MapTagIDs {
			if _, ok := cfg.MapTags[tagID]; !ok {
				return newError("CONFIG_BAD_SCENARIO_MAP_TAG", "scenario %s references missing map tag %s", id, tagID)
			}
		}
	}
	if err := validateScenarioWeights(cfg); err != nil {
		return err
	}
	for id, route := range cfg.Routes {
		if id == "" || route.ID == "" || route.ID != id || !oneOf(route.Side, SideT, SideCT) {
			return newError("CONFIG_BAD_ROUTE_NODE", "route %s has invalid id or side", id)
		}
		if route.MinPlayers <= 0 || route.MaxPlayers <= 0 || route.MinPlayers > route.MaxPlayers {
			return newError("CONFIG_BAD_ROUTE_LIMIT", "route %s has invalid player limits", id)
		}
	}
	for id, tpl := range cfg.RouteTemplates {
		if id == "" || tpl.ID == "" || tpl.ID != id {
			return newError("CONFIG_DUP_ROUTE_TEMPLATE", "route template key/id mismatch: %s/%s", id, tpl.ID)
		}
		if tpl.MapID != cfg.MapID {
			return newError("CONFIG_BAD_ROUTE_TEMPLATE", "template %s has map %s", id, tpl.MapID)
		}
		if !oneOf(tpl.Side, SideT, SideCT) {
			return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s has invalid side %s", id, tpl.Side)
		}
		if tpl.RecommendedMin <= 0 || tpl.RecommendedMax <= 0 || tpl.RecommendedMin > tpl.RecommendedMax || tpl.RecommendedMax > 5 {
			return newError("CONFIG_BAD_TEMPLATE_LIMIT", "template %s has invalid recommended limits", id)
		}
		for _, scenarioID := range tpl.ScenarioIDs {
			if _, ok := cfg.Scenarios[scenarioID]; !ok {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references missing scenario %s", id, scenarioID)
			}
		}
		for _, tagID := range tpl.MapTagIDs {
			if _, ok := cfg.MapTags[tagID]; !ok {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references missing map tag %s", id, tagID)
			}
		}
		if len(tpl.RouteIDs) == 0 || len(tpl.RouteAllocations) == 0 {
			return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s has no route references", id)
		}
		total := 0
		for _, routeID := range tpl.RouteIDs {
			route, ok := cfg.Routes[routeID]
			if !ok || route.Side != tpl.Side {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references missing or cross-side route %s", id, routeID)
			}
			count, ok := tpl.RouteAllocations[routeID]
			if !ok || count < route.MinPlayers || count > route.MaxPlayers {
				return newError("CONFIG_BAD_TEMPLATE_LIMIT", "template %s allocation for %s violates route limits", id, routeID)
			}
			total += count
		}
		if len(tpl.RouteAllocations) != len(tpl.RouteIDs) || total != 5 {
			return newError("CONFIG_BAD_TEMPLATE_LIMIT", "template %s route allocations total %d instead of 5", id, total)
		}
		for _, setupID := range tpl.CommonCTSetupIDs {
			setup, ok := cfg.RouteTemplates[setupID]
			if tpl.Side != SideT || !ok || setup.Side != SideCT {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references invalid CT setup %s", id, setupID)
			}
		}
		for _, fallbackID := range tpl.FailureFallbacks {
			fallback, ok := cfg.RouteTemplates[fallbackID]
			if !ok || fallback.Side != tpl.Side {
				return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "template %s references invalid fallback %s", id, fallbackID)
			}
		}
	}
	for id, modifier := range cfg.EncounterModifiers {
		if _, ok := cfg.Scenarios[modifier.ScenarioID]; !ok {
			return newError("CONFIG_BAD_ENCOUNTER_MODIFIER", "modifier %s references missing scenario %s", id, modifier.ScenarioID)
		}
		if modifier.ReasonCode == "" {
			return newError("CONFIG_BAD_REASON_CODE", "modifier %s has empty reason code", id)
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
			node, ok := cfg.Nodes[nodeID]
			if !ok || !finiteInRange(node.X, 0, 1) || !finiteInRange(node.Y, 0, 1) {
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
	if err := validateDefaultTemplates(cfg); err != nil {
		return err
	}
	if cfg.MapID == DefaultMapID && cfg.Version == DefaultMapVersion {
		if err := validateFormalDust2Coverage(cfg); err != nil {
			return err
		}
	}
	return nil
}

func validateCombatConstants(constants CombatConstants) error {
	for key, value := range constants.Values {
		if value.Key != "" && value.Key != key {
			return newError("CONFIG_DUP_COMBAT_CONST", "combat const map key %s contains row %s", key, value.Key)
		}
		expected, required := expectedCombatConstTypes[key]
		if required && value.ValueType != expected {
			return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s must be %s", key, expected)
		}
		switch value.ValueType {
		case "Int":
			n, err := strconv.Atoi(value.Value)
			if err != nil {
				return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s is not an int", key)
			}
			if err := validateNumericRange(key, float64(n), value.MinValue, value.MaxValue); err != nil {
				return err
			}
		case "Float":
			n, err := strconv.ParseFloat(value.Value, 64)
			if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
				return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s is not a finite float", key)
			}
			if err := validateNumericRange(key, n, value.MinValue, value.MaxValue); err != nil {
				return err
			}
		case "String":
			if value.Value == "" {
				return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s is an empty string", key)
			}
		case "Bool":
			if _, err := strconv.ParseBool(value.Value); err != nil {
				return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s is not a bool", key)
			}
		default:
			return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s has unsupported type %s", key, value.ValueType)
		}
	}
	if constants.Int("CommunicationDelay", -1) != 0 {
		return newError("CONFIG_UNSUPPORTED_COMMUNICATION_DELAY", "CommunicationDelay must be zero in the first version")
	}
	for _, key := range []string{"MaxStateTransitions", "MaxScheduledActions", "MaxEffectsPerTimestamp", "MaxNoOpTransitions", "MaxRoundTimeline"} {
		if constants.Int(key, 0) <= 0 {
			return newError("CONFIG_BAD_COMBAT_CONST_RANGE", "combat const %s must be positive", key)
		}
	}
	for _, pair := range [][2]string{
		{"MinPickupTime", "MaxPickupTime"}, {"MinCombatDuration", "MaxCombatDuration"},
		{"MinDamagePotential", "MaxDamagePotential"}, {"MinExposureModifier", "MaxExposureModifier"},
		{"MinIntelTTL", "MaxIntelTTL"}, {"MinAttribute", "MaxAttribute"}, {"MinHP", "MaxHP"},
		{"MinStamina", "MaxStamina"}, {"MinFocus", "MaxFocus"}, {"MinHitChance", "MaxHitChance"},
		{"MinPlantTime", "MaxPlantTime"}, {"MinDefuseTime", "MaxDefuseTime"}, {"MinMoveTime", "MaxMoveTime"},
	} {
		if constants.Float(pair[0], math.NaN()) > constants.Float(pair[1], math.NaN()) {
			return newError("CONFIG_BAD_COMBAT_CONST_RANGE", "%s exceeds %s", pair[0], pair[1])
		}
	}
	if constants.Int("BasePickupTime", 0) <= 0 || constants.Int("BasePlantTime", 0) <= 0 ||
		constants.Int("BaseDefuseTime", 0) <= 0 || constants.Int("BombExplodeTime", 0) <= 0 ||
		constants.Int("BasePickupTime", 0) < constants.Int("MinPickupTime", 0) ||
		constants.Int("BasePickupTime", 0) > constants.Int("MaxPickupTime", 0) ||
		constants.Int("BasePlantTime", 0) < constants.Int("MinPlantTime", 0) ||
		constants.Int("BasePlantTime", 0) > constants.Int("MaxPlantTime", 0) ||
		constants.Int("BaseDefuseTime", 0) < constants.Int("MinDefuseTime", 0) ||
		constants.Int("BaseDefuseTime", 0) > constants.Int("MaxDefuseTime", 0) {
		return newError("CONFIG_BAD_BOMB_CONST", "bomb action constants violate configured bounds")
	}
	return nil
}

func validateNumericRange(key string, value float64, minRaw, maxRaw string) error {
	if minRaw != "" {
		minValue, err := strconv.ParseFloat(minRaw, 64)
		if err != nil || math.IsNaN(minValue) || math.IsInf(minValue, 0) {
			return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s has invalid min_value", key)
		}
		if value < minValue {
			return newError("CONFIG_BAD_COMBAT_CONST_RANGE", "combat const %s is below min_value", key)
		}
	}
	if maxRaw != "" {
		maxValue, err := strconv.ParseFloat(maxRaw, 64)
		if err != nil || math.IsNaN(maxValue) || math.IsInf(maxValue, 0) {
			return newError("CONFIG_BAD_COMBAT_CONST_TYPE", "combat const %s has invalid max_value", key)
		}
		if value > maxValue {
			return newError("CONFIG_BAD_COMBAT_CONST_RANGE", "combat const %s is above max_value", key)
		}
	}
	return nil
}

func validateScenarioWeights(cfg *MapConfig) error {
	for scenarioID := range cfg.Scenarios {
		weights := make(map[string]int, len(scenarioWeightAttributes))
		for _, modifier := range cfg.EncounterModifiers {
			if modifier.ScenarioID != scenarioID || modifier.Factor != "ScenarioWeight" {
				continue
			}
			if !hasString(scenarioWeightAttributes, modifier.Attribute) || modifier.Weight < 0 {
				return newError("CONFIG_BAD_SCENARIO_WEIGHT", "scenario %s has invalid attribute weight %s=%d", scenarioID, modifier.Attribute, modifier.Weight)
			}
			if _, exists := weights[modifier.Attribute]; exists {
				return newError("CONFIG_BAD_SCENARIO_WEIGHT", "scenario %s repeats attribute %s", scenarioID, modifier.Attribute)
			}
			weights[modifier.Attribute] = modifier.Weight
		}
		if len(weights) != len(scenarioWeightAttributes) {
			return newError("CONFIG_BAD_SCENARIO_WEIGHT", "scenario %s does not define all ten attributes", scenarioID)
		}
		total := 0
		for _, attribute := range scenarioWeightAttributes {
			total += weights[attribute]
		}
		if total != 100 {
			return newError("CONFIG_BAD_SCENARIO_WEIGHT", "scenario %s weight total is %d instead of 100", scenarioID, total)
		}
	}
	return nil
}

func validateDefaultTemplates(cfg *MapConfig) error {
	tID := cfg.CombatConstants.Values["DefaultStrategyTemplateID"].Value
	ctID := cfg.CombatConstants.Values["DefaultCTSetupTemplateID"].Value
	tTemplate, tOK := cfg.RouteTemplates[tID]
	ctTemplate, ctOK := cfg.RouteTemplates[ctID]
	if !tOK || tTemplate.Side != SideT || !ctOK || ctTemplate.Side != SideCT {
		return newError("CONFIG_BAD_TEMPLATE_SCENARIO", "default T/CT templates are missing or have the wrong side")
	}
	return nil
}

func validateFormalDust2Coverage(cfg *MapConfig) error {
	for _, id := range requiredDust2Templates {
		if template, ok := cfg.RouteTemplates[id]; !ok || template.Side != SideT {
			return incompleteDust2("missing T template %s", id)
		}
	}
	ctSites := map[string]bool{}
	ctCount := 0
	for _, template := range cfg.RouteTemplates {
		if template.Side == SideCT {
			ctCount++
			ctSites[template.TargetSite] = true
		}
	}
	if ctCount < 3 || !ctSites["A"] || !ctSites["B"] || !ctSites["None"] {
		return incompleteDust2("CT setup coverage must include A, B and Mid/default")
	}
	for _, id := range requiredDust2Routes {
		if _, ok := cfg.Routes[id]; !ok {
			return incompleteDust2("missing route %s", id)
		}
	}
	for _, id := range requiredDust2Nodes {
		if _, ok := cfg.Nodes[id]; !ok {
			return incompleteDust2("missing core node %s", id)
		}
	}
	for _, id := range requiredDust2Edges {
		if _, ok := cfg.Edges[id]; !ok {
			return incompleteDust2("missing core edge %s", id)
		}
	}
	for _, id := range requiredDust2Visibility {
		if _, ok := cfg.Visibility[id]; !ok {
			return incompleteDust2("missing key visibility %s", id)
		}
	}
	phases := map[string]bool{}
	for _, scenario := range cfg.Scenarios {
		phases[scenario.Phase] = true
	}
	for _, phase := range []string{"OpeningDuel", "MidControl", "SiteEntry", "Retake", "BombResolution"} {
		if !phases[phase] {
			return incompleteDust2("missing scenario phase %s", phase)
		}
	}
	if !allRouteNodesReachable(cfg) {
		return newError("CONFIG_UNREACHABLE_NODE", "a route node is unreachable from both spawns")
	}
	return nil
}

func allRouteNodesReachable(cfg *MapConfig) bool {
	adjacent := make(map[string][]string, len(cfg.Nodes))
	for _, edge := range cfg.Edges {
		adjacent[edge.FromNode] = append(adjacent[edge.FromNode], edge.ToNode)
		if edge.Bidirectional {
			adjacent[edge.ToNode] = append(adjacent[edge.ToNode], edge.FromNode)
		}
	}
	visited := map[string]bool{"T_SPAWN": true, "CT_SPAWN": true}
	queue := []string{"T_SPAWN", "CT_SPAWN"}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacent[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	for _, route := range cfg.Routes {
		for _, nodeID := range route.Nodes {
			if !visited[nodeID] {
				return false
			}
		}
	}
	return true
}

func incompleteDust2(format string, args ...interface{}) *EngineError {
	return newError("CONFIG_INCOMPLETE_DUST2_COVERAGE", format, args...)
}

func hasUsableNodeGeometry(node MapNode) bool {
	switch node.Shape {
	case "Circle":
		return node.Radius > 0
	case "Polygon":
		return validPolygonPoints(node.Points)
	default:
		return false
	}
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func oneOf(value string, allowed ...string) bool {
	return hasString(allowed, value)
}

func finiteInRange(value, minValue, maxValue float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= minValue && value <= maxValue
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
