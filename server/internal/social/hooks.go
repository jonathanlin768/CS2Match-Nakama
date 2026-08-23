package social

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

func rejectClientChat(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ *rtapi.Envelope) (*rtapi.Envelope, error) {
	return nil, runtime.NewError("direct messages only support server-generated contact exchange cards", 7)
}

func beforeDeleteFriends(service *Service) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, *api.DeleteFriendsRequest) (*api.DeleteFriendsRequest, error) {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.DeleteFriendsRequest) (*api.DeleteFriendsRequest, error) {
		id, _, err := caller(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, runtime.NewError("missing delete friends request", 3)
		}
		if err := service.RevokeFriends(ctx, id, in.Ids, in.Usernames); err != nil {
			if typed, ok := err.(*SocialError); ok {
				return nil, runtime.NewError(typed.Code+": "+typed.Message, 3)
			}
			return nil, err
		}
		return in, nil
	}
}

func Register(initializer runtime.Initializer, service *Service, exchangeEnabled bool) error {
	if exchangeEnabled {
		for name, handler := range map[string]rpcFunc{
			"SocialSetContactProfile":        RPCSetContactProfile(service),
			"SocialGetContactProfile":        RPCGetContactProfile(service),
			"SocialGetContactExchange":       RPCGetContactExchange(service),
			"SocialRequestContactExchange":   RPCRequestContactExchange(service),
			"SocialRespondContactExchange":   RPCRespondContactExchange(service),
			"SocialListContactExchangeInbox": RPCListContactExchangeInbox(service),
		} {
			if err := initializer.RegisterRpc(name, handler); err != nil {
				return err
			}
		}
	}
	if err := initializer.RegisterBeforeDeleteFriends(beforeDeleteFriends(service)); err != nil {
		return err
	}
	for _, name := range []string{"ChannelMessageSend", "ChannelMessageUpdate", "ChannelMessageRemove"} {
		if err := initializer.RegisterBeforeRt(name, rejectClientChat); err != nil {
			return err
		}
	}
	return nil
}
