package matchengine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

type MatchEngine struct {
	input          *MatchInput
	state          *MatchScoreState
	roundSimulator roundSimulator
	stats          map[string]*playerStatAccumulator
	roster         map[string]PlayerProfile
	strategyMemory map[string]StrategyMemory
}

type playerStatAccumulator struct {
	PlayerID       string
	ConfigPlayerID string
	PlayerName     string
	TeamID         string
	Side           string
	Kills          int
	Deaths         int
	Assists        int
	Damage         int
	FK             int
	MK             int
	Plants         int
	Defuses        int
}

func newProductionMatchEngine(input *MatchInput) *MatchEngine {
	return newMatchEngine(input, nil)
}

func newMatchEngine(input *MatchInput, simulator roundSimulator) *MatchEngine {
	engine := &MatchEngine{input: input, roundSimulator: simulator}
	if engine.roundSimulator == nil {
		engine.roundSimulator = &causalRoundEngine{owner: engine}
	}
	return engine
}

func (e *MatchEngine) simulateMatch(ctx context.Context) (*MatchResult, error) {
	if err := e.validateInput(); err != nil {
		return nil, err
	}

	startTime := e.input.StartTime
	if startTime == 0 {
		startTime = time.Now().UnixMilli()
	}
	var err error
	e.state, err = NewMatchScoreState(e.input.TeamA.TeamID, e.input.TeamB.TeamID, e.input.InitialSideByTeam)
	if err != nil {
		return nil, err
	}
	e.roster = e.buildRoster()
	e.stats = e.buildStats()
	e.strategyMemory = map[string]StrategyMemory{
		e.input.TeamA.TeamID: newStrategyMemory(),
		e.input.TeamB.TeamID: newStrategyMemory(),
	}

	result := &MatchResult{
		MatchInfo: &MatchInfo{
			MatchID:    e.input.MatchID,
			MapID:      e.input.MapID,
			MapName:    e.input.MapName,
			MapVersion: e.input.MapVersion,
			RuleSetID:  e.input.RuleSet.RuleSetID,
			Seed:       e.input.Seed,
			TeamAID:    e.input.TeamA.TeamID,
			TeamBID:    e.input.TeamB.TeamID,
			TeamAName:  e.input.TeamA.Name,
			TeamBName:  e.input.TeamB.Name,
			StartTime:  startTime,
		},
	}

	if result.MatchInfo.MapName == "" {
		result.MatchInfo.MapName = e.input.MapConfig.MapName
	}

	for roundNumber := 1; roundNumber <= e.input.RuleSet.RegulationMaxRounds; roundNumber++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if roundNumber == e.input.RuleSet.RegulationHalfRounds+1 {
			e.swapSides()
		}
		simulation, err := e.roundSimulator.SimulateRound(ctx, e.buildRoundInput(roundContext{
			roundNumber:  roundNumber,
			phase:        "regulation",
			half:         regulationHalf(e.input.RuleSet, roundNumber),
			isSideSwitch: roundNumber == e.input.RuleSet.RegulationHalfRounds+1,
		}))
		if err != nil {
			return nil, err
		}
		round, err := e.consumeRoundSimulation(simulation)
		if err != nil {
			return nil, err
		}
		result.Rounds = append(result.Rounds, round)
		if e.state.RegulationComplete(e.input.RuleSet, len(result.Rounds)) {
			break
		}
	}

	if e.state.ShouldEnterOvertime(e.input.RuleSet, len(result.Rounds)) {
		if err := e.playOvertime(ctx, result); err != nil {
			return nil, err
		}
	}

	result.TotalRounds = len(result.Rounds)
	result.MatchInfo.TotalRounds = result.TotalRounds
	result.FinalScoreTeamA = e.state.Score(e.input.TeamA.TeamID)
	result.FinalScoreTeamB = e.state.Score(e.input.TeamB.TeamID)
	result.WinnerTeamID = e.winnerTeamID()
	result.Winner = e.winnerSide(result.WinnerTeamID)
	e.addMatchBoundaryEvents(result)
	result.MatchInfo.FinalScoreTeamA = result.FinalScoreTeamA
	result.MatchInfo.FinalScoreTeamB = result.FinalScoreTeamB
	result.MatchInfo.WinnerTeamID = result.WinnerTeamID
	result.FinalStats = e.buildFinalStats(result.TotalRounds, result.WinnerTeamID)
	if len(result.Rounds) > 0 {
		lastRound := result.Rounds[len(result.Rounds)-1]
		result.FinalStats.ScoreT = lastRound.ScoreT
		result.FinalStats.ScoreCT = lastRound.ScoreCT
	}
	result.FinalStats.ScoreTeamA = result.FinalScoreTeamA
	result.FinalStats.ScoreTeamB = result.FinalScoreTeamB
	result.Report = BuildMatchExplainableReport(result.Rounds)
	return result, nil
}

