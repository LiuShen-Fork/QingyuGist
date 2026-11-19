# Configuration

QingyuGist has no configuration file. Every option ships with a default compiled
into the binary, so a fresh instance starts without any configuration at all.
Anything you need to change is set through an environment variable; site-level
settings — name, registration, uploads, OAuth providers, links — live in the
[admin panel](admin-panel.md) and are stored in the database.

The [configuration cheat sheet](cheat-sheet.md) lists every option, its
environment variable and its default value.

## Environment Variables

Usage with Docker Compose:
```yml
services:
  qingyugist:
    # ...
    environment:
      QYG_LOG_LEVEL: "info"
      # etc.
```

Usage via command line:
```shell
QYG_LOG_LEVEL=info ./qingyugist
```

The data directory is the only option worth setting explicitly in most
deployments: `QYG_HOME` points at it, and defaults to `~/.qingyugist`.

### Using Docker Compose secrets

You can use Docker Compose secrets to not expose sensitive information in your compose file, using a `.env` file.

```dotenv
# file secrets.env
QYG_GITLAB_CLIENT_KEY=your_gitlab_client_key
QYG_GITLAB_SECRET=your_gitlab_secret_key
```

And then use it in your compose file :

```yml
services:
  qingyugist:
    # ...
    secrets:
      - qingyugist_secrets

secrets:
  qingyugist_secrets:
    file: ./secrets.env
```

The image loads `/run/secrets/qingyugist_secrets` at startup when it exists, so
the file above needs no further wiring.

## Admin panel

Settings made in the admin panel are written to the database and apply
immediately, without a restart. They take precedence over the corresponding
`QYG_*` environment variables, which stay useful as bootstrap values for a
fresh instance.

## Upgrading from a configuration file

Releases before 1.0.4 read a YAML `config.yml`. That file is no longer loaded:
rename each option to its `QYG_*` environment variable (the cheat sheet maps
one to the other) and delete the file. `--config` is still accepted so the git
hooks installed by older releases keep working, but its value is ignored.
