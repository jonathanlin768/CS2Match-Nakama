package match

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"windypath.com/cs2match/server/internal/framework/rpcregistry"
)

func TestSimuMatchRejectsInvalidModeBeforeSimulation(t *testing.T) {
	// A missing engine dependency makes unintended simulation fail this test.
	service := &Service{}
	for _, mode := range []MatchMode{"", "unknown", "Computer"} {
		t.Run(string(mode), func(t *testing.T) {
			result, err := service.SimuMatch(context.Background(), "user", SimuMatchRequest{Mode: mode})
			var typed *MatchError
			if result != nil || !errors.As(err, &typed) || typed.Code != "INVALID_MODE" {
				t.Fatalf("expected INVALID_MODE without a result, got result=%v err=%v", result, err)
			}
		})
	}
}

func TestSimuMatchModeJSONContract(t *testing.T) {
	for _, mode := range []MatchMode{MatchModeComputer, MatchModeTutorial, "unknown"} {
		t.Run(string(mode), func(t *testing.T) {
			body := `{"mode":"` + string(mode) + `"}`
			var req SimuMatchRequest
			if err := decodeStrictRequest(body, &req); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(req)
			if err != nil || req.Mode != mode || string(encoded) != body {
				t.Fatalf("string mode contract changed: mode=%q encoded=%s err=%v", req.Mode, encoded, err)
			}
		})
	}
}

func TestMatchRPCsRequireUserIdentity(t *testing.T) {
	service := &Service{}
	for _, tc := range []struct {
		name    string
		handler rpcregistry.Handler
		payload string
	}{
		{name: "DebugSimuMatch", handler: RPCDebugSimuMatch(service), payload: `{"map_id":"de_dust2","seed":42}`},
		{name: "SimuMatch", handler: RPCSimuMatch(service), payload: `{"mode":"computer"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := tc.handler(context.Background(), nil, nil, nil, tc.payload)
			var response MatchError
			if jsonErr := json.Unmarshal([]byte(body), &response); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			if err == nil || response.Code != "UNAUTHORIZED" {
				t.Fatalf("RPC did not reject missing identity: body=%s err=%v", body, err)
			}
		})
	}
}
