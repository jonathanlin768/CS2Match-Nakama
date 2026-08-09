package match

import (
	"context"
	"fmt"
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
	defaultTeamAName        = "Falcons"
	defaultTeamBName        = "Vitality"
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
	teamA, err := s.buildTeam("team_a", defaultTeamAName, teamAPlayerIDs)
	if err != nil {
		return nil, err
	}
	teamB, err := s.buildTeam("team_b", defaultTeamBName, teamBPlayerIDs)
	if err != nil {
		return nil, err
	}

	seed := req.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	input := &matchengine.MatchInput{
		MatchID:    fmt.Sprintf("debug_%d", time.Now().UnixMilli()),
		MapID:      req.MapID,
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
	return &DebugSimuMatchResponse{
		MatchResult: result,
		DebugEnabled: combatConstBool(
			state.config.CombatConstants,
			battleReportDebugLogKey,
			false,
		),
	}, nil
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

func (s *Service) buildTeam(teamID string, name string, playerIDs []string) (matchengine.TeamInput, error) {
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
		team.Players = append(team.Players, playerFromConfig(p))
	}
	return team, nil
}

func playerFromConfig(p *cfg.Player) matchengine.PlayerProfile {
	return matchengine.PlayerProfile{
		PlayerID:    p.Id,
		DisplayName: p.Name,
		Portrait:    p.Portrait,
		RoleTags:    append([]string(nil), p.Positions...),
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
	}
}

func defaultTeamPlayerIDs() ([]string, []string, error) {
	if cfg.Global == nil || cfg.Global.TbPlayer == nil {
		return nil, nil, &MatchError{Code: "CONFIG_NOT_INITIALIZED", Message: "TbPlayer is not initialized"}
	}
	players := cfg.Global.TbPlayer.GetDataList()
	teamAPlayerIDs := make([]string, 0, 5)
	teamBPlayerIDs := make([]string, 0, 5)
	seenPlayerIDs := make(map[string]struct{}, 10)

	for index, player := range players {
		if player == nil {
			return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("TbPlayer row %d is empty", index+1)}
		}

		var target *[]string
		switch {
		case strings.EqualFold(strings.TrimSpace(player.Team), defaultTeamAName):
			target = &teamAPlayerIDs
		case strings.EqualFold(strings.TrimSpace(player.Team), defaultTeamBName):
			target = &teamBPlayerIDs
		default:
			continue
		}

		playerID := strings.TrimSpace(player.Id)
		if playerID == "" {
			return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("TbPlayer row %d has no player id", index+1)}
		}
		if _, exists := seenPlayerIDs[playerID]; exists {
			return nil, nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("duplicate default player id: %s", playerID)}
		}
		seenPlayerIDs[playerID] = struct{}{}
		*target = append(*target, playerID)
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
