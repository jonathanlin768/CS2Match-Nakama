package matchengine

import (
	"math"
	"sort"
	"strings"
)

type DecisionTriggerType string

const (
	TriggerEncounterEnd      DecisionTriggerType = "EncounterEnd"
	TriggerResourceBand      DecisionTriggerType = "ResourceBandChanged"
	TriggerControlChanged    DecisionTriggerType = "ControlChanged"
	TriggerEmptySite         DecisionTriggerType = "EmptySite"
	TriggerBombChanged       DecisionTriggerType = "BombChanged"
	TriggerRouteBlocked      DecisionTriggerType = "RouteBlocked"
	TriggerForceExecute      DecisionTriggerType = "ForceExecuteThreshold"
	TriggerPostPlantArrival  DecisionTriggerType = "PostPlantArrival"
	TriggerDefuseInterrupted DecisionTriggerType = "DefuseInterrupted"
)

type DecisionTrigger struct {
	Type     DecisionTriggerType
	SourceID string
	Reason   ReasonRecord
}

type DecisionFingerprint struct {
	Timeline          int
	AliveBySide       map[string]int
	ResourceBands     map[string]string
	KnownControl      map[string]string
	BombStatus        BombRuntimeStatus
	BombNodeID        string
	ActiveEncounters  int
	EmptySiteIntelIDs []string
}

type DecisionType string

const (
	DecisionContinue        DecisionType = "Continue"
	DecisionRotate          DecisionType = "Rotate"
	DecisionForceExecute    DecisionType = "ForceExecute"
	DecisionRecoverBomb     DecisionType = "RecoverBomb"
	DecisionPlant           DecisionType = "Plant"
	DecisionGatherIntel     DecisionType = "GatherIntel"
	DecisionHoldFlank       DecisionType = "HoldFlank"
	DecisionInterceptRotate DecisionType = "InterceptRotate"
	DecisionReinforce       DecisionType = "Reinforce"
	DecisionHold            DecisionType = "Hold"
	DecisionRetake          DecisionType = "Retake"
	DecisionDefuse          DecisionType = "Defuse"
	DecisionSave            DecisionType = "Save"
	DecisionDenyDefuse      DecisionType = "DenyDefuse"
)

type DecisionCandidate struct {
	Type               DecisionType
	Side               string
	ActorIDs           []string
	TargetNode         string
	RouteID            string
	Site               string
	Score              float64
	DeterministicScore float64
	RandomNoise        float64
	Rotation           bool
	Reasons            []ReasonRecord
}

type DecisionResolution struct {
	Candidate DecisionCandidate
	Actions   []ScheduledAction
}

func CaptureDecisionFingerprint(state *RoundState) DecisionFingerprint {
	fingerprint := DecisionFingerprint{
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
			if record.Type == string(IntelEmptySite) && record.ExpiresAt > state.Timeline {
				fingerprint.EmptySiteIntelIDs = append(fingerprint.EmptySiteIntelIDs, record.ID)
			}
		}
	}
	sort.Strings(fingerprint.EmptySiteIntelIDs)
	return fingerprint
}

