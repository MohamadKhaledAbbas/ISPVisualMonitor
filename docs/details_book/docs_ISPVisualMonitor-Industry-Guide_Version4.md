# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## 0. Why this document exists

ISP operations require fast fault isolation, topology awareness, and low-noise alerting. Many monitoring systems fail for ISPs because they:
- provide device metrics but not **service impact**
- create alert storms with no correlation
- lack topology and **root cause analysis (RCA)**
- are hard to deploy in **self‑hosted** environments

This guide defines a production-grade architecture, data model, operational practices, and a concrete blueprint for building a trusted monitoring product.

---

## 1. Product vision & non-goals

### 1.1 Vision (what ISPs pay for)
A paid-grade monitoring platform should deliver:
1. **Fast detection** of outages and degradations (seconds to <1 minute)
2. **RCA and impact**: which upstream component explains the most failures
3. **SLA reporting**: per site, per link, per customer/service
4. **Noise control**: deduplication, flap detection, suppression via dependencies
5. **Operational workflows**: acknowledge, assign, annotate, export
6. **Trust and security**: open-source agent, audit logs, strong crypto, RBAC
7. **Hybrid deployment**: cloud-hosted or self-hosted with the same code

### 1.2 Non-goals for early versions (to keep velocity)
- Becoming a full NMS for *all* vendors at once
- Full SDN controller capabilities
- DPI / packet capture as a core feature
- Billing/CRM replacement

---

## 2. Design principles (industry-grade requirements)

### 2.1 Reliability > features
- A reliable “down/up + impact + SLA + RCA” system beats an unstable “kitchen sink” system.
- Every feature must come with:
  - failure modes
  - rate limits
  - testing strategy
  - rollback plan

### 2.2 Observe the observer
The monitoring system must expose metrics about itself:
- ingest lag
- agent heartbeat status
- queue depth
- alert engine throughput
- DB performance

### 2.3 Separation of control plane and data plane
- **Control plane** is the product UI, rules, inventory, topology, storage.
- **Data plane** is distributed agents doing collection and probing near routers.

This is mandatory for:
- private management networks
- PoP-local reachability
- cloud-hosted deployments
- resilience during WAN failures

### 2.4 Protocol plurality (don’t bet everything on one protocol)
Even if you start with RouterOS API:
- add **active probing** for SLA (ICMP/TCP/HTTP)
- strongly consider **SNMP** for lightweight counters and multi-vendor future
- add **syslog** for event timelines and fast signals

### 2.5 Topology as a first-class citizen
Without topology you cannot:
- suppress downstream noise
- compute impact
- do meaningful RCA
- build dependency graphs and maps that reflect reality

---

## 3. Reference architecture (MikroTik-first, extensible)

### 3.1 High-level components

**Control plane (ISPVisualMonitor):**
- Web UI (TypeScript)
- Backend API (Go)
- Inventory + Topology + Alerting + Reporting modules
- Storage:
  - Postgres (relational truth)
  - Time-series store (TimescaleDB recommended for SQL-first approach)
  - Optional queue (NATS / RabbitMQ / Redis Streams)

**Data plane (ISPVisualMonitor-Agent):**
- Collector plugins:
  - RouterOS API collector (primary)
  - SNMP collector (recommended)
  - Syslog receiver/forwarder (optional)
  - Flow collector (future)
- Probers:
  - ping
  - traceroute (on demand and/or sampled)
  - TCP/HTTP checks (optional)
- Buffering:
  - disk-backed queue for offline tolerance
- Transport:
  - HTTPS/gRPC + mTLS to control plane ingest

### 3.2 Why the agent is open source (trust model)
Open-sourcing the agent helps you:
- prove exactly what data is collected
- pass security review faster
- allow ISPs to fork and self-audit
- attract community contributions for new collectors

**Key rule:** Keep the **server** closed or open as you decide, but the **agent** should be independently buildable and verifiable (reproducible builds, signed releases).

---

## 4. Deployment models

### 4.1 Self-hosted (on-prem / ISP DC)
Recommended for many ISPs because:
- management networks are internal
- compliance and data locality
- easier router reachability

**Baseline stack:**
- Docker Compose (small/mid)
- Reverse proxy (Traefik/Nginx)
- Postgres + TimescaleDB
- Optional: NATS for buffering and fan-out

### 4.2 Cloud-hosted SaaS (you host)
Recommended when:
- you want centralized operations
- you target many tenants with a standard platform

**Baseline stack:**
- Kubernetes (or managed containers)
- managed Postgres/Timescale
- object storage for backups/exports
- per-tenant logical isolation and strong RBAC

### 4.3 Hybrid reality
Many ISPs will want:
- on-prem collectors (agents)
- cloud UI/reporting
- strict controls over what leaves their network

This architecture supports that.

---

## 5. ISP routing fundamentals (needed for topology + RCA)

This section is not “how to become an ISP”, but what your monitoring system must understand to be useful.

