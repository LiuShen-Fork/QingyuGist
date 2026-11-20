// Package genericoauth implements a goth provider for any OAuth 2.0 server that
// exposes a JSON user-info endpoint (Authentik, Keycloak, Gitea, Zitadel, a
// self-written login service, ...). The field names of that JSON document are
// configurable from the admin panel.
package genericoauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

const (
	defaultIDField    = "id"
	defaultNameField  = "username"
	defaultEmailField = "email"
)

// Provider is the goth.Provider implementation for a generic OAuth 2.0 server.
type Provider struct {
	ClientKey    string
	Secret       string
	CallbackURL  string
	HTTPClient   *http.Client
	providerName string

	config      *oauth2.Config
	userinfoURL string
	idField     string
	nameField   string
	emailField  string
}

// New creates the provider. Empty field names fall back to the common ones, so
// a server following the usual conventions needs only urls and credentials.
func New(clientKey, secret, callbackURL, authURL, tokenURL, userinfoURL, scopes, idField, nameField, emailField string) *Provider {
	p := &Provider{
		ClientKey:    clientKey,
		Secret:       secret,
		CallbackURL:  callbackURL,
		providerName: "oauth2",
		userinfoURL:  userinfoURL,
		idField:      orDefault(idField, defaultIDField),
		nameField:    orDefault(nameField, defaultNameField),
		emailField:   orDefault(emailField, defaultEmailField),
	}

	p.config = &oauth2.Config{
		ClientID:     clientKey,
		ClientSecret: secret,
		RedirectURL:  callbackURL,
		Endpoint:     oauth2.Endpoint{AuthURL: authURL, TokenURL: tokenURL},
		Scopes:       splitScopes(scopes),
	}

	return p
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func splitScopes(scopes string) []string {
	return strings.FieldsFunc(scopes, func(r rune) bool {
		return r == ',' || r == ' '
	})
}

// Name returns the name goth knows this provider by ("oauth2").
func (p *Provider) Name() string {
	return p.providerName
}

// SetName changes the provider name, kept for the goth.Provider contract.
func (p *Provider) SetName(name string) {
	p.providerName = name
}

func (p *Provider) Client() *http.Client {
	return goth.HTTPClientWithFallBack(p.HTTPClient)
}

// Debug is a no-op.
func (p *Provider) Debug(bool) {}

// BeginAuth returns the authorization url of the configured server.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	return &Session{AuthURL: p.config.AuthCodeURL(state)}, nil
}

// UnmarshalSession restores the session stored between the two redirects.
func (p *Provider) UnmarshalSession(data string) (goth.Session, error) {
	session := Session{}
	err := json.Unmarshal([]byte(data), &session)
	return &session, err
}

// FetchUser reads the user-info endpoint and maps the configured fields onto a
// goth.User.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	s := session.(*Session)
	user := goth.User{
		AccessToken:  s.AccessToken,
		Provider:     p.Name(),
		RefreshToken: s.RefreshToken,
		ExpiresAt:    s.ExpiresAt,
	}

	if user.AccessToken == "" {
		return user, fmt.Errorf("%s cannot get user information without an access token", p.providerName)
	}
	if p.userinfoURL == "" {
		return user, fmt.Errorf("%s has no user-info endpoint configured", p.providerName)
	}

	req, err := http.NewRequest(http.MethodGet, p.userinfoURL, nil)
	if err != nil {
		return user, err
	}
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.Client().Do(req)
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return user, fmt.Errorf("%s responded with a %d trying to fetch user information: %s", p.providerName, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := p.userFromReader(resp.Body, &user); err != nil {
		return user, err
	}
	if user.UserID == "" {
		return user, fmt.Errorf("%s user-info response has no %q field", p.providerName, p.idField)
	}

	return user, nil
}

func (p *Provider) userFromReader(r io.Reader, user *goth.User) error {
	payload := map[string]any{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return err
	}

	user.UserID = firstField(payload, p.idField, "id", "sub", "user_id")
	user.Name = firstField(payload, p.nameField, "name", "username", "login", "preferred_username", "display_name", "nickname")
	user.Email = firstField(payload, p.emailField, "email", "mail")

	user.NickName = user.Name
	if user.NickName == "" && user.Email != "" {
		user.NickName = strings.SplitN(user.Email, "@", 2)[0]
	}

	return nil
}

func firstField(payload map[string]any, paths ...string) string {
	for _, path := range paths {
		if value := stringAt(payload, path); value != "" {
			return value
		}
	}
	return ""
}

// stringAt walks a dotted path ("profile.avatar_url") through nested objects.
func stringAt(payload map[string]any, path string) string {
	var current any = payload
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = object[part]
		if !ok {
			return ""
		}
	}

	switch value := current.(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		if value {
			return "true"
		}
		return "false"
	}
	return ""
}

// RefreshTokenAvailable reports that refresh tokens are supported.
func (p *Provider) RefreshTokenAvailable() bool { return true }

// RefreshToken exchanges a refresh token for a fresh access token.
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	return p.config.TokenSource(goth.ContextForClient(p.Client()), token).Token()
}
