package matchengine

import "context"

// RoundInput is the immutable per-round snapshot derived by the match layer.
// ScoreByTeam is copied from matchScoreState so round simulation cannot mutate
// the authoritative match score or side assignment.
type RoundInput struct {
	MatchID          string
	RoundNumber      int
	MapID            string
	MapVersion       string
	Seed             int64
	RuleSet          RuleSet
	MapConfig        *MapConfig
	WeaponSpecs      map[string]WeaponSpec
	SideLoadouts     map[string]WeaponLoadout
	TeamT            TeamInput
	TeamCT           TeamInput
	TeamAID          string
	TeamBID          string
	ScoreByTeam      map[string]int
	StrategyMemoryT  StrategyMemory
	StrategyMemoryCT StrategyMemory

	phase                string
	half                 int
	overtimeBlock        int
	overtimeRoundInBlock int
	isSideSwitch         bool
}

type roundTerminal struct {
	WinnerTeamID string
	WinnerSide   string
	WinReason    string
	Reason       reasonRecord
}

type roundSimulationResult struct {
	Round        *RoundResult
	Terminal     *roundTerminal
	PhaseHistory []roundPhase
}

type roundSimulator interface {
	SimulateRound(context.Context, *RoundInput) (*roundSimulationResult, error)
}

// causalRoundEngine is the production roundSimulator assembly point. The
// implementation is progressively replaced by the discrete causal pipeline;
// the match layer only depends on the terminal/result contract.
type causalRoundEngine struct{}

func (r *causalRoundEngine) SimulateRound(ctx context.Context, input *RoundInput) (*roundSimulationResult, error) {
	return runCausalRound(ctx, input)
}
