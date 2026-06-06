package cfg

import (
	"embed"
	"encoding/json"
	"fmt"
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
	return nil
}

// TableCount returns the number of loaded tables
func TableCount() int {
	count := 0
	if Global != nil {
		if Global.Tbitem != nil {
			count++
		}
	}
	return count
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
