# Admin panel

The first user created on a QingyuGist instance has access to the admin panel.
Open your user menu and choose **Admin**. The panel is available at
`/admin`.

## Dashboard

The dashboard shows the instance at a glance: gist and user counts, repository,
attachment and search-index disk usage, heap and reserved memory, goroutine
count, uptime and the last search reindex.

## Actions

Maintenance tasks run in the background, so it is safe to leave the page while
one is in progress. Besides the manually triggered actions, the search index is
rebuilt automatically every `search.reindex-days` days (7 by default) and
expired gists are deleted on a schedule.

**Rebuild the contribution heatmap** backfills activity from existing gists.
New gists update the heatmap automatically.

## Configuration

Settings changed here are stored in the database and apply without restarting
the instance. They take precedence over the `QYG_*` environment variables the
process started with, so anything you change here no longer needs an environment
variable. Secrets are shown masked; leaving a masked value untouched keeps the
stored secret.

### Site

- Site name and description used in the interface, page titles and metadata.
- External URL, used for generated links.
- Logo and favicon. Enter a file name from `<qingyugist-home>/custom/`, or a full
  `http(s)://` URL.
- Custom CSS and JavaScript injected into every page.
- Source repository URL, shown as a GitHub icon at the bottom of the sidebar.
- Sidebar links: icon, label and URL for each custom link, plus the source
  repository icon. Existing `custom.static-links` entries keep working.
- Site announcement: Markdown shown above the content on every page while it is
  set. Leave it empty to hide the banner.

### Authentication

- Registration mode: open, OAuth-only, or closed. OAuth-only hides the local
  registration form while keeping sign-in and invitation links available.
- Require login, allow individual gists without login, disable the login form
  (for OAuth-only instances), and disable Gravatar.

### Third-party sign-in

One tab per provider — GitHub, Google, GitLab, Gitea, OpenID Connect and a
generic OAuth 2.0 — with a switch at the top of each tab and a green dot on the
tab once it is on. A provider only appears as a button on the sign-in page when
it is switched on and its credentials are filled in, and a provider that is
switched off cannot be reached through `/oauth/<provider>` either.

Each provider has a display name, so the button can read *GitLab* or the name of
your own instance. The OpenID Connect tab takes the discovery URL and, if the
provider returns groups, the claim that carries them plus the group whose
members become admins. The OAuth 2.0 tab takes the authorize, token and
user-info endpoints, the scope to request, and the JSON field names for the
user id, name, email and avatar — dotted paths such as `data.user.login` work.

### LDAP

- Server URL, bind DN and bind credentials.
- Search base and search filter, with `%s` standing for the username.

### Uploads

- Enable or disable file uploads.
- Maximum size per file and total size per gist.
- Allow or block media files.
- Blocked file extensions (comma-separated). The default list blocks common
  executable and archive formats.

### Search

- Backend: none, the built-in Bleve index, or Meilisearch. Changing it takes
  effect after a restart.
- Meilisearch host and API key, also used after a restart.
- Which fields a query searches by default.
- Reindex interval in days. The index is checked hourly and rebuilt when the
  interval has elapsed.
- Use **Actions → Rebuild search index** to force a rebuild after a manual data
  import.

### Security

- CORS: allow or deny cross-origin requests, choose allowed origins
  (comma-separated, `*` for any), and whether credentials are allowed.

### Moderation

When enabled, a public gist submitted from the web editor or pushed over git is
stored privately and marked *pending* while an asynchronous review asks an
OpenAI-compatible **Responses** endpoint whether it may be published. The review
never runs on the request path, so a push or a form post returns immediately.
Approved gists become public; rejected gists stay private and their author is
notified with the reason. Private and unlisted gists are not reviewed.

Editing or pushing to an already-approved gist starts a new review; the gist
stays online until a verdict arrives, and only a rejection takes it down.

- API endpoint, API key and model. The endpoint is the full Responses URL, for
  example `https://api.openai.com/v1/responses`.
- Reasoning effort, left out of the request when set to *None* for models that
  do not support it.
- Maximum output tokens and maximum input size sent per review.
- Review prompt and request timeout.
- Daily review quota: calls that would exceed it are skipped and the gist stays
  pending, then is picked up again every ten minutes. `0` removes the cap.
- **Fail closed** keeps gists private if the moderation service is unavailable.

Each attempt is recorded on **Admin → Moderation**, together with the verdict,
the reason, the model and the latency. The page also shows how many gists are
waiting and offers **Review now** to retry them without waiting for the next
scheduled attempt.

### Notifications

**Admin → Notifications** publishes a Markdown announcement to every account,
including accounts created later. Recipients see it under the bell in the header
and can mark it read individually or all at once. An optional link turns the
title into a pointer to another page.

### Git and API

- Default branch for new repositories.
- HTTP Git access: turning it off leaves only SSH, and takes effect after a
  restart.
- SSH host name and SSH user name used in clone URLs, for an instance that is
  reached through a proxy or a dedicated SSH host. Leave them empty to derive
  both from the request.
- Enable or disable the HTTP API.
