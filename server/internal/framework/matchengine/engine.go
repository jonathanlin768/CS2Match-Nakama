package matchengine

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

type MatchEngine struct {
	input  *MatchInput
	state  *matchState
	stats  map[string]*playerStatAccumulator
	roster map[string]PlayerProfile
}

type matchState struct {
	scoreTeamA int
	scoreTeamB int
	sides      map[string]string
}

type roundPlayer struct {
	profile PlayerProfile
	teamID  string
	side    string
	weapon  WeaponLoadout
	alive   bool
	kills   int
	deaths  int
	damage  int
	hasBomb bool
	nodeID  string
}

type playerStatAccumulator struct {
	PlayerID   string
	PlayerName string
	TeamID     string
	Side       string
	Kills      int
	Deaths     int
	Damage     int
	FK         int
	MK         int
	Plants     int
	Defuses    int
}

func NewMatchEngine(input *MatchInput) *MatchEngine {
	return &MatchEngine{input: input}
}

func (e *MatchEngine) StartMatch(ctx context.Context) (*MatchResult, error) {
	if err := e.validateInput(); err != nil {
		return nil, err
	}

	startTime := e.input.StartTime
	if startTime == 0 {
		startTime = time.Now().UnixMilli()
	}
	e.state = &matchState{
		sides: map[string]string{
			e.input.TeamA.TeamID: e.input.InitialSideByTeam[e.input.TeamA.TeamID],
			e.input.TeamB.TeamID: e.input.InitialSideByTeam[e.input.TeamB.TeamID],
		},
	}
	e.roster = e.buildRoster()
	e.stats = e.buildStats()

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
		round := e.simulateRound(roundContext{
			roundNumber:  roundNumber,
			phase:        "regulation",
			half:         regulationHalf(e.input.RuleSet, roundNumber),
			isSideSwitch: roundNumber == e.input.RuleSet.RegulationHalfRounds+1,
		})
		result.Rounds = append(result.Rounds, round)
		if e.state.scoreTeamA >= e.input.RuleSet.RegulationWinRounds || e.state.scoreTeamB >= e.input.RuleSet.RegulationWinRounds {
			break
		}
	}

	if e.state.scoreTeamA == e.state.scoreTeamB && e.input.RuleSet.OvertimeEnabled {
		e.playOvertime(ctx, result)
	}

	result.TotalRounds = len(result.Rounds)
	result.MatchInfo.TotalRounds = result.TotalRounds
	result.FinalScoreTeamA = e.state.scoreTeamA
	result.FinalScoreTeamB = e.state.scoreTeamB
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
		Message:   fmt.Sprintf("%s 对阵 %s，地图 %s，比赛开始。", e.input.TeamA.Name, e.input.TeamB.Name, e.input.MapName),
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
		Message:    fmt.Sprintf("比赛结束，%s 以 %d:%d 获胜。", e.teamName(result.WinnerTeamID), result.FinalScoreTeamA, result.FinalScoreTeamB),
		ScoreTeamA: result.FinalScoreTeamA,
		ScoreTeamB: result.FinalScoreTeamB,
	})
	sortEvents(last.Events)
}

func (e *MatchEngine) playOvertime(ctx context.Context, result *MatchResult) {
	roundNumber := len(result.Rounds) + 1
	for block := 1; block <= 20; block++ {
		for inBlock := 1; inBlock <= e.input.RuleSet.OvertimeBlockRounds; inBlock++ {
			if err := ctx.Err(); err != nil {
				return
			}
			if inBlock == e.input.RuleSet.OvertimeHalfRounds+1 {
				e.swapSides()
			}
			round := e.simulateRound(roundContext{
				roundNumber:          roundNumber,
				phase:                "overtime",
				half:                 2 + block,
				overtimeBlock:        block,
				overtimeRoundInBlock: inBlock,
				isSideSwitch:         inBlock == e.input.RuleSet.OvertimeHalfRounds+1,
			})
			result.Rounds = append(result.Rounds, round)
			roundNumber++
		}
		if e.state.scoreTeamA != e.state.scoreTeamB {
			return
		}
	}
}

