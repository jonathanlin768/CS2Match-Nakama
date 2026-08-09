package matchengine

import (
	"math"
	"sort"
)

type UtilityEffectType string

const (
	UtilityVisibilitySuppression UtilityEffectType = "VisibilitySuppression"
	UtilityOpeningInitiative     UtilityEffectType = "OpeningInitiative"
	UtilityExposureReduction     UtilityEffectType = "ExposureReduction"
	UtilitySyncPeekQuality       UtilityEffectType = "SyncPeekQuality"
	UtilityPlantCover            UtilityEffectType = "PlantCover"
	UtilityDefuseCover           UtilityEffectType = "DefuseCover"
)

type UtilityRequest struct {
	Type             UtilityEffectType
	ActorIDs         []string
	ScopeActionID    string
	ScopeEncounterID string
	StartAt          int
	Duration         int
	BaseCost         int
}

type UtilityWindow struct {
	ID               string
	Type             UtilityEffectType
	Side             string
	ActorIDs         []string
	ScopeActionID    string
	ScopeEncounterID string
	StartAt          int
	EndAt            int
	Cost             int
	Modifier         float64
}

type UtilitySpendResult struct {
	Applied bool
	Window  *UtilityWindow
	Reason  ReasonRecord
}

func InitializeUtilityBudget(state *RoundState, side string, template RouteTemplate) int {
	if state == nil || state.Utility[side] == nil {
		return 0
	}
	capBudget := state.constants.Int("UtilityBudget", 0)
	players, utilityTotal, supportCount, grenadeCount := 0, 0, 0, 0
	for _, player := range state.Players {
		if player.Side != side {
			continue
		}
		players++
		utilityTotal += player.Profile.Attributes.Utility
		grenadeCount += len(player.Weapon.Grenades)
		if hasFold(player.Profile.RoleTags, "Support") {
			supportCount++
		}
	}
	if players == 0 {
		state.Utility[side].Budget = 0
		return 0
	}
	averageUtility := float64(utilityTotal) / float64(players)
	budget := float64(capBudget)*0.35 + float64(capBudget)*0.35*averageUtility/100 + float64(capBudget)*0.2*math.Min(1, float64(grenadeCount)/float64(players*4)) + float64(capBudget)*0.02*float64(supportCount)
	if hasFold(template.RequiredRoles, "Support") {
		budget += float64(capBudget) * 0.05
	}
	state.Utility[side].Budget = clampInt(int(math.Round(budget)), 0, capBudget)
	state.Utility[side].Spent = 0
	state.Utility[side].Windows = nil
	return state.Utility[side].Budget
}

