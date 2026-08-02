package matchengine

import "sort"

const (
	PriorityCombatPulseCommit = 100
	PriorityCombatAftermath   = 90
	PriorityBombDrop          = 80
)

type damageContribution struct {
	effect  Effect
	applied int
}

// ApplyCombatPulseCommit atomically applies every damage effect produced from
// one immutable combat-pulse snapshot. Death, kill attribution, bomb drop and
// external-action interruption are derived only after all HP changes commit.
func ApplyCombatPulseCommit(state *RoundState, action ScheduledAction, effects []Effect) (*AppliedBatch, error) {
	if state == nil || action.ID == "" || action.Type != ActionCombatPulse || action.Priority != PriorityCombatPulseCommit {
		return nil, newError("INVALID_EFFECT_BATCH", "combat pulse commit requires a priority-100 CombatPulse action")
	}
	if action.ResolveAt != state.Timeline {
		return nil, newError("INVALID_EFFECT_BATCH", "combat pulse resolve time %d does not match timeline %d", action.ResolveAt, state.Timeline)
	}
	if err := state.ValidateEffectBatchSize(len(effects)); err != nil {
		return nil, err
	}

	normalized := append([]Effect(nil), effects...)
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].TargetID != normalized[j].TargetID {
			return normalized[i].TargetID < normalized[j].TargetID
		}
		return normalized[i].ID < normalized[j].ID
	})
	byTarget := make(map[string][]Effect)
	for _, effect := range normalized {
		if effect.ID == "" || effect.SourceActionID != action.ID || effect.Type != EffectDamage || effect.Timestamp != state.Timeline || effect.Amount < 0 {
			return nil, newError("INVALID_EFFECT_BATCH", "combat pulse contains an invalid damage effect")
		}
		if state.Players[effect.ActorID] == nil || state.Players[effect.TargetID] == nil {
			return nil, newError("INVALID_EFFECT_BATCH", "damage effect references an unknown player")
		}
		byTarget[effect.TargetID] = append(byTarget[effect.TargetID], effect)
	}

	targetIDs := make([]string, 0, len(byTarget))
	for targetID := range byTarget {
		targetIDs = append(targetIDs, targetID)
	}
	sort.Strings(targetIDs)
	derivedCount := 0
	for _, targetID := range targetIDs {
		target := state.Players[targetID]
		if !target.Alive || target.HP <= 0 {
			continue
		}
		totalRaw := 0
		for _, effect := range byTarget[targetID] {
			totalRaw += effect.Amount
		}
		if totalRaw >= target.HP {
			derivedCount++
			if target.HasBomb {
				derivedCount++
			}
		}
	}
	if err := state.ValidateEffectBatchSize(len(effects) + derivedCount); err != nil {
		return nil, err
	}

	batch := &AppliedBatch{Timestamp: state.Timeline}
	contributions := make(map[string][]damageContribution, len(targetIDs))
	for _, targetID := range targetIDs {
		target := state.Players[targetID]
		if !target.Alive || target.HP <= 0 {
			continue
		}
		allocated := allocateActualDamage(target.HP, byTarget[targetID])
		totalApplied := 0
		for _, contribution := range allocated {
			totalApplied += contribution.applied
			attacker := state.Players[contribution.effect.ActorID]
			attacker.Damage += contribution.applied
			batch.Effects = append(batch.Effects, AppliedEffect{Effect: contribution.effect, AppliedAmount: contribution.applied})
			if contribution.applied > 0 {
				batch.Events = append(batch.Events, damageEvent(state, action, contribution))
			}
		}
		target.HP -= totalApplied
		contributions[targetID] = allocated
	}

	deathOrdinal := 0
	bombOrdinal := 0
	firstKill := true
	lastKillSide := ""
	lastKillAt := -1000
	for _, event := range state.Events {
		if event.EventType == EventKill {
			firstKill = false
			lastKillAt = int(event.Timestamp)
			if attacker := state.Players[event.AttackerID]; attacker != nil {
				lastKillSide = attacker.Side
			}
		}
	}
	for _, targetID := range targetIDs {
		target := state.Players[targetID]
		if target.HP > 0 || !target.Alive {
			if hasAppliedDamage(contributions[targetID]) {
				interruptExternalAction(state, target)
			}
			continue
		}
		killer := selectKillContribution(contributions[targetID])
		if killer == nil {
			continue
		}
		target.Alive = false
		target.Deaths++
		state.Players[killer.effect.ActorID].Kills++
		cancelCurrentAction(state, target)

		death := Effect{
			ID: NewEffectID(actionSeed(state), action.ID, EffectDeath, deathOrdinal), SourceActionID: action.ID,
			Type: EffectDeath, Priority: PriorityCombatPulseCommit, Timestamp: state.Timeline,
			ActorID: killer.effect.ActorID, TargetID: targetID,
			StringValue: killer.effect.StringValue, ReasonRecords: append([]ReasonRecord(nil), killer.effect.ReasonRecords...),
		}
		deathOrdinal++
		batch.Effects = append(batch.Effects, AppliedEffect{Effect: death, AppliedAmount: 1})
		killerSide := state.Players[death.ActorID].Side
		isTrade := lastKillSide != "" && lastKillSide != killerSide && state.Timeline-lastKillAt <= 5
		batch.Events = append(batch.Events, killEvent(state, action, death, firstKill, isTrade))
		firstKill = false
		lastKillSide, lastKillAt = killerSide, state.Timeline

		if target.HasBomb {
			bombDrop := Effect{
				ID: NewEffectID(actionSeed(state), action.ID, EffectBombDrop, bombOrdinal), SourceActionID: action.ID,
				Type: EffectBombDrop, Priority: PriorityBombDrop, Timestamp: state.Timeline,
				ActorID: targetID, TargetID: targetID, NodeID: projectedNodeID(target.Location),
			}
			bombOrdinal++
			if err := state.Bomb.Drop(target.Location, state.Timeline); err != nil {
				return nil, err
			}
			target.HasBomb = false
			batch.Effects = append(batch.Effects, AppliedEffect{Effect: bombDrop, AppliedAmount: 1})
			batch.Events = append(batch.Events, bombDropEvent(state, action, bombDrop))
		}
	}
	sortEvents(batch.Events)
	state.Events = append(state.Events, batch.Events...)
	if err := state.ClampAndValidate(); err != nil {
		return nil, err
	}
	return batch, nil
}

