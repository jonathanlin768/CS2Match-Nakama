package match

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

func TestDebugSimuMatchUsesPlayersFromLubanTableAndSeed(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)

	res1, err := service.DebugSimuMatch(context.Background(), "user", DebugSimuMatchRequest{MapID: matchengine.DefaultMapID, Seed: 777})
	if err != nil {
		t.Fatalf("first simulation failed: %v", err)
	}
	res2, err := service.DebugSimuMatch(context.Background(), "user", DebugSimuMatchRequest{MapID: matchengine.DefaultMapID, Seed: 777})
	if err != nil {
		t.Fatalf("second simulation failed: %v", err)
	}
	if res1.MatchInfo.Seed != 777 || res2.MatchInfo.Seed != 777 {
		t.Fatalf("seed not preserved: %d/%d", res1.MatchInfo.Seed, res2.MatchInfo.Seed)
	}
	if !res1.DebugEnabled || !res2.DebugEnabled {
		t.Fatalf("battle report debug log should be enabled by CombatConst")
	}
	encoded, err := json.Marshal(res1)
	if err != nil {
		t.Fatalf("marshal response failed: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if enabled, ok := payload["debug_enabled"].(bool); !ok || !enabled {
		t.Fatalf("response JSON missing enabled debug flag: %s", encoded)
	}
	if _, ok := payload["match_info"]; !ok {
		t.Fatalf("embedded match report missing from response JSON: %s", encoded)
	}
	if res1.WinnerTeamID != res2.WinnerTeamID || res1.TotalRounds != res2.TotalRounds {
		t.Fatalf("same seed produced different summary")
	}
	if res1.MatchInfo.TeamAName != defaultTeamAName || res1.MatchInfo.TeamBName != defaultTeamBName {
		t.Fatalf("unexpected matchup: %s vs %s", res1.MatchInfo.TeamAName, res1.MatchInfo.TeamBName)
	}

	seen := map[string]bool{}
	for _, p := range res1.Rounds[0].PlayerStates {
		seen[p.PlayerID] = true
	}
	teamAPlayerIDs, teamBPlayerIDs, err := defaultTeamPlayerIDs()
	if err != nil {
		t.Fatalf("default lineups failed: %v", err)
	}
	for _, id := range append(teamAPlayerIDs, teamBPlayerIDs...) {
		if !seen[id] {
			t.Fatalf("default player %s missing from round snapshot", id)
		}
	}
	expectedFalcons := []string{"player_kyousuke", "player_kyxsan", "player_monesy", "player_niko", "player_teses"}
	expectedVitality := []string{"player_apex", "player_flamez", "player_zywoo", "player_mezii", "player_ropz"}
	if !reflect.DeepEqual(teamAPlayerIDs, expectedFalcons) {
		t.Fatalf("unexpected Falcons lineup: %v", teamAPlayerIDs)
	}
	if !reflect.DeepEqual(teamBPlayerIDs, expectedVitality) {
		t.Fatalf("unexpected Vitality lineup: %v", teamBPlayerIDs)
	}
	for _, state := range res1.Rounds[0].PlayerStates {
		if state.PlayerID == "player_niko" && state.Portrait != "portraits/player_niko.jpg" {
			t.Fatalf("portrait not copied from TbPlayer: %q", state.Portrait)
		}
	}
}

func TestDebugSimuMatchWeaponsFollowSideSwitch(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	res, err := service.DebugSimuMatch(context.Background(), "user", DebugSimuMatchRequest{MapID: matchengine.DefaultMapID, Seed: 888})
	if err != nil {
		t.Fatalf("simulation failed: %v", err)
	}
	if len(res.Rounds) < 13 {
		t.Fatalf("expected at least 13 rounds, got %d", len(res.Rounds))
	}
	round1 := res.Rounds[0]
	round13 := res.Rounds[12]
	if round1.TeamTID != "team_a" || round13.TeamCTID != "team_a" {
		t.Fatalf("unexpected side assignment: r1 T=%s, r13 CT=%s", round1.TeamTID, round13.TeamCTID)
	}
	if weaponForTeam(round1, "team_a") != matchengine.WeaponAK47 {
		t.Fatalf("team_a should use AK-47 on T side in round 1")
	}
	if weaponForTeam(round13, "team_a") != matchengine.WeaponM4A1S {
		t.Fatalf("team_a should use M4A1-S on CT side in round 13")
	}
}

