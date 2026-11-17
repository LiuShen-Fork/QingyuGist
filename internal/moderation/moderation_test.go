package moderation

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseVerdict(t *testing.T) {
	tests := []struct {
		name    string
		content string
		allowed bool
		reason  string
		wantErr bool
	}{
		{"plain", `{"allowed":true,"reason":"fine"}`, true, "fine", false},
		{"fenced", "```json\n{\"allowed\":false,\"reason\":\"spam\"}\n```", false, "spam", false},
		{"chatty", "Verdict: {\"allowed\": true, \"reason\": \"ok\"} — done", true, "ok", false},
		{"no json", "looks fine to me", false, "", true},
		{"broken json", `{"allowed":`, false, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, reason, err := parseVerdict(tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if allowed != tt.allowed || reason != tt.reason {
				t.Fatalf("got (%v, %q), want (%v, %q)", allowed, reason, tt.allowed, tt.reason)
			}
		})
	}
}

func TestTruncateKeepsValidUTF8(t *testing.T) {
	input := strings.Repeat("中文内容", 10)
	got := truncate(input, 11)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate produced invalid UTF-8: %q", got)
	}
	if len(got) > 11+len("…") {
		t.Fatalf("truncate returned %d bytes, want at most 11 plus ellipsis", len(got))
	}
}
