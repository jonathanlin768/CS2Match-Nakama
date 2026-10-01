package matchengine

import (
	"math"
	"sort"
)

const (
	encounterActive = "Active"
	encounterEnded  = "Ended"
)

type threatCandidate struct {
	PlayerID    string
	Side        string
	ThreatScore float64
	Exposure    float64
}

type encounterCandidatePlan struct {
	ID             string
	SourceActionID string
	ScenarioID     string
	NodeID         string
	Actors         []threatCandidate
	PriorityScore  float64
}

type encounterSchedule struct {
	EncounterID  string
	PulseActions []scheduledAction
	EndAction    scheduledAction
}

type combatEndResult struct {
	EncounterID       string
	LocalWinnerSide   string
	EndReason         string
	DecisionTriggered bool
}

func buildEncounterCandidate(state *roundState, sourceActionID, scenarioID, nodeID string, actorIDs []string) (encounterCandidatePlan, error) {
	if state == nil || state.scenarios[scenarioID].ID == "" || nodeID == "" {
		return encounterCandidatePlan{}, newError("INVALID_ENCOUNTER", "encounter requires configured scenario and contact node")
	}
	ids := append([]string(nil), actorIDs...)
	sort.Strings(ids)
	ids = uniqueStrings(ids)
	actors := make([]threatCandidate, 0, len(ids))
	sides := map[string]bool{}
	for _, actorID := range ids {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.EngagementID != "" || !actionCanEnterEncounter(player.Action.Status) {
			continue
		}
		candidate := threatCandidate{PlayerID: actorID, Side: player.Side, ThreatScore: encounterThreatScore(player), Exposure: playerExposure(player)}
		actors = append(actors, candidate)
		sides[player.Side] = true
	}
	if !sides[SideT] || !sides[SideCT] {
		return encounterCandidatePlan{}, newError("INVALID_ENCOUNTER", "encounter has no legal opposing actors")
	}
	sort.SliceStable(actors, func(i, j int) bool {
		if actors[i].ThreatScore != actors[j].ThreatScore {
			return actors[i].ThreatScore > actors[j].ThreatScore
		}
		return actors[i].PlayerID < actors[j].PlayerID
	})
	priority := 0.0
	stableIDs := make([]string, 0, len(actors))
	for _, actor := range actors {
		priority += actor.ThreatScore
		stableIDs = append(stableIDs, actor.PlayerID)
	}
	sort.Strings(stableIDs)
	id := stableObjectID("enc", state.Seed, sourceActionID, scenarioID, nodeID, joinStrings(stableIDs))
	return encounterCandidatePlan{ID: id, SourceActionID: sourceActionID, ScenarioID: scenarioID, NodeID: nodeID, Actors: actors, PriorityScore: priority}, nil
}

func arbitrateEncounterCandidates(candidates []encounterCandidatePlan) []encounterCandidatePlan {
	ordered := append([]encounterCandidatePlan(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].PriorityScore != ordered[j].PriorityScore {
			return ordered[i].PriorityScore > ordered[j].PriorityScore
		}
		return ordered[i].ID < ordered[j].ID
	})
	usedActors, usedNodes := map[string]bool{}, map[string]bool{}
	accepted := make([]encounterCandidatePlan, 0, len(ordered))
	for _, candidate := range ordered {
		conflict := usedNodes[candidate.NodeID]
		for _, actor := range candidate.Actors {
			if usedActors[actor.PlayerID] {
				conflict = true
			}
		}
		if conflict {
			continue
		}
		accepted = append(accepted, candidate)
		usedNodes[candidate.NodeID] = true
		for _, actor := range candidate.Actors {
			usedActors[actor.PlayerID] = true
		}
	}
	sort.SliceStable(accepted, func(i, j int) bool { return accepted[i].ID < accepted[j].ID })
	return accepted
}

