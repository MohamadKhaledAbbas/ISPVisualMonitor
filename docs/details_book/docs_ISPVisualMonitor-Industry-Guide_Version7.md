# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## 0. Why this document exists
*(unchanged; see earlier sections)*

## 1. Product vision & non-goals
*(unchanged; see earlier sections)*

## 2. Design principles (industry-grade requirements)
*(unchanged; see earlier sections)*

## 3. Reference architecture (MikroTik-first, extensible)
*(unchanged; see earlier sections)*

## 4. Deployment models
*(unchanged; see earlier sections)*

## 5. ISP routing fundamentals (needed for topology + RCA)
*(unchanged; see earlier sections)*

## 6. Telemetry methods (what to collect and why)
*(unchanged; see earlier sections)*

## 7. Data plane agent (design blueprint)
*(unchanged; see earlier sections)*

## 8. Control plane backend (design blueprint)
*(unchanged; see earlier sections)*

## 9. Data model essentials (minimum production schema concepts)
*(unchanged; see earlier sections)*

## 10. Root Cause Analysis (RCA) — concept overview
*(unchanged; see earlier sections)*

## 11. Alerting (what makes it paid-grade)
*(unchanged; see earlier sections)*

## 12. Frontend (map + graph + investigation workflow)
*(unchanged; see earlier sections)*

## 13. Security & compliance (minimum bar)
*(unchanged; see earlier sections)*

## 14. Testing & simulation (EVE-NG + CHR)
*(unchanged; see earlier sections)*

## 15. Roadmap features that make ISPs “move up a level”
*(unchanged; see earlier sections)*

## 16. How to use this document
*(unchanged; see earlier sections)*

---

# Part II — Implementation Blueprint (Concrete Design)
*(unchanged; see earlier sections)*

# Part III — Storage, Schemas, and Ingest Contracts (Concrete)
*(unchanged; see earlier sections)*

# Part IV — Alert Rules, RCA, PPPoE Impact, and “ISP-grade” Features
*(unchanged; see earlier sections)*

# Part V — Security, Agent Bootstrap, and Deployment Reference
*(unchanged; see earlier sections)*

# Part VI — Rule DSL, Evaluation Engine, Correlation, and Diagnostics
*(unchanged; see earlier sections)*

---

# Part VII — Multi‑Tenancy, Isolation, Scaling, and Additional Telemetry

## 51. Multi‑tenancy: supporting many ISPs safely

### 51.1 Tenancy goals
A multi-tenant architecture must guarantee:
- **data isolation**: tenant A can never see tenant B data
- **resource isolation**: tenant A cannot starve tenant B
- **operational isolation**: per-tenant retention, limits, and policies
- **auditability**: prove isolation in security reviews

### 51.2 Three isolation levels (choose by product phase)

#### Level 1: Application-enforced tenancy (fastest MVP)
- every query includes `tenant_id` filter
- every write includes `tenant_id`
- simplest to build; risk depends on discipline and tests

**Must-haves if using Level 1**
- centralized DB access layer that always injects `tenant_id`
- tests that attempt cross-tenant access
- code review rules: “no raw SQL without tenant guard”

#### Level 2: Postgres Row-Level Security (RLS) (recommended for SaaS)
- enable RLS on all tenant-owned tables
- set `app.tenant_id` via `SET LOCAL` each request
- policies ensure tenant isolation in DB itself

**Pros**
- strong safety net
- reduces “oops we forgot tenant filter” class of bugs

**Cons**
- more complex debugging
- must be careful with admin queries and migrations

#### Level 3: Database-per-tenant (enterprise, largest customers)
- physically separate DBs per tenant
- strongest isolation; heavy ops overhead
- good for “premium” self-hosted or regulated customers

### 51.3 Recommended path
- Start with Level 1 (app-enforced) to ship faster.
- Design DB schema and access layer so you can migrate to Level 2 (RLS) without rewriting everything.

### 51.4 Per-tenant limits (prevent noisy neighbors)
Store tenant quotas:
- max devices
- max agents
- max metric ingest rate
- max retention days
- max diagnostics runs per hour

Enforce at:
- UI (prevent configuration)
- backend (hard enforcement)
- ingest (rate limiting)

---

## 52. Scaling by workload type (where systems usually break)

Monitoring platforms break in predictable places:
1. **Ingest throughput** (metrics/events per second)
2. **DB write amplification** (indexes + constraints)
3. **Alert evaluation** (windowed queries across many entities)
4. **UI queries** (dashboards without rollups)
5. **Topology/RCA computations** (graph operations)

### 52.1 Design for “cheap writes, cheap reads”
- batch ingest
- minimize indexes on raw hypertables
- continuous aggregates for reads
- cache expensive UI queries

### 52.2 Separate “hot path” from “cold path”
- hot path:
  - ingest -> write -> minimal compute
  - alert evaluation on schedule
- cold path:
  - reports, exports
  - historical queries
  - ML/anomaly experiments

### 52.3 Worker pool pattern
Even in a monolith, implement background workers:
- `ingest_worker` (optional if ingest writes to queue)
- `rule_eval_worker`
- `incident_correlation_worker`
- `notification_worker`
- `reporting_worker`

Each worker has:
- concurrency config
- backpressure behavior
- metrics and logs

