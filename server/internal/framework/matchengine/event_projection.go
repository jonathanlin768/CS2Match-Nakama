package matchengine

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

func ProjectReasonRecord(record ReasonRecord, actionID, effectID string) (*EventReason, error) {
	if record.Code == "" {
		return nil, newError("INVALID_REASON", "reason code is required")
	}
	for _, change := range record.StateChanges {
		if change.Field == "" || !validReasonValue(change.Before) || !validReasonValue(change.After) || hiddenStateField(change.Field) {
			return nil, newError("INVALID_REASON", "reason contains an invalid or hidden state change")
		}
	}
	mainFactor := record.MainFactor
	if mainFactor == "" {
		mainFactor = record.Source
	}
	scoreDelta := record.ScoreDelta
	if scoreDelta == 0 && record.Value != 0 {
		scoreDelta = record.Value
	}
	if actionID == "" {
		actionID = record.SourceActionID
	}
	if effectID == "" {
		effectID = record.SourceEffectID
	}
	return &EventReason{
		Code: record.Code, MainFactor: mainFactor, Modifiers: append([]ReasonModifier(nil), record.Modifiers...), ScoreDelta: scoreDelta,
		Probability: cloneFloat(record.Probability), Formula: record.Formula, Inputs: copyStringFloatMap(record.Inputs),
		StateChanges: cloneReasonStateChanges(record.StateChanges), SourceActionID: actionID, SourceEffectID: effectID, Detail: record.Detail,
	}, nil
}

func validReasonValue(value ReasonValue) bool {
	set := 0
	if value.Number != nil {
		set++
	}
	if value.String != nil {
		set++
	}
	if value.Bool != nil {
		set++
	}
	switch value.Kind {
	case "Number":
		return set == 1 && value.Number != nil
	case "String":
		return set == 1 && value.String != nil
	case "Bool":
		return set == 1 && value.Bool != nil
	case "Null":
		return set == 0
	default:
		return false
	}
}

func hiddenStateField(field string) bool {
	lower := strings.ToLower(field)
	return strings.Contains(lower, "intent") || strings.Contains(lower, "actionqueue") || strings.Contains(lower, "actualcontrol") || strings.Contains(lower, "hidden")
}

func NumberReasonValue(value float64) ReasonValue { return ReasonValue{Kind: "Number", Number: &value} }
func StringReasonValue(value string) ReasonValue  { return ReasonValue{Kind: "String", String: &value} }
func BoolReasonValue(value bool) ReasonValue      { return ReasonValue{Kind: "Bool", Bool: &value} }
func NullReasonValue() ReasonValue                { return ReasonValue{Kind: "Null"} }

func snapshotForEvent(state *RoundState) *EventStateSnapshot {
	if state == nil {
		return nil
	}
	projection := &RoundPublicProjection{Bomb: projectBombState(state.Bomb)}
	for _, playerID := range sortedPlayerIDs(state) {
		player := state.Players[playerID]
		projection.Players = append(projection.Players, &PlayerState{
			PlayerID: player.Profile.PlayerID, PlayerName: player.Profile.DisplayName, DisplayName: player.Profile.DisplayName,
			Portrait: player.Profile.Portrait, TeamID: player.TeamID, Side: player.Side, IsAlive: player.Alive, Alive: player.Alive,
			HP: player.HP, Stamina: player.Stamina, Focus: player.Focus, CurrentNode: projectedNodeID(player.Location), HasBomb: player.HasBomb,
			Kills: player.Kills, Deaths: player.Deaths, Damage: player.Damage, RoleTags: append([]string(nil), player.Profile.RoleTags...), Weapon: cloneLoadout(player.Weapon),
		})
	}
	var nodeIDs []string
	for nodeID := range state.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	for _, nodeID := range nodeIDs {
		node := state.Nodes[nodeID]
		projection.Controls = append(projection.Controls, &NodeControlState{
			NodeID: nodeID, Status: string(node.ActualControl), UpdatedAt: node.UpdatedAt,
			KnownByT: controlKnownAt(node.KnownControl[SideT], state.Timeline), KnownByCT: controlKnownAt(node.KnownControl[SideCT], state.Timeline),
		})
	}
	scoreA, scoreB := state.ScoreByTeam[state.TeamAID], state.ScoreByTeam[state.TeamBID]
	return &EventStateSnapshot{
		ScoreTeamA: scoreA, ScoreTeamB: scoreB, ScoreT: state.ScoreByTeam[state.TeamTID], ScoreCT: state.ScoreByTeam[state.TeamCTID],
		Players: projection.Players, Bomb: projection.Bomb, Controls: projection.Controls,
	}
}

func eventLocation(state *RoundState, location PlayerLocation, eventID, sourceObjectID string) *Location {
	seed := deriveSeed(state.Seed, "event_location", eventID, sourceObjectID)
	if location.Edge != nil {
		return &Location{
			Name: location.Edge.DisplayName, X: clampProbability(location.Edge.X), Y: clampProbability(location.Edge.Y),
			SourceType: "OnEdge", SourceID: location.Edge.EdgeID, Seed: seed,
		}
	}
	node := state.Nodes[location.NodeID]
	if node == nil {
		return nil
	}
	x, y, sourceType := node.Node.X, node.Node.Y, "Node"
	switch node.Node.Shape {
	case "Circle":
		angle := stableUnit(seed, "angle") * 2 * math.Pi
		radius := math.Sqrt(stableUnit(seed, "radius")) * node.Node.Radius
		x += math.Cos(angle) * radius
		y += math.Sin(angle) * radius
		sourceType = "Area"
	case "Polygon":
		if polygon := parseEventPolygon(node.Node.Points); len(polygon) >= 3 {
			minX, maxX, minY, maxY := eventPolygonBounds(polygon)
			for attempt := 0; attempt < 16; attempt++ {
				candidateX := minX + stableUnit(seed, "polygon_x", attempt)*(maxX-minX)
				candidateY := minY + stableUnit(seed, "polygon_y", attempt)*(maxY-minY)
				if eventPointInPolygon(candidateX, candidateY, polygon) {
					x, y = candidateX, candidateY
					break
				}
			}
			sourceType = "Area"
		}
	}
	return &Location{Name: node.Node.Name, X: clampProbability(x), Y: clampProbability(y), SourceType: sourceType, SourceID: node.Node.ID, Floor: node.Node.Floor, Seed: seed}
}

func BuildExplainableReport(round *RoundResult) *ExplainableReport {
	if round == nil {
		return &ExplainableReport{}
	}
	report := &ExplainableReport{StrategySummary: "template=" + round.StrategyTemplateID + "; route=" + round.RouteMain}
	winnerID := round.WinnerTeamID
	seenWin, seenLoss := map[string]bool{}, map[string]bool{}
	var decisions []string
	for _, event := range round.Events {
		if event == nil {
			continue
		}
		if keyReportEvent(event.EventType) {
			report.KeyEvents = append(report.KeyEvents, event)
		}
		if event.EventType == EventRotate || event.EventType == EventReinforce {
			decisions = append(decisions, event.EventType+":"+event.SourceActionID)
		}
		if event.Reason == nil {
			continue
		}
		key := event.Reason.Code + ":" + event.Reason.SourceActionID + ":" + event.Reason.SourceEffectID
		actorTeam := event.AttackerTeamID
		if event.EventType == EventRoundEnd {
			actorTeam = winnerID
		}
		if actorTeam == winnerID && !seenWin[key] {
			report.WinFactors = append(report.WinFactors, cloneEventReason(event.Reason))
			seenWin[key] = true
		} else if actorTeam != "" && actorTeam != winnerID && !seenLoss[key] {
			report.LossReasons = append(report.LossReasons, cloneEventReason(event.Reason))
			seenLoss[key] = true
		}
		if event.VictimTeamID != "" && event.VictimTeamID != winnerID && !seenLoss[key] {
			report.LossReasons = append(report.LossReasons, cloneEventReason(event.Reason))
			seenLoss[key] = true
		}
	}
	if len(decisions) > 0 {
		sort.Strings(decisions)
		report.StrategySummary += "; decisions=" + strings.Join(decisions, ",")
	}
	return report
}

