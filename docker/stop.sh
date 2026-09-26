#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

docker compose -f "$SCRIPT_DIR/compose.yml" down

printf '\nfarelah services have stopped.\n'

