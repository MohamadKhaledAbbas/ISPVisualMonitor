#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

COUNT="${1:-50}"
USER_PREFIX="${USER_PREFIX:-cust}"
PASS="${PASS:-pass123}"
PROFILE="${PROFILE:-default}"

echo "Creating $COUNT PPPoE users: ${USER_PREFIX}001.. using profile=$PROFILE"
for i in $(seq -w 1 "$COUNT"); do
  u="${USER_PREFIX}${i}"
  "$DIR/routeros-run.sh" "/ppp secret add name=$u password=$PASS service=pppoe profile=$PROFILE"
done

echo "Done."