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
	if TableCount() < 13 {
		t.Fatalf("expected at least 13 tables, got %d", TableCount())
	}
	players := Global.TbPlayer.GetDataList()
	if len(players) != 20 {
		t.Fatalf("expected 20 players, got %d", len(players))
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
	if p.CardImage != "player-cards/niko2.png" {
		t.Fatalf("unexpected card image: %s", p.CardImage)
	}
	if p.AvatarCropWidth <= 0 || p.AvatarCropHeight <= 0 {
		t.Fatalf("expected configured avatar crop, got %v x %v", p.AvatarCropWidth, p.AvatarCropHeight)
	}
	legacy := GetPlayer("player_chopper")
	if legacy == nil || legacy.CardImage != "" || legacy.AvatarCropWidth != 0 || legacy.Portrait == "" {
		t.Fatalf("expected player_chopper to retain portrait-only fallback: %+v", legacy)
	}
	t.Logf("Loaded %d tables and %d players, first: %s (entry=%d)", TableCount(), len(players), p.Name, p.Entry)
}

func TestTeamLookupAndTutorialValidation(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if team := GetTeam("team_vitality"); team == nil || team.Name != "Team Vitality" {
		t.Fatalf("unexpected team: %#v", team)
	}
	if got := len(PlayersByTeam("team_spirit")); got != 5 {
		t.Fatalf("expected 5 Spirit players, got %d", got)
	}
	tutorial := EnabledTutorialBattle()
	if tutorial == nil || tutorial.Id != "tutorial_default" || tutorial.Version != 1 {
		t.Fatalf("unexpected tutorial: %#v", tutorial)
	}
}
