package matchengine

import (
	"math"
	"sort"
	"strings"
)

type ReasonRecord struct {
	Code           string
	Source         string
	Value          float64
	Weight         float64
	Detail         string
	MainFactor     string
	Modifiers      []ReasonModifier
	ScoreDelta     float64
	Probability    *float64
	Formula        string
	Inputs         map[string]float64
	StateChanges   []ReasonStateChange
	SourceActionID string
	SourceEffectID string
}

type StrategyMemory struct {
	PreviousSuccess map[string]int
	RecentTemplates []string
	CounterReads    map[string]int
	SideTendency    map[string]float64
	TeamStyle       map[string]float64
}

type StrategyScore struct {
	TemplateID           string
	TemplateBaseWeight   float64
	LineupFitScore       float64
	CurrentScorePressure float64
	PreviousSuccessBonus float64
	RepeatPenalty        float64
	CounterReadRisk      float64
	DeterministicScore   float64
	RandomNoise          float64
	FinalScore           float64
	Reasons              []ReasonRecord
}

type StrategySelection struct {
	Template       RouteTemplate
	Score          StrategyScore
	AttemptOrdinal int
	UsedDefault    bool
}

type RoleAssignmentResult struct {
	Assignments []RoleAssignment
	GapRoles    []string
	Penalty     float64
	Reasons     []ReasonRecord
}

func ScoreStrategyCandidates(config *MapConfig, team TeamInput, side string, scoreFor, scoreAgainst int, memory StrategyMemory, roundSeed int64, attemptOrdinal int) ([]StrategyScore, error) {
	if config == nil || !validSide(side) {
		return nil, newError("INVALID_STRATEGY_INPUT", "strategy scoring requires config and side")
	}
	templateIDs := make([]string, 0, len(config.RouteTemplates))
	for templateID, template := range config.RouteTemplates {
		if template.Side == side {
			templateIDs = append(templateIDs, templateID)
		}
	}
	sort.Strings(templateIDs)
	if len(templateIDs) == 0 {
		return nil, newError("INVALID_OPENING_PLAN", "no %s strategy templates are configured", side)
	}
	results := make([]StrategyScore, 0, len(templateIDs))
	for _, templateID := range templateIDs {
		template := config.RouteTemplates[templateID]
		base := strategyTemplateBase(config, template)
		fit := lineupFit(team, template)
		pressure := scorePressure(template, scoreFor, scoreAgainst)
		success := math.Min(float64(memory.PreviousSuccess[templateID])*config.CombatConstants.Float("SuccessBonusPerRound", 0), config.CombatConstants.Float("MaxPreviousSuccessBonus", 0))
		repeats := recentTemplateCount(memory.RecentTemplates, templateID, config.CombatConstants.Int("StrategyRepeatWindow", 0))
		repeatPenalty := float64(maxInt(0, repeats-config.CombatConstants.Int("RepeatFreeCount", 0))) * config.CombatConstants.Float("StrategyRepeatPenalty", 0)
		counterRisk := float64(memory.CounterReads[templateID]) * config.CombatConstants.Float("CounterMemoryBonus", 0)
		deterministic := base + fit + pressure + success - repeatPenalty - counterRisk
		amplitude := strategyNoiseAmplitude(team, config.CombatConstants)
		rawNoise := (stableUnit(roundSeed, "strategy", side, team.TeamID, templateID, attemptOrdinal)*2 - 1) * amplitude
		results = append(results, StrategyScore{
			TemplateID: templateID, TemplateBaseWeight: base, LineupFitScore: fit, CurrentScorePressure: pressure,
			PreviousSuccessBonus: success, RepeatPenalty: repeatPenalty, CounterReadRisk: counterRisk,
			DeterministicScore: deterministic, RandomNoise: rawNoise,
			Reasons: []ReasonRecord{
				{Code: "TEMPLATE_BASE", Source: templateID, Value: base, Weight: 1},
				{Code: "LINEUP_FIT", Source: team.TeamID, Value: fit, Weight: 1},
				{Code: "SCORE_PRESSURE", Source: team.TeamID, Value: pressure, Weight: 1},
				{Code: "PREVIOUS_SUCCESS", Source: templateID, Value: success, Weight: 1},
				{Code: "REPEAT_PENALTY", Source: templateID, Value: -repeatPenalty, Weight: 1},
				{Code: "COUNTER_READ_RISK", Source: templateID, Value: -counterRisk, Weight: 1},
			},
		})
	}
	protectDecisiveStrategyTrend(results, config.CombatConstants.Float("DecisiveScoreGap", 0))
	minScore := config.CombatConstants.Float("MinStrategyWeight", 1)
	maxScore := config.CombatConstants.Float("MaxStrategyWeight", 100)
	for index := range results {
		results[index].FinalScore = clampFloat(results[index].DeterministicScore+results[index].RandomNoise, minScore, maxScore)
		results[index].Reasons = append(results[index].Reasons, ReasonRecord{Code: "BOUNDED_RANDOM_NOISE", Source: results[index].TemplateID, Value: results[index].RandomNoise, Weight: 1})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].FinalScore != results[j].FinalScore {
			return results[i].FinalScore > results[j].FinalScore
		}
		return results[i].TemplateID < results[j].TemplateID
	})
	return results, nil
}

