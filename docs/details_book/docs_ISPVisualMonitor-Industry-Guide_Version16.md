# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XV
Earlier parts define the complete system blueprint and repo alignment. This part adds concrete schema examples (ingest + rules), an API endpoint catalog for the UI, polling profiles, and agent/backend capability negotiation.

---

# Part XVI — Concrete Schemas (Ingest + Rules), UI API Catalog, Polling Profiles, and Capability Negotiation

## 115. Ingest JSON Schema (v1) — concrete example (illustrative)

> This is not a full JSON Schema draft with `$schema` and definitions, but a structured example you can translate into a real schema. The goal is to make the contract explicit.

### 115.1 `IngestBatch` shape
- `schema_version`: string (e.g., `"1.0"`)
- `sent_at`: RFC3339 timestamp
- `agent`:
  - `id`: uuid
  - `name`: string (optional)
  - `version`: string (semver)
  - `capabilities`: array of strings (enum)
- `metrics`: array of `MetricSample`
- `events`: array of `Event`
- `inventory_patch`: optional object

Validation rules:
- max batch size (bytes)
- max metrics per batch (e.g., 50k)
- max events per batch (e.g., 10k)
- timestamps cannot be too far in the future (clock skew guard)

### 115.2 `MetricSample` shape
- `ts`: RFC3339 timestamp
- `metric_name`: string (must be in registry or allow “unknown with warning”)
- `entity_type`: enum (`device`, `interface`, `site`, `service`, `bgp_peer`, `ospf_neighbor`, `probe_target`)
- `entity_id`: uuid
- one of:
  - `value_num`: number
  - `value_text`: string
- optional:
  - `unit`: string
  - `labels`: object (string->string), low cardinality
  - `source`: enum (`routeros_api`, `snmp`, `syslog_derived`, `probe`, `synthetic`)
  - `quality`: enum (`good`, `estimated`, `stale`)

Validation rules:
- numeric bounds for known metrics (optional)
- `labels` size limit
- entity_id must exist (or allow “unknown entity” with buffering; recommended to reject)

### 115.3 `Event` shape
- `id`: uuid
- `ts`: RFC3339 timestamp
- `event_type`: string (enum list)
- `severity`: enum (`info`, `warn`, `critical`)
- `entity_type`: enum
- `entity_id`: uuid
- `payload`: object
- optional:
  - `raw`: string (syslog line)
  - `correlation_hint`: object (optional)

Validation rules:
- `payload` size limit
- `raw` size limit (truncate safely)

### 115.4 `inventory_patch` shape (optional)
- `devices`: optional array of device updates
- `interfaces`: optional array of interface updates
- `bgp_peers`: optional array of peer updates
- `ospf_neighbors`: optional array of neighbor updates

Rule:
- inventory_patch must be bounded; avoid sending full inventory too often
- use “diff” semantics (only changed/new)

---

## 116. Rule schema (v1) — concrete example (illustrative)

### 116.1 Rule header
- `id`: uuid
- `name`: string
- `enabled`: bool
- `severity`: enum or severity mapping
- `scope`: selector object
- `condition`: condition object
- `for_seconds`: int
- `cooldown_seconds`: int
- `annotations`:
  - `summary_template`
  - `description`
  - `runbook_url`

### 116.2 Scope selectors (examples)
A selector object could allow one of:
- `by_device_ids`: array uuid
- `by_site_ids`: array uuid
- `by_tags`: array string
- `by_roles`: array string
- `by_entity_type`: enum
- combinations:
  - AND semantics across fields
  - OR semantics inside arrays

### 116.3 Condition types (starter set)
Implement a small set of condition types:

1. `threshold`
   - metric_name
   - operator: `>`, `>=`, `<`, `<=`, `==`, `!=`
   - threshold_num
   - aggregation: `avg`, `max`, `min`, `p95`
   - lookback_seconds

2. `enum_not_equal`
   - metric_name
   - bad_values: array string
   - lookback_seconds

3. `rate_of_change`
   - metric_name
   - operator
   - threshold_num
   - lookback_seconds

4. `composite`
   - op: `and|or`
   - conditions: array condition objects

### 116.4 Example rules (in your future docs)
- device down:
  - composite AND:
    - ping loss == 100% for 30s
    - api up == 0 for 30s (optional evidence)
- congestion:
  - AND:
    - utilization_p95 > 95% for 5m
    - rtt_p95 > baseline+X for 5m

**Guidance:** keep the evaluator deterministic and explainable.

---

## 117. UI API endpoint catalog (v1)

This is a recommended list of endpoints to keep the UI simple and stable.

### 117.1 Auth and identity
- `POST /api/v1/auth/login` (local auth)
- `POST /api/v1/auth/logout`
- `GET  /api/v1/me`
- `GET  /api/v1/audit-logs` (admin/auditor)

For OIDC:
- `GET /api/v1/auth/oidc/login`
- `GET /api/v1/auth/oidc/callback`

### 117.2 Inventory
- `GET /api/v1/sites`
- `POST /api/v1/sites`
- `GET /api/v1/devices`
- `POST /api/v1/devices`
- `GET /api/v1/devices/{id}`
- `PATCH /api/v1/devices/{id}`
- `GET /api/v1/devices/{id}/interfaces`
- `GET /api/v1/devices/{id}/bgp-peers`
- `GET /api/v1/devices/{id}/ospf-neighbors`

