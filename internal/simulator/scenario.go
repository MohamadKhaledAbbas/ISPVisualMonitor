package simulator

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// ScenarioName identifies a named simulation scenario.
type ScenarioName string

const (
	ScenarioHealthy          ScenarioName = "healthy"
	ScenarioRouterDown       ScenarioName = "router-down"
	ScenarioLinkSaturation   ScenarioName = "link-saturation"
	ScenarioUpstreamOutage   ScenarioName = "upstream-outage"
	ScenarioPacketLoss       ScenarioName = "packet-loss"
	ScenarioSessionSpike     ScenarioName = "session-spike"
	ScenarioFlappingInterface ScenarioName = "flapping-interface"
)

// AllScenarios returns the list of all supported scenario names.
func AllScenarios() []ScenarioName {
	return []ScenarioName{
		ScenarioHealthy,
		ScenarioRouterDown,
		ScenarioLinkSaturation,
		ScenarioUpstreamOutage,
		ScenarioPacketLoss,
		ScenarioSessionSpike,
		ScenarioFlappingInterface,
	}
}

// ScenarioOverride captures per-device or per-interface modifications that a
// scenario applies on top of the baseline telemetry.
type ScenarioOverride struct {
	// Device-level overrides
	RouterStatus   *string  // force status (e.g. "offline")
	CPUMultiplier  float64  // multiply baseline CPU (1.0 = no change)
	MemMultiplier  float64  // multiply baseline Memory
	TempMultiplier float64  // multiply baseline Temp

	// Interface-level overrides keyed by interface ID
	InterfaceDown         map[uuid.UUID]bool    // force interface down
	InterfaceUtilOverride map[uuid.UUID]float64 // force utilization %
	InterfaceErrorRate    map[uuid.UUID]float64 // force error multiplier

	// PPPoE overrides
	PPPoESessionOverride map[uuid.UUID]int // force session count by router ID

	// Alerts to inject
	Alerts []ScenarioAlert
}

// ScenarioAlert represents an alert that a scenario wants to create.
type ScenarioAlert struct {
	Name        string
	Description string
	Severity    string // critical, warning, info
	TargetType  string // router, interface
	TargetID    uuid.UUID
	Metadata    string // JSON
}

// NewBaseOverride returns an override with no modifications (identity).
func NewBaseOverride() *ScenarioOverride {
	return &ScenarioOverride{
		CPUMultiplier:         1.0,
		MemMultiplier:         1.0,
		TempMultiplier:        1.0,
		InterfaceDown:         make(map[uuid.UUID]bool),
		InterfaceUtilOverride: make(map[uuid.UUID]float64),
		InterfaceErrorRate:    make(map[uuid.UUID]float64),
		PPPoESessionOverride:  make(map[uuid.UUID]int),
	}
}

// ScenarioEngine manages the active scenario and computes overrides to be
// applied during telemetry generation.
type ScenarioEngine struct {
	active   ScenarioName
	topology *Topology
	tick     int // incremented each generation cycle (used for flapping)
}

// NewScenarioEngine creates a new engine starting in the healthy scenario.
func NewScenarioEngine(topo *Topology) *ScenarioEngine {
	return &ScenarioEngine{
		active:   ScenarioHealthy,
		topology: topo,
	}
}

// Active returns the currently active scenario name.
func (e *ScenarioEngine) Active() ScenarioName { return e.active }

// Activate switches to a named scenario. Returns error for unknown names.
func (e *ScenarioEngine) Activate(name ScenarioName) error {
	for _, s := range AllScenarios() {
		if s == name {
			e.active = name
			e.tick = 0
			return nil
		}
	}
	return fmt.Errorf("unknown scenario: %s", name)
}

// Reset returns to the healthy baseline.
func (e *ScenarioEngine) Reset() {
	e.active = ScenarioHealthy
	e.tick = 0
}

// Tick advances the internal counter (call once per generation cycle).
func (e *ScenarioEngine) Tick() { e.tick++ }

// Override returns the current ScenarioOverride based on the active scenario.
func (e *ScenarioEngine) Override() *ScenarioOverride {
	switch e.active {
	case ScenarioRouterDown:
		return e.routerDown()
	case ScenarioLinkSaturation:
		return e.linkSaturation()
	case ScenarioUpstreamOutage:
		return e.upstreamOutage()
	case ScenarioPacketLoss:
		return e.packetLoss()
	case ScenarioSessionSpike:
		return e.sessionSpike()
	case ScenarioFlappingInterface:
		return e.flappingInterface()
	default:
		return NewBaseOverride()
	}
}

// --- scenario implementations ---

func (e *ScenarioEngine) routerDown() *ScenarioOverride {
	o := NewBaseOverride()
	offline := "offline"
	o.RouterStatus = &offline
	// TRI-ACCESS-01 goes offline (same as demo-scenarios.sh)
	// All its interfaces go down
	for _, iface := range e.topology.DeviceByID(RouterTRIACCESS01).Interfaces {
		o.InterfaceDown[iface.ID] = true
	}
	o.Alerts = []ScenarioAlert{
		{
			Name:        "TRI-ACCESS-01 offline",
			Description: "Router TRI-ACCESS-01 unreachable. Last poll timed out after 30s.",
			Severity:    "critical",
			TargetType:  "router",
			TargetID:    RouterTRIACCESS01,
			Metadata:    `{"consecutive_failures": 3}`,
		},
	}
	return o
}

func (e *ScenarioEngine) linkSaturation() *ScenarioOverride {
	o := NewBaseOverride()
	// BEY-CORE-01 sfp2-uplink (to upstream) saturated at 93 %
	uplinkID := uuid.MustParse("f0000000-0000-0000-0001-000000000003")
	o.InterfaceUtilOverride[uplinkID] = 93
	o.CPUMultiplier = 1.4
	o.MemMultiplier = 1.15
	o.Alerts = []ScenarioAlert{
		{
			Name:        "Core link congestion — BEY-CORE-01 sfp2-uplink",
			Description: "Upstream interface utilization at 93%. Sustained for 15 minutes.",
			Severity:    "warning",
			TargetType:  "interface",
			TargetID:    uplinkID,
			Metadata:    `{"utilization_percent": 93, "threshold": 85}`,
		},
	}
	return o
}

func (e *ScenarioEngine) upstreamOutage() *ScenarioOverride {
	o := NewBaseOverride()
	ispBID := uuid.MustParse("f0000000-0000-0000-0004-000000000002")
	o.InterfaceDown[ispBID] = true
	o.Alerts = []ScenarioAlert{
		{
			Name:        "Upstream ISP-B link down",
			Description: "BEY-UPSTREAM-01 ether2-isp-b is down. All traffic failing over to ISP-A.",
			Severity:    "critical",
			TargetType:  "interface",
			TargetID:    ispBID,
			Metadata:    `{"failover_active": true, "isp": "ISP-B"}`,
		},
	}
	return o
}

func (e *ScenarioEngine) packetLoss() *ScenarioOverride {
	o := NewBaseOverride()
	triUplinkID := uuid.MustParse("f0000000-0000-0000-0006-000000000001")
	o.InterfaceErrorRate[triUplinkID] = 50.0 // 50× error multiplier
	o.Alerts = []ScenarioAlert{
		{
			Name:        "Packet loss spike — TRI-EDGE-01 sfp1-core",
			Description: "Error rate spiked to 3.2% on Tripoli uplink. Possible fiber degradation.",
			Severity:    "warning",
			TargetType:  "interface",
			TargetID:    triUplinkID,
			Metadata:    `{"error_rate_percent": 3.2, "threshold": 1.0}`,
		},
	}
	return o
}

