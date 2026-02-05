# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–VIII
Earlier parts define the overall architecture, schemas, alerting/RCA, security, multi-tenancy, and self-hosted operations. This section appends the next “network + lab + MikroTik practice” chapters.

---

# Part IX — MikroTik/ISP Lab Topologies, OSPF & BGP Patterns, and How Monitoring Uses Them

This part gives you a concrete, MikroTik-friendly ISP simulation topology you can implement in **EVE-NG** and directly use to validate:
- your topology model (manual + inferred edges)
- alert rules (BGP down, link down, SLA loss)
- RCA and suppression
- capacity and degradation detection

> Note: This guide is not a RouterOS CLI cookbook, but it will explain what you should build and why. You can later add RouterOS script snippets in a separate “Lab Cookbook” doc.

---

## 67. A reference ISP topology for simulation (EVE‑NG + CHR)

### 67.1 Goals of the reference topology
- Minimal number of routers but exhibits real ISP behaviors:
  - core redundancy
  - IGP inside (OSPF)
  - edge policy (eBGP)
  - internal route distribution (iBGP + route reflectors)
  - access/aggregation sites
  - customer edges / sample endpoints
- Can simulate:
  - single link cut
  - router failure
  - upstream provider failure
  - congestion and packet loss
  - route flaps

### 67.2 Topology roles (recommended nodes)
You can start with 8–12 routers:

**Core (2–4 routers):**
- `CORE1`, `CORE2` (OSPF Area 0)
- Optional: `CORE3`, `CORE4` (for ring/mesh)

**Route Reflectors (1–2 routers):**
- `RR1` (can be CORE1 if small, but dedicated is more realistic)
- Optional `RR2` (HA)

**Edge (2 routers):**
- `EDGE1`, `EDGE2`
- connect to upstream providers using eBGP

**Aggregation/Access (2–4 routers):**
- `AGG1`, `AGG2` (connect sites/PoPs)
- `ACCESS1`, `ACCESS2` (simulate access concentrators / PPPoE termination)

**Upstream provider simulation:**
- 1 FRRouting (FRR) Linux node: `UPSTREAM1`
- optional second upstream: `UPSTREAM2`

**Endpoints (Linux):**
- `TOOLS` (runs ISPVisualMonitor-Agent in the lab)
- `TARGETS` (hosts services; responds to ping/http)
- Optional “customer host” nodes behind ACCESS routers

### 67.3 Addressing plan (must be consistent)
Use a clean addressing plan:
- Loopbacks: `10.255.0.0/16` (router IDs)
  - CORE1 lo: `10.255.0.1/32`
  - CORE2 lo: `10.255.0.2/32`
  - EDGE1 lo: `10.255.0.11/32`, etc.
- Point-to-point links:
  - Use /31 or /30 from `10.0.0.0/16`
- PoP LANs:
  - `10.10.X.0/24` per site

**Monitoring benefit:** loopbacks are stable; you can probe them and identify routers reliably.

---

## 68. OSPF design for ISP core (what and why)

### 68.1 OSPF basics for monitoring systems
OSPF is a link-state IGP; routers form adjacencies (neighbors) and exchange LSAs. When a link fails:
- adjacency drops
- routing recalculates
- traffic reroutes (if redundancy exists)

**What monitoring needs from OSPF**
- neighbor state changes (Full → Down)
- interface status and errors on OSPF links
- SPF churn symptoms (optional)

### 68.2 ISP patterns
**Common ISP OSPF pattern**
- Core links in **Area 0**
- ABRs if you have multiple areas (often not needed in small simulation)
- Use loopback as router-id

**Why use OSPF in your simulation**
- It creates real neighbor adjacency events
- It creates stable “IGP up but BGP down” or “IGP down” scenarios
- It helps you test RCA (link down → OSPF down → site impact)

### 68.3 Monitoring use cases
Alert rules:
- `ospf.neighbor_state != Full for 60s` (warn/critical depending on role)
Correlations:
- if interface down and OSPF neighbor down, confidence increases
RCA:
- OSPF neighbor down in core explains many downstream unreachable probes

---

## 69. BGP design for ISP edges (iBGP/eBGP)

### 69.1 Why BGP is critical for your product’s value
ISPs live and die by BGP:
- upstream reachability
- traffic engineering
- route leaks and prefix drops
- customer outages due to peering issues

A monitoring tool that does not properly handle BGP is not “ISP-grade”.

### 69.2 eBGP (edge to upstream)
In the lab:
- EDGE1 ↔ UPSTREAM1 via eBGP
- EDGE2 ↔ UPSTREAM2 via eBGP (optional)

**Monitoring signals**
- BGP state (Established vs Idle/Connect/Active)
- session uptime
- last error
- prefix counts (in/out)

**Alerting**
- session down for 60s
- prefix count sudden drop > X% for 2m (advanced)

### 69.3 iBGP inside the ISP
In ISPs, iBGP distributes external routes.
But full mesh does not scale; use **route reflectors**.

