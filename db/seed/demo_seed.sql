-- =============================================================================
-- ISP Visual Monitor — Demo Seed Data
-- Provides a realistic small-ISP topology for demo / development.
--
-- Topology:
--   1 tenant  (LebanonNet ISP)
--   3 sites   (Beirut-DC, Tripoli-POP, Sidon-POP)
--   10 routers covering core, edge, access, upstream, pppoe roles
--   Interface inventory per router
--   Active + historical alerts / incidents
--   Baseline telemetry for dashboards
--
-- Safe to re-run: uses ON CONFLICT DO NOTHING.
-- =============================================================================

-- ============================================================================
-- 0. EXTENSIONS (idempotent)
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================================================
-- 1. DEMO TENANT
-- ============================================================================
INSERT INTO tenants (id, name, slug, contact_email, subscription_tier, max_devices, status)
VALUES (
  'a0000000-0000-0000-0000-000000000001',
  'LebanonNet ISP',
  'lebanonnet',
  'noc@lebanonnet.demo',
  'enterprise',
  200,
  'active'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. DEMO USER (password: password — bcrypt cost 10)
-- ============================================================================
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, email_verified)
VALUES (
  'b0000000-0000-0000-0000-000000000001',
  'a0000000-0000-0000-0000-000000000001',
  'demo@lebanonnet.demo',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'Demo',
  'Admin',
  'active',
  true
) ON CONFLICT DO NOTHING;

