package matchengine

import (
	"context"
	"fmt"
	"sort"
)

type causalRoundRuntime struct {
	input              *RoundInput
	state              *RoundState
	decisionCandidates map[string]DecisionCandidate
	forcedDecisionUsed map[string]bool
	ordinal            int
}

func runCausalRound(ctx context.Context, input *RoundInput) (*RoundSimulationResult, error) {
	if input == nil || input.MapConfig == nil {
		return nil, newError("INVALID_ROUND_INPUT", "causal round requires an immutable input snapshot")
	}
	tScore, ctScore := input.ScoreByTeam[input.TeamT.TeamID], input.ScoreByTeam[input.TeamCT.TeamID]
	tSelection, err := SelectStrategyTemplate(input.MapConfig, input.TeamT, SideT, tScore, ctScore, input.StrategyMemoryT, input.Seed, 0)
	if err != nil {
		return nil, err
	}
	ctSelection, err := SelectCTSetup(input.MapConfig, input.TeamCT, ctScore, tScore, input.StrategyMemoryCT, input.Seed, 0)
	if err != nil {
		return nil, err
	}
	plan, reasons, err := BuildRoundPlan(input, tSelection.Template, ctSelection.Template)
	if err != nil {
		return nil, err
	}
	state, err := NewRoundState(input, plan)
	if err != nil {
		return nil, err
	}
	runtime := &causalRoundRuntime{input: input, state: state, decisionCandidates: map[string]DecisionCandidate{}, forcedDecisionUsed: map[string]bool{}}
	runtime.appendRoundStart(reasons, tSelection.Score, ctSelection.Score)
	if _, err := DeployOpeningActions(state); err != nil {
		return nil, err
	}
	if err := runtime.reevaluatePhase(); err != nil {
		return nil, err
	}

	for state.Terminal == nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := state.RecordTransition(); err != nil {
			return nil, err
		}
		beforeState := StateFingerprint(state)
		beforeDecision := CaptureDecisionFingerprint(state)
		if err := runtime.ensureActions(); err != nil {
			return nil, err
		}
		if err := runtime.reevaluatePhase(); err != nil {
			return nil, err
		}
		terminal, err := EvaluateRoundTerminal(state, AppliedBatch{Timestamp: state.Timeline})
		if err != nil {
			return nil, err
		}
		if terminal != nil {
			if err := runtime.enterRoundEnd(terminal); err != nil {
				return nil, err
			}
			break
		}

		nextAt, kind := state.NextTime()
		if ready, ok := state.Scheduler.Peek(); ok && ready.ResolveAt <= state.Timeline {
			nextAt, kind = state.Timeline, NextTimeAction
		}
		if kind == NextTimeNone {
			if err := state.RecordNoOp(); err != nil {
				return nil, err
			}
			if _, err := ScheduleNoProgressRecovery(state); err != nil {
				return nil, err
			}
			terminal, err = EvaluateRoundTerminal(state, AppliedBatch{Timestamp: state.Timeline})
			if err != nil {
				return nil, err
			}
			if terminal != nil {
				if err := runtime.enterRoundEnd(terminal); err != nil {
					return nil, err
				}
				break
			}
			return nil, newError("SCHEDULER_STALLED", "round has no action, deadline or eligible recovery")
		}
		if err := state.AdvanceTimeline(nextAt); err != nil {
			return nil, err
		}
		batch, sourceActionID, err := runtime.resolveTimestamp(kind)
		if err != nil {
			return nil, err
		}
		if err := DecayIntelAndControl(state, state.Timeline); err != nil {
			return nil, err
		}
		triggers := DetectDecisionTriggers(state, beforeDecision, nil)
		if len(triggers) > 0 {
			if err := runtime.scheduleDecisions(); err != nil {
				return nil, err
			}
		}
		if err := runtime.reevaluatePhase(); err != nil {
			return nil, err
		}
		terminal, err = EvaluateRoundTerminal(state, batch)
		if err != nil {
			return nil, err
		}
		if terminal != nil {
			if err := runtime.enterRoundEnd(terminal); err != nil {
				return nil, err
			}
			break
		}
		if _, err := ObserveStateProgress(state, beforeState, sourceActionID); err != nil {
			return nil, err
		}
	}

	round, err := ProjectRoundResult(state, input)
	if err != nil {
		return nil, err
	}
	round.RouteMain = plan.OpeningRoutes[plan.BombCarrierID]
	round.Report = BuildExplainableReport(round)
	return &RoundSimulationResult{Round: round, Terminal: state.Terminal, PhaseHistory: append([]RoundPhase(nil), state.PhaseHistory...)}, nil
}

