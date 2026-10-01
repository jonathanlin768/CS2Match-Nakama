package matchengine

import (
	"math"
	"sort"
	"strings"
)

type decisionTriggerType string

const (
	triggerEncounterEnd      decisionTriggerType = "EncounterEnd"
	triggerResourceBand      decisionTriggerType = "ResourceBandChanged"
	triggerControlChanged    decisionTriggerType = "ControlChanged"
	triggerEmptySite         decisionTriggerType = "EmptySite"
	triggerBombChanged       decisionTriggerType = "BombChanged"
	triggerRouteBlocked      decisionTriggerType = "RouteBlocked"
	triggerForceExecute      decisionTriggerType = "ForceExecuteThreshold"
	triggerPostPlantArrival  decisionTriggerType = "PostPlantArrival"
	triggerDefuseInterrupted decisionTriggerType = "DefuseInterrupted"
)

type decisionTrigger struct {
	Type     decisionTriggerType
	SourceID string
	Reason   reasonRecord
}

type decisionFingerprint struct {
	Timeline          int
	AliveBySide       map[string]int
	ResourceBands     map[string]string
	KnownControl      map[string]string
	BombStatus        bombRuntimeStatus
	BombNodeID        string
	ActiveEncounters  int
	EmptySiteIntelIDs []string
}

type decisionType string

const (
	decisionContinue        decisionType = "Continue"
	decisionRotate          decisionType = "Rotate"
	decisionForceExecute    decisionType = "ForceExecute"
	decisionRecoverBomb     decisionType = "RecoverBomb"
	decisionPlant           decisionType = "Plant"
	decisionGatherIntel     decisionType = "GatherIntel"
	decisionHoldFlank       decisionType = "HoldFlank"
	decisionInterceptRotate decisionType = "InterceptRotate"
	decisionReinforce       decisionType = "Reinforce"
	decisionHold            decisionType = "Hold"
	decisionRetake          decisionType = "Retake"
	decisionDefuse          decisionType = "Defuse"
	decisionSave            decisionType = "Save"
	decisionDenyDefuse      decisionType = "DenyDefuse"
)

type decisionCandidate struct {
	Type               decisionType
	Side               string
	ActorIDs           []string
	TargetNode         string
	RouteID            string
	Site               string
	Score              float64
	DeterministicScore float64
	RandomNoise        float64
	Rotation           bool
	Reasons            []reasonRecord
}

type decisionResolution struct {
	Candidate decisionCandidate
	Actions   []scheduledAction
}

func captureDecisionFingerprint(state *roundState) decisionFingerprint {
	fingerprint := decisionFingerprint{
		Timeline: state.Timeline, AliveBySide: map[string]int{SideT: 0, SideCT: 0}, ResourceBands: map[string]string{},
		KnownControl: map[string]string{}, BombStatus: state.Bomb.Status, BombNodeID: projectedNodeID(state.Bomb.Location), ActiveEncounters: len(state.ActiveEngagements),
	}
	playerIDs := make([]string, 0, len(state.Players))
	for playerID := range state.Players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)
	for _, playerID := range playerIDs {
		player := state.Players[playerID]
		if player.Alive {
			fingerprint.AliveBySide[player.Side]++
		}
		fingerprint.ResourceBands[playerID] = resourceBand(player)
	}
	nodeIDs := make([]string, 0, len(state.Nodes))
	for nodeID := range state.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	for _, nodeID := range nodeIDs {
		for _, side := range []string{SideT, SideCT} {
			known := state.Nodes[nodeID].KnownControl[side]
			fingerprint.KnownControl[side+":"+nodeID] = string(known.Status)
		}
	}
	for _, side := range []string{SideT, SideCT} {
		for _, record := range state.Intel[side].Records {
			if record.Type == string(intelEmptySite) && record.ExpiresAt > state.Timeline {
				fingerprint.EmptySiteIntelIDs = append(fingerprint.EmptySiteIntelIDs, record.ID)
			}
		}
	}
	sort.Strings(fingerprint.EmptySiteIntelIDs)
	return fingerprint
}

