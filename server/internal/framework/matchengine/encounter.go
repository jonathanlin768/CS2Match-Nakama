package matchengine

import (
	"math"
	"sort"
)

const (
	EncounterActive = "Active"
	EncounterEnded  = "Ended"
)

type ThreatCandidate struct {
	PlayerID    string
	Side        string
	ThreatScore float64
	Exposure    float64
}

type EncounterCandidatePlan struct {
	ID             string
	SourceActionID string
	ScenarioID     string
	NodeID         string
	Actors         []ThreatCandidate
	PriorityScore  float64
}

type EncounterSchedule struct {
	EncounterID  string
	PulseActions []ScheduledAction
	EndAction    ScheduledAction
}

type CombatEndResult struct {
	EncounterID       string
	LocalWinnerSide   string
	EndReason         string
	DecisionTriggered bool
}

func BuildEncounterCandidate(state *RoundState, sourceActionID, scenarioID, nodeID string, actorIDs []string) (EncounterCandidatePlan, error) {
	if state == nil || state.scenarios[scenarioID].ID == "" || nodeID == "" {
		return EncounterCandidatePlan{}, newError("INVALID_ENCOUNTER", "encounter requires configured scenario and contact node")
	}
	ids := append([]string(nil), actorIDs...)
	sort.Strings(ids)
	ids = uniqueStrings(ids)
	actors := make([]ThreatCandidate, 0, len(ids))
	sides := map[string]bool{}
	for _, actorID := range ids {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.EngagementID != "" || !actionCanEnterEncounter(player.Action.Status) {
			continue
		}
		candidate := ThreatCandidate{PlayerID: actorID, Side: player.Side, ThreatScore: encounterThreatScore(player), Exposure: playerExposure(player)}
		actors = append(actors, candidate)
		sides[player.Side] = true
	}
	if !sides[SideT] || !sides[SideCT] {
		return EncounterCandidatePlan{}, newError("INVALID_ENCOUNTER", "encounter has no legal opposing actors")
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
	return EncounterCandidatePlan{ID: id, SourceActionID: sourceActionID, ScenarioID: scenarioID, NodeID: nodeID, Actors: actors, PriorityScore: priority}, nil
}

func ArbitrateEncounterCandidates(candidates []EncounterCandidatePlan) []EncounterCandidatePlan {
	ordered := append([]EncounterCandidatePlan(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].PriorityScore != ordered[j].PriorityScore {
			return ordered[i].PriorityScore > ordered[j].PriorityScore
		}
		return ordered[i].ID < ordered[j].ID
	})
	usedActors, usedNodes := map[string]bool{}, map[string]bool{}
	accepted := make([]EncounterCandidatePlan, 0, len(ordered))
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

func StartEncounter(state *RoundState, candidate EncounterCandidatePlan) (*EncounterSchedule, error) {
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
	utilityReasons := make([]ReasonRecord, 0, 2)
	for _, side := range []string{SideT, SideCT} {
		result, err := spendScopedUtility(state, side, UtilityOpeningInitiative, actorIDs, "", candidate.ID, duration, 5)
		if err != nil {
			return nil, err
		}
		if result.Reason.Code != "" {
			utilityReasons = append(utilityReasons, result.Reason)
		}
	}
	temporary := &EncounterState{ID: candidate.ID, ScenarioID: candidate.ScenarioID, ActorIDs: actorIDs, NodeID: candidate.NodeID, StartedAt: state.Timeline}
	scores, err := CalculateEncounterScorePair(state, temporary, candidate.ScenarioID, deriveSeed(state.Seed, "encounter", candidate.ID))
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
	encounter := &EncounterState{
		ID: candidate.ID, SourceActionID: candidate.SourceActionID, ScenarioID: candidate.ScenarioID, ActorIDs: actorIDs,
		NodeID: candidate.NodeID, StartedAt: state.Timeline, EndsAt: state.Timeline + duration, MaxPulses: pulseCount,
		InitiativeSide: initiativeSide, Status: EncounterActive,
		Reasons: append(append(append([]ReasonRecord(nil), scores[SideT].Reasons...), scores[SideCT].Reasons...), utilityReasons...),
	}
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		cancelCurrentAction(state, player)
		player.Action.Version++
		player.Action.CurrentActionID = encounter.ID
		player.Action.Status = ActionEngaged
		player.Action.BusyUntil = encounter.EndsAt
		player.Action.Busy = BusyInterval{ActionID: encounter.ID, StartAt: state.Timeline, EndAt: encounter.EndsAt}
		player.EngagementID = encounter.ID
		player.Posture = initialEncounterPosture(player, initiativeSide)
	}
	state.ActiveEngagements[encounter.ID] = encounter

	schedule := &EncounterSchedule{EncounterID: encounter.ID}
	for pulseIndex := 1; pulseIndex <= pulseCount; pulseIndex++ {
		resolveAt := state.Timeline + int(math.Ceil(float64(duration*pulseIndex)/float64(pulseCount)))
		if len(encounter.PulseTimes) > 0 && resolveAt <= encounter.PulseTimes[len(encounter.PulseTimes)-1] {
			resolveAt = encounter.PulseTimes[len(encounter.PulseTimes)-1] + 1
		}
		resolveAt = minDamageInt(resolveAt, encounter.EndsAt)
		encounter.PulseTimes = append(encounter.PulseTimes, resolveAt)
		action := encounterAction(state, encounter, ActionCombatPulse, resolveAt, PriorityCombatPulseCommit, pulseIndex-1)
		if err := state.Scheduler.Schedule(action); err != nil {
			return nil, err
		}
		schedule.PulseActions = append(schedule.PulseActions, action)
	}
	endAction := encounterAction(state, encounter, ActionCombatEnd, encounter.EndsAt, 50, 0)
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

func EndEncounter(state *RoundState, encounterID, reason string) (CombatEndResult, error) {
	encounter := state.ActiveEngagements[encounterID]
	if encounter == nil || encounter.Status != EncounterActive {
		return CombatEndResult{}, newError("INVALID_ENCOUNTER", "encounter %s is not active", encounterID)
	}
	alive := map[string][]string{SideT: {}, SideCT: {}}
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player != nil && player.Alive {
			alive[player.Side] = append(alive[player.Side], actorID)
		}
	}
	localWinner := ""
	control := ControlUnknown
	if len(alive[SideT]) > 0 && len(alive[SideCT]) == 0 {
		localWinner, control = SideT, ControlT
	} else if len(alive[SideCT]) > 0 && len(alive[SideT]) == 0 {
		localWinner, control = SideCT, ControlCT
	}
	if node := state.Nodes[encounter.NodeID]; node != nil {
		beforeControl := node.ActualControl
		observed := map[string][]string{SideT: alive[SideT], SideCT: alive[SideCT]}
		if err := node.ResolveContest(control, state.Timeline, state.constants.Int("ControlIntelTTL", 1), observed); err != nil {
			return CombatEndResult{}, err
		}
		if control != ControlUnknown && control != beforeControl {
			effectID := NewEffectID(state.Seed, encounter.ID, EffectControl, 0)
			eventID := NewEventID(state.Seed, encounter.ID, effectID, EventControlGained, 0)
			reason, _ := ProjectReasonRecord(ReasonRecord{Code: "ENCOUNTER_CONTROL_RESOLVED", Source: encounter.ID, Value: 1, Weight: 1, StateChanges: []ReasonStateChange{{Field: "control.status", Before: StringReasonValue(string(beforeControl)), After: StringReasonValue(string(control))}}}, encounter.ID, effectID)
			event := &GameEvent{EventID: eventID, SourceActionID: encounter.ID, SourceEffectID: effectID, Timestamp: int64(state.Timeline), EventType: EventControlGained, Message: "node control resolved from encounter", Reason: reason, Location: eventLocation(state, PlayerLocation{NodeID: encounter.NodeID}, eventID, effectID), State: snapshotForEvent(state), sortPriority: 40, sortActionType: string(ActionCombatEnd)}
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
			player.Action.Status = ActionIdle
			player.Action.BusyUntil = state.Timeline
			player.Action.Busy = BusyInterval{}
		}
		player.EngagementID = ""
		if player.Alive {
			player.Posture = PostureDefault
		}
	}
	encounter.Status = EncounterEnded
	delete(state.ActiveEngagements, encounterID)
	return CombatEndResult{EncounterID: encounterID, LocalWinnerSide: localWinner, EndReason: reason, DecisionTriggered: true}, nil
}

