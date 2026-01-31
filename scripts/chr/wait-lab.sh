#!/usr/bin/env bash
set -euo pipefail

# Wait until CHR routers are reachable (WebFig + API TCP ports).
# Usage:
#   chmod +x scripts/chr/wait-lab.sh
#   ./scripts/chr/wait-lab.sh

routers=(
  "core   127.0.0.1 8728 18081"
  "edge   127.0.0.1 8738 18082"
  "access 127.0.0.1 8748 18083"
  "pppoe  127.0.0.1 8758 18084"
)

retries="${RETRIES:-120}"   # 120 * 2s = 4 minutes default
sleep_s="${SLEEP_S:-2}"

check_port () {
  local host="$1"
  local port="$2"
  nc -z "$host" "$port" >/dev/null 2>&1
}

echo "Waiting for CHR lab to become ready (retries=$retries, sleep=${sleep_s}s)..."

for r in "${routers[@]}"; do
  name="$(awk '{print $1}' <<< "$r")"
  host="$(awk '{print $2}' <<< "$r")"
  api_port="$(awk '{print $3}' <<< "$r")"
  web_port="$(awk '{print $4}' <<< "$r")"

  echo ""
  echo "==> $name: waiting for API $host:$api_port and WebFig $host:$web_port"

  ok_api=0
  ok_web=0
  for i in $(seq 1 "$retries"); do
    if [ "$ok_api" -eq 0 ] && check_port "$host" "$api_port"; then ok_api=1; fi
    if [ "$ok_web" -eq 0 ] && check_port "$host" "$web_port"; then ok_web=1; fi

    if [ "$ok_api" -eq 1 ] && [ "$ok_web" -eq 1 ]; then
      echo "  [OK] $name is reachable"
      break
    fi

    if [ "$i" -eq "$retries" ]; then
      echo "  [FAIL] $name did not become ready in time"
      echo "  Debug:"
      echo "    docker logs isp-chr-$name-01 --tail=200"
      exit 1
    fi

    sleep "$sleep_s"
  done
done

echo ""
echo "CHR lab is ready."