func SelectStrategyTemplate(config *MapConfig, team TeamInput, side string, scoreFor, scoreAgainst int, memory StrategyMemory, roundSeed int64, attemptOrdinal int) (StrategySelection, error) {
	scores, err := ScoreStrategyCandidates(config, team, side, scoreFor, scoreAgainst, memory, roundSeed, attemptOrdinal)
	if err != nil {
		return StrategySelection{}, err
	}
	winner := scores[0]
	return StrategySelection{Template: config.RouteTemplates[winner.TemplateID], Score: winner, AttemptOrdinal: attemptOrdinal}, nil
}

// SelectCTSetup intentionally has no T plan/template argument. Its seed and
// inputs are restricted to CT-owned state and match memory.
func SelectCTSetup(config *MapConfig, ctTeam TeamInput, scoreFor, scoreAgainst int, memory StrategyMemory, roundSeed int64, attemptOrdinal int) (StrategySelection, error) {
	ctSeed := deriveSeed(roundSeed, "ct_setup", attemptOrdinal)
	return SelectStrategyTemplate(config, ctTeam, SideCT, scoreFor, scoreAgainst, memory, ctSeed, attemptOrdinal)
}

func AssignRoles(team TeamInput, template RouteTemplate) RoleAssignmentResult {
	playerIDs := make([]string, 0, len(team.Players))
	profiles := make(map[string]PlayerProfile, len(team.Players))
	for _, profile := range team.Players {
		playerIDs = append(playerIDs, profile.PlayerID)
		profiles[profile.PlayerID] = profile
	}
	sort.Strings(playerIDs)
	required := append([]string(nil), template.RequiredRoles...)
	sort.Strings(required)
	available := append([]string(nil), playerIDs...)
	result := RoleAssignmentResult{}
	for _, role := range required {
		bestIndex, bestScore, exact := -1, -1.0, false
		for index, playerID := range available {
			profile := profiles[playerID]
			score := roleFitScore(profile, role)
			isExact := hasFold(profile.RoleTags, role)
			if score > bestScore || (score == bestScore && (bestIndex < 0 || playerID < available[bestIndex])) {
				bestIndex, bestScore, exact = index, score, isExact
			}
		}
		if bestIndex < 0 {
			result.GapRoles = append(result.GapRoles, role)
			continue
		}
		playerID := available[bestIndex]
		result.Assignments = append(result.Assignments, RoleAssignment{PlayerID: playerID, Role: role})
		available = append(available[:bestIndex], available[bestIndex+1:]...)
		if !exact {
			result.GapRoles = append(result.GapRoles, role)
			result.Penalty += 5
			result.Reasons = append(result.Reasons, ReasonRecord{Code: "ROLE_GAP", Source: role, Value: -5, Weight: 1, Detail: playerID + " is the stable fallback"})
		}
	}
	for _, playerID := range available {
		role := "Rifler"
		if tags := profiles[playerID].RoleTags; len(tags) > 0 {
			sorted := append([]string(nil), tags...)
			sort.Strings(sorted)
			role = sorted[0]
		}
		result.Assignments = append(result.Assignments, RoleAssignment{PlayerID: playerID, Role: role})
	}
	sort.Slice(result.Assignments, func(i, j int) bool { return result.Assignments[i].PlayerID < result.Assignments[j].PlayerID })
	return result
}

