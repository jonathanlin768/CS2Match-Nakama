package matchengine

import (
	"math"
	"sort"
)

type combatModifierInput struct {
	RoleTagModifier     float64
	WeaponModifier      float64
	PostureModifier     float64
	VisibilityModifier  float64
	TeamSupportModifier float64
	StaminaPenalty      float64
	DamagePenalty       float64
	SuppressionPenalty  float64
}

type playerCombatScoreBreakdown struct {
	WeightedPlayerScore float64
	combatModifierInput
	PlayerCombatScore float64
	Reasons           []reasonRecord
}

type survivalModifierInput struct {
	CoverModifier       float64
	TeamSupportModifier float64
	MovementExposure    float64
	DamagePenalty       float64
}

type targetSurvivalBreakdown struct {
	Positioning float64
	Reaction    float64
	Focus       float64
	survivalModifierInput
	TargetSurvivalScore float64
}

type coordinationMetrics struct {
	TradeCoverage float64
	Crossfire     float64
	Spacing       float64
	SyncPeek      float64
	Isolation     float64
}

func (metrics coordinationMetrics) Total() float64 {
	return metrics.TradeCoverage + metrics.Crossfire + metrics.Spacing + metrics.SyncPeek - metrics.Isolation
}

type encounterScoreBreakdown struct {
	Side                        string
	PlayerScoreSum              float64
	TeamModifier                float64
	ScenarioModifier            float64
	UtilityModifier             float64
	MomentumModifier            float64
	TimePressureModifier        float64
	BoundedRandomNoise          float64
	DeterministicEncounterScore float64
	FinalScore                  float64
	Reasons                     []reasonRecord
}

type pulsePlayerSnapshot struct {
	PlayerID   string
	TeamID     string
	Side       string
	Profile    PlayerProfile
	Weapon     WeaponLoadout
	WeaponSpec WeaponSpec
	Alive      bool
	HP         int
	Stamina    int
	Focus      int
	Suppressed bool
	Posture    combatPosture
	Location   playerLocation
	X          float64
	Y          float64
}

type combatPulseSnapshot struct {
	EncounterID    string
	Scenario       Scenario
	Timeline       int
	Players        map[string]pulsePlayerSnapshot
	ActorIDs       []string
	VisibleTargets map[string]map[string]bool
}

type targetCandidate struct {
	PlayerID    string
	ThreatScore float64
	Exposure    float64
	Distance    float64
	HP          int
	TieRoll     float64
}

type attackProbabilities struct {
	RawHitChance     float64
	HitChance        float64
	BaseShotDamage   float64
	BurstCapacity    int
	DamagePotential  float64
	ExposureModifier float64
	RawKillChance    float64
	KillChance       float64
}

type attackWindowResult struct {
	ActorID       string
	TargetID      string
	Hit           bool
	LethalWindow  bool
	HitRoll       float64
	LethalRoll    float64
	DamageRoll    float64
	Probabilities attackProbabilities
	Effect        effect
}

type combatPulseResult struct {
	EncounterID string
	PulseIndex  int
	Snapshot    combatPulseSnapshot
	Attacks     []attackWindowResult
	Batch       *appliedBatch
	ShouldEnd   bool
	EndReason   string
}

