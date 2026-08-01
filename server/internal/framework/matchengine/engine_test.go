package matchengine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTestInput(seed int64, winners []string) *MatchInput {
	cfg := makeTestMapConfig()
	return &MatchInput{
		MatchID:    "test_match",
		MapID:      DefaultMapID,
		MapName:    DefaultMapName,
		MapVersion: "test-map-v1",
		Seed:       seed,
		RuleSet:    DefaultMR12RuleSet(cfg.CombatConstants),
		TeamA:      makeTestTeam("team_a", "Team A", 82),
		TeamB:      makeTestTeam("team_b", "Team B", 80),
		InitialSideByTeam: map[string]string{
			"team_a": SideT,
			"team_b": SideCT,
		},
		MapConfig:          cfg,
		WeaponSpecs:        makeTestWeapons(),
		SideLoadouts:       makeTestLoadouts(),
		ForcedRoundWinners: winners,
	}
}

func makeTestTeam(teamID, name string, base int) TeamInput {
	players := make([]PlayerProfile, 0, 5)
	for i := 0; i < 5; i++ {
		value := base - i
		players = append(players, PlayerProfile{
			PlayerID:    teamID + "_p" + string(rune('1'+i)),
			DisplayName: name + " P" + string(rune('1'+i)),
			RoleTags:    []string{"Rifler"},
			Attributes: PlayerAttributes{
				Entry: value, Aim: value, Trade: value, Clutch: value, Firepower: value, Gamesense: value,
				Reaction: value, Positioning: value, Awareness: value, Teamplay: value, Utility: value, Composure: value, Mobility: value, Endurance: value, Discipline: value,
			},
		})
	}
	return TeamInput{TeamID: teamID, Name: name, Players: players}
}

func makeTestWeapons() map[string]WeaponSpec {
	return map[string]WeaponSpec{
		WeaponAK47:  {ID: WeaponAK47, DisplayName: "AK-47", Damage: 36, RoundsPerMinute: 600, MagazineSize: 30, ArmorPenetration: 0.775, RangeModifier: 0.98},
		WeaponM4A1S: {ID: WeaponM4A1S, DisplayName: "M4A1-S", Damage: 38, RoundsPerMinute: 600, MagazineSize: 20, ArmorPenetration: 0.7, RangeModifier: 0.99},
	}
}

func makeTestLoadouts() map[string]WeaponLoadout {
	return map[string]WeaponLoadout{
		SideT:  {Primary: WeaponAK47, Armor: true, Helmet: true},
		SideCT: {Primary: WeaponM4A1S, Armor: true, Helmet: true, HasKit: true},
	}
}

