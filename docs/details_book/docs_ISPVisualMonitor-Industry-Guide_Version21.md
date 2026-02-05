# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XXI
Earlier parts cover ticketing integrations and a glossary. This part adds: a concrete webhook specification with example payloads, an integration catalog (Grafana/Prometheus/Loki), an “operator handbook” outline for self-hosted ISPs, a threat-model appendix (STRIDE), and a v1.0 acceptance checklist.

---

# Part XXII — Webhook Spec, Integration Catalog, Operator Handbook Outline, Threat Model (STRIDE), and v1.0 Acceptance Checklist

## 148. Ticketing/automation webhook specification (v1)

### 148.1 Webhook overview
The webhook is an outbound integration that fires on incident lifecycle events:
- incident opened
- incident updated (severity/root cause/impact changed)
- incident acknowledged
- incident closed

### 148.2 Delivery guarantees
- at-least-once delivery (retries)
- idempotent consumer required
- exponential backoff
- DLQ (failed deliveries stored for manual resend)

### 148.3 Authentication and integrity
Support at least one:
- **HMAC signature** (recommended)
  - header: `X-ISPVM-Signature: sha256=<hex>`
  - computed over raw request body using shared secret
- Optional:
  - mTLS to webhook endpoint (enterprise)
  - JWT bearer token

### 148.4 Headers
- `Content-Type: application/json`
- `X-ISPVM-Event-Type: incident.opened|incident.updated|incident.acked|incident.closed`
- `X-ISPVM-Delivery-ID: <uuid>` (unique per attempt)
- `X-ISPVM-Tenant-ID: <uuid>` (if multi-tenant)
- `X-ISPVM-Signature: sha256=<hex>` (if enabled)

### 148.5 Payload schema (conceptual)
Top-level fields:
- `event_id` (uuid)
- `event_type`
- `sent_at`
- `tenant`:
  - `id`, `name` (optional)
- `incident`:
  - `id`
  - `title`
  - `status`
  - `severity`
  - `starts_at`
  - `ends_at` (nullable)
  - `url` (deep link)
  - `ack`:
    - `acked_by` (nullable)
    - `acked_at` (nullable)
  - `root_causes[]`:
    - entity_type, entity_id, name (optional), confidence
  - `impact`:
    - sites_affected
    - devices_affected
    - services_affected
    - estimated_subscribers_affected (nullable)
  - `evidence_top[]` (short list)
  - `labels` (optional map)

### 148.6 Example payloads
> These are examples, not a final API guarantee. Keep them stable once published.

#### 148.6.1 `incident.opened`
```json name=docs/examples/webhooks/incident.opened.json
{
  "event_id": "2b5a7f50-5b8c-4d13-8c65-93f5ed2d8a23",
  "event_type": "incident.opened",
  "sent_at": "2026-02-05T12:00:10Z",
  "tenant": {
    "id": "b2f5c2b8-7aa4-4f03-9f8c-9d3a4b5e6c01",
    "name": "ExampleISP"
  },
  "incident": {
    "id": "a3d2b1c0-1111-4444-8888-0123456789ab",
    "title": "Transit outage: UPSTREAM1 degraded",
    "status": "open",
    "severity": "critical",
    "starts_at": "2026-02-05T11:59:30Z",
    "ends_at": null,
    "url": "https://monitor.example.com/incidents/a3d2b1c0-1111-4444-8888-0123456789ab",
    "ack": {
      "acked_by": null,
      "acked_at": null
    },
    "root_causes": [
      {
        "entity_type": "bgp_peer",
        "entity_id": "6e1b2a3c-2222-4444-9999-abcdefabcdef",
        "name": "EDGE1-UPSTREAM1",
        "confidence": 0.86
      }
    ],
    "impact": {
      "sites_affected": 12,
      "devices_affected": 43,
      "services_affected": 1,
      "estimated_subscribers_affected": null
    },
    "evidence_top": [
      "BGP peer EDGE1-UPSTREAM1 state=Idle for 90s",
      "Probe loss to 1.1.1.1 = 100% from POP-A agent",
      "Core OSPF neighbors stable"
    ],
    "labels": {
      "category": "transit",
      "provider": "UPSTREAM1"
    }
  }
}
```

#### 148.6.2 `incident.closed`
```json name=docs/examples/webhooks/incident.closed.json
{
  "event_id": "ae4c0cc0-3c6f-4a1b-b8c9-3cc7a2d0e4d2",
  "event_type": "incident.closed",
  "sent_at": "2026-02-05T12:30:10Z",
  "tenant": {
    "id": "b2f5c2b8-7aa4-4f03-9f8c-9d3a4b5e6c01",
    "name": "ExampleISP"
  },
  "incident": {
    "id": "a3d2b1c0-1111-4444-8888-0123456789ab",
    "title": "Transit outage: UPSTREAM1 degraded",
    "status": "closed",
    "severity": "critical",
    "starts_at": "2026-02-05T11:59:30Z",
    "ends_at": "2026-02-05T12:28:40Z",
    "url": "https://monitor.example.com/incidents/a3d2b1c0-1111-4444-8888-0123456789ab",
    "ack": {
      "acked_by": "noc.user@example.com",
      "acked_at": "2026-02-05T12:01:00Z"
    },
    "root_causes": [
      {
        "entity_type": "bgp_peer",
        "entity_id": "6e1b2a3c-2222-4444-9999-abcdefabcdef",
        "name": "EDGE1-UPSTREAM1",
        "confidence": 0.86
      }
    ],
    "impact": {
      "sites_affected": 12,
      "devices_affected": 43,
      "services_affected": 1,
      "estimated_subscribers_affected": null
    },
    "evidence_top": [
      "BGP peer recovered; state=Established",
      "Probe loss returned to <1%"
    ],
    "labels": {
      "category": "transit",
      "provider": "UPSTREAM1"
    }
  }
}
```

