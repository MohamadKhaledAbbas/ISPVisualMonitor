package dto

import "github.com/google/uuid"

// TopologyResponse represents the full network topology
type TopologyResponse struct {
	Routers []RouterNode   `json:"routers"`
	Links   []TopologyLink `json:"links"`
}

// RouterNode represents a router node in the topology
type RouterNode struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ManagementIP string    `json:"management_ip"`
	Status       string    `json:"status"`
	Vendor       *string   `json:"vendor,omitempty"`
	Model        *string   `json:"model,omitempty"`
	Location     *GeoPoint `json:"location,omitempty"`
}

// TopologyLink represents a link between interfaces
type TopologyLink struct {
	ID                uuid.UUID `json:"id"`
	Name              *string   `json:"name,omitempty"`
	SourceInterfaceID uuid.UUID `json:"source_interface_id"`
	TargetInterfaceID uuid.UUID `json:"target_interface_id"`
	SourceRouterID    uuid.UUID `json:"source_router_id"`
	TargetRouterID    uuid.UUID `json:"target_router_id"`
	LinkType          string    `json:"link_type"`
	CapacityMbps      *int64    `json:"capacity_mbps,omitempty"`
	LatencyMs         *float64  `json:"latency_ms,omitempty"`
	Status            string    `json:"status"`
}

// GeoJSONResponse represents topology as GeoJSON
type GeoJSONResponse struct {
	Type     string       `json:"type"`
	Features []GeoFeature `json:"features"`
}

// GeoFeature represents a GeoJSON feature
type GeoFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoGeometry            `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoGeometry represents GeoJSON geometry
type GeoGeometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}
