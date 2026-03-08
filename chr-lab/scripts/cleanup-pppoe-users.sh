#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

USER_PREFIX="${USER_PREFIX:-cust}"

echo "Removing PPP secrets with prefix: $USER_PREFIX"
"$DIR/routeros-run.sh" "/ppp secret remove [find where name~\"^${USER_PREFIX}\"]"

echo "Disconnecting active PPP sessions with prefix: $USER_PREFIX"
"$DIR/routeros-run.sh" "/ppp active remove [find where name~\"^${USER_PREFIX}\"]"

echo "Done."