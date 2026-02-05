# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XXIII
Earlier parts cover integrations, STRIDE templates, blue-team checklists, and demo-kit automation. This part moves from “guide” into “deliverables”: (1) a full telemetry registry v1 appendix, (2) a rule pack v1 importable set (conceptual), and (3) a more complete operator handbook (first draft structure + key sections).

> Note: The rule pack below is expressed as JSON-like objects for clarity. In your implementation you can store it as JSON in DB and provide import/export endpoints.

---

# Part XXIV — Telemetry Registry v1, Rule Pack v1, and Operator Handbook (Draft)

## 160. Telemetry registry v1 (appendix-grade)

### 160.1 Registry format
Each registry entry should specify:
- `name`
- `entity_type`
- `type` (gauge/counter/enum/bool)
- `unit`
- `description`
- `expected_frequency` (fast/medium/slow)
- `sources` (routeros_api/snmp/probe/syslog_derived)
- `labels_allowed` (low-cardinality list)
- `notes` (edge cases, interpretation)

### 160.2 Device metrics
- **`device.up`**
  - entity_type: device
  - type: bool
  - unit: none
  - description: Device reachability from agent POV (ping-based primary)
  - expected_frequency: fast
  - sources: probe
  - notes: Do not set based on API alone.

- **`device.api.up`**
  - entity_type: device
  - type: bool
  - description: RouterOS API reachability (API-SSL recommended)
  - expected_frequency: fast
  - sources: routeros_api
  - notes: Used as evidence; mgmt-plane-only outages exist.

- **`device.cpu_pct`**
  - entity_type: device
  - type: gauge
  - unit: percent
  - expected_frequency: medium
  - sources: routeros_api, snmp
  - labels_allowed: none
  - notes: Interpret with traffic and routing churn context.

- **`device.uptime_seconds`**
  - entity_type: device
  - type: gauge
  - unit: seconds
  - expected_frequency: slow
  - sources: routeros_api, snmp
  - notes: Used to detect reboots and correlate with incidents.

### 160.3 Interface metrics
- **`interface.oper_up`**
  - entity_type: interface
  - type: bool
  - expected_frequency: fast
  - sources: routeros_api, snmp
  - notes: oper state “running” or equivalent.

- **`interface.rx_bps` / `interface.tx_bps`**
  - entity_type: interface
  - type: gauge
  - unit: bps
  - expected_frequency: medium
  - sources: snmp preferred, routeros_api acceptable
  - notes: Prefer counters → rate calculation; beware counter resets.

- **`interface.rx_errors_total` / `interface.tx_errors_total`**
  - entity_type: interface
  - type: counter
  - unit: count
  - expected_frequency: medium
  - sources: snmp, routeros_api
  - notes: Use derivative per second for alerting.

- **`interface.utilization_pct`**
  - entity_type: interface
  - type: gauge
  - unit: percent
  - expected_frequency: medium
  - sources: derived
  - notes: requires capacity_bps in inventory/link metadata.

### 160.4 BGP metrics (peer entities)
- **`bgp.peer.up`**
  - entity_type: bgp_peer
  - type: bool
  - expected_frequency: fast
  - sources: routeros_api
  - notes: up implies Established, but keep `state` for detail.

- **`bgp.peer.state`**
  - entity_type: bgp_peer
  - type: enum
  - expected_frequency: fast
  - sources: routeros_api
  - notes: store raw state string; dashboards map to colored states.

- **`bgp.peer.prefixes_in` / `bgp.peer.prefixes_out`**
  - type: gauge
  - expected_frequency: medium
  - sources: routeros_api
  - notes: baseline required to detect anomalies.

### 160.5 OSPF metrics (neighbor entities)
- **`ospf.neighbor.up`**
  - entity_type: ospf_neighbor
  - type: bool
  - expected_frequency: fast
  - sources: routeros_api
- **`ospf.neighbor.state`**
  - type: enum
  - expected_frequency: fast
  - notes: Full is healthy; other states indicate convergence or problems.

### 160.6 Probe metrics
- **`probe.ping.loss_pct`**
  - entity_type: probe_target
  - type: gauge
  - unit: percent
  - expected_frequency: fast
  - sources: probe
  - notes: Use multiple targets to locate faults.

- **`probe.ping.rtt_ms`**
  - type: gauge
  - unit: ms
  - expected_frequency: fast

### 160.7 Agent self-metrics (platform observability)
These should be Prometheus metrics and optionally sent as platform metrics:
- `agent_last_push_seconds`
- `agent_buffer_depth`
- `agent_jobs_running`
- `agent_poll_errors_total`

---