type roundContext struct {
	roundNumber          int
	phase                string
	half                 int
	overtimeBlock        int
	overtimeRoundInBlock int
	isSideSwitch         bool
}

func (e *MatchEngine) simulateRound(ctx roundContext) *RoundResult {
	roundSeed := deriveSeed(e.input.Seed, e.input.MapVersion, e.input.RuleSet.RuleSetID, ctx.roundNumber)
	decisionRNG := newDerivedRand(roundSeed, "decision")
	encounterRNG := newDerivedRand(roundSeed, "encounter")
	bombRNG := newDerivedRand(roundSeed, "bomb")

	teamTID := e.teamIDBySide(SideT)
	teamCTID := e.teamIDBySide(SideCT)
	route := e.selectRoute(decisionRNG)
	template := e.selectTemplate(route, decisionRNG)
	scenario := e.selectScenario(template, decisionRNG)
	players := e.buildRoundPlayers(teamTID, teamCTID, route)

	events := []*GameEvent{{
		Timestamp: 0,
		EventType: EventRoundStart,
		Message:   fmt.Sprintf("第 %d 回合开始：%s 作为 T 方执行 %s。", ctx.roundNumber, e.teamName(teamTID), route.Name),
		Extra: map[string]interface{}{
			"phase":                ctx.phase,
			"overtime_block":       ctx.overtimeBlock,
			"strategy_template_id": template.ID,
			"scenario_id":          scenario.ID,
		},
	}}
	if ctx.isSideSwitch {
		eventType := EventSideSwitch
		message := fmt.Sprintf("第 %d 回合换边：%s 为 T，%s 为 CT。", ctx.roundNumber, e.teamName(teamTID), e.teamName(teamCTID))
		if ctx.roundNumber == e.input.RuleSet.RegulationHalfRounds+1 {
			eventType = EventHalfTime
			message = fmt.Sprintf("半场换边：%s 转为 T，%s 转为 CT。", e.teamName(teamTID), e.teamName(teamCTID))
		}
		events = append(events, &GameEvent{Timestamp: 0, EventType: eventType, Message: message})
	}
	if ctx.overtimeRoundInBlock == 1 {
		events = append(events, &GameEvent{
			Timestamp: 0,
			EventType: EventOvertime,
			Message:   fmt.Sprintf("加时 OT%d 开始，先完整进行 6 个小局。", ctx.overtimeBlock),
		})
	}

	winnerTeamID, scoreDelta := e.resolveRoundWinner(ctx.roundNumber, teamTID, teamCTID, route, scenario, encounterRNG)
	winnerSide := e.state.sides[winnerTeamID]
	events, bomb := e.resolveRoundEvents(events, players, route, scenario, winnerTeamID, winnerSide, scoreDelta, roundSeed, encounterRNG, bombRNG)
	tAlive := len(filterRoundPlayers(players, SideT, true))
	ctAlive := len(filterRoundPlayers(players, SideCT, true))

	if winnerTeamID == e.input.TeamA.TeamID {
		e.state.scoreTeamA++
	} else {
		e.state.scoreTeamB++
	}

	winReason := inferWinReason(winnerSide, bomb, tAlive, ctAlive)
	roundEndAt := e.roundEndTime(events)
	if winReason == "timeout" {
		roundEndAt = e.input.RuleSet.RoundTimeLimit
	}
	events = append(events, &GameEvent{
		Timestamp:  int64(clampInt(roundEndAt, 1, e.input.RuleSet.RoundTimeLimit+e.input.RuleSet.BombExplodeTime)),
		EventType:  EventRoundEnd,
		Message:    fmt.Sprintf("%s 赢下第 %d 回合（%s），比分 %d:%d。", e.teamName(winnerTeamID), ctx.roundNumber, winReason, e.state.scoreTeamA, e.state.scoreTeamB),
		ScoreTeamA: e.state.scoreTeamA,
		ScoreTeamB: e.state.scoreTeamB,
		Reason: &EventReason{
			Code:       winReason,
			MainFactor: "team_round_score",
			ScoreDelta: scoreDelta,
		},
		Bomb: cloneBomb(bomb),
	})
	sortEvents(events)

	round := &RoundResult{
		RoundNumber:          ctx.roundNumber,
		Phase:                ctx.phase,
		Half:                 ctx.half,
		OvertimeBlock:        ctx.overtimeBlock,
		OvertimeRoundInBlock: ctx.overtimeRoundInBlock,
		IsSideSwitch:         ctx.isSideSwitch,
		Seed:                 roundSeed,
		SideAttacking:        SideT,
		TeamTID:              teamTID,
		TeamCTID:             teamCTID,
		Winner:               winnerSide,
		WinnerTeamID:         winnerTeamID,
		WinReason:            winReason,
		ScoreTeamA:           e.state.scoreTeamA,
		ScoreTeamB:           e.state.scoreTeamB,
		ScoreT:               e.scoreForSide(SideT),
		ScoreCT:              e.scoreForSide(SideCT),
		RouteMain:            route.ID,
		StrategyTemplateID:   template.ID,
		Events:               events,
		PlayerStates:         e.buildPlayerStates(players),
		Bomb:                 bomb,
		FinalControls:        e.buildControls(route, bomb, winnerSide),
	}
	return round
}

