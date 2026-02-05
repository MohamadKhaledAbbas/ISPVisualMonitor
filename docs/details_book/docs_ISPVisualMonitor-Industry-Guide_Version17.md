# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XVI
Earlier parts define architecture, contracts, schemas, and APIs. This part focuses on operational pitfalls (Timescale/Postgres), dashboard query performance, seed/demo data strategy, and CI/CD + release engineering for trust.

---

# Part XVII — DB/Timescale Gotchas, Dashboard Performance, Demo Seed Data, and Release Engineering

## 121. TimescaleDB & Postgres operational pitfalls (and how to avoid them)

### 121.1 Hypertable design pitfalls
**Pitfall:** too many indexes on raw hypertables  
- Every insert must update every index → write amplification.

**Guidance**
- keep raw hypertable indexes minimal:
  - `(tenant_id, metric_name, ts DESC)`
  - `(tenant_id, entity_type, entity_id, ts DESC)`
- move complex queries to continuous aggregates

**Pitfall:** high-cardinality labels in `labels` JSONB  
- Makes queries slow and increases storage.

**Guidance**
- keep labels small and stable (provider, region, role)
- store interface names in inventory, not as metric labels
- never store subscriber identifiers as labels

### 121.2 Chunk interval tuning
Timescale stores data in “chunks.”
- too small: too many chunks, overhead
- too large: compression and retention less efficient

**Guidance**
- start with chunk interval aligned with retention and query patterns:
  - e.g., 1 day chunks for high rate metrics
- re-evaluate after measuring insert rate

### 121.3 Compression tradeoffs
Compression saves space but costs CPU on queries.

**Guidance**
- compress chunks older than e.g., 7 days
- keep most recent data uncompressed for dashboards
- test p95 query latency after enabling compression

### 121.4 Retention jobs and maintenance windows
Retention policies delete old chunks.
**Pitfall:** running retention during peak hours can cause IO spikes.

**Guidance**
- run retention jobs in off-peak windows
- monitor DB IO and vacuum behavior

### 121.5 Vacuum and autovacuum tuning
Monitoring platforms insert constantly.
**Pitfall:** autovacuum not tuned → bloat.

**Guidance**
- set sensible autovacuum thresholds for hypertables
- monitor bloat and dead tuples
- consider Timescale guidance for continuous ingest workloads

### 121.6 Connection pooling
**Pitfall:** too many DB connections from backend/worker → exhaustion.

**Guidance**
- use pgx pool limits
- consider PgBouncer in mid/large deployments
- keep worker concurrency in line with DB capacity

### 121.7 “Slow query surprises”
**Pitfall:** dashboards query raw data for long ranges.

**Guidance**
- for ranges > 24h, use rollups (1m/1h)
- enforce query limits server-side
- implement “max datapoints” downsampling in API

---

## 122. Dashboard query performance patterns (backend-driven UX)

### 122.1 Never let the browser build complex queries
The backend should:
- provide aggregated time-series
- enforce tenant limits
- cache expensive summaries

### 122.2 Standard dashboard endpoints
Implement endpoints that return:
- series data with downsampling (e.g., max 600 points)
- summary tiles:
  - current status counts
  - top 10 congested links
  - top 10 erroring interfaces
- incident trends (incidents/day)

### 122.3 Downsampling strategy
Given a time range:
- compute bucket size = range / max_points
- query continuous aggregate at appropriate resolution
- return:
  - min/avg/max per bucket
  - p95 if needed

This yields smooth graphs without huge payloads.

### 122.4 Cache strategy
Cache:
- site summary statuses (10–30s)
- map overlay data (10–30s)
- “top N” lists (30–60s)

Use:
- in-memory cache for single instance
- Redis for multi-instance (optional later)

---

## 123. Demo and seed data strategy (repeatable environments)

### 123.1 Why seed data matters
To demonstrate value quickly, you need:
- pre-made sites and coordinates
- pre-made device roles/tags
- pre-made links and dependencies
- pre-made probe targets
- pre-enabled alert packs

### 123.2 Seed data sources
- static JSON/YAML fixtures in repo
- migration “seed step” for dev/demo only
- UI import (CSV) for real customers

### 123.3 Recommended seed bundles
Ship multiple seed bundles:
- `demo_small.json`: 3 sites, 6 routers
- `demo_mid.json`: 10 sites, 40 routers
- `lab_eve.json`: matches your EVE-NG topology

### 123.4 Safe seeding practices
- seeding should be optional and disabled in production by default
- never ship default credentials
- demo tenant separated from real tenants

---

## 124. Release engineering (agent and server) — what “trusted by ISPs” requires

### 124.1 Container image hygiene (server)
- build minimal images (distroless or alpine where appropriate)
- pin dependencies
- include build metadata:
  - git commit
  - build date
  - version
- scan images for vulnerabilities (Trivy/Grype)
- publish SBOM (Syft)

### 124.2 Agent release process (open-source trust)
ISPs will scrutinize the agent most.

Recommended:
- publish source + tagged releases
- build binaries for:
  - linux amd64/arm64
- publish container images (for k8s)
- sign artifacts:
  - cosign for containers
  - gpg/sigstore for binaries
- provide checksums and verification instructions

### 124.3 Reproducible builds (advanced but powerful)
Even partial reproducibility helps:
- pinned Go version
- pinned module versions
- deterministic build flags
- CI pipelines that can be rerun

### 124.4 Compatibility policy
Publish:
- supported backend versions for each agent version
- deprecation schedule
- minimum supported RouterOS versions (documented)

### 124.5 CI/CD pipeline outline
**Agent CI**
- lint + unit tests
- schema conformance tests (ingest payload)
- build + sign artifacts
- publish SBOM

**Server CI**
- lint + unit tests
- integration tests (DB + migrations)
- performance smoke tests (ingest)
- build container + scan + sign

---

## 125. Operational “gotchas” outside the DB

### 125.1 Clock sync is mandatory
If agent/router clocks drift:
- event timelines become misleading
- correlation breaks

Provide:
- NTP checks and “clock skew” alerts

### 125.2 Notification provider failures
Telegram/email may fail during outages.
You must:
- retry with backoff
- show failures in UI
- optionally switch to fallback channel

### 125.3 Partial outages and split-brain visibility
If a PoP is isolated:
- local agent sees failure
- central backend may not be reachable

Agents must buffer; backend must show:
- “agent offline from backend” vs “routers down”
This is an important operator distinction.

---

## 126. “Continue” note
Next chapters can include:
- a full set of “operator dashboards” (what pages to build first)
- incident postmortem workflow and exports
- a deeper “service model” (internet service entities, upstream providers, PoP services)
- advanced correlation: combining syslog bursts + probe changes + routing state
- future adoption of gRPC or NATS for ingest and diagnostics

Say **continue** to proceed.