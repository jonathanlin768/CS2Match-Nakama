package rpcregistry

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

type recordingRegistrar struct {
	names   []string
	results []string
	failAt  string
	err     error
}

func echoHandler(prefix string) Handler {
	return func(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		return prefix + payload, nil
	}
}

func (r *recordingRegistrar) RegisterRpc(name string, handler Handler) error {
	r.names = append(r.names, name)
	if name == r.failAt {
		return r.err
	}
	result, err := handler(context.Background(), nil, nil, nil, name)
	if err != nil {
		return err
	}
	r.results = append(r.results, result)
	return nil
}

func TestRegisterPreservesOrderAndHandlerBindings(t *testing.T) {
	registrar := &recordingRegistrar{}
	entries := []Entry{
		{Name: "First", Handler: echoHandler("first:")},
		{Name: "Second", Handler: echoHandler("second:")},
	}
	if err := Register(registrar, nil, entries); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(registrar.names, []string{"First", "Second"}) ||
		!reflect.DeepEqual(registrar.results, []string{"first:First", "second:Second"}) {
		t.Fatalf("registration order or handler binding changed: names=%v results=%v", registrar.names, registrar.results)
	}
}

func TestRegisterStopsAndPreservesOriginalError(t *testing.T) {
	wantErr := errors.New("registration rejected")
	registrar := &recordingRegistrar{failAt: "Second", err: wantErr}
	err := Register(registrar, nil, []Entry{
		{Name: "First", Handler: echoHandler("")},
		{Name: "Second", Handler: echoHandler("")},
		{Name: "NeverRegistered", Handler: echoHandler("")},
	})
	if !errors.Is(err, wantErr) || !strings.Contains(err.Error(), "Second") {
		t.Fatalf("registration error lost its cause or RPC name: %v", err)
	}
	if !reflect.DeepEqual(registrar.names, []string{"First", "Second"}) {
		t.Fatalf("registration continued after failure: %v", registrar.names)
	}
}