func calculatePlayerCombatScore(state *roundState, playerID, scenarioID string, modifiers combatModifierInput) (playerCombatScoreBreakdown, error) {
	player := state.Players[playerID]
	if player == nil || state.scenarios[scenarioID].ID == "" {
		return playerCombatScoreBreakdown{}, newError("INVALID_COMBAT_SCORE", "combat score requires player and scenario")
	}
	weighted, err := weightedScenarioPlayerScore(player.Profile.Attributes, scenarioID, state.encounterModifiers)
	if err != nil {
		return playerCombatScoreBreakdown{}, err
	}
	score := weighted + modifiers.RoleTagModifier + modifiers.WeaponModifier + modifiers.PostureModifier + modifiers.VisibilityModifier + modifiers.TeamSupportModifier - modifiers.StaminaPenalty - modifiers.DamagePenalty - modifiers.SuppressionPenalty
	return playerCombatScoreBreakdown{
		WeightedPlayerScore: weighted, combatModifierInput: modifiers, PlayerCombatScore: score,
		Reasons: []reasonRecord{
			{Code: "WEIGHTED_PLAYER_SCORE", Source: playerID, Value: weighted, Weight: 1},
			{Code: "ROLE_TAG", Source: playerID, Value: modifiers.RoleTagModifier, Weight: 1},
			{Code: "WEAPON", Source: player.Weapon.Primary, Value: modifiers.WeaponModifier, Weight: 1},
			{Code: "POSTURE", Source: string(player.Posture), Value: modifiers.PostureModifier, Weight: 1},
			{Code: "VISIBILITY", Source: scenarioID, Value: modifiers.VisibilityModifier, Weight: 1},
			{Code: "TEAM_SUPPORT", Source: player.Side, Value: modifiers.TeamSupportModifier, Weight: 1},
			{Code: "STAMINA_PENALTY", Source: playerID, Value: -modifiers.StaminaPenalty, Weight: 1},
			{Code: "DAMAGE_PENALTY", Source: playerID, Value: -modifiers.DamagePenalty, Weight: 1},
			{Code: "SUPPRESSION_PENALTY", Source: playerID, Value: -modifiers.SuppressionPenalty, Weight: 1},
		},
	}, nil
}

func weightedScenarioPlayerScore(attributes PlayerAttributes, scenarioID string, modifiers map[string]EncounterModifier) (float64, error) {
	weights := map[string]int{}
	for _, modifier := range modifiers {
		if modifier.ScenarioID == scenarioID && modifier.Factor == "ScenarioWeight" && (modifier.Side == "Both" || modifier.Side == "") {
			if _, duplicate := weights[modifier.Attribute]; duplicate {
				return 0, newError("CONFIG_BAD_WEIGHT_SUM", "scenario %s duplicates weight %s", scenarioID, modifier.Attribute)
			}
			weights[modifier.Attribute] = modifier.Weight
		}
	}
	if len(weights) != len(scenarioWeightAttributes) {
		return 0, newError("CONFIG_BAD_WEIGHT_SUM", "scenario %s does not define ten weights", scenarioID)
	}
	totalWeight := 0
	weighted := 0.0
	for _, attribute := range scenarioWeightAttributes {
		weight, ok := weights[attribute]
		if !ok {
			return 0, newError("CONFIG_BAD_WEIGHT_SUM", "scenario %s misses weight %s", scenarioID, attribute)
		}
		totalWeight += weight
		weighted += float64(playerAttribute(attributes, attribute) * weight)
	}
	if totalWeight != 100 {
		return 0, newError("CONFIG_BAD_WEIGHT_SUM", "scenario %s weights sum to %d", scenarioID, totalWeight)
	}
	return weighted / 100, nil
}

func calculateTargetSurvivalScore(player *roundPlayerState, modifiers survivalModifierInput) targetSurvivalBreakdown {
	positioning := float64(player.Profile.Attributes.Positioning)
	reaction := float64(player.Profile.Attributes.Reaction)
	focus := float64(player.Focus)
	score := positioning + reaction + modifiers.CoverModifier + modifiers.TeamSupportModifier + focus - modifiers.MovementExposure - modifiers.DamagePenalty
	return targetSurvivalBreakdown{Positioning: positioning, Reaction: reaction, Focus: focus, survivalModifierInput: modifiers, TargetSurvivalScore: score}
}

func calculateCoordination(state *roundState, encounter *encounterState, side string) coordinationMetrics {
	participants := make([]*roundPlayerState, 0)
	locations := map[string]int{}
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.Side != side {
			continue
		}
		participants = append(participants, player)
		locations[coordinationLocationKey(player.Location)]++
	}
	metrics := coordinationMetrics{}
	if len(participants) >= 2 {
		metrics.TradeCoverage = float64(len(participants)-1) * 2
		metrics.Spacing = float64(len(locations)-1) * 1.5
		if len(locations) >= 2 {
			metrics.Crossfire = float64(len(locations)) * 2
		}
		syncCount := 0
		for _, player := range participants {
			if player.Posture == postureMoving || player.Posture == postureEngaged {
				syncCount++
			}
		}
		if syncCount >= 2 {
			metrics.SyncPeek = float64(syncCount) * 1.5
		}
	} else if len(participants) == 1 {
		metrics.Isolation = 6
	}
	return metrics
}

