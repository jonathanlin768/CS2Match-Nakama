package matchengine

import (
	"context"
	"math"
	"runtime"
	"sync"
)

// CalibrationSummary contains distribution metrics produced only from completed
// causal rounds. Opportunity counts make zero-valued conditional rates explicit.
type CalibrationSummary struct {
	Samples                     int            `json:"samples"`
	TerminalCounts              map[string]int `json:"terminal_counts"`
	TimeoutBombNodes            map[string]int `json:"timeout_bomb_nodes,omitempty"`
	TimeoutAverageAliveT        float64        `json:"timeout_average_alive_t"`
	TimeoutAverageAliveCT       float64        `json:"timeout_average_alive_ct"`
	TWinRate                    float64        `json:"t_win_rate"`
	AverageKills                float64        `json:"average_kills"`
	PlantRate                   float64        `json:"plant_rate"`
	DefuseRate                  float64        `json:"defuse_rate"` // Defuses divided by planted rounds.
	ExplosionRate               float64        `json:"explosion_rate"`
	FirstKillOpportunities      int            `json:"first_kill_opportunities"`
	FirstKillWinnerRate         float64        `json:"first_kill_winner_rate"`
	FiveVThreeOpportunities     int            `json:"five_v_three_opportunities"`
	FiveVThreeConversionRate    float64        `json:"five_v_three_conversion_rate"`
	ThreeVFiveOpportunities     int            `json:"three_v_five_opportunities"`
	ThreeVFiveComebackRate      float64        `json:"three_v_five_comeback_rate"`
	StrongTeamOpportunities     int            `json:"strong_team_opportunities"`
	StrongTeamWinRate           float64        `json:"strong_team_win_rate"`
	AverageRoundDurationSeconds float64        `json:"average_round_duration_seconds"`
}

