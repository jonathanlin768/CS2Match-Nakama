package match

import (
	"github.com/heroiclabs/nakama-common/runtime"
	"windypath.com/cs2match/server/internal/framework/rpcregistry"
)

// RegisterRPCs binds this module's handlers to the initialized match service.
func RegisterRPCs(initializer runtime.Initializer, logger runtime.Logger, service *Service) error {
	return rpcregistry.Register(initializer, logger, []rpcregistry.Entry{
		{Name: "DebugSimuMatch", Handler: RPCDebugSimuMatch(service)},
		{Name: "SimuMatch", Handler: RPCSimuMatch(service)},
	})
}
