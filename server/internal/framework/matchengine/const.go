package matchengine

import (
	"math/rand"
)

const (
	// RoundTimeLimit 回合时间上限（秒）。
	RoundTimeLimit = 115

	// BombPlantTime 下包时间（秒），MVP 阶段简化为立即下包。
	BombPlantTime = 5

	// BombExplodeTime 炸弹爆炸时间（秒）。
	BombExplodeTime = 40

	// SideT 进攻方阵营。
	SideT = "T"

	// SideCT 防守方阵营。
	SideCT = "CT"

	// WeaponT T 方默认主战武器。
	WeaponT = "AK-47"

	// WeaponCT CT 方默认主战武器。
	WeaponCT = "M4A4"
)

// dust2AttackRoutes 定义 Dust2 的 6 条进攻路线。
// MVP 阶段先硬编码，后续迁移到 Luban TbRoute 配表。
var dust2AttackRoutes = []*RouteConfig{
	{ID: "A_Long", Name: "A大", TargetSite: "A", BaseTime: 20, MinPlayers: 1, MaxPlayers: 3},
	{ID: "A_Short", Name: "A小", TargetSite: "A", BaseTime: 15, MinPlayers: 1, MaxPlayers: 2},
	{ID: "B_Tunnel", Name: "B洞", TargetSite: "B", BaseTime: 15, MinPlayers: 1, MaxPlayers: 3},
	{ID: "B_Door", Name: "B门", TargetSite: "B", BaseTime: 10, MinPlayers: 1, MaxPlayers: 2},
	{ID: "Rush_B", Name: "RushB", TargetSite: "B", BaseTime: 10, MinPlayers: 2, MaxPlayers: 5},
	{ID: "Mid_SplitB", Name: "中夹B", TargetSite: "B", BaseTime: 20, MinPlayers: 2, MaxPlayers: 4},
}

// routePositions 为每条进攻路线预定义一个地图坐标（比例 0.0 ~ 1.0）。
// 击杀位置会以此为基础加随机偏移，模拟不同阵亡点。
var routePositions = map[string]*Location{
	"A_Long":     {Name: "A大", X: 0.22, Y: 0.24},
	"A_Short":    {Name: "A小", X: 0.38, Y: 0.32},
	"B_Tunnel":   {Name: "B洞", X: 0.72, Y: 0.78},
	"B_Door":     {Name: "B门", X: 0.80, Y: 0.62},
	"Rush_B":     {Name: "RushB", X: 0.78, Y: 0.82},
	"Mid_SplitB": {Name: "中夹B", X: 0.55, Y: 0.68},
}

// RandomRouteLocation 返回带轻微随机偏移的路线位置，避免所有击杀点完全重叠。
func RandomRouteLocation(routeID string, rng *rand.Rand) *Location {
	base := routePositions[routeID]
	if base == nil {
		base = &Location{Name: routeID, X: 0.5, Y: 0.5}
	}
	jitter := func(v float64) float64 {
		offset := rng.Float64()*0.08 - 0.04 // ±4%
		v += offset
		if v < 0.02 {
			v = 0.02
		}
		if v > 0.98 {
			v = 0.98
		}
		return v
	}
	return &Location{
		Name: base.Name,
		X:    jitter(base.X),
		Y:    jitter(base.Y),
	}
}

// DefaultMapName 返回地图显示名称。
func DefaultMapName(mapID string) string {
	if mapID == "de_dust2" {
		return "Dust II"
	}
	return mapID
}

// IsSupportedMap 检查当前 MVP 是否支持该地图。
func IsSupportedMap(mapID string) bool {
	return mapID == "de_dust2"
}

// GetRouteByID 根据路线 ID 获取配置。
func GetRouteByID(id string) *RouteConfig {
	for _, r := range dust2AttackRoutes {
		if r.ID == id {
			return r
		}
	}
	return nil
}
