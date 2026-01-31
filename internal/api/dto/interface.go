package dto

import (
	"time"

	"github.com/google/uuid"
)

// InterfaceDTO represents a network interface in API responses
type InterfaceDTO struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	RouterID    uuid.UUID `json:"router_id"`
	RouterName  string    `json:"router_name,omitempty"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IfIndex     *int      `json:"if_index,omitempty"`
	IfType      *string   `json:"if_type,omitempty"`
	SpeedMbps   *int64    `json:"speed_mbps,omitempty"`
	MTU         *int      `json:"mtu,omitempty"`
	MACAddress  *string   `json:"mac_address,omitempty"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	SubnetMask  *string   `json:"subnet_mask,omitempty"`
	Status      string    `json:"status"`
	AdminStatus string    `json:"admin_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
