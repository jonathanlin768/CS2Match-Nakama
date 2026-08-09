package match

import (
	"strconv"
	"strings"

	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

func buildMapConfigFromTables(mapID string) (*matchengine.MapConfig, error) {
	if cfg.Global == nil {
		return nil, &MatchError{Code: "CONFIG_NOT_INITIALIZED", Message: "config tables are not initialized"}
	}
	out := &matchengine.MapConfig{
		MapID:              mapID,
		MapName:            matchengine.DefaultMapName,
		Version:            matchengine.DefaultMapVersion,
		RouteTemplates:     map[string]matchengine.RouteTemplate{},
		Scenarios:          map[string]matchengine.Scenario{},
		MapTags:            map[string]matchengine.MapTag{},
		EncounterModifiers: map[string]matchengine.EncounterModifier{},
		Nodes:              map[string]matchengine.MapNode{},
		Edges:              map[string]matchengine.MapEdge{},
		Visibility:         map[string]matchengine.Visibility{},
		Routes:             map[string]matchengine.Route{},
		CombatConstants:    matchengine.CombatConstants{Values: map[string]matchengine.CombatConstValue{}},
	}

	for _, row := range cfg.Global.TbRouteTemplate.GetDataList() {
		if row.MapId != mapID {
			continue
		}
		out.RouteTemplates[row.Id] = matchengine.RouteTemplate{
			ID:               row.Id,
			MapID:            row.MapId,
			Side:             row.Side,
			TargetSite:       row.TargetSite,
			Tempo:            row.Tempo,
			RecommendedMin:   int(row.RecommendedMin),
			RecommendedMax:   int(row.RecommendedMax),
			RequiredRoles:    append([]string(nil), row.RequiredRoles...),
			KeyAttributes:    parseKeyAttributes(row.KeyAttributes),
			RouteIDs:         append([]string(nil), row.RouteIds...),
			RouteAllocations: copyRouteAllocations(row.RouteAllocations),
			ScenarioIDs:      append([]string(nil), row.ScenarioIds...),
			MapTagIDs:        append([]string(nil), row.MapTagIds...),
			CommonCTSetupIDs: append([]string(nil), row.CommonCtSetupIds...),
			SuccessNextPhase: row.SuccessNextPhase,
			FailureFallbacks: append([]string(nil), row.FailureFallbacks...),
		}
	}
	for _, row := range cfg.Global.TbScenario.GetDataList() {
		out.Scenarios[row.Id] = matchengine.Scenario{
			ID:             row.Id,
			Route:          row.Route,
			Phase:          row.Phase,
			Range:          row.Range,
			Site:           row.Site,
			Tempo:          row.Tempo,
			Posture:        row.Posture,
			UtilityContext: row.UtilityContext,
			MapTagIDs:      append([]string(nil), row.MapTagIds...),
			BaseTimeCost:   int(row.BaseTimeCost),
			BaseWeight:     int(row.BaseWeight),
		}
	}
	for _, row := range cfg.Global.TbMapTag.GetDataList() {
		if row.MapId != mapID {
			continue
		}
		out.MapTags[row.Id] = matchengine.MapTag{
			ID:          row.Id,
			MapID:       row.MapId,
			Category:    row.Category,
			Value:       row.Value,
			Side:        row.Side,
			Weight:      int(row.Weight),
			ReasonCode:  row.ReasonCode,
			Description: row.Description,
		}
	}
	for _, row := range cfg.Global.TbEncounterModifier.GetDataList() {
		out.EncounterModifiers[row.Id] = matchengine.EncounterModifier{
			ID:         row.Id,
			ScenarioID: row.ScenarioId,
			Factor:     row.Factor,
			Side:       row.Side,
			Attribute:  row.Attribute,
			Weight:     int(row.Weight),
			ReasonCode: row.ReasonCode,
		}
	}
	for _, row := range cfg.Global.TbMapNode.GetDataList() {
		if row.MapId != mapID {
			continue
		}
		out.Nodes[row.Id] = matchengine.MapNode{
			ID:          row.Id,
			MapID:       row.MapId,
			Name:        row.Name,
			Zone:        row.Zone,
			Site:        row.Site,
			NodeType:    row.NodeType,
			DefaultSide: row.DefaultSide,
			X:           float64(row.X),
			Y:           float64(row.Y),
			Floor:       row.Floor,
			AreaUsages:  append([]string(nil), row.AreaUsages...),
			Shape:       row.Shape,
			Radius:      float64(row.Radius),
			Points:      row.Points,
		}
	}
	for _, row := range cfg.Global.TbMapEdge.GetDataList() {
		out.Edges[row.Id] = matchengine.MapEdge{
			ID:             row.Id,
			FromNode:       row.FromNode,
			ToNode:         row.ToNode,
			BaseTime:       int(row.BaseTime),
			StaminaCost:    int(row.StaminaCost),
			Risk:           int(row.Risk),
			Noise:          int(row.Noise),
			RiskPoints:     append([]string(nil), row.RiskPoints...),
			InterceptNodes: append([]string(nil), row.InterceptNodes...),
			Bidirectional:  row.Bidirectional,
		}
	}
	for _, row := range cfg.Global.TbVisibility.GetDataList() {
		out.Visibility[row.Id] = matchengine.Visibility{
			ID:               row.Id,
			FromNode:         row.FromNode,
			ToNode:           row.ToNode,
			Visible:          row.Visible,
			Range:            row.Range,
			AngleAdvantage:   row.AngleAdvantage,
			Elevation:        row.Elevation,
			CoverModifier:    int(row.CoverModifier),
			ExposureModifier: int(row.ExposureModifier),
		}
	}
	for _, row := range cfg.Global.TbRoute.GetDataList() {
		out.Routes[row.Id] = matchengine.Route{
			ID:         row.Id,
			Name:       row.Name,
			Side:       row.Side,
			TargetSite: row.TargetSite,
			Nodes:      append([]string(nil), row.Nodes...),
			MinPlayers: int(row.MinPlayers),
			MaxPlayers: int(row.MaxPlayers),
			StyleTags:  append([]string(nil), row.StyleTags...),
		}
	}
	for _, row := range cfg.Global.TbCombatConst.GetDataList() {
		out.CombatConstants.Values[row.Key] = matchengine.CombatConstValue{
			Key:       row.Key,
			Category:  row.Category,
			ValueType: row.ValueType,
			Value:     row.Value,
			MinValue:  row.MinValue,
			MaxValue:  row.MaxValue,
			Unit:      row.Unit,
		}
	}
	if err := matchengine.ValidateMapConfig(out); err != nil {
		if me, ok := err.(*matchengine.EngineError); ok {
			return out, &MatchError{Code: me.Code, Message: me.Message}
		}
		return out, err
	}
	return out, nil
}

func parseKeyAttributes(raw string) map[string]float64 {
	out := map[string]float64{}
	for _, item := range strings.Split(raw, ";") {
		parts := strings.Split(item, "=")
		if len(parts) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			continue
		}
		out[strings.TrimSpace(parts[0])] = value
	}
	return out
}

func copyRouteAllocations(source map[string]int32) map[string]int {
	out := make(map[string]int, len(source))
	for routeID, count := range source {
		out[routeID] = int(count)
	}
	return out
}
