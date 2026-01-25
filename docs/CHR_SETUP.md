# MikroTik CHR Lab Setup Guide

## Overview

This guide explains how to set up a MikroTik Cloud Hosted Router (CHR) lab environment for testing and developing the ISP Visual Monitor system. The lab includes multiple CHR instances with different roles (core, edge, access, PPPoE server) to simulate a realistic ISP network topology.

## Prerequisites

- Docker and Docker Compose installed
- At least 4GB of available RAM
- Basic understanding of MikroTik RouterOS
- (Optional) MikroTik RouterOS CLI tool for advanced testing

## Architecture

The CHR lab consists of:

1. **chr-core-01**: Core backbone router
   - Role: `core_router`
   - IP: 172.25.0.10
   - API Port: 8728
   - SNMP Port: 161

2. **chr-edge-01**: Edge router for external connectivity
   - Role: `edge_router`
   - IP: 172.25.0.11
   - API Port: 8738
   - SNMP Port: 162

3. **chr-access-01**: Access router with NAT
   - Roles: `access_router`, `nat_gateway`
   - IP: 172.25.0.12
   - API Port: 8748
   - SNMP Port: 163

4. **chr-pppoe-01**: PPPoE and DHCP server
   - Roles: `pppoe_server`, `dhcp_server`
   - IP: 172.25.0.13
   - API Port: 8758
   - SNMP Port: 164

## Quick Start

### 1. Start the CHR Lab

```bash
# From the repository root
docker-compose -f docker-compose.chr.yml up -d

# Check status
docker-compose -f docker-compose.chr.yml ps

# View logs
docker-compose -f docker-compose.chr.yml logs -f
```

### 2. Wait for CHR Instances to Boot

CHR instances take 30-60 seconds to fully boot. Wait until you see "Login:" in the logs.

```bash
# Watch logs for chr-pppoe-01
docker-compose -f docker-compose.chr.yml logs -f chr-pppoe-01
```

### 3. Configure Each CHR Instance

Access each CHR instance and run the configuration script:

```bash
# Option 1: Via docker exec
docker exec -it isp-chr-pppoe-01 /bin/sh

# Option 2: Via SSH (after initial setup)
ssh admin@localhost -p 2224
```

Inside the CHR instance, download and run the configuration script:

```bash
# Copy the configuration script content
# Then paste it into the RouterOS CLI or
# Import it via /import command
```

Alternatively, copy the script to the container and import it:

```bash
# Copy script to container
docker cp scripts/chr/setup-chr-monitoring.rsc isp-chr-pppoe-01:/setup.rsc

# Import via docker exec
docker exec -it isp-chr-pppoe-01 ros script run setup.rsc
```

### 4. Verify Configuration

Check that monitoring capabilities are enabled:

```bash
# Test SNMP (replace with actual router IP)
snmpwalk -v 2c -c public 172.25.0.13 system

# Test API access with curl
curl -u monitor:monitor123 http://localhost:8758/rest/system/resource

# Test via MikroTik CLI
ros -user=monitor -pass=monitor123 -host=localhost:8758 /system/resource/print
```

## Connecting ISP Monitor to CHR Lab

### 1. Apply Database Migrations

```bash
# Start PostgreSQL if not running
docker-compose -f docker-compose.chr.yml up -d postgres

# Apply migrations
docker-compose -f docker-compose.chr.yml exec postgres \
  psql -U ispmonitor -d ispmonitor -f /docker-entrypoint-initdb.d/001_initial_schema.sql

docker-compose -f docker-compose.chr.yml exec postgres \
  psql -U ispmonitor -d ispmonitor -f /docker-entrypoint-initdb.d/002_enhanced_router_schema.sql
```

### 2. Insert Test Router Configurations

Create a SQL file with router configurations:

```sql
-- Insert tenant
INSERT INTO tenants (name, slug, contact_email, subscription_tier, max_devices, status)
VALUES ('Test ISP', 'test-isp', 'admin@test-isp.local', 'enterprise', 100, 'active')
RETURNING id;

-- Use the tenant ID from above in subsequent queries
-- Replace <tenant_id> with actual UUID

-- Insert PPPoE Server Router
INSERT INTO routers (tenant_id, name, management_ip, vendor, status, polling_enabled, polling_interval_seconds)
VALUES ('<tenant_id>', 'chr-pppoe-01', '172.25.0.13', 'MikroTik', 'active', true, 60);

-- Get router ID and insert capabilities
INSERT INTO router_capabilities (
  router_id, tenant_id,
  api_enabled, api_type, api_endpoint, api_port, api_username, api_password, api_use_tls,
  snmp_enabled, snmp_version, snmp_community, snmp_port,
  preferred_method, fallback_order
) VALUES (
  '<router_id>', '<tenant_id>',
  true, 'mikrotik', '172.25.0.13', 8758, 'monitor', 'monitor123', false,
  true, 'v2c', 'public', 164,
  'api', ARRAY['api', 'snmp']
);

-- Assign roles
INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '<router_id>', id, 100, true FROM router_roles WHERE code = 'pppoe_server';

INSERT INTO router_role_assignments (router_id, role_id, priority, is_primary)
SELECT '<router_id>', id, 200, false FROM router_roles WHERE code = 'dhcp_server';
```

Execute the SQL:

```bash
docker-compose -f docker-compose.chr.yml exec postgres \
  psql -U ispmonitor -d ispmonitor -f /path/to/test_routers.sql
```

### 3. Start ISP Monitor

```bash
# Build and start the monitor
docker-compose -f docker-compose.chr.yml up -d isp-monitor-dev

# Check logs
docker-compose -f docker-compose.chr.yml logs -f isp-monitor-dev
```

## Testing Scenarios