func SpendUtility(state *RoundState, side string, request UtilityRequest) (UtilitySpendResult, error) {
	teamUtility := state.Utility[side]
	if teamUtility == nil || request.BaseCost <= 0 || request.Duration <= 0 || request.StartAt < state.Timeline || request.StartAt+request.Duration > state.constants.Int("MaxRoundTimeline", 0) {
		return UtilitySpendResult{}, newError("INVALID_UTILITY", "utility request is invalid")
	}
	if request.ScopeActionID == "" && request.ScopeEncounterID == "" {
		return UtilitySpendResult{}, newError("INVALID_UTILITY", "utility must be scoped to an action or encounter")
	}
	actors := validUtilityActors(state, side, request.ActorIDs)
	if len(actors) == 0 {
		return UtilitySpendResult{}, newError("INVALID_UTILITY", "utility has no valid actors")
	}
	supportCount, utilityTotal := 0, 0
	for _, actorID := range actors {
		player := state.Players[actorID]
		utilityTotal += player.Profile.Attributes.Utility
		if hasFold(player.Profile.RoleTags, "Support") {
			supportCount++
		}
	}
	efficiency := 1 + float64(supportCount)*0.15
	cost := maxInt(1, int(math.Ceil(float64(request.BaseCost)/efficiency)))
	remaining := teamUtility.Budget - teamUtility.Spent
	if cost > remaining {
		return UtilitySpendResult{Reason: ReasonRecord{Code: "LOW_UTILITY", Source: side, Value: float64(remaining - cost), Weight: 1, Detail: "insufficient scoped UtilityBudget"}}, nil
	}
	averageUtility := float64(utilityTotal) / float64(len(actors))
	modifier := utilityBaseModifier(request.Type) * (0.5 + averageUtility/200)
	window := UtilityWindow{
		ID:   stableObjectID("util", state.Seed, side, string(request.Type), request.ScopeActionID, request.ScopeEncounterID, request.StartAt, request.Duration, teamUtility.Spent),
		Type: request.Type, Side: side, ActorIDs: actors, ScopeActionID: request.ScopeActionID, ScopeEncounterID: request.ScopeEncounterID,
		StartAt: request.StartAt, EndAt: request.StartAt + request.Duration, Cost: cost, Modifier: modifier,
	}
	teamUtility.Spent += cost
	teamUtility.Windows = append(teamUtility.Windows, window)
	sort.SliceStable(teamUtility.Windows, func(i, j int) bool { return teamUtility.Windows[i].ID < teamUtility.Windows[j].ID })
	copy := window
	copy.ActorIDs = append([]string(nil), window.ActorIDs...)
	return UtilitySpendResult{Applied: true, Window: &copy, Reason: ReasonRecord{Code: "UTILITY_SPENT", Source: window.ID, Value: -float64(cost), Weight: modifier}}, nil
}

func spendScopedUtility(state *RoundState, side string, effectType UtilityEffectType, actorIDs []string, actionID, encounterID string, duration, divisor int) (UtilitySpendResult, error) {
	if state == nil || len(validUtilityActors(state, side, actorIDs)) == 0 {
		return UtilitySpendResult{}, nil
	}
	remainingTimeline := state.constants.Int("MaxRoundTimeline", 0) - state.Timeline
	duration = minDamageInt(duration, remainingTimeline)
	if duration <= 0 {
		return UtilitySpendResult{}, nil
	}
	if divisor <= 0 {
		divisor = 10
	}
	return SpendUtility(state, side, UtilityRequest{
		Type: effectType, ActorIDs: actorIDs, ScopeActionID: actionID, ScopeEncounterID: encounterID,
		StartAt: state.Timeline, Duration: duration, BaseCost: maxInt(1, state.constants.Int("UtilityBudget", 0)/divisor),
	})
}

func addUtilityReason(event *GameEvent, result UtilitySpendResult) {
	if event == nil || event.Reason == nil || result.Reason.Code == "" {
		return
	}
	event.Reason.Modifiers = append(event.Reason.Modifiers, ReasonModifier{Code: result.Reason.Code, Value: result.Reason.Value, Detail: result.Reason.Detail})
}

func ScopedUtilityModifier(state *RoundState, side string, effectType UtilityEffectType, actionID, encounterID string, at int) float64 {
	teamUtility := state.Utility[side]
	if teamUtility == nil {
		return 0
	}
	total := 0.0
	for _, window := range teamUtility.Windows {
		if window.Type != effectType || at < window.StartAt || at >= window.EndAt {
			continue
		}
		if window.ScopeActionID != "" && window.ScopeActionID != actionID {
			continue
		}
		if window.ScopeEncounterID != "" && window.ScopeEncounterID != encounterID {
			continue
		}
		total += window.Modifier
	}
	return total
}

func utilityBaseModifier(effectType UtilityEffectType) float64 {
	switch effectType {
	case UtilityOpeningInitiative:
		return 8
	case UtilityVisibilitySuppression, UtilityExposureReduction, UtilitySyncPeekQuality, UtilityPlantCover, UtilityDefuseCover:
		return 0.2
	default:
		return 0
	}
}

func validUtilityActors(state *RoundState, side string, actorIDs []string) []string {
	var out []string
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player != nil && player.Alive && player.Side == side && len(player.Weapon.Grenades) > 0 {
			out = append(out, actorID)
		}
	}
	sort.Strings(out)
	return uniqueStrings(out)
}