## 161. Rule Pack v1 (importable starter set)

### 161.1 Philosophy
- conservative thresholds
- minimal false positives
- role-based severity
- easy to tune per tenant

### 161.2 Example rule objects (conceptual)

#### 161.2.1 Device unreachable (core/edge critical)
```json name=docs/rule-packs/v1/device_unreachable_core_edge.json
{
  "id": "00000000-0000-0000-0000-000000000101",
  "name": "Device unreachable (core/edge)",
  "enabled": true,
  "severity": "critical",
  "scope": {
    "by_roles": ["core", "edge"]
  },
  "condition": {
    "type": "threshold",
    "metric_name": "probe.ping.loss_pct",
    "operator": "==",
    "threshold_num": 100,
    "aggregation": "max",
    "lookback_seconds": 30
  },
  "for_seconds": 30,
  "cooldown_seconds": 300,
  "annotations": {
    "summary_template": "Device {{device.name}} unreachable (ping loss 100%)",
    "runbook_url": "https://docs.example.com/runbooks/device-unreachable"
  }
}
```

#### 161.2.2 BGP peer down (upstream)
```json name=docs/rule-packs/v1/bgp_peer_down_upstream.json
{
  "id": "00000000-0000-0000-0000-000000000201",
  "name": "BGP upstream peer down",
  "enabled": true,
  "severity": "critical",
  "scope": {
    "by_entity_type": "bgp_peer",
    "by_tags": ["peer_type:upstream"]
  },
  "condition": {
    "type": "enum_not_equal",
    "metric_name": "bgp.peer.state",
    "bad_values": ["Established"],
    "lookback_seconds": 60
  },
  "for_seconds": 60,
  "cooldown_seconds": 300,
  "annotations": {
    "summary_template": "BGP upstream peer {{bgp_peer.name}} is down (state != Established)",
    "runbook_url": "https://docs.example.com/runbooks/bgp-peer-down"
  }
}
```

#### 161.2.3 OSPF neighbor not Full (core)
```json name=docs/rule-packs/v1/ospf_neighbor_not_full_core.json
{
  "id": "00000000-0000-0000-0000-000000000301",
  "name": "OSPF neighbor degraded (core)",
  "enabled": true,
  "severity": "critical",
  "scope": {
    "by_entity_type": "ospf_neighbor",
    "by_tags": ["role:core"]
  },
  "condition": {
    "type": "enum_not_equal",
    "metric_name": "ospf.neighbor.state",
    "bad_values": ["Full"],
    "lookback_seconds": 60
  },
  "for_seconds": 60,
  "cooldown_seconds": 300,
  "annotations": {
    "summary_template": "OSPF neighbor {{ospf_neighbor.neighbor_router_id}} not Full",
    "runbook_url": "https://docs.example.com/runbooks/ospf-neighbor-down"
  }
}
```

#### 161.2.4 Utilization high (warn/critical tiers)
```json name=docs/rule-packs/v1/interface_utilization_high_warn.json
{
  "id": "00000000-0000-0000-0000-000000000401",
  "name": "Interface utilization high (warn)",
  "enabled": true,
  "severity": "warn",
  "scope": {
    "by_entity_type": "interface",
    "by_tags": ["uplink:true"]
  },
  "condition": {
    "type": "threshold",
    "metric_name": "interface.utilization_pct",
    "operator": ">=",
    "threshold_num": 80,
    "aggregation": "p95",
    "lookback_seconds": 600
  },
  "for_seconds": 600,
  "cooldown_seconds": 1800,
  "annotations": {
    "summary_template": "Uplink {{interface.name}} utilization high (p95 >= 80%)",
    "runbook_url": "https://docs.example.com/runbooks/capacity-high"
  }
}
```

```json name=docs/rule-packs/v1/interface_utilization_high_critical.json
{
  "id": "00000000-0000-0000-0000-000000000402",
  "name": "Interface utilization high (critical)",
  "enabled": true,
  "severity": "critical",
  "scope": {
    "by_entity_type": "interface",
    "by_tags": ["uplink:true"]
  },
  "condition": {
    "type": "threshold",
    "metric_name": "interface.utilization_pct",
    "operator": ">=",
    "threshold_num": 95,
    "aggregation": "p95",
    "lookback_seconds": 300
  },
  "for_seconds": 300,
  "cooldown_seconds": 1800,
  "annotations": {
    "summary_template": "Uplink {{interface.name}} congested (p95 >= 95%)",
    "runbook_url": "https://docs.example.com/runbooks/congestion"
  }
}
```

