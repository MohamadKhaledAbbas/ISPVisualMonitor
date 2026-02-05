# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–VII
For brevity inside this file, earlier parts are retained as written in previous iterations. This section continues appending new chapters without removing previous guidance.

*(In your local copy, keep the full text from earlier parts; this document is designed to be cumulative.)*

---

# Part VIII — Self‑Hosted Reference Deployment, Backups/DR, Upgrades, and Sizing

> This part is written as a “reference architecture” for **self-hosted ISPs** using Docker Compose.
> It is intentionally explicit: ISPs will judge you on how deployable and operable your system is.

## 60. Self‑Hosted reference deployment (Docker Compose)

### 60.1 Goals of the reference deployment
- “One command” bring-up for evaluation and pilot
- Durable storage (DB and state)
- TLS termination with a reverse proxy
- Clear upgrade procedure
- Minimal moving parts, but with an upgrade path to HA

### 60.2 Recommended baseline services (small/mid ISP)
**Always include**
- `proxy` (Traefik or Nginx)
- `backend` (Go API + UI)
- `db` (Postgres + TimescaleDB)
- `migrate` (schema migrations)
- `worker` (rule eval + incident correlation + notifications)  
  *(You can embed worker inside backend at first, but separate container is cleaner.)*

**Optional / scale features**
- `nats` (JetStream) for buffering and decoupling
- `redis` for caching (optional)
- `loki/promtail` for logs (optional)
- `prometheus/grafana` for platform monitoring (optional)

### 60.3 Network layout (Compose)
Use two networks:
- `public` network: proxy only
- `private` network: backend, db, nats, worker

Only expose:
- proxy ports 80/443 to the outside.

### 60.4 Storage layout
Persistent volumes:
- `db_data` (Postgres/Timescale)
- `backend_data` (exports, generated reports)
- optional `proxy_certs` (if proxy stores certs)
- optional `worker_data` (if it stores state)

**Rule:** Do not store secrets in volumes unless encrypted. Prefer environment variables or a secrets file.

### 60.5 Configuration management (industry-grade)
ISPs will ask: “How do we manage config and secrets?”
Provide:
- `.env` template for non-sensitive values
- support Docker secrets (or mounted files) for sensitive values
- clear docs for generating keys/certs

### 60.6 Example environment variables (reference)
**Backend**
- `DATABASE_URL=postgres://...`
- `PUBLIC_BASE_URL=https://monitor.example.com`
- `AUTH_MODE=local|oidc`
- `OIDC_ISSUER_URL=...` (if OIDC)
- `OIDC_CLIENT_ID=...`
- `OIDC_CLIENT_SECRET_FILE=/run/secrets/oidc_secret`
- `ENCRYPTION_MASTER_KEY_FILE=/run/secrets/master_key`
- `INGEST_MAX_BODY_BYTES=10485760`
- `LOG_LEVEL=info`
- `FEATURE_FLAGS=...` (optional)

**Worker**
- `DATABASE_URL=...`
- `QUEUE_URL=...` (if using NATS/Redis)
- `EVAL_FAST_INTERVAL=10s`
- `EVAL_MEDIUM_INTERVAL=60s`
- `EVAL_SLOW_INTERVAL=300s`

**DB**
- `POSTGRES_PASSWORD_FILE=/run/secrets/db_password`

### 60.7 TLS options for self-hosted
Two real-world approaches:

**Option A: ACME/Let’s Encrypt**
- works if ISP exposes service publicly and has valid DNS
- easiest for most deployments

**Option B: Internal CA**
- common for on-prem, private DNS
- you must document:
  - how to generate CA
  - how to distribute trust roots to browsers
  - certificate rotation plan

**Recommendation:** Support both. Many ISPs will run internal-only.

---

## 61. Backup & Restore (must be documented like a product)

### 61.1 What must be backed up
Minimum:
- Postgres/Timescale database
- encryption keys (master key) **stored separately**
- uploaded exports / generated reports (if any)
- optional: CA private key (if you run internal CA) — extreme care

**Critical:** If the master key is lost, encrypted credentials become unrecoverable.

### 61.2 Backup strategy (practical)
- nightly full DB backup (pg_dump or physical backups)
- WAL archiving if you need point-in-time restore (PITR)
- store backups off-host (S3, NFS, external storage)
- test restores regularly (quarterly at least)

