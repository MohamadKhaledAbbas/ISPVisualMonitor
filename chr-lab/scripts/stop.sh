#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_env.sh"

docker compose -f "$COMPOSE_FILE" stop
echo "Stopped CHR lab ($COMPOSE_FILE)"