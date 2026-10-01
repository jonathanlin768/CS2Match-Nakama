package matchengine

import (
	"math"
	"sort"
)

const (
	priorityPlantComplete  = 70
	priorityDefuseComplete = 60
	priorityBombExplode    = 50
)

type siteContestDecisionType string

const (
	siteContestEncounter siteContestDecisionType = "Encounter"
	siteContestPlant     siteContestDecisionType = "Plant"
	siteContestWithdraw  siteContestDecisionType = "Withdraw"
)

type siteContestDecision struct {
	Type       siteContestDecisionType
	Site       string
	ActorIDs   []string
	PlantScore float64
	PlantRisk  float64
	Reasons    []reasonRecord
}

type bombRecoverySchedule struct {
	Action         scheduledAction
	Path           pathResult
	MoveDuration   int
	PickupDuration int
}

type bombActionResult struct {
	Applied  bool
	Event    *GameEvent
	FollowUp *scheduledAction
}

func planSiteContest(state *roundState, site string, actorIDs []string) (siteContestDecision, error) {
	nodeID := plantNodeForSite(state, site)
	if nodeID == "" {
		return siteContestDecision{}, newError("INVALID_PLANT", "site %s has no plant node", site)
	}
	actors := liveActorsAtNode(state, actorIDs, nodeID)
	carrier := state.Players[state.Bomb.CarrierID]
	if carrier == nil || !carrier.Alive || carrier.Side != SideT || carrier.Location.NodeID != nodeID {
		return siteContestDecision{Type: siteContestWithdraw, Site: site, ActorIDs: actors, Reasons: []reasonRecord{{Code: "NO_CARRIER_AT_SITE", Source: site, Value: -1, Weight: 1}}}, nil
	}
	node := state.Nodes[nodeID]
	threats := visibleThreatsAtSite(state, nodeID, SideCT)
	plantScore, plantRisk := calculatePlantScore(state, carrier.Profile.PlayerID, site, threats)
	if node.ActualControl == controlCT || node.ActualControl == controlContested || len(threats) > 0 {
		return siteContestDecision{Type: siteContestEncounter, Site: site, ActorIDs: append(actors, threats...), PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []reasonRecord{{Code: "VISIBLE_SITE_THREAT", Source: nodeID, Value: -plantRisk, Weight: 1}}}, nil
	}
	if plantScore >= plantRisk {
		return siteContestDecision{Type: siteContestPlant, Site: site, ActorIDs: actors, PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []reasonRecord{{Code: "PLANT_WINDOW", Source: nodeID, Value: plantScore - plantRisk, Weight: 1}}}, nil
	}
	return siteContestDecision{Type: siteContestWithdraw, Site: site, ActorIDs: actors, PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []reasonRecord{{Code: "PLANT_RISK", Source: nodeID, Value: plantScore - plantRisk, Weight: 1}}}, nil
}

func calculatePlantScore(state *roundState, actorID, site string, visibleThreatIDs []string) (float64, float64) {
	actor := state.Players[actorID]
	if actor == nil {
		return 0, math.Inf(1)
	}
	cover := scopedUtilityModifier(state, SideT, utilityPlantCover, "", "site:"+site, state.Timeline)
	score := float64(actor.Profile.Attributes.Composure+actor.Profile.Attributes.Discipline+actor.Focus)/3 + cover*20
	risk := float64(len(visibleThreatIDs))*25 + playerExposure(actor)
	if node := state.Nodes[actor.Location.NodeID]; node != nil && node.ActualControl == controlCT {
		risk += 30
	}
	return score, risk
}

func canAttemptPlant(state *roundState, actorID, site string) error {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideT || !actor.HasBomb || state.Bomb.CarrierID != actorID || state.Bomb.Status != bombCarried {
		return newError("INVALID_PLANT", "actor is not the live unique bomb carrier")
	}
	if actor.EngagementID != "" || actor.Action.CurrentActionID != "" {
		return newError("INVALID_PLANT", "bomb carrier is busy or engaged")
	}
	if state.Timeline > state.RoundDeadline {
		return newError("INVALID_PLANT", "plant starts after RoundDeadline")
	}
	node := state.Nodes[actor.Location.NodeID]
	if node == nil || node.Node.Site != site || !hasString(node.Node.AreaUsages, "Plant") {
		return newError("INVALID_PLANT", "bomb carrier is not inside the configured plant site")
	}
	if node.ActualControl == controlCT || node.ActualControl == controlContested {
		return newError("SITE_CONTEST_REQUIRED", "plant site must be contested before planting")
	}
	return nil
}

