#!/bin/bash
# Post-create script for Codespaces / Dev Container
# This runs once after the container is created.

set -e

echo "========================================"
echo "  ISPVisualMonitor — Post-Create Setup"
echo "========================================"

cd /workspace

# 1. Install Go dependencies
echo "[1/5] Installing Go dependencies..."
go mod download

# 2. Install frontend dependencies
echo "[2/5] Installing frontend dependencies..."
cd web && npm ci && cd ..

# 3. Wait for PostgreSQL
echo "[3/5] Waiting for PostgreSQL..."
for i in $(seq 1 30); do
  if pg_isready -h postgres -U ispmonitor -q 2>/dev/null; then
    echo "  PostgreSQL is ready."
    break
  fi
  sleep 1
  if [ $i -eq 30 ]; then
    echo "  Warning: PostgreSQL may not be ready yet. Run 'make demo-seed' manually."
  fi
done

# 4. Load demo seed data
echo "[4/5] Loading demo seed data..."
if PGPASSWORD=ispmonitor psql -h postgres -U ispmonitor -d ispmonitor -f db/seed/demo_seed.sql; then
  echo "  Demo seed data loaded."
else
  echo "  Warning: Could not load seed data. Run 'make demo-seed' after services are up."
fi

# 5. Print instructions
echo "[5/5] Setup complete!"
echo ""
echo "========================================"
echo "  Quick Start"
echo "========================================"
echo ""
echo "  Start the app:"
echo "    make demo-start"
echo ""
echo "  Or start services individually:"
echo "    go run ./cmd/ispmonitor     # API on :8080"
echo "    cd web && npm run dev       # Frontend on :5173"
echo ""
echo "  Reset demo data:"
echo "    make demo-reset"
echo ""
echo "  Run demo scenarios:"
echo "    make demo-scenarios"
echo ""
echo "  Full docs: docs/DEMO.md"
echo "========================================"
