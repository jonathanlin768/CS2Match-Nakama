package matchengine

import (
	"reflect"
	"testing"
)

func TestEncounterCandidateFiltersBusyActorsAndThreatOrderIsStable(t *testing.T) {
	state := makeTestRoundState(t, 701)
	state.Players["team_a_p5"].EngagementID = "other"
	candidate, err := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_b_p2", "team_a_p5", "team_a_p2", "team_b_p1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidate.Actors) != 3 {
		t.Fatalf("busy actor was included: %+v", candidate.Actors)
	}
	for index := 1; index < len(candidate.Actors); index++ {
		previous, current := candidate.Actors[index-1], candidate.Actors[index]
		if previous.ThreatScore < current.ThreatScore || previous.ThreatScore == current.ThreatScore && previous.PlayerID > current.PlayerID {
			t.Fatalf("threat ordering is unstable: %+v", candidate.Actors)
		}
	}
	reordered, err := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1", "team_b_p2", "team_a_p5"})
	if err != nil || !reflect.DeepEqual(candidate, reordered) {
		t.Fatalf("actor input order changed candidate: equal=%v err=%v", reflect.DeepEqual(candidate, reordered), err)
	}
}

func TestEncounterArbitrationLocksSharedActorsAndAllowsDisjointConcurrency(t *testing.T) {
	state := makeTestRoundState(t, 702)
	a, _ := BuildEncounterCandidate(state, "a", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1"})
	shared, _ := BuildEncounterCandidate(state, "shared", "SCN_A", "A_LONG", []string{"team_a_p2", "team_b_p2"})
	b, _ := BuildEncounterCandidate(state, "b", "SCN_A", "B_SITE", []string{"team_a_p3", "team_b_p3"})
	overlap, _ := BuildEncounterCandidate(state, "b-overlap", "SCN_A", "B_SITE", []string{"team_a_p4", "team_b_p4"})
	accepted := ArbitrateEncounterCandidates([]EncounterCandidatePlan{shared, b, overlap, a})
	if len(accepted) != 2 {
		t.Fatalf("arbitration accepted shared actor twice or blocked disjoint area: %+v", accepted)
	}
	seenActors := map[string]bool{}
	for _, candidate := range accepted {
		for _, actor := range candidate.Actors {
			if seenActors[actor.PlayerID] {
				t.Fatalf("shared actor %s accepted twice", actor.PlayerID)
			}
			seenActors[actor.PlayerID] = true
		}
	}
	for _, candidate := range accepted {
		if _, err := StartEncounter(state, candidate); err != nil {
			t.Fatal(err)
		}
	}
	if len(state.ActiveEngagements) != 2 {
		t.Fatalf("disjoint encounters did not run concurrently: %d", len(state.ActiveEngagements))
	}
}

func TestEncounterStartOnlySchedulesFuturePulsesAndDoesNotBlockOtherActions(t *testing.T) {
	state := makeTestRoundState(t, 703)
	candidate, err := BuildEncounterCandidate(state, "a-contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_a_p3", "team_b_p1", "team_b_p2"})
	if err != nil {
		t.Fatal(err)
	}
	beforeEvents := len(state.Events)
	schedule, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if len(schedule.PulseActions) < 1 || len(schedule.PulseActions) > state.constants.Int("MaxEncounterPulses", 0) || schedule.EndAction.ResolveAt <= state.Timeline {
		t.Fatalf("invalid encounter schedule: %+v", schedule)
	}
	if len(state.Events) != beforeEvents {
		t.Fatal("Encounter.Start pre-generated future combat events")
	}
	for _, pulse := range schedule.PulseActions {
		if pulse.ResolveAt <= state.Timeline || pulse.ResolveAt > schedule.EndAction.ResolveAt {
			t.Fatalf("pulse is not at an absolute time: %+v", pulse)
		}
	}
	carrier := state.Players[state.Bomb.CarrierID]
	carrier.Location = PlayerLocation{NodeID: "B_SITE"}
	state.Bomb.Location = carrier.Location
	plant := ScheduledAction{ID: "b-plant", Type: ActionPlantComplete, ActorIDs: []string{carrier.Profile.PlayerID}, From: carrier.Location, StartAt: 0, ResolveAt: 4, Priority: 70, MinRequiredActors: 1}
	if err := BeginExclusiveAction(state, &plant, ActionPlanting); err != nil {
		t.Fatalf("A encounter globally blocked disjoint B Plant action start: %v", err)
	}
	if err := state.Scheduler.Schedule(plant); err != nil {
		t.Fatal(err)
	}
	if err := state.Bomb.StartPlant(carrier.Profile.PlayerID, plant.ID, "B", 0, 4); err != nil {
		t.Fatalf("A encounter globally blocked disjoint B plant: %v", err)
	}
	if state.Phase != PhaseOpeningDeploy {
		t.Fatal("encounter mutated the global phase into an action lock")
	}
}

func TestCombatEndIsLocalReleasesActorsAndNeverSetsRoundWinner(t *testing.T) {
	state := makeTestRoundState(t, 704)
	candidate, _ := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1"})
	schedule, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = schedule.EndAction.ResolveAt
	victim := state.Players["team_b_p1"]
	victim.HP, victim.Alive = 0, false
	result, err := EndEncounter(state, candidate.ID, "local_elimination")
	if err != nil {
		t.Fatal(err)
	}
	if result.LocalWinnerSide != SideT || !result.DecisionTriggered || state.Terminal != nil || len(state.ActiveEngagements) != 0 {
		t.Fatalf("CombatEnd became a round terminal or failed cleanup: result=%+v terminal=%+v", result, state.Terminal)
	}
	if state.Players["team_a_p2"].EngagementID != "" || state.Players["team_a_p2"].Action.CurrentActionID != "" || state.Nodes["A_SITE"].ActualControl != ControlT {
		t.Fatalf("CombatEnd did not release/update local control: player=%+v node=%+v", state.Players["team_a_p2"], state.Nodes["A_SITE"])
	}
	if !state.Players["team_b_p3"].Alive {
		t.Fatal("local elimination killed an unrelated round actor")
	}
}

func TestEncounterExitChecksPulseDurationAndLocalElimination(t *testing.T) {
	state := makeTestRoundState(t, 705)
	candidate, _ := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1"})
	_, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	encounter := state.ActiveEngagements[candidate.ID]
	encounter.PulsesResolved = encounter.MaxPulses
	if end, reason := EncounterShouldEnd(state, encounter); !end || reason != "pulse_limit" {
		t.Fatalf("pulse limit not enforced: %v/%s", end, reason)
	}
	encounter.PulsesResolved = 0
	state.Timeline = encounter.EndsAt
	if end, reason := EncounterShouldEnd(state, encounter); !end || reason != "duration_limit" {
		t.Fatalf("duration limit not enforced: %v/%s", end, reason)
	}
	state.Timeline = 0
	state.Players["team_b_p1"].Alive, state.Players["team_b_p1"].HP = false, 0
	if end, reason := EncounterShouldEnd(state, encounter); !end || reason != "local_elimination" {
		t.Fatalf("local elimination not enforced: %v/%s", end, reason)
	}
}
