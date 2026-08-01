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

func weaponForTeam(round *matchengine.RoundResult, teamID string) string {
	for _, state := range round.PlayerStates {
		if state.TeamID == teamID {
			return state.Weapon.Primary
		}
	}
	return ""
}