func detectDecisionTriggers(state *roundState, before decisionFingerprint, explicit []decisionTriggerType) []decisionTrigger {
	after := captureDecisionFingerprint(state)
	triggers := make([]decisionTrigger, 0)
	add := func(triggerType decisionTriggerType, source string) {
		triggers = append(triggers, decisionTrigger{Type: triggerType, SourceID: source, Reason: reasonRecord{Code: string(triggerType), Source: source, Value: 1, Weight: 1}})
	}
	if after.ActiveEncounters < before.ActiveEncounters {
		add(triggerEncounterEnd, "encounter")
	}
	if !equalStringMap(after.ResourceBands, before.ResourceBands) || !equalIntMap(after.AliveBySide, before.AliveBySide) {
		add(triggerResourceBand, "players")
	}
	if !equalStringMap(after.KnownControl, before.KnownControl) {
		add(triggerControlChanged, "control")
	}
	if !equalStrings(after.EmptySiteIntelIDs, before.EmptySiteIntelIDs) {
		add(triggerEmptySite, "intel")
	}
	if after.BombStatus != before.BombStatus || after.BombNodeID != before.BombNodeID {
		add(triggerBombChanged, after.BombNodeID)
	}
	threshold := state.constants.Int("ForceExecuteThreshold", 0)
	if before.Timeline < state.RoundDeadline-threshold && after.Timeline >= state.RoundDeadline-threshold {
		add(triggerForceExecute, "round_timer")
	}
	for _, triggerType := range explicit {
		add(triggerType, "explicit")
	}
	sort.SliceStable(triggers, func(i, j int) bool {
		if triggers[i].Type != triggers[j].Type {
			return triggers[i].Type < triggers[j].Type
		}
		return triggers[i].SourceID < triggers[j].SourceID
	})
	return triggers
}

func scoreDecisionCandidates(view decisionView, routes map[string]Route, constants CombatConstants, rolls rollSource) []decisionCandidate {
	available := availableDecisionPlayers(view)
	var candidates []decisionCandidate
	if view.Side == SideT {
		candidates = append(candidates, tDecisionCandidates(view, available, routes, constants)...)
	} else {
		candidates = append(candidates, ctDecisionCandidates(view, available, routes)...)
	}
	if len(candidates) == 0 && len(available) > 0 {
		candidates = append(candidates, decisionCandidate{Type: decisionHold, Side: view.Side, ActorIDs: []string{available[0].PlayerID}, TargetNode: projectedNodeID(available[0].Location), DeterministicScore: 1})
	}
	amplitude := decisionNoiseAmplitude(view, constants)
	for index := range candidates {
		candidate := &candidates[index]
		candidate.ActorIDs = sortedUnique(candidate.ActorIDs)
		candidate.RandomNoise = (rolls.Unit("decision", view.Side, string(candidate.Type), candidate.RouteID, candidate.TargetNode, joinStrings(candidate.ActorIDs))*2 - 1) * amplitude
	}
	protectDecisionTrend(candidates, constants.Float("DecisiveScoreGap", 0))
	for index := range candidates {
		candidates[index].Score = candidates[index].DeterministicScore + candidates[index].RandomNoise
		candidates[index].Reasons = append(candidates[index].Reasons, reasonRecord{Code: "BOUNDED_RANDOM_NOISE", Source: string(candidates[index].Type), Value: candidates[index].RandomNoise, Weight: 1})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].Type != candidates[j].Type {
			return candidates[i].Type < candidates[j].Type
		}
		return joinStrings(candidates[i].ActorIDs) < joinStrings(candidates[j].ActorIDs)
	})
	return candidates
}

