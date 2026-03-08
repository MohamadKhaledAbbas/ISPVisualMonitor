#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ROUTER_HOST=127.0.0.1 ROUTER_PORT=2221 ROUTER_USER=admin ROUTER_PASS=admin \
#     ./routeros-run.sh '/ppp active print'

ROUTER_HOST="${ROUTER_HOST:-127.0.0.1}"
ROUTER_PORT="${ROUTER_PORT:-2221}"
ROUTER_USER="${ROUTER_USER:-admin}"
ROUTER_PASS="${ROUTER_PASS:-admin}"

CMD="${1:-}"
if [[ -z "$CMD" ]]; then
  echo "Usage: $0 '<routeros command>'" >&2
  exit 1
fi

if ! command -v sshpass >/dev/null 2>&1; then
  echo "Missing dependency: sshpass" >&2
  echo "Install: sudo apt update && sudo apt install -y sshpass" >&2
  exit 1
fi

sshpass -p "$ROUTER_PASS" ssh \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null \
  -p "$ROUTER_PORT" \
  "$ROUTER_USER@$ROUTER_HOST" \
  "$CMD"