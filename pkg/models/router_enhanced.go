package models

import (
	"github.com/google/uuid"
)

// EnhancedRouter extends the base Router with capabilities and roles
type EnhancedRouter struct {
	Router                                    // Embedded base router
	Capabilities *RouterCapabilities          `json:"capabilities,omitempty"`
	Roles        []RouterRoleAssignment       `json:"roles,omitempty"`
	Dependencies []RouterDependency           `json:"dependencies,omitempty"`
}

// GetPrimaryRole returns the primary role assignment for this router
func (er *EnhancedRouter) GetPrimaryRole() *RouterRoleAssignment {
	for i := range er.Roles {
		if er.Roles[i].IsPrimary {
			return &er.Roles[i]
		}
	}
	// If no primary, return first role by priority
	if len(er.Roles) > 0 {
		return &er.Roles[0]
	}
	return nil
}

// GetRoleCodes returns a list of role codes assigned to this router
func (er *EnhancedRouter) GetRoleCodes() []string {
	codes := make([]string, 0, len(er.Roles))
	for _, role := range er.Roles {
		if role.Role != nil {
			codes = append(codes, role.Role.Code)
		}
	}
	return codes
}

// HasRole checks if router has a specific role
func (er *EnhancedRouter) HasRole(roleCode string) bool {
	for _, role := range er.Roles {
		if role.Role != nil && role.Role.Code == roleCode {
			return true
		}
	}
	return false
}

// RequiresSessionTracking returns true if any of the router's roles require session tracking
func (er *EnhancedRouter) RequiresSessionTracking() bool {
	for _, role := range er.Roles {
		if role.Role != nil && IsSessionTrackingRole(role.Role.Code) {
			return true
		}
	}
	return false
}

// GetUpstreamRouters returns routers that this router depends on (upstream)
func (er *EnhancedRouter) GetUpstreamRouters() []uuid.UUID {
	upstreams := []uuid.UUID{}
	for _, dep := range er.Dependencies {
		if dep.DependencyType == DependencyTypeUpstream && dep.IsActive {
			upstreams = append(upstreams, dep.TargetRouterID)
		}
	}
	return upstreams
}

// GetDownstreamRouters returns routers that depend on this router (downstream)
func (er *EnhancedRouter) GetDownstreamRouters() []uuid.UUID {
	downstreams := []uuid.UUID{}
	for _, dep := range er.Dependencies {
		if dep.DependencyType == DependencyTypeDownstream && dep.IsActive {
			downstreams = append(downstreams, dep.TargetRouterID)
		}
	}
	return downstreams
}

// GetPeerRouters returns peer routers
func (er *EnhancedRouter) GetPeerRouters() []uuid.UUID {
	peers := []uuid.UUID{}
	for _, dep := range er.Dependencies {
		if dep.DependencyType == DependencyTypePeer && dep.IsActive {
			peers = append(peers, dep.TargetRouterID)
		}
	}
	return peers
}

// GetFailoverRouters returns failover routers
func (er *EnhancedRouter) GetFailoverRouters() []uuid.UUID {
	failovers := []uuid.UUID{}
	for _, dep := range er.Dependencies {
		if dep.DependencyType == DependencyTypeFailover && dep.IsActive {
			failovers = append(failovers, dep.TargetRouterID)
		}
	}
	return failovers
}

// CanPoll returns true if the router has at least one active polling method configured
func (er *EnhancedRouter) CanPoll() bool {
	if er.Capabilities == nil {
		return false
	}
	return er.Capabilities.HasActivePollingMethod()
}

// GetPreferredPollingMethod returns the preferred polling method for this router
func (er *EnhancedRouter) GetPreferredPollingMethod() string {
	if er.Capabilities == nil {
		return ""
	}
	return er.Capabilities.PreferredMethod
}

// GetPollingMethodOrder returns the order of polling methods to try
func (er *EnhancedRouter) GetPollingMethodOrder() []string {
	if er.Capabilities == nil {
		return []string{}
	}
	return er.Capabilities.GetPreferredMethodOrder()
}

// GetMetricTypesToCollect returns metric types based on assigned roles
func (er *EnhancedRouter) GetMetricTypesToCollect() []string {
	metricsMap := make(map[string]bool)
	
	// Always collect basic metrics
	metricsMap["interface_metrics"] = true
	metricsMap["cpu_usage"] = true
	metricsMap["memory_usage"] = true
	
	// Add role-specific metrics
	for _, role := range er.Roles {
		if role.Role != nil {
			roleMetrics := GetRoleMetricTypes(role.Role.Code)
			for _, metric := range roleMetrics {
				metricsMap[metric] = true
			}
		}
	}
	
	// Convert map to slice
	metrics := make([]string, 0, len(metricsMap))
	for metric := range metricsMap {
		metrics = append(metrics, metric)
	}
	
	return metrics
}

// RouterWithRolesSummary provides a lightweight representation with role information
type RouterWithRolesSummary struct {
	Router
	RoleCodes        []string `json:"role_codes"`
	RoleNames        []string `json:"role_names"`
	PrimaryRoleCode  *string  `json:"primary_role_code,omitempty"`
	HasCapabilities  bool     `json:"has_capabilities"`
	CanPoll          bool     `json:"can_poll"`
	PreferredMethod  *string  `json:"preferred_method,omitempty"`
}
