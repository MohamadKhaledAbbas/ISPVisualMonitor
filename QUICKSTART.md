# Quick Start Guide

## GitHub Codespaces / Dev Containers

If you're using GitHub Codespaces or VS Code Dev Containers and see "docker: command not found":

**Supported path**:
```bash
make demo-start  # Automatically detects environment
```

If you are in a recovery container, rebuild first so the sidecar services start correctly:
- Press `Cmd/Ctrl + Shift + P`
- Select `Codespaces: Rebuild Container`

See [CODESPACES.md](CODESPACES.md) for detailed troubleshooting.

---

## Starting Services

The ISP Visual Monitor requires PostgreSQL and other services to run. Here's how to start them:

### Option 1: Start Everything (Recommended for Demo)
```bash
make demo-start
```

This will:
- Start PostgreSQL, Redis, and all backend services
- Load demo data into the database
- Start both API server and frontend

### Option 2: Start Infrastructure Only
```bash
make docker-up
```

This starts PostgreSQL and Redis, then you can run the API and frontend separately.

### Option 3: Using Docker Compose Directly
```bash
docker compose up -d postgres redis
```

## Verifying Services

Check if PostgreSQL is running:
```bash
# Method 1: Using psql
psql -h localhost -U ispmonitor -d ispmonitor -c "SELECT version();"

# Method 2: Using nc/telnet
nc -zv localhost 5432
```

## Running Demo Scenarios

After services are running, you can run demo scenarios:

```bash
# Reset all routers to healthy state
./scripts/demo-scenarios.sh healthy

# Simulate a router going offline  
./scripts/demo-scenarios.sh router-offline

# List all available scenarios
./scripts/demo-scenarios.sh
```

## Environment Variables

If you're having connection issues, you may need to set:

```bash
export DB_HOST=postgres      # For devcontainer/docker network
export DB_HOST=localhost     # For local development
```

## Troubleshooting

### "Cannot connect to PostgreSQL"
- **Cause**: Services aren't running
- **Fix**: Run `make docker-up` or `make demo-start`

### "psql: command not found"
- **Cause**: PostgreSQL client not installed
- **Fix**: Already fixed! Run `sudo apk add postgresql-client`

### "docker: command not found" in demo scripts
- **Cause**: You are likely in a recovery container instead of the normal devcontainer
- **Fix**: Rebuild the container, then run `make demo-start`
