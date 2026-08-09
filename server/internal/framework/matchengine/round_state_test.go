package matchengine

import "testing"

func makeTestRoundInput(seed int64) *RoundInput {
	match := makeTestInput(seed)
	return &RoundInput{
		MatchID: match.MatchID, RoundNumber: 1, MapID: match.MapID, MapVersion: match.MapVersion, Seed: seed,
		RuleSet: match.RuleSet, MapConfig: match.MapConfig, WeaponSpecs: match.WeaponSpecs, SideLoadouts: match.SideLoadouts,
		TeamT: match.TeamA, TeamCT: match.TeamB, ScoreByTeam: map[string]int{match.TeamA.TeamID: 0, match.TeamB.TeamID: 0},
	}
}

func makeTestRoundState(t *testing.T, seed int64) *RoundState {
	t.Helper()
	input := makeTestRoundInput(seed)
	state, err := NewRoundState(input, RoundPlan{TStrategyTemplateID: "TPL_A", CTSetupTemplateID: "TPL_CT", BombCarrierID: input.TeamT.Players[0].PlayerID})
	if err != nil {
		t.Fatalf("NewRoundState() error = %v", err)
	}
	return state
}

func TestNewRoundStateBuildsAuthoritativePlayersAndUniqueBombCarrier(t *testing.T) {
	state := makeTestRoundState(t, 101)
	if len(state.Players) != 10 || len(state.Nodes) != len(makeTestMapConfig().Nodes) {
		t.Fatalf("unexpected state size: players=%d nodes=%d", len(state.Players), len(state.Nodes))
	}
	carrierCount := 0
	for _, player := range state.Players {
		if !player.Alive || player.HP != 100 || player.Stamina != 100 || player.Focus != 100 {
			t.Fatalf("player not initialized from combat constants: %+v", player)
		}
		if player.Side == SideT && player.Weapon.Primary != WeaponAK47 {
			t.Fatalf("T loadout mismatch: %+v", player.Weapon)
		}
		if player.Side == SideCT && player.Weapon.Primary != WeaponM4A1S {
			t.Fatalf("CT loadout mismatch: %+v", player.Weapon)
		}
		if player.HasBomb {
			carrierCount++
		}
	}
	if carrierCount != 1 || state.Bomb.Status != BombCarried || state.Bomb.CarrierID != state.Plan.BombCarrierID {
		t.Fatalf("bomb carrier invariant failed: count=%d bomb=%+v", carrierCount, state.Bomb)
	}
}

func TestNodeControlContestResolutionAndTTLDecay(t *testing.T) {
	state := makeTestRoundState(t, 102)
	node := state.Nodes["A_SITE"]
	node.ActualControl = ControlContested
	if err := node.ResolveContest(ControlT, 12, 8, map[string][]string{SideT: {"team_a_p2", "team_a_p1"}}); err != nil {
		t.Fatalf("ResolveContest() error = %v", err)
	}
	known := node.KnownControl[SideT]
	if node.ActualControl != ControlT || known.ExpiresAt != 20 || known.ObservedBy[0] != "team_a_p1" {
		t.Fatalf("unexpected resolved control: node=%+v known=%+v", node, known)
	}
	node.DecayKnownControl(SideT, 19)
	if node.KnownControl[SideT].Status != ControlT {
		t.Fatal("known control decayed before TTL")
	}
	node.DecayKnownControl(SideT, 20)
	if node.KnownControl[SideT].Status != ControlUnknown {
		t.Fatal("known control did not decay at TTL")
	}
	if err := node.ResolveContest(ControlContested, 21, 8, nil); err == nil {
		t.Fatal("contest resolved to Contested without error")
	}
}

