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
*(unchanged; see previous sections)*

---

## 3. Reference architecture (MikroTik-first, extensible)
*(unchanged; see previous sections)*

---

## 4. Deployment models
*(unchanged; see previous sections)*

---

## 5. ISP routing fundamentals (needed for topology + RCA)
*(unchanged; see previous sections)*

---

## 6. Telemetry methods (what to collect and why)
*(unchanged; see previous sections)*

---

## 7. Data plane agent (design blueprint)
*(unchanged; see previous sections)*

---

## 8. Control plane backend (design blueprint)
*(unchanged; see previous sections)*

---

## 9. Data model essentials (minimum production schema concepts)
*(unchanged; see previous sections)*

---

## 10. Root Cause Analysis (RCA) — concept overview
*(unchanged; see previous sections)*

---

## 11. Alerting (what makes it paid-grade)
*(unchanged; see previous sections)*

---

## 12. Frontend (map + graph + investigation workflow)
*(unchanged; see previous sections)*

---

## 13. Security & compliance (minimum bar)
*(unchanged; see previous sections)*

---

## 14. Testing & simulation (EVE-NG + CHR)
*(unchanged; see previous sections)*

---

## 15. Roadmap features that make ISPs “move up a level”
*(unchanged; see previous sections)*

---

## 16. How to use this document
*(unchanged; see previous sections)*

---

# Part II — Implementation Blueprint (Concrete Design)
*(unchanged; see previous sections)*

---

# Part III — Storage, Schemas, and Ingest Contracts (Concrete)
*(unchanged; see previous sections)*

---

# Part IV — Alert Rules, RCA, PPPoE Impact, and “ISP-grade” Features
*(unchanged; see previous sections)*

---

# Part V — Security, Agent Bootstrap, and Deployment Reference
*(unchanged; see previous sections)*

---

# Part VI — Rule DSL, Evaluation Engine, Correlation, and Diagnostics

## 45. Rule specification: a pragmatic DSL/JSON that works

### 45.1 Why you need a formal spec
A paid-grade monitoring system cannot have “hidden” rules in code only. You need:
- portable rule definitions (export/import)
- versioning and review
- explainability
- testability (“this rule should fire under these conditions”)

### 45.2 Recommended v1 approach: JSON rules + a small expression language
Do *not* implement a large query language initially.
Instead:
- define a JSON schema for common rule types
- allow limited expressions for combinations

#### 45.2.1 Base rule fields (common)
Every rule should include:
- `id`, `name`, `enabled`
- `severity` (info/warn/critical)
- `scope` (by tags/sites/devices/interfaces/services)
- `for` duration (how long condition must hold)
- `cooldown` duration (minimum time between notifications)
- `notify_policy_id`
- `annotations`:
  - `summary_template`
  - `runbook_url` (important in production)

### 45.3 Canonical entity selectors (“scope”)
Selectors must be predictable and fast. A good set:
- by tag: `device.tags contains "core"`
- by site: `site.id in [...]`
- by role: `device.role == "edge"`
- by explicit list: `device.id in [...]`
- by link: `link.id == ...` (later)
- by “critical uplinks only” flag: `interface.is_uplink == true` (needs a field or tag)

**Important:** allow ISPs to label “uplink interfaces” manually. Auto-detection is not reliable early.

### 45.4 Rule types you should implement early

#### A) Reachability / Device down
**Signal sources**
- ping from agent to mgmt IP
- RouterOS API connect success
- optionally SNMP response

**Rule design**
- primary: ping loss 100% for 30–60s → down
- secondary: API/SNMP unreachable adds evidence

**Why not “API unreachable = down”?**
Because management plane can break while forwarding plane still works.

#### B) Link/interface down
- interface running == false for 30s (or immediate for uplinks)
- correlation with OSPF neighbor down strengthens evidence

#### C) BGP session down
- BGP state != Established for 30s–2m
- plus a “flap rate” rule

#### D) SLA degradation
- loss > X% for Y
- RTT > threshold for Y
- jitter > threshold (optional)

#### E) Capacity & saturation
- utilization > 85% for 5m
- utilization > 95% for 1m (critical)

#### F) Error-rate degradation
- interface errors/drops per second above threshold for 5–10m
- or a baseline deviation (later)

### 45.5 Example rule definitions (illustrative JSON)
These examples are *conceptual* and define the shape.

#### 45.5.1 Device down (ping-based)
- scope: all devices tagged `core` or `edge`
- condition: ping loss == 100% over 3 probes
- for: 30s

#### 45.5.2 BGP session down
- scope: devices role=edge
- condition: bgp.session_state != Established
- for: 60s

#### 45.5.3 SLA loss rule (site to DNS)
- scope: all sites
- condition: ping.loss_pct > 2
- for: 5m
- severity: warn, critical at >10%

