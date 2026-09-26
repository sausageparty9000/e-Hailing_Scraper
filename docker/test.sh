#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
docker compose -f "$SCRIPT_DIR/compose.yml" run --build --rm --no-deps backend-test