func BuildRoundPlan(input *RoundInput, tTemplate, ctTemplate RouteTemplate) (RoundPlan, []ReasonRecord, error) {
	if input == nil || input.MapConfig == nil || tTemplate.Side != SideT || ctTemplate.Side != SideCT {
		return RoundPlan{}, nil, newError("INVALID_OPENING_PLAN", "round plan requires T and CT templates")
	}
	tRoles, ctRoles := AssignRoles(input.TeamT, tTemplate), AssignRoles(input.TeamCT, ctTemplate)
	plan := RoundPlan{
		TStrategyTemplateID: tTemplate.ID, CTSetupTemplateID: ctTemplate.ID,
		RoleAssignments: append(append([]RoleAssignment(nil), tRoles.Assignments...), ctRoles.Assignments...), OpeningRoutes: map[string]string{},
	}
	if err := assignOpeningRoutes(plan.OpeningRoutes, input.TeamT, tTemplate, input.MapConfig); err != nil {
		return RoundPlan{}, nil, err
	}
	if err := assignOpeningRoutes(plan.OpeningRoutes, input.TeamCT, ctTemplate, input.MapConfig); err != nil {
		return RoundPlan{}, nil, err
	}
	carrier, err := SelectBombCarrier(input.TeamT, tRoles.Assignments, plan.OpeningRoutes, input.MapConfig)
	if err != nil {
		return RoundPlan{}, nil, err
	}
	plan.BombCarrierID = carrier
	reasons := append(append([]ReasonRecord(nil), tRoles.Reasons...), ctRoles.Reasons...)
	return plan, reasons, nil
}

func SelectBombCarrier(team TeamInput, assignments []RoleAssignment, routes map[string]string, config *MapConfig) (string, error) {
	roles := make(map[string]string, len(assignments))
	for _, assignment := range assignments {
		roles[assignment.PlayerID] = assignment.Role
	}
	profiles := append([]PlayerProfile(nil), team.Players...)
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].PlayerID < profiles[j].PlayerID })
	bestID, bestScore := "", math.Inf(-1)
	for _, profile := range profiles {
		route, ok := config.Routes[routes[profile.PlayerID]]
		if !ok || len(route.Nodes) == 0 {
			continue
		}
		if _, feedback, err := FindBoundedPath(config, "T_SPAWN", route.Nodes[len(route.Nodes)-1], len(config.Nodes)*4); err != nil || feedback != nil {
			continue
		}
		score := float64(profile.Attributes.Composure+profile.Attributes.Discipline) / 2
		if strings.EqualFold(roles[profile.PlayerID], "Entry") || hasFold(profile.RoleTags, "Entry") {
			score -= 40
		}
		if strings.EqualFold(roles[profile.PlayerID], "Support") || hasFold(profile.RoleTags, "Support") {
			score += 10
		}
		if score > bestScore || (score == bestScore && profile.PlayerID < bestID) {
			bestID, bestScore = profile.PlayerID, score
		}
	}
	if bestID == "" {
		return "", newError("INVALID_OPENING_PLAN", "no reachable T bomb carrier")
	}
	return bestID, nil
}

func DeployOpeningActions(state *RoundState) ([]ScheduledAction, error) {
	if state == nil || state.Timeline != 0 {
		return nil, newError("INVALID_OPENING_PLAN", "opening deployment must run at timeline zero")
	}
	playerIDs := make([]string, 0, len(state.Players))
	for playerID := range state.Players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)
	actions := make([]ScheduledAction, 0, len(playerIDs))
	for ordinal, playerID := range playerIDs {
		player := state.Players[playerID]
		routeID := state.Plan.OpeningRoutes[playerID]
		route, ok := state.routes[routeID]
		if !ok || len(route.Nodes) == 0 {
			return nil, newError("INVALID_OPENING_PLAN", "player %s has no legal route", playerID)
		}
		nextNode := ""
		if player.Location.NodeID == route.Nodes[0] && len(route.Nodes) > 1 {
			nextNode = route.Nodes[1]
		} else if player.Location.NodeID != route.Nodes[0] {
			path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, player.Location.NodeID, route.Nodes[0], len(state.Nodes)*4)
			if err != nil || feedback != nil || len(path.NodeIDs) < 2 {
				return nil, newError("INVALID_OPENING_PLAN", "player %s cannot reach route %s", playerID, routeID)
			}
			nextNode = path.NodeIDs[1]
		}
		if nextNode == "" {
			action, err := startOpeningHold(state, playerID, routeID, ordinal)
			if err != nil {
				return nil, err
			}
			actions = append(actions, action)
			continue
		}
		edgeID, ok := edgeBetween(state.mapEdges, player.Location.NodeID, nextNode)
		if !ok {
			return nil, newError("INVALID_OPENING_PLAN", "route %s has no edge from %s to %s", routeID, player.Location.NodeID, nextNode)
		}
		templateID := state.Plan.TStrategyTemplateID
		if player.Side == SideCT {
			templateID = state.Plan.CTSetupTemplateID
		}
		template := state.routeTemplates[templateID]
		action, err := StartMovement(state, []string{playerID}, edgeID, MoveProfile{Tempo: template.Tempo}, "opening:"+playerID+":"+routeID, ordinal)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func DecayStrategyMemoryForSideSwitch(memory StrategyMemory) StrategyMemory {
	out := StrategyMemory{
		PreviousSuccess: copyStringIntMap(memory.PreviousSuccess), CounterReads: copyStringIntMap(memory.CounterReads),
		SideTendency: map[string]float64{}, TeamStyle: copyStringFloatMap(memory.TeamStyle),
	}
	for key, value := range out.PreviousSuccess {
		out.PreviousSuccess[key] = value / 2
	}
	for key, value := range out.CounterReads {
		out.CounterReads[key] = value / 2
	}
	for key, value := range memory.SideTendency {
		out.SideTendency[key] = value * 0.5
	}
	return out
}

