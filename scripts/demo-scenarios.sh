#!/bin/bash
# ISP Visual Monitor — Trigger demo scenarios by updating database state.
# These simulate real-world ISP events for demo / development.
#
# Usage:  bash scripts/demo-scenarios.sh [scenario] [--host HOST]
#
# Scenarios:
#   healthy          Reset all routers to active, clear active alerts
#   router-offline   Mark TRI-ACCESS-01 as offline, create alert
#   core-congestion  Spike interface utilization on core uplink
#   upstream-failure Set upstream ISP-B interface to down
#   packet-loss      Inject packet loss metrics on Tripoli uplink
#   high-sessions    Push PPPoE session count near capacity

set -euo pipefail

# Auto-detect database host
if [[ -n "${DB_HOST:-}" ]]; then
  # Use explicitly set DB_HOST
  :
elif ping -c 1 postgres >/dev/null 2>&1; then
  # In Docker network, postgres hostname is resolvable
  DB_HOST="postgres"
elif nc -z postgres 5432 >/dev/null 2>&1; then
  # Postgres hostname exists and port is open
  DB_HOST="postgres"
else
  # Default to localhost
  DB_HOST="localhost"
fi

DB_HOST="${DB_HOST}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-ispmonitor}"
DB_PASSWORD="${DB_PASSWORD:-ispmonitor}"
DB_NAME="${DB_NAME:-ispmonitor}"
DB_CONNECT_TIMEOUT="${DB_CONNECT_TIMEOUT:-5}"

DEMO_TENANT="a0000000-0000-0000-0000-000000000001"

# Parse flags
SCENARIO="${1:-}"
shift || true
while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) DB_HOST="$2"; shift 2 ;;
    --port) DB_PORT="$2"; shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

export PGPASSWORD="$DB_PASSWORD"
export PGCONNECT_TIMEOUT="$DB_CONNECT_TIMEOUT"

echo "Target DB: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME} (connect timeout: ${DB_CONNECT_TIMEOUT}s)"

print_usage() {
  echo "ISP Visual Monitor — Demo Scenarios"
  echo ""
  echo "Usage: bash scripts/demo-scenarios.sh <scenario>"
  echo ""
  echo "Available scenarios:"
  echo "  healthy          Reset to healthy baseline"
  echo "  router-offline   Simulate router going offline"
  echo "  core-congestion  Simulate core link congestion"
  echo "  upstream-failure Simulate upstream provider failure"
  echo "  packet-loss      Simulate packet loss spike"
  echo "  high-sessions    Simulate high PPPoE session count"
  echo ""
  echo "Options:"
  echo "  --host HOST      Database host (default: localhost)"
  echo "  --port PORT      Database port (default: 5432)"
}

run_psql_stdin() {
  if command -v psql >/dev/null 2>&1; then
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
      echo "ERROR: Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
      echo ""
      echo "Troubleshooting:"
      echo "  1. Start services with: ./lab.sh"
      echo "  2. Check if postgres is running: docker ps | grep postgres"
      echo "  3. Set DB_HOST if needed: export DB_HOST=postgres"
      echo ""
      exit 1
    fi
    psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f -
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "Local psql not found; using docker-compose exec postgres psql..."
    docker-compose up -d postgres >/dev/null
    docker-compose exec -T postgres pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null
    docker-compose exec -T -e PGPASSWORD="$DB_PASSWORD" -e PGCONNECT_TIMEOUT="$DB_CONNECT_TIMEOUT" postgres \
      psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f -
  elif command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    echo "Local psql not found; using docker compose exec postgres psql..."
    docker compose up -d postgres >/dev/null
    docker compose exec -T postgres pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null
    docker compose exec -T -e PGPASSWORD="$DB_PASSWORD" -e PGCONNECT_TIMEOUT="$DB_CONNECT_TIMEOUT" postgres \
      psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f -
  else
    echo "ERROR: Neither local 'psql' nor Docker Compose is available."
    echo "Install postgresql-client with: sudo apk add postgresql-client"
    echo "  (or sudo apt-get install postgresql-client on Debian/Ubuntu)"
    exit 127
  fi
}

run_sql() {
  printf "%s\n" "$1" | run_psql_stdin
}

if [[ -z "$SCENARIO" ]]; then
  print_usage
  exit 0
fi

