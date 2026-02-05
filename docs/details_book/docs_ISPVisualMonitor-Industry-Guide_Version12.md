# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XI
Earlier parts cover architecture, storage, alerting/RCA, security, scaling, lab design, productization, starter packs, telemetry matrix, and roadmap. This part provides privacy/trust documentation templates, performance testing plans, UI/UX guidance for maps/graphs, and demo playbooks.

---

# Part XII — Privacy/Trust Policy, Performance Test Plan, UI/UX for Topology, and Sales-Grade Demo Playbooks

## 89. “Data Collected & Privacy” policy (template you can ship)

ISPs will ask exactly:
- What do you collect?
- Where is it stored?
- How long is it kept?
- Who can see it?
- How can it be deleted?

This section can be adapted into a `docs/privacy-and-data-collected.md` file later.

### 89.1 Data minimization principles
1. Collect only what you need to provide monitoring value.
2. Prefer aggregated counts over user-level details.
3. Make subscriber-level data **opt-in**, time-limited, and audited.
4. Always allow per-tenant configuration of retention.
5. Provide a “data inventory” list per version.

### 89.2 Categories of data collected

#### A) Device metadata (inventory)
Examples:
- device name, mgmt IP
- vendor/model
- RouterOS version
- site assignment, tags

**Purpose:** topology, dashboards, grouping  
**Sensitivity:** medium (network internal details)

#### B) Operational metrics (time-series)
Examples:
- ping loss/RTT
- interface counters and errors
- CPU/memory usage
- BGP session state

**Purpose:** health, alerts, SLA reports  
**Sensitivity:** low-to-medium (still internal network telemetry)

#### C) Events (discrete)
Examples:
- link up/down from syslog
- BGP state transitions
- agent connectivity events

**Purpose:** timelines, RCA evidence  
**Sensitivity:** medium (contains network details)

#### D) Credentials and secrets
Examples:
- RouterOS API credentials, SNMPv3 keys

**Purpose:** collection access  
**Sensitivity:** critical  
**Policy:** encrypted at rest; access-controlled; never exposed to non-admin roles; audited.

#### E) Subscriber/customer data (optional)
Examples:
- PPPoE usernames, IPs, session logs (if enabled)

**Purpose:** customer impact and support correlation  
**Sensitivity:** high / personal data  
**Policy:** off by default; opt-in; minimized; strict RBAC; shorter retention; audited access.

### 89.3 Storage and retention defaults (recommended)
Define defaults that can be customized:

- raw metrics (10–30s resolution): 30 days
- 1m aggregates: 12 months
- 1h aggregates: 3–5 years (SLA)
- events: 90–180 days
- incident records: 1–3 years
- subscriber-level data: 7–30 days (opt-in)

### 89.4 Data access control model (RBAC)
Recommended baseline roles:
- `Admin`: manage everything, including credentials
- `NOC`: view incidents, acknowledge, view topology, run diagnostics
- `ReadOnly`: view dashboards and reports
- `Auditor`: view audit logs and compliance exports

Sensitive areas:
- credentials
- subscriber-level data
- exports

### 89.5 Data deletion and tenant offboarding
You should support:
- delete a device and associated telemetry (within retention constraints)
- delete a tenant (hard delete or scheduled purge)
- export tenant data (optional enterprise)

Even if you cannot implement full purge early, document:
- what is possible now
- what is planned

### 89.6 Telemetry transparency (agent trust)
Publish:
- a list of exact metric names and event types
- sample payloads
- how to disable certain collectors
- how to run the agent in “restricted mode” (no diagnostics)

---

## 90. Performance test plan (how to prove you’re production-grade)

### 90.1 Why performance testing is part of “industry-grade”
ISPs will not deploy tools that:
- fall behind during incidents
- crash under alert storms
- freeze dashboards

A performance plan provides confidence and guides capacity planning.

### 90.2 What to test (workload dimensions)
1. **Ingest throughput**
   - metrics/sec
   - events/sec
2. **DB performance**
   - write latency
   - chunk compression effects
   - query latency for dashboards
3. **Rule evaluation**
   - number of rules × number of entities
   - window sizes (5m/10m)
4. **Incident correlation**
   - 1 event vs 10k events in 1 minute
5. **Notification dispatch**
   - burst of 500 notifications
6. **UI**
   - map load time with 500 sites
   - incident list load with 10k incidents

### 90.3 Load generation approaches

#### A) Synthetic load generator (recommended)
Build a small tool:
- produces ingest payloads matching your schema
- simulates N agents and M routers
- produces realistic distributions (not uniform)

