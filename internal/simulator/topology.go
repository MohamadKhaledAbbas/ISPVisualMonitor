package simulator

import (
	"github.com/google/uuid"
)

// Deterministic UUIDs for the demo topology (matches demo_seed.sql).
var (
	DemoTenantID = uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	// Router IDs (match demo_seed.sql)
	RouterBEYCORE01    = uuid.MustParse("e0000000-0000-0000-0000-000000000001")
	RouterBEYCORE02    = uuid.MustParse("e0000000-0000-0000-0000-000000000002")
	RouterBEYEDGE01    = uuid.MustParse("e0000000-0000-0000-0000-000000000003")
	RouterBEYUPSTREAM  = uuid.MustParse("e0000000-0000-0000-0000-000000000004")
	RouterBEYPPPOE01   = uuid.MustParse("e0000000-0000-0000-0000-000000000005")
	RouterTRIEDGE01    = uuid.MustParse("e0000000-0000-0000-0000-000000000006")
	RouterTRIACCESS01  = uuid.MustParse("e0000000-0000-0000-0000-000000000007")
	RouterTRIPPPOE01   = uuid.MustParse("e0000000-0000-0000-0000-000000000008")
	RouterSIDEDGE01    = uuid.MustParse("e0000000-0000-0000-0000-000000000009")
	RouterSIDPPPOE01   = uuid.MustParse("e0000000-0000-0000-0000-00000000000a")
)

// DeviceRole describes the role of a simulated router in the topology.
type DeviceRole string

const (
	RoleCore     DeviceRole = "core"
	RoleEdge     DeviceRole = "edge"
	RoleBorder   DeviceRole = "border"
	RoleAccess   DeviceRole = "access"
	RolePPPoE    DeviceRole = "pppoe"
	RoleUpstream DeviceRole = "upstream"
)

// SimDevice represents a simulated network device in the reference topology.
type SimDevice struct {
	ID           uuid.UUID
	Name         string
	Role         DeviceRole
	ManagementIP string
	Site         string
	Vendor       string
	Model        string

	// Baseline metrics (healthy state)
	BaselineCPU      float64 // percent
	BaselineMemory   float64 // percent
	BaselineTemp     float64 // celsius
	UptimeSeconds    int64
	MaxPPPoESessions int // 0 for non-PPPoE routers

	// Interfaces belonging to this device
	Interfaces []SimInterface
}

// SimInterface represents a simulated network interface.
type SimInterface struct {
	ID          uuid.UUID
	Name        string
	Description string
	SpeedMbps   int64
	Status      string // up, down
	AdminStatus string // up, down

	// Baseline throughput (healthy)
	BaselineInMbps  float64
	BaselineOutMbps float64
}

// SimLink represents a simulated link between two interfaces.
type SimLink struct {
	ID                uuid.UUID
	Name              string
	SourceInterfaceID uuid.UUID
	TargetInterfaceID uuid.UUID
	CapacityMbps      int64
}

// Topology holds the complete reference ISP topology for simulation.
type Topology struct {
	TenantID uuid.UUID
	Devices  []SimDevice
	Links    []SimLink
}