func (e *MatchEngine) addMatchBoundaryEvents(result *MatchResult) {
	if len(result.Rounds) == 0 {
		return
	}
	first := result.Rounds[0]
	first.Events = append(first.Events, &GameEvent{
		Timestamp: 0,
		EventType: EventMatchStart,
		Message:   fmt.Sprintf("%s vs %s on %s: match started", e.input.TeamA.Name, e.input.TeamB.Name, e.input.MapName),
		Extra: map[string]interface{}{
			"seed":        e.input.Seed,
			"rule_set":    e.input.RuleSet.RuleSetID,
			"map_version": e.input.MapVersion,
		},
	})
	sortEvents(first.Events)

	last := result.Rounds[len(result.Rounds)-1]
	last.Events = append(last.Events, &GameEvent{
		Timestamp:  int64(e.roundEndTime(last.Events)),
		EventType:  EventMatchEnd,
		Message:    fmt.Sprintf("match ended: %s wins %d:%d", e.teamName(result.WinnerTeamID), result.FinalScoreTeamA, result.FinalScoreTeamB),
		ScoreTeamA: result.FinalScoreTeamA,
		ScoreTeamB: result.FinalScoreTeamB,
	})
	sortEvents(last.Events)
}

func (e *MatchEngine) playOvertime(ctx context.Context, result *MatchResult) error {
	roundNumber := len(result.Rounds) + 1
	for block := 1; block <= 20; block++ {
		for inBlock := 1; inBlock <= e.input.RuleSet.OvertimeBlockRounds; inBlock++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			if inBlock == e.input.RuleSet.OvertimeHalfRounds+1 {
				e.swapSides()
			}
			simulation, err := e.roundSimulator.SimulateRound(ctx, e.buildRoundInput(roundContext{
				roundNumber:          roundNumber,
				phase:                "overtime",
				half:                 2 + block,
				overtimeBlock:        block,
				overtimeRoundInBlock: inBlock,
				isSideSwitch:         inBlock == e.input.RuleSet.OvertimeHalfRounds+1,
			}))
			if err != nil {
				return err
			}
			round, err := e.consumeRoundSimulation(simulation)
			if err != nil {
				return err
			}
			result.Rounds = append(result.Rounds, round)
			roundNumber++
		}
		if e.state.OvertimeDecided(e.input.RuleSet, e.input.RuleSet.OvertimeBlockRounds) {
			return nil
		}
	}
	return nil
}

type roundContext struct {
	roundNumber          int
	phase                string
	half                 int
	overtimeBlock        int
	overtimeRoundInBlock int
	isSideSwitch         bool
}

