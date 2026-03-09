#!/bin/bash
# ISP Visual Monitor — Load demo seed data into the database.
# Usage:  bash scripts/demo-seed.sh [--host HOST] [--port PORT]
#
# Requires: psql (from postgresql-client)
# Environment variables respected: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SEED_FILE="$PROJECT_DIR/db/seed/demo_seed.sql"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-ispmonitor}"
DB_PASSWORD="${DB_PASSWORD:-ispmonitor}"
DB_NAME="${DB_NAME:-ispmonitor}"

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) DB_HOST="$2"; shift 2 ;;
    --port) DB_PORT="$2"; shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

echo "========================================"
echo "  ISPVisualMonitor — Demo Seed Loader"
echo "========================================"
echo ""
echo "  Target: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "  Seed:   ${SEED_FILE}"
echo ""

if [ ! -f "$SEED_FILE" ]; then
  echo "ERROR: Seed file not found at $SEED_FILE"
  exit 1
fi

export PGPASSWORD="$DB_PASSWORD"

echo "Loading seed data..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SEED_FILE" 2>&1 | tail -20

echo ""
echo "Done. Demo data loaded."
echo ""
echo "  Tenant:   LebanonNet ISP"
echo "  Login:    demo@lebanonnet.demo / Demo@12345"
echo "  Routers:  10 devices across 3 sites"
echo "  Alerts:   7 (mix of active/acknowledged/resolved)"
echo ""