case "$SCENARIO" in

  healthy)
    echo "Scenario: healthy — resetting all routers to active..."
    run_sql "
      UPDATE routers SET status = 'active' WHERE tenant_id = '$DEMO_TENANT';
      UPDATE interfaces SET status = 'up' WHERE tenant_id = '$DEMO_TENANT';
      UPDATE links SET status = 'up' WHERE tenant_id = '$DEMO_TENANT';
      UPDATE alerts SET status = 'resolved', resolved_at = NOW() WHERE tenant_id = '$DEMO_TENANT' AND status IN ('active', 'acknowledged');
    "
    echo "Done. All routers active, alerts resolved."
    ;;

  router-offline)
    echo "Scenario: router-offline — TRI-ACCESS-01 going down..."
    run_sql "
      UPDATE routers SET status = 'offline' WHERE id = 'e0000000-0000-0000-0000-000000000007';
      UPDATE interfaces SET status = 'down' WHERE router_id = 'e0000000-0000-0000-0000-000000000007';
      INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
      VALUES (
        uuid_generate_v4(), '$DEMO_TENANT',
        'TRI-ACCESS-01 offline',
        'Router TRI-ACCESS-01 unreachable. Last poll timed out after 30s.',
        'critical', 'active', 'router', 'e0000000-0000-0000-0000-000000000007', NOW(),
        '{\"consecutive_failures\": 3}'
      );
    "
    echo "Done. TRI-ACCESS-01 is offline."
    ;;

  core-congestion)
    echo "Scenario: core-congestion — spiking traffic on core uplink..."
    run_sql "
      INSERT INTO interface_metrics (tenant_id, interface_id, timestamp, in_octets, out_octets, in_packets, out_packets, in_errors, out_errors, utilization_percent)
      SELECT
        '$DEMO_TENANT',
        'f0000000-0000-0000-0001-000000000003',
        ts,
        (25000 * 125000 * (0.85 + random() * 0.12))::bigint,
        (25000 * 125000 * (0.70 + random() * 0.15))::bigint,
        (80000 + random() * 40000)::bigint,
        (60000 + random() * 30000)::bigint,
        (random() * 10)::bigint,
        (random() * 5)::bigint,
        85 + random() * 12
      FROM generate_series(NOW() - INTERVAL '15 minutes', NOW(), INTERVAL '1 minute') AS ts;

      INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
      VALUES (
        uuid_generate_v4(), '$DEMO_TENANT',
        'Core link congestion — BEY-CORE-01 sfp1-uplink',
        'Upstream interface utilization at 93%. Sustained for 15 minutes.',
        'warning', 'active', 'interface', 'f0000000-0000-0000-0001-000000000003', NOW(),
        '{\"utilization_percent\": 93, \"threshold\": 85}'
      );
    "
    echo "Done. Core uplink congestion active."
    ;;

  upstream-failure)
    echo "Scenario: upstream-failure — ISP-B link going down..."
    run_sql "
      UPDATE interfaces SET status = 'down' WHERE id = 'f0000000-0000-0000-0004-000000000002';
      INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
      VALUES (
        uuid_generate_v4(), '$DEMO_TENANT',
        'Upstream ISP-B link down',
        'BEY-UPSTREAM-01 ether2-isp-b is down. All traffic failing over to ISP-A.',
        'critical', 'active', 'interface', 'f0000000-0000-0000-0004-000000000002', NOW(),
        '{\"failover_active\": true, \"isp\": \"ISP-B\"}'
      );
    "
    echo "Done. Upstream ISP-B link is down."
    ;;

  packet-loss)
    echo "Scenario: packet-loss — injecting errors on Tripoli uplink..."
    run_sql "
      INSERT INTO interface_metrics (tenant_id, interface_id, timestamp, in_octets, out_octets, in_packets, out_packets, in_errors, out_errors, utilization_percent)
      SELECT
        '$DEMO_TENANT',
        'f0000000-0000-0000-0006-000000000001',
        ts,
        (10000 * 125000 * (0.4 + random() * 0.2))::bigint,
        (10000 * 125000 * (0.2 + random() * 0.15))::bigint,
        (30000 + random() * 20000)::bigint,
        (15000 + random() * 10000)::bigint,
        (500 + random() * 800)::bigint,
        (200 + random() * 400)::bigint,
        40 + random() * 20
      FROM generate_series(NOW() - INTERVAL '10 minutes', NOW(), INTERVAL '1 minute') AS ts;

      INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
      VALUES (
        uuid_generate_v4(), '$DEMO_TENANT',
        'Packet loss spike — TRI-EDGE-01 sfp1-core',
        'Error rate spiked to 3.2% on Tripoli uplink. Possible fiber degradation.',
        'warning', 'active', 'interface', 'f0000000-0000-0000-0006-000000000001', NOW(),
        '{\"error_rate_percent\": 3.2, \"threshold\": 1.0}'
      );
    "
    echo "Done. Packet loss injected on Tripoli uplink."
    ;;

  high-sessions)
    echo "Scenario: high-sessions — pushing PPPoE sessions near capacity..."
    run_sql "
      INSERT INTO pppoe_sessions (tenant_id, router_id, session_id, username, calling_station_id, framed_ip_address, status, connect_time, bytes_in, bytes_out)
      SELECT
        '$DEMO_TENANT',
        'e0000000-0000-0000-0000-000000000005',
        'pppoe-surge-' || n,
        'surge' || lpad(n::text, 4, '0') || '@lebanonnet.demo',
        '00:SU:RG:' || lpad(to_hex(n / 256), 2, '0') || ':' || lpad(to_hex(n % 256), 2, '0') || ':01',
        ('10.200.' || (n / 256) || '.' || (n % 256))::inet,
        'active',
        NOW() - (random() * INTERVAL '1 hour'),
        (random() * 500000000)::bigint,
        (random() * 100000000)::bigint
      FROM generate_series(1, 200) AS n
      ON CONFLICT DO NOTHING;

      INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
      VALUES (
        uuid_generate_v4(), '$DEMO_TENANT',
        'PPPoE session count critical — BEY-PPPOE-01',
        'Active sessions at 950+. Approaching maximum capacity of 1000.',
        'critical', 'active', 'router', 'e0000000-0000-0000-0000-000000000005', NOW(),
        '{\"active_sessions\": 950, \"max_sessions\": 1000}'
      );
    "
    echo "Done. PPPoE sessions near capacity."
    ;;

  *)
    echo "Unknown scenario: $SCENARIO"
    echo ""
    print_usage
    exit 1
    ;;
esac
