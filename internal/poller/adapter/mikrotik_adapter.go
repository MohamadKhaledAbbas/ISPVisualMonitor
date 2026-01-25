package adapter

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"gopkg.in/routeros.v2"
)

// MikroTikAdapter implements polling via MikroTik RouterOS API
type MikroTikAdapter struct {
	config AdapterConfig
}

// NewMikroTikAdapter creates a new MikroTik adapter
func NewMikroTikAdapter(config AdapterConfig) *MikroTikAdapter {
	return &MikroTikAdapter{
		config: config,
	}
}

// GetAdapterName returns the adapter name
func (a *MikroTikAdapter) GetAdapterName() string {
	return "mikrotik_api"
}

// CanHandle checks if this adapter can handle the router
func (a *MikroTikAdapter) CanHandle(router *models.EnhancedRouter) bool {
	if router.Capabilities == nil || router.Capabilities.API == nil {
		return false
	}
	
	api := router.Capabilities.API
	return api.Enabled && api.Type == "mikrotik"
}

// GetSupportedMetrics returns supported metric types
func (a *MikroTikAdapter) GetSupportedMetrics() []string {
	return []string{
		"interface_metrics",
		"cpu_usage",
		"memory_usage",
		"uptime",
		"pppoe_sessions",
		"nat_sessions",
		"dhcp_leases",
		"system_info",
	}
}

// Poll performs RouterOS API polling
func (a *MikroTikAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
	startTime := time.Now()
	result := NewPollResult(router.ID, router.TenantID, a.GetAdapterName())
	
	// Create RouterOS client
	client, err := a.createClient(router)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create client: %v", err)
		return result, err
	}
	defer client.Close()
	
	// Poll system resources
	if err := a.pollSystemResources(client, result); err != nil {
		log.Printf("Warning: Failed to poll system resources: %v", err)
	}
	
	// Poll interfaces
	if err := a.pollInterfaces(client, result); err != nil {
		log.Printf("Warning: Failed to poll interfaces: %v", err)
	}
	
	// Poll PPPoE sessions if router has pppoe_server role
	if router.HasRole(models.RoleCodePPPoEServer) {
		if err := a.pollPPPoESessions(client, router, result); err != nil {
			log.Printf("Warning: Failed to poll PPPoE sessions: %v", err)
		}
	}
	
	// Poll NAT sessions if router has nat_gateway role
	if router.HasRole(models.RoleCodeNATGateway) {
		if err := a.pollNATSessions(client, router, result); err != nil {
			log.Printf("Warning: Failed to poll NAT sessions: %v", err)
		}
	}
	
	// Poll DHCP leases if router has dhcp_server role
	if router.HasRole(models.RoleCodeDHCPServer) {
		if err := a.pollDHCPLeases(client, router, result); err != nil {
			log.Printf("Warning: Failed to poll DHCP leases: %v", err)
		}
	}
	
	// Calculate response time
	result.ResponseTimeMs = int(time.Since(startTime).Milliseconds())
	result.Success = true
	
	return result, nil
}

// HealthCheck tests RouterOS API connectivity
func (a *MikroTikAdapter) HealthCheck(ctx context.Context, router *models.EnhancedRouter) error {
	client, err := a.createClient(router)
	if err != nil {
		return err
	}
	defer client.Close()
	
	// Try a simple command
	_, err = client.Run("/system/resource/print")
	return err
}