func calculateEncounterScorePair(state *roundState, encounter *encounterState, scenarioID string, seed int64) (map[string]encounterScoreBreakdown, error) {
	scenario := state.scenarios[scenarioID]
	if scenario.ID == "" {
		return nil, newError("INVALID_ENCOUNTER", "scenario %s is missing", scenarioID)
	}
	results := map[string]encounterScoreBreakdown{}
	for _, side := range []string{SideT, SideCT} {
		coordination := calculateCoordination(state, encounter, side)
		playerSum := 0.0
		for _, actorID := range encounter.ActorIDs {
			player := state.Players[actorID]
			if player == nil || !player.Alive || player.Side != side {
				continue
			}
			breakdown, err := calculatePlayerCombatScore(state, actorID, scenarioID, defaultCombatModifiers(state, player, scenario, coordination))
			if err != nil {
				return nil, err
			}
			playerSum += breakdown.PlayerCombatScore
		}
		teamModifier := coordination.Total()
		scenarioModifier := float64(scenario.BaseWeight) / 5
		utilityModifier := scopedUtilityModifier(state, side, utilityOpeningInitiative, "", encounter.ID, state.Timeline)
		momentumModifier := float64(state.MomentumT)
		if side == SideCT {
			momentumModifier = float64(state.MomentumCT)
		}
		momentumModifier /= 10
		timePressure := encounterTimePressure(state, side)
		deterministic := playerSum + teamModifier + scenarioModifier + utilityModifier + momentumModifier + timePressure
		noise := (stableUnit(seed, side, "encounter_noise")*2 - 1) * strategyNoiseAmplitude(teamInputFromState(state, side), state.constants)
		results[side] = encounterScoreBreakdown{
			Side: side, PlayerScoreSum: playerSum, TeamModifier: teamModifier, ScenarioModifier: scenarioModifier,
			UtilityModifier: utilityModifier, MomentumModifier: momentumModifier, TimePressureModifier: timePressure,
			BoundedRandomNoise: noise, DeterministicEncounterScore: deterministic,
			Reasons: []reasonRecord{
				{Code: "PLAYER_COMBAT_SUM", Source: encounter.ID, Value: playerSum, Weight: 1},
				{Code: "TEAM_COORDINATION", Source: side, Value: teamModifier, Weight: 1},
				{Code: "SCENARIO", Source: scenarioID, Value: scenarioModifier, Weight: 1},
				{Code: "UTILITY_ONCE", Source: encounter.ID, Value: utilityModifier, Weight: 1},
				{Code: "MOMENTUM_ONCE", Source: side, Value: momentumModifier, Weight: 1},
				{Code: "TIME_PRESSURE_ONCE", Source: side, Value: timePressure, Weight: 1},
			},
		}
	}
	protectEncounterTrend(results, state.constants.Float("DecisiveScoreGap", 0))
	for _, side := range []string{SideT, SideCT} {
		result := results[side]
		result.FinalScore = result.DeterministicEncounterScore + result.BoundedRandomNoise
		result.Reasons = append(result.Reasons, reasonRecord{Code: "BOUNDED_RANDOM_NOISE", Source: encounter.ID, Value: result.BoundedRandomNoise, Weight: 1})
		if math.Abs(results[SideT].DeterministicEncounterScore-results[SideCT].DeterministicEncounterScore) <= state.constants.Float("CloseScoreGap", 0) {
			result.Reasons = append(result.Reasons, reasonRecord{Code: "CLOSE_SCORE", Source: encounter.ID, Value: 1, Weight: 1})
		}
		results[side] = result
	}
	return results, nil
}

