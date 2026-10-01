package matchengine

import (
	"container/heap"
	"math"
	"sort"
	"strings"
)

type moveProfile struct {
	Tempo            string
	FormationPenalty float64
}

type pathResult struct {
	NodeIDs       []string
	EdgeIDs       []string
	TotalBaseTime int
}

type decisionFeedback struct {
	Code    string
	Message string
}

type interceptContext struct {
	EnemyActorIDs        []string
	Visible              bool
	ObserverPosture      combatPosture
	KnownEnemyConfidence int
}

type interceptCheckResult struct {
	Scheduled   bool
	Probability float64
	Roll        float64
	ActionID    string
	ReasonCode  string
}

type encounterCandidate struct {
	SourceActionID string
	NodeID         string
	ActorIDs       []string
	StartedAt      int
}

func resolveMoveDuration(edge MapEdge, actors []*roundPlayerState, profile moveProfile, constants CombatConstants) int {
	if len(actors) == 0 {
		return clampInt(edge.BaseTime, constants.Int("MinMoveTime", 1), constants.Int("MaxMoveTime", edge.BaseTime))
	}
	maxStamina := constants.Int("MaxStamina", 100)
	mobility, stamina := 0, 0
	for _, actor := range actors {
		mobility += actor.Profile.Attributes.Mobility
		stamina += actor.Stamina
	}
	avgMobility := float64(mobility) / float64(len(actors))
	avgStamina := float64(stamina) / float64(len(actors))
	staminaPenalty := math.Max(0, float64(maxStamina)-avgStamina) / 20
	mobilityModifier := (avgMobility - 50) / 20
	duration := float64(edge.BaseTime)*tempoFactor(profile.Tempo) + profile.FormationPenalty + staminaPenalty - mobilityModifier
	return clampInt(int(math.Round(duration)), constants.Int("MinMoveTime", 1), constants.Int("MaxMoveTime", edge.BaseTime))
}

func tempoFactor(tempo string) float64 {
	switch strings.ToLower(tempo) {
	case "fast", "fastdefault", "rush":
		return 0.85
	case "slow", "slowdefault", "defaultslow":
		return 1.15
	default:
		return 1
	}
}

func startMovement(state *roundState, actorIDs []string, edgeID string, profile moveProfile, intentID string, ordinal int) (scheduledAction, error) {
	if state == nil || len(actorIDs) == 0 {
		return scheduledAction{}, newError("INVALID_MOVE", "movement requires state and actors")
	}
	edge, ok := stateEdge(state, edgeID)
	if !ok {
		return scheduledAction{}, newError("CONFIG_MISSING_EDGE", "edge %s is missing", edgeID)
	}
	actors := make([]*roundPlayerState, 0, len(actorIDs))
	fromNode := ""
	for _, actorID := range actorIDs {
		actor := state.Players[actorID]
		if actor == nil || !actor.Alive || actor.Location.NodeID == "" {
			return scheduledAction{}, newError("INVALID_MOVE", "actor %s is not on a movable node", actorID)
		}
		if fromNode == "" {
			fromNode = actor.Location.NodeID
		} else if actor.Location.NodeID != fromNode {
			return scheduledAction{}, newError("INVALID_MOVE", "group actors do not share a start node")
		}
		actors = append(actors, actor)
	}
	toNode, ok := edgeDestination(edge, fromNode)
	if !ok {
		return scheduledAction{}, newError("INVALID_MOVE", "edge %s is not traversable from %s", edgeID, fromNode)
	}
	duration := resolveMoveDuration(edge, actors, profile, state.constants)
	action := scheduledAction{
		IntentID: intentID, Type: actionMovementArrive, ActorIDs: append([]string(nil), actorIDs...), StartAt: state.Timeline,
		ResolveAt: state.Timeline + duration, Priority: 20, MinRequiredActors: len(actorIDs), ToNodeID: toNode, Payload: actionPayload{EdgeID: edgeID},
	}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionMoving); err != nil {
		return scheduledAction{}, err
	}
	edgeLocation, err := resolveOnEdgeLocation(state, edge, fromNode, toNode, 0, action.ID)
	if err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	action.From = playerLocation{Edge: edgeLocation}
	for _, actorID := range action.ActorIDs {
		state.Players[actorID].Location = playerLocation{Edge: cloneOnEdge(edgeLocation)}
		if state.Players[actorID].HasBomb {
			state.Bomb.Location = clonePlayerLocation(state.Players[actorID].Location)
		}
		state.Players[actorID].Posture = postureMoving
		state.Players[actorID].Intent = intent{ID: intentID, Type: intentMove, TargetID: toNode, CreatedAt: state.Timeline}
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	return action, nil
}