func (e *MatchEngine) resolveRoundEvents(events []*GameEvent, players []*roundPlayer, route Route, scenario Scenario, winnerTeamID, winnerSide string, scoreDelta float64, roundSeed int64, encounterRNG, bombRNG *rand.Rand) ([]*GameEvent, *BombPublicState) {
	timestamp := clampInt(scenario.BaseTimeCost, 8, 25)
	if timestamp == 0 {
		timestamp = 12
	}
	firstKill := true
	lastKillSide := ""
	lastKillAt := -99
	plantChance := 35
	if winnerSide == SideT {
		plantChance = 70
	}
	plantPlanned := route.TargetSite != "None" && bombRNG.Intn(100) < plantChance
	winnerTargetAlive := 1 + encounterRNG.Intn(3)
	loserTargetAlive := 0
	if plantPlanned {
		loserTargetAlive = 1 + encounterRNG.Intn(3)
	} else if winnerSide == SideCT && encounterRNG.Intn(100) < 20 {
		loserTargetAlive = 1
	}

	for i := 0; i < len(players); i++ {
		aliveT := filterRoundPlayers(players, SideT, true)
		aliveCT := filterRoundPlayers(players, SideCT, true)
		targetT := survivorTarget(SideT, winnerSide, winnerTargetAlive, loserTargetAlive)
		targetCT := survivorTarget(SideCT, winnerSide, winnerTargetAlive, loserTargetAlive)
		if len(aliveT) <= targetT && len(aliveCT) <= targetCT {
			break
		}
		killerSide := pickKillSide(winnerSide, encounterRNG)
		if killerSide == SideT && len(aliveCT) <= targetCT {
			killerSide = SideCT
		}
		if killerSide == SideCT && len(aliveT) <= targetT {
			killerSide = SideT
		}
		var attacker, victim *roundPlayer
		if killerSide == SideT {
			attacker = pickPlayer(aliveT, encounterRNG, true)
			victim = pickPlayer(aliveCT, encounterRNG, false)
		} else {
			attacker = pickPlayer(aliveCT, encounterRNG, true)
			victim = pickPlayer(aliveT, encounterRNG, false)
		}
		if attacker == nil || victim == nil {
			break
		}
		victim.alive = false
		victim.deaths++
		victim.hasBomb = false
		attacker.kills++
		attacker.damage += 100
		e.recordKill(attacker, victim, firstKill)
		location := e.sampleEventLocation(route, "KILL", fmt.Sprintf("kill_%d_%s", i, victim.profile.PlayerID), roundSeed)
		isTrade := lastKillSide != "" && lastKillSide != attacker.side && timestamp-lastKillAt <= 5
		events = append(events, &GameEvent{
			Timestamp:      int64(timestamp),
			EventType:      EventKill,
			AttackerID:     attacker.profile.PlayerID,
			AttackerName:   attacker.profile.DisplayName,
			AttackerTeamID: attacker.teamID,
			VictimID:       victim.profile.PlayerID,
			VictimName:     victim.profile.DisplayName,
			VictimTeamID:   victim.teamID,
			Weapon:         e.weaponName(attacker.weapon.Primary),
			Location:       location,
			IsFirstKill:    firstKill,
			IsTrade:        isTrade,
			Message:        fmt.Sprintf("[%s] %s 使用 %s 击杀了 %s。", formatTime(int64(timestamp)), attacker.profile.DisplayName, e.weaponName(attacker.weapon.Primary), victim.profile.DisplayName),
			Reason: &EventReason{
				Code:       "combat_duel",
				MainFactor: scenario.Phase,
				ScoreDelta: scoreDelta,
			},
		})
		firstKill = false
		lastKillSide = attacker.side
		lastKillAt = timestamp
		timestamp += 3 + encounterRNG.Intn(8)
	}

	bomb := &BombPublicState{Status: BombStatusCarried}
	if carrier := firstAliveBySide(players, SideT); carrier != nil {
		carrier.hasBomb = true
		bomb.CarrierID = carrier.profile.PlayerID
	} else {
		bomb.Status = BombStatusDropped
	}
	tAlive := len(filterRoundPlayers(players, SideT, true))
	ctAlive := len(filterRoundPlayers(players, SideCT, true))
	shouldPlant := plantPlanned && tAlive > 0 && ctAlive > 0
	if shouldPlant {
		plantAt := clampInt(timestamp+e.input.RuleSet.BasePlantTime, 15, e.input.RuleSet.RoundTimeLimit)
		plantNode := e.plantNode(route.TargetSite)
		bomb = &BombPublicState{
			Status:    BombStatusPlanted,
			Site:      route.TargetSite,
			NodeID:    plantNode.ID,
			CarrierID: bomb.CarrierID,
			PlantedAt: plantAt,
			ExplodeAt: plantAt + e.input.RuleSet.BombExplodeTime,
		}
		events = append(events, &GameEvent{
			Timestamp: int64(plantAt),
			EventType: EventBombPlant,
			Location:  e.sampleNodeLocation(plantNode, roundSeed, "bomb_plant"),
			Message:   fmt.Sprintf("%s 在 %s 包点完成下包。", e.teamName(e.teamIDBySide(SideT)), route.TargetSite),
			Bomb:      cloneBomb(bomb),
		})
		if winnerSide == SideCT {
			defuseAt := plantAt + e.input.RuleSet.BaseDefuseTime + bombRNG.Intn(4)
			bomb.Status = BombStatusDefused
			events = append(events, &GameEvent{
				Timestamp: int64(defuseAt),
				EventType: EventBombDefuse,
				Location:  e.sampleNodeLocation(plantNode, roundSeed, "bomb_defuse"),
				Message:   fmt.Sprintf("%s 完成拆包。", e.teamName(winnerTeamID)),
				Bomb:      cloneBomb(bomb),
			})
			if defuser := firstAliveBySide(players, SideCT); defuser != nil {
				e.stats[defuser.profile.PlayerID].Defuses++
			}
		} else {
			explodeAt := bomb.ExplodeAt
			bomb.Status = BombStatusExplode
			events = append(events, &GameEvent{
				Timestamp: int64(explodeAt),
				EventType: EventBombExplode,
				Location:  e.sampleNodeLocation(plantNode, roundSeed, "bomb_explode"),
				Message:   "炸弹爆炸，T 方守包成功。",
				Bomb:      cloneBomb(bomb),
			})
		}
		if bomb.CarrierID != "" {
			if acc := e.stats[bomb.CarrierID]; acc != nil {
				acc.Plants++
			}
		}
	}
	return events, bomb
}

