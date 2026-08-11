package social

import "testing"

func TestPairKeyIsDeterministic(t *testing.T) {
	a := pairKey("user-b", "user-a")
	b := pairKey("user-a", "user-b")
	if a != b || a != "user-a_user-b" {
		t.Fatalf("unexpected pair keys %q %q", a, b)
	}
}
func TestContactValidationAndFiltering(t *testing.T) {
	p := ContactProfile{QQ: "12345678", WeChat: "Player_code"}
	if err := validateProfile(p); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}
	if err := profileHasChannels(p, []string{"qq", "wechat"}); err != nil {
		t.Fatal(err)
	}
	filtered := filterProfile(p, []string{"qq"})
	if filtered.QQ == "" || filtered.WeChat != "" {
		t.Fatalf("unexpected filtered profile %#v", filtered)
	}
}
func TestChannelNormalizationRejectsArbitraryFields(t *testing.T) {
	channels, err := normalizeChannels([]string{"wechat", "qq", "qq"})
	if err != nil || len(channels) != 2 || channels[0] != "qq" {
		t.Fatalf("unexpected channels %v %v", channels, err)
	}
	if _, err := normalizeChannels([]string{"phone"}); err == nil {
		t.Fatal("phone should not be accepted")
	}
}

func TestProfileSummaryNeverReturnsPlaintext(t *testing.T) {
	summary := summarizeProfile(ContactProfile{QQ: "12345678", WeChat: "Alpha_123"})
	if !summary.QQConfigured || !summary.WeChatConfigured || summary.QQMasked == "12345678" || summary.WeChatMasked == "Alpha_123" {
		t.Fatalf("profile was not safely summarized: %#v", summary)
	}
}
