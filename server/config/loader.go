package cfg

import (
	"embed"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

//go:embed data/*.json
var configData embed.FS

// Global is the singleton config tables instance, available after Init()
var Global *Tables

// Init loads all config table data. Call from InitModule.
func Init() error {
	entries, err := configData.ReadDir("data")
	if err != nil {
		return fmt.Errorf("config: failed to read embedded config data: %w", err)
	}

	dataMap := make(map[string][]map[string]interface{})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		raw, err := configData.ReadFile("data/" + entry.Name())
		if err != nil {
			return fmt.Errorf("config: failed to read %s: %w", entry.Name(), err)
		}

		var rows []map[string]interface{}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return fmt.Errorf("config: failed to parse %s: %w", entry.Name(), err)
		}

		tableName := strings.TrimSuffix(entry.Name(), ".json")
		dataMap[tableName] = rows
	}

	loader := func(tableName string) ([]map[string]interface{}, error) {
		data, ok := dataMap[tableName]
		if !ok {
			return nil, fmt.Errorf("config: table %s not found", tableName)
		}
		return data, nil
	}

	tables, err := NewTables(loader)
	if err != nil {
		return fmt.Errorf("config: failed to init tables: %w", err)
	}

	Global = tables
	return Validate()
}

// TableCount returns the number of loaded tables
func TableCount() int {
	if Global == nil {
		return 0
	}
	value := reflect.ValueOf(Global).Elem()
	count := 0
	for index := 0; index < value.NumField(); index++ {
		if !value.Field(index).IsNil() {
			count++
		}
	}
	return count
}

// GetPlayer returns a player config by id.
func GetPlayer(id string) *Player {
	if Global == nil || Global.TbPlayer == nil {
		return nil
	}
	return Global.TbPlayer.Get(id)
}

// GetTeam returns a team config by stable id.
func GetTeam(id string) *Team {
	if Global == nil || Global.TbTeam == nil {
		return nil
	}
	return Global.TbTeam.Get(id)
}

// PlayersByTeam returns a deterministic lineup candidate list for a team.
func PlayersByTeam(teamID string) []*Player {
	if Global == nil || Global.TbPlayer == nil {
		return nil
	}
	players := make([]*Player, 0)
	for _, player := range Global.TbPlayer.GetDataList() {
		if player != nil && player.TeamId == teamID {
			players = append(players, player)
		}
	}
	sort.Slice(players, func(i, j int) bool { return players[i].Id < players[j].Id })
	return players
}

// GetTutorialBattle returns a tutorial config by id.
func GetTutorialBattle(id string) *TutorialBattle {
	if Global == nil || Global.TbTutorialBattle == nil {
		return nil
	}
	return Global.TbTutorialBattle.Get(id)
}

// EnabledTutorialBattle returns the first enabled tutorial config.
func EnabledTutorialBattle() *TutorialBattle {
	if Global == nil || Global.TbTutorialBattle == nil {
		return nil
	}
	for _, tutorial := range Global.TbTutorialBattle.GetDataList() {
		if tutorial != nil && tutorial.Enabled {
			return tutorial
		}
	}
	return nil
}

// Validate checks cross-table constraints which must hold before the module starts.
func Validate() error {
	if Global == nil || Global.TbTeam == nil || Global.TbPlayer == nil || Global.TbTutorialBattle == nil {
		return fmt.Errorf("config: Team, Player and TutorialBattle tables are required")
	}
	for _, player := range Global.TbPlayer.GetDataList() {
		if player == nil || GetTeam(player.TeamId) == nil {
			return fmt.Errorf("config: player %v references unknown team", player)
		}
	}
	for _, tutorial := range Global.TbTutorialBattle.GetDataList() {
		if tutorial == nil || !tutorial.Enabled {
			continue
		}
		if tutorial.Budget <= 0 || tutorial.RosterSize != 5 || GetTeam(tutorial.OpponentTeamId) == nil {
			return fmt.Errorf("config: tutorial %s has invalid budget, roster size or opponent team", tutorial.Id)
		}
		seen := make(map[string]struct{})
		priceByPlayer := make(map[string]int)
		pools := map[int][]string{5: tutorial.Tier5PlayerIds, 4: tutorial.Tier4PlayerIds, 3: tutorial.Tier3PlayerIds, 2: tutorial.Tier2PlayerIds, 1: tutorial.Tier1PlayerIds}
		for price, ids := range pools {
			for _, id := range ids {
				if GetPlayer(id) == nil {
					return fmt.Errorf("config: tutorial %s references unknown player %s", tutorial.Id, id)
				}
				if _, duplicate := seen[id]; duplicate {
					return fmt.Errorf("config: tutorial %s repeats player %s across price tiers", tutorial.Id, id)
				}
				seen[id], priceByPlayer[id] = struct{}{}, price
			}
		}
		if len(tutorial.OpponentPlayerIds) != int(tutorial.RosterSize) {
			return fmt.Errorf("config: tutorial %s opponent requires exactly %d players", tutorial.Id, tutorial.RosterSize)
		}
		opponents := make(map[string]struct{}, len(tutorial.OpponentPlayerIds))
		for _, id := range tutorial.OpponentPlayerIds {
			player := GetPlayer(id)
			if player == nil || player.TeamId != tutorial.OpponentTeamId {
				return fmt.Errorf("config: tutorial %s has invalid opponent player %s", tutorial.Id, id)
			}
			if _, duplicate := opponents[id]; duplicate {
				return fmt.Errorf("config: tutorial %s repeats opponent player %s", tutorial.Id, id)
			}
			opponents[id] = struct{}{}
		}
		prices := make([]int, 0, len(priceByPlayer))
		for _, price := range priceByPlayer {
			prices = append(prices, price)
		}
		sort.Ints(prices)
		if len(prices) < int(tutorial.RosterSize) {
			return fmt.Errorf("config: tutorial %s cannot form a full roster", tutorial.Id)
		}
		cost := 0
		for i := 0; i < int(tutorial.RosterSize); i++ {
			cost += prices[i]
		}
		if cost > int(tutorial.Budget) {
			return fmt.Errorf("config: tutorial %s cannot form a roster within budget", tutorial.Id)
		}
	}
	return nil
}

// GetFirstItem returns the first item (for debug logging)
func GetFirstItem() *item {
	if Global == nil || Global.Tbitem == nil {
		return nil
	}
	list := Global.Tbitem.GetDataList()
	if len(list) == 0 {
		return nil
	}
	return list[0]
}