func (e *MatchEngine) resolveRoundWinner(roundNumber int, teamTID, teamCTID string, route Route, scenario Scenario, rng *rand.Rand) (string, float64) {
	if roundNumber-1 < len(e.input.ForcedRoundWinners) {
		forced := e.input.ForcedRoundWinners[roundNumber-1]
		if forced == e.input.TeamA.TeamID || forced == e.input.TeamB.TeamID {
			return forced, 99
		}
	}
	tPower := e.teamPower(teamTID, scenario, route, SideT)
	ctPower := e.teamPower(teamCTID, scenario, route, SideCT)
	delta := tPower - ctPower
	scale := e.input.MapConfig.CombatConstants.Float("CombatScale", 12)
	if scale <= 0 {
		scale = 12
	}
	tWinChance := 1 / (1 + math.Exp(-delta/scale))
	if rng.Float64() < tWinChance {
		return teamTID, delta
	}
	return teamCTID, delta
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
		if p.PlayerID == "" || p.DisplayName == "" {
			return newError("INVALID_LINEUP", "player id and display name are required")
		}
	}
	return nil
}

func (e *MatchEngine) selectRoute(rng *rand.Rand) Route {
	var routes []Route
	for _, route := range e.input.MapConfig.Routes {
		if route.Side == SideT {
			routes = append(routes, route)
		}
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].ID < routes[j].ID })
	return routes[rng.Intn(len(routes))]
}

