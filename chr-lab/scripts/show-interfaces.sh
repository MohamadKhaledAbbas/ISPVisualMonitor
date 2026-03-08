#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Interfaces (brief) ==="
"$DIR/routeros-run.sh" "/interface print"

echo
echo "=== Interface Traffic (RX/TX) ==="
"$DIR/routeros-run.sh" "/interface monitor-traffic [find] once"