func startEncounter(state *roundState, candidate encounterCandidatePlan) (*encounterSchedule, error) {
	if state == nil || candidate.ID == "" || state.ActiveEngagements[candidate.ID] != nil {
		return nil, newError("INVALID_ENCOUNTER", "encounter cannot start")
	}
	actorIDs := make([]string, 0, len(candidate.Actors))
	for _, actor := range candidate.Actors {
		player := state.Players[actor.PlayerID]
		if player == nil || !player.Alive || player.EngagementID != "" {
			return nil, newError("ENCOUNTER_ACTOR_CONFLICT", "actor %s is already unavailable", actor.PlayerID)
		}
		actorIDs = append(actorIDs, actor.PlayerID)
	}
	sort.Strings(actorIDs)
	scenario := state.scenarios[candidate.ScenarioID]
	duration := clampInt(scenario.BaseTimeCost, state.constants.Int("MinCombatDuration", 1), state.constants.Int("MaxCombatDuration", 1))
	utilityReasons := make([]reasonRecord, 0, 2)
	for _, side := range []string{SideT, SideCT} {
		result, err := spendScopedUtility(state, side, utilityOpeningInitiative, actorIDs, "", candidate.ID, duration, 5)
		if err != nil {
			return nil, err
		}
		if result.Reason.Code != "" {
			utilityReasons = append(utilityReasons, result.Reason)
		}
	}
	temporary := &encounterState{ID: candidate.ID, ScenarioID: candidate.ScenarioID, ActorIDs: actorIDs, NodeID: candidate.NodeID, StartedAt: state.Timeline}
	scores, err := calculateEncounterScorePair(state, temporary, candidate.ScenarioID, deriveSeed(state.Seed, "encounter", candidate.ID))
	if err != nil {
		return nil, err
	}
	initiativeSide := SideT
	if scores[SideCT].FinalScore > scores[SideT].FinalScore {
		initiativeSide = SideCT
	}
	if math.Abs(scores[SideT].FinalScore-scores[SideCT].FinalScore) >= state.constants.Float("DecisiveScoreGap", 0) {
		duration = maxInt(state.constants.Int("MinCombatDuration", 1), duration-1)
	}
	pulseWindow := maxInt(1, state.constants.Int("PulseFireWindow", 1))
	pulseCount := clampInt(int(math.Ceil(float64(duration)/float64(pulseWindow))), 1, state.constants.Int("MaxEncounterPulses", 1))
	pulseCount = minDamageInt(pulseCount, maxInt(1, duration))
	encounter := &encounterState{
		ID: candidate.ID, SourceActionID: candidate.SourceActionID, ScenarioID: candidate.ScenarioID, ActorIDs: actorIDs,
		NodeID: candidate.NodeID, StartedAt: state.Timeline, EndsAt: state.Timeline + duration, MaxPulses: pulseCount,
		InitiativeSide: initiativeSide, Status: encounterActive,
		Reasons: append(append(append([]reasonRecord(nil), scores[SideT].Reasons...), scores[SideCT].Reasons...), utilityReasons...),
	}
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		cancelCurrentAction(state, player)
		player.Action.Version++
		player.Action.CurrentActionID = encounter.ID
		player.Action.Status = actionEngaged
		player.Action.BusyUntil = encounter.EndsAt
		player.Action.Busy = busyInterval{ActionID: encounter.ID, StartAt: state.Timeline, EndAt: encounter.EndsAt}
		player.EngagementID = encounter.ID
		player.Posture = initialEncounterPosture(player, initiativeSide)
	}
	state.ActiveEngagements[encounter.ID] = encounter

	schedule := &encounterSchedule{EncounterID: encounter.ID}
	for pulseIndex := 1; pulseIndex <= pulseCount; pulseIndex++ {
		resolveAt := state.Timeline + int(math.Ceil(float64(duration*pulseIndex)/float64(pulseCount)))
		if len(encounter.PulseTimes) > 0 && resolveAt <= encounter.PulseTimes[len(encounter.PulseTimes)-1] {
			resolveAt = encounter.PulseTimes[len(encounter.PulseTimes)-1] + 1
		}
		resolveAt = minDamageInt(resolveAt, encounter.EndsAt)
		encounter.PulseTimes = append(encounter.PulseTimes, resolveAt)
		action := encounterAction(state, encounter, actionCombatPulse, resolveAt, priorityCombatPulseCommit, pulseIndex-1)
		if err := state.Scheduler.Schedule(action); err != nil {
			return nil, err
		}
		schedule.PulseActions = append(schedule.PulseActions, action)
	}
	endAction := encounterAction(state, encounter, actionCombatEnd, encounter.EndsAt, 50, 0)
	endAction.ActorIDs = nil
	endAction.VersionByActor = nil
	endAction.ParentActionID = ""
	endAction.Payload.ParticipantIDs = append([]string(nil), actorIDs...)
	if err := state.Scheduler.Schedule(endAction); err != nil {
		return nil, err
	}
	schedule.EndAction = endAction
	return schedule, nil
}