func (e *MatchEngine) buildRoundInput(ctx roundContext) *RoundInput {
	scoreSnapshot := map[string]int{
		e.input.TeamA.TeamID: e.state.Score(e.input.TeamA.TeamID),
		e.input.TeamB.TeamID: e.state.Score(e.input.TeamB.TeamID),
	}
	return &RoundInput{
		MatchID:              e.input.MatchID,
		RoundNumber:          ctx.roundNumber,
		MapID:                e.input.MapID,
		MapVersion:           e.input.MapVersion,
		Seed:                 deriveSeed(e.input.Seed, e.input.MapVersion, e.input.RuleSet.RuleSetID, ctx.roundNumber),
		RuleSet:              e.input.RuleSet,
		MapConfig:            e.input.MapConfig,
		WeaponSpecs:          e.input.WeaponSpecs,
		SideLoadouts:         e.input.SideLoadouts,
		TeamT:                e.team(e.state.TeamForSide(SideT)),
		TeamCT:               e.team(e.state.TeamForSide(SideCT)),
		TeamAID:              e.input.TeamA.TeamID,
		TeamBID:              e.input.TeamB.TeamID,
		ScoreByTeam:          scoreSnapshot,
		StrategyMemoryT:      cloneStrategyMemory(e.strategyMemory[e.state.TeamForSide(SideT)]),
		StrategyMemoryCT:     cloneStrategyMemory(e.strategyMemory[e.state.TeamForSide(SideCT)]),
		phase:                ctx.phase,
		half:                 ctx.half,
		overtimeBlock:        ctx.overtimeBlock,
		overtimeRoundInBlock: ctx.overtimeRoundInBlock,
		isSideSwitch:         ctx.isSideSwitch,
	}
}

func (e *MatchEngine) consumeRoundSimulation(simulation *RoundSimulationResult) (*RoundResult, error) {
	if simulation == nil || simulation.Round == nil || simulation.Terminal == nil {
		return nil, newError("SIMULATION_INVARIANT_ERROR", "round simulator returned no round or terminal")
	}
	terminal := simulation.Terminal
	expectedSide := e.state.SideByTeam[terminal.WinnerTeamID]
	if expectedSide == "" || (terminal.WinnerSide != "" && terminal.WinnerSide != expectedSide) || terminal.WinReason == "" {
		return nil, newError("SIMULATION_INVARIANT_ERROR", "round terminal is inconsistent with current teams/sides")
	}
	if err := e.state.ApplyRoundWinner(terminal.WinnerTeamID); err != nil {
		return nil, err
	}
	round := simulation.Round
	round.WinnerTeamID = terminal.WinnerTeamID
	round.Winner = expectedSide
	round.WinReason = terminal.WinReason
	round.ScoreTeamA = e.state.Score(e.input.TeamA.TeamID)
	round.ScoreTeamB = e.state.Score(e.input.TeamB.TeamID)
	round.ScoreT = e.state.Score(round.TeamTID)
	round.ScoreCT = e.state.Score(round.TeamCTID)
	for _, event := range round.Events {
		if event.State != nil {
			event.State.ScoreTeamA = round.ScoreTeamA
			event.State.ScoreTeamB = round.ScoreTeamB
			event.State.ScoreT = round.ScoreT
			event.State.ScoreCT = round.ScoreCT
		}
		if event.EventType == EventRoundEnd {
			event.ScoreTeamA = round.ScoreTeamA
			event.ScoreTeamB = round.ScoreTeamB
		}
	}
	e.aggregateRoundStats(round)
	e.updateStrategyMemory(round)
	return round, nil
}

