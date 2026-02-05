# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XVII
Earlier parts cover architecture, data contracts, storage ops, and release engineering. This part defines the “operator dashboards” your UI should implement, a service model for ISP business impact, incident postmortems/exports, and advanced correlation ideas.

---

# Part XVIII — Operator Dashboards, Service Model, Postmortems, and Advanced Correlation

## 127. Operator dashboards (what to build first to be useful)

ISPs will interact with your platform primarily through a few screens. These dashboards must be:
- fast
- low-noise
- actionable
- role-aware (NOC vs admin)

### 127.1 NOC Home dashboard (primary screen)
Purpose: “What is broken right now and what should I do first?”

Recommended widgets:
1. **Active incidents list (top)**
   - severity, start time, status
   - root cause candidate (if known)
   - impacted sites/customers estimate
   - ack/assign shortcuts

2. **Service health tiles**
   - Internet Transit (overall)
   - Core routing (OSPF/iBGP)
   - Access services (PPPoE)
   - DNS service (if monitored)
   Each tile shows: OK/degraded/down + last 24h trend sparkline.

3. **Map mini-view**
   - only show sites with issues (filter)
   - click to expand to full map

4. **Top degradations**
   - highest packet loss sites
   - highest RTT sites
   - top congested links

5. **Agent health**
   - offline agents
   - agents with high buffer

Design note:
- the home page should load in <2 seconds for mid-size tenants.
- use cached summary endpoints.

### 127.2 Map dashboard (geo + link overlay)
Purpose: “Where is the issue physically?”

Core features:
- site status and incident overlay
- link utilization overlay
- filters by tag/role/provider
- click site → site detail view

Advanced (later):
- time travel (“status at time T”)
- heatmaps (loss/RTT by region)

### 127.3 Topology dashboard (dependency graph)
Purpose: “What depends on what? What is the blast radius?”

Features:
- role-based view (core/agg/access)
- show redundancy groups (ANY/ALL)
- incident overlay showing suppressed/impacted nodes

### 127.4 Device detail dashboard
Purpose: “What is happening on this router?”

Sections:
- status summary and last change
- interface list with utilization/errors
- BGP peers list with states and prefix counts
- OSPF neighbors list
- recent events timeline (syslog-derived and system events)
- probes related to this device (if loopback target exists)

### 127.5 Link/interface detail dashboard
Purpose: “Is this link healthy? Is it congested or degrading?”

Sections:
- utilization time-series (1h/24h/7d)
- error rate time-series
- SLA correlation (loss/RTT) for paths that depend on this link
- incidents involving this link historically

### 127.6 Incidents dashboard (work queue)
Purpose: “How do we manage incidents operationally?”

Features:
- filters (severity, site, status, assignee)
- SLA impact estimate per incident
- export incident timeline (PDF/CSV later)
- postmortem generation (template)

### 127.7 Capacity planning dashboard
Purpose: “What will break soon?”

Widgets:
- top links by p95 utilization
- forecasted saturation (days to 90%)
- top PoPs by growth rate
- recommended upgrade list

### 127.8 Reports dashboard (SLA)
Purpose: “What do we report to customers or management?”

Outputs:
- uptime per site/link/service
- MTTR/MTBF trends
- monthly summaries
- scheduled exports

---

## 128. ISP service model (turn telemetry into business impact)

To make your product “paid-grade,” you need a model that maps network components to **services**.

### 128.1 Why service modeling matters
Operators don’t want “router X down.” They want:
- “Internet service in PoP A degraded”
- “Upstream Provider Y down”
- “PPPoE service impacted for 3,200 subscribers”

### 128.2 Define service entities
Introduce a `services` table:
- `id`, `tenant_id`
- `name`
- `service_type` enum:
  - `internet_transit`
  - `core_routing`
  - `dns`
  - `pppoe_access`
  - `site_connectivity`
  - `customer_service` (later)
- `site_id` nullable (global vs per site)
- `description`
- `enabled`

Then define dependencies:
- `service depends_on device/link/probe_target/service`

### 128.3 Service health computation
Service health should be derived from:
- probe results (primary)
- routing health (secondary)
- link utilization/errors (secondary)
- device reachability (supporting)

Example:
- `internet_transit` service:
  - depends_on `edge BGP` + upstream probe targets (1.1.1.1/8.8.8.8)
- `core_routing` service:
  - depends_on OSPF adjacency health and iBGP RR health
- `pppoe_access` service:
  - depends_on access concentrators + churn metrics

### 128.4 Service “SLOs” and SLAs
You can allow per service:
- availability definition:
  - e.g., loss < 20% AND at least 1 of 2 probes is up
- latency objectives:
  - e.g., p95 RTT < 40ms
- allowed downtime per month

Then compute:
- SLA compliance per month
- incident-driven breach explanations

---

## 129. Incident postmortems (MTTR reduction and enterprise credibility)

### 129.1 Why postmortems are valuable
ISPs need internal accountability and continuous improvement.
Providing postmortem tooling makes you “enterprise-grade.”

### 129.2 Postmortem template
For each incident:
- Summary
- Impact
- Timeline (auto-generated from events/metrics)
- Root cause (chosen by operator; can override algorithm)
- Contributing factors
- Detection and response notes
- Mitigations performed
- Action items (tickets)
- SLA breach analysis
- Lessons learned

### 129.3 Auto-generated timeline
Use stored events + alert transitions + probe changes to generate:
- “10:01:23 BGP EDGE1-UPSTREAM1 down”
- “10:01:45 Loss to 1.1.1.1 rose to 100% from PoP A”
- “10:02:10 Incident created”
- “10:08:40 Recovery observed”
- “10:10:00 Incident closed”

### 129.4 Export formats
Start with:
- Markdown export
- JSON export (for integration)
Later:
- PDF export (pandoc pipeline)

---

## 130. Advanced correlation ideas (v1.2+ roadmap)

### 130.1 Burst correlation (syslog storm)
If 200 “interface down” syslogs occur within 30 seconds in the same site:
- create a single “site power outage” suspicion
- raise confidence for site-level root cause
- suppress per-interface spam

### 130.2 Multi-signal correlation: congestion
Congestion is best detected by:
- utilization high
- RTT rising
- loss rising
- drops increasing
- sometimes CPU high

Correlate these into a single “congestion incident” with evidence.

### 130.3 Prefix drop anomalies
If prefixes drop but sessions remain established:
- route leak or upstream filtering
- this is high severity and often missed by simplistic tools

Requires:
- stable baseline
- careful thresholding
- operator review

### 130.4 Configuration change correlation
If a failure occurs soon after a config change:
- highlight “change as potential cause”
- include config snapshot hash change event in evidence

This is extremely valuable in NOC operations.

---

## 131. “Continue” note
Next chapters can include:
- how to implement config change detection safely (hashing vs storing configs)
- how to build a “change calendar” and maintenance workflows
- detailed gRPC/NATS transport designs for large scale
- advanced security: per-tenant keys, HSM/KMS integration, Vault patterns
- deeper service modeling: customer services, upstream provider SLAs

Say **continue** to proceed.