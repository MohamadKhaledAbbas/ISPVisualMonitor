#!/bin/bash
# ISP Visual Monitor - Development Startup Script
# Starts all services needed for local development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  ISP Visual Monitor - Dev Startup"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

if ! command -v docker >/dev/null 2>&1; then
  echo -e "${YELLOW}Docker CLI not found in this container.${NC}"
  echo "Falling back to devcontainer startup mode..."
  exec bash "$SCRIPT_DIR/devcontainer-start.sh"
fi

# Step 1: Start infrastructure via Docker
echo -e "${YELLOW}[1/4] Starting infrastructure (PostgreSQL, Redis, Prometheus, Grafana)...${NC}"
cd "$PROJECT_DIR"
docker compose up -d postgres redis prometheus grafana 2>&1 | tail -5
echo ""

# Step 2: Wait for PostgreSQL to be healthy
echo -e "${YELLOW}[2/4] Waiting for PostgreSQL to be ready...${NC}"
for i in $(seq 1 30); do
  if docker exec ispmonitor-postgres pg_isready -U ispmonitor -q 2>/dev/null; then
    echo -e "${GREEN}  PostgreSQL is ready!${NC}"
    break
  fi
  sleep 1
  if [ $i -eq 30 ]; then
    echo -e "${RED}  PostgreSQL failed to start!${NC}"
    exit 1
  fi
done
echo ""

# Step 3: Start Go API server
echo -e "${YELLOW}[3/4] Starting Go API server on port 8080...${NC}"
# Kill existing API process if running
pkill -f "go run ./cmd/ispmonitor" 2>/dev/null || true
sleep 1

cd "$PROJECT_DIR"
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ispmonitor
export DB_PASSWORD=ispmonitor
export DB_NAME=ispmonitor
export DB_SSLMODE=disable
export REDIS_URL=redis://localhost:6379
export JWT_SECRET=dev-secret-key-change-in-production
export ALLOWED_ORIGINS='*'
export LOG_LEVEL=debug

nohup go run ./cmd/ispmonitor > /tmp/ispmonitor-api.log 2>&1 &
API_PID=$!
echo "  API PID: $API_PID (logs: /tmp/ispmonitor-api.log)"

# Wait for port 8080 to be listening (initial startup)
echo "  Waiting for API server to start listening..."
for i in $(seq 1 10); do
  if lsof -i :8080 > /dev/null 2>&1; then
    echo "  API server is listening on port 8080"
    break
  fi
  sleep 1
  if [ $i -eq 10 ]; then
    echo -e "${RED}  API server failed to start listening! Check /tmp/ispmonitor-api.log${NC}"
    exit 1
  fi
done

# Wait for API to be ready
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

# Step 4: Start Vite frontend dev server
echo -e "${YELLOW}[4/4] Starting frontend dev server...${NC}"
# Kill existing Vite process if running
pkill -f "vite" 2>/dev/null || true
sleep 1

cd "$PROJECT_DIR/web"
nohup npx vite --host 0.0.0.0 > /tmp/ispmonitor-frontend.log 2>&1 &
FRONTEND_PID=$!
echo "  Frontend PID: $FRONTEND_PID (logs: /tmp/ispmonitor-frontend.log)"

# Wait for frontend
sleep 3
FRONTEND_PORT=$(grep -oP 'localhost:\K[0-9]+' /tmp/ispmonitor-frontend.log | head -1)
if [ -z "$FRONTEND_PORT" ]; then
  FRONTEND_PORT="5173"
fi

echo ""
echo "========================================"
echo -e "${GREEN}  All services started successfully!${NC}"
echo "========================================"
echo ""
echo "  Services:"
echo "  ─────────────────────────────────────"
echo -e "  ${GREEN}Frontend (App):${NC}     http://localhost:${FRONTEND_PORT}"
echo -e "  ${GREEN}Go API:${NC}             http://localhost:8080"
echo -e "  ${GREEN}Prometheus:${NC}         http://localhost:9090"
echo -e "  ${GREEN}Grafana:${NC}            http://localhost:3000"
echo -e "  ${GREEN}PostgreSQL:${NC}         localhost:5432"
echo -e "  ${GREEN}Redis:${NC}              localhost:6379"
echo ""
echo "  Login Credentials:"
echo "  ─────────────────────────────────────"
echo "  Email:     demo@lebanonnet.demo"
echo "  Password:  password"
echo ""
echo "  Useful commands:"
echo "  ─────────────────────────────────────"
echo "  Logs (API):       tail -f /tmp/ispmonitor-api.log"
echo "  Logs (Frontend):  tail -f /tmp/ispmonitor-frontend.log"
echo "  Stop all:         bash scripts/dev-stop.sh"
echo ""
