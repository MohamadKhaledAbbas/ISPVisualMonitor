#!/bin/bash
# ISP Visual Monitor - Stop all development services

echo "Stopping ISP Visual Monitor services..."

# Stop Go API
pkill -f "go run ./cmd/ispmonitor" 2>/dev/null && echo "  Stopped: Go API" || echo "  Go API was not running"

# Stop Vite frontend
pkill -f "vite" 2>/dev/null && echo "  Stopped: Frontend" || echo "  Frontend was not running"

# Stop Docker services
cd "$(dirname "${BASH_SOURCE[0]}")/.."
docker compose down 2>&1 | tail -3 && echo "  Stopped: Docker services"

echo "All services stopped."
