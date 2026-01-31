#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_env.sh"
SERVICE="${1:-}"

if [[ -n "$SERVICE" ]]; then
  docker compose -f "$COMPOSE_FILE" logs -f --tail=200 "$SERVICE"
else
  docker compose -f "$COMPOSE_FILE" logs -f --tail=200
fi