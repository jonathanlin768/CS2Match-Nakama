package matchengine

import "testing"

func eliminateSideForTerminal(state *RoundState, side string) {
	for _, player := range state.Players {
		if player.Side == side {
			player.Alive, player.HP = false, 0
			player.Action = PlayerActionState{Status: ActionIdle}
		}
	}
}

func TestEvaluateRoundTerminalPriorityAndInternalCodes(t *testing.T) {
	t.Run("dropped bomb T elimination beats timeout", func(t *testing.T) {
		state := makeTestRoundState(t, 1201)
		carrier := state.Players[state.Bomb.CarrierID]
		carrier.HasBomb = false
		if err := state.Bomb.Drop(carrier.Location, 10); err != nil {
			t.Fatal(err)
		}
		eliminateSideForTerminal(state, SideT)
		state.Timeline = state.RoundDeadline
		before := StateFingerprint(state)
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{Timestamp: state.Timeline})
		if err != nil || terminal == nil || terminal.WinnerSide != SideCT || terminal.WinReason != WinReasonElimination || terminal.Reason.Code != "T_ELIMINATED_BOMB_DROPPED" {
			t.Fatalf("terminal priority/code mismatch: %+v/%v", terminal, err)
		}
		if after := StateFingerprint(state); after != before {
			t.Fatal("pure terminal evaluation mutated state")
		}
	})

	t.Run("preplant CT elimination", func(t *testing.T) {
		state := makeTestRoundState(t, 1202)
		eliminateSideForTerminal(state, SideCT)
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{})
		if err != nil || terminal == nil || terminal.WinnerSide != SideT || terminal.WinReason != WinReasonElimination || terminal.Reason.Code != "CT_ELIMINATED" {
			t.Fatalf("CT elimination mismatch: %+v/%v", terminal, err)
		}
	})

	t.Run("preplant timeout", func(t *testing.T) {
		state := makeTestRoundState(t, 1203)
		state.Timeline = state.RoundDeadline
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{})
		if err != nil || terminal == nil || terminal.WinnerSide != SideCT || terminal.WinReason != WinReasonTimeout {
			t.Fatalf("timeout mismatch: %+v/%v", terminal, err)
		}
	})

	t.Run("postplant T elimination does not end round", func(t *testing.T) {
		state := preparePostPlantState(t, 1204)
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{})
		if err != nil || terminal != nil {
			t.Fatalf("postplant T elimination ended round: %+v/%v", terminal, err)
		}
	})

	t.Run("postplant CT elimination secures bomb", func(t *testing.T) {
		state := preparePostPlantState(t, 1205)
		eliminateSideForTerminal(state, SideCT)
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{})
		if err != nil || terminal == nil || terminal.WinnerSide != SideT || terminal.WinReason != WinReasonBombSecured {
			t.Fatalf("bomb secured mismatch: %+v/%v", terminal, err)
		}
	})
}

func TestEvaluateRoundTerminalBombOutcomesComeFromAppliedLifecycle(t *testing.T) {
	t.Run("defused", func(t *testing.T) {
		state := preparePostPlantState(t, 1210)
		state.Timeline = state.BombDeadline - CalculateDefuseTime(state, "team_b_p1")
		action, err := StartDefuseAction(state, "team_b_p1", "terminal-defuse", 0)
		if err != nil {
			t.Fatal(err)
		}
		state.Timeline = action.ResolveAt
		result, err := ResolveDefuseComplete(state, action)
		if err != nil || !result.Applied {
			t.Fatalf("defuse fixture failed: %+v/%v", result, err)
		}
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{Timestamp: state.Timeline, Events: []*GameEvent{result.Event}})
		if err != nil || terminal == nil || terminal.WinnerSide != SideCT || terminal.WinReason != WinReasonBombDefused {
			t.Fatalf("defuse terminal mismatch: %+v/%v", terminal, err)
		}
	})

	t.Run("exploded", func(t *testing.T) {
		state := preparePostPlantState(t, 1211)
		state.Timeline = state.BombDeadline
		action := ScheduledAction{ID: "terminal-explode", Type: ActionBombExplode, ResolveAt: state.BombDeadline, Priority: PriorityBombExplode}
		result, err := ResolveBombExplode(state, action)
		if err != nil || !result.Applied {
			t.Fatalf("explode fixture failed: %+v/%v", result, err)
		}
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{Timestamp: state.Timeline, Events: []*GameEvent{result.Event}})
		if err != nil || terminal == nil || terminal.WinnerSide != SideT || terminal.WinReason != WinReasonBombExploded {
			t.Fatalf("explosion terminal mismatch: %+v/%v", terminal, err)
		}
	})
}

func TestEvaluateRoundTerminalRejectsInconsistentStateWithoutRepair(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*RoundState)
	}{
		{name: "AliveHP", mutate: func(state *RoundState) { state.Players["team_a_p1"].Alive = false }},
		{name: "BombTimer", mutate: func(state *RoundState) { state.BombDeadline = 9 }},
		{name: "FutureEvent", mutate: func(state *RoundState) {
			state.Events = append(state.Events, &GameEvent{Timestamp: 1, EventType: EventDamage})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := makeTestRoundState(t, 1220)
			test.mutate(state)
			beforeHP := state.Players["team_a_p1"].HP
			_, err := EvaluateRoundTerminal(state, AppliedBatch{})
			assertEngineErrorCode(t, err, "SIMULATION_INVARIANT_ERROR")
			if state.Players["team_a_p1"].HP != beforeHP {
				t.Fatal("terminal validation repaired invalid state")
			}
		})
	}
}

func TestNoProgressRequiresFailedRecoveryEligibility(t *testing.T) {
	state := makeTestRoundState(t, 1230)
	state.mapEdges = map[string]MapEdge{}
	valid, err := ValidNoProgress(state)
	if err != nil || !valid {
		t.Fatalf("unreachable carrier was not a valid business condition: %t/%v", valid, err)
	}
	terminal, err := EvaluateRoundTerminal(state, AppliedBatch{})
	if err != nil || terminal != nil {
		t.Fatalf("business condition bypassed eligibility: %+v/%v", terminal, err)
	}
	setConstInt(&state.constants, "MaxNoOpTransitions", 1)
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	action, err := ScheduleNoProgressRecovery(state)
	if err != nil || action != nil || state.RecoveryAttempt.Status != RecoveryFailed || !state.NoProgressEligible {
		t.Fatalf("failed recovery proof mismatch: action=%+v recovery=%+v err=%v", action, state.RecoveryAttempt, err)
	}
	terminal, err = EvaluateRoundTerminal(state, AppliedBatch{})
	if err != nil || terminal == nil || terminal.WinReason != WinReasonNoProgressTimeout || terminal.Reason.Code != "NO_PROGRESS_CONFIRMED" {
		t.Fatalf("eligible no-progress mismatch: %+v/%v", terminal, err)
	}
}
