package match

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

func TestLubanCalibrationShortSample(t *testing.T) {
	input := lubanCalibrationInput(t, 18001)
	if input.MapConfig.CombatConstants.Int("MaxEncounterPulses", 0) != 5 || input.MapConfig.CombatConstants.Float("CombatScale", 0) != 50 || input.MapConfig.CombatConstants.Int("RoundTimeLimit", 0) != 90 || input.MapConfig.Edges["EDGE_T_LONG"].BaseTime != 4 {
		t.Fatalf("calibrated Luban values were not loaded: raw_pulses=%+v raw_edge=%+v pulses=%d scale=%f edge=%d", cfg.Global.TbCombatConst.Get("MaxEncounterPulses"), cfg.Global.TbMapEdge.Get("EDGE_CT_A"), input.MapConfig.CombatConstants.Int("MaxEncounterPulses", 0), input.MapConfig.CombatConstants.Float("CombatScale", 0), input.MapConfig.Edges["EDGE_CT_A"].BaseTime)
	}
	summary, err := matchengine.CalibrateRounds(context.Background(), input, 64)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Samples != 64 || summary.AverageRoundDurationSeconds <= 0 {
		t.Fatalf("invalid short calibration: %+v", summary)
	}
	t.Log(calibrationJSON(t, summary))
}

