package matchengine

import (
	"reflect"
	"sort"
	"testing"
)

func TestStrategyScoreIsStableExplainedAndDecisiveNoiseCannotReverse(t *testing.T) {
	config := makeTestMapConfig()
	high := config.RouteTemplates["TPL_A"]
	high.ID, high.ScenarioIDs = "T_HIGH", []string{"SCN_HIGH"}
	low := high
	low.ID, low.ScenarioIDs = "T_LOW", []string{"SCN_LOW"}
	config.RouteTemplates = map[string]RouteTemplate{low.ID: low, high.ID: high, "TPL_CT": config.RouteTemplates["TPL_CT"]}
	config.Scenarios["SCN_HIGH"] = Scenario{ID: "SCN_HIGH", BaseWeight: 60}
	config.Scenarios["SCN_LOW"] = Scenario{ID: "SCN_LOW", BaseWeight: 5}
	setConstFloat(&config.CombatConstants, "DecisiveScoreGap", "18")
	setConstFloat(&config.CombatConstants, "MaxRandomNoise", "40")
	memory := StrategyMemory{
		PreviousSuccess: map[string]int{"T_HIGH": 2}, RecentTemplates: []string{"T_HIGH", "T_HIGH", "T_HIGH"},
		CounterReads: map[string]int{"T_HIGH": 1},
	}
	first, err := ScoreStrategyCandidates(config, makeTestTeam("t", "T", 82), SideT, 4, 8, memory, 9001, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ScoreStrategyCandidates(config, makeTestTeam("t", "T", 82), SideT, 4, 8, memory, 9001, 0)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("strategy scores are not reproducible: equal=%v err=%v", reflect.DeepEqual(first, second), err)
	}
	if first[0].TemplateID != "T_HIGH" || first[0].FinalScore <= first[1].FinalScore {
		t.Fatalf("decisive deterministic lead was reversed: %+v", first)
	}
	wantReasons := map[string]bool{"TEMPLATE_BASE": true, "LINEUP_FIT": true, "SCORE_PRESSURE": true, "PREVIOUS_SUCCESS": true, "REPEAT_PENALTY": true, "COUNTER_READ_RISK": true, "BOUNDED_RANDOM_NOISE": true}
	for _, reason := range first[0].Reasons {
		delete(wantReasons, reason.Code)
	}
	if len(wantReasons) != 0 {
		t.Fatalf("strategy explanation is incomplete: %v", wantReasons)
	}

	config.RouteTemplates = reverseTemplateMap(config.RouteTemplates)
	reordered, err := ScoreStrategyCandidates(config, makeTestTeam("t", "T", 82), SideT, 4, 8, memory, 9001, 0)
	if err != nil || !reflect.DeepEqual(first, reordered) {
		t.Fatalf("template map order changed scoring: equal=%v err=%v", reflect.DeepEqual(first, reordered), err)
	}
}

func TestStrategyNoiseUsesDisciplineIGLAndAwareness(t *testing.T) {
	config := makeTestMapConfig()
	team := makeTestTeam("t", "T", 80)
	team.Players[0].RoleTags = []string{"IGL"}
	stable := strategyNoiseAmplitude(team, config.CombatConstants)
	for index := range team.Players {
		team.Players[index].Attributes.Discipline = 0
		team.Players[index].Attributes.Awareness = 0
		team.Players[index].Attributes.Gamesense = 0
	}
	volatile := strategyNoiseAmplitude(team, config.CombatConstants)
	if stable >= volatile {
		t.Fatalf("discipline/IGL/awareness did not jointly constrain noise: stable=%v volatile=%v", stable, volatile)
	}
}

func TestRoleAssignmentBombCarrierAndGapReasonsAreStable(t *testing.T) {
	team := makeTestTeam("team_a", "A", 85)
	tags := [][]string{{"Entry"}, {"Support"}, {"IGL"}, {"Lurker"}, {"AWPer"}}
	for index := range team.Players {
		team.Players[index].RoleTags = tags[index]
	}
	template := makeTestMapConfig().RouteTemplates["TPL_A"]
	template.RequiredRoles = []string{"Entry", "Support", "IGL", "Lurker", "AWPer"}
	roles := AssignRoles(team, template)
	if len(roles.Assignments) != 5 || len(roles.GapRoles) != 0 || roles.Penalty != 0 {
		t.Fatalf("exact role assignment failed: %+v", roles)
	}
	config := makeTestMapConfig()
	routes := map[string]string{}
	for _, profile := range team.Players {
		routes[profile.PlayerID] = "D2_A_LONG"
	}
	carrier, err := SelectBombCarrier(team, roles.Assignments, routes, config)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range roles.Assignments {
		if assignment.PlayerID == carrier && assignment.Role == "Entry" {
			t.Fatal("Entry was selected as carrier despite reachable non-Entry options")
		}
	}
	template.RequiredRoles = []string{"Coach"}
	gap := AssignRoles(team, template)
	if len(gap.GapRoles) != 1 || gap.GapRoles[0] != "Coach" || gap.Penalty == 0 || len(gap.Reasons) == 0 || gap.Reasons[0].Code != "ROLE_GAP" {
		t.Fatalf("role gap was not explained: %+v", gap)
	}
}

