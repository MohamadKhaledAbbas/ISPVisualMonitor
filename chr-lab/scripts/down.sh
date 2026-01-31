#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_env.sh"

docker compose -f "$COMPOSE_FILE" down
echo "Stopped CHR lab ($COMPOSE_FILE)"