func (runtime *causalRoundRuntime) appendRoundStart(roleReasons []ReasonRecord, tScore, ctScore StrategyScore) {
	state := runtime.state
	record := strategyScoreReason(tScore)
	if len(roleReasons) > 0 {
		record.Modifiers = append(record.Modifiers, ReasonModifier{Code: roleReasons[0].Code, Value: roleReasons[0].Value, Detail: roleReasons[0].Detail})
	}
	actionID := stableObjectID("act", state.Seed, "strategy", state.Plan.TStrategyTemplateID)
	reason, _ := ProjectReasonRecord(record, actionID, "")
	event := &GameEvent{
		EventID: NewEventID(state.Seed, actionID, "", EventRoundStart, 0), SourceActionID: actionID, Timestamp: 0, EventType: EventRoundStart,
		Message: "round started from authoritative opening plan", Reason: reason,
		Extra: map[string]interface{}{"strategy_template_id": state.Plan.TStrategyTemplateID, "ct_setup_template_id": state.Plan.CTSetupTemplateID},
	}
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	for ordinal, score := range []StrategyScore{tScore, ctScore} {
		if score.PreviousSuccessBonus == 0 && score.RepeatPenalty == 0 && score.CounterReadRisk == 0 {
			continue
		}
		adjustActionID := stableObjectID("act", state.Seed, "strategy_adjusted", score.TemplateID)
		adjustReason, _ := ProjectReasonRecord(strategyScoreReason(score), adjustActionID, "")
		adjusted := &GameEvent{EventID: NewEventID(state.Seed, adjustActionID, "", EventStrategyAdjusted, ordinal), SourceActionID: adjustActionID, Timestamp: 0, EventType: EventStrategyAdjusted, Message: "strategy score adjusted by completed-round memory", Reason: adjustReason, sortActionType: "StrategySelection"}
		teamID := state.TeamTID
		if ordinal == 1 {
			teamID = state.TeamCTID
		}
		adjusted.AttackerTeamID = teamID
		state.Events = append(state.Events, adjusted)
	}
}

func strategyScoreReason(score StrategyScore) ReasonRecord {
	return ReasonRecord{
		Code: "STRATEGY_SCORE", MainFactor: score.TemplateID, ScoreDelta: score.FinalScore,
		Modifiers: []ReasonModifier{{Code: "PREVIOUS_SUCCESS", Value: score.PreviousSuccessBonus}, {Code: "REPEAT_PENALTY", Value: -score.RepeatPenalty}, {Code: "COUNTER_READ_RISK", Value: -score.CounterReadRisk}, {Code: "RANDOM_NOISE", Value: score.RandomNoise}},
		Formula:   "FinalScore = Base + LineupFit + ScorePressure + PreviousSuccess - RepeatPenalty - CounterReadRisk + RandomNoise",
		Inputs:    map[string]float64{"template_base": score.TemplateBaseWeight, "lineup_fit": score.LineupFitScore, "score_pressure": score.CurrentScorePressure, "previous_success": score.PreviousSuccessBonus, "repeat_penalty": score.RepeatPenalty, "counter_read_risk": score.CounterReadRisk, "random_noise": score.RandomNoise},
	}
}

