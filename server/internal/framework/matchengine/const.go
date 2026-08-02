package matchengine

const (
	DefaultMapID      = "de_dust2"
	DefaultMapName    = "Dust II"
	DefaultMapVersion = "draft-dust2-semantic-v1"

	WeaponAK47  = "AK47"
	WeaponM4A1S = "M4A1S"
)

func DefaultMR12RuleSet(_ CombatConstants) RuleSet {
	return RuleSet{
		RuleSetID:            "cs2_mr12_ot_mr3_v1",
		RegulationHalfRounds: 12,
		RegulationWinRounds:  13,
		RegulationMaxRounds:  24,
		OvertimeEnabled:      true,
		OvertimeHalfRounds:   3,
		OvertimeBlockRounds:  6,
	}
}

func IsSupportedMap(mapID string) bool {
	return mapID == DefaultMapID
}
