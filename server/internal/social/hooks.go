package social

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

func rejectClientChat(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ *rtapi.Envelope) (*rtapi.Envelope, error) {
	return nil, runtime.NewError("direct messages only support server-generated contact exchange cards", 7)
}

func Register(initializer runtime.Initializer, service *Service, exchangeEnabled bool) error {
	if exchangeEnabled {
		for name, handler := range map[string]rpcFunc{
			"SocialSetContactProfile":      RPCSetContactProfile(service),
			"SocialGetContactExchange":     RPCGetContactExchange(service),
			"SocialRequestContactExchange": RPCRequestContactExchange(service),
			"SocialRespondContactExchange": RPCRespondContactExchange(service),
		} {
			if err := initializer.RegisterRpc(name, handler); err != nil {
				return err
			}
		}
	}
	for _, name := range []string{"ChannelMessageSend", "ChannelMessageUpdate", "ChannelMessageRemove"} {
		if err := initializer.RegisterBeforeRt(name, rejectClientChat); err != nil {
			return err
		}
	}
	return nil
}