#### 161.2.5 SLA loss (warn/critical)
```json name=docs/rule-packs/v1/probe_loss_warn.json
{
  "id": "00000000-0000-0000-0000-000000000501",
  "name": "Probe loss degradation (warn)",
  "enabled": true,
  "severity": "warn",
  "scope": {
    "by_entity_type": "probe_target",
    "by_tags": ["class:internet"]
  },
  "condition": {
    "type": "threshold",
    "metric_name": "probe.ping.loss_pct",
    "operator": ">",
    "threshold_num": 2,
    "aggregation": "avg",
    "lookback_seconds": 300
  },
  "for_seconds": 300,
  "cooldown_seconds": 900,
  "annotations": {
    "summary_template": "Internet probe loss > 2% (5m avg)",
    "runbook_url": "https://docs.example.com/runbooks/loss-degradation"
  }
}
```

```json name=docs/rule-packs/v1/probe_loss_critical.json
{
  "id": "00000000-0000-0000-0000-000000000502",
  "name": "Probe loss outage (critical)",
  "enabled": true,
  "severity": "critical",
  "scope": {
    "by_entity_type": "probe_target",
    "by_tags": ["class:internet"]
  },
  "condition": {
    "type": "threshold",
    "metric_name": "probe.ping.loss_pct",
    "operator": ">=",
    "threshold_num": 20,
    "aggregation": "avg",
    "lookback_seconds": 60
  },
  "for_seconds": 60,
  "cooldown_seconds": 900,
  "annotations": {
    "summary_template": "Internet probe loss >= 20% (1m avg)",
    "runbook_url": "https://docs.example.com/runbooks/loss-outage"
  }
}
```

### 161.3 Import/export mechanics (recommended)
- Provide `POST /api/v1/alert-rules/import` accepting:
  - `rules[]`
  - `mode` = `merge|replace`
- Export endpoint:
  - `GET /api/v1/alert-rules/export?pack=v1`
- Validate:
  - rule schema
  - metric_name exists or warn
  - scope selectors valid

---

## 162. Operator handbook (first draft with key content)

### 162.1 Installation (self-hosted)
- prerequisites:
  - Linux VM with Docker + Compose
  - DNS record and TLS plan
  - persistent storage sized
- deploy steps:
  - set secrets (db password, master key)
  - run compose up
  - verify health endpoints
- post-deploy:
  - create admin user
  - create first tenant (if multi-tenant)

### 162.2 Agent enrollment and placement
- enroll agent using one-time token
- place one agent per PoP
- firewall rules required:
  - agent → backend: 443
  - agent → routers: 8729 (API-SSL), 161/162 (SNMP if used), syslog port if used
- verify:
  - agent heartbeat in UI
  - sample probe results

### 162.3 Adding devices
- create credential set
- add device mgmt IP and site
- assign to agent (or auto)
- apply tags:
  - role tags
  - uplink interface tags
- verify:
  - API up
  - interface inventory discovered
  - BGP peers discovered (if enabled)

### 162.4 Topology and dependencies
- start manual:
  - create site links
  - create dependency groups for redundancy
- enable inference proposals:
  - review and accept OSPF/BGP neighbor edges
- validate RCA:
  - simulate a link cut in lab and verify suppression

### 162.5 Alert packs
- enable pack v1
- tune thresholds:
  - utilization
  - SLA loss/RTT
- configure notifications:
  - telegram channel for critical
  - email for warnings (optional)
- test notifications with a controlled incident

### 162.6 Incident response workflow
- NOC home: triage by severity + impact
- open incident: verify root cause evidence
- run diagnostics (ping/traceroute) from correct PoP agent
- acknowledge and assign
- add notes with actions taken
- close incident when recovered and stable

### 162.7 Maintenance & change workflows
- create maintenance window before planned work
- create change record for major changes
- after completion:
  - review incidents during window
  - generate postmortem if outage occurred

### 162.8 Backups and DR
- nightly DB backups
- test restore
- keep master key safe
- document RTO/RPO targets

### 162.9 Troubleshooting quick reference
- agent offline:
  - check network to backend
  - check cert expiry
  - check buffer depth
- device missing metrics:
  - credential set
  - firewall allowed ports
  - API enabled
- alerts not firing:
  - rule enabled?
  - metric present?
  - evaluation worker running?

---

## 163. “Continue” note
Next chapters (if you want a complete deliverable set) can include:
- a full telemetry registry v1 as a separate file with all entries
- a complete rule pack v1 (20–50 rules) ready to import
- complete operator handbook prose (step-by-step)
- UI wireframe spec and component inventory
- “golden lab” scripts for EVE-NG to simulate failures reliably

Say **continue** to proceed.