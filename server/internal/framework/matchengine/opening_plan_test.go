package matchengine

import "testing"

func TestOpeningAndCTSetupSeedsAreIndependentAndStable(t *testing.T) {
	roundSeed := deriveSeed(int64(99), DefaultMapVersion, "cs2_mr12_ot_mr3_v1", 1)
	if deriveSeed(roundSeed, "opening_plan", 0) == deriveSeed(roundSeed, "ct_setup", 0) {
		t.Fatal("T opening and CT setup must not share a derived seed stream")
	}
	if deriveSeed(roundSeed, "opening_plan", 0) == deriveSeed(roundSeed, "opening_plan", 1) {
		t.Fatal("attempt ordinal must change the opening-plan seed")
	}

	input := makeTestInput(99)
	ctAlternative := input.MapConfig.RouteTemplates["TPL_CT"]
	ctAlternative.ID = "TPL_CT_ALT"
	input.MapConfig.RouteTemplates[ctAlternative.ID] = ctAlternative
	engine := newProductionMatchEngine(input)
	first, err := engine.planOpening(roundSeed, SideCT)
	if err != nil {
		t.Fatalf("first CT setup failed: %v", err)
	}
	second, err := engine.planOpening(roundSeed, SideCT)
	if err != nil {
		t.Fatalf("second CT setup failed: %v", err)
	}
	if first.Template.ID != second.Template.ID || first.PrimaryRoute.ID != second.PrimaryRoute.ID || first.AttemptOrdinal != second.AttemptOrdinal {
		t.Fatalf("same root seed produced different CT setup: %+v / %+v", first, second)
	}
}

func TestCTSetupSelectionDoesNotReadCurrentTRoundPlan(t *testing.T) {
	input := makeTestInput(123)
	ctAlternative := input.MapConfig.RouteTemplates["TPL_CT"]
	ctAlternative.ID = "TPL_CT_ALT"
	input.MapConfig.RouteTemplates[ctAlternative.ID] = ctAlternative
	roundSeed := deriveSeed(input.Seed, input.MapVersion, input.RuleSet.RuleSetID, 1)

	before, err := newProductionMatchEngine(input).planOpening(roundSeed, SideCT)
	if err != nil {
		t.Fatalf("CT setup failed before T change: %v", err)
	}
	tTemplate := input.MapConfig.RouteTemplates["TPL_A"]
	tTemplate.TargetSite = "B"
	tTemplate.CommonCTSetupIDs = []string{"UNRELATED_HIDDEN_T_PRIOR"}
	tTemplate.RouteAllocations = map[string]int{"D2_A_LONG": 2}
	input.MapConfig.RouteTemplates["TPL_A"] = tTemplate
	after, err := newProductionMatchEngine(input).planOpening(roundSeed, SideCT)
	if err != nil {
		t.Fatalf("CT setup failed after T change: %v", err)
	}
	if before.Template.ID != after.Template.ID || before.PrimaryRoute.ID != after.PrimaryRoute.ID || before.AttemptOrdinal != after.AttemptOrdinal {
		t.Fatalf("CT setup leaked the current T hidden plan: before=%+v after=%+v", before, after)
	}
}

func TestOpeningPlanRetriesThenUsesOnlyConfiguredDefault(t *testing.T) {
	input := makeTestInput(321)
	bad := input.MapConfig.RouteTemplates["TPL_A"]
	bad.ID = "TPL_BAD"
	bad.RouteIDs = []string{"D2_CT_A"}
	bad.RouteAllocations = map[string]int{"D2_CT_A": 5}
	input.MapConfig.RouteTemplates[bad.ID] = bad
	roundSeed := deriveSeed(input.Seed, input.MapVersion, input.RuleSet.RuleSetID, 1)

	selection, err := newProductionMatchEngine(input).planOpening(roundSeed, SideT)
	if err != nil {
		t.Fatalf("configured default should recover the invalid candidate: %v", err)
	}
	if !selection.UsedDefault || selection.Template.ID != "TPL_A" || selection.AttemptOrdinal != 1 {
		t.Fatalf("expected fallback to configured T default after attempts 0/1, got %+v", selection)
	}

	defaultTemplate := input.MapConfig.RouteTemplates["TPL_A"]
	defaultTemplate.RouteAllocations = map[string]int{"D2_A_LONG": 4}
	input.MapConfig.RouteTemplates["TPL_A"] = defaultTemplate
	_, err = newProductionMatchEngine(input).planOpening(roundSeed, SideT)
	assertEngineErrorCode(t, err, "INVALID_OPENING_PLAN")
}
