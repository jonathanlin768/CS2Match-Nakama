package matchengine

import "testing"

func completedMemoryRound() *RoundResult {
	return &RoundResult{
		RoundNumber: 1, TeamTID: "team_a", TeamCTID: "team_b", WinnerTeamID: "team_a", Winner: SideT, WinReason: WinReasonElimination,
		StrategyTemplateID: "TPL_A", CTSetupTemplateID: "TPL_CT", Bomb: &BombPublicState{Site: "A"},
		Events: []*GameEvent{{EventType: EventRoundEnd}},
	}
}

func TestStrategyMemoryUpdatesOnlyFromCompletedRoundAndDecaysAtSideSwitch(t *testing.T) {
	memory := newStrategyMemory()
	if _, err := UpdateStrategyMemoryFromRound(memory, &RoundResult{StrategyTemplateID: "TPL_A"}, "team_a", 3); err == nil {
		t.Fatal("incomplete result entered strategy memory")
	}
	updated, err := UpdateStrategyMemoryFromRound(memory, completedMemoryRound(), "team_a", 3)
	if err != nil {
		t.Fatal(err)
	}
	if updated.PreviousSuccess["TPL_A"] != 1 || updated.CounterReads["TPL_A"] != 1 || len(updated.RecentTemplates) != 1 || updated.SideTendency["site:A"] != 1 {
		t.Fatalf("completed result did not update memory: %+v", updated)
	}
	decayed := DecayStrategyMemoryForSideSwitch(StrategyMemory{PreviousSuccess: map[string]int{"TPL_A": 5}, CounterReads: map[string]int{"TPL_A": 3}, SideTendency: map[string]float64{"site:A": 3}, TeamStyle: map[string]float64{"aggression": 2}})
	if decayed.PreviousSuccess["TPL_A"] != 2 || decayed.CounterReads["TPL_A"] != 1 || decayed.SideTendency["site:A"] != 1.5 || decayed.TeamStyle["aggression"] != 2 {
		t.Fatalf("side-switch decay boundary mismatch: %+v", decayed)
	}
}

func TestStatsAggregateOnlyActualEventDamageAndBombActions(t *testing.T) {
	input := makeTestInput(1701)
	engine := newProductionMatchEngine(input)
	engine.stats = engine.buildStats()
	round := &RoundResult{Events: []*GameEvent{
		{EventType: EventDamage, AttackerID: "team_a_p1", Extra: map[string]interface{}{"damage": 70}},
		{EventType: EventDamage, AttackerID: "team_a_p1", Extra: map[string]interface{}{"damage": 30}},
		{EventType: EventKill, AttackerID: "team_a_p1", VictimID: "team_b_p1", IsFirstKill: true, Extra: map[string]interface{}{"assist_ids": []string{"team_a_p2"}}},
		{EventType: EventBombPlant, AttackerID: "team_a_p2"},
		{EventType: EventBombDefuse, AttackerID: "team_b_p2"},
	}}
	engine.aggregateRoundStats(round)
	if got := engine.stats["team_a_p1"]; got.Damage != 100 || got.Kills != 1 || got.FK != 1 || got.MK != 0 {
		t.Fatalf("damage/kill aggregation mismatch: %+v", got)
	}
	if engine.stats["team_b_p1"].Deaths != 1 || engine.stats["team_a_p2"].Assists != 1 || engine.stats["team_a_p2"].Plants != 1 || engine.stats["team_b_p2"].Defuses != 1 {
		t.Fatalf("death/bomb aggregation mismatch: stats=%+v", engine.stats)
	}
}

func TestExplainableReportGroupsOnlyExistingReasonsByPerspective(t *testing.T) {
	winnerReason := &EventReason{Code: "ACTUAL_WIN_FACTOR", SourceActionID: "act-win"}
	lossReason := &EventReason{Code: "ACTUAL_LOSS_FACTOR", SourceActionID: "act-loss"}
	round := &RoundResult{
		WinnerTeamID: "team_a", StrategyTemplateID: "TPL_A", RouteMain: "ROUTE_A",
		Events: []*GameEvent{
			{EventID: "evt-win", SourceActionID: "act-win", EventType: EventKill, AttackerTeamID: "team_a", VictimTeamID: "team_b", Reason: winnerReason},
			{EventID: "evt-loss", SourceActionID: "act-loss", EventType: EventRotate, AttackerTeamID: "team_b", Reason: lossReason},
		},
	}
	report := BuildExplainableReport(round)
	if len(report.KeyEvents) != 2 || len(report.WinFactors) != 1 || report.WinFactors[0].Code != "ACTUAL_WIN_FACTOR" || len(report.LossReasons) != 2 {
		t.Fatalf("report perspective aggregation mismatch: %+v", report)
	}
	for _, reason := range append(append([]*EventReason{}, report.WinFactors...), report.LossReasons...) {
		if reason.Code != "ACTUAL_WIN_FACTOR" && reason.Code != "ACTUAL_LOSS_FACTOR" {
			t.Fatalf("report fabricated a reason: %+v", reason)
		}
	}
}
