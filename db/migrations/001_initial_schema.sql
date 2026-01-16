-- ISP Visual Monitor - Database Schema
-- PostgreSQL 15+ with PostGIS extension

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For full-text search

-- ============================================================================
-- TENANT MANAGEMENT
-- ============================================================================

-- Tenants table (ISP organizations)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    contact_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50),
    subscription_tier VARCHAR(50) DEFAULT 'basic', -- basic, professional, enterprise
    max_devices INTEGER DEFAULT 100,
    max_users INTEGER DEFAULT 10,
    status VARCHAR(50) DEFAULT 'active', -- active, suspended, trial
    trial_ends_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_slug ON tenants(slug);

-- ============================================================================
-- USER MANAGEMENT & RBAC
-- ============================================================================

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, suspended
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_email_per_tenant UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false, -- System roles cannot be modified
    is_custom BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_role_per_tenant UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);

-- Permissions table (global, not tenant-scoped)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_permissions_category ON permissions(category);

-- Role-Permission mapping
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role assignment
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- API Keys for programmatic access
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL, -- For display purposes
    permissions TEXT[], -- Array of permission names
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- ============================================================================
-- NETWORK TOPOLOGY
-- ============================================================================

-- Geographic Regions
CREATE TABLE regions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    parent_region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    boundary GEOGRAPHY(POLYGON),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_region_per_tenant UNIQUE(tenant_id, name)
);

CREATE INDEX idx_regions_tenant_id ON regions(tenant_id);
CREATE INDEX idx_regions_parent ON regions(parent_region_id);
CREATE INDEX idx_regions_boundary ON regions USING GIST(boundary);

-- Points of Presence (POPs)
CREATE TABLE pops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50), -- Short code (e.g., NYC-DC1)
    location GEOGRAPHY(POINT) NOT NULL,
    address TEXT,
    pop_type VARCHAR(50), -- datacenter, co-location, edge, customer-site
    capacity_gbps INTEGER,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_pop_per_tenant UNIQUE(tenant_id, code)
);

CREATE INDEX idx_pops_tenant_id ON pops(tenant_id);
CREATE INDEX idx_pops_region_id ON pops(region_id);
CREATE INDEX idx_pops_location ON pops USING GIST(location);
CREATE INDEX idx_pops_status ON pops(status);

-- Routers
CREATE TABLE routers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    pop_id UUID REFERENCES pops(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    management_ip INET NOT NULL,
    location GEOGRAPHY(POINT),
    router_type VARCHAR(50), -- core, edge, border, access
    vendor VARCHAR(100), -- cisco, juniper, arista, etc.
    model VARCHAR(100),
    os_version VARCHAR(100),
    serial_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, maintenance, down
    polling_enabled BOOLEAN DEFAULT true,
    polling_interval_seconds INTEGER DEFAULT 300,
    snmp_version VARCHAR(10) DEFAULT 'v2c',
    snmp_community VARCHAR(255),
    snmp_port INTEGER DEFAULT 161,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_polled_at TIMESTAMP,
    CONSTRAINT unique_router_name UNIQUE(tenant_id, name),
    CONSTRAINT unique_management_ip UNIQUE(tenant_id, management_ip)
);

CREATE INDEX idx_routers_tenant_id ON routers(tenant_id);
CREATE INDEX idx_routers_pop_id ON routers(pop_id);
CREATE INDEX idx_routers_location ON routers USING GIST(location);
CREATE INDEX idx_routers_status ON routers(status);
CREATE INDEX idx_routers_polling ON routers(polling_enabled, last_polled_at);
CREATE INDEX idx_routers_search ON routers USING gin(name gin_trgm_ops);

-- Router Interfaces
CREATE TABLE interfaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    if_index INTEGER,
    if_type VARCHAR(50), -- ethernet, optical, tunnel, etc.
    speed_mbps BIGINT,
    mtu INTEGER,
    mac_address MACADDR,
    ip_address INET,
    subnet_mask INET,
    status VARCHAR(50) DEFAULT 'up', -- up, down, admin-down, testing
    admin_status VARCHAR(50) DEFAULT 'up',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_interface UNIQUE(router_id, name)
);

CREATE INDEX idx_interfaces_tenant_id ON interfaces(tenant_id);
CREATE INDEX idx_interfaces_router_id ON interfaces(router_id);
CREATE INDEX idx_interfaces_status ON interfaces(status);
CREATE INDEX idx_interfaces_ip ON interfaces(ip_address);

-- Network Links (connections between interfaces)
CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255),
    source_interface_id UUID NOT NULL REFERENCES interfaces(id) ON DELETE CASCADE,
    target_interface_id UUID NOT NULL REFERENCES interfaces(id) ON DELETE CASCADE,
    link_type VARCHAR(50), -- physical, logical, vpn, tunnel
    capacity_mbps BIGINT,
    latency_ms DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'up',
    path_geometry GEOGRAPHY(LINESTRING), -- Visual path on map
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_link UNIQUE(source_interface_id, target_interface_id),
    CONSTRAINT no_self_link CHECK (source_interface_id != target_interface_id)
);