func createCombatPulseSnapshot(state *roundState, encounter *encounterState) (combatPulseSnapshot, error) {
	if encounter == nil || encounter.Status != encounterActive {
		return combatPulseSnapshot{}, newError("INVALID_COMBAT_PULSE", "pulse requires active encounter")
	}
	snapshot := combatPulseSnapshot{EncounterID: encounter.ID, Scenario: state.scenarios[encounter.ScenarioID], Timeline: state.Timeline, Players: map[string]pulsePlayerSnapshot{}, VisibleTargets: map[string]map[string]bool{}}
	for _, actorID := range encounter.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.EngagementID != encounter.ID {
			continue
		}
		snapshot.ActorIDs = append(snapshot.ActorIDs, actorID)
		x, y := runtimeXY(state, player.Location)
		snapshot.Players[actorID] = pulsePlayerSnapshot{
			PlayerID: actorID, TeamID: player.TeamID, Side: player.Side, Profile: player.Profile, Weapon: cloneLoadout(player.Weapon),
			WeaponSpec: state.weaponSpecs[player.Weapon.Primary], Alive: player.Alive, HP: player.HP, Stamina: player.Stamina, Focus: player.Focus,
			Suppressed: player.Suppressed, Posture: player.Posture, Location: clonePlayerLocation(player.Location), X: x, Y: y,
		}
	}
	sort.Strings(snapshot.ActorIDs)
	for _, actorID := range snapshot.ActorIDs {
		snapshot.VisibleTargets[actorID] = map[string]bool{}
		for _, targetID := range snapshot.ActorIDs {
			if snapshot.Players[actorID].Side == snapshot.Players[targetID].Side {
				continue
			}
			snapshot.VisibleTargets[actorID][targetID] = visibilityForLocations(state, snapshot.Players[actorID].Location, snapshot.Players[targetID].Location).Visible
		}
	}
	return snapshot, nil
}

func selectCombatTarget(snapshot combatPulseSnapshot, actorID string, pulseSeed int64, constants CombatConstants) (targetCandidate, bool) {
	actor, ok := snapshot.Players[actorID]
	if !ok || !actor.Alive {
		return targetCandidate{}, false
	}
	var candidates []targetCandidate
	for _, targetID := range snapshot.ActorIDs {
		target := snapshot.Players[targetID]
		if !target.Alive || target.Side == actor.Side || snapshot.VisibleTargets != nil && snapshot.VisibleTargets[actorID] != nil && !snapshot.VisibleTargets[actorID][targetID] {
			continue
		}
		candidate := targetCandidate{
			PlayerID: targetID, ThreatScore: snapshotThreatScore(target), Exposure: snapshotExposure(target),
			Distance: math.Hypot(actor.X-target.X, actor.Y-target.Y), HP: target.HP,
			TieRoll: stableUnit(pulseSeed, actorID, targetID, "target"),
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return targetCandidate{}, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].ThreatScore != candidates[j].ThreatScore {
			return candidates[i].ThreatScore > candidates[j].ThreatScore
		}
		if candidates[i].Exposure != candidates[j].Exposure {
			return candidates[i].Exposure > candidates[j].Exposure
		}
		if candidates[i].Distance != candidates[j].Distance {
			return candidates[i].Distance < candidates[j].Distance
		}
		if candidates[i].HP != candidates[j].HP {
			return candidates[i].HP < candidates[j].HP
		}
		return candidates[i].PlayerID < candidates[j].PlayerID
	})
	closeGap := constants.Float("CloseScoreGap", 0)
	closeCount := 1
	for closeCount < len(candidates) && candidates[0].ThreatScore-candidates[closeCount].ThreatScore <= closeGap {
		closeCount++
	}
	if closeCount > 1 {
		close := candidates[:closeCount]
		sort.SliceStable(close, func(i, j int) bool {
			if close[i].TieRoll != close[j].TieRoll {
				return close[i].TieRoll > close[j].TieRoll
			}
			return close[i].PlayerID < close[j].PlayerID
		})
		return close[0], true
	}
	return candidates[0], true
}

