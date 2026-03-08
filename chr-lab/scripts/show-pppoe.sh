#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== PPP Active Sessions ==="
"$DIR/routeros-run.sh" "/ppp active print detail"

echo
echo "=== PPP Secrets (users) ==="
"$DIR/routeros-run.sh" "/ppp secret print detail"