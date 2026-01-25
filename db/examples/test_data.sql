-- Sample Test Data for ISP Visual Monitor
-- This script demonstrates the enhanced multi-role router capabilities
-- Run this after applying both migration scripts

-- ============================================================================
-- 1. CREATE TEST TENANT
-- ============================================================================

INSERT INTO tenants (id, name, slug, contact_email, subscription_tier, max_devices, status)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'Example ISP', 'example-isp', 'admin@example-isp.com', 'enterprise', 100, 'active')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. CREATE TEST ROUTERS
-- ============================================================================

-- Core Router
INSERT INTO routers (
    id, tenant_id, name, management_ip, vendor, model, os_version,
    status, polling_enabled, polling_interval_seconds
) VALUES (
    '22222222-2222-2222-2222-222222222221',
    '11111111-1111-1111-1111-111111111111',
    'core-router-01',
    '172.25.0.10',
    'MikroTik',
    'CCR1036-12G-4S',
    '7.11',
    'active',
    true,
    60
) ON CONFLICT (id) DO NOTHING;

-- Edge Router
INSERT INTO routers (
    id, tenant_id, name, management_ip, vendor, model, os_version,
    status, polling_enabled, polling_interval_seconds
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'edge-router-01',
    '172.25.0.11',
    'MikroTik',
    'CCR1036-8G-2S+',
    '7.11',
    'active',
    true,
    60
) ON CONFLICT (id) DO NOTHING;

-- Access Router with NAT
INSERT INTO routers (
    id, tenant_id, name, management_ip, vendor, model, os_version,
    status, polling_enabled, polling_interval_seconds
) VALUES (
    '22222222-2222-2222-2222-222222222223',
    '11111111-1111-1111-1111-111111111111',
    'access-router-01',
    '172.25.0.12',
    'MikroTik',
    'RB4011iGS+',
    '7.11',
    'active',
    true,
    60
) ON CONFLICT (id) DO NOTHING;

