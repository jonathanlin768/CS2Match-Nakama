package matchengine

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func uniformAttributes(value int) PlayerAttributes {
	return PlayerAttributes{Entry: value, Aim: value, Trade: value, Clutch: value, Firepower: value, Gamesense: value, Reaction: value, Positioning: value, Awareness: value, Teamplay: value, Utility: value, Composure: value, Mobility: value, Endurance: value, Discipline: value}
}

func TestPlayerCombatAndSurvivalFrozenFormulaGolden(t *testing.T) {
	state := makeTestRoundState(t, 801)
	player := state.Players["team_a_p2"]
	player.Profile.Attributes = uniformAttributes(80)
	modifiers := CombatModifierInput{RoleTagModifier: 5, WeaponModifier: 4, PostureModifier: 3, VisibilityModifier: 2, TeamSupportModifier: 1, StaminaPenalty: 6, DamagePenalty: 7, SuppressionPenalty: 8}
	combat, err := CalculatePlayerCombatScore(state, player.Profile.PlayerID, "SCN_A", modifiers)
	if err != nil {
		t.Fatal(err)
	}
	if combat.WeightedPlayerScore != 80 || combat.PlayerCombatScore != 74 {
		t.Fatalf("PlayerCombatScore golden mismatch: %+v", combat)
	}
	player.Focus = 70
	survival := CalculateTargetSurvivalScore(player, SurvivalModifierInput{CoverModifier: 5, TeamSupportModifier: 4, MovementExposure: 3, DamagePenalty: 2})
	if survival.TargetSurvivalScore != 234 {
		t.Fatalf("TargetSurvivalScore golden mismatch: %+v", survival)
	}
	combatType := reflect.TypeOf(CombatModifierInput{})
	for _, forbidden := range []string{"Utility", "Momentum", "TimePressure"} {
		for index := 0; index < combatType.NumField(); index++ {
			if strings.Contains(combatType.Field(index).Name, forbidden) {
				t.Fatalf("%s was double-counted in PlayerCombatScore", forbidden)
			}
		}
	}
	survivalType := reflect.TypeOf(TargetSurvivalBreakdown{})
	for _, forbidden := range []string{"Suppression", "WeightedSurvival"} {
		for index := 0; index < survivalType.NumField(); index++ {
			if strings.Contains(survivalType.Field(index).Name, forbidden) {
				t.Fatalf("forbidden survival component %s found", forbidden)
			}
		}
	}
}

func TestConfiguredCombatModifiersStayWithinFrozenFormulaTerms(t *testing.T) {
	state := makeTestRoundState(t, 706)
	scenario := state.scenarios["SCN_A"]
	tag := state.mapTags["TAG_A"]
	tag.Category = "Posture"
	tag.Side = SideT
	state.mapTags[tag.ID] = tag
	modifier := state.encounterModifiers["MOD_A"]
	modifier.Factor = "SyncPeek"
	modifier.Side = SideT
	state.encounterModifiers[modifier.ID] = modifier
	tPosture, tVisibility, tSupport := configuredCombatTerms(state, state.Players["team_a_p1"], scenario)
	ctPosture, ctVisibility, ctSupport := configuredCombatTerms(state, state.Players["team_b_p1"], scenario)
	if tPosture != 10 || tVisibility != 0 || tSupport <= 0 || ctPosture != 0 || ctVisibility != 0 || ctSupport != 0 {
		t.Fatalf("scenario config escaped or missed frozen terms: T=%f/%f/%f CT=%f/%f/%f", tPosture, tVisibility, tSupport, ctPosture, ctVisibility, ctSupport)
	}
}

func TestScenarioWeightsExactlyTenAndSumOneHundred(t *testing.T) {
	state := makeTestRoundState(t, 802)
	score, err := WeightedScenarioPlayerScore(uniformAttributes(73), "SCN_A", state.encounterModifiers)
	if err != nil || score != 73 {
		t.Fatalf("ten-weight score = %v/%v", score, err)
	}
	bad := make(map[string]EncounterModifier, len(state.encounterModifiers))
	for key, modifier := range state.encounterModifiers {
		bad[key] = modifier
	}
	for key, modifier := range bad {
		if modifier.Factor == "ScenarioWeight" {
			modifier.Weight++
			bad[key] = modifier
			break
		}
	}
	assertEngineErrorCode(t, func() error { _, err := WeightedScenarioPlayerScore(uniformAttributes(73), "SCN_A", bad); return err }(), "CONFIG_BAD_WEIGHT_SUM")
}