func scheduleDecision(state *roundState, candidate decisionCandidate, ordinal int) (scheduledAction, decisionCandidate, error) {
	atDecisionLimit := state.DecisionCount >= state.constants.Int("MaxDecisionCount", 0)
	if atDecisionLimit && forceableDecision(candidate.Type) {
		candidate = forceReachableCandidate(candidate, state)
	}
	if candidate.Rotation && state.RotationCount[candidate.Side] >= state.constants.Int("MaxRotationsPerTeam", 0) && forceableDecision(candidate.Type) {
		candidate = forceReachableCandidate(candidate, state)
	}
	delay := state.constants.Int("DecisionDelay", 0)
	action := scheduledAction{
		ID:       newActionID(state.Seed, actionDecisionResolve, string(candidate.Type), state.Timeline, state.Timeline+delay, candidate.ActorIDs, ordinal),
		IntentID: "decision:" + string(candidate.Type), Type: actionDecisionResolve, StartAt: state.Timeline, ResolveAt: state.Timeline + delay,
		Priority: 20, Payload: actionPayload{DecisionType: string(candidate.Type), TargetID: candidate.TargetNode, RouteID: candidate.RouteID, Site: candidate.Site, ParticipantIDs: candidate.ActorIDs},
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		return scheduledAction{}, decisionCandidate{}, err
	}
	if !atDecisionLimit {
		state.DecisionCount++
	}
	if candidate.Rotation {
		if err := state.RecordRotation(candidate.Side); err != nil {
			return scheduledAction{}, decisionCandidate{}, err
		}
	}
	return action, candidate, nil
}

func resolveDecision(state *roundState, action scheduledAction, candidate decisionCandidate) (decisionResolution, error) {
	if action.Type != actionDecisionResolve || action.ResolveAt != state.Timeline || action.Payload.DecisionType != string(candidate.Type) {
		return decisionResolution{}, newError("INVALID_DECISION", "decision action does not match candidate")
	}
	resolution := decisionResolution{Candidate: candidate}
	for ordinal, actorID := range candidate.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.EngagementID != "" || player.Action.CurrentActionID != "" {
			continue
		}
		intentID := action.ID + ":" + actorID
		var next scheduledAction
		var err error
		if player.Location.Edge != nil {
			next, err = resumeInterruptedMovement(state, actorID, intentID, ordinal)
			if err != nil {
				return decisionResolution{}, err
			}
			resolution.Actions = append(resolution.Actions, next)
			continue
		}
		switch candidate.Type {
		case decisionPlant:
			next, err = startPlantAction(state, actorID, candidate.Site, intentID, ordinal)
		case decisionDefuse:
			next, err = startDefuseAction(state, actorID, intentID, ordinal)
		case decisionRecoverBomb:
			var schedule bombRecoverySchedule
			var feedback *decisionFeedback
			schedule, feedback, err = scheduleBombRecovery(state, actorID, intentID, ordinal)
			if feedback != nil {
				continue
			}
			next = schedule.Action
		case decisionHold, decisionHoldFlank, decisionSave:
			next, err = startHoldAction(state, actorID, candidate.TargetNode, intentID, ordinal)
		default:
			next, err = startDecisionMove(state, actorID, candidate.TargetNode, candidate.RouteID, intentID, ordinal)
		}
		if err != nil {
			return decisionResolution{}, err
		}
		resolution.Actions = append(resolution.Actions, next)
	}
	return resolution, nil
}

func startHoldAction(state *roundState, actorID, nodeID, intentID string, ordinal int) (scheduledAction, error) {
	player := state.Players[actorID]
	if player == nil || !player.Alive || player.Action.CurrentActionID != "" || player.Location.NodeID != nodeID {
		return scheduledAction{}, newError("INVALID_DECISION", "hold actor is not at target node")
	}
	action := scheduledAction{IntentID: intentID, Type: actionHoldStart, ActorIDs: []string{actorID}, From: player.Location, StartAt: state.Timeline, ResolveAt: state.Timeline, Priority: 30, MinRequiredActors: 1, Payload: actionPayload{TargetID: nodeID}}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionHolding); err != nil {
		return scheduledAction{}, err
	}
	player.Intent = intent{ID: intentID, Type: intentHold, TargetID: nodeID, CreatedAt: state.Timeline}
	player.Posture = postureHolding
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	return action, nil
}

