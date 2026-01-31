#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Interfaces ==="
"$DIR/routeros-run.sh" "/interface print"

echo
echo "=== IP Addresses ==="
"$DIR/routeros-run.sh" "/ip address print"

echo
echo "=== Routes ==="
"$DIR/routeros-run.sh" "/ip route print detail"

echo
echo "=== NAT Rules ==="
"$DIR/routeros-run.sh" "/ip firewall nat print detail"

echo
echo "=== Filter Rules ==="
"$DIR/routeros-run.sh" "/ip firewall filter print detail"