package identity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"regexp"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

var usernamePattern = regexp.MustCompile(`^[A-Z0-9]{8}$`)

// GenerateUsername creates the opaque account code shown as 玩家#CODE by clients.
func GenerateUsername() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = codeAlphabet[int(bytes[i])%len(codeAlphabet)]
	}
	return string(bytes), nil
}

func ensureCreationUsername(create bool, username *string) error {
	if !create {
		return nil
	}
	if *username == "" {
		generated, err := GenerateUsername()
		if err != nil {
			return runtime.NewError("could not generate player code", 13)
		}
		*username = generated
	}
	if !usernamePattern.MatchString(*username) {
		return runtime.NewError("username must be an 8-character uppercase player code", 3)
	}
	return nil
}

func BeforeAuthenticateDevice(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateDeviceRequest) (*api.AuthenticateDeviceRequest, error) {
	if in == nil {
		return nil, runtime.NewError("request is required", 3)
	}
	create := in.Create != nil && in.Create.Value
	if err := ensureCreationUsername(create, &in.Username); err != nil {
		return nil, err
	}
	return in, nil
}

func BeforeAuthenticateEmail(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AuthenticateEmailRequest) (*api.AuthenticateEmailRequest, error) {
	if in == nil {
		return nil, runtime.NewError("request is required", 3)
	}
	create := in.Create != nil && in.Create.Value
	if err := ensureCreationUsername(create, &in.Username); err != nil {
		return nil, err
	}
	return in, nil
}

func BeforeUpdateAccount(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.UpdateAccountRequest) (*api.UpdateAccountRequest, error) {
	if in == nil {
		return nil, runtime.NewError("request is required", 3)
	}
	if in.Username != nil || in.DisplayName != nil {
		return nil, runtime.NewError("player code and display name cannot be changed", 7)
	}
	return in, nil
}

func Register(initializer runtime.Initializer) error {
	if err := initializer.RegisterBeforeAuthenticateDevice(BeforeAuthenticateDevice); err != nil {
		return err
	}
	if err := initializer.RegisterBeforeAuthenticateEmail(BeforeAuthenticateEmail); err != nil {
		return err
	}
	return initializer.RegisterBeforeUpdateAccount(BeforeUpdateAccount)
}
