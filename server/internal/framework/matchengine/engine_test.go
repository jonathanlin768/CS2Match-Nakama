package matchengine

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTestInput(seed int64) *MatchInput {
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
		MapConfig:    cfg,
		WeaponSpecs:  makeTestWeapons(),
		SideLoadouts: makeTestLoadouts(),
	}
}

func makeTestTeam(teamID, name string, base int) TeamInput {
	players := make([]PlayerProfile, 0, 5)
	for i := 0; i < 5; i++ {
		value := base - i
		players = append(players, PlayerProfile{
			PlayerID:       teamID + "_p" + string(rune('1'+i)),
			ConfigPlayerID: "config_p" + string(rune('1'+i)),
			DisplayName:    name + " P" + string(rune('1'+i)),
			RoleTags:       []string{"Rifler"},
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
		valueType := expectedCombatConstTypes[key]
		value := "10"
		if valueType == "String" {
			value = "TPL_A"
		}
		consts.Values[key] = CombatConstValue{Key: key, ValueType: valueType, Value: value}
	}
	consts.Values["RoundTimeLimit"] = CombatConstValue{Key: "RoundTimeLimit", ValueType: "Int", Value: "115"}
	consts.Values["BombExplodeTime"] = CombatConstValue{Key: "BombExplodeTime", ValueType: "Int", Value: "40"}
	consts.Values["BasePlantTime"] = CombatConstValue{Key: "BasePlantTime", ValueType: "Int", Value: "4"}
	consts.Values["BaseDefuseTime"] = CombatConstValue{Key: "BaseDefuseTime", ValueType: "Int", Value: "10"}
	consts.Values["BasePickupTime"] = CombatConstValue{Key: "BasePickupTime", ValueType: "Int", Value: "2"}
	consts.Values["DecisionDelay"] = CombatConstValue{Key: "DecisionDelay", ValueType: "Int", Value: "2"}
	consts.Values["ForceExecuteThreshold"] = CombatConstValue{Key: "ForceExecuteThreshold", ValueType: "Int", Value: "20"}
	consts.Values["MaxDecisionCount"] = CombatConstValue{Key: "MaxDecisionCount", ValueType: "Int", Value: "3"}
	consts.Values["MaxEncounterPulses"] = CombatConstValue{Key: "MaxEncounterPulses", ValueType: "Int", Value: "3"}
	consts.Values["CommunicationDelay"] = CombatConstValue{Key: "CommunicationDelay", ValueType: "Int", Value: "0"}
	consts.Values["MaxRoundTimeline"] = CombatConstValue{Key: "MaxRoundTimeline", ValueType: "Int", Value: "180"}
	consts.Values["MaxStateTransitions"] = CombatConstValue{Key: "MaxStateTransitions", ValueType: "Int", Value: "1000"}
	consts.Values["MaxScheduledActions"] = CombatConstValue{Key: "MaxScheduledActions", ValueType: "Int", Value: "1000"}
	consts.Values["MaxEffectsPerTimestamp"] = CombatConstValue{Key: "MaxEffectsPerTimestamp", ValueType: "Int", Value: "100"}
	consts.Values["MaxNoOpTransitions"] = CombatConstValue{Key: "MaxNoOpTransitions", ValueType: "Int", Value: "4"}
	consts.Values["MaxRotationsPerTeam"] = CombatConstValue{Key: "MaxRotationsPerTeam", ValueType: "Int", Value: "3"}
	consts.Values["SoundIntelMinConfidence"] = CombatConstValue{Key: "SoundIntelMinConfidence", ValueType: "Int", Value: "30"}
	consts.Values["SoundIntelMaxConfidence"] = CombatConstValue{Key: "SoundIntelMaxConfidence", ValueType: "Int", Value: "70"}
	consts.Values["DeathIntelMaxConfidence"] = CombatConstValue{Key: "DeathIntelMaxConfidence", ValueType: "Int", Value: "70"}
	consts.Values["ControlIntelTTL"] = CombatConstValue{Key: "ControlIntelTTL", ValueType: "Int", Value: "12"}
	consts.Values["UtilityBudget"] = CombatConstValue{Key: "UtilityBudget", ValueType: "Int", Value: "100"}
	consts.Values["MinStrategyWeight"] = CombatConstValue{Key: "MinStrategyWeight", ValueType: "Float", Value: "1"}
	consts.Values["MaxStrategyWeight"] = CombatConstValue{Key: "MaxStrategyWeight", ValueType: "Float", Value: "100"}
	consts.Values["MaxRandomNoise"] = CombatConstValue{Key: "MaxRandomNoise", ValueType: "Float", Value: "8"}
	consts.Values["CloseScoreGap"] = CombatConstValue{Key: "CloseScoreGap", ValueType: "Float", Value: "6"}
	consts.Values["DecisiveScoreGap"] = CombatConstValue{Key: "DecisiveScoreGap", ValueType: "Float", Value: "18"}
	consts.Values["StrategyRepeatPenalty"] = CombatConstValue{Key: "StrategyRepeatPenalty", ValueType: "Float", Value: "3"}
	consts.Values["SuccessBonusPerRound"] = CombatConstValue{Key: "SuccessBonusPerRound", ValueType: "Float", Value: "1.5"}
	consts.Values["MaxPreviousSuccessBonus"] = CombatConstValue{Key: "MaxPreviousSuccessBonus", ValueType: "Float", Value: "6"}
	consts.Values["CounterMemoryBonus"] = CombatConstValue{Key: "CounterMemoryBonus", ValueType: "Float", Value: "3"}
	consts.Values["DefaultStrategyTemplateID"] = CombatConstValue{Key: "DefaultStrategyTemplateID", ValueType: "String", Value: "TPL_A"}
	consts.Values["DefaultCTSetupTemplateID"] = CombatConstValue{Key: "DefaultCTSetupTemplateID", ValueType: "String", Value: "TPL_CT"}
	for _, pair := range [][3]string{
		{"MinPickupTime", "1", "Int"}, {"MaxPickupTime", "4", "Int"},
		{"MinCombatDuration", "2", "Int"}, {"MaxCombatDuration", "12", "Int"},
		{"MinDamagePotential", "0.2", "Float"}, {"MaxDamagePotential", "1.2", "Float"},
		{"MinExposureModifier", "0.5", "Float"}, {"MaxExposureModifier", "1.5", "Float"},
		{"MinIntelTTL", "3", "Int"}, {"MaxIntelTTL", "20", "Int"},
		{"MinAttribute", "0", "Int"}, {"MaxAttribute", "100", "Int"},
		{"MinHP", "0", "Int"}, {"MaxHP", "100", "Int"},
		{"MinStamina", "0", "Int"}, {"MaxStamina", "100", "Int"},
		{"MinFocus", "0", "Int"}, {"MaxFocus", "100", "Int"},
		{"MinHitChance", "0.05", "Float"}, {"MaxHitChance", "0.95", "Float"}, {"MaxKillChance", "0.85", "Float"},
		{"MinPlantTime", "3", "Int"}, {"MaxPlantTime", "8", "Int"},
		{"MinDefuseTime", "5", "Int"}, {"MaxDefuseTime", "12", "Int"},
		{"MinMoveTime", "1", "Int"}, {"MaxMoveTime", "30", "Int"},
	} {
		consts.Values[pair[0]] = CombatConstValue{Key: pair[0], ValueType: pair[2], Value: pair[1]}
	}
	modifiers := map[string]EncounterModifier{
		"MOD_A": {ID: "MOD_A", ScenarioID: "SCN_A", Factor: "LongRange", Side: "Both", Attribute: "aim", Weight: 8, ReasonCode: "aim_long_range"},
	}
	for index, attribute := range scenarioWeightAttributes {
		id := "WEIGHT_" + strings.ToUpper(attribute)
		modifiers[id] = EncounterModifier{ID: id, ScenarioID: "SCN_A", Factor: "ScenarioWeight", Side: "Both", Attribute: attribute, Weight: 10, ReasonCode: "WEIGHT_" + strings.ToUpper(attribute)}
		_ = index
	}

	return &MapConfig{
		MapID:   DefaultMapID,
		MapName: DefaultMapName,
		Version: "test-map-v1",
		RouteTemplates: map[string]RouteTemplate{
			"TPL_A":  {ID: "TPL_A", MapID: DefaultMapID, Side: SideT, TargetSite: "A", Tempo: "Default", RecommendedMin: 2, RecommendedMax: 5, RouteIDs: []string{"D2_A_LONG"}, RouteAllocations: map[string]int{"D2_A_LONG": 5}, ScenarioIDs: []string{"SCN_A"}, MapTagIDs: []string{"TAG_A"}, SuccessNextPhase: "SiteEntry"},
			"TPL_CT": {ID: "TPL_CT", MapID: DefaultMapID, Side: SideCT, TargetSite: "None", Tempo: "Default", RecommendedMin: 1, RecommendedMax: 5, RouteIDs: []string{"D2_CT_A"}, RouteAllocations: map[string]int{"D2_CT_A": 5}, ScenarioIDs: []string{"SCN_A"}, MapTagIDs: []string{"TAG_A"}, SuccessNextPhase: "Advance"},
		},
		Scenarios: map[string]Scenario{
			"SCN_A": {ID: "SCN_A", Route: "A_Long", Phase: "SiteEntry", Range: "Long", Site: "A", Tempo: "SlowDefault", Posture: "T_Executing", UtilityContext: "Even", MapTagIDs: []string{"TAG_A"}, BaseTimeCost: 12, BaseWeight: 10},
		},
		MapTags: map[string]MapTag{
			"TAG_A": {ID: "TAG_A", MapID: DefaultMapID, Category: "Range", Value: "LongRange", Side: "Both", Weight: 10, ReasonCode: "long_range_duel"},
		},
		EncounterModifiers: modifiers,
		Nodes: map[string]MapNode{
			"T_SPAWN":   {ID: "T_SPAWN", MapID: DefaultMapID, Name: "T Spawn", Site: "None", NodeType: "Spawn", DefaultSide: SideT, X: 0.5, Y: 0.9, Floor: "Ground", Shape: "None"},
			"CT_SPAWN":  {ID: "CT_SPAWN", MapID: DefaultMapID, Name: "CT Spawn", Site: "None", NodeType: "Spawn", DefaultSide: SideCT, X: 0.5, Y: 0.1, Floor: "Ground", Shape: "None"},
			"LONG_DOOR": {ID: "LONG_DOOR", MapID: DefaultMapID, Name: "Long Door", Site: "A", NodeType: "Connector", DefaultSide: "Both", X: 0.65, Y: 0.65, Floor: "Ground", Shape: "None"},
			"A_LONG":    {ID: "A_LONG", MapID: DefaultMapID, Name: "A Long", Site: "A", NodeType: "Lane", DefaultSide: "Both", X: 0.85, Y: 0.55, Floor: "Ground", Shape: "Circle", Radius: 0.03, AreaUsages: []string{"KillSample"}},
			"A_SITE":    {ID: "A_SITE", MapID: DefaultMapID, Name: "A Site", Site: "A", NodeType: "Site", DefaultSide: SideCT, X: 0.75, Y: 0.25, Floor: "Ground", Shape: "Circle", Radius: 0.05, AreaUsages: []string{"Plant", "KillSample"}},
			"B_SITE":    {ID: "B_SITE", MapID: DefaultMapID, Name: "B Site", Site: "B", NodeType: "Site", DefaultSide: SideCT, X: 0.2, Y: 0.2, Floor: "Ground", Shape: "Circle", Radius: 0.05, AreaUsages: []string{"Plant", "KillSample"}},
		},
		Edges: map[string]MapEdge{
			"E1": {ID: "E1", FromNode: "T_SPAWN", ToNode: "LONG_DOOR", BaseTime: 8, Bidirectional: true},
			"E2": {ID: "E2", FromNode: "LONG_DOOR", ToNode: "A_LONG", BaseTime: 8, Bidirectional: true},
			"E3": {ID: "E3", FromNode: "A_LONG", ToNode: "A_SITE", BaseTime: 8, Bidirectional: true},
			"E4": {ID: "E4", FromNode: "CT_SPAWN", ToNode: "A_SITE", BaseTime: 6, Bidirectional: true},
		},
		Visibility: map[string]Visibility{},
		Routes: map[string]Route{
			"D2_A_LONG": {ID: "D2_A_LONG", Name: "A Long Execute", Side: SideT, TargetSite: "A", Nodes: []string{"T_SPAWN", "LONG_DOOR", "A_LONG", "A_SITE"}, MinPlayers: 2, MaxPlayers: 5},
			"D2_CT_A":   {ID: "D2_CT_A", Name: "CT A Hold", Side: SideCT, TargetSite: "A", Nodes: []string{"CT_SPAWN", "A_SITE"}, MinPlayers: 1, MaxPlayers: 5},
		},
		CombatConstants: consts,
	}
}

func TestSameSeedProducesSameMatch(t *testing.T) {
	input := makeTestInput(12345)
	res1, err := newProductionMatchEngine(input).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("first simulate failed: %v", err)
	}
	res2, err := newProductionMatchEngine(input).simulateMatch(context.Background())
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

func TestPlayerVisualFieldsDoNotChangeSimulation(t *testing.T) {
	plain := makeTestInput(24680)
	visual := makeTestInput(24680)
	for _, team := range []*TeamInput{&visual.TeamA, &visual.TeamB} {
		for playerIndex := range team.Players {
			team.Players[playerIndex].Portrait = "portraits/fallback.png"
			team.Players[playerIndex].CardImage = "player-cards/card.png"
			team.Players[playerIndex].AvatarCrop = &ImageCrop{X: 0.2, Y: 0.08, Width: 0.6, Height: 0.56}
		}
	}
	plainResult, err := newProductionMatchEngine(plain).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("plain simulate failed: %v", err)
	}
	visualResult, err := newProductionMatchEngine(visual).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("visual simulate failed: %v", err)
	}
	if plainResult.WinnerTeamID != visualResult.WinnerTeamID || plainResult.TotalRounds != visualResult.TotalRounds || len(plainResult.Rounds) != len(visualResult.Rounds) {
		t.Fatal("visual-only fields changed match summary")
	}
	for index := range plainResult.Rounds {
		left, right := plainResult.Rounds[index], visualResult.Rounds[index]
		if left.WinnerTeamID != right.WinnerTeamID || left.ScoreTeamA != right.ScoreTeamA || left.ScoreTeamB != right.ScoreTeamB || len(left.Events) != len(right.Events) {
			t.Fatalf("visual-only fields changed round %d", index+1)
		}
		for eventIndex := range left.Events {
			if left.Events[eventIndex].Timestamp != right.Events[eventIndex].Timestamp || left.Events[eventIndex].Message != right.Events[eventIndex].Message {
				t.Fatalf("visual-only fields changed event %d in round %d", eventIndex, index+1)
			}
		}
	}
}

