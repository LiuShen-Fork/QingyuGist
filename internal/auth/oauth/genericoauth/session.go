package genericoauth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/markbates/goth"
)

// Session keeps the state of one authorization round trip.
type Session struct {
	AuthURL      string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// GetAuthURL returns the url built by BeginAuth.
func (s Session) GetAuthURL() (string, error) {
	if s.AuthURL == "" {
		return "", errors.New(goth.NoAuthUrlErrorMessage)
	}
	return s.AuthURL, nil
}

// Authorize exchanges the callback code for an access token.
func (s *Session) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	p := provider.(*Provider)
	token, err := p.config.Exchange(goth.ContextForClient(p.Client()), params.Get("code"))
	if err != nil {
		return "", err
	}
	if !token.Valid() {
		return "", errors.New("invalid token received from provider")
	}

	s.AccessToken = token.AccessToken
	s.RefreshToken = token.RefreshToken
	s.ExpiresAt = token.Expiry

	return token.AccessToken, nil
}

// Marshal serialises the session for the state cookie.
func (s Session) Marshal() string {
	value, _ := json.Marshal(s)
	return string(value)
}

// String is the same as Marshal.
func (s Session) String() string {
	return s.Marshal()
}
