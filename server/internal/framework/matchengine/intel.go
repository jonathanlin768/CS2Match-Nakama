package matchengine

import "sort"

type IntelType string

const (
	IntelDirectVisibility IntelType = "DirectVisibility"
	IntelEncounter        IntelType = "Encounter"
	IntelSound            IntelType = "Sound"
	IntelDeath            IntelType = "Death"
	IntelEmptySite        IntelType = "EmptySiteAssumption"
	IntelBomb             IntelType = "Bomb"
	IntelControl          IntelType = "Control"
)

type IntelObservation struct {
	Type           IntelType
	TargetID       string
	NodeID         string
	AreaID         string
	SourceActionID string
	SourceEventID  string
	ObservedBy     []string
	Confidence     int
	At             int
	TTL            int
}

type DecisionPlayerView struct {
	PlayerID   string
	TeamID     string
	Side       string
	Alive      bool
	HP         int
	Stamina    int
	Focus      int
	Location   PlayerLocation
	Intent     Intent
	Action     PlayerActionState
	Attributes PlayerAttributes
	RoleTags   []string
	HasBomb    bool
}

type DecisionControlView struct {
	NodeID    string
	Status    ControlStatus
	UpdatedAt int
	ExpiresAt int
}

type DecisionView struct {
	Side                string
	Timeline            int
	RoundDeadline       int
	BombDeadline        int
	OwnPlayers          []DecisionPlayerView
	PublicDeadPlayerIDs []string
	KnownControls       []DecisionControlView
	Intel               []IntelRecord
	BombIntel           *IntelRecord
	BombStatus          BombRuntimeStatus
	BombNodeID          string
	BombSite            string
	DecisionCount       int
	RotationCount       int
}

type RollSource interface {
	Unit(parts ...interface{}) float64
}

type IdentityRollSource struct{ Seed int64 }

func (source IdentityRollSource) Unit(parts ...interface{}) float64 {
	all := make([]interface{}, 0, len(parts)+1)
	all = append(all, source.Seed)
	all = append(all, parts...)
	return stableUnit(all...)
}

func RecordIntel(state *RoundState, side string, observation IntelObservation) (IntelRecord, error) {
	if state == nil || !validSide(side) || state.Intel[side] == nil {
		return IntelRecord{}, newError("INVALID_INTEL", "intel requires a valid round side")
	}
	if state.constants.Int("CommunicationDelay", -1) != 0 {
		return IntelRecord{}, newError("CONFIG_UNSUPPORTED_COMMUNICATION_DELAY", "CommunicationDelay must be zero")
	}
	if observation.At < 0 || observation.At > state.Timeline {
		return IntelRecord{}, newError("INVALID_INTEL", "intel observation timestamp is not observable")
	}
	confidence, err := normalizeIntelConfidence(state, side, observation)
	if err != nil {
		return IntelRecord{}, err
	}
	minTTL, maxTTL := state.constants.Int("MinIntelTTL", 1), state.constants.Int("MaxIntelTTL", 1)
	ttl := clampInt(observation.TTL, minTTL, maxTTL)
	observers := validObservers(state, side, observation.ObservedBy)
	if len(observers) == 0 && observation.Type != IntelEmptySite {
		return IntelRecord{}, newError("INVALID_INTEL", "intel has no live friendly observer")
	}
	record := IntelRecord{
		ID:   stableObjectID("intel", state.Seed, side, string(observation.Type), observation.TargetID, observation.NodeID, observation.AreaID, observation.SourceActionID, observation.SourceEventID, observation.At),
		Type: string(observation.Type), TargetID: observation.TargetID, NodeID: observation.NodeID, AreaID: observation.AreaID,
		Source: intelSource(observation), SourceActionID: observation.SourceActionID, SourceEventID: observation.SourceEventID,
		Confidence: confidence, ObservedBy: observers, LastSeenAt: observation.At, ExpiresAt: observation.At + ttl,
	}
	teamIntel := state.Intel[side]
	teamIntel.Records = append(teamIntel.Records, record)
	sortIntelRecords(teamIntel.Records)
	rebuildIntelIndexes(teamIntel, state.Timeline)
	return record, nil
}