func (e *MatchEngine) validateInput() error {
	if e.input == nil {
		return newError("INVALID_MATCH_INPUT", "input is nil")
	}
	if e.input.MatchID == "" || e.input.MapID == "" {
		return newError("INVALID_MATCH_INPUT", "match_id and map_id are required")
	}
	if e.input.Seed == 0 {
		return newError("INVALID_MATCH_INPUT", "seed must be non-zero before calling matchengine")
	}
	if err := ValidateRuleSet(e.input.RuleSet); err != nil {
		return err
	}
	if err := ValidateMapConfig(e.input.MapConfig); err != nil {
		return err
	}
	if e.input.MapID != e.input.MapConfig.MapID {
		return newError("INVALID_MATCH_INPUT", "input map %s does not match config %s", e.input.MapID, e.input.MapConfig.MapID)
	}
	if err := validateTeam(e.input.TeamA); err != nil {
		return err
	}
	if err := validateTeam(e.input.TeamB); err != nil {
		return err
	}
	if e.input.TeamA.TeamID == e.input.TeamB.TeamID {
		return newError("INVALID_LINEUP", "team ids must be unique within a match")
	}
	seen := map[string]struct{}{}
	for _, p := range append(e.input.TeamA.Players, e.input.TeamB.Players...) {
		if _, ok := seen[p.PlayerID]; ok {
			return newError("INVALID_LINEUP", "duplicate player id: %s", p.PlayerID)
		}
		seen[p.PlayerID] = struct{}{}
	}
	if e.input.InitialSideByTeam == nil {
		return newError("INVALID_MATCH_INPUT", "initial side assignment is required")
	}
	sideA := e.input.InitialSideByTeam[e.input.TeamA.TeamID]
	sideB := e.input.InitialSideByTeam[e.input.TeamB.TeamID]
	if !validSide(sideA) || !validSide(sideB) || sideA == sideB {
		return newError("INVALID_MATCH_INPUT", "initial sides must assign one T and one CT")
	}
	for _, side := range []string{SideT, SideCT} {
		loadout, ok := e.input.SideLoadouts[side]
		if !ok || loadout.Primary == "" {
			return newError("INVALID_MATCH_INPUT", "missing side loadout for %s", side)
		}
		if _, ok := e.input.WeaponSpecs[loadout.Primary]; !ok {
			return newError("INVALID_MATCH_INPUT", "weapon spec missing for %s", loadout.Primary)
		}
	}
	return nil
}

func validateTeam(team TeamInput) error {
	if team.TeamID == "" || team.Name == "" {
		return newError("INVALID_LINEUP", "team id and name are required")
	}
	if len(team.Players) != 5 {
		return newError("INVALID_LINEUP", "team %s must contain exactly 5 players", team.TeamID)
	}
	for _, p := range team.Players {
		if p.PlayerID == "" || p.ConfigPlayerID == "" || p.DisplayName == "" {
			return newError("INVALID_LINEUP", "player id, config player id, and display name are required")
		}
	}
	return nil
}

func (e *MatchEngine) buildFinalStats(totalRounds int, winnerTeamID string) *FinalStats {
	stats := make([]*PlayerMatchStats, 0, len(e.stats))
	ids := make([]string, 0, len(e.stats))
	for id := range e.stats {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		acc := e.stats[id]
		adr := 0.0
		if totalRounds > 0 {
			adr = float64(acc.Damage) / float64(totalRounds)
		}
		stats = append(stats, &PlayerMatchStats{
			PlayerID:       acc.PlayerID,
			ConfigPlayerID: acc.ConfigPlayerID,
			PlayerName:     acc.PlayerName,
			TeamID:         acc.TeamID,
			Side:           acc.Side,
			Kills:          acc.Kills,
			Deaths:         acc.Deaths,
			Assists:        acc.Assists,
			Damage:         acc.Damage,
			ADR:            math.Round(adr*10) / 10,
			FK:             acc.FK,
			MK:             acc.MK,
			Plants:         acc.Plants,
			Defuses:        acc.Defuses,
		})
	}
	return &FinalStats{WinnerTeamID: winnerTeamID, PlayerStats: stats}
}

func (e *MatchEngine) buildRoster() map[string]PlayerProfile {
	roster := map[string]PlayerProfile{}
	for _, p := range append(e.input.TeamA.Players, e.input.TeamB.Players...) {
		roster[p.PlayerID] = p
	}
	return roster
}

