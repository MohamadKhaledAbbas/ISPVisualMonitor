# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–X
Earlier parts establish the platform architecture, storage, alerting/RCA, security, scaling, lab topologies, and productization strategy. This part provides *concrete starter packs*, a MikroTik telemetry matrix, and phased milestones.

---

# Part XI — Starter Alert Packs, MikroTik Telemetry Matrix, Hardening Checklists, and Roadmap

## 83. Starter alert packs (opinionated defaults that work)

A major differentiator is shipping a set of **prebuilt rules** that ISPs can enable quickly and then tune. These should be conservative to avoid noise.

### 83.1 General design rules for defaults
- defaults must be safe in noisy networks
- use `for` durations to avoid transient spikes
- include hysteresis on recoveries
- include “role-based severities”
- every rule has:
  - summary template
  - runbook link (even if generic)
  - evidence fields (what metric triggered it)

### 83.2 Pack A: Core/IGP health (OSPF)
**Target:** core + aggregation routers

1. **OSPF neighbor down (critical for core)**
   - scope: `role:core`, `role:agg`
   - condition: `ospf.neighbor_state != Full`
   - for: 60s
   - severity:
     - core: critical
     - agg: warn/critical depending on adjacency count

2. **OSPF adjacency flapping**
   - condition: > 3 transitions in 10m
   - severity: warn
   - action: suppress raw neighbor alerts, create “flapping” incident

### 83.3 Pack B: Edge and upstream health (BGP)
**Target:** edge routers and route reflectors

1. **eBGP down**
   - scope: `role:edge`
   - condition: `bgp.session_state != Established`
   - for: 60s
   - severity: critical
   - note: include remote AS and provider tag

2. **iBGP down to RR**
   - scope: `role:core`, `role:edge`
   - condition: RR session down
   - for: 60s
   - severity: critical if RR is single; warn if redundant RR exists

3. **Prefix count sudden drop (advanced, optional)**
   - condition: prefixes_in decreased by > 30% vs 10m baseline
   - for: 2m
   - severity: critical
   - warning: requires stable baselines; default off

### 83.4 Pack C: Device reachability and management plane
1. **Device unreachable (ping-based)**
   - condition: loss == 100% over last 3 probes
   - for: 30s
   - severity by role:
     - core/edge: critical
     - access: critical
     - customer-edge: warn/critical depending on policy

2. **Management plane degraded**
   - condition: ping OK but API unreachable for 5m
   - severity: warn
   - value: identifies “mgmt outage” vs real device down

### 83.5 Pack D: Link utilization and congestion
1. **High utilization warn**
   - condition: utilization > 80% avg over 10m
   - severity: warn

2. **High utilization critical**
   - condition: utilization > 95% avg over 5m
   - severity: critical

3. **Congestion symptom rule**
   - condition: utilization high AND ping RTT p95 increased by > X
   - severity: critical
   - value: ties congestion to experience impact

### 83.6 Pack E: Physical degradation (errors/drops)
1. **Interface error rate high**
   - condition: errors_per_sec > threshold for 10m
   - severity: warn
2. **Drops increasing**
   - condition: drops_per_sec > threshold for 10m
   - severity: warn/critical if correlated with SLA loss

### 83.7 Pack F: SLA (experience) rules
1. **Loss degradation**
   - condition: loss_pct > 2 for 5m
   - severity: warn
2. **Loss outage**
   - condition: loss_pct > 20 for 1m
   - severity: critical
3. **RTT degradation**
   - condition: rtt_p95 > threshold (regional) for 10m
   - severity: warn

**Important:** Provide per-site or per-region thresholds (rural wireless has different baselines than metro fiber).

### 83.8 Pack G: PPPoE (when enabled)
1. **PPPoE sessions drop spike**
   - condition: sessions_count drops by > X% within 5m
   - severity: critical
2. **PPPoE churn high**
   - condition: disconnects per minute > threshold
   - severity: warn/critical

---

## 84. MikroTik telemetry matrix (API vs SNMP vs Syslog vs Probes)

This matrix helps you decide what to implement first and why.

### 84.1 Categories
- **Device identity & health**
- **Interfaces & link stats**
- **Routing**
- **Services (PPPoE/DHCP)**
- **Events**
- **Experience (SLA)**

### 84.2 Recommended collection sources (high-level)
1. **Device identity & health**
   - RouterOS API: yes (rich, structured)
   - SNMP: yes (standard, lightweight)
   - Syslog: no
   - Probes: no

2. **Interfaces & link stats**
   - SNMP: preferred for counters/errors at scale
   - RouterOS API: acceptable, especially early
   - Syslog: events only (link up/down)
   - Probes: experience signals only

