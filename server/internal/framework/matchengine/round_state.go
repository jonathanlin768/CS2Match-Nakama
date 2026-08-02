package matchengine

import (
	"fmt"
	"sort"
)

type RoundPhase string
type CombatPosture string
type RecoveryStatus string
type ControlStatus string
type BombRuntimeStatus string

const (
	PhaseOpeningDeploy RoundPhase = "OpeningDeploy"
	PhaseAdvance       RoundPhase = "Advance"
	PhaseClash         RoundPhase = "Clash"
	PhaseRotate        RoundPhase = "RotateDecision"
	PhaseSiteContest   RoundPhase = "SiteContest"
	PhasePlanting      RoundPhase = "Planting"
	PhasePostPlant     RoundPhase = "PostPlant"
	PhaseRoundEnd      RoundPhase = "RoundEnd"

	PostureDefault  CombatPosture = "Default"
	PostureHolding  CombatPosture = "Holding"
	PostureMoving   CombatPosture = "Moving"
	PostureEngaged  CombatPosture = "Engaged"
	PostureRetaking CombatPosture = "Retaking"

	RecoveryNotAttempted RecoveryStatus = "NotAttempted"
	RecoveryRunning      RecoveryStatus = "Running"
	RecoveryFailed       RecoveryStatus = "Failed"
	RecoverySucceeded    RecoveryStatus = "Succeeded"

	ControlUnknown   ControlStatus = "Unknown"
	ControlT         ControlStatus = "T"
	ControlCT        ControlStatus = "CT"
	ControlContested ControlStatus = "Contested"

	BombCarried  BombRuntimeStatus = "Carried"
	BombDropped  BombRuntimeStatus = "Dropped"
	BombPlanting BombRuntimeStatus = "Planting"
	BombPlanted  BombRuntimeStatus = "Planted"
	BombDefusing BombRuntimeStatus = "Defusing"
	BombDefused  BombRuntimeStatus = "Defused"
	BombExploded BombRuntimeStatus = "Exploded"
)

type RoleAssignment struct {
	PlayerID string
	Role     string
}

type RoundPlan struct {
	TStrategyTemplateID string
	CTSetupTemplateID   string
	RoleAssignments     []RoleAssignment
	OpeningRoutes       map[string]string
	BombCarrierID       string
}

type NoProgressRecoveryState struct {
	CycleID          string
	Status           RecoveryStatus
	RecoveryActionID string
	StartedAt        int
	CompletedAt      int
	ResultCode       string
}

type RoundPlayerState struct {
	Profile PlayerProfile
	TeamID  string
	Side    string
	Weapon  WeaponLoadout

	Alive      bool
	HP         int
	Stamina    int
	Focus      int
	Suppressed bool
	Momentum   int

	Location     PlayerLocation
	Posture      CombatPosture
	Intent       Intent
	Action       PlayerActionState
	EngagementID string

	HasBomb bool
	Kills   int
	Deaths  int
	Damage  int
}

type KnownControlState struct {
	Status     ControlStatus
	UpdatedAt  int
	ObservedBy []string
	ExpiresAt  int
}

type NodeRuntimeState struct {
	Node          MapNode
	ActualControl ControlStatus
	KnownControl  map[string]KnownControlState
	UpdatedAt     int
}

func (n *NodeRuntimeState) ResolveContest(result ControlStatus, at, ttl int, observedBy map[string][]string) error {
	if result == ControlContested || !oneOf(string(result), string(ControlUnknown), string(ControlT), string(ControlCT)) {
		return newError("SIMULATION_INVARIANT_ERROR", "node %s contest did not resolve to a legal control state", n.Node.ID)
	}
	n.ActualControl = result
	n.UpdatedAt = at
	for _, side := range []string{SideT, SideCT} {
		observers := append([]string(nil), observedBy[side]...)
		if len(observers) == 0 {
			continue
		}
		sort.Strings(observers)
		n.KnownControl[side] = KnownControlState{Status: result, UpdatedAt: at, ObservedBy: observers, ExpiresAt: at + ttl}
	}
	return nil
}

func (n *NodeRuntimeState) DecayKnownControl(side string, timeline int) {
	known, ok := n.KnownControl[side]
	if !ok || known.ExpiresAt == 0 || timeline < known.ExpiresAt {
		return
	}
	known.Status = ControlUnknown
	known.ObservedBy = nil
	known.UpdatedAt = timeline
	n.KnownControl[side] = known
}

type IntelRecord struct {
	ID             string
	Type           string
	TargetID       string
	NodeID         string
	AreaID         string
	Source         string
	SourceActionID string
	SourceEventID  string
	Confidence     int
	ObservedBy     []string
	LastSeenAt     int
	ExpiresAt      int
}

