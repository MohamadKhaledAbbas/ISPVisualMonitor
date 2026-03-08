#!/usr/bin/env bash
set -euo pipefail

# WARNING: This removes containers and volumes for the lab (full reset).

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/_env.sh"

echo "Resetting CHR lab: bringing down + removing volumes..."
docker compose -f "$COMPOSE_FILE" down -v
echo "Done. Bring it up again with: ./scripts/up.sh"