### 5.1 IGP vs EGP
- **IGP (Interior Gateway Protocol)**: inside one administrative domain (your ISP).
  - Example: OSPF, IS-IS
- **EGP (Exterior Gateway Protocol)**: between administrative domains (ASNs).
  - Example: BGP (eBGP)

In many ISPs:
- OSPF runs in the core and aggregation.
- BGP runs at edges (upstreams, peers, customers) and internally (iBGP).

### 5.2 OSPF in ISP networks (what it is, why it matters)
**OSPF** computes shortest paths within the ISP.
- A “neighbor” relationship (adjacency) must be established between routers.
- When a link fails, OSPF adjacency drops; routes recompute.

**Monitoring importance:**
- OSPF neighbor down is often an early and precise signal of a link/router issue.
- OSPF instability correlates with user-visible outages.

### 5.3 BGP (iBGP vs eBGP)
- **eBGP**: ISP ↔ upstream transit / ISP ↔ customer / ISP ↔ IX peer
- **iBGP**: inside the ISP to distribute external routes, often with **route reflectors**.

**Monitoring importance:**
- BGP session down can mean transit outage or misconfiguration.
- Prefix count changes can reveal upstream filtering/route leaks.
- Flaps cause churn and customer pain.

### 5.4 Implication for monitoring design
To claim ��RCA” and “impact,” your system must model:
- which routers are core vs edge
- which links are transit vs access
- which customers depend on which aggregation/core path

You can infer some relationships (neighbors), but operators must be able to override/label dependencies.

---

## 6. Telemetry methods (what to collect and why)

### 6.1 RouterOS API (primary for MikroTik)
**Best suited for:**
- Router identity (version, board name), uptime
- Interfaces list, running state
- Routing neighbors (BGP/OSPF), route counts
- PPPoE session counts, DHCP leases (if used)
- Queue statistics (if ISPs use QoS)
- Firewall counters (careful; can be heavy)

**Risks & mitigations:**
- **Chatty polling**: mitigate with batching, caching, and tiered schedules.
- **Security**: enforce TLS (8729), least privilege, IP restrictions.

### 6.2 SNMP (recommended)
**Best suited for:**
- standardized interface counters and errors
- CPU/memory/temperature sensors
- cross-vendor compatibility later

Prefer **SNMPv3** in production (auth + privacy).

### 6.3 Syslog (events)
**Best suited for:**
- event timelines: link up/down, auth failures, routing events
- “push” signals (less polling)

Requires careful parsing, normalization, and deduplication.

### 6.4 Active probing (mandatory for SLA and “real” monitoring)
Metrics are not enough. You need:
- ICMP ping (loss + RTT)
- TCP connect checks for services
- periodic and on-demand traceroute

**Key:** Multi-target probes help distinguish:
- last-mile/customer edge issue
- aggregation issue
- upstream/transit issue
- DNS/service-specific issue

---

## 7. Data plane agent (design blueprint)

### 7.1 Goals
- lightweight and safe for routers
- secure
- resilient when backend unreachable
- supports multiple collection plugins
- supports multi-PoP deployment

### 7.2 Internal modules (recommended)
- **Config manager**: device list, polling profiles, credentials references
- **Scheduler**: job orchestration with jitter, timeouts, concurrency caps
- **Collectors**:
  - routeros_api
  - snmp
  - syslog (optional)
- **Probers**: ping, traceroute, tcp/http
- **Normalizer**: converts vendor outputs into canonical metrics/events
- **Buffer**: disk queue for offline
- **Transport**: HTTPS/gRPC + mTLS, retry with backoff
- **Health endpoint**: agent self-metrics

### 7.3 Polling profiles (tiered collection)
Example (tune per ISP):
- **Fast (10–30s)**:
  - device reachability
  - BGP session state (edge/core)
  - critical interface running state
- **Medium (60–120s)**:
  - interface throughput counters
  - errors/drops
  - PPPoE session count (if used)
- **Slow (5–15m)**:
  - inventory refresh
  - sensor readings
- **Daily/weekly**:
  - config snapshot fingerprint (hash) for change detection

### 7.4 Concurrency, timeouts, and safety
- global concurrency limit (e.g., 50 jobs)
- per-device limit (e.g., 1–3)
- strict timeouts (e.g., 2–5s connect, 5–10s query)
- backoff on failures to avoid pounding a degraded router

### 7.5 Offline buffering
Agent must:
- buffer metrics/events locally when backend unreachable
- flush later with ordering and dedup keys
- expose “buffer depth” as a health metric

---

## 8. Control plane backend (design blueprint)

### 8.1 API separation
- **UI API**: auth, inventory, topology, dashboards, queries
- **Ingest API**: fast write path for metrics/events from agents

### 8.2 Suggested modules (start modular monolith)
- auth & RBAC
- tenancy (orgs)
- inventory (devices, tags, sites)
- topology graph
- ingest + validation
- alert engine + incidents
- notification channels
- reporting

