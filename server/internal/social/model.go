package social

import "fmt"

const (
	profileCollection  = "social_contact_profile"
	exchangeCollection = "social_contact_exchange"
	profileKey         = "profile"
	systemOwnerID      = "00000000-0000-0000-0000-000000000000"
	exchangeTTLSeconds = int64(7 * 24 * 60 * 60)
	requestCooldownSec = int64(30)
)

type ContactProfile struct {
	QQ        string `json:"qq,omitempty"`
	WeChat    string `json:"wechat,omitempty"`
	UpdatedAt int64  `json:"updated_at"`
}

type SetContactProfileRequest struct {
	QQ     string `json:"qq,omitempty"`
	WeChat string `json:"wechat,omitempty"`
}

type ContactProfileSummary struct {
	QQConfigured     bool   `json:"qq_configured"`
	WeChatConfigured bool   `json:"wechat_configured"`
	QQMasked         string `json:"qq_masked,omitempty"`
	WeChatMasked     string `json:"wechat_masked,omitempty"`
}

type ExchangeRequest struct {
	RequestID   string   `json:"request_id"`
	RequesterID string   `json:"requester_id"`
	RecipientID string   `json:"recipient_id"`
	Channels    []string `json:"channels"`
	Status      string   `json:"status"`
	Version     int64    `json:"version"`
	RequestedAt int64    `json:"requested_at"`
	RespondedAt int64    `json:"responded_at,omitempty"`
	ExpiresAt   int64    `json:"expires_at"`
}

type ExchangeView struct {
	ExchangeRequest
	MyContact     ContactProfile `json:"my_contact,omitempty"`
	FriendContact ContactProfile `json:"friend_contact,omitempty"`
}

type FriendRequest struct {
	FriendID string `json:"friend_id"`
}

type RequestExchangeRequest struct {
	FriendID string   `json:"friend_id"`
	Channels []string `json:"channels"`
}

type RespondExchangeRequest struct {
	FriendID  string `json:"friend_id"`
	RequestID string `json:"request_id"`
	Accept    bool   `json:"accept"`
}

type SocialError struct{ Code, Message string }

func (e *SocialError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }
