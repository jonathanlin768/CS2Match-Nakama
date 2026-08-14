package match

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

type mapConfigState struct {
	config *matchengine.MapConfig
	err    *MatchError
}

const (
	defaultTeamAID          = "team_falcons"
	defaultTeamBID          = "team_vitality"
	defaultTeamAName        = "Team Falcons"
	defaultTeamBName        = "Team Vitality"
	battleReportDebugLogKey = "BattleReportDebugLog"
)

// Service 是 match 子系统的业务入口。
type Service struct {
	engine    *matchengine.Service
	logger    runtime.Logger
	mapConfig map[string]mapConfigState
}

// NewService 创建 match 业务服务，并缓存地图配置校验结果。
func NewService(engine *matchengine.Service, logger runtime.Logger) *Service {
	s := &Service{
		engine:    engine,
		logger:    logger,
		mapConfig: map[string]mapConfigState{},
	}
	s.cacheMapConfig(matchengine.DefaultMapID)
	return s
}

// DebugSimuMatch 执行完整调试对战并返回整场战报。
func (s *Service) DebugSimuMatch(ctx context.Context, userID string, req DebugSimuMatchRequest) (*DebugSimuMatchResponse, error) {
	if req.MapID == "" {
		req.MapID = matchengine.DefaultMapID
	}
	state, ok := s.mapConfig[req.MapID]
	if !ok {
		return nil, &MatchError{Code: "INVALID_MAP", Message: fmt.Sprintf("unsupported map: %s", req.MapID)}
	}
	if state.err != nil {
		return nil, state.err
	}

	teamAPlayerIDs, teamBPlayerIDs, err := defaultTeamPlayerIDs()
	if err != nil {
		return nil, err
	}
	teamA, err := s.buildNamedTeam("team_a", defaultTeamAName, teamAPlayerIDs)
	if err != nil {
		return nil, err
	}
	teamB, err := s.buildNamedTeam("team_b", defaultTeamBName, teamBPlayerIDs)
	if err != nil {
		return nil, err
	}

	seed := req.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	result, err := s.simulate(ctx, fmt.Sprintf("debug_%d", time.Now().UnixMilli()), req.MapID, seed, teamA, teamB)
	if err != nil {
		return nil, err
	}
	return &DebugSimuMatchResponse{
		MatchResult: result,
		DebugEnabled: combatConstBool(
			state.config.CombatConstants,
			battleReportDebugLogKey,
			false,
		),
	}, nil
}

// SimuMatch runs a server-authoritative tutorial or computer battle.
func (s *Service) SimuMatch(ctx context.Context, userID string, req SimuMatchRequest) (*SimuMatchResponse, error) {
	var mapID string
	var teamA, teamB matchengine.TeamInput
	var err error
	switch req.Mode {
	case "computer":
		mapID = matchengine.DefaultMapID
		teamAIDs, teamBIDs, lineupErr := defaultTeamPlayerIDs()
		if lineupErr != nil {
			return nil, lineupErr
		}
		teamA, err = s.buildConfigTeam(defaultTeamAID, teamAIDs)
		if err == nil {
			teamB, err = s.buildConfigTeam(defaultTeamBID, teamBIDs)
		}
	case "tutorial":
		tutorial := cfg.GetTutorialBattle(req.TutorialConfigID)
		if tutorial == nil || !tutorial.Enabled {
			return nil, &MatchError{Code: "INVALID_TUTORIAL_CONFIG", Message: "tutorial config is unavailable"}
		}
		if req.ConfigVersion != tutorial.Version {
			return nil, &MatchError{Code: "CONFIG_VERSION_MISMATCH", Message: "tutorial config has changed; reload and try again"}
		}
		if err = validateTutorialLineup(tutorial, req.PlayerIDs); err != nil {
			return nil, err
		}
		mapID = tutorial.MapId
		teamA, err = s.buildNamedTeam("tutorial_players", "你的临时阵容", req.PlayerIDs)
		if err == nil {
			teamB, err = s.buildConfigTeam(tutorial.OpponentTeamId, tutorial.OpponentPlayerIds)
		}
	default:
		return nil, &MatchError{Code: "INVALID_MODE", Message: "mode must be tutorial or computer"}
	}
	if err != nil {
		return nil, err
	}
	result, err := s.simulate(ctx, fmt.Sprintf("simu_%s_%d", userID, time.Now().UnixMilli()), mapID, time.Now().UnixNano(), teamA, teamB)
	if err != nil {
		return nil, err
	}
	return &SimuMatchResponse{
		MatchResult: result,
		DebugEnabled: combatConstBool(
			s.mapConfig[mapID].config.CombatConstants,
			battleReportDebugLogKey,
			false,
		),
	}, nil
}