**Implementation note:** keep multi-threshold in one rule to avoid duplication.

---

## 46. Rule evaluation engine (how to implement it safely)

### 46.1 Avoid evaluating rules “on ingestion” only
If you evaluate rules only at ingestion time:
- missing samples can prevent evaluation
- backfilled data can fire late alerts incorrectly
- complex conditions become hard

Recommended approach:
- ingest writes metrics/events quickly
- a separate **evaluation scheduler** evaluates rules periodically

### 46.2 Evaluation scheduling
- every 10s: evaluate fast rules (reachability, BGP)
- every 60s: evaluate medium rules (errors, utilization)
- every 5m: evaluate slow rules (SLA windows, trend rules)

### 46.3 Windowed query model
To evaluate a rule, engine queries:
- last N minutes of metrics for each entity in scope
- compute aggregates:
  - avg, max, min
  - percent of samples failing
  - consecutive failures count

### 46.4 Stateful transitions (PENDING → FIRING)
Store per-alert instance state:
- when condition first observed true (pending_since)
- when fired (starts_at)
- last notified at (cooldown)

This prevents spam and provides consistent behavior across restarts.

### 46.5 Hysteresis (recovery thresholds)
For thresholds, define separate recover threshold:
- fire at > 100ms RTT for 5m
- recover at < 80ms RTT for 5m

This reduces alert flapping.

### 46.6 Flap detection implementation
Maintain a small history:
- count transitions in a rolling window
- if transitions > X in Y minutes, fire “flapping” alert and suppress individual flaps

---

## 47. Correlation: turning alerts into incidents

### 47.1 What correlation must achieve
During a PoP outage, 200 devices might generate alerts. Correlation must:
- group them
- identify root cause candidates
- suppress downstream notifications
- keep evidence timeline

### 47.2 Correlation keys (practical v1)
Group alerts by:
- tenant
- time window (e.g., 60–120s)
- topology neighborhood (same site or dependent subtree)

Start simple:
1. group by site_id for site-level outages
2. group by root cause candidate computed by RCA

### 47.3 Incident creation policy
- create incident when:
  - a critical alert fires OR
  - N warn alerts fire in M minutes in the same site
- update incident as more alerts arrive
- close incident when:
  - all related alerts recovered for > cooldown

### 47.4 Acknowledgement semantics
- ack applies to an incident or a specific alert
- ack should:
  - optionally silence notifications for that incident
  - still update UI state
  - expire (optional) to avoid “stale ack forever”

---

## 48. Diagnostics execution model (safe and auditable)

Diagnostics are a “premium” feature if done well.

### 48.1 Types of diagnostics
- ping a target from an agent PoP (agent OS ping)
- traceroute from agent PoP (agent OS)
- (optional) run ping/traceroute from router itself via RouterOS API (higher risk)

### 48.2 Safety constraints
- rate limit diagnostics per user and per tenant
- allow only safe commands (no arbitrary shell)
- ensure outputs are sanitized
- audit every run (who, when, what target)

### 48.3 Recommended implementation
Start with agent OS-based probes:
- agent has built-in ping/traceroute functions
- backend requests diagnostic job via:
  - HTTP long poll
  - websocket
  - queue (NATS) if present
- agent returns results + hop details

Store:
- diagnostic run record
- output payload
- related incident id

### 48.4 Target selection and reachability
Operators may want traceroute to:
- customer gateway
- DNS servers
- upstream IPs
- public endpoints (8.8.8.8, 1.1.1.1)

Ensure:
- agent has routing to those targets
- for private targets, use correct PoP agent

---

## 49. UI design patterns that support RCA

### 49.1 Evidence panel (must-have)
When an incident is opened, show:
- root cause candidate
- confidence score
- evidence list with timestamps:
  - interface down
  - OSPF down
  - ping loss spike
  - syslog link-down

### 49.2 Timeline view (metrics + events)
A good timeline overlays:
- ping loss/RTT
- interface utilization
- BGP state transitions
- event markers (syslog, neighbor down)

### 49.3 Blast radius visualization
- map: highlight impacted sites
- graph: highlight dependency subtree
- numbers: impacted routers, links, subscribers (if PPPoE modeled)

### 49.4 “Why suppressed?” UI
If an alert is suppressed, show:
- suppressed_by incident
- upstream root cause entity
- link to upstream evidence

This builds operator trust.

---

## 50. “Continue” note
Next chapters will cover:
- multi-tenant isolation deep dive (RLS vs app-level enforcement)
- per-tenant encryption keys and secrets strategy
- Docker Compose profiles with concrete service configs (example)
- scaling ingest with queues and worker pools
- adding SNMP/syslog alongside RouterOS API without bloat
- advanced topology inference (LLDP, OSPF, BGP graph heuristics)
- compliance and data minimization for subscriber-level PPPoE telemetry

Say **continue** to proceed.