// DefaultTopology returns the reference small-ISP topology matching the demo seed data.
// It defines 10 routers across 3 sites with realistic interface configurations.
func DefaultTopology() *Topology {
	return &Topology{
		TenantID: DemoTenantID,
		Devices: []SimDevice{
			// === Beirut DC (BEY-DC1) ===
			{
				ID: RouterBEYCORE01, Name: "BEY-CORE-01", Role: RoleCore,
				ManagementIP: "10.255.0.1", Site: "BEY-DC1",
				Vendor: "MikroTik", Model: "CCR2216-1G-12XS-2XQ",
				BaselineCPU: 30, BaselineMemory: 55, BaselineTemp: 42, UptimeSeconds: 2592000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0001-000000000001"), Name: "ether1-mgmt", Description: "Management", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 5, BaselineOutMbps: 5},
					{ID: uuid.MustParse("f0000000-0000-0000-0001-000000000002"), Name: "sfp1-core02", Description: "To BEY-CORE-02", SpeedMbps: 25000, Status: "up", AdminStatus: "up", BaselineInMbps: 8000, BaselineOutMbps: 6000},
					{ID: uuid.MustParse("f0000000-0000-0000-0001-000000000003"), Name: "sfp2-uplink", Description: "To BEY-UPSTREAM-01", SpeedMbps: 25000, Status: "up", AdminStatus: "up", BaselineInMbps: 12000, BaselineOutMbps: 8000},
				},
			},
			{
				ID: RouterBEYCORE02, Name: "BEY-CORE-02", Role: RoleCore,
				ManagementIP: "10.255.0.2", Site: "BEY-DC1",
				Vendor: "MikroTik", Model: "CCR2216-1G-12XS-2XQ",
				BaselineCPU: 28, BaselineMemory: 52, BaselineTemp: 40, UptimeSeconds: 2592000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0002-000000000001"), Name: "ether1-mgmt", Description: "Management", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 5, BaselineOutMbps: 5},
					{ID: uuid.MustParse("f0000000-0000-0000-0002-000000000002"), Name: "sfp1-core01", Description: "To BEY-CORE-01", SpeedMbps: 25000, Status: "up", AdminStatus: "up", BaselineInMbps: 6000, BaselineOutMbps: 8000},
					{ID: uuid.MustParse("f0000000-0000-0000-0002-000000000003"), Name: "sfp2-edge", Description: "To BEY-EDGE-01", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 4000, BaselineOutMbps: 3000},
				},
			},
			{
				ID: RouterBEYEDGE01, Name: "BEY-EDGE-01", Role: RoleEdge,
				ManagementIP: "10.255.0.3", Site: "BEY-DC1",
				Vendor: "MikroTik", Model: "CCR2116-12G-4S+",
				BaselineCPU: 40, BaselineMemory: 60, BaselineTemp: 45, UptimeSeconds: 1296000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0003-000000000001"), Name: "ether1-mgmt", Description: "Management", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 5, BaselineOutMbps: 5},
					{ID: uuid.MustParse("f0000000-0000-0000-0003-000000000002"), Name: "sfp1-core02", Description: "To BEY-CORE-02", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 3000, BaselineOutMbps: 4000},
					{ID: uuid.MustParse("f0000000-0000-0000-0003-000000000003"), Name: "sfp2-pppoe", Description: "To BEY-PPPOE-01", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 2500, BaselineOutMbps: 3500},
				},
			},
			{
				ID: RouterBEYUPSTREAM, Name: "BEY-UPSTREAM-01", Role: RoleUpstream,
				ManagementIP: "10.255.0.4", Site: "BEY-DC1",
				Vendor: "MikroTik", Model: "CCR1072-1G-8S+",
				BaselineCPU: 22, BaselineMemory: 40, BaselineTemp: 38, UptimeSeconds: 5184000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0004-000000000001"), Name: "ether1-isp-a", Description: "ISP-A Upstream", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 7000, BaselineOutMbps: 3000},
					{ID: uuid.MustParse("f0000000-0000-0000-0004-000000000002"), Name: "ether2-isp-b", Description: "ISP-B Upstream", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 5000, BaselineOutMbps: 2000},
					{ID: uuid.MustParse("f0000000-0000-0000-0004-000000000003"), Name: "sfp1-core01", Description: "To BEY-CORE-01", SpeedMbps: 25000, Status: "up", AdminStatus: "up", BaselineInMbps: 8000, BaselineOutMbps: 12000},
				},
			},
			{
				ID: RouterBEYPPPOE01, Name: "BEY-PPPOE-01", Role: RolePPPoE,
				ManagementIP: "10.255.0.5", Site: "BEY-DC1",
				Vendor: "MikroTik", Model: "CHR",
				BaselineCPU: 55, BaselineMemory: 68, BaselineTemp: 50, UptimeSeconds: 864000,
				MaxPPPoESessions: 1000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0005-000000000001"), Name: "ether1-edge", Description: "To BEY-EDGE-01", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 3500, BaselineOutMbps: 2500},
					{ID: uuid.MustParse("f0000000-0000-0000-0005-000000000002"), Name: "ether2-subs", Description: "Subscriber VLAN", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 2000, BaselineOutMbps: 3000},
				},
			},
			// === Tripoli POP (TRI-POP1) ===
			{
				ID: RouterTRIEDGE01, Name: "TRI-EDGE-01", Role: RoleEdge,
				ManagementIP: "10.255.1.1", Site: "TRI-POP1",
				Vendor: "MikroTik", Model: "CCR2116-12G-4S+",
				BaselineCPU: 35, BaselineMemory: 55, BaselineTemp: 44, UptimeSeconds: 1728000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0006-000000000001"), Name: "sfp1-core", Description: "To BEY-CORE-01", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 2500, BaselineOutMbps: 1500},
					{ID: uuid.MustParse("f0000000-0000-0000-0006-000000000002"), Name: "ether1-access", Description: "To TRI-ACCESS-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 400, BaselineOutMbps: 600},
				},
			},
			{
				ID: RouterTRIACCESS01, Name: "TRI-ACCESS-01", Role: RoleAccess,
				ManagementIP: "10.255.1.2", Site: "TRI-POP1",
				Vendor: "MikroTik", Model: "CRS326-24G-2S+",
				BaselineCPU: 25, BaselineMemory: 45, BaselineTemp: 38, UptimeSeconds: 2160000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0007-000000000001"), Name: "ether1-edge", Description: "To TRI-EDGE-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 600, BaselineOutMbps: 400},
					{ID: uuid.MustParse("f0000000-0000-0000-0007-000000000002"), Name: "ether2-pppoe", Description: "To TRI-PPPOE-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 300, BaselineOutMbps: 500},
				},
			},
			{
				ID: RouterTRIPPPOE01, Name: "TRI-PPPOE-01", Role: RolePPPoE,
				ManagementIP: "10.255.1.3", Site: "TRI-POP1",
				Vendor: "MikroTik", Model: "CHR",
				BaselineCPU: 48, BaselineMemory: 62, BaselineTemp: 46, UptimeSeconds: 1296000,
				MaxPPPoESessions: 500,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0008-000000000001"), Name: "ether1-access", Description: "To TRI-ACCESS-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 500, BaselineOutMbps: 300},
					{ID: uuid.MustParse("f0000000-0000-0000-0008-000000000002"), Name: "ether2-subs", Description: "Subscriber VLAN", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 250, BaselineOutMbps: 400},
				},
			},
			// === Sidon POP (SID-POP1) ===
			{
				ID: RouterSIDEDGE01, Name: "SID-EDGE-01", Role: RoleEdge,
				ManagementIP: "10.255.2.1", Site: "SID-POP1",
				Vendor: "MikroTik", Model: "CCR2116-12G-4S+",
				BaselineCPU: 30, BaselineMemory: 50, BaselineTemp: 42, UptimeSeconds: 1900000,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-0009-000000000001"), Name: "sfp1-core", Description: "To BEY-CORE-02", SpeedMbps: 10000, Status: "up", AdminStatus: "up", BaselineInMbps: 2000, BaselineOutMbps: 1200},
					{ID: uuid.MustParse("f0000000-0000-0000-0009-000000000002"), Name: "ether1-pppoe", Description: "To SID-PPPOE-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 350, BaselineOutMbps: 550},
				},
			},
			{
				ID: RouterSIDPPPOE01, Name: "SID-PPPOE-01", Role: RolePPPoE,
				ManagementIP: "10.255.2.2", Site: "SID-POP1",
				Vendor: "MikroTik", Model: "CHR",
				BaselineCPU: 42, BaselineMemory: 58, BaselineTemp: 44, UptimeSeconds: 1500000,
				MaxPPPoESessions: 500,
				Interfaces: []SimInterface{
					{ID: uuid.MustParse("f0000000-0000-0000-000a-000000000001"), Name: "ether1-edge", Description: "To SID-EDGE-01", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 550, BaselineOutMbps: 350},
					{ID: uuid.MustParse("f0000000-0000-0000-000a-000000000002"), Name: "ether2-subs", Description: "Subscriber VLAN", SpeedMbps: 1000, Status: "up", AdminStatus: "up", BaselineInMbps: 280, BaselineOutMbps: 450},
				},
			},
		},
		Links: []SimLink{
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000001"), Name: "CORE-01↔CORE-02", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0001-000000000002"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0002-000000000002"), CapacityMbps: 25000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000002"), Name: "CORE-01↔UPSTREAM", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0001-000000000003"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0004-000000000003"), CapacityMbps: 25000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000003"), Name: "CORE-02↔BEY-EDGE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0002-000000000003"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0003-000000000002"), CapacityMbps: 10000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000004"), Name: "BEY-EDGE↔BEY-PPPOE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0003-000000000003"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0005-000000000001"), CapacityMbps: 10000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000005"), Name: "CORE-01↔TRI-EDGE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0001-000000000002"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0006-000000000001"), CapacityMbps: 10000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000006"), Name: "TRI-EDGE↔TRI-ACCESS", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0006-000000000002"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0007-000000000001"), CapacityMbps: 1000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000007"), Name: "TRI-ACCESS↔TRI-PPPOE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0007-000000000002"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0008-000000000001"), CapacityMbps: 1000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000008"), Name: "CORE-02↔SID-EDGE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0002-000000000003"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-0009-000000000001"), CapacityMbps: 10000},
			{ID: uuid.MustParse("aa000000-0000-0000-0000-000000000009"), Name: "SID-EDGE↔SID-PPPOE", SourceInterfaceID: uuid.MustParse("f0000000-0000-0000-0009-000000000002"), TargetInterfaceID: uuid.MustParse("f0000000-0000-0000-000a-000000000001"), CapacityMbps: 1000},
		},
	}
}

// DeviceByID returns the device with the given ID, or nil if not found.
func (t *Topology) DeviceByID(id uuid.UUID) *SimDevice {
	for i := range t.Devices {
		if t.Devices[i].ID == id {
			return &t.Devices[i]
		}
	}
	return nil
}

// PPPoEDevices returns all devices with the PPPoE role.
func (t *Topology) PPPoEDevices() []SimDevice {
	var out []SimDevice
	for _, d := range t.Devices {
		if d.Role == RolePPPoE {
			out = append(out, d)
		}
	}
	return out
}
