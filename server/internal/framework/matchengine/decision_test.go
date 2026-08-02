package matchengine

import (
	"reflect"
	"testing"
)

func TestDecisionTriggersReadCurrentAuthoritativeState(t *testing.T) {
	state := makeTestRoundState(t, 901)
	state.Timeline = 90
	state.ActiveEngagements["enc"] = &EncounterState{ID: "enc", Status: EncounterActive}
	before := CaptureDecisionFingerprint(state)
	delete(state.ActiveEngagements, "enc")
	state.Players["team_a_p2"].HP = 20
	state.Players["team_a_p2"].Focus = 20
	state.Nodes["A_SITE"].KnownControl[SideT] = KnownControlState{Status: ControlCT, UpdatedAt: 95, ExpiresAt: 110}
	state.Events = append(state.Events, &GameEvent{EventID: "drop", EventType: EventBombDrop})
	state.Players[state.Bomb.CarrierID].HasBomb = false
	if err := state.Bomb.Drop(PlayerLocation{NodeID: "A_LONG"}, 95); err != nil {
		t.Fatal(err)
	}
	state.Timeline = 95
	if _, err := RecordIntel(state, SideT, IntelObservation{Type: IntelEmptySite, NodeID: "B_SITE", At: 95, TTL: 10}); err != nil {
		t.Fatal(err)
	}
	state.Timeline = 96
	triggers := DetectDecisionTriggers(state, before, []DecisionTriggerType{TriggerRouteBlocked, TriggerPostPlantArrival, TriggerDefuseInterrupted})
	types := map[DecisionTriggerType]bool{}
	for _, trigger := range triggers {
		types[trigger.Type] = true
	}
	for _, want := range []DecisionTriggerType{TriggerEncounterEnd, TriggerResourceBand, TriggerControlChanged, TriggerEmptySite, TriggerBombChanged, TriggerForceExecute, TriggerRouteBlocked, TriggerPostPlantArrival, TriggerDefuseInterrupted} {
		if !types[want] {
			t.Fatalf("missing decision trigger %s: %+v", want, triggers)
		}
	}
}

