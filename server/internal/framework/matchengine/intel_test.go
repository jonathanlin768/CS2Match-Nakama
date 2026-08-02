package matchengine

import (
	"reflect"
	"testing"
)

func TestIntelSourcesConfidenceAndImmediateSharing(t *testing.T) {
	state := makeTestRoundState(t, 501)
	state.Timeline = 10
	observer := "team_a_p2"
	target := "team_b_p1"
	direct, err := RecordIntel(state, SideT, IntelObservation{Type: IntelDirectVisibility, TargetID: target, NodeID: "A_SITE", ObservedBy: []string{observer}, At: 10, TTL: 10})
	if err != nil || direct.Confidence != 100 {
		t.Fatalf("direct intel = %+v/%v", direct, err)
	}
	death, err := RecordIntel(state, SideT, IntelObservation{Type: IntelDeath, TargetID: target, NodeID: "A_SITE", SourceActionID: "combat", ObservedBy: []string{observer}, Confidence: 100, At: 10, TTL: 10})
	if err != nil || death.Confidence != 70 {
		t.Fatalf("death intel cap = %+v/%v", death, err)
	}
	empty, err := RecordIntel(state, SideT, IntelObservation{Type: IntelEmptySite, NodeID: "B_SITE", At: 10, TTL: 10})
	if err != nil || empty.Confidence != 20 {
		t.Fatalf("empty-site golden confidence = %+v/%v", empty, err)
	}

	soundEvent := &GameEvent{EventID: "real-sound", EventType: "MOVE_SOUND", SourceActionID: "move"}
	state.Events = append(state.Events, soundEvent)
	sound, err := RecordIntel(state, SideT, IntelObservation{Type: IntelSound, TargetID: target, AreaID: "Mid", SourceEventID: soundEvent.EventID, ObservedBy: []string{observer}, Confidence: 5, At: 10, TTL: 10})
	if err != nil || sound.Confidence != 30 || sound.NodeID != "" || sound.AreaID != "Mid" {
		t.Fatalf("sound intel = %+v/%v", sound, err)
	}
	if _, err := RecordIntel(state, SideT, IntelObservation{Type: IntelSound, TargetID: "invented", AreaID: "Mid", SourceEventID: soundEvent.EventID, ObservedBy: []string{observer}, Confidence: 50, At: 10, TTL: 10}); err == nil {
		t.Fatal("sound intel invented a nonexistent enemy")
	}
	if _, err := RecordIntel(state, SideT, IntelObservation{Type: IntelSound, TargetID: target, NodeID: "MID", AreaID: "Mid", SourceEventID: soundEvent.EventID, ObservedBy: []string{observer}, Confidence: 50, At: 10, TTL: 10}); err == nil {
		t.Fatal("sound intel leaked an exact node")
	}
	view, err := BuildDecisionView(state, SideT)
	if err != nil || len(view.Intel) != 4 {
		t.Fatalf("team-wide immediate view = %d records/%v", len(view.Intel), err)
	}
}

func TestIntelDecayObserverDeathAndKnownControlTTL(t *testing.T) {
	state := makeTestRoundState(t, 502)
	state.Timeline = 5
	observer := "team_a_p2"
	record, err := RecordIntel(state, SideT, IntelObservation{Type: IntelDirectVisibility, TargetID: "team_b_p1", NodeID: "A_SITE", ObservedBy: []string{observer}, Confidence: 90, At: 5, TTL: 3})
	if err != nil {
		t.Fatal(err)
	}
	DegradeIntelForObserverDeath(state, SideT, observer)
	degraded := state.Intel[SideT].KnownEnemies[record.TargetID]
	if degraded.Confidence != 70 || len(degraded.ObservedBy) != 0 {
		t.Fatalf("observer death did not degrade intel: %+v", degraded)
	}
	state.Nodes["A_SITE"].KnownControl[SideT] = KnownControlState{Status: ControlCT, UpdatedAt: 5, ExpiresAt: 8, ObservedBy: []string{observer}}
	if err := DecayIntelAndControl(state, 8); err != nil {
		t.Fatal(err)
	}
	if len(state.Intel[SideT].Records) != 0 || state.Nodes["A_SITE"].KnownControl[SideT].Status != ControlUnknown {
		t.Fatalf("priority-10 decay left expired knowledge: intel=%v control=%+v", state.Intel[SideT].Records, state.Nodes["A_SITE"].KnownControl[SideT])
	}
}