func TestBombStateLegalLifecycleAndAbsoluteTimes(t *testing.T) {
	location := PlayerLocation{NodeID: "A_SITE"}
	bomb := BombState{Status: BombCarried, CarrierID: "t1", Location: location}
	if err := bomb.StartPlant("t1", "plant-1", "A", 20, 24); err != nil {
		t.Fatalf("StartPlant() error = %v", err)
	}
	if err := bomb.CompletePlant(location, 24, 64); err != nil {
		t.Fatalf("CompletePlant() error = %v", err)
	}
	if err := bomb.StartDefuse("ct1", "defuse-1", 54, 64); err != nil {
		t.Fatalf("StartDefuse() equality error = %v", err)
	}
	if err := bomb.CompleteDefuse(64); err != nil || bomb.Status != BombDefused {
		t.Fatalf("CompleteDefuse() = %v, status=%s", err, bomb.Status)
	}
	if err := bomb.Explode(64); err == nil {
		t.Fatal("defused bomb exploded")
	}

	dropped := BombState{Status: BombPlanting, CarrierID: "t1", PlantActionID: "plant-2", PlantActorID: "t1"}
	if err := dropped.Drop(PlayerLocation{Edge: &OnEdgeLocation{EdgeID: "E3", FromNode: "A_LONG", ToNode: "A_SITE", Progress: 0.4}}, 31); err != nil {
		t.Fatalf("Drop() error = %v", err)
	}
	if dropped.Status != BombDropped || dropped.CarrierID != "" || dropped.DroppedAt != 31 || dropped.PlantActionID != "" {
		t.Fatalf("drop transition left stale state: %+v", dropped)
	}
}

func TestRoundStateCentralClampAndInvariants(t *testing.T) {
	state := makeTestRoundState(t, 103)
	player := state.Players["team_a_p2"]
	player.HP, player.Stamina, player.Focus, player.Momentum = 150, -5, 130, 150
	state.MomentumT, state.MomentumCT = 200, -200
	if err := state.ClampAndValidate(); err != nil {
		t.Fatalf("ClampAndValidate() error = %v", err)
	}
	if player.HP != 100 || player.Stamina != 0 || player.Focus != 100 || player.Momentum != 100 || state.MomentumT != 100 || state.MomentumCT != -100 {
		t.Fatalf("central clamp mismatch: player=%+v momentum=%d/%d", player, state.MomentumT, state.MomentumCT)
	}
	if clampProbability(-0.1) != 0 || clampProbability(1.1) != 1 || clampProbability(0.4) != 0.4 {
		t.Fatal("probability clamp mismatch")
	}

	player.Location = PlayerLocation{}
	assertEngineErrorCode(t, state.ClampAndValidate(), "SIMULATION_INVARIANT_ERROR")
	player.Location = PlayerLocation{NodeID: "T_SPAWN"}
	player.HP, player.Alive = 50, false
	assertEngineErrorCode(t, state.ClampAndValidate(), "SIMULATION_INVARIANT_ERROR")
}

func TestRoundProjectionIsStableAndCannotDriveInternalState(t *testing.T) {
	state := makeTestRoundState(t, 104)
	state.Events = []*GameEvent{{EventID: "evt", SourceActionID: "act", EventType: EventDamage, Extra: map[string]interface{}{"damage": 10}}}
	projection, err := ProjectRoundState(state)
	if err != nil {
		t.Fatalf("ProjectRoundState() error = %v", err)
	}
	for index := 1; index < len(projection.Players); index++ {
		if projection.Players[index-1].PlayerID > projection.Players[index].PlayerID {
			t.Fatal("player projection is not sorted")
		}
	}
	first := projection.Players[0]
	internal := state.Players[first.PlayerID]
	first.HP = 1
	first.RoleTags[0] = "mutated"
	first.Weapon.Grenades = append(first.Weapon.Grenades, "fake")
	projection.Bomb.CarrierID = "mutated"
	projection.Controls[0].Status = "mutated"
	if internal.HP == 1 || internal.Profile.RoleTags[0] == "mutated" || len(internal.Weapon.Grenades) != 0 || state.Bomb.CarrierID == "mutated" || string(state.Nodes[projection.Controls[0].NodeID].ActualControl) == "mutated" {
		t.Fatal("public DTO mutation flowed back into authoritative state")
	}

	result, err := ProjectRoundResult(state, makeTestRoundInput(104))
	if err != nil {
		t.Fatalf("ProjectRoundResult() error = %v", err)
	}
	result.Events[0].Extra["damage"] = 999
	result.PlayerStates[0].HP = 2
	if state.Events[0].Extra["damage"] == 999 || state.Players[result.PlayerStates[0].PlayerID].HP == 2 {
		t.Fatal("RoundResult mutation flowed back into authoritative state")
	}
}
