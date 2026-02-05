# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XIX
Earlier parts cover change detection, maintenance workflows, and key management. This part provides a RouterOS change-detection allowlist, a concrete workflow spec for maintenance/change/incidents, SIEM integration patterns, and compliance-oriented retention and audit export signing.

---

# Part XX — RouterOS Change Allowlist, Workflow Spec, SIEM Integration, and Compliance Add‑Ons

## 137. RouterOS change detection allowlist (what to hash, what to exclude)

### 137.1 Goal: “useful correlation” without storing secrets
The aim is to detect *meaningful* changes that commonly cause outages, while excluding:
- passwords and secrets
- certificates private keys
- user databases
- anything that increases legal/privacy risk without operational value

### 137.2 Allowlist categories (recommended)

#### A) Interface admin and L2/L3 critical settings
Include:
- interface admin state (enabled/disabled)
- bridge membership changes (if used in core/access)
- VLAN interfaces / tagging changes (if you rely on VLANs)
- IP address assignments on critical interfaces
- VRF assignment changes (if used)

Why:
- interface disables and VLAN mistakes cause immediate outages.

#### B) Routing protocol configuration
Include:
- BGP peer definitions (remote addr/as, enable/disable)
- BGP filters/policies references (names/ids, not full content if huge)
- OSPF instances, areas, interface templates
- route reflector configuration (if present)
- default route presence/next-hop (summary-level)

Why:
- routing changes are a major outage source.

#### C) Firewall/NAT “summary only” (optional)
Include:
- rule counts by chain
- last modified time if available
- hashes of normalized rule list *excluding comments* (advanced)

Why:
- firewall changes can break access, but full firewall config can be large and sensitive.
Start with summary-level detection.

#### D) Services that affect monitoring and operations
Include:
- API/API-SSL enablement and listen addresses (no credentials)
- SNMP enablement and version (no communities / v3 keys)
- syslog remote targets (if configured)
- NTP servers and time sync settings

Why:
- misconfig here breaks visibility and correlation.

#### E) System identity metadata (low risk)
Include:
- router identity/name (if relevant)
- RouterOS version changes
- reboot events (already an event type)

Why:
- upgrades and reboots correlate with incidents.

### 137.3 Explicit exclude list (must not collect/store)
Exclude:
- user password hashes
- secrets/keys
- SNMP communities and SNMPv3 auth/priv keys
- PPP secrets (PPPoE credentials)
- certificate private keys
- any RADIUS shared secrets
- any exported config files that contain secrets, unless you implement redaction

### 137.4 Fingerprinting algorithm (practical)
For each category:
1. collect normalized JSON:
   - stable key ordering
   - remove volatile fields (counters, timestamps)
2. compute SHA-256 hash per category
3. compute overall hash = SHA-256(concat(category_hashes))

Store:
- per-category hash
- overall hash
- timestamp

When a category hash changes:
- emit event `config.changed.detected` with `category`.

---

## 138. Workflow specification: incidents, maintenance, and changes (concrete UX rules)

### 138.1 Primary states (entities)
- **Alert instance**: ok/pending/firing/suppressed
- **Incident**: open/closed
- **Maintenance window**: active/inactive
- **Change record**: planned/in-progress/completed/canceled

### 138.2 Core workflow: “alert → incident”
1. Alert fires
2. Correlation engine decides:
   - attach to existing incident, or
   - create new incident
3. RCA runs and sets root cause candidate(s)
4. Notifications dispatched (unless suppressed)

### 138.3 Acknowledgement rules (avoid confusion)
- Ack can be applied to:
  - incident (preferred)
  - alert instance (advanced)
- Ack results:
  - stops repeated notifications for that incident
  - still updates incident and UI state
- Ack should be visible in:
  - incident header (who/when)
  - audit log

Optional:
- ack expiry (e.g., 4h) to prevent stale acknowledgements.

### 138.4 Maintenance window behavior (what should happen)
When maintenance is active for scoped entities:
- alerts may still evaluate and change state
- incidents may still be created, but flagged:
  - `during_maintenance_id`
- notifications are suppressed by default
- UI displays “maintenance” badge prominently

Allow override:
- “critical alerts still notify” (for high-severity categories)

### 138.5 Change record behavior
When a change record is active:
- similar behavior to maintenance, but with different semantics:
  - incidents flagged `during_change_id`
- after change completes:
  - show all incidents during window
  - allow operator to link root cause to the change explicitly