func TestCoordinationUsesActualParticipantsAndEncounterComponentsOnce(t *testing.T) {
	state := makeUtilityRoundState(t, 803, true)
	encounter := &EncounterState{ID: "enc-score", ScenarioID: "SCN_A", ActorIDs: []string{"team_a_p1", "team_a_p2", "team_b_p1"}, NodeID: "A_SITE", StartedAt: 0, Status: EncounterActive}
	state.Players["team_a_p1"].Location = PlayerLocation{NodeID: "A_SITE"}
	state.Players["team_a_p2"].Location = PlayerLocation{NodeID: "A_LONG"}
	state.Players["team_b_p1"].Location = PlayerLocation{NodeID: "A_SITE"}
	metricsT := CalculateCoordination(state, encounter, SideT)
	metricsCT := CalculateCoordination(state, encounter, SideCT)
	if metricsT.Crossfire <= 0 || metricsT.TradeCoverage <= 0 || metricsCT.Isolation <= 0 {
		t.Fatalf("actual participant coordination mismatch: T=%+v CT=%+v", metricsT, metricsCT)
	}
	if _, err := SpendUtility(state, SideT, UtilityRequest{Type: UtilityOpeningInitiative, ActorIDs: []string{"team_a_p1"}, ScopeEncounterID: encounter.ID, StartAt: 0, Duration: 5, BaseCost: 1}); err != nil {
		t.Fatal(err)
	}
	state.MomentumT = 20
	scores, err := CalculateEncounterScorePair(state, encounter, "SCN_A", 123)
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{SideT, SideCT} {
		score := scores[side]
		want := score.PlayerScoreSum + score.TeamModifier + score.ScenarioModifier + score.UtilityModifier + score.MomentumModifier + score.TimePressureModifier
		if math.Abs(score.DeterministicEncounterScore-want) > 0.000001 {
			t.Fatalf("EncounterScore component counted more than once: %+v", score)
		}
		counts := map[string]int{}
		for _, reason := range score.Reasons {
			counts[reason.Code]++
		}
		for _, code := range []string{"UTILITY_ONCE", "MOMENTUM_ONCE", "TIME_PRESSURE_ONCE"} {
			if counts[code] != 1 {
				t.Fatalf("%s reason count = %d", code, counts[code])
			}
		}
	}
}

func TestEncounterCloseAndDecisiveTrendProtectionDoesNotSkipRolls(t *testing.T) {
	results := map[string]EncounterScoreBreakdown{
		SideT:  {Side: SideT, DeterministicEncounterScore: 100, BoundedRandomNoise: -100},
		SideCT: {Side: SideCT, DeterministicEncounterScore: 50, BoundedRandomNoise: 100},
	}
	protectEncounterTrend(results, 18)
	if results[SideT].DeterministicEncounterScore+results[SideT].BoundedRandomNoise <= results[SideCT].DeterministicEncounterScore+results[SideCT].BoundedRandomNoise {
		t.Fatalf("decisive encounter noise reversed trend: %+v", results)
	}
	state := makeTestRoundState(t, 808)
	for _, player := range state.Players {
		player.Profile.Attributes = uniformAttributes(70)
	}
	encounter := &EncounterState{ID: "close", ScenarioID: "SCN_A", ActorIDs: []string{"team_a_p2", "team_b_p1"}, NodeID: "A_SITE", Status: EncounterActive}
	scores, err := CalculateEncounterScorePair(state, encounter, "SCN_A", 55)
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{SideT, SideCT} {
		foundClose := false
		for _, reason := range scores[side].Reasons {
			if reason.Code == "CLOSE_SCORE" {
				foundClose = true
			}
		}
		if !foundClose {
			t.Fatalf("close deterministic score had no explanation: %+v", scores[side])
		}
	}
	// Trend protection constrains only score noise; attack probability still
	// samples independent hit/lethal/damage rolls (covered by distinct values).
	if stableUnit(55, "actor", "target", "hit") == stableUnit(55, "actor", "target", "lethal") || stableUnit(55, "actor", "target", "lethal") == stableUnit(55, "actor", "target", "damage") {
		t.Fatal("combat rolls collapsed into the encounter score noise stream")
	}
}

