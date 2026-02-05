# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XII
Earlier parts cover: architecture, storage, alerting/RCA, security, multi-tenancy, self-hosted operations, lab topologies, productization, starter packs, telemetry matrix, privacy policy templates, performance plans, and demo playbooks. This part dives into redundancy modeling, advanced RCA, PPPoE/RADIUS integration patterns, and packaging/pricing considerations.

---

# Part XIII — Redundancy Modeling, Advanced RCA, PPPoE/RADIUS Patterns, and Packaging

## 95. Redundancy and dependency semantics (rings, dual uplinks, ECMP)

### 95.1 Why redundancy breaks naive RCA
A naive dependency model says:
- “Access depends_on Aggregation”
- “Aggregation depends_on Core”

But in real ISPs:
- Access has **two uplinks**
- Aggregation has **ring topology**
- Traffic uses **ECMP** or fast reroute
- A single upstream failure shouldn’t imply “everything down”

Therefore you must represent redundancy explicitly.

### 95.2 Dependency group model (recommended)
Instead of simple edges, represent:
- a dependent entity requires **any of** a set (OR)
- or requires **all of** a set (AND)
- or requires “N of M” (k-of-n) for advanced HA

#### 95.2.1 Practical representation in DB
Option A (simple, JSONB):
- `dependency_groups`
  - `id`, `tenant_id`
  - `entity_type`, `entity_id` (dependent)
  - `mode` = `any|all|kofn`
  - `k` (nullable)
  - `members` (array of {entity_type, entity_id})
  - `created_at`

Option B (normalized):
- `dependency_groups` (header)
- `dependency_group_members` (rows)
This is more queryable and scalable.

### 95.3 How to use redundancy in suppression
When an upstream fails, only suppress downstream if:
- all upstreams in an `any` group are down
- at least k upstreams down in a `kofn` group

Example:
- Access depends_on ANY(AGG1, AGG2)
- If AGG1 down but AGG2 up, do not suppress access alerts (it should still work).
- If both AGG1 and AGG2 down, suppress downstream and set impact.

### 95.4 Link redundancy vs routing redundancy
A link can fail while routing remains stable due to alternate path.
Therefore, track:
- physical link status (interface/link)
- logical reachability (probe)
- routing adjacency (OSPF/BGP)

Your “service availability” should be based on experience + reachability, not just one interface status.

---

## 96. Advanced RCA: multi-cause incidents and evidence graphs

### 96.1 Why single root cause is sometimes wrong
In real outages:
- a core link fails AND an upstream drops routes
- a PoP loses power while another link is congested
- a maintenance window overlaps a failure

Your RCA must allow:
- more than one root cause candidate
- evidence-based confidence per candidate
- “primary” vs “contributing” causes

### 96.2 Multi-cause RCA model
Store in incident:
- `root_causes[]` list:
  - entity type/id
  - confidence
  - evidence references
  - classification (primary/contributing)

### 96.3 Evidence graph
Instead of just a list, treat evidence as a graph:
- nodes: metrics/events/alerts
- edges: “supports”, “correlates_with”, “caused_by”
This is too heavy for v1, but you can implement a simplified form:
- evidence list grouped by entity and time

### 96.4 Confidence calibration (trust feature)
ISPs will distrust “AI-like” claims if wrong.
Make confidence:
- conservative
- explainable
- based on weighted evidence

Also track:
- historical accuracy (postmortem feedback) if possible:
  - operator selects “true root cause”
  - system learns weights (optional later)

---

## 97. PPPoE modeling deep dive (without going off the rails)

### 97.1 The “PPPoE value triangle”
PPPoE monitoring value comes from:
1. **Impact**: how many subscribers affected
2. **Experience**: latency/loss by region/customer segment
3. **Support efficiency**: fast correlation to PoP/link issue

### 97.2 Entities you may introduce (phase-based)
Phase 1 (counts only):
- `pppoe_concentrator` = device role access
- metrics:
  - active sessions
  - churn rate

Phase 2 (subscriber segments):
- session counts by profile/plan (e.g., “10M”, “50M”, “business”)
- still no subscriber IDs

