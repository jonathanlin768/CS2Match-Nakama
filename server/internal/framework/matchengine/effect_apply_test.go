package matchengine

import (
	"reflect"
	"testing"
)

func combatAction(state *RoundState, id string, actors ...string) ScheduledAction {
	return ScheduledAction{ID: id, Type: ActionCombatPulse, ActorIDs: actors, StartAt: state.Timeline, ResolveAt: state.Timeline, Priority: PriorityCombatPulseCommit}
}

func damageEffect(state *RoundState, action ScheduledAction, actor, target string, amount, ordinal int) Effect {
	return Effect{ID: NewEffectID(state.Seed, action.ID, EffectDamage, ordinal), SourceActionID: action.ID, Type: EffectDamage, Priority: PriorityCombatPulseCommit, Timestamp: state.Timeline, ActorID: actor, TargetID: target, Amount: amount}
}

func TestCombatPulseCommitAllowsSamePulseTrade(t *testing.T) {
	state := makeTestRoundState(t, 301)
	tID, ctID := "team_a_p2", "team_b_p1"
	action := combatAction(state, "pulse-trade", tID, ctID)
	batch, err := ApplyCombatPulseCommit(state, action, []Effect{
		damageEffect(state, action, tID, ctID, 100, 0),
		damageEffect(state, action, ctID, tID, 100, 1),
	})
	if err != nil {
		t.Fatalf("ApplyCombatPulseCommit() error = %v", err)
	}
	if state.Players[tID].Alive || state.Players[ctID].Alive || state.Players[tID].Kills != 1 || state.Players[ctID].Kills != 1 {
		t.Fatalf("same-pulse trade was not atomic: T=%+v CT=%+v", state.Players[tID], state.Players[ctID])
	}
	if countEvents(batch.Events, EventKill) != 2 {
		t.Fatalf("trade emitted %d kill events", countEvents(batch.Events, EventKill))
	}
}

