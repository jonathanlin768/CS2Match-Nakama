package matchengine

import (
	"reflect"
	"testing"
)

func makeUtilityRoundState(t *testing.T, seed int64, support bool) *RoundState {
	t.Helper()
	input := makeTestRoundInput(seed)
	input.SideLoadouts[SideT] = WeaponLoadout{Primary: WeaponAK47, Armor: true, Helmet: true, Grenades: []string{"smoke", "flash"}}
	input.SideLoadouts[SideCT] = WeaponLoadout{Primary: WeaponM4A1S, Armor: true, Helmet: true, HasKit: true, Grenades: []string{"smoke", "flash"}}
	if support {
		input.TeamT.Players[0].RoleTags = []string{"Support"}
	}
	state, err := NewRoundState(input, RoundPlan{TStrategyTemplateID: "TPL_A", CTSetupTemplateID: "TPL_CT", BombCarrierID: input.TeamT.Players[1].PlayerID})
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestUtilityBudgetUsesAttributesLoadoutSupportAndConfiguredCap(t *testing.T) {
	plain := makeTestRoundState(t, 601)
	equipped := makeUtilityRoundState(t, 601, false)
	support := makeUtilityRoundState(t, 601, true)
	capBudget := support.constants.Int("UtilityBudget", 0)
	if plain.Utility[SideT].Budget >= equipped.Utility[SideT].Budget || equipped.Utility[SideT].Budget >= support.Utility[SideT].Budget || support.Utility[SideT].Budget > capBudget {
		t.Fatalf("utility initialization mismatch: plain=%d equipped=%d support=%d cap=%d", plain.Utility[SideT].Budget, equipped.Utility[SideT].Budget, support.Utility[SideT].Budget, capBudget)
	}
}

func TestScopedUtilityWindowsCoverAllEffectsAndNeverBecomeGlobalPower(t *testing.T) {
	state := makeUtilityRoundState(t, 602, true)
	types := []UtilityEffectType{UtilityVisibilitySuppression, UtilityOpeningInitiative, UtilityExposureReduction, UtilitySyncPeekQuality, UtilityPlantCover, UtilityDefuseCover}
	for index, effectType := range types {
		result, err := SpendUtility(state, SideT, UtilityRequest{Type: effectType, ActorIDs: []string{"team_a_p1"}, ScopeEncounterID: "encounter-1", StartAt: index, Duration: 10, BaseCost: 1})
		if err != nil || !result.Applied || result.Window == nil || result.Window.Modifier <= 0 {
			t.Fatalf("utility %s not applied: %+v/%v", effectType, result, err)
		}
		if ScopedUtilityModifier(state, SideT, effectType, "", "encounter-1", index) <= 0 || ScopedUtilityModifier(state, SideT, effectType, "", "encounter-2", index) != 0 {
			t.Fatalf("utility %s escaped its encounter scope", effectType)
		}
	}
	if _, ok := reflect.TypeOf(*state).FieldByName("TeamPower"); ok {
		t.Fatal("Utility introduced a forbidden global TeamPower field")
	}
}

func TestSupportEfficiencyBudgetExhaustionAndNoCrossEncounterReuse(t *testing.T) {
	plain := makeUtilityRoundState(t, 603, false)
	support := makeUtilityRoundState(t, 603, true)
	request := UtilityRequest{Type: UtilityVisibilitySuppression, ActorIDs: []string{"team_a_p1"}, ScopeEncounterID: "enc-1", StartAt: 0, Duration: 5, BaseCost: 20}
	plainResult, err := SpendUtility(plain, SideT, request)
	if err != nil {
		t.Fatal(err)
	}
	supportResult, err := SpendUtility(support, SideT, request)
	if err != nil {
		t.Fatal(err)
	}
	if !plainResult.Applied || !supportResult.Applied || supportResult.Window.Cost >= plainResult.Window.Cost {
		t.Fatalf("Support did not improve efficiency: plain=%+v support=%+v", plainResult, supportResult)
	}
	if ScopedUtilityModifier(support, SideT, request.Type, "", "enc-2", 2) != 0 || ScopedUtilityModifier(support, SideT, request.Type, "", "enc-1", 5) != 0 {
		t.Fatal("utility window was reused by another encounter or after expiry")
	}
	support.Utility[SideT].Budget = support.Utility[SideT].Spent
	low, err := SpendUtility(support, SideT, UtilityRequest{Type: UtilityPlantCover, ActorIDs: []string{"team_a_p1"}, ScopeActionID: "plant", StartAt: 0, Duration: 4, BaseCost: 10})
	if err != nil || low.Applied || low.Reason.Code != "LOW_UTILITY" || ScopedUtilityModifier(support, SideT, UtilityPlantCover, "plant", "", 1) != 0 {
		t.Fatalf("exhausted budget still applied: %+v/%v", low, err)
	}
}

func TestUtilitySpendIsSeedStable(t *testing.T) {
	left := makeUtilityRoundState(t, 604, true)
	right := makeUtilityRoundState(t, 604, true)
	request := UtilityRequest{Type: UtilityDefuseCover, ActorIDs: []string{"team_a_p1"}, ScopeActionID: "defuse", StartAt: 3, Duration: 6, BaseCost: 15}
	leftResult, err := SpendUtility(left, SideT, request)
	if err != nil {
		t.Fatal(err)
	}
	rightResult, err := SpendUtility(right, SideT, request)
	if err != nil || !reflect.DeepEqual(leftResult, rightResult) {
		t.Fatalf("same seed utility differs: left=%+v right=%+v err=%v", leftResult, rightResult, err)
	}
}

func TestProductionEncounterAndBombActionsSpendScopedUtility(t *testing.T) {
	state := makeUtilityRoundState(t, 605, true)
	candidate, err := BuildEncounterCandidate(state, "utility-contact", "SCN_A", "A_SITE", []string{"team_a_p1", "team_b_p1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := StartEncounter(state, candidate); err != nil {
		t.Fatal(err)
	}
	if state.Utility[SideT].Spent == 0 || state.Utility[SideCT].Spent == 0 || ScopedUtilityModifier(state, SideT, UtilityOpeningInitiative, "", candidate.ID, state.Timeline) == 0 {
		t.Fatalf("production encounter did not consume scoped utility: T=%+v CT=%+v", state.Utility[SideT], state.Utility[SideCT])
	}

	plantState, carrier := preparePlantState(t, 606, "A")
	plantState.Players[carrier].Weapon.Grenades = []string{"smoke"}
	before := plantState.Utility[SideT].Spent
	action, err := StartPlantAction(plantState, carrier, "A", "utility-plant", 0)
	if err != nil {
		t.Fatal(err)
	}
	if plantState.Utility[SideT].Spent <= before || len(plantState.Events) == 0 || plantState.Events[len(plantState.Events)-1].Reason == nil || len(plantState.Events[len(plantState.Events)-1].Reason.Modifiers) == 0 {
		t.Fatalf("production plant did not expose scoped utility spend: action=%+v utility=%+v events=%+v", action, plantState.Utility[SideT], plantState.Events)
	}
}
