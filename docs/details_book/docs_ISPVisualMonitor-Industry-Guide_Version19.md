# ISPVisualMonitor — Industry‑Grade Architecture & Implementation Guide (MikroTik‑first)

**Audience:** Network/ISP engineers, SRE/DevOps, backend engineers, security reviewers  
**Scope:** MikroTik RouterOS first; extensible to other vendors later  
**Repos:**  
- Control plane: `MohamadKhaledAbbas/ISPVisualMonitor`  
- Data plane agent (open source): `MohamadKhaledAbbas/ISPVisualMonitor-Agent`

**Status:** Living document. Say **“continue”** to append the next chapters.

---

## Parts I–XVIII
Earlier parts cover dashboards, service modeling, and postmortems. This part covers configuration change detection, maintenance workflows/change calendars, and more advanced security patterns (per-tenant keys, Vault/KMS), plus transport options (gRPC/NATS) for higher scale.

---

# Part XIX — Change Detection, Maintenance Workflows, Advanced Security, and High‑Scale Transport

## 132. Configuration change detection (safe and ISP-friendly)

### 132.1 Why config change detection matters
Many outages are self-inflicted:
- routing policy change
- firewall rule change
- interface disable
- BGP filters updated incorrectly

If your system can say:
- “This started 2 minutes after a config change on EDGE1”
you become dramatically more valuable.

### 132.2 Levels of config change detection

#### Level 1 (safe MVP): “fingerprint only”
- periodically collect a *normalized* subset of config or state
- compute a hash (fingerprint)
- store only:
  - fingerprint
  - timestamp
  - change/no change

Pros:
- minimal sensitive data storage
- low risk for self-hosted adoption
- good enough to correlate incidents with changes

Cons:
- doesn’t tell you what changed unless you also collect diffs

#### Level 2: “diff of selected sections”
- collect a small allowlist of configuration sections:
  - BGP peers config
  - OSPF config
  - IP addresses and routes policy
  - critical interface config
- store canonical JSON and compute diffs

Pros:
- actionable
Cons:
- higher sensitivity and storage

#### Level 3: “full config snapshot” (high sensitivity)
- store full export / config script
- strong RBAC + encryption + retention required
- many ISPs will resist unless they fully trust you

### 132.3 MikroTik-specific considerations
RouterOS configuration can be exported in various forms.
For a safe approach:
- collect operational state relevant to outages:
  - BGP peer definitions and filters (if accessible safely)
  - interface admin state
  - routing table summary and policy counts
- avoid collecting secrets:
  - passwords
  - keys
  - SNMP communities
  - certificates private keys

### 132.4 Change event generation
When fingerprint changes:
- emit event: `config.changed.detected`
- payload:
  - old_hash, new_hash
  - category (bgp/ospf/interfaces/system)
  - device_id
- optionally attach “change summary” if Level 2

### 132.5 Change correlation rule
Add a correlation rule:
- if incident starts within X minutes of config change event on impacted root cause candidate:
  - add evidence weight: +N
  - show “Possible change-related outage” label

This is explainable and useful.

---

## 133. Maintenance windows and change calendars (operator workflows)

### 133.1 Maintenance windows (already discussed; deepen here)
Maintenance windows should support:
- time range
- scope:
  - site(s)
  - device(s)
  - tags
  - services
- behavior:
  - suppress notifications
  - optionally mark alerts as “maintenance” not “incident”
  - still collect metrics (always collect)

### 133.2 Change calendar (why it matters)
Large ISPs plan changes:
- link upgrades
- routing policy changes
- PoP maintenance

A change calendar allows:
- correlation (“outage during scheduled maintenance”)
- better communication
- auditing and compliance

### 133.3 Change record entity (recommended)
Create:
- `changes`
  - `id`
  - `tenant_id`
  - `title`
  - `description`
  - `starts_at`, `ends_at`
  - `scope` (sites/devices/tags/services)
  - `owner_user_id`
  - `status` (planned/in-progress/completed/canceled)
  - `created_at`

And:
- `change_incidents` mapping if incidents occurred during that window.

### 133.4 How to use change records in alerting
During change window:
- suppress notifications for scoped entities
- but still create incidents internally marked:
  - `during_change_id`

This prevents missing real outages while avoiding spam.

---

## 134. Advanced security: per-tenant keys, Vault/KMS, and secret rotation

### 134.1 Why a single master key may not be enough
A single encryption master key for all tenants:
- is simpler
- but increases blast radius if compromised

For enterprise/SaaS, you should plan per-tenant keys.

### 134.2 Key management levels

#### Level 1: One master key (acceptable early)
- environment-provided `ENCRYPTION_MASTER_KEY`
- encrypt all credential blobs

#### Level 2: Master key + per-tenant derived keys (recommended mid-term)
- derive tenant key from master using HKDF:
  - `tenant_key = HKDF(master_key, tenant_id)`
- store encrypted blobs with tenant key
- this limits cross-tenant impact if a blob leaks

#### Level 3: External KMS/Vault (enterprise/SaaS)
- store only ciphertext in DB
- use KMS to encrypt/decrypt
- audit every decrypt operation centrally
- support rotation without downtime

### 134.3 Secret rotation strategy
- rotate router credentials periodically (ISP policy)
- rotate agent certs automatically (mTLS)
- rotate encryption keys with re-encryption process:
  - read ciphertext, decrypt with old, encrypt with new
  - run in background job
  - maintain “key version” metadata

### 134.4 Audit and access control for secrets
- only Admin role can create/edit credential sets
- never show plaintext after creation (write-only UI)
- log every secret usage event (optional high security mode):
  - “agent X requested credential set Y at time T”

---

## 135. High-scale transport options (HTTP → gRPC → NATS)

### 135.1 When HTTP is enough
HTTP ingest is sufficient for:
- small/mid deployments
- moderate ingest rates
- simple operations

If you do:
- batching + compression
- queue decoupling (optional)
HTTP can go quite far.

### 135.2 gRPC ingest (benefits and tradeoffs)
Benefits:
- strongly typed protobuf schemas
- better performance for high throughput
- streaming options

Tradeoffs:
- more complex debugging for operators
- proxying and load balancers sometimes require tuning

### 135.3 NATS JetStream ingest (event-driven architecture)
Pattern:
- agent publishes to NATS subject:
  - `tenant.<id>.ingest`
- workers consume and persist

Benefits:
- built-in buffering and replay
- good for multi-worker scaling
- decouples ingest from DB

Tradeoffs:
- adds operational component
- multi-tenant security requires careful subject permissions

### 135.4 Recommended adoption path
- v0.1–v0.5: HTTP ingest + direct DB writes
- v0.8: optional queue (NATS) for decoupling
- v1.x: consider gRPC for SaaS/high scale, keep HTTP compatibility for self-hosted

---

## 136. “Continue” note
Next chapters can include:
- full “change detection allowlist” for RouterOS (what to hash, what to exclude)
- a full incident/change/maintenance UI workflow specification
- SIEM integration patterns (export events to syslog/CEF/JSON)
- compliance add-ons: data retention per category, legal holds, audit export signing
- advanced service models: upstream provider SLAs, customer SLAs, multi-service PoPs

Say **continue** to proceed.