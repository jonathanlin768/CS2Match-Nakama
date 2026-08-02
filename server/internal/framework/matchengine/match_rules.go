package matchengine

// MatchScoreState owns team-identity scores and the current side assignment.
// Round simulation supplies only a natural winning team; this type deliberately
// knows nothing about combat events or how a round winner was produced.
type MatchScoreState struct {
	TeamAID     string
	TeamBID     string
	ScoreByTeam map[string]int
	SideByTeam  map[string]string
}

func NewMatchScoreState(teamAID, teamBID string, initialSideByTeam map[string]string) (*MatchScoreState, error) {
	if teamAID == "" || teamBID == "" || teamAID == teamBID {
		return nil, newError("INVALID_MATCH_INPUT", "two distinct team ids are required")
	}
	sideA := initialSideByTeam[teamAID]
	sideB := initialSideByTeam[teamBID]
	if !validSide(sideA) || !validSide(sideB) || sideA == sideB {
		return nil, newError("INVALID_MATCH_INPUT", "initial sides must assign one T and one CT")
	}
	return &MatchScoreState{
		TeamAID: teamAID,
		TeamBID: teamBID,
		ScoreByTeam: map[string]int{
			teamAID: 0,
			teamBID: 0,
		},
		SideByTeam: map[string]string{
			teamAID: sideA,
			teamBID: sideB,
		},
	}, nil
}

func (s *MatchScoreState) ApplyRoundWinner(teamID string) error {
	if teamID != s.TeamAID && teamID != s.TeamBID {
		return newError("SIMULATION_INVARIANT_ERROR", "round winner %q is not a match team", teamID)
	}
	s.ScoreByTeam[teamID]++
	return nil
}

func (s *MatchScoreState) Score(teamID string) int {
	return s.ScoreByTeam[teamID]
}

func (s *MatchScoreState) TeamForSide(side string) string {
	for _, teamID := range []string{s.TeamAID, s.TeamBID} {
		if s.SideByTeam[teamID] == side {
			return teamID
		}
	}
	return ""
}

func (s *MatchScoreState) SwitchSides() {
	for _, teamID := range []string{s.TeamAID, s.TeamBID} {
		if s.SideByTeam[teamID] == SideT {
			s.SideByTeam[teamID] = SideCT
		} else {
			s.SideByTeam[teamID] = SideT
		}
	}
}

func (s *MatchScoreState) RegulationComplete(rule RuleSet, roundsPlayed int) bool {
	return s.Score(s.TeamAID) >= rule.RegulationWinRounds ||
		s.Score(s.TeamBID) >= rule.RegulationWinRounds ||
		roundsPlayed >= rule.RegulationMaxRounds
}

func (s *MatchScoreState) ShouldEnterOvertime(rule RuleSet, roundsPlayed int) bool {
	return rule.OvertimeEnabled &&
		roundsPlayed >= rule.RegulationMaxRounds &&
		s.Score(s.TeamAID) == s.Score(s.TeamBID)
}

// OvertimeDecided intentionally refuses to end a match inside an MR3 block.
func (s *MatchScoreState) OvertimeDecided(rule RuleSet, roundsInBlock int) bool {
	return roundsInBlock >= rule.OvertimeBlockRounds && s.Score(s.TeamAID) != s.Score(s.TeamBID)
}
