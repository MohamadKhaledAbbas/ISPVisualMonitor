# Adapter Development Guide

## Overview

The ISP Visual Monitor uses a pluggable adapter pattern to support multiple vendor-specific APIs and protocols. This guide explains how to develop new adapters for additional network equipment vendors.

## Architecture

### Adapter Interface

All adapters must implement the `PollerAdapter` interface defined in `internal/poller/adapter/adapter.go`:

```go
type PollerAdapter interface {
    GetAdapterName() string
    CanHandle(router *models.EnhancedRouter) bool
    Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error)
    HealthCheck(ctx context.Context, router *models.EnhancedRouter) error
    GetSupportedMetrics() []string
}
```

### Key Components

1. **Adapter Interface**: Defines the contract all adapters must follow
2. **PollResult**: Standardized result structure containing metrics and session data
3. **Registry**: Manages adapter registration and selection
4. **Router Capabilities**: Configuration for connection methods (API, SNMP, SSH, etc.)

## Creating a New Adapter

### Step 1: Choose a Template

Start by copying an existing adapter as a template:

```bash
# For API-based vendors (Cisco RESTCONF, Juniper NETCONF)
cp internal/poller/adapter/mikrotik_adapter.go internal/poller/adapter/cisco_adapter.go

# For SNMP-only vendors
cp internal/poller/adapter/snmp_adapter.go internal/poller/adapter/generic_snmp_adapter.go
```

### Step 2: Implement the Interface

#### Example: Cisco IOS XE RESTCONF Adapter

```go
package adapter

import (
    "context"
    "fmt"
    "time"
    
    "github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

type CiscoRESTCONFAdapter struct {
    config AdapterConfig
}

func NewCiscoRESTCONFAdapter(config AdapterConfig) *CiscoRESTCONFAdapter {
    return &CiscoRESTCONFAdapter{
        config: config,
    }
}

func (a *CiscoRESTCONFAdapter) GetAdapterName() string {
    return "cisco_restconf"
}

func (a *CiscoRESTCONFAdapter) CanHandle(router *models.EnhancedRouter) bool {
    if router.Capabilities == nil || router.Capabilities.API == nil {
        return false
    }
    
    api := router.Capabilities.API
    return api.Enabled && api.Type == "cisco_restconf"
}

func (a *CiscoRESTCONFAdapter) GetSupportedMetrics() []string {
    return []string{
        "interface_metrics",
        "cpu_usage",
        "memory_usage",
        "uptime",
        "system_info",
    }
}

func (a *CiscoRESTCONFAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
    startTime := time.Now()
    result := NewPollResult(router.ID, router.TenantID, a.GetAdapterName())
    
    // Create HTTP client
    client, err := a.createRESTCONFClient(router)
    if err != nil {
        result.ErrorMessage = fmt.Sprintf("Failed to create client: %v", err)
        return result, err
    }
    
    // Poll system information
    if err := a.pollSystemInfo(client, result); err != nil {
        log.Printf("Warning: Failed to poll system info: %v", err)
    }
    
    // Poll interfaces
    if err := a.pollInterfaces(client, result); err != nil {
        log.Printf("Warning: Failed to poll interfaces: %v", err)
    }
    
    result.ResponseTimeMs = int(time.Since(startTime).Milliseconds())
    result.Success = true
    
    return result, nil
}

func (a *CiscoRESTCONFAdapter) HealthCheck(ctx context.Context, router *models.EnhancedRouter) error {
    // Implement health check (e.g., GET /restconf/data/ietf-system:system-state)
    return nil
}

// Private helper methods
func (a *CiscoRESTCONFAdapter) createRESTCONFClient(router *models.EnhancedRouter) (*http.Client, error) {
    // Create HTTP client with authentication
    return &http.Client{
        Timeout: time.Duration(router.Capabilities.API.TimeoutSeconds) * time.Second,
    }, nil
}

func (a *CiscoRESTCONFAdapter) pollSystemInfo(client *http.Client, result *PollResult) error {
    // GET /restconf/data/ietf-system:system-state
    // Parse JSON response and populate result.Metrics
    return nil
}

func (a *CiscoRESTCONFAdapter) pollInterfaces(client *http.Client, result *PollResult) error {
    // GET /restconf/data/ietf-interfaces:interfaces-state
    // Parse JSON response and populate result.Interfaces
    return nil
}
```

### Step 3: Register the Adapter

Edit `internal/poller/adapter/registry.go` to register your new adapter:

```go
func NewRegistry(config AdapterConfig) *Registry {
    r := &Registry{
        adapters: []PollerAdapter{},
        config:   config,
    }
    
    // Register adapters in priority order
    r.Register(NewMikroTikAdapter(config))
    r.Register(NewCiscoRESTCONFAdapter(config))  // Add your adapter
    r.Register(NewSNMPAdapter(config))
    
    return r
}
```

### Step 4: Add Router Capability Support

Update the database to support your adapter type:

```sql
-- Add new API type
ALTER TABLE router_capabilities 
  DROP CONSTRAINT IF EXISTS router_capabilities_api_type_check;

ALTER TABLE router_capabilities 
  ADD CONSTRAINT router_capabilities_api_type_check 
  CHECK (api_type IN ('mikrotik', 'cisco_restconf', 'juniper_netconf', 'arista_eapi', 'your_new_type'));
```

### Step 5: Test Your Adapter

Create unit tests:

```go
// internal/poller/adapter/cisco_adapter_test.go
package adapter

import (
    "context"
    "testing"
    
    "github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

func TestCiscoRESTCONFAdapter_CanHandle(t *testing.T) {
    adapter := NewCiscoRESTCONFAdapter(DefaultAdapterConfig())
    
    tests := []struct {
        name     string
        router   *models.EnhancedRouter
        expected bool
    }{
        {
            name: "Cisco RESTCONF enabled",
            router: &models.EnhancedRouter{
                Router: models.Router{},
                Capabilities: &models.RouterCapabilities{
                    API: &models.APICapability{
                        Enabled: true,
                        Type:    "cisco_restconf",
                    },
                },
            },
            expected: true,
        },
        {
            name: "Different API type",
            router: &models.EnhancedRouter{
                Router: models.Router{},
                Capabilities: &models.RouterCapabilities{
                    API: &models.APICapability{
                        Enabled: true,
                        Type:    "mikrotik",
                    },
                },
            },
            expected: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := adapter.CanHandle(tt.router)
            if result != tt.expected {
                t.Errorf("CanHandle() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Best Practices

### 1. Error Handling

Always return meaningful errors and populate `PollResult.ErrorMessage`:

```go
if err != nil {
    result.ErrorMessage = fmt.Sprintf("Failed to connect: %v", err)
    return result, err
}
```

### 2. Timeouts

Respect the configured timeout:

```go
ctx, cancel := context.WithTimeout(ctx, time.Duration(a.config.TimeoutSeconds)*time.Second)
defer cancel()
```

### 3. Partial Success

If some data collection fails, log warnings but continue:

```go
if err := a.pollSystemInfo(client, result); err != nil {
    log.Printf("Warning: Failed to poll system info: %v", err)
    // Continue with other data collection
}
```

### 4. Role-Based Polling

Check router roles and collect appropriate metrics:

```go
if router.HasRole(models.RoleCodePPPoEServer) {
    if err := a.pollPPPoESessions(client, router, result); err != nil {
        log.Printf("Warning: Failed to poll PPPoE sessions: %v", err)
    }
}
```

### 5. Credential Security

Never log credentials. Use them only for authentication:

```go
// BAD
log.Printf("Connecting with password: %s", apiCfg.Password)

// GOOD
log.Printf("Connecting to %s", router.ManagementIP)
```

### 6. Connection Pooling

Reuse connections when possible:

```go
type CiscoAdapter struct {
    config AdapterConfig
    client *http.Client  // Shared client
}
```

## Vendor-Specific Examples

### Juniper NETCONF

```go
type JuniperNETCONFAdapter struct {
    config AdapterConfig
}

func (a *JuniperNETCONFAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
    // Use netconf library
    session := netconf.NewSession(/* ... */)
    
    // RPC call: <get-interface-information/>
    reply, err := session.Exec(netconf.RawMethod("<get-interface-information/>"))
    
    // Parse XML response
    // ...
}
```

### Arista eAPI

```go
type AristaeAPIAdapter struct {
    config AdapterConfig
}

func (a *AristaeAPIAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
    // Use JSON-RPC over HTTPS
    commands := []string{
        "show version",
        "show interfaces",
        "show system resources",
    }
    
    // Send JSON-RPC request
    jsonRPC := map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  "runCmds",
        "params": map[string]interface{}{
            "version": 1,
            "cmds":    commands,
            "format":  "json",
        },
        "id": 1,
    }
    
    // Parse JSON response
    // ...
}
```

### Generic SNMP with Vendor MIBs

```go
type CiscoPPPoESNMPAdapter struct {
    SNMPAdapter  // Embed base SNMP adapter
}

func (a *CiscoPPPoESNMPAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
    result, err := a.SNMPAdapter.Poll(ctx, router)
    if err != nil {
        return result, err
    }
    
    // Add Cisco-specific PPPoE MIB polling
    if router.HasRole(models.RoleCodePPPoEServer) {
        client, _ := a.createSNMPClient(router)
        a.pollCiscoPPPoEMIB(client, result)
    }
    
    return result, nil
}
```

## Testing with Real Devices

### 1. Create Test Configuration

```sql
-- Add test router
INSERT INTO routers (tenant_id, name, management_ip, vendor, status, polling_enabled)
VALUES ('<tenant_id>', 'test-cisco-01', '192.168.1.10', 'Cisco', 'active', true);

-- Add capabilities
INSERT INTO router_capabilities (
    router_id, tenant_id,
    api_enabled, api_type, api_endpoint, api_username, api_password,
    preferred_method
) VALUES (
    '<router_id>', '<tenant_id>',
    true, 'cisco_restconf', 'https://192.168.1.10', 'admin', 'password',
    'api'
);
```

### 2. Enable Debug Logging

```go
// Add verbose logging to your adapter
log.Printf("[DEBUG] Connecting to %s", router.ManagementIP)
log.Printf("[DEBUG] Using credentials: %s", apiCfg.Username)
log.Printf("[DEBUG] Response: %s", string(responseBody))
```

### 3. Monitor Polling

```bash
# Watch poller logs
docker-compose logs -f isp-monitor-dev | grep -i cisco

# Check polling history
psql -c "SELECT * FROM polling_history WHERE router_id = '<router_id>' ORDER BY poll_started_at DESC LIMIT 10;"
```

## Common Issues and Solutions

### Issue: Authentication Fails

**Solution**: Verify credentials and authentication method:

```go
// Test credentials manually first
curl -u username:password https://router-ip/restconf/data/...

// Add detailed error logging
log.Printf("Auth error: %v", err)
log.Printf("Status code: %d", resp.StatusCode)
```

### Issue: Timeout Errors

**Solution**: Increase timeout or optimize queries:

```go
// Increase timeout
config := AdapterConfig{
    TimeoutSeconds: 60,  // Increase from 30
}

// Or batch multiple queries
responses, err := client.BatchGet([]string{oid1, oid2, oid3})
```

### Issue: Partial Data

**Solution**: Implement graceful degradation:

```go
// Don't fail entire poll if one metric fails
var errors []string
if err := a.pollMetric1(...); err != nil {
    errors = append(errors, fmt.Sprintf("metric1: %v", err))
}
if err := a.pollMetric2(...); err != nil {
    errors = append(errors, fmt.Sprintf("metric2: %v", err))
}

if len(errors) > 0 {
    result.ErrorMessage = strings.Join(errors, "; ")
}
result.Success = len(errors) < totalMetrics/2  // Success if >50% work
```

## Documentation Requirements

When submitting a new adapter, include:

1. **README section** describing:
   - Vendor and model support
   - Required firmware/API versions
   - Supported metrics and features
   - Configuration example

2. **Test results** showing:
   - Successful polling examples
   - Performance benchmarks
   - Error handling verification

3. **Migration guide** if needed:
   - Database schema changes
   - Configuration updates
   - Breaking changes

## Contributing Your Adapter

1. Fork the repository
2. Create a feature branch: `git checkout -b adapter/vendor-name`
3. Implement and test your adapter
4. Add tests and documentation
5. Submit a pull request

Include in your PR description:
- Vendor/model compatibility matrix
- Sample configuration
- Test results
- Any special requirements or dependencies

## Resources

- [MikroTik RouterOS API Documentation](https://help.mikrotik.com/docs/display/ROS/API)
- [Cisco RESTCONF Developer Guide](https://developer.cisco.com/docs/ios-xe/)
- [Juniper NETCONF Documentation](https://www.juniper.net/documentation/us/en/software/junos/netconf/)
- [Arista eAPI Documentation](https://www.arista.com/en/um-eos/eos-section-6-5-eapi)
- [SNMP MIB Repository](http://www.oidview.com/mibs/0/md-0.html)

## Support

For questions or issues with adapter development:

1. Check existing adapters for examples
2. Review the troubleshooting section
3. Open a GitHub issue with the `adapter-development` label
4. Join our community discussions

Happy adapter development! 🚀