func (e *MatchEngine) selectTemplate(route Route, rng *rand.Rand) RouteTemplate {
	var templates []RouteTemplate
	for _, template := range e.input.MapConfig.RouteTemplates {
		if template.TargetSite == route.TargetSite || template.TargetSite == "None" {
			templates = append(templates, template)
		}
	}
	if len(templates) == 0 {
		for _, template := range e.input.MapConfig.RouteTemplates {
			templates = append(templates, template)
		}
	}
	sort.Slice(templates, func(i, j int) bool { return templates[i].ID < templates[j].ID })
	return templates[rng.Intn(len(templates))]
}

func (e *MatchEngine) selectScenario(template RouteTemplate, rng *rand.Rand) Scenario {
	var scenarios []Scenario
	for _, id := range template.ScenarioIDs {
		if scenario, ok := e.input.MapConfig.Scenarios[id]; ok {
			scenarios = append(scenarios, scenario)
		}
	}
	if len(scenarios) == 0 {
		for _, scenario := range e.input.MapConfig.Scenarios {
			scenarios = append(scenarios, scenario)
		}
	}
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	return scenarios[rng.Intn(len(scenarios))]
}

func (e *MatchEngine) buildRoundPlayers(teamTID, teamCTID string, route Route) []*roundPlayer {
	var players []*roundPlayer
	teams := []TeamInput{e.input.TeamA, e.input.TeamB}
	for _, team := range teams {
		side := SideCT
		if team.TeamID == teamTID {
			side = SideT
		}
		for idx, profile := range team.Players {
			nodeID := firstRouteNode(route, side)
			players = append(players, &roundPlayer{
				profile: profile,
				teamID:  team.TeamID,
				side:    side,
				weapon:  e.input.SideLoadouts[side],
				alive:   true,
				hasBomb: side == SideT && idx == 0,
				nodeID:  nodeID,
			})
		}
	}
	sort.Slice(players, func(i, j int) bool { return players[i].profile.PlayerID < players[j].profile.PlayerID })
	return players
}