func allocateActualDamage(hp int, effects []Effect) []damageContribution {
	ordered := append([]Effect(nil), effects...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Amount != ordered[j].Amount {
			return ordered[i].Amount > ordered[j].Amount
		}
		return ordered[i].ID < ordered[j].ID
	})
	totalRaw := 0
	for _, effect := range ordered {
		totalRaw += effect.Amount
	}
	actualLoss := minDamageInt(hp, totalRaw)
	out := make([]damageContribution, len(ordered))
	allocated := 0
	if totalRaw > 0 {
		for index, effect := range ordered {
			share := actualLoss * effect.Amount / totalRaw
			out[index] = damageContribution{effect: effect, applied: share}
			allocated += share
		}
	}
	for index := 0; allocated < actualLoss && index < len(out); index++ {
		out[index].applied++
		allocated++
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].effect.ID < out[j].effect.ID })
	return out
}

func selectKillContribution(contributions []damageContribution) *damageContribution {
	eligible := append([]damageContribution(nil), contributions...)
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].applied != eligible[j].applied {
			return eligible[i].applied > eligible[j].applied
		}
		return eligible[i].effect.ID < eligible[j].effect.ID
	})
	if len(eligible) == 0 || eligible[0].applied <= 0 {
		return nil
	}
	return &eligible[0]
}

func hasAppliedDamage(contributions []damageContribution) bool {
	for _, contribution := range contributions {
		if contribution.applied > 0 {
			return true
		}
	}
	return false
}