func makeTestMapConfig() *MapConfig {
	consts := CombatConstants{Values: map[string]CombatConstValue{}}
	for _, key := range requiredCombatConstKeys {
		valueType := "Int"
		value := "10"
		if strings.Contains(key, "Scale") || strings.Contains(key, "Noise") || strings.Contains(key, "Gap") || strings.Contains(key, "Chance") || strings.Contains(key, "Penalty") || strings.Contains(key, "Bonus") || strings.Contains(key, "Weight") {
			valueType = "Float"
			value = "10"
		}
		consts.Values[key] = CombatConstValue{Key: key, ValueType: valueType, Value: value}
	}
	consts.Values["RoundTimeLimit"] = CombatConstValue{Key: "RoundTimeLimit", ValueType: "Int", Value: "115"}
	consts.Values["BombExplodeTime"] = CombatConstValue{Key: "BombExplodeTime", ValueType: "Int", Value: "40"}
	consts.Values["BasePlantTime"] = CombatConstValue{Key: "BasePlantTime", ValueType: "Int", Value: "4"}
	consts.Values["BaseDefuseTime"] = CombatConstValue{Key: "BaseDefuseTime", ValueType: "Int", Value: "10"}
	consts.Values["BasePickupTime"] = CombatConstValue{Key: "BasePickupTime", ValueType: "Int", Value: "2"}
	consts.Values["MaxDecisionCount"] = CombatConstValue{Key: "MaxDecisionCount", ValueType: "Int", Value: "3"}
	consts.Values["MaxEncounterPulses"] = CombatConstValue{Key: "MaxEncounterPulses", ValueType: "Int", Value: "3"}

	return &MapConfig{
		MapID:   DefaultMapID,
		MapName: DefaultMapName,
		Version: "test-map-v1",
		RouteTemplates: map[string]RouteTemplate{
			"TPL_A": {ID: "TPL_A", MapID: DefaultMapID, TargetSite: "A", Tempo: "Default", RecommendedMin: 2, RecommendedMax: 5, ScenarioIDs: []string{"SCN_A"}, MapTagIDs: []string{"TAG_A"}, SuccessNextPhase: "SiteEntry"},
		},
		Scenarios: map[string]Scenario{
			"SCN_A": {ID: "SCN_A", Route: "A_Long", Phase: "SiteEntry", Range: "Long", Site: "A", Tempo: "SlowDefault", Posture: "T_Executing", UtilityContext: "Even", MapTagIDs: []string{"TAG_A"}, BaseTimeCost: 12, BaseWeight: 10},
		},
		MapTags: map[string]MapTag{
			"TAG_A": {ID: "TAG_A", MapID: DefaultMapID, Category: "Range", Value: "LongRange", Side: "Both", Weight: 10, ReasonCode: "long_range_duel"},
		},
		EncounterModifiers: map[string]EncounterModifier{
			"MOD_A": {ID: "MOD_A", ScenarioID: "SCN_A", Factor: "LongRange", Side: "Both", Attribute: "aim", Weight: 8, ReasonCode: "aim_long_range"},
		},
		Nodes: map[string]MapNode{
			"T_SPAWN":   {ID: "T_SPAWN", MapID: DefaultMapID, Name: "T Spawn", Site: "None", NodeType: "Spawn", DefaultSide: SideT, X: 0.5, Y: 0.9, Floor: "Ground", Shape: "None"},
			"LONG_DOOR": {ID: "LONG_DOOR", MapID: DefaultMapID, Name: "Long Door", Site: "A", NodeType: "Connector", DefaultSide: "Both", X: 0.65, Y: 0.65, Floor: "Ground", Shape: "None"},
			"A_LONG":    {ID: "A_LONG", MapID: DefaultMapID, Name: "A Long", Site: "A", NodeType: "Lane", DefaultSide: "Both", X: 0.85, Y: 0.55, Floor: "Ground", Shape: "Circle", Radius: 0.03, AreaUsages: []string{"KillSample"}},
			"A_SITE":    {ID: "A_SITE", MapID: DefaultMapID, Name: "A Site", Site: "A", NodeType: "Site", DefaultSide: SideCT, X: 0.75, Y: 0.25, Floor: "Ground", Shape: "Circle", Radius: 0.05, AreaUsages: []string{"Plant", "KillSample"}},
			"B_SITE":    {ID: "B_SITE", MapID: DefaultMapID, Name: "B Site", Site: "B", NodeType: "Site", DefaultSide: SideCT, X: 0.2, Y: 0.2, Floor: "Ground", Shape: "Circle", Radius: 0.05, AreaUsages: []string{"Plant", "KillSample"}},
		},
		Edges: map[string]MapEdge{
			"E1": {ID: "E1", FromNode: "T_SPAWN", ToNode: "LONG_DOOR", BaseTime: 8, Bidirectional: true},
			"E2": {ID: "E2", FromNode: "LONG_DOOR", ToNode: "A_LONG", BaseTime: 8, Bidirectional: true},
			"E3": {ID: "E3", FromNode: "A_LONG", ToNode: "A_SITE", BaseTime: 8, Bidirectional: true},
		},
		Visibility: map[string]Visibility{},
		Routes: map[string]Route{
			"D2_A_LONG": {ID: "D2_A_LONG", Name: "A Long Execute", Side: SideT, TargetSite: "A", Nodes: []string{"T_SPAWN", "LONG_DOOR", "A_LONG", "A_SITE"}, MinPlayers: 2, MaxPlayers: 5},
		},
		CombatConstants: consts,
	}
}

func TestRegulationEarlyWin(t *testing.T) {
	winners := repeatWinner("team_a", 13)
	res, err := NewMatchEngine(makeTestInput(100, winners)).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}
	if res.WinnerTeamID != "team_a" {
		t.Fatalf("expected team_a winner, got %s", res.WinnerTeamID)
	}
	if res.TotalRounds != 13 {
		t.Fatalf("expected 13 rounds, got %d", res.TotalRounds)
	}
	if res.FinalScoreTeamA != 13 || res.FinalScoreTeamB != 0 {
		t.Fatalf("unexpected score %d:%d", res.FinalScoreTeamA, res.FinalScoreTeamB)
	}
}

