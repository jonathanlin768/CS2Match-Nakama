package matchengine

import (
	"math"
	"reflect"
	"testing"
)

func TestResolveMoveDurationUsesSemanticInputsAndClamp(t *testing.T) {
	state := makeTestRoundState(t, 401)
	actor := state.Players["team_a_p2"]
	actor.Profile.Attributes.Mobility = 80
	edge := MapEdge{BaseTime: 10}
	if got := ResolveMoveDuration(edge, []*RoundPlayerState{actor}, MoveProfile{Tempo: "Default"}, state.constants); got != 9 {
		t.Fatalf("default move duration = %d, want 9", got)
	}
	if got := ResolveMoveDuration(edge, []*RoundPlayerState{actor}, MoveProfile{Tempo: "Fast"}, state.constants); got != 7 {
		t.Fatalf("fast move duration = %d, want 7", got)
	}
	actor.Stamina = 0
	if got := ResolveMoveDuration(edge, []*RoundPlayerState{actor}, MoveProfile{Tempo: "Slow", FormationPenalty: 2}, state.constants); got != 17 {
		t.Fatalf("slow exhausted formation duration = %d, want 17", got)
	}
	setConstInt(&state.constants, "MaxMoveTime", 12)
	if got := ResolveMoveDuration(edge, []*RoundPlayerState{actor}, MoveProfile{Tempo: "Slow", FormationPenalty: 20}, state.constants); got != 12 {
		t.Fatalf("duration did not clamp to max: %d", got)
	}
}

