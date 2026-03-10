#!/bin/bash
# ISP Visual Monitor - Devcontainer Startup Script
# For use inside devcontainers where Docker is not available
# Services (postgres, redis) should come from the devcontainer sidecars or a host Docker stack

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  ISP Visual Monitor - Devcontainer Mode"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

find_service_host() {
  local command_name="$1"
  shift

  local host
  for host in "$@"; do
    case "$command_name" in
      psql)
        if PGPASSWORD=ispmonitor psql -h "$host" -U ispmonitor -d ispmonitor -c '\q' >/dev/null 2>&1; then
          echo "$host"
          return 0
        fi
        ;;
      redis)
        if command -v redis-cli >/dev/null 2>&1; then
          if redis-cli -h "$host" ping >/dev/null 2>&1; then
            echo "$host"
            return 0
          fi
        elif command -v nc >/dev/null 2>&1; then
          if nc -z "$host" 6379 >/dev/null 2>&1; then
            echo "$host"
            return 0
          fi
        fi
        ;;
    esac
  done

  return 1
}

# Check if running in devcontainer
if [ -z "$REMOTE_CONTAINERS" ] && [ -z "$CODESPACES" ]; then
  echo -e "${YELLOW}Warning: This script is designed for devcontainer environments${NC}"
  echo "For local development, use: make demo-start"
  echo ""
fi

# Step 1: Check if PostgreSQL is accessible
echo -e "${YELLOW}[1/4] Checking PostgreSQL connection...${NC}"

DB_HOST="$(find_service_host psql postgres localhost 127.0.0.1 || true)"

if [ -z "$DB_HOST" ]; then
  echo -e "${RED}  PostgreSQL is not accessible!${NC}"
  echo ""
  if [ "${CODESPACES_RECOVERY_CONTAINER:-false}" = "true" ]; then
    echo -e "${YELLOW}Detected Codespaces recovery container.${NC}"
    echo "This mode does not run docker-compose sidecar services (postgres/redis)."
    echo ""
    echo "Fix:"
    echo "  1. Command Palette -> Codespaces: Rebuild Container"
    echo "  2. If it reopens in recovery again, run: Codespaces: View Creation Log"
    echo "  3. Share the creation log error so we can fix the devcontainer config"
    echo ""
    exit 1
  fi

  echo "The database services need to be started from outside this container."
  echo ""
  echo "If using GitHub Codespaces or VS Code Dev Containers:"
  echo "  1. Services should start automatically from the devcontainer compose setup"
  echo "  2. Check .devcontainer/docker-compose.devcontainer.yml"
  echo "  3. Try rebuilding the container"
  echo ""
  echo "If using local Docker:"
  echo "  1. From your host machine, run: docker compose up -d"
  echo "  2. Or start just postgres: docker compose up -d postgres redis"
  echo ""
  exit 1
fi

echo -e "${GREEN}  PostgreSQL found at: $DB_HOST${NC}"

REDIS_HOST="$(find_service_host redis redis localhost 127.0.0.1 || true)"
if [ -z "$REDIS_HOST" ]; then
  echo -e "${RED}  Redis is not accessible!${NC}"
  echo ""
  echo "The application expects Redis for cache and background coordination."
  echo ""
  echo "If using GitHub Codespaces or VS Code Dev Containers:"
  echo "  1. Command Palette -> Codespaces: Rebuild Container"
  echo "  2. Verify the normal devcontainer starts instead of the recovery container"
  echo ""
  echo "If using local Docker:"
  echo "  1. From your host machine, run: docker compose up -d postgres redis"
  echo ""
  exit 1
fi

echo -e "${GREEN}  Redis found at: $REDIS_HOST${NC}"

# Step 2: Load demo seed data if not already loaded
echo -e "${YELLOW}[2/4] Checking demo seed data...${NC}"
cd "$PROJECT_DIR"

TENANT_COUNT=$(PGPASSWORD=ispmonitor psql -h "$DB_HOST" -U ispmonitor -d ispmonitor -t -c "SELECT COUNT(*) FROM tenants;" 2>/dev/null | tr -d ' ')

