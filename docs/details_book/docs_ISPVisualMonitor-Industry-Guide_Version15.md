# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XIV
Earlier parts cover the full platform blueprint. This part translates the blueprint into concrete inventory extensions (BGP peers, OSPF neighbors, probe targets), rule schema validation strategy, and a recommended internal module layout for both repos.

---

# Part XV — Inventory Extensions (Peers/Probes), Rule Schema Validation, and Repo Module Layout

## 108. Inventory extensions: BGP peers, OSPF neighbors, and probe targets

Earlier we recommended modeling peers/neighbors as first-class entities to avoid high-cardinality labels. This section outlines what you should store.

### 108.1 BGP peer entity model

#### 108.1.1 Why a peer entity
- one device can have many peers
- you want stable identifiers for rules and dashboards
- you want metadata like remote AS, peer role (upstream/customer/ix)

#### 108.1.2 Suggested table: `bgp_peers`
Fields:
- `id` (uuid)
- `tenant_id`
- `device_id` (router that owns this peer)
- `name` (RouterOS peer name)
- `remote_address` (inet)
- `local_as` (int)
- `remote_as` (int)
- `peer_type` (enum):
  - `upstream`
  - `customer`
  - `ix_peer`
  - `internal` (iBGP)
- `description` (optional)
- `site_id` (optional, derived from device)
- `enabled` (bool)
- `last_seen_at` (timestamptz)

Constraints:
- unique `(device_id, name)` or `(device_id, remote_address, remote_as)`

#### 108.1.3 Peer discovery and drift
Agent discovers peers via RouterOS API:
- insert new peers (inventory_patch)
- update changed metadata
- mark missing peers as “not seen” after N intervals (don’t delete immediately)

This prevents churn from temporary config changes.

### 108.2 OSPF neighbor entity model

#### 108.2.1 Suggested table: `ospf_neighbors`
Fields:
- `id` (uuid)
- `tenant_id`
- `device_id`
- `neighbor_router_id` (inet) — OSPF router-id
- `neighbor_ip` (inet) — adjacency IP (if available)
- `interface_name` (text)
- `area` (text or int)
- `enabled` (bool)
- `last_seen_at`

Constraints:
- unique `(device_id, neighbor_router_id, area)`

#### 108.2.2 Why store interface_name
This helps correlation:
- interface down → neighbor down on same interface
- increases RCA evidence confidence

### 108.3 Probe targets entity model

#### 108.3.1 Why probe targets matter
SLA and “experience monitoring” requires:
- stable targets
- per-site or per-agent target assignment
- classification (dns, gateway, upstream, public)

#### 108.3.2 Suggested tables
- `probe_targets`
  - `id`
  - `tenant_id`
  - `name`
  - `target_type` (enum):
    - `router_loopback`
    - `site_gateway`
    - `dns`
    - `upstream_test`
    - `public_internet`
    - `customer_sample`
  - `address` (inet or hostname)
  - `port` (nullable, for tcp/http)
  - `protocol` (icmp/tcp/http)
  - `enabled`
  - `created_at`

- `probe_assignments`
  - `probe_target_id`
  - `agent_id` (or `site_id`)
  - `interval_seconds`
  - `timeout_ms`
  - `packet_count`
  - `created_at`

This makes probes configurable without redeploying agents.

#### 108.3.3 Probe-to-incident semantics
When probe results degrade:
- create events with:
  - entity_type = probe_target
  - entity_id = probe_target_id
- rules evaluate probe metrics and correlate with topology.

---

## 109. Rule schema validation and testing (how to prevent bad rules)

### 109.1 Why validation is required
In production, a broken rule can:
- create alert storms
- miss outages
- crash evaluation workers due to invalid expressions

### 109.2 Strategy: JSON schema + compile step
1. Define a JSON Schema for rule definitions.
2. On rule creation/update:
   - validate JSON schema
   - “compile” into an internal representation (AST)
   - store compiled form (optional) or store validation checksum
3. Provide a “dry run” mode:
   - evaluate rule over last 1h without sending notifications
   - show how many entities would fire

### 109.3 Unit tests for rule evaluation
Build a suite:
- “given these metric samples, this rule must fire”
- “given jitter, must not flap”
- “cooldown prevents repeated notifications”

### 109.4 Regression fixtures
Keep fixture datasets:
- small timeseries samples for each rule type
- incident correlation fixtures (burst scenarios)

