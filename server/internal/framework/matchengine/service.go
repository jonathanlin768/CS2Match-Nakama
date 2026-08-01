package matchengine

import (
	"context"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Service 是 matchengine 的入口工厂。
type Service struct {
	logger runtime.Logger
}

// NewService 创建引擎服务。
func NewService(logger runtime.Logger) *Service {
	return &Service{logger: logger}
}

// Simulate 接收输入并执行完整推演，返回 MatchResult。
func (s *Service) Simulate(ctx context.Context, input *MatchInput) (*MatchResult, error) {
	if input == nil {
		return nil, newError("INVALID_MATCH_INPUT", "input is nil")
	}
	engine := NewMatchEngine(input)
	return engine.StartMatch(ctx)
}
