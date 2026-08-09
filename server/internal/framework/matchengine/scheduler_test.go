package matchengine

import (
	"reflect"
	"strconv"
	"testing"
)

func TestSchedulerHeapIsDeterministicAcrossInsertionOrder(t *testing.T) {
	actions := []ScheduledAction{
		{ID: "e", Type: ActionMoveStart, StartAt: 0, ResolveAt: 8, Priority: 20, ActorIDs: []string{"p2"}},
		{ID: "d", Type: ActionCombatPulse, StartAt: 0, ResolveAt: 8, Priority: 100, ActorIDs: []string{"p3"}},
		{ID: "c", Type: ActionCombatPulse, StartAt: 0, ResolveAt: 8, Priority: 100, ActorIDs: []string{"p1"}},
		{ID: "b", Type: ActionCombatPulse, StartAt: 0, ResolveAt: 8, Priority: 100, ActorIDs: []string{"p1"}},
		{ID: "a", Type: ActionRoundExpire, StartAt: 0, ResolveAt: 7, Priority: 0},
	}
	orders := [][]int{{0, 1, 2, 3, 4}, {4, 3, 2, 1, 0}, {2, 0, 4, 1, 3}}
	var want []string
	for orderIndex, order := range orders {
		constants := makeTestMapConfig().CombatConstants
		scheduler := NewActionScheduler(constants)
		for _, index := range order {
			if err := scheduler.Schedule(actions[index]); err != nil {
				t.Fatalf("Schedule() error = %v", err)
			}
		}
		var got []string
		for scheduler.Len() > 0 {
			action, _ := scheduler.Pop()
			got = append(got, action.ID)
		}
		if orderIndex == 0 {
			want = got
		} else if !reflect.DeepEqual(got, want) {
			t.Fatalf("insertion order changed pop sequence: got=%v want=%v", got, want)
		}
	}
	if !reflect.DeepEqual(want, []string{"a", "b", "c", "d", "e"}) {
		t.Fatalf("unexpected ordering keys: %v", want)
	}
}

func TestEventProjectionSortAndStableIDs(t *testing.T) {
	actors := []string{"p2", "p1"}
	actionID := NewActionID(77, ActionCombatPulse, "intent", 3, 9, actors, 0)
	if actionID != NewActionID(77, ActionCombatPulse, "intent", 3, 9, []string{"p1", "p2"}, 0) {
		t.Fatal("actor input order changed ActionID")
	}
	if actionID == NewActionID(77, ActionCombatPulse, "intent", 3, 9, actors, 1) {
		t.Fatal("ordinal did not distinguish ActionID")
	}
	effectID := NewEffectID(77, actionID, EffectDamage, 0)
	eventID := NewEventID(77, actionID, effectID, EventDamage, 0)
	if effectID == "" || eventID == "" || eventID != NewEventID(77, actionID, effectID, EventDamage, 0) {
		t.Fatal("effect/event IDs are not stable")
	}
	state := makeTestRoundState(t, 77)
	lifecycle, err := newActionLifecycleEvent(state, ScheduledAction{ID: actionID, Type: ActionMoveStart, Priority: 20, ActorIDs: actors}, "MOVE_START", "movement started", 0)
	if err != nil {
		t.Fatal(err)
	}
	if lifecycle.SourceActionID == "" || lifecycle.SourceEffectID != "" || lifecycle.EventID == "" {
		t.Fatalf("action lifecycle provenance mismatch: %+v", lifecycle)
	}

	base := []*GameEvent{
		{EventID: "4", SourceActionID: "b", SourceEffectID: "a", Timestamp: 8, sortPriority: 80, sortActionType: "z", sortMinActorID: "p2"},
		{EventID: "3", SourceActionID: "a", SourceEffectID: "b", Timestamp: 8, sortPriority: 100, sortActionType: "b", sortMinActorID: "p2"},
		{EventID: "2", SourceActionID: "a", SourceEffectID: "a", Timestamp: 8, sortPriority: 100, sortActionType: "a", sortMinActorID: "p2"},
		{EventID: "1", SourceActionID: "a", SourceEffectID: "a", Timestamp: 7, sortPriority: 1, sortActionType: "z", sortMinActorID: "p9"},
	}
	permutations := [][]int{{0, 1, 2, 3}, {3, 2, 1, 0}, {1, 3, 0, 2}}
	for _, permutation := range permutations {
		events := make([]*GameEvent, 0, len(base))
		for _, index := range permutation {
			copy := *base[index]
			events = append(events, &copy)
		}
		sortEvents(events)
		got := []string{events[0].EventID, events[1].EventID, events[2].EventID, events[3].EventID}
		if !reflect.DeepEqual(got, []string{"1", "2", "3", "4"}) {
			t.Fatalf("event sort mismatch: %v", got)
		}
	}
}

