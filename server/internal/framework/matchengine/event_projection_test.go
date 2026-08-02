package matchengine

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestReasonProjectionPreservesAuditFieldsAndZeroProbability(t *testing.T) {
	zero := 0.0
	record := ReasonRecord{
		Code: "AUDIT", MainFactor: "aim", ScoreDelta: 1.25, Modifiers: []ReasonModifier{{Code: "cover", Value: -0.5}},
		Probability: &zero, Formula: "score = aim + cover", Inputs: map[string]float64{"aim": 80, "cover": -0.5},
		StateChanges: []ReasonStateChange{{Field: "player.hp", Before: NumberReasonValue(100), After: NumberReasonValue(75)}}, Detail: "actual applied effect",
	}
	reason, err := ProjectReasonRecord(record, "act-1", "eff-1")
	if err != nil {
		t.Fatal(err)
	}
	if reason.ScoreDelta != 1.25 || reason.Probability == nil || *reason.Probability != 0 || reason.SourceActionID != "act-1" || reason.SourceEffectID != "eff-1" || reason.Formula == "" || len(reason.Inputs) != 2 || len(reason.StateChanges) != 1 {
		t.Fatalf("reason projection lost audit fields: %+v", reason)
	}
	encoded, err := json.Marshal(reason)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"probability":0`) {
		t.Fatalf("actual zero probability was omitted: %s", encoded)
	}
	ruleReason, err := ProjectReasonRecord(ReasonRecord{Code: "RULE", Source: "terminal"}, "act-rule", "")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ = json.Marshal(ruleReason)
	if strings.Contains(string(encoded), "probability") {
		t.Fatalf("non-applicable probability was serialized: %s", encoded)
	}
}

func TestReasonProjectionRejectsNonCanonicalOrHiddenStateChanges(t *testing.T) {
	number, text := 1.0, "bad"
	tests := []ReasonRecord{
		{Code: "BAD", StateChanges: []ReasonStateChange{{Field: "player.hp", Before: ReasonValue{Kind: "Number", Number: &number, String: &text}, After: NumberReasonValue(0)}}},
		{Code: "HIDDEN", StateChanges: []ReasonStateChange{{Field: "enemy.Intent", Before: NullReasonValue(), After: StringReasonValue("Move")}}},
	}
	for _, record := range tests {
		if _, err := ProjectReasonRecord(record, "act", "eff"); err == nil {
			t.Fatalf("invalid reason accepted: %+v", record)
		}
	}
}

func TestEventLocationUsesSemanticSourceAndIdentitySeed(t *testing.T) {
	state := makeTestRoundState(t, 1601)
	location := PlayerLocation{NodeID: "A_SITE"}
	first := eventLocation(state, location, "evt-1", "eff-1")
	second := eventLocation(state, location, "evt-1", "eff-1")
	if !reflect.DeepEqual(first, second) || first.SourceType != "Area" || first.SourceID != "A_SITE" || first.Seed == 0 {
		t.Fatalf("node area location is not identity-derived: %+v/%+v", first, second)
	}
	edge, err := ResolveOnEdgeLocation(state, state.mapEdges["E2"], "LONG_DOOR", "A_LONG", 0.4, "edge-event")
	if err != nil {
		t.Fatal(err)
	}
	onEdge := eventLocation(state, PlayerLocation{Edge: edge}, "evt-edge", "eff-edge")
	if onEdge.SourceType != "OnEdge" || onEdge.SourceID != "E2" || onEdge.X != edge.X || onEdge.Y != edge.Y {
		t.Fatalf("OnEdge source was replaced by a route/final location: %+v", onEdge)
	}
}

func TestCausalRoundEventsHaveStableSourcesSnapshotsAndReport(t *testing.T) {
	result, err := runCausalRound(context.Background(), makeTestRoundInput(1602))
	if err != nil {
		t.Fatal(err)
	}
	if result.Round.Report == nil || len(result.Round.Report.KeyEvents) == 0 || !strings.Contains(result.Round.Report.StrategySummary, result.Round.StrategyTemplateID) {
		t.Fatalf("missing explainable report: %+v", result.Round.Report)
	}
	for _, event := range result.Round.Events {
		if event.Reason == nil {
			continue
		}
		if event.EventID == "" || event.SourceActionID == "" || event.Reason.SourceActionID != event.SourceActionID {
			t.Fatalf("reason event has unstable/mismatched source: %+v", event)
		}
		if oneOf(event.EventType, EventDamage, EventKill, EventBombDrop, EventBombPlant, EventBombDefuse, EventBombExplode) && event.SourceEffectID == "" {
			t.Fatalf("effect-derived event has no effect source: %+v", event)
		}
		if oneOf(event.EventType, EventRoundStart, EventBombPlant, EventBombDefuse, EventBombExplode, EventRoundEnd) && event.State == nil {
			t.Fatalf("critical event has no public snapshot: %+v", event)
		}
	}
	for _, keyEvent := range result.Round.Report.KeyEvents {
		found := false
		for _, event := range result.Round.Events {
			if keyEvent == event {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("report fabricated or copied a key event outside Round.Events")
		}
	}
}

func TestPlantAndDefuseInterruptEventsComeFromRealActionLifecycle(t *testing.T) {
	plantState, carrier := preparePlantState(t, 1603, "A")
	plant, err := StartPlantAction(plantState, carrier, "A", "plant-lifecycle", 0)
	if err != nil {
		t.Fatal(err)
	}
	plantState.Timeline++
	combat := combatAction(plantState, "plant-lifecycle-hit", "team_b_p1")
	if _, err := ApplyCombatPulseCommit(plantState, combat, []Effect{damageEffect(plantState, combat, "team_b_p1", carrier, 1, 0)}); err != nil {
		t.Fatal(err)
	}
	assertLifecycleEvent(t, plantState.Events, EventPlantStart, plant.ID)
	assertLifecycleEvent(t, plantState.Events, EventPlantInterrupt, plant.ID)

	defuseState := preparePostPlantState(t, 1604)
	defuseState.Timeline = defuseState.BombDeadline - CalculateDefuseTime(defuseState, "team_b_p1")
	defuse, err := StartDefuseAction(defuseState, "team_b_p1", "defuse-lifecycle", 0)
	if err != nil {
		t.Fatal(err)
	}
	defuseState.Timeline++
	combat = combatAction(defuseState, "defuse-lifecycle-hit", "team_a_p2")
	defuseState.Players["team_a_p2"].Alive, defuseState.Players["team_a_p2"].HP = true, 100
	if _, err := ApplyCombatPulseCommit(defuseState, combat, []Effect{damageEffect(defuseState, combat, "team_a_p2", "team_b_p1", 1, 0)}); err != nil {
		t.Fatal(err)
	}
	assertLifecycleEvent(t, defuseState.Events, EventDefuseStart, defuse.ID)
	assertLifecycleEvent(t, defuseState.Events, EventDefuseInterrupt, defuse.ID)
}

func assertLifecycleEvent(t *testing.T, events []*GameEvent, eventType, actionID string) {
	t.Helper()
	for _, event := range events {
		if event.EventType == eventType && event.SourceActionID == actionID && event.Reason != nil {
			return
		}
	}
	t.Fatalf("missing %s lifecycle event for %s", eventType, actionID)
}
