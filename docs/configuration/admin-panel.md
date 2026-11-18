# Admin panel

The first user created on an Opengist instance has access to the admin panel.
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
the instance. They take precedence over the corresponding values in
`config.yml`, so the configuration file is optional for these options. Secrets
are shown masked; leaving a masked value untouched keeps the stored secret.

### Site

- Site name and description used in the interface, page titles and metadata.
- External URL, used for generated links.
- Logo and favicon. Enter a file name from `<opengist-home>/custom/`, or a full
  `http(s)://` URL.
- Custom CSS and JavaScript injected into every page.

### Authentication

- Registration mode: open, OAuth-only, or closed. OAuth-only hides the local
  registration form while keeping sign-in and invitation links available.
- Require login, allow individual gists without login, disable the login form
  (for OAuth-only instances), and disable Gravatar.

### Uploads

- Enable or disable file uploads.
- Maximum size per file and total size per gist.
- Allow or block media files.
- Blocked file extensions (comma-separated). The default list blocks common
  executable and archive formats.

### Search

- Reindex interval in days. The index is checked hourly and rebuilt when the
  interval has elapsed.
- Use **Actions → Rebuild search index** to force a rebuild after a manual data
  import.

### Security

- CORS: allow or deny cross-origin requests, choose allowed origins
  (comma-separated, `*` for any), and whether credentials are allowed.

### Moderation

When enabled, a newly published public gist is made private and marked
*pending* while it is reviewed asynchronously by an OpenAI-compatible chat
completions endpoint. Approved gists become public; rejected gists stay private.
Private gists are not reviewed.

- API endpoint, API key and model.
- Review prompt and request timeout.
- **Fail closed** keeps gists private if the moderation service is unavailable.

### Git and API

- Default branch for new repositories.
- Enable or disable the HTTP API.
