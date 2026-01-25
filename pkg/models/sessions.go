package models

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// PPPoESession represents an active or historical PPPoE session
type PPPoESession struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RouterID           uuid.UUID  `json:"router_id" db:"router_id"`
	SessionID          *string    `json:"session_id,omitempty" db:"session_id"` // Router-specific session ID
	Username           string     `json:"username" db:"username"`
	CallingStationID   *string    `json:"calling_station_id,omitempty" db:"calling_station_id"` // MAC address
	FramedIPAddress    *net.IP    `json:"framed_ip_address,omitempty" db:"framed_ip_address"`
	NASIPAddress       *net.IP    `json:"nas_ip_address,omitempty" db:"nas_ip_address"`
	NASPort            *string    `json:"nas_port,omitempty" db:"nas_port"`
	ServiceType        *string    `json:"service_type,omitempty" db:"service_type"`
	SessionTimeSeconds *int64     `json:"session_time_seconds,omitempty" db:"session_time_seconds"`
	IdleTimeSeconds    *int64     `json:"idle_time_seconds,omitempty" db:"idle_time_seconds"`
	BytesIn            int64      `json:"bytes_in" db:"bytes_in"`
	BytesOut           int64      `json:"bytes_out" db:"bytes_out"`
	PacketsIn          int64      `json:"packets_in" db:"packets_in"`
	PacketsOut         int64      `json:"packets_out" db:"packets_out"`
	Status             string     `json:"status" db:"status"` // active, disconnected, idle
	ConnectTime        *time.Time `json:"connect_time,omitempty" db:"connect_time"`
	DisconnectTime     *time.Time `json:"disconnect_time,omitempty" db:"disconnect_time"`
	DisconnectCause    *string    `json:"disconnect_cause,omitempty" db:"disconnect_cause"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// NATSession represents an active NAT translation session
type NATSession struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	TenantID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RouterID              uuid.UUID  `json:"router_id" db:"router_id"`
	Protocol              string     `json:"protocol" db:"protocol"` // tcp, udp, icmp
	SrcAddress            net.IP     `json:"src_address" db:"src_address"`
	SrcPort               *int       `json:"src_port,omitempty" db:"src_port"`
	DstAddress            net.IP     `json:"dst_address" db:"dst_address"`
	DstPort               *int       `json:"dst_port,omitempty" db:"dst_port"`
	TranslatedSrcAddress  *net.IP    `json:"translated_src_address,omitempty" db:"translated_src_address"`
	TranslatedSrcPort     *int       `json:"translated_src_port,omitempty" db:"translated_src_port"`
	State                 *string    `json:"state,omitempty" db:"state"` // established, syn_sent, etc.
	Bytes                 int64      `json:"bytes" db:"bytes"`
	Packets               int64      `json:"packets" db:"packets"`
	TimeoutSeconds        *int       `json:"timeout_seconds,omitempty" db:"timeout_seconds"`
	EstablishedAt         *time.Time `json:"established_at,omitempty" db:"established_at"`
	LastSeenAt            *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	Status                string     `json:"status" db:"status"` // active, closed
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
}

// DHCPLease represents a DHCP lease assignment
type DHCPLease struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RouterID    uuid.UUID  `json:"router_id" db:"router_id"`
	MACAddress  string     `json:"mac_address" db:"mac_address"`
	IPAddress   net.IP     `json:"ip_address" db:"ip_address"`
	Hostname    *string    `json:"hostname,omitempty" db:"hostname"`
	LeaseStart  time.Time  `json:"lease_start" db:"lease_start"`
	LeaseEnd    time.Time  `json:"lease_end" db:"lease_end"`
	LeaseState  *string    `json:"lease_state,omitempty" db:"lease_state"` // active, expired, released, offered
	DHCPPool    *string    `json:"dhcp_pool,omitempty" db:"dhcp_pool"`
	ClientID    *string    `json:"client_id,omitempty" db:"client_id"`
	VendorClass *string    `json:"vendor_class,omitempty" db:"vendor_class"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// RoleSpecificMetric represents flexible storage for role-specific metrics
