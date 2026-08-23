package social

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

type fakeSocialRepo struct {
	profiles       map[string]ContactProfile
	exchanges      map[string]ExchangeRequest
	friends        map[string]map[string]bool
	usernames      map[string]string
	profileWrites  int
	exchangeWrites int
	revoked        []string
	failWrite      bool
	cursor         string
}

func newFakeSocialRepo() *fakeSocialRepo {
	return &fakeSocialRepo{
		profiles:  map[string]ContactProfile{},
		exchanges: map[string]ExchangeRequest{},
		friends:   map[string]map[string]bool{},
		usernames: map[string]string{},
	}
}

func (r *fakeSocialRepo) IsFriend(_ context.Context, userID, friendID string) (bool, error) {
	return r.friends[userID][friendID], nil
}
func (r *fakeSocialRepo) ReadProfile(_ context.Context, userID string) (ContactProfile, string, error) {
	return normalizeLegacyProfile(r.profiles[userID]), "profile-version", nil
}
func (r *fakeSocialRepo) WriteProfile(_ context.Context, userID string, profile ContactProfile, _ string) error {
	if r.failWrite {
		return errors.New("write failed")
	}
	r.profileWrites++
	r.profiles[userID] = profile
	return nil
}
func (r *fakeSocialRepo) ReadExchange(_ context.Context, a, b string) (ExchangeRequest, string, error) {
	return r.exchanges[pairKey(a, b)], "exchange-version", nil
}
func (r *fakeSocialRepo) WriteExchange(_ context.Context, a, b string, exchange ExchangeRequest, _ string) error {
	if r.failWrite {
		return errors.New("write failed")
	}
	r.exchangeWrites++
	r.exchanges[pairKey(a, b)] = exchange
	return nil
}
func (r *fakeSocialRepo) ListFriends(_ context.Context, userID string, limit int, _ string) ([]*api.Friend, string, error) {
	result := make([]*api.Friend, 0, len(r.friends[userID]))
	for friendID, active := range r.friends[userID] {
		if active {
			result = append(result, &api.Friend{User: &api.User{Id: friendID, Username: r.usernames[friendID]}})
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, r.cursor, nil
}
func (r *fakeSocialRepo) ReadExchanges(_ context.Context, userID string, friendIDs []string) (map[string]ExchangeRequest, error) {
	result := map[string]ExchangeRequest{}
	for _, friendID := range friendIDs {
		if exchange, ok := r.exchanges[pairKey(userID, friendID)]; ok {
			result[friendID] = exchange
		}
	}
	return result, nil
}
func (r *fakeSocialRepo) RevokeExchange(_ context.Context, userID, friendID string, revokedAt int64) error {
	if r.failWrite {
		return errors.New("write failed")
	}
	r.revoked = append(r.revoked, friendID)
	exchange := r.exchanges[pairKey(userID, friendID)]
	if exchange.RequestID != "" {
		exchange.Status = "revoked"
		exchange.Version++
		exchange.RespondedAt = revokedAt
		exchange.RequesterProfileRevision = 0
		exchange.AcceptedRequesterProfileRevision = 0
		exchange.AcceptedRecipientProfileRevision = 0
		r.exchanges[pairKey(userID, friendID)] = exchange
	}
	return nil
}

type fakeSocialNakama struct {
	accounts map[string]*api.Account
	users    map[string]*api.User
	cards    []map[string]interface{}
	cardErr  error
	writes   []*runtime.StorageWrite
}

func newFakeSocialNakama() *fakeSocialNakama {
	return &fakeSocialNakama{accounts: map[string]*api.Account{}, users: map[string]*api.User{}}
}
func (n *fakeSocialNakama) AccountGetId(_ context.Context, userID string) (*api.Account, error) {
	return n.accounts[userID], nil
}
func (n *fakeSocialNakama) UsersGetUsername(_ context.Context, usernames []string) ([]*api.User, error) {
	result := make([]*api.User, 0, len(usernames))
	for _, username := range usernames {
		if user := n.users[username]; user != nil {
			result = append(result, user)
		}
	}
	return result, nil
}
func (n *fakeSocialNakama) ChannelIdBuild(_ context.Context, sender, target string, _ runtime.ChannelType) (string, error) {
	if n.cardErr != nil {
		return "", n.cardErr
	}
	return pairKey(sender, target), nil
}
func (n *fakeSocialNakama) ChannelMessageSend(_ context.Context, _ string, content map[string]interface{}, _, _ string, _ bool) (*rtapi.ChannelMessageAck, error) {
	if n.cardErr != nil {
		return nil, n.cardErr
	}
	n.cards = append(n.cards, content)
	return &rtapi.ChannelMessageAck{}, nil
}
func (n *fakeSocialNakama) FriendsList(context.Context, string, int, *int, string) ([]*api.Friend, string, error) {
	return nil, "", nil
}
func (n *fakeSocialNakama) StorageRead(context.Context, []*runtime.StorageRead) ([]*api.StorageObject, error) {
	return nil, nil
}

func (n *fakeSocialNakama) StorageWrite(_ context.Context, writes []*runtime.StorageWrite) ([]*api.StorageObjectAck, error) {
	n.writes = append(n.writes, writes...)
	return nil, nil
}

func newConsentService(repo *fakeSocialRepo, nk *fakeSocialNakama) *Service {
	return &Service{repo: repo, nk: nk, now: func() time.Time { return time.Unix(1_700_000_000, 0) }}
}

func makeFriends(repo *fakeSocialRepo, a, b string) {
	if repo.friends[a] == nil {
		repo.friends[a] = map[string]bool{}
	}
	if repo.friends[b] == nil {
		repo.friends[b] = map[string]bool{}
	}
	repo.friends[a][b] = true
	repo.friends[b][a] = true
}

func TestProfileRevisionLifecycle(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	nk.accounts["a"] = &api.Account{Email: "a@example.com"}
	repo.profiles["a"] = ContactProfile{QQ: "12345678", WeChat: "Alpha_123"}
	service := newConsentService(repo, nk)

	view, err := service.GetProfile(context.Background(), "a")
	if err != nil || view.Revision != 1 || view.QQ != "12345678" {
		t.Fatalf("legacy profile was not normalized: %#v %v", view, err)
	}
	view, err = service.SetProfile(context.Background(), "a", SetContactProfileRequest{QQ: " 12345678 ", WeChat: "Alpha_123"})
	if err != nil || view.Revision != 1 || repo.profileWrites != 0 {
		t.Fatalf("no-op save changed revision: %#v writes=%d err=%v", view, repo.profileWrites, err)
	}
	view, err = service.SetProfile(context.Background(), "a", SetContactProfileRequest{QQ: "22345678", WeChat: "Alpha_123"})
	if err != nil || view.Revision != 2 {
		t.Fatalf("real change did not increment revision: %#v %v", view, err)
	}
	view, err = service.SetProfile(context.Background(), "a", SetContactProfileRequest{QQ: "22345678"})
	if err != nil || view.Revision != 3 || view.WeChat != "" {
		t.Fatalf("single channel clear failed: %#v %v", view, err)
	}
	view, err = service.SetProfile(context.Background(), "a", SetContactProfileRequest{})
	if err != nil || view.Revision != 4 || view.QQ != "" {
		t.Fatalf("full clear failed: %#v %v", view, err)
	}
}

func TestProfileAllowsEitherSingleContactChannel(t *testing.T) {
	for _, test := range []struct {
		name string
		req  SetContactProfileRequest
		want ContactProfile
	}{
		{name: "qq only", req: SetContactProfileRequest{QQ: "12345678"}, want: ContactProfile{QQ: "12345678"}},
		{name: "wechat only", req: SetContactProfileRequest{WeChat: "Alpha_123"}, want: ContactProfile{WeChat: "Alpha_123"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeSocialRepo()
			nk := newFakeSocialNakama()
			nk.accounts["a"] = &api.Account{Email: "a@example.com"}
			view, err := newConsentService(repo, nk).SetProfile(context.Background(), "a", test.req)
			if err != nil || view.QQ != test.want.QQ || view.WeChat != test.want.WeChat || view.Revision != 1 {
				t.Fatalf("single-channel profile was rejected: view=%#v err=%v", view, err)
			}
		})
	}
}

func TestProfileGetterIsBoundToCaller(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	nk.accounts["a"] = &api.Account{Email: "a@example.com"}
	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
	repo.profiles["b"] = ContactProfile{QQ: "87654321", Revision: 1}
	service := newConsentService(repo, nk)
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "a")
	ctx = context.WithValue(ctx, runtime.RUNTIME_CTX_USERNAME, "A")

	payload, err := RPCGetContactProfile(service)(ctx, nil, nil, nil, `{}`)
	if err != nil || !containsAll(payload, "12345678") || containsAll(payload, "87654321") {
		t.Fatalf("getter escaped caller identity: payload=%s err=%v", payload, err)
	}
	if _, err := RPCGetContactProfile(service)(ctx, nil, nil, nil, `{"user_id":"b"}`); err == nil {
		t.Fatal("getter accepted a target user id")
	}
}

func TestExchangeConsentBindsBothProfileRevisions(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	makeFriends(repo, "a", "b")
	repo.profiles["a"] = ContactProfile{QQ: "12345678", WeChat: "Alpha_123", Revision: 1}
	repo.profiles["b"] = ContactProfile{QQ: "87654321", WeChat: "Bravo_456", Revision: 1}
	service := newConsentService(repo, nk)

	requested, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"qq", "wechat"}})
	if err != nil || requested.RequesterProfileRevision != 1 {
		t.Fatalf("request did not bind requester revision: %#v %v", requested, err)
	}
	repeated, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"wechat", "qq"}})
	if err != nil || repeated.RequestID != requested.RequestID || repo.exchangeWrites != 1 {
		t.Fatalf("request was not idempotent: %#v writes=%d err=%v", repeated, repo.exchangeWrites, err)
	}
	accepted, err := service.RespondExchange(context.Background(), "b", "B", RespondExchangeRequest{FriendID: "a", RequestID: requested.RequestID, Accept: true})
	if err != nil || accepted.Status != "accepted" || accepted.FriendContact.QQ != "12345678" {
		t.Fatalf("valid request was not accepted bidirectionally: %#v %v", accepted, err)
	}
	if accepted.AcceptedRequesterProfileRevision != 1 || accepted.AcceptedRecipientProfileRevision != 1 {
		t.Fatalf("accepted revisions missing: %#v", accepted.ExchangeRequest)
	}
	if _, err := service.RespondExchange(context.Background(), "b", "B", RespondExchangeRequest{FriendID: "a", RequestID: requested.RequestID, Accept: false}); err == nil {
		t.Fatal("a second concurrent-style response was accepted")
	}

	repo.profiles["b"] = ContactProfile{QQ: "99999999", WeChat: "Bravo_456", Revision: 2}
	stale, err := service.GetExchange(context.Background(), "a", "b")
	if err != nil || stale.Status != "stale" || stale.FriendContact.QQ != "" || stale.MyContact.QQ != "" {
		t.Fatalf("profile change leaked contact data: %#v %v", stale, err)
	}
}