func (runtime *causalRoundRuntime) appendDecisionEvent(action ScheduledAction, candidate DecisionCandidate) {
	eventType := ""
	switch candidate.Type {
	case DecisionReinforce, DecisionRetake:
		eventType = EventReinforce
	case DecisionRotate, DecisionForceExecute, DecisionInterceptRotate:
		eventType = EventRotate
	}
	if eventType == "" {
		return
	}
	record := ReasonRecord{Code: string(candidate.Type), Source: candidate.TargetNode, Value: candidate.Score, Weight: 1}
	if len(candidate.Reasons) > 0 {
		record = candidate.Reasons[0]
	}
	reason, _ := ProjectReasonRecord(record, action.ID, "")
	eventID := NewEventID(runtime.state.Seed, action.ID, "", eventType, 0)
	event := &GameEvent{EventID: eventID, SourceActionID: action.ID, Timestamp: int64(runtime.state.Timeline), EventType: eventType, Message: fmt.Sprintf("%s executes %s toward %s", candidate.Side, candidate.Type, candidate.TargetNode), Reason: reason, Extra: map[string]interface{}{"decision_type": string(candidate.Type), "target_node": candidate.TargetNode}, sortPriority: action.Priority, sortActionType: string(action.Type), sortMinActorID: action.MinActorID()}
	if len(candidate.ActorIDs) > 0 {
		if actor := runtime.state.Players[candidate.ActorIDs[0]]; actor != nil {
			event.AttackerID, event.AttackerName, event.AttackerTeamID = actor.Profile.PlayerID, actor.Profile.DisplayName, actor.TeamID
			event.Location = eventLocation(runtime.state, actor.Location, eventID, action.ID)
			event.Message = fmt.Sprintf("%s executes %s toward %s", actor.TeamID, candidate.Type, candidate.TargetNode)
		}
	}
	runtime.state.Events = append(runtime.state.Events, event)
}

func (runtime *causalRoundRuntime) resolveTimestamp(kind NextTimeKind) (AppliedBatch, string, error) {
	state := runtime.state
	batch := AppliedBatch{Timestamp: state.Timeline}
	sources := map[string]bool{}
	if kind != NextTimeAction {
		return batch, "", nil
	}
	for {
		action, ok := state.Scheduler.Peek()
		if !ok || action.ResolveAt > state.Timeline {
			break
		}
		_, _ = state.Scheduler.Pop()
		validActors := validActionActors(state, action)
		if len(action.ActorIDs) > 0 && len(validActors) < maxInt(1, action.MinRequiredActors) {
			cancelActionForActors(state, action)
			continue
		}
		applied, err := runtime.resolveAction(action, validActors)
		if err != nil {
			return AppliedBatch{}, "", err
		}
		if applied != nil {
			batch.Effects = append(batch.Effects, applied.Effects...)
			batch.Events = append(batch.Events, applied.Events...)
		}
		sources[action.ID] = true
	}
	sourceID := ""
	if len(sources) == 1 {
		for id := range sources {
			sourceID = id
		}
	} else if len(sources) > 1 {
		sourceID = "mixed"
	}
	return batch, sourceID, nil
}

func (runtime *causalRoundRuntime) resolveAction(action ScheduledAction, validActors []string) (*AppliedBatch, error) {
	state := runtime.state
	switch action.Type {
	case ActionHoldStart:
		CompleteActionForActors(state, action, validActors)
	case ActionMovementArrive:
		if err := CompleteMovement(state, action, validActors); err != nil {
			return nil, err
		}
	case ActionInterceptCheck:
		candidate, err := ResolveInterceptCheck(state, action)
		if err != nil {
			return nil, err
		}
		if candidate != nil {
			if err := runtime.startEncounter(candidate.SourceActionID, candidate.NodeID, candidate.ActorIDs); err != nil {
				return nil, err
			}
		}
	case ActionCombatPulse:
		if state.ActiveEngagements[action.Payload.TargetID] == nil {
			return nil, nil
		}
		result, err := ResolveCombatPulse(state, action)
		if err != nil {
			return nil, err
		}
		if result.ShouldEnd && state.ActiveEngagements[result.EncounterID] != nil {
			if _, err := EndEncounter(state, result.EncounterID, result.EndReason); err != nil {
				return nil, err
			}
		}
		return result.Batch, nil
	case ActionCombatEnd:
		if state.ActiveEngagements[action.Payload.TargetID] != nil {
			if _, err := EndEncounter(state, action.Payload.TargetID, "scheduled_end"); err != nil {
				return nil, err
			}
		}
	case ActionDecisionResolve:
		candidate, ok := runtime.decisionCandidates[action.ID]
		delete(runtime.decisionCandidates, action.ID)
		if !ok {
			return nil, newError("SIMULATION_INVARIANT_ERROR", "decision action %s has no frozen candidate", action.ID)
		}
		resolution, err := ResolveDecision(state, action, candidate)
		if err != nil {
			if transientDecisionError(err) {
				return nil, nil
			}
			return nil, err
		}
		if len(resolution.Actions) > 0 {
			runtime.appendDecisionEvent(action, candidate)
		}
	case ActionPlantComplete:
		result, err := ResolvePlantComplete(state, action)
		if err != nil {
			return nil, err
		}
		return bombResultBatch(state.Timeline, result), nil
	case ActionPickupComplete:
		result, err := ResolveBombPickup(state, action)
		if err != nil {
			return nil, err
		}
		return bombResultBatch(state.Timeline, result), nil
	case ActionDefuseComplete:
		result, err := ResolveDefuseComplete(state, action)
		if err != nil {
			return nil, err
		}
		return bombResultBatch(state.Timeline, result), nil
	case ActionBombExplode:
		result, err := ResolveBombExplode(state, action)
		if err != nil {
			return nil, err
		}
		return bombResultBatch(state.Timeline, result), nil
	case ActionIntelDecay, ActionControlDecay, ActionRoundExpire:
		// Deadlines and knowledge expiry are applied once after the whole batch.
	default:
		return nil, newError("UNKNOWN_ACTION_TYPE", "cannot resolve action type %s", action.Type)
	}
	return nil, nil
}

