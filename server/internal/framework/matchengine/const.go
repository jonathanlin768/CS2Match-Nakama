package matchengine

const (
	SideT  = "T"
	SideCT = "CT"

	EventMatchStart       = "MATCH_START"
	EventRoundStart       = "ROUND_START"
	EventHalfTime         = "HALF_TIME"
	EventSideSwitch       = "SIDE_SWITCH"
	EventOvertime         = "OVERTIME_START"
	EventDamage           = "DAMAGE"
	EventKill             = "KILL"
	EventStrategyAdjusted = "STRATEGY_ADJUSTED"
	EventRotate           = "ROTATE"
	EventReinforce        = "REINFORCE"
	EventControlGained    = "CONTROL_GAINED"
	EventBombDrop         = "BOMB_DROP"
	EventBombPickup       = "BOMB_PICKUP"
	EventPlantStart       = "BOMB_PLANT_START"
	EventPlantInterrupt   = "BOMB_PLANT_INTERRUPT"
	EventDefuseStart      = "DEFUSE_START"
	EventDefuseInterrupt  = "DEFUSE_INTERRUPT"
	EventBombPlant        = "BOMB_PLANT"
	EventBombDefuse       = "BOMB_DEFUSE"
	EventBombExplode      = "BOMB_EXPLODE"
	EventRoundEnd         = "ROUND_END"
	EventMatchEnd         = "MATCH_END"
	BombStatusCarried     = "Carried"
	BombStatusPlanted     = "Planted"
	BombStatusDefused     = "Defused"
	BombStatusExplode     = "Exploded"
	BombStatusDropped     = "Dropped"
)

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
