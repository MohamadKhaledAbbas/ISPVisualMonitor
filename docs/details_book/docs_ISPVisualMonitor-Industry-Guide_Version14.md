# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XIII
Earlier parts cover architecture, storage, rules/RCA, security, scaling, ops, lab, productization, starter packs, privacy, performance, redundancy, PPPoE patterns, and packaging guidance. This part defines the *canonical telemetry registry*, schema versioning, ingest payload shapes, tagging conventions, and an operator runbook library.

---

# Part XIV — Canonical Telemetry Registry, Schemas, Tagging Style Guide, and Runbooks

## 102. Canonical metric & event registry (the “contract” of your platform)

### 102.1 Why you need a registry
If you don’t standardize metric names and meanings:
- rules become vendor-specific
- dashboards don’t transfer across tenants
- migrations become painful
- other collector plugins won’t be compatible

A registry is:
- a list of metric names and event types
- each with definition, units, expected labels, and entity type
- versioned and documented

### 102.2 Registry design rules
- metric names are stable, lowercase, dot-separated
- units are explicit and consistent
- entity type is explicit (device/interface/site/service)
- labels are low-cardinality
- avoid embedding IDs in labels (store IDs in entity fields)

### 102.3 Recommended metric name list (starter)
Below is a recommended initial registry. You can extend it later.

#### 102.3.1 Device metrics (entity_type = device)
- `device.up` (0/1)
- `device.api.up` (0/1) — RouterOS API reachable
- `device.snmp.up` (0/1) — SNMP reachable (if used)
- `device.uptime_seconds` (gauge)
- `device.cpu_pct` (0–100)
- `device.mem_used_bytes` (gauge)
- `device.mem_total_bytes` (gauge)
- `device.temp_c` (gauge) — if sensors exist
- `device.clock.skew_seconds` (gauge) — optional (needs time sync logic)

#### 102.3.2 Interface metrics (entity_type = interface)
- `interface.oper_up` (0/1)
- `interface.admin_up` (0/1)
- `interface.rx_bps` (gauge)
- `interface.tx_bps` (gauge)
- `interface.rx_bytes_total` (counter)
- `interface.tx_bytes_total` (counter)
- `interface.rx_errors_total` (counter)
- `interface.tx_errors_total` (counter)
- `interface.rx_discards_total` (counter)
- `interface.tx_discards_total` (counter)
- `interface.utilization_pct` (0–100) — derived using capacity_bps

#### 102.3.3 Routing metrics (entity_type = bgp_peer / ospf_neighbor OR device with labels)
You have two design options:
- model peers/neighbors as separate entities (cleaner)
- or store as device metrics with labels (simpler but can be high-cardinality)

**Recommendation:** model as separate entities in inventory:
- `bgp_peers` table, `ospf_neighbors` table

Then metrics:
- `bgp.peer.up` (0/1)
- `bgp.peer.state` (text or enum mapped to numeric)
- `bgp.peer.uptime_seconds` (gauge)
- `bgp.peer.prefixes_in` (gauge)
- `bgp.peer.prefixes_out` (gauge)

- `ospf.neighbor.up` (0/1)
- `ospf.neighbor.state` (text/enum)
- `ospf.neighbor.uptime_seconds` (gauge)

#### 102.3.4 Probe metrics (entity_type = probe_target or site/service)
- `probe.ping.loss_pct` (0–100)
- `probe.ping.rtt_ms` (gauge)
- `probe.ping.jitter_ms` (gauge) — optional
- `probe.tcp.up` (0/1)
- `probe.http.up` (0/1)
- `probe.http.latency_ms` (gauge)

#### 102.3.5 Service metrics (PPPoE, entity_type = service or device)
- `pppoe.sessions.active` (gauge)
- `pppoe.sessions.churn_per_min` (gauge)
- `dhcp.leases.active` (gauge)

### 102.4 Canonical event types (starter)
Event types should be stable strings:
- `device.reachable.changed`
- `device.api.changed`
- `interface.oper.changed`
- `bgp.peer.state.changed`
- `ospf.neighbor.state.changed`
- `probe.ping.degraded` (optional)
- `agent.heartbeat.missed`
- `agent.buffer.high`
- `config.changed.detected` (daily hash change)

Event payloads should include:
- old/new values
- optional correlation IDs (incident_id)
- raw syslog message when applicable

---

## 103. Schema versioning strategy (rules + ingest + registry)

### 103.1 Why schema versioning matters
Agents and backend will evolve separately. A stable versioning plan allows:
- old agents to keep working
- safe server upgrades
- gradual feature rollout

### 103.2 Recommended versioning approach
- Ingest API path includes version: `/ingest/v1/batch`
- Payload includes:
  - `schema_version` (e.g., `1.0`)
  - `agent_version` (semantic)
  - `capabilities` (collector types)
- Rule definitions include:
  - `rule_schema_version`

### 103.3 Backward compatibility rules
- server must accept at least N-2 agent versions (policy)
- changes to payload fields should be additive
- breaking changes require:
  - new endpoint `/v2`
  - migration period

