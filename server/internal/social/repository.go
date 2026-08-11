package social

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

type Repository struct{ nk runtime.NakamaModule }

func NewRepository(nk runtime.NakamaModule) *Repository { return &Repository{nk: nk} }

func pairKey(a, b string) string {
	ids := []string{strings.TrimSpace(a), strings.TrimSpace(b)}
	sort.Strings(ids)
	return ids[0] + "_" + ids[1]
}

func (r *Repository) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	state := 0
	friends, _, err := r.nk.FriendsList(ctx, userID, 100, &state, "")
	if err != nil {
		return false, err
	}
	for _, friend := range friends {
		if friend != nil && friend.User != nil && friend.User.Id == friendID && friend.State != nil && friend.State.Value == 0 {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repository) ReadProfile(ctx context.Context, userID string) (ContactProfile, string, error) {
	objects, err := r.nk.StorageRead(ctx, []*runtime.StorageRead{{Collection: profileCollection, Key: profileKey, UserID: userID}})
	if err != nil {
		return ContactProfile{}, "", err
	}
	if len(objects) == 0 {
		return ContactProfile{}, "", nil
	}
	var profile ContactProfile
	if err := json.Unmarshal([]byte(objects[0].Value), &profile); err != nil {
		return ContactProfile{}, "", err
	}
	return profile, objects[0].Version, nil
}

func (r *Repository) WriteProfile(ctx context.Context, userID string, profile ContactProfile, version string) error {
	value, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	if version == "" {
		version = "*"
	}
	_, err = r.nk.StorageWrite(ctx, []*runtime.StorageWrite{{Collection: profileCollection, Key: profileKey, UserID: userID, Value: string(value), Version: version, PermissionRead: runtime.STORAGE_PERMISSION_NO_READ, PermissionWrite: runtime.STORAGE_PERMISSION_NO_WRITE}})
	return err
}

func (r *Repository) ReadExchange(ctx context.Context, a, b string) (ExchangeRequest, string, error) {
	objects, err := r.nk.StorageRead(ctx, []*runtime.StorageRead{{Collection: exchangeCollection, Key: pairKey(a, b), UserID: systemOwnerID}})
	if err != nil {
		return ExchangeRequest{}, "", err
	}
	if len(objects) == 0 {
		return ExchangeRequest{}, "", nil
	}
	var exchange ExchangeRequest
	if err := json.Unmarshal([]byte(objects[0].Value), &exchange); err != nil {
		return ExchangeRequest{}, "", err
	}
	return exchange, objects[0].Version, nil
}

func (r *Repository) WriteExchange(ctx context.Context, a, b string, exchange ExchangeRequest, version string) error {
	value, err := json.Marshal(exchange)
	if err != nil {
		return err
	}
	if version == "" {
		version = "*"
	}
	_, err = r.nk.StorageWrite(ctx, []*runtime.StorageWrite{{Collection: exchangeCollection, Key: pairKey(a, b), UserID: systemOwnerID, Value: string(value), Version: version, PermissionRead: runtime.STORAGE_PERMISSION_NO_READ, PermissionWrite: runtime.STORAGE_PERMISSION_NO_WRITE}})
	return err
}