func (e *MatchEngine) buildStats() map[string]*playerStatAccumulator {
	stats := map[string]*playerStatAccumulator{}
	for _, team := range []TeamInput{e.input.TeamA, e.input.TeamB} {
		for _, p := range team.Players {
			stats[p.PlayerID] = &playerStatAccumulator{
				PlayerID:       p.PlayerID,
				ConfigPlayerID: p.ConfigPlayerID,
				PlayerName:     p.DisplayName,
				TeamID:         team.TeamID,
				Side:           e.input.InitialSideByTeam[team.TeamID],
			}
		}
	}
	return stats
}

func (e *MatchEngine) swapSides() {
	for teamID, memory := range e.strategyMemory {
		e.strategyMemory[teamID] = DecayStrategyMemoryForSideSwitch(memory)
	}
	e.state.SwitchSides()
}

func (e *MatchEngine) winnerTeamID() string {
	if e.state.Score(e.input.TeamA.TeamID) > e.state.Score(e.input.TeamB.TeamID) {
		return e.input.TeamA.TeamID
	}
	if e.state.Score(e.input.TeamB.TeamID) > e.state.Score(e.input.TeamA.TeamID) {
		return e.input.TeamB.TeamID
	}
	return ""
}

func (e *MatchEngine) winnerSide(teamID string) string {
	if teamID == "" {
		return ""
	}
	return e.state.SideByTeam[teamID]
}

func (e *MatchEngine) teamName(teamID string) string {
	if teamID == e.input.TeamA.TeamID {
		return e.input.TeamA.Name
	}
	if teamID == e.input.TeamB.TeamID {
		return e.input.TeamB.Name
	}
	return teamID
}

func (e *MatchEngine) team(teamID string) TeamInput {
	if teamID == e.input.TeamA.TeamID {
		return e.input.TeamA
	}
	return e.input.TeamB
}

func (e *MatchEngine) roundEndTime(events []*GameEvent) int {
	maxTimestamp := 1
	for _, ev := range events {
		if int(ev.Timestamp) > maxTimestamp {
			maxTimestamp = int(ev.Timestamp)
		}
	}
	return maxTimestamp + 2
}

func regulationHalf(rule RuleSet, roundNumber int) int {
	if roundNumber <= rule.RegulationHalfRounds {
		return 1
	}
	return 2
}

func validSide(side string) bool {
	return side == SideT || side == SideCT
}

func sortEvents(events []*GameEvent) {
	priority := map[string]int{
		EventMatchStart:  0,
		EventRoundStart:  0,
		EventHalfTime:    1,
		EventSideSwitch:  1,
		EventOvertime:    1,
		EventKill:        5,
		EventBombPlant:   6,
		EventBombDefuse:  7,
		EventBombExplode: 8,
		EventRoundEnd:    9,
		EventMatchEnd:    10,
	}
	sort.SliceStable(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.Timestamp != b.Timestamp {
			return a.Timestamp < b.Timestamp
		}
		aPriority, bPriority := a.sortPriority, b.sortPriority
		if aPriority == 0 {
			aPriority = -priority[a.EventType]
		}
		if bPriority == 0 {
			bPriority = -priority[b.EventType]
		}
		if aPriority != bPriority {
			return aPriority > bPriority
		}
		if a.sortActionType != b.sortActionType {
			return a.sortActionType < b.sortActionType
		}
		if a.sortMinActorID != b.sortMinActorID {
			return a.sortMinActorID < b.sortMinActorID
		}
		if a.SourceActionID != b.SourceActionID {
			return a.SourceActionID < b.SourceActionID
		}
		if a.SourceEffectID != b.SourceEffectID {
			return a.SourceEffectID < b.SourceEffectID
		}
		if a.EventID != b.EventID {
			return a.EventID < b.EventID
		}
		if a.EventType != b.EventType {
			return a.EventType < b.EventType
		}
		return a.Message < b.Message
	})
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