func tDecisionCandidates(view decisionView, players []decisionPlayerView, routes map[string]Route, constants CombatConstants) []decisionCandidate {
	var out []decisionCandidate
	for _, player := range players {
		current := projectedNodeID(player.Location)
		out = append(out, decisionCandidate{Type: decisionContinue, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: resourceExecutionScore(player), Reasons: []reasonRecord{{Code: "CURRENT_RESOURCES", Source: player.PlayerID, Value: resourceExecutionScore(player), Weight: 1}}})
		if player.HasBomb && (current == "A_SITE" || current == "B_SITE") {
			out = append(out, decisionCandidate{Type: decisionPlant, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, Site: strings.TrimSuffix(current, "_SITE"), DeterministicScore: 100})
		}
		if hasFold(player.RoleTags, "Lurker") {
			gatherTarget := bestIntelNode(view)
			if gatherTarget == "" {
				gatherTarget = current
			}
			out = append(out,
				decisionCandidate{Type: decisionGatherIntel, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: gatherTarget, DeterministicScore: 45 + intelGapScore(view)},
				decisionCandidate{Type: decisionHoldFlank, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: 40 + float64(player.Attributes.Awareness)/10},
				decisionCandidate{Type: decisionInterceptRotate, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: gatherTarget, DeterministicScore: 35 + intelConfidenceScore(view), Rotation: true},
			)
		}
	}
	if view.BombStatus == bombDropped && len(players) > 0 {
		out = append(out, decisionCandidate{Type: decisionRecoverBomb, Side: SideT, ActorIDs: []string{players[0].PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 120})
	}
	if (view.BombStatus == bombPlanted || view.BombStatus == bombDefusing) && len(players) > 0 {
		out = append(out, decisionCandidate{Type: decisionDenyDefuse, Side: SideT, ActorIDs: []string{players[0].PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 100})
	}
	if view.BombStatus != bombDropped {
		for _, route := range sortedSideRoutes(routes, SideT) {
			if len(players) == 0 || len(route.Nodes) == 0 {
				continue
			}
			typeValue := decisionRotate
			score := 30 + routeValue(view, route)
			actorIDs := []string{players[0].PlayerID}
			if view.RoundDeadline-view.Timeline <= constants.Int("ForceExecuteThreshold", 0) {
				typeValue, score = decisionForceExecute, 200+routeValue(view, route)
				actorIDs = make([]string, 0, len(players))
				for _, player := range players {
					actorIDs = append(actorIDs, player.PlayerID)
				}
			}
			out = append(out, decisionCandidate{Type: typeValue, Side: SideT, ActorIDs: actorIDs, TargetNode: route.Nodes[len(route.Nodes)-1], RouteID: route.ID, DeterministicScore: score, Rotation: typeValue == decisionRotate})
		}
	}
	return out
}

func ctDecisionCandidates(view decisionView, players []decisionPlayerView, routes map[string]Route) []decisionCandidate {
	var out []decisionCandidate
	for _, player := range players {
		current := projectedNodeID(player.Location)
		out = append(out, decisionCandidate{Type: decisionHold, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: resourceExecutionScore(player)})
		if (view.BombStatus == bombPlanted || view.BombStatus == bombDefusing) && current == view.BombNodeID {
			out = append(out, decisionCandidate{Type: decisionDefuse, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: current, Site: view.BombSite, DeterministicScore: 100 + float64(player.Attributes.Composure)/10})
		}
		if view.BombStatus == bombPlanted && view.BombNodeID != "" && current != view.BombNodeID {
			out = append(out, decisionCandidate{Type: decisionRetake, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 90, Rotation: true})
		}
	}
	intelNode := bestIntelNode(view)
	for _, route := range sortedSideRoutes(routes, SideCT) {
		if len(players) == 0 || len(route.Nodes) == 0 {
			continue
		}
		target := route.Nodes[len(route.Nodes)-1]
		out = append(out, decisionCandidate{Type: decisionReinforce, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: target, RouteID: route.ID, DeterministicScore: 35 + routeValue(view, route), Rotation: true})
		if intelNode != "" {
			out = append(out, decisionCandidate{Type: decisionInterceptRotate, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: intelNode, RouteID: route.ID, DeterministicScore: 30 + intelConfidenceScore(view), Rotation: true, Reasons: []reasonRecord{{Code: "KNOWN_INTEL_ONLY", Source: intelNode, Value: intelConfidenceScore(view), Weight: 1}}})
		}
	}
	if view.BombStatus == bombPlanted && view.BombDeadline-view.Timeline <= 5 && len(players) > 0 {
		out = append(out, decisionCandidate{Type: decisionSave, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: projectedNodeID(players[0].Location), DeterministicScore: 70})
	}
	return out
}

func startDecisionMove(state *roundState, actorID, targetNode, routeID, intentID string, ordinal int) (scheduledAction, error) {
	player := state.Players[actorID]
	if player.Location.NodeID == targetNode {
		return startHoldAction(state, actorID, targetNode, intentID, ordinal)
	}
	path, feedback, err := findBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, player.Location.NodeID, targetNode, len(state.Nodes)*4)
	if err != nil {
		return scheduledAction{}, err
	}
	if feedback != nil || len(path.EdgeIDs) == 0 {
		return scheduledAction{}, newError("DECISION_UNREACHABLE", "decision target %s is unreachable", targetNode)
	}
	tempo := "Default"
	if route := state.routes[routeID]; strings.Contains(strings.ToLower(strings.Join(route.StyleTags, ",")), "fast") {
		tempo = "Fast"
	}
	action, err := startMovement(state, []string{actorID}, path.EdgeIDs[0], moveProfile{Tempo: tempo}, intentID, ordinal)
	if err == nil {
		// Decision movement owns an ultimate target. startMovement records the
		// current edge endpoint, so restore the frozen decision destination for
		// subsequent causal path segments.
		state.Players[actorID].Intent.TargetID = targetNode
	}
	return action, err
}

func forceReachableCandidate(candidate decisionCandidate, state *roundState) decisionCandidate {
	candidate.Rotation = false
	if candidate.Side == SideT {
		candidate = forceExecuteCandidate(state, candidate)
	} else {
		candidate.Type = decisionHold
		if len(candidate.ActorIDs) > 0 {
			candidate.TargetNode = projectedNodeID(state.Players[candidate.ActorIDs[0]].Location)
		}
	}
	candidate.Reasons = append(candidate.Reasons, reasonRecord{Code: "DECISION_LIMIT_FORCE_EXECUTION", Source: candidate.Side, Value: 1, Weight: 1})
	return candidate
}

func forceableDecision(decisionType decisionType) bool {
	switch decisionType {
	case decisionContinue, decisionRotate, decisionForceExecute, decisionGatherIntel, decisionHoldFlank, decisionInterceptRotate, decisionReinforce, decisionHold, decisionRetake, decisionSave:
		return true
	default:
		return false
	}
}

func forceExecuteCandidate(state *roundState, candidate decisionCandidate) decisionCandidate {
	actorIDs := sortedLivePlayerIDs(state, SideT)
	if len(actorIDs) == 0 {
		candidate.Type = decisionForceExecute
		return candidate
	}
	anchorID := actorIDs[0]
	if carrier := state.Players[state.Bomb.CarrierID]; carrier != nil && carrier.Alive && carrier.Side == SideT {
		anchorID = carrier.Profile.PlayerID
	}
	anchor := state.Players[anchorID]
	fromNode := projectedNodeID(anchor.Location)
	bestRouteID, bestTarget, bestDuration := "", "", int(^uint(0)>>1)
	for _, route := range sortedSideRoutes(state.routes, SideT) {
		if route.TargetSite == "" || route.TargetSite == "None" || len(route.Nodes) == 0 {
			continue
		}
		target := route.Nodes[len(route.Nodes)-1]
		path, feedback, err := findBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, fromNode, target, len(state.Nodes)*4)
		if err != nil || feedback != nil {
			continue
		}
		if path.TotalBaseTime < bestDuration || path.TotalBaseTime == bestDuration && route.ID < bestRouteID {
			bestRouteID, bestTarget, bestDuration = route.ID, target, path.TotalBaseTime
		}
	}
	candidate.Type = decisionForceExecute
	candidate.ActorIDs = actorIDs
	if bestRouteID != "" {
		candidate.RouteID = bestRouteID
		candidate.TargetNode = bestTarget
	}
	return candidate
}