func normalizeIntelConfidence(state *RoundState, side string, observation IntelObservation) (int, error) {
	confidence := observation.Confidence
	switch observation.Type {
	case IntelDirectVisibility:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "direct visibility requires an observed enemy and exact node")
		}
		if confidence == 0 {
			confidence = 100
		}
		confidence = clampInt(confidence, 1, 100)
	case IntelEncounter:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.SourceActionID == "" {
			return 0, newError("INVALID_INTEL", "encounter intel requires enemy and source action")
		}
		if confidence == 0 {
			confidence = 75
		}
		confidence = clampInt(confidence, 1, 89)
	case IntelSound:
		if observation.SourceEventID == "" || !hasEventID(state.Events, observation.SourceEventID) || observation.AreaID == "" || observation.NodeID != "" {
			return 0, newError("INVALID_INTEL", "sound intel requires a real event and area-only location")
		}
		if observation.TargetID != "" {
			if err := validateEnemyTarget(state, side, observation.TargetID); err != nil {
				return 0, err
			}
		}
		confidence = clampInt(confidence, state.constants.Int("SoundIntelMinConfidence", 30), state.constants.Int("SoundIntelMaxConfidence", 70))
	case IntelDeath:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.SourceActionID == "" {
			return 0, newError("INVALID_INTEL", "death intel requires a real attacker and source action")
		}
		if confidence == 0 {
			confidence = state.constants.Int("DeathIntelMaxConfidence", 70)
		}
		confidence = clampInt(confidence, 1, state.constants.Int("DeathIntelMaxConfidence", 70))
	case IntelEmptySite:
		if observation.TargetID != "" || observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "empty-site assumption cannot name an enemy")
		}
		if confidence == 0 {
			confidence = 20
		}
		confidence = clampInt(confidence, 1, 29)
	case IntelBomb:
		if observation.SourceEventID == "" || !hasEventID(state.Events, observation.SourceEventID) {
			return 0, newError("INVALID_INTEL", "bomb intel requires a real bomb event")
		}
		if observation.NodeID == "" && observation.AreaID == "" {
			return 0, newError("INVALID_INTEL", "bomb intel requires an observed area or node")
		}
		confidence = clampInt(confidence, 1, 100)
	case IntelControl:
		if observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "control intel requires a node")
		}
		confidence = clampInt(confidence, 1, 100)
	default:
		return 0, newError("INVALID_INTEL", "unsupported intel type %s", observation.Type)
	}
	return confidence, nil
}

func DegradeIntelForObserverDeath(state *RoundState, side, observerID string) {
	if state == nil || state.Intel[side] == nil {
		return
	}
	for index := range state.Intel[side].Records {
		record := &state.Intel[side].Records[index]
		record.ObservedBy = removeSortedString(record.ObservedBy, observerID)
		if record.Confidence > 1 {
			record.Confidence = maxInt(1, record.Confidence-20)
		}
	}
	rebuildIntelIndexes(state.Intel[side], state.Timeline)
}

func DecayIntelAndControl(state *RoundState, at int) error {
	if state == nil || at < state.Timeline {
		return newError("INVALID_INTEL", "decay cannot run before current timeline")
	}
	for _, side := range []string{SideT, SideCT} {
		teamIntel := state.Intel[side]
		kept := teamIntel.Records[:0]
		for _, record := range teamIntel.Records {
			if record.ExpiresAt > at {
				kept = append(kept, record)
			}
		}
		teamIntel.Records = kept
		rebuildIntelIndexes(teamIntel, at)
	}
	for _, node := range state.Nodes {
		node.DecayKnownControl(SideT, at)
		node.DecayKnownControl(SideCT, at)
	}
	return nil
}

func BuildDecisionView(state *RoundState, side string) (DecisionView, error) {
	if state == nil || !validSide(side) {
		return DecisionView{}, newError("INVALID_DECISION_VIEW", "decision view requires a valid side")
	}
	view := DecisionView{Side: side, Timeline: state.Timeline, RoundDeadline: state.RoundDeadline, BombDeadline: state.BombDeadline, DecisionCount: state.DecisionCount, RotationCount: state.RotationCount[side]}
	playerIDs := make([]string, 0, len(state.Players))
	for playerID := range state.Players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)
	for _, playerID := range playerIDs {
		player := state.Players[playerID]
		if player.Side == side {
			view.OwnPlayers = append(view.OwnPlayers, DecisionPlayerView{
				PlayerID: playerID, TeamID: player.TeamID, Side: player.Side, Alive: player.Alive, HP: player.HP, Stamina: player.Stamina,
				Focus: player.Focus, Location: clonePlayerLocation(player.Location), Intent: player.Intent, Action: player.Action,
				Attributes: player.Profile.Attributes, RoleTags: append([]string(nil), player.Profile.RoleTags...), HasBomb: player.HasBomb,
			})
		} else if !player.Alive {
			view.PublicDeadPlayerIDs = append(view.PublicDeadPlayerIDs, playerID)
		}
	}
	nodeIDs := make([]string, 0, len(state.Nodes))
	for nodeID := range state.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	for _, nodeID := range nodeIDs {
		known, ok := state.Nodes[nodeID].KnownControl[side]
		if !ok || known.ExpiresAt <= state.Timeline || known.Status == ControlUnknown {
			continue
		}
		view.KnownControls = append(view.KnownControls, DecisionControlView{NodeID: nodeID, Status: known.Status, UpdatedAt: known.UpdatedAt, ExpiresAt: known.ExpiresAt})
	}
	for _, record := range state.Intel[side].Records {
		if record.ExpiresAt <= state.Timeline {
			continue
		}
		view.Intel = append(view.Intel, cloneIntelRecord(record))
	}
	sortIntelRecords(view.Intel)
	if bomb := state.Intel[side].BombIntel; bomb != nil && bomb.ExpiresAt > state.Timeline {
		copy := cloneIntelRecord(*bomb)
		view.BombIntel = &copy
	}
	if side == SideT || state.Bomb.Status == BombPlanted || state.Bomb.Status == BombDefusing || state.Bomb.Status == BombDefused || state.Bomb.Status == BombExploded {
		view.BombStatus, view.BombNodeID, view.BombSite = state.Bomb.Status, projectedNodeID(state.Bomb.Location), state.Bomb.PlantedSite
	} else if view.BombIntel != nil {
		view.BombStatus, view.BombNodeID = state.Bomb.Status, view.BombIntel.NodeID
	}
	return view, nil
}