func strategyTemplateBase(config *MapConfig, template RouteTemplate) float64 {
	if len(template.ScenarioIDs) == 0 {
		return config.CombatConstants.Float("MinStrategyWeight", 1)
	}
	total, count := 0, 0
	for _, scenarioID := range template.ScenarioIDs {
		if scenario, ok := config.Scenarios[scenarioID]; ok {
			total += scenario.BaseWeight
			count++
		}
	}
	if count == 0 {
		return config.CombatConstants.Float("MinStrategyWeight", 1)
	}
	return float64(total) / float64(count)
}

func lineupFit(team TeamInput, template RouteTemplate) float64 {
	if len(team.Players) == 0 {
		return 0
	}
	weighted, totalWeight := 0.0, 0.0
	keys := make([]string, 0, len(template.KeyAttributes))
	for key := range template.KeyAttributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		weight := template.KeyAttributes[key]
		for _, profile := range team.Players {
			weighted += float64(playerAttribute(profile.Attributes, key)) * weight / float64(len(team.Players))
		}
		totalWeight += weight
	}
	if totalWeight == 0 {
		for _, profile := range team.Players {
			weighted += float64(profile.Attributes.Teamplay+profile.Attributes.Discipline) / 2 / float64(len(team.Players))
		}
		totalWeight = 1
	}
	return weighted / totalWeight * 0.4
}

func scorePressure(template RouteTemplate, scoreFor, scoreAgainst int) float64 {
	delta := float64(scoreAgainst - scoreFor)
	factor := 0.25
	if tempoFactor(template.Tempo) < 1 && delta > 0 {
		factor = 0.5
	}
	return clampFloat(delta*factor, -6, 6)
}

func strategyNoiseAmplitude(team TeamInput, constants CombatConstants) float64 {
	if len(team.Players) == 0 {
		return 0
	}
	discipline, awareness, igl, iglCount := 0, 0, 0, 0
	for _, profile := range team.Players {
		discipline += profile.Attributes.Discipline
		awareness += profile.Attributes.Awareness
		if hasFold(profile.RoleTags, "IGL") {
			igl += profile.Attributes.Gamesense
			iglCount++
		}
	}
	if iglCount == 0 {
		for _, profile := range team.Players {
			igl += profile.Attributes.Gamesense
		}
		iglCount = len(team.Players)
	}
	stability := (float64(discipline)/float64(len(team.Players)) + float64(awareness)/float64(len(team.Players)) + float64(igl)/float64(iglCount)) / 300
	return constants.Float("MaxRandomNoise", 0) * (1 - 0.75*clampProbability(stability))
}

func protectDecisiveStrategyTrend(scores []StrategyScore, decisiveGap float64) {
	if len(scores) < 2 {
		return
	}
	order := append([]StrategyScore(nil), scores...)
	sort.Slice(order, func(i, j int) bool {
		if order[i].DeterministicScore != order[j].DeterministicScore {
			return order[i].DeterministicScore > order[j].DeterministicScore
		}
		return order[i].TemplateID < order[j].TemplateID
	})
	gap := order[0].DeterministicScore - order[1].DeterministicScore
	if gap < decisiveGap {
		return
	}
	bound := math.Max(0, (gap-0.000001)/2)
	for index := range scores {
		scores[index].RandomNoise = clampFloat(scores[index].RandomNoise, -bound, bound)
	}
}

func roleFitScore(profile PlayerProfile, role string) float64 {
	score := float64(roleAttribute(profile.Attributes, role))
	if hasFold(profile.RoleTags, role) {
		score += 100
	}
	return score
}