func resourceBand(player *roundPlayerState) string {
	minimum := minDamageInt(player.HP, minDamageInt(player.Focus, player.Stamina))
	switch {
	case !player.Alive || player.HP == 0:
		return "dead"
	case minimum < 30:
		return "low"
	case minimum < 60:
		return "medium"
	default:
		return "high"
	}
}

func availableDecisionPlayers(view decisionView) []decisionPlayerView {
	var out []decisionPlayerView
	for _, player := range view.OwnPlayers {
		if player.Alive && player.Action.CurrentActionID == "" {
			out = append(out, player)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PlayerID < out[j].PlayerID })
	return out
}

func resourceExecutionScore(player decisionPlayerView) float64 {
	return float64(player.HP+player.Focus+player.Stamina) / 6
}

func intelGapScore(view decisionView) float64 { return math.Max(0, 20-float64(len(view.Intel))*2) }

func intelConfidenceScore(view decisionView) float64 {
	maxConfidence := 0
	for _, record := range view.Intel {
		maxConfidence = maxInt(maxConfidence, record.Confidence)
	}
	return float64(maxConfidence) / 5
}

func bestIntelNode(view decisionView) string {
	records := append([]intelRecord(nil), view.Intel...)
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Confidence != records[j].Confidence {
			return records[i].Confidence > records[j].Confidence
		}
		return records[i].ID < records[j].ID
	})
	for _, record := range records {
		if record.NodeID != "" {
			return record.NodeID
		}
	}
	return ""
}

