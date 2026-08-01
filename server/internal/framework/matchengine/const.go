package matchengine

const (
	DefaultMapID      = "de_dust2"
	DefaultMapName    = "Dust II"
	DefaultMapVersion = "draft-dust2-semantic-v1"

	WeaponAK47  = "AK47"
	WeaponM4A1S = "M4A1S"
)

func DefaultMR12RuleSet(constants CombatConstants) RuleSet {
	return RuleSet{
		RuleSetID:             "cs2_mr12_ot_mr3_v1",
		RegulationHalfRounds:  12,
		RegulationWinRounds:   13,
		RegulationMaxRounds:   24,
		OvertimeEnabled:       true,
		OvertimeHalfRounds:    3,
		OvertimeBlockRounds:   6,
		RoundTimeLimit:        constants.Int("RoundTimeLimit", 115),
		BombExplodeTime:       constants.Int("BombExplodeTime", 40),
		BasePlantTime:         constants.Int("BasePlantTime", 4),
		BaseDefuseTime:        constants.Int("BaseDefuseTime", 10),
		BasePickupTime:        constants.Int("BasePickupTime", 2),
		ForceExecuteThreshold: constants.Int("ForceExecuteThreshold", 20),
		MaxDecisionCount:      constants.Int("MaxDecisionCount", 3),
		MaxEncounterPulses:    constants.Int("MaxEncounterPulses", 3),
	}
}

func IsSupportedMap(mapID string) bool {
	return mapID == DefaultMapID
}