type TeamIntel struct {
	Records      []IntelRecord
	KnownEnemies map[string]IntelRecord
	KnownControl map[string]IntelRecord
	SoundCues    []IntelRecord
	BombIntel    *IntelRecord
}

type EncounterState struct {
	ID             string
	SourceActionID string
	ScenarioID     string
	ActorIDs       []string
	NodeID         string
	StartedAt      int
	EndsAt         int
	PulseTimes     []int
	PulsesResolved int
	MaxPulses      int
	InitiativeSide string
	Status         string
	Reasons        []ReasonRecord
}

type TeamUtilityState struct {
	Budget  int
	Spent   int
	Windows []UtilityWindow
}

type BombState struct {
	Status         BombRuntimeStatus
	CarrierID      string
	Location       PlayerLocation
	DroppedAt      int
	PlantedSite    string
	PlantedAt      int
	ExplodeAt      int
	PlantActionID  string
	PlantActorID   string
	PlantStartAt   int
	PlantFinishAt  int
	DefuseActionID string
	DefuseActorID  string
	DefuseStartAt  int
	DefuseFinishAt int
}

func (b *BombState) Drop(location PlayerLocation, at int) error {
	if !location.Valid() || (b.Status != BombCarried && b.Status != BombPlanting) {
		return newError("SIMULATION_INVARIANT_ERROR", "bomb cannot drop from status %s", b.Status)
	}
	b.Status = BombDropped
	b.CarrierID = ""
	b.Location = location
	b.DroppedAt = at
	b.clearActions()
	return nil
}

func (b *BombState) StartPlant(actorID, actionID, site string, startAt, finishAt int) error {
	if b.Status != BombCarried || b.CarrierID != actorID || actionID == "" || finishAt <= startAt {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid plant start")
	}
	b.Status = BombPlanting
	b.PlantActorID, b.PlantActionID = actorID, actionID
	b.PlantedSite, b.PlantStartAt, b.PlantFinishAt = site, startAt, finishAt
	return nil
}

func (b *BombState) CompletePlant(location PlayerLocation, at, explodeAt int) error {
	if b.Status != BombPlanting || !location.Valid() || at != b.PlantFinishAt || explodeAt <= at {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid plant completion")
	}
	b.Status = BombPlanted
	b.CarrierID = ""
	b.Location = location
	b.PlantedAt, b.ExplodeAt = at, explodeAt
	return nil
}

func (b *BombState) InterruptPlant(actorID, actionID string) bool {
	if b.Status != BombPlanting || b.PlantActorID != actorID || b.PlantActionID != actionID {
		return false
	}
	b.Status = BombCarried
	b.PlantActorID, b.PlantActionID = "", ""
	b.PlantStartAt, b.PlantFinishAt = 0, 0
	b.PlantedSite = ""
	return true
}

func (b *BombState) StartDefuse(actorID, actionID string, startAt, finishAt int) error {
	if b.Status != BombPlanted || actorID == "" || actionID == "" || finishAt <= startAt || finishAt > b.ExplodeAt {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid defuse start")
	}
	b.Status = BombDefusing
	b.DefuseActorID, b.DefuseActionID = actorID, actionID
	b.DefuseStartAt, b.DefuseFinishAt = startAt, finishAt
	return nil
}

func (b *BombState) CompleteDefuse(at int) error {
	if b.Status != BombDefusing || at != b.DefuseFinishAt || at > b.ExplodeAt {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid defuse completion")
	}
	b.Status = BombDefused
	return nil
}

func (b *BombState) InterruptDefuse(actorID, actionID string) bool {
	if b.Status != BombDefusing || b.DefuseActorID != actorID || b.DefuseActionID != actionID {
		return false
	}
	b.Status = BombPlanted
	b.DefuseActorID, b.DefuseActionID = "", ""
	b.DefuseStartAt, b.DefuseFinishAt = 0, 0
	return true
}

func (b *BombState) Pickup(actorID string, location PlayerLocation, at int) error {
	if b.Status != BombDropped || actorID == "" || !location.Valid() || !sameLocation(b.Location, location) {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid bomb pickup")
	}
	b.Status = BombCarried
	b.CarrierID = actorID
	b.Location = location
	b.DroppedAt = 0
	_ = at
	return nil
}

func (b *BombState) Explode(at int) error {
	if (b.Status != BombPlanted && b.Status != BombDefusing) || at != b.ExplodeAt {
		return newError("SIMULATION_INVARIANT_ERROR", "invalid bomb explosion")
	}
	b.Status = BombExploded
	return nil
}

func (b *BombState) clearActions() {
	b.PlantActionID, b.PlantActorID = "", ""
	b.DefuseActionID, b.DefuseActorID = "", ""
}

