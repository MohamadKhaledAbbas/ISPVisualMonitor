#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Print export to stdout; you can redirect to a file:
# ./scripts/export-config.sh > chr-export.rsc

"$DIR/routeros-run.sh" "/export show-sensitive"