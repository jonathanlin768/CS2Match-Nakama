package matchengine

import (
	"container/heap"
	"sort"
)

type scheduledActionHeap []scheduledAction

func (h scheduledActionHeap) Len() int { return len(h) }
func (h scheduledActionHeap) Less(i, j int) bool {
	a, b := h[i], h[j]
	if a.ResolveAt != b.ResolveAt {
		return a.ResolveAt < b.ResolveAt
	}
	if a.Priority != b.Priority {
		return a.Priority > b.Priority
	}
	if a.Type != b.Type {
		return a.Type < b.Type
	}
	if a.MinActorID() != b.MinActorID() {
		return a.MinActorID() < b.MinActorID()
	}
	return a.ID < b.ID
}
func (h scheduledActionHeap) Swap(i, j int)           { h[i], h[j] = h[j], h[i] }
func (h *scheduledActionHeap) Push(value interface{}) { *h = append(*h, value.(scheduledAction)) }
func (h *scheduledActionHeap) Pop() interface{} {
	old := *h
	value := old[len(old)-1]
	*h = old[:len(old)-1]
	return value
}

type actionScheduler struct {
	actions              scheduledActionHeap
	scheduledCount       int
	maxScheduled         int
	interceptByTraversal map[string]bool
}

func newActionScheduler(constants CombatConstants) *actionScheduler {
	scheduler := &actionScheduler{maxScheduled: constants.Int("MaxScheduledActions", 0), interceptByTraversal: map[string]bool{}}
	heap.Init(&scheduler.actions)
	return scheduler
}

func (s *actionScheduler) Len() int { return len(s.actions) }

// snapshot copies the queue slice so callers can inspect or sort it without
// changing heap order. Nested action payloads remain read-only.
func (s *actionScheduler) snapshot() []scheduledAction {
	return append([]scheduledAction(nil), s.actions...)
}

func (s *actionScheduler) hasType(kind actionType) bool {
	for _, action := range s.actions {
		if action.Type == kind {
			return true
		}
	}
	return false
}

func (s *actionScheduler) resolveAt(actionID string) int {
	for _, action := range s.actions {
		if action.ID == actionID {
			return action.ResolveAt
		}
	}
	return 0
}

func (s *actionScheduler) clear() { s.actions = nil }

// claimInterceptCheck reserves the one allowed check for a traversal before
// its remaining configuration and eligibility checks run.
func (s *actionScheduler) claimInterceptCheck(traversalID string) bool {
	if s.interceptByTraversal[traversalID] {
		return false
	}
	s.interceptByTraversal[traversalID] = true
	return true
}

func (s *actionScheduler) Schedule(action scheduledAction) error {
	if action.ID == "" || action.ResolveAt < action.StartAt || action.ResolveAt < 0 {
		return newError("INVALID_ACTION", "action has invalid identity or timing")
	}
	if s.maxScheduled <= 0 || s.scheduledCount >= s.maxScheduled {
		return newError("SCHEDULER_LIMIT_EXCEEDED", "MaxScheduledActions exceeded")
	}
	action.ActorIDs = append([]string(nil), action.ActorIDs...)
	sort.Strings(action.ActorIDs)
	action.VersionByActor = copyIntMap(action.VersionByActor)
	action.Payload.ParticipantIDs = append([]string(nil), action.Payload.ParticipantIDs...)
	sort.Strings(action.Payload.ParticipantIDs)
	heap.Push(&s.actions, action)
	s.scheduledCount++
	return nil
}

func (s *actionScheduler) Peek() (scheduledAction, bool) {
	if len(s.actions) == 0 {
		return scheduledAction{}, false
	}
	return s.actions[0], true
}

func (s *actionScheduler) Pop() (scheduledAction, bool) {
	if len(s.actions) == 0 {
		return scheduledAction{}, false
	}
	return heap.Pop(&s.actions).(scheduledAction), true
}

