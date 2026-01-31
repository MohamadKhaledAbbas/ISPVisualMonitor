#!/usr/bin/env bash
set -euo pipefail

# Wait for CHR services to become reachable (useful because CHR can boot slowly under emulation).

HOST="${HOST:-127.0.0.1}"
API_PORT="${API_PORT:-8728}"
WEBFIG_PORT="${WEBFIG_PORT:-18081}"

RETRIES="${RETRIES:-180}"   # 180 * 2s = 6 minutes
SLEEP_S="${SLEEP_S:-2}"

wait_port() {
  local name="$1"
  local port="$2"
  for i in $(seq 1 "$RETRIES"); do
    if nc -z "$HOST" "$port" >/dev/null 2>&1; then
      echo "[OK] $name is reachable on $HOST:$port"
      return 0
    fi
    sleep "$SLEEP_S"
  done
  echo "[FAIL] Timed out waiting for $name on $HOST:$port"
  return 1
}

wait_port "API" "$API_PORT"
wait_port "WebFig" "$WEBFIG_PORT"

echo "CHR router looks ready."