func sortedSideRoutes(routes map[string]Route, side string) []Route {
	var out []Route
	for _, route := range routes {
		if route.Side == side {
			out = append(out, route)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func routeValue(view decisionView, route Route) float64 {
	value := float64(100-len(route.Nodes)*5) / 10
	for _, control := range view.KnownControls {
		if len(route.Nodes) > 0 && control.NodeID == route.Nodes[len(route.Nodes)-1] && string(control.Status) == view.Side {
			value += 5
		}
	}
	return value
}

func decisionNoiseAmplitude(view decisionView, constants CombatConstants) float64 {
	if len(view.OwnPlayers) == 0 {
		return 0
	}
	discipline, awareness, igl, iglCount := 0, 0, 0, 0
	for _, player := range view.OwnPlayers {
		discipline += player.Attributes.Discipline
		awareness += player.Attributes.Awareness
		if hasFold(player.RoleTags, "IGL") {
			igl += player.Attributes.Gamesense
			iglCount++
		}
	}
	if iglCount == 0 {
		for _, player := range view.OwnPlayers {
			igl += player.Attributes.Gamesense
		}
		iglCount = len(view.OwnPlayers)
	}
	stability := (float64(discipline)/float64(len(view.OwnPlayers)) + float64(awareness)/float64(len(view.OwnPlayers)) + float64(igl)/float64(iglCount)) / 300
	return constants.Float("MaxRandomNoise", 0) * (1 - 0.75*clampProbability(stability))
}

func protectDecisionTrend(candidates []decisionCandidate, decisiveGap float64) {
	if len(candidates) < 2 {
		return
	}
	order := append([]decisionCandidate(nil), candidates...)
	sort.SliceStable(order, func(i, j int) bool { return order[i].DeterministicScore > order[j].DeterministicScore })
	gap := order[0].DeterministicScore - order[1].DeterministicScore
	if gap < decisiveGap {
		return
	}
	bound := math.Max(0, (gap-0.000001)/2)
	for index := range candidates {
		candidates[index].RandomNoise = clampFloat(candidates[index].RandomNoise, -bound, bound)
	}
}

func sortedUnique(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return uniqueStrings(out)
}

func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func equalIntMap(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
