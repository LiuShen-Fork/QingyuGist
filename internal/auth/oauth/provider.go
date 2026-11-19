package oauth

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/LiuShen-Fork/QingyuGist/internal/config"
	"github.com/LiuShen-Fork/QingyuGist/internal/db"
	"github.com/LiuShen-Fork/QingyuGist/internal/web/context"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog/log"
)

const (
	GitHubProviderString = "github"
	GoogleProviderString = "google"
	GitLabProviderString = "gitlab"
	GiteaProviderString  = "gitea"
	OpenIDConnectString  = "openid-connect"
	OAuth2ProviderString = "oauth2"
)

type Provider interface {
	RegisterProvider() error
	BeginAuthHandler(ctx *context.Context)
	UserHasProvider(user *db.User) bool
}

type CallbackProvider interface {
	GetProvider() string
	GetProviderUser() *goth.User
	GetProviderUserID(user *db.User) bool
	GetProviderUserSSHKeys() ([]string, error)
	UpdateUserDB(user *db.User)
	IsAdmin() bool
}

func ProviderEnabled(provider string) bool {
	switch provider {
	case GitHubProviderString:
		return config.C.GithubOAuthEnabled()
	case GoogleProviderString:
		return config.C.GoogleOAuthEnabled()
	case GitLabProviderString:
		return config.C.GitlabOAuthEnabled()
	case GiteaProviderString:
		return config.C.GiteaOAuthEnabled()
	case OpenIDConnectString:
		return config.C.OpenIDConnectEnabled()
	case OAuth2ProviderString:
		return config.C.OAuth2OAuthEnabled()
	}

	return false
}

func DefineProvider(provider string, url string) (Provider, error) {
	switch provider {
	case GitHubProviderString:
		return NewGitHubProvider(url), nil
	case GoogleProviderString:
		return NewGoogleProvider(url), nil
	case GitLabProviderString:
		return NewGitLabProvider(url), nil
	case GiteaProviderString:
		return NewGiteaProvider(url), nil
	case OpenIDConnectString:
		return NewOIDCProvider(url), nil
	case OAuth2ProviderString:
		return NewOAuth2Provider(url), nil
	}

	return nil, fmt.Errorf("unsupported provider %s", provider)
}

func CompleteUserAuth(ctx *context.Context) (CallbackProvider, error) {
	injectCodeVerifier(ctx)

	user, err := gothic.CompleteUserAuth(ctx.Response(), ctx.Request())
	if err != nil {
		return nil, err
	}

	switch user.Provider {
	case GitHubProviderString:
		return NewGitHubCallbackProvider(&user), nil
	case GoogleProviderString:
		return NewGoogleCallbackProvider(&user), nil
	case GitLabProviderString:
		return NewGitLabCallbackProvider(&user), nil
	case GiteaProviderString:
		return NewGiteaCallbackProvider(&user), nil
	case OpenIDConnectString:
		return NewOIDCCallbackProvider(&user), nil
	case OAuth2ProviderString:
		return NewOAuth2CallbackProvider(&user), nil
	}

	return nil, fmt.Errorf("unsupported provider %s", user.Provider)
}

func NewCallbackProviderFromSession(provider string, userID string, nickname string, email string, avatarURL string) (CallbackProvider, error) {
	user := &goth.User{
		Provider:  provider,
		UserID:    userID,
		NickName:  nickname,
		Email:     email,
		AvatarURL: avatarURL,
	}

	switch provider {
	case GitHubProviderString:
		return NewGitHubCallbackProvider(user), nil
	case GoogleProviderString:
		return NewGoogleCallbackProvider(user), nil
	case GitLabProviderString:
		return NewGitLabCallbackProvider(user), nil
	case GiteaProviderString:
		return NewGiteaCallbackProvider(user), nil
	case OpenIDConnectString:
		return NewOIDCCallbackProvider(user), nil
	case OAuth2ProviderString:
		return NewOAuth2CallbackProvider(user), nil
	}

	return nil, fmt.Errorf("unsupported provider %s", provider)
}

// ProviderDisplayName returns the name the admin panel gave a provider, so the
// ui says "Sign in with <name>" instead of the protocol behind it.
func ProviderDisplayName(provider string) string {
	switch provider {
	case GitHubProviderString:
		return nameOrDefault(config.C.GithubName, "GitHub")
	case GoogleProviderString:
		return nameOrDefault(config.C.GoogleName, "Google")
	case GitLabProviderString:
		return nameOrDefault(config.C.GitlabName, "GitLab")
	case GiteaProviderString:
		return nameOrDefault(config.C.GiteaName, "Gitea")
	case OpenIDConnectString:
		return nameOrDefault(config.C.OIDCProviderName, "OpenID")
	case OAuth2ProviderString:
		return nameOrDefault(config.C.OAuth2Name, "OAuth 2.0")
	}

	return provider
}

func nameOrDefault(name, fallback string) string {
	if strings.TrimSpace(name) == "" {
		return fallback
	}
	return strings.TrimSpace(name)
}

func urlJoin(base string, elem ...string) string {
	joined, err := url.JoinPath(base, elem...)
	if err != nil {
		log.Error().Err(err).Msg("Cannot join url")
	}

	return joined
}

func readKeys(response *http.Response) ([]string, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not get user keys %v", err)
	}

	keys := strings.Split(string(body), "\n")
	if len(keys[len(keys)-1]) == 0 {
		keys = keys[:len(keys)-1]
	}

	return keys, nil
}