type RoundState struct {
	RoundNumber int
	Seed        int64
	Phase       RoundPhase
	Timeline    int

	RoundDeadline int
	BombDeadline  int
	MapID         string
	TeamTID       string
	TeamCTID      string
	TeamAID       string
	TeamBID       string
	ScoreByTeam   map[string]int
	Plan          RoundPlan

	Players            map[string]*RoundPlayerState
	Bomb               BombState
	Nodes              map[string]*NodeRuntimeState
	Intel              map[string]*TeamIntel
	ActiveEngagements  map[string]*EncounterState
	Scheduler          *ActionScheduler
	MomentumT          int
	MomentumCT         int
	Utility            map[string]*TeamUtilityState
	DecisionCount      int
	RotationCount      map[string]int
	TransitionCount    int
	NoOpCount          int
	NoProgressEligible bool
	RecoveryOrdinal    int
	RecoveryAttempt    NoProgressRecoveryState
	Events             []*GameEvent
	Terminal           *RoundTerminal
	PhaseHistory       []RoundPhase
	constants          CombatConstants
	mapEdges           map[string]MapEdge
	routes             map[string]Route
	routeTemplates     map[string]RouteTemplate
	scenarios          map[string]Scenario
	mapTags            map[string]MapTag
	encounterModifiers map[string]EncounterModifier
	visibility         map[string]Visibility
	weaponSpecs        map[string]WeaponSpec
}

func NewRoundState(input *RoundInput, plan RoundPlan) (*RoundState, error) {
	if input == nil || input.MapConfig == nil {
		return nil, newError("INVALID_ROUND_INPUT", "round input/map config is nil")
	}
	constants := input.MapConfig.CombatConstants
	state := &RoundState{
		RoundNumber:        input.RoundNumber,
		Seed:               input.Seed,
		Phase:              PhaseOpeningDeploy,
		RoundDeadline:      constants.Int("RoundTimeLimit", 0),
		MapID:              input.MapID,
		TeamTID:            input.TeamT.TeamID,
		TeamCTID:           input.TeamCT.TeamID,
		TeamAID:            input.TeamAID,
		TeamBID:            input.TeamBID,
		ScoreByTeam:        copyIntMap(input.ScoreByTeam),
		Plan:               plan,
		Players:            map[string]*RoundPlayerState{},
		Nodes:              map[string]*NodeRuntimeState{},
		Intel:              map[string]*TeamIntel{},
		ActiveEngagements:  map[string]*EncounterState{},
		Scheduler:          NewActionScheduler(constants),
		Utility:            map[string]*TeamUtilityState{},
		RotationCount:      map[string]int{SideT: 0, SideCT: 0},
		RecoveryAttempt:    NoProgressRecoveryState{Status: RecoveryNotAttempted},
		PhaseHistory:       []RoundPhase{PhaseOpeningDeploy},
		constants:          constants,
		mapEdges:           input.MapConfig.Edges,
		routes:             input.MapConfig.Routes,
		routeTemplates:     input.MapConfig.RouteTemplates,
		scenarios:          input.MapConfig.Scenarios,
		mapTags:            input.MapConfig.MapTags,
		encounterModifiers: input.MapConfig.EncounterModifiers,
		visibility:         input.MapConfig.Visibility,
		weaponSpecs:        input.WeaponSpecs,
	}
	if state.TeamAID == "" {
		state.TeamAID = input.TeamT.TeamID
	}
	if state.TeamBID == "" {
		state.TeamBID = input.TeamCT.TeamID
	}
	for _, side := range []string{SideT, SideCT} {
		state.Intel[side] = &TeamIntel{KnownEnemies: map[string]IntelRecord{}, KnownControl: map[string]IntelRecord{}}
		state.Utility[side] = &TeamUtilityState{Budget: constants.Int("UtilityBudget", 0)}
	}
	for id, node := range input.MapConfig.Nodes {
		actual := ControlUnknown
		if node.DefaultSide == SideT {
			actual = ControlT
		} else if node.DefaultSide == SideCT {
			actual = ControlCT
		}
		state.Nodes[id] = &NodeRuntimeState{Node: node, ActualControl: actual, KnownControl: map[string]KnownControlState{}}
	}
	if err := state.addTeamPlayers(input.TeamT, SideT, input.SideLoadouts[SideT], "T_SPAWN"); err != nil {
		return nil, err
	}
	if err := state.addTeamPlayers(input.TeamCT, SideCT, input.SideLoadouts[SideCT], "CT_SPAWN"); err != nil {
		return nil, err
	}
	carrier, ok := state.Players[plan.BombCarrierID]
	if !ok || carrier.Side != SideT || !carrier.Alive {
		return nil, newError("INVALID_OPENING_PLAN", "bomb carrier %s is not a live T player", plan.BombCarrierID)
	}
	carrier.HasBomb = true
	state.Bomb = BombState{Status: BombCarried, CarrierID: carrier.Profile.PlayerID, Location: carrier.Location}
	InitializeUtilityBudget(state, SideT, input.MapConfig.RouteTemplates[plan.TStrategyTemplateID])
	InitializeUtilityBudget(state, SideCT, input.MapConfig.RouteTemplates[plan.CTSetupTemplateID])
	if err := state.ClampAndValidate(); err != nil {
		return nil, err
	}
	return state, nil
}