func (s *Service) simulate(ctx context.Context, matchID, mapID string, seed int64, teamA, teamB matchengine.TeamInput) (*matchengine.MatchResult, error) {
	if err := validateMatchTeamIdentities(teamA, teamB); err != nil {
		return nil, err
	}
	state, ok := s.mapConfig[mapID]
	if !ok {
		s.cacheMapConfig(mapID)
		state, ok = s.mapConfig[mapID]
	}
	if !ok {
		return nil, &MatchError{Code: "INVALID_MAP", Message: fmt.Sprintf("unsupported map: %s", mapID)}
	}
	if state.err != nil {
		return nil, state.err
	}
	input := &matchengine.MatchInput{
		MatchID:    matchID,
		MapID:      mapID,
		MapName:    state.config.MapName,
		MapVersion: state.config.Version,
		Seed:       seed,
		RuleSet:    matchengine.DefaultMR12RuleSet(state.config.CombatConstants),
		TeamA:      teamA,
		TeamB:      teamB,
		InitialSideByTeam: map[string]string{
			teamA.TeamID: matchengine.SideT,
			teamB.TeamID: matchengine.SideCT,
		},
		MapConfig:    state.config,
		WeaponSpecs:  defaultWeaponSpecs(),
		SideLoadouts: defaultSideLoadouts(),
	}

	result, err := s.engine.Simulate(ctx, input)
	if err != nil {
		if me, ok := err.(*matchengine.EngineError); ok {
			return nil, &MatchError{Code: me.Code, Message: me.Message}
		}
		return nil, &MatchError{Code: "SIMULATION_ERROR", Message: err.Error()}
	}
	return result, nil
}

func validateMatchTeamIdentities(teamA, teamB matchengine.TeamInput) error {
	if strings.TrimSpace(teamA.TeamID) == "" || strings.TrimSpace(teamB.TeamID) == "" {
		return &MatchError{Code: "INVALID_TEAM", Message: "both team ids are required"}
	}
	if teamA.TeamID == teamB.TeamID {
		return &MatchError{Code: "INVALID_LINEUP", Message: "team ids must be unique within a match"}
	}

	seenPlayerIDs := make(map[string]struct{}, len(teamA.Players)+len(teamB.Players))
	for _, team := range []matchengine.TeamInput{teamA, teamB} {
		for _, player := range team.Players {
			expectedID, err := matchPlayerID(team.TeamID, player.ConfigPlayerID)
			if err != nil {
				return err
			}
			if player.PlayerID != expectedID {
				return &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("player instance id %q does not match team/config identity", player.PlayerID)}
			}
			if _, duplicate := seenPlayerIDs[player.PlayerID]; duplicate {
				return &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("duplicate player instance id: %s", player.PlayerID)}
			}
			seenPlayerIDs[player.PlayerID] = struct{}{}
		}
	}
	return nil
}

func combatConstBool(constants matchengine.CombatConstants, key string, fallback bool) bool {
	value, ok := constants.Values[key]
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value.Value)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

func (s *Service) cacheMapConfig(mapID string) {
	mapConfig, err := buildMapConfigFromTables(mapID)
	if err != nil {
		me := &MatchError{Code: "CONFIG_ERROR", Message: err.Error()}
		if typed, ok := err.(*MatchError); ok {
			me = typed
		}
		s.mapConfig[mapID] = mapConfigState{config: mapConfig, err: me}
		if s.logger != nil {
			s.logger.Warn("map config validation failed for %s: %s", mapID, me.Error())
		}
		return
	}
	s.mapConfig[mapID] = mapConfigState{config: mapConfig}
}