func startPlantAction(state *roundState, actorID, site, intentID string, ordinal int) (scheduledAction, error) {
	if err := canAttemptPlant(state, actorID, site); err != nil {
		return scheduledAction{}, err
	}
	actor := state.Players[actorID]
	utilityResult, err := spendScopedUtility(state, SideT, utilityPlantCover, []string{actorID}, "", "site:"+site, state.constants.Int("BasePlantTime", 1), 10)
	if err != nil {
		return scheduledAction{}, err
	}
	cover := scopedUtilityModifier(state, SideT, utilityPlantCover, "", "site:"+site, state.Timeline)
	duration := int(math.Round(float64(state.constants.Int("BasePlantTime", 1)) - cover - float64(actor.Profile.Attributes.Discipline-50)/100))
	duration = clampInt(duration, state.constants.Int("MinPlantTime", 1), state.constants.Int("MaxPlantTime", 1))
	action := scheduledAction{IntentID: intentID, Type: actionPlantComplete, ActorIDs: []string{actorID}, From: actor.Location, StartAt: state.Timeline, ResolveAt: state.Timeline + duration, Priority: priorityPlantComplete, MinRequiredActors: 1, Payload: actionPayload{Site: site}}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionPlanting); err != nil {
		return scheduledAction{}, err
	}
	actor.Intent = intent{ID: intentID, Type: intentPlant, TargetID: site, CreatedAt: state.Timeline}
	if err := state.Bomb.StartPlant(actorID, action.ID, site, action.StartAt, action.ResolveAt); err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		state.Bomb.InterruptPlant(actorID, action.ID)
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	event, _ := newActionLifecycleEvent(state, action, EventPlantStart, "bomb plant started", 0)
	addUtilityReason(event, utilityResult)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return action, nil
}

func resolvePlantComplete(state *roundState, action scheduledAction) (bombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Timeline > state.RoundDeadline {
		return bombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	if state.Bomb.Status != bombPlanting || state.Bomb.PlantActionID != action.ID || !sameLocation(actor.Location, action.From) {
		return bombActionResult{}, nil
	}
	explodeAt := state.Timeline + state.constants.Int("BombExplodeTime", 1)
	if err := state.Bomb.CompletePlant(actor.Location, state.Timeline, explodeAt); err != nil {
		return bombActionResult{}, err
	}
	actor.HasBomb = false
	completeActionForActors(state, action, valid)
	state.BombDeadline = explodeAt
	state.Phase = phasePostPlant
	explode := scheduledAction{ID: newActionID(state.Seed, actionBombExplode, action.ID, state.Timeline, explodeAt, nil, 0), IntentID: action.ID, Type: actionBombExplode, StartAt: state.Timeline, ResolveAt: explodeAt, Priority: priorityBombExplode, Payload: actionPayload{Site: state.Bomb.PlantedSite}}
	if err := state.Scheduler.Schedule(explode); err != nil {
		return bombActionResult{}, err
	}
	event := bombLifecycleEffectEvent(state, action, EventBombPlant, "bomb planted", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return bombActionResult{Applied: true, Event: event, FollowUp: &explode}, nil
}

func scheduleBombRecovery(state *roundState, actorID, intentID string, ordinal int) (bombRecoverySchedule, *decisionFeedback, error) {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideT || actor.Action.CurrentActionID != "" || state.Bomb.Status != bombDropped {
		return bombRecoverySchedule{}, nil, newError("INVALID_PICKUP", "bomb recovery actor is unavailable")
	}
	targetNode := bombRecoveryNode(state.Bomb.Location)
	if targetNode == "" {
		return bombRecoverySchedule{}, nil, newError("INVALID_PICKUP", "dropped bomb has no semantic location")
	}
	if node := state.Nodes[targetNode]; node != nil && (node.ActualControl == controlCT || node.ActualControl == controlContested) {
		return bombRecoverySchedule{}, &decisionFeedback{Code: "SITE_CONTEST_REQUIRED", Message: "bomb location is enemy-controlled or contested"}, nil
	}
	path, feedback, err := findBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, actor.Location.NodeID, targetNode, len(state.Nodes)*4)
	if err != nil || feedback != nil {
		return bombRecoverySchedule{}, feedback, err
	}
	pickupDuration := clampInt(state.constants.Int("BasePickupTime", 1), state.constants.Int("MinPickupTime", 1), state.constants.Int("MaxPickupTime", 1))
	moveDuration := path.TotalBaseTime
	resolveAt := state.Timeline + moveDuration + pickupDuration
	action := scheduledAction{IntentID: intentID, Type: actionPickupComplete, ActorIDs: []string{actorID}, From: actor.Location, ToNodeID: targetNode, StartAt: state.Timeline, ResolveAt: resolveAt, Priority: priorityPlantComplete, MinRequiredActors: 1, Payload: actionPayload{TargetID: targetNode}}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionMoving); err != nil {
		return bombRecoverySchedule{}, nil, err
	}
	actor.Intent = intent{ID: intentID, Type: intentPickupBomb, TargetID: targetNode, CreatedAt: state.Timeline}
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return bombRecoverySchedule{}, nil, err
	}
	return bombRecoverySchedule{Action: action, Path: path, MoveDuration: moveDuration, PickupDuration: pickupDuration}, nil, nil
}