This ensures refactors don’t break alert correctness.

---

## 110. Schema validation for ingest payloads (agent → backend)

### 110.1 Contract testing
To keep agent and backend compatible:
- publish JSON schemas for ingest payload v1
- in agent CI:
  - validate produced payload against schema
- in backend CI:
  - validate sample payload fixtures

### 110.2 Compatibility tests
Implement a simple “golden payload” test:
- for each agent version supported:
  - ingest sample payload into backend test instance
  - ensure acceptance and correct storage

This builds confidence for upgrades.

---

## 111. Repo module layout recommendations (Control plane repo)

This section proposes an internal structure that matches the document’s architecture.

### 111.1 Control plane (Go) — suggested packages
- `internal/platform/`
  - config loader, env parsing
  - logger
  - HTTP server setup
  - DB connection + migrations runner
  - tracing/metrics helpers

- `internal/auth/`
  - JWT/OIDC
  - RBAC checks
  - middleware

- `internal/tenancy/`
  - tenant context propagation
  - tenant quotas
  - tenant settings (retention, limits)

- `internal/inventory/`
  - devices, sites, interfaces
  - credential references
  - bgp_peers, ospf_neighbors, probe_targets
  - repositories + services + handlers

- `internal/ingest/`
  - ingest handler
  - payload validation
  - write path (direct DB or queue)
  - idempotency / dedup helpers

- `internal/metrics/`
  - storage interface for samples (Timescale)
  - rollup jobs management

- `internal/alerts/`
  - rule definitions
  - evaluator
  - state machine for alerts

- `internal/incidents/`
  - correlation engine
  - RCA integration
  - incident lifecycle

- `internal/topology/`
  - topology_edges
  - dependency groups
  - inference proposals

- `internal/notifications/`
  - email, telegram, webhook senders
  - retry + DLQ
  - templates

- `internal/reporting/`
  - SLA reports
  - export scheduling

- `internal/diagnostics/`
  - jobs dispatch to agents
  - results storage
  - audit logging

### 111.2 Frontend (TypeScript)
Recommended structure:
- `src/pages/` (routes)
- `src/components/`
- `src/features/` (map, incidents, inventory, reports)
- `src/api/` (client + types)
- `src/state/` (redux/zustand/query)
- `src/styles/`

### 111.3 API definition discipline
- generate OpenAPI for UI API
- keep ingest schema in `api/ingest/` as JSON schema or protobuf

---

## 112. Repo module layout recommendations (Agent repo)

### 112.1 Agent core packages
- `internal/agent/`
  - startup
  - config manager
  - enrollment/mTLS bootstrap
  - heartbeat

- `internal/scheduler/`
  - job definitions
  - jitter logic
  - concurrency limits
  - backoff/circuit breaker

- `internal/collectors/routeros/`
  - RouterOS API client wrapper
  - pollers: device, interfaces, bgp, ospf, pppoe

- `internal/collectors/snmp/` (phase 2)
- `internal/collectors/syslog/` (phase 3)

- `internal/probes/`
  - ping
  - traceroute
  - tcp/http checks

- `internal/normalize/`
  - canonical metric mapping
  - canonical event mapping
  - registry references

- `internal/buffer/`
  - disk queue
  - retry semantics

- `internal/transport/`
  - ingest client
  - batching
  - compression
  - mTLS config

- `internal/telemetry/`
  - agent self-metrics (Prometheus)
  - logs

### 112.2 Agent configuration model
Agent should support:
- static config file (self-hosted simple mode)
- dynamic config from backend (preferred)
- a safe merge strategy:
  - local overrides + server assignments

---

## 113. Acceptance criteria for “repo aligned with architecture”
When your repos match the guide, you should be able to say:

- The agent can output canonical metrics/events and validate against ingest schema.
- The backend can ingest and store with tenant isolation.
- Rule definitions are validated, compiled, and tested.
- RCA uses dependency groups and evidence scoring.
- UI can show map, graph, incident evidence and timelines.
- Deployment has a clear compose reference, backup/restore and upgrade docs.

---

## 114. “Continue” note
Next chapters can provide:
- example JSON schemas (rule schema and ingest schema) in detail
- example OpenAPI endpoint catalog for UI API
- concrete polling profiles configuration format
- a “capability negotiation” protocol between agent and backend
- detailed operational pitfalls and how to avoid them (DB locks, retention jobs, etc.)

Say **continue** to proceed.