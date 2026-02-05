# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–IX
Earlier parts cover core architecture, schemas, alerting/RCA, security, scaling, operations, and lab topologies. This part focuses on **productization**: what you open-source, how you build trust, and what features make ISPs pay.

---

# Part X — Productization: Trust, Open‑Source Agent Strategy, and Commercial “ISP‑Grade” Features

## 75. Open‑source agent strategy (how to maximize trust without losing control)

### 75.1 What should be open source
To maximize operator trust, open-source:
- agent core runtime
- collector plugins that read from devices (RouterOS API, SNMP, syslog)
- transport code that sends metrics/events upstream
- data normalization (metric naming, event mapping)
- configuration format and schema

This allows an ISP to confirm:
- what is collected
- what is sent
- what is stored locally (buffer)
- how secrets are handled on the agent side

### 75.2 What can remain closed-source (if you want a commercial model)
It is common to keep the control plane proprietary while keeping the agent open.
You can keep closed:
- advanced RCA algorithms (if you consider it IP)
- advanced anomaly detection models
- some UI dashboards and reporting packs
- enterprise features (SAML, multi-region, compliance exports)

However, be careful:
- if the backend is closed, you must still document:
  - data retention
  - what data is stored
  - encryption controls
  - deletion policies

### 75.3 Agent licensing
Common patterns:
- Agent under Apache-2.0 (business-friendly) or MPL-2.0 (forces changes to be shared on modified files)
- Commercial license for backend
- Contributor License Agreement (CLA) optional

**Practical recommendation:** Apache-2.0 is simplest for adoption.

### 75.4 Reproducible builds and signed releases (trust multiplier)
ISPs will care about supply chain:
- publish build instructions
- produce deterministic builds when possible
- sign release binaries (Sigstore/cosign or GPG)
- publish SBOM (Software Bill of Materials)

### 75.5 Security vulnerability process
To be industry-grade, add:
- `SECURITY.md` with disclosure process
- CVE handling guidelines (even if informal)
- patch release cadence expectations

---

## 76. ISP onboarding experience (make adoption easy)

### 76.1 Why onboarding is a “paid feature”
ISPs judge tooling by:
- time to value (first map + first alert)
- how hard it is to integrate with their network
- clarity of docs and operational support

### 76.2 Recommended onboarding flow (wizard concept)
1. Create tenant/org
2. Add sites (or import CSV)
3. Enroll first agent (token + mTLS)
4. Add first router(s)
   - choose credential set
   - test connectivity
5. Tag devices by role
6. Add links (manual) or accept inferred neighbors
7. Enable starter alert packs
8. Configure notification channels
9. Run synthetic “failure test” (optional) to confirm alerts work

### 76.3 Templates and “starter packs”
Ship opinionated templates:
- Small ISP template (single agent)
- Multi-PoP template (agents per site)
- PPPoE template (access concentrator focus)
- Transit template (BGP edge focus)

Templates accelerate adoption and reduce misconfiguration.

---

## 77. Commercial feature blueprint (what makes it hard to resist)

This section describes features that directly translate to business value.

### 77.1 RCA + impact packs
Operators pay when your system answers:
- What failed?
- Where is the root cause?
- How many customers are impacted?
- What is the estimated downtime and SLA impact?
- What should the NOC do next?

Features:
- root cause confidence + evidence list
- blast radius estimation
- automatic suppression of downstream alerts
- incident timelines with “first symptom” detection
- “affected customer count” (PPPoE/leases/service model)

### 77.2 SLA reporting packs
ISPs sell SLAs. Provide:
- per site uptime
- per link uptime and congestion time
- per service (internet, voice, IPTV) availability
- exports (PDF/CSV) and scheduled email
- compliance metrics: MTTR, MTBF

**Key differentiator:** show SLA *and* root cause correlation:
- “SLA breached due to upstream transit outage”

### 77.3 Capacity planning and forecasting
“Prevent downtime” sells well:
- utilization heatmaps by link and PoP
- 95th percentile billing style charts
- days-to-saturation forecast
- recommended upgrade list (top 10 links to upgrade soon)

