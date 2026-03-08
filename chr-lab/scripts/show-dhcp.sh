#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== DHCP Server ==="
"$DIR/routeros-run.sh" "/ip dhcp-server print detail"

echo
echo "=== DHCP Leases ==="
"$DIR/routeros-run.sh" "/ip dhcp-server lease print detail"