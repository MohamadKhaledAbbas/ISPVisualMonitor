#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== IP Services ==="
"$DIR/routeros-run.sh" "/ip service print detail"

echo
echo "=== SNMP ==="
"$DIR/routeros-run.sh" "/snmp print"