func calculateAttackProbabilities(combatScore, survivalScore float64, weapon WeaponSpec, targetArmor bool, scenario Scenario, visibility Visibility, constants CombatConstants) attackProbabilities {
	combatScale := constants.Float("CombatScale", 1)
	rawHit := 1 / (1 + math.Exp(-(combatScore-survivalScore)/combatScale))
	hit := clampFloat(rawHit, constants.Float("MinHitChance", 0), constants.Float("MaxHitChance", 1))
	rangeFactor := weaponRangeFactor(weapon, scenario.Range)
	armorFactor := 1.0
	if targetArmor {
		armorFactor = 0.5 + clampProbability(weapon.ArmorPenetration)*0.5
	}
	baseShotDamage := float64(weapon.Damage) * rangeFactor * armorFactor
	burst := clampInt(int(math.Floor(float64(weapon.RoundsPerMinute)/60*float64(constants.Int("PulseFireWindow", 1)))), 1, maxInt(1, weapon.MagazineSize))
	burstPotential := 1 + math.Log2(float64(burst))*0.25
	damagePotential := clampFloat(baseShotDamage*burstPotential/float64(constants.Int("MaxHP", 100)), constants.Float("MinDamagePotential", 0), constants.Float("MaxDamagePotential", 1))
	exposure := clampFloat(1+float64(visibility.ExposureModifier)/100, constants.Float("MinExposureModifier", 0), constants.Float("MaxExposureModifier", 2))
	rawKill := hit * damagePotential * exposure
	kill := clampFloat(math.Min(rawKill, hit), 0, constants.Float("MaxKillChance", 1))
	return attackProbabilities{RawHitChance: rawHit, HitChance: hit, BaseShotDamage: baseShotDamage, BurstCapacity: burst, DamagePotential: damagePotential, ExposureModifier: exposure, RawKillChance: rawKill, KillChance: kill}
}

func resolveAttackWindow(state *roundState, snapshot combatPulseSnapshot, action scheduledAction, actorID string, target targetCandidate, pulseSeed int64, ordinal int) (attackWindowResult, error) {
	actorSnapshot, targetSnapshot := snapshot.Players[actorID], snapshot.Players[target.PlayerID]
	actorState, targetState := state.Players[actorID], state.Players[target.PlayerID]
	coordination := calculateCoordination(state, state.ActiveEngagements[snapshot.EncounterID], actorSnapshot.Side)
	combatBreakdown, err := calculatePlayerCombatScore(state, actorID, snapshot.Scenario.ID, defaultCombatModifiers(state, actorState, snapshot.Scenario, coordination))
	if err != nil {
		return attackWindowResult{}, err
	}
	survival := calculateTargetSurvivalScore(targetState, defaultSurvivalModifiers(state, targetState, snapshot.Scenario, calculateCoordination(state, state.ActiveEngagements[snapshot.EncounterID], targetSnapshot.Side)))
	visibility := visibilityForLocations(state, actorSnapshot.Location, targetSnapshot.Location)
	probabilities := calculateAttackProbabilities(combatBreakdown.PlayerCombatScore, survival.TargetSurvivalScore, actorSnapshot.WeaponSpec, targetSnapshot.Weapon.Armor, snapshot.Scenario, visibility, state.constants)
	hitRoll := stableUnit(pulseSeed, actorID, target.PlayerID, "hit")
	lethalRoll := stableUnit(pulseSeed, actorID, target.PlayerID, "lethal")
	damageRoll := stableUnit(pulseSeed, actorID, target.PlayerID, "damage")
	result := attackWindowResult{ActorID: actorID, TargetID: target.PlayerID, HitRoll: hitRoll, LethalRoll: lethalRoll, DamageRoll: damageRoll, Probabilities: probabilities}
	if hitRoll >= probabilities.HitChance {
		result.Effect = effect{ID: newEffectID(state.Seed, action.ID, effectMiss, ordinal), SourceActionID: action.ID, Type: effectMiss, Priority: priorityCombatPulseCommit, Timestamp: state.Timeline, ActorID: actorID, TargetID: target.PlayerID, StringValue: actorSnapshot.WeaponSpec.ID}
		return result, nil
	}
	result.Hit = true
	conditionalLethal := 0.0
	if probabilities.HitChance > 0 {
		conditionalLethal = probabilities.KillChance / probabilities.HitChance
	}
	result.LethalWindow = lethalRoll < conditionalLethal
	damage := maxInt(1, int(math.Round(probabilities.BaseShotDamage*(0.5+damageRoll)*math.Min(2, 1+float64(probabilities.BurstCapacity-1)*0.08))))
	if result.LethalWindow {
		damage = maxInt(damage, targetSnapshot.HP)
	}
	result.Effect = effect{
		ID: newEffectID(state.Seed, action.ID, effectDamage, ordinal), SourceActionID: action.ID, Type: effectDamage,
		Priority: priorityCombatPulseCommit, Timestamp: state.Timeline, ActorID: actorID, TargetID: target.PlayerID,
		Amount: damage, StringValue: actorSnapshot.WeaponSpec.ID,
		ReasonRecords: []reasonRecord{
			{Code: "HIT_CHANCE", Source: actorID, Value: probabilities.HitChance, Weight: 1, Probability: floatPointer(probabilities.HitChance), Formula: "Hit = HitRoll < HitChance", Inputs: map[string]float64{"hit_roll": hitRoll, "hit_chance": probabilities.HitChance, "kill_chance": probabilities.KillChance}},
			{Code: "KILL_CHANCE", Source: target.PlayerID, Value: probabilities.KillChance, Weight: 1},
			{Code: "WEAPON_DAMAGE", Source: actorSnapshot.WeaponSpec.ID, Value: probabilities.BaseShotDamage, Weight: 1},
		},
	}
	return result, nil
}