// resumeInterruptedMovement converts an authoritative OnEdge location into a
// timed arrival at the traversal's intended endpoint. It preserves elapsed
// travel instead of snapping the actor to a node after an encounter.
func resumeInterruptedMovement(state *roundState, actorID, intentID string, ordinal int) (scheduledAction, error) {
	player := state.Players[actorID]
	if player == nil || !player.Alive || player.Action.CurrentActionID != "" || player.Location.Edge == nil {
		return scheduledAction{}, newError("INVALID_MOVE", "edge recovery requires an available on-edge actor")
	}
	edgeLocation := cloneOnEdge(player.Location.Edge)
	edge, ok := state.mapEdges[edgeLocation.EdgeID]
	if !ok {
		return scheduledAction{}, newError("CONFIG_MISSING_EDGE", "edge %s is missing", edgeLocation.EdgeID)
	}
	remainingFraction := 1 - clampProbability(edgeLocation.Progress)
	duration := clampInt(int(math.Ceil(float64(edge.BaseTime)*remainingFraction)), state.constants.Int("MinMoveTime", 1), state.constants.Int("MaxMoveTime", edge.BaseTime))
	action := scheduledAction{
		IntentID: intentID, Type: actionMovementArrive, ActorIDs: []string{actorID}, From: playerLocation{Edge: edgeLocation}, ToNodeID: edgeLocation.ToNode,
		StartAt: state.Timeline, ResolveAt: state.Timeline + duration, Priority: 20, MinRequiredActors: 1, Payload: actionPayload{EdgeID: edge.ID},
	}
	action.ID = newActionID(state.Seed, action.Type, intentID, action.StartAt, action.ResolveAt, action.ActorIDs, ordinal)
	if err := beginExclusiveAction(state, &action, actionMoving); err != nil {
		return scheduledAction{}, err
	}
	player.Intent = intent{ID: intentID, Type: intentMove, TargetID: edgeLocation.ToNode, CreatedAt: state.Timeline}
	player.Posture = postureMoving
	if err := state.Scheduler.Schedule(action); err != nil {
		cancelActionForActors(state, action)
		return scheduledAction{}, err
	}
	return action, nil
}

func updateMovementProgress(state *roundState, action scheduledAction, at int, actorIDs []string) error {
	if at < action.StartAt || at > action.ResolveAt || action.ResolveAt <= action.StartAt {
		return newError("INVALID_MOVE", "movement progress timestamp is outside action interval")
	}
	edge, ok := stateEdge(state, action.Payload.EdgeID)
	if !ok || action.From.Edge == nil {
		return newError("CONFIG_MISSING_EDGE", "movement edge %s is missing", action.Payload.EdgeID)
	}
	progress := float64(at-action.StartAt) / float64(action.ResolveAt-action.StartAt)
	location, err := resolveOnEdgeLocation(state, edge, action.From.Edge.FromNode, action.From.Edge.ToNode, progress, action.ID)
	if err != nil {
		return err
	}
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player != nil && player.Alive && player.Action.CurrentActionID == action.ID {
			player.Location = playerLocation{Edge: cloneOnEdge(location)}
			if player.HasBomb {
				state.Bomb.Location = clonePlayerLocation(player.Location)
			}
		}
	}
	return nil
}

func completeMovement(state *roundState, action scheduledAction, actorIDs []string) error {
	if state.Timeline != action.ResolveAt {
		return newError("INVALID_MOVE", "movement completion is not at ResolveAt")
	}
	if _, ok := state.Nodes[action.ToNodeID]; !ok {
		return newError("CONFIG_UNREACHABLE_NODE", "movement target %s is missing", action.ToNodeID)
	}
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.Action.CurrentActionID != action.ID {
			continue
		}
		player.Location = playerLocation{NodeID: action.ToNodeID}
		if player.HasBomb {
			state.Bomb.Location = clonePlayerLocation(player.Location)
		}
		player.Posture = postureDefault
	}
	completeActionForActors(state, action, actorIDs)
	return nil
}

