#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

docker compose -f "$SCRIPT_DIR/compose.yml" up --build -d

printf '\nfarelah is running at http://localhost:%s\n' "${APP_PORT:-3000}"
printf 'Stop it with: sh %s/stop.sh\n' "$SCRIPT_DIR"
