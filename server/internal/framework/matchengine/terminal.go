package matchengine

import "fmt"

const (
	WinReasonElimination       = "elimination"
	WinReasonTimeout           = "timeout"
	WinReasonBombDefused       = "bomb_defused"
	WinReasonBombExploded      = "bomb_exploded"
	WinReasonBombSecured       = "bomb_secured"
	WinReasonNoProgressTimeout = "no_progress_timeout"
)

// EvaluateRoundTerminal is deliberately pure: it only inspects the state
// produced by the just-applied timestamp batch and never rolls, repairs or
// advances the simulation.
func EvaluateRoundTerminal(state *RoundState, applied AppliedBatch) (*RoundTerminal, error) {
	if err := ValidateTerminalInvariants(state); err != nil {
		return nil, err
	}
	tAlive, ctAlive := liveCount(state, SideT), liveCount(state, SideCT)
	if state.Bomb.Status == BombDefused {
		return newRoundTerminal(state, SideCT, WinReasonBombDefused, "BOMB_DEFUSED"), nil
	}
	if state.Bomb.Status == BombExploded {
		return newRoundTerminal(state, SideT, WinReasonBombExploded, "BOMB_EXPLODED"), nil
	}
	if !bombIsPostPlant(state.Bomb.Status) {
		if tAlive == 0 {
			code := "T_ELIMINATED"
			if state.Bomb.Status == BombDropped {
				code = "T_ELIMINATED_BOMB_DROPPED"
			}
			return newRoundTerminal(state, SideCT, WinReasonElimination, code), nil
		}
		if ctAlive == 0 {
			return newRoundTerminal(state, SideT, WinReasonElimination, "CT_ELIMINATED"), nil
		}
		if state.Timeline >= state.RoundDeadline {
			return newRoundTerminal(state, SideCT, WinReasonTimeout, "ROUND_TIME_EXPIRED"), nil
		}
		if state.NoProgressEligible {
			valid, err := ValidNoProgress(state)
			if err != nil {
				return nil, err
			}
			if valid {
				return newRoundTerminal(state, SideCT, WinReasonNoProgressTimeout, "NO_PROGRESS_CONFIRMED"), nil
			}
		}
		return nil, nil
	}
	if ctAlive == 0 && !batchContainsEvent(applied, EventBombDefuse) {
		return newRoundTerminal(state, SideT, WinReasonBombSecured, "CT_ELIMINATED_POST_PLANT"), nil
	}
	return nil, nil
}

func newRoundTerminal(state *RoundState, side, publicCode, code string) *RoundTerminal {
	teamID := state.TeamTID
	if side == SideCT {
		teamID = state.TeamCTID
	}
	return &RoundTerminal{
		WinnerTeamID: teamID,
		WinnerSide:   side,
		WinReason:    publicCode,
		Reason:       ReasonRecord{Code: code, Source: "EvaluateRoundTerminal", Value: 1, Weight: 1},
	}
}

func liveCount(state *RoundState, side string) int {
	count := 0
	for _, player := range state.Players {
		if player != nil && player.Side == side && player.Alive {
			count++
		}
	}
	return count
}

func bombIsPostPlant(status BombRuntimeStatus) bool {
	return status == BombPlanted || status == BombDefusing || status == BombDefused || status == BombExploded
}

func batchContainsEvent(batch AppliedBatch, eventType string) bool {
	for _, event := range batch.Events {
		if event != nil && event.EventType == eventType {
			return true
		}
	}
	return false
}

// ValidateTerminalInvariants is strict and non-mutating. Invalid terminal
// state is a simulation error; it is never clamped into a plausible result.
func ValidateTerminalInvariants(state *RoundState) error {
	if state == nil || state.Scheduler == nil || state.RoundDeadline <= 0 || state.Timeline < 0 {
		return terminalInvariant("round state, scheduler or timer is invalid")
	}
	if state.Timeline > state.constants.Int("MaxRoundTimeline", 0) {
		return terminalInvariant("timeline exceeds MaxRoundTimeline")
	}
	carrierCount := 0
	for id, player := range state.Players {
		if player == nil || !oneOf(player.Side, SideT, SideCT) {
			return terminalInvariant("player %s is nil or has an invalid side", id)
		}
		if !player.Location.Valid() || !runtimeLocationConfigured(state, player.Location) {
			return terminalInvariant("player %s has an invalid semantic location", id)
		}
		if player.Alive != (player.HP > 0) {
			return terminalInvariant("player %s Alive/HP state disagrees", id)
		}
		if !player.Alive && player.Action.CurrentActionID != "" {
			return terminalInvariant("dead player %s has a running action", id)
		}
		if player.Action.CurrentActionID == "" && player.Action.Busy.ActionID != "" {
			return terminalInvariant("player %s has a ghost busy interval", id)
		}
		if player.HasBomb {
			carrierCount++
			if state.Bomb.CarrierID != id || !sameLocation(player.Location, state.Bomb.Location) {
				return terminalInvariant("player %s disagrees with bomb carrier state", id)
			}
		}
	}
	if err := validateTerminalBomb(state, carrierCount); err != nil {
		return err
	}
	lastTimestamp := int64(-1)
	roundEndCount := 0
	for _, event := range state.Events {
		if event == nil || event.Timestamp < lastTimestamp || event.Timestamp > int64(state.Timeline) {
			return terminalInvariant("event timeline is nil, unsorted or in the future")
		}
		lastTimestamp = event.Timestamp
		if event.EventType == EventRoundEnd {
			roundEndCount++
		}
	}
	if roundEndCount > 1 || (roundEndCount == 1 && state.Terminal == nil) {
		return terminalInvariant("ROUND_END lifecycle disagrees with terminal state")
	}
	if state.Phase == PhaseRoundEnd && state.Terminal == nil {
		return terminalInvariant("RoundEnd phase has no terminal")
	}
	return nil
}

