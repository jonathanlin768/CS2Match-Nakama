// Package rpcregistry registers explicit RPC lists with consistent diagnostics.
package rpcregistry

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Handler is Nakama's RPC callback signature.
type Handler = func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error)

type Entry struct {
	Name    string
	Handler Handler
}

// Registrar is the registration capability used from runtime.Initializer.
type Registrar interface {
	RegisterRpc(string, Handler) error
}

// Register preserves list order and stops at the first registration failure.
func Register(registrar Registrar, logger runtime.Logger, entries []Entry) error {
	for _, entry := range entries {
		if err := registrar.RegisterRpc(entry.Name, entry.Handler); err != nil {
			return fmt.Errorf("register RPC %q: %w", entry.Name, err)
		}
		if logger != nil {
			logger.Info("%s RPC registered", entry.Name)
		}
	}
	return nil
}