### 117.3 Topology
- `GET /api/v1/topology/edges`
- `POST /api/v1/topology/edges` (manual edge)
- `GET /api/v1/topology/dependency-groups`
- `POST /api/v1/topology/dependency-groups`
- `GET /api/v1/topology/proposals` (inferred edges pending approval)
- `POST /api/v1/topology/proposals/{id}/accept`
- `POST /api/v1/topology/proposals/{id}/reject`

### 117.4 Probes
- `GET /api/v1/probe-targets`
- `POST /api/v1/probe-targets`
- `GET /api/v1/probe-targets/{id}`
- `POST /api/v1/probe-assignments`

### 117.5 Metrics and dashboards
- `GET /api/v1/metrics/query` (time range + metric_name + entity)
- `GET /api/v1/metrics/summary` (site/device summary status)
- `GET /api/v1/dashboards/site/{id}`
- `GET /api/v1/dashboards/device/{id}`

**Important:** Provide server-side aggregation endpoints so the UI doesn’t run heavy queries.

### 117.6 Alert rules and alert instances
- `GET /api/v1/alert-rules`
- `POST /api/v1/alert-rules`
- `GET /api/v1/alerts` (instances)
- `POST /api/v1/alerts/{id}/ack`

### 117.7 Incidents
- `GET /api/v1/incidents`
- `GET /api/v1/incidents/{id}`
- `POST /api/v1/incidents/{id}/ack`
- `POST /api/v1/incidents/{id}/close`
- `POST /api/v1/incidents/{id}/notes`

### 117.8 Notifications
- `GET /api/v1/notification-channels`
- `POST /api/v1/notification-channels`
- `GET /api/v1/notification-policies`
- `POST /api/v1/notification-policies`
- `GET /api/v1/notification-deliveries` (admin)

### 117.9 Reporting
- `GET /api/v1/reports/sla?site_id=...&range=...`
- `POST /api/v1/reports/sla/schedule`
- `GET /api/v1/reports/exports`

### 117.10 Diagnostics
- `POST /api/v1/diagnostics/ping`
- `POST /api/v1/diagnostics/traceroute`
- `GET  /api/v1/diagnostics/runs/{id}`

Diagnostics should require elevated role and be audited.

---

## 118. Polling profiles configuration (agent + backend)

### 118.1 Why profiles are required
ISPs differ:
- small networks can poll more frequently
- large networks need conservative polling
- wireless PoPs may need different thresholds

Profiles let you tune without code changes.

### 118.2 Recommended profile model
- `polling_profiles`
  - `id`
  - `tenant_id`
  - `name`
  - `fast_interval_seconds`
  - `medium_interval_seconds`
  - `slow_interval_seconds`
  - `timeouts_ms` (connect/query)
  - `max_concurrency_global`
  - `max_concurrency_per_device`
  - `jitter_pct` (0–30%)
  - `enabled_collectors[]` (routeros_api, snmp, probes, syslog)
  - `enabled_metric_families[]` (interfaces, routing, system, services)

- `device_polling_profile`
  - `device_id`
  - `polling_profile_id`

### 118.3 Suggested defaults
- fast: 10–30s
- medium: 60–120s
- slow: 5–15m
- jitter: 10–20%
- connect timeout: 3–5s
- query timeout: 5–10s

### 118.4 Adaptive polling (advanced but valuable)
If device is flapping or overloaded:
- automatically reduce poll frequency
- open circuit breaker
- raise an event “monitoring degraded”

---

## 119. Capability negotiation (agent ↔ backend)

### 119.1 Why you need negotiation
Agents may differ:
- some support SNMP, some don’t
- some can run traceroute (Linux), some can’t
- some are placed in PoPs with special reachability

Backend needs to know:
- which agent can monitor which devices
- which probes can be executed where

### 119.2 Agent capabilities model
Agent reports:
- collector support: `routeros_api`, `snmp`, `syslog`
- probe support: `icmp_ping`, `traceroute`, `tcp_check`, `http_check`
- constraints:
  - max batch size it can send
  - max probes per second it can run
- network hints (optional, careful):
  - site affinity
  - reachable CIDRs (do not expose sensitive topology unnecessarily)

### 119.3 Assignment algorithm (v1)
Backend assigns devices to agents using:
- site affinity (same site preferred)
- capability match (routeros_api required initially)
- agent health (last seen, buffer depth)
- load balancing (max devices per agent)

### 119.4 Heartbeat endpoint concept
- agent sends `POST /ingest/v1/heartbeat`
  - includes capabilities and health metrics
- backend updates:
  - last_seen
  - status
  - agent load

This supports HA and monitoring of collectors themselves.

---

## 120. “Continue” note
Next chapters can include:
- operational pitfalls and “gotchas” for Timescale hypertables, indexes, compression
- best practices for avoiding slow queries in dashboards
- a concrete “seed data” approach for demos (sites/devices/links)
- a full example OpenAPI sketch or protobuf for ingest (if you switch to gRPC)
- recommendations for CI/CD and release engineering (agent signing, server images, SBOM)

Say **continue** to proceed.