func resolveOnEdgeLocation(state *roundState, edge MapEdge, fromNode, toNode string, progress float64, identity string) (*onEdgeLocation, error) {
	from, fromOK := state.Nodes[fromNode]
	to, toOK := state.Nodes[toNode]
	if !fromOK || !toOK {
		return nil, newError("CONFIG_UNREACHABLE_NODE", "edge %s has missing endpoints", edge.ID)
	}
	progress = clampProbability(progress)
	x := from.Node.X + (to.Node.X-from.Node.X)*progress
	y := from.Node.Y + (to.Node.Y-from.Node.Y)*progress
	curve := math.Sin(math.Pi * progress)
	x = clampProbability(x + stableSignedUnit(state.Seed, "edge_x", edge.ID, identity)*0.008*curve)
	y = clampProbability(y + stableSignedUnit(state.Seed, "edge_y", edge.ID, identity)*0.008*curve)
	return &onEdgeLocation{
		EdgeID: edge.ID, FromNode: fromNode, ToNode: toNode, Progress: progress, X: x, Y: y,
		DisplayName: from.Node.Name + " → " + to.Node.Name,
	}, nil
}

func findBoundedPath(config *MapConfig, fromNode, toNode string, maxExpansions int) (pathResult, *decisionFeedback, error) {
	if config == nil || config.Nodes[fromNode].ID == "" || config.Nodes[toNode].ID == "" {
		return pathResult{}, nil, newError("CONFIG_UNREACHABLE_NODE", "path endpoint is not configured: %s -> %s", fromNode, toNode)
	}
	if fromNode == toNode {
		return pathResult{NodeIDs: []string{fromNode}}, nil, nil
	}
	if maxExpansions <= 0 {
		maxExpansions = len(config.Nodes) * 4
	}
	adjacency := buildAdjacency(config)
	queue := &pathHeap{{nodeID: fromNode, distance: 0, nodeIDs: []string{fromNode}}}
	heap.Init(queue)
	bestDistance := map[string]int{fromNode: 0}
	bestKey := map[string]string{fromNode: ""}
	expansions := 0
	for queue.Len() > 0 && expansions < maxExpansions {
		current := heap.Pop(queue).(pathCandidate)
		if current.distance != bestDistance[current.nodeID] || strings.Join(current.edgeIDs, ",") != bestKey[current.nodeID] {
			continue
		}
		expansions++
		if current.nodeID == toNode {
			return pathResult{NodeIDs: current.nodeIDs, EdgeIDs: current.edgeIDs, TotalBaseTime: current.distance}, nil, nil
		}
		for _, next := range adjacency[current.nodeID] {
			distance := current.distance + next.weight
			edgeIDs := append(append([]string(nil), current.edgeIDs...), next.edgeID)
			key := strings.Join(edgeIDs, ",")
			oldDistance, seen := bestDistance[next.toNode]
			if seen && (distance > oldDistance || (distance == oldDistance && key >= bestKey[next.toNode])) {
				continue
			}
			bestDistance[next.toNode], bestKey[next.toNode] = distance, key
			nodeIDs := append(append([]string(nil), current.nodeIDs...), next.toNode)
			heap.Push(queue, pathCandidate{nodeID: next.toNode, distance: distance, edgeIDs: edgeIDs, nodeIDs: nodeIDs})
		}
	}
	return pathResult{}, &decisionFeedback{Code: "UNREACHABLE", Message: "no configured semantic path from " + fromNode + " to " + toNode}, nil
}