func (e *MatchEngine) teamPower(teamID string, scenario Scenario, route Route, side string) float64 {
	team := e.team(teamID)
	total := 0.0
	for _, p := range team.Players {
		attr := p.Attributes
		total += float64(attr.Aim)*0.28 + float64(attr.Firepower)*0.24 + float64(attr.Entry)*0.16 + float64(attr.Trade)*0.12 + float64(attr.Clutch)*0.1 + float64(attr.Gamesense)*0.1
	}
	avg := total / float64(len(team.Players))
	if side == SideCT {
		avg += 1.5
	}
	if scenario.Range == "Long" && side == SideT {
		avg += 1
	}
	if route.TargetSite == "B" && side == SideT {
		avg += 0.5
	}
	return avg
}

func (e *MatchEngine) sampleEventLocation(route Route, eventType, eventID string, roundSeed int64) *Location {
	if eventType == EventKill {
		for _, nodeID := range route.Nodes {
			node, ok := e.input.MapConfig.Nodes[nodeID]
			if !ok {
				continue
			}
			if hasUsage(node, "KillSample") || node.NodeType == "Lane" || node.NodeType == "Site" {
				return e.sampleNodeLocation(node, roundSeed, eventID)
			}
		}
	}
	if len(route.Nodes) > 0 {
		if node, ok := e.input.MapConfig.Nodes[route.Nodes[len(route.Nodes)-1]]; ok {
			return e.sampleNodeLocation(node, roundSeed, eventID)
		}
	}
	return &Location{Name: route.Name, X: 0.5, Y: 0.5}
}

func (e *MatchEngine) sampleNodeLocation(node MapNode, roundSeed int64, eventID string) *Location {
	rng := newDerivedRand(roundSeed, "event_location", eventID, node.ID)
	switch node.Shape {
	case "Circle":
		angle := rng.Float64() * 2 * math.Pi
		radius := math.Sqrt(rng.Float64()) * node.Radius
		return &Location{Name: node.Name, X: clampFloat(node.X+math.Cos(angle)*radius, 0.02, 0.98), Y: clampFloat(node.Y+math.Sin(angle)*radius, 0.02, 0.98)}
	case "Polygon":
		points := parsePolygon(node.Points)
		if len(points) >= 3 {
			minX, maxX, minY, maxY := polygonBounds(points)
			for i := 0; i < 20; i++ {
				x := minX + rng.Float64()*(maxX-minX)
				y := minY + rng.Float64()*(maxY-minY)
				if pointInPolygon(x, y, points) {
					return &Location{Name: node.Name, X: x, Y: y}
				}
			}
		}
	}
	return &Location{Name: node.Name, X: clampFloat(node.X+rng.Float64()*0.02-0.01, 0.02, 0.98), Y: clampFloat(node.Y+rng.Float64()*0.02-0.01, 0.02, 0.98)}
}

func (e *MatchEngine) plantNode(site string) MapNode {
	var nodes []MapNode
	for _, node := range e.input.MapConfig.Nodes {
		if node.Site == site && hasUsage(node, "Plant") {
			nodes = append(nodes, node)
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	if len(nodes) > 0 {
		return nodes[0]
	}
	return MapNode{ID: site + "_SITE", Name: site + " Site", Site: site, X: 0.5, Y: 0.5, Shape: "None"}
}

func (e *MatchEngine) buildPlayerStates(players []*roundPlayer) []*PlayerState {
	states := make([]*PlayerState, 0, len(players))
	for _, p := range players {
		states = append(states, &PlayerState{
			PlayerID:    p.profile.PlayerID,
			PlayerName:  p.profile.DisplayName,
			DisplayName: p.profile.DisplayName,
			Portrait:    p.profile.Portrait,
			TeamID:      p.teamID,
			Side:        p.side,
			IsAlive:     p.alive,
			Alive:       p.alive,
			HP:          boolToHP(p.alive),
			Stamina:     100,
			Focus:       100,
			CurrentNode: p.nodeID,
			HasBomb:     p.hasBomb,
			Kills:       p.kills,
			Deaths:      p.deaths,
			Damage:      p.damage,
			RoleTags:    append([]string(nil), p.profile.RoleTags...),
			Weapon:      p.weapon,
		})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].PlayerID < states[j].PlayerID })
	return states
}