CREATE INDEX idx_links_tenant_id ON links(tenant_id);
CREATE INDEX idx_links_source ON links(source_interface_id);
CREATE INDEX idx_links_target ON links(target_interface_id);
CREATE INDEX idx_links_status ON links(status);
CREATE INDEX idx_links_geometry ON links USING GIST(path_geometry);

-- ============================================================================
-- MONITORING DATA (Time-Series)
-- ============================================================================

-- Interface Metrics (time-series data)
CREATE TABLE interface_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    interface_id UUID NOT NULL REFERENCES interfaces(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    in_octets BIGINT,
    out_octets BIGINT,
    in_packets BIGINT,
    out_packets BIGINT,
    in_errors BIGINT,
    out_errors BIGINT,
    in_discards BIGINT,
    out_discards BIGINT,
    utilization_percent DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Partition by month for better performance
CREATE INDEX idx_interface_metrics_interface_time ON interface_metrics(interface_id, timestamp DESC);
CREATE INDEX idx_interface_metrics_tenant_time ON interface_metrics(tenant_id, timestamp DESC);

-- Router Metrics
CREATE TABLE router_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cpu_percent DECIMAL(5,2),
    memory_percent DECIMAL(5,2),
    uptime_seconds BIGINT,
    temperature_celsius DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_router_metrics_router_time ON router_metrics(router_id, timestamp DESC);
CREATE INDEX idx_router_metrics_tenant_time ON router_metrics(tenant_id, timestamp DESC);

-- ============================================================================
-- ALERTS & NOTIFICATIONS
-- ============================================================================

-- Alert Rules
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL, -- interface_utilization, device_down, link_down, etc.
    target_type VARCHAR(50), -- router, interface, link, pop
    target_id UUID, -- References various tables
    condition JSONB NOT NULL, -- Flexible condition definition
    threshold_value DECIMAL(10,2),
    severity VARCHAR(50) DEFAULT 'warning', -- critical, warning, info
    enabled BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_rules_tenant_id ON alert_rules(tenant_id);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX idx_alert_rules_target ON alert_rules(target_type, target_id);

-- Alerts (generated from rules)
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(50) DEFAULT 'warning',
    status VARCHAR(50) DEFAULT 'active', -- active, acknowledged, resolved
    target_type VARCHAR(50),
    target_id UUID,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP,
    acknowledged_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP,
    metadata JSONB
);

CREATE INDEX idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_triggered ON alerts(triggered_at DESC);
CREATE INDEX idx_alerts_target ON alerts(target_type, target_id);

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- ============================================================================
-- ROW LEVEL SECURITY
-- ============================================================================

-- Enable RLS on tenant-scoped tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE routers ENABLE ROW LEVEL SECURITY;
ALTER TABLE interfaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE links ENABLE ROW LEVEL SECURITY;
ALTER TABLE pops ENABLE ROW LEVEL SECURITY;
ALTER TABLE regions ENABLE ROW LEVEL SECURITY;
ALTER TABLE interface_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE router_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;

-- Create RLS policies (example for routers)
CREATE POLICY tenant_isolation_policy ON routers
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID);

-- ============================================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================================

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update trigger to relevant tables
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_routers_updated_at BEFORE UPDATE ON routers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_interfaces_updated_at BEFORE UPDATE ON interfaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_links_updated_at BEFORE UPDATE ON links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SEED DATA (Default Permissions)
-- ============================================================================

INSERT INTO permissions (name, category, description) VALUES
    -- Network Management
    ('network.devices.read', 'network', 'View network devices'),
    ('network.devices.create', 'network', 'Add new network devices'),
    ('network.devices.update', 'network', 'Modify network device configuration'),
    ('network.devices.delete', 'network', 'Remove network devices'),
    ('network.topology.read', 'network', 'View network topology'),
    ('network.topology.update', 'network', 'Modify network topology'),
    
    -- Monitoring
    ('monitoring.metrics.read', 'monitoring', 'View metrics and graphs'),
    ('monitoring.alerts.read', 'monitoring', 'View alerts'),
    ('monitoring.alerts.acknowledge', 'monitoring', 'Acknowledge alerts'),
    ('monitoring.alerts.configure', 'monitoring', 'Create and modify alert rules'),
    
    -- User Management
    ('users.read', 'users', 'View user list'),
    ('users.create', 'users', 'Create new users'),
    ('users.update', 'users', 'Modify user details'),
    ('users.delete', 'users', 'Remove users'),
    ('users.roles.assign', 'users', 'Assign roles to users'),
    
    -- Configuration
    ('config.tenant.read', 'config', 'View tenant settings'),
    ('config.tenant.update', 'config', 'Modify tenant settings'),
    ('config.polling.update', 'config', 'Configure polling parameters'),
    ('config.integration.manage', 'config', 'Manage webhooks and integrations'),
    
    -- Data
    ('data.export', 'data', 'Export data and reports'),
    ('data.import', 'data', 'Import device configurations');