func floatPointer(value float64) *float64 { return &value }

func resolveCombatPulse(state *roundState, action scheduledAction) (*combatPulseResult, error) {
	if state == nil || action.Type != actionCombatPulse || action.ResolveAt != state.Timeline {
		return nil, newError("INVALID_COMBAT_PULSE", "invalid combat pulse action")
	}
	encounter := state.ActiveEngagements[action.Payload.TargetID]
	if encounter == nil || encounter.Status != encounterActive {
		return nil, newError("INVALID_COMBAT_PULSE", "encounter is not active")
	}
	snapshot, err := createCombatPulseSnapshot(state, encounter)
	if err != nil {
		return nil, err
	}
	pulseIndex := encounter.PulsesResolved
	pulseSeed := deriveSeed(state.Seed, "encounter", encounter.ID, pulseIndex, action.ResolveAt)
	result := &combatPulseResult{EncounterID: encounter.ID, PulseIndex: pulseIndex, Snapshot: snapshot}
	damageEffects := make([]effect, 0, len(snapshot.ActorIDs))
	for ordinal, actorID := range snapshot.ActorIDs {
		target, ok := selectCombatTarget(snapshot, actorID, pulseSeed, state.constants)
		if !ok {
			continue
		}
		attack, err := resolveAttackWindow(state, snapshot, action, actorID, target, pulseSeed, ordinal)
		if err != nil {
			return nil, err
		}
		result.Attacks = append(result.Attacks, attack)
		if attack.Effect.Type == effectDamage {
			damageEffects = append(damageEffects, attack.Effect)
		}
	}
	batch, err := applyCombatPulseCommit(state, action, damageEffects)
	if err != nil {
		return nil, err
	}
	missCount := 0
	for _, attack := range result.Attacks {
		if attack.Effect.Type == effectMiss {
			missCount++
		}
	}
	if err := state.ValidateEffectBatchSize(len(batch.Effects) + missCount); err != nil {
		return nil, err
	}
	for _, attack := range result.Attacks {
		if attack.Effect.Type == effectMiss {
			batch.Effects = append(batch.Effects, appliedEffect{Effect: attack.Effect})
		}
	}
	applyCombatAftermath(state, snapshot, result.Attacks, batch)
	encounter.PulsesResolved++
	result.Batch = batch
	result.ShouldEnd, result.EndReason = encounterShouldEnd(state, encounter)
	return result, nil
}

