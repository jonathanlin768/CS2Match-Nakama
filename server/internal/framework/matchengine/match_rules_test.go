package matchengine

import "testing"

func newTestScoreState(t *testing.T) *MatchScoreState {
	t.Helper()
	state, err := NewMatchScoreState("team_a", "team_b", map[string]string{
		"team_a": SideT,
		"team_b": SideCT,
	})
	if err != nil {
		t.Fatalf("new score state: %v", err)
	}
	return state
}

func TestMatchScoreStateRegulationEarlyWin(t *testing.T) {
	state := newTestScoreState(t)
	rule := DefaultMR12RuleSet(makeTestMapConfig().CombatConstants)
	for round := 1; round <= 13; round++ {
		if err := state.ApplyRoundWinner("team_a"); err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		if round < 13 && state.RegulationComplete(rule, round) {
			t.Fatalf("regulation ended early at round %d", round)
		}
	}
	if !state.RegulationComplete(rule, 13) {
		t.Fatal("expected regulation to end when team_a reaches 13")
	}
	if state.Score("team_a") != 13 || state.Score("team_b") != 0 {
		t.Fatalf("unexpected score %d:%d", state.Score("team_a"), state.Score("team_b"))
	}
}

func TestMatchScoreStateSideSwitchPreservesTeamScores(t *testing.T) {
	state := newTestScoreState(t)
	for i := 0; i < 8; i++ {
		_ = state.ApplyRoundWinner("team_a")
	}
	for i := 0; i < 4; i++ {
		_ = state.ApplyRoundWinner("team_b")
	}
	state.SwitchSides()
	if state.TeamForSide(SideT) != "team_b" || state.TeamForSide(SideCT) != "team_a" {
		t.Fatalf("wrong sides after switch: T=%s CT=%s", state.TeamForSide(SideT), state.TeamForSide(SideCT))
	}
	if err := state.ApplyRoundWinner("team_a"); err != nil {
		t.Fatal(err)
	}
	if state.Score("team_a") != 9 || state.Score("team_b") != 4 {
		t.Fatalf("side switch changed team scores: %d:%d", state.Score("team_a"), state.Score("team_b"))
	}
}

func TestMatchScoreStateOvertimeRequiresCompleteBlock(t *testing.T) {
	state := newTestScoreState(t)
	rule := DefaultMR12RuleSet(makeTestMapConfig().CombatConstants)
	for i := 0; i < 12; i++ {
		_ = state.ApplyRoundWinner("team_a")
		_ = state.ApplyRoundWinner("team_b")
	}
	if !state.ShouldEnterOvertime(rule, 24) {
		t.Fatal("12-12 should enter overtime")
	}

	winners := []string{"team_b", "team_b", "team_a", "team_b", "team_a", "team_b"}
	for i, winner := range winners {
		if i == rule.OvertimeHalfRounds {
			state.SwitchSides()
		}
		if err := state.ApplyRoundWinner(winner); err != nil {
			t.Fatal(err)
		}
		if i < len(winners)-1 && state.OvertimeDecided(rule, i+1) {
			t.Fatalf("overtime ended inside block after %d rounds", i+1)
		}
	}
	if !state.OvertimeDecided(rule, len(winners)) {
		t.Fatal("expected overtime to end after the complete decisive block")
	}
	if state.Score("team_b") <= state.Score("team_a") {
		t.Fatalf("expected team_b lead, got %d:%d", state.Score("team_a"), state.Score("team_b"))
	}
}