-- PPPoE Server
INSERT INTO routers (
    id, tenant_id, name, management_ip, vendor, model, os_version,
    status, polling_enabled, polling_interval_seconds
) VALUES (
    '22222222-2222-2222-2222-222222222224',
    '11111111-1111-1111-1111-111111111111',
    'pppoe-server-01',
    '172.25.0.13',
    'MikroTik',
    'CCR2004-16G-2S+',
    '7.11',
    'active',
    true,
    30
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 3. CONFIGURE ROUTER CAPABILITIES
-- ============================================================================

-- Core Router - MikroTik API + SNMP
INSERT INTO router_capabilities (
    router_id, tenant_id,
    snmp_enabled, snmp_version, snmp_community, snmp_port, snmp_timeout_seconds, snmp_retries,
    api_enabled, api_type, api_endpoint, api_port, api_username, api_password, api_use_tls, api_verify_cert, api_timeout_seconds,
    preferred_method, fallback_order
) VALUES (
    '22222222-2222-2222-2222-222222222221',
    '11111111-1111-1111-1111-111111111111',
    true, 'v2c', 'public', 161, 10, 3,
    true, 'mikrotik', '172.25.0.10', 8728, 'monitor', 'monitor123', false, false, 30,
    'api',
    ARRAY['api', 'snmp']
) ON CONFLICT (router_id) DO NOTHING;

-- Edge Router - MikroTik API + SNMP
INSERT INTO router_capabilities (
    router_id, tenant_id,
    snmp_enabled, snmp_version, snmp_community, snmp_port, snmp_timeout_seconds, snmp_retries,
    api_enabled, api_type, api_endpoint, api_port, api_username, api_password, api_use_tls, api_verify_cert, api_timeout_seconds,
    preferred_method, fallback_order
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    true, 'v2c', 'public', 162, 10, 3,
    true, 'mikrotik', '172.25.0.11', 8738, 'monitor', 'monitor123', false, false, 30,
    'api',
    ARRAY['api', 'snmp']
) ON CONFLICT (router_id) DO NOTHING;

-- Access Router - MikroTik API + SNMP
INSERT INTO router_capabilities (
    router_id, tenant_id,
    snmp_enabled, snmp_version, snmp_community, snmp_port, snmp_timeout_seconds, snmp_retries,
    api_enabled, api_type, api_endpoint, api_port, api_username, api_password, api_use_tls, api_verify_cert, api_timeout_seconds,
    preferred_method, fallback_order
) VALUES (
    '22222222-2222-2222-2222-222222222223',
    '11111111-1111-1111-1111-111111111111',
    true, 'v2c', 'public', 163, 10, 3,
    true, 'mikrotik', '172.25.0.12', 8748, 'monitor', 'monitor123', false, false, 30,
    'api',
    ARRAY['api', 'snmp']
) ON CONFLICT (router_id) DO NOTHING;

-- PPPoE Server - MikroTik API + SNMP (with faster polling)
INSERT INTO router_capabilities (
    router_id, tenant_id,
    snmp_enabled, snmp_version, snmp_community, snmp_port, snmp_timeout_seconds, snmp_retries,
    api_enabled, api_type, api_endpoint, api_port, api_username, api_password, api_use_tls, api_verify_cert, api_timeout_seconds,
    preferred_method, fallback_order
) VALUES (
    '22222222-2222-2222-2222-222222222224',
    '11111111-1111-1111-1111-111111111111',
    true, 'v2c', 'public', 164, 10, 3,
    true, 'mikrotik', '172.25.0.13', 8758, 'monitor', 'monitor123', false, false, 30,
    'api',
    ARRAY['api', 'snmp']
) ON CONFLICT (router_id) DO NOTHING;

-- ============================================================================
-- 4. ASSIGN ROUTER ROLES
-- ============================================================================

-- Core Router
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222221', id, 100, true
FROM router_roles WHERE code = 'core_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

-- Edge Router (with load balancer role)
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222222', id, 100, true
FROM router_roles WHERE code = 'edge_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222222', id, 200, false
FROM router_roles WHERE code = 'load_balancer'
ON CONFLICT (router_id, role_id) DO NOTHING;

-- Access Router (multi-role: access + NAT + firewall)
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222223', id, 100, true
FROM router_roles WHERE code = 'access_router'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222223', id, 200, false
FROM router_roles WHERE code = 'nat_gateway'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222223', id, 300, false
FROM router_roles WHERE code = 'firewall'
ON CONFLICT (router_id, role_id) DO NOTHING;

-- PPPoE Server (multi-role: pppoe_server + dhcp_server)
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222224', id, 100, true
FROM router_roles WHERE code = 'pppoe_server'
ON CONFLICT (router_id, role_id) DO NOTHING;

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '22222222-2222-2222-2222-222222222224', id, 200, false
FROM router_roles WHERE code = 'dhcp_server'
ON CONFLICT (router_id, role_id) DO NOTHING;

-- ============================================================================
-- 5. DEFINE ROUTER DEPENDENCIES
-- ============================================================================

-- Core router is upstream of edge router
INSERT INTO router_dependencies (
    tenant_id, source_router_id, target_router_id, dependency_type, weight, is_active, description
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',  -- edge router
    '22222222-2222-2222-2222-222222222221',  -- core router
    'upstream',
    100,
    true,
    'Edge router routes traffic through core'
) ON CONFLICT (source_router_id, target_router_id, dependency_type) DO NOTHING;

-- Edge router is upstream of access router
INSERT INTO router_dependencies (
    tenant_id, source_router_id, target_router_id, dependency_type, weight, is_active, description
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222223',  -- access router
    '22222222-2222-2222-2222-222222222222',  -- edge router
    'upstream',
    100,
    true,
    'Access router connects through edge router'
) ON CONFLICT (source_router_id, target_router_id, dependency_type) DO NOTHING;

-- Access router is upstream of PPPoE server
INSERT INTO router_dependencies (
    tenant_id, source_router_id, target_router_id, dependency_type, weight, is_active, description
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222224',  -- pppoe server
    '22222222-2222-2222-2222-222222222223',  -- access router
    'upstream',
    100,
    true,
    'PPPoE traffic routes through access router'
) ON CONFLICT (source_router_id, target_router_id, dependency_type) DO NOTHING;

-- ============================================================================
-- 6. SAMPLE PPPoE SESSIONS (for demonstration)
-- ============================================================================

INSERT INTO pppoe_sessions (
    tenant_id, router_id, session_id, username, calling_station_id,
    framed_ip_address, status, connect_time, bytes_in, bytes_out
) VALUES 
    (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222224',
        'pppoe-001',
        'user001@example.com',
        '00:11:22:33:44:55',
        '10.10.10.10',
        'active',
        NOW() - INTERVAL '2 hours',
        1024000000,
        512000000
    ),
    (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222224',
        'pppoe-002',
        'user002@example.com',
        '00:11:22:33:44:66',
        '10.10.10.11',
        'active',
        NOW() - INTERVAL '1 hour',
        512000000,
        256000000
    ),
    (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222224',
        'pppoe-003',
        'user003@example.com',
        '00:11:22:33:44:77',
        '10.10.10.12',
        'active',
        NOW() - INTERVAL '30 minutes',
        256000000,
        128000000
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 7. QUERY TO VERIFY SETUP
-- ============================================================================

-- Check routers with roles
SELECT 
    r.name,
    r.management_ip,
    r.vendor,
    array_agg(rr.code ORDER BY rra.priority) as roles,
    rc.preferred_method as polling_method
FROM routers r
LEFT JOIN router_role_assignments rra ON r.id = rra.router_id
LEFT JOIN router_roles rr ON rra.role_id = rr.id
LEFT JOIN router_capabilities rc ON r.id = rc.router_id
WHERE r.tenant_id = '11111111-1111-1111-1111-111111111111'
GROUP BY r.id, r.name, r.management_ip, r.vendor, rc.preferred_method
ORDER BY r.name;

-- Check router dependencies
SELECT 
    r1.name as from_router,
    rd.dependency_type,
    r2.name as to_router,
    rd.description
FROM router_dependencies rd
JOIN routers r1 ON rd.source_router_id = r1.id
JOIN routers r2 ON rd.target_router_id = r2.id
WHERE rd.tenant_id = '11111111-1111-1111-1111-111111111111'
ORDER BY r1.name, rd.dependency_type;

-- Check active PPPoE sessions
SELECT 
    r.name as router,
    ps.username,
    ps.framed_ip_address,
    ps.connect_time,
    ps.status,
    (ps.bytes_in + ps.bytes_out) / 1048576 as total_mb
FROM pppoe_sessions ps
JOIN routers r ON ps.router_id = r.id
WHERE ps.tenant_id = '11111111-1111-1111-1111-111111111111'
  AND ps.status = 'active'
ORDER BY ps.connect_time DESC;

-- ============================================================================
-- NOTES
-- ============================================================================

-- To use this test data with the CHR lab:
-- 1. Start the CHR lab: docker-compose -f docker-compose.chr.yml up -d
-- 2. Configure CHR instances with: scripts/chr/setup-chr-monitoring.rsc
-- 3. Apply this SQL: psql -U ispmonitor -d ispmonitor -f db/examples/test_data.sql
-- 4. Start the monitor: The poller will automatically discover and poll these routers
-- 5. Check polling history: SELECT * FROM polling_history ORDER BY poll_started_at DESC;