---

## 9. Data model essentials (minimum production schema concepts)

### 9.1 Entities
- Organization (tenant / ISP)
- Site (PoP) with geo coordinates
- Device (router)
- Interface
- Link (physical or logical)
- Service/Customer (optional early, crucial later for PPPoE ISPs)
- Agent (collector identity)
- Metric samples (time-series)
- Events (discrete)
- Alerts (stateful)
- Incidents (correlated alerts)
- Maintenance windows
- Audit logs

### 9.2 Canonical metric types (examples)
- `device.up` (boolean)
- `ping.rtt_ms`, `ping.loss_pct`
- `interface.rx_bps`, `interface.tx_bps`
- `interface.errors`, `interface.drops`
- `bgp.session_state`, `bgp.prefixes_in`, `bgp.prefixes_out`
- `osfp.neighbor_state` (optional later)
- `pppoe.active_sessions` (if PPPoE)

### 9.3 Events vs metrics
- **Metrics**: continuous values over time (RTT, bps)
- **Events**: discrete facts (link down at 10:31:02)

Good incident timelines combine both.

---

## 10. Root Cause Analysis (RCA) — concept overview

### 10.1 Dependency graph
To do RCA, model dependencies:
- a customer depends on access router
- access depends on aggregation
- aggregation depends on core
- core depends on transit providers

### 10.2 Practical v1 RCA strategy
- detect which nodes/links are down
- compute impacted nodes by graph traversal
- root cause candidates are nodes whose failure explains the most downstream impact
- suppress downstream alerts if upstream cause is active, but show them as “impacted/suppressed” in UI

---

## 11. Alerting (what makes it paid-grade)

### 11.1 Stateful alerts (not just notifications)
Alerts should have state transitions:
- OK → DEGRADED → DOWN → RECOVERING → OK

### 11.2 Noise controls
- dedup (one alert per entity per type)
- flap detection
- maintenance windows
- dependency suppression
- correlation window

### 11.3 Impact-aware notifications
Every alert should include:
- what happened
- when it started
- likely root cause
- how many customers/services are impacted
- evidence (metrics/events)
- suggested next steps (run traceroute, check BGP peer, check optics)

---

## 12. Frontend (map + graph + investigation workflow)

### 12.1 Map view
- Sites as nodes with status
- Links between sites with utilization/health color
- Filters: tag, region, provider, service type

### 12.2 Dependency graph
- show physical and logical dependencies
- highlight incident blast radius
- show suppression reasoning

### 12.3 NOC workflow
- incident list
- acknowledge/assign
- timeline view
- drill down device/link
- run diagnostics via agent (safe/permissioned)

---

## 13. Security & compliance (minimum bar)

### 13.1 Agent identity and transport
- mTLS agent ↔ backend
- agent certificate rotation
- agent heartbeat and revocation

### 13.2 Credential storage
- encrypt at rest
- least privilege
- audit access to secrets
- allow on-prem integration with Vault later

### 13.3 Audit logs
Record:
- who changed alert rules
- who added devices
- who exported reports
- who ran diagnostics

---

## 14. Testing & simulation (EVE-NG + CHR)

EVE-NG + CHR is recommended for:
- realistic routing events
- controlled failure injection
- topology experiments

Build a lab:
- core/edge/customer routers
- Linux agent node
- Linux traffic endpoints
- optional FRR upstream

Test:
- link cuts
- BGP flap
- OSPF neighbor down
- congestion and packet loss
- partial outages (one PoP isolated)

---

## 15. Roadmap features that make ISPs “move up a level”

- SLA per customer/service (PPPoE-focused)
- capacity planning: 95th percentile, saturation prediction
- anomaly detection (baseline vs spike)
- config change detection (hash snapshots)
- root cause with evidence scoring
- playbooks / suggested actions
- integrations: Telegram, Slack, email, PagerDuty, webhooks
- customer portal (optional)
- flow-based insights (later)

---

## 16. How to use this document

This is Chapter 0–16 (foundation). Next chapters will turn this into a concrete implementation blueprint:
- exact service boundaries and module layout for Go
- recommended database schema (tables + indexes)
- ingest payload formats (metrics/events) and idempotency rules
- agent plugin interface design
- MikroTik RouterOS API collection plan (what to poll, how often, rate limits)
- RCA algorithm details (v1 and v2)
- deployment reference (Docker Compose profiles)
- operational runbooks (backup/restore, upgrades, migrations)

Say **“continue”** when ready.

---

# Part II — Implementation Blueprint (Concrete Design)

## 17. Target “MVP that feels paid-grade”

### 17.1 MVP scope (minimum that ISPs will still respect)
**You are MikroTik-first and API-first.** A credible MVP for ISP use typically includes:

1. **Inventory + site model**
   - Sites/PoPs with coordinates
   - Routers assigned to sites
   - Tags: `core`, `agg`, `access`, `edge`, `customer-edge`, `upstream`