func TestDecisionViewExcludesEnemyHiddenStateAndActualControl(t *testing.T) {
	state := makeTestRoundState(t, 503)
	state.Timeline = 12
	enemy := state.Players["team_b_p1"]
	enemy.Location = PlayerLocation{NodeID: "B_SITE"}
	enemy.Intent = Intent{ID: "secret-intent", Type: IntentMove, TargetID: "A_SITE"}
	enemy.Action = PlayerActionState{CurrentActionID: "secret-action", Version: 7, Status: ActionMoving}
	state.Nodes["B_SITE"].ActualControl = ControlCT
	state.Nodes["B_SITE"].KnownControl[SideT] = KnownControlState{Status: ControlUnknown, UpdatedAt: 1, ExpiresAt: 20}
	view, err := BuildDecisionView(state, SideT)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.OwnPlayers) != 5 || len(view.KnownControls) != 0 || len(view.Intel) != 0 {
		t.Fatalf("hidden reinforcement leaked through decision view: %+v", view)
	}
	for _, player := range view.OwnPlayers {
		if player.PlayerID == enemy.Profile.PlayerID || player.Intent.ID == "secret-intent" || player.Action.CurrentActionID == "secret-action" {
			t.Fatal("enemy hidden intent/action leaked through own player view")
		}
	}
	view.OwnPlayers[0].Location.NodeID = "mutated"
	if state.Players[view.OwnPlayers[0].PlayerID].Location.NodeID == "mutated" {
		t.Fatal("DecisionView retains a mutable RoundState reference")
	}
}

func TestLowConfidenceOnlyChangesBoundedScoreAndBombIntelDoesNotLeak(t *testing.T) {
	state := makeTestRoundState(t, 504)
	state.Timeline = 10
	low, err := RecordIntel(state, SideCT, IntelObservation{Type: IntelEmptySite, NodeID: "A_SITE", Confidence: 20, At: 10, TTL: 10})
	if err != nil {
		t.Fatal(err)
	}
	if IntelScoreModifier(low, 5) != 1 || CanTriggerDeterministicIntelAction(low, state.constants) {
		t.Fatalf("low confidence became deterministic: modifier=%v", IntelScoreModifier(low, 5))
	}
	if _, err := RecordIntel(state, SideCT, IntelObservation{Type: IntelBomb, NodeID: "T_SPAWN", SourceEventID: "not-real", ObservedBy: []string{"team_b_p1"}, Confidence: 100, At: 10, TTL: 10}); err == nil {
		t.Fatal("unobserved bomb leaked exact node")
	}
	view, err := BuildDecisionView(state, SideCT)
	if err != nil || view.BombIntel != nil {
		t.Fatalf("unobserved bomb appeared in decision view: %+v/%v", view.BombIntel, err)
	}
}

func TestDecisionViewAndIdentityRollAreStableAcrossMapOrder(t *testing.T) {
	state := makeTestRoundState(t, 505)
	state.Timeline = 1
	_, err := RecordIntel(state, SideT, IntelObservation{Type: IntelDirectVisibility, TargetID: "team_b_p1", NodeID: "A_SITE", ObservedBy: []string{"team_a_p2"}, At: 1, TTL: 10})
	if err != nil {
		t.Fatal(err)
	}
	first, err := BuildDecisionView(state, SideT)
	if err != nil {
		t.Fatal(err)
	}
	state.Nodes = reverseNodeMap(state.Nodes)
	state.Players = reversePlayerMap(state.Players)
	second, err := BuildDecisionView(state, SideT)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("map order changed DecisionView: equal=%v err=%v", reflect.DeepEqual(first, second), err)
	}
	rolls := IdentityRollSource{Seed: state.Seed}
	if rolls.Unit("decision", "p1") != rolls.Unit("decision", "p1") || rolls.Unit("decision", "p1") == rolls.Unit("decision", "p2") {
		t.Fatal("identity-derived RollSource is not stable/independent")
	}
}

func reverseNodeMap(source map[string]*NodeRuntimeState) map[string]*NodeRuntimeState {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	result := map[string]*NodeRuntimeState{}
	for index := len(keys) - 1; index >= 0; index-- {
		result[keys[index]] = source[keys[index]]
	}
	return result
}

func reversePlayerMap(source map[string]*RoundPlayerState) map[string]*RoundPlayerState {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	result := map[string]*RoundPlayerState{}
	for index := len(keys) - 1; index >= 0; index-- {
		result[keys[index]] = source[keys[index]]
	}
	return result
}