func scheduleInterceptCheck(state *roundState, movement scheduledAction, context interceptContext) (interceptCheckResult, error) {
	if state == nil || movement.ID == "" || movement.Type != actionMovementArrive {
		return interceptCheckResult{}, newError("INVALID_INTERCEPT", "intercept requires a movement traversal")
	}
	if !state.Scheduler.claimInterceptCheck(movement.ID) {
		return interceptCheckResult{ReasonCode: "ALREADY_EVALUATED"}, nil
	}
	edge, ok := stateEdge(state, movement.Payload.EdgeID)
	if !ok {
		return interceptCheckResult{}, newError("CONFIG_MISSING_EDGE", "edge %s is missing", movement.Payload.EdgeID)
	}
	enemies := legalInterceptEnemies(state, movement.ActorIDs, context.EnemyActorIDs)
	probability := interceptProbability(edge, context)
	roll := stableUnit(state.Seed, "intercept_roll", movement.ID, edge.ID)
	result := interceptCheckResult{Probability: probability, Roll: roll, ReasonCode: "ROLL_MISSED"}
	if len(enemies) == 0 {
		result.ReasonCode = "NO_LEGAL_OPPONENTS"
		return result, nil
	}
	if roll >= probability {
		return result, nil
	}
	window := movement.ResolveAt - movement.StartAt
	resolveAt := movement.StartAt + maxInt(1, int(math.Round(float64(window)*(0.25+stableUnit(state.Seed, "intercept_time", movement.ID)*0.5))))
	if resolveAt >= movement.ResolveAt {
		resolveAt = movement.ResolveAt - 1
	}
	action := scheduledAction{
		ID:             newActionID(state.Seed, actionInterceptCheck, movement.ID, movement.StartAt, resolveAt, movement.ActorIDs, 0),
		ParentActionID: movement.ID, IntentID: movement.IntentID, Type: actionInterceptCheck, ActorIDs: append([]string(nil), movement.ActorIDs...),
		From: movement.From, StartAt: movement.StartAt, ResolveAt: resolveAt, Priority: 60, VersionByActor: copyIntMap(movement.VersionByActor),
		Payload: actionPayload{EdgeID: edge.ID, ParticipantIDs: enemies},
	}
	if err := state.Scheduler.Schedule(action); err != nil {
		return interceptCheckResult{}, err
	}
	result.Scheduled, result.ActionID, result.ReasonCode = true, action.ID, "INTERCEPT_SCHEDULED"
	return result, nil
}

func resolveInterceptCheck(state *roundState, action scheduledAction) (*encounterCandidate, error) {
	if action.Type != actionInterceptCheck || action.ParentActionID == "" || state.Timeline != action.ResolveAt {
		return nil, newError("INVALID_INTERCEPT", "invalid intercept action")
	}
	movers := validActionActors(state, action)
	if len(movers) == 0 {
		return nil, nil
	}
	if err := updateParentMovementProgress(state, action, movers); err != nil {
		return nil, err
	}
	enemies := legalInterceptEnemies(state, movers, action.Payload.ParticipantIDs)
	if len(enemies) == 0 {
		return nil, nil
	}
	actors := append(append([]string(nil), movers...), enemies...)
	sort.Strings(actors)
	return &encounterCandidate{SourceActionID: action.ID, NodeID: action.Payload.EdgeID, ActorIDs: actors, StartedAt: state.Timeline}, nil
}

func interruptMovementForEncounter(state *roundState, movement scheduledAction, actorIDs []string, at int) error {
	if err := updateMovementProgress(state, movement, at, actorIDs); err != nil {
		return err
	}
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player == nil || player.Action.CurrentActionID != movement.ID {
			continue
		}
		player.Action.Version++
		player.Action.CurrentActionID = ""
		player.Action.Status = actionEngaged
		player.Action.BusyUntil = at
		player.Action.Busy = busyInterval{}
		player.Posture = postureEngaged
	}
	return nil
}

func updateParentMovementProgress(state *roundState, intercept scheduledAction, actorIDs []string) error {
	parent := scheduledAction{
		ID: intercept.ParentActionID, Type: actionMovementArrive, ActorIDs: intercept.ActorIDs, From: intercept.From,
		StartAt: intercept.StartAt, ResolveAt: state.Scheduler.resolveAt(intercept.ParentActionID), Payload: actionPayload{EdgeID: intercept.Payload.EdgeID},
	}
	if parent.ResolveAt <= parent.StartAt {
		return newError("INVALID_INTERCEPT", "parent movement timing is unavailable")
	}
	return updateMovementProgress(state, parent, intercept.ResolveAt, actorIDs)
}

