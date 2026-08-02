package matchengine

import "testing"

func TestValidateMapConfigTemplateErrorMatrix(t *testing.T) {
	tests := []struct {
		name string
		edit func(*MapConfig)
		code string
	}{
		{
			name: "invalid side",
			edit: func(cfg *MapConfig) {
				tpl := cfg.RouteTemplates["TPL_A"]
				tpl.Side = "Both"
				cfg.RouteTemplates["TPL_A"] = tpl
			},
			code: "CONFIG_BAD_TEMPLATE_SCENARIO",
		},
		{
			name: "missing route reference",
			edit: func(cfg *MapConfig) {
				tpl := cfg.RouteTemplates["TPL_A"]
				tpl.RouteIDs = []string{"MISSING_ROUTE"}
				tpl.RouteAllocations = map[string]int{"MISSING_ROUTE": 5}
				cfg.RouteTemplates["TPL_A"] = tpl
			},
			code: "CONFIG_BAD_TEMPLATE_SCENARIO",
		},
		{
			name: "allocation does not close",
			edit: func(cfg *MapConfig) {
				tpl := cfg.RouteTemplates["TPL_A"]
				tpl.RouteAllocations = map[string]int{"D2_A_LONG": 4}
				cfg.RouteTemplates["TPL_A"] = tpl
			},
			code: "CONFIG_BAD_TEMPLATE_LIMIT",
		},
		{
			name: "wrong-side default",
			edit: func(cfg *MapConfig) {
				value := cfg.CombatConstants.Values["DefaultStrategyTemplateID"]
				value.Value = "TPL_CT"
				cfg.CombatConstants.Values["DefaultStrategyTemplateID"] = value
			},
			code: "CONFIG_BAD_TEMPLATE_SCENARIO",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := makeTestMapConfig()
			test.edit(cfg)
			assertEngineErrorCode(t, ValidateMapConfig(cfg), test.code)
		})
	}
}

func TestValidateMapConfigScenarioWeightAndReferenceMatrix(t *testing.T) {
	tests := []struct {
		name string
		edit func(*MapConfig)
		code string
	}{
		{
			name: "invalid scenario tag",
			edit: func(cfg *MapConfig) {
				scenario := cfg.Scenarios["SCN_A"]
				scenario.Phase = "Unknown"
				cfg.Scenarios["SCN_A"] = scenario
			},
			code: "CONFIG_BAD_SCENARIO_TAG",
		},
		{
			name: "missing map tag reference",
			edit: func(cfg *MapConfig) {
				scenario := cfg.Scenarios["SCN_A"]
				scenario.MapTagIDs = []string{"MISSING_TAG"}
				cfg.Scenarios["SCN_A"] = scenario
			},
			code: "CONFIG_BAD_SCENARIO_MAP_TAG",
		},
		{
			name: "missing attribute weight",
			edit: func(cfg *MapConfig) { delete(cfg.EncounterModifiers, "WEIGHT_DISCIPLINE") },
			code: "CONFIG_BAD_SCENARIO_WEIGHT",
		},
		{
			name: "weight total not normalized",
			edit: func(cfg *MapConfig) {
				modifier := cfg.EncounterModifiers["WEIGHT_AIM"]
				modifier.Weight = 11
				cfg.EncounterModifiers["WEIGHT_AIM"] = modifier
			},
			code: "CONFIG_BAD_SCENARIO_WEIGHT",
		},
		{
			name: "empty modifier reason code",
			edit: func(cfg *MapConfig) {
				modifier := cfg.EncounterModifiers["MOD_A"]
				modifier.ReasonCode = ""
				cfg.EncounterModifiers["MOD_A"] = modifier
			},
			code: "CONFIG_BAD_REASON_CODE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := makeTestMapConfig()
			test.edit(cfg)
			assertEngineErrorCode(t, ValidateMapConfig(cfg), test.code)
		})
	}
}

