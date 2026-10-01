package matchengine

import "sort"

type intelType string

const (
	intelDirectVisibility intelType = "DirectVisibility"
	intelEncounter        intelType = "Encounter"
	intelSound            intelType = "Sound"
	intelDeath            intelType = "Death"
	intelEmptySite        intelType = "EmptySiteAssumption"
	intelBomb             intelType = "Bomb"
	intelControl          intelType = "Control"
)

type intelObservation struct {
	Type           intelType
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

type decisionPlayerView struct {
	PlayerID   string
	TeamID     string
	Side       string
	Alive      bool
	HP         int
	Stamina    int
	Focus      int
	Location   playerLocation
	Intent     intent
	Action     playerActionState
	Attributes PlayerAttributes
	RoleTags   []string
	HasBomb    bool
}

type decisionControlView struct {
	NodeID    string
	Status    controlStatus
	UpdatedAt int
	ExpiresAt int
}

type decisionView struct {
	Side                string
	Timeline            int
	RoundDeadline       int
	BombDeadline        int
	OwnPlayers          []decisionPlayerView
	PublicDeadPlayerIDs []string
	KnownControls       []decisionControlView
	Intel               []intelRecord
	BombIntel           *intelRecord
	BombStatus          bombRuntimeStatus
	BombNodeID          string
	BombSite            string
	DecisionCount       int
	RotationCount       int
}

type rollSource interface {
	Unit(parts ...interface{}) float64
}

type identityRollSource struct{ Seed int64 }

func (source identityRollSource) Unit(parts ...interface{}) float64 {
	all := make([]interface{}, 0, len(parts)+1)
	all = append(all, source.Seed)
	all = append(all, parts...)
	return stableUnit(all...)
}

func recordIntel(state *roundState, side string, observation intelObservation) (intelRecord, error) {
	if state == nil || !validSide(side) || state.Intel[side] == nil {
		return intelRecord{}, newError("INVALID_INTEL", "intel requires a valid round side")
	}
	if state.constants.Int("CommunicationDelay", -1) != 0 {
		return intelRecord{}, newError("CONFIG_UNSUPPORTED_COMMUNICATION_DELAY", "CommunicationDelay must be zero")
	}
	if observation.At < 0 || observation.At > state.Timeline {
		return intelRecord{}, newError("INVALID_INTEL", "intel observation timestamp is not observable")
	}
	confidence, err := normalizeIntelConfidence(state, side, observation)
	if err != nil {
		return intelRecord{}, err
	}
	minTTL, maxTTL := state.constants.Int("MinIntelTTL", 1), state.constants.Int("MaxIntelTTL", 1)
	ttl := clampInt(observation.TTL, minTTL, maxTTL)
	observers := validObservers(state, side, observation.ObservedBy)
	if len(observers) == 0 && observation.Type != intelEmptySite {
		return intelRecord{}, newError("INVALID_INTEL", "intel has no live friendly observer")
	}
	record := intelRecord{
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

func normalizeIntelConfidence(state *roundState, side string, observation intelObservation) (int, error) {
	confidence := observation.Confidence
	switch observation.Type {
	case intelDirectVisibility:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "direct visibility requires an observed enemy and exact node")
		}
		if confidence == 0 {
			confidence = 100
		}
		confidence = clampInt(confidence, 1, 100)
	case intelEncounter:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.SourceActionID == "" {
			return 0, newError("INVALID_INTEL", "encounter intel requires enemy and source action")
		}
		if confidence == 0 {
			confidence = 75
		}
		confidence = clampInt(confidence, 1, 89)
	case intelSound:
		if observation.SourceEventID == "" || !hasEventID(state.Events, observation.SourceEventID) || observation.AreaID == "" || observation.NodeID != "" {
			return 0, newError("INVALID_INTEL", "sound intel requires a real event and area-only location")
		}
		if observation.TargetID != "" {
			if err := validateEnemyTarget(state, side, observation.TargetID); err != nil {
				return 0, err
			}
		}
		confidence = clampInt(confidence, state.constants.Int("SoundIntelMinConfidence", 30), state.constants.Int("SoundIntelMaxConfidence", 70))
	case intelDeath:
		if err := validateEnemyTarget(state, side, observation.TargetID); err != nil || observation.SourceActionID == "" {
			return 0, newError("INVALID_INTEL", "death intel requires a real attacker and source action")
		}
		if confidence == 0 {
			confidence = state.constants.Int("DeathIntelMaxConfidence", 70)
		}
		confidence = clampInt(confidence, 1, state.constants.Int("DeathIntelMaxConfidence", 70))
	case intelEmptySite:
		if observation.TargetID != "" || observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "empty-site assumption cannot name an enemy")
		}
		if confidence == 0 {
			confidence = 20
		}
		confidence = clampInt(confidence, 1, 29)
	case intelBomb:
		if observation.SourceEventID == "" || !hasEventID(state.Events, observation.SourceEventID) {
			return 0, newError("INVALID_INTEL", "bomb intel requires a real bomb event")
		}
		if observation.NodeID == "" && observation.AreaID == "" {
			return 0, newError("INVALID_INTEL", "bomb intel requires an observed area or node")
		}
		confidence = clampInt(confidence, 1, 100)
	case intelControl:
		if observation.NodeID == "" {
			return 0, newError("INVALID_INTEL", "control intel requires a node")
		}
		confidence = clampInt(confidence, 1, 100)
	default:
		return 0, newError("INVALID_INTEL", "unsupported intel type %s", observation.Type)
	}
	return confidence, nil
}