func TestExchangeConsentAllowsAsymmetricContactChannels(t *testing.T) {
	tests := []struct {
		name       string
		requester  ContactProfile
		recipient  ContactProfile
		wantAFromB ContactProfile
		wantBFromA ContactProfile
	}{
		{
			name:       "requester has both and recipient has only QQ",
			requester:  ContactProfile{QQ: "12345678", WeChat: "Alpha_123", Revision: 1},
			recipient:  ContactProfile{QQ: "87654321", Revision: 1},
			wantAFromB: ContactProfile{QQ: "87654321"},
			wantBFromA: ContactProfile{QQ: "12345678", WeChat: "Alpha_123"},
		},
		{
			name:       "requester has only QQ and recipient has only WeChat",
			requester:  ContactProfile{QQ: "12345678", Revision: 1},
			recipient:  ContactProfile{WeChat: "Bravo_456", Revision: 1},
			wantAFromB: ContactProfile{WeChat: "Bravo_456"},
			wantBFromA: ContactProfile{QQ: "12345678"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeSocialRepo()
			nk := newFakeSocialNakama()
			makeFriends(repo, "a", "b")
			repo.profiles["a"] = test.requester
			repo.profiles["b"] = test.recipient
			service := newConsentService(repo, nk)

			// Deliberately submit only QQ. The server must derive all non-empty
			// requester channels instead of trusting a client-side subset.
			requested, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"qq"}})
			if err != nil || !equalChannels(requested.Channels, profileChannels(test.requester)) {
				t.Fatalf("request channels were not derived from the profile: %#v err=%v", requested, err)
			}
			acceptedByB, err := service.RespondExchange(context.Background(), "b", "B", RespondExchangeRequest{FriendID: "a", RequestID: requested.RequestID, Accept: true})
			if err != nil || acceptedByB.Status != "accepted" || !equalChannels(acceptedByB.Channels, []string{"qq", "wechat"}) {
				t.Fatalf("asymmetric profiles were not accepted: %#v err=%v", acceptedByB, err)
			}
			if acceptedByB.FriendContact.QQ != test.wantBFromA.QQ || acceptedByB.FriendContact.WeChat != test.wantBFromA.WeChat {
				t.Fatalf("recipient saw unexpected requester contact: %#v", acceptedByB.FriendContact)
			}
			acceptedByA, err := service.GetExchange(context.Background(), "a", "b")
			if err != nil || acceptedByA.FriendContact.QQ != test.wantAFromB.QQ || acceptedByA.FriendContact.WeChat != test.wantAFromB.WeChat {
				t.Fatalf("requester saw unexpected recipient contact: %#v err=%v", acceptedByA.FriendContact, err)
			}
		})
	}
}