3. **Routing (BGP/OSPF)**
   - RouterOS API: preferred
   - Syslog: useful for quick “neighbor down” evidence
   - SNMP: limited for routing detail (varies)
   - Probes: detect reachability impact

4. **PPPoE/DHCP service stats**
   - RouterOS API: preferred
   - SNMP: limited/varies
   - Syslog: session events possible but noisy
   - Probes: customer experience, not session truth

5. **Events**
   - Syslog: primary for event timelines
   - RouterOS API: can fetch some state but not event stream
   - SNMP traps: possible later (complex)
   - Probes: produce synthetic events (probe failure)

6. **Experience (SLA)**
   - Probes: primary
   - Everything else: supporting evidence

### 84.3 Implementation recommendation (phased)
- Phase 1: RouterOS API + Probes
- Phase 2: Add SNMP for interface counters/health scaling
- Phase 3: Add Syslog for richer incident timelines
- Phase 4: Flow telemetry for advanced paid features

---

## 85. Hardening checklists (RouterOS, agents, backend)

### 85.1 RouterOS hardening (monitoring perspective)
- API:
  - prefer API-SSL (8729)
  - disable plain API (8728) if possible
  - restrict by firewall to agent IPs only
- Users:
  - create dedicated monitoring user
  - least privilege, no write perms
  - rotate credentials
- Management network:
  - mgmt VLAN/VRF
  - no mgmt exposure to public internet
- Logging:
  - configure syslog to agent collector if used
- NTP:
  - ensure accurate clocks (event correlation needs time)

### 85.2 Agent host hardening
- run as non-root where possible
- read-only filesystem where possible
- minimal Linux capabilities
- secure storage for buffered data (disk permissions)
- outbound-only connectivity to backend (prefer)
- log redaction (never log secrets)

### 85.3 Backend hardening
- TLS everywhere
- strict request size limits
- rate limits per agent and per user
- encryption master key management
- audit logs enabled and protected
- DB credentials rotated
- vulnerability scanning for container images

---

## 86. Roadmap milestones (v0.1 → v1.0 with acceptance criteria)

A credible roadmap helps you build systematically and communicate value.

### 86.1 v0.1 — “Lab MVP”
**Goal:** works in EVE-NG reliably

Acceptance criteria:
- agent collects RouterOS API device health + BGP state
- agent performs ping probes
- backend ingests batches and stores metrics
- UI shows:
  - site list + basic map points
  - device statuses
- one alert: device unreachable
- basic incident list (even if manual grouping)

### 86.2 v0.3 — “Pilot-ready MVP”
Acceptance criteria:
- multi-site map + link visualization (manual links)
- alert packs A–F basic:
  - device down, BGP down, utilization, loss/RTT
- dedup + for-duration + cooldown
- incidents auto-created by correlation window
- maintenance windows implemented
- self-hosted compose deployment documented
- backups documented (even if manual)

### 86.3 v0.5 — “Production pilot”
Acceptance criteria:
- RCA v1 (dependency suppression + root cause candidate)
- evidence panel and timeline view
- Timescale rollups + retention policies
- agent mTLS enrollment flow
- notification channels: telegram + email + webhooks
- audit logs for admin actions
- basic SLA reports per site/device

### 86.4 v0.8 — “Production-ready”
Acceptance criteria:
- multi-agent support with site affinity + failover
- SNMP integration for interface counters at scale
- syslog ingestion for event timelines
- performance testing for 300 routers
- upgrade/migration runbooks
- DR basics (restore tested)

### 86.5 v1.0 — “Commercial-grade”
Acceptance criteria:
- packaging tiers (self-hosted + SaaS)
- enterprise auth (OIDC; SAML later)
- advanced reporting pack (monthly SLA, MTTR)
- capacity forecasting
- PPPoE pack (counts + churn + impact estimation) if target market needs it
- signed agent releases + SBOM
- documented support policy and security disclosure

---

## 87. Documentation set you should ship (ISPs judge docs)
Minimum docs for production credibility:
- Installation (compose)
- Agent enrollment and placement
- RouterOS hardening guide for monitoring
- Data collected (privacy and trust)
- Alert packs and tuning guide
- RCA explanation and limitations
- Backup/restore and DR
- Upgrade guide
- Troubleshooting guide (common issues)
- Security policy (SECURITY.md)

---

## 88. “Continue” note
Next chapters will cover:
- a “data collected and privacy” policy template (ISP-friendly)
- a ���support contract expectations” appendix (what enterprises ask)
- deeper PPPoE modeling guidance (service entities, impact, churn)
- how to present topology on maps (UI/UX) and graph layout strategies
- performance test plan: load generation for agents and backend
- how to build a “demo environment” for sales (EVE-NG lab + scripted failure demos)

Say **continue** to proceed.