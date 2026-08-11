package social

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

var qqPattern = regexp.MustCompile(`^[1-9][0-9]{4,11}$`)
var wechatPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{5,19}$`)

type Service struct {
	repo *Repository
	nk   runtime.NakamaModule
	now  func() time.Time
}

func NewService(nk runtime.NakamaModule) *Service {
	return &Service{repo: NewRepository(nk), nk: nk, now: time.Now}
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
	if profile.QQ == "" && profile.WeChat == "" {
		return &SocialError{Code: "EMPTY_PROFILE", Message: "set at least one contact method"}
	}
	return nil
}

func (s *Service) SetProfile(ctx context.Context, userID string, req SetContactProfileRequest) (ContactProfileSummary, error) {
	if err := s.requireFormalAccount(ctx, userID); err != nil {
		return ContactProfileSummary{}, err
	}
	profile := ContactProfile{QQ: strings.TrimSpace(req.QQ), WeChat: strings.TrimSpace(req.WeChat), UpdatedAt: s.now().Unix()}
	if err := validateProfile(profile); err != nil {
		return ContactProfileSummary{}, err
	}
	_, version, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ContactProfileSummary{}, err
	}
	if err = s.repo.WriteProfile(ctx, userID, profile, version); err != nil {
		return ContactProfileSummary{}, err
	}
	return summarizeProfile(profile), nil
}

func (s *Service) GetExchange(ctx context.Context, userID, friendID string) (ExchangeView, error) {
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
	isFriend, friendErr := s.repo.IsFriend(ctx, userID, friendID)
	if friendErr != nil {
		return ExchangeView{}, friendErr
	}
	if !isFriend {
		exchange.Status = "expired"
		return ExchangeView{ExchangeRequest: exchange}, nil
	}
	if exchange.Status == "pending" && s.now().Unix() > exchange.ExpiresAt {
		exchange.Status = "expired"
	}
	view := ExchangeView{ExchangeRequest: exchange}
	if exchange.Status == "accepted" {
		mine, _, readErr := s.repo.ReadProfile(ctx, userID)
		if readErr != nil {
			return ExchangeView{}, readErr
		}
		friend, _, readErr := s.repo.ReadProfile(ctx, friendID)
		if readErr != nil {
			return ExchangeView{}, readErr
		}
		view.MyContact = filterProfile(mine, exchange.Channels)
		view.FriendContact = filterProfile(friend, exchange.Channels)
	}
	return view, nil
}

func (s *Service) RequestExchange(ctx context.Context, userID, username string, req RequestExchangeRequest) (ExchangeView, error) {
	if err := s.requireFriend(ctx, userID, req.FriendID); err != nil {
		return ExchangeView{}, err
	}
	channels, err := normalizeChannels(req.Channels)
	if err != nil {
		return ExchangeView{}, err
	}
	profile, _, err := s.repo.ReadProfile(ctx, userID)
	if err != nil {
		return ExchangeView{}, err
	}
	if err = profileHasChannels(profile, channels); err != nil {
		return ExchangeView{}, err
	}
	current, storageVersion, err := s.repo.ReadExchange(ctx, userID, req.FriendID)
	if err != nil {
		return ExchangeView{}, err
	}
	now := s.now().Unix()
	if current.Status == "pending" && current.ExpiresAt >= now {
		if current.RequesterID == userID && equalChannels(current.Channels, channels) {
			return ExchangeView{ExchangeRequest: current}, nil
		}
		return ExchangeView{}, &SocialError{Code: "PENDING_EXISTS", Message: "another contact exchange request is already pending"}
	}
	if current.RequestedAt > 0 && now-current.RequestedAt < requestCooldownSec {
		return ExchangeView{}, &SocialError{Code: "RATE_LIMITED", Message: "please wait before sending another request"}
	}
	requestID, err := newRequestID()
	if err != nil {
		return ExchangeView{}, err
	}
	exchange := ExchangeRequest{RequestID: requestID, RequesterID: userID, RecipientID: req.FriendID, Channels: channels, Status: "pending", Version: current.Version + 1, RequestedAt: now, ExpiresAt: now + exchangeTTLSeconds}
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
		profile, _, readErr := s.repo.ReadProfile(ctx, userID)
		if readErr != nil {
			return ExchangeView{}, readErr
		}
		if readErr = profileHasChannels(profile, exchange.Channels); readErr != nil {
			return ExchangeView{}, readErr
		}
		exchange.Status = "accepted"
	} else {
		exchange.Status = "declined"
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

func profileHasChannels(profile ContactProfile, channels []string) error {
	for _, c := range channels {
		if c == "qq" && profile.QQ == "" {
			return &SocialError{Code: "PROFILE_INCOMPLETE", Message: "set your QQ number first"}
		}
		if c == "wechat" && profile.WeChat == "" {
			return &SocialError{Code: "PROFILE_INCOMPLETE", Message: "set your WeChat ID first"}
		}
	}
	return nil
}
func filterProfile(profile ContactProfile, channels []string) ContactProfile {
	out := ContactProfile{UpdatedAt: profile.UpdatedAt}
	for _, c := range channels {
		if c == "qq" {
			out.QQ = profile.QQ
		}
		if c == "wechat" {
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
func equalChannels(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func newRequestID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func (s *Service) sendCard(ctx context.Context, senderID, senderUsername, targetID string, exchange ExchangeRequest, action string) {
	channelID, err := s.nk.ChannelIdBuild(ctx, senderID, targetID, runtime.DirectMessage)
	if err != nil {
		return
	}
	_, _ = s.nk.ChannelMessageSend(ctx, channelID, map[string]interface{}{"type": "contact_exchange", "request_id": exchange.RequestID, "action": action, "version": exchange.Version}, senderID, senderUsername, true)
}