func runtimeLocation(state *roundState, location playerLocation) *Location {
	if location.Edge != nil {
		return &Location{Name: location.Edge.DisplayName, X: location.Edge.X, Y: location.Edge.Y}
	}
	if node := state.Nodes[location.NodeID]; node != nil {
		return &Location{Name: node.Node.Name, X: clampProbability(node.Node.X), Y: clampProbability(node.Node.Y)}
	}
	return nil
}

func stateEdge(state *roundState, edgeID string) (MapEdge, bool) {
	edge, ok := state.mapEdges[edgeID]
	return edge, ok
}

func edgeDestination(edge MapEdge, fromNode string) (string, bool) {
	if edge.FromNode == fromNode {
		return edge.ToNode, true
	}
	if edge.Bidirectional && edge.ToNode == fromNode {
		return edge.FromNode, true
	}
	return "", false
}

type pathArc struct {
	edgeID string
	toNode string
	weight int
}

func buildAdjacency(config *MapConfig) map[string][]pathArc {
	adjacency := make(map[string][]pathArc, len(config.Nodes))
	edgeIDs := make([]string, 0, len(config.Edges))
	for edgeID := range config.Edges {
		edgeIDs = append(edgeIDs, edgeID)
	}
	sort.Strings(edgeIDs)
	for _, edgeID := range edgeIDs {
		edge := config.Edges[edgeID]
		adjacency[edge.FromNode] = append(adjacency[edge.FromNode], pathArc{edgeID: edgeID, toNode: edge.ToNode, weight: edge.BaseTime})
		if edge.Bidirectional {
			adjacency[edge.ToNode] = append(adjacency[edge.ToNode], pathArc{edgeID: edgeID, toNode: edge.FromNode, weight: edge.BaseTime})
		}
	}
	return adjacency
}

type pathCandidate struct {
	nodeID   string
	distance int
	edgeIDs  []string
	nodeIDs  []string
}
type pathHeap []pathCandidate

func (h pathHeap) Len() int { return len(h) }
func (h pathHeap) Less(i, j int) bool {
	if h[i].distance != h[j].distance {
		return h[i].distance < h[j].distance
	}
	iKey, jKey := strings.Join(h[i].edgeIDs, ","), strings.Join(h[j].edgeIDs, ",")
	if iKey != jKey {
		return iKey < jKey
	}
	return h[i].nodeID < h[j].nodeID
}
func (h pathHeap) Swap(i, j int)           { h[i], h[j] = h[j], h[i] }
func (h *pathHeap) Push(value interface{}) { *h = append(*h, value.(pathCandidate)) }
func (h *pathHeap) Pop() interface{} {
	old := *h
	value := old[len(old)-1]
	*h = old[:len(old)-1]
	return value
}

func interceptProbability(edge MapEdge, context interceptContext) float64 {
	probability := float64(edge.Risk)/250 + float64(edge.Noise)/500 + float64(clampInt(context.KnownEnemyConfidence, 0, 100))/1000
	if context.Visible {
		probability += 0.2
	}
	if context.ObserverPosture == postureHolding {
		probability += 0.1
	}
	return clampProbability(probability)
}

func legalInterceptEnemies(state *roundState, movers, candidates []string) []string {
	moverSides := map[string]bool{}
	for _, actorID := range movers {
		if player := state.Players[actorID]; player != nil {
			moverSides[player.Side] = true
		}
	}
	var out []string
	for _, actorID := range candidates {
		player := state.Players[actorID]
		if player != nil && player.Alive && !moverSides[player.Side] {
			out = append(out, actorID)
		}
	}
	sort.Strings(out)
	return out
}

func stableUnit(parts ...interface{}) float64 {
	return float64(uint64(deriveSeed(parts...))%1000000) / 1000000
}

func stableSignedUnit(parts ...interface{}) float64 { return stableUnit(parts...)*2 - 1 }

func cloneOnEdge(edge *onEdgeLocation) *onEdgeLocation {
	if edge == nil {
		return nil
	}
	copy := *edge
	return &copy
}