func TestExchangeRequiresOnlyOneContactPerParticipant(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	makeFriends(repo, "a", "b")
	service := newConsentService(repo, nk)

	_, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b"})
	var socialErr *SocialError
	if !errors.As(err, &socialErr) || socialErr.Code != "PROFILE_INCOMPLETE" {
		t.Fatalf("empty requester profile returned the wrong error: %v", err)
	}

	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
	requested, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"qq"}})
	if err != nil {
		t.Fatalf("single-channel requester could not create a request: %v", err)
	}
	_, err = service.RespondExchange(context.Background(), "b", "B", RespondExchangeRequest{FriendID: "a", RequestID: requested.RequestID, Accept: true})
	socialErr = nil
	if !errors.As(err, &socialErr) || socialErr.Code != "PROFILE_INCOMPLETE" {
		t.Fatalf("empty recipient profile returned the wrong error: %v", err)
	}
}

func TestRequesterChangeMakesPendingRequestStale(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	makeFriends(repo, "a", "b")
	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
	repo.profiles["b"] = ContactProfile{QQ: "87654321", Revision: 1}
	service := newConsentService(repo, nk)
	requested, _ := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"qq"}})
	repo.profiles["a"] = ContactProfile{QQ: "22345678", Revision: 2}

	view, err := service.RespondExchange(context.Background(), "b", "B", RespondExchangeRequest{FriendID: "a", RequestID: requested.RequestID, Accept: true})
	var socialErr *SocialError
	if !errors.As(err, &socialErr) || socialErr.Code != "PROFILE_CHANGED" || view.Status != "stale" {
		t.Fatalf("changed requester was not made stale: %#v %v", view, err)
	}
	if view.MyContact.QQ != "" || view.FriendContact.QQ != "" {
		t.Fatalf("stale response leaked contact data: %#v", view)
	}
}

