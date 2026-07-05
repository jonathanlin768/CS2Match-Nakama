package matchengine

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// MatchEngine 是单场比赛的推演实例。
type MatchEngine struct {
	input  *MatchInput
	rng    *rand.Rand
	state  *MatchState
	events []*GameEvent
}

// NewMatchEngine 创建一个新的比赛推演实例。
func NewMatchEngine(input *MatchInput) *MatchEngine {
	seed := input.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &MatchEngine{
		input: input,
		rng:   rand.New(rand.NewSource(seed)),
		state: &MatchState{},
	}
}

// StartMatch 执行完整推演并返回结果。MVP 阶段只推演一回合。
func (e *MatchEngine) StartMatch(ctx context.Context) (*MatchResult, error) {
	nowMs := time.Now().UnixMilli()
	matchInfo := &MatchInfo{
		MatchID:     e.input.MatchID,
		MapID:       e.input.MapID,
		MapName:     DefaultMapName(e.input.MapID),
		TeamAName:   e.input.TeamA.Name,
		TeamBName:   e.input.TeamB.Name,
		StartTime:   nowMs,
		TotalRounds: 1,
	}

	round := e.simulateRound(1, SideT)
	e.state.ScoreT = round.ScoreT
	e.state.ScoreCT = round.ScoreCT

	winner := SideCT
	if e.state.ScoreT > e.state.ScoreCT {
		winner = SideT
	}

	finalStats := e.buildFinalStats()

	return &MatchResult{
		MatchInfo:   matchInfo,
		Rounds:      []*RoundResult{round},
		FinalStats:  finalStats,
		Winner:      winner,
		TotalRounds: 1,
	}, nil
}

// simulateRound 推演一个回合。
func (e *MatchEngine) simulateRound(roundNumber int, sideAttacking string) *RoundResult {
	route := dust2AttackRoutes[e.rng.Intn(len(dust2AttackRoutes))]

	// 复制一份选手状态用于本回合。
	all := e.combatants()
	aliveT := filterBySide(all, SideT)
	aliveCT := filterBySide(all, SideCT)

	e.events = nil

	// ROUND_START 事件
	e.addEvent(&GameEvent{
		Timestamp: 0,
		EventType: "ROUND_START",
		Message:   fmt.Sprintf("Round %d 开始。T方选择 %s 进攻%s区。", roundNumber, route.Name, route.TargetSite),
	})

	// 按综合属性排序，高者先开枪
	sortCombatantsByPower(aliveT)
	sortCombatantsByPower(aliveCT)

	currentTime := int64(route.BaseTime + e.rng.Intn(5))
	firstKill := true

	pairs := min(len(aliveT), len(aliveCT))
	for i := 0; i < pairs; i++ {
		t := aliveT[i]
		ct := aliveCT[i]
		if !t.Alive || !ct.Alive {
			continue
		}

		winner, loser := e.resolveDuel(t, ct)
		loser.Alive = false
		loser.Deaths++
		winner.Kills++
		winner.Damage += 100

		weapon := WeaponT
		location := route.Name
		loc := RandomRouteLocation(route.ID, e.rng)
		if winner.Side == SideCT {
			weapon = WeaponCT
		}

		if firstKill {
			winner.FirstKills++
		}

		e.addEvent(&GameEvent{
			Timestamp:    currentTime,
			EventType:    "KILL",
			AttackerID:   winner.PlayerID,
			AttackerName: winner.Name,
			VictimID:     loser.PlayerID,
			VictimName:   loser.Name,
			Weapon:       weapon,
			Location:     loc,
			IsFirstKill:  firstKill,
			Message:      fmt.Sprintf("[%s] %s (%s) 击杀了 %s [%s]", formatTime(currentTime), winner.Name, weapon, loser.Name, location),
		})

		firstKill = false
		currentTime += int64(1 + e.rng.Intn(3))
	}

	// 统计存活
	survivorsT := countAlive(aliveT)
	survivorsCT := countAlive(aliveCT)

	var winner, reason string
	if survivorsT > survivorsCT {
		winner = SideT
		reason = "elimination"
	} else if survivorsCT > survivorsT {
		winner = SideCT
		reason = "elimination"
	} else {
		// 人数相等时默认 T 方通过下包获胜
		winner = SideT
		reason = "bomb_exploded"
	}

	scoreT, scoreCT := 0, 0
	if winner == SideT {
		scoreT = 1
	} else {
		scoreCT = 1
	}

	endTime := currentTime + int64(2+e.rng.Intn(4))
	if endTime > RoundTimeLimit {
		endTime = RoundTimeLimit
	}

	winnerText := "T方"
	if winner == SideCT {
		winnerText = "CT方"
	}
	e.addEvent(&GameEvent{
		Timestamp: endTime,
		EventType: "ROUND_END",
		Message:   fmt.Sprintf("%s获胜（%s）。比分 T %d - %d CT", winnerText, reason, scoreT, scoreCT),
	})

	return &RoundResult{
		RoundNumber:   roundNumber,
		SideAttacking: sideAttacking,
		Winner:        winner,
		WinReason:     reason,
		ScoreT:        scoreT,
		ScoreCT:       scoreCT,
		RouteMain:     route.ID,
		RouteSub:      "",
		Events:        append([]*GameEvent(nil), e.events...),
		PlayerStates:  buildPlayerStates(all),
	}
}

