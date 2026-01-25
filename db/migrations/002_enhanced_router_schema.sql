-- ISP Visual Monitor - Enhanced Router Schema Migration
-- This migration adds support for:
-- 1. Multi-role routers (routers can have multiple roles)
-- 2. Router capabilities (SNMP, API, SSH, etc.)
-- 3. Router dependencies (upstream/downstream relationships)
-- 4. Session tracking (PPPoE, NAT, etc.)

-- ============================================================================
-- ROUTER ROLES
-- ============================================================================

-- Router Roles table - defines standard router role types
CREATE TABLE router_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE, -- e.g., core_router, pppoe_server
    description TEXT,
    category VARCHAR(50), -- routing, access, security, management
    icon VARCHAR(100), -- Icon name for UI
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_router_roles_code ON router_roles(code);
CREATE INDEX idx_router_roles_category ON router_roles(category);

-- Insert standard router roles
INSERT INTO router_roles (name, code, description, category) VALUES
    ('Core Router', 'core_router', 'Core backbone router handling inter-POP traffic', 'routing'),
    ('Edge Router', 'edge_router', 'Edge router connecting to external networks/providers', 'routing'),
    ('Border Router', 'border_router', 'Border router for external BGP peering', 'routing'),
    ('Access Router', 'access_router', 'Access router for customer connections', 'access'),
    ('NAT Gateway', 'nat_gateway', 'Network Address Translation gateway', 'security'),
    ('PPPoE Server', 'pppoe_server', 'PPPoE authentication and session server', 'access'),
    ('DHCP Server', 'dhcp_server', 'DHCP address assignment server', 'access'),
    ('VPN Gateway', 'vpn_gateway', 'VPN concentrator and gateway', 'security'),
    ('Load Balancer', 'load_balancer', 'Traffic load balancing device', 'routing'),
    ('Bandwidth Shaper', 'bandwidth_shaper', 'QoS and traffic shaping device', 'management'),
    ('Access Controller', 'access_controller', 'Access control and authentication', 'security'),
    ('Firewall', 'firewall', 'Firewall and security gateway', 'security');

-- Router Role Assignments - Many-to-many relationship between routers and roles
CREATE TABLE router_role_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES router_roles(id) ON DELETE CASCADE,
    priority INTEGER DEFAULT 100, -- Lower = higher priority, for display ordering
    is_primary BOOLEAN DEFAULT false, -- Mark one role as primary
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID REFERENCES users(id),
    notes TEXT,
    CONSTRAINT unique_router_role UNIQUE(router_id, role_id)
);

CREATE INDEX idx_router_role_assignments_router ON router_role_assignments(router_id);
CREATE INDEX idx_router_role_assignments_role ON router_role_assignments(role_id);
CREATE INDEX idx_router_role_assignments_priority ON router_role_assignments(router_id, priority);

-- ============================================================================
-- ROUTER DEPENDENCIES
-- ============================================================================

-- Router Dependencies - Model relationships between routers
CREATE TABLE router_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    target_router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    dependency_type VARCHAR(50) NOT NULL, -- upstream, downstream, peer, failover, backup
    weight INTEGER DEFAULT 100, -- For load balancing or priority
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_router_dependency UNIQUE(source_router_id, target_router_id, dependency_type),
    CONSTRAINT no_self_dependency CHECK (source_router_id != target_router_id)
);

CREATE INDEX idx_router_dependencies_tenant ON router_dependencies(tenant_id);
CREATE INDEX idx_router_dependencies_source ON router_dependencies(source_router_id);
CREATE INDEX idx_router_dependencies_target ON router_dependencies(target_router_id);
CREATE INDEX idx_router_dependencies_type ON router_dependencies(dependency_type);

-- ============================================================================
-- ROUTER CAPABILITIES
-- ============================================================================