func (s *Service) buildConfigTeam(teamID string, playerIDs []string) (matchengine.TeamInput, error) {
	teamConfig := cfg.GetTeam(teamID)
	if teamConfig == nil {
		return matchengine.TeamInput{}, &MatchError{Code: "INVALID_TEAM", Message: fmt.Sprintf("team not found: %s", teamID)}
	}
	return s.buildTeam(teamID, teamConfig.Name, playerIDs)
}

func (s *Service) buildNamedTeam(teamID string, name string, playerIDs []string) (matchengine.TeamInput, error) {
	return s.buildTeam(teamID, name, playerIDs)
}

func (s *Service) buildTeam(teamID string, name string, playerIDs []string) (matchengine.TeamInput, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return matchengine.TeamInput{}, &MatchError{Code: "INVALID_TEAM", Message: "team id is required"}
	}
	if err := validateUniqueConfigPlayerIDs(teamID, playerIDs); err != nil {
		return matchengine.TeamInput{}, err
	}
	team := matchengine.TeamInput{
		TeamID:  teamID,
		Name:    name,
		Players: make([]matchengine.PlayerProfile, 0, len(playerIDs)),
	}
	for _, id := range playerIDs {
		p := cfg.GetPlayer(id)
		if p == nil {
			return matchengine.TeamInput{}, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("player not found: %s", id)}
		}
		profile, profileErr := playerFromConfig(teamID, p)
		if profileErr != nil {
			return matchengine.TeamInput{}, profileErr
		}
		team.Players = append(team.Players, profile)
	}
	return team, nil
}

func validateUniqueConfigPlayerIDs(teamID string, playerIDs []string) error {
	seen := make(map[string]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if _, duplicate := seen[playerID]; duplicate {
			return &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("duplicate player in team %s: %s", teamID, playerID)}
		}
		seen[playerID] = struct{}{}
	}
	return nil
}

func validateTutorialLineup(tutorial *cfg.TutorialBattle, playerIDs []string) error {
	if len(playerIDs) != int(tutorial.RosterSize) {
		return &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("select exactly %d players", tutorial.RosterSize)}
	}
	prices := make(map[string]int)
	for price, ids := range map[int][]string{5: tutorial.Tier5PlayerIds, 4: tutorial.Tier4PlayerIds, 3: tutorial.Tier3PlayerIds, 2: tutorial.Tier2PlayerIds, 1: tutorial.Tier1PlayerIds} {
		for _, id := range ids {
			prices[id] = price
		}
	}
	seen, total := make(map[string]struct{}, len(playerIDs)), 0
	for _, id := range playerIDs {
		if _, duplicate := seen[id]; duplicate {
			return &MatchError{Code: "INVALID_LINEUP", Message: "player ids must be unique"}
		}
		price, ok := prices[id]
		if !ok || cfg.GetPlayer(id) == nil {
			return &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("player is not in tutorial pool: %s", id)}
		}
		seen[id], total = struct{}{}, total+price
	}
	if total > int(tutorial.Budget) {
		return &MatchError{Code: "BUDGET_EXCEEDED", Message: fmt.Sprintf("lineup costs %d, budget is %d", total, tutorial.Budget)}
	}
	return nil
}

func matchPlayerID(teamID string, configPlayerID string) (string, error) {
	teamID = strings.TrimSpace(teamID)
	configPlayerID = strings.TrimSpace(configPlayerID)
	if teamID == "" || configPlayerID == "" {
		return "", &MatchError{Code: "INVALID_LINEUP", Message: "team id and config player id are required"}
	}
	return url.PathEscape(teamID) + "/" + url.PathEscape(configPlayerID), nil
}