Phase 3 (subscriber-level, opt-in):
- subscriber ID (username) hashed or stored with strict controls
- session events and durations
- careful retention and RBAC

### 97.3 Subscriber impact computation
Impact can be computed even without subscriber IDs:
- if concentrator sessions drop from 3200 → 50, impact ~ 3150 subscribers
- if PoP depends_on that concentrator, propagate impact to service entity

### 97.4 Common PPPoE incidents and how to detect them
- concentrator down (sessions drop to near zero)
- BRAS CPU exhaustion (sessions stable but latency spikes)
- authentication/RADIUS issues (session connect failures, churn)
- upstream congestion (sessions stable but experience degrades)

Your system should separate:
- “sessions down” vs “experience degraded”

---

## 98. RADIUS integration patterns (future but plan now)

Many PPPoE ISPs use RADIUS. Even if you don’t integrate early, plan the architecture.

### 98.1 Integration goals
- detect AAA outages
- correlate auth failures to incidents
- optionally import subscriber/service metadata (careful with privacy)

### 98.2 Data sources
- RADIUS server health probes (TCP/UDP checks)
- syslog from RADIUS server
- optional RADIUS accounting logs (high volume, privacy sensitive)

### 98.3 Recommended approach
Start with:
- monitor RADIUS server availability (probe + service metrics)
- monitor auth failure rate from concentrators (if RouterOS exposes)
Then later:
- integrate accounting logs for deeper insights

### 98.4 Privacy-by-design rules
If you ingest RADIUS accounting:
- avoid storing plaintext usernames unless required
- hash identifiers for correlation
- short retention by default
- explicit customer opt-in with policy acknowledgement

---

## 99. Packaging and pricing strategy (pragmatic guidance)

You asked for “industry grade that ISPs will pay for.” Packaging matters.

### 99.1 Common model: open agent + commercial server
- agent: open source
- server: paid with optional free tier

### 99.2 Feature tiers (example blueprint)
**Free / Community**
- basic inventory + map
- basic alerts (device/link down)
- limited retention (e.g., 7–14 days)
- limited device count
- community support

**Pro**
- RCA + suppression
- SLA reporting pack
- capacity planning
- telegram/email/webhook notifications
- 1 year rollups

**Enterprise**
- OIDC/SAML, audit exports
- HA/DR guidance and support
- multi-region SaaS
- advanced PPPoE pack
- compliance features
- 24/7 support

### 99.3 Avoid “crippling” the free tier too much
Free tier should still demonstrate value, or ISPs won’t trust it.
Better to limit:
- scale and retention
not core correctness.

### 99.4 Licensing pitfalls
If agent is open source:
- ensure enterprise customers can validate it
- avoid license terms that scare legal teams unnecessarily

Apache-2.0 tends to reduce friction.

---

## 100. Developer documentation to ship with the product (repo-level)

You should publish docs in a structured way:

**Control plane repo docs**
- `docs/architecture.md` (high-level)
- `docs/deployment/self-hosted.md`
- `docs/configuration.md`
- `docs/api.md` (OpenAPI)
- `docs/security.md`
- `docs/backup-and-restore.md`
- `docs/upgrade.md`

**Agent repo docs**
- `docs/installation.md`
- `docs/configuration.md`
- `docs/collected-data.md`
- `docs/security.md`
- `docs/troubleshooting.md`
- `SECURITY.md`
- `CONTRIBUTING.md`

### 100.1 Make docs “operator-first”
ISPs are operator-driven. Your docs should answer:
- how to deploy
- how to add routers
- how to verify agent reachability
- what ports are needed
- how to tune polling and alerts
- how to handle incidents

---

## 101. “Continue” note
Next chapters can go into “implementation detail mode”:
- define canonical metric/event name registry
- provide JSON schemas for ingest payloads and rules
- define a stable versioning strategy for schemas
- provide recommended defaults for polling rates and jitter
- create an “ISPVisualMonitor style guide” (naming, tags, roles, site conventions)
- provide a sample “operator runbook library” for common incidents (BGP down, congestion, fiber errors)

Say **continue** to proceed.