func TestLegacyAndNonAcceptedExchangesNeverDisclose(t *testing.T) {
	statuses := []string{"pending", "declined", "expired", "revoked", "cancelled"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			repo := newFakeSocialRepo()
			nk := newFakeSocialNakama()
			makeFriends(repo, "a", "b")
			repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
			repo.profiles["b"] = ContactProfile{QQ: "87654321", Revision: 1}
			repo.exchanges[pairKey("a", "b")] = ExchangeRequest{RequestID: "r", RequesterID: "a", RecipientID: "b", Channels: []string{"qq"}, Status: status, ExpiresAt: 1_800_000_000}
			view, err := newConsentService(repo, nk).GetExchange(context.Background(), "a", "b")
			if err != nil || view.FriendContact.QQ != "" || view.MyContact.QQ != "" {
				t.Fatalf("%s disclosed contact data: %#v %v", status, view, err)
			}
		})
	}

	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	makeFriends(repo, "a", "b")
	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
	repo.profiles["b"] = ContactProfile{QQ: "87654321", Revision: 1}
	repo.exchanges[pairKey("a", "b")] = ExchangeRequest{RequestID: "legacy", RequesterID: "a", RecipientID: "b", Channels: []string{"qq"}, Status: "accepted"}
	legacy, err := newConsentService(repo, nk).GetExchange(context.Background(), "a", "b")
	if err != nil || legacy.Status != "stale" || legacy.FriendContact.QQ != "" {
		t.Fatalf("legacy accepted exchange was not safely downgraded: %#v %v", legacy, err)
	}
}