func (s *actionScheduler) PopNextValid(state *roundState) (scheduledAction, []string, bool) {
	for {
		action, ok := s.Pop()
		if !ok {
			return scheduledAction{}, nil, false
		}
		validActors := validActionActors(state, action)
		if len(action.ActorIDs) == 0 || len(validActors) >= maxInt(1, action.MinRequiredActors) {
			return action, validActors, true
		}
		cancelActionForActors(state, action)
	}
}

func beginExclusiveAction(state *roundState, action *scheduledAction, status actionStatus) error {
	if state == nil || action == nil || action.ID == "" || len(action.ActorIDs) == 0 {
		return newError("INVALID_ACTION", "exclusive action requires actors")
	}
	action.ActorIDs = append([]string(nil), action.ActorIDs...)
	sort.Strings(action.ActorIDs)
	action.VersionByActor = make(map[string]int, len(action.ActorIDs))
	for _, actorID := range action.ActorIDs {
		player, ok := state.Players[actorID]
		if !ok || !player.Alive || player.Action.CurrentActionID != "" {
			return newError("INVALID_ACTION", "actor %s is not available", actorID)
		}
	}
	for _, actorID := range action.ActorIDs {
		player := state.Players[actorID]
		player.Action.Version++
		player.Action.CurrentActionID = action.ID
		player.Action.Status = status
		player.Action.BusyUntil = action.ResolveAt
		player.Action.Busy = busyInterval{ActionID: action.ID, StartAt: action.StartAt, EndAt: action.ResolveAt}
		action.VersionByActor[actorID] = player.Action.Version
	}
	if action.MinRequiredActors <= 0 {
		action.MinRequiredActors = len(action.ActorIDs)
	}
	return nil
}

func completeActionForActors(state *roundState, action scheduledAction, actorIDs []string) {
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player == nil || player.Action.CurrentActionID != action.ID {
			continue
		}
		player.Action.CurrentActionID = ""
		player.Action.Status = actionIdle
		player.Action.BusyUntil = state.Timeline
		player.Action.Busy = busyInterval{}
	}
}

func validActionActors(state *roundState, action scheduledAction) []string {
	valid := make([]string, 0, len(action.ActorIDs))
	expectedActionID := action.ID
	if action.ParentActionID != "" {
		expectedActionID = action.ParentActionID
	}
	for _, actorID := range action.ActorIDs {
		player := state.Players[actorID]
		if player == nil || !player.Alive || player.Action.CurrentActionID != expectedActionID || player.Action.Version != action.VersionByActor[actorID] {
			continue
		}
		if action.From.Valid() && !locationMatchesAction(player.Location, action.From, action.Type) {
			continue
		}
		valid = append(valid, actorID)
	}
	sort.Strings(valid)
	return valid
}

func locationMatchesAction(current, expected playerLocation, actionType actionType) bool {
	if expected.Edge != nil && (actionType == actionMovementArrive || actionType == actionInterceptCheck) {
		return current.Edge != nil && current.Edge.EdgeID == expected.Edge.EdgeID && current.Edge.FromNode == expected.Edge.FromNode && current.Edge.ToNode == expected.Edge.ToNode
	}
	return sameLocation(current, expected)
}

func cancelActionForActors(state *roundState, action scheduledAction) {
	for _, actorID := range action.ActorIDs {
		player := state.Players[actorID]
		if player != nil && player.Action.CurrentActionID == action.ID {
			player.Action.Version++
			player.Action.CurrentActionID = ""
			player.Action.Status = actionIdle
			player.Action.BusyUntil = state.Timeline
			player.Action.Busy = busyInterval{}
		}
	}
}

func copyIntMap(source map[string]int) map[string]int {
	if source == nil {
		return nil
	}
	out := make(map[string]int, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func sameLocation(a, b playerLocation) bool {
	if a.NodeID != b.NodeID || (a.Edge == nil) != (b.Edge == nil) {
		return false
	}
	if a.Edge == nil {
		return true
	}
	return a.Edge.EdgeID == b.Edge.EdgeID && a.Edge.FromNode == b.Edge.FromNode && a.Edge.ToNode == b.Edge.ToNode && a.Edge.Progress == b.Edge.Progress
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