func TestLubanCalibrationLongSample(t *testing.T) {
	if os.Getenv("SIMU_LONG_CALIBRATION") != "1" {
		t.Skip("set SIMU_LONG_CALIBRATION=1 to run the explicit 10k production-config calibration")
	}
	summary, err := matchengine.CalibrateRounds(context.Background(), lubanCalibrationInput(t, 19001), 10000)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Samples != 10000 {
		t.Fatalf("long calibration did not finish: %+v", summary)
	}
	t.Log("baseline:", calibrationJSON(t, summary))
	strongWeakInput := lubanCalibrationInput(t, 29001)
	strongWeakInput.TeamT = adjustCalibrationTeam(strongWeakInput.TeamT, 15)
	strongWeakInput.TeamCT = adjustCalibrationTeam(strongWeakInput.TeamCT, -15)
	strongWeak, err := matchengine.CalibrateRounds(context.Background(), strongWeakInput, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if strongWeak.StrongTeamOpportunities != 2000 {
		t.Fatalf("controlled strong/weak calibration was not recognized: %+v", strongWeak)
	}
	t.Log("controlled strong-vs-weak:", calibrationJSON(t, strongWeak))
}

func adjustCalibrationTeam(team matchengine.TeamInput, delta int) matchengine.TeamInput {
	players := append([]matchengine.PlayerProfile(nil), team.Players...)
	for index := range players {
		a := &players[index].Attributes
		a.Entry = clampCalibrationAttribute(a.Entry + delta)
		a.Aim = clampCalibrationAttribute(a.Aim + delta)
		a.Trade = clampCalibrationAttribute(a.Trade + delta)
		a.Clutch = clampCalibrationAttribute(a.Clutch + delta)
		a.Firepower = clampCalibrationAttribute(a.Firepower + delta)
		a.Gamesense = clampCalibrationAttribute(a.Gamesense + delta)
		a.Reaction = clampCalibrationAttribute(a.Reaction + delta)
		a.Positioning = clampCalibrationAttribute(a.Positioning + delta)
		a.Awareness = clampCalibrationAttribute(a.Awareness + delta)
		a.Teamplay = clampCalibrationAttribute(a.Teamplay + delta)
		a.Utility = clampCalibrationAttribute(a.Utility + delta)
		a.Composure = clampCalibrationAttribute(a.Composure + delta)
		a.Mobility = clampCalibrationAttribute(a.Mobility + delta)
		a.Endurance = clampCalibrationAttribute(a.Endurance + delta)
		a.Discipline = clampCalibrationAttribute(a.Discipline + delta)
	}
	team.Players = players
	return team
}

func clampCalibrationAttribute(value int) int {
	return min(100, max(0, value))
}

func TestLubanCalibrationCandidates(t *testing.T) {
	if os.Getenv("SIMU_CALIBRATION_CANDIDATES") != "1" {
		t.Skip("diagnostic candidate sweep")
	}
	tests := []struct {
		name                string
		plantTime, minPlant int
		ctA, ctMid, ctB     int
		boost               string
	}{
		{name: "plant2-current", plantTime: 2, minPlant: 2, ctA: 10, ctMid: 8, ctB: 11},
		{name: "plant3", plantTime: 3, minPlant: 3, ctA: 10, ctMid: 8, ctB: 11},
		{name: "plant4", plantTime: 4, minPlant: 3, ctA: 10, ctMid: 8, ctB: 11},
		{name: "plant4-ct-faster", plantTime: 4, minPlant: 3, ctA: 9, ctMid: 7, ctB: 10, boost: "maximum"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := lubanCalibrationInput(t, 20001)
			setCalibrationConstant(input, "BasePlantTime", "Int", fmt.Sprint(test.plantTime))
			setCalibrationConstant(input, "MinPlantTime", "Int", fmt.Sprint(test.minPlant))
			setCalibrationEdge(input, "EDGE_CT_A", test.ctA)
			setCalibrationEdge(input, "EDGE_CT_MID", test.ctMid)
			setCalibrationEdge(input, "EDGE_CT_B_DOOR", test.ctB)
			applyCalibrationBoost(input, test.boost)
			summary, err := matchengine.CalibrateRounds(context.Background(), input, 1024)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(calibrationJSON(t, summary))
		})
	}
	for _, scale := range []string{"12", "32", "50"} {
		t.Run("strong-weak-scale-"+scale, func(t *testing.T) {
			input := lubanCalibrationInput(t, 24001)
			setCalibrationConstant(input, "CombatScale", "Float", scale)
			input.TeamT = adjustCalibrationTeam(input.TeamT, 16)
			input.TeamCT = adjustCalibrationTeam(input.TeamCT, -16)
			summary, err := matchengine.CalibrateRounds(context.Background(), input, 1024)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(calibrationJSON(t, summary))
		})
	}
	t.Run("baseline-scale-50", func(t *testing.T) {
		input := lubanCalibrationInput(t, 25001)
		setCalibrationConstant(input, "CombatScale", "Float", "50")
		summary, err := matchengine.CalibrateRounds(context.Background(), input, 1024)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(calibrationJSON(t, summary))
	})
	for _, variant := range []struct {
		name                string
		tEdgeAdd, ctEdgeSub int
	}{
		{name: "scale50-t-plus1", tEdgeAdd: 1},
		{name: "scale50-t-plus2", tEdgeAdd: 2},
		{name: "scale50-t-plus1-ct-minus1", tEdgeAdd: 1, ctEdgeSub: 1},
	} {
		t.Run(variant.name, func(t *testing.T) {
			input := lubanCalibrationInput(t, 26001)
			setCalibrationConstant(input, "CombatScale", "Float", "50")
			for _, edgeID := range []string{"EDGE_T_LONG", "EDGE_LONG_A", "EDGE_A_LONG_SITE", "EDGE_T_MID", "EDGE_MID_CAT", "EDGE_CAT_SHORT", "EDGE_SHORT_A", "EDGE_T_B_UPPER", "EDGE_B_UPPER_TUNNEL", "EDGE_TUNNEL_B"} {
				edge := input.MapConfig.Edges[edgeID]
				setCalibrationEdge(input, edgeID, edge.BaseTime+variant.tEdgeAdd)
			}
			for _, edgeID := range []string{"EDGE_CT_A", "EDGE_CT_MID", "EDGE_CT_B_DOOR"} {
				edge := input.MapConfig.Edges[edgeID]
				setCalibrationEdge(input, edgeID, edge.BaseTime-variant.ctEdgeSub)
			}
			summary, err := matchengine.CalibrateRounds(context.Background(), input, 1024)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(calibrationJSON(t, summary))
		})
	}
	for _, controlledStrength := range []int{0, 16, 20} {
		name := "scale50-t-plus1-bomb28"
		if controlledStrength > 0 {
			name += fmt.Sprintf("-strong%d", controlledStrength)
		}
		t.Run(name, func(t *testing.T) {
			input := lubanCalibrationInput(t, 27001)
			setCalibrationConstant(input, "CombatScale", "Float", "50")
			setCalibrationConstant(input, "BombExplodeTime", "Int", "28")
			for _, edgeID := range []string{"EDGE_T_LONG", "EDGE_LONG_A", "EDGE_A_LONG_SITE", "EDGE_T_MID", "EDGE_MID_CAT", "EDGE_CAT_SHORT", "EDGE_SHORT_A", "EDGE_T_B_UPPER", "EDGE_B_UPPER_TUNNEL", "EDGE_TUNNEL_B"} {
				edge := input.MapConfig.Edges[edgeID]
				setCalibrationEdge(input, edgeID, edge.BaseTime+1)
			}
			if controlledStrength > 0 {
				input.TeamT = adjustCalibrationTeam(input.TeamT, controlledStrength)
				input.TeamCT = adjustCalibrationTeam(input.TeamCT, -controlledStrength)
			}
			summary, err := matchengine.CalibrateRounds(context.Background(), input, 1024)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(calibrationJSON(t, summary))
		})
	}
}