func TestDecisionCandidatesCoverTCTAndCTCannotReadHiddenRotate(t *testing.T) {
	state := makeTestRoundState(t, 902)
	state.Players["team_a_p2"].Profile.RoleTags = []string{"Lurker"}
	state.Timeline = 100
	tView, err := BuildDecisionView(state, SideT)
	if err != nil {
		t.Fatal(err)
	}
	tCandidates := ScoreDecisionCandidates(tView, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	assertDecisionTypes(t, tCandidates, DecisionContinue, DecisionGatherIntel, DecisionHoldFlank, DecisionInterceptRotate, DecisionForceExecute)

	state.Bomb.Status, state.Bomb.Location, state.Bomb.PlantedSite, state.BombDeadline = BombPlanted, PlayerLocation{NodeID: "A_SITE"}, "A", 105
	state.Players["team_b_p1"].Location = PlayerLocation{NodeID: "A_SITE"}
	state.Players["team_b_p2"].Location = PlayerLocation{NodeID: "CT_SPAWN"}
	state.Intel[SideCT].Records = []IntelRecord{{ID: "known", Type: string(IntelDirectVisibility), TargetID: "team_a_p2", NodeID: "A_LONG", Confidence: 80, ExpiresAt: 110}}
	rebuildIntelIndexes(state.Intel[SideCT], state.Timeline)
	ctView, err := BuildDecisionView(state, SideCT)
	if err != nil {
		t.Fatal(err)
	}
	before := ScoreDecisionCandidates(ctView, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	assertDecisionTypes(t, before, DecisionHold, DecisionDefuse, DecisionRetake, DecisionReinforce, DecisionInterceptRotate, DecisionSave)
	state.Players["team_a_p2"].Intent = Intent{ID: "secret-rotate", Type: IntentMove, TargetID: "B_SITE"}
	state.Players["team_a_p2"].Action = PlayerActionState{CurrentActionID: "secret-action", Status: ActionMoving, Version: 8}
	ctViewAfter, err := BuildDecisionView(state, SideCT)
	if err != nil {
		t.Fatal(err)
	}
	after := ScoreDecisionCandidates(ctViewAfter, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	if !reflect.DeepEqual(before, after) {
		t.Fatal("CT InterceptRotate scoring read T hidden rotate action")
	}
}

func TestDecisionDelayResolvesIntoRealMovement(t *testing.T) {
	state := makeTestRoundState(t, 903)
	candidate := DecisionCandidate{Type: DecisionRotate, Side: SideT, ActorIDs: []string{"team_a_p2"}, TargetNode: "LONG_DOOR", RouteID: "D2_A_LONG", Rotation: true, DeterministicScore: 50}
	action, normalized, err := ScheduleDecision(state, candidate, 0)
	if err != nil {
		t.Fatal(err)
	}
	if action.ResolveAt-state.Timeline != state.constants.Int("DecisionDelay", 0) || state.Players["team_a_p2"].Action.CurrentActionID != "" {
		t.Fatalf("decision did not wait independently: %+v", action)
	}
	state.Timeline = action.ResolveAt
	resolution, err := ResolveDecision(state, action, normalized)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Actions) != 1 || resolution.Actions[0].Type != ActionMovementArrive || state.Players["team_a_p2"].Intent.Type != IntentMove || state.Players["team_a_p2"].Location.Edge == nil {
		t.Fatalf("decision produced narrative instead of real Move: %+v", resolution)
	}
}

func TestDecisionAndRotationLimitsForceReachableActionWithoutWinner(t *testing.T) {
	state := makeTestRoundState(t, 904)
	state.DecisionCount = state.constants.Int("MaxDecisionCount", 0)
	state.RotationCount[SideT] = state.constants.Int("MaxRotationsPerTeam", 0)
	candidate := DecisionCandidate{Type: DecisionRotate, Side: SideT, ActorIDs: []string{"team_a_p2"}, TargetNode: "LONG_DOOR", RouteID: "D2_A_LONG", Rotation: true}
	action, normalized, err := ScheduleDecision(state, candidate, 0)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Type != DecisionForceExecute || normalized.Rotation || state.DecisionCount != state.constants.Int("MaxDecisionCount", 0) || state.Terminal != nil {
		t.Fatalf("limit fabricated outcome or kept rotating: candidate=%+v terminal=%+v", normalized, state.Terminal)
	}
	state.Timeline = action.ResolveAt
	resolution, err := ResolveDecision(state, action, normalized)
	if err != nil || len(resolution.Actions) == 0 || normalized.TargetNode == "LONG_DOOR" {
		t.Fatalf("limit did not force a reachable execution action: %+v/%v", resolution, err)
	}
}

func TestForceExecuteThresholdSelectsRealSiteMovement(t *testing.T) {
	state := makeTestRoundState(t, 906)
	setConstInt(&state.constants, "ForceExecuteThreshold", 30)
	state.Timeline = state.RoundDeadline - 30
	view, err := BuildDecisionView(state, SideT)
	if err != nil {
		t.Fatal(err)
	}
	candidates := ScoreDecisionCandidates(view, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	if len(candidates) == 0 || candidates[0].Type != DecisionForceExecute || candidates[0].TargetNode != "A_SITE" || len(candidates[0].ActorIDs) != 5 {
		t.Fatalf("low-time decision did not dominate with a real team execute: %+v", candidates)
	}
	action, normalized, err := ScheduleDecision(state, candidates[0], 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = action.ResolveAt
	resolution, err := ResolveDecision(state, action, normalized)
	if err != nil || len(resolution.Actions) == 0 {
		t.Fatalf("force execute produced no causal movement: %+v/%v", resolution, err)
	}
	for _, movement := range resolution.Actions {
		if movement.Type != ActionMovementArrive || movement.ToNodeID == "" {
			t.Fatalf("force execute emitted a non-movement action: %+v", movement)
		}
	}
}

func TestDecisionMovementPersistsUltimateTargetAcrossEdges(t *testing.T) {
	state := makeTestRoundState(t, 907)
	candidate := DecisionCandidate{Type: DecisionForceExecute, Side: SideT, ActorIDs: []string{"team_a_p2"}, TargetNode: "A_SITE", RouteID: "D2_A_LONG"}
	action, normalized, err := ScheduleDecision(state, candidate, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = action.ResolveAt
	resolution, err := ResolveDecision(state, action, normalized)
	if err != nil || len(resolution.Actions) != 1 {
		t.Fatalf("decision movement setup failed: %+v/%v", resolution, err)
	}
	first := resolution.Actions[0]
	if state.Players["team_a_p2"].Intent.TargetID != "A_SITE" {
		t.Fatalf("decision lost ultimate target on first edge: %+v", state.Players["team_a_p2"].Intent)
	}
	state.Timeline = first.ResolveAt
	if err := CompleteMovement(state, first, first.ActorIDs); err != nil {
		t.Fatal(err)
	}
	runtime := &causalRoundRuntime{state: state, decisionCandidates: map[string]DecisionCandidate{}}
	started, err := runtime.ensurePrePlantActions()
	if err != nil || !started || state.Players["team_a_p2"].Location.Edge == nil || state.Players["team_a_p2"].Intent.TargetID != "A_SITE" {
		t.Fatalf("decision did not continue toward ultimate target: started=%t player=%+v err=%v", started, state.Players["team_a_p2"], err)
	}
}

func TestCurrentDamageBombControlIntelAndTimeChangeNextAction(t *testing.T) {
	state := makeTestRoundState(t, 905)
	state.Players["team_a_p2"].Profile.RoleTags = []string{"Lurker"}
	baselineView, _ := BuildDecisionView(state, SideT)
	baseline := ScoreDecisionCandidates(baselineView, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	state.Players["team_a_p2"].HP, state.Players["team_a_p2"].Focus, state.Players["team_a_p2"].Stamina = 20, 20, 20
	state.Nodes["A_SITE"].KnownControl[SideT] = KnownControlState{Status: ControlCT, UpdatedAt: 20, ExpiresAt: 100}
	state.Timeline = 100
	state.Players[state.Bomb.CarrierID].HasBomb = false
	if err := state.Bomb.Drop(PlayerLocation{NodeID: "A_LONG"}, 100); err != nil {
		t.Fatal(err)
	}
	changedView, _ := BuildDecisionView(state, SideT)
	changed := ScoreDecisionCandidates(changedView, state.routes, state.constants, IdentityRollSource{Seed: state.Seed})
	if baseline[0].Type == changed[0].Type || changed[0].Type != DecisionRecoverBomb {
		t.Fatalf("authoritative changes did not alter actual next decision: baseline=%+v changed=%+v", baseline[0], changed[0])
	}
}

func assertDecisionTypes(t *testing.T, candidates []DecisionCandidate, expected ...DecisionType) {
	t.Helper()
	seen := map[DecisionType]bool{}
	for _, candidate := range candidates {
		seen[candidate.Type] = true
	}
	for _, decisionType := range expected {
		if !seen[decisionType] {
			t.Fatalf("missing %s candidate: %+v", decisionType, candidates)
		}
	}
}