// createClient creates and connects a RouterOS API client
func (a *MikroTikAdapter) createClient(router *models.EnhancedRouter) (*routeros.Client, error) {
	if router.Capabilities == nil || router.Capabilities.API == nil {
		return nil, fmt.Errorf("API not configured")
	}
	
	apiCfg := router.Capabilities.API
	
	// Determine address
	address := router.ManagementIP
	if apiCfg.Port != nil && *apiCfg.Port != 0 {
		address = fmt.Sprintf("%s:%d", address, *apiCfg.Port)
	} else {
		// Default API port
		if apiCfg.UseTLS {
			address = fmt.Sprintf("%s:8729", address)
		} else {
			address = fmt.Sprintf("%s:8728", address)
		}
	}
	
	// Set timeout
	timeout := time.Duration(apiCfg.TimeoutSeconds) * time.Second
	
	// Connect
	var client *routeros.Client
	var err error
	
	if apiCfg.UseTLS {
		// NOTE: The current version of gopkg.in/routeros.v2 doesn't have full TLS support
		// For production use, consider using a newer library or implementing custom TLS dialing
		// For now, fall back to regular connection
		log.Printf("Warning: TLS requested for %s but library doesn't support TLS config, using regular connection", router.ManagementIP)
		client, err = routeros.Dial(address, apiCfg.Username, apiCfg.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	} else {
		client, err = routeros.DialTimeout(address, apiCfg.Username, apiCfg.Password, timeout)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	
	return client, nil
}

// pollSystemResources polls system resource information
func (a *MikroTikAdapter) pollSystemResources(client *routeros.Client, result *PollResult) error {
	reply, err := client.Run("/system/resource/print")
	if err != nil {
		return err
	}
	
	if len(reply.Re) > 0 {
		res := reply.Re[0].Map
		
		// CPU
		if cpuLoad, ok := res["cpu-load"]; ok {
			if cpu, err := strconv.ParseFloat(cpuLoad, 64); err == nil {
				result.Metrics["cpu_percent"] = cpu
			}
		}
		
		// Memory
		if totalMem, ok := res["total-memory"]; ok {
			if freeMem, ok2 := res["free-memory"]; ok2 {
				total, _ := strconv.ParseInt(totalMem, 10, 64)
				free, _ := strconv.ParseInt(freeMem, 10, 64)
				used := total - free
				
				result.Metrics["memory_total_mb"] = total / (1024 * 1024)
				result.Metrics["memory_used_mb"] = used / (1024 * 1024)
				result.Metrics["memory_free_mb"] = free / (1024 * 1024)
				
				if total > 0 {
					result.Metrics["memory_percent"] = float64(used) / float64(total) * 100.0
				}
			}
		}
		
		// Uptime
		if uptime, ok := res["uptime"]; ok {
			// Parse uptime string (e.g., "1w2d3h4m5s")
			// For simplicity, store as string
			result.Metrics["uptime"] = uptime
		}
		
		// Board name
		if boardName, ok := res["board-name"]; ok {
			result.Metrics["board_name"] = boardName
		}
		
		// Version
		if version, ok := res["version"]; ok {
			result.Metrics["version"] = version
		}
	}
	
	return nil
}

// pollInterfaces polls interface statistics
func (a *MikroTikAdapter) pollInterfaces(client *routeros.Client, result *PollResult) error {
	reply, err := client.Run("/interface/print", "=stats")
	if err != nil {
		return err
	}
	
	interfaces := []InterfaceStatus{}
	
	for _, re := range reply.Re {
		iface := InterfaceStatus{
			Name: re.Map["name"],
		}
		
		// Status
		if running, ok := re.Map["running"]; ok {
			if running == "true" {
				iface.Status = "up"
			} else {
				iface.Status = "down"
			}
		}
		
		if disabled, ok := re.Map["disabled"]; ok {
			if disabled == "true" {
				iface.AdminStatus = "down"
			} else {
				iface.AdminStatus = "up"
			}
		}
		
		// Traffic stats
		if rxBytes, ok := re.Map["rx-byte"]; ok {
			if val, err := strconv.ParseInt(rxBytes, 10, 64); err == nil {
				iface.InOctets = val
			}
		}
		
		if txBytes, ok := re.Map["tx-byte"]; ok {
			if val, err := strconv.ParseInt(txBytes, 10, 64); err == nil {
				iface.OutOctets = val
			}
		}
		
		if rxPackets, ok := re.Map["rx-packet"]; ok {
			if val, err := strconv.ParseInt(rxPackets, 10, 64); err == nil {
				iface.InPackets = val
			}
		}
		
		if txPackets, ok := re.Map["tx-packet"]; ok {
			if val, err := strconv.ParseInt(txPackets, 10, 64); err == nil {
				iface.OutPackets = val
			}
		}
		
		interfaces = append(interfaces, iface)
	}
	
	result.Interfaces = interfaces
	result.Metrics["interface_count"] = len(interfaces)
	
	return nil
}

// pollPPPoESessions polls active PPPoE sessions
func (a *MikroTikAdapter) pollPPPoESessions(client *routeros.Client, router *models.EnhancedRouter, result *PollResult) error {
	reply, err := client.Run("/ppp/active/print")
	if err != nil {
		return err
	}
	
	sessions := []models.PPPoESession{}
	
	for _, re := range reply.Re {
		session := models.PPPoESession{
			TenantID: router.TenantID,
			RouterID: router.ID,
			Username: re.Map["name"],
			Status:   models.SessionStatusActive,
		}
		
		// Session ID
		if id, ok := re.Map[".id"]; ok {
			session.SessionID = &id
		}
		
		// Calling station (MAC)
		if callingStation, ok := re.Map["caller-id"]; ok {
			session.CallingStationID = &callingStation
		}
		
		// IP Address
		if address, ok := re.Map["address"]; ok {
			if ip := net.ParseIP(address); ip != nil {
				session.FramedIPAddress = &ip
			}
		}
		
		// Uptime
		if uptime, ok := re.Map["uptime"]; ok {
			// Store as string for now
			// TODO: Parse uptime string to seconds
			_ = uptime
		}
		
		sessions = append(sessions, session)
	}
	
	result.PPPoESessions = sessions
	
	// Set PPPoE metrics
	result.SetPPPoEMetrics(PPPoEMetrics{
		TotalSessions:  len(sessions),
		ActiveSessions: len(sessions),
	})
	
	return nil
}

// pollNATSessions polls active NAT connections
func (a *MikroTikAdapter) pollNATSessions(client *routeros.Client, router *models.EnhancedRouter, result *PollResult) error {
	reply, err := client.Run("/ip/firewall/connection/print", "=count-only")
	if err != nil {
		return err
	}
	
	// Get connection count
	count := 0
	if len(reply.Re) > 0 {
		// MikroTik returns count in a specific way
		// This is simplified
		count = len(reply.Re)
	}
	
	// Set NAT metrics
	result.SetNATMetrics(NATMetrics{
		TotalSessions: count,
	})
	
	// Note: Full NAT session details would require more processing
	// For now, just store the count
	
	return nil
}

// pollDHCPLeases polls DHCP leases
func (a *MikroTikAdapter) pollDHCPLeases(client *routeros.Client, router *models.EnhancedRouter, result *PollResult) error {
	reply, err := client.Run("/ip/dhcp-server/lease/print")
	if err != nil {
		return err
	}
	
	leases := []models.DHCPLease{}
	
	for _, re := range reply.Re {
		lease := models.DHCPLease{
			TenantID: router.TenantID,
			RouterID: router.ID,
		}
		
		// MAC Address
		if mac, ok := re.Map["mac-address"]; ok {
			lease.MACAddress = mac
		}
		
		// IP Address
		if address, ok := re.Map["address"]; ok {
			if ip := net.ParseIP(address); ip != nil {
				lease.IPAddress = ip
			}
		}
		
		// Hostname
		if hostname, ok := re.Map["host-name"]; ok {
			lease.Hostname = &hostname
		}
		
		// Status
		if status, ok := re.Map["status"]; ok {
			if status == "bound" {
				state := models.LeaseStateActive
				lease.LeaseState = &state
			}
		}
		
		leases = append(leases, lease)
	}
	
	result.DHCPLeases = leases
	result.Metrics["dhcp_lease_count"] = len(leases)
	
	return nil
}