### 138.6 Post-incident workflow
After incident is closed:
- offer “Generate postmortem” (auto template)
- operator confirms root cause and contributing factors
- action items created (internal list or external integration)

---

## 139. SIEM integration patterns (exporting security/ops events)

ISPs (and enterprises) often have SIEM tooling.

### 139.1 What to export to SIEM
- authentication events (login success/fail)
- RBAC/permission changes
- credential set create/update/delete
- agent enroll/revoke
- diagnostics runs (who traced what)
- incident state changes (optional)
- audit logs (primary)

### 139.2 Export methods
1. **Syslog output**
   - emit your audit logs as syslog lines
   - compatible with many SIEMs

2. **Webhook**
   - POST JSON to SIEM collector endpoint
   - include signing key / HMAC for integrity

3. **File export**
   - periodic JSONL export for offline ingestion

### 139.3 Log format guidance
Use structured JSON:
- `ts`
- `tenant_id`
- `actor_type`, `actor_id`
- `action`
- `entity_type`, `entity_id`
- `ip_address`
- `user_agent` (for UI actions)
- `details`

This makes SIEM parsing reliable.

---

## 140. Compliance add-ons (retention by category, legal holds, signed exports)

### 140.1 Retention by category (enterprise-friendly)
Allow per tenant:
- metrics retention (raw + rollups)
- events retention
- incident retention
- audit log retention
- subscriber data retention (if used)

This allows “data minimization” compliance.

### 140.2 Legal hold concept (advanced)
Some tenants may require:
- “do not delete incident/audit data for date range”
Implement as:
- legal hold policy that overrides retention deletion for selected categories

### 140.3 Signed audit exports
If audit logs are used for compliance, export integrity matters.

Approach:
- export audit logs as JSONL
- compute file hash (SHA-256)
- sign the hash using:
  - GPG key or Sigstore
- store signature metadata and allow verification

This is a strong enterprise differentiator.

---

## 141. “Continue” note
Next chapters can include:
- detailed integration with external ticketing (Jira/ServiceNow) as incident sinks
- detailed “action items” tracking and SLO improvements
- implementing “operator feedback” loops to improve RCA weights
- deeper upstream provider SLA modeling and per-provider reporting
- a full “appendix: glossary” for ISP terms to help new users

Say **continue** to proceed.
````````markdown name=docs/ISPVisualMonitor-Industry-Guide.md
# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XX
Earlier parts cover RouterOS change allowlists, workflow specs, SIEM integration, and compliance add-ons. This part adds: ticketing integrations, action-item/SLO tracking, RCA feedback loops, upstream provider SLA modeling, and a glossary appendix.

---

# Part XXI — Ticketing Integrations, Action Items & SLOs, RCA Feedback Loops, Provider SLAs, and Glossary

## 142. Ticketing integrations (Jira/ServiceNow/Zendesk) as incident sinks

### 142.1 Why ticketing integration matters
ISPs already have workflows:
- NOC opens tickets
- field teams execute repairs
- management tracks MTTR

If your system can create/update tickets automatically, you become “part of the process.”

### 142.2 Integration modes
1. **Webhook-first (recommended)**
   - on incident open/update/close, POST JSON to a configurable webhook
   - ISPs can route this to their ticketing automation

2. **Native connectors (later)**
   - Jira: create issue, update status, add comments
   - ServiceNow: incident records, CMDB linking
   - Zendesk: support tickets

Start webhook-first to keep scope controlled.

### 142.3 What to send (payload fields)
- incident id, title, severity
- started_at, last_updated_at
- root cause candidate(s) + confidence
- impacted sites count, impacted subscribers estimate
- evidence summary (top 5 signals)
- URL to incident in ISPVisualMonitor

### 142.4 Idempotency and updates
Ticket creation must be idempotent:
- use incident id as external correlation key
- on repeated calls, update existing ticket not create duplicates

### 142.5 Bi-directional integration (optional)
Later, allow:
- ticket status updates to close/ack incident
- comments synced back
This is powerful but complex; treat as enterprise feature.

---

## 143. Action items and SLO tracking (operational maturity)

### 143.1 Why action items matter
Postmortems without action items don’t reduce outages.
Your platform can help ISPs mature:
- track recurring causes
- measure improvements

### 143.2 Action item model (lightweight)
Tables:
- `action_items`
  - `id`, `tenant_id`
  - `incident_id` (nullable)
  - `title`, `description`
  - `owner_user_id`
  - `due_at`
  - `status` (open/in_progress/done/canceled)
  - `created_at`, `updated_at`

