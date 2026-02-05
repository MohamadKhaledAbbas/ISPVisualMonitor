# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XXII
Earlier parts cover webhook spec, integration catalog, operator handbook, STRIDE notes, and v1.0 checklist. This part adds: concrete integration examples (Jira + ServiceNow), a more formal STRIDE table template, ISP “blue team” hardening checks, and a demo-kit automation blueprint.

---

# Part XXIII — Concrete Integration Examples, Formal STRIDE Template, Blue‑Team Checks, and Demo‑Kit Automation

## 154. Concrete integration example: webhook → Jira (minimal, reliable)

### 154.1 Goal
Demonstrate a practical, low-friction workflow:
- ISPVisualMonitor sends incident webhooks
- a small receiver service creates/updates a Jira issue
- no bi-directional complexity required

### 154.2 Architecture
- ISPVisualMonitor → HTTPS webhook → `jira-bridge` (small service)
- `jira-bridge` → Jira REST API
- Jira issue key stored in `jira-bridge` DB (or a small KV store)

### 154.3 Mapping rules
- incident severity → Jira priority
- incident status:
  - opened/updated → keep issue open
  - closed → transition issue to Done/Resolved (optional)
- labels:
  - `ispvm`
  - provider/site tags

### 154.4 Idempotency keying
Use incident ID as the stable external key:
- Jira issue summary includes `[ISPVM] <incident_title>`
- `jira-bridge` stores mapping:
  - incident_id → jira_issue_key

If mapping exists:
- update issue fields
- add comment “Incident updated: root cause changed…”

### 154.5 Example webhook receiver behavior
On `incident.opened`:
- create Jira issue with description including:
  - deep link URL
  - start time
  - impact counts
  - evidence_top
On `incident.updated`:
- update description and add comment with diff summary
On `incident.closed`:
- add comment with duration and close reason
- optionally transition workflow

### 154.6 What to document for customers
- required Jira credentials (API token)
- required Jira project key
- how to map priorities

---

## 155. Concrete integration example: webhook → ServiceNow (incident record)

### 155.1 Typical ServiceNow mapping
ServiceNow has:
- `incident` records
- assignment groups
- categories and subcategories

Map ISPVisualMonitor incidents to:
- category: `Network`
- subcategory: `Monitoring`
- short_description: incident title
- description: evidence + impact + URL

### 155.2 Assignment routing (important)
Route based on:
- site tag or region tag
- provider tag
- role/core vs access classification

This can be done:
- in a custom bridge service
- or inside ServiceNow integration logic

### 155.3 Idempotency and updates
Use external correlation ID:
- `u_ispvm_incident_id` custom field in ServiceNow

On updates:
- append work_notes
On close:
- set state resolved/closed with resolution notes.

---

## 156. Formal STRIDE template (table form)

Use this template to produce an appendix that enterprises can review.

### 156.1 STRIDE table columns
- Asset/Component
- Threat (S/T/R/I/D/E)
- Scenario
- Impact
- Likelihood (Low/Med/High)
- Mitigations
- Residual risk
- Monitoring/Detection

### 156.2 Example STRIDE entries (starter)

| Asset | Threat | Scenario | Impact | Likelihood | Mitigations | Residual | Detection |
|------|--------|----------|--------|------------|-------------|----------|-----------|
| Agent identity | Spoofing | Attacker uses stolen cert to impersonate agent | False data + missed outages | Med | mTLS, cert rotation, revocation, rate limits | Low/Med | Audit logs, anomaly on agent_id |
| Ingest endpoint | DoS | Agent misconfig floods ingest | Ingest backlog, alert lag | Med/High | rate limits, queue, backoff | Med | queue depth alerts |
| Tenant data | Info disclosure | Missing tenant filter leaks data | Critical breach | Low if RLS, Med if app-only | RLS policies, tests | Low | security tests, audit review |
| Credentials | Tampering | DB compromise modifies encrypted blobs | Monitoring access loss | Med | integrity checks, key versions, audit | Med | secret access audit anomalies |

(Expand to full table for v1.0 enterprise review.)

---

## 157. ISP “blue team” hardening checks (deployments and operations)

This section is designed as a checklist ISPs can apply.

### 157.1 Network perimeter
- [ ] Backend exposed only via 443
- [ ] Admin UI restricted by IP allowlist (optional)
- [ ] Agent to backend connectivity is outbound-only where possible
- [ ] Router management network is isolated (mgmt VLAN/VRF)
- [ ] Firewall restricts RouterOS API/SNMP to agent IPs only

### 157.2 TLS and certificates
- [ ] TLS 1.2+ enforced
- [ ] Valid certificates (ACME or internal CA)
- [ ] Agent cert rotation tested
- [ ] Certificate expiry alerts configured

### 157.3 Secrets management
- [ ] Master key stored outside repo and outside container image
- [ ] Docker secrets or mounted files used for sensitive env vars
- [ ] Credential access audited (admin actions)

### 157.4 OS hardening (backend host)
- [ ] Automatic security updates (policy-driven)
- [ ] minimal open ports
- [ ] log shipping to centralized system (optional)
- [ ] filesystem permissions for volumes

### 157.5 Database hardening
- [ ] DB not exposed publicly
- [ ] strong DB password or cert auth
- [ ] backup encryption at rest
- [ ] PITR configured (if needed)

### 157.6 Monitoring the monitoring system
- [ ] Prometheus scraping enabled
- [ ] alerts for:
  - ingest lag
  - agent offline
  - DB disk usage
  - queue depth
  - notification failures

---

## 158. Demo kit automation blueprint (sales-grade repeatability)

### 158.1 Why a demo kit
A repeatable demo kit allows you to:
- show the product consistently
- run scripted outages
- generate screenshots for marketing
- test regression in a realistic scenario

### 158.2 Proposed demo kit structure
Create a separate repo or folder:
- `demo-kit/`
  - `compose/` (backend + db + proxy + optional grafana)
  - `eve-ng/` (lab topology export files)
  - `seed/` (demo seed bundles)
  - `scripts/`
    - `start.sh` / `start.ps1`
    - `seed.sh`
    - `simulate-core-link-cut.sh`
    - `simulate-upstream-down.sh`
    - `simulate-congestion.sh`
  - `docs/` (demo steps and narration)

### 158.3 Simulation scripts (principles)
Scripts should:
- be idempotent (re-run safe)
- verify prerequisites
- print “expected outcomes”:
  - “incident should appear within 60s”
- optionally call UI API to confirm incident existence (automated validation)

### 158.4 Demo storyline (recommended narrative)
1. Show NOC Home: all green
2. Trigger core failure
3. Show RCA + suppression
4. Trigger upstream failure
5. Show service-based incident classification
6. Show capacity planning and forecast
7. Export a sample SLA report and incident postmortem

This story sells the value proposition.

---

## 159. “Continue” note
Next chapters could include:
- a complete “operator handbook” written out (not just an outline)
- a full “telemetry registry v1” as a separate appendix document
- a “rule pack v1” definition set ready to import
- detailed UI wireframe specs (components, routes, APIs)
- a full threat model with diagrams (data flow, trust boundaries)

Say **continue** to proceed.