func TestSchedulerActorVersionsAndFrozenGroupMinimum(t *testing.T) {
	state := makeTestRoundState(t, 201)
	actorID := "team_a_p1"
	action := ScheduledAction{ID: "stale", Type: ActionMovementArrive, ActorIDs: []string{actorID}, From: state.Players[actorID].Location, StartAt: 0, ResolveAt: 5, Priority: 20}
	if err := BeginExclusiveAction(state, &action, ActionMoving); err != nil {
		t.Fatal(err)
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		t.Fatal(err)
	}
	state.Players[actorID].Action.Version++
	if _, _, ok := state.Scheduler.PopNextValid(state); ok {
		t.Fatal("stale action was not silently discarded")
	}

	state = makeTestRoundState(t, 202)
	group := ScheduledAction{ID: "group-continue", Type: ActionMovementArrive, ActorIDs: []string{"team_a_p1", "team_a_p2", "team_a_p3"}, From: PlayerLocation{NodeID: "T_SPAWN"}, StartAt: 0, ResolveAt: 5, Priority: 20, MinRequiredActors: 2}
	if err := BeginExclusiveAction(state, &group, ActionMoving); err != nil {
		t.Fatal(err)
	}
	if err := state.Scheduler.Schedule(group); err != nil {
		t.Fatal(err)
	}
	state.Players["team_a_p3"].Alive = false
	state.Players["team_a_p3"].HP = 0
	got, actors, ok := state.Scheduler.PopNextValid(state)
	if !ok || got.ID != group.ID || !reflect.DeepEqual(actors, []string{"team_a_p1", "team_a_p2"}) {
		t.Fatalf("group at frozen minimum did not continue: action=%+v actors=%v ok=%v", got, actors, ok)
	}

	state = makeTestRoundState(t, 203)
	group.ID = "group-cancel"
	group.MinRequiredActors = 3
	if err := BeginExclusiveAction(state, &group, ActionMoving); err != nil {
		t.Fatal(err)
	}
	if err := state.Scheduler.Schedule(group); err != nil {
		t.Fatal(err)
	}
	state.Players["team_a_p3"].Alive = false
	state.Players["team_a_p3"].HP = 0
	if _, _, ok := state.Scheduler.PopNextValid(state); ok {
		t.Fatal("group below frozen minimum continued")
	}
	for _, actor := range []string{"team_a_p1", "team_a_p2"} {
		if state.Players[actor].Action.CurrentActionID != "" {
			t.Fatalf("cancelled group left actor %s busy", actor)
		}
	}
}

