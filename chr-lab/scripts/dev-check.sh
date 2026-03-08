#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Waiting for CHR to be reachable..."
"$DIR/wait-ready.sh"

echo
echo "==> Checking local ports..."
"$DIR/ports-check.sh"

echo
echo "==> RouterOS quick health dump..."
"$DIR/show-health.sh"