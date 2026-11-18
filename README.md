# QingyuGist · 清羽Gist

<img height="108px" src="https://raw.githubusercontent.com/LiuShen-Fork/QingyuGist/master/public/img/qingyugist.svg" alt="QingyuGist" align="right" />

QingyuGist (清羽Gist) is a self-hosted pastebin powered by Git. Every snippet is
a real Git repository, so it can be read and edited from the web interface or
with ordinary `git clone`, `pull` and `push`.

This repository is an independent hard fork. It keeps the data format of the
upstream project, but the interface, defaults and deployment tooling have been
reworked around a simpler, calmer experience.

[Repository](https://github.com/LiuShen-Fork/QingyuGist) · [Documentation](docs/) · [Releases](https://github.com/LiuShen-Fork/QingyuGist/releases)

![CI](https://github.com/LiuShen-Fork/QingyuGist/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/github/license/LiuShen-Fork/QingyuGist?color=blue)

## Features

- Public, unlisted and private gists, with revisions, likes, forks and topics
- Clone, pull and push over HTTP or SSH
- Syntax highlighting, Markdown, CSV and notebook previews
- Full-text search, including Chinese word segmentation, with a scheduled reindex
- Explore feeds and contribution heatmaps, globally and per user
- OAuth2 login with GitHub, GitLab, Gitea and OpenID Connect; passkeys and TOTP
- Per-instance branding: name, description, logo, favicon, custom CSS and JavaScript
- Admin panel with live settings, storage and runtime dashboard, and user management
- Optional AI moderation for public gists, upload quotas and CORS controls
- Responsive light/dark interface, English and Simplified Chinese out of the box

## Quick start

### Docker

Images are published to GitHub Container Registry on every release:

```shell
docker pull ghcr.io/liushen-fork/qingyugist:latest
```

```yaml
services:
  qingyugist:
    image: ghcr.io/liushen-fork/qingyugist:latest
    container_name: qingyugist
    restart: unless-stopped
    ports:
      - "6157:6157"
      - "2222:2222"
    volumes:
      - "$HOME/.qingyugist:/qingyugist"
```

Then open <http://localhost:6157>. The first account created becomes the
instance administrator.

To run the container as a specific user, set `UID` and `GID`:

```yaml
    environment:
      UID: 1001
      GID: 1001
```

### Binary

Download the archive for your platform from the
[latest release](https://github.com/LiuShen-Fork/QingyuGist/releases/latest),
extract it and run the binary:

```shell
tar xzf qingyugist-1.0.0-linux-amd64.tar.gz
./qingyugist
```

`qingyugist --config config.yml` uses an explicit YAML configuration file. All
options can also be supplied through `QYG_*` environment variables; see
[`config.yml`](config.yml) and the [configuration docs](docs/configuration/configure.md).

### From source

Requirements: Git 2.28+, Go 1.27+, Node.js 26+ and Make.

```shell
git clone https://github.com/LiuShen-Fork/QingyuGist.git
cd QingyuGist
make
./qingyugist
```

The frontend assets are embedded into the binary, so the single executable is
the only runtime artifact.

## Configuration

Most day-to-day options live in the admin panel at `/admin/settings` and take
effect immediately: registration mode, upload limits, moderation, CORS, search
reindexing, branding and the footer. A configuration file is still supported for
bootstrap values such as the data directory, database URL, ports and OAuth
providers.

Data lives in `~/.qingyugist` by default. Set `QYG_HOME` to use another
directory. Existing installations are migrated automatically: the old database,
session key and SSH host key are detected, and the repository symlink used by
existing Git hooks remains available under its historical name.

## Upgrading from OpenGist

Upgrading an existing OpenGist v1.15.x installation does not require a manual
database conversion. Stop the old process, back up the data directory, then
start QingyuGist with the same data:

- The old `~/.opengist` directory, `OG_*` environment variables, YAML keys
  such as `opengist-home`, and the legacy Docker volume `/opengist` are
  still accepted.
- The SQLite database and search index are renamed in place; the session secret
  and SSH host key are copied. Back up the data directory first, because rolling
  back to the old binary means restoring that backup or renaming the database
  back to its old name.
- Existing repositories, Git hooks, SSH keys, access tokens, passwords and
  gists are unchanged. Existing `/-/...` bookmarks redirect to the new routes,
  and existing Git remotes continue to work.
- Existing browser sessions are invalidated by the cookie rename; users simply
  sign in again. The search index may be rebuilt in the background after the
  first start.

Two deployment details need a manual update: the executable is now named
`qingyugist` (update systemd or other service units), and the container image is
`ghcr.io/liushen-fork/qingyugist`. New installations should use the new names;
the compatibility paths are intended for upgrades only.

## Development

```shell
make install
make watch
```

Run the test suite with:

```shell
go test -p 1 ./internal/...
```

See [development docs](docs/contributing/development.md) for the Vite workflow,
database matrix and test environment.

## License

QingyuGist is licensed under the [AGPL-3.0 license](LICENSE).