if [ "$TENANT_COUNT" = "0" ]; then
  echo "  Loading demo seed data..."
  if PGPASSWORD=ispmonitor psql -h "$DB_HOST" -U ispmonitor -d ispmonitor -f db/seed/demo_seed.sql > /dev/null 2>&1; then
    echo -e "${GREEN}  Demo seed data loaded!${NC}"
  else
    echo -e "${YELLOW}  Warning: Could not load seed data${NC}"
  fi
else
  echo -e "${GREEN}  Demo data already loaded ($TENANT_COUNT tenants)${NC}"
fi
echo ""

# Step 3: Start Go API server
echo -e "${YELLOW}[3/4] Starting Go API server on port 8080...${NC}"

# Kill existing API process if running
pkill -f "go run ./cmd/ispmonitor" 2>/dev/null || true
sleep 1

cd "$PROJECT_DIR"

# Export environment variables
export DB_HOST="$DB_HOST"
export DB_PORT=5432
export DB_USER=ispmonitor
export DB_PASSWORD=ispmonitor
export DB_NAME=ispmonitor
export DB_SSLMODE=disable
export REDIS_URL="redis://$REDIS_HOST:6379"
export JWT_SECRET=dev-secret-key-change-in-production
export ALLOWED_ORIGINS='*'
export LOG_LEVEL=debug
export APP_MODE=${APP_MODE:-demo}
export ENABLE_SIMULATOR=${ENABLE_SIMULATOR:-true}
export ENABLE_REAL_AGENT=${ENABLE_REAL_AGENT:-false}
export USE_SEED_DATA=${USE_SEED_DATA:-true}
export BYPASS_LICENSE=${BYPASS_LICENSE:-true}

# Start API server in background
nohup go run ./cmd/ispmonitor > /tmp/ispmonitor-api.log 2>&1 &
API_PID=$!
echo "  API PID: $API_PID (logs: /tmp/ispmonitor-api.log)"

# Wait for API to be ready
echo "  Waiting for API server to start..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}  API server is ready!${NC}"
    break
  fi
  sleep 1
  if [ $i -eq 30 ]; then
    echo -e "${RED}  API server failed to start! Check /tmp/ispmonitor-api.log${NC}"
    tail -20 /tmp/ispmonitor-api.log
    exit 1
  fi
done
echo ""

# Step 4: Start frontend dev server
echo -e "${YELLOW}[4/4] Starting frontend dev server on port 5173...${NC}"

# Kill existing frontend process if running
pkill -f "vite" 2>/dev/null || true
sleep 1

cd "$PROJECT_DIR/web"

# Start frontend in background
nohup npm run dev > /tmp/ispmonitor-frontend.log 2>&1 &
FRONTEND_PID=$!
echo "  Frontend PID: $FRONTEND_PID (logs: /tmp/ispmonitor-frontend.log)"

# Wait for frontend to be ready
echo "  Waiting for frontend server to start..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:5173 > /dev/null 2>&1; then
    echo -e "${GREEN}  Frontend server is ready!${NC}"
    break
  fi
  sleep 1
  if [ $i -eq 30 ]; then
    echo -e "${YELLOW}  Frontend server may still be starting...${NC}"
  fi
done
echo ""

echo -e "${GREEN}========================================"
echo "  ISP Visual Monitor is running!"
echo "========================================${NC}"
echo ""
echo "  Frontend:    http://localhost:5173"
echo "  API:         http://localhost:8080"
echo "  Health:      http://localhost:8080/health"
echo ""
echo "  API logs:    tail -f /tmp/ispmonitor-api.log"
echo "  Frontend:    tail -f /tmp/ispmonitor-frontend.log"
echo ""
echo "  Demo Login:"
echo "    Email:    admin@demo.local"
echo "    Password: demo123"
echo ""
echo "  Stop services: make demo-stop"
echo -e "${GREEN}========================================${NC}"
