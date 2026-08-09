package matchengine

import "testing"

func TestStateFingerprintIsMapOrderStableAndTracksCausalDimensions(t *testing.T) {
	left := makeTestRoundState(t, 1301)
	right := makeTestRoundState(t, 1301)
	reordered := make(map[string]*RoundPlayerState, len(right.Players))
	ids := sortedPlayerIDs(right)
	for index := len(ids) - 1; index >= 0; index-- {
		reordered[ids[index]] = right.Players[ids[index]]
	}
	right.Players = reordered
	if StateFingerprint(left) != StateFingerprint(right) {
		t.Fatal("fingerprint depends on map insertion order")
	}
	right.Players["team_a_p1"].Focus--
	if StateFingerprint(left) == StateFingerprint(right) {
		t.Fatal("fingerprint ignored player resources")
	}
}

func TestRecoveryCycleOrdinalDoesNotRollBackAndOwnProgressIsRetained(t *testing.T) {
	state := makeTestRoundState(t, 1302)
	setConstInt(&state.constants, "MaxNoOpTransitions", 1)
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	firstCycle := state.RecoveryAttempt.CycleID
	action, err := ScheduleNoProgressRecovery(state)
	if err != nil || action == nil || state.RecoveryAttempt.Status != RecoveryRunning {
		t.Fatalf("recovery action was not scheduled: %+v/%+v/%v", action, state.RecoveryAttempt, err)
	}
	before := StateFingerprint(state)
	state.Timeline++
	progressed, err := ObserveStateProgress(state, before, action.ID)
	if err != nil || !progressed || state.RecoveryAttempt.CycleID != firstCycle || state.NoOpCount != 1 {
		t.Fatalf("recovery-owned progress cleared proof: progressed=%t recovery=%+v err=%v", progressed, state.RecoveryAttempt, err)
	}
	if err := CompleteNoProgressRecovery(state, action.ID, true, "REACHABILITY_RESTORED"); err != nil {
		t.Fatal(err)
	}
	if state.RecoveryAttempt.Status != RecoverySucceeded || state.NoOpCount != 0 || state.NoProgressEligible {
		t.Fatalf("successful recovery lifecycle mismatch: %+v", state.RecoveryAttempt)
	}
	resetRecoveryCycle(state)
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	if state.RecoveryAttempt.CycleID == firstCycle || state.RecoveryOrdinal != 2 {
		t.Fatalf("RecoveryOrdinal rolled back: first=%s current=%+v ordinal=%d", firstCycle, state.RecoveryAttempt, state.RecoveryOrdinal)
	}
}

func TestExternalProgressClearsRecoveryCycleAndEligibility(t *testing.T) {
	state := makeTestRoundState(t, 1303)
	setConstInt(&state.constants, "MaxNoOpTransitions", 1)
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	action, err := ScheduleNoProgressRecovery(state)
	if err != nil || action == nil {
		t.Fatalf("recovery fixture failed: %+v/%v", action, err)
	}
	before := StateFingerprint(state)
	state.Players["team_b_p1"].Focus--
	progressed, err := ObserveStateProgress(state, before, "external-combat")
	if err != nil || !progressed || state.RecoveryAttempt.CycleID != "" || state.NoOpCount != 0 || state.NoProgressEligible {
		t.Fatalf("external progress retained stale recovery: progressed=%t recovery=%+v err=%v", progressed, state.RecoveryAttempt, err)
	}
}

func TestPostPlantNoOpNeverBecomesNoProgressEligible(t *testing.T) {
	state := preparePostPlantState(t, 1304)
	setConstInt(&state.constants, "MaxNoOpTransitions", 1)
	if err := state.RecordNoOp(); err != nil {
		t.Fatal(err)
	}
	action, err := ScheduleNoProgressRecovery(state)
	if err != nil || action != nil || state.NoProgressEligible || state.RecoveryAttempt.ResultCode != "POST_PLANT_USES_BOMB_DEADLINE" {
		t.Fatalf("postplant entered preplant no-progress: action=%+v recovery=%+v err=%v", action, state.RecoveryAttempt, err)
	}
}
