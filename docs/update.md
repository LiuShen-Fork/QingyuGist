# Update QingyuGist

## Upgrading from OpenGist

QingyuGist is a hard fork, but it reads the existing data format. On the first
start, the binary detects the old home directory and automatically moves the
database, session key and search index to their QingyuGist names. The old secret
and SSH host key are copied rather than deleted, so rolling back remains safe.
The pre-fork `OG_*` environment variables and the old `opengist-home` YAML key
are still accepted.

Docker users can keep the existing volume mount: the entrypoint uses the old
`/opengist` volume while the new `/qingyugist` volume is empty. Move the data to
a `/qingyugist` volume when convenient.

## Make a backup

Before updating, always make sure to backup the QingyuGist home directory, where all the data is stored.

You can do so by copying the `~/.qingyugist` directory (default location).

```shell
cp -r ~/.qingyugist ~/.qingyugist.bak
```

## Install the new version

### With Docker

Pull the last version of QingyuGist
```shell
docker pull ghcr.io/liushen-fork/qingyugist:latest
```

And restart the container, using `docker compose up -d` for example if you use docker compose.

### Via binary

Stop the running instance; then like your first installation of QingyuGist, download the archive for your system from the release page [here](https://github.com/LiuShen-Fork/QingyuGist/releases/latest), and extract it.

```shell
# example for linux amd64
wget https://github.com/LiuShen-Fork/QingyuGist/releases/download/v1.15.3/qingyugist1.15.3-linux-amd64.tar.gz

tar xzvf qingyugist1.15.3-linux-amd64.tar.gz
cd qingyugist
chmod +x qingyugist
./qingyugist # with or without `--config config.yml`
```

### From source

Stop the running instance; then pull the last changes from the master branch, and build the new version.

```shell
git switch master
git pull
make
./qingyugist
```

## Restore the backup

If you have any issue with the new version, you can restore the backup you made before updating.

```shell
rm -rf ~/.qingyugist
cp -r ~/.qingyugist.bak ~/.qingyugist
```

Then run the old version of QingyuGist again.