func TestOnEdgeLocationProgressAndDisplayAreReproducible(t *testing.T) {
	state := makeTestRoundState(t, 402)
	actor := "team_a_p2"
	action, err := StartMovement(state, []string{actor}, "E1", MoveProfile{Tempo: "Default"}, "move-long", 0)
	if err != nil {
		t.Fatal(err)
	}
	midpoint := action.StartAt + (action.ResolveAt-action.StartAt)/2
	if err := UpdateMovementProgress(state, action, midpoint, []string{actor}); err != nil {
		t.Fatal(err)
	}
	location := *state.Players[actor].Location.Edge
	if location.Progress <= 0 || location.Progress >= 1 || location.X < 0 || location.X > 1 || location.Y < 0 || location.Y > 1 || location.DisplayName == "" {
		t.Fatalf("invalid OnEdgeLocation: %+v", location)
	}
	state2 := makeTestRoundState(t, 402)
	action2, err := StartMovement(state2, []string{actor}, "E1", MoveProfile{Tempo: "Default"}, "move-long", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := UpdateMovementProgress(state2, action2, midpoint, []string{actor}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(location, *state2.Players[actor].Location.Edge) {
		t.Fatalf("same identity produced different edge display: %+v vs %+v", location, *state2.Players[actor].Location.Edge)
	}
}

func TestBoundedSemanticPathIsStableAndUnreachableReturnsFeedback(t *testing.T) {
	config := makeTestMapConfig()
	path, feedback, err := FindBoundedPath(config, "T_SPAWN", "A_SITE", 100)
	if err != nil || feedback != nil {
		t.Fatalf("FindBoundedPath() = %+v/%v", feedback, err)
	}
	if !reflect.DeepEqual(path.EdgeIDs, []string{"E1", "E2", "E3"}) || path.TotalBaseTime != 24 {
		t.Fatalf("unexpected semantic path: %+v", path)
	}
	reordered := *config
	reordered.Edges = map[string]MapEdge{}
	for _, id := range []string{"E4", "E3", "E2", "E1"} {
		reordered.Edges[id] = config.Edges[id]
	}
	path2, _, err := FindBoundedPath(&reordered, "T_SPAWN", "A_SITE", 100)
	if err != nil || !reflect.DeepEqual(path, path2) {
		t.Fatalf("map insertion order changed path: %+v vs %+v (%v)", path, path2, err)
	}
	config.Nodes["ISLAND"] = MapNode{ID: "ISLAND", Name: "Island"}
	_, feedback, err = FindBoundedPath(config, "T_SPAWN", "ISLAND", 100)
	if err != nil || feedback == nil || feedback.Code != "UNREACHABLE" {
		t.Fatalf("unreachable path did not return decision feedback: %+v/%v", feedback, err)
	}
}

func TestInterceptCheckIsSingleDeterministicCandidateNotAKill(t *testing.T) {
	state := makeTestRoundState(t, 403)
	edge := state.mapEdges["E1"]
	edge.Risk, edge.Noise = 100, 100
	state.mapEdges["E1"] = edge
	movement, err := StartMovement(state, []string{"team_a_p2"}, "E1", MoveProfile{}, "intercepted-move", 0)
	if err != nil {
		t.Fatal(err)
	}
	context := InterceptContext{EnemyActorIDs: []string{"team_b_p1"}, Visible: true, ObserverPosture: PostureHolding, KnownEnemyConfidence: 100}
	result, err := ScheduleInterceptCheck(state, movement, context)
	if err != nil || !result.Scheduled || math.Abs(result.Probability-1) > 0.000001 {
		t.Fatalf("ScheduleInterceptCheck() = %+v/%v", result, err)
	}
	second, err := ScheduleInterceptCheck(state, movement, context)
	if err != nil || second.Scheduled || second.ReasonCode != "ALREADY_EVALUATED" {
		t.Fatalf("second intercept check was not suppressed: %+v/%v", second, err)
	}
	var intercept ScheduledAction
	for state.Scheduler.Len() > 0 {
		action, _ := state.Scheduler.Pop()
		if action.ID == result.ActionID {
			intercept = action
			break
		}
	}
	state.Timeline = intercept.ResolveAt
	candidate, err := ResolveInterceptCheck(state, intercept)
	if err != nil || candidate == nil || !reflect.DeepEqual(candidate.ActorIDs, []string{"team_a_p2", "team_b_p1"}) {
		t.Fatalf("ResolveInterceptCheck() = %+v/%v", candidate, err)
	}
	if countEvents(state.Events, EventKill) != 0 {
		t.Fatal("risk/intercept created a KILL without CombatPulse")
	}
}

func TestIndependentMovementContinuesDuringAnotherEncounter(t *testing.T) {
	state := makeTestRoundState(t, 404)
	first, err := StartMovement(state, []string{"team_a_p2"}, "E1", MoveProfile{}, "move-one", 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := StartMovement(state, []string{"team_a_p3"}, "E1", MoveProfile{}, "move-two", 0)
	if err != nil {
		t.Fatal(err)
	}
	interruptAt := first.StartAt + 2
	if err := InterruptMovementForEncounter(state, first, []string{"team_a_p2"}, interruptAt); err != nil {
		t.Fatal(err)
	}
	state.Timeline = second.ResolveAt
	if err := UpdateMovementProgress(state, second, second.ResolveAt, []string{"team_a_p3"}); err != nil {
		t.Fatal(err)
	}
	if err := CompleteMovement(state, second, []string{"team_a_p3"}); err != nil {
		t.Fatal(err)
	}
	if state.Players["team_a_p2"].Location.Edge == nil || state.Players["team_a_p3"].Location.NodeID != "LONG_DOOR" {
		t.Fatalf("encounter globally froze unrelated movement: first=%+v second=%+v", state.Players["team_a_p2"].Location, state.Players["team_a_p3"].Location)
	}
}

func TestEdgeDeathAndBombDropUseRuntimeLocation(t *testing.T) {
	state := makeTestRoundState(t, 405)
	carrier := state.Bomb.CarrierID
	movement, err := StartMovement(state, []string{carrier}, "E1", MoveProfile{}, "carrier-move", 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = movement.StartAt + 2
	if err := UpdateMovementProgress(state, movement, state.Timeline, []string{carrier}); err != nil {
		t.Fatal(err)
	}
	location := *state.Players[carrier].Location.Edge
	combat := combatAction(state, "edge-kill", "team_b_p1")
	batch, err := ApplyCombatPulseCommit(state, combat, []Effect{damageEffect(state, combat, "team_b_p1", carrier, 100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if state.Bomb.Status != BombDropped || state.Bomb.Location.Edge == nil || state.Bomb.Location.Edge.EdgeID != "E1" {
		t.Fatalf("edge death did not leave bomb on edge: %+v", state.Bomb)
	}
	var kill *GameEvent
	for _, event := range batch.Events {
		if event.EventType == EventKill {
			kill = event
		}
	}
	if kill == nil || kill.Location == nil || kill.Location.Name != location.DisplayName || kill.Location.X != location.X || kill.Location.Y != location.Y {
		t.Fatalf("kill location did not use runtime edge location: event=%+v edge=%+v", kill, location)
	}
}