func roleAttribute(attributes PlayerAttributes, role string) int {
	switch strings.ToLower(role) {
	case "entry":
		return attributes.Entry
	case "support":
		return (attributes.Utility + attributes.Teamplay) / 2
	case "igl":
		return (attributes.Gamesense + attributes.Discipline) / 2
	case "lurker":
		return (attributes.Awareness + attributes.Gamesense) / 2
	case "anchor":
		return (attributes.Positioning + attributes.Discipline) / 2
	case "awper":
		return (attributes.Aim + attributes.Positioning) / 2
	default:
		return (attributes.Aim + attributes.Teamplay) / 2
	}
}

func playerAttribute(attributes PlayerAttributes, name string) int {
	switch strings.ToLower(name) {
	case "aim":
		return attributes.Aim
	case "reaction":
		return attributes.Reaction
	case "gamesense":
		return attributes.Gamesense
	case "positioning":
		return attributes.Positioning
	case "awareness":
		return attributes.Awareness
	case "teamplay":
		return attributes.Teamplay
	case "utility":
		return attributes.Utility
	case "composure":
		return attributes.Composure
	case "mobility":
		return attributes.Mobility
	case "endurance":
		return attributes.Endurance
	case "discipline":
		return attributes.Discipline
	default:
		return 0
	}
}

func recentTemplateCount(recent []string, templateID string, window int) int {
	start := maxInt(0, len(recent)-window)
	count := 0
	for _, id := range recent[start:] {
		if id == templateID {
			count++
		}
	}
	return count
}

func assignOpeningRoutes(target map[string]string, team TeamInput, template RouteTemplate, config *MapConfig) error {
	playerIDs := make([]string, 0, len(team.Players))
	for _, profile := range team.Players {
		playerIDs = append(playerIDs, profile.PlayerID)
	}
	sort.Strings(playerIDs)
	routeIDs := make([]string, 0, len(template.RouteAllocations))
	for routeID := range template.RouteAllocations {
		routeIDs = append(routeIDs, routeID)
	}
	sort.Strings(routeIDs)
	index := 0
	for _, routeID := range routeIDs {
		count := template.RouteAllocations[routeID]
		route, ok := config.Routes[routeID]
		if !ok || route.Side != template.Side || count < route.MinPlayers || count > route.MaxPlayers || index+count > len(playerIDs) {
			return newError("INVALID_OPENING_PLAN", "template %s has invalid route allocation %s", template.ID, routeID)
		}
		for _, playerID := range playerIDs[index : index+count] {
			target[playerID] = routeID
		}
		index += count
	}
	if index != len(playerIDs) {
		return newError("INVALID_OPENING_PLAN", "template %s does not allocate all players", template.ID)
	}
	return nil
}

func startOpeningHold(state *RoundState, playerID, routeID string, ordinal int) (ScheduledAction, error) {
	player := state.Players[playerID]
	intentID := "opening-hold:" + playerID + ":" + routeID
	action := ScheduledAction{IntentID: intentID, Type: ActionHoldStart, ActorIDs: []string{playerID}, From: player.Location, StartAt: 0, ResolveAt: 0, Priority: 30, MinRequiredActors: 1, Payload: ActionPayload{TargetID: player.Location.NodeID}}
	action.ID = NewActionID(state.Seed, action.Type, intentID, 0, 0, action.ActorIDs, ordinal)
	if err := BeginExclusiveAction(state, &action, ActionHolding); err != nil {
		return ScheduledAction{}, err
	}
	player.Intent = Intent{ID: intentID, Type: IntentHold, TargetID: player.Location.NodeID, CreatedAt: 0}
	player.Posture = PostureHolding
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	return action, nil
}

func edgeBetween(edges map[string]MapEdge, from, to string) (string, bool) {
	ids := make([]string, 0, len(edges))
	for id := range edges {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		edge := edges[id]
		if edge.FromNode == from && edge.ToNode == to || edge.Bidirectional && edge.FromNode == to && edge.ToNode == from {
			return id, true
		}
	}
	return "", false
}

func runtimeNodes(state *RoundState) map[string]MapNode {
	out := make(map[string]MapNode, len(state.Nodes))
	for id, node := range state.Nodes {
		out[id] = node.Node
	}
	return out
}

func hasFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func copyStringIntMap(source map[string]int) map[string]int {
	out := map[string]int{}
	for k, v := range source {
		out[k] = v
	}
	return out
}
func copyStringFloatMap(source map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range source {
		out[k] = v
	}
	return out
}
