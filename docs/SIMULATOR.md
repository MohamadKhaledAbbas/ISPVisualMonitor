# Simulator — Developer Guide

The ISPVisualMonitor simulator generates realistic ISP telemetry inside the
repository so dashboards, alerting, maps, and incidents can be exercised
without a real agent or MikroTik devices.

---

## Quick Start

### Option A — via the main app (recommended)

```bash
# Start the full stack in demo mode (simulator auto-starts)
make demo-start
```

Under the hood this sets `ENABLE_SIMULATOR=true` which causes the main
`cmd/ispmonitor` binary to launch the simulator as a background goroutine.

### Option B — standalone simulator binary

```bash
# Requires database to be running (docker compose up -d postgres)
go run ./cmd/simulator                  # starts with default (healthy)
go run ./cmd/simulator router-down      # starts with a scenario
```

### Environment Variables

| Variable         | Default     | Description                                |
|------------------|-------------|--------------------------------------------|
| `SIM_MODE`       | `scenario`  | `deterministic`, `scenario`, or `random`   |
| `SIM_SEED`       | `42`        | RNG seed (deterministic & scenario modes)  |
| `SIM_INTERVAL`   | `30s`       | Telemetry generation interval              |
| `SIM_SCENARIO`   | `healthy`   | Initial active scenario                    |
| `DB_HOST`        | `localhost` | PostgreSQL host                            |
| `DB_PORT`        | `5432`      | PostgreSQL port                            |
| `DB_USER`        | `ispmonitor`| PostgreSQL user                            |
| `DB_PASSWORD`    | `ispmonitor`| PostgreSQL password                        |
| `DB_NAME`        | `ispmonitor`| PostgreSQL database                        |

---

## Reference Topology

The simulator ships with a built-in ISP topology that matches the demo seed
data (`db/seed/demo_seed.sql`):

| Device            | Role     | Site     | Interfaces |
|-------------------|----------|----------|------------|
| BEY-CORE-01       | core     | BEY-DC1  | 3          |
| BEY-CORE-02       | core     | BEY-DC1  | 3          |
| BEY-EDGE-01       | edge     | BEY-DC1  | 3          |
| BEY-UPSTREAM-01   | upstream | BEY-DC1  | 3          |
| BEY-PPPOE-01      | pppoe    | BEY-DC1  | 2          |
| TRI-EDGE-01       | edge     | TRI-POP1 | 2          |
| TRI-ACCESS-01     | access   | TRI-POP1 | 2          |
| TRI-PPPOE-01      | pppoe    | TRI-POP1 | 2          |
| SID-EDGE-01       | edge     | SID-POP1 | 2          |
| SID-PPPOE-01      | pppoe    | SID-POP1 | 2          |

**10 devices · 24 interfaces · 9 links · 3 sites**

---

## Telemetry Generated

Each generation cycle writes to the same tables the real poller would:

| Table               | What                                            |
|---------------------|-------------------------------------------------|
| `router_metrics`    | CPU %, memory %, uptime, temperature per router  |
| `interface_metrics`  | In/out octets, packets, errors, discards, util % |
| `routers`           | Status updates (`active`, `offline`)             |
| `interfaces`        | Status updates (`up`, `down`)                    |
| `alerts`            | Scenario-driven alert insertion                  |

---

## Simulation Modes

### Deterministic (`SIM_MODE=deterministic`)
Uses a fixed RNG seed so every run produces the exact same sequence of
metrics. Ideal for automated tests and regression testing.

### Scenario (`SIM_MODE=scenario`)
Uses the seed for reproducibility but applies the active scenario's
overrides. This is the default mode.

### Random (`SIM_MODE=random`)
Uses a time-based seed so output varies each run. Useful for exploratory
testing and realistic long-running demos.

---

## Scenarios

| Name                  | Effect                                                    |
|-----------------------|-----------------------------------------------------------|
| `healthy`             | All routers active, all interfaces up, no alerts          |
| `router-down`         | TRI-ACCESS-01 goes offline, interfaces down, critical alert |
| `link-saturation`     | Core uplink at 93 % utilization, CPU spike, warning alert |
| `upstream-outage`     | ISP-B link goes down, failover alert                      |
| `packet-loss`         | Tripoli uplink error rate spikes, warning alert           |
| `session-spike`       | BEY-PPPOE-01 sessions jump to 950/1000, critical alert   |
| `flapping-interface`  | TRI-ACCESS-01 ether1 toggles up/down each cycle           |

### Triggering a Scenario

**Standalone binary:**

```bash
go run ./cmd/simulator router-down
```

**Environment variable:**

```bash
SIM_SCENARIO=upstream-outage go run ./cmd/simulator
```

### Resetting to Healthy

The simulator exposes a `ResetToHealthy` method that:
1. Switches the active scenario back to `healthy`
2. Resets the RNG (in deterministic mode)
3. Restores all routers to `active`, interfaces to `up`, and resolves all alerts

---

## Architecture

```
cmd/simulator/main.go          Standalone entry point
cmd/ispmonitor/main.go          Integration point (ENABLE_SIMULATOR=true)

internal/simulator/
  simulator.go                  Service orchestration, cycle loop
  topology.go                   Reference ISP topology (10 devices, 9 links)
  telemetry.go                  Metric generation (router, interface, PPPoE)
  scenario.go                   Scenario engine + 7 named scenarios
  writer.go                     Database writer (same tables as real poller)
  simulator_test.go             Unit tests (18 tests)

configs/simulator.yaml          Configuration template
```

---

## Validation Checklist

- [ ] `make demo-start` — simulator auto-starts, logs show cycle output
- [ ] `go run ./cmd/simulator` — standalone binary runs and writes metrics
- [ ] Query `router_metrics` table — new rows appear every 30 s
- [ ] Query `interface_metrics` table — new rows appear every 30 s
- [ ] Trigger `router-down` — TRI-ACCESS-01 shows offline, alert created
- [ ] Trigger `healthy` — all routers return to active, alerts resolved
- [ ] Run with `SIM_MODE=deterministic SIM_SEED=42` twice — same metrics

---

## Files Reference

| File / Directory              | Purpose                              |
|-------------------------------|--------------------------------------|
| `internal/simulator/`         | Simulator module (all Go source)     |
| `cmd/simulator/`              | Standalone entry point               |
| `configs/simulator.yaml`      | Configuration template               |
| `docs/SIMULATOR.md`           | This document                        |
| `db/seed/demo_seed.sql`       | Demo topology seed data              |
| `scripts/demo-scenarios.sh`   | Shell-based scenario triggers        |
| `pkg/config/deployment.go`    | `ENABLE_SIMULATOR` flag definition   |