func defaultCombatModifiers(state *roundState, player *roundPlayerState, scenario Scenario, coordination coordinationMetrics) combatModifierInput {
	weapon := state.weaponSpecs[player.Weapon.Primary]
	role := 0.0
	if hasFold(player.Profile.RoleTags, "Entry") && scenario.Phase == "SiteEntry" {
		role += 5
	}
	if hasFold(player.Profile.RoleTags, "Support") {
		role += 2
	}
	posture := map[combatPosture]float64{postureHolding: 6, postureEngaged: 3, postureMoving: -5, postureRetaking: 2}[player.Posture]
	visibilityModifier := 0.0
	if player.Location.Edge != nil {
		visibilityModifier -= 4
	}
	configuredPosture, configuredVisibility, configuredSupport := configuredCombatTerms(state, player, scenario)
	return combatModifierInput{
		RoleTagModifier: role, WeaponModifier: float64(weapon.Damage-30)/2 + float64(weapon.RoundsPerMinute)/600,
		PostureModifier: posture + configuredPosture, VisibilityModifier: visibilityModifier + configuredVisibility, TeamSupportModifier: coordination.Total() + configuredSupport,
		StaminaPenalty:     float64(state.constants.Int("MaxStamina", 100)-player.Stamina) / 10,
		DamagePenalty:      float64(state.constants.Int("MaxHP", 100)-player.HP) / 10,
		SuppressionPenalty: boolFloat(player.Suppressed) * 8,
	}
}

func configuredCombatTerms(state *roundState, player *roundPlayerState, scenario Scenario) (posture, visibility, support float64) {
	tagIDs := append([]string(nil), scenario.MapTagIDs...)
	sort.Strings(tagIDs)
	for _, tagID := range tagIDs {
		tag := state.mapTags[tagID]
		if tag.ID == "" || tag.Side != player.Side {
			continue
		}
		switch tag.Category {
		case "Angle", "Posture", "Timing":
			posture += float64(tag.Weight)
		case "Risk":
			visibility -= float64(tag.Weight)
		case "Site":
			support += float64(tag.Weight)
		}
	}
	modifierIDs := make([]string, 0, len(state.encounterModifiers))
	for id := range state.encounterModifiers {
		modifierIDs = append(modifierIDs, id)
	}
	sort.Strings(modifierIDs)
	for _, id := range modifierIDs {
		modifier := state.encounterModifiers[id]
		if modifier.ScenarioID != scenario.ID || modifier.Factor == "ScenarioWeight" || modifier.Side != "" && modifier.Side != "Both" && modifier.Side != player.Side {
			continue
		}
		attributeScale := 1.0
		if modifier.Attribute != "" {
			attributeScale = clampProbability(float64(playerAttribute(player.Profile.Attributes, modifier.Attribute)) / 100)
		}
		value := float64(modifier.Weight) * attributeScale
		switch modifier.Factor {
		case "CTHoldAngle":
			posture += value
		case "RotateRisk":
			visibility += value
		case "SyncPeek", "IntelControl", "SitePressure", "CloseExecute", "RetakeCover", "BombComposure":
			support += value
		}
	}
	return posture, visibility, support
}

func defaultSurvivalModifiers(state *roundState, player *roundPlayerState, _ Scenario, coordination coordinationMetrics) survivalModifierInput {
	cover := 0.0
	if player.Posture == postureHolding {
		cover = 8
	}
	movement := 0.0
	if player.Location.Edge != nil || player.Posture == postureMoving {
		movement = 10
	}
	return survivalModifierInput{CoverModifier: cover, TeamSupportModifier: coordination.Total(), MovementExposure: movement, DamagePenalty: float64(state.constants.Int("MaxHP", 100)-player.HP) / 10}
}

func protectEncounterTrend(results map[string]encounterScoreBreakdown, decisiveGap float64) {
	gap := math.Abs(results[SideT].DeterministicEncounterScore - results[SideCT].DeterministicEncounterScore)
	if gap < decisiveGap {
		return
	}
	bound := math.Max(0, (gap-0.000001)/2)
	for _, side := range []string{SideT, SideCT} {
		result := results[side]
		result.BoundedRandomNoise = clampFloat(result.BoundedRandomNoise, -bound, bound)
		results[side] = result
	}
}