2. **Core health**
   - Router reachability (agent ping or API connect)
   - Interface running status (critical uplinks)
   - BGP peer state (edge + route reflectors)
   - Active probes to key endpoints

3. **Map + topology**
   - Map of sites with status
   - Logical dependency graph (initially manual with optional inference)

4. **Alerting**
   - Router down / link down / BGP down / SLA breach (loss/latency)
   - Dedup + flap protection + maintenance windows
   - Basic RCA suppression (“downstream suppressed by upstream”)

5. **Incident workflow**
   - Incidents list
   - Acknowledge + notes
   - Timeline of events/metrics

6. **Reporting (basic)**
   - Uptime % per site/device (weekly/monthly)

**Strong recommendation:** Even if RouterOS API is your primary, include **active probes** in MVP. SLA without probing is not credible.

### 17.2 What to defer (but design for)
- PPPoE per-user status (valuable, but can explode scope)
- NetFlow/IPFIX
- Advanced anomaly detection (do basic thresholding first)
- Full auto-discovery of topology (start with semi-manual + inference hints)

---

## 18. Go backend: “modular monolith first” blueprint

The fastest path to industry-grade is a modular monolith that can later be split into services.

### 18.1 Suggested repository layout (conceptual)
A proven layout for Go systems:

- `cmd/ispvisualmonitor/` — main entrypoint
- `internal/` — non-exported application modules
  - `internal/auth/`
  - `internal/tenancy/`
  - `internal/inventory/`
  - `internal/topology/`
  - `internal/ingest/`
  - `internal/alerts/`
  - `internal/incidents/`
  - `internal/notifications/`
  - `internal/reporting/`
  - `internal/diagnostics/` (on-demand traceroute/ping via agent)
  - `internal/platform/` (logging, config, db, tracing)
- `pkg/` — shareable library code (use sparingly; keep core in internal)
- `api/` — OpenAPI/Swagger, protobuf definitions
- `web/` or `frontend/` — UI

**Goal:** avoid “god packages”. Each module owns:
- data model (domain types)
- storage interface (repository)
- service logic
- HTTP handlers

### 18.2 API boundaries
Define two major API groups:

1. **UI API**
   - `/api/v1/...`
   - Auth required (user tokens)
   - Pagination and filtering
   - Cached where possible

2. **Agent Ingest API**
   - `/ingest/v1/...`
   - Auth required (agent mTLS identity + signed token)
   - Very strict payload validation
   - Idempotent writes

### 18.3 Idempotency rules (critical)
Agents will retry. Backend must ensure:
- metrics can be inserted twice without breaking (time-series naturally tolerates duplicates if deduped by key)
- events can be deduped by a stable `event_id` generated at agent side
- alert engine is stateful and should handle repeated same input

**Recommended:**
- Each metric sample has:
  - `tenant_id`, `source_agent_id`, `metric_name`, `entity_id`, `timestamp`, `value`
- Each event has:
  - `event_id` (UUID), `tenant_id`, `timestamp`, `type`, `entity_id`, `payload`

Backend can maintain:
- a short-term `seen_event_ids` table/cache to drop duplicates.

---

## 19. Time-series storage choice (recommended approach)

### 19.1 TimescaleDB (recommended for your case)
Reasons:
- you already use Postgres/PLpgSQL in your repo composition
- easy SQL queries for reports and dashboards
- retention policies with hypertables
- continuous aggregates for rollups (hour/day)

**Industry pattern:**
- Postgres/Timescale for metrics + inventory
- optional Prometheus for system/agent metrics (internal observability)

### 19.2 Retention strategy (must be configurable)
ISPs care about retention and cost. Example:
- raw 10s metrics: retain 14–30 days
- 1m rollups: retain 6–12 months
- 1h rollups: retain 2–5 years (for SLA)

TimescaleDB can automate this with:
- compression
- retention jobs
- continuous aggregates

---

## 20. Agent: plugin model and job scheduling (concrete)

### 20.1 Plugin interface (conceptual requirements)
A collector plugin should provide:
- supported capabilities (routeros_api/snmp/syslog)
- discovery actions (optional)
- polling actions grouped by frequency class

**Collector output must be normalized** into:
- `MetricSample[]`
- `Event[]`
- optionally `InventoryPatch` (new interfaces, new peers, etc.)

### 20.2 Polling loops pattern
Use multiple schedules:
- `fast`: 10–30s
- `medium`: 60–120s
- `slow`: 5–15m

Each schedule:
- selects devices based on tags or profiles
- obeys concurrency limits
- adds jitter
- sends results in batches

### 20.3 “Batching” transport contract
Do not send one HTTP request per metric. Send batches:
- reduces overhead
- reduces TLS handshake cost
- improves backend write efficiency

A batch payload typically contains:
- agent identity info
- time window
- list of metric samples
- list of events

---

## 21. MikroTik RouterOS API collection plan (v1)

This section defines *what* you should query, not the exact code.