func (e *MatchEngine) buildControls(route Route, bomb *BombPublicState, winnerSide string) []*NodeControlState {
	var controls []*NodeControlState
	for _, nodeID := range route.Nodes {
		status := "Contested"
		if nodeID == bomb.NodeID || nodeID == route.Nodes[len(route.Nodes)-1] {
			if winnerSide == SideT {
				status = "TControlled"
			} else {
				status = "CTControlled"
			}
		}
		controls = append(controls, &NodeControlState{NodeID: nodeID, Status: status, KnownByT: true, KnownByCT: true, UpdatedAt: e.input.RuleSet.RoundTimeLimit})
	}
	return controls
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
			PlayerID:   acc.PlayerID,
			PlayerName: acc.PlayerName,
			TeamID:     acc.TeamID,
			Side:       acc.Side,
			Kills:      acc.Kills,
			Deaths:     acc.Deaths,
			Damage:     acc.Damage,
			ADR:        math.Round(adr*10) / 10,
			FK:         acc.FK,
			MK:         acc.MK,
			Plants:     acc.Plants,
			Defuses:    acc.Defuses,
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
				PlayerID:   p.PlayerID,
				PlayerName: p.DisplayName,
				TeamID:     team.TeamID,
				Side:       e.input.InitialSideByTeam[team.TeamID],
			}
		}
	}
	return stats
}

func (e *MatchEngine) recordKill(attacker, victim *roundPlayer, firstKill bool) {
	if acc := e.stats[attacker.profile.PlayerID]; acc != nil {
		acc.Kills++
		acc.Damage += 100
		if firstKill {
			acc.FK++
		}
	}
	if acc := e.stats[victim.profile.PlayerID]; acc != nil {
		acc.Deaths++
	}
}

func (e *MatchEngine) swapSides() {
	for teamID, side := range e.state.sides {
		if side == SideT {
			e.state.sides[teamID] = SideCT
		} else {
			e.state.sides[teamID] = SideT
		}
	}
}

func (e *MatchEngine) winnerTeamID() string {
	if e.state.scoreTeamA > e.state.scoreTeamB {
		return e.input.TeamA.TeamID
	}
	if e.state.scoreTeamB > e.state.scoreTeamA {
		return e.input.TeamB.TeamID
	}
	return ""
}

func (e *MatchEngine) winnerSide(teamID string) string {
	if teamID == "" {
		return ""
	}
	return e.state.sides[teamID]
}

func (e *MatchEngine) scoreForSide(side string) int {
	teamID := e.teamIDBySide(side)
	if teamID == e.input.TeamA.TeamID {
		return e.state.scoreTeamA
	}
	return e.state.scoreTeamB
}

func (e *MatchEngine) teamIDBySide(side string) string {
	for teamID, currentSide := range e.state.sides {
		if currentSide == side {
			return teamID
		}
	}
	return ""
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

func (e *MatchEngine) weaponName(id string) string {
	if spec, ok := e.input.WeaponSpecs[id]; ok && spec.DisplayName != "" {
		return spec.DisplayName
	}
	return id
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
		if events[i].Timestamp != events[j].Timestamp {
			return events[i].Timestamp < events[j].Timestamp
		}
		if priority[events[i].EventType] != priority[events[j].EventType] {
			return priority[events[i].EventType] < priority[events[j].EventType]
		}
		return events[i].Message < events[j].Message
	})
}

func pickKillSide(winnerSide string, rng *rand.Rand) string {
	if rng.Intn(100) < 68 {
		return winnerSide
	}
	if winnerSide == SideT {
		return SideCT
	}
	return SideT
}