**Route reflector (RR) pattern**
- all core/edge routers peer with RR
- RR reflects routes between clients

**Monitoring signals**
- RR session health is high impact (many routers depend on RR)
- RR down leads to widespread routing changes

**RCA**
If many iBGP sessions drop simultaneously, root cause may be:
- RR failure
- core reachability problem (IGP issue)

### 69.4 What to monitor in RouterOS for BGP (API-first)
At minimum:
- peer state (string/enum)
- established time / uptime
- prefixes received/advertised (if available)
- peer remote address and AS
- whether it’s eBGP/iBGP (infer by local AS vs remote AS)

---

## 70. How monitoring interprets routing vs reachability

### 70.1 The “control plane vs data plane” distinction
- Control plane: OSPF/BGP sessions, routing tables
- Data plane: actual packet forwarding (ping, TCP)

**Common outages**
- BGP down but ping to router still works (transit outage)
- OSPF stable but packet loss high (congestion)
- API down but ping ok (management issue)

Your system must:
- represent these distinctly
- not collapse everything into a single “device down”

### 70.2 Health model (recommended)
Define separate health aspects:
- `reachability` (ping/API)
- `routing_health` (OSPF/BGP)
- `link_health` (interface + errors + utilization)
- `service_health` (SLA probes, PPPoE counts)

Then compute:
- `overall_status` using weighted logic + role tags

Example:
- For EDGE routers:
  - BGP transit down may be critical even if router reachable
- For ACCESS concentrators:
  - PPPoE session drop spike may be critical even if BGP fine

---

## 71. Tagging and role classification (MikroTik-first practicality)

### 71.1 Why tags matter
Tags drive:
- which rules apply
- which probes run
- RCA dependency defaults
- UI filtering (map layers)

### 71.2 Recommended minimum tags
- `role:core`, `role:rr`, `role:edge`, `role:agg`, `role:access`
- `uplink:true` per interface (or interface tag)
- `site:<name>` (or site_id already)
- `provider:<upstream-name>` for edge links

### 71.3 Default dependency rules by role (starter)
If operator hasn’t defined dependencies yet, apply defaults:
- access depends_on aggregation
- aggregation depends_on core
- edge depends_on core + upstream (separate service node for upstream)
- rr depends_on core

These defaults should have low confidence and be editable.

---

## 72. Lab failure scenarios (what to test and what you should see)

### 72.1 Scenario A: Core link cut
Action: bring down CORE1–CORE2 link.
Expected:
- interface down event
- OSPF neighbor down
- reroute if redundant path exists
- if no redundancy, downstream probes fail
Your platform should:
- raise one incident “CoreLink CORE1–CORE2 down”
- suppress downstream device alerts as impacted

### 72.2 Scenario B: Upstream failure
Action: shut down UPSTREAM1 BGP session or node.
Expected:
- EDGE1 eBGP down
- internet probe loss (to 1.1.1.1) increases
- internal reachability still ok
Your platform should:
- show “Upstream provider down” incident
- classify as service outage (transit)
- show impact (internet service degraded) not “all routers down”

### 72.3 Scenario C: RR failure
Action: shut down RR1.
Expected:
- many iBGP sessions drop
- routing instability
Your platform should:
- detect high-impact root cause at RR1
- show blast radius across edges/cores

### 72.4 Scenario D: Congestion
Action: generate traffic saturating a link (iperf).
Expected:
- utilization > 90%
- ping RTT and loss degrade across affected path
Your platform should:
- create degradation incident
- predict saturation and warn earlier if trend exists

---

## 73. Agent hardening: RouterOS API rate limits, pooling, circuit breakers

### 73.1 Why this matters
Routers are production-critical. Monitoring must not create outages.
Aggressive polling can:
- consume CPU
- exhaust API session limits
- increase control plane jitter

### 73.2 Session pooling rules
- reuse sessions per router where safe
- keep a max session count per router (often 1)
- reconnect on errors with exponential backoff
- never loop retry fast in a tight loop

### 73.3 Circuit breaker
If a router is failing repeatedly:
- mark router collector as “open circuit” for N seconds
- reduce poll frequency temporarily
- emit an event “monitoring degraded for device X (API errors)”
This prevents cascading overload.

### 73.4 Safe defaults for production
- 1 API session per router
- 5s connect timeout
- 10s query timeout
- max 1–3 concurrent jobs per router
- jitter per schedule bucket

---

## 74. “Continue” note
Next chapters will cover:
- commercialization-grade features and packaging (what to sell, what to keep OSS)
- subscription tiers and enterprise expectations (SAML/OIDC, audit exports, HA)
- agent open-source governance (security reviews, release signing, reproducible builds)
- a “product-grade” onboarding workflow for ISPs (wizard, discovery, templates)
- how to monitor PPPoE without leaking PII (privacy-by-design patterns)
- a future vendor abstraction layer (Cisco/Juniper/etc.) without rewriting the system

Say **continue** to proceed.