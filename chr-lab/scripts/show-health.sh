#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Identity ==="
"$DIR/routeros-run.sh" "/system identity print"

echo
echo "=== Resource ==="
"$DIR/routeros-run.sh" "/system resource print"

echo
echo "=== Clock ==="
"$DIR/routeros-run.sh" "/system clock print"

echo
echo "=== Packages ==="
"$DIR/routeros-run.sh" "/system package print"