func TestLethalHitDoesNotCancelSnapshotNonLethalRetaliation(t *testing.T) {
	state := makeTestRoundState(t, 302)
	t1, t2, ct := "team_a_p2", "team_a_p3", "team_b_p1"
	action := combatAction(state, "pulse-retaliation", t1, ct)
	_, err := ApplyCombatPulseCommit(state, action, []Effect{
		damageEffect(state, action, t1, ct, 100, 0),
		damageEffect(state, action, ct, t2, 35, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[ct].Alive || state.Players[t2].HP != 65 || state.Players[ct].Damage != 35 {
		t.Fatalf("snapshot retaliation was cancelled: CT=%+v target=%+v", state.Players[ct], state.Players[t2])
	}
}

func TestMultiAttackerActualDamageAndKillAttributionAreStable(t *testing.T) {
	t1, t2, target := "team_a_p2", "team_a_p3", "team_b_p1"
	run := func(reverse bool) (*RoundState, *AppliedBatch) {
		state := makeTestRoundState(t, 303)
		action := combatAction(state, "pulse-multi", t1, t2)
		effects := []Effect{
			damageEffect(state, action, t1, target, 80, 0),
			damageEffect(state, action, t2, target, 40, 1),
		}
		if reverse {
			effects[0], effects[1] = effects[1], effects[0]
		}
		batch, err := ApplyCombatPulseCommit(state, action, effects)
		if err != nil {
			t.Fatal(err)
		}
		return state, batch
	}
	left, leftBatch := run(false)
	right, rightBatch := run(true)
	if left.Players[t1].Damage != 67 || left.Players[t2].Damage != 33 || left.Players[t1].Kills != 1 {
		t.Fatalf("proportional actual damage mismatch: %d/%d kills=%d", left.Players[t1].Damage, left.Players[t2].Damage, left.Players[t1].Kills)
	}
	if left.Players[t1].Damage != right.Players[t1].Damage || left.Players[t2].Damage != right.Players[t2].Damage || eventIDs(leftBatch.Events) == nil || !reflect.DeepEqual(eventIDs(leftBatch.Events), eventIDs(rightBatch.Events)) {
		t.Fatal("input effect order changed atomic commit output")
	}
}

func TestCombatDeathInvalidatesExternalActionAndDropsBomb(t *testing.T) {
	for _, status := range []ActionStatus{ActionMoving, ActionPlanting, ActionDefusing} {
		t.Run(string(status), func(t *testing.T) {
			state := makeTestRoundState(t, 304)
			victim := "team_a_p1"
			if status == ActionDefusing {
				victim = "team_b_p1"
				state.Players[state.Bomb.CarrierID].HasBomb = false
				state.Bomb = BombState{Status: BombDefusing, Location: PlayerLocation{NodeID: "A_SITE"}, PlantedSite: "A", PlantedAt: 10, ExplodeAt: 40, DefuseActorID: victim, DefuseActionID: "external", DefuseStartAt: 20, DefuseFinishAt: 30}
			}
			player := state.Players[victim]
			player.Action = PlayerActionState{CurrentActionID: "external", Version: 4, Status: status, BusyUntil: 20}
			action := combatAction(state, "pulse-interrupt", "team_b_p2")
			_, err := ApplyCombatPulseCommit(state, action, []Effect{damageEffect(state, action, "team_b_p2", victim, 100, 0)})
			if err != nil {
				t.Fatal(err)
			}
			if player.Action.CurrentActionID != "" || player.Action.Version != 5 || player.Alive {
				t.Fatalf("death did not invalidate %s: %+v", status, player.Action)
			}
			if status == ActionPlanting && (state.Bomb.Status != BombDropped || state.Bomb.CarrierID != "" || player.HasBomb) {
				t.Fatalf("planting carrier death did not drop bomb: %+v", state.Bomb)
			}
		})
	}
}

func TestSameSecondBombOrderingContracts(t *testing.T) {
	location := PlayerLocation{NodeID: "A_SITE"}
	bomb := BombState{Status: BombPlanted, Location: location, PlantedSite: "A", PlantedAt: 10, ExplodeAt: 40}
	if err := bomb.StartDefuse("ct1", "defuse", 30, 40); err != nil {
		t.Fatal(err)
	}
	if err := bomb.CompleteDefuse(40); err != nil {
		t.Fatalf("DefuseComplete == Explode must succeed: %v", err)
	}
	if err := bomb.Explode(40); err == nil {
		t.Fatal("explosion remained valid after equal-time defuse")
	}

	state := makeTestRoundState(t, 305)
	carrier := state.Bomb.CarrierID
	state.RoundDeadline = 40
	if err := state.Bomb.StartPlant(carrier, "plant", "A", 36, 40); err != nil {
		t.Fatal(err)
	}
	if err := state.Bomb.CompletePlant(location, 40, 80); err != nil {
		t.Fatalf("PlantComplete == RoundExpire must apply first: %v", err)
	}
	if state.Bomb.Status != BombPlanted {
		t.Fatalf("equal-time plant did not supersede round expiry: %s", state.Bomb.Status)
	}
}

func TestDefuserDeathMakesEqualTimeCompletionStale(t *testing.T) {
	state := makeTestRoundState(t, 306)
	carrier := state.Bomb.CarrierID
	state.Players[carrier].HasBomb = false
	defuser := "team_b_p1"
	state.Bomb = BombState{Status: BombPlanted, Location: PlayerLocation{NodeID: "A_SITE"}, PlantedSite: "A", PlantedAt: 0, ExplodeAt: 40}
	if err := state.Bomb.StartDefuse(defuser, "defuse-complete", 30, 40); err != nil {
		t.Fatal(err)
	}
	state.Timeline = 40
	defuse := ScheduledAction{ID: "defuse-complete", Type: ActionDefuseComplete, ActorIDs: []string{defuser}, From: state.Players[defuser].Location, StartAt: 30, ResolveAt: 40, Priority: 70}
	if err := BeginExclusiveAction(state, &defuse, ActionDefusing); err != nil {
		t.Fatal(err)
	}
	if err := state.Scheduler.Schedule(defuse); err != nil {
		t.Fatal(err)
	}
	combat := combatAction(state, "pulse-defuser-death", "team_a_p2")
	if _, err := ApplyCombatPulseCommit(state, combat, []Effect{damageEffect(state, combat, "team_a_p2", defuser, 100, 0)}); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := state.Scheduler.PopNextValid(state); ok {
		t.Fatal("dead defuser's equal-time completion remained valid")
	}
	if state.Bomb.Status != BombPlanted {
		t.Fatalf("stale completion mutated bomb: %s", state.Bomb.Status)
	}
}

func countEvents(events []*GameEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

func eventIDs(events []*GameEvent) []string {
	ids := make([]string, len(events))
	for index, event := range events {
		ids[index] = event.EventID
	}
	return ids
}