func endEncounter(state *roundState, encounterID, reason string) (combatEndResult, error) {
	encounter := state.ActiveEngagements[encounterID]
	if encounter == nil || encounter.Status != encounterActive {
		return combatEndResult{}, newError("INVALID_ENCOUNTER", "encounter %s is not active", encounterID)
	}
	alive := map[string][]string{SideT: {}, SideCT: {}}
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player != nil && player.Alive {
			alive[player.Side] = append(alive[player.Side], actorID)
		}
	}
	localWinner := ""
	control := controlUnknown
	if len(alive[SideT]) > 0 && len(alive[SideCT]) == 0 {
		localWinner, control = SideT, controlT
	} else if len(alive[SideCT]) > 0 && len(alive[SideT]) == 0 {
		localWinner, control = SideCT, controlCT
	}
	if node := state.Nodes[encounter.NodeID]; node != nil {
		beforeControl := node.ActualControl
		observed := map[string][]string{SideT: alive[SideT], SideCT: alive[SideCT]}
		if err := node.ResolveContest(control, state.Timeline, state.constants.Int("ControlIntelTTL", 1), observed); err != nil {
			return combatEndResult{}, err
		}
		if control != controlUnknown && control != beforeControl {
			effectID := newEffectID(state.Seed, encounter.ID, effectControl, 0)
			eventID := newEventID(state.Seed, encounter.ID, effectID, EventControlGained, 0)
			reason, _ := projectReasonRecord(reasonRecord{Code: "ENCOUNTER_CONTROL_RESOLVED", Source: encounter.ID, Value: 1, Weight: 1, StateChanges: []ReasonStateChange{{Field: "control.status", Before: stringReasonValue(string(beforeControl)), After: stringReasonValue(string(control))}}}, encounter.ID, effectID)
			event := &GameEvent{EventID: eventID, SourceActionID: encounter.ID, SourceEffectID: effectID, Timestamp: int64(state.Timeline), EventType: EventControlGained, Message: "node control resolved from encounter", Reason: reason, Location: eventLocation(state, playerLocation{NodeID: encounter.NodeID}, eventID, effectID), State: snapshotForEvent(state), sortPriority: 40, sortActionType: string(actionCombatEnd)}
			if len(alive[localWinner]) > 0 {
				actor := state.Players[alive[localWinner][0]]
				event.AttackerID, event.AttackerName, event.AttackerTeamID = actor.Profile.PlayerID, actor.Profile.DisplayName, actor.TeamID
			}
			state.Events = append(state.Events, event)
		}
	}
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player == nil {
			continue
		}
		if player.Action.CurrentActionID == encounter.ID {
			player.Action.Version++
			player.Action.CurrentActionID = ""
			player.Action.Status = actionIdle
			player.Action.BusyUntil = state.Timeline
			player.Action.Busy = busyInterval{}
		}
		player.EngagementID = ""
		if player.Alive {
			player.Posture = postureDefault
		}
	}
	encounter.Status = encounterEnded
	delete(state.ActiveEngagements, encounterID)
	return combatEndResult{EncounterID: encounterID, LocalWinnerSide: localWinner, EndReason: reason, DecisionTriggered: true}, nil
}

func encounterShouldEnd(state *roundState, encounter *encounterState) (bool, string) {
	if encounter == nil || encounter.Status != encounterActive {
		return true, "inactive"
	}
	aliveT, aliveCT := 0, 0
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive {
			continue
		}
		if player.Side == SideT {
			aliveT++
		} else {
			aliveCT++
		}
	}
	if aliveT == 0 || aliveCT == 0 {
		return true, "local_elimination"
	}
	if encounter.PulsesResolved >= encounter.MaxPulses {
		return true, "pulse_limit"
	}
	if state.Timeline >= encounter.EndsAt {
		return true, "duration_limit"
	}
	return false, ""
}

func encounterAction(state *roundState, encounter *encounterState, actionType actionType, resolveAt, priority, ordinal int) scheduledAction {
	versions := make(map[string]int, len(encounter.ActorIDs))
	for _, actorID := range encounter.ActorIDs {
		versions[actorID] = state.Players[actorID].Action.Version
	}
	action := scheduledAction{
		ParentActionID: encounter.ID, IntentID: encounter.ID, Type: actionType, ActorIDs: append([]string(nil), encounter.ActorIDs...),
		StartAt: encounter.StartedAt, ResolveAt: resolveAt, Priority: priority, VersionByActor: versions, MinRequiredActors: 1,
		Payload: actionPayload{ScenarioID: encounter.ScenarioID, TargetID: encounter.ID},
	}
	action.ID = newActionID(state.Seed, actionType, encounter.ID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	return action
}

func actionCanEnterEncounter(status actionStatus) bool {
	switch status {
	case actionIdle, actionMoving, actionHolding, actionPlanting, actionDefusing:
		return true
	default:
		return false
	}
}

func encounterThreatScore(player *roundPlayerState) float64 {
	attributes := player.Profile.Attributes
	return float64(attributes.Firepower+attributes.Aim+attributes.Reaction+attributes.Awareness)/4 + float64(player.HP)/10 + playerExposure(player)
}

func playerExposure(player *roundPlayerState) float64 {
	exposure := 0.0
	if player.Location.Edge != nil {
		exposure += 20
	}
	switch player.Posture {
	case postureMoving:
		exposure += 15
	case postureHolding:
		exposure -= 10
	}
	return exposure
}

func initialEncounterPosture(player *roundPlayerState, initiativeSide string) combatPosture {
	if player.Side == initiativeSide {
		return postureEngaged
	}
	return postureHolding
}

func joinStrings(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += ","
		}
		result += value
	}
	return result
}