// CalibrateRounds runs the production causal RoundEngine across consecutive
// seeds. Sides alternate per sample so team strength and side advantage are not
// conflated. The supplied input and all of its configuration maps are read-only.
func CalibrateRounds(ctx context.Context, template *RoundInput, samples int) (*CalibrationSummary, error) {
	if template == nil || template.MapConfig == nil {
		return nil, newError("INVALID_CALIBRATION_INPUT", "round input/map config is nil")
	}
	if samples <= 0 {
		return nil, newError("INVALID_CALIBRATION_INPUT", "sample count must be positive")
	}

	results := make([]*calibrationRound, samples)
	workerCount := runtime.GOMAXPROCS(0)
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > samples {
		workerCount = samples
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	var workers sync.WaitGroup
	var errorMu sync.Mutex
	var firstErr error

	worker := func() {
		defer workers.Done()
		for {
			select {
			case <-workCtx.Done():
				return
			case index, ok := <-jobs:
				if !ok {
					return
				}
				input := calibrationInput(template, index)
				result, err := runCausalRound(workCtx, input)
				if err != nil {
					errorMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errorMu.Unlock()
					cancel()
					return
				}
				results[index] = &calibrationRound{input: input, result: result}
			}
		}
	}

	workers.Add(workerCount)
	for range workerCount {
		go worker()
	}
sendJobs:
	for index := range samples {
		select {
		case jobs <- index:
		case <-workCtx.Done():
			break sendJobs
		}
	}
	close(jobs)
	workers.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return aggregateCalibration(results)
}

type calibrationRound struct {
	input  *RoundInput
	result *RoundSimulationResult
}

func calibrationInput(template *RoundInput, index int) *RoundInput {
	input := *template
	input.ScoreByTeam = copyIntMap(template.ScoreByTeam)
	input.Seed = template.Seed + int64(index)
	if input.Seed == 0 {
		input.Seed = int64(index + 1)
	}
	if input.TeamAID == "" {
		input.TeamAID = template.TeamT.TeamID
	}
	if input.TeamBID == "" {
		input.TeamBID = template.TeamCT.TeamID
	}
	if index%2 == 1 {
		input.TeamT, input.TeamCT = template.TeamCT, template.TeamT
		input.StrategyMemoryT, input.StrategyMemoryCT = template.StrategyMemoryCT, template.StrategyMemoryT
	}
	return &input
}

func aggregateCalibration(rounds []*calibrationRound) (*CalibrationSummary, error) {
	summary := &CalibrationSummary{Samples: len(rounds), TerminalCounts: map[string]int{}, TimeoutBombNodes: map[string]int{}}
	var tWins, kills, plants, defuses, explosions, firstKillWins int
	var fiveVThreeConversions, threeVFiveComebacks, strongTeamWins, duration int
	timeoutAliveT, timeoutAliveCT := 0, 0
	for _, sample := range rounds {
		if sample == nil || sample.result == nil || sample.result.Round == nil || sample.result.Terminal == nil {
			return nil, newError("SIMULATION_INVARIANT_ERROR", "calibration received an incomplete round result")
		}
		round := sample.result.Round
		terminal := sample.result.Terminal
		summary.TerminalCounts[terminal.WinReason]++
		if terminal.WinReason == WinReasonTimeout {
			nodeID := "unknown"
			if round.Bomb != nil && round.Bomb.NodeID != "" {
				nodeID = round.Bomb.NodeID
			}
			summary.TimeoutBombNodes[nodeID]++
			aliveT, aliveCT := aliveBySide(round.PlayerStates)
			timeoutAliveT += aliveT
			timeoutAliveCT += aliveCT
		}
		if terminal.WinnerSide == SideT {
			tWins++
		}
		firstKillTeamID, advantagedTeamID, disadvantagedTeamID := "", "", ""
		tAlive, ctAlive := len(sample.input.TeamT.Players), len(sample.input.TeamCT.Players)
		for _, event := range round.Events {
			switch event.EventType {
			case EventKill:
				kills++
				if firstKillTeamID == "" {
					firstKillTeamID = event.AttackerTeamID
				}
				if event.State != nil && len(event.State.Players) > 0 {
					tAlive, ctAlive = aliveBySide(event.State.Players)
				} else {
					if event.VictimTeamID == round.TeamTID {
						tAlive--
					} else if event.VictimTeamID == round.TeamCTID {
						ctAlive--
					}
				}
				if advantagedTeamID == "" {
					switch {
					case tAlive == 5 && ctAlive == 3:
						advantagedTeamID, disadvantagedTeamID = round.TeamTID, round.TeamCTID
					case ctAlive == 5 && tAlive == 3:
						advantagedTeamID, disadvantagedTeamID = round.TeamCTID, round.TeamTID
					}
				}
			case EventBombPlant:
				plants++
			case EventBombDefuse:
				defuses++
			case EventBombExplode:
				explosions++
			}
		}
		if firstKillTeamID != "" {
			summary.FirstKillOpportunities++
			if firstKillTeamID == terminal.WinnerTeamID {
				firstKillWins++
			}
		}
		if advantagedTeamID != "" {
			summary.FiveVThreeOpportunities++
			summary.ThreeVFiveOpportunities++
			if advantagedTeamID == terminal.WinnerTeamID {
				fiveVThreeConversions++
			}
			if disadvantagedTeamID == terminal.WinnerTeamID {
				threeVFiveComebacks++
			}
		}
		if strongTeamID, ok := strongerTeam(sample.input.TeamT, sample.input.TeamCT); ok {
			summary.StrongTeamOpportunities++
			if strongTeamID == terminal.WinnerTeamID {
				strongTeamWins++
			}
		}
		duration += roundDuration(round.Events)
	}

	denominator := float64(summary.Samples)
	summary.TWinRate = float64(tWins) / denominator
	summary.AverageKills = float64(kills) / denominator
	summary.PlantRate = float64(plants) / denominator
	summary.DefuseRate = ratio(defuses, plants)
	summary.ExplosionRate = float64(explosions) / denominator
	summary.FirstKillWinnerRate = ratio(firstKillWins, summary.FirstKillOpportunities)
	summary.FiveVThreeConversionRate = ratio(fiveVThreeConversions, summary.FiveVThreeOpportunities)
	summary.ThreeVFiveComebackRate = ratio(threeVFiveComebacks, summary.ThreeVFiveOpportunities)
	summary.StrongTeamWinRate = ratio(strongTeamWins, summary.StrongTeamOpportunities)
	summary.AverageRoundDurationSeconds = float64(duration) / denominator
	timeoutCount := summary.TerminalCounts[WinReasonTimeout]
	summary.TimeoutAverageAliveT = ratio(timeoutAliveT, timeoutCount)
	summary.TimeoutAverageAliveCT = ratio(timeoutAliveCT, timeoutCount)
	return summary, nil
}

func roundDuration(events []*GameEvent) int {
	duration := int64(0)
	for _, event := range events {
		if event != nil && event.EventType == EventRoundEnd && event.Timestamp > duration {
			duration = event.Timestamp
		}
	}
	return int(duration)
}

func aliveBySide(players []*PlayerState) (int, int) {
	tAlive, ctAlive := 0, 0
	for _, player := range players {
		if player == nil || !player.Alive {
			continue
		}
		if player.Side == SideT {
			tAlive++
		} else if player.Side == SideCT {
			ctAlive++
		}
	}
	return tAlive, ctAlive
}

func strongerTeam(left, right TeamInput) (string, bool) {
	leftScore, rightScore := teamAttributeTotal(left), teamAttributeTotal(right)
	if leftScore == rightScore {
		return "", false
	}
	average := float64(leftScore+rightScore) / 2
	if average <= 0 || math.Abs(float64(leftScore-rightScore))/average < 0.10 {
		return "", false
	}
	if leftScore > rightScore {
		return left.TeamID, true
	}
	return right.TeamID, true
}

func teamAttributeTotal(team TeamInput) int {
	total := 0
	for _, player := range team.Players {
		a := player.Attributes
		total += a.Entry + a.Aim + a.Trade + a.Clutch + a.Firepower + a.Gamesense + a.Reaction + a.Positioning + a.Awareness + a.Teamplay + a.Utility + a.Composure + a.Mobility + a.Endurance + a.Discipline
	}
	return total
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