func degradeIntelForObserverDeath(state *roundState, side, observerID string) {
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

func decayIntelAndControl(state *roundState, at int) error {
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

func buildDecisionView(state *roundState, side string) (decisionView, error) {
	if state == nil || !validSide(side) {
		return decisionView{}, newError("INVALID_DECISION_VIEW", "decision view requires a valid side")
	}
	view := decisionView{Side: side, Timeline: state.Timeline, RoundDeadline: state.RoundDeadline, BombDeadline: state.BombDeadline, DecisionCount: state.DecisionCount, RotationCount: state.RotationCount[side]}
	playerIDs := make([]string, 0, len(state.Players))
	for playerID := range state.Players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)
	for _, playerID := range playerIDs {
		player := state.Players[playerID]
		if player.Side == side {
			view.OwnPlayers = append(view.OwnPlayers, decisionPlayerView{
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
		if !ok || known.ExpiresAt <= state.Timeline || known.Status == controlUnknown {
			continue
		}
		view.KnownControls = append(view.KnownControls, decisionControlView{NodeID: nodeID, Status: known.Status, UpdatedAt: known.UpdatedAt, ExpiresAt: known.ExpiresAt})
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
	if side == SideT || state.Bomb.Status == bombPlanted || state.Bomb.Status == bombDefusing || state.Bomb.Status == bombDefused || state.Bomb.Status == bombExploded {
		view.BombStatus, view.BombNodeID, view.BombSite = state.Bomb.Status, projectedNodeID(state.Bomb.Location), state.Bomb.PlantedSite
	} else if view.BombIntel != nil {
		view.BombStatus, view.BombNodeID = state.Bomb.Status, view.BombIntel.NodeID
	}
	return view, nil
}

func intelScoreModifier(record intelRecord, maxMagnitude float64) float64 {
	return maxMagnitude * float64(clampInt(record.Confidence, 0, 100)) / 100
}

func canTriggerDeterministicIntelAction(record intelRecord, constants CombatConstants) bool {
	if record.Type == string(intelEmptySite) {
		return false
	}
	return record.Confidence >= constants.Int("SoundIntelMinConfidence", 30)
}

func rebuildIntelIndexes(teamIntel *teamIntel, timeline int) {
	teamIntel.KnownEnemies = map[string]intelRecord{}
	teamIntel.KnownControl = map[string]intelRecord{}
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
		if record.Type == string(intelControl) || record.Type == string(intelEmptySite) {
			current, ok := teamIntel.KnownControl[record.NodeID]
			if !ok || betterIntel(record, current) {
				teamIntel.KnownControl[record.NodeID] = cloneIntelRecord(record)
			}
		}
		if record.Type == string(intelSound) {
			teamIntel.SoundCues = append(teamIntel.SoundCues, cloneIntelRecord(record))
		}
		if record.Type == string(intelBomb) && (teamIntel.BombIntel == nil || betterIntel(record, *teamIntel.BombIntel)) {
			copy := cloneIntelRecord(record)
			teamIntel.BombIntel = &copy
		}
	}
}

func betterIntel(candidate, current intelRecord) bool {
	if candidate.LastSeenAt != current.LastSeenAt {
		return candidate.LastSeenAt > current.LastSeenAt
	}
	if candidate.Confidence != current.Confidence {
		return candidate.Confidence > current.Confidence
	}
	return candidate.ID < current.ID
}

func sortIntelRecords(records []intelRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ID < records[j].ID })
}

func cloneIntelRecord(record intelRecord) intelRecord {
	record.ObservedBy = append([]string(nil), record.ObservedBy...)
	return record
}

func clonePlayerLocation(location playerLocation) playerLocation {
	location.Edge = cloneOnEdge(location.Edge)
	return location
}

func validObservers(state *roundState, side string, observers []string) []string {
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

func validateEnemyTarget(state *roundState, side, targetID string) error {
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

func intelSource(observation intelObservation) string {
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