func playerFromConfig(teamID string, p *cfg.Player) (matchengine.PlayerProfile, error) {
	if p == nil {
		return matchengine.PlayerProfile{}, &MatchError{Code: "INVALID_LINEUP", Message: "player config is required"}
	}
	playerID, err := matchPlayerID(teamID, p.Id)
	if err != nil {
		return matchengine.PlayerProfile{}, err
	}
	crop := &matchengine.ImageCrop{
		X: float64(p.AvatarCropX), Y: float64(p.AvatarCropY),
		Width: float64(p.AvatarCropWidth), Height: float64(p.AvatarCropHeight),
	}
	if p.CardImage == "" || !crop.Valid() {
		crop = nil
	}
	return matchengine.PlayerProfile{
		PlayerID:       playerID,
		ConfigPlayerID: p.Id,
		DisplayName:    p.Name,
		Portrait:       p.Portrait,
		CardImage:      p.CardImage,
		AvatarCrop:     crop,
		RoleTags:       append([]string(nil), p.Positions...),
		Attributes: matchengine.PlayerAttributes{
			Entry:       int(p.Entry),
			Aim:         int(p.Aim),
			Trade:       int(p.Trade),
			Clutch:      int(p.Clutch),
			Firepower:   int(p.Firepower),
			Gamesense:   int(p.Gamesense),
			Reaction:    int(p.Reaction),
			Positioning: int(p.Positioning),
			Awareness:   int(p.Awareness),
			Teamplay:    int(p.Teamplay),
			Utility:     int(p.Utility),
			Composure:   int(p.Composure),
			Mobility:    int(p.Mobility),
			Endurance:   int(p.Endurance),
			Discipline:  int(p.Discipline),
		},
	}, nil
}

func defaultTeamPlayerIDs() ([]string, []string, error) {
	if cfg.Global == nil || cfg.Global.TbPlayer == nil {
		return nil, nil, &MatchError{Code: "CONFIG_NOT_INITIALIZED", Message: "TbPlayer is not initialized"}
	}
	players := cfg.Global.TbPlayer.GetDataList()
	teamAPlayerIDs := make([]string, 0, 5)
	teamBPlayerIDs := make([]string, 0, 5)
	for index, player := range players {
		if player == nil {
			return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("TbPlayer row %d is empty", index+1)}
		}

		var target *[]string
		switch {
		case strings.EqualFold(strings.TrimSpace(player.TeamId), defaultTeamAID):
			target = &teamAPlayerIDs
		case strings.EqualFold(strings.TrimSpace(player.TeamId), defaultTeamBID):
			target = &teamBPlayerIDs
		default:
			continue
		}

		playerID := strings.TrimSpace(player.Id)
		if playerID == "" {
			return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("TbPlayer row %d has no player id", index+1)}
		}
		*target = append(*target, playerID)
	}
	if err := validateUniqueConfigPlayerIDs(defaultTeamAID, teamAPlayerIDs); err != nil {
		return nil, nil, err
	}
	if err := validateUniqueConfigPlayerIDs(defaultTeamBID, teamBPlayerIDs); err != nil {
		return nil, nil, err
	}

	if len(teamAPlayerIDs) != 5 {
		return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("team %s requires exactly 5 players, got %d", defaultTeamAName, len(teamAPlayerIDs))}
	}
	if len(teamBPlayerIDs) != 5 {
		return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("team %s requires exactly 5 players, got %d", defaultTeamBName, len(teamBPlayerIDs))}
	}

	return teamAPlayerIDs, teamBPlayerIDs, nil
}

func defaultWeaponSpecs() map[string]matchengine.WeaponSpec {
	return map[string]matchengine.WeaponSpec{
		matchengine.WeaponAK47: {
			ID:               matchengine.WeaponAK47,
			DisplayName:      "AK-47",
			Damage:           36,
			RoundsPerMinute:  600,
			MagazineSize:     30,
			ArmorPenetration: 0.775,
			RangeModifier:    0.98,
		},
		matchengine.WeaponM4A1S: {
			ID:               matchengine.WeaponM4A1S,
			DisplayName:      "M4A1-S",
			Damage:           38,
			RoundsPerMinute:  600,
			MagazineSize:     20,
			ArmorPenetration: 0.70,
			RangeModifier:    0.99,
		},
	}
}

func defaultSideLoadouts() map[string]matchengine.WeaponLoadout {
	return map[string]matchengine.WeaponLoadout{
		matchengine.SideT: {
			Primary:  matchengine.WeaponAK47,
			Armor:    true,
			Helmet:   true,
			Grenades: []string{"flash", "smoke"},
		},
		matchengine.SideCT: {
			Primary:  matchengine.WeaponM4A1S,
			Armor:    true,
			Helmet:   true,
			HasKit:   true,
			Grenades: []string{"flash", "smoke"},
		},
	}
}

func SideT() string { return matchengine.SideT }

func SideCT() string { return matchengine.SideCT }

func (e *MatchError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
