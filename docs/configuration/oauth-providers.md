# Third-party sign-in

QingyuGist can authenticate users through GitHub, Google, GitLab, Gitea, any OpenID
Connect provider, and any OAuth 2.0 server.

Every provider is configured in **Admin → Configuration → Third-party sign-in**:
one tab per provider, a switch that stays off until you turn it on, and a
display name used on the sign-in button. A provider shows up on the sign-in page
only while it is switched on and its client credentials are filled in.

The callback URL has the form `https://your.instance/oauth/<provider>/callback`,
with `<provider>` being `github`, `google`, `gitlab`, `gitea`,
`openid-connect` or `oauth2`.

The `QYG_*` variables below still work as bootstrap values, but the admin panel
takes precedence, so a value entered there no longer needs an environment
variable.

## GitHub

* Create an OAuth app in your [GitHub account settings](https://github.com/settings/applications/new)
* Set the *Authorization callback URL* to `https://your.instance/oauth/github/callback`
* Copy the *Client ID* and the *Client Secret* into the **GitHub** tab

## Google

* Create credentials of type *OAuth client ID* in the [Google Cloud
  console](https://console.cloud.google.com/apis/credentials)
* Add `https://your.instance/oauth/google/callback` as an authorized redirect URI
* Copy the *Client ID* and the *Client Secret* into the **Google** tab

## GitLab

* Create an application in the [GitLab user
  settings](https://gitlab.com/-/user_settings/applications)
* Set the *Redirect URI* to `https://your.instance/oauth/gitlab/callback`
* Copy the *Application ID* and the *Secret* into the **GitLab** tab
* Enter the base URL of your GitLab instance; it defaults to
  `https://gitlab.com/`

## Gitea

* Create an application in *Settings → Applications* of your Gitea instance
* Set the *Redirect URI* to `https://your.instance/oauth/gitea/callback`
* Copy the *Client ID* and the *Client Secret* into the **Gitea** tab
* Enter the base URL of your Gitea instance; it defaults to `https://gitea.com/`

## OpenID Connect

* Register a client in your OpenID provider
* Add `https://your.instance/oauth/openid-connect/callback` as a redirect URI
* Copy the *Client ID* and the *Client Secret* into the **OpenID Connect** tab
* Enter the discovery endpoint, usually something like
  `https://auth.example.com/.well-known/openid-configuration`

### Admin group

When your provider returns group memberships, QingyuGist can grant admin rights
from them. Enter the name of the claim that carries the groups (often `groups`)
and the group whose members should become admins. The result is re-evaluated on
every login, so removing somebody from the group removes the rights.

## OAuth 2.0

For any authorization server that is not covered above, fill in the authorize,
token and user-info endpoints, the scope to request, and the names of the JSON
fields that carry the user id, name, email and avatar. Field names may be dotted
paths into the user-info response, such as `data.user.login`.

## Environment variables

| Provider | Variables |
| --- | --- |
| GitHub | `QYG_GITHUB_CLIENT_KEY`, `QYG_GITHUB_SECRET`, `QYG_GITHUB_NAME` |
| Google | `QYG_GOOGLE_CLIENT_KEY`, `QYG_GOOGLE_SECRET`, `QYG_GOOGLE_NAME` |
| GitLab | `QYG_GITLAB_CLIENT_KEY`, `QYG_GITLAB_SECRET`, `QYG_GITLAB_URL`, `QYG_GITLAB_NAME` |
| Gitea | `QYG_GITEA_CLIENT_KEY`, `QYG_GITEA_SECRET`, `QYG_GITEA_URL`, `QYG_GITEA_NAME` |
| OpenID Connect | `QYG_OIDC_CLIENT_KEY`, `QYG_OIDC_SECRET`, `QYG_OIDC_DISCOVERY_URL`, `QYG_OIDC_PROVIDER_NAME`, `QYG_OIDC_GROUP_CLAIM_NAME`, `QYG_OIDC_ADMIN_GROUP` |
| OAuth 2.0 | `QYG_OAUTH2_CLIENT_KEY`, `QYG_OAUTH2_SECRET`, `QYG_OAUTH2_AUTH_URL`, `QYG_OAUTH2_TOKEN_URL`, `QYG_OAUTH2_USERINFO_URL`, `QYG_OAUTH2_SCOPES`, `QYG_OAUTH2_ID_FIELD`, `QYG_OAUTH2_NAME_FIELD`, `QYG_OAUTH2_EMAIL_FIELD`, `QYG_OAUTH2_AVATAR_FIELD`, `QYG_OAUTH2_NAME` |