### 21.1 Collection categories
1. **Device identity & health**
   - uptime
   - RouterOS version
   - CPU load
   - memory usage
   - board model (for inventory)

2. **Interfaces**
   - list interfaces
   - running status
   - rx/tx rates (beware: rates can be derived from counters; choose method)
   - errors/drops (important for physical degradation)

3. **Routing neighbors**
   - BGP peers:
     - state
     - uptime
     - prefixes in/out
     - last error
   - OSPF neighbors:
     - adjacency state

4. **Services (optional for v1)**
   - PPPoE active sessions count
   - DHCP lease counts

### 21.2 Polling frequency recommendations
- BGP peer state: 10–30s (edge/core)
- Interface running status (uplinks): 10–30s
- Counters/errors: 60–120s
- Inventory refresh: 5–15m
- Config snapshot hash: daily

### 21.3 Safety rules for RouterOS API
- cap API sessions per router
- reuse sessions if safe and stable
- protect routers with timeouts and backoff
- if a router starts failing API calls, reduce poll rate automatically (“circuit breaker”)

---

## 22. Topology: start manual + inference (practical approach)

### 22.1 Why pure auto-discovery is hard
- ISPs have mixed designs, tunnels, L2 domains, MPLS, partial visibility
- LLDP is not always enabled
- management network may not see L2 neighbors

### 22.2 Recommended approach
**Phase 1 (MVP):** manual link definitions + tagging
- Operators define:
  - site nodes
  - site-to-site links (capacity, provider, medium)
  - device roles and uplinks
- Agent provides evidence (BGP/OSPF neighbor relationships)

**Phase 2:** inference helpers
- infer adjacency edges from OSPF/BGP neighbors
- infer “site dependency” from router role and neighbor graph
- allow operator review/approval before applying changes

---

## 23. Incident pipeline (how data becomes “actionable”)

### 23.1 Stages
1. ingest metrics/events
2. update entity states (device up/down, link health)
3. evaluate alert rules
4. dedup and correlate into incidents
5. compute RCA candidate(s)
6. compute impact set (customers/services)
7. notify (channels) + update UI

### 23.2 Incident data model essentials
Incident should store:
- start time, last update, end time
- current severity
- root cause candidate + confidence
- impacted entities count
- timeline references (events, metric points)
- acknowledgements & notes

---

# Part III — Storage, Schemas, and Ingest Contracts (Concrete)

## 25. Database architecture (Postgres + TimescaleDB)

### 25.1 Why a single Postgres cluster works well initially
For an early production system (even up to hundreds of routers), one Postgres cluster can host:
- relational truth (inventory, topology, users, incidents)
- time-series metrics (Timescale hypertables)
- materialized rollups for reporting

Benefits:
- one backup story
- one query language (SQL)
- easier self-hosted adoption

Scaling later:
- separate DBs (inventory vs time-series)
- read replicas
- sharding by tenant (advanced)

### 25.2 Extensions and features
Recommended Postgres/Timescale features:
- `timescaledb` extension
- `pgcrypto` (for some encryption primitives; not a full KMS replacement)
- `uuid-ossp` or application-generated UUIDs
- Partitioning / hypertables for metrics
- Row-level security (RLS) if you want stronger tenant boundaries (optional early)

### 25.3 Naming conventions
Choose consistent conventions:
- table names: snake_case plural (`devices`, `sites`, `metric_samples`)
- primary keys: UUID
- foreign keys: `<entity>_id`
- timestamps: `timestamptz` (`timestamp with time zone`)
- do not store local times; store UTC

---

## 26. Core relational schema (inventory, topology, tenants)

This section outlines the *conceptual schema*. You can implement with migrations (Goose, Atlas, Flyway, etc.)

### 26.1 Tenancy and auth
- `tenants`
  - `id` (uuid pk)
  - `name`
  - `created_at`
- `users`
  - `id`
  - `tenant_id`
  - `email`
  - `display_name`
  - `password_hash` (if local auth)
  - `created_at`, `disabled_at`
- `roles` / `user_roles`
  - RBAC mapping; keep simple early (admin, noc, read-only)

**Recommendation:** If you plan enterprise adoption, design for OIDC (Keycloak) early:
- keep `users` but allow external subject mapping:
  - `oidc_issuer`, `oidc_subject`

### 26.2 Sites (PoPs)
- `sites`
  - `id`
  - `tenant_id`
  - `name`
  - `latitude`, `longitude`
  - `address` (optional)
  - `tags` (optional normalized table or jsonb)

### 26.3 Devices
- `devices`
  - `id`
  - `tenant_id`
  - `site_id`
  - `name` (router identity)
  - `mgmt_ip` (inet)
  - `vendor` (`mikrotik` now)
  - `model`
  - `os_version`
  - `role` (core/edge/agg/access/customer-edge)
  - `enabled` (boolean)
  - `created_at`, `updated_at`

