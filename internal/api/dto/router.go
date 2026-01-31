package dto

import (
	"time"

	"github.com/google/uuid"
)

// GeoPoint represents geographic coordinates
type GeoPoint struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// RouterDTO represents a router in API responses
type RouterDTO struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Name         string     `json:"name"`
	Hostname     *string    `json:"hostname,omitempty"`
	ManagementIP string     `json:"management_ip"`
	Vendor       *string    `json:"vendor,omitempty"`
	Model        *string    `json:"model,omitempty"`
	OSVersion    *string    `json:"os_version,omitempty"`
	Status       string     `json:"status"`
	Location     *GeoPoint  `json:"location,omitempty"`
	POPID        *uuid.UUID `json:"pop_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateRouterRequest represents the request to create a router
type CreateRouterRequest struct {
	Name          string    `json:"name" validate:"required"`
	Hostname      *string   `json:"hostname,omitempty"`
	ManagementIP  string    `json:"management_ip" validate:"required,ip"`
	Vendor        string    `json:"vendor" validate:"required"`
	Model         *string   `json:"model,omitempty"`
	POPID         *string   `json:"pop_id,omitempty"`
	Location      *GeoPoint `json:"location,omitempty"`
	SNMPVersion   *string   `json:"snmp_version,omitempty"`
	SNMPCommunity *string   `json:"snmp_community,omitempty"`
}

// UpdateRouterRequest represents the request to update a router
type UpdateRouterRequest struct {
	Name         *string   `json:"name,omitempty"`
	Hostname     *string   `json:"hostname,omitempty"`
	ManagementIP *string   `json:"management_ip,omitempty" validate:"omitempty,ip"`
	Vendor       *string   `json:"vendor,omitempty"`
	Model        *string   `json:"model,omitempty"`
	Status       *string   `json:"status,omitempty"`
	Location     *GeoPoint `json:"location,omitempty"`
	POPID        *string   `json:"pop_id,omitempty"`
}