---

## 149. Integration catalog (observability + NOC tooling)

### 149.1 Prometheus (recommended)
Use Prometheus to monitor:
- backend API latency and errors
- ingest throughput
- rule eval durations
- DB connection pool usage
- agent heartbeats
- buffer depth

Expose metrics:
- `/metrics` on backend and agent

### 149.2 Grafana
Grafana can be used to:
- visualize platform self-metrics
- build custom dashboards for ISPs
- optionally query Timescale directly (read-only user)

**Caution:** Don’t make Grafana a requirement for product dashboards; keep product UI primary.

### 149.3 Loki (logs)
Use Loki/Promtail for:
- backend logs
- agent logs
- correlation during outages

### 149.4 Alertmanager (optional)
If you integrate with Prometheus stack:
- Alertmanager can handle “platform internal alerts”
- but for ISP device alerts, your own incident system should be primary

### 149.5 Export to external NOC systems
- Webhook integration (primary)
- CSV/PDF exports for reporting
- Syslog export for audit logs

---

## 150. Operator handbook outline (self-hosted ISP)

This is a doc set outline you can ship as `docs/operator-handbook/`.

### 150.1 Sections
1. **Architecture overview**
2. **Installation**
   - compose profiles
   - TLS options
3. **Agent deployment**
   - enrollment
   - placement per PoP
4. **Adding routers**
   - credential sets
   - connectivity testing
5. **Tagging and topology**
   - roles
   - uplinks
   - dependencies
6. **Alert packs**
   - enabling and tuning
7. **Incident response**
   - acknowledge/assign
   - using evidence and diagnostics
8. **Maintenance and change workflows**
9. **Backups and restore**
10. **Upgrades**
11. **Troubleshooting**
12. **Security**
   - hardening
   - audits
   - SIEM export

### 150.2 Runbook library linkage
Every alert should link to a runbook in this handbook.

---

## 151. Threat model appendix (STRIDE) — agent + backend

### 151.1 Why STRIDE helps
Enterprises like structured security reasoning. STRIDE covers:
- Spoofing
- Tampering
- Repudiation
- Information Disclosure
- Denial of Service
- Elevation of Privilege

### 151.2 Agent threats (examples)
**Spoofing**
- attacker impersonates agent → injects fake metrics
Mitigation:
- mTLS, cert pinning, revocation

**Tampering**
- attacker modifies buffered payloads on disk
Mitigation:
- file permissions, optional payload signing, secure host

**Repudiation**
- agent claims it didn’t send data
Mitigation:
- audit logs on backend, delivery IDs

**Information disclosure**
- secrets in logs
Mitigation:
- secret redaction, structured logging policies

**DoS**
- agent floods backend
Mitigation:
- rate limiting, payload size limits, backoff

**EoP**
- agent runs as root and gets exploited
Mitigation:
- run as non-root, minimal capabilities, patching

### 151.3 Backend threats (examples)
**Spoofing**
- stolen user tokens
Mitigation:
- OIDC, short-lived tokens, MFA (enterprise)

**Tampering**
- database injection
Mitigation:
- parameterized queries, RLS, code review

**Repudiation**
- admins deny changes
Mitigation:
- immutable audit logs, signed exports

**Information disclosure**
- cross-tenant leaks
Mitigation:
- tenant isolation (RLS), tests, code guards

**DoS**
- ingest storms
Mitigation:
- queue, rate limits, worker pools

**EoP**
- RBAC bypass
Mitigation:
- centralized authorization middleware, audits, tests

---

## 152. v1.0 acceptance checklist (launch readiness)

Use this as a pre-launch checklist.

### 152.1 Functional correctness
- [ ] Inventory CRUD works (sites/devices/interfaces/links)
- [ ] Agent enrollment (mTLS) works and rotation is tested
- [ ] Ingest accepts batches and stores metrics/events correctly
- [ ] Rollups and retention run without breaking dashboards
- [ ] Alert packs function and are tunable
- [ ] Incidents correlate and suppress downstream alerts
- [ ] RCA v1 identifies plausible root causes with evidence
- [ ] Notifications work (telegram + email + webhook) with retries
- [ ] Diagnostics (ping/traceroute) works and is audited
- [ ] Maintenance windows suppress notifications properly
- [ ] Change detection fingerprints produce events and correlate

### 152.2 Security
- [ ] RBAC enforced everywhere
- [ ] Secrets encrypted at rest; master key management documented
- [ ] Audit logs enabled; signed exports supported (enterprise or optional)
- [ ] Agent binaries/images signed; SBOM published
- [ ] Rate limits configured on ingest and auth endpoints

### 152.3 Operability
- [ ] Compose deployment documented and tested
- [ ] Backup/restore tested end-to-end
- [ ] Upgrade process documented and tested (N-1 to N)
- [ ] Observability: `/metrics` exposed and dashboards exist (internal)
- [ ] Performance benchmarks documented (router count targets)

### 152.4 Documentation
- [ ] Operator handbook published
- [ ] Data collected & privacy policy published
- [ ] Runbooks linked from alert rules
- [ ] Troubleshooting guide covers common failures

---

## 153. “Continue” note
Next chapters could include:
- concrete external integration examples (Jira webhook receiver, ServiceNow mapping)
- a complete STRIDE table (assets, threats, mitigations) as a formal appendix
- “blue team” security checks for ISPs (hardening scripts)
- a full “demo kit” repository structure and automation scripts

Say **continue** to proceed.