func TestInboxClassifiesAuthorityWithoutPII(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	for _, friendID := range []string{"b", "c", "d"} {
		makeFriends(repo, "a", friendID)
		repo.usernames[friendID] = "USER" + friendID
	}
	repo.cursor = "next"
	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 2}
	repo.profiles["d"] = ContactProfile{QQ: "87654321", Revision: 2}
	repo.exchanges[pairKey("a", "b")] = ExchangeRequest{RequestID: "received", RequesterID: "b", RecipientID: "a", Channels: []string{"qq"}, Status: "pending", ExpiresAt: 1_800_000_000}
	repo.exchanges[pairKey("a", "c")] = ExchangeRequest{RequestID: "sent", RequesterID: "a", RecipientID: "c", Channels: []string{"wechat"}, Status: "pending", ExpiresAt: 1_800_000_000}
	repo.exchanges[pairKey("a", "d")] = ExchangeRequest{RequestID: "stale", RequesterID: "a", RecipientID: "d", Channels: []string{"qq"}, Status: "accepted", AcceptedRequesterProfileRevision: 1, AcceptedRecipientProfileRevision: 1}

	inbox, err := newConsentService(repo, nk).ListInbox(context.Background(), "a", ListInboxRequest{Limit: 100})
	if err != nil || len(inbox.Received) != 1 || len(inbox.Sent) != 1 || len(inbox.Reapproval) != 1 || inbox.IncomingPendingCount != 1 || inbox.Cursor != "next" {
		t.Fatalf("unexpected inbox: %#v %v", inbox, err)
	}
	encoded, _ := json.Marshal(inbox)
	if contains(string(encoded), "12345678") || contains(string(encoded), "87654321") {
		t.Fatalf("inbox leaked contact data: %s", encoded)
	}
}

func TestCardFailureDoesNotRollbackCanonicalRequest(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	nk.cardErr = errors.New("channel unavailable")
	makeFriends(repo, "a", "b")
	repo.profiles["a"] = ContactProfile{QQ: "12345678", Revision: 1}
	service := newConsentService(repo, nk)
	view, err := service.RequestExchange(context.Background(), "a", "A", RequestExchangeRequest{FriendID: "b", Channels: []string{"qq"}})
	if err != nil || view.Status != "pending" || repo.exchanges[pairKey("a", "b")].RequestID == "" {
		t.Fatalf("card failure rolled back request: %#v %v", view, err)
	}
}

func TestRepositoryWritesPrivateStorageOnly(t *testing.T) {
	nk := newFakeSocialNakama()
	repo := NewRepository(nk)
	if err := repo.WriteProfile(context.Background(), "a", ContactProfile{QQ: "12345678", Revision: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if err := repo.WriteExchange(context.Background(), "a", "b", ExchangeRequest{RequestID: "r", Status: "pending"}, ""); err != nil {
		t.Fatal(err)
	}
	if len(nk.writes) != 2 {
		t.Fatalf("unexpected storage writes: %d", len(nk.writes))
	}
	for _, write := range nk.writes {
		if write.PermissionRead != runtime.STORAGE_PERMISSION_NO_READ || write.PermissionWrite != runtime.STORAGE_PERMISSION_NO_WRITE {
			t.Fatalf("storage write was not private: %#v", write)
		}
	}
}

func TestDeleteFriendsRevokesByIDAndUsername(t *testing.T) {
	repo := newFakeSocialRepo()
	nk := newFakeSocialNakama()
	nk.users["USERC"] = &api.User{Id: "c", Username: "USERC"}
	repo.exchanges[pairKey("a", "b")] = ExchangeRequest{RequestID: "ab", Status: "accepted"}
	repo.exchanges[pairKey("a", "c")] = ExchangeRequest{RequestID: "ac", Status: "pending"}
	service := newConsentService(repo, nk)
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "a")

	request := &api.DeleteFriendsRequest{Ids: []string{"b"}, Usernames: []string{"USERC"}}
	returned, err := beforeDeleteFriends(service)(ctx, nil, nil, nil, request)
	if err != nil || returned != request || len(repo.revoked) != 2 {
		t.Fatalf("delete hook did not revoke all targets: %#v %v", repo.revoked, err)
	}
	if repo.exchanges[pairKey("a", "b")].Status != "revoked" || repo.exchanges[pairKey("a", "c")].Status != "revoked" {
		t.Fatal("revoked state was not persisted")
	}

	makeFriends(repo, "a", "b")
	view, err := service.GetExchange(context.Background(), "a", "b")
	if err != nil || view.Status != "revoked" || view.FriendContact.QQ != "" {
		t.Fatalf("re-added friendship revived authorization: %#v %v", view, err)
	}
}

func TestDeleteFriendsBlocksWhenRevocationFails(t *testing.T) {
	repo := newFakeSocialRepo()
	repo.failWrite = true
	nk := newFakeSocialNakama()
	service := newConsentService(repo, nk)
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "a")
	if returned, err := beforeDeleteFriends(service)(ctx, nil, nil, nil, &api.DeleteFriendsRequest{Ids: []string{"b"}}); err == nil || returned != nil {
		t.Fatalf("delete proceeded after revocation failure: %#v %v", returned, err)
	}
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if len(needle) > 0 && !contains(value, needle) {
			return false
		}
	}
	return true
}

func contains(value, needle string) bool {
	for index := 0; index+len(needle) <= len(value); index++ {
		if value[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}