func EncounterShouldEnd(state *RoundState, encounter *EncounterState) (bool, string) {
	if encounter == nil || encounter.Status != EncounterActive {
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

func encounterAction(state *RoundState, encounter *EncounterState, actionType ActionType, resolveAt, priority, ordinal int) ScheduledAction {
	versions := make(map[string]int, len(encounter.ActorIDs))
	for _, actorID := range encounter.ActorIDs {
		versions[actorID] = state.Players[actorID].Action.Version
	}
	action := ScheduledAction{
		ParentActionID: encounter.ID, IntentID: encounter.ID, Type: actionType, ActorIDs: append([]string(nil), encounter.ActorIDs...),
		StartAt: encounter.StartedAt, ResolveAt: resolveAt, Priority: priority, VersionByActor: versions, MinRequiredActors: 1,
		Payload: ActionPayload{ScenarioID: encounter.ScenarioID, TargetID: encounter.ID},
	}
	action.ID = NewActionID(state.Seed, actionType, encounter.ID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	return action
}

func actionCanEnterEncounter(status ActionStatus) bool {
	switch status {
	case ActionIdle, ActionMoving, ActionHolding, ActionPlanting, ActionDefusing:
		return true
	default:
		return false
	}
}

func encounterThreatScore(player *RoundPlayerState) float64 {
	attributes := player.Profile.Attributes
	return float64(attributes.Firepower+attributes.Aim+attributes.Reaction+attributes.Awareness)/4 + float64(player.HP)/10 + playerExposure(player)
}

func playerExposure(player *RoundPlayerState) float64 {
	exposure := 0.0
	if player.Location.Edge != nil {
		exposure += 20
	}
	switch player.Posture {
	case PostureMoving:
		exposure += 15
	case PostureHolding:
		exposure -= 10
	}
	return exposure
}

func initialEncounterPosture(player *RoundPlayerState, initiativeSide string) CombatPosture {
	if player.Side == initiativeSide {
		return PostureEngaged
	}
	return PostureHolding
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
