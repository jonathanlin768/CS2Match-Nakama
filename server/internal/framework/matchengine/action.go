package matchengine

import (
	"fmt"
	"sort"
	"strings"
)

type intentType string
type actionType string
type actionStatus string
type effectType string

const (
	intentMove       intentType = "Move"
	intentHold       intentType = "Hold"
	intentEngage     intentType = "Engage"
	intentPlant      intentType = "Plant"
	intentDefuse     intentType = "Defuse"
	intentPickupBomb intentType = "PickupBomb"

	actionMoveStart       actionType = "MoveStart"
	actionMovementArrive  actionType = "MovementArrive"
	actionInterceptCheck  actionType = "InterceptCheck"
	actionHoldStart       actionType = "HoldStart"
	actionEncounterStart  actionType = "EncounterStart"
	actionCombatPulse     actionType = "CombatPulse"
	actionCombatEnd       actionType = "CombatEnd"
	actionDecisionResolve actionType = "DecisionResolve"
	actionPlantStart      actionType = "PlantStart"
	actionPlantComplete   actionType = "PlantComplete"
	actionPickupComplete  actionType = "PickupComplete"
	actionDefuseStart     actionType = "DefuseStart"
	actionDefuseComplete  actionType = "DefuseComplete"
	actionBombExplode     actionType = "BombExplode"
	actionRoundExpire     actionType = "RoundExpire"
	actionIntelDecay      actionType = "IntelDecay"
	actionControlDecay    actionType = "ControlDecay"

	actionIdle     actionStatus = "Idle"
	actionMoving   actionStatus = "Moving"
	actionHolding  actionStatus = "Holding"
	actionEngaged  actionStatus = "Engaged"
	actionPlanting actionStatus = "Planting"
	actionDefusing actionStatus = "Defusing"

	effectDamage    effectType = "Damage"
	effectMiss      effectType = "Miss"
	effectDeath     effectType = "Death"
	effectBombDrop  effectType = "BombDrop"
	effectMove      effectType = "Move"
	effectControl   effectType = "Control"
	effectBombState effectType = "BombState"
)

type intent struct {
	ID        string
	Type      intentType
	TargetID  string
	Priority  int
	CreatedAt int
}

type onEdgeLocation struct {
	EdgeID      string
	FromNode    string
	ToNode      string
	Progress    float64
	X           float64
	Y           float64
	DisplayName string
}

type playerLocation struct {
	NodeID string
	Edge   *onEdgeLocation
}

func (l playerLocation) Valid() bool {
	return (l.NodeID != "") != (l.Edge != nil)
}

type busyInterval struct {
	ActionID string
	StartAt  int
	EndAt    int
}

type playerActionState struct {
	CurrentActionID string
	Version         int
	Status          actionStatus
	BusyUntil       int
	Busy            busyInterval
}

type actionPayload struct {
	ScenarioID     string
	Site           string
	EdgeID         string
	TargetID       string
	RouteID        string
	DecisionType   string
	ParticipantIDs []string
}

type scheduledAction struct {
	ID                string
	ParentActionID    string
	IntentID          string
	Type              actionType
	ActorIDs          []string
	From              playerLocation
	ToNodeID          string
	StartAt           int
	ResolveAt         int
	Priority          int
	VersionByActor    map[string]int
	MinRequiredActors int
	Payload           actionPayload
}

func (a scheduledAction) MinActorID() string {
	if len(a.ActorIDs) == 0 {
		return ""
	}
	actors := append([]string(nil), a.ActorIDs...)
	sort.Strings(actors)
	return actors[0]
}

type effect struct {
	ID             string
	SourceActionID string
	Type           effectType
	Priority       int
	Timestamp      int
	ActorID        string
	TargetID       string
	Amount         int
	NodeID         string
	StringValue    string
	ReasonRecords  []reasonRecord
}

type appliedEffect struct {
	Effect        effect
	AppliedAmount int
}

type appliedBatch struct {
	Timestamp int
	Effects   []appliedEffect
	Events    []*GameEvent
}

func stableObjectID(prefix string, parts ...interface{}) string {
	return fmt.Sprintf("%s_%016x", prefix, uint64(deriveSeed(parts...)))
}

func newActionID(roundSeed int64, actionType actionType, intentID string, startAt, resolveAt int, actorIDs []string, ordinal int) string {
	actors := append([]string(nil), actorIDs...)
	sort.Strings(actors)
	return stableObjectID("act", roundSeed, string(actionType), intentID, startAt, resolveAt, strings.Join(actors, ","), ordinal)
}

func newEffectID(roundSeed int64, actionID string, effectType effectType, ordinal int) string {
	return stableObjectID("eff", roundSeed, actionID, string(effectType), ordinal)
}

func newEventID(roundSeed int64, actionID, effectID, eventType string, ordinal int) string {
	return stableObjectID("evt", roundSeed, "event", actionID, effectID, eventType, ordinal)
}

func newActionLifecycleEvent(state *roundState, action scheduledAction, eventType, message string, ordinal int) (*GameEvent, error) {
	if state == nil || action.ID == "" || eventType == "" {
		return nil, newError("INVALID_ACTION_EVENT", "action lifecycle event requires state, action and event type")
	}
	reason, err := projectReasonRecord(reasonRecord{Code: eventType, Source: string(action.Type), Value: 1, Weight: 1, SourceActionID: action.ID}, action.ID, "")
	if err != nil {
		return nil, err
	}
	event := &GameEvent{
		EventID: newEventID(state.Seed, action.ID, "", eventType, ordinal), SourceActionID: action.ID,
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