### 77.4 Degradation and “early warning”
Proactive alerts:
- rising error rate trend on fiber interface
- increasing RTT baseline by region
- repeated micro-outages (flapping)
- rising CPU baseline (routing churn or misconfig)

These reduce churn and customer complaints.

### 77.5 NOC workflow features
Industry-grade operations features:
- incident ownership (assigned teams/users)
- acknowledgement and escalation
- maintenance windows
- runbooks attached to alert rules
- incident postmortem export (timeline + evidence)
- integration with ticketing systems (later: Jira, Zendesk)

---

## 78. PPPoE-focused “next-level” features (MikroTik ISPs)

### 78.1 Value proposition for PPPoE ISPs
PPPoE ISPs compete on:
- uptime
- latency quality
- fast fault resolution
- customer support efficiency

Your system can provide:
- per concentrator session counts and churn
- region heatmaps of customer experience (loss/RTT)
- correlation: “these complaints map to this PoP link congestion”
- “subscriber impact estimate” even without per-subscriber tracking

### 78.2 Subscriber impact without storing personal data
A privacy-friendly approach:
- collect only counts:
  - active sessions per concentrator
  - sessions by profile/plan (optional)
- avoid usernames
- optionally hash identifiers if you must correlate long-term
- store per-subscriber only when explicitly enabled

### 78.3 Customer experience scoring (CX score)
Create a simple, explainable score per site:
- loss weight
- RTT weight
- incident frequency weight
- congestion time weight

Show it as:
- PoP “quality” leaderboard
- trend over time
- “what changed” explanation

This is very attractive to management.

---

## 79. Vendor abstraction (future-proofing without rewriting)

### 79.1 The key pattern: “normalized telemetry contract”
Define canonical metrics/events that all collectors must output.
Example canonical metrics:
- `interface.rx_bps`
- `interface.errors`
- `device.cpu_pct`
- `bgp.session_state`
- `ping.loss_pct`

Vendor-specific collectors map their data to canonical names.

### 79.2 Collector capability flags
Each agent collector advertises:
- supported protocols (routeros_api/snmp/syslog)
- supported metric families (bgp, ospf, interfaces, services)
- safe diagnostics capabilities

Backend can then:
- assign devices to agents based on capability

### 79.3 Avoid “vendor logic” in backend
Keep vendor quirks in:
- agent collectors
- normalization layer
- adapter libraries

Backend remains generic: it stores canonical metrics/events and evaluates rules on them.

This is how you scale to other vendors later.

---

## 80. Enterprise expectations (what large ISPs will ask for)

Prepare for these questions:
1. Do you support **OIDC/SAML**? (Keycloak, Azure AD)
2. Can you run **air-gapped**?
3. Can you do **HA** and **DR**?
4. Do you support **audit exports**?
5. Do you have a **security disclosure** process?
6. Can we restrict which data leaves the network?
7. Can we rotate keys/certs?
8. Can we integrate with our SIEM?

Even if you don’t implement all early, document the roadmap and the design choices that make it possible.

---

## 81. Packaging strategy (cloud-hosted + self-hosted without forks)

### 81.1 One codebase, multiple deployment profiles
Support:
- `compose-small` profile: single VM
- `compose-mid` profile: separate DB host, optional NATS
- `k8s` profile: scalable SaaS/large ISP

### 81.2 Config compatibility
Use:
- the same env vars
- the same migration tool
- the same ingest contract
- feature flags only when necessary

### 81.3 Support policy
Define:
- supported versions (N-2 policy)
- upgrade support window
- how long you support a major version

This matters for enterprise sales.

---

## 82. “Continue” note
Next chapters will cover:
- a concrete “starter alert pack” list (rules, thresholds, default severities)
- a concrete “MikroTik telemetry matrix” (what to collect via API vs SNMP vs syslog)
- security hardening checklist for RouterOS devices and for agents
- a phased roadmap (v0.1, v0.5, v1.0) with milestones and acceptance criteria
- guidance for writing documentation and marketing material that ISPs trust

Say **continue** to proceed.