func TestNextTimeIncludesActionsDeadlinesAndTTL(t *testing.T) {
	state := makeTestRoundState(t, 204)
	state.Intel[SideT].Records = []IntelRecord{{ExpiresAt: 8}}
	state.Nodes["A_SITE"].KnownControl[SideT] = KnownControlState{Status: ControlCT, ExpiresAt: 6}
	if err := state.Scheduler.Schedule(ScheduledAction{ID: "later", Type: ActionMoveStart, StartAt: 0, ResolveAt: 10, Priority: 20}); err != nil {
		t.Fatal(err)
	}
	if at, kind := state.NextTime(); at != 6 || kind != NextTimeControl {
		t.Fatalf("NextTime() = %d/%s, want 6/%s", at, kind, NextTimeControl)
	}
	state.Timeline = 6
	if at, kind := state.NextTime(); at != 8 || kind != NextTimeIntel {
		t.Fatalf("NextTime() = %d/%s, want 8/%s", at, kind, NextTimeIntel)
	}
	state.Bomb.Status, state.BombDeadline = BombPlanted, 7
	if at, kind := state.NextTime(); at != 7 || kind != NextTimeBomb {
		t.Fatalf("bomb deadline did not supersede later work: %d/%s", at, kind)
	}
	state.Bomb.Status = BombCarried
	state.Scheduler = NewActionScheduler(state.constants)
	state.Intel[SideT].Records = nil
	state.Nodes["A_SITE"].KnownControl = map[string]KnownControlState{}
	state.Timeline, state.RoundDeadline = 0, 115
	decisionAt := state.RoundDeadline - state.constants.Int("ForceExecuteThreshold", 0)
	if at, kind := state.NextTime(); at != decisionAt || kind != NextTimeDecision {
		t.Fatalf("empty queue skipped force-execute decision deadline: %d/%s", at, kind)
	}
	state.Timeline = decisionAt
	if at, kind := state.NextTime(); at != 115 || kind != NextTimeRound {
		t.Fatalf("post-decision empty queue did not advance to round deadline: %d/%s", at, kind)
	}
}

func TestSchedulerAndRoundGuardsReturnStableCodes(t *testing.T) {
	state := makeTestRoundState(t, 205)
	setConstInt(&state.constants, "MaxStateTransitions", 1)
	setConstInt(&state.constants, "MaxNoOpTransitions", 1)
	setConstInt(&state.constants, "MaxRotationsPerTeam", 1)
	setConstInt(&state.constants, "MaxEffectsPerTimestamp", 1)
	setConstInt(&state.constants, "MaxRoundTimeline", 10)
	if err := state.RecordTransition(); err != nil {
		t.Fatal(err)
	}
	assertEngineErrorCode(t, state.RecordTransition(), "STATE_TRANSITION_LIMIT_EXCEEDED")
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	if state.NoOpCount != 1 || state.RecoveryAttempt.CycleID == "" || state.RecoveryAttempt.Status != RecoveryNotAttempted {
		t.Fatalf("NoOp threshold did not freeze and create recovery cycle: count=%d recovery=%+v", state.NoOpCount, state.RecoveryAttempt)
	}
	if err := state.RecordRotation(SideT); err != nil {
		t.Fatal(err)
	}
	assertEngineErrorCode(t, state.RecordRotation(SideT), "ROTATION_LIMIT_EXCEEDED")
	assertEngineErrorCode(t, state.ValidateEffectBatchSize(2), "EFFECT_LIMIT_EXCEEDED")
	assertEngineErrorCode(t, state.AdvanceTimeline(11), "TIMELINE_LIMIT_EXCEEDED")

	constants := makeTestMapConfig().CombatConstants
	setConstInt(&constants, "MaxScheduledActions", 1)
	scheduler := NewActionScheduler(constants)
	if err := scheduler.Schedule(ScheduledAction{ID: "one", Type: ActionMoveStart, ResolveAt: 1}); err != nil {
		t.Fatal(err)
	}
	assertEngineErrorCode(t, scheduler.Schedule(ScheduledAction{ID: "two", Type: ActionMoveStart, ResolveAt: 2}), "SCHEDULER_LIMIT_EXCEEDED")
}

func setConstInt(constants *CombatConstants, key string, value int) {
	entry := constants.Values[key]
	entry.ValueType = "Int"
	entry.Value = strconv.Itoa(value)
	constants.Values[key] = entry
}
