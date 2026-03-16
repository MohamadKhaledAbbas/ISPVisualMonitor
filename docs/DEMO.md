# Demo Mode — Developer Guide

This document explains how to run ISPVisualMonitor in **demo mode**: a
self-contained environment that works without real MikroTik routers, a
production license, or an external agent.

---

## Quick Start (Codespaces)

1. **Open in Codespaces** — click the green *Code* button on GitHub and choose
   *Create codespace on main*.
2. Wait for the post-create script to finish (Go deps, npm install, seed data).
3. Run:

   ```bash
   make demo-start
   ```

4. Open the forwarded port **5173** (frontend) in your browser.
5. Log in with:

   | Field    | Value                     |
   |----------|---------------------------|
   | Email    | `demo@lebanonnet.demo`    |
   | Password | `password`                |

That's it — dashboard, map, alerts, and routers pages are populated.

---

## Quick Start (local)

```bash
# 1. Clone
git clone https://github.com/MohamadKhaledAbbas/ISPVisualMonitor.git
cd ISPVisualMonitor

# 2. Copy env and set JWT secret
cp .env.example .env
# edit .env — at minimum set JWT_SECRET to any non-default value

# 3. Start in demo mode
make demo-start
```

The command starts PostgreSQL + Redis via Docker Compose, launches the Go API
server, starts the Vite frontend dev server, and loads seed data.

---

## What Is Demo Mode?

Demo mode is controlled by these environment variables:

| Variable            | Default   | Demo Value | Description                          |
|---------------------|-----------|------------|--------------------------------------|
| `APP_MODE`          | `normal`  | `demo`     | Activates demo-mode behaviour        |
| `ENABLE_SIMULATOR`  | `false`   | `true`     | Enables simulated telemetry          |
| `ENABLE_REAL_AGENT` | `true`    | `false`    | Disables connections to real routers |
| `USE_SEED_DATA`     | `false`   | `true`     | Marks seed data as active            |
| `BYPASS_LICENSE`    | `false`   | `true`     | Skips license validation             |

All of these are defined in `pkg/config/deployment.go` and loaded from the
environment. In production, none of these flags should be set.

---

## Seed Data

The seed dataset (`db/seed/demo_seed.sql`) provides:

| Entity         | Count | Details                                        |
|----------------|------:|------------------------------------------------|
| Tenant         |     1 | LebanonNet ISP                                 |
| User           |     1 | demo@lebanonnet.demo / password                |
| Regions        |     3 | Beirut Metro, North Lebanon, South Lebanon     |
| Sites (POPs)   |     3 | BEY-DC1, TRI-POP1, SID-POP1                   |
| Routers        |    10 | core, edge, border/upstream, access, pppoe     |
| Interfaces     |   30+ | ethernet + optical per router                  |
| Links          |     9 | physical inter-router connections               |
| Alerts         |     7 | active / acknowledged / resolved mix            |
| Router metrics |  ~2 h | CPU, memory, uptime, temp (every 5 min)        |
| Interface metrics | ~2 h | octets, packets, errors (every 5 min)        |
| PPPoE sessions |    50 | active subscriber sessions                     |

### Load seed data manually

```bash
make demo-seed
# or
bash scripts/demo-seed.sh
```

### Reset to clean baseline

```bash
make demo-reset
# or
bash scripts/demo-reset.sh
```

This deletes all demo tenant data and re-runs the seed script.

---

## Demo Scenarios

Repeatable scenarios simulate real-world ISP events by modifying the database.

```bash
# List all scenarios
make demo-scenarios

# Run a specific scenario
bash scripts/demo-scenarios.sh <scenario>
```

| Scenario           | What it does                                         |
|--------------------|------------------------------------------------------|
| `healthy`          | Reset all routers to active, resolve all alerts      |
| `router-offline`   | Mark TRI-ACCESS-01 as offline + create critical alert|
| `core-congestion`  | Spike utilization on core uplink to ~93%             |
| `upstream-failure`  | Set upstream ISP-B interface to down                 |
| `packet-loss`      | Inject high error rates on Tripoli uplink            |
| `high-sessions`    | Push PPPoE sessions near 1000-session capacity       |

### Suggested demo flow

1. Start with `healthy` baseline.
2. Trigger `core-congestion` — observe warning alert and metric spike.
3. Trigger `upstream-failure` — observe critical alert.
4. Reset with `healthy`.
5. Trigger `router-offline` — show map marker going offline.
6. Trigger `high-sessions` — show PPPoE capacity warning.

---

## Port Forwarding (Codespaces)

| Port  | Service         | Auto-forward |
|------:|-----------------|:------------:|
|  5173 | Frontend (Vite) | ✅ (browser)  |
|  8080 | Go API server   | ✅            |
|  3000 | Grafana         | ✅            |
|  9090 | Prometheus      | ✅            |
|  5432 | PostgreSQL      | ❌            |
|  6379 | Redis           | ❌            |

Forwarded ports are defined in `.devcontainer/devcontainer.json`.

---

## Architecture in Demo Mode

```
┌──────────────────────┐     ┌───────────────┐
│  Frontend (Vite)     │────▶│  Go API :8080 │
│  :5173               │     │               │
└──────────────────────┘     └───────┬───────┘
                                     │
                              ┌──────┴──────┐
                              │ PostgreSQL   │  ◀── seed data
                              │ :5432        │
                              └──────┬──────┘
                              ┌──────┴──────┐
                              │ Redis :6379  │
                              └─────────────┘
```

In demo mode the **poller does not connect to real routers**. Telemetry comes
from seed data and scenario scripts.

---

## What Is Intentionally Stubbed or Bypassed

- **License validation** — skipped when `BYPASS_LICENSE=true`.
- **Real router polling** — disabled when `ENABLE_REAL_AGENT=false`. The
  poller service still starts but has no targets.
- **OIDC / external auth** — uses local auth provider only.
- **Email / webhook notifications** — not configured in demo; SMTP vars are
  left empty.

---

## Files Reference

| File/Directory | Purpose |
|---|---|
| `.devcontainer/` | Codespaces / Dev Container config |
| `db/seed/demo_seed.sql` | Comprehensive demo dataset |
| `scripts/demo-seed.sh` | Load seed data |
| `scripts/demo-reset.sh` | Reset demo to clean baseline |
| `scripts/demo-scenarios.sh` | Trigger demo scenarios |
| `pkg/config/deployment.go` | Demo mode env var config |
| `.env.example` | All env vars including demo mode |
| `docs/DEMO.md` | This document |
