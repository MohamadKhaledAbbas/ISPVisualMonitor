package adapter

import (
	"context"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// PollerAdapter is the interface that all polling adapters must implement
type PollerAdapter interface {
	// GetAdapterName returns the name of this adapter (e.g., "snmp", "mikrotik_api")
	GetAdapterName() string
	
	// CanHandle determines if this adapter can handle polling for the given router
	CanHandle(router *models.EnhancedRouter) bool
	
	// Poll performs the actual polling operation and returns results
	Poll(ctx context.Context, router *models.EnhancedRouter) (*PollResult, error)
	
	// HealthCheck tests connectivity to the router without full polling
	HealthCheck(ctx context.Context, router *models.EnhancedRouter) error
	
	// GetSupportedMetrics returns the list of metric types this adapter can collect
	GetSupportedMetrics() []string
}

// PollResult represents the result of a polling operation
type PollResult struct {
	RouterID    uuid.UUID              `json:"router_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Success     bool                   `json:"success"`
	AdapterUsed string                 `json:"adapter_used"`
	Timestamp   time.Time              `json:"timestamp"`
	
	// Generic metrics
	Metrics map[string]interface{} `json:"metrics"`
	
	// Role-specific data
	PPPoESessions []models.PPPoESession `json:"pppoe_sessions,omitempty"`
	NATSessions   []models.NATSession   `json:"nat_sessions,omitempty"`
	DHCPLeases    []models.DHCPLease    `json:"dhcp_leases,omitempty"`
	Interfaces    []InterfaceStatus     `json:"interfaces,omitempty"`
	
	// Performance metrics
	ResponseTimeMs int    `json:"response_time_ms"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

// InterfaceStatus represents the status and metrics of a network interface
type InterfaceStatus struct {
	Name              string  `json:"name"`
	Description       string  `json:"description,omitempty"`
	IfIndex           int     `json:"if_index,omitempty"`
	Status            string  `json:"status"`           // up, down, admin-down
	AdminStatus       string  `json:"admin_status"`     // up, down
	Speed             int64   `json:"speed_mbps,omitempty"`
	MTU               int     `json:"mtu,omitempty"`
	InOctets          int64   `json:"in_octets"`
	OutOctets         int64   `json:"out_octets"`
	InPackets         int64   `json:"in_packets"`
	OutPackets        int64   `json:"out_packets"`
	InErrors          int64   `json:"in_errors"`
	OutErrors         int64   `json:"out_errors"`
	InDiscards        int64   `json:"in_discards"`
	OutDiscards       int64   `json:"out_discards"`
	UtilizationPercent float64 `json:"utilization_percent,omitempty"`
}

// SystemMetrics represents general system health metrics
type SystemMetrics struct {
	CPUPercent         float64 `json:"cpu_percent"`
	MemoryPercent      float64 `json:"memory_percent"`
	MemoryTotalMB      int64   `json:"memory_total_mb"`
	MemoryUsedMB       int64   `json:"memory_used_mb"`
	MemoryFreeMB       int64   `json:"memory_free_mb"`
	UptimeSeconds      int64   `json:"uptime_seconds"`
	TemperatureCelsius float64 `json:"temperature_celsius,omitempty"`
}

// PPPoEMetrics represents PPPoE server specific metrics
type PPPoEMetrics struct {
	TotalSessions       int     `json:"total_sessions"`
	ActiveSessions      int     `json:"active_sessions"`
	MaxSessions         int     `json:"max_sessions"`
	AuthSuccesses       int64   `json:"auth_successes"`
	AuthFailures        int64   `json:"auth_failures"`
	ThroughputInMbps    float64 `json:"throughput_in_mbps"`
	ThroughputOutMbps   float64 `json:"throughput_out_mbps"`
}

// NATMetrics represents NAT gateway specific metrics
type NATMetrics struct {
	TotalSessions     int     `json:"total_sessions"`
	MaxSessions       int     `json:"max_sessions"`
	PoolUtilization   float64 `json:"pool_utilization_percent"`
	PortExhaustionPct float64 `json:"port_exhaustion_percent"`
	NewSessionsPerSec float64 `json:"new_sessions_per_sec"`
}

// AdapterConfig holds common configuration for adapters
type AdapterConfig struct {
	TimeoutSeconds int
	RetryAttempts  int
	RetryDelay     time.Duration
}

// DefaultAdapterConfig returns sensible defaults
func DefaultAdapterConfig() AdapterConfig {
	return AdapterConfig{
		TimeoutSeconds: 30,
		RetryAttempts:  3,
		RetryDelay:     2 * time.Second,
	}
}

// NewPollResult creates a new PollResult with default values
func NewPollResult(routerID, tenantID uuid.UUID, adapterName string) *PollResult {
	return &PollResult{
		RouterID:       routerID,
		TenantID:       tenantID,
		Success:        false,
		AdapterUsed:    adapterName,
		Timestamp:      time.Now(),
		Metrics:        make(map[string]interface{}),
		PPPoESessions:  []models.PPPoESession{},
		NATSessions:    []models.NATSession{},
		DHCPLeases:     []models.DHCPLease{},
		Interfaces:     []InterfaceStatus{},
	}
}

// SetSystemMetrics sets system metrics in the result
func (pr *PollResult) SetSystemMetrics(sm SystemMetrics) {
	pr.Metrics["cpu_percent"] = sm.CPUPercent
	pr.Metrics["memory_percent"] = sm.MemoryPercent
	pr.Metrics["memory_total_mb"] = sm.MemoryTotalMB
	pr.Metrics["memory_used_mb"] = sm.MemoryUsedMB
	pr.Metrics["memory_free_mb"] = sm.MemoryFreeMB
	pr.Metrics["uptime_seconds"] = sm.UptimeSeconds
	if sm.TemperatureCelsius > 0 {
		pr.Metrics["temperature_celsius"] = sm.TemperatureCelsius
	}
}

// SetPPPoEMetrics sets PPPoE metrics in the result
func (pr *PollResult) SetPPPoEMetrics(pm PPPoEMetrics) {
	pr.Metrics["pppoe_total_sessions"] = pm.TotalSessions
	pr.Metrics["pppoe_active_sessions"] = pm.ActiveSessions
	pr.Metrics["pppoe_max_sessions"] = pm.MaxSessions
	pr.Metrics["pppoe_auth_successes"] = pm.AuthSuccesses
	pr.Metrics["pppoe_auth_failures"] = pm.AuthFailures
	pr.Metrics["pppoe_throughput_in_mbps"] = pm.ThroughputInMbps
	pr.Metrics["pppoe_throughput_out_mbps"] = pm.ThroughputOutMbps
}

// SetNATMetrics sets NAT metrics in the result
func (pr *PollResult) SetNATMetrics(nm NATMetrics) {
	pr.Metrics["nat_total_sessions"] = nm.TotalSessions
	pr.Metrics["nat_max_sessions"] = nm.MaxSessions
	pr.Metrics["nat_pool_utilization"] = nm.PoolUtilization
	pr.Metrics["nat_port_exhaustion_pct"] = nm.PortExhaustionPct
	pr.Metrics["nat_new_sessions_per_sec"] = nm.NewSessionsPerSec
}

// GetMetricsCount returns the total number of metrics collected
func (pr *PollResult) GetMetricsCount() int {
	count := len(pr.Metrics)
	count += len(pr.PPPoESessions)
	count += len(pr.NATSessions)
	count += len(pr.DHCPLeases)
	count += len(pr.Interfaces)
	return count
}