### 143.3 SLOs (internal targets)
SLOs are internal operational goals (not customer SLAs necessarily).

Examples:
- “Detect critical outages within 60 seconds”
- “Acknowledge critical incidents within 5 minutes”
- “MTTR for transit outage < 30 minutes”

### 143.4 Metrics to compute SLOs
- detection latency:
  - incident_started_at vs first symptom timestamp
- response latency:
  - ack timestamp - incident started
- repair time:
  - incident closed - incident started

Expose:
- per month SLO compliance
- trends and leaderboards

This becomes a management dashboard and a paid feature.

---

## 144. Operator feedback loops (improving RCA over time)

### 144.1 Why feedback is needed
Automated RCA will be wrong sometimes.
Allow operators to correct it:
- increases trust
- improves algorithm tuning

### 144.2 Feedback capture
In the incident UI:
- “Confirm root cause” (select entity or “unknown”)
- “Contributing factors” list
- “Was suppression correct?” yes/no
- optional notes

Store:
- `incident_feedback`
  - incident_id
  - chosen_root_cause_entity
  - correctness flags
  - timestamp, author

### 144.3 Using feedback to tune evidence weights (safe approach)
Do not do “black box ML” early. Do:
- per evidence type success rate
- adjust weights slowly and conservatively
- allow per-tenant static tuning (some ISPs trust syslog more than probes)

### 144.4 Explain changes
If weights change, record:
- what changed
- why (based on feedback)
- when
So operators can trust the evolution.

---

## 145. Upstream provider SLA modeling (ISP business reporting)

### 145.1 Why provider SLAs are valuable
ISPs pay upstreams and peers; they need:
- proof of outages
- performance reports
- negotiation leverage

### 145.2 Model upstream providers as services
Create entities:
- `providers`
  - id, name, contact info
- `provider_services`
  - per provider transit link(s)
  - associated BGP peers
  - probe targets (public endpoints and provider next-hop)

Then compute:
- provider availability
- provider latency/loss metrics per PoP
- incidents attributed to provider outages

### 145.3 Attribution rules (v1)
An “Upstream Provider Outage” is likely when:
- eBGP down to provider OR prefixes drop drastically
AND
- internal core routing is healthy
AND
- multiple PoPs show loss to public endpoints

### 145.4 Provider report outputs
Monthly provider report includes:
- total downtime minutes (derived from incident windows)
- number of incidents
- p95 RTT to provider test endpoints
- top affected PoPs
- evidence attachments (event timeline)

This is a strong selling point.

---

## 146. Appendix: Glossary (ISP-friendly)

### 146.1 Routing and topology
- **AS (Autonomous System):** A network under one admin domain with an ASN.
- **BGP:** Border Gateway Protocol (inter-domain routing).
- **eBGP:** BGP between different ASNs (ISP ↔ upstream/customer/peer).
- **iBGP:** BGP within the same ASN (distributes external routes internally).
- **Route Reflector (RR):** iBGP scaling mechanism to avoid full mesh.
- **OSPF:** Interior Gateway Protocol used inside ISPs.
- **IGP:** Interior Gateway Protocol (OSPF/IS-IS).
- **ECMP:** Equal-Cost Multi-Path routing (multiple next-hops).
- **PoP:** Point of Presence (site).
- **Blast radius:** Set of impacted entities/services.

### 146.2 Monitoring concepts
- **Alert:** A stateful condition for an entity (firing, ok, suppressed).
- **Incident:** A correlated group of alerts representing a real-world event.
- **RCA:** Root Cause Analysis (identify upstream cause(s)).
- **Suppression:** Muting downstream alerts when upstream cause is active.
- **SLA:** Service Level Agreement (customer-facing commitments).
- **SLO:** Service Level Objective (internal target).
- **MTTR:** Mean Time To Repair/Recover.
- **MTBF:** Mean Time Between Failures.

### 146.3 Access service concepts
- **PPPoE:** Point-to-Point Protocol over Ethernet.
- **Concentrator:** Device terminating PPPoE sessions (access router).
- **RADIUS:** AAA protocol for authentication/accounting.

---

## 147. “Continue” note
Next chapters (if you want to go even further) can include:
- full webhook spec for ticketing + example payloads
- external integration catalog (Grafana, Prometheus, Loki)
- a complete “operator handbook” structure for self-hosted ISPs
- threat modeling appendix (STRIDE) for agent + backend
- a concrete acceptance test suite checklist for v1.0 launch

Say **continue** to proceed.