func interruptExternalAction(state *RoundState, player *RoundPlayerState) {
	player.Action.Version++
	switch player.Action.Status {
	case ActionMoving, ActionPlanting, ActionDefusing:
		interruptBombActionForPlayer(state, player)
		player.Action.CurrentActionID = ""
		player.Action.Status = ActionIdle
		player.Action.BusyUntil = state.Timeline
		player.Action.Busy = BusyInterval{}
	}
}

func cancelCurrentAction(state *RoundState, player *RoundPlayerState) {
	interruptBombActionForPlayer(state, player)
	player.Action.Version++
	player.Action.CurrentActionID = ""
	player.Action.Status = ActionIdle
	player.Action.BusyUntil = state.Timeline
	player.Action.Busy = BusyInterval{}
}

func interruptBombActionForPlayer(state *RoundState, player *RoundPlayerState) {
	if state == nil || player == nil || player.Action.CurrentActionID == "" {
		return
	}
	if player.Action.Status == ActionPlanting {
		actionID := player.Action.CurrentActionID
		if state.Bomb.InterruptPlant(player.Profile.PlayerID, actionID) {
			action := ScheduledAction{ID: actionID, Type: ActionPlantComplete, ActorIDs: []string{player.Profile.PlayerID}, Priority: PriorityPlantComplete}
			event, _ := newActionLifecycleEvent(state, action, EventPlantInterrupt, "bomb plant interrupted", 0)
			event.VictimID, event.VictimName, event.VictimTeamID = player.Profile.PlayerID, player.Profile.DisplayName, player.TeamID
			event.Bomb, event.State = projectBombState(state.Bomb), snapshotForEvent(state)
			state.Events = append(state.Events, event)
		}
	}
	if player.Action.Status == ActionDefusing {
		actionID := player.Action.CurrentActionID
		if state.Bomb.InterruptDefuse(player.Profile.PlayerID, actionID) {
			action := ScheduledAction{ID: actionID, Type: ActionDefuseComplete, ActorIDs: []string{player.Profile.PlayerID}, Priority: PriorityDefuseComplete}
			event, _ := newActionLifecycleEvent(state, action, EventDefuseInterrupt, "bomb defuse interrupted", 0)
			event.VictimID, event.VictimName, event.VictimTeamID = player.Profile.PlayerID, player.Profile.DisplayName, player.TeamID
			event.Bomb, event.State = projectBombState(state.Bomb), snapshotForEvent(state)
			state.Events = append(state.Events, event)
		}
	}
}

func damageEvent(state *RoundState, action ScheduledAction, contribution damageContribution) *GameEvent {
	attacker := state.Players[contribution.effect.ActorID]
	victim := state.Players[contribution.effect.TargetID]
	eventID := NewEventID(actionSeed(state), action.ID, contribution.effect.ID, EventDamage, 0)
	reasonRecord := firstEffectReason(contribution.effect, "DAMAGE_APPLIED")
	reasonRecord.Formula = "AppliedDamage = min(TargetHP, RawDamageShare)"
	reasonRecord.Inputs = map[string]float64{"raw_damage": float64(contribution.effect.Amount), "applied_damage": float64(contribution.applied)}
	reasonRecord.StateChanges = []ReasonStateChange{{Field: "player.hp", Before: NumberReasonValue(float64(victim.HP + contribution.applied)), After: NumberReasonValue(float64(victim.HP))}}
	reason, _ := ProjectReasonRecord(reasonRecord, action.ID, contribution.effect.ID)
	event := &GameEvent{
		EventID: eventID, SourceActionID: action.ID, SourceEffectID: contribution.effect.ID,
		Timestamp: int64(state.Timeline), EventType: EventDamage, AttackerID: attacker.Profile.PlayerID, AttackerName: attacker.Profile.DisplayName,
		AttackerTeamID: attacker.TeamID, VictimID: victim.Profile.PlayerID, VictimName: victim.Profile.DisplayName, VictimTeamID: victim.TeamID,
		Weapon: contribution.effect.StringValue, Message: "damage applied", Location: eventLocation(state, victim.Location, eventID, contribution.effect.ID), Extra: map[string]interface{}{"damage": contribution.applied},
		Reason: reason, sortPriority: PriorityCombatPulseCommit, sortActionType: string(action.Type) + "/00Damage", sortMinActorID: action.MinActorID(),
	}
	event.State = snapshotForEvent(state)
	return event
}

