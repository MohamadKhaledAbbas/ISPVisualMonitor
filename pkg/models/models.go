package models

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents an ISP organization
type Tenant struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Slug             string     `json:"slug" db:"slug"`
	ContactEmail     string     `json:"contact_email" db:"contact_email"`
	ContactPhone     *string    `json:"contact_phone,omitempty" db:"contact_phone"`
	SubscriptionTier string     `json:"subscription_tier" db:"subscription_tier"`
	MaxDevices       int        `json:"max_devices" db:"max_devices"`
	MaxUsers         int        `json:"max_users" db:"max_users"`
	Status           string     `json:"status" db:"status"`
	TrialEndsAt      *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	FirstName     *string    `json:"first_name,omitempty" db:"first_name"`
	LastName      *string    `json:"last_name,omitempty" db:"last_name"`
	Status        string     `json:"status" db:"status"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Role represents a role with permissions
type Role struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	IsCustom    bool       `json:"is_custom" db:"is_custom"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Router represents a network router
type Router struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	TenantID               uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	POPID                  *uuid.UUID `json:"pop_id,omitempty" db:"pop_id"`
	Name                   string     `json:"name" db:"name"`
	Hostname               *string    `json:"hostname,omitempty" db:"hostname"`
	ManagementIP           string     `json:"management_ip" db:"management_ip"`
	Location               *string    `json:"location,omitempty" db:"location"` // PostGIS POINT as string
	RouterType             string     `json:"router_type" db:"router_type"`
	Vendor                 *string    `json:"vendor,omitempty" db:"vendor"`
	Model                  *string    `json:"model,omitempty" db:"model"`
	OSVersion              *string    `json:"os_version,omitempty" db:"os_version"`
	SerialNumber           *string    `json:"serial_number,omitempty" db:"serial_number"`
	Status                 string     `json:"status" db:"status"`
	PollingEnabled         bool       `json:"polling_enabled" db:"polling_enabled"`
	PollingIntervalSeconds int        `json:"polling_interval_seconds" db:"polling_interval_seconds"`
	SNMPVersion            string     `json:"snmp_version" db:"snmp_version"`
	SNMPCommunity          *string    `json:"-" db:"snmp_community"` // Sensitive
	SNMPPort               int        `json:"snmp_port" db:"snmp_port"`
	Description            *string    `json:"description,omitempty" db:"description"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	LastPolledAt           *time.Time `json:"last_polled_at,omitempty" db:"last_polled_at"`
}

// Interface represents a router interface
type Interface struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	RouterID    uuid.UUID `json:"router_id" db:"router_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	IfIndex     *int      `json:"if_index,omitempty" db:"if_index"`
	IfType      *string   `json:"if_type,omitempty" db:"if_type"`
	SpeedMbps   *int64    `json:"speed_mbps,omitempty" db:"speed_mbps"`
	MTU         *int      `json:"mtu,omitempty" db:"mtu"`
	MACAddress  *string   `json:"mac_address,omitempty" db:"mac_address"`
	IPAddress   *string   `json:"ip_address,omitempty" db:"ip_address"`
	SubnetMask  *string   `json:"subnet_mask,omitempty" db:"subnet_mask"`
	Status      string    `json:"status" db:"status"`
	AdminStatus string    `json:"admin_status" db:"admin_status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Link represents a connection between two interfaces
type Link struct {
	ID                uuid.UUID `json:"id" db:"id"`
	TenantID          uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name              *string   `json:"name,omitempty" db:"name"`
	SourceInterfaceID uuid.UUID `json:"source_interface_id" db:"source_interface_id"`
	TargetInterfaceID uuid.UUID `json:"target_interface_id" db:"target_interface_id"`
	LinkType          string    `json:"link_type" db:"link_type"`
	CapacityMbps      *int64    `json:"capacity_mbps,omitempty" db:"capacity_mbps"`
	LatencyMs         *float64  `json:"latency_ms,omitempty" db:"latency_ms"`
	Status            string    `json:"status" db:"status"`
	PathGeometry      *string   `json:"path_geometry,omitempty" db:"path_geometry"` // PostGIS LINESTRING
	Description       *string   `json:"description,omitempty" db:"description"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// POP represents a Point of Presence
type POP struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RegionID     *uuid.UUID `json:"region_id,omitempty" db:"region_id"`
	Name         string     `json:"name" db:"name"`
	Code         *string    `json:"code,omitempty" db:"code"`
	Location     string     `json:"location" db:"location"` // PostGIS POINT as string
	Address      *string    `json:"address,omitempty" db:"address"`
	POPType      *string    `json:"pop_type,omitempty" db:"pop_type"`
	CapacityGbps *int       `json:"capacity_gbps,omitempty" db:"capacity_gbps"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Alert represents an alert
type Alert struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RuleID         *uuid.UUID `json:"rule_id,omitempty" db:"rule_id"`
	Name           string     `json:"name" db:"name"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Severity       string     `json:"severity" db:"severity"`
	Status         string     `json:"status" db:"status"`
	TargetType     *string    `json:"target_type,omitempty" db:"target_type"`
	TargetID       *uuid.UUID `json:"target_id,omitempty" db:"target_id"`
	TriggeredAt    time.Time  `json:"triggered_at" db:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AcknowledgedBy *uuid.UUID `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata       *string    `json:"metadata,omitempty" db:"metadata"` // JSONB as string
}