func survivorTarget(side, winnerSide string, winnerAlive, loserAlive int) int {
	if side == winnerSide {
		return winnerAlive
	}
	return loserAlive
}

func pickPlayer(players []*roundPlayer, rng *rand.Rand, preferStrong bool) *roundPlayer {
	if len(players) == 0 {
		return nil
	}
	sort.Slice(players, func(i, j int) bool {
		powerI := playerPower(players[i].profile)
		powerJ := playerPower(players[j].profile)
		if powerI == powerJ {
			return players[i].profile.PlayerID < players[j].profile.PlayerID
		}
		if preferStrong {
			return powerI > powerJ
		}
		return powerI < powerJ
	})
	limit := minInt(3, len(players))
	return players[rng.Intn(limit)]
}

func playerPower(p PlayerProfile) int {
	return p.Attributes.Aim + p.Attributes.Firepower + p.Attributes.Entry + p.Attributes.Trade + p.Attributes.Clutch + p.Attributes.Gamesense
}

func filterRoundPlayers(players []*roundPlayer, side string, alive bool) []*roundPlayer {
	var out []*roundPlayer
	for _, p := range players {
		if p.side == side && p.alive == alive {
			out = append(out, p)
		}
	}
	return out
}

func firstAliveBySide(players []*roundPlayer, side string) *roundPlayer {
	list := filterRoundPlayers(players, side, true)
	sort.Slice(list, func(i, j int) bool { return list[i].profile.PlayerID < list[j].profile.PlayerID })
	if len(list) == 0 {
		return nil
	}
	return list[0]
}

func firstRouteNode(route Route, side string) string {
	if len(route.Nodes) == 0 {
		return ""
	}
	if side == SideCT {
		return route.Nodes[len(route.Nodes)-1]
	}
	return route.Nodes[0]
}

func inferWinReason(winnerSide string, bomb *BombPublicState, tAlive, ctAlive int) string {
	if bomb != nil {
		switch bomb.Status {
		case BombStatusDefused:
			return "bomb_defused"
		case BombStatusExplode:
			return "bomb_exploded"
		}
	}
	if tAlive == 0 || ctAlive == 0 {
		return "elimination"
	}
	if winnerSide == SideCT {
		return "timeout"
	}
	return "elimination"
}

func cloneBomb(b *BombPublicState) *BombPublicState {
	if b == nil {
		return nil
	}
	cp := *b
	return &cp
}

func boolToHP(alive bool) int {
	if alive {
		return 100
	}
	return 0
}

func hasUsage(node MapNode, usage string) bool {
	for _, item := range node.AreaUsages {
		if item == usage {
			return true
		}
	}
	return false
}

type point struct {
	x float64
	y float64
}

func parsePolygon(points string) []point {
	var out []point
	for _, pair := range splitNonEmpty(points, ";") {
		xy := splitNonEmpty(pair, ",")
		if len(xy) != 2 {
			continue
		}
		var p point
		_, _ = fmt.Sscanf(xy[0], "%f", &p.x)
		_, _ = fmt.Sscanf(xy[1], "%f", &p.y)
		out = append(out, p)
	}
	return out
}

func polygonBounds(points []point) (minX, maxX, minY, maxY float64) {
	minX, maxX, minY, maxY = points[0].x, points[0].x, points[0].y, points[0].y
	for _, p := range points[1:] {
		minX = math.Min(minX, p.x)
		maxX = math.Max(maxX, p.x)
		minY = math.Min(minY, p.y)
		maxY = math.Max(maxY, p.y)
	}
	return
}

func pointInPolygon(x, y float64, poly []point) bool {
	inside := false
	j := len(poly) - 1
	for i := 0; i < len(poly); i++ {
		xi, yi := poly[i].x, poly[i].y
		xj, yj := poly[j].x, poly[j].y
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatTime(seconds int64) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