func validateTerminalBomb(state *RoundState, carrierCount int) error {
	bomb := state.Bomb
	if !bomb.Location.Valid() || !runtimeLocationConfigured(state, bomb.Location) {
		return terminalInvariant("bomb has an invalid semantic location")
	}
	switch bomb.Status {
	case BombCarried, BombPlanting:
		if carrierCount != 1 || bomb.CarrierID == "" || state.BombDeadline != 0 {
			return terminalInvariant("carried/planting bomb carrier or deadline is inconsistent")
		}
		if bomb.Status == BombPlanting && (bomb.PlantActionID == "" || bomb.PlantActorID != bomb.CarrierID || bomb.PlantFinishAt <= bomb.PlantStartAt) {
			return terminalInvariant("plant action lifecycle is inconsistent")
		}
	case BombDropped:
		if carrierCount != 0 || bomb.CarrierID != "" || bomb.DroppedAt > state.Timeline || state.BombDeadline != 0 {
			return terminalInvariant("dropped bomb state is inconsistent")
		}
	case BombPlanted, BombDefusing, BombDefused, BombExploded:
		if carrierCount != 0 || bomb.CarrierID != "" || bomb.PlantedSite == "" || bomb.PlantedAt > state.Timeline || bomb.ExplodeAt <= bomb.PlantedAt || state.BombDeadline != bomb.ExplodeAt {
			return terminalInvariant("post-plant bomb state or timer is inconsistent")
		}
		if !hasEvent(state.Events, EventBombPlant) {
			return terminalInvariant("post-plant bomb has no applied BOMB_PLANT event")
		}
		if bomb.Status == BombDefusing {
			actor := state.Players[bomb.DefuseActorID]
			if actor == nil || !actor.Alive || actor.Side != SideCT || actor.Action.CurrentActionID != bomb.DefuseActionID || bomb.DefuseFinishAt <= bomb.DefuseStartAt || bomb.DefuseFinishAt > bomb.ExplodeAt {
				return terminalInvariant("defuse action lifecycle is inconsistent")
			}
		}
		if bomb.Status == BombDefused && !hasEvent(state.Events, EventBombDefuse) {
			return terminalInvariant("Defused state has no applied BOMB_DEFUSE event")
		}
		if bomb.Status == BombExploded && !hasEvent(state.Events, EventBombExplode) {
			return terminalInvariant("Exploded state has no applied BOMB_EXPLODE event")
		}
	default:
		return terminalInvariant("unknown bomb status %s", bomb.Status)
	}
	return nil
}

func runtimeLocationConfigured(state *RoundState, location PlayerLocation) bool {
	if location.NodeID != "" {
		_, ok := state.Nodes[location.NodeID]
		return ok
	}
	if location.Edge == nil || location.Edge.Progress < 0 || location.Edge.Progress > 1 {
		return false
	}
	edge, ok := state.mapEdges[location.Edge.EdgeID]
	return ok && location.Edge.FromNode != "" && location.Edge.ToNode != "" &&
		(edge.FromNode == location.Edge.FromNode && edge.ToNode == location.Edge.ToNode ||
			edge.Bidirectional && edge.ToNode == location.Edge.FromNode && edge.FromNode == location.Edge.ToNode)
}

func hasEvent(events []*GameEvent, eventType string) bool {
	for _, event := range events {
		if event != nil && event.EventType == eventType {
			return true
		}
	}
	return false
}

func terminalInvariant(format string, args ...interface{}) error {
	return &EngineError{Code: "SIMULATION_INVARIANT_ERROR", Message: fmt.Sprintf(format, args...)}
}
