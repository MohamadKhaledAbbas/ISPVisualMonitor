package adapter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/gosnmp/gosnmp"
)

// SNMPAdapter implements polling via SNMP protocol
type SNMPAdapter struct {
	config AdapterConfig
}

// NewSNMPAdapter creates a new SNMP adapter
func NewSNMPAdapter(config AdapterConfig) *SNMPAdapter {
	return &SNMPAdapter{
		config: config,
	}
}

// GetAdapterName returns the adapter name
func (a *SNMPAdapter) GetAdapterName() string {
	return "snmp"
}

// CanHandle checks if this adapter can handle the router
func (a *SNMPAdapter) CanHandle(router *models.EnhancedRouter) bool {
	if router.Capabilities == nil {
		return false
	}
	if router.Capabilities.SNMP == nil {
		return false
	}
	return router.Capabilities.SNMP.Enabled
}

// GetSupportedMetrics returns supported metric types
func (a *SNMPAdapter) GetSupportedMetrics() []string {
	return []string{
		"interface_metrics",
		"cpu_usage",
		"memory_usage",
		"uptime",
		"system_info",
	}
}

// Poll performs SNMP polling
func (a *SNMPAdapter) Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error) {
	startTime := time.Now()
	result := NewPollResult(router.ID, router.TenantID, a.GetAdapterName())
	
	// Create SNMP client
	client, err := a.createSNMPClient(router)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create SNMP client: %v", err)
		return result, err
	}
	
	// Connect to SNMP agent
	err = client.Connect()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to connect: %v", err)
		return result, err
	}
	defer client.Conn.Close()
	
	// Poll system information
	if err := a.pollSystemInfo(client, result); err != nil {
		log.Printf("Warning: Failed to poll system info: %v", err)
	}
	
	// Poll interface statistics
	if err := a.pollInterfaces(client, result); err != nil {
		log.Printf("Warning: Failed to poll interfaces: %v", err)
	}
	
	// Calculate response time
	result.ResponseTimeMs = int(time.Since(startTime).Milliseconds())
	result.Success = true
	
	return result, nil
}

// HealthCheck tests SNMP connectivity
func (a *SNMPAdapter) HealthCheck(ctx context.Context, router *models.EnhancedRouter) error {
	client, err := a.createSNMPClient(router)
	if err != nil {
		return err
	}
	
	err = client.Connect()
	if err != nil {
		return err
	}
	defer client.Conn.Close()
	
	// Try to get system description (1.3.6.1.2.1.1.1.0)
	oids := []string{"1.3.6.1.2.1.1.1.0"}
	result, err := client.Get(oids)
	if err != nil {
		return err
	}
	
	if len(result.Variables) == 0 {
		return fmt.Errorf("no SNMP response")
	}
	
	return nil
}