func resolveBombPickup(state *roundState, action scheduledAction) (bombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Bomb.Status != bombDropped {
		return bombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	actor.Location = clonePlayerLocation(state.Bomb.Location)
	if err := state.Bomb.Pickup(actor.Profile.PlayerID, actor.Location, state.Timeline); err != nil {
		return bombActionResult{}, err
	}
	actor.HasBomb = true
	completeActionForActors(state, action, valid)
	event := bombLifecycleEffectEvent(state, action, EventBombPickup, "bomb picked up", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return bombActionResult{Applied: true, Event: event}, nil
}

func calculateDefuseTime(state *roundState, actorID string) int {
	actor := state.Players[actorID]
	duration := float64(state.constants.Int("BaseDefuseTime", 1))
	if actor != nil && actor.Weapon.HasKit {
		duration *= 0.5
	}
	cover := scopedUtilityModifier(state, SideCT, utilityDefuseCover, "", "site:"+state.Bomb.PlantedSite, state.Timeline)
	duration -= cover
	return clampInt(int(math.Round(duration)), state.constants.Int("MinDefuseTime", 1), state.constants.Int("MaxDefuseTime", 1))
}

func calculateDefuseScore(state *roundState, actorID string) float64 {
	actor := state.Players[actorID]
	if actor == nil {
		return math.Inf(-1)
	}
	score := float64(actor.Focus+actor.Profile.Attributes.Composure+actor.Profile.Attributes.Discipline) / 3
	if actor.Weapon.HasKit {
		score += 20
	}
	if canDenyDefuse(state, actor.Location.NodeID, state.Timeline+calculateDefuseTime(state, actorID)) {
		score -= 30
	}
	return score
}

func canAttemptDefuse(state *roundState, actorID string) error {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideCT || actor.EngagementID != "" || actor.Action.CurrentActionID != "" {
		return newError("INVALID_DEFUSE", "defuser is unavailable")
	}
	if state.Bomb.Status != bombPlanted || actor.Location.NodeID == "" || actor.Location.NodeID != state.Bomb.Location.NodeID {
		return newError("INVALID_DEFUSE", "defuser has not reached the planted bomb")
	}
	if state.Timeline+calculateDefuseTime(state, actorID) > state.BombDeadline {
		return newError("INVALID_DEFUSE", "defuse cannot finish before bomb deadline")
	}
	if canDenyDefuse(state, actor.Location.NodeID, state.Timeline+calculateDefuseTime(state, actorID)) {
		return newError("DEFUSE_CONTEST_REQUIRED", "a live T can deny the defuse before completion")
	}
	return nil
}

func startDefuseAction(state *roundState, actorID, intentID string, ordinal int) (scheduledAction, error) {
	if err := canAttemptDefuse(state, actorID); err != nil {
		return scheduledAction{}, err
	}
	actor := state.Players[actorID]
	utilityResult, err := spendScopedUtility(state, SideCT, utilityDefuseCover, []string{actorID}, "", "site:"+state.Bomb.PlantedSite, state.constants.Int("BaseDefuseTime", 1), 10)
	if err != nil {
		return scheduledAction{}, err
	}
	duration := calculateDefuseTime(state, actorID)
	action := scheduledAction{IntentID: intentID, Type: actionDefuseComplete, ActorIDs: []string{actorID}, From: actor.Location, StartAt: state.Timeline, ResolveAt: state.Timeline + duration, Priority: priorityDefuseComplete, MinRequiredActors: 1, Payload: actionPayload{Site: state.Bomb.PlantedSite}}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionDefusing); err != nil {
		return scheduledAction{}, err
	}
	actor.Intent = intent{ID: intentID, Type: intentDefuse, TargetID: state.Bomb.PlantedSite, CreatedAt: state.Timeline}
	if err := state.Bomb.StartDefuse(actorID, action.ID, action.StartAt, action.ResolveAt); err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		state.Bomb.InterruptDefuse(actorID, action.ID)
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	event, _ := newActionLifecycleEvent(state, action, EventDefuseStart, "bomb defuse started", 0)
	addUtilityReason(event, utilityResult)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return action, nil
}

func resolveDefuseComplete(state *roundState, action scheduledAction) (bombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Timeline > state.BombDeadline {
		return bombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	if !actor.Alive || state.Bomb.Status != bombDefusing || state.Bomb.DefuseActionID != action.ID || !sameLocation(actor.Location, action.From) {
		return bombActionResult{}, nil
	}
	if err := state.Bomb.CompleteDefuse(state.Timeline); err != nil {
		return bombActionResult{}, err
	}
	completeActionForActors(state, action, valid)
	event := bombLifecycleEffectEvent(state, action, EventBombDefuse, "bomb defused", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return bombActionResult{Applied: true, Event: event}, nil
}

func resolveBombExplode(state *roundState, action scheduledAction) (bombActionResult, error) {
	if state.Timeline != action.ResolveAt || state.Bomb.Status == bombDefused || state.Bomb.ExplodeAt != action.ResolveAt {
		return bombActionResult{}, nil
	}
	if err := state.Bomb.Explode(state.Timeline); err != nil {
		return bombActionResult{}, err
	}
	event := bombLifecycleEffectEvent(state, action, EventBombExplode, "bomb exploded", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return bombActionResult{Applied: true, Event: event}, nil
}

func canDenyDefuse(state *roundState, bombNode string, finishAt int) bool {
	for _, player := range state.Players {
		if !player.Alive || player.Side != SideT {
			continue
		}
		from := player.Location.NodeID
		if from == "" {
			from = projectedNodeID(player.Location)
		}
		path, feedback, err := findBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, from, bombNode, len(state.Nodes)*4)
		if err == nil && feedback == nil && state.Timeline+path.TotalBaseTime <= finishAt {
			return true
		}
	}
	return false
}

func bombLifecycleEffectEvent(state *roundState, action scheduledAction, eventType, message string, ordinal int) *GameEvent {
	effectID := newEffectID(state.Seed, action.ID, effectBombState, ordinal)
	eventID := newEventID(state.Seed, action.ID, effectID, eventType, ordinal)
	reason, _ := projectReasonRecord(reasonRecord{Code: eventType, Source: string(action.Type), Value: 1, Weight: 1}, action.ID, effectID)
	event := &GameEvent{EventID: eventID, SourceActionID: action.ID, SourceEffectID: effectID, Timestamp: int64(state.Timeline), EventType: eventType, Message: message, Reason: reason, sortPriority: action.Priority, sortActionType: string(action.Type), sortMinActorID: action.MinActorID()}
	if len(action.ActorIDs) > 0 {
		if actor := state.Players[action.ActorIDs[0]]; actor != nil {
			event.AttackerID, event.AttackerName, event.AttackerTeamID = actor.Profile.PlayerID, actor.Profile.DisplayName, actor.TeamID
			event.Location = eventLocation(state, actor.Location, eventID, effectID)
		}
	} else {
		event.Location = eventLocation(state, state.Bomb.Location, eventID, effectID)
	}
	return event
}

func plantNodeForSite(state *roundState, site string) string {
	ids := make([]string, 0, len(state.Nodes))
	for id := range state.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		node := state.Nodes[id].Node
		if node.Site == site && hasString(node.AreaUsages, "Plant") {
			return id
		}
	}
	return ""
}

func liveActorsAtNode(state *roundState, actorIDs []string, nodeID string) []string {
	var out []string
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player != nil && player.Alive && player.Location.NodeID == nodeID {
			out = append(out, actorID)
		}
	}
	sort.Strings(out)
	return uniqueStrings(out)
}

func visibleThreatsAtSite(state *roundState, nodeID, side string) []string {
	var out []string
	for actorID, player := range state.Players {
		if !player.Alive || player.Side != side {
			continue
		}
		if player.Location.NodeID == nodeID || configuredVisible(state, player.Location, nodeID) {
			out = append(out, actorID)
		}
	}
	sort.Strings(out)
	return out
}

func configuredVisible(state *roundState, from playerLocation, toNode string) bool {
	fromNode := projectedNodeID(from)
	for _, visibility := range state.visibility {
		if visibility.Visible && (visibility.FromNode == fromNode && visibility.ToNode == toNode || visibility.FromNode == toNode && visibility.ToNode == fromNode) {
			return true
		}
	}
	return false
}

func bombRecoveryNode(location playerLocation) string {
	if location.NodeID != "" {
		return location.NodeID
	}
	if location.Edge == nil {
		return ""
	}
	if location.Edge.Progress < 0.5 {
		return location.Edge.FromNode
	}
	return location.Edge.ToNode
}
