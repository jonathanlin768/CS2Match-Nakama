package matchengine

import "testing"

func preparePlantState(t *testing.T, seed int64, site string) (*roundState, string) {
	t.Helper()
	state := makeTestRoundState(t, seed)
	carrierID := state.Bomb.CarrierID
	nodeID := site + "_SITE"
	carrier := state.Players[carrierID]
	carrier.Location = playerLocation{NodeID: nodeID}
	state.Bomb.Location = carrier.Location
	state.Nodes[nodeID].ActualControl = controlUnknown
	return state, carrierID
}

func TestSiteContestChoosesThreatEncounterOrPlantWindow(t *testing.T) {
	state, carrier := preparePlantState(t, 1001, "A")
	state.Nodes["A_SITE"].ActualControl = controlCT
	decision, err := planSiteContest(state, "A", []string{carrier})
	if err != nil || decision.Type != siteContestEncounter {
		t.Fatalf("CT-controlled site did not require encounter: %+v/%v", decision, err)
	}
	state.Nodes["A_SITE"].ActualControl = controlUnknown
	for _, player := range state.Players {
		if player.Side == SideCT {
			player.Location = playerLocation{NodeID: "CT_SPAWN"}
		}
	}
	decision, err = planSiteContest(state, "A", []string{carrier})
	if err != nil || decision.Type != siteContestPlant || decision.PlantScore < decision.PlantRisk {
		t.Fatalf("clear site did not create plant decision: %+v/%v", decision, err)
	}
}

func TestPlantCompleteAtRoundDeadlineWinsOrderingAndSchedulesExplosion(t *testing.T) {
	state, carrier := preparePlantState(t, 1002, "A")
	state.Timeline = 111
	action, err := startPlantAction(state, carrier, "A", "plant-a", 0)
	if err != nil {
		t.Fatal(err)
	}
	if action.ResolveAt != state.RoundDeadline {
		t.Fatalf("test plant does not finish at RoundDeadline: %d/%d", action.ResolveAt, state.RoundDeadline)
	}
	state.Timeline = action.ResolveAt
	result, err := resolvePlantComplete(state, action)
	if err != nil || !result.Applied || result.FollowUp == nil {
		t.Fatalf("equal-time plant failed: %+v/%v", result, err)
	}
	if state.Bomb.Status != bombPlanted || state.BombDeadline != state.Timeline+state.constants.Int("BombExplodeTime", 0) || result.FollowUp.ResolveAt != state.BombDeadline || state.Phase != phasePostPlant {
		t.Fatalf("plant did not switch authoritative bomb clock: bomb=%+v deadline=%d follow=%+v", state.Bomb, state.BombDeadline, result.FollowUp)
	}
	if at, kind := state.NextTime(); at != state.BombDeadline || kind != nextTimeBomb {
		t.Fatalf("RoundDeadline remained active after plant: %d/%s", at, kind)
	}
}