// createSNMPClient creates and configures an SNMP client
func (a *SNMPAdapter) createSNMPClient(router *models.EnhancedRouter) (*gosnmp.GoSNMP, error) {
	if router.Capabilities == nil || router.Capabilities.SNMP == nil {
		return nil, fmt.Errorf("SNMP not configured")
	}
	
	snmpCfg := router.Capabilities.SNMP
	
	client := &gosnmp.GoSNMP{
		Target:    router.ManagementIP,
		Port:      uint16(snmpCfg.Port),
		Timeout:   time.Duration(snmpCfg.TimeoutSeconds) * time.Second,
		Retries:   snmpCfg.Retries,
		Transport: "udp",
	}
	
	// Configure based on version
	switch snmpCfg.Version {
	case "v1":
		client.Version = gosnmp.Version1
		if snmpCfg.Community != nil {
			client.Community = *snmpCfg.Community
		}
	case "v2c":
		client.Version = gosnmp.Version2c
		if snmpCfg.Community != nil {
			client.Community = *snmpCfg.Community
		}
	case "v3":
		client.Version = gosnmp.Version3
		if snmpCfg.V3Username != nil {
			client.SecurityModel = gosnmp.UserSecurityModel
			client.MsgFlags = gosnmp.AuthPriv
			
			params := &gosnmp.UsmSecurityParameters{
				UserName: *snmpCfg.V3Username,
			}
			
			// Set authentication
			if snmpCfg.V3AuthProtocol != nil && snmpCfg.V3AuthPassword != nil {
				params.AuthenticationPassphrase = *snmpCfg.V3AuthPassword
				switch *snmpCfg.V3AuthProtocol {
				case "MD5":
					params.AuthenticationProtocol = gosnmp.MD5
				case "SHA":
					params.AuthenticationProtocol = gosnmp.SHA
				case "SHA-224":
					params.AuthenticationProtocol = gosnmp.SHA224
				case "SHA-256":
					params.AuthenticationProtocol = gosnmp.SHA256
				case "SHA-384":
					params.AuthenticationProtocol = gosnmp.SHA384
				case "SHA-512":
					params.AuthenticationProtocol = gosnmp.SHA512
				}
			}
			
			// Set privacy
			if snmpCfg.V3PrivProtocol != nil && snmpCfg.V3PrivPassword != nil {
				params.PrivacyPassphrase = *snmpCfg.V3PrivPassword
				switch *snmpCfg.V3PrivProtocol {
				case "DES":
					params.PrivacyProtocol = gosnmp.DES
				case "AES":
					params.PrivacyProtocol = gosnmp.AES
				case "AES-192":
					params.PrivacyProtocol = gosnmp.AES192
				case "AES-256":
					params.PrivacyProtocol = gosnmp.AES256
				}
			}
			
			client.SecurityParameters = params
		}
	default:
		return nil, fmt.Errorf("unsupported SNMP version: %s", snmpCfg.Version)
	}
	
	return client, nil
}

// pollSystemInfo polls basic system information
func (a *SNMPAdapter) pollSystemInfo(client *gosnmp.GoSNMP, result *PollResult) error {
	// Standard System MIB OIDs
	oids := []string{
		"1.3.6.1.2.1.1.1.0",  // sysDescr
		"1.3.6.1.2.1.1.3.0",  // sysUpTime
		"1.3.6.1.2.1.1.5.0",  // sysName
		"1.3.6.1.2.1.25.1.5.0", // hrSystemUptime (if available)
	}
	
	response, err := client.Get(oids)
	if err != nil {
		return err
	}
	
	for _, variable := range response.Variables {
		switch variable.Name {
		case "1.3.6.1.2.1.1.1.0":
			result.Metrics["system_description"] = fmt.Sprintf("%s", variable.Value)
		case "1.3.6.1.2.1.1.3.0":
			if uptime, ok := variable.Value.(uint32); ok {
				result.Metrics["uptime_seconds"] = int64(uptime / 100) // Convert timeticks to seconds
			}
		case "1.3.6.1.2.1.1.5.0":
			result.Metrics["system_name"] = fmt.Sprintf("%s", variable.Value)
		}
	}
	
	return nil
}

// pollInterfaces polls interface statistics
func (a *SNMPAdapter) pollInterfaces(client *gosnmp.GoSNMP, result *PollResult) error {
	// Walk the interface table (IF-MIB)
	// This is a simplified implementation - in production, you'd want to:
	// 1. First walk ifIndex to get all interface indices
	// 2. Then bulk get all stats for each interface
	// 3. Calculate rates from previous polls
	
	// For now, just get basic interface count
	oids := []string{
		"1.3.6.1.2.1.2.1.0", // ifNumber
	}
	
	response, err := client.Get(oids)
	if err != nil {
		return err
	}
	
	for _, variable := range response.Variables {
		if variable.Name == "1.3.6.1.2.1.2.1.0" {
			if ifCount, ok := variable.Value.(int); ok {
				result.Metrics["interface_count"] = ifCount
			}
		}
	}
	
	// TODO: Implement full interface statistics collection
	// This would include walking:
	// - ifDescr (1.3.6.1.2.1.2.2.1.2)
	// - ifOperStatus (1.3.6.1.2.1.2.2.1.8)
	// - ifInOctets (1.3.6.1.2.1.2.2.1.10)
	// - ifOutOctets (1.3.6.1.2.1.2.2.1.16)
	// - etc.
	
	return nil
}
