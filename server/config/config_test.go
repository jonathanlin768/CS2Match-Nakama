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
	if TableCount() != 2 {
		t.Fatalf("expected 2 tables, got %d", TableCount())
	}
	players := Global.TbPlayer.GetDataList()
	if len(players) != 10 {
		t.Fatalf("expected 10 players, got %d", len(players))
	}
	p := GetPlayer("player_niko")
	if p == nil {
		t.Fatal("player_niko not found")
	}
	if p.Name != "NiKo" {
		t.Fatalf("expected NiKo, got %s", p.Name)
	}
	t.Logf("Loaded %d players, first: %s (entry=%d)", len(players), p.Name, p.Entry)
}
