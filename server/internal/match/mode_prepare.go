package match

import (
	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

// preparedMatch contains mode-specific inputs for the shared simulation flow.
type preparedMatch struct {
	MapID string
	TeamA matchengine.TeamInput
	TeamB matchengine.TeamInput
}

func (s *Service) prepareMatch(req SimuMatchRequest) (*preparedMatch, error) {
	switch req.Mode {
	case MatchModeComputer:
		return s.prepareComputerMatch()
	case MatchModeTutorial:
		return s.prepareTutorialMatch(req)
	default:
		return nil, &MatchError{Code: "INVALID_MODE", Message: "mode must be tutorial or computer"}
	}
}

func (s *Service) prepareComputerMatch() (*preparedMatch, error) {
	teamAIDs, teamBIDs, err := defaultTeamPlayerIDs()
	if err != nil {
		return nil, err
	}
	teamA, err := s.buildConfigTeam(defaultTeamAID, teamAIDs)
	if err != nil {
		return nil, err
	}
	teamB, err := s.buildConfigTeam(defaultTeamBID, teamBIDs)
	if err != nil {
		return nil, err
	}
	return &preparedMatch{MapID: matchengine.DefaultMapID, TeamA: teamA, TeamB: teamB}, nil
}

func (s *Service) prepareTutorialMatch(req SimuMatchRequest) (*preparedMatch, error) {
	tutorial := cfg.GetTutorialBattle(req.TutorialConfigID)
	if tutorial == nil || !tutorial.Enabled {
		return nil, &MatchError{Code: "INVALID_TUTORIAL_CONFIG", Message: "tutorial config is unavailable"}
	}
	if req.ConfigVersion != tutorial.Version {
		return nil, &MatchError{Code: "CONFIG_VERSION_MISMATCH", Message: "tutorial config has changed; reload and try again"}
	}
	if err := validateTutorialLineup(tutorial, req.PlayerIDs); err != nil {
		return nil, err
	}
	teamA, err := s.buildNamedTeam("tutorial_players", "你的临时阵容", req.PlayerIDs)
	if err != nil {
		return nil, err
	}
	teamB, err := s.buildConfigTeam(tutorial.OpponentTeamId, tutorial.OpponentPlayerIds)
	if err != nil {
		return nil, err
	}
	return &preparedMatch{MapID: tutorial.MapId, TeamA: teamA, TeamB: teamB}, nil
}
