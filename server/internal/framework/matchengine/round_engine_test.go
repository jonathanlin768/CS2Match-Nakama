package matchengine

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func TestCausalRoundEngineRunsAuthoritativeLoopToTerminal(t *testing.T) {
	input := makeTestRoundInput(1401)
	result, err := runCausalRound(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.Round == nil || result.Terminal == nil || result.Round.WinnerTeamID != result.Terminal.WinnerTeamID || result.Round.WinReason != result.Terminal.WinReason {
		t.Fatalf("round/terminal contract mismatch: %+v", result)
	}
	if violations := roundInvariantViolations(result.Round); len(violations) > 0 {
		t.Fatalf("causal round invariant violations: %v", violations)
	}
	if len(result.PhaseHistory) < 3 || result.PhaseHistory[0] != PhaseOpeningDeploy || result.PhaseHistory[len(result.PhaseHistory)-1] != PhaseRoundEnd {
		t.Fatalf("phase projection did not cycle to RoundEnd: %v", result.PhaseHistory)
	}
	if !hasPhase(result.PhaseHistory, PhaseAdvance) || !hasPhase(result.PhaseHistory, PhaseClash) {
		t.Fatalf("opening movement/contact phases missing: %v", result.PhaseHistory)
	}
}

func TestCausalRoundEngineSameSeedDeepEqual(t *testing.T) {
	left, err := runCausalRound(context.Background(), makeTestRoundInput(1402))
	if err != nil {
		t.Fatal(err)
	}
	right, err := runCausalRound(context.Background(), makeTestRoundInput(1402))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatal("same round input/seed did not produce a deep-equal causal result")
	}
}

func TestProductionMatchDeepEqualAcrossMapInsertionOrder(t *testing.T) {
	leftInput := makeTestInput(1403)
	leftInput.StartTime = 1700000000000
	rightInput := makeTestInput(1403)
	rightInput.StartTime = leftInput.StartTime
	rightInput.MapConfig = reverseMapConfigInsertion(rightInput.MapConfig)
	rightInput.WeaponSpecs = reverseStringMap(rightInput.WeaponSpecs)
	rightInput.SideLoadouts = reverseStringMap(rightInput.SideLoadouts)
	rightInput.InitialSideByTeam = reverseStringMap(rightInput.InitialSideByTeam)

	left, err := newProductionMatchEngine(leftInput).simulateMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	right, err := newProductionMatchEngine(rightInput).simulateMatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatal("action/event/reason/location/terminal/scores/stats changed with map insertion order")
	}
}

func reverseMapConfigInsertion(source *MapConfig) *MapConfig {
	copy := *source
	copy.RouteTemplates = reverseStringMap(source.RouteTemplates)
	copy.Scenarios = reverseStringMap(source.Scenarios)
	copy.MapTags = reverseStringMap(source.MapTags)
	copy.EncounterModifiers = reverseStringMap(source.EncounterModifiers)
	copy.Nodes = reverseStringMap(source.Nodes)
	copy.Edges = reverseStringMap(source.Edges)
	copy.Visibility = reverseStringMap(source.Visibility)
	copy.Routes = reverseStringMap(source.Routes)
	copy.CombatConstants.Values = reverseStringMap(source.CombatConstants.Values)
	return &copy
}

func reverseStringMap[T any](source map[string]T) map[string]T {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	out := make(map[string]T, len(source))
	for _, key := range keys {
		out[key] = source[key]
	}
	return out
}

func TestCausalRoundEngineSupportsRepeatedClashAndDecisionPhases(t *testing.T) {
	coveredRepeatedClash, coveredDecision := false, false
	for seed := int64(1410); seed < 1440 && (!coveredRepeatedClash || !coveredDecision); seed++ {
		result, err := runCausalRound(context.Background(), makeTestRoundInput(seed))
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		clashes := 0
		for _, phase := range result.PhaseHistory {
			if phase == PhaseClash {
				clashes++
			}
		}
		coveredRepeatedClash = coveredRepeatedClash || clashes >= 2
		coveredDecision = coveredDecision || hasPhase(result.PhaseHistory, PhaseRotate)
	}
	if !coveredRepeatedClash || !coveredDecision {
		t.Fatalf("phase loop coverage missing: repeated_clash=%t decision=%t", coveredRepeatedClash, coveredDecision)
	}
}

func TestScenarioSelectionUsesCurrentContactPhase(t *testing.T) {
	state := makeTestRoundState(t, 1441)
	state.scenarios = map[string]Scenario{
		"OPEN":   {ID: "OPEN", Phase: "OpeningDuel", Site: "A"},
		"ENTRY":  {ID: "ENTRY", Phase: "SiteEntry", Site: "A"},
		"RETAKE": {ID: "RETAKE", Phase: "Retake", Site: "A"},
	}
	runtime := &causalRoundRuntime{state: state}
	if got := runtime.scenarioForContact("A_SITE"); got != "ENTRY" {
		t.Fatalf("site contact used %s instead of SiteEntry", got)
	}
	state.Bomb.Status = BombPlanted
	state.Bomb.PlantedSite = "A"
	if got := runtime.scenarioForContact("A_SITE"); got != "RETAKE" {
		t.Fatalf("post-plant contact used %s instead of Retake", got)
	}
}

func hasPhase(phases []RoundPhase, target RoundPhase) bool {
	for _, phase := range phases {
		if phase == target {
			return true
		}
	}
	return false
}
