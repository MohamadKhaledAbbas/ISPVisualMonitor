package models

import (
	"time"

	"github.com/google/uuid"
)

// RouterRole represents a standard router role type
type RouterRole struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Code        string     `json:"code" db:"code"` // core_router, pppoe_server, etc.
	Description *string    `json:"description,omitempty" db:"description"`
	Category    *string    `json:"category,omitempty" db:"category"` // routing, access, security, management
	Icon        *string    `json:"icon,omitempty" db:"icon"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// RouterRoleAssignment represents a router's assigned role
type RouterRoleAssignment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	RouterID   uuid.UUID  `json:"router_id" db:"router_id"`
	RoleID     uuid.UUID  `json:"role_id" db:"role_id"`
	Priority   int        `json:"priority" db:"priority"`       // Lower = higher priority
	IsPrimary  bool       `json:"is_primary" db:"is_primary"`   // Mark one role as primary
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty" db:"assigned_by"`
	Notes      *string    `json:"notes,omitempty" db:"notes"`
	
	// Expanded fields (from joins)
	Role *RouterRole `json:"role,omitempty" db:"-"`
}

// RouterDependency represents a relationship between two routers
type RouterDependency struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	SourceRouterID uuid.UUID  `json:"source_router_id" db:"source_router_id"`
	TargetRouterID uuid.UUID  `json:"target_router_id" db:"target_router_id"`
	DependencyType string     `json:"dependency_type" db:"dependency_type"` // upstream, downstream, peer, failover, backup
	Weight         int        `json:"weight" db:"weight"`                   // For load balancing or priority
	Description    *string    `json:"description,omitempty" db:"description"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	
	// Expanded fields (from joins)
	SourceRouter *Router `json:"source_router,omitempty" db:"-"`
	TargetRouter *Router `json:"target_router,omitempty" db:"-"`
}

// Standard router role codes
const (
	RoleCodeCoreRouter        = "core_router"
	RoleCodeEdgeRouter        = "edge_router"
	RoleCodeBorderRouter      = "border_router"
	RoleCodeAccessRouter      = "access_router"
	RoleCodeNATGateway        = "nat_gateway"
	RoleCodePPPoEServer       = "pppoe_server"
	RoleCodeDHCPServer        = "dhcp_server"
	RoleCodeVPNGateway        = "vpn_gateway"
	RoleCodeLoadBalancer      = "load_balancer"
	RoleCodeBandwidthShaper   = "bandwidth_shaper"
	RoleCodeAccessController  = "access_controller"
	RoleCodeFirewall          = "firewall"
)

// Dependency type constants
const (
	DependencyTypeUpstream   = "upstream"
	DependencyTypeDownstream = "downstream"
	DependencyTypePeer       = "peer"
	DependencyTypeFailover   = "failover"
	DependencyTypeBackup     = "backup"
)

// IsSessionTrackingRole returns true if the role requires session tracking
func IsSessionTrackingRole(roleCode string) bool {
	switch roleCode {
	case RoleCodePPPoEServer, RoleCodeNATGateway, RoleCodeDHCPServer:
		return true
	default:
		return false
	}
}

// GetRoleMetricTypes returns the types of metrics that should be collected for a role
func GetRoleMetricTypes(roleCode string) []string {
	switch roleCode {
	case RoleCodePPPoEServer:
		return []string{"pppoe_sessions", "pppoe_throughput", "authentication_failures"}
	case RoleCodeNATGateway:
		return []string{"nat_sessions", "nat_pool_utilization", "port_exhaustion"}
	case RoleCodeDHCPServer:
		return []string{"dhcp_leases", "dhcp_pool_utilization", "lease_conflicts"}
	case RoleCodeVPNGateway:
		return []string{"vpn_tunnels", "vpn_throughput", "tunnel_failures"}
	case RoleCodeLoadBalancer:
		return []string{"backend_health", "connection_distribution", "session_persistence"}
	case RoleCodeBandwidthShaper:
		return []string{"queue_depth", "dropped_packets", "shaped_flows"}
	case RoleCodeFirewall:
		return []string{"blocked_connections", "rule_hits", "threat_detections"}
	default:
		return []string{"interface_metrics", "cpu_usage", "memory_usage"}
	}
}