-- ============================================================================
-- 3. REGIONS (geographic grouping)
-- ============================================================================
INSERT INTO regions (id, tenant_id, name, description) VALUES
  ('c0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'Beirut Metro', 'Greater Beirut area'),
  ('c0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'North Lebanon', 'Tripoli and surroundings'),
  ('c0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'South Lebanon', 'Sidon and surroundings')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 4. POPS (Points of Presence) — 3 sites
-- ============================================================================
INSERT INTO pops (id, tenant_id, region_id, name, code, location, address, pop_type, capacity_gbps, status) VALUES
  (
    'd0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000001',
    'Beirut Data Center',
    'BEY-DC1',
    ST_SetSRID(ST_MakePoint(35.5018, 33.8938), 4326),
    '123 Downtown St, Beirut',
    'datacenter',
    40,
    'active'
  ),
  (
    'd0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000002',
    'Tripoli POP',
    'TRI-POP1',
    ST_SetSRID(ST_MakePoint(35.8497, 34.4332), 4326),
    '45 Al-Mina Rd, Tripoli',
    'co-location',
    10,
    'active'
  ),
  (
    'd0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000003',
    'Sidon POP',
    'SID-POP1',
    ST_SetSRID(ST_MakePoint(35.3716, 33.5571), 4326),
    '78 Sea Rd, Sidon',
    'edge',
    10,
    'active'
  )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 5. ROUTERS — 10 devices across 3 sites
-- ============================================================================

-- Beirut-DC (core site) — 5 routers
INSERT INTO routers (id, tenant_id, pop_id, name, hostname, management_ip, location, router_type, vendor, model, os_version, status, polling_enabled, polling_interval_seconds) VALUES
  ('e0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001',
   'BEY-CORE-01', 'bey-core-01.lab', '10.0.0.1',
   ST_SetSRID(ST_MakePoint(35.5018, 33.8938), 4326), 'core', 'MikroTik', 'CCR2216-1G-12XS-2XQ', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001',
   'BEY-CORE-02', 'bey-core-02.lab', '10.0.0.2',
   ST_SetSRID(ST_MakePoint(35.5030, 33.8945), 4326), 'core', 'MikroTik', 'CCR2216-1G-12XS-2XQ', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001',
   'BEY-EDGE-01', 'bey-edge-01.lab', '10.0.1.1',
   ST_SetSRID(ST_MakePoint(35.5010, 33.8930), 4326), 'edge', 'MikroTik', 'CCR1036-8G-2S+', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001',
   'BEY-UPSTREAM-01', 'bey-upstream-01.lab', '10.0.2.1',
   ST_SetSRID(ST_MakePoint(35.5025, 33.8950), 4326), 'border', 'MikroTik', 'CCR1072-1G-8S+', '7.14', 'active', true, 120),

  ('e0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001',
   'BEY-PPPOE-01', 'bey-pppoe-01.lab', '10.0.3.1',
   ST_SetSRID(ST_MakePoint(35.5035, 33.8935), 4326), 'access', 'MikroTik', 'CCR2004-16G-2S+', '7.14', 'active', true, 30)
ON CONFLICT DO NOTHING;

-- Tripoli-POP — 3 routers
INSERT INTO routers (id, tenant_id, pop_id, name, hostname, management_ip, location, router_type, vendor, model, os_version, status, polling_enabled, polling_interval_seconds) VALUES
  ('e0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000002',
   'TRI-EDGE-01', 'tri-edge-01.lab', '10.1.0.1',
   ST_SetSRID(ST_MakePoint(35.8497, 34.4332), 4326), 'edge', 'MikroTik', 'CCR1036-8G-2S+', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000002',
   'TRI-ACCESS-01', 'tri-access-01.lab', '10.1.1.1',
   ST_SetSRID(ST_MakePoint(35.8510, 34.4340), 4326), 'access', 'MikroTik', 'RB4011iGS+', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000002',
   'TRI-PPPOE-01', 'tri-pppoe-01.lab', '10.1.2.1',
   ST_SetSRID(ST_MakePoint(35.8505, 34.4325), 4326), 'access', 'MikroTik', 'CCR2004-16G-2S+', '7.14', 'active', true, 30)
ON CONFLICT DO NOTHING;

-- Sidon-POP — 2 routers
INSERT INTO routers (id, tenant_id, pop_id, name, hostname, management_ip, location, router_type, vendor, model, os_version, status, polling_enabled, polling_interval_seconds) VALUES
  ('e0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000003',
   'SID-EDGE-01', 'sid-edge-01.lab', '10.2.0.1',
   ST_SetSRID(ST_MakePoint(35.3716, 33.5571), 4326), 'edge', 'MikroTik', 'CCR1036-8G-2S+', '7.14', 'active', true, 60),

  ('e0000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000003',
   'SID-ACCESS-01', 'sid-access-01.lab', '10.2.1.1',
   ST_SetSRID(ST_MakePoint(35.3720, 33.5580), 4326), 'access', 'MikroTik', 'RB4011iGS+', '7.14', 'offline', true, 60)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 6. INTERFACES — representative inventory per router
-- ============================================================================

-- Helper: BEY-CORE-01 interfaces
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0001-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'ether1-wan',   'ethernet', 10000, 'up',   '203.0.113.1'),
  ('f0000000-0000-0000-0001-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'ether2-core',  'ethernet', 10000, 'up',   '10.0.0.1'),
  ('f0000000-0000-0000-0001-000000000003', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'sfp1-uplink',  'optical',  25000, 'up',   '10.255.0.1'),
  ('f0000000-0000-0000-0001-000000000004', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'sfp2-tri',     'optical',  10000, 'up',   '10.255.1.1')
ON CONFLICT DO NOTHING;

-- BEY-CORE-02
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0002-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000002', 'ether1-core',  'ethernet', 10000, 'up',   '10.0.0.2'),
  ('f0000000-0000-0000-0002-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000002', 'sfp1-sid',     'optical',  10000, 'up',   '10.255.2.1'),
  ('f0000000-0000-0000-0002-000000000003', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000002', 'ether2-edge',  'ethernet', 10000, 'up',   '10.0.0.10')
ON CONFLICT DO NOTHING;

-- BEY-EDGE-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0003-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000003', 'ether1-core',  'ethernet', 10000, 'up',   '10.0.0.11'),
  ('f0000000-0000-0000-0003-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000003', 'ether2-pppoe', 'ethernet', 1000,  'up',   '10.0.3.254'),
  ('f0000000-0000-0000-0003-000000000003', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000003', 'ether3-mgmt',  'ethernet', 1000,  'up',   '10.0.1.1')
ON CONFLICT DO NOTHING;

-- BEY-UPSTREAM-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0004-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000004', 'ether1-isp-a', 'ethernet', 10000, 'up',   '198.51.100.2'),
  ('f0000000-0000-0000-0004-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000004', 'ether2-isp-b', 'ethernet', 10000, 'up',   '198.51.200.2'),
  ('f0000000-0000-0000-0004-000000000003', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000004', 'ether3-core',  'ethernet', 10000, 'up',   '10.0.2.1')
ON CONFLICT DO NOTHING;

-- BEY-PPPOE-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0005-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000005', 'ether1-edge',  'ethernet', 1000, 'up',   '10.0.3.1'),
  ('f0000000-0000-0000-0005-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000005', 'pppoe-pool',   'ethernet', 1000, 'up',   '10.100.0.1')
ON CONFLICT DO NOTHING;

-- TRI-EDGE-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0006-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000006', 'sfp1-core',    'optical',  10000, 'up',   '10.255.1.2'),
  ('f0000000-0000-0000-0006-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000006', 'ether1-access','ethernet', 1000,  'up',   '10.1.0.1'),
  ('f0000000-0000-0000-0006-000000000003', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000006', 'ether2-pppoe', 'ethernet', 1000,  'up',   '10.1.2.254')
ON CONFLICT DO NOTHING;

-- TRI-ACCESS-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0007-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000007', 'ether1-edge',  'ethernet', 1000, 'up',   '10.1.0.2'),
  ('f0000000-0000-0000-0007-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000007', 'ether2-lan',   'ethernet', 1000, 'up',   '10.1.1.1')
ON CONFLICT DO NOTHING;

-- TRI-PPPOE-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0008-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000008', 'ether1-edge',  'ethernet', 1000, 'up',   '10.1.2.1'),
  ('f0000000-0000-0000-0008-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000008', 'pppoe-pool',   'ethernet', 1000, 'up',   '10.101.0.1')
ON CONFLICT DO NOTHING;

-- SID-EDGE-01
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0009-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000009', 'sfp1-core',    'optical',  10000, 'up',   '10.255.2.2'),
  ('f0000000-0000-0000-0009-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000009', 'ether1-access','ethernet', 1000,  'up',   '10.2.0.1')
ON CONFLICT DO NOTHING;

-- SID-ACCESS-01 (offline router)
INSERT INTO interfaces (id, tenant_id, router_id, name, if_type, speed_mbps, status, ip_address) VALUES
  ('f0000000-0000-0000-0010-000000000001', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000010', 'ether1-edge',  'ethernet', 1000, 'down', '10.2.1.1'),
  ('f0000000-0000-0000-0010-000000000002', 'a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000010', 'ether2-lan',   'ethernet', 1000, 'down', '10.2.1.254')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 7. LINKS — inter-router connections
-- ============================================================================
INSERT INTO links (id, tenant_id, name, source_interface_id, target_interface_id, link_type, capacity_mbps, status) VALUES
  -- Core-01 <-> Core-02
  ('aa000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
   'BEY-CORE-01 <-> BEY-CORE-02',
   'f0000000-0000-0000-0001-000000000002', 'f0000000-0000-0000-0002-000000000001',
   'physical', 10000, 'up'),

  -- Core-01 <-> Tripoli (SFP)
  ('aa000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
   'BEY-CORE-01 <-> TRI-EDGE-01',
   'f0000000-0000-0000-0001-000000000004', 'f0000000-0000-0000-0006-000000000001',
   'physical', 10000, 'up'),

  -- Core-02 <-> Sidon (SFP)
  ('aa000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
   'BEY-CORE-02 <-> SID-EDGE-01',
   'f0000000-0000-0000-0002-000000000002', 'f0000000-0000-0000-0009-000000000001',
   'physical', 10000, 'up'),

  -- Core-02 <-> Edge-01
  ('aa000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
   'BEY-CORE-02 <-> BEY-EDGE-01',
   'f0000000-0000-0000-0002-000000000003', 'f0000000-0000-0000-0003-000000000001',
   'physical', 10000, 'up'),

  -- Edge-01 <-> PPPoE-01
  ('aa000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
   'BEY-EDGE-01 <-> BEY-PPPOE-01',
   'f0000000-0000-0000-0003-000000000002', 'f0000000-0000-0000-0005-000000000001',
   'physical', 1000, 'up'),

  -- Core-01 <-> Upstream-01
  ('aa000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
   'BEY-CORE-01 <-> BEY-UPSTREAM-01',
   'f0000000-0000-0000-0001-000000000003', 'f0000000-0000-0000-0004-000000000003',
   'physical', 25000, 'up'),

  -- TRI-EDGE-01 <-> TRI-ACCESS-01
  ('aa000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
   'TRI-EDGE-01 <-> TRI-ACCESS-01',
   'f0000000-0000-0000-0006-000000000002', 'f0000000-0000-0000-0007-000000000001',
   'physical', 1000, 'up'),

  -- TRI-EDGE-01 <-> TRI-PPPOE-01
  ('aa000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001',
   'TRI-EDGE-01 <-> TRI-PPPOE-01',
   'f0000000-0000-0000-0006-000000000003', 'f0000000-0000-0000-0008-000000000001',
   'physical', 1000, 'up'),

  -- SID-EDGE-01 <-> SID-ACCESS-01
  ('aa000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001',
   'SID-EDGE-01 <-> SID-ACCESS-01',
   'f0000000-0000-0000-0009-000000000002', 'f0000000-0000-0000-0010-000000000001',
   'physical', 1000, 'down')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 8. ROUTER ROLE ASSIGNMENTS
-- ============================================================================
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000001', id, 100, true FROM router_roles WHERE code = 'core_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000002', id, 100, true FROM router_roles WHERE code = 'core_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000003', id, 100, true FROM router_roles WHERE code = 'edge_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000004', id, 100, true FROM router_roles WHERE code = 'border_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000005', id, 100, true FROM router_roles WHERE code = 'pppoe_server'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000005', id, 200, false FROM router_roles WHERE code = 'dhcp_server'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000006', id, 100, true FROM router_roles WHERE code = 'edge_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000007', id, 100, true FROM router_roles WHERE code = 'access_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000008', id, 100, true FROM router_roles WHERE code = 'pppoe_server'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000009', id, 100, true FROM router_roles WHERE code = 'edge_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT 'e0000000-0000-0000-0000-000000000010', id, 100, true FROM router_roles WHERE code = 'access_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

-- ============================================================================
-- 9. ROUTER DEPENDENCIES
-- ============================================================================
INSERT INTO router_dependencies (tenant_id, source_router_id, target_router_id, dependency_type, weight, is_active, description) VALUES
  -- Upstream -> Core
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000004', 'upstream', 100, true, 'Core-01 routes to upstream'),
  -- Core pair
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000002', 'peer', 100, true, 'Core HA pair'),
  -- Edge -> Core
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000002', 'upstream', 100, true, 'Edge routes through Core-02'),
  -- PPPoE -> Edge
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000005', 'e0000000-0000-0000-0000-000000000003', 'upstream', 100, true, 'PPPoE routes through Edge'),
  -- Tripoli -> Core
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000006', 'e0000000-0000-0000-0000-000000000001', 'upstream', 100, true, 'Tripoli Edge connects to Core-01'),
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000007', 'e0000000-0000-0000-0000-000000000006', 'upstream', 100, true, 'Tripoli Access routes through Tripoli Edge'),
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000008', 'e0000000-0000-0000-0000-000000000006', 'upstream', 100, true, 'Tripoli PPPoE routes through Tripoli Edge'),
  -- Sidon -> Core
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000009', 'e0000000-0000-0000-0000-000000000002', 'upstream', 100, true, 'Sidon Edge connects to Core-02'),
  ('a0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000010', 'e0000000-0000-0000-0000-000000000009', 'upstream', 100, true, 'Sidon Access routes through Sidon Edge')
ON CONFLICT (source_router_id, target_router_id, dependency_type) DO NOTHING;

-- ============================================================================
-- 10. ALERTS — mix of active, acknowledged, and resolved
-- ============================================================================
INSERT INTO alerts (id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata) VALUES
  ('bb000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
   'SID-ACCESS-01 offline',
   'Router SID-ACCESS-01 has been unreachable for 25 minutes. Last poll failed with timeout.',
   'critical', 'active', 'router', 'e0000000-0000-0000-0000-000000000010',
   NOW() - INTERVAL '25 minutes',
   '{"consecutive_failures": 5, "last_successful_poll": "2025-01-01T10:00:00Z"}'),

  ('bb000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
   'High CPU on BEY-CORE-01',
   'CPU utilization at 87% for the last 10 minutes. Threshold: 80%.',
   'warning', 'active', 'router', 'e0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '10 minutes',
   '{"cpu_percent": 87, "threshold": 80}'),

  ('bb000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
   'Core link congestion BEY-CORE-01 sfp1-uplink',
   'Interface utilization at 92%. Possible congestion on upstream link.',
   'warning', 'active', 'interface', 'f0000000-0000-0000-0001-000000000003',
   NOW() - INTERVAL '8 minutes',
   '{"utilization_percent": 92, "threshold": 85}'),

  ('bb000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
   'Packet loss on TRI-EDGE-01 sfp1-core',
   'Packet loss rate 2.3% on Tripoli uplink over last 5 minutes.',
   'warning', 'acknowledged', 'interface', 'f0000000-0000-0000-0006-000000000001',
   NOW() - INTERVAL '45 minutes',
   '{"packet_loss_percent": 2.3, "threshold": 1.0}'),

  ('bb000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
   'High PPPoE session count on BEY-PPPOE-01',
   'Active PPPoE sessions at 847 out of 1000 capacity (84.7%).',
   'info', 'active', 'router', 'e0000000-0000-0000-0000-000000000005',
   NOW() - INTERVAL '15 minutes',
   '{"active_sessions": 847, "max_sessions": 1000}'),

  ('bb000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
   'Upstream ISP-B link flap',
   'Interface ether2-isp-b on BEY-UPSTREAM-01 went down and recovered. 3 flaps in 1 hour.',
   'warning', 'resolved', 'interface', 'f0000000-0000-0000-0004-000000000002',
   NOW() - INTERVAL '2 hours',
   '{"flap_count": 3, "time_window": "1h"}'),

  ('bb000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
   'Memory usage warning on TRI-PPPOE-01',
   'Memory utilization at 78%. Approaching warning threshold of 80%.',
   'info', 'resolved', 'router', 'e0000000-0000-0000-0000-000000000008',
   NOW() - INTERVAL '3 hours',
   '{"memory_percent": 78, "threshold": 80}')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 11. ROUTER METRICS — baseline telemetry (last 2 hours, every 5 min)
-- ============================================================================
INSERT INTO router_metrics (tenant_id, router_id, timestamp, cpu_percent, memory_percent, uptime_seconds, temperature_celsius)
SELECT
  'a0000000-0000-0000-0000-000000000001',
  r.id,
  ts,
  -- CPU: base + noise, higher for core routers
  CASE WHEN r.router_type = 'core' THEN 35 + random() * 25
       WHEN r.router_type = 'border' THEN 20 + random() * 15
       ELSE 15 + random() * 20
  END,
  -- Memory: base + noise
  CASE WHEN r.router_type = 'core' THEN 55 + random() * 15
       ELSE 40 + random() * 20
  END,
  -- Uptime: 30-90 days in seconds
  (30 + floor(random() * 60)::int) * 86400,
  -- Temperature
  38 + random() * 12
FROM routers r
CROSS JOIN generate_series(
  NOW() - INTERVAL '2 hours',
  NOW(),
  INTERVAL '5 minutes'
) AS ts
WHERE r.tenant_id = 'a0000000-0000-0000-0000-000000000001'
  AND r.status = 'active';

-- ============================================================================
-- 12. INTERFACE METRICS — baseline traffic
-- ============================================================================
INSERT INTO interface_metrics (tenant_id, interface_id, timestamp, in_octets, out_octets, in_packets, out_packets, in_errors, out_errors, utilization_percent)
SELECT
  'a0000000-0000-0000-0000-000000000001',
  i.id,
  ts,
  -- in_octets: scale by interface speed
  (i.speed_mbps * 125000 * (0.3 + random() * 0.4))::bigint,
  (i.speed_mbps * 125000 * (0.15 + random() * 0.25))::bigint,
  (10000 + random() * 50000)::bigint,
  (5000 + random() * 25000)::bigint,
  (random() * 5)::bigint,
  (random() * 3)::bigint,
  30 + random() * 40
FROM interfaces i
CROSS JOIN generate_series(
  NOW() - INTERVAL '2 hours',
  NOW(),
  INTERVAL '5 minutes'
) AS ts
WHERE i.tenant_id = 'a0000000-0000-0000-0000-000000000001'
  AND i.status = 'up';

-- ============================================================================
-- 13. PPPoE SESSIONS — realistic session pool
-- ============================================================================
INSERT INTO pppoe_sessions (tenant_id, router_id, session_id, username, calling_station_id, framed_ip_address, status, connect_time, bytes_in, bytes_out)
SELECT
  'a0000000-0000-0000-0000-000000000001',
  router_id,
  'pppoe-demo-' || n,
  'user' || lpad(n::text, 4, '0') || '@lebanonnet.demo',
  '00:DE:MO:' || lpad(to_hex(n / 256), 2, '0') || ':' || lpad(to_hex(n % 256), 2, '0') || ':01',
  ('10.100.' || (n / 256) || '.' || (n % 256))::inet,
  'active',
  NOW() - (random() * INTERVAL '8 hours'),
  (random() * 2000000000)::bigint,
  (random() * 500000000)::bigint
FROM (
  SELECT 'e0000000-0000-0000-0000-000000000005'::uuid AS router_id, generate_series(1, 30) AS n
  UNION ALL
  SELECT 'e0000000-0000-0000-0000-000000000008'::uuid, generate_series(31, 50)
) sub
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 14. SUMMARY
-- ============================================================================
-- After running this seed:
--   1 tenant   : LebanonNet ISP
--   1 user     : demo@lebanonnet.demo / password
--   3 regions  : Beirut Metro, North Lebanon, South Lebanon
--   3 POPs     : BEY-DC1, TRI-POP1, SID-POP1
--   10 routers : 2 core, 1 edge, 1 upstream/border, 2 pppoe, 2 access, 2 edge (remote)
--   30+ interfaces
--   9 links
--   7 alerts (3 active, 1 acknowledged, 2 resolved, 1 info)
--   ~2 hours of router + interface metrics
--   50 active PPPoE sessions
