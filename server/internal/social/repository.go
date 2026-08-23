package social

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type repositoryNakama interface {
	FriendsList(ctx context.Context, userID string, limit int, state *int, cursor string) ([]*api.Friend, string, error)
	StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error)
	StorageWrite(ctx context.Context, writes []*runtime.StorageWrite) ([]*api.StorageObjectAck, error)
}

type Repository struct{ nk repositoryNakama }

func NewRepository(nk repositoryNakama) *Repository { return &Repository{nk: nk} }

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
	profile = normalizeLegacyProfile(profile)
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

func (r *Repository) ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]*api.Friend, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	state := 0
	return r.nk.FriendsList(ctx, userID, limit, &state, cursor)
}

func (r *Repository) ReadExchanges(ctx context.Context, userID string, friendIDs []string) (map[string]ExchangeRequest, error) {
	result := make(map[string]ExchangeRequest, len(friendIDs))
	if len(friendIDs) == 0 {
		return result, nil
	}
	reads := make([]*runtime.StorageRead, 0, len(friendIDs))
	friendByKey := make(map[string]string, len(friendIDs))
	for _, friendID := range friendIDs {
		key := pairKey(userID, friendID)
		reads = append(reads, &runtime.StorageRead{Collection: exchangeCollection, Key: key, UserID: systemOwnerID})
		friendByKey[key] = friendID
	}
	objects, err := r.nk.StorageRead(ctx, reads)
	if err != nil {
		return nil, err
	}
	for _, object := range objects {
		friendID, ok := friendByKey[object.Key]
		if !ok {
			continue
		}
		var exchange ExchangeRequest
		if err := json.Unmarshal([]byte(object.Value), &exchange); err != nil {
			return nil, err
		}
		result[friendID] = exchange
	}
	return result, nil
}

func (r *Repository) RevokeExchange(ctx context.Context, userID, friendID string, revokedAt int64) error {
	exchange, version, err := r.ReadExchange(ctx, userID, friendID)
	if err != nil || exchange.RequestID == "" {
		return err
	}
	exchange.Status = "revoked"
	exchange.Version++
	exchange.RespondedAt = revokedAt
	exchange.RequesterProfileRevision = 0
	exchange.AcceptedRequesterProfileRevision = 0
	exchange.AcceptedRecipientProfileRevision = 0
	return r.WriteExchange(ctx, userID, friendID, exchange, version)
}
