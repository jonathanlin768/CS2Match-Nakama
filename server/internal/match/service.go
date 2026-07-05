package match

import (
	"context"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

// Service 是 match 子系统的业务入口。
type Service struct {
	engine *matchengine.Service
	logger runtime.Logger
}

// NewService 创建 match 业务服务。
func NewService(engine *matchengine.Service, logger runtime.Logger) *Service {
	return &Service{engine: engine, logger: logger}
}

// DebugSimuMatch 执行一回合调试对战并返回战报。
func (s *Service) DebugSimuMatch(ctx context.Context, userID string, req DebugSimuMatchRequest) (*DebugSimuMatchResponse, error) {
	if req.MapID == "" {
		req.MapID = "de_dust2"
	}
	if !matchengine.IsSupportedMap(req.MapID) {
		return nil, &MatchError{Code: "INVALID_MAP", Message: fmt.Sprintf("unsupported map: %s", req.MapID)}
	}

	teamT, err := s.buildTeam(SideT(), "测试T队", []string{
		"player_niko",
		"player_s1mple",
		"player_zywoo",
		"player_device",
		"player_ropz",
	})
	if err != nil {
		return nil, err
	}

	teamCT, err := s.buildTeam(SideCT(), "测试CT队", []string{
		"player_rain",
		"player_apex",
		"player_xyp9x",
		"player_karrigan",
		"player_broky",
	})
	if err != nil {
		return nil, err
	}

	input := &matchengine.MatchInput{
		MatchID: fmt.Sprintf("debug_%d", time.Now().UnixMilli()),
		TeamA:   teamT,
		TeamB:   teamCT,
		MapID:   req.MapID,
		Seed:    time.Now().UnixNano(),
	}

	result, err := s.engine.Simulate(ctx, input)
	if err != nil {
		return nil, &MatchError{Code: "SIMULATION_ERROR", Message: err.Error()}
	}

	return &DebugSimuMatchResponse{
		MatchInfo:  result.MatchInfo,
		Rounds:     result.Rounds,
		FinalStats: result.FinalStats,
		Winner:     result.Winner,
	}, nil
}

func (s *Service) buildTeam(side string, name string, playerIDs []string) (*matchengine.Team, error) {
	team := &matchengine.Team{
		ID:      side,
		Name:    name,
		Side:    side,
		Players: make([]*matchengine.Combatant, 0, len(playerIDs)),
	}
	for _, id := range playerIDs {
		p := cfg.GetPlayer(id)
		if p == nil {
			return nil, &MatchError{Code: "INVALID_LINEUP", Message: fmt.Sprintf("player not found: %s", id)}
		}
		team.Players = append(team.Players, &matchengine.Combatant{
			PlayerID:  p.Id,
			Name:      p.Name,
			Side:      side,
			Entry:     p.Entry,
			Aim:       p.Aim,
			Trade:     p.Trade,
			Clutch:    p.Clutch,
			Firepower: p.Firepower,
			Gamesense: p.Gamesense,
			Alive:     true,
		})
	}
	return team, nil
}

// SideT 返回进攻方阵营字符串。
func SideT() string { return matchengine.SideT }

// SideCT 返回防守方阵营字符串。
func SideCT() string { return matchengine.SideCT }

// Error implements the error interface for MatchError.
func (e *MatchError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
