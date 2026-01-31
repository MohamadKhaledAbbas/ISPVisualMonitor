#!/usr/bin/env bash
set -euo pipefail

HOST="${HOST:-127.0.0.1}"
WINBOX_PORT="${WINBOX_PORT:-8291}"
WEBFIG_PORT="${WEBFIG_PORT:-18081}"

cat <<EOF
Winbox connection:
  Connect To: $HOST
  Port      : $WINBOX_PORT

WebFig:
  http://$HOST:$WEBFIG_PORT

Note:
- Winbox "Neighbors" discovery usually will NOT show Docker NAT devices.
- You must connect manually by host+port.
EOF