### 61.3 Restore procedure (runbook essentials)
Document steps:
1. stop backend/worker
2. restore DB snapshot to clean DB volume
3. restore encryption master key (same as original)
4. start DB
5. start migrate (if needed)
6. start backend/worker
7. verify system health and agent connectivity

### 61.4 Backup retention
Example policy:
- daily backups keep 14 days
- weekly backups keep 8 weeks
- monthly backups keep 12 months

---

## 62. Disaster Recovery (DR) and high availability (HA)

### 62.1 DR goals by ISP size
**Small ISP:**  
- recover in hours (RTO 4–12h), tolerate some data loss (RPO 24h)

**Mid ISP:**  
- recover < 2h, RPO < 1h

**Large ISP:**  
- active/active or active/passive; RTO < 30m; PITR

### 62.2 HA roadmap
Phase approach:
1. Single node with backups (pilot)
2. DB with replication + failover (Patroni or managed DB)
3. Multiple backend instances behind proxy
4. Queue-based ingest to tolerate DB slowness
5. Multi-region (SaaS)

### 62.3 Agent behavior during DR events
Agents must:
- buffer locally when backend unreachable
- keep collecting locally (if configured)
- resume pushing when backend returns
- expose local buffer depth and last successful push

---

## 63. Upgrades & migrations (the difference between hobby and product)

### 63.1 Versioning policy
Adopt semantic versioning:
- MAJOR: breaking changes
- MINOR: backward compatible features
- PATCH: bugfixes

### 63.2 Migration strategy
- schema migrations must be idempotent
- avoid long locks on hypertables
- for big changes:
  - add new columns/tables
  - backfill asynchronously
  - switch reads to new schema
  - later drop old fields

### 63.3 Rolling upgrade readiness
Even if you start single-node, design for:
- multiple backend instances
- multiple workers
- no “in-memory only” state that breaks during restart

### 63.4 Change management artifacts
ISPs appreciate:
- release notes with upgrade steps
- known issues
- rollback instructions
- database migration notes (duration/impact)

---

## 64. Capacity planning and sizing (routers → CPU/RAM/disk)

Sizing depends on:
- polling frequency
- metrics per router
- number of interfaces
- retention
- number of probe targets

### 64.1 A practical “starter” sizing model
Assume per router:
- 20 metrics every 30s (core health + interfaces + BGP)
- plus pings to 3 targets every 10s from PoP agent

This is a moderate load.

### 64.2 Small ISP (10–50 routers)
**Backend/DB on one VM**
- 2–4 vCPU
- 8–16 GB RAM
- 100–300 GB SSD (depends on retention)

**Agents**
- 1–2 vCPU
- 1–2 GB RAM
- small SSD for buffering (10–20GB)

### 64.3 Mid ISP (50–300 routers)
**Backend**
- 4–8 vCPU
- 16–32 GB RAM

**DB**
- 8 vCPU
- 32 GB RAM
- NVMe SSD strongly recommended
- consider separate DB host

**Agents**
- 1 per PoP:
  - 2–4 vCPU
  - 4–8 GB RAM

### 64.4 Large ISP (300–2000+ routers)
**Split roles**
- dedicated DB cluster
- multiple ingest workers
- queue mandatory
- multi-agent in many PoPs

At this scale, you’ll want:
- careful metric cardinality control
- rollups and retention tuning
- caching for UI dashboards

### 64.5 Disk growth guidance
Time-series storage grows fast. Control it with:
- retention policies
- compression
- rollups
- avoid high-cardinality labels

ISPs will ask: “How big will this get in 6 months?” Provide a sizing calculator later.

---

## 65. Reference operational checklists (what an ISP expects)

### 65.1 Pre-production checklist
- TLS enabled and verified
- RBAC configured
- audit logs enabled
- backup configured and tested
- alert channels tested
- maintenance windows policy set
- agent enroll/rotate tested
- retention configured

### 65.2 Production readiness checklist
- monitoring of the monitoring system (Prometheus recommended)
- DB vacuum/maintenance plan
- incident on-call workflow
- upgrade playbook
- DR playbook
- security patch policy

---

## 66. “Continue” note
Next chapters will cover:
- concrete “MikroTik-first” device model: how to identify core/edge/access and how to tag them
- a recommended EVE-NG lab topology *with* OSPF/iBGP/eBGP examples to test RCA
- collection hardening: API rate limits, session reuse, circuit breakers
- a feature blueprint that differentiates your product commercially:
  - SLA packs
  - customer-impact heatmaps
  - competitor-grade reporting
  - NOC workflow features

Say **continue** to proceed.