### 26.4 Device credentials (references, not plaintext)
A strong pattern is to store:
- device credential **references** in DB
- actual secrets encrypted and access-controlled

Tables:
- `credential_sets`
  - `id`
  - `tenant_id`
  - `name`
  - `type` (`routeros_api`, `snmpv3`, `ssh`)
  - `encrypted_blob` (bytea) — encrypted at rest using a master key
  - `created_at`

- `device_credentials`
  - `device_id`
  - `credential_set_id`
  - `priority` (if multiple creds)
  - `created_at`

**Note:** For MVP you can simplify, but document the “production recommended” approach.

### 26.5 Interfaces
You may store interfaces as discovered inventory:
- `device_interfaces`
  - `id`
  - `device_id`
  - `name` (ether1, sfp-sfpplus1, bridge1)
  - `type`
  - `mac_address` (optional)
  - `last_seen_at`
  - `admin_status` (enabled/disabled)
  - `oper_status` (running)
  - `speed_bps` (optional)
  - unique constraint: (`device_id`, `name`)

### 26.6 Links (physical/logical)
Start with manual links:
- `links`
  - `id`
  - `tenant_id`
  - `name`
  - `site_a_id`, `site_b_id`
  - `capacity_bps`
  - `provider` (optional)
  - `medium` (fiber, wireless, leased line)
  - `created_at`

And optionally endpoint mapping:
- `link_endpoints`
  - `link_id`
  - `device_interface_id`
  - `side` (`A`/`B`)

### 26.7 Topology graph edges (for RCA)
A flexible table:
- `topology_edges`
  - `id`
  - `tenant_id`
  - `from_entity_type` (device/site/service)
  - `from_entity_id`
  - `to_entity_type`
  - `to_entity_id`
  - `edge_type` (depends_on, connected_to, bgp_peer, ospf_neighbor)
  - `source` (manual, inferred, discovered)
  - `confidence` (0–1)
  - `created_at`, `updated_at`

This allows:
- manual dependencies
- inferred dependencies
- future auto-discovery improvements

---

## 27. Time-series schema (metrics) with TimescaleDB

### 27.1 Metric sample table (hypertable)
You need a canonical metric store. Recommended columns:

- `metric_samples`
  - `tenant_id` (uuid)
  - `ts` (timestamptz) — time of measurement
  - `metric_name` (text) — e.g. `ping.rtt_ms`
  - `entity_type` (text) — device/interface/site/service
  - `entity_id` (uuid)
  - `value_num` (double precision) — numeric values
  - `value_text` (text) — optional (for enums like BGP state)
  - `unit` (text) — optional; or infer from metric_name
  - `labels` (jsonb) — optional low-cardinality labels (avoid high-cardinality explosion)
  - `source_agent_id` (uuid)
  - `ingested_at` (timestamptz default now())

Make it a hypertable on `ts`.

### 27.2 Index strategy (important)
Indexes often needed:
- `(tenant_id, metric_name, ts DESC)`
- `(tenant_id, entity_type, entity_id, ts DESC)`
- optionally partial indexes for “hot” metrics

### 27.3 Rollups (continuous aggregates)
Define rollups:
- 1m rollup for graphs
- 1h/day rollup for SLA

Example rollup outputs:
- avg RTT, max RTT, loss percent
- avg rx/tx bps, p95
- uptime percent derived from `device.up`

### 27.4 Retention and compression
Policies:
- raw data retention 30 days
- 1m rollups 12 months
- 1h rollups 3–5 years

Enable compression on older chunks.

---

## 28. Events schema (discrete facts)

### 28.1 Canonical events table
- `events`
  - `id` (uuid) — event_id from agent
  - `tenant_id`
  - `ts` (timestamptz)
  - `event_type` (text) — link_down, bgp_down, device_reboot
  - `entity_type`, `entity_id`
  - `severity` (info/warn/critical)
  - `payload` (jsonb)
  - `source_agent_id`
  - `ingested_at`

### 28.2 Deduplication
- Primary key on `id` (UUID) makes events naturally idempotent.

---

## 29. Alerting schema (stateful)

### 29.1 Alert definitions (rules)
- `alert_rules`
  - `id`
  - `tenant_id`
  - `name`
  - `enabled`
  - `scope` (tags/sites/devices)
  - `condition` (DSL or JSON)
  - `severity`
  - `notification_policy_id`
  - `created_at`, `updated_at`

### 29.2 Alert instances (state machine)
- `alerts`
  - `id`
  - `tenant_id`
  - `rule_id`
  - `entity_type`, `entity_id`
  - `state` (ok/degraded/down/recovering)
  - `starts_at`
  - `last_state_change_at`
  - `last_evaluated_at`
  - `fingerprint` (dedup key)
  - `suppressed_by_incident_id` (nullable)
  - `details` (jsonb)

### 29.3 Incidents (correlated)
- `incidents`
  - `id`
  - `tenant_id`
  - `title`
  - `status` (open/closed)
  - `severity`
  - `root_cause_entity_type`, `root_cause_entity_id`
  - `root_cause_confidence`
  - `starts_at`, `ends_at`
  - `created_at`, `updated_at`

