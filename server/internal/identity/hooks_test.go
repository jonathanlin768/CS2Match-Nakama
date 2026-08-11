package identity

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGenerateUsernameContract(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 128; i++ {
		code, err := GenerateUsername()
		if err != nil || !usernamePattern.MatchString(code) {
			t.Fatalf("invalid generated code %q: %v", code, err)
		}
		if seen[code] {
			t.Fatalf("unexpected duplicate in sample: %s", code)
		}
		seen[code] = true
	}
}

func TestAuthCreationNormalizesAndRejectsCustomNames(t *testing.T) {
	request := &api.AuthenticateDeviceRequest{Create: wrapperspb.Bool(true)}
	out, err := BeforeAuthenticateDevice(context.Background(), nil, nil, nil, request)
	if err != nil || !usernamePattern.MatchString(out.Username) {
		t.Fatalf("device username not generated: %#v %v", out, err)
	}
	bad := &api.AuthenticateEmailRequest{Create: wrapperspb.Bool(true), Username: "PlayerName"}
	if _, err := BeforeAuthenticateEmail(context.Background(), nil, nil, nil, bad); err == nil {
		t.Fatal("custom username should be rejected")
	}
}

func TestUpdateAccountRejectsIdentityChanges(t *testing.T) {
	if _, err := BeforeUpdateAccount(context.Background(), nil, nil, nil, &api.UpdateAccountRequest{DisplayName: wrapperspb.String("hello")}); err == nil {
		t.Fatal("display name update should be rejected")
	}
	if _, err := BeforeUpdateAccount(context.Background(), nil, nil, nil, &api.UpdateAccountRequest{AvatarUrl: wrapperspb.String("https://example.test/a.png")}); err != nil {
		t.Fatalf("unrelated profile field should be allowed: %v", err)
	}
}