// resolveDuel 比较两名选手综合属性，高者获胜；相等时随机。
func (e *MatchEngine) resolveDuel(a, b *Combatant) (winner, loser *Combatant) {
	powerA := combatPower(a)
	powerB := combatPower(b)
	if powerA > powerB {
		return a, b
	}
	if powerB > powerA {
		return b, a
	}
	if e.rng.Intn(2) == 0 {
		return a, b
	}
	return b, a
}

// combatPower 计算选手综合战力。MVP 阶段使用 Entry + Aim + Firepower。
func combatPower(c *Combatant) int32 {
	return c.Entry + c.Aim + c.Firepower
}

func (e *MatchEngine) addEvent(ev *GameEvent) {
	e.events = append(e.events, ev)
}

func (e *MatchEngine) combatants() []*Combatant {
	var out []*Combatant
	for _, p := range e.input.TeamA.Players {
		out = append(out, cloneCombatant(p))
	}
	for _, p := range e.input.TeamB.Players {
		out = append(out, cloneCombatant(p))
	}
	return out
}

func cloneCombatant(c *Combatant) *Combatant {
	return &Combatant{
		PlayerID:  c.PlayerID,
		Name:      c.Name,
		Side:      c.Side,
		Entry:     c.Entry,
		Aim:       c.Aim,
		Trade:     c.Trade,
		Clutch:    c.Clutch,
		Firepower: c.Firepower,
		Gamesense: c.Gamesense,
		Alive:     true,
	}
}

func filterBySide(all []*Combatant, side string) []*Combatant {
	var out []*Combatant
	for _, c := range all {
		if c.Side == side {
			out = append(out, c)
		}
	}
	return out
}

func sortCombatantsByPower(list []*Combatant) {
	sort.Slice(list, func(i, j int) bool {
		return combatPower(list[i]) > combatPower(list[j])
	})
}

func countAlive(list []*Combatant) int {
	n := 0
	for _, c := range list {
		if c.Alive {
			n++
		}
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildPlayerStates(all []*Combatant) []*PlayerState {
	states := make([]*PlayerState, 0, len(all))
	for _, c := range all {
		states = append(states, &PlayerState{
			PlayerID:   c.PlayerID,
			PlayerName: c.Name,
			Side:       c.Side,
			IsAlive:    c.Alive,
			Kills:      c.Kills,
			Deaths:     c.Deaths,
			Damage:     c.Damage,
		})
	}
	return states
}

func (e *MatchEngine) buildFinalStats() *FinalStats {
	stats := &FinalStats{}
	for _, c := range e.input.TeamA.Players {
		stats.PlayerStats = append(stats.PlayerStats, buildPlayerMatchStats(c, SideT))
	}
	for _, c := range e.input.TeamB.Players {
		stats.PlayerStats = append(stats.PlayerStats, buildPlayerMatchStats(c, SideCT))
	}
	stats.ScoreT = e.state.ScoreT
	stats.ScoreCT = e.state.ScoreCT
	return stats
}

func buildPlayerMatchStats(c *Combatant, side string) *PlayerMatchStats {
	mk := 0
	if c.Kills >= 3 {
		mk = 1
	}
	adr := 0.0
	if c.Damage > 0 {
		adr = float64(c.Damage) // 1 回合，直接等于总伤害
	}
	return &PlayerMatchStats{
		PlayerID:   c.PlayerID,
		PlayerName: c.Name,
		Side:       side,
		Kills:      c.Kills,
		Deaths:     c.Deaths,
		ADR:        adr,
		FK:         c.FirstKills,
		MK:         mk,
	}
}

func formatTime(seconds int64) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