func applyCombatAftermath(state *roundState, snapshot combatPulseSnapshot, attacks []attackWindowResult, batch *appliedBatch) {
	for _, attack := range attacks {
		if attacker := state.Players[attack.ActorID]; attacker != nil && attacker.Alive {
			attacker.Focus = clampInt(attacker.Focus-2, state.constants.Int("MinFocus", 0), state.constants.Int("MaxFocus", 100))
			attacker.Stamina = clampInt(attacker.Stamina-2, state.constants.Int("MinStamina", 0), state.constants.Int("MaxStamina", 100))
			attacker.Posture = postureEngaged
		}
		if attack.Hit {
			if target := state.Players[attack.TargetID]; target != nil && target.Alive {
				target.Focus = clampInt(target.Focus-8, state.constants.Int("MinFocus", 0), state.constants.Int("MaxFocus", 100))
				target.Suppressed = true
				target.Posture = postureHolding
			}
		}
	}
	for _, event := range batch.Events {
		if event.EventType != EventKill {
			continue
		}
		attacker, victim := state.Players[event.AttackerID], state.Players[event.VictimID]
		if attacker == nil || victim == nil {
			continue
		}
		if attacker.Side == SideT {
			state.MomentumT = clampInt(state.MomentumT+5, -100, 100)
			state.MomentumCT = clampInt(state.MomentumCT-5, -100, 100)
		} else {
			state.MomentumCT = clampInt(state.MomentumCT+5, -100, 100)
			state.MomentumT = clampInt(state.MomentumT-5, -100, 100)
		}
	}
	_ = snapshot
}

func teamInputFromState(state *roundState, side string) TeamInput {
	team := TeamInput{}
	ids := make([]string, 0)
	for id, player := range state.Players {
		if player.Side == side {
			ids = append(ids, id)
			team.TeamID = player.TeamID
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		team.Players = append(team.Players, state.Players[id].Profile)
	}
	return team
}

func encounterTimePressure(state *roundState, side string) float64 {
	remaining := state.RoundDeadline - state.Timeline
	threshold := state.constants.Int("ForceExecuteThreshold", 1)
	if remaining >= threshold {
		return 0
	}
	pressure := float64(threshold-remaining) / float64(maxInt(1, threshold)) * 5
	if side == SideT {
		return pressure
	}
	return -pressure
}

func coordinationLocationKey(location playerLocation) string {
	if location.NodeID != "" {
		return "node:" + location.NodeID
	}
	if location.Edge != nil {
		return "edge:" + location.Edge.EdgeID
	}
	return "unknown"
}

func snapshotThreatScore(player pulsePlayerSnapshot) float64 {
	attributes := player.Profile.Attributes
	return float64(attributes.Firepower+attributes.Aim+attributes.Reaction+attributes.Awareness) / 4
}

func snapshotExposure(player pulsePlayerSnapshot) float64 {
	exposure := 0.0
	if player.Location.Edge != nil {
		exposure += 20
	}
	if player.Posture == postureMoving {
		exposure += 15
	}
	if player.Posture == postureHolding {
		exposure -= 10
	}
	return exposure
}

func runtimeXY(state *roundState, location playerLocation) (float64, float64) {
	if location.Edge != nil {
		return location.Edge.X, location.Edge.Y
	}
	if node := state.Nodes[location.NodeID]; node != nil {
		return node.Node.X, node.Node.Y
	}
	return 0, 0
}

func visibilityForLocations(state *roundState, from, to playerLocation) Visibility {
	fromNode, toNode := projectedNodeID(from), projectedNodeID(to)
	ids := make([]string, 0, len(state.visibility))
	for id := range state.visibility {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		visibility := state.visibility[id]
		if visibility.FromNode == fromNode && visibility.ToNode == toNode || visibility.FromNode == toNode && visibility.ToNode == fromNode {
			return visibility
		}
	}
	return Visibility{Visible: true}
}

func weaponRangeFactor(weapon WeaponSpec, scenarioRange string) float64 {
	switch scenarioRange {
	case "Long":
		return weapon.RangeModifier
	case "Medium":
		return math.Sqrt(maxFloat(0, weapon.RangeModifier))
	default:
		return 1
	}
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
