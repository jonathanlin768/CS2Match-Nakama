package matchengine

import "sort"

type RoundPublicProjection struct {
	Players  []*PlayerState
	Bomb     *BombPublicState
	Controls []*NodeControlState
}

func ProjectRoundState(state *RoundState) (*RoundPublicProjection, error) {
	if state == nil {
		return nil, newError("SIMULATION_INVARIANT_ERROR", "cannot project nil round state")
	}
	if err := state.ClampAndValidate(); err != nil {
		return nil, err
	}
	projection := &RoundPublicProjection{}
	playerIDs := make([]string, 0, len(state.Players))
	for playerID := range state.Players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)
	for _, playerID := range playerIDs {
		player := state.Players[playerID]
		projection.Players = append(projection.Players, &PlayerState{
			PlayerID: player.Profile.PlayerID, PlayerName: player.Profile.DisplayName, DisplayName: player.Profile.DisplayName,
			Portrait: player.Profile.Portrait, TeamID: player.TeamID, Side: player.Side, IsAlive: player.Alive, Alive: player.Alive,
			HP: player.HP, Stamina: player.Stamina, Focus: player.Focus, CurrentNode: projectedNodeID(player.Location), HasBomb: player.HasBomb,
			Kills: player.Kills, Deaths: player.Deaths, Damage: player.Damage, RoleTags: append([]string(nil), player.Profile.RoleTags...), Weapon: cloneLoadout(player.Weapon),
		})
	}
	projection.Bomb = projectBombState(state.Bomb)
	nodeIDs := make([]string, 0, len(state.Nodes))
	for nodeID := range state.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	for _, nodeID := range nodeIDs {
		node := state.Nodes[nodeID]
		projection.Controls = append(projection.Controls, &NodeControlState{
			NodeID: nodeID, Status: string(node.ActualControl), UpdatedAt: node.UpdatedAt,
			KnownByT:  controlKnownAt(node.KnownControl[SideT], state.Timeline),
			KnownByCT: controlKnownAt(node.KnownControl[SideCT], state.Timeline),
		})
	}
	return projection, nil
}

func ProjectRoundResult(state *RoundState, input *RoundInput) (*RoundResult, error) {
	if input == nil {
		return nil, newError("INVALID_ROUND_INPUT", "cannot project round result without input metadata")
	}
	projection, err := ProjectRoundState(state)
	if err != nil {
		return nil, err
	}
	result := &RoundResult{
		RoundNumber: input.RoundNumber, Phase: input.phase, Half: input.half, OvertimeBlock: input.overtimeBlock,
		OvertimeRoundInBlock: input.overtimeRoundInBlock, IsSideSwitch: input.isSideSwitch, Seed: input.Seed,
		SideAttacking: SideT, TeamTID: state.TeamTID, TeamCTID: state.TeamCTID,
		StrategyTemplateID: state.Plan.TStrategyTemplateID, Events: cloneGameEvents(state.Events),
		CTSetupTemplateID: state.Plan.CTSetupTemplateID,
		PlayerStates:      projection.Players, Bomb: projection.Bomb, FinalControls: projection.Controls,
	}
	if state.Terminal != nil {
		result.Winner = state.Terminal.WinnerSide
		result.WinnerTeamID = state.Terminal.WinnerTeamID
		result.WinReason = state.Terminal.WinReason
	}
	return result, nil
}

func projectBombState(bomb BombState) *BombPublicState {
	status := string(bomb.Status)
	switch bomb.Status {
	case BombPlanting:
		status = BombStatusCarried
	case BombDefusing:
		status = BombStatusPlanted
	case BombExploded:
		status = BombStatusExplode
	}
	return &BombPublicState{
		Status: status, CarrierID: bomb.CarrierID, NodeID: projectedNodeID(bomb.Location), Site: bomb.PlantedSite,
		PlantedAt: bomb.PlantedAt, ExplodeAt: bomb.ExplodeAt, DroppedAt: bomb.DroppedAt,
	}
}

func projectedNodeID(location PlayerLocation) string {
	if location.NodeID != "" {
		return location.NodeID
	}
	if location.Edge != nil {
		if location.Edge.Progress >= 0.5 {
			return location.Edge.ToNode
		}
		return location.Edge.FromNode
	}
	return ""
}

func controlKnownAt(known KnownControlState, timeline int) bool {
	return known.Status != "" && known.Status != ControlUnknown && (known.ExpiresAt == 0 || timeline < known.ExpiresAt)
}

func cloneLoadout(loadout WeaponLoadout) WeaponLoadout {
	loadout.Grenades = append([]string(nil), loadout.Grenades...)
	return loadout
}

func cloneGameEvents(events []*GameEvent) []*GameEvent {
	out := make([]*GameEvent, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		copy := *event
		if event.Location != nil {
			location := *event.Location
			copy.Location = &location
		}
		if event.Reason != nil {
			copy.Reason = cloneEventReason(event.Reason)
		}
		if event.Bomb != nil {
			bomb := *event.Bomb
			copy.Bomb = &bomb
		}
		if event.Extra != nil {
			copy.Extra = make(map[string]interface{}, len(event.Extra))
			for key, value := range event.Extra {
				copy.Extra[key] = value
			}
		}
		if event.State != nil {
			copy.State = cloneEventStateSnapshot(event.State)
		}
		out = append(out, &copy)
	}
	return out
}

func cloneEventStateSnapshot(snapshot *EventStateSnapshot) *EventStateSnapshot {
	if snapshot == nil {
		return nil
	}
	copy := *snapshot
	copy.Players = make([]*PlayerState, 0, len(snapshot.Players))
	for _, player := range snapshot.Players {
		if player == nil {
			continue
		}
		playerCopy := *player
		playerCopy.RoleTags = append([]string(nil), player.RoleTags...)
		playerCopy.Weapon = cloneLoadout(player.Weapon)
		copy.Players = append(copy.Players, &playerCopy)
	}
	if snapshot.Bomb != nil {
		bomb := *snapshot.Bomb
		copy.Bomb = &bomb
	}
	copy.Controls = make([]*NodeControlState, 0, len(snapshot.Controls))
	for _, control := range snapshot.Controls {
		if control != nil {
			controlCopy := *control
			copy.Controls = append(copy.Controls, &controlCopy)
		}
	}
	return &copy
}