func TestHalftimeSideSwitchKeepsTeamScore(t *testing.T) {
	winners := append(repeatWinner("team_a", 8), repeatWinner("team_b", 4)...)
	winners = append(winners, "team_a")
	res, err := NewMatchEngine(makeTestInput(101, winners)).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}
	if len(res.Rounds) < 13 {
		t.Fatalf("expected at least 13 rounds")
	}
	round13 := res.Rounds[12]
	if round13.TeamTID != "team_b" || round13.TeamCTID != "team_a" {
		t.Fatalf("round 13 sides wrong: T=%s CT=%s", round13.TeamTID, round13.TeamCTID)
	}
	if round13.ScoreTeamA != 9 || round13.ScoreTeamB != 4 {
		t.Fatalf("team score did not carry through switch: %d:%d", round13.ScoreTeamA, round13.ScoreTeamB)
	}
}

func TestOvertimeBlockSideRulesAndWinner(t *testing.T) {
	winners := regulationTieWinners()
	winners = append(winners, "team_b", "team_b", "team_a", "team_b", "team_a", "team_b")
	res, err := NewMatchEngine(makeTestInput(102, winners)).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}
	if res.TotalRounds != 30 {
		t.Fatalf("expected full OT block for 30 rounds, got %d", res.TotalRounds)
	}
	if res.WinnerTeamID != "team_b" {
		t.Fatalf("expected team_b winner, got %s", res.WinnerTeamID)
	}
	for i := 24; i < 27; i++ {
		if res.Rounds[i].Phase != "overtime" || res.Rounds[i].TeamCTID != "team_a" {
			t.Fatalf("OT1 first half should continue second-half sides at round %d", i+1)
		}
	}
	for i := 27; i < 30; i++ {
		if res.Rounds[i].TeamTID != "team_a" {
			t.Fatalf("OT1 second half should switch sides at round %d", i+1)
		}
	}
}

func TestOvertimeDoesNotEndEarlyInsideBlock(t *testing.T) {
	winners := regulationTieWinners()
	winners = append(winners, repeatWinner("team_a", 4)...)
	res, err := NewMatchEngine(makeTestInput(103, winners)).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}
	if res.TotalRounds != 30 {
		t.Fatalf("expected complete OT block, got %d rounds", res.TotalRounds)
	}
}

func TestSameSeedProducesSameMatch(t *testing.T) {
	input := makeTestInput(12345, nil)
	res1, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("first simulate failed: %v", err)
	}
	res2, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("second simulate failed: %v", err)
	}
	if res1.WinnerTeamID != res2.WinnerTeamID || res1.TotalRounds != res2.TotalRounds {
		t.Fatalf("match summary differs")
	}
	for i := range res1.Rounds {
		a, b := res1.Rounds[i], res2.Rounds[i]
		if a.WinnerTeamID != b.WinnerTeamID || a.TeamTID != b.TeamTID || a.ScoreTeamA != b.ScoreTeamA || len(a.Events) != len(b.Events) {
			t.Fatalf("round %d differs", i+1)
		}
		for j := range a.Events {
			if a.Events[j].Message != b.Events[j].Message || a.Events[j].Timestamp != b.Events[j].Timestamp {
				t.Fatalf("event %d in round %d differs", j, i+1)
			}
		}
	}
}

