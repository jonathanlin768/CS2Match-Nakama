package matchengine

import (
	"context"
	"reflect"
	"testing"
)

type scriptedTerminalSimulator struct {
	winners      []string
	inputs       []*RoundInput
	mutateScores bool
}

func (s *scriptedTerminalSimulator) SimulateRound(_ context.Context, input *RoundInput) (*RoundSimulationResult, error) {
	index := len(s.inputs)
	snapshot := *input
	snapshot.ScoreByTeam = map[string]int{}
	for teamID, score := range input.ScoreByTeam {
		snapshot.ScoreByTeam[teamID] = score
	}
	s.inputs = append(s.inputs, &snapshot)
	if s.mutateScores {
		for teamID := range input.ScoreByTeam {
			input.ScoreByTeam[teamID] = 999
		}
	}
	if index >= len(s.winners) {
		return nil, newError("TEST_SCRIPT_EXHAUSTED", "no terminal scripted for round %d", input.RoundNumber)
	}
	winnerTeamID := s.winners[index]
	winnerSide := SideCT
	if winnerTeamID == input.TeamT.TeamID {
		winnerSide = SideT
	}
	return &RoundSimulationResult{
		Round: &RoundResult{
			RoundNumber:          input.RoundNumber,
			Phase:                input.phase,
			Half:                 input.half,
			OvertimeBlock:        input.overtimeBlock,
			OvertimeRoundInBlock: input.overtimeRoundInBlock,
			IsSideSwitch:         input.isSideSwitch,
			Seed:                 input.Seed,
			SideAttacking:        SideT,
			TeamTID:              input.TeamT.TeamID,
			TeamCTID:             input.TeamCT.TeamID,
		},
		Terminal: &RoundTerminal{WinnerTeamID: winnerTeamID, WinnerSide: winnerSide, WinReason: "test_terminal"},
	}, nil
}

func TestMatchOrchestrationEarlyStopKeepsTeamIdentityScoreAcrossSideSwitch(t *testing.T) {
	winners := make([]string, 13)
	for index := range winners {
		winners[index] = "team_a"
	}
	fake := &scriptedTerminalSimulator{winners: winners, mutateScores: true}
	result, err := newMatchEngine(makeTestInput(501), fake).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("match orchestration failed: %v", err)
	}
	if result.TotalRounds != 13 || result.FinalScoreTeamA != 13 || result.FinalScoreTeamB != 0 {
		t.Fatalf("unexpected early-stop result: rounds=%d score=%d:%d", result.TotalRounds, result.FinalScoreTeamA, result.FinalScoreTeamB)
	}
	if fake.inputs[12].TeamCT.TeamID != "team_a" || fake.inputs[12].ScoreByTeam["team_a"] != 12 {
		t.Fatalf("side switch changed/reset team score snapshot: teamCT=%s score=%v", fake.inputs[12].TeamCT.TeamID, fake.inputs[12].ScoreByTeam)
	}
	round13 := result.Rounds[12]
	if round13.ScoreTeamA != 13 || round13.ScoreTeamB != 0 || round13.ScoreT != 0 || round13.ScoreCT != 13 {
		t.Fatalf("round score projection confused side and team identity: %+v", round13)
	}
	for _, round := range result.Rounds {
		for _, event := range round.Events {
			if event.EventType != EventMatchStart && event.EventType != EventMatchEnd {
				t.Fatalf("MatchRules fake must not claim to generate combat events: %+v", event)
			}
		}
	}
}

func TestMatchOrchestrationCompletesFullBlocksAndMultipleOvertimes(t *testing.T) {
	winners := make([]string, 0, 36)
	for index := 0; index < 12; index++ {
		winners = append(winners, "team_a")
	}
	for index := 0; index < 12; index++ {
		winners = append(winners, "team_b")
	}
	winners = append(winners, "team_a", "team_b", "team_a", "team_b", "team_a", "team_b")
	winners = append(winners, "team_a", "team_a", "team_a", "team_b", "team_a", "team_b")

	fake := &scriptedTerminalSimulator{winners: winners}
	result, err := newMatchEngine(makeTestInput(777), fake).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("multi-OT orchestration failed: %v", err)
	}
	if result.TotalRounds != 36 || result.FinalScoreTeamA != 19 || result.FinalScoreTeamB != 17 {
		t.Fatalf("unexpected multi-OT result: rounds=%d score=%d:%d", result.TotalRounds, result.FinalScoreTeamA, result.FinalScoreTeamB)
	}
	if result.Rounds[29].OvertimeBlock != 1 || result.Rounds[35].OvertimeBlock != 2 || result.Rounds[35].OvertimeRoundInBlock != 6 {
		t.Fatalf("overtime ended inside a block: r30=%+v r36=%+v", result.Rounds[29], result.Rounds[35])
	}
}

func TestFormalMatchContractsCannotScriptWinnersOrOverrideRoundConstants(t *testing.T) {
	matchInputType := reflect.TypeOf(MatchInput{})
	for _, forbidden := range []string{"ForcedRoundWinners", "RoundCount", "MaxRounds"} {
		if _, ok := matchInputType.FieldByName(forbidden); ok {
			t.Fatalf("MatchInput exposes forbidden formal control %s", forbidden)
		}
	}
	ruleType := reflect.TypeOf(RuleSet{})
	for _, forbidden := range []string{
		"RoundTimeLimit", "BombExplodeTime", "BasePlantTime", "BaseDefuseTime", "BasePickupTime",
		"ForceExecuteThreshold", "MaxDecisionCount", "MaxEncounterPulses",
	} {
		if _, ok := ruleType.FieldByName(forbidden); ok {
			t.Fatalf("RuleSet overrides CombatConstants through %s", forbidden)
		}
	}
	serviceType := reflect.TypeOf(&Service{})
	if serviceType.NumMethod() != 1 || serviceType.Method(0).Name != "Simulate" {
		t.Fatalf("Simulate must be the only exported formal engine entry, got %v", serviceType.NumMethod())
	}
}
