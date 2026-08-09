package matchengine

import (
	"math"
	"sort"
)

const (
	PriorityPlantComplete  = 70
	PriorityDefuseComplete = 60
	PriorityBombExplode    = 50
)

type SiteContestDecisionType string

const (
	SiteContestEncounter SiteContestDecisionType = "Encounter"
	SiteContestPlant     SiteContestDecisionType = "Plant"
	SiteContestWithdraw  SiteContestDecisionType = "Withdraw"
)

type SiteContestDecision struct {
	Type       SiteContestDecisionType
	Site       string
	ActorIDs   []string
	PlantScore float64
	PlantRisk  float64
	Reasons    []ReasonRecord
}

type BombRecoverySchedule struct {
	Action         ScheduledAction
	Path           PathResult
	MoveDuration   int
	PickupDuration int
}

type BombActionResult struct {
	Applied  bool
	Event    *GameEvent
	FollowUp *ScheduledAction
}

func PlanSiteContest(state *RoundState, site string, actorIDs []string) (SiteContestDecision, error) {
	nodeID := plantNodeForSite(state, site)
	if nodeID == "" {
		return SiteContestDecision{}, newError("INVALID_PLANT", "site %s has no plant node", site)
	}
	actors := liveActorsAtNode(state, actorIDs, nodeID)
	carrier := state.Players[state.Bomb.CarrierID]
	if carrier == nil || !carrier.Alive || carrier.Side != SideT || carrier.Location.NodeID != nodeID {
		return SiteContestDecision{Type: SiteContestWithdraw, Site: site, ActorIDs: actors, Reasons: []ReasonRecord{{Code: "NO_CARRIER_AT_SITE", Source: site, Value: -1, Weight: 1}}}, nil
	}
	node := state.Nodes[nodeID]
	threats := visibleThreatsAtSite(state, nodeID, SideCT)
	plantScore, plantRisk := CalculatePlantScore(state, carrier.Profile.PlayerID, site, threats)
	if node.ActualControl == ControlCT || node.ActualControl == ControlContested || len(threats) > 0 {
		return SiteContestDecision{Type: SiteContestEncounter, Site: site, ActorIDs: append(actors, threats...), PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []ReasonRecord{{Code: "VISIBLE_SITE_THREAT", Source: nodeID, Value: -plantRisk, Weight: 1}}}, nil
	}
	if plantScore >= plantRisk {
		return SiteContestDecision{Type: SiteContestPlant, Site: site, ActorIDs: actors, PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []ReasonRecord{{Code: "PLANT_WINDOW", Source: nodeID, Value: plantScore - plantRisk, Weight: 1}}}, nil
	}
	return SiteContestDecision{Type: SiteContestWithdraw, Site: site, ActorIDs: actors, PlantScore: plantScore, PlantRisk: plantRisk, Reasons: []ReasonRecord{{Code: "PLANT_RISK", Source: nodeID, Value: plantScore - plantRisk, Weight: 1}}}, nil
}

func CalculatePlantScore(state *RoundState, actorID, site string, visibleThreatIDs []string) (float64, float64) {
	actor := state.Players[actorID]
	if actor == nil {
		return 0, math.Inf(1)
	}
	cover := ScopedUtilityModifier(state, SideT, UtilityPlantCover, "", "site:"+site, state.Timeline)
	score := float64(actor.Profile.Attributes.Composure+actor.Profile.Attributes.Discipline+actor.Focus)/3 + cover*20
	risk := float64(len(visibleThreatIDs))*25 + playerExposure(actor)
	if node := state.Nodes[actor.Location.NodeID]; node != nil && node.ActualControl == ControlCT {
		risk += 30
	}
	return score, risk
}