func TestValidateMapConfigCombatConstMatrix(t *testing.T) {
	tests := []struct {
		name string
		edit func(*MapConfig)
		code string
	}{
		{
			name: "bad value type",
			edit: func(cfg *MapConfig) {
				value := cfg.CombatConstants.Values["RoundTimeLimit"]
				value.Value = "not-an-int"
				cfg.CombatConstants.Values["RoundTimeLimit"] = value
			},
			code: "CONFIG_BAD_COMBAT_CONST_TYPE",
		},
		{
			name: "out of configured range",
			edit: func(cfg *MapConfig) {
				value := cfg.CombatConstants.Values["RoundTimeLimit"]
				value.MinValue = "120"
				cfg.CombatConstants.Values["RoundTimeLimit"] = value
			},
			code: "CONFIG_BAD_COMBAT_CONST_RANGE",
		},
		{
			name: "unsupported communication delay",
			edit: func(cfg *MapConfig) {
				value := cfg.CombatConstants.Values["CommunicationDelay"]
				value.Value = "1"
				cfg.CombatConstants.Values["CommunicationDelay"] = value
			},
			code: "CONFIG_UNSUPPORTED_COMMUNICATION_DELAY",
		},
		{
			name: "bomb action below bound",
			edit: func(cfg *MapConfig) {
				value := cfg.CombatConstants.Values["BasePlantTime"]
				value.Value = "2"
				cfg.CombatConstants.Values["BasePlantTime"] = value
			},
			code: "CONFIG_BAD_BOMB_CONST",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := makeTestMapConfig()
			test.edit(cfg)
			assertEngineErrorCode(t, ValidateMapConfig(cfg), test.code)
		})
	}
}

func TestValidateMapConfigKillSampleFallbackIsWarning(t *testing.T) {
	cfg := makeTestMapConfig()
	node := cfg.Nodes["LONG_DOOR"]
	node.AreaUsages = []string{"KillSample"}
	cfg.Nodes["LONG_DOOR"] = node
	if err := ValidateMapConfig(cfg); err != nil {
		t.Fatalf("missing KillSample geometry should degrade to x/y: %v", err)
	}
	if len(cfg.Warnings) != 1 || cfg.Warnings[0].Code != "CONFIG_MISSING_KILL_SAMPLE" {
		t.Fatalf("expected CONFIG_MISSING_KILL_SAMPLE warning, got %+v", cfg.Warnings)
	}
}

func TestValidateMapConfigFormalDust2Coverage(t *testing.T) {
	cfg := makeTestMapConfig()
	cfg.Version = DefaultMapVersion
	assertEngineErrorCode(t, ValidateMapConfig(cfg), "CONFIG_INCOMPLETE_DUST2_COVERAGE")
}

