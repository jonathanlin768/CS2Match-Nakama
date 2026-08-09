package matchengine

import (
	"container/heap"
	"sort"
)

type scheduledActionHeap []ScheduledAction

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
func (h *scheduledActionHeap) Push(value interface{}) { *h = append(*h, value.(ScheduledAction)) }
func (h *scheduledActionHeap) Pop() interface{} {
	old := *h
	value := old[len(old)-1]
	*h = old[:len(old)-1]
	return value
}

type ActionScheduler struct {
	actions              scheduledActionHeap
	scheduledCount       int
	maxScheduled         int
	interceptByTraversal map[string]bool
}

func NewActionScheduler(constants CombatConstants) *ActionScheduler {
	scheduler := &ActionScheduler{maxScheduled: constants.Int("MaxScheduledActions", 0), interceptByTraversal: map[string]bool{}}
	heap.Init(&scheduler.actions)
	return scheduler
}

func (s *ActionScheduler) Len() int { return len(s.actions) }

func (s *ActionScheduler) Schedule(action ScheduledAction) error {
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

func (s *ActionScheduler) Peek() (ScheduledAction, bool) {
	if len(s.actions) == 0 {
		return ScheduledAction{}, false
	}
	return s.actions[0], true
}

func (s *ActionScheduler) Pop() (ScheduledAction, bool) {
	if len(s.actions) == 0 {
		return ScheduledAction{}, false
	}
	return heap.Pop(&s.actions).(ScheduledAction), true
}

func (s *ActionScheduler) PopNextValid(state *RoundState) (ScheduledAction, []string, bool) {
	for {
		action, ok := s.Pop()
		if !ok {
			return ScheduledAction{}, nil, false
		}
		validActors := validActionActors(state, action)
		if len(action.ActorIDs) == 0 || len(validActors) >= maxInt(1, action.MinRequiredActors) {
			return action, validActors, true
		}
		cancelActionForActors(state, action)
	}
}

func BeginExclusiveAction(state *RoundState, action *ScheduledAction, status ActionStatus) error {
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
		player.Action.Busy = BusyInterval{ActionID: action.ID, StartAt: action.StartAt, EndAt: action.ResolveAt}
		action.VersionByActor[actorID] = player.Action.Version
	}
	if action.MinRequiredActors <= 0 {
		action.MinRequiredActors = len(action.ActorIDs)
	}
	return nil
}

func CompleteActionForActors(state *RoundState, action ScheduledAction, actorIDs []string) {
	for _, actorID := range actorIDs {
		player := state.Players[actorID]
		if player == nil || player.Action.CurrentActionID != action.ID {
			continue
		}
		player.Action.CurrentActionID = ""
		player.Action.Status = ActionIdle
		player.Action.BusyUntil = state.Timeline
		player.Action.Busy = BusyInterval{}
	}
}

func validActionActors(state *RoundState, action ScheduledAction) []string {
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

func locationMatchesAction(current, expected PlayerLocation, actionType ActionType) bool {
	if expected.Edge != nil && (actionType == ActionMovementArrive || actionType == ActionInterceptCheck) {
		return current.Edge != nil && current.Edge.EdgeID == expected.Edge.EdgeID && current.Edge.FromNode == expected.Edge.FromNode && current.Edge.ToNode == expected.Edge.ToNode
	}
	return sameLocation(current, expected)
}

func cancelActionForActors(state *RoundState, action ScheduledAction) {
	for _, actorID := range action.ActorIDs {
		player := state.Players[actorID]
		if player != nil && player.Action.CurrentActionID == action.ID {
			player.Action.Version++
			player.Action.CurrentActionID = ""
			player.Action.Status = ActionIdle
			player.Action.BusyUntil = state.Timeline
			player.Action.Busy = BusyInterval{}
		}
	}
}

type NextTimeKind string

const (
	NextTimeNone     NextTimeKind = "None"
	NextTimeAction   NextTimeKind = "Action"
	NextTimeRound    NextTimeKind = "RoundDeadline"
	NextTimeBomb     NextTimeKind = "BombDeadline"
	NextTimeIntel    NextTimeKind = "IntelDecay"
	NextTimeControl  NextTimeKind = "ControlDecay"
	NextTimeDecision NextTimeKind = "DecisionDeadline"
)

func (s *RoundState) NextTime() (int, NextTimeKind) {
	next, kind := int(^uint(0)>>1), NextTimeNone
	consider := func(at int, candidate NextTimeKind) {
		if at > s.Timeline && (at < next || (at == next && candidate < kind)) {
			next, kind = at, candidate
		}
	}
	if action, ok := s.Scheduler.Peek(); ok {
		consider(action.ResolveAt, NextTimeAction)
	}
	if s.Bomb.Status == BombPlanted || s.Bomb.Status == BombDefusing {
		consider(s.BombDeadline, NextTimeBomb)
	} else {
		consider(s.RoundDeadline-s.constants.Int("ForceExecuteThreshold", 0), NextTimeDecision)
		consider(s.RoundDeadline, NextTimeRound)
	}
	for _, intel := range s.Intel {
		for _, record := range intel.Records {
			consider(record.ExpiresAt, NextTimeIntel)
		}
	}
	for _, node := range s.Nodes {
		for _, known := range node.KnownControl {
			consider(known.ExpiresAt, NextTimeControl)
		}
	}
	if kind == NextTimeNone {
		return s.Timeline, kind
	}
	return next, kind
}

func (s *RoundState) RecordTransition() error {
	s.TransitionCount++
	if s.TransitionCount > s.constants.Int("MaxStateTransitions", 0) {
		return newError("STATE_TRANSITION_LIMIT_EXCEEDED", "MaxStateTransitions exceeded")
	}
	return nil
}

func (s *RoundState) RecordRotation(side string) error {
	s.RotationCount[side]++
	if s.RotationCount[side] > s.constants.Int("MaxRotationsPerTeam", 0) {
		return newError("ROTATION_LIMIT_EXCEEDED", "MaxRotationsPerTeam exceeded for %s", side)
	}
	return nil
}

func (s *RoundState) ValidateEffectBatchSize(count int) error {
	if count > s.constants.Int("MaxEffectsPerTimestamp", 0) {
		return newError("EFFECT_LIMIT_EXCEEDED", "MaxEffectsPerTimestamp exceeded")
	}
	return nil
}

func (s *RoundState) AdvanceTimeline(at int) error {
	if at < s.Timeline {
		return newError("INVALID_TIMELINE", "timeline cannot move backwards")
	}
	if at > s.constants.Int("MaxRoundTimeline", 0) {
		return newError("TIMELINE_LIMIT_EXCEEDED", "MaxRoundTimeline exceeded")
	}
	s.Timeline = at
	return nil
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

func sameLocation(a, b PlayerLocation) bool {
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