---

## 53. Queue usage: when you need it and how to adopt it safely

### 53.1 When to introduce a queue
You likely need a queue when:
- > 50–100 routers per tenant with frequent metrics
- multiple agents sending large batches
- you want stronger delivery guarantees
- notifications must be decoupled from evaluation

### 53.2 Minimal queue adoption strategy
- keep ingest endpoint
- ingest endpoint writes payload to queue (NATS/Redis Streams)
- workers consume and write to DB
- return 202 Accepted quickly

This makes ingest stable during DB slowdowns.

### 53.3 Choosing a queue
Pragmatic choices:
- **NATS JetStream**: lightweight, fast, good for event streams
- **RabbitMQ**: mature, reliable, more ops overhead
- **Redis Streams**: easy if you already use Redis

For an ISP monitoring platform:
- NATS JetStream is often a very good fit.

---

## 54. Adding SNMP alongside RouterOS API without bloat

### 54.1 Principle: one normalized metric model
Collectors can be many, but metrics must be canonical.
Example:
- RouterOS API collector produces `interface.rx_bps`
- SNMP collector produces `interface.rx_bps`
Same metric name, same entity_id.

Backend doesn’t care how it was obtained; it cares about:
- quality
- timestamps
- coverage

### 54.2 What to prefer SNMP for
- interface counters (ifInOctets/ifOutOctets and 64-bit variants)
- error counters
- standard CPU/memory where available
- cross-vendor future

### 54.3 Avoid cardinality explosions
SNMP can produce huge label cardinality (interface names, indexes).
Best practice:
- map SNMP ifIndex -> your `device_interface_id`
- store interface name in inventory table, not as metric label

---

## 55. Syslog ingestion: building a timeline safely

### 55.1 Two approaches
1. Agent receives syslog locally and forwards normalized events
2. Backend receives syslog directly (less common in self-hosted)

Given your agent model:
- prefer syslog ingestion into the agent.

### 55.2 Normalization pipeline
Syslog messages must be converted to:
- event_type
- severity
- entity mapping (device_id, interface_id if possible)
- payload (original message stored for audit)

### 55.3 Dedup & noise controls
Syslog is noisy. Implement:
- repeated message suppression (same message within N seconds)
- severity filtering
- allow “event-only” mode (store but don’t alert directly)

---

## 56. Topology inference: from “manual” to “semi-automatic”

### 56.1 Discovery sources in MikroTik environments
- RouterOS API:
  - BGP peers and OSPF neighbors
  - interface lists
  - IP addresses
- SNMP:
  - interface tables
- Optional:
  - LLDP (if enabled)
  - ARP tables (careful: dynamic and noisy)

### 56.2 Inference heuristics (v1)
- If router A has OSPF neighbor router B:
  - propose `topology_edges` type `ospf_neighbor`
- If router A has BGP session to router B:
  - propose `topology_edges` type `bgp_peer`
- If two routers are in same site:
  - propose “intra-site adjacency” (low confidence)

Then allow operator approval:
- UI shows proposed links/dependencies
- operator accepts/rejects and can edit endpoints

### 56.3 Avoid dangerous assumptions
Do not assume:
- BGP peer means physical adjacency (it can be loopback)
- OSPF neighbor always means direct fiber (it can be VLAN/tunnel)
Therefore: separate “neighbor” from “physical link.”

---

## 57. Active probes at scale (SLA without overload)

### 57.1 Probe design principles
- probes must be distributed (agents in PoPs)
- probes must be limited (avoid flooding)
- probes must be multi-target (for RCA)
- probes must have stability (consistent schedule)

### 57.2 Recommended probe target classes
Per tenant, define:
- **core reachability targets**:
  - loopbacks of core routers
- **service targets**:
  - DNS resolvers
  - gateway IPs
  - external well-known endpoints (1.1.1.1/8.8.8.8) for upstream health
- **customer sample targets** (optional, careful):
  - only a sampled subset to avoid privacy and load issues

### 57.3 Probe frequency guidance
- core targets: 10–30s
- service targets: 30–60s
- customer samples: 60–300s (if used)
- traceroute: on incident start or on-demand; not continuous

### 57.4 From probes to SLAs
SLA usually wants:
- availability (up/down)
- latency percentile
- packet loss

Compute from probe metrics:
- define “available” if loss < X and consecutive failures < N
- use rollups for reporting
- show P95 latency

---

## 58. Data minimization & privacy (subscriber data)
If you monitor PPPoE subscribers:
- treat it as sensitive personal data
- ensure strict RBAC (NOC vs admin)
- retention policies (e.g., session logs keep 7–30 days)
- audit access to subscriber-level views
- consider hashing or minimizing PII (no names unless needed)

Even if you don’t target GDPR regions, ISPs often require these controls.

---

## 59. “Continue” note
Next chapters will cover:
- concrete Docker Compose example (self-hosted reference)
- certificate management options for self-hosted (internal CA vs ACME)
- backup/restore and disaster recovery (DR) playbooks
- upgrade strategy (schema migration, versioning, rollback)
- how to structure agent configs and multi-PoP assignment
- performance sizing guidance (CPU/RAM/disk) by router count
- commercialization considerations (licensing agent, enterprise features)

Say **continue** to proceed.