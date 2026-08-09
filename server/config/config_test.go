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
	if len(players) != 10 {
		t.Fatalf("expected 10 players, got %d", len(players))
	}
	p := GetPlayer("player_niko")
	if p == nil {
		t.Fatal("player_niko not found")
	}
	if p.Team != "Falcons" {
		t.Fatalf("expected Falcons, got %s", p.Team)
	}
	if p.Portrait != "portraits/player_niko.jpg" {
		t.Fatalf("unexpected portrait: %s", p.Portrait)
	}
	t.Logf("Loaded %d tables and %d players, first: %s (entry=%d)", TableCount(), len(players), p.Name, p.Entry)
}