-- Router Capabilities - Store connection methods and credentials
CREATE TABLE router_capabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    router_id UUID NOT NULL UNIQUE REFERENCES routers(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- SNMP Capability
    snmp_enabled BOOLEAN DEFAULT false,
    snmp_version VARCHAR(10), -- v1, v2c, v3
    snmp_community VARCHAR(255), -- For v1/v2c (encrypted)
    snmp_port INTEGER DEFAULT 161,
    snmp_v3_username VARCHAR(255),
    snmp_v3_auth_protocol VARCHAR(50), -- MD5, SHA, SHA-224, SHA-256, SHA-384, SHA-512
    snmp_v3_auth_password VARCHAR(255), -- (encrypted)
    snmp_v3_priv_protocol VARCHAR(50), -- DES, AES, AES-192, AES-256
    snmp_v3_priv_password VARCHAR(255), -- (encrypted)
    snmp_timeout_seconds INTEGER DEFAULT 10,
    snmp_retries INTEGER DEFAULT 3,
    
    -- API Capability (Vendor-specific APIs)
    api_enabled BOOLEAN DEFAULT false,
    api_type VARCHAR(50), -- mikrotik, cisco_restconf, juniper_netconf, arista_eapi
    api_endpoint VARCHAR(255), -- URL or connection string
    api_port INTEGER,
    api_username VARCHAR(255),
    api_password VARCHAR(255), -- (encrypted)
    api_use_tls BOOLEAN DEFAULT true,
    api_verify_cert BOOLEAN DEFAULT true,
    api_timeout_seconds INTEGER DEFAULT 30,
    
    -- SSH Capability
    ssh_enabled BOOLEAN DEFAULT false,
    ssh_host VARCHAR(255),
    ssh_port INTEGER DEFAULT 22,
    ssh_username VARCHAR(255),
    ssh_password VARCHAR(255), -- (encrypted)
    ssh_private_key TEXT, -- (encrypted)
    ssh_timeout_seconds INTEGER DEFAULT 30,
    
    -- NETCONF Capability
    netconf_enabled BOOLEAN DEFAULT false,
    netconf_port INTEGER DEFAULT 830,
    netconf_username VARCHAR(255),
    netconf_password VARCHAR(255), -- (encrypted)
    
    -- Syslog Capability (passive monitoring)
    syslog_enabled BOOLEAN DEFAULT false,
    syslog_facility VARCHAR(50),
    syslog_severity VARCHAR(50),
    
    -- NetFlow/sFlow Capability (passive monitoring)
    netflow_enabled BOOLEAN DEFAULT false,
    netflow_version INTEGER, -- 5, 9, 10 (IPFIX)
    netflow_port INTEGER DEFAULT 2055,
    
    -- IPFIX Capability
    ipfix_enabled BOOLEAN DEFAULT false,
    ipfix_port INTEGER DEFAULT 4739,
    
    -- Metadata
    preferred_method VARCHAR(50), -- api, snmp, ssh, netconf
    fallback_order TEXT[], -- Array of method names in fallback order
    last_tested_at TIMESTAMP,
    last_test_success BOOLEAN,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_router_capabilities_tenant ON router_capabilities(tenant_id);
CREATE INDEX idx_router_capabilities_router ON router_capabilities(router_id);

-- ============================================================================
-- SESSION TRACKING (PPPoE, NAT, etc.)
-- ============================================================================

-- PPPoE Sessions
CREATE TABLE pppoe_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    session_id VARCHAR(255), -- Router-specific session ID
    username VARCHAR(255) NOT NULL,
    calling_station_id VARCHAR(100), -- MAC address or phone number
    framed_ip_address INET,
    nas_ip_address INET,
    nas_port VARCHAR(100),
    service_type VARCHAR(100),
    session_time_seconds BIGINT,
    idle_time_seconds BIGINT,
    bytes_in BIGINT DEFAULT 0,
    bytes_out BIGINT DEFAULT 0,
    packets_in BIGINT DEFAULT 0,
    packets_out BIGINT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active', -- active, disconnected, idle
    connect_time TIMESTAMP,
    disconnect_time TIMESTAMP,
    disconnect_cause VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pppoe_sessions_tenant ON pppoe_sessions(tenant_id);
CREATE INDEX idx_pppoe_sessions_router ON pppoe_sessions(router_id);
CREATE INDEX idx_pppoe_sessions_username ON pppoe_sessions(username);
CREATE INDEX idx_pppoe_sessions_status ON pppoe_sessions(status);
CREATE INDEX idx_pppoe_sessions_connect_time ON pppoe_sessions(connect_time DESC);
CREATE INDEX idx_pppoe_sessions_framed_ip ON pppoe_sessions(framed_ip_address);

-- NAT Sessions
CREATE TABLE nat_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    protocol VARCHAR(20) NOT NULL, -- tcp, udp, icmp
    src_address INET NOT NULL,
    src_port INTEGER,
    dst_address INET NOT NULL,
    dst_port INTEGER,
    translated_src_address INET,
    translated_src_port INTEGER,
    state VARCHAR(50), -- established, syn_sent, syn_recv, fin_wait, etc.
    bytes BIGINT DEFAULT 0,
    packets BIGINT DEFAULT 0,
    timeout_seconds INTEGER,
    established_at TIMESTAMP,
    last_seen_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_nat_sessions_tenant ON nat_sessions(tenant_id);
CREATE INDEX idx_nat_sessions_router ON nat_sessions(router_id);
CREATE INDEX idx_nat_sessions_src_address ON nat_sessions(src_address);
CREATE INDEX idx_nat_sessions_dst_address ON nat_sessions(dst_address);
CREATE INDEX idx_nat_sessions_status ON nat_sessions(status);
CREATE INDEX idx_nat_sessions_last_seen ON nat_sessions(last_seen_at DESC);

-- DHCP Leases
CREATE TABLE dhcp_leases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    mac_address MACADDR NOT NULL,
    ip_address INET NOT NULL,
    hostname VARCHAR(255),
    lease_start TIMESTAMP NOT NULL,
    lease_end TIMESTAMP NOT NULL,
    lease_state VARCHAR(50), -- active, expired, released, offered
    dhcp_pool VARCHAR(100),
    client_id VARCHAR(255),
    vendor_class VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dhcp_leases_tenant ON dhcp_leases(tenant_id);
CREATE INDEX idx_dhcp_leases_router ON dhcp_leases(router_id);
CREATE INDEX idx_dhcp_leases_mac ON dhcp_leases(mac_address);
CREATE INDEX idx_dhcp_leases_ip ON dhcp_leases(ip_address);
CREATE INDEX idx_dhcp_leases_state ON dhcp_leases(lease_state);

-- ============================================================================
-- ENHANCED METRICS TABLES
-- ============================================================================

-- Role-specific metrics
CREATE TABLE role_specific_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    role_code VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metrics JSONB NOT NULL, -- Flexible JSON storage for role-specific data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_role_metrics_tenant_time ON role_specific_metrics(tenant_id, timestamp DESC);
CREATE INDEX idx_role_metrics_router_time ON role_specific_metrics(router_id, timestamp DESC);
CREATE INDEX idx_role_metrics_role ON role_specific_metrics(role_code);
CREATE INDEX idx_role_metrics_jsonb ON role_specific_metrics USING gin(metrics);

-- ============================================================================
-- POLLING HISTORY
-- ============================================================================

-- Polling History - Track polling attempts and results
CREATE TABLE polling_history (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    poll_started_at TIMESTAMP NOT NULL,
    poll_completed_at TIMESTAMP,
    adapter_used VARCHAR(100), -- snmp, mikrotik_api, ssh, etc.
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metrics_collected INTEGER DEFAULT 0,
    response_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_polling_history_tenant ON polling_history(tenant_id);
CREATE INDEX idx_polling_history_router ON polling_history(router_id);
CREATE INDEX idx_polling_history_started ON polling_history(poll_started_at DESC);
CREATE INDEX idx_polling_history_success ON polling_history(success);

-- ============================================================================
-- MIGRATION HELPER: Auto-assign roles based on router_type
-- ============================================================================

-- Migrate existing routers to have role assignments based on router_type
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary, assigned_by, notes)
SELECT 
    r.id,
    rr.id,
    100,
    true,
    NULL,
    'Auto-migrated from router_type'
FROM routers r
JOIN router_roles rr ON (
    (r.router_type = 'core' AND rr.code = 'core_router') OR
    (r.router_type = 'edge' AND rr.code = 'edge_router') OR
    (r.router_type = 'border' AND rr.code = 'border_router') OR
    (r.router_type = 'access' AND rr.code = 'access_router')
)
WHERE NOT EXISTS (
    SELECT 1 FROM router_role_assignments rra 
    WHERE rra.router_id = r.id
);

-- Migrate existing SNMP configuration to router_capabilities
INSERT INTO router_capabilities (
    router_id, 
    tenant_id, 
    snmp_enabled, 
    snmp_version, 
    snmp_community, 
    snmp_port,
    preferred_method,
    fallback_order
)
SELECT 
    r.id,
    r.tenant_id,
    r.polling_enabled,
    r.snmp_version,
    r.snmp_community,
    r.snmp_port,
    'snmp',
    ARRAY['snmp']
FROM routers r
WHERE NOT EXISTS (
    SELECT 1 FROM router_capabilities rc 
    WHERE rc.router_id = r.id
);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

CREATE TRIGGER update_router_capabilities_updated_at BEFORE UPDATE ON router_capabilities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_router_dependencies_updated_at BEFORE UPDATE ON router_dependencies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pppoe_sessions_updated_at BEFORE UPDATE ON pppoe_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dhcp_leases_updated_at BEFORE UPDATE ON dhcp_leases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- VIEWS FOR CONVENIENCE
-- ============================================================================

-- View: Routers with their roles
CREATE VIEW routers_with_roles AS
SELECT 
    r.id,
    r.tenant_id,
    r.name,
    r.management_ip,
    r.vendor,
    r.status,
    array_agg(rr.code ORDER BY rra.priority) as role_codes,
    array_agg(rr.name ORDER BY rra.priority) as role_names,
    (SELECT rr2.code FROM router_role_assignments rra2 
     JOIN router_roles rr2 ON rra2.role_id = rr2.id 
     WHERE rra2.router_id = r.id AND rra2.is_primary = true 
     LIMIT 1) as primary_role_code
FROM routers r
LEFT JOIN router_role_assignments rra ON r.id = rra.router_id
LEFT JOIN router_roles rr ON rra.role_id = rr.id
GROUP BY r.id;

-- View: Active PPPoE sessions summary
CREATE VIEW pppoe_sessions_summary AS
SELECT 
    router_id,
    COUNT(*) as total_sessions,
    COUNT(*) FILTER (WHERE status = 'active') as active_sessions,
    SUM(bytes_in) as total_bytes_in,
    SUM(bytes_out) as total_bytes_out,
    AVG(session_time_seconds) as avg_session_time
FROM pppoe_sessions
WHERE status = 'active'
GROUP BY router_id;

-- View: NAT sessions summary
CREATE VIEW nat_sessions_summary AS
SELECT 
    router_id,
    protocol,
    COUNT(*) as session_count,
    SUM(bytes) as total_bytes,
    SUM(packets) as total_packets
FROM nat_sessions
WHERE status = 'active'
GROUP BY router_id, protocol;

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE router_roles IS 'Standard router role types (core, edge, PPPoE server, NAT gateway, etc.)';
COMMENT ON TABLE router_role_assignments IS 'Many-to-many mapping allowing routers to have multiple roles';
COMMENT ON TABLE router_dependencies IS 'Models upstream/downstream/peer relationships between routers';
COMMENT ON TABLE router_capabilities IS 'Connection methods and credentials for each router (SNMP, API, SSH, etc.)';
COMMENT ON TABLE pppoe_sessions IS 'Active and historical PPPoE sessions';
COMMENT ON TABLE nat_sessions IS 'Active NAT translation sessions';
COMMENT ON TABLE dhcp_leases IS 'DHCP lease assignments';
COMMENT ON TABLE role_specific_metrics IS 'Flexible storage for role-specific metrics (e.g., PPPoE stats, NAT counters)';
COMMENT ON TABLE polling_history IS 'History of polling attempts with success/failure tracking';