func TestRoundTerminalStateMatchesWinReason(t *testing.T) {
	seenReasons := map[string]bool{}
	for seed := int64(1); seed <= 200; seed++ {
		res, err := NewMatchEngine(makeTestInput(seed, nil)).StartMatch(context.Background())
		if err != nil {
			t.Fatalf("seed %d simulate failed: %v", seed, err)
		}
		for _, round := range res.Rounds {
			tAlive, ctAlive := 0, 0
			for _, player := range round.PlayerStates {
				if !player.Alive {
					continue
				}
				if player.Side == SideT {
					tAlive++
				} else if player.Side == SideCT {
					ctAlive++
				}
			}

			hasPlant := false
			roundEndAt := int64(-1)
			for _, event := range round.Events {
				if event.EventType == EventBombPlant {
					hasPlant = true
				}
				if event.EventType == EventRoundEnd {
					roundEndAt = event.Timestamp
				}
			}

			seenReasons[round.WinReason] = true
			switch round.WinReason {
			case "elimination":
				loserAlive := ctAlive
				if round.Winner == SideCT {
					loserAlive = tAlive
				}
				if loserAlive != 0 {
					t.Fatalf("seed %d round %d: elimination left %d losing players alive", seed, round.RoundNumber, loserAlive)
				}
				if hasPlant {
					t.Fatalf("seed %d round %d: elimination round contains a bomb plant", seed, round.RoundNumber)
				}
			case "timeout":
				if round.Winner != SideCT || tAlive == 0 || ctAlive == 0 || hasPlant {
					t.Fatalf("seed %d round %d: invalid timeout state winner=%s alive=%d:%d plant=%v", seed, round.RoundNumber, round.Winner, tAlive, ctAlive, hasPlant)
				}
				if roundEndAt != int64(makeTestInput(seed, nil).RuleSet.RoundTimeLimit) {
					t.Fatalf("seed %d round %d: timeout ended at %d", seed, round.RoundNumber, roundEndAt)
				}
			case "bomb_defused":
				if round.Winner != SideCT || tAlive == 0 || ctAlive == 0 || !hasPlant || round.Bomb == nil || round.Bomb.Status != BombStatusDefused {
					t.Fatalf("seed %d round %d: invalid defuse state winner=%s alive=%d:%d plant=%v bomb=%+v", seed, round.RoundNumber, round.Winner, tAlive, ctAlive, hasPlant, round.Bomb)
				}
			case "bomb_exploded":
				if round.Winner != SideT || tAlive == 0 || ctAlive == 0 || !hasPlant || round.Bomb == nil || round.Bomb.Status != BombStatusExplode {
					t.Fatalf("seed %d round %d: invalid explosion state winner=%s alive=%d:%d plant=%v bomb=%+v", seed, round.RoundNumber, round.Winner, tAlive, ctAlive, hasPlant, round.Bomb)
				}
			default:
				t.Fatalf("seed %d round %d: unsupported win reason %q", seed, round.RoundNumber, round.WinReason)
			}
		}
	}

	for _, reason := range []string{"elimination", "timeout", "bomb_defused", "bomb_exploded"} {
		if !seenReasons[reason] {
			t.Fatalf("expected deterministic seed sweep to cover %s", reason)
		}
	}
}

func TestValidateMapConfigBadRouteNode(t *testing.T) {
	cfg := makeTestMapConfig()
	route := cfg.Routes["D2_A_LONG"]
	route.Nodes = append(route.Nodes, "MISSING_NODE")
	cfg.Routes["D2_A_LONG"] = route
	err := ValidateMapConfig(cfg)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if me, ok := err.(*EngineError); !ok || me.Code != "CONFIG_BAD_ROUTE_NODE" {
		t.Fatalf("expected CONFIG_BAD_ROUTE_NODE, got %v", err)
	}
}

func TestValidateMapConfigMissingCombatConst(t *testing.T) {
	cfg := makeTestMapConfig()
	delete(cfg.CombatConstants.Values, "RoundTimeLimit")
	err := ValidateMapConfig(cfg)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if me, ok := err.(*EngineError); !ok || me.Code != "CONFIG_MISSING_COMBAT_CONST" {
		t.Fatalf("expected CONFIG_MISSING_COMBAT_CONST, got %v", err)
	}
}

func TestValidateMapConfigBadRiskAndVisibilityRefs(t *testing.T) {
	cfg := makeTestMapConfig()
	cfg.Edges["E1"] = MapEdge{ID: "E1", FromNode: "T_SPAWN", ToNode: "LONG_DOOR", RiskPoints: []string{"MISSING_RISK"}, Bidirectional: true}
	err := ValidateMapConfig(cfg)
	if me, ok := err.(*EngineError); !ok || me.Code != "CONFIG_BAD_RISK_POINT" {
		t.Fatalf("expected CONFIG_BAD_RISK_POINT, got %v", err)
	}

	cfg = makeTestMapConfig()
	cfg.Visibility["BAD_VIS"] = Visibility{ID: "BAD_VIS", FromNode: "T_SPAWN", ToNode: "MISSING_NODE", Visible: true}
	err = ValidateMapConfig(cfg)
	if me, ok := err.(*EngineError); !ok || me.Code != "CONFIG_BAD_VISIBILITY_NODE" {
		t.Fatalf("expected CONFIG_BAD_VISIBILITY_NODE, got %v", err)
	}
}

func TestMatchengineDoesNotImportConfigPackage(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s failed: %v", file, err)
		}
		if strings.Contains(string(body), "windypath.com/cs2match/config") {
			t.Fatalf("%s imports generated config package", file)
		}
	}
}

func repeatWinner(teamID string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = teamID
	}
	return out
}

func regulationTieWinners() []string {
	winners := make([]string, 0, 24)
	for i := 0; i < 12; i++ {
		winners = append(winners, "team_a", "team_b")
	}
	return winners
}