func killEvent(state *RoundState, action ScheduledAction, death Effect, firstKill, trade bool) *GameEvent {
	attacker := state.Players[death.ActorID]
	victim := state.Players[death.TargetID]
	eventID := NewEventID(actionSeed(state), action.ID, death.ID, EventKill, 0)
	reasonRecord := firstEffectReason(death, "PLAYER_KILLED")
	reasonRecord.StateChanges = []ReasonStateChange{{Field: "player.alive", Before: BoolReasonValue(true), After: BoolReasonValue(false)}}
	reason, _ := ProjectReasonRecord(reasonRecord, action.ID, death.ID)
	event := &GameEvent{
		EventID: eventID, SourceActionID: action.ID, SourceEffectID: death.ID,
		Timestamp: int64(state.Timeline), EventType: EventKill, AttackerID: attacker.Profile.PlayerID, AttackerName: attacker.Profile.DisplayName,
		AttackerTeamID: attacker.TeamID, VictimID: victim.Profile.PlayerID, VictimName: victim.Profile.DisplayName, VictimTeamID: victim.TeamID,
		Weapon: death.StringValue, IsFirstKill: firstKill, IsTrade: trade, Message: "player killed", Location: eventLocation(state, victim.Location, eventID, death.ID),
		Reason: reason, Bomb: projectBombState(state.Bomb), sortPriority: PriorityCombatPulseCommit, sortActionType: string(action.Type) + "/10Kill", sortMinActorID: action.MinActorID(),
	}
	event.State = snapshotForEvent(state)
	return event
}

func projectEffectReason(effect Effect) *EventReason {
	if len(effect.ReasonRecords) == 0 {
		return nil
	}
	reason, _ := ProjectReasonRecord(effect.ReasonRecords[0], effect.SourceActionID, effect.ID)
	return reason
}

func firstEffectReason(effect Effect, fallback string) ReasonRecord {
	if len(effect.ReasonRecords) > 0 {
		return effect.ReasonRecords[0]
	}
	return ReasonRecord{Code: fallback, Source: string(effect.Type), Value: float64(effect.Amount), Weight: 1}
}

func bombDropEvent(state *RoundState, action ScheduledAction, drop Effect) *GameEvent {
	carrier := state.Players[drop.ActorID]
	eventID := NewEventID(actionSeed(state), action.ID, drop.ID, EventBombDrop, 0)
	reason, _ := ProjectReasonRecord(ReasonRecord{Code: "BOMB_CARRIER_KILLED", Source: carrier.Profile.PlayerID, Value: 1, Weight: 1, StateChanges: []ReasonStateChange{{Field: "bomb.status", Before: StringReasonValue(string(BombCarried)), After: StringReasonValue(string(BombDropped))}}}, action.ID, drop.ID)
	return &GameEvent{
		EventID: eventID, SourceActionID: action.ID, SourceEffectID: drop.ID,
		Timestamp: int64(state.Timeline), EventType: EventBombDrop, VictimID: carrier.Profile.PlayerID, VictimName: carrier.Profile.DisplayName,
		VictimTeamID: carrier.TeamID, Message: "bomb dropped", Location: eventLocation(state, carrier.Location, eventID, drop.ID), Bomb: projectBombState(state.Bomb), Reason: reason, State: snapshotForEvent(state),
		sortPriority: PriorityBombDrop, sortActionType: string(action.Type), sortMinActorID: action.MinActorID(),
	}
}

func actionSeed(state *RoundState) int64 {
	return state.Seed
}

func minDamageInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