type RoleSpecificMetric struct {
	ID        int64              `json:"id" db:"id"`
	TenantID  uuid.UUID          `json:"tenant_id" db:"tenant_id"`
	RouterID  uuid.UUID          `json:"router_id" db:"router_id"`
	RoleCode  string             `json:"role_code" db:"role_code"`
	Timestamp time.Time          `json:"timestamp" db:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics" db:"metrics"` // JSONB
	CreatedAt time.Time          `json:"created_at" db:"created_at"`
}

// PollingHistory tracks polling attempts and their results
type PollingHistory struct {
	ID                int64      `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RouterID          uuid.UUID  `json:"router_id" db:"router_id"`
	PollStartedAt     time.Time  `json:"poll_started_at" db:"poll_started_at"`
	PollCompletedAt   *time.Time `json:"poll_completed_at,omitempty" db:"poll_completed_at"`
	AdapterUsed       *string    `json:"adapter_used,omitempty" db:"adapter_used"` // snmp, mikrotik_api, etc.
	Success           bool       `json:"success" db:"success"`
	ErrorMessage      *string    `json:"error_message,omitempty" db:"error_message"`
	MetricsCollected  int        `json:"metrics_collected" db:"metrics_collected"`
	ResponseTimeMs    *int       `json:"response_time_ms,omitempty" db:"response_time_ms"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// PPPoESessionSummary represents aggregated PPPoE session statistics
type PPPoESessionSummary struct {
	RouterID         uuid.UUID `json:"router_id" db:"router_id"`
	TotalSessions    int       `json:"total_sessions" db:"total_sessions"`
	ActiveSessions   int       `json:"active_sessions" db:"active_sessions"`
	TotalBytesIn     int64     `json:"total_bytes_in" db:"total_bytes_in"`
	TotalBytesOut    int64     `json:"total_bytes_out" db:"total_bytes_out"`
	AvgSessionTime   *float64  `json:"avg_session_time,omitempty" db:"avg_session_time"`
}

// NATSessionSummary represents aggregated NAT session statistics
type NATSessionSummary struct {
	RouterID     uuid.UUID `json:"router_id" db:"router_id"`
	Protocol     string    `json:"protocol" db:"protocol"`
	SessionCount int       `json:"session_count" db:"session_count"`
	TotalBytes   int64     `json:"total_bytes" db:"total_bytes"`
	TotalPackets int64     `json:"total_packets" db:"total_packets"`
}

// Session status constants
const (
	SessionStatusActive       = "active"
	SessionStatusDisconnected = "disconnected"
	SessionStatusIdle         = "idle"
	SessionStatusClosed       = "closed"
)

// DHCP lease state constants
const (
	LeaseStateActive   = "active"
	LeaseStateExpired  = "expired"
	LeaseStateReleased = "released"
	LeaseStateOffered  = "offered"
)

// NAT protocols
const (
	NATProtocolTCP  = "tcp"
	NATProtocolUDP  = "udp"
	NATProtocolICMP = "icmp"
)

// GetTotalBytes returns total bytes transferred in the session
func (p *PPPoESession) GetTotalBytes() int64 {
	return p.BytesIn + p.BytesOut
}

// GetTotalPackets returns total packets transferred in the session
func (p *PPPoESession) GetTotalPackets() int64 {
	return p.PacketsIn + p.PacketsOut
}

// IsActive returns true if the session is currently active
func (p *PPPoESession) IsActive() bool {
	return p.Status == SessionStatusActive
}

// IsExpired returns true if the DHCP lease has expired
func (d *DHCPLease) IsExpired() bool {
	return time.Now().After(d.LeaseEnd)
}

// TimeRemaining returns the duration until the lease expires
func (d *DHCPLease) TimeRemaining() time.Duration {
	if d.IsExpired() {
		return 0
	}
	return time.Until(d.LeaseEnd)
}
