package match

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

// RPCDebugSimuMatch 返回 DebugSimuMatch 的 RPC 处理函数。
func RPCDebugSimuMatch(service *Service) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		var req DebugSimuMatchRequest
		if payload != "" {
			if err := decodeDebugSimuMatchRequest(payload, &req); err != nil {
				return marshalError(&MatchError{Code: "INVALID_REQUEST", Message: err.Error()})
			}
		}

		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok || userID == "" {
			return marshalError(&MatchError{Code: "UNAUTHORIZED", Message: "missing user id"})
		}

		resp, err := service.DebugSimuMatch(ctx, userID, req)
		if err != nil {
			if me, ok := err.(*MatchError); ok {
				return marshalError(me)
			}
			return marshalError(&MatchError{Code: "SIMULATION_ERROR", Message: err.Error()})
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			logger.Error("DebugSimuMatch RPC failed to marshal response: %v", err)
			return "", err
		}
		return string(bytes), nil
	}
}

// RPCSimuMatch returns the production battle RPC handler.
func RPCSimuMatch(service *Service) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		var req SimuMatchRequest
		if payload == "" {
			return marshalError(&MatchError{Code: "INVALID_REQUEST", Message: "request body is required"})
		}
		if err := decodeStrictRequest(payload, &req); err != nil {
			return marshalError(&MatchError{Code: "INVALID_REQUEST", Message: err.Error()})
		}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok || userID == "" {
			return marshalError(&MatchError{Code: "UNAUTHORIZED", Message: "missing user id"})
		}
		resp, err := service.SimuMatch(ctx, userID, req)
		if err != nil {
			if typed, ok := err.(*MatchError); ok {
				return marshalError(typed)
			}
			return marshalError(&MatchError{Code: "SIMULATION_ERROR", Message: err.Error()})
		}
		bytes, err := json.Marshal(resp)
		if err != nil {
			logger.Error("SimuMatch RPC failed to marshal response: %v", err)
			return "", err
		}
		return string(bytes), nil
	}
}

func decodeDebugSimuMatchRequest(payload string, target *DebugSimuMatchRequest) error {
	return decodeStrictRequest(payload, target)
}

func decodeStrictRequest(payload string, target interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request contains multiple JSON values")
		}
		return err
	}
	return nil
}

func marshalError(me *MatchError) (string, error) {
	bytes, err := json.Marshal(me)
	if err != nil {
		return "", err
	}
	return string(bytes), fmt.Errorf("%s: %s", me.Code, me.Message)
}
