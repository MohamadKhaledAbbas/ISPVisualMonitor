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
*(unchanged; see previous sections)*

---

# Part III — Storage, Schemas, and Ingest Contracts (Concrete)
*(unchanged; see previous sections)*

---

# Part IV — Alert Rules, RCA, PPPoE Impact, and “ISP-grade” Features
*(unchanged; see previous sections)*

---

# Part V — Security, Agent Bootstrap, and Deployment Reference

## 38. Agent security model (mTLS, identity, and bootstrap)

### 38.1 Threat model (what can go wrong)
In ISP networks, assume:
- agents may run on shared infrastructure
- internal networks may be partially compromised
- credentials leakage is catastrophic (router takeover)
- attackers may attempt to impersonate an agent and inject fake data

Therefore:
- agent identity must be cryptographically strong
- secrets must be minimized and rotated
- backend must validate and rate limit

### 38.2 Recommended trust chain
1. **Agent identity**: X.509 certificate (client cert)
2. **Backend identity**: HTTPS server cert
3. **Mutual TLS**: agent ↔ backend ingest endpoint

Benefits:
- strong agent authentication
- no long-lived API keys in configs
- simplified rotation and revocation

### 38.3 Bootstrap options (choose one)
**Option A (best for production): “One-time enrollment token”**
- Admin creates an agent record in UI.
- Backend generates an enrollment token valid for e.g. 10 minutes.
- Agent starts with:
  - backend URL
  - enrollment token
- Agent uses token to request a certificate (or to fetch a pre-issued cert).
- After enrollment, agent uses only mTLS.

**Option B (simple self-hosted): “Pre-shared agent key”**
- Backend stores agent API key.
- Agent uses key to authenticate.
- Still use TLS, but not mTLS.
- Easier, less secure, suitable only for small labs.

**Option C (enterprise): “SPIFFE/SPIRE”**
- Use SPIFFE identities and automatic cert rotation.
- Great long-term, heavy initially.

### 38.4 Certificate lifecycle
Minimum lifecycle you need:
- issue
- rotate (e.g., 30 days)
- revoke (if agent compromised)
- audit: who enrolled agent and when

Backend should track:
- agent last_seen
- cert serial
- cert expiry
- status: active/revoked

### 38.5 Transport hardening
- enforce TLS 1.2+ (TLS 1.3 preferred)
- HSTS on UI endpoints
- strict cipher suites
- request size limits (e.g., 5–20MB)
- per-agent rate limits

---

## 39. Router credential security (MikroTik best practices)

### 39.1 RouterOS API security checklist
On MikroTik routers:
- enable API only on management interface / mgmt VRF
- restrict by firewall to agent IPs only
- prefer **API-SSL** (8729) over plain API (8728)
- use strong usernames/passwords; ideally dedicated monitoring user
- limit permissions:
  - separate read-only vs diagnostic role
  - do not allow write permissions for monitoring user

### 39.2 Least privilege model (practical)
Define credential sets:
- `monitor_readonly`: can read interface stats, routing state, system health
- `diagnostics`: can run traceroute/ping if you choose to trigger from agent via router (careful)
- `config_audit`: if you later need config snapshots, keep separate and highly protected

Most of the time:
- agents should use `monitor_readonly`.

### 39.3 Secrets in the backend
Recommended:
- master key provided by environment (not in DB)
- encrypt credential blobs before storing
- audit every access path that decrypts secrets
- support key rotation later

**Important:** `pgcrypto` alone is not a full secrets manager; it’s acceptable for early versions, but plan for Vault/KMS in advanced deployments.

---

## 40. Notification security & reliability

### 40.1 Notification channels
Support (in order of typical ISP adoption):
- Telegram
- Email (SMTP)
- Webhooks (for custom NOC systems)
- Slack/MS Teams (some)
- PagerDuty/Opsgenie (larger ISPs)

### 40.2 Delivery guarantees
Notifications should have:
- retry policy with backoff
- dead-letter queue (DLQ) or “failed notifications” table
- audit and visibility: “notification sent at X, status Y”

### 40.3 Avoid alert loops
Webhooks can call back into your system; protect against loops with:
- idempotency keys
- rate limits

---

## 41. Self-hosted Docker Compose reference architecture (baseline)

This is a conceptual blueprint; actual `docker-compose.yml` can be generated later.

### 41.1 Services list
- `reverse-proxy` (Traefik/Nginx)
- `backend` (Go API + UI static hosting)
- `frontend` (optional separate if SPA)
- `db` (Postgres + TimescaleDB)
- `migrations` (runs on startup)
- `queue` (optional: NATS/Redis)
- `worker` (optional: alert evaluator, RCA worker)
- `agent` (separate deployment, runs in PoPs)

### 41.2 Minimum environment variables
Backend:
- `DATABASE_URL`
- `TENANT_MODE` (single/multi)
- `AUTH_MODE` (local/oidc)
- `ENCRYPTION_MASTER_KEY`
- `PUBLIC_BASE_URL`
- `INGEST_MAX_BODY_BYTES`
- `LOG_LEVEL`

DB:
- `POSTGRES_PASSWORD`
- volume mounts for durability

### 41.3 Persistent volumes
Must persist:
- Postgres data directory
- optional: backend uploads/exports
- optional: certificate store (if self-managing PKI)

### 41.4 Upgrade and migration strategy
- always run migrations in a controlled step
- support rollback (where possible)
- document downtime expectations
- keep schema changes backward-compatible when feasible

---

## 42. Agent deployment patterns (PoP placement)

### 42.1 Placement rule
Place agents where they can reach:
- router management IPs (private ranges)
- syslog streams if used
- probe targets (customer gateways, DNS servers)

Typical deployments:
- 1 agent per PoP (recommended)
- 2 agents per PoP for HA (active/active or active/standby)
- regional agents for small ISPs

### 42.2 High availability approach
Two patterns:

**Pattern A: Active/Active collectors**
- both agents collect; backend dedups by agent_id and sample timestamps
- safer for redundancy; more load

**Pattern B: Active/Standby**
- standby agent monitors active agent (or backend controls leader election)
- lower router load; more complexity

Start with Pattern A but keep per-device concurrency low.

### 42.3 Multi-agent ownership model
A device should have:
- preferred agent(s) list
- fallback agent list
- constraints:
  - site affinity
  - network reachability
  - load balancing

Store in DB:
- device ↔ agent assignment
- agent capabilities (supports snmp/syslog/routeros_api)

---

## 43. Observability of your platform (SRE readiness)

### 43.1 Logs, metrics, traces
Minimum:
- structured JSON logs (backend + agent)
- internal metrics endpoint (Prometheus format)
- traces (OpenTelemetry) optional but valuable early

### 43.2 System SLOs (examples)
Define SLOs you can measure:
- outage detection latency p95 < 30s
- ingest acceptance p95 < 500ms
- alert notification dispatch p95 < 30s
- dashboard query latency p95 < 2s

### 43.3 Runbooks
Every “paid-grade” system needs runbooks:
- DB down
- ingest backlog
- agent offline
- certificate expiry
- notification provider failure

---

## 44. “Continue” note
Next chapters will cover:
- concrete rule DSL examples (router down, BGP down, ping loss)
- evaluation engine pseudocode and scheduling
- incident correlation strategies
- UI workflow design: map + graph + timeline + evidence panels
- diagnostics execution model (safe on-demand traceroute)
- multi-tenant isolation patterns (RLS, per-tenant keys)
- backups, DR, and compliance considerations

Say **continue** to proceed.