package matchengine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// StateFingerprint covers every causal dimension used by NoOp detection. It
// is stable across Go map insertion order.
func StateFingerprint(state *RoundState) string {
	if state == nil {
		return "nil"
	}
	var parts []string
	parts = append(parts,
		fmt.Sprintf("time:%d:%d:%d", state.Timeline, state.RoundDeadline, state.BombDeadline),
		fmt.Sprintf("phase:%s", state.Phase),
		fmt.Sprintf("resource:%d:%d", state.MomentumT, state.MomentumCT),
		fmt.Sprintf("bomb:%s:%s:%s:%d:%s:%d:%d:%s:%s", state.Bomb.Status, state.Bomb.CarrierID, locationKey(state.Bomb.Location), state.Bomb.DroppedAt, state.Bomb.PlantedSite, state.Bomb.PlantedAt, state.Bomb.ExplodeAt, state.Bomb.PlantActionID, state.Bomb.DefuseActionID),
		fmt.Sprintf("decision:%d:%d:%d", state.DecisionCount, state.RotationCount[SideT], state.RotationCount[SideCT]),
		fmt.Sprintf("recovery:%d:%t:%s:%s:%s:%d:%d:%s", state.NoOpCount, state.NoProgressEligible, state.RecoveryAttempt.CycleID, state.RecoveryAttempt.Status, state.RecoveryAttempt.RecoveryActionID, state.RecoveryAttempt.StartedAt, state.RecoveryAttempt.CompletedAt, state.RecoveryAttempt.ResultCode),
	)
	for _, id := range sortedPlayerIDs(state) {
		player := state.Players[id]
		parts = append(parts, fmt.Sprintf("player:%s:%s:%t:%d:%d:%d:%t:%d:%s:%s:%d:%s:%s:%s:%d:%d:%d", id, player.Side, player.Alive, player.HP, player.Stamina, player.Focus, player.Suppressed, player.Momentum, locationKey(player.Location), player.Intent.ID, player.Intent.CreatedAt, player.Intent.Type, player.Intent.TargetID, player.Action.CurrentActionID, player.Action.Version, player.Action.BusyUntil, player.Damage))
	}
	nodeIDs := make([]string, 0, len(state.Nodes))
	for id := range state.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	for _, id := range nodeIDs {
		node := state.Nodes[id]
		parts = append(parts, fmt.Sprintf("node:%s:%s:%d", id, node.ActualControl, node.UpdatedAt))
		for _, side := range []string{SideT, SideCT} {
			known := node.KnownControl[side]
			parts = append(parts, fmt.Sprintf("known:%s:%s:%s:%d:%d:%s", id, side, known.Status, known.UpdatedAt, known.ExpiresAt, strings.Join(append([]string(nil), known.ObservedBy...), ",")))
		}
	}
	for _, side := range []string{SideT, SideCT} {
		intel := state.Intel[side]
		if intel == nil {
			continue
		}
		records := append([]IntelRecord(nil), intel.Records...)
		sort.SliceStable(records, func(i, j int) bool { return records[i].ID < records[j].ID })
		for _, record := range records {
			parts = append(parts, fmt.Sprintf("intel:%s:%s:%s:%s:%s:%d:%d:%d", side, record.ID, record.Type, record.TargetID, record.NodeID, record.Confidence, record.LastSeenAt, record.ExpiresAt))
		}
	}
	actions := append([]ScheduledAction(nil), state.Scheduler.actions...)
	sort.SliceStable(actions, func(i, j int) bool { return actions[i].ID < actions[j].ID })
	for _, action := range actions {
		parts = append(parts, fmt.Sprintf("action:%s:%s:%d:%d:%s:%s:%s", action.ID, action.Type, action.StartAt, action.ResolveAt, strings.Join(action.ActorIDs, ","), action.ToNodeID, action.Payload.TargetID))
	}
	for _, event := range state.Events {
		if event != nil {
			parts = append(parts, fmt.Sprintf("event:%s:%s:%d:%s:%s", event.EventID, event.EventType, event.Timestamp, event.SourceActionID, event.SourceEffectID))
		}
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(digest[:])
}

// ObserveStateProgress updates NoOp/recovery bookkeeping after one attempted
// transition. Progress created by the current recovery action intentionally
// keeps that cycle alive until its completion is classified.
func ObserveStateProgress(state *RoundState, beforeFingerprint, sourceActionID string) (bool, error) {
	progressed := beforeFingerprint != StateFingerprint(state)
	if !progressed {
		return false, state.RecordNoOp()
	}
	if state.RecoveryAttempt.Status == RecoveryRunning && sourceActionID == state.RecoveryAttempt.RecoveryActionID {
		return true, nil
	}
	resetRecoveryCycle(state)
	return true, nil
}

func (s *RoundState) RecordNoOp() error {
	limit := s.constants.Int("MaxNoOpTransitions", 0)
	if limit <= 0 {
		return newError("INVALID_COMBAT_CONSTANT", "MaxNoOpTransitions must be positive")
	}
	if s.NoOpCount < limit {
		s.NoOpCount++
	}
	if s.NoOpCount == limit && s.RecoveryAttempt.CycleID == "" {
		s.RecoveryAttempt = NoProgressRecoveryState{
			CycleID: stableObjectID("recovery", s.Seed, "no_progress_recovery", s.RecoveryOrdinal),
			Status:  RecoveryNotAttempted,
		}
		s.RecoveryOrdinal++
	}
	return nil
}

// ScheduleNoProgressRecovery attempts exactly one deterministic recovery
// action for the active cycle. An unavailable action records a failed proof
// and enables one explicit pure terminal check.
func ScheduleNoProgressRecovery(state *RoundState) (*ScheduledAction, error) {
	if state == nil || state.RecoveryAttempt.CycleID == "" || state.RecoveryAttempt.Status != RecoveryNotAttempted {
		return nil, nil
	}
	if bombIsPostPlant(state.Bomb.Status) {
		failRecovery(state, "POST_PLANT_USES_BOMB_DEADLINE")
		state.NoProgressEligible = false
		return nil, nil
	}
	intentID := state.RecoveryAttempt.CycleID
	if state.Bomb.Status == BombDropped {
		for _, actorID := range sortedLivePlayerIDs(state, SideT) {
			schedule, feedback, err := ScheduleBombRecovery(state, actorID, intentID, state.RecoveryOrdinal)
			if err != nil {
				if engineError, ok := err.(*EngineError); ok && engineError.Code == "INVALID_PICKUP" {
					continue
				}
				return nil, err
			}
			if feedback != nil {
				continue
			}
			startRecovery(state, schedule.Action.ID)
			return &schedule.Action, nil
		}
		failRecovery(state, "NO_REACHABLE_BOMB_RECOVERY")
		return nil, nil
	}
	carrier := state.Players[state.Bomb.CarrierID]
	if carrier == nil || !carrier.Alive || carrier.Location.NodeID == "" {
		failRecovery(state, "NO_LIVE_NODE_CARRIER")
		return nil, nil
	}
	siteNode, path, ok := nearestReachablePlantSite(state, carrier.Location.NodeID)
	if !ok {
		failRecovery(state, "NO_REACHABLE_PLANT_SITE")
		return nil, nil
	}
	if len(path.EdgeIDs) == 0 {
		site := state.Nodes[siteNode].Node.Site
		action, err := StartPlantAction(state, carrier.Profile.PlayerID, site, intentID, state.RecoveryOrdinal)
		if err != nil {
			failRecovery(state, "PLANT_RECOVERY_REJECTED")
			return nil, nil
		}
		startRecovery(state, action.ID)
		return &action, nil
	}
	action, err := StartMovement(state, []string{carrier.Profile.PlayerID}, path.EdgeIDs[0], MoveProfile{Tempo: "Fast"}, intentID, state.RecoveryOrdinal)
	if err != nil {
		return nil, err
	}
	startRecovery(state, action.ID)
	return &action, nil
}

func CompleteNoProgressRecovery(state *RoundState, actionID string, succeeded bool, resultCode string) error {
	if state == nil || state.RecoveryAttempt.Status != RecoveryRunning || state.RecoveryAttempt.RecoveryActionID != actionID {
		return newError("SIMULATION_INVARIANT_ERROR", "recovery completion does not match the running cycle")
	}
	state.RecoveryAttempt.CompletedAt = state.Timeline
	state.RecoveryAttempt.ResultCode = resultCode
	if succeeded {
		state.RecoveryAttempt.Status = RecoverySucceeded
		state.NoProgressEligible = false
		state.NoOpCount = 0
		return nil
	}
	state.RecoveryAttempt.Status = RecoveryFailed
	state.NoProgressEligible = true
	return nil
}

// ValidNoProgress checks only the business condition. Eligibility is checked
// separately by EvaluateRoundTerminal.
func ValidNoProgress(state *RoundState) (bool, error) {
	if err := validateRuntimeGraph(state); err != nil {
		return false, err
	}
	if state == nil || bombIsPostPlant(state.Bomb.Status) || state.Bomb.Status == BombPlanting || liveCount(state, SideT) == 0 {
		return false, nil
	}
	if pendingReachabilityAction(state) {
		return false, nil
	}
	plantNodes := configuredPlantNodes(state)
	if len(plantNodes) == 0 {
		return false, newError("CONFIG_NO_PLANT_SITE", "runtime graph has no plant site")
	}
	if state.Bomb.Status == BombCarried {
		carrier := state.Players[state.Bomb.CarrierID]
		if carrier == nil || !carrier.Alive {
			return false, terminalInvariant("carried bomb has no live carrier")
		}
		from := projectedNodeID(carrier.Location)
		return !canReachAny(state, from, plantNodes), nil
	}
	if state.Bomb.Status == BombDropped {
		bombNode := bombRecoveryNode(state.Bomb.Location)
		if !canReachAny(state, bombNode, plantNodes) {
			return true, nil
		}
		for _, actorID := range sortedLivePlayerIDs(state, SideT) {
			if graphReachable(state, projectedNodeID(state.Players[actorID].Location), bombNode) {
				return false, nil
			}
		}
		return true, nil
	}
	return false, nil
}

func startRecovery(state *RoundState, actionID string) {
	state.RecoveryAttempt.Status = RecoveryRunning
	state.RecoveryAttempt.RecoveryActionID = actionID
	state.RecoveryAttempt.StartedAt = state.Timeline
	state.NoProgressEligible = false
}

func failRecovery(state *RoundState, code string) {
	state.RecoveryAttempt.Status = RecoveryFailed
	state.RecoveryAttempt.CompletedAt = state.Timeline
	state.RecoveryAttempt.ResultCode = code
	state.NoProgressEligible = true
}

func resetRecoveryCycle(state *RoundState) {
	state.NoOpCount = 0
	state.NoProgressEligible = false
	state.RecoveryAttempt = NoProgressRecoveryState{Status: RecoveryNotAttempted}
}

func validateRuntimeGraph(state *RoundState) error {
	if state == nil || len(state.Nodes) == 0 {
		return newError("CONFIG_UNREACHABLE_NODE", "runtime graph has no nodes")
	}
	for edgeID, edge := range state.mapEdges {
		if edge.ID == "" || edge.ID != edgeID || edge.BaseTime <= 0 || state.Nodes[edge.FromNode] == nil || state.Nodes[edge.ToNode] == nil {
			return newError("CONFIG_BAD_EDGE_NODE", "runtime edge %s is invalid", edgeID)
		}
	}
	return nil
}

func pendingReachabilityAction(state *RoundState) bool {
	for _, action := range state.Scheduler.actions {
		if !oneOf(string(action.Type), string(ActionMovementArrive), string(ActionPickupComplete), string(ActionPlantComplete)) {
			continue
		}
		for _, actorID := range validActionActors(state, action) {
			if state.Players[actorID].Side == SideT {
				return true
			}
		}
	}
	return false
}

func nearestReachablePlantSite(state *RoundState, from string) (string, PathResult, bool) {
	bestNode := ""
	best := PathResult{}
	for _, nodeID := range configuredPlantNodes(state) {
		path, feedback, err := FindBoundedPath(&MapConfig{Nodes: runtimeNodes(state), Edges: state.mapEdges}, from, nodeID, len(state.Nodes)*4)
		if err != nil || feedback != nil {
			continue
		}
		if bestNode == "" || path.TotalBaseTime < best.TotalBaseTime || path.TotalBaseTime == best.TotalBaseTime && nodeID < bestNode {
			bestNode, best = nodeID, path
		}
	}
	return bestNode, best, bestNode != ""
}

func configuredPlantNodes(state *RoundState) []string {
	var nodes []string
	for id, node := range state.Nodes {
		if node != nil && node.Node.Site != "" && node.Node.Site != "None" && hasString(node.Node.AreaUsages, "Plant") {
			nodes = append(nodes, id)
		}
	}
	sort.Strings(nodes)
	return nodes
}

func canReachAny(state *RoundState, from string, targets []string) bool {
	for _, target := range targets {
		if graphReachable(state, from, target) {
			return true
		}
	}
	return false
}

func graphReachable(state *RoundState, from, to string) bool {
	if state.Nodes[from] == nil || state.Nodes[to] == nil {
		return false
	}
	seen := map[string]bool{from: true}
	queue := []string{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == to {
			return true
		}
		for _, edge := range state.mapEdges {
			next := ""
			if edge.FromNode == current {
				next = edge.ToNode
			} else if edge.Bidirectional && edge.ToNode == current {
				next = edge.FromNode
			}
			if next != "" && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

func sortedLivePlayerIDs(state *RoundState, side string) []string {
	var ids []string
	for id, player := range state.Players {
		if player != nil && player.Alive && player.Side == side {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func sortedPlayerIDs(state *RoundState) []string {
	ids := make([]string, 0, len(state.Players))
	for id := range state.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func locationKey(location PlayerLocation) string {
	if location.NodeID != "" {
		return "node:" + location.NodeID
	}
	if location.Edge == nil {
		return "invalid"
	}
	return fmt.Sprintf("edge:%s:%s:%s:%.6f:%.6f:%.6f", location.Edge.EdgeID, location.Edge.FromNode, location.Edge.ToNode, location.Edge.Progress, location.Edge.X, location.Edge.Y)
}