func DetectDecisionTriggers(state *RoundState, before DecisionFingerprint, explicit []DecisionTriggerType) []DecisionTrigger {
	after := CaptureDecisionFingerprint(state)
	triggers := make([]DecisionTrigger, 0)
	add := func(triggerType DecisionTriggerType, source string) {
		triggers = append(triggers, DecisionTrigger{Type: triggerType, SourceID: source, Reason: ReasonRecord{Code: string(triggerType), Source: source, Value: 1, Weight: 1}})
	}
	if after.ActiveEncounters < before.ActiveEncounters {
		add(TriggerEncounterEnd, "encounter")
	}
	if !equalStringMap(after.ResourceBands, before.ResourceBands) || !equalIntMap(after.AliveBySide, before.AliveBySide) {
		add(TriggerResourceBand, "players")
	}
	if !equalStringMap(after.KnownControl, before.KnownControl) {
		add(TriggerControlChanged, "control")
	}
	if !equalStrings(after.EmptySiteIntelIDs, before.EmptySiteIntelIDs) {
		add(TriggerEmptySite, "intel")
	}
	if after.BombStatus != before.BombStatus || after.BombNodeID != before.BombNodeID {
		add(TriggerBombChanged, after.BombNodeID)
	}
	threshold := state.constants.Int("ForceExecuteThreshold", 0)
	if before.Timeline < state.RoundDeadline-threshold && after.Timeline >= state.RoundDeadline-threshold {
		add(TriggerForceExecute, "round_timer")
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

func ScoreDecisionCandidates(view DecisionView, routes map[string]Route, constants CombatConstants, rolls RollSource) []DecisionCandidate {
	available := availableDecisionPlayers(view)
	var candidates []DecisionCandidate
	if view.Side == SideT {
		candidates = append(candidates, tDecisionCandidates(view, available, routes, constants)...)
	} else {
		candidates = append(candidates, ctDecisionCandidates(view, available, routes)...)
	}
	if len(candidates) == 0 && len(available) > 0 {
		candidates = append(candidates, DecisionCandidate{Type: DecisionHold, Side: view.Side, ActorIDs: []string{available[0].PlayerID}, TargetNode: projectedNodeID(available[0].Location), DeterministicScore: 1})
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
		candidates[index].Reasons = append(candidates[index].Reasons, ReasonRecord{Code: "BOUNDED_RANDOM_NOISE", Source: string(candidates[index].Type), Value: candidates[index].RandomNoise, Weight: 1})
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

func ScheduleDecision(state *RoundState, candidate DecisionCandidate, ordinal int) (ScheduledAction, DecisionCandidate, error) {
	atDecisionLimit := state.DecisionCount >= state.constants.Int("MaxDecisionCount", 0)
	if atDecisionLimit && forceableDecision(candidate.Type) {
		candidate = forceReachableCandidate(candidate, state)
	}
	if candidate.Rotation && state.RotationCount[candidate.Side] >= state.constants.Int("MaxRotationsPerTeam", 0) && forceableDecision(candidate.Type) {
		candidate = forceReachableCandidate(candidate, state)
	}
	delay := state.constants.Int("DecisionDelay", 0)
	action := ScheduledAction{
		ID:       NewActionID(state.Seed, ActionDecisionResolve, string(candidate.Type), state.Timeline, state.Timeline+delay, candidate.ActorIDs, ordinal),
		IntentID: "decision:" + string(candidate.Type), Type: ActionDecisionResolve, StartAt: state.Timeline, ResolveAt: state.Timeline + delay,
		Priority: 20, Payload: ActionPayload{DecisionType: string(candidate.Type), TargetID: candidate.TargetNode, RouteID: candidate.RouteID, Site: candidate.Site, ParticipantIDs: candidate.ActorIDs},
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		return ScheduledAction{}, DecisionCandidate{}, err
	}
	if !atDecisionLimit {
		state.DecisionCount++
	}
	if candidate.Rotation {
		if err := state.RecordRotation(candidate.Side); err != nil {
			return ScheduledAction{}, DecisionCandidate{}, err
		}
	}
	return action, candidate, nil
}

func ResolveDecision(state *RoundState, action ScheduledAction, candidate DecisionCandidate) (DecisionResolution, error) {
	if action.Type != ActionDecisionResolve || action.ResolveAt != state.Timeline || action.Payload.DecisionType != string(candidate.Type) {
		return DecisionResolution{}, newError("INVALID_DECISION", "decision action does not match candidate")
	}
	resolution := DecisionResolution{Candidate: candidate}
	for ordinal, actorID := range candidate.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.EngagementID != "" || player.Action.CurrentActionID != "" {
			continue
		}
		intentID := action.ID + ":" + actorID
		var next ScheduledAction
		var err error
		if player.Location.Edge != nil {
			next, err = ResumeInterruptedMovement(state, actorID, intentID, ordinal)
			if err != nil {
				return DecisionResolution{}, err
			}
			resolution.Actions = append(resolution.Actions, next)
			continue
		}
		switch candidate.Type {
		case DecisionPlant:
			next, err = StartPlantAction(state, actorID, candidate.Site, intentID, ordinal)
		case DecisionDefuse:
			next, err = StartDefuseAction(state, actorID, intentID, ordinal)
		case DecisionRecoverBomb:
			var schedule BombRecoverySchedule
			var feedback *DecisionFeedback
			schedule, feedback, err = ScheduleBombRecovery(state, actorID, intentID, ordinal)
			if feedback != nil {
				continue
			}
			next = schedule.Action
		case DecisionHold, DecisionHoldFlank, DecisionSave:
			next, err = StartHoldAction(state, actorID, candidate.TargetNode, intentID, ordinal)
		default:
			next, err = startDecisionMove(state, actorID, candidate.TargetNode, candidate.RouteID, intentID, ordinal)
		}
		if err != nil {
			return DecisionResolution{}, err
		}
		resolution.Actions = append(resolution.Actions, next)
	}
	return resolution, nil
}

func StartHoldAction(state *RoundState, actorID, nodeID, intentID string, ordinal int) (ScheduledAction, error) {
	player := state.Players[actorID]
	if player == nil || !player.Alive || player.Action.CurrentActionID != "" || player.Location.NodeID != nodeID {
		return ScheduledAction{}, newError("INVALID_DECISION", "hold actor is not at target node")
	}
	action := ScheduledAction{IntentID: intentID, Type: ActionHoldStart, ActorIDs: []string{actorID}, From: player.Location, StartAt: state.Timeline, ResolveAt: state.Timeline, Priority: 30, MinRequiredActors: 1, Payload: ActionPayload{TargetID: nodeID}}
	action.ID = NewActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := BeginExclusiveAction(state, &action, ActionHolding); err != nil {
		return ScheduledAction{}, err
	}
	player.Intent = Intent{ID: intentID, Type: IntentHold, TargetID: nodeID, CreatedAt: state.Timeline}
	player.Posture = PostureHolding
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	return action, nil
}

func tDecisionCandidates(view DecisionView, players []DecisionPlayerView, routes map[string]Route, constants CombatConstants) []DecisionCandidate {
	var out []DecisionCandidate
	for _, player := range players {
		current := projectedNodeID(player.Location)
		out = append(out, DecisionCandidate{Type: DecisionContinue, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: resourceExecutionScore(player), Reasons: []ReasonRecord{{Code: "CURRENT_RESOURCES", Source: player.PlayerID, Value: resourceExecutionScore(player), Weight: 1}}})
		if player.HasBomb && (current == "A_SITE" || current == "B_SITE") {
			out = append(out, DecisionCandidate{Type: DecisionPlant, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, Site: strings.TrimSuffix(current, "_SITE"), DeterministicScore: 100})
		}
		if hasFold(player.RoleTags, "Lurker") {
			gatherTarget := bestIntelNode(view)
			if gatherTarget == "" {
				gatherTarget = current
			}
			out = append(out,
				DecisionCandidate{Type: DecisionGatherIntel, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: gatherTarget, DeterministicScore: 45 + intelGapScore(view)},
				DecisionCandidate{Type: DecisionHoldFlank, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: 40 + float64(player.Attributes.Awareness)/10},
				DecisionCandidate{Type: DecisionInterceptRotate, Side: SideT, ActorIDs: []string{player.PlayerID}, TargetNode: gatherTarget, DeterministicScore: 35 + intelConfidenceScore(view), Rotation: true},
			)
		}
	}
	if view.BombStatus == BombDropped && len(players) > 0 {
		out = append(out, DecisionCandidate{Type: DecisionRecoverBomb, Side: SideT, ActorIDs: []string{players[0].PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 120})
	}
	if (view.BombStatus == BombPlanted || view.BombStatus == BombDefusing) && len(players) > 0 {
		out = append(out, DecisionCandidate{Type: DecisionDenyDefuse, Side: SideT, ActorIDs: []string{players[0].PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 100})
	}
	if view.BombStatus != BombDropped {
		for _, route := range sortedSideRoutes(routes, SideT) {
			if len(players) == 0 || len(route.Nodes) == 0 {
				continue
			}
			typeValue := DecisionRotate
			score := 30 + routeValue(view, route)
			actorIDs := []string{players[0].PlayerID}
			if view.RoundDeadline-view.Timeline <= constants.Int("ForceExecuteThreshold", 0) {
				typeValue, score = DecisionForceExecute, 200+routeValue(view, route)
				actorIDs = make([]string, 0, len(players))
				for _, player := range players {
					actorIDs = append(actorIDs, player.PlayerID)
				}
			}
			out = append(out, DecisionCandidate{Type: typeValue, Side: SideT, ActorIDs: actorIDs, TargetNode: route.Nodes[len(route.Nodes)-1], RouteID: route.ID, DeterministicScore: score, Rotation: typeValue == DecisionRotate})
		}
	}
	return out
}

func ctDecisionCandidates(view DecisionView, players []DecisionPlayerView, routes map[string]Route) []DecisionCandidate {
	var out []DecisionCandidate
	for _, player := range players {
		current := projectedNodeID(player.Location)
		out = append(out, DecisionCandidate{Type: DecisionHold, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: current, DeterministicScore: resourceExecutionScore(player)})
		if (view.BombStatus == BombPlanted || view.BombStatus == BombDefusing) && current == view.BombNodeID {
			out = append(out, DecisionCandidate{Type: DecisionDefuse, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: current, Site: view.BombSite, DeterministicScore: 100 + float64(player.Attributes.Composure)/10})
		}
		if view.BombStatus == BombPlanted && view.BombNodeID != "" && current != view.BombNodeID {
			out = append(out, DecisionCandidate{Type: DecisionRetake, Side: SideCT, ActorIDs: []string{player.PlayerID}, TargetNode: view.BombNodeID, DeterministicScore: 90, Rotation: true})
		}
	}
	intelNode := bestIntelNode(view)
	for _, route := range sortedSideRoutes(routes, SideCT) {
		if len(players) == 0 || len(route.Nodes) == 0 {
			continue
		}
		target := route.Nodes[len(route.Nodes)-1]
		out = append(out, DecisionCandidate{Type: DecisionReinforce, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: target, RouteID: route.ID, DeterministicScore: 35 + routeValue(view, route), Rotation: true})
		if intelNode != "" {
			out = append(out, DecisionCandidate{Type: DecisionInterceptRotate, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: intelNode, RouteID: route.ID, DeterministicScore: 30 + intelConfidenceScore(view), Rotation: true, Reasons: []ReasonRecord{{Code: "KNOWN_INTEL_ONLY", Source: intelNode, Value: intelConfidenceScore(view), Weight: 1}}})
		}
	}
	if view.BombStatus == BombPlanted && view.BombDeadline-view.Timeline <= 5 && len(players) > 0 {
		out = append(out, DecisionCandidate{Type: DecisionSave, Side: SideCT, ActorIDs: []string{players[0].PlayerID}, TargetNode: projectedNodeID(players[0].Location), DeterministicScore: 70})
	}
	return out
}

func startDecisionMove(state *RoundState, actorID, targetNode, routeID, intentID string, ordinal int) (ScheduledAction, error) {
	player := state.Players[actorID]
	if player.Location.NodeID == targetNode {
		return StartHoldAction(state, actorID, targetNode, intentID, ordinal)
	}
	path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, player.Location.NodeID, targetNode, len(state.Nodes)*4)
	if err != nil {
		return ScheduledAction{}, err
	}
	if feedback != nil || len(path.EdgeIDs) == 0 {
		return ScheduledAction{}, newError("DECISION_UNREACHABLE", "decision target %s is unreachable", targetNode)
	}
	tempo := "Default"
	if route := state.routes[routeID]; strings.Contains(strings.ToLower(strings.Join(route.StyleTags, ",")), "fast") {
		tempo = "Fast"
	}
	action, err := StartMovement(state, []string{actorID}, path.EdgeIDs[0], MoveProfile{Tempo: tempo}, intentID, ordinal)
	if err == nil {
		// Decision movement owns an ultimate target. StartMovement records the
		// current edge endpoint, so restore the frozen decision destination for
		// subsequent causal path segments.
		state.Players[actorID].Intent.TargetID = targetNode
	}
	return action, err
}

func forceReachableCandidate(candidate DecisionCandidate, state *RoundState) DecisionCandidate {
	candidate.Rotation = false
	if candidate.Side == SideT {
		candidate = forceExecuteCandidate(state, candidate)
	} else {
		candidate.Type = DecisionHold
		if len(candidate.ActorIDs) > 0 {
			candidate.TargetNode = projectedNodeID(state.Players[candidate.ActorIDs[0]].Location)
		}
	}
	candidate.Reasons = append(candidate.Reasons, ReasonRecord{Code: "DECISION_LIMIT_FORCE_EXECUTION", Source: candidate.Side, Value: 1, Weight: 1})
	return candidate
}

func forceableDecision(decisionType DecisionType) bool {
	switch decisionType {
	case DecisionContinue, DecisionRotate, DecisionForceExecute, DecisionGatherIntel, DecisionHoldFlank, DecisionInterceptRotate, DecisionReinforce, DecisionHold, DecisionRetake, DecisionSave:
		return true
	default:
		return false
	}
}

func forceExecuteCandidate(state *RoundState, candidate DecisionCandidate) DecisionCandidate {
	actorIDs := sortedLivePlayerIDs(state, SideT)
	if len(actorIDs) == 0 {
		candidate.Type = DecisionForceExecute
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
		path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, fromNode, target, len(state.Nodes)*4)
		if err != nil || feedback != nil {
			continue
		}
		if path.TotalBaseTime < bestDuration || path.TotalBaseTime == bestDuration && route.ID < bestRouteID {
			bestRouteID, bestTarget, bestDuration = route.ID, target, path.TotalBaseTime
		}
	}
	candidate.Type = DecisionForceExecute
	candidate.ActorIDs = actorIDs
	if bestRouteID != "" {
		candidate.RouteID = bestRouteID
		candidate.TargetNode = bestTarget
	}
	return candidate
}

func resourceBand(player *RoundPlayerState) string {
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

func availableDecisionPlayers(view DecisionView) []DecisionPlayerView {
	var out []DecisionPlayerView
	for _, player := range view.OwnPlayers {
		if player.Alive && player.Action.CurrentActionID == "" {
			out = append(out, player)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PlayerID < out[j].PlayerID })
	return out
}

func resourceExecutionScore(player DecisionPlayerView) float64 {
	return float64(player.HP+player.Focus+player.Stamina) / 6
}

func intelGapScore(view DecisionView) float64 { return math.Max(0, 20-float64(len(view.Intel))*2) }

func intelConfidenceScore(view DecisionView) float64 {
	maxConfidence := 0
	for _, record := range view.Intel {
		maxConfidence = maxInt(maxConfidence, record.Confidence)
	}
	return float64(maxConfidence) / 5
}

func bestIntelNode(view DecisionView) string {
	records := append([]IntelRecord(nil), view.Intel...)
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

func routeValue(view DecisionView, route Route) float64 {
	value := float64(100-len(route.Nodes)*5) / 10
	for _, control := range view.KnownControls {
		if len(route.Nodes) > 0 && control.NodeID == route.Nodes[len(route.Nodes)-1] && string(control.Status) == view.Side {
			value += 5
		}
	}
	return value
}

func decisionNoiseAmplitude(view DecisionView, constants CombatConstants) float64 {
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

func protectDecisionTrend(candidates []DecisionCandidate, decisiveGap float64) {
	if len(candidates) < 2 {
		return
	}
	order := append([]DecisionCandidate(nil), candidates...)
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
