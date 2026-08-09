package matchengine

func newStrategyMemory() StrategyMemory {
	return StrategyMemory{PreviousSuccess: map[string]int{}, CounterReads: map[string]int{}, SideTendency: map[string]float64{}, TeamStyle: map[string]float64{}}
}

func cloneStrategyMemory(memory StrategyMemory) StrategyMemory {
	out := StrategyMemory{
		PreviousSuccess: copyStringIntMap(memory.PreviousSuccess), RecentTemplates: append([]string(nil), memory.RecentTemplates...),
		CounterReads: copyStringIntMap(memory.CounterReads), SideTendency: copyStringFloatMap(memory.SideTendency), TeamStyle: copyStringFloatMap(memory.TeamStyle),
	}
	return out
}

// UpdateStrategyMemoryFromRound accepts only a completed causal RoundResult;
// no in-progress or predicted outcome can enter the next round's planner.
func UpdateStrategyMemoryFromRound(memory StrategyMemory, round *RoundResult, teamID string, repeatWindow int) (StrategyMemory, error) {
	if round == nil || round.WinnerTeamID == "" || round.WinReason == "" || !roundHasEvent(round, EventRoundEnd) {
		return memory, newError("INCOMPLETE_ROUND_RESULT", "strategy memory requires a completed RoundResult")
	}
	out := cloneStrategyMemory(memory)
	if out.PreviousSuccess == nil {
		out = newStrategyMemory()
	}
	templateID := round.StrategyTemplateID
	if teamID == round.TeamCTID {
		templateID = round.CTSetupTemplateID
	}
	if templateID == "" {
		return out, nil
	}
	out.RecentTemplates = append(out.RecentTemplates, templateID)
	if repeatWindow > 0 && len(out.RecentTemplates) > repeatWindow {
		out.RecentTemplates = append([]string(nil), out.RecentTemplates[len(out.RecentTemplates)-repeatWindow:]...)
	}
	out.CounterReads[templateID]++
	if round.WinnerTeamID == teamID {
		out.PreviousSuccess[templateID]++
	}
	if round.Bomb != nil && round.Bomb.Site != "" {
		out.SideTendency["site:"+round.Bomb.Site]++
	}
	return out, nil
}

func (e *MatchEngine) updateStrategyMemory(round *RoundResult) {
	window := e.input.MapConfig.CombatConstants.Int("StrategyRepeatWindow", 0)
	for _, teamID := range []string{e.input.TeamA.TeamID, e.input.TeamB.TeamID} {
		memory, err := UpdateStrategyMemoryFromRound(e.strategyMemory[teamID], round, teamID, window)
		if err == nil {
			e.strategyMemory[teamID] = memory
		}
	}
}

func (e *MatchEngine) aggregateRoundStats(round *RoundResult) {
	killsThisRound := map[string]int{}
	for _, event := range round.Events {
		if event == nil {
			continue
		}
		switch event.EventType {
		case EventDamage:
			if acc := e.stats[event.AttackerID]; acc != nil {
				acc.Damage += eventDamageAmount(event)
			}
		case EventKill:
			if attacker := e.stats[event.AttackerID]; attacker != nil {
				attacker.Kills++
				killsThisRound[event.AttackerID]++
				if event.IsFirstKill {
					attacker.FK++
				}
			}
			if victim := e.stats[event.VictimID]; victim != nil {
				victim.Deaths++
			}
		case EventBombPlant:
			if planter := e.stats[event.AttackerID]; planter != nil {
				planter.Plants++
			}
		case EventBombDefuse:
			if defuser := e.stats[event.AttackerID]; defuser != nil {
				defuser.Defuses++
			}
		}
	}
	for playerID, kills := range killsThisRound {
		if kills >= 2 {
			e.stats[playerID].MK++
		}
	}
}

func eventDamageAmount(event *GameEvent) int {
	if event == nil || event.Extra == nil {
		return 0
	}
	switch value := event.Extra["damage"].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func roundHasEvent(round *RoundResult, eventType string) bool {
	for _, event := range round.Events {
		if event != nil && event.EventType == eventType {
			return true
		}
	}
	return false
}
