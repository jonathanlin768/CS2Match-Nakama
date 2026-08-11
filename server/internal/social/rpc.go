package social

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

type rpcFunc func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error)

func decode(payload string, target interface{}) error {
	d := json.NewDecoder(strings.NewReader(payload))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
func caller(ctx context.Context) (string, string, error) {
	id, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	name, _ := ctx.Value(runtime.RUNTIME_CTX_USERNAME).(string)
	if id == "" {
		return "", "", runtime.NewError("missing user id", 16)
	}
	return id, name, nil
}
func respond(value interface{}, err error) (string, error) {
	if err != nil {
		if typed, ok := err.(*SocialError); ok {
			b, _ := json.Marshal(map[string]string{"code": typed.Code, "message": typed.Message})
			return string(b), runtime.NewError(typed.Message, 3)
		}
		return "", err
	}
	b, marshalErr := json.Marshal(value)
	return string(b), marshalErr
}

func RPCSetContactProfile(service *Service) rpcFunc {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		var req SetContactProfileRequest
		if err := decode(payload, &req); err != nil {
			return respond(nil, &SocialError{Code: "INVALID_REQUEST", Message: err.Error()})
		}
		id, _, err := caller(ctx)
		if err != nil {
			return "", err
		}
		value, err := service.SetProfile(ctx, id, req)
		return respond(value, err)
	}
}
func RPCGetContactExchange(service *Service) rpcFunc {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		var req FriendRequest
		if err := decode(payload, &req); err != nil {
			return respond(nil, &SocialError{Code: "INVALID_REQUEST", Message: err.Error()})
		}
		id, _, err := caller(ctx)
		if err != nil {
			return "", err
		}
		value, err := service.GetExchange(ctx, id, req.FriendID)
		return respond(value, err)
	}
}
func RPCRequestContactExchange(service *Service) rpcFunc {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		var req RequestExchangeRequest
		if err := decode(payload, &req); err != nil {
			return respond(nil, &SocialError{Code: "INVALID_REQUEST", Message: err.Error()})
		}
		id, name, err := caller(ctx)
		if err != nil {
			return "", err
		}
		value, err := service.RequestExchange(ctx, id, name, req)
		return respond(value, err)
	}
}
func RPCRespondContactExchange(service *Service) rpcFunc {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		var req RespondExchangeRequest
		if err := decode(payload, &req); err != nil {
			return respond(nil, &SocialError{Code: "INVALID_REQUEST", Message: err.Error()})
		}
		id, name, err := caller(ctx)
		if err != nil {
			return "", err
		}
		value, err := service.RespondExchange(ctx, id, name, req)
		return respond(value, err)
	}
}