func TestRoundTerminalStateMatchesWinReason(t *testing.T) {
	seenReasons := map[string]bool{}
	for seed := int64(1); seed <= 40; seed++ {
		res, err := newProductionMatchEngine(makeTestInput(seed)).simulateMatch(context.Background())
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
				if roundEndAt != int64(makeTestInput(seed).MapConfig.CombatConstants.Int("RoundTimeLimit", 0)) {
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
			case "bomb_secured":
				if round.Winner != SideT || ctAlive != 0 || !hasPlant || round.Bomb == nil || round.Bomb.Status != BombStatusPlanted {
					t.Fatalf("seed %d round %d: invalid secured state winner=%s alive=%d:%d plant=%v bomb=%+v", seed, round.RoundNumber, round.Winner, tAlive, ctAlive, hasPlant, round.Bomb)
				}
			default:
				t.Fatalf("seed %d round %d: unsupported win reason %q", seed, round.RoundNumber, round.WinReason)
			}
		}
	}

	if len(seenReasons) < 2 {
		t.Fatalf("seed sweep did not exercise distinct causal terminals: %+v", seenReasons)
	}
}

func TestSharedConfigPlayersRemainDistinctAndDeterministic(t *testing.T) {
	input := makeTestInput(424242)
	input.StartTime = 1_700_000_000_000
	input.TeamA.Players[0].ConfigPlayerID = "player_shared"
	input.TeamB.Players[0].ConfigPlayerID = "player_shared"

	first, err := newProductionMatchEngine(input).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("first simulation failed: %v", err)
	}
	second, err := newProductionMatchEngine(input).simulateMatch(context.Background())
	if err != nil {
		t.Fatalf("second simulation failed: %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first result: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second result: %v", err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("same instance identities and seed produced different reports")
	}

	sharedStates := map[string]*PlayerState{}
	for _, state := range first.Rounds[0].PlayerStates {
		if state.ConfigPlayerID == "player_shared" {
			sharedStates[state.PlayerID] = state
		}
	}
	if len(sharedStates) != 2 {
		t.Fatalf("expected two shared-config instances, got %+v", sharedStates)
	}
	sharedStats := map[string]*PlayerMatchStats{}
	for _, stats := range first.FinalStats.PlayerStats {
		if stats.ConfigPlayerID == "player_shared" {
			sharedStats[stats.PlayerID] = stats
		}
	}
	if len(sharedStats) != 2 {
		t.Fatalf("expected two independent shared-config stats, got %+v", sharedStats)
	}
}

func TestDuplicateMatchPlayerIDIsRejectedEvenWhenConfigIDsDiffer(t *testing.T) {
	input := makeTestInput(99)
	input.TeamB.Players[0].PlayerID = input.TeamA.Players[0].PlayerID
	input.TeamB.Players[0].ConfigPlayerID = "different_config_player"
	_, err := newProductionMatchEngine(input).simulateMatch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "duplicate player id") {
		t.Fatalf("expected duplicate instance id error, got %v", err)
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
