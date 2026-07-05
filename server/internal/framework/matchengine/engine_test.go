package matchengine

import (
	"context"
	"testing"
)

func makeTestTeam(side string, ids []string) *Team {
	players := make([]*Combatant, 0, len(ids))
	for i, id := range ids {
		// 构造不同属性，id 越小属性越高
		power := int32(90 - i*10)
		players = append(players, &Combatant{
			PlayerID:  id,
			Name:      id,
			Side:      side,
			Entry:     power,
			Aim:       power,
			Firepower: power,
			Alive:     true,
		})
	}
	return &Team{ID: side, Name: side, Side: side, Players: players}
}

func TestSameSeedProducesSameRoute(t *testing.T) {
	input := &MatchInput{
		MatchID: "match_same_seed",
		TeamA:   makeTestTeam(SideT, []string{"t1", "t2", "t3", "t4", "t5"}),
		TeamB:   makeTestTeam(SideCT, []string{"c1", "c2", "c3", "c4", "c5"}),
		MapID:   "de_dust2",
		Seed:    12345,
	}

	res1, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("first simulate failed: %v", err)
	}
	res2, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("second simulate failed: %v", err)
	}

	if res1.Rounds[0].RouteMain != res2.Rounds[0].RouteMain {
		t.Fatalf("route should be deterministic for same seed: %s vs %s", res1.Rounds[0].RouteMain, res2.Rounds[0].RouteMain)
	}
}

func TestHigherPowerWins(t *testing.T) {
	weak := &Combatant{PlayerID: "weak", Name: "weak", Side: SideCT, Entry: 10, Aim: 10, Firepower: 10, Alive: true}
	strong := &Combatant{PlayerID: "strong", Name: "strong", Side: SideT, Entry: 90, Aim: 90, Firepower: 90, Alive: true}

	input := &MatchInput{
		MatchID: "match_duel",
		TeamA:   &Team{ID: "T", Name: "T", Side: SideT, Players: []*Combatant{strong}},
		TeamB:   &Team{ID: "CT", Name: "CT", Side: SideCT, Players: []*Combatant{weak}},
		MapID:   "de_dust2",
		Seed:    1,
	}

	res, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}

	round := res.Rounds[0]
	if len(round.Events) < 3 {
		t.Fatalf("expected at least 3 events, got %d", len(round.Events))
	}

	killFound := false
	for _, ev := range round.Events {
		if ev.EventType == "KILL" && ev.AttackerID == "strong" && ev.VictimID == "weak" {
			killFound = true
			break
		}
	}
	if !killFound {
		t.Fatalf("expected strong to kill weak")
	}
}

func TestRoundResultStructure(t *testing.T) {
	input := &MatchInput{
		MatchID: "match_structure",
		TeamA:   makeTestTeam(SideT, []string{"t1", "t2", "t3", "t4", "t5"}),
		TeamB:   makeTestTeam(SideCT, []string{"c1", "c2", "c3", "c4", "c5"}),
		MapID:   "de_dust2",
		Seed:    42,
	}

	res, err := NewMatchEngine(input).StartMatch(context.Background())
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}

	if res.MatchInfo == nil || res.MatchInfo.MatchID != "match_structure" {
		t.Fatalf("match_info missing or wrong id")
	}
	if len(res.Rounds) != 1 {
		t.Fatalf("expected 1 round, got %d", len(res.Rounds))
	}
	round := res.Rounds[0]
	if round.RoundNumber != 1 {
		t.Fatalf("expected round number 1, got %d", round.RoundNumber)
	}
	if round.Winner != SideT && round.Winner != SideCT {
		t.Fatalf("unexpected winner: %s", round.Winner)
	}
	if len(round.Events) == 0 {
		t.Fatalf("round events should not be empty")
	}
	if len(round.PlayerStates) != 10 {
		t.Fatalf("expected 10 player states, got %d", len(round.PlayerStates))
	}
	if res.FinalStats == nil || len(res.FinalStats.PlayerStats) != 10 {
		t.Fatalf("final stats missing or wrong length")
	}
}