func TestPlantingCarrierDeathSameSecondDropsBombAndInvalidatesComplete(t *testing.T) {
	state, carrier := preparePlantState(t, 1003, "A")
	action, err := startPlantAction(state, carrier, "A", "plant-death", 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = action.ResolveAt
	combat := combatAction(state, "plant-death-pulse", "team_b_p1")
	if _, err := applyCombatPulseCommit(state, combat, []effect{damageEffect(state, combat, "team_b_p1", carrier, 100, 0)}); err != nil {
		t.Fatal(err)
	}
	result, err := resolvePlantComplete(state, action)
	if err != nil || result.Applied || state.Bomb.Status != bombDropped || state.Players[carrier].HasBomb {
		t.Fatalf("death did not precede plant complete: result=%+v bomb=%+v err=%v", result, state.Bomb, err)
	}
}

func TestNonlethalDamageInterruptsPlantWithoutDroppingBomb(t *testing.T) {
	state, carrier := preparePlantState(t, 1004, "A")
	action, err := startPlantAction(state, carrier, "A", "plant-interrupt", 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = action.StartAt + 1
	combat := combatAction(state, "plant-hit", "team_b_p1")
	if _, err := applyCombatPulseCommit(state, combat, []effect{damageEffect(state, combat, "team_b_p1", carrier, 10, 0)}); err != nil {
		t.Fatal(err)
	}
	if state.Bomb.Status != bombCarried || !state.Players[carrier].HasBomb || state.Players[carrier].Action.CurrentActionID != "" {
		t.Fatalf("nonlethal plant interruption mismatch: bomb=%+v player=%+v", state.Bomb, state.Players[carrier])
	}
}

func TestDroppedBombRecoveryConsumesPathAndPickupTime(t *testing.T) {
	state := makeTestRoundState(t, 1005)
	oldCarrier := state.Bomb.CarrierID
	state.Players[oldCarrier].HasBomb = false
	if err := state.Bomb.Drop(playerLocation{NodeID: "A_LONG"}, 0); err != nil {
		t.Fatal(err)
	}
	state.Nodes["A_LONG"].ActualControl = controlUnknown
	recoverer := "team_a_p2"
	schedule, feedback, err := scheduleBombRecovery(state, recoverer, "recover", 0)
	if err != nil || feedback != nil {
		t.Fatalf("ScheduleBombRecovery() = %+v/%+v/%v", schedule, feedback, err)
	}
	if schedule.MoveDuration != 16 || schedule.PickupDuration != 2 || schedule.Action.ResolveAt != 18 {
		t.Fatalf("recovery did not consume Move+Pickup time: %+v", schedule)
	}
	state.Timeline = schedule.Action.ResolveAt
	result, err := resolveBombPickup(state, schedule.Action)
	if err != nil || !result.Applied || state.Bomb.CarrierID != recoverer || !state.Players[recoverer].HasBomb || state.Players[recoverer].Location.NodeID != "A_LONG" {
		t.Fatalf("bomb pickup failed: result=%+v bomb=%+v player=%+v err=%v", result, state.Bomb, state.Players[recoverer], err)
	}
}

func TestBombRecoveryRequiresContestAndReachability(t *testing.T) {
	state := makeTestRoundState(t, 1006)
	state.Players[state.Bomb.CarrierID].HasBomb = false
	if err := state.Bomb.Drop(playerLocation{NodeID: "A_SITE"}, 0); err != nil {
		t.Fatal(err)
	}
	state.Nodes["A_SITE"].ActualControl = controlCT
	_, feedback, err := scheduleBombRecovery(state, "team_a_p2", "recover", 0)
	if err != nil || feedback == nil || feedback.Code != "SITE_CONTEST_REQUIRED" {
		t.Fatalf("enemy-controlled bomb skipped contest: %+v/%v", feedback, err)
	}
	state.Nodes["A_SITE"].ActualControl = controlUnknown
	state.mapEdges = map[string]MapEdge{}
	_, feedback, err = scheduleBombRecovery(state, "team_a_p2", "recover", 0)
	if err != nil || feedback == nil || feedback.Code != "UNREACHABLE" {
		t.Fatalf("unreachable bomb was teleported: %+v/%v", feedback, err)
	}
}

func preparePostPlantState(t *testing.T, seed int64) *roundState {
	t.Helper()
	state, carrier := preparePlantState(t, seed, "A")
	action, err := startPlantAction(state, carrier, "A", "plant", 0)
	if err != nil {
		t.Fatal(err)
	}
	state.Timeline = action.ResolveAt
	if result, err := resolvePlantComplete(state, action); err != nil || !result.Applied {
		t.Fatalf("prepare plant failed: %+v/%v", result, err)
	}
	for _, player := range state.Players {
		if player.Side == SideT {
			player.Alive, player.HP = false, 0
			player.Action = playerActionState{Status: actionIdle}
		}
	}
	state.Players["team_b_p1"].Location = playerLocation{NodeID: "A_SITE"}
	return state
}

func TestKitDefuseEqualExplosionCompletesFirst(t *testing.T) {
	state := preparePostPlantState(t, 1007)
	defuser := "team_b_p1"
	state.Timeline = state.BombDeadline - calculateDefuseTime(state, defuser)
	action, err := startDefuseAction(state, defuser, "defuse", 0)
	if err != nil {
		t.Fatal(err)
	}
	if action.ResolveAt != state.BombDeadline {
		t.Fatalf("defuse does not finish at explosion: %d/%d", action.ResolveAt, state.BombDeadline)
	}
	state.Timeline = action.ResolveAt
	result, err := resolveDefuseComplete(state, action)
	if err != nil || !result.Applied || state.Bomb.Status != bombDefused {
		t.Fatalf("equal-time defuse failed: %+v/%v bomb=%+v", result, err, state.Bomb)
	}
	explode := scheduledAction{ID: "explode", Type: actionBombExplode, ResolveAt: state.BombDeadline, Priority: priorityBombExplode}
	explosion, err := resolveBombExplode(state, explode)
	if err != nil || explosion.Applied {
		t.Fatalf("defused bomb exploded: %+v/%v", explosion, err)
	}
}

func TestDefuserSameSecondDeathFailsButOtherCTDeathDoesNot(t *testing.T) {
	for _, killDefuser := range []bool{true, false} {
		state := preparePostPlantState(t, 1008)
		defuser := "team_b_p1"
		state.Timeline = state.BombDeadline - calculateDefuseTime(state, defuser)
		action, err := startDefuseAction(state, defuser, "defuse", 0)
		if err != nil {
			t.Fatal(err)
		}
		attacker := state.Players["team_a_p2"]
		attacker.Alive, attacker.HP = true, 100
		victim := "team_b_p2"
		if killDefuser {
			victim = defuser
		}
		state.Timeline = action.ResolveAt
		combat := combatAction(state, "defuse-pulse", attacker.Profile.PlayerID)
		if _, err := applyCombatPulseCommit(state, combat, []effect{damageEffect(state, combat, attacker.Profile.PlayerID, victim, 100, 0)}); err != nil {
			t.Fatal(err)
		}
		result, err := resolveDefuseComplete(state, action)
		if err != nil {
			t.Fatal(err)
		}
		if killDefuser && (result.Applied || state.Bomb.Status != bombPlanted) {
			t.Fatalf("dead defuser completed: %+v bomb=%+v", result, state.Bomb)
		}
		if !killDefuser && (!result.Applied || state.Bomb.Status != bombDefused) {
			t.Fatalf("other CT death cancelled surviving defuser: %+v bomb=%+v", result, state.Bomb)
		}
	}
}