This avoids depending on EVE-NG scale.

#### B) EVE-NG “real” lab
Use EVE for functional correctness and small-scale perf.

### 90.4 Example test targets (measurable)
- ingest p95 latency < 300ms at 5k samples/sec
- rule eval cycle completes < 5s for 300 routers, 50 rules
- UI dashboard loads < 2s for last 1h graphs
- DB disk growth stays within expectations with compression

### 90.5 Regression testing
Add performance smoke tests in CI:
- run ingest batch tests
- run rule evaluation benchmarks

---

## 91. UI/UX guidance: maps + dependency graphs for NOC usability

### 91.1 Why map-only UIs fail
A map alone cannot show:
- routing dependencies
- redundant paths
- logical service relationships

Therefore the UI should offer:
- **Map view** (geo)
- **Graph view** (logical dependency)
- **Timeline view** (metrics + events)
- **Table views** (NOC workflows, filtering)

### 91.2 Map view design (industry patterns)
Features:
- clustering for dense site regions
- status rings (OK/degraded/down)
- link overlays:
  - utilization color
  - health color
  - “suppressed” styling
- layers:
  - show/hide: core, access, customer-edge
- time range selector for “what was status at time T” (later)

**Performance tip:** precompute summary status per site and cache it.

### 91.3 Dependency graph design (must be actionable)
Graph must:
- show root cause candidate at top
- show impacted subtree
- support redundancy:
  - multiple upstreams
- show evidence and statuses on edges

Graph layout:
- hierarchical (core → agg → access → services)
- avoid force-directed layouts for large graphs (too chaotic)
- allow “collapse” by site or role

### 91.4 Incident page layout (recommended)
Three columns:
1. Summary + root cause + impact numbers
2. Evidence + timeline
3. Entities list (affected sites/devices/services), searchable

### 91.5 “Explainability” UI
Make it clear:
- which signals triggered alert
- why suppressed
- confidence score and evidence weights

ISPs trust tools that explain themselves.

---

## 92. Demo playbooks (how to present a compelling product)

You need a repeatable demo environment to sell.

### 92.1 Demo environment goals
- deploy in < 30 minutes
- show multiple PoPs and routers
- show realistic outages and RCA in a scripted way
- produce screenshots and reports

### 92.2 Recommended demo setup
- EVE-NG lab from Part IX (core, edge, upstream FRR)
- Agent runs in `TOOLS` node
- Backend runs on host or separate VM
- Predefined sites with coordinates (fake but realistic)
- Predefined alert packs enabled

### 92.3 Scripted demos (what to show)

#### Demo 1: Core link failure with suppression + RCA
Steps:
1. show map “all green”
2. cut core link
3. show incident created
4. show RCA root cause “CoreLink A–B”
5. show impacted sites count
6. show suppressed downstream alerts (“why suppressed”)

#### Demo 2: Upstream BGP failure (internet degraded but internal ok)
Steps:
1. show internal probes still OK
2. upstream probe to 1.1.1.1 fails
3. BGP edge down alert fires
4. incident classified “Transit outage”
5. show SLA impact on “Internet Service” entity

#### Demo 3: Congestion leading to degradation + forecast
Steps:
1. saturate a link with iperf
2. show utilization > 95%
3. show RTT rise and loss
4. show “degradation incident”
5. show “capacity planning” view with predicted saturation

#### Demo 4: PPPoE churn spike (optional)
Steps:
1. simulate session churn (in lab or synthetic)
2. show “session drop spike” alert
3. show estimated subscriber impact

### 92.4 What not to demo early
- complex ML anomaly models (hard to explain)
- vendor-wide claims beyond MikroTik before you support them well

---

## 93. Support and enterprise expectations (appendix guidance)

### 93.1 What ISPs expect in paid support
- guaranteed response times
- upgrade assistance
- incident assistance (help interpreting alerts)
- security patches quickly
- roadmap commitments (to a degree)

### 93.2 Support tiers (example)
- Community: GitHub issues, best effort
- Standard: business hours support, patch releases
- Premium: 24/7, dedicated Slack channel, DR planning help

Even if you don’t implement this now, document the intended model.

---

## 94. “Continue” note
Next chapters can go even deeper into:
- subscriber/service modeling for PPPoE with RADIUS integration patterns
- topology model extensions for rings, ECMP, and redundancy groups
- advanced RCA (probabilistic scoring, multi-cause incidents)
- pricing and packaging strategies (what to include per tier)
- writing the *actual* developer documentation for your repos:
  - API reference
  - agent configuration reference
  - contribution guidelines

Say **continue** to proceed.