func TestValidateMapConfigStableStructuralErrorMatrix(t *testing.T) {
	tests := []struct {
		name string
		edit func(*MapConfig)
		code string
	}{
		{"duplicate template id", func(cfg *MapConfig) {
			tpl := cfg.RouteTemplates["TPL_A"]
			tpl.ID = "TPL_OTHER"
			cfg.RouteTemplates["TPL_A"] = tpl
		}, "CONFIG_DUP_ROUTE_TEMPLATE"},
		{"duplicate scenario id", func(cfg *MapConfig) {
			scenario := cfg.Scenarios["SCN_A"]
			scenario.ID = "SCN_OTHER"
			cfg.Scenarios["SCN_A"] = scenario
		}, "CONFIG_DUP_SCENARIO"},
		{"duplicate map tag id", func(cfg *MapConfig) {
			tag := cfg.MapTags["TAG_A"]
			tag.ID = "TAG_OTHER"
			cfg.MapTags["TAG_A"] = tag
		}, "CONFIG_DUP_MAP_TAG"},
		{"bad map tag", func(cfg *MapConfig) {
			tag := cfg.MapTags["TAG_A"]
			tag.Category = "Unknown"
			cfg.MapTags["TAG_A"] = tag
		}, "CONFIG_BAD_MAP_TAG"},
		{"missing modifier scenario", func(cfg *MapConfig) {
			modifier := cfg.EncounterModifiers["MOD_A"]
			modifier.ScenarioID = "MISSING_SCENARIO"
			cfg.EncounterModifiers["MOD_A"] = modifier
		}, "CONFIG_BAD_ENCOUNTER_MODIFIER"},
		{"duplicate node id", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.ID = "OTHER_NODE"
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_DUP_NODE"},
		{"bad node enum", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.Floor = "Space"
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_BAD_NODE_ENUM"},
		{"bad node coordinate", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.X = 2
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_BAD_NODE_COORD"},
		{"bad node shape", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.Shape = "Triangle"
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_BAD_NODE_SHAPE"},
		{"bad circle", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.Radius = 0
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_BAD_NODE_CIRCLE"},
		{"bad polygon", func(cfg *MapConfig) {
			node := cfg.Nodes["A_LONG"]
			node.Shape = "Polygon"
			node.Points = "0,0;1,1"
			cfg.Nodes["A_LONG"] = node
		}, "CONFIG_BAD_NODE_POLYGON"},
		{"bad edge node", func(cfg *MapConfig) {
			edge := cfg.Edges["E1"]
			edge.FromNode = "MISSING_NODE"
			cfg.Edges["E1"] = edge
		}, "CONFIG_BAD_EDGE_NODE"},
		{"bad risk point", func(cfg *MapConfig) {
			edge := cfg.Edges["E1"]
			edge.RiskPoints = []string{"MISSING_NODE"}
			cfg.Edges["E1"] = edge
		}, "CONFIG_BAD_RISK_POINT"},
		{"bad intercept node", func(cfg *MapConfig) {
			edge := cfg.Edges["E1"]
			edge.InterceptNodes = []string{"MISSING_NODE"}
			cfg.Edges["E1"] = edge
		}, "CONFIG_BAD_INTERCEPT_NODE"},
		{"bad visibility node", func(cfg *MapConfig) {
			cfg.Visibility["VIS"] = Visibility{ID: "VIS", FromNode: "MISSING_NODE", ToNode: "A_LONG", Visible: true}
		}, "CONFIG_BAD_VISIBILITY_NODE"},
		{"bad route node", func(cfg *MapConfig) {
			route := cfg.Routes["D2_A_LONG"]
			route.Nodes = append(route.Nodes, "MISSING_NODE")
			cfg.Routes["D2_A_LONG"] = route
		}, "CONFIG_BAD_ROUTE_NODE"},
		{"disconnected route", func(cfg *MapConfig) {
			route := cfg.Routes["D2_A_LONG"]
			route.Nodes = append(route.Nodes, "B_SITE")
			cfg.Routes["D2_A_LONG"] = route
		}, "CONFIG_ROUTE_NOT_CONNECTED"},
		{"bad route limit", func(cfg *MapConfig) {
			route := cfg.Routes["D2_A_LONG"]
			route.MinPlayers = 6
			cfg.Routes["D2_A_LONG"] = route
		}, "CONFIG_BAD_ROUTE_LIMIT"},
		{"missing plant site", func(cfg *MapConfig) {
			node := cfg.Nodes["A_SITE"]
			node.AreaUsages = []string{"KillSample"}
			cfg.Nodes["A_SITE"] = node
		}, "CONFIG_NO_PLANT_SITE"},
		{"duplicate combat const key", func(cfg *MapConfig) {
			value := cfg.CombatConstants.Values["RoundTimeLimit"]
			value.Key = "OtherKey"
			cfg.CombatConstants.Values["RoundTimeLimit"] = value
		}, "CONFIG_DUP_COMBAT_CONST"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := makeTestMapConfig()
			test.edit(cfg)
			assertEngineErrorCode(t, ValidateMapConfig(cfg), test.code)
		})
	}
}

func assertEngineErrorCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s, got nil", want)
	}
	engineErr, ok := err.(*EngineError)
	if !ok || engineErr.Code != want {
		t.Fatalf("expected %s, got %v", want, err)
	}
}
