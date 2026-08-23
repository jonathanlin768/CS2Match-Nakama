package social

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

var qqPattern = regexp.MustCompile(`^[1-9][0-9]{4,11}$`)
var wechatPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{5,19}$`)

type socialRepository interface {
	IsFriend(ctx context.Context, userID, friendID string) (bool, error)
	ReadProfile(ctx context.Context, userID string) (ContactProfile, string, error)
	WriteProfile(ctx context.Context, userID string, profile ContactProfile, version string) error
	ReadExchange(ctx context.Context, a, b string) (ExchangeRequest, string, error)
	WriteExchange(ctx context.Context, a, b string, exchange ExchangeRequest, version string) error
	ListFriends(ctx context.Context, userID string, limit int, cursor string) ([]*api.Friend, string, error)
	ReadExchanges(ctx context.Context, userID string, friendIDs []string) (map[string]ExchangeRequest, error)
	RevokeExchange(ctx context.Context, userID, friendID string, revokedAt int64) error
}

type socialNakama interface {
	repositoryNakama
	AccountGetId(ctx context.Context, userID string) (*api.Account, error)
	UsersGetUsername(ctx context.Context, usernames []string) ([]*api.User, error)
	ChannelIdBuild(ctx context.Context, sender string, target string, chanType runtime.ChannelType) (string, error)
	ChannelMessageSend(ctx context.Context, channelID string, content map[string]interface{}, senderID, senderUsername string, persist bool) (*rtapi.ChannelMessageAck, error)
}

type Service struct {
	repo   socialRepository
	nk     socialNakama
	logger runtime.Logger
	now    func() time.Time
}

func NewService(nk socialNakama, logger runtime.Logger) *Service {
	return &Service{repo: NewRepository(nk), nk: nk, logger: logger, now: time.Now}
}

func normalizeChannels(channels []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(channels))
	for _, channel := range channels {
		channel = strings.ToLower(strings.TrimSpace(channel))
		if channel != "qq" && channel != "wechat" {
			return nil, &SocialError{Code: "INVALID_CHANNEL", Message: "channels must be qq or wechat"}
		}
		if !seen[channel] {
			seen[channel] = true
			result = append(result, channel)
		}
	}
	if len(result) == 0 {
		return nil, &SocialError{Code: "INVALID_CHANNEL", Message: "select at least one contact channel"}
	}
	sort.Strings(result)
	return result, nil
}

func validateProfile(profile ContactProfile) error {
	if profile.QQ != "" && !qqPattern.MatchString(profile.QQ) {
		return &SocialError{Code: "INVALID_QQ", Message: "QQ number must contain 5 to 12 digits"}
	}
	if profile.WeChat != "" && !wechatPattern.MatchString(profile.WeChat) {
		return &SocialError{Code: "INVALID_WECHAT", Message: "WeChat ID format is invalid"}
	}
	return nil
}

func normalizeLegacyProfile(profile ContactProfile) ContactProfile {
	if profile.Revision == 0 && (profile.QQ != "" || profile.WeChat != "") {
		profile.Revision = 1
	}
	return profile
}

func (s *Service) GetProfile(ctx context.Context, userID string) (ContactProfileView, error) {
	if err := s.requireFormalAccount(ctx, userID); err != nil {
		return ContactProfileView{}, err
	}
	profile, _, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ContactProfileView{}, err
	}
	return profileView(profile), nil
}

func (s *Service) SetProfile(ctx context.Context, userID string, req SetContactProfileRequest) (ContactProfileView, error) {
	if err := s.requireFormalAccount(ctx, userID); err != nil {
		return ContactProfileView{}, err
	}
	current, version, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ContactProfileView{}, err
	}
	current = normalizeLegacyProfile(current)
	next := ContactProfile{QQ: strings.TrimSpace(req.QQ), WeChat: strings.TrimSpace(req.WeChat)}
	if err := validateProfile(next); err != nil {
		return ContactProfileView{}, err
	}
	if current.QQ == next.QQ && current.WeChat == next.WeChat {
		return profileView(current), nil
	}
	next.Revision = current.Revision + 1
	next.UpdatedAt = s.now().Unix()
	if err = s.repo.WriteProfile(ctx, userID, next, version); err != nil {
		return ContactProfileView{}, &SocialError{Code: "CONFLICT", Message: "contact profile changed; reload and try again"}
	}
	return profileView(next), nil
}

func (s *Service) effectiveExchange(ctx context.Context, userID, friendID string, exchange ExchangeRequest) (ExchangeRequest, error) {
	if exchange.RequestID == "" {
		exchange.Status = "none"
		return exchange, nil
	}
	isFriend, err := s.repo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return ExchangeRequest{}, err
	}
	if !isFriend {
		exchange.Status = "revoked"
		return exchange, nil
	}
	if exchange.Status == "pending" && s.now().Unix() > exchange.ExpiresAt {
		exchange.Status = "expired"
		return exchange, nil
	}
	if exchange.Status != "accepted" {
		return exchange, nil
	}
	if exchange.AcceptedRequesterProfileRevision <= 0 || exchange.AcceptedRecipientProfileRevision <= 0 {
		exchange.Status = "stale"
		return exchange, nil
	}
	requester, _, err := s.repo.ReadProfile(ctx, exchange.RequesterID)
	if err != nil {
		return ExchangeRequest{}, err
	}
	recipient, _, err := s.repo.ReadProfile(ctx, exchange.RecipientID)
	if err != nil {
		return ExchangeRequest{}, err
	}
	if requester.Revision != exchange.AcceptedRequesterProfileRevision || recipient.Revision != exchange.AcceptedRecipientProfileRevision {
		exchange.Status = "stale"
	}
	return exchange, nil
}

func (s *Service) GetExchange(ctx context.Context, userID, friendID string) (ExchangeView, error) {
	if userID == "" || friendID == "" || userID == friendID {
		return ExchangeView{}, &SocialError{Code: "INVALID_FRIEND", Message: "a different friend id is required"}
	}
	exchange, _, err := s.repo.ReadExchange(ctx, userID, friendID)
	if err != nil {
		return ExchangeView{}, err
	}
	if exchange.RequestID == "" {
		if err := s.requireFriend(ctx, userID, friendID); err != nil {
			return ExchangeView{}, err
		}
		return ExchangeView{ExchangeRequest: ExchangeRequest{Status: "none"}}, nil
	}
	exchange, err = s.effectiveExchange(ctx, userID, friendID, exchange)
	if err != nil {
		return ExchangeView{}, err
	}
	view := ExchangeView{ExchangeRequest: exchange}
	if exchange.Status != "accepted" {
		return view, nil
	}
	mine, _, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ExchangeView{}, err
	}
	friend, _, err := s.repo.ReadProfile(ctx, friendID)
	if err != nil {
		return ExchangeView{}, err
	}
	view.MyContact = filterProfile(mine, exchange.Channels)
	view.FriendContact = filterProfile(friend, exchange.Channels)
	return view, nil
}

func (s *Service) RequestExchange(ctx context.Context, userID, username string, req RequestExchangeRequest) (ExchangeView, error) {
	if err := s.requireFriend(ctx, userID, req.FriendID); err != nil {
		return ExchangeView{}, err
	}
	if len(req.Channels) > 0 {
		if _, err := normalizeChannels(req.Channels); err != nil {
			return ExchangeView{}, err
		}
	}
	profile, _, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ExchangeView{}, err
	}
	if err = requireAnyContact(profile); err != nil {
		return ExchangeView{}, err
	}
	channels := profileChannels(profile)
	current, storageVersion, err := s.repo.ReadExchange(ctx, userID, req.FriendID)
	if err != nil {
		return ExchangeView{}, err
	}
	now := s.now().Unix()
	if current.Status == "pending" && current.ExpiresAt >= now {
		if current.RequesterID == userID && equalChannels(current.Channels, channels) && current.RequesterProfileRevision == profile.Revision {
			return ExchangeView{ExchangeRequest: current}, nil
		}
		return ExchangeView{}, &SocialError{Code: "PENDING_EXISTS", Message: "another contact exchange request is already pending"}
	}
	if current.Status == "accepted" {
		effective, effectiveErr := s.effectiveExchange(ctx, userID, req.FriendID, current)
		if effectiveErr != nil {
			return ExchangeView{}, effectiveErr
		}
		if effective.Status == "accepted" {
			return ExchangeView{}, &SocialError{Code: "ALREADY_AUTHORIZED", Message: "contact exchange is already authorized"}
		}
	}
	if current.RequestedAt > 0 && now-current.RequestedAt < requestCooldownSec {
		return ExchangeView{}, &SocialError{Code: "RATE_LIMITED", Message: "please wait before sending another request"}
	}
	requestID, err := newRequestID()
	if err != nil {
		return ExchangeView{}, err
	}
	exchange := ExchangeRequest{
		RequestID: requestID, RequesterID: userID, RecipientID: req.FriendID,
		Channels: channels, Status: "pending", Version: current.Version + 1,
		RequestedAt: now, ExpiresAt: now + exchangeTTLSeconds,
		RequesterProfileRevision: profile.Revision,
	}
	if err = s.repo.WriteExchange(ctx, userID, req.FriendID, exchange, storageVersion); err != nil {
		return ExchangeView{}, &SocialError{Code: "CONFLICT", Message: "exchange state changed; reload and try again"}
	}
	s.sendCard(ctx, userID, username, req.FriendID, exchange, "requested")
	return ExchangeView{ExchangeRequest: exchange}, nil
}

func (s *Service) RespondExchange(ctx context.Context, userID, username string, req RespondExchangeRequest) (ExchangeView, error) {
	if err := s.requireFriend(ctx, userID, req.FriendID); err != nil {
		return ExchangeView{}, err
	}
	exchange, storageVersion, err := s.repo.ReadExchange(ctx, userID, req.FriendID)
	if err != nil {
		return ExchangeView{}, err
	}
	if exchange.RequestID != req.RequestID || exchange.Status != "pending" {
		return ExchangeView{}, &SocialError{Code: "INVALID_STATE", Message: "request is no longer pending"}
	}
	if exchange.RecipientID != userID {
		return ExchangeView{}, &SocialError{Code: "FORBIDDEN", Message: "only the recipient may respond"}
	}
	if s.now().Unix() > exchange.ExpiresAt {
		return ExchangeView{}, &SocialError{Code: "EXPIRED", Message: "request has expired"}
	}
	if req.Accept {
		requester, _, readErr := s.repo.ReadProfile(ctx, exchange.RequesterID)
		if readErr != nil {
			return ExchangeView{}, readErr
		}
		if requester.Revision != exchange.RequesterProfileRevision {
			exchange.Status = "stale"
			exchange.Version++
			exchange.RespondedAt = s.now().Unix()
			if writeErr := s.repo.WriteExchange(ctx, userID, req.FriendID, exchange, storageVersion); writeErr != nil {
				return ExchangeView{}, &SocialError{Code: "CONFLICT", Message: "exchange state changed; reload and try again"}
			}
			s.sendCard(ctx, userID, username, req.FriendID, exchange, "stale")
			return ExchangeView{ExchangeRequest: exchange}, &SocialError{Code: "PROFILE_CHANGED", Message: "requester contact profile changed; request again"}
		}
		recipient, _, readErr := s.repo.ReadProfile(ctx, userID)
		if readErr != nil {
			return ExchangeView{}, readErr
		}
		if readErr = requireAnyContact(requester); readErr != nil {
			return ExchangeView{}, readErr
		}
		if readErr = requireAnyContact(recipient); readErr != nil {
			return ExchangeView{}, readErr
		}
		exchange.Channels = mergeChannels(profileChannels(requester), profileChannels(recipient))
		exchange.Status = "accepted"
		exchange.AcceptedRequesterProfileRevision = requester.Revision
		exchange.AcceptedRecipientProfileRevision = recipient.Revision
	} else {
		exchange.Status = "declined"
		exchange.AcceptedRequesterProfileRevision = 0
		exchange.AcceptedRecipientProfileRevision = 0
	}
	exchange.Version++
	exchange.RespondedAt = s.now().Unix()
	if err = s.repo.WriteExchange(ctx, userID, req.FriendID, exchange, storageVersion); err != nil {
		return ExchangeView{}, &SocialError{Code: "CONFLICT", Message: "exchange state changed; reload and try again"}
	}
	action := "declined"
	if req.Accept {
		action = "accepted"
	}
	s.sendCard(ctx, userID, username, req.FriendID, exchange, action)
	return s.GetExchange(ctx, userID, req.FriendID)
}

func (s *Service) ListInbox(ctx context.Context, userID string, req ListInboxRequest) (ContactExchangeInbox, error) {
	friends, cursor, err := s.repo.ListFriends(ctx, userID, req.Limit, req.Cursor)
	if err != nil {
		return ContactExchangeInbox{}, err
	}
	friendIDs := make([]string, 0, len(friends))
	friendByID := make(map[string]*api.User, len(friends))
	for _, friend := range friends {
		if friend == nil || friend.User == nil || friend.User.Id == "" {
			continue
		}
		friendIDs = append(friendIDs, friend.User.Id)
		friendByID[friend.User.Id] = friend.User
	}
	exchanges, err := s.repo.ReadExchanges(ctx, userID, friendIDs)
	if err != nil {
		return ContactExchangeInbox{}, err
	}
	inbox := ContactExchangeInbox{Received: []InboxItem{}, Sent: []InboxItem{}, Reapproval: []InboxItem{}, Cursor: cursor}
	for _, friendID := range friendIDs {
		exchange, ok := exchanges[friendID]
		if !ok || exchange.RequestID == "" {
			continue
		}
		effective, effectiveErr := s.effectiveExchange(ctx, userID, friendID, exchange)
		if effectiveErr != nil {
			return ContactExchangeInbox{}, effectiveErr
		}
		item := inboxItem(friendID, friendByID[friendID].Username, effective)
		switch effective.Status {
		case "pending":
			if effective.RecipientID == userID {
				inbox.Received = append(inbox.Received, item)
				inbox.IncomingPendingCount++
			} else if effective.RequesterID == userID {
				inbox.Sent = append(inbox.Sent, item)
			}
		case "stale", "revoked":
			inbox.Reapproval = append(inbox.Reapproval, item)
		}
	}
	return inbox, nil
}

func (s *Service) RevokeFriends(ctx context.Context, userID string, ids, usernames []string) error {
	targets := make(map[string]struct{}, len(ids)+len(usernames))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && id != userID {
			targets[id] = struct{}{}
		}
	}
	cleanNames := make([]string, 0, len(usernames))
	for _, username := range usernames {
		if username = strings.TrimSpace(username); username != "" {
			cleanNames = append(cleanNames, username)
		}
	}
	if len(cleanNames) > 0 {
		users, err := s.nk.UsersGetUsername(ctx, cleanNames)
		if err != nil {
			return err
		}
		for _, user := range users {
			if user != nil && user.Id != "" && user.Id != userID {
				targets[user.Id] = struct{}{}
			}
		}
	}
	for targetID := range targets {
		if err := s.repo.RevokeExchange(ctx, userID, targetID, s.now().Unix()); err != nil {
			return &SocialError{Code: "REVOCATION_FAILED", Message: "contact authorization could not be revoked"}
		}
	}
	return nil
}

func (s *Service) requireFriend(ctx context.Context, userID, friendID string) error {
	if userID == "" || friendID == "" || userID == friendID {
		return &SocialError{Code: "INVALID_FRIEND", Message: "a different friend id is required"}
	}
	ok, err := s.repo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !ok {
		return &SocialError{Code: "NOT_FRIENDS", Message: "contact exchange is available to current friends only"}
	}
	return nil
}

func (s *Service) requireFormalAccount(ctx context.Context, userID string) error {
	account, err := s.nk.AccountGetId(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil || strings.TrimSpace(account.Email) == "" {
		return &SocialError{Code: "FORMAL_ACCOUNT_REQUIRED", Message: "link or log in with email before saving contact information"}
	}
	return nil
}

func profileChannels(profile ContactProfile) []string {
	channels := make([]string, 0, 2)
	if profile.QQ != "" {
		channels = append(channels, "qq")
	}
	if profile.WeChat != "" {
		channels = append(channels, "wechat")
	}
	return channels
}

func requireAnyContact(profile ContactProfile) error {
	if len(profileChannels(profile)) == 0 {
		return &SocialError{Code: "PROFILE_INCOMPLETE", Message: "save at least one QQ number or WeChat ID first"}
	}
	return nil
}

func mergeChannels(left, right []string) []string {
	seen := make(map[string]bool, len(left)+len(right))
	for _, channel := range append(append([]string{}, left...), right...) {
		seen[channel] = true
	}
	channels := make([]string, 0, len(seen))
	for _, channel := range []string{"qq", "wechat"} {
		if seen[channel] {
			channels = append(channels, channel)
		}
	}
	return channels
}

func filterProfile(profile ContactProfile, channels []string) ContactProfile {
	out := ContactProfile{UpdatedAt: profile.UpdatedAt, Revision: profile.Revision}
	for _, channel := range channels {
		if channel == "qq" {
			out.QQ = profile.QQ
		}
		if channel == "wechat" {
			out.WeChat = profile.WeChat
		}
	}
	return out
}

func summarizeProfile(profile ContactProfile) ContactProfileSummary {
	summary := ContactProfileSummary{QQConfigured: profile.QQ != "", WeChatConfigured: profile.WeChat != ""}
	if len(profile.QQ) >= 4 {
		summary.QQMasked = profile.QQ[:2] + strings.Repeat("*", len(profile.QQ)-4) + profile.QQ[len(profile.QQ)-2:]
	}
	if profile.WeChat != "" {
		summary.WeChatMasked = profile.WeChat[:1] + strings.Repeat("*", max(2, len(profile.WeChat)-1))
	}
	return summary
}

func profileView(profile ContactProfile) ContactProfileView {
	return ContactProfileView{ContactProfile: profile, Summary: summarizeProfile(profile)}
}

func inboxItem(friendID, username string, exchange ExchangeRequest) InboxItem {
	return InboxItem{
		FriendID: friendID, Username: username, RequestID: exchange.RequestID,
		RequesterID: exchange.RequesterID, RecipientID: exchange.RecipientID,
		Channels: exchange.Channels, Status: exchange.Status, Version: exchange.Version,
		RequestedAt: exchange.RequestedAt, ExpiresAt: exchange.ExpiresAt,
	}
}

func equalChannels(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func newRequestID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func (s *Service) sendCard(ctx context.Context, senderID, senderUsername, targetID string, exchange ExchangeRequest, action string) {
	channelID, err := s.nk.ChannelIdBuild(ctx, senderID, targetID, runtime.DirectMessage)
	if err == nil {
		_, err = s.nk.ChannelMessageSend(ctx, channelID, map[string]interface{}{
			"type": "contact_exchange", "request_id": exchange.RequestID,
			"action": action, "version": exchange.Version,
		}, senderID, senderUsername, true)
	}
	if err != nil && s.logger != nil {
		s.logger.Warn("Contact exchange card delivery failed action=%s request_id=%s: %v", action, exchange.RequestID, err)
	}
}
