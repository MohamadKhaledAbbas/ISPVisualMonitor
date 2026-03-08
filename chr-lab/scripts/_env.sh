#!/usr/bin/env bash
set -euo pipefail

# Shared helper to locate compose file regardless of where scripts are run from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change this if your compose filename differs
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/docker-compose.chr.dev.yml}"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "Compose file not found: $COMPOSE_FILE" >&2
  echo "Tip: set COMPOSE_FILE=/full/path/to/your-compose.yml" >&2
  exit 1
fi