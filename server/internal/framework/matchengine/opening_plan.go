package matchengine

import "sort"

type openingSelection struct {
	Template       RouteTemplate
	PrimaryRoute   Route
	AttemptOrdinal int
	UsedDefault    bool
}

func (e *MatchEngine) planOpening(roundSeed int64, side string) (openingSelection, error) {
	seedLabel, defaultKey := "opening_plan", "DefaultStrategyTemplateID"
	if side == SideCT {
		seedLabel, defaultKey = "ct_setup", "DefaultCTSetupTemplateID"
	}
	defaultID := e.input.MapConfig.CombatConstants.Values[defaultKey].Value
	candidates := make([]RouteTemplate, 0, len(e.input.MapConfig.RouteTemplates))
	for _, template := range e.input.MapConfig.RouteTemplates {
		if template.Side == side && template.ID != defaultID {
			candidates = append(candidates, template)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })

	for attemptOrdinal := 0; attemptOrdinal <= 1; attemptOrdinal++ {
		if len(candidates) == 0 {
			continue
		}
		rng := newDerivedRand(roundSeed, seedLabel, attemptOrdinal)
		candidate := candidates[rng.Intn(len(candidates))]
		selection, err := e.resolveOpeningTemplate(candidate, side, rng)
		if err == nil {
			selection.AttemptOrdinal = attemptOrdinal
			return selection, nil
		}
	}

	defaultTemplate, ok := e.input.MapConfig.RouteTemplates[defaultID]
	if !ok || defaultTemplate.Side != side {
		return openingSelection{}, newError("INVALID_OPENING_PLAN", "%s default template %s is missing or has the wrong side", side, defaultID)
	}
	selection, err := e.resolveOpeningTemplate(defaultTemplate, side, newDerivedRand(roundSeed, seedLabel, "default"))
	if err != nil {
		return openingSelection{}, newError("INVALID_OPENING_PLAN", "%s default template %s is invalid: %v", side, defaultID, err)
	}
	selection.AttemptOrdinal = 1
	selection.UsedDefault = true
	return selection, nil
}

func (e *MatchEngine) resolveOpeningTemplate(template RouteTemplate, side string, rng interface{ Intn(int) int }) (openingSelection, error) {
	if template.Side != side || len(template.RouteIDs) == 0 || len(template.RouteAllocations) != len(template.RouteIDs) {
		return openingSelection{}, newError("INVALID_OPENING_PLAN", "template %s has an invalid side or route set", template.ID)
	}
	routeIDs := append([]string(nil), template.RouteIDs...)
	sort.Strings(routeIDs)
	total := 0
	for _, routeID := range routeIDs {
		route, ok := e.input.MapConfig.Routes[routeID]
		count, allocated := template.RouteAllocations[routeID]
		if !ok || !allocated || route.Side != side || count < route.MinPlayers || count > route.MaxPlayers || !e.routeReachableFromSpawn(route, side) {
			return openingSelection{}, newError("INVALID_OPENING_PLAN", "template %s has invalid allocation or reachability for route %s", template.ID, routeID)
		}
		total += count
	}
	if total != 5 {
		return openingSelection{}, newError("INVALID_OPENING_PLAN", "template %s allocates %d players instead of 5", template.ID, total)
	}
	primaryID := routeIDs[rng.Intn(len(routeIDs))]
	return openingSelection{Template: template, PrimaryRoute: e.input.MapConfig.Routes[primaryID]}, nil
}

func (e *MatchEngine) routeReachableFromSpawn(route Route, side string) bool {
	spawnID := "T_SPAWN"
	if side == SideCT {
		spawnID = "CT_SPAWN"
	}
	if _, ok := e.input.MapConfig.Nodes[spawnID]; !ok {
		return false
	}
	adjacent := make(map[string][]string, len(e.input.MapConfig.Nodes))
	for _, edge := range e.input.MapConfig.Edges {
		adjacent[edge.FromNode] = append(adjacent[edge.FromNode], edge.ToNode)
		if edge.Bidirectional {
			adjacent[edge.ToNode] = append(adjacent[edge.ToNode], edge.FromNode)
		}
	}
	visited := map[string]bool{spawnID: true}
	queue := []string{spawnID}
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
	for _, nodeID := range route.Nodes {
		if !visited[nodeID] {
			return false
		}
	}
	return true
}
