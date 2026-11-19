#!/bin/sh

set -e

load_secrets() {
    if [ -f "/run/secrets/qingyugist_secrets" ]; then
        set -a
        . /run/secrets/qingyugist_secrets
        set +a
    fi
}

# Existing compose files still mount the pre-fork data directory. Use it when
# the new location is empty, so an image upgrade never looks like data loss.
QYG_DATA_HOME=/qingyugist
LEGACY_DATA_HOME=/opengist
if [ -d "$LEGACY_DATA_HOME" ] && [ -z "$(ls -A "$QYG_DATA_HOME" 2>/dev/null)" ]; then
    QYG_DATA_HOME=$LEGACY_DATA_HOME
fi
export QYG_HOME=$QYG_DATA_HOME

if [ "$(id -u)" -ne 0 ]; then
    load_secrets
    exec env HOME="$QYG_DATA_HOME" QYG_HOME="$QYG_DATA_HOME" /app/qingyugist/qingyugist
fi

export USER=qingyugist
UID=${UID:-1000}
GID=${GID:-1000}
groupmod -o -g "$GID" $USER
usermod -o -u "$UID" $USER

chown -R "$USER:$USER" "$QYG_DATA_HOME"

load_secrets

exec su $USER -c "QYG_HOME=$QYG_DATA_HOME /app/qingyugist/qingyugist"
