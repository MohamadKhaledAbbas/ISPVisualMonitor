#!/usr/bin/env bash
set -euo pipefail

# Checks from the host/WSL that ports are open.
# Adjust ports to match your compose mappings.

HOST="${HOST:-127.0.0.1}"
API_PORT="${API_PORT:-8728}"
WEBFIG_PORT="${WEBFIG_PORT:-18081}"
SSH_PORT="${SSH_PORT:-2221}"
WINBOX_PORT="${WINBOX_PORT:-8291}"

check() {
  local name="$1"
  local port="$2"
  if nc -z "$HOST" "$port" >/dev/null 2>&1; then
    echo "[OK] $name $HOST:$port"
  else
    echo "[FAIL] $name $HOST:$port"
    return 1
  fi
}

check "API" "$API_PORT"
check "WebFig" "$WEBFIG_PORT"
check "SSH" "$SSH_PORT"
check "Winbox" "$WINBOX_PORT"