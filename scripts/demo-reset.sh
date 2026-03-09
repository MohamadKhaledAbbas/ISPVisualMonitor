#!/bin/bash
# ISP Visual Monitor — Reset demo environment to a clean baseline.
# Drops all demo data and re-seeds from scratch.
#
# Usage:  bash scripts/demo-reset.sh [--host HOST] [--port PORT]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

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

export PGPASSWORD="$DB_PASSWORD"

DEMO_TENANT="a0000000-0000-0000-0000-000000000001"

echo "========================================"
echo "  ISPVisualMonitor — Demo Reset"
echo "========================================"
echo ""
echo "  This will DELETE all demo data and re-seed."
echo "  Tenant ID: $DEMO_TENANT"
echo ""

read -rp "Continue? [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

echo ""
echo "Removing demo data..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<SQL
-- Remove demo data in dependency order
DELETE FROM interface_metrics WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM router_metrics WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM pppoe_sessions WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM nat_sessions WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM dhcp_leases WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM polling_history WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM role_specific_metrics WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM alerts WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM alert_rules WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM links WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM router_role_assignments WHERE router_id IN (SELECT id FROM routers WHERE tenant_id = '$DEMO_TENANT');
DELETE FROM router_capabilities WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM router_dependencies WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM interfaces WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM routers WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM pops WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM regions WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE tenant_id = '$DEMO_TENANT');
DELETE FROM api_keys WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM audit_logs WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM users WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM roles WHERE tenant_id = '$DEMO_TENANT';
DELETE FROM tenants WHERE id = '$DEMO_TENANT';
SQL

echo "Demo data removed."
echo ""

echo "Re-seeding..."
bash "$SCRIPT_DIR/demo-seed.sh" --host "$DB_HOST" --port "$DB_PORT"

echo ""
echo "Demo environment reset complete."