func BuildMatchExplainableReport(rounds []*RoundResult) *ExplainableReport {
	report := &ExplainableReport{}
	var summaries []string
	for _, round := range rounds {
		if round == nil || round.Report == nil {
			continue
		}
		report.KeyEvents = append(report.KeyEvents, round.Report.KeyEvents...)
		report.WinFactors = append(report.WinFactors, cloneEventReasons(round.Report.WinFactors)...)
		report.LossReasons = append(report.LossReasons, cloneEventReasons(round.Report.LossReasons)...)
		summaries = append(summaries, "R"+strconv.Itoa(round.RoundNumber)+"["+round.Report.StrategySummary+"]")
	}
	report.StrategySummary = strings.Join(summaries, "; ")
	return report
}

func keyReportEvent(eventType string) bool {
	return oneOf(eventType, EventRoundStart, EventStrategyAdjusted, EventKill, EventRotate, EventReinforce, EventControlGained, EventBombDrop, EventBombPickup, EventPlantStart, EventPlantInterrupt, EventBombPlant, EventDefuseStart, EventDefuseInterrupt, EventBombDefuse, EventBombExplode, EventRoundEnd)
}

func cloneEventReason(reason *EventReason) *EventReason {
	if reason == nil {
		return nil
	}
	copy := *reason
	copy.Modifiers = append([]ReasonModifier(nil), reason.Modifiers...)
	copy.Probability = cloneFloat(reason.Probability)
	copy.Inputs = copyStringFloatMap(reason.Inputs)
	copy.StateChanges = cloneReasonStateChanges(reason.StateChanges)
	return &copy
}

func cloneEventReasons(reasons []*EventReason) []*EventReason {
	out := make([]*EventReason, 0, len(reasons))
	for _, reason := range reasons {
		out = append(out, cloneEventReason(reason))
	}
	return out
}

func cloneReasonStateChanges(changes []ReasonStateChange) []ReasonStateChange {
	out := make([]ReasonStateChange, len(changes))
	for index, change := range changes {
		out[index] = ReasonStateChange{Field: change.Field, Before: cloneReasonValue(change.Before), After: cloneReasonValue(change.After)}
	}
	return out
}

func cloneReasonValue(value ReasonValue) ReasonValue {
	copy := value
	copy.Number = cloneFloat(value.Number)
	if value.String != nil {
		text := *value.String
		copy.String = &text
	}
	if value.Bool != nil {
		flag := *value.Bool
		copy.Bool = &flag
	}
	return copy
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

type eventPoint struct{ x, y float64 }

func parseEventPolygon(raw string) []eventPoint {
	var points []eventPoint
	for _, pair := range strings.Split(raw, ";") {
		values := strings.Split(strings.TrimSpace(pair), ",")
		if len(values) != 2 {
			continue
		}
		x, xErr := strconv.ParseFloat(strings.TrimSpace(values[0]), 64)
		y, yErr := strconv.ParseFloat(strings.TrimSpace(values[1]), 64)
		if xErr == nil && yErr == nil {
			points = append(points, eventPoint{x: x, y: y})
		}
	}
	return points
}

func eventPolygonBounds(points []eventPoint) (float64, float64, float64, float64) {
	minX, maxX, minY, maxY := points[0].x, points[0].x, points[0].y, points[0].y
	for _, point := range points[1:] {
		minX, maxX = math.Min(minX, point.x), math.Max(maxX, point.x)
		minY, maxY = math.Min(minY, point.y), math.Max(maxY, point.y)
	}
	return minX, maxX, minY, maxY
}

func eventPointInPolygon(x, y float64, polygon []eventPoint) bool {
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		left, right := polygon[i], polygon[j]
		if (left.y > y) != (right.y > y) && x < (right.x-left.x)*(y-left.y)/(right.y-left.y)+left.x {
			inside = !inside
		}
	}
	return inside
}