### 103.4 Registry versioning
- publish registry as a versioned document:
  - `docs/telemetry-registry/v1.md`
- when adding new metrics/events:
  - keep old ones valid
  - mark deprecated metrics with removal date

---

## 104. Ingest payload shape (concrete JSON outline)

This is a “shape contract” you can implement as JSON schema.

### 104.1 Batch request top-level
Fields:
- `schema_version`
- `sent_at`
- `agent`:
  - `id`
  - `name` (optional)
  - `version`
  - `capabilities[]` (routeros_api, snmp, syslog, probes)
- `metrics[]`
- `events[]`
- optional `inventory_patch`:
  - new interfaces
  - new peers
  - device metadata updates

### 104.2 Metric sample shape
- `ts`
- `metric_name`
- `entity_type`
- `entity_id`
- `value_num` or `value_text`
- optional:
  - `unit`
  - `labels` (json object; low cardinality)
  - `quality` (good/estimated/stale)
  - `source` (routeros_api/snmp/probe/syslog-derived)

### 104.3 Event shape
- `id` (uuid)
- `ts`
- `event_type`
- `entity_type`
- `entity_id`
- `severity`
- `payload` (json)
- optional:
  - `raw` (syslog raw line, if applicable)

---

## 105. Tagging & naming style guide (consistency creates value)

### 105.1 Why you need a style guide
ISPs will use your tags and naming to:
- filter dashboards
- apply rule scopes
- compute dependencies

Inconsistent tagging makes the system hard to operate.

### 105.2 Device naming conventions
Recommend:
- `ROLE-SITE-NUMBER`
Examples:
- `CORE-BAGHDAD-1`
- `EDGE-BAGHDAD-1`
- `AGG-MOSUL-1`
- `ACCESS-POPA-1`

### 105.3 Site naming conventions
- short, stable, human-friendly
- include city/region if possible

### 105.4 Role tags (required)
- `role:core`
- `role:rr`
- `role:edge`
- `role:agg`
- `role:access`
- `role:customer-edge`

### 105.5 Interface tags (recommended)
- `uplink:true`
- `medium:fiber|wireless|leased`
- `provider:<name>` (for edge uplinks)
- `critical:true` (if you want extra alert sensitivity)

### 105.6 Dependency conventions
- model dependencies at service level when possible:
  - `service:internet-pop-a depends_on EDGE1`
- keep manual override capability for operators

---

## 106. Operator runbook library (what ISPs actually need)

Runbooks make alerts actionable and reduce MTTR.

### 106.1 Runbook structure
For each runbook:
- Summary: what it means
- Impact: what users experience
- Common causes
- Immediate checks (2–5 steps)
- Deep checks (optional)
- Mitigation actions
- Escalation criteria
- Related metrics and dashboards

### 106.2 Runbook: BGP peer down (edge)
**Meaning**
- Loss of transit/peering connectivity with an upstream or peer.

**Impact**
- internet outage or degraded reachability for some/all prefixes.

**Immediate checks**
1. Check interface oper status and errors on uplink.
2. Check OSPF/IGP reachability to loopbacks (if peer uses loopbacks).
3. Check ping to peer IP and next-hop.
4. Check prefix count before/after (if available).
5. Check syslog for link-down or BGP reset reason.

**Mitigation**
- fail over to secondary upstream if configured
- contact upstream provider with timestamps and evidence

**Escalate when**
- multiple edges affected or prefixes drop network-wide

### 106.3 Runbook: High packet loss to core probe target
**Meaning**
- forwarding plane degradation, congestion, or physical layer errors.

**Immediate checks**
1. Identify which PoP agent sees the loss (one PoP vs all).
2. Check interface utilization on the path.
3. Check error/drops counters trending.
4. Run traceroute from affected PoP agent.
5. Compare RTT baseline vs spike.

**Mitigation**
- traffic engineering change (BGP policy) if available
- rate-limit offending traffic (if DDoS)
- schedule link upgrade if sustained congestion

### 106.4 Runbook: Interface error rate rising
**Meaning**
- physical degradation: dirty fiber, bad SFP, wireless interference.

**Immediate checks**
1. Confirm which interface and medium.
2. Compare errors to historical baseline.
3. Check link partner interface (if monitored).
4. If fiber: clean connectors, replace patch lead, swap SFP.
5. If wireless: check signal/noise, alignment, interference.

### 106.5 Runbook: PPPoE sessions drop spike
**Meaning**
- access concentrator issue, upstream outage to concentrator, AAA/RADIUS issues.

**Immediate checks**
1. Confirm concentrator reachability and CPU/memory.
2. Check uplink interface status/utilization.
3. Check recent config changes.
4. Check RADIUS health (if used).
5. Check churn rate and geographic correlation (which PoPs).

---

## 107. “Continue” note
Next chapters can include:
- adding “peer entities” and “probe targets” to inventory schema concretely
- rule JSON schema examples and validation approach
- sample dashboard definitions (map overlays, capacity charts, SLA views)
- a proposed internal package/module layout for both repos aligned to these contracts
- how to build migration tooling and fixture data for labs

Say **continue** to proceed.