func TestTargetSelectionStableOrderingAndIdentityRoll(t *testing.T) {
	constants := makeTestMapConfig().CombatConstants
	setConstFloat(&constants, "CloseScoreGap", "0")
	snapshot := CombatPulseSnapshot{ActorIDs: []string{"a", "z", "b"}, Players: map[string]PulsePlayerSnapshot{
		"a": {PlayerID: "a", Side: SideT, Alive: true, X: 0, Y: 0, Profile: PlayerProfile{Attributes: uniformAttributes(50)}},
		"z": {PlayerID: "z", Side: SideCT, Alive: true, HP: 80, X: 1, Y: 1, Profile: PlayerProfile{Attributes: uniformAttributes(90)}, Posture: PostureHolding},
		"b": {PlayerID: "b", Side: SideCT, Alive: true, HP: 20, X: 0.1, Y: 0.1, Profile: PlayerProfile{Attributes: uniformAttributes(50)}, Posture: PostureMoving},
	}}
	target, ok := SelectCombatTarget(snapshot, "a", 99, constants)
	if !ok || target.PlayerID != "z" {
		t.Fatalf("ThreatScore ordering was not primary: %+v", target)
	}
	target2, _ := SelectCombatTarget(snapshot, "a", 99, constants)
	if !reflect.DeepEqual(target, target2) {
		t.Fatal("identity-derived target selection is not stable")
	}
	snapshot.Players = reversePulsePlayerMap(snapshot.Players)
	target3, _ := SelectCombatTarget(snapshot, "a", 99, constants)
	if !reflect.DeepEqual(target, target3) {
		t.Fatal("pulse player map order changed target selection")
	}
}

func TestTargetOrderingUsesExposureDistanceHPAndPlayerIDAfterThreat(t *testing.T) {
	constants := makeTestMapConfig().CombatConstants
	setConstFloat(&constants, "CloseScoreGap", "-1")
	base := PulsePlayerSnapshot{Side: SideCT, Alive: true, HP: 50, Profile: PlayerProfile{Attributes: uniformAttributes(60)}}
	snapshot := CombatPulseSnapshot{ActorIDs: []string{"actor", "exposed", "near", "lowhp", "alpha"}, Players: map[string]PulsePlayerSnapshot{
		"actor": {PlayerID: "actor", Side: SideT, Alive: true, X: 0, Y: 0, Profile: PlayerProfile{Attributes: uniformAttributes(50)}},
	}}
	exposed, near, lowHP, alpha := base, base, base, base
	exposed.PlayerID, exposed.Posture, exposed.X = "exposed", PostureMoving, 1
	near.PlayerID, near.X = "near", 0.1
	lowHP.PlayerID, lowHP.X, lowHP.HP = "lowhp", 0.1, 20
	alpha.PlayerID, alpha.X, alpha.HP = "alpha", 0.1, 20
	snapshot.Players["exposed"], snapshot.Players["near"], snapshot.Players["lowhp"], snapshot.Players["alpha"] = exposed, near, lowHP, alpha
	target, ok := SelectCombatTarget(snapshot, "actor", 7, constants)
	if !ok || target.PlayerID != "exposed" {
		t.Fatalf("exposure did not outrank distance/HP/ID after equal threat: %+v", target)
	}
	exposed.Posture = PostureDefault
	snapshot.Players["exposed"] = exposed
	target, _ = SelectCombatTarget(snapshot, "actor", 7, constants)
	if target.PlayerID != "alpha" {
		t.Fatalf("distance/HP/PlayerID stable ordering mismatch: %+v", target)
	}
	snapshot.VisibleTargets = map[string]map[string]bool{"actor": {"alpha": false, "lowhp": true, "near": true, "exposed": true}}
	target, _ = SelectCombatTarget(snapshot, "actor", 7, constants)
	if target.PlayerID == "alpha" {
		t.Fatal("invisible target was not filtered")
	}
}

func TestHitKillDamageWeaponArmorRangeAndNoTargetHPFactor(t *testing.T) {
	constants := makeTestMapConfig().CombatConstants
	ak := makeTestWeapons()[WeaponAK47]
	weak := ak
	weak.Damage, weak.RoundsPerMinute, weak.MagazineSize, weak.ArmorPenetration, weak.RangeModifier = 20, 300, 10, 0.2, 0.8
	scenario := Scenario{Range: "Long"}
	visibility := Visibility{ExposureModifier: 20}
	unarmored := CalculateAttackProbabilities(100, 80, ak, false, scenario, visibility, constants)
	armored := CalculateAttackProbabilities(100, 80, ak, true, scenario, visibility, constants)
	weakResult := CalculateAttackProbabilities(100, 80, weak, true, scenario, visibility, constants)
	if unarmored.KillChance > unarmored.HitChance || unarmored.KillChance > constants.Float("MaxKillChance", 1) || armored.BaseShotDamage >= unarmored.BaseShotDamage || weakResult.DamagePotential >= armored.DamagePotential || weakResult.BurstCapacity >= armored.BurstCapacity {
		t.Fatalf("probability/weapon contract mismatch: unarmored=%+v armored=%+v weak=%+v", unarmored, armored, weakResult)
	}
	lowHP := CalculateAttackProbabilities(100, 80, ak, true, scenario, visibility, constants)
	if !reflect.DeepEqual(armored, lowHP) {
		t.Fatal("target HP changed KillChance via a forbidden TargetHPFactor")
	}
}