func applyCalibrationBoost(input *matchengine.RoundInput, level string) {
	fast, execute, ctHold, tModifier := 0, 0, 0, 0
	switch level {
	case "moderate":
		fast, execute, ctHold, tModifier = 15, 18, 8, 15
	case "strong":
		fast, execute, ctHold, tModifier = 22, 25, 5, 20
	case "maximum":
		fast, execute, ctHold, tModifier = 25, 25, 0, 25
	default:
		return
	}
	for id, weight := range map[string]int{"D2_FAST_TIMING": fast, "D2_T_EXECUTE": execute, "D2_CT_HOLD_ANGLE": ctHold} {
		tag := input.MapConfig.MapTags[id]
		tag.Weight = weight
		input.MapConfig.MapTags[id] = tag
	}
	for id, modifier := range input.MapConfig.EncounterModifiers {
		if modifier.Factor == "ScenarioWeight" {
			continue
		}
		if modifier.Side == matchengine.SideT && modifier.Weight > 0 {
			modifier.Weight = tModifier
		}
		if modifier.Side == matchengine.SideCT && modifier.Factor == "CTHoldAngle" {
			modifier.Weight = ctHold
		}
		input.MapConfig.EncounterModifiers[id] = modifier
	}
}

func setCalibrationEdge(input *matchengine.RoundInput, id string, baseTime int) {
	edge := input.MapConfig.Edges[id]
	edge.BaseTime = baseTime
	input.MapConfig.Edges[id] = edge
}

func setCalibrationConstant(input *matchengine.RoundInput, key, valueType, value string) {
	entry := input.MapConfig.CombatConstants.Values[key]
	entry.ValueType, entry.Value = valueType, value
	input.MapConfig.CombatConstants.Values[key] = entry
}

func lubanCalibrationInput(t *testing.T, seed int64) *matchengine.RoundInput {
	t.Helper()
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	mapConfig, err := buildMapConfigFromTables(matchengine.DefaultMapID)
	if err != nil {
		t.Fatalf("map config build failed: %v", err)
	}
	teamAPlayerIDs, teamBPlayerIDs, err := defaultTeamPlayerIDs()
	if err != nil {
		t.Fatal(err)
	}
	service := &Service{}
	teamA, err := service.buildTeam("team_a", defaultTeamAName, teamAPlayerIDs)
	if err != nil {
		t.Fatal(err)
	}
	teamB, err := service.buildTeam("team_b", defaultTeamBName, teamBPlayerIDs)
	if err != nil {
		t.Fatal(err)
	}
	return &matchengine.RoundInput{
		MatchID: "calibration", RoundNumber: 1, MapID: mapConfig.MapID, MapVersion: mapConfig.Version, Seed: seed,
		RuleSet: matchengine.DefaultMR12RuleSet(mapConfig.CombatConstants), MapConfig: mapConfig,
		WeaponSpecs: defaultWeaponSpecs(), SideLoadouts: defaultSideLoadouts(), TeamT: teamA, TeamCT: teamB,
		TeamAID: teamA.TeamID, TeamBID: teamB.TeamID, ScoreByTeam: map[string]int{teamA.TeamID: 0, teamB.TeamID: 0},
	}
}

func calibrationJSON(t *testing.T, summary *matchengine.CalibrationSummary) string {
	t.Helper()
	payload, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
