#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_env.sh"

docker compose -f "$COMPOSE_FILE" start
echo "Started CHR lab with $COMPOSE_FILE"