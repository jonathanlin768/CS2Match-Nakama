package match

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
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
		teamID := "team_a"
		if slices.Contains(teamBPlayerIDs, id) {
			teamID = "team_b"
		}
		instanceID, instanceErr := matchPlayerID(teamID, id)
		if instanceErr != nil {
			t.Fatalf("build player instance id: %v", instanceErr)
		}
		if !seen[instanceID] {
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
		if state.ConfigPlayerID == "player_niko" && state.Portrait != "portraits/player_niko.jpg" {
			t.Fatalf("portrait not copied from TbPlayer: %q", state.Portrait)
		}
		if state.ConfigPlayerID == "player_niko" {
			if state.CardImage != "player-cards/niko2.png" || state.AvatarCrop == nil || !state.AvatarCrop.Valid() {
				t.Fatalf("card crop not copied from TbPlayer: card=%q crop=%+v", state.CardImage, state.AvatarCrop)
			}
		}
	}
}

func TestPlayerFromConfigRejectsInvalidAvatarCrop(t *testing.T) {
	profile, err := playerFromConfig("team/a", &cfg.Player{Id: "player/p", Name: "P", Portrait: "portraits/p.png", CardImage: "player-cards/p.png", AvatarCropX: 0.8, AvatarCropWidth: 0.6, AvatarCropHeight: 0.2})
	if err != nil {
		t.Fatalf("player mapping failed: %v", err)
	}
	if profile.CardImage != "player-cards/p.png" || profile.AvatarCrop != nil || profile.Portrait != "portraits/p.png" {
		t.Fatalf("invalid crop should preserve source paths and omit crop: %+v", profile)
	}
	if profile.PlayerID != "team%2Fa/player%2Fp" || profile.ConfigPlayerID != "player/p" {
		t.Fatalf("player identities were not escaped and preserved: %+v", profile)
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

func TestSimuMatchComputerUsesConfiguredTeams(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	res, err := service.SimuMatch(context.Background(), "user", SimuMatchRequest{Mode: "computer"})
	if err != nil {
		t.Fatalf("computer simulation failed: %v", err)
	}
	if res.MatchInfo.TeamAID != defaultTeamAID || res.MatchInfo.TeamBID != defaultTeamBID {
		t.Fatalf("unexpected configured teams: %s vs %s", res.MatchInfo.TeamAID, res.MatchInfo.TeamBID)
	}
	if res.MatchInfo.TeamAName != defaultTeamAName || res.MatchInfo.TeamBName != defaultTeamBName {
		t.Fatalf("unexpected team names: %s vs %s", res.MatchInfo.TeamAName, res.MatchInfo.TeamBName)
	}
	if !res.DebugEnabled {
		t.Fatal("production SimuMatch response lost the configured battle report debug flag")
	}
}

func TestSimuMatchTutorialValidatesAuthoritativeRoster(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	valid := []string{"player_donk", "player_sh1ro", "player_zont1x", "player_magixx", "player_chopper"}
	res, err := service.SimuMatch(context.Background(), "guest-user", SimuMatchRequest{Mode: "tutorial", TutorialConfigID: "tutorial_default", ConfigVersion: 1, PlayerIDs: valid})
	if err != nil {
		t.Fatalf("tutorial simulation failed: %v", err)
	}
	if res.MatchInfo.TeamAID != "tutorial_players" || res.MatchInfo.TeamBID != "team_vitality" {
		t.Fatalf("unexpected tutorial matchup: %s vs %s", res.MatchInfo.TeamAID, res.MatchInfo.TeamBID)
	}
	seen := map[string]bool{}
	for _, state := range res.Rounds[0].PlayerStates {
		seen[state.PlayerID] = true
		if state.ConfigPlayerID == "" {
			t.Fatalf("state %s lost config player id", state.PlayerID)
		}
	}
	for _, id := range valid {
		instanceID, instanceErr := matchPlayerID("tutorial_players", id)
		if instanceErr != nil {
			t.Fatalf("build player instance id: %v", instanceErr)
		}
		if !seen[instanceID] {
			t.Fatalf("selected player %s missing from match input", id)
		}
	}

	cases := []struct {
		name    string
		request SimuMatchRequest
		code    string
	}{
		{"version", SimuMatchRequest{Mode: "tutorial", TutorialConfigID: "tutorial_default", ConfigVersion: 2, PlayerIDs: valid}, "CONFIG_VERSION_MISMATCH"},
		{"duplicate", SimuMatchRequest{Mode: "tutorial", TutorialConfigID: "tutorial_default", ConfigVersion: 1, PlayerIDs: []string{"player_donk", "player_donk", "player_zont1x", "player_magixx", "player_chopper"}}, "INVALID_LINEUP"},
		{"budget", SimuMatchRequest{Mode: "tutorial", TutorialConfigID: "tutorial_default", ConfigVersion: 1, PlayerIDs: []string{"player_donk", "player_monesy", "player_b1t", "player_sh1ro", "player_niko"}}, "BUDGET_EXCEEDED"},
		{"unknown", SimuMatchRequest{Mode: "tutorial", TutorialConfigID: "missing", ConfigVersion: 1, PlayerIDs: valid}, "INVALID_TUTORIAL_CONFIG"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, gotErr := service.SimuMatch(context.Background(), "guest-user", tc.request)
			typed, ok := gotErr.(*MatchError)
			if !ok || typed.Code != tc.code {
				t.Fatalf("expected %s, got %v", tc.code, gotErr)
			}
		})
	}
}

func TestValidateTutorialLineupAllowsOpponentOverlap(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	tutorial := &cfg.TutorialBattle{
		RosterSize:        5,
		Budget:            15,
		Tier5PlayerIds:    []string{"player_donk"},
		Tier4PlayerIds:    []string{"player_sh1ro"},
		Tier3PlayerIds:    []string{"player_zont1x"},
		Tier2PlayerIds:    []string{"player_magixx"},
		Tier1PlayerIds:    []string{"player_chopper"},
		OpponentPlayerIds: []string{"player_donk"},
	}
	err := validateTutorialLineup(tutorial, []string{"player_donk", "player_sh1ro", "player_zont1x", "player_magixx", "player_chopper"})
	if err != nil {
		t.Fatalf("opponent overlap should be allowed, got %v", err)
	}
}

func TestBuildTeamsNamespacesSharedConfigPlayer(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	teamA, err := service.buildNamedTeam("tutorial_players", "Tutorial", []string{"player_zywoo"})
	if err != nil {
		t.Fatalf("build team A: %v", err)
	}
	teamB, err := service.buildNamedTeam("team_vitality", "Vitality", []string{"player_zywoo"})
	if err != nil {
		t.Fatalf("build team B: %v", err)
	}
	left, right := teamA.Players[0], teamB.Players[0]
	if left.PlayerID == right.PlayerID || left.ConfigPlayerID != "player_zywoo" || right.ConfigPlayerID != "player_zywoo" {
		t.Fatalf("shared config player was not split into unique instances: left=%+v right=%+v", left, right)
	}
	if left.Attributes != right.Attributes || left.CardImage != right.CardImage || !reflect.DeepEqual(left.AvatarCrop, right.AvatarCrop) {
		t.Fatalf("shared config snapshots differ: left=%+v right=%+v", left, right)
	}
}

func TestSimulateSharedConfigPlayerProducesTenDistinctActors(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	teamA, err := service.buildNamedTeam("tutorial_players", "Tutorial", []string{"player_zywoo", "player_donk", "player_sh1ro", "player_zont1x", "player_magixx"})
	if err != nil {
		t.Fatalf("build team A: %v", err)
	}
	teamB, err := service.buildConfigTeam("team_vitality", []string{"player_zywoo", "player_apex", "player_flamez", "player_mezii", "player_ropz"})
	if err != nil {
		t.Fatalf("build team B: %v", err)
	}
	result, err := service.simulate(context.Background(), "shared_player_test", matchengine.DefaultMapID, 20260813, teamA, teamB)
	if err != nil {
		t.Fatalf("simulate shared player match: %v", err)
	}
	if len(result.Rounds) == 0 || result.FinalStats == nil || len(result.FinalStats.PlayerStats) != 10 {
		t.Fatalf("incomplete match result: rounds=%d final=%+v", len(result.Rounds), result.FinalStats)
	}
	actors := map[string]string{}
	shared := map[string]bool{}
	for _, state := range result.Rounds[0].PlayerStates {
		if _, duplicate := actors[state.PlayerID]; duplicate {
			t.Fatalf("duplicate actor state: %s", state.PlayerID)
		}
		actors[state.PlayerID] = state.TeamID
		if state.ConfigPlayerID == "player_zywoo" {
			shared[state.PlayerID] = true
		}
	}
	if len(actors) != 10 || len(shared) != 2 {
		t.Fatalf("expected 10 actors and two ZyWOo instances, actors=%v shared=%v", actors, shared)
	}
	for _, round := range result.Rounds {
		if round.Bomb != nil && round.Bomb.CarrierID != "" {
			if _, ok := actors[round.Bomb.CarrierID]; !ok {
				t.Fatalf("bomb carrier does not reference an instance id: %s", round.Bomb.CarrierID)
			}
		}
		for _, event := range round.Events {
			for _, actorID := range []string{event.AttackerID, event.VictimID} {
				if actorID != "" {
					if _, ok := actors[actorID]; !ok {
						t.Fatalf("event %s references non-instance actor %s", event.EventType, actorID)
					}
				}
			}
		}
	}
}

func TestBuildTeamRejectsDuplicateConfigPlayer(t *testing.T) {
	if err := cfg.Init(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	service := NewService(matchengine.NewService(nil), nil)
	_, err := service.buildNamedTeam("tutorial_players", "Tutorial", []string{"player_zywoo", "player_zywoo"})
	typed, ok := err.(*MatchError)
	if !ok || typed.Code != "INVALID_LINEUP" {
		t.Fatalf("expected duplicate config player error, got %v", err)
	}
}

func TestMatchPlayerIDRejectsEmptyComponents(t *testing.T) {
	for _, tc := range [][2]string{{"", "player"}, {"team", ""}, {" ", "player"}} {
		if _, err := matchPlayerID(tc[0], tc[1]); err == nil {
			t.Fatalf("expected empty component error for %#v", tc)
		}
	}
}

func TestValidateMatchTeamIdentitiesAllowsSharedConfigPlayer(t *testing.T) {
	left := matchengine.TeamInput{TeamID: "tutorial_players", Players: []matchengine.PlayerProfile{{
		PlayerID: "tutorial_players/player_zywoo", ConfigPlayerID: "player_zywoo",
	}}}
	right := matchengine.TeamInput{TeamID: "team_vitality", Players: []matchengine.PlayerProfile{{
		PlayerID: "team_vitality/player_zywoo", ConfigPlayerID: "player_zywoo",
	}}}
	if err := validateMatchTeamIdentities(left, right); err != nil {
		t.Fatalf("shared config player should have two valid match identities: %v", err)
	}
}

func TestValidateMatchTeamIdentitiesRejectsBoundaryViolations(t *testing.T) {
	validPlayer := matchengine.PlayerProfile{PlayerID: "team_a/player_one", ConfigPlayerID: "player_one"}
	for name, teams := range map[string][2]matchengine.TeamInput{
		"same team id": {
			{TeamID: "team_a", Players: []matchengine.PlayerProfile{validPlayer}},
			{TeamID: "team_a", Players: []matchengine.PlayerProfile{validPlayer}},
		},
		"forged instance id": {
			{TeamID: "team_a", Players: []matchengine.PlayerProfile{validPlayer}},
			{TeamID: "team_b", Players: []matchengine.PlayerProfile{{PlayerID: "team_a/player_one", ConfigPlayerID: "player_one"}}},
		},
		"empty config id": {
			{TeamID: "team_a", Players: []matchengine.PlayerProfile{validPlayer}},
			{TeamID: "team_b", Players: []matchengine.PlayerProfile{{PlayerID: "team_b/player_two"}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateMatchTeamIdentities(teams[0], teams[1]); err == nil {
				t.Fatal("expected identity validation failure")
			}
		})
	}
}

func TestPerTeamDuplicateValidationAllowsCrossTeamOverlap(t *testing.T) {
	if err := validateUniqueConfigPlayerIDs("team_a", []string{"player_zywoo"}); err != nil {
		t.Fatalf("team A validation failed: %v", err)
	}
	if err := validateUniqueConfigPlayerIDs("team_b", []string{"player_zywoo"}); err != nil {
		t.Fatalf("same config id on team B should be allowed: %v", err)
	}
	if err := validateUniqueConfigPlayerIDs("team_a", []string{"player_zywoo", "player_zywoo"}); err == nil {
		t.Fatal("same-team duplicate should still fail")
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
	profile, err := playerFromConfig(defaultTeamAID, row)
	if err != nil {
		t.Fatalf("player mapping failed: %v", err)
	}
	got := profile.Attributes
	want := matchengine.PlayerAttributes{
		Entry: 68, Aim: 91, Trade: 75, Clutch: 68, Firepower: 86, Gamesense: 80,
		Reaction: 80, Positioning: 80, Awareness: 80, Teamplay: 75, Utility: 78,
		Composure: 68, Mobility: 80, Endurance: 77, Discipline: 80,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("player attributes were derived or dropped: got %+v want %+v", got, want)
	}
	if !reflect.DeepEqual(profile.RoleTags, row.Positions) {
		t.Fatalf("role tags not mapped: got %v want %v", profile.RoleTags, row.Positions)
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