func TestOpeningPlanDeploysTenLegalActionsAtTimelineZero(t *testing.T) {
	input := makeTestRoundInput(9002)
	tTemplate := input.MapConfig.RouteTemplates["TPL_A"]
	ctTemplate := input.MapConfig.RouteTemplates["TPL_CT"]
	plan, _, err := BuildRoundPlan(input, tTemplate, ctTemplate)
	if err != nil {
		t.Fatal(err)
	}
	state, err := NewRoundState(input, plan)
	if err != nil {
		t.Fatal(err)
	}
	actions, err := DeployOpeningActions(state)
	if err != nil {
		t.Fatal(err)
	}
	if state.Timeline != 0 || len(actions) != 10 {
		t.Fatalf("opening deployment consumed timer or missed players: timeline=%d actions=%d", state.Timeline, len(actions))
	}
	seen := map[string]bool{}
	for _, action := range actions {
		if len(action.ActorIDs) != 1 || action.StartAt != 0 || action.ID == "" {
			t.Fatalf("invalid opening action: %+v", action)
		}
		seen[action.ActorIDs[0]] = true
	}
	for playerID, player := range state.Players {
		if !seen[playerID] || player.Intent.ID == "" || player.Action.CurrentActionID == "" {
			t.Fatalf("player %s has no real opening intent/action: %+v", playerID, player)
		}
	}
}

func TestConfiguredTemplateFamiliesReachableAndCTSelectionIndependent(t *testing.T) {
	config := makeTestMapConfig()
	baseT, baseCT := config.RouteTemplates["TPL_A"], config.RouteTemplates["TPL_CT"]
	for _, id := range []string{"A_Long_Rush", "A_Short_Split", "B_Tunnel_Explode", "Mid_To_B", "Default_Pick", "Fake_A_Go_B"} {
		template := baseT
		template.ID = id
		config.RouteTemplates[id] = template
	}
	for _, id := range []string{"CT_2A_1Mid_2B", "CT_1A_2Mid_2B", "CT_2A_2Mid_1B"} {
		template := baseCT
		template.ID = id
		config.RouteTemplates[id] = template
	}
	for _, template := range config.RouteTemplates {
		for routeID := range template.RouteAllocations {
			route := config.Routes[routeID]
			spawn := "T_SPAWN"
			if template.Side == SideCT {
				spawn = "CT_SPAWN"
			}
			if _, feedback, err := FindBoundedPath(config, spawn, route.Nodes[len(route.Nodes)-1], 100); err != nil || feedback != nil {
				t.Fatalf("template %s route %s unreachable: %+v/%v", template.ID, routeID, feedback, err)
			}
		}
	}
	ctTeam := makeTestTeam("ct", "CT", 80)
	before, err := SelectCTSetup(config, ctTeam, 2, 4, StrategyMemory{}, 777, 0)
	if err != nil {
		t.Fatal(err)
	}
	tTemplate := config.RouteTemplates["A_Long_Rush"]
	tTemplate.TargetSite = "SECRET_CHANGED_T_PLAN"
	config.RouteTemplates["A_Long_Rush"] = tTemplate
	after, err := SelectCTSetup(config, ctTeam, 2, 4, StrategyMemory{}, 777, 0)
	if err != nil || before.Template.ID != after.Template.ID || !reflect.DeepEqual(before.Score, after.Score) {
		t.Fatalf("CT setup read current hidden T plan: before=%+v after=%+v err=%v", before, after, err)
	}
	if deriveSeed(777, "opening_plan", 0) == deriveSeed(777, "ct_setup", 0) || deriveSeed(777, "ct_setup", 0) == deriveSeed(777, "ct_setup", 1) {
		t.Fatal("T/CT attempts do not have independent stable seed identities")
	}
}

func TestSideSwitchMemoryDecayPreservesTeamStyle(t *testing.T) {
	memory := StrategyMemory{
		PreviousSuccess: map[string]int{"TPL_A": 5}, CounterReads: map[string]int{"TPL_A": 3}, SideTendency: map[string]float64{"A": 8},
		TeamStyle: map[string]float64{"aggression": 0.7}, RecentTemplates: []string{"TPL_A"},
	}
	decayed := DecayStrategyMemoryForSideSwitch(memory)
	if decayed.PreviousSuccess["TPL_A"] != 2 || decayed.CounterReads["TPL_A"] != 1 || decayed.SideTendency["A"] != 4 || decayed.TeamStyle["aggression"] != 0.7 || len(decayed.RecentTemplates) != 0 {
		t.Fatalf("side-switch memory boundary mismatch: %+v", decayed)
	}
}

func TestCTOpeningFallsBackOnlyToConfiguredCTDefault(t *testing.T) {
	input := makeTestInput(9003)
	bad := input.MapConfig.RouteTemplates["TPL_CT"]
	bad.ID = "CT_BAD"
	bad.RouteIDs = []string{"D2_A_LONG"}
	bad.RouteAllocations = map[string]int{"D2_A_LONG": 5}
	input.MapConfig.RouteTemplates[bad.ID] = bad
	roundSeed := deriveSeed(input.Seed, input.MapVersion, input.RuleSet.RuleSetID, 1)
	selection, err := newProductionMatchEngine(input).planOpening(roundSeed, SideCT)
	if err != nil {
		t.Fatal(err)
	}
	if !selection.UsedDefault || selection.Template.ID != "TPL_CT" || selection.Template.Side != SideCT {
		t.Fatalf("CT did not recover with its own configured default: %+v", selection)
	}
}

func reverseTemplateMap(source map[string]RouteTemplate) map[string]RouteTemplate {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := map[string]RouteTemplate{}
	for index := len(keys) - 1; index >= 0; index-- {
		result[keys[index]] = source[keys[index]]
	}
	return result
}

func setConstFloat(constants *CombatConstants, key, value string) {
	entry := constants.Values[key]
	entry.ValueType = "Float"
	entry.Value = value
	constants.Values[key] = entry
}
