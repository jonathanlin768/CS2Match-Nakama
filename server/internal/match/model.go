// Package match 是匹配对战业务子系统。
// 它负责把客户端的 RPC 请求转换为 matchengine 的输入，并返回标准化战报。
package match

import (
	"windypath.com/cs2match/server/internal/framework/matchengine"
)

// DebugSimuMatchRequest 是测试 RPC 请求体。
type DebugSimuMatchRequest struct {
	MapID string `json:"map_id"`
}

// DebugSimuMatchResponse 是测试 RPC 响应体。
type DebugSimuMatchResponse struct {
	MatchInfo  *matchengine.MatchInfo     `json:"match_info"`
	Rounds     []*matchengine.RoundResult `json:"rounds"`
	FinalStats *matchengine.FinalStats    `json:"final_stats"`
	Winner     string                     `json:"winner"`
}

// MatchError 是统一的错误响应结构。
type MatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MatchPlayerRef 表示阵容中的一个选手引用。
type MatchPlayerRef struct {
	PlayerID   string `json:"player_id"`
	InstanceID string `json:"instance_id"`
}
