package oauth

import (
	gocontext "context"
	"fmt"
	"strings"

	"github.com/LiuShen-Fork/QingyuGist/internal/auth/oauth/genericoauth"
	"github.com/LiuShen-Fork/QingyuGist/internal/config"
	"github.com/LiuShen-Fork/QingyuGist/internal/db"
	"github.com/LiuShen-Fork/QingyuGist/internal/web/context"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type OAuth2Provider struct {
	Provider
	URL string
}

func (p *OAuth2Provider) RegisterProvider() error {
	if config.C.OAuth2AuthUrl == "" || config.C.OAuth2TokenUrl == "" || config.C.OAuth2UserinfoUrl == "" {
		return fmt.Errorf("the generic OAuth 2.0 provider needs an authorize, token and user-info url")
	}

	goth.UseProviders(
		genericoauth.New(
			config.C.OAuth2ClientKey,
			config.C.OAuth2Secret,
			urlJoin(p.URL, "/oauth/oauth2/callback"),
			config.C.OAuth2AuthUrl,
			config.C.OAuth2TokenUrl,
			config.C.OAuth2UserinfoUrl,
			config.C.OAuth2Scopes,
			config.C.OAuth2IdField,
			config.C.OAuth2NameField,
			config.C.OAuth2EmailField,
			config.C.OAuth2AvatarField,
		),
	)

	return nil
}

func (p *OAuth2Provider) BeginAuthHandler(ctx *context.Context) {
	ctxValue := gocontext.WithValue(ctx.Request().Context(), gothic.ProviderParamKey, OAuth2ProviderString)
	ctx.SetRequest(ctx.Request().WithContext(ctxValue))

	gothic.BeginAuthHandler(ctx.Response(), ctx.Request())
}

func (p *OAuth2Provider) UserHasProvider(user *db.User) bool {
	return user.OAuth2ID != ""
}

func NewOAuth2Provider(url string) *OAuth2Provider {
	return &OAuth2Provider{URL: url}
}

type OAuth2CallbackProvider struct {
	CallbackProvider
	User *goth.User
}

func (p *OAuth2CallbackProvider) GetProvider() string {
	return OAuth2ProviderString
}

func (p *OAuth2CallbackProvider) GetProviderUser() *goth.User {
	return p.User
}

func (p *OAuth2CallbackProvider) GetProviderUserID(user *db.User) bool {
	return user.OAuth2ID != ""
}

// A generic server has no known SSH key endpoint.
func (p *OAuth2CallbackProvider) GetProviderUserSSHKeys() ([]string, error) {
	return nil, nil
}

func (p *OAuth2CallbackProvider) UpdateUserDB(user *db.User) {
	user.OAuth2ID = p.User.UserID
	if p.User.AvatarURL != "" {
		user.AvatarURL = p.User.AvatarURL
	}
}

// IsAdmin is false for generic servers: groups are provider specific.
func (p *OAuth2CallbackProvider) IsAdmin() bool {
	return false
}

func NewOAuth2CallbackProvider(user *goth.User) CallbackProvider {
	return &OAuth2CallbackProvider{User: user}
}

// oauth2ProviderName is the label shown for provider errors and flash messages.
func oauth2ProviderName() string {
	if name := strings.TrimSpace(config.C.OAuth2Name); name != "" {
		return name
	}
	return "OAuth 2.0"
}