func (e *ScenarioEngine) sessionSpike() *ScenarioOverride {
	o := NewBaseOverride()
	o.PPPoESessionOverride[RouterBEYPPPOE01] = 950
	o.CPUMultiplier = 1.5
	o.MemMultiplier = 1.3
	o.Alerts = []ScenarioAlert{
		{
			Name:        "PPPoE session count critical — BEY-PPPOE-01",
			Description: "Active sessions at 950+. Approaching maximum capacity of 1000.",
			Severity:    "critical",
			TargetType:  "router",
			TargetID:    RouterBEYPPPOE01,
			Metadata:    `{"active_sessions": 950, "max_sessions": 1000}`,
		},
	}
	return o
}

func (e *ScenarioEngine) flappingInterface() *ScenarioOverride {
	o := NewBaseOverride()
	triAccessEther1 := uuid.MustParse("f0000000-0000-0000-0007-000000000001")

	// Alternate up/down every tick to simulate flapping
	if e.tick%2 == 0 {
		o.InterfaceDown[triAccessEther1] = true
	}
	o.Alerts = []ScenarioAlert{
		{
			Name:        "Interface flapping — TRI-ACCESS-01 ether1-edge",
			Description: fmt.Sprintf("Interface has toggled %d times in the last observation window.", e.tick+1),
			Severity:    "warning",
			TargetType:  "interface",
			TargetID:    triAccessEther1,
			Metadata:    fmt.Sprintf(`{"flap_count": %d, "observation_window": "%s"}`, e.tick+1, (time.Duration(e.tick+1)*5*time.Minute).String()),
		},
	}
	return o
}

// ApplyRouterMetrics applies scenario overrides to a RouterMetrics value.
func (o *ScenarioOverride) ApplyRouterMetrics(m *RouterMetrics, dev *SimDevice) {
	// If this device is the downed router, zero everything
	if o.RouterStatus != nil && *o.RouterStatus == "offline" && dev.ID == RouterTRIACCESS01 {
		m.CPUPercent = 0
		m.MemoryPercent = 0
		m.TemperatureCelsius = 0
		m.UptimeSeconds = 0
		return
	}
	m.CPUPercent = math.Min(m.CPUPercent*o.CPUMultiplier, 100)
	m.MemoryPercent = math.Min(m.MemoryPercent*o.MemMultiplier, 100)
	m.TemperatureCelsius = m.TemperatureCelsius * o.TempMultiplier
}

// ApplyInterfaceMetrics applies scenario overrides to an InterfaceMetrics value.
func (o *ScenarioOverride) ApplyInterfaceMetrics(m *InterfaceMetrics, iface *SimInterface) {
	if o.InterfaceDown[iface.ID] {
		m.InOctets = 0
		m.OutOctets = 0
		m.InPackets = 0
		m.OutPackets = 0
		m.UtilizationPercent = 0
		return
	}
	if util, ok := o.InterfaceUtilOverride[iface.ID]; ok {
		m.UtilizationPercent = util
		// Scale octets to match utilization
		speedBytes := float64(iface.SpeedMbps) * 125000
		m.InOctets = int64(speedBytes * util / 100 * 300) // 5-min sample
		m.OutOctets = int64(speedBytes * util * 0.75 / 100 * 300)
		m.InPackets = m.InOctets / 750
		m.OutPackets = m.OutOctets / 750
	}
	if errMul, ok := o.InterfaceErrorRate[iface.ID]; ok {
		m.InErrors = int64(float64(m.InErrors+10) * errMul)
		m.OutErrors = int64(float64(m.OutErrors+5) * errMul)
	}
}

// ApplyPPPoESnapshot applies scenario overrides to a PPPoESnapshot.
func (o *ScenarioOverride) ApplyPPPoESnapshot(snap *PPPoESnapshot) {
	if count, ok := o.PPPoESessionOverride[snap.RouterID]; ok {
		snap.ActiveSessions = count
	}
}