func TestPlayerFromConfigMapsExplicitTenAttributes(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	row := cfg.GetPlayer("player_niko")
	if row == nil {
		t.Fatal("player_niko missing from generated config")
	}
	profile := playerFromConfig(row)
	got := profile.Attributes
	want := matchengine.PlayerAttributes{
		Entry: 68, Aim: 91, Trade: 75, Clutch: 68, Firepower: 86, Gamesense: 80,
		Reaction: 80, Positioning: 80, Awareness: 80, Teamplay: 75, Utility: 78,
		Composure: 68, Mobility: 80, Endurance: 77, Discipline: 80,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("player attributes were derived or dropped: got %+v want %+v", got, want)
	}
	if !reflect.DeepEqual(profile.RoleTags, []string{"AWPer", "IGL"}) {
		t.Fatalf("role tags not mapped: %v", profile.RoleTags)
	}
}

func TestBuildMapConfigMapsTemplateSideRoutesAndAllocations(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	mapConfig, err := buildMapConfigFromTables(matchengine.DefaultMapID)
	if err != nil {
		t.Fatalf("map config build failed: %v", err)
	}
	if mapConfig.Version != matchengine.DefaultMapVersion || len(mapConfig.Warnings) != 0 {
		t.Fatalf("formal Dust2 config should validate cleanly: version=%s warnings=%+v", mapConfig.Version, mapConfig.Warnings)
	}
	template := mapConfig.RouteTemplates["Default_Pick"]
	if template.Side != matchengine.SideT {
		t.Fatalf("template side not mapped: %q", template.Side)
	}
	if len(template.RouteIDs) != 4 || len(template.RouteAllocations) != 4 {
		t.Fatalf("template routes not mapped: ids=%v allocations=%v", template.RouteIDs, template.RouteAllocations)
	}
	total := 0
	for _, count := range template.RouteAllocations {
		total += count
	}
	if total != 5 {
		t.Fatalf("template allocations do not close to five: %v", template.RouteAllocations)
	}
	if !reflect.DeepEqual(template.CommonCTSetupIDs, []string{"CT_2A_1Mid_2B", "CT_3A_1Mid_1B", "CT_1A_1Mid_3B"}) {
		t.Fatalf("CT setup priors not mapped: %v", template.CommonCTSetupIDs)
	}
}

func TestCombatConstBool(t *testing.T) {
	constants := matchengine.CombatConstants{Values: map[string]matchengine.CombatConstValue{
		"Enabled":  {Value: "true"},
		"Disabled": {Value: "false"},
		"Invalid":  {Value: "not-a-bool"},
	}}
	if !combatConstBool(constants, "Enabled", false) {
		t.Fatal("true value should enable the switch")
	}
	if combatConstBool(constants, "Disabled", true) {
		t.Fatal("false value should disable the switch")
	}
	if combatConstBool(constants, "Invalid", false) || combatConstBool(constants, "Missing", false) {
		t.Fatal("invalid and missing values should use the safe disabled fallback")
	}
}

func TestDebugSimuMatchRequestRejectsRoundTruncationAndWinnerScripts(t *testing.T) {
	for _, payload := range []string{
		`{"map_id":"de_dust2","round_count":1}`,
		`{"map_id":"de_dust2","max_rounds":1}`,
		`{"map_id":"de_dust2","forced_round_winners":["team_a"]}`,
	} {
		var request DebugSimuMatchRequest
		if err := decodeDebugSimuMatchRequest(payload, &request); err == nil {
			t.Fatalf("request must reject non-contract control fields: %s", payload)
		}
	}
	var request DebugSimuMatchRequest
	if err := decodeDebugSimuMatchRequest(`{"map_id":"de_dust2","seed":42}`, &request); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
}

func weaponForTeam(round *matchengine.RoundResult, teamID string) string {
	for _, state := range round.PlayerStates {
		if state.TeamID == teamID {
			return state.Weapon.Primary
		}
	}
	return ""
}