func CanAttemptPlant(state *RoundState, actorID, site string) error {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideT || !actor.HasBomb || state.Bomb.CarrierID != actorID || state.Bomb.Status != BombCarried {
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
	if node.ActualControl == ControlCT || node.ActualControl == ControlContested {
		return newError("SITE_CONTEST_REQUIRED", "plant site must be contested before planting")
	}
	return nil
}

func StartPlantAction(state *RoundState, actorID, site, intentID string, ordinal int) (ScheduledAction, error) {
	if err := CanAttemptPlant(state, actorID, site); err != nil {
		return ScheduledAction{}, err
	}
	actor := state.Players[actorID]
	utilityResult, err := spendScopedUtility(state, SideT, UtilityPlantCover, []string{actorID}, "", "site:"+site, state.constants.Int("BasePlantTime", 1), 10)
	if err != nil {
		return ScheduledAction{}, err
	}
	cover := ScopedUtilityModifier(state, SideT, UtilityPlantCover, "", "site:"+site, state.Timeline)
	duration := int(math.Round(float64(state.constants.Int("BasePlantTime", 1)) - cover - float64(actor.Profile.Attributes.Discipline-50)/100))
	duration = clampInt(duration, state.constants.Int("MinPlantTime", 1), state.constants.Int("MaxPlantTime", 1))
	action := ScheduledAction{IntentID: intentID, Type: ActionPlantComplete, ActorIDs: []string{actorID}, From: actor.Location, StartAt: state.Timeline, ResolveAt: state.Timeline + duration, Priority: PriorityPlantComplete, MinRequiredActors: 1, Payload: ActionPayload{Site: site}}
	action.ID = NewActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := BeginExclusiveAction(state, &action, ActionPlanting); err != nil {
		return ScheduledAction{}, err
	}
	actor.Intent = Intent{ID: intentID, Type: IntentPlant, TargetID: site, CreatedAt: state.Timeline}
	if err := state.Bomb.StartPlant(actorID, action.ID, site, action.StartAt, action.ResolveAt); err != nil {
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		state.Bomb.InterruptPlant(actorID, action.ID)
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	event, _ := newActionLifecycleEvent(state, action, EventPlantStart, "bomb plant started", 0)
	addUtilityReason(event, utilityResult)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return action, nil
}

func ResolvePlantComplete(state *RoundState, action ScheduledAction) (BombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Timeline > state.RoundDeadline {
		return BombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	if state.Bomb.Status != BombPlanting || state.Bomb.PlantActionID != action.ID || !sameLocation(actor.Location, action.From) {
		return BombActionResult{}, nil
	}
	explodeAt := state.Timeline + state.constants.Int("BombExplodeTime", 1)
	if err := state.Bomb.CompletePlant(actor.Location, state.Timeline, explodeAt); err != nil {
		return BombActionResult{}, err
	}
	actor.HasBomb = false
	CompleteActionForActors(state, action, valid)
	state.BombDeadline = explodeAt
	state.Phase = PhasePostPlant
	explode := ScheduledAction{ID: NewActionID(state.Seed, ActionBombExplode, action.ID, state.Timeline, explodeAt, nil, 0), IntentID: action.ID, Type: ActionBombExplode, StartAt: state.Timeline, ResolveAt: explodeAt, Priority: PriorityBombExplode, Payload: ActionPayload{Site: state.Bomb.PlantedSite}}
	if err := state.Scheduler.Schedule(explode); err != nil {
		return BombActionResult{}, err
	}
	event := bombLifecycleEffectEvent(state, action, EventBombPlant, "bomb planted", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return BombActionResult{Applied: true, Event: event, FollowUp: &explode}, nil
}

func ScheduleBombRecovery(state *RoundState, actorID, intentID string, ordinal int) (BombRecoverySchedule, *DecisionFeedback, error) {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideT || actor.Action.CurrentActionID != "" || state.Bomb.Status != BombDropped {
		return BombRecoverySchedule{}, nil, newError("INVALID_PICKUP", "bomb recovery actor is unavailable")
	}
	targetNode := bombRecoveryNode(state.Bomb.Location)
	if targetNode == "" {
		return BombRecoverySchedule{}, nil, newError("INVALID_PICKUP", "dropped bomb has no semantic location")
	}
	if node := state.Nodes[targetNode]; node != nil && (node.ActualControl == ControlCT || node.ActualControl == ControlContested) {
		return BombRecoverySchedule{}, &DecisionFeedback{Code: "SITE_CONTEST_REQUIRED", Message: "bomb location is enemy-controlled or contested"}, nil
	}
	path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, actor.Location.NodeID, targetNode, len(state.Nodes)*4)
	if err != nil || feedback != nil {
		return BombRecoverySchedule{}, feedback, err
	}
	pickupDuration := clampInt(state.constants.Int("BasePickupTime", 1), state.constants.Int("MinPickupTime", 1), state.constants.Int("MaxPickupTime", 1))
	moveDuration := path.TotalBaseTime
	resolveAt := state.Timeline + moveDuration + pickupDuration
	action := ScheduledAction{IntentID: intentID, Type: ActionPickupComplete, ActorIDs: []string{actorID}, From: actor.Location, ToNodeID: targetNode, StartAt: state.Timeline, ResolveAt: resolveAt, Priority: PriorityPlantComplete, MinRequiredActors: 1, Payload: ActionPayload{TargetID: targetNode}}
	action.ID = NewActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := BeginExclusiveAction(state, &action, ActionMoving); err != nil {
		return BombRecoverySchedule{}, nil, err
	}
	actor.Intent = Intent{ID: intentID, Type: IntentPickupBomb, TargetID: targetNode, CreatedAt: state.Timeline}
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return BombRecoverySchedule{}, nil, err
	}
	return BombRecoverySchedule{Action: action, Path: path, MoveDuration: moveDuration, PickupDuration: pickupDuration}, nil, nil
}

func ResolveBombPickup(state *RoundState, action ScheduledAction) (BombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Bomb.Status != BombDropped {
		return BombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	actor.Location = clonePlayerLocation(state.Bomb.Location)
	if err := state.Bomb.Pickup(actor.Profile.PlayerID, actor.Location, state.Timeline); err != nil {
		return BombActionResult{}, err
	}
	actor.HasBomb = true
	CompleteActionForActors(state, action, valid)
	event := bombLifecycleEffectEvent(state, action, EventBombPickup, "bomb picked up", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return BombActionResult{Applied: true, Event: event}, nil
}

func CalculateDefuseTime(state *RoundState, actorID string) int {
	actor := state.Players[actorID]
	duration := float64(state.constants.Int("BaseDefuseTime", 1))
	if actor != nil && actor.Weapon.HasKit {
		duration *= 0.5
	}
	cover := ScopedUtilityModifier(state, SideCT, UtilityDefuseCover, "", "site:"+state.Bomb.PlantedSite, state.Timeline)
	duration -= cover
	return clampInt(int(math.Round(duration)), state.constants.Int("MinDefuseTime", 1), state.constants.Int("MaxDefuseTime", 1))
}

func CalculateDefuseScore(state *RoundState, actorID string) float64 {
	actor := state.Players[actorID]
	if actor == nil {
		return math.Inf(-1)
	}
	score := float64(actor.Focus+actor.Profile.Attributes.Composure+actor.Profile.Attributes.Discipline) / 3
	if actor.Weapon.HasKit {
		score += 20
	}
	if CanDenyDefuse(state, actor.Location.NodeID, state.Timeline+CalculateDefuseTime(state, actorID)) {
		score -= 30
	}
	return score
}

func CanAttemptDefuse(state *RoundState, actorID string) error {
	actor := state.Players[actorID]
	if actor == nil || !actor.Alive || actor.Side != SideCT || actor.EngagementID != "" || actor.Action.CurrentActionID != "" {
		return newError("INVALID_DEFUSE", "defuser is unavailable")
	}
	if state.Bomb.Status != BombPlanted || actor.Location.NodeID == "" || actor.Location.NodeID != state.Bomb.Location.NodeID {
		return newError("INVALID_DEFUSE", "defuser has not reached the planted bomb")
	}
	if state.Timeline+CalculateDefuseTime(state, actorID) > state.BombDeadline {
		return newError("INVALID_DEFUSE", "defuse cannot finish before bomb deadline")
	}
	if CanDenyDefuse(state, actor.Location.NodeID, state.Timeline+CalculateDefuseTime(state, actorID)) {
		return newError("DEFUSE_CONTEST_REQUIRED", "a live T can deny the defuse before completion")
	}
	return nil
}

func StartDefuseAction(state *RoundState, actorID, intentID string, ordinal int) (ScheduledAction, error) {
	if err := CanAttemptDefuse(state, actorID); err != nil {
		return ScheduledAction{}, err
	}
	actor := state.Players[actorID]
	utilityResult, err := spendScopedUtility(state, SideCT, UtilityDefuseCover, []string{actorID}, "", "site:"+state.Bomb.PlantedSite, state.constants.Int("BaseDefuseTime", 1), 10)
	if err != nil {
		return ScheduledAction{}, err
	}
	duration := CalculateDefuseTime(state, actorID)
	action := ScheduledAction{IntentID: intentID, Type: ActionDefuseComplete, ActorIDs: []string{actorID}, From: actor.Location, StartAt: state.Timeline, ResolveAt: state.Timeline + duration, Priority: PriorityDefuseComplete, MinRequiredActors: 1, Payload: ActionPayload{Site: state.Bomb.PlantedSite}}
	action.ID = NewActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := BeginExclusiveAction(state, &action, ActionDefusing); err != nil {
		return ScheduledAction{}, err
	}
	actor.Intent = Intent{ID: intentID, Type: IntentDefuse, TargetID: state.Bomb.PlantedSite, CreatedAt: state.Timeline}
	if err := state.Bomb.StartDefuse(actorID, action.ID, action.StartAt, action.ResolveAt); err != nil {
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		state.Bomb.InterruptDefuse(actorID, action.ID)
		cancelActionForActors(state, action)
		return ScheduledAction{}, err
	}
	event, _ := newActionLifecycleEvent(state, action, EventDefuseStart, "bomb defuse started", 0)
	addUtilityReason(event, utilityResult)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return action, nil
}

func ResolveDefuseComplete(state *RoundState, action ScheduledAction) (BombActionResult, error) {
	valid := validActionActors(state, action)
	if len(valid) != 1 || state.Timeline != action.ResolveAt || state.Timeline > state.BombDeadline {
		return BombActionResult{}, nil
	}
	actor := state.Players[valid[0]]
	if !actor.Alive || state.Bomb.Status != BombDefusing || state.Bomb.DefuseActionID != action.ID || !sameLocation(actor.Location, action.From) {
		return BombActionResult{}, nil
	}
	if err := state.Bomb.CompleteDefuse(state.Timeline); err != nil {
		return BombActionResult{}, err
	}
	CompleteActionForActors(state, action, valid)
	event := bombLifecycleEffectEvent(state, action, EventBombDefuse, "bomb defused", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return BombActionResult{Applied: true, Event: event}, nil
}

func ResolveBombExplode(state *RoundState, action ScheduledAction) (BombActionResult, error) {
	if state.Timeline != action.ResolveAt || state.Bomb.Status == BombDefused || state.Bomb.ExplodeAt != action.ResolveAt {
		return BombActionResult{}, nil
	}
	if err := state.Bomb.Explode(state.Timeline); err != nil {
		return BombActionResult{}, err
	}
	event := bombLifecycleEffectEvent(state, action, EventBombExplode, "bomb exploded", 0)
	event.Bomb = projectBombState(state.Bomb)
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	return BombActionResult{Applied: true, Event: event}, nil
}

func CanDenyDefuse(state *RoundState, bombNode string, finishAt int) bool {
	for _, player := range state.Players {
		if !player.Alive || player.Side != SideT {
			continue
		}
		from := player.Location.NodeID
		if from == "" {
			from = projectedNodeID(player.Location)
		}
		path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, from, bombNode, len(state.Nodes)*4)
		if err == nil && feedback == nil && state.Timeline+path.TotalBaseTime <= finishAt {
			return true
		}
	}
	return false
}

func bombLifecycleEffectEvent(state *RoundState, action ScheduledAction, eventType, message string, ordinal int) *GameEvent {
	effectID := NewEffectID(state.Seed, action.ID, EffectBombState, ordinal)
	eventID := NewEventID(state.Seed, action.ID, effectID, eventType, ordinal)
	reason, _ := ProjectReasonRecord(ReasonRecord{Code: eventType, Source: string(action.Type), Value: 1, Weight: 1}, action.ID, effectID)
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

func plantNodeForSite(state *RoundState, site string) string {
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

func liveActorsAtNode(state *RoundState, actorIDs []string, nodeID string) []string {
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

func visibleThreatsAtSite(state *RoundState, nodeID, side string) []string {
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

func configuredVisible(state *RoundState, from PlayerLocation, toNode string) bool {
	fromNode := projectedNodeID(from)
	for _, visibility := range state.visibility {
		if visibility.Visible && (visibility.FromNode == fromNode && visibility.ToNode == toNode || visibility.FromNode == toNode && visibility.ToNode == fromNode) {
			return true
		}
	}
	return false
}

func bombRecoveryNode(location PlayerLocation) string {
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
