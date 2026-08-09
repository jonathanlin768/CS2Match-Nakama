package matchengine

import (
	"fmt"
	"sort"
	"strings"
)

type IntentType string
type ActionType string
type ActionStatus string
type EffectType string

const (
	IntentMove       IntentType = "Move"
	IntentHold       IntentType = "Hold"
	IntentEngage     IntentType = "Engage"
	IntentPlant      IntentType = "Plant"
	IntentDefuse     IntentType = "Defuse"
	IntentPickupBomb IntentType = "PickupBomb"

	ActionMoveStart       ActionType = "MoveStart"
	ActionMovementArrive  ActionType = "MovementArrive"
	ActionInterceptCheck  ActionType = "InterceptCheck"
	ActionHoldStart       ActionType = "HoldStart"
	ActionEncounterStart  ActionType = "EncounterStart"
	ActionCombatPulse     ActionType = "CombatPulse"
	ActionCombatEnd       ActionType = "CombatEnd"
	ActionDecisionResolve ActionType = "DecisionResolve"
	ActionPlantStart      ActionType = "PlantStart"
	ActionPlantComplete   ActionType = "PlantComplete"
	ActionPickupComplete  ActionType = "PickupComplete"
	ActionDefuseStart     ActionType = "DefuseStart"
	ActionDefuseComplete  ActionType = "DefuseComplete"
	ActionBombExplode     ActionType = "BombExplode"
	ActionRoundExpire     ActionType = "RoundExpire"
	ActionIntelDecay      ActionType = "IntelDecay"
	ActionControlDecay    ActionType = "ControlDecay"

	ActionIdle     ActionStatus = "Idle"
	ActionMoving   ActionStatus = "Moving"
	ActionHolding  ActionStatus = "Holding"
	ActionEngaged  ActionStatus = "Engaged"
	ActionPlanting ActionStatus = "Planting"
	ActionDefusing ActionStatus = "Defusing"

	EffectDamage    EffectType = "Damage"
	EffectMiss      EffectType = "Miss"
	EffectDeath     EffectType = "Death"
	EffectBombDrop  EffectType = "BombDrop"
	EffectMove      EffectType = "Move"
	EffectControl   EffectType = "Control"
	EffectBombState EffectType = "BombState"
)

type Intent struct {
	ID        string
	Type      IntentType
	TargetID  string
	Priority  int
	CreatedAt int
}

type OnEdgeLocation struct {
	EdgeID      string
	FromNode    string
	ToNode      string
	Progress    float64
	X           float64
	Y           float64
	DisplayName string
}

type PlayerLocation struct {
	NodeID string
	Edge   *OnEdgeLocation
}

func (l PlayerLocation) Valid() bool {
	return (l.NodeID != "") != (l.Edge != nil)
}

type BusyInterval struct {
	ActionID string
	StartAt  int
	EndAt    int
}

type PlayerActionState struct {
	CurrentActionID string
	Version         int
	Status          ActionStatus
	BusyUntil       int
	Busy            BusyInterval
}

type ActionPayload struct {
	ScenarioID     string
	Site           string
	EdgeID         string
	TargetID       string
	RouteID        string
	DecisionType   string
	ParticipantIDs []string
}

type ScheduledAction struct {
	ID                string
	ParentActionID    string
	IntentID          string
	Type              ActionType
	ActorIDs          []string
	From              PlayerLocation
	ToNodeID          string
	StartAt           int
	ResolveAt         int
	Priority          int
	VersionByActor    map[string]int
	MinRequiredActors int
	Payload           ActionPayload
}

func (a ScheduledAction) MinActorID() string {
	if len(a.ActorIDs) == 0 {
		return ""
	}
	actors := append([]string(nil), a.ActorIDs...)
	sort.Strings(actors)
	return actors[0]
}

type Effect struct {
	ID             string
	SourceActionID string
	Type           EffectType
	Priority       int
	Timestamp      int
	ActorID        string
	TargetID       string
	Amount         int
	NodeID         string
	StringValue    string
	ReasonRecords  []ReasonRecord
}

type AppliedEffect struct {
	Effect        Effect
	AppliedAmount int
}

type AppliedBatch struct {
	Timestamp int
	Effects   []AppliedEffect
	Events    []*GameEvent
}

func stableObjectID(prefix string, parts ...interface{}) string {
	return fmt.Sprintf("%s_%016x", prefix, uint64(deriveSeed(parts...)))
}

func NewActionID(roundSeed int64, actionType ActionType, intentID string, startAt, resolveAt int, actorIDs []string, ordinal int) string {
	actors := append([]string(nil), actorIDs...)
	sort.Strings(actors)
	return stableObjectID("act", roundSeed, string(actionType), intentID, startAt, resolveAt, strings.Join(actors, ","), ordinal)
}

func NewEffectID(roundSeed int64, actionID string, effectType EffectType, ordinal int) string {
	return stableObjectID("eff", roundSeed, actionID, string(effectType), ordinal)
}

func NewEventID(roundSeed int64, actionID, effectID, eventType string, ordinal int) string {
	return stableObjectID("evt", roundSeed, "event", actionID, effectID, eventType, ordinal)
}

func newActionLifecycleEvent(state *RoundState, action ScheduledAction, eventType, message string, ordinal int) (*GameEvent, error) {
	if state == nil || action.ID == "" || eventType == "" {
		return nil, newError("INVALID_ACTION_EVENT", "action lifecycle event requires state, action and event type")
	}
	reason, err := ProjectReasonRecord(ReasonRecord{Code: eventType, Source: string(action.Type), Value: 1, Weight: 1, SourceActionID: action.ID}, action.ID, "")
	if err != nil {
		return nil, err
	}
	event := &GameEvent{
		EventID: NewEventID(state.Seed, action.ID, "", eventType, ordinal), SourceActionID: action.ID,
		Timestamp: int64(state.Timeline), EventType: eventType, Message: message, Reason: reason,
		sortPriority: action.Priority, sortActionType: string(action.Type), sortMinActorID: action.MinActorID(),
	}
	if len(action.ActorIDs) > 0 {
		if actor := state.Players[action.ActorIDs[0]]; actor != nil {
			event.AttackerID, event.AttackerName, event.AttackerTeamID = actor.Profile.PlayerID, actor.Profile.DisplayName, actor.TeamID
			event.Location = eventLocation(state, actor.Location, event.EventID, action.ID)
		}
	}
	return event, nil
}