func TestAttackWindowsOnlyProduceMissOrDamageBeforeCommit(t *testing.T) {
	state := makeTestRoundState(t, 804)
	candidate, _ := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1"})
	schedule, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	action := schedule.PulseActions[0]
	state.Timeline = action.ResolveAt
	snapshot, err := CreateCombatPulseSnapshot(state, state.ActiveEngagements[candidate.ID])
	if err != nil {
		t.Fatal(err)
	}
	target, _ := SelectCombatTarget(snapshot, "team_a_p2", 1, state.constants)
	foundMiss, foundNonLethal, foundLethal := false, false, false
	for seed := int64(0); seed < 10000 && !(foundMiss && foundNonLethal && foundLethal); seed++ {
		attack, err := ResolveAttackWindow(state, snapshot, action, "team_a_p2", target, seed, 0)
		if err != nil {
			t.Fatal(err)
		}
		if attack.Effect.Type == EffectMiss {
			foundMiss = true
			continue
		}
		if attack.Effect.Type != EffectDamage {
			t.Fatalf("attack directly produced %s before HP commit", attack.Effect.Type)
		}
		if attack.LethalWindow {
			foundLethal = attack.Effect.Amount >= snapshot.Players[target.PlayerID].HP
		} else {
			foundNonLethal = attack.Effect.Amount > 0
		}
	}
	if !foundMiss || !foundNonLethal || !foundLethal {
		t.Fatalf("did not cover miss/nonlethal/lethal windows: %v/%v/%v", foundMiss, foundNonLethal, foundLethal)
	}
}

func TestLowHPDiesNaturallyFromNonLethalDamageWindow(t *testing.T) {
	state := makeTestRoundState(t, 807)
	candidate, _ := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_b_p1"})
	schedule, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	action := schedule.PulseActions[0]
	state.Timeline = action.ResolveAt
	targetID := "team_b_p1"
	state.Players[targetID].HP = 5
	snapshot, err := CreateCombatPulseSnapshot(state, state.ActiveEngagements[candidate.ID])
	if err != nil {
		t.Fatal(err)
	}
	target, _ := SelectCombatTarget(snapshot, "team_a_p2", 1, state.constants)
	var damage Effect
	for seed := int64(0); seed < 10000; seed++ {
		attack, err := ResolveAttackWindow(state, snapshot, action, "team_a_p2", target, seed, 0)
		if err != nil {
			t.Fatal(err)
		}
		if attack.Hit && !attack.LethalWindow && attack.Effect.Amount >= 5 {
			damage = attack.Effect
			break
		}
	}
	if damage.ID == "" {
		t.Fatal("could not produce a non-lethal damage window")
	}
	if _, err := ApplyCombatPulseCommit(state, action, []Effect{damage}); err != nil {
		t.Fatal(err)
	}
	if state.Players[targetID].Alive {
		t.Fatal("low HP was not naturally eliminated by actual damage")
	}
}

