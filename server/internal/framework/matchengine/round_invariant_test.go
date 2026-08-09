package matchengine

import (
	"fmt"
	"strings"
	"testing"
)

func TestRoundInvariantSkeletonAcceptsCausalFixture(t *testing.T) {
	round := validInvariantFixture()
	if violations := roundInvariantViolations(round); len(violations) > 0 {
		t.Fatalf("valid fixture rejected: %s", strings.Join(violations, "; "))
	}
}

func TestRoundInvariantSkeletonDetectsViolations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*RoundResult)
		want   string
	}{
		{
			name: "gameplay event after round end",
			mutate: func(round *RoundResult) {
				round.Events = append(round.Events, &GameEvent{Timestamp: 12, EventType: EventKill, VictimID: "p2"})
			},
			want: "after ROUND_END",
		},
		{
			name: "death before damage",
			mutate: func(round *RoundResult) {
				round.Events[0], round.Events[1] = round.Events[1], round.Events[0]
				round.Events[0].Timestamp = 9
			},
			want: "no prior DAMAGE",
		},
		{
			name: "bomb carrier mismatch",
			mutate: func(round *RoundResult) {
				round.Bomb.CarrierID = "p2"
			},
			want: "carrier",
		},
		{
			name: "alive player with zero hp",
			mutate: func(round *RoundResult) {
				round.PlayerStates[0].HP = 0
			},
			want: "alive with non-positive HP",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			round := validInvariantFixture()
			tt.mutate(round)
			violations := strings.Join(roundInvariantViolations(round), "; ")
			if !strings.Contains(violations, tt.want) {
				t.Fatalf("expected %q violation, got %q", tt.want, violations)
			}
		})
	}
}

func validInvariantFixture() *RoundResult {
	return &RoundResult{
		Winner:    SideT,
		WinReason: "elimination",
		Events: []*GameEvent{
			{Timestamp: 10, EventType: "DAMAGE", AttackerID: "p1", VictimID: "p2"},
			{Timestamp: 10, EventType: EventKill, AttackerID: "p1", VictimID: "p2"},
			{Timestamp: 11, EventType: EventRoundEnd},
		},
		PlayerStates: []*PlayerState{
			{PlayerID: "p1", Side: SideT, Alive: true, IsAlive: true, HP: 100, HasBomb: true},
			{PlayerID: "p2", Side: SideCT, Alive: false, IsAlive: false, HP: 0},
		},
		Bomb: &BombPublicState{Status: BombStatusCarried, CarrierID: "p1"},
	}
}

func roundInvariantViolations(round *RoundResult) []string {
	if round == nil {
		return []string{"round is nil"}
	}
	var violations []string
	roundEnded := false
	damaged := map[string]bool{}
	lastTimestamp := int64(-1)
	for _, event := range round.Events {
		if event.Timestamp < lastTimestamp {
			violations = append(violations, "event timestamps are not monotonic")
		}
		lastTimestamp = event.Timestamp
		if roundEnded && event.EventType != EventMatchEnd {
			violations = append(violations, fmt.Sprintf("%s appears after ROUND_END", event.EventType))
		}
		switch event.EventType {
		case "DAMAGE":
			damaged[event.VictimID] = true
		case EventKill:
			if !damaged[event.VictimID] {
				violations = append(violations, fmt.Sprintf("KILL for %s has no prior DAMAGE", event.VictimID))
			}
		case EventRoundEnd:
			if roundEnded {
				violations = append(violations, "multiple ROUND_END events")
			}
			roundEnded = true
		}
	}
	if !roundEnded {
		violations = append(violations, "missing ROUND_END")
	}

	hasBomb := ""
	for _, player := range round.PlayerStates {
		if player.Alive && player.HP <= 0 {
			violations = append(violations, fmt.Sprintf("player %s is alive with non-positive HP", player.PlayerID))
		}
		if !player.Alive && player.HP > 0 {
			violations = append(violations, fmt.Sprintf("player %s is dead with positive HP", player.PlayerID))
		}
		if player.Alive != player.IsAlive {
			violations = append(violations, fmt.Sprintf("player %s alive flags disagree", player.PlayerID))
		}
		if player.HasBomb {
			if hasBomb != "" {
				violations = append(violations, "multiple players carry the bomb")
			}
			hasBomb = player.PlayerID
		}
	}
	if round.Bomb == nil {
		return append(violations, "bomb state is nil")
	}
	if round.Bomb.Status == BombStatusCarried {
		if hasBomb == "" || hasBomb != round.Bomb.CarrierID {
			violations = append(violations, "bomb carrier does not match player HasBomb state")
		}
	} else if hasBomb != "" {
		violations = append(violations, "player carries bomb while bomb is not Carried")
	}
	return violations
}