- `incident_alerts`
  - `incident_id`
  - `alert_id`

- `incident_notes`
  - `id`
  - `incident_id`
  - `author_user_id`
  - `note`
  - `created_at`

---

## 30. Ingest API contract (HTTP-first, production-safe)

### 30.1 Why define the contract early
If you define the ingest payloads now:
- you can evolve the backend and agent independently
- you can build test fixtures and replay traffic
- you can add other vendor agents later

### 30.2 Core design constraints
- batch-based
- compressible (gzip)
- versioned (v1)
- idempotent
- bounded (max payload size, max metrics per request)

### 30.3 Example ingest endpoints (conceptual)
- `POST /ingest/v1/batch`
  - contains:
    - `agent_id`
    - `tenant_id` (or inferred from agent identity)
    - `sent_at`
    - `metrics[]`
    - `events[]`
    - optional `inventory_patch`

### 30.4 Idempotency for metric samples
Metrics are tricky: you may not want a strict “unique” constraint on every sample (costly).
Practical approaches:
1. **Allow duplicates** and rely on aggregation queries (acceptable at small scale)
2. **Dedup in backend** using a short-lived cache keyed by:
   - (tenant, metric_name, entity_id, ts, value)
3. **Dedup in agent** by ensuring each sample time is unique per interval

In practice:
- keep it simple: avoid duplicates by design in agent,
- tolerate a small number of duplicates in storage.

### 30.5 Backpressure and rate limits
Backend must protect itself:
- limit requests per agent
- enforce payload limits
- reject with 429 + retry-after when overloaded
- agent should backoff and buffer

---

## 31. Audit logs (trust feature)

### 31.1 What to audit
- login events
- device additions/removals
- credential set created/updated
- alert rule changes
- incident acknowledgements and closes
- report exports
- diagnostics runs (traceroute/ping execution)

### 31.2 Table outline
- `audit_logs`
  - `id`
  - `tenant_id`
  - `actor_type` (user/agent/system)
  - `actor_id`
  - `action` (text)
  - `entity_type`, `entity_id`
  - `ts`
  - `details` (jsonb)

---

# Part IV — Alert Rules, RCA, PPPoE Impact, and “ISP-grade” Features

## 33. Alert rules engine design

### 33.1 Requirements for an ISP-grade alert engine
1. **Stateful** evaluation (alerts have life cycles)
2. **Hysteresis** (avoid flip-flopping due to jitter)
3. **Flap detection** and escalation
4. **Dependency suppression** (don’t spam downstream)
5. **Maintenance windows** (scheduled silencing)
6. **Per-tenant customization** (ISPs have different thresholds)
7. **Explainability** (why did this alert fire?)

### 33.2 Avoid writing “Prometheus clones” too early
You can build an industry-grade rule system without reinventing PromQL.

Recommended v1 approach:
- a small DSL or JSON condition model supporting common cases:
  - threshold rules
  - rate-of-change rules
  - boolean state rules
  - “for” duration (condition must hold for X time)

### 33.3 Rule condition model (practical JSON form)
Think of rules in this shape:

- **scope**: which entities
- **metric**: what measurement
- **operator**: `>`, `<`, `==`, `!=`, `in`
- **threshold**: numeric or enum
- **window**: lookback window for evaluation (e.g., 5m)
- **for**: how long it must be true before firing (e.g., 1m)
- **recover**: recover threshold/hysteresis (optional)
- **severity mapping**: warn/critical based on levels

Example rule types you need:
- interface down
- ping loss > X for Y
- ping RTT > X for Y
- BGP peer not established
- interface error rate above baseline

### 33.4 Key concept: “signal selection”
A big mistake is using one signal as truth. For example:
- Device “down” should not be only “API unreachable”
- combine:
  - ping fail
  - API fail
  - syslog last event
  - last seen time
into a robust decision.

### 33.5 Alert state machine (recommended)
For each alert instance (rule + entity):
- `OK`: condition false
- `PENDING`: condition true but not for long enough
- `FIRING`: condition true and “for” satisfied
- `RECOVERING`: condition false but within cool-down (optional)
- `SUPPRESSED`: firing, but muted by maintenance or dependency

In UI and notifications, present:
- a single incident, not 50 raw alerts.

---

## 34. Root Cause Analysis (RCA) in detail

### 34.1 RCA goals
When an outage occurs:
- identify the minimal set of upstream failures that explains the impact
- suppress noise from impacted downstream entities
- show blast radius in map/graph
- attach evidence for confidence

### 34.2 Inputs to RCA
RCA consumes:
- current alert states (down/degraded)
- topology edges (depends_on, connected_to)
- routing neighbor states (BGP/OSPF)
- active probe results
- syslog events

### 34.3 Graph model and edge semantics
Define edge semantics precisely:

- `connected_to`:
  - physical/logical adjacency
  - not always a dependency; can be redundant

- `depends_on`:
  - strict upstream dependency used for suppression/impact propagation
  - example:
    - access router depends_on aggregation router
    - customer service depends_on access router

- `bgp_peer` / `ospf_neighbor`:
  - evidence edges; not always dependencies but important hints

**Recommendation:** Keep `depends_on` edges explicit and operator-editable.

### 34.4 Practical RCA algorithm v1 (greedy coverage)
1. Build set `DownEntities` from firing “down” alerts.
2. For each candidate upstream entity `c` in `DownEntities`:
   - compute `ImpactSet(c)` = all downstream entities reachable via `depends_on`.
3. Score candidates by:
   - size of impact set covered
   - plus evidence weight (e.g., interface down on uplink is stronger than ping loss alone)
4. Pick top candidate(s) that cover most impacted entities.
5. Mark downstream alerts as suppressed_by = incident root cause.

This is greedy set cover. It’s not perfect, but it’s explainable and effective.

### 34.5 Evidence scoring (to improve trust)
Assign weights to evidence types (example):
- interface down on uplink: +10
- OSPF neighbor down on the same link: +7
- BGP down to upstream: +7
- ping loss from multiple agents: +6
- API unreachable but ping ok: +3 (could be management plane issue)
- syslog link-down event: +6

Compute a confidence score 0–1:
- normalized weight sum / max expected

Show this in UI: “Root cause candidate: CoreLink A-B (0.86 confidence). Evidence: interface down, OSPF down, 120 sites impacted.”

### 34.6 Redundancy handling
ISPs often have redundancy (ring topologies). Dependency edges must support:
- multiple upstreams
- “requires any” vs “requires all” semantics

Model options:
- (simple) store multiple `depends_on` edges and treat dependency as “any upstream ok”
- (advanced) dependency groups:
  - service depends_on (uplink1 OR uplink2)

Plan for this early:
- else you’ll suppress incorrectly in redundant topologies.

---

## 35. PPPoE-centric modeling (what PPPoE ISPs will pay for)

Many MikroTik ISPs run PPPoE for last-mile authentication and service delivery.

### 35.1 Concepts
- **Subscriber**: customer account
- **Session**: active PPPoE connection (with IP, uptime, rate limits)
- **Access concentrator**: router terminating PPPoE (BRAS-like role)
- **AAA/RADIUS**: external auth/billing integration (not required v1)

### 35.2 Valuable PPPoE monitoring features
1. **Active sessions count per concentrator**
2. **Session churn** (spikes can indicate instability)
3. **Per-service impact**:
   - if concentrator down, how many subscribers impacted
4. **Radius reachability** (if applicable)
5. **Customer experience metrics**:
   - ping to CPE gateway
   - latency trends by region
6. **Top complaints correlation**:
   - “high latency in PoP X” matches link utilization

### 35.3 How to represent PPPoE in your model (incrementally)
Start with:
- per router:
  - `pppoe.active_sessions`
  - `pppoe.sessions_flap_rate` (derived)
Later:
- per subscriber:
  - only if you can do it safely and at scale (privacy considerations)

**Important:** Subscriber-level monitoring is sensitive. Many ISPs will require:
- access control
- data minimization
- retention limits
- audit logging

### 35.4 “Service entity” abstraction
To compute business impact, define `service` entities:
- internet access service per site / per customer group
- PPPoE service per concentrator
- upstream transit service per provider

Then define dependencies:
- `service depends_on device`
- `customer depends_on service`

Now alerts can say:
- “Service INTERNET-POP-A degraded; estimated 3,200 subscribers impacted.”

---

## 36. Predictive signals (prevent downtime)

ISPs pay for prevention:
- capacity forecasting
- error rate trending
- temperature anomalies

### 36.1 Capacity planning
From interface bps metrics:
- compute utilization % = bps / capacity_bps
- compute 95th percentile per day/week
- alert on sustained >80%, >90%
- “days-to-saturation” estimate using slope of trend

### 36.2 Degradation precursors
- increasing interface errors/drops
- increasing ping jitter/loss
- rising CPU under steady traffic
- thermal rise (if sensors exist)

### 36.3 “Might fail soon” logic (v1)
Keep it explainable:
- alert if error rate doubled week-over-week AND above threshold
- alert if utilization slope predicts 90% within N days
- alert if temperature above safe threshold for X minutes

Explain “why” in the alert.

---

## 37. “Continue” note
Next chapters will cover:
- concrete alert rule DSL examples and evaluation pseudocode
- notification routing policies (NOC vs escalation)
- agent security implementation details (mTLS, cert rotation, bootstrap)
- Docker Compose reference for self-hosted (service list, env vars, volumes)
- multi-agent and multi-PoP deployment strategy
- UI design: map rendering, graph layout, timelines
- operations: SLOs, backups, upgrade/migrations, disaster recovery

Say **continue** to proceed.