func bombResultBatch(at int, result BombActionResult) *AppliedBatch {
	if !result.Applied || result.Event == nil {
		return nil
	}
	return &AppliedBatch{Timestamp: at, Events: []*GameEvent{result.Event}}
}

func transientDecisionError(err error) bool {
	engineError, ok := err.(*EngineError)
	if !ok {
		return false
	}
	return oneOf(engineError.Code, "INVALID_DECISION", "DECISION_UNREACHABLE", "INVALID_PLANT", "SITE_CONTEST_REQUIRED", "INVALID_DEFUSE", "DEFUSE_CONTEST_REQUIRED", "INVALID_PICKUP")
}

func (runtime *causalRoundRuntime) ensureActions() error {
	state := runtime.state
	if state.Terminal != nil {
		return nil
	}
	started, err := runtime.discoverEncounters()
	if err != nil || started {
		return err
	}
	if state.Bomb.Status == BombPlanting || state.Bomb.Status == BombDefusing {
		return nil
	}
	if bombIsPostPlant(state.Bomb.Status) {
		return runtime.ensurePostPlantActions()
	}
	started, err = runtime.ensurePrePlantActions()
	if err != nil || started {
		return err
	}
	return runtime.scheduleDecisions()
}

func (runtime *causalRoundRuntime) ensurePrePlantActions() (bool, error) {
	state := runtime.state
	if state.Bomb.Status == BombDropped {
		for _, actorID := range sortedLivePlayerIDs(state, SideT) {
			player := state.Players[actorID]
			if player.Action.CurrentActionID != "" || player.EngagementID != "" {
				continue
			}
			if player.Location.Edge != nil {
				if _, err := ResumeInterruptedMovement(state, actorID, stableObjectID("intent", state.Seed, "auto_recover_edge", state.Timeline, actorID), runtime.nextOrdinal()); err != nil {
					return false, err
				}
				return true, nil
			}
			schedule, feedback, err := ScheduleBombRecovery(state, actorID, stableObjectID("intent", state.Seed, "auto_recover", state.Timeline), runtime.nextOrdinal())
			if err != nil {
				if transientDecisionError(err) {
					continue
				}
				return false, err
			}
			if feedback == nil {
				_ = schedule
				return true, nil
			}
		}
	}
	started := false
	for _, actorID := range sortedPlayerIDs(state) {
		player := state.Players[actorID]
		if !player.Alive || player.Action.CurrentActionID != "" || player.EngagementID != "" {
			continue
		}
		if player.Location.Edge != nil {
			ultimateTarget := player.Intent.TargetID
			if _, err := ResumeInterruptedMovement(state, actorID, stableObjectID("intent", state.Seed, "resume_edge", state.Timeline, actorID), runtime.nextOrdinal()); err != nil {
				return false, err
			}
			if ultimateTarget != "" {
				player.Intent.TargetID = ultimateTarget
			}
			started = true
			continue
		}
		if player.Location.NodeID == "" {
			continue
		}
		if player.HasBomb {
			node := state.Nodes[player.Location.NodeID]
			if node != nil && node.Node.Site != "" && node.Node.Site != "None" && hasString(node.Node.AreaUsages, "Plant") && node.ActualControl != ControlCT && node.ActualControl != ControlContested && len(visibleThreatsAtSite(state, player.Location.NodeID, SideCT)) == 0 {
				if _, err := StartPlantAction(state, actorID, node.Node.Site, stableObjectID("intent", state.Seed, "auto_plant", state.Timeline, actorID), runtime.nextOrdinal()); err != nil {
					if !transientDecisionError(err) {
						return false, err
					}
				} else {
					return true, nil
				}
			}
		}
		if player.Intent.Type == IntentMove && player.Intent.TargetID != "" && player.Intent.TargetID != player.Location.NodeID {
			targetNode := player.Intent.TargetID
			path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, player.Location.NodeID, targetNode, len(state.Nodes)*4)
			if err != nil {
				return false, err
			}
			if feedback == nil && len(path.EdgeIDs) > 0 {
				if _, err := StartMovement(state, []string{actorID}, path.EdgeIDs[0], MoveProfile{Tempo: "Fast"}, stableObjectID("intent", state.Seed, "continue_decision", targetNode, actorID, state.Timeline), runtime.nextOrdinal()); err != nil {
					return false, err
				}
				player.Intent.TargetID = targetNode
				started = true
				continue
			}
		}
		route := state.routes[state.Plan.OpeningRoutes[actorID]]
		nextNode := nextRouteNode(route, player.Location.NodeID)
		if nextNode == "" {
			continue
		}
		edgeID, ok := edgeBetween(state.mapEdges, player.Location.NodeID, nextNode)
		if !ok {
			return false, newError("CONFIG_ROUTE_NOT_CONNECTED", "route %s has no edge from %s to %s", route.ID, player.Location.NodeID, nextNode)
		}
		templateID := state.Plan.TStrategyTemplateID
		if player.Side == SideCT {
			templateID = state.Plan.CTSetupTemplateID
		}
		if _, err := StartMovement(state, []string{actorID}, edgeID, MoveProfile{Tempo: state.routeTemplates[templateID].Tempo}, stableObjectID("intent", state.Seed, "route", route.ID, actorID, state.Timeline), runtime.nextOrdinal()); err != nil {
			return false, err
		}
		started = true
	}
	return started, nil
}

func (runtime *causalRoundRuntime) ensurePostPlantActions() error {
	state := runtime.state
	bombNode := projectedNodeID(state.Bomb.Location)
	for _, side := range []string{SideT, SideCT} {
		for _, actorID := range sortedLivePlayerIDs(state, side) {
			player := state.Players[actorID]
			if player.Action.CurrentActionID != "" || player.EngagementID != "" || player.Location.NodeID == "" {
				continue
			}
			if side == SideCT && player.Location.NodeID == bombNode {
				if err := CanAttemptDefuse(state, actorID); err == nil {
					if _, err := StartDefuseAction(state, actorID, stableObjectID("intent", state.Seed, "auto_defuse", state.Timeline, actorID), runtime.nextOrdinal()); err != nil {
						return err
					}
					return nil
				}
			}
			if player.Location.NodeID == bombNode {
				player.Posture = PostureHolding
				continue
			}
			path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, player.Location.NodeID, bombNode, len(state.Nodes)*4)
			if err != nil {
				return err
			}
			if feedback != nil || len(path.EdgeIDs) == 0 {
				continue
			}
			if _, err := StartMovement(state, []string{actorID}, path.EdgeIDs[0], MoveProfile{Tempo: "Fast"}, stableObjectID("intent", state.Seed, "postplant", side, actorID, state.Timeline), runtime.nextOrdinal()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (runtime *causalRoundRuntime) discoverEncounters() (bool, error) {
	state := runtime.state
	var candidates []EncounterCandidatePlan
	seen := map[string]bool{}
	ids := sortedLivePlayerIDs(state, SideT)
	for _, tID := range ids {
		tPlayer := state.Players[tID]
		if tPlayer.EngagementID != "" {
			continue
		}
		for _, ctID := range sortedLivePlayerIDs(state, SideCT) {
			ctPlayer := state.Players[ctID]
			if ctPlayer.EngagementID != "" || !playersInContact(state, tPlayer, ctPlayer) {
				continue
			}
			nodeID := projectedNodeID(tPlayer.Location)
			key := nodeID + ":" + tID + ":" + ctID
			if seen[key] {
				continue
			}
			seen[key] = true
			actors := contactActors(state, tPlayer.Location, ctPlayer.Location)
			scenarioID := runtime.scenarioForContact(nodeID)
			candidate, err := BuildEncounterCandidate(state, stableObjectID("contact", state.Seed, state.Timeline, key), scenarioID, nodeID, actors)
			if err != nil {
				continue
			}
			candidates = append(candidates, candidate)
		}
	}
	accepted := ArbitrateEncounterCandidates(candidates)
	for _, candidate := range accepted {
		if _, err := StartEncounter(state, candidate); err != nil {
			return false, err
		}
	}
	return len(accepted) > 0, nil
}

func (runtime *causalRoundRuntime) startEncounter(sourceActionID, nodeID string, actorIDs []string) error {
	scenarioID := runtime.scenarioForContact(nodeID)
	candidate, err := BuildEncounterCandidate(runtime.state, sourceActionID, scenarioID, nodeID, actorIDs)
	if err != nil {
		return err
	}
	_, err = StartEncounter(runtime.state, candidate)
	return err
}

func (runtime *causalRoundRuntime) scenarioForContact(nodeID string) string {
	state := runtime.state
	wantedSite := ""
	isPlantSite := false
	if node := state.Nodes[nodeID]; node != nil && node.Node.Site != "None" {
		wantedSite = node.Node.Site
		isPlantSite = hasString(node.Node.AreaUsages, "Plant")
	}
	ids := make([]string, 0, len(state.scenarios))
	for id := range state.scenarios {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	best := ""
	bestScore := -1
	for _, id := range ids {
		scenario := state.scenarios[id]
		score := 0
		if wantedSite != "" && scenario.Site == wantedSite {
			score += 4
		}
		if bombIsPostPlant(state.Bomb.Status) && (scenario.Phase == "Retake" || scenario.Phase == "BombResolution") {
			score += 8
		}
		if !bombIsPostPlant(state.Bomb.Status) {
			switch {
			case isPlantSite && scenario.Phase == "SiteEntry":
				score += 8
			case !isPlantSite && (scenario.Phase == "OpeningDuel" || scenario.Phase == "MidControl"):
				score += 4
			}
		}
		if score > bestScore {
			best, bestScore = id, score
		}
	}
	return best
}

func (runtime *causalRoundRuntime) scheduleDecisions() error {
	state := runtime.state
	if runtime.forcedDecisionUsed == nil {
		runtime.forcedDecisionUsed = map[string]bool{}
	}
	maxDecisions := state.constants.Int("MaxDecisionCount", 0)
	for _, side := range []string{SideT, SideCT} {
		if hasPendingDecision(state, side, runtime.decisionCandidates) {
			continue
		}
		atDecisionLimit := state.DecisionCount >= maxDecisions
		if atDecisionLimit && runtime.forcedDecisionUsed[side] {
			continue
		}
		view, err := BuildDecisionView(state, side)
		if err != nil {
			return err
		}
		candidates := ScoreDecisionCandidates(view, state.routes, state.constants, IdentityRollSource{Seed: deriveSeed(state.Seed, "decision", side, state.DecisionCount, state.Timeline)})
		if len(candidates) == 0 {
			continue
		}
		action, normalized, err := ScheduleDecision(state, candidates[0], runtime.nextOrdinal())
		if err != nil {
			return err
		}
		runtime.decisionCandidates[action.ID] = normalized
		if atDecisionLimit {
			runtime.forcedDecisionUsed[side] = true
		}
	}
	return nil
}

func (runtime *causalRoundRuntime) reevaluatePhase() error {
	state := runtime.state
	phase := PhaseAdvance
	switch {
	case state.Terminal != nil:
		phase = PhaseRoundEnd
	case state.Bomb.Status == BombPlanted || state.Bomb.Status == BombDefusing || state.Bomb.Status == BombDefused || state.Bomb.Status == BombExploded:
		phase = PhasePostPlant
	case state.Bomb.Status == BombPlanting:
		phase = PhasePlanting
	case len(state.ActiveEngagements) > 0:
		phase = PhaseClash
	case carrierAtContestSite(state):
		phase = PhaseSiteContest
	case schedulerHasType(state.Scheduler, ActionDecisionResolve):
		phase = PhaseRotate
	case state.Timeline == 0:
		phase = PhaseOpeningDeploy
	}
	if state.Phase != phase {
		state.Phase = phase
		state.PhaseHistory = append(state.PhaseHistory, phase)
	}
	return nil
}

func (runtime *causalRoundRuntime) enterRoundEnd(terminal *RoundTerminal) error {
	state := runtime.state
	if terminal == nil {
		return newError("SIMULATION_INVARIANT_ERROR", "cannot enter RoundEnd without terminal")
	}
	for _, player := range state.Players {
		if player != nil && player.Action.CurrentActionID != "" {
			cancelCurrentAction(state, player)
		}
	}
	state.Scheduler.actions = nil
	state.Terminal = terminal
	state.Phase = PhaseRoundEnd
	state.PhaseHistory = append(state.PhaseHistory, PhaseRoundEnd)
	terminalActionID := stableObjectID("act", state.Seed, "terminal", terminal.Reason.Code, state.Timeline)
	reason, _ := ProjectReasonRecord(terminal.Reason, terminalActionID, "")
	event := &GameEvent{
		EventID: NewEventID(state.Seed, terminalActionID, "", EventRoundEnd, 0), SourceActionID: terminalActionID, Timestamp: int64(state.Timeline), EventType: EventRoundEnd,
		Message: fmt.Sprintf("%s wins the round by %s", terminal.WinnerSide, terminal.WinReason),
		Reason:  reason,
		Bomb:    projectBombState(state.Bomb),
	}
	event.State = snapshotForEvent(state)
	state.Events = append(state.Events, event)
	sortEvents(state.Events)
	return ValidateTerminalInvariants(state)
}

func (runtime *causalRoundRuntime) nextOrdinal() int {
	runtime.ordinal++
	return runtime.ordinal
}

func nextRouteNode(route Route, current string) string {
	for index, nodeID := range route.Nodes {
		if nodeID == current && index+1 < len(route.Nodes) {
			return route.Nodes[index+1]
		}
	}
	return ""
}

func playersInContact(state *RoundState, left, right *RoundPlayerState) bool {
	if left.Location.NodeID != "" && left.Location.NodeID == right.Location.NodeID {
		return true
	}
	return configuredVisible(state, left.Location, projectedNodeID(right.Location)) || configuredVisible(state, right.Location, projectedNodeID(left.Location))
}

func contactActors(state *RoundState, left, right PlayerLocation) []string {
	var ids []string
	for id, player := range state.Players {
		if !player.Alive || player.EngagementID != "" {
			continue
		}
		if sameLocation(player.Location, left) || sameLocation(player.Location, right) || configuredVisible(state, player.Location, projectedNodeID(left)) || configuredVisible(state, player.Location, projectedNodeID(right)) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func hasPendingDecision(state *RoundState, side string, candidates map[string]DecisionCandidate) bool {
	for _, action := range state.Scheduler.actions {
		candidate, ok := candidates[action.ID]
		if ok && action.Type == ActionDecisionResolve && candidate.Side == side {
			return true
		}
	}
	return false
}

func schedulerHasType(scheduler *ActionScheduler, actionType ActionType) bool {
	for _, action := range scheduler.actions {
		if action.Type == actionType {
			return true
		}
	}
	return false
}

func carrierAtContestSite(state *RoundState) bool {
	carrier := state.Players[state.Bomb.CarrierID]
	if carrier == nil || !carrier.Alive || carrier.Location.NodeID == "" {
		return false
	}
	node := state.Nodes[carrier.Location.NodeID]
	return node != nil && node.Node.Site != "" && node.Node.Site != "None" && hasString(node.Node.AreaUsages, "Plant")
}