func TestCombatPulseSnapshotAftermathAndNaturalFirstKill(t *testing.T) {
	state := makeTestRoundState(t, 805)
	tID, ctID := "team_a_p2", "team_b_p1"
	action := combatAction(state, "natural-first", tID, ctID)
	batch, err := ApplyCombatPulseCommit(state, action, []Effect{
		{ID: NewEffectID(state.Seed, action.ID, EffectDamage, 0), SourceActionID: action.ID, Type: EffectDamage, Priority: 100, Timestamp: 0, ActorID: tID, TargetID: ctID, Amount: 100, StringValue: WeaponAK47},
		{ID: NewEffectID(state.Seed, action.ID, EffectDamage, 1), SourceActionID: action.ID, Type: EffectDamage, Priority: 100, Timestamp: 0, ActorID: ctID, TargetID: tID, Amount: 30, StringValue: WeaponM4A1S},
	})
	if err != nil {
		t.Fatal(err)
	}
	var kill *GameEvent
	for _, event := range batch.Events {
		if event.EventType == EventKill {
			kill = event
		}
	}
	if kill == nil || !kill.IsFirstKill || kill.AttackerID != tID || state.Players[tID].HP != 70 || state.Players[ctID].Alive {
		t.Fatalf("first kill was pre-drawn or retaliation disappeared: kill=%+v T=%+v CT=%+v", kill, state.Players[tID], state.Players[ctID])
	}
	state.Timeline = 3
	tradeAction := combatAction(state, "natural-trade", "team_b_p2")
	tradeBatch, err := ApplyCombatPulseCommit(state, tradeAction, []Effect{{ID: NewEffectID(state.Seed, tradeAction.ID, EffectDamage, 0), SourceActionID: tradeAction.ID, Type: EffectDamage, Priority: 100, Timestamp: 3, ActorID: "team_b_p2", TargetID: tID, Amount: 100, StringValue: WeaponM4A1S}})
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range tradeBatch.Events {
		if event.EventType == EventKill && (event.IsFirstKill || !event.IsTrade) {
			t.Fatalf("natural trade markers mismatch: %+v", event)
		}
	}
}

func TestResolveCombatPulseUsesOneSnapshotAndUpdatesResources(t *testing.T) {
	state := makeTestRoundState(t, 806)
	candidate, _ := BuildEncounterCandidate(state, "contact", "SCN_A", "A_SITE", []string{"team_a_p2", "team_a_p3", "team_b_p1", "team_b_p2"})
	schedule, err := StartEncounter(state, candidate)
	if err != nil {
		t.Fatal(err)
	}
	action := schedule.PulseActions[0]
	state.Timeline = action.ResolveAt
	result, err := ResolveCombatPulse(state, action)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Attacks) == 0 || result.Batch == nil || result.Snapshot.Timeline != action.ResolveAt {
		t.Fatalf("pulse did not resolve from a single snapshot: %+v", result)
	}
	for _, attack := range result.Attacks {
		if result.Snapshot.Players[attack.ActorID].Alive != true {
			t.Fatal("non-snapshot actor attacked")
		}
		if player := state.Players[attack.ActorID]; player.Alive && (player.Focus >= result.Snapshot.Players[attack.ActorID].Focus || player.Stamina >= result.Snapshot.Players[attack.ActorID].Stamina) {
			t.Fatalf("combat aftermath did not consume resources: before=%+v after=%+v", result.Snapshot.Players[attack.ActorID], player)
		}
	}
}

func TestCombatAftermathUpdatesSuppressionMomentumAndPosture(t *testing.T) {
	state := makeTestRoundState(t, 809)
	attacker, target := state.Players["team_a_p2"], state.Players["team_b_p1"]
	snapshot := CombatPulseSnapshot{Players: map[string]PulsePlayerSnapshot{
		attacker.Profile.PlayerID: {PlayerID: attacker.Profile.PlayerID, Focus: attacker.Focus, Stamina: attacker.Stamina},
		target.Profile.PlayerID:   {PlayerID: target.Profile.PlayerID, Focus: target.Focus, Stamina: target.Stamina},
	}}
	batch := &AppliedBatch{Events: []*GameEvent{{EventType: EventKill, AttackerID: attacker.Profile.PlayerID, VictimID: target.Profile.PlayerID}}}
	applyCombatAftermath(state, snapshot, []AttackWindowResult{{ActorID: attacker.Profile.PlayerID, TargetID: target.Profile.PlayerID, Hit: true}}, batch)
	if attacker.Focus != 98 || attacker.Stamina != 98 || attacker.Posture != PostureEngaged || !target.Suppressed || target.Focus != 92 || target.Posture != PostureHolding || state.MomentumT != 5 || state.MomentumCT != -5 {
		t.Fatalf("combat aftermath mismatch: attacker=%+v target=%+v momentum=%d/%d", attacker, target, state.MomentumT, state.MomentumCT)
	}
}

func reversePulsePlayerMap(source map[string]PulsePlayerSnapshot) map[string]PulsePlayerSnapshot {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	result := map[string]PulsePlayerSnapshot{}
	for index := len(keys) - 1; index >= 0; index-- {
		result[keys[index]] = source[keys[index]]
	}
	return result
}
