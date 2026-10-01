package matchengine_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"

	"windypath.com/cs2match/server/internal/framework/matchengine"
)

// These tests deliberately use only the API available to an importing package.
func TestServicePublicContractMatchesRecordedReports(t *testing.T) {
	body, err := os.ReadFile("testdata/service_results.sha256.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]string
	if err := json.Unmarshal(body, &expected); err != nil {
		t.Fatal(err)
	}
	for _, seed := range []int64{11, 42, 20261001} {
		name := strconv.FormatInt(seed, 10)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			input := loadServiceInput(t)
			input.Seed = seed
			before := marshalContractJSON(t, input)
			result, err := matchengine.NewService(nil).Simulate(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(marshalContractJSON(t, result))
			if got := hex.EncodeToString(digest[:]); got != expected[name] {
				t.Fatalf("complete report changed for seed %d: got %s, want %s", seed, got, expected[name])
			}
			if !bytes.Equal(before, marshalContractJSON(t, input)) {
				t.Fatal("Simulate mutated the caller's input snapshot")
			}
			if result.TotalRounds != len(result.Rounds) || result.TotalRounds < 13 || result.FinalStats == nil || len(result.FinalStats.PlayerStats) != 10 {
				t.Fatalf("incomplete match result: rounds=%d", result.TotalRounds)
			}
			if result.WinnerTeamID != input.TeamA.TeamID && result.WinnerTeamID != input.TeamB.TeamID {
				t.Fatalf("winner %q is not one of the input teams", result.WinnerTeamID)
			}
		})
	}
}

func TestServicePublicContractRejectsNilInput(t *testing.T) {
	result, err := matchengine.NewService(nil).Simulate(context.Background(), nil)
	var engineError *matchengine.EngineError
	if result != nil || !errors.As(err, &engineError) || engineError.Code != "INVALID_MATCH_INPUT" {
		t.Fatalf("expected structured input error without a result, got result=%v err=%v", result, err)
	}
}

func TestServicePublicContractHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := matchengine.NewService(nil).Simulate(ctx, loadServiceInput(t))
	if result != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation without a partial result, got result=%v err=%v", result, err)
	}
}

func loadServiceInput(t *testing.T) *matchengine.MatchInput {
	t.Helper()
	body, err := os.ReadFile("testdata/service_input.json")
	if err != nil {
		t.Fatal(err)
	}
	var input matchengine.MatchInput
	if err := json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	return &input
}

func marshalContractJSON(t *testing.T, value any) []byte {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