### 1. Test Basic Polling

The poller should automatically discover and poll configured routers every 60 seconds.

Watch the poller logs:

```bash
docker-compose -f docker-compose.chr.yml logs -f isp-monitor-dev | grep -i poll
```

### 2. Simulate PPPoE Activity

Use the provided simulation script to generate realistic PPPoE connection patterns:

```bash
# From repository root
./scripts/chr/simulate-pppoe-activity.sh
```

Select from the menu:
1. Random activity - Simulates users connecting/disconnecting randomly
2. Steady load - Maintains constant number of connections
3. Peak hour - Simulates traffic ramp-up and ramp-down

### 3. Query Active Sessions

Check that PPPoE sessions are being tracked:

```bash
# Via database
docker-compose -f docker-compose.chr.yml exec postgres \
  psql -U ispmonitor -d ispmonitor -c "SELECT * FROM pppoe_sessions WHERE status = 'active';"

# Via API (once implemented)
curl http://localhost:8080/api/v1/routers/<router_id>/pppoe-sessions
```

### 4. Test Adapter Fallback

Disable API and verify SNMP fallback:

```bash
# Disable API on CHR
docker exec -it isp-chr-pppoe-01 \
  ros /ip/service/disable api

# Watch logs to see adapter fallback
docker-compose -f docker-compose.chr.yml logs -f isp-monitor-dev
```

### 5. Test Multi-Role Behavior

Add NAT gateway role to PPPoE server:

```sql
INSERT INTO router_role_assignments (router_id, role_id, priority)
SELECT '<router_id>', id, 300 FROM router_roles WHERE code = 'nat_gateway';
```

Verify that both PPPoE and NAT metrics are collected.

## Advanced Configuration

### Enable Syslog Forwarding

Configure CHR to forward logs to ISP Monitor:

```routeros
/system logging action add name=remote remote=172.25.0.1 remote-port=514 target=remote
/system logging add action=remote topics=system,info
/system logging add action=remote topics=pppoe,info
```

### Enable NetFlow Export

Configure NetFlow v9 export:

```routeros
/ip traffic-flow set enabled=yes interfaces=all
/ip traffic-flow target add address=172.25.0.1:2055 version=9
```

### Create Router Dependencies

Model router relationships:

```sql
-- chr-edge-01 is upstream of chr-access-01
INSERT INTO router_dependencies (tenant_id, source_router_id, target_router_id, dependency_type, weight)
VALUES ('<tenant_id>', '<access_router_id>', '<edge_router_id>', 'upstream', 100);

-- chr-pppoe-01 is downstream of chr-access-01
INSERT INTO router_dependencies (tenant_id, source_router_id, target_router_id, dependency_type, weight)
VALUES ('<tenant_id>', '<access_router_id>', '<pppoe_router_id>', 'downstream', 100);
```

## Troubleshooting

### CHR Won't Boot

```bash
# Check logs
docker-compose -f docker-compose.chr.yml logs chr-pppoe-01

# Restart container
docker-compose -f docker-compose.chr.yml restart chr-pppoe-01
```

### API Connection Fails

```bash
# Verify API is enabled
docker exec -it isp-chr-pppoe-01 ros /ip/service/print

# Check firewall rules
docker exec -it isp-chr-pppoe-01 ros /ip/firewall/filter/print

# Test from monitor container
docker-compose -f docker-compose.chr.yml exec isp-monitor-dev \
  curl -v -u monitor:monitor123 http://172.25.0.13:8758/rest/system/resource
```

### SNMP Not Working

```bash
# Verify SNMP is enabled
docker exec -it isp-chr-pppoe-01 ros /snmp/print

# Test SNMP from host
snmpwalk -v 2c -c public localhost:164 system

# Check community settings
docker exec -it isp-chr-pppoe-01 ros /snmp/community/print
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose -f docker-compose.chr.yml ps postgres

# Test connection
docker-compose -f docker-compose.chr.yml exec postgres \
  psql -U ispmonitor -d ispmonitor -c "SELECT 1;"

# Check monitor can connect
docker-compose -f docker-compose.chr.yml exec isp-monitor-dev \
  pg_isready -h postgres -p 5432
```

## Cleanup

### Stop Lab

```bash
# Stop all containers
docker-compose -f docker-compose.chr.yml down

# Stop and remove volumes (deletes all data)
docker-compose -f docker-compose.chr.yml down -v
```

### Reset Router Configuration

```bash
# Reset to factory defaults
docker exec -it isp-chr-pppoe-01 ros /system/reset-configuration
```

## Performance Considerations

- Each CHR instance requires ~512MB RAM
- Total lab setup needs ~4GB RAM + overhead
- Polling interval affects database growth
- Consider using TimescaleDB for time-series data in production

## Security Notes

⚠️ **This lab setup is for development and testing only!**

- Default passwords are used (change in production)
- SNMP community is "public" (use SNMPv3 in production)
- API runs without TLS (enable API-SSL in production)
- No firewall rules restrict access (implement strict rules in production)

## Next Steps

1. Explore the adapter pattern by modifying `internal/poller/adapter/mikrotik_adapter.go`
2. Add support for additional MikroTik API endpoints
3. Implement Cisco/Juniper adapters using the same pattern
4. Create custom dashboards for role-specific metrics
5. Set up alerting rules for session limits and failures

## References

- [MikroTik RouterOS Documentation](https://help.mikrotik.com/docs/)
- [MikroTik API Documentation](https://help.mikrotik.com/docs/display/ROS/API)
- [RouterOS SNMP Configuration](https://help.mikrotik.com/docs/display/ROS/SNMP)
- [PPPoE Server Setup](https://help.mikrotik.com/docs/display/ROS/PPPoE)
