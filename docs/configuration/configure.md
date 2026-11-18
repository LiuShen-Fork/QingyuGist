# Configuration

QingyuGist provides flexible configuration options through either a YAML file and/or environment variables.
You would only need to specify the configuration options you want to change — for any config option left untouched, 
QingyuGist will simply apply the default values.

The [configuration cheat sheet](cheat-sheet.md) lists all available configuration options.


## Configuration via YAML file

The configuration file must be specified when launching the application, using the `--config` flag followed by the path 
to your YAML file.

Usage with Docker Compose :
```yml
services:
  qingyugist:
    # ...
    volumes:
    # ...
    - "/path/to/config.yml:/config.yml"
```

Usage via command line :
```shell
./qingyugist --config /path/to/config.yml
```

You can start by copying and/or modifying the provided [config.yml](https://github.com/LiuShen-Fork/Qingyu-Gist/blob/master/config.yml) file.


## Configuration via Environment Variables

Usage with Docker Compose :
```yml
services:
  qingyugist:
    # ...
    environment:
      QYG_LQYG_LEVEL: "info"
      # etc.
```

Usage via command line :
```shell
QYG_LQYG_LEVEL=info ./qingyugist
```

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
