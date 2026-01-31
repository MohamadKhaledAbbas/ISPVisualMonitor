// Package license provides license validation for ISP Visual Monitor.
package license

import (
	"context"
	"time"
)

// LicenseValidator defines the interface for license validation
type LicenseValidator interface {
	// Validate validates a license key and returns license information
	Validate(ctx context.Context, licenseKey string) (*LicenseInfo, error)
	// GetFeatures returns the features enabled by the current license
	GetFeatures(ctx context.Context) (*Features, error)
	// IsValid returns true if the current license is valid
	IsValid() bool
	// GetInfo returns the current license information
	GetInfo() *LicenseInfo
}

// LicenseInfo contains license details
type LicenseInfo struct {
	LicenseID    string    `json:"license_id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Plan         string    `json:"plan"` // starter, professional, enterprise
	MaxRouters   int       `json:"max_routers"`
	MaxUsers     int       `json:"max_users"`
	ExpiresAt    time.Time `json:"expires_at"`
	IssuedAt     time.Time `json:"issued_at"`
	Features     Features  `json:"features"`
}

// Features defines the available license features
type Features struct {
	MultiTenant      bool `json:"multi_tenant"`
	AdvancedAlerts   bool `json:"advanced_alerts"`
	CustomDashboards bool `json:"custom_dashboards"`
	APIAccess        bool `json:"api_access"`
	SSO              bool `json:"sso"`
	AuditLog         bool `json:"audit_log"`
	PrioritySupport  bool `json:"priority_support"`
}

// Plan represents license plan tiers
type Plan string

const (
	// PlanStarter is the basic plan
	PlanStarter Plan = "starter"
	// PlanProfessional is the mid-tier plan
	PlanProfessional Plan = "professional"
	// PlanEnterprise is the top-tier plan
	PlanEnterprise Plan = "enterprise"
)

// PlanLimits defines the limits for each plan
var PlanLimits = map[Plan]struct {
	MaxRouters int
	MaxUsers   int
}{
	PlanStarter:      {MaxRouters: 10, MaxUsers: 5},
	PlanProfessional: {MaxRouters: 100, MaxUsers: 25},
	PlanEnterprise:   {MaxRouters: -1, MaxUsers: -1}, // unlimited
}

// DefaultFeatures returns the default features for a plan
func DefaultFeatures(plan Plan) Features {
	switch plan {
	case PlanEnterprise:
		return Features{
			MultiTenant:      true,
			AdvancedAlerts:   true,
			CustomDashboards: true,
			APIAccess:        true,
			SSO:              true,
			AuditLog:         true,
			PrioritySupport:  true,
		}
	case PlanProfessional:
		return Features{
			MultiTenant:      false,
			AdvancedAlerts:   true,
			CustomDashboards: true,
			APIAccess:        true,
			SSO:              false,
			AuditLog:         true,
			PrioritySupport:  false,
		}
	default: // PlanStarter
		return Features{
			MultiTenant:      false,
			AdvancedAlerts:   false,
			CustomDashboards: false,
			APIAccess:        true,
			SSO:              false,
			AuditLog:         false,
			PrioritySupport:  false,
		}
	}
}

// IsExpired returns true if the license has expired
func (l *LicenseInfo) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// DaysUntilExpiry returns the number of days until the license expires
func (l *LicenseInfo) DaysUntilExpiry() int {
	if l.IsExpired() {
		return 0
	}
	return int(time.Until(l.ExpiresAt).Hours() / 24)
}

// CanAddRouter returns true if another router can be added
func (l *LicenseInfo) CanAddRouter(currentCount int) bool {
	if l.MaxRouters < 0 {
		return true // unlimited
	}
	return currentCount < l.MaxRouters
}

// CanAddUser returns true if another user can be added
func (l *LicenseInfo) CanAddUser(currentCount int) bool {
	if l.MaxUsers < 0 {
		return true // unlimited
	}
	return currentCount < l.MaxUsers
}
