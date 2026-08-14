package cfg

import (
	"testing"
)

func TestConfigLoad(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if Global == nil {
		t.Fatal("Global is nil")
	}
	if Global.Tbitem == nil {
		t.Fatal("Tbitem not loaded")
	}
	if Global.TbPlayer == nil {
		t.Fatal("TbPlayer not loaded")
	}
	if Global.TbMapNode == nil {
		t.Fatal("TbMapNode not loaded")
	}
	if TableCount() < 11 {
		t.Fatalf("expected at least 11 tables, got %d", TableCount())
	}
	players := Global.TbPlayer.GetDataList()
	if len(players) == 0 {
		t.Fatal("expected at least one player")
	}
	p := GetPlayer("player_niko")
	if p == nil {
		t.Fatal("player_niko not found")
	}
	if p.TeamId != "team_falcons" {
		t.Fatalf("expected team_falcons, got %s", p.TeamId)
	}
	if p.Portrait != "portraits/player_niko.jpg" {
		t.Fatalf("unexpected portrait: %s", p.Portrait)
	}
	t.Logf("Loaded %d tables and %d players, first: %s (entry=%d)", TableCount(), len(players), p.Name, p.Entry)
}

func TestValidateAllowsTutorialCandidateOpponentOverlap(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	tutorial := GetTutorialBattle("tutorial_default")
	if tutorial == nil {
		t.Fatal("tutorial_default not found")
	}
	candidateIDs := append([]string{}, tutorial.Tier5PlayerIds...)
	candidateIDs = append(candidateIDs, tutorial.Tier4PlayerIds...)
	candidateIDs = append(candidateIDs, tutorial.Tier3PlayerIds...)
	candidateIDs = append(candidateIDs, tutorial.Tier2PlayerIds...)
	candidateIDs = append(candidateIDs, tutorial.Tier1PlayerIds...)
	hasOverlap := false
	for _, candidateID := range candidateIDs {
		for _, opponentID := range tutorial.OpponentPlayerIds {
			if candidateID == opponentID {
				hasOverlap = true
				break
			}
		}
	}
	if !hasOverlap {
		t.Fatal("tutorial fixture should contain a candidate/opponent overlap")
	}
	if err := Validate(); err != nil {
		t.Fatalf("candidate/opponent overlap should be valid: %v", err)
	}
}

func TestValidateStillRejectsDuplicateOpponentPlayer(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	tutorial := GetTutorialBattle("tutorial_default")
	if tutorial == nil {
		t.Fatal("tutorial_default not found")
	}
	original := append([]string(nil), tutorial.OpponentPlayerIds...)
	t.Cleanup(func() { tutorial.OpponentPlayerIds = original })
	tutorial.OpponentPlayerIds[1] = tutorial.OpponentPlayerIds[0]
	if err := Validate(); err == nil {
		t.Fatal("duplicate opponent player should remain invalid")
	}
}