func IntelScoreModifier(record IntelRecord, maxMagnitude float64) float64 {
	return maxMagnitude * float64(clampInt(record.Confidence, 0, 100)) / 100
}

func CanTriggerDeterministicIntelAction(record IntelRecord, constants CombatConstants) bool {
	if record.Type == string(IntelEmptySite) {
		return false
	}
	return record.Confidence >= constants.Int("SoundIntelMinConfidence", 30)
}

func rebuildIntelIndexes(teamIntel *TeamIntel, timeline int) {
	teamIntel.KnownEnemies = map[string]IntelRecord{}
	teamIntel.KnownControl = map[string]IntelRecord{}
	teamIntel.SoundCues = nil
	teamIntel.BombIntel = nil
	for _, record := range teamIntel.Records {
		if record.ExpiresAt <= timeline {
			continue
		}
		if record.TargetID != "" {
			current, ok := teamIntel.KnownEnemies[record.TargetID]
			if !ok || betterIntel(record, current) {
				teamIntel.KnownEnemies[record.TargetID] = cloneIntelRecord(record)
			}
		}
		if record.Type == string(IntelControl) || record.Type == string(IntelEmptySite) {
			current, ok := teamIntel.KnownControl[record.NodeID]
			if !ok || betterIntel(record, current) {
				teamIntel.KnownControl[record.NodeID] = cloneIntelRecord(record)
			}
		}
		if record.Type == string(IntelSound) {
			teamIntel.SoundCues = append(teamIntel.SoundCues, cloneIntelRecord(record))
		}
		if record.Type == string(IntelBomb) && (teamIntel.BombIntel == nil || betterIntel(record, *teamIntel.BombIntel)) {
			copy := cloneIntelRecord(record)
			teamIntel.BombIntel = &copy
		}
	}
}

func betterIntel(candidate, current IntelRecord) bool {
	if candidate.LastSeenAt != current.LastSeenAt {
		return candidate.LastSeenAt > current.LastSeenAt
	}
	if candidate.Confidence != current.Confidence {
		return candidate.Confidence > current.Confidence
	}
	return candidate.ID < current.ID
}

func sortIntelRecords(records []IntelRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ID < records[j].ID })
}

func cloneIntelRecord(record IntelRecord) IntelRecord {
	record.ObservedBy = append([]string(nil), record.ObservedBy...)
	return record
}

func clonePlayerLocation(location PlayerLocation) PlayerLocation {
	location.Edge = cloneOnEdge(location.Edge)
	return location
}

func validObservers(state *RoundState, side string, observers []string) []string {
	var out []string
	for _, observerID := range observers {
		player := state.Players[observerID]
		if player != nil && player.Alive && player.Side == side {
			out = append(out, observerID)
		}
	}
	sort.Strings(out)
	return uniqueStrings(out)
}

func validateEnemyTarget(state *RoundState, side, targetID string) error {
	target := state.Players[targetID]
	if target == nil || target.Side == side {
		return newError("INVALID_INTEL", "intel target %s is not a real enemy", targetID)
	}
	return nil
}

func hasEventID(events []*GameEvent, eventID string) bool {
	for _, event := range events {
		if event != nil && event.EventID == eventID {
			return true
		}
	}
	return false
}

func intelSource(observation IntelObservation) string {
	if observation.SourceEventID != "" {
		return observation.SourceEventID
	}
	return observation.SourceActionID
}

func removeSortedString(values []string, remove string) []string {
	out := values[:0]
	for _, value := range values {
		if value != remove {
			out = append(out, value)
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