func (s *RoundState) addTeamPlayers(team TeamInput, side string, loadout WeaponLoadout, spawnID string) error {
	if len(team.Players) != 5 {
		return newError("INVALID_LINEUP", "team %s must contain five players", team.TeamID)
	}
	if _, ok := s.Nodes[spawnID]; !ok {
		return newError("CONFIG_UNREACHABLE_NODE", "spawn %s is missing", spawnID)
	}
	for _, profile := range team.Players {
		if _, duplicate := s.Players[profile.PlayerID]; duplicate {
			return newError("INVALID_LINEUP", "duplicate player %s", profile.PlayerID)
		}
		s.Players[profile.PlayerID] = &RoundPlayerState{
			Profile: profile, TeamID: team.TeamID, Side: side, Weapon: loadout,
			Alive: true, HP: s.constants.Int("MaxHP", 0), Stamina: s.constants.Int("MaxStamina", 0), Focus: s.constants.Int("MaxFocus", 0),
			Location: PlayerLocation{NodeID: spawnID}, Posture: PostureDefault, Action: PlayerActionState{Status: ActionIdle},
		}
	}
	return nil
}

func (s *RoundState) ClampAndValidate() error {
	minHP, maxHP := s.constants.Int("MinHP", 0), s.constants.Int("MaxHP", 0)
	minStamina, maxStamina := s.constants.Int("MinStamina", 0), s.constants.Int("MaxStamina", 0)
	minFocus, maxFocus := s.constants.Int("MinFocus", 0), s.constants.Int("MaxFocus", 0)
	s.MomentumT = clampInt(s.MomentumT, -100, 100)
	s.MomentumCT = clampInt(s.MomentumCT, -100, 100)
	if s.Timeline < 0 {
		s.Timeline = 0
	}
	if s.Timeline > s.constants.Int("MaxRoundTimeline", 0) {
		return newError("TIMELINE_LIMIT_EXCEEDED", "MaxRoundTimeline exceeded")
	}
	carrierCount := 0
	for id, player := range s.Players {
		player.HP = clampInt(player.HP, minHP, maxHP)
		player.Stamina = clampInt(player.Stamina, minStamina, maxStamina)
		player.Focus = clampInt(player.Focus, minFocus, maxFocus)
		player.Momentum = clampInt(player.Momentum, -100, 100)
		if !player.Location.Valid() {
			return newError("SIMULATION_INVARIANT_ERROR", "player %s location is not Node XOR Edge", id)
		}
		if player.HP == 0 {
			player.Alive = false
		}
		if !player.Alive && player.HP != 0 {
			return newError("SIMULATION_INVARIANT_ERROR", "dead player %s has non-zero HP", id)
		}
		if !player.Alive && player.Action.CurrentActionID != "" {
			return newError("SIMULATION_INVARIANT_ERROR", "dead player %s has a running action", id)
		}
		if player.HasBomb {
			carrierCount++
			if s.Bomb.CarrierID != id || (s.Bomb.Status != BombCarried && s.Bomb.Status != BombPlanting) {
				return newError("SIMULATION_INVARIANT_ERROR", "player/bomb carrier state disagrees for %s", id)
			}
		}
	}
	if s.Bomb.Status == BombCarried || s.Bomb.Status == BombPlanting {
		if carrierCount != 1 {
			return newError("SIMULATION_INVARIANT_ERROR", "carried bomb must have exactly one player carrier")
		}
	} else if carrierCount != 0 || s.Bomb.CarrierID != "" {
		return newError("SIMULATION_INVARIANT_ERROR", "non-carried bomb still has a carrier")
	}
	if s.Phase == PhaseRoundEnd && s.Terminal == nil {
		return newError("SIMULATION_INVARIANT_ERROR", "RoundEnd phase has no terminal")
	}
	return nil
}

func clampProbability(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func (s *RoundState) String() string {
	return fmt.Sprintf("round=%d phase=%s t=%d players=%d", s.RoundNumber, s.Phase, s.Timeline, len(s.Players))
}
