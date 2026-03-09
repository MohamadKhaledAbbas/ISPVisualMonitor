package simulator

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Topology tests
// ---------------------------------------------------------------------------

func TestDefaultTopology(t *testing.T) {
	topo := DefaultTopology()

	if topo.TenantID != DemoTenantID {
		t.Errorf("expected tenant %s, got %s", DemoTenantID, topo.TenantID)
	}
	if len(topo.Devices) != 10 {
		t.Errorf("expected 10 devices, got %d", len(topo.Devices))
	}
	if len(topo.Links) != 9 {
		t.Errorf("expected 9 links, got %d", len(topo.Links))
	}

	// Verify all interface IDs are unique
	seen := make(map[uuid.UUID]bool)
	for _, dev := range topo.Devices {
		for _, iface := range dev.Interfaces {
			if seen[iface.ID] {
				t.Errorf("duplicate interface ID: %s", iface.ID)
			}
			seen[iface.ID] = true
		}
	}
}

func TestTopologyDeviceByID(t *testing.T) {
	topo := DefaultTopology()

	dev := topo.DeviceByID(RouterBEYCORE01)
	if dev == nil || dev.Name != "BEY-CORE-01" {
		t.Error("DeviceByID failed for BEY-CORE-01")
	}

	if topo.DeviceByID(uuid.Nil) != nil {
		t.Error("DeviceByID should return nil for unknown ID")
	}
}

func TestTopologyPPPoEDevices(t *testing.T) {
	topo := DefaultTopology()
	pppoe := topo.PPPoEDevices()
	if len(pppoe) != 3 {
		t.Errorf("expected 3 PPPoE devices, got %d", len(pppoe))
	}
}

// ---------------------------------------------------------------------------
// Telemetry generator tests
// ---------------------------------------------------------------------------

func TestTelemetryDeterministic(t *testing.T) {
	gen1 := NewTelemetryGenerator(42)
	gen2 := NewTelemetryGenerator(42)

	topo := DefaultTopology()
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dev := &topo.Devices[0]

	m1 := gen1.GenerateRouterMetrics(dev, topo.TenantID, ts)
	m2 := gen2.GenerateRouterMetrics(dev, topo.TenantID, ts)

	if m1.CPUPercent != m2.CPUPercent {
		t.Errorf("deterministic mode produced different CPU: %f vs %f", m1.CPUPercent, m2.CPUPercent)
	}
	if m1.MemoryPercent != m2.MemoryPercent {
		t.Errorf("deterministic mode produced different memory: %f vs %f", m1.MemoryPercent, m2.MemoryPercent)
	}
}

func TestTelemetryReset(t *testing.T) {
	gen := NewTelemetryGenerator(99)
	topo := DefaultTopology()
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dev := &topo.Devices[0]

	first := gen.GenerateRouterMetrics(dev, topo.TenantID, ts)
	// Consume more values
	gen.GenerateRouterMetrics(dev, topo.TenantID, ts)
	gen.GenerateRouterMetrics(dev, topo.TenantID, ts)

	gen.Reset()
	afterReset := gen.GenerateRouterMetrics(dev, topo.TenantID, ts)

	if first.CPUPercent != afterReset.CPUPercent {
		t.Errorf("reset did not replay: %f vs %f", first.CPUPercent, afterReset.CPUPercent)
	}
}

func TestGenerateInterfaceMetrics(t *testing.T) {
	gen := NewTelemetryGenerator(42)
	topo := DefaultTopology()
	ts := time.Now()
	dev := &topo.Devices[0]
	iface := &dev.Interfaces[1] // sfp1-core02, 25 Gbps

	m := gen.GenerateInterfaceMetrics(iface, topo.TenantID, ts)

	if m.InOctets <= 0 {
		t.Error("InOctets should be positive")
	}
	if m.UtilizationPercent < 0 || m.UtilizationPercent > 100 {
		t.Errorf("utilization out of range: %f", m.UtilizationPercent)
	}
}

func TestGeneratePPPoESnapshot(t *testing.T) {
	gen := NewTelemetryGenerator(42)
	topo := DefaultTopology()
	ts := time.Now()

	// Non-PPPoE device should return zero sessions
	coreDev := &topo.Devices[0]
	snap := gen.GeneratePPPoESnapshot(coreDev, topo.TenantID, ts)
	if snap.ActiveSessions != 0 {
		t.Errorf("non-PPPoE device should have 0 sessions, got %d", snap.ActiveSessions)
	}

	// PPPoE device should have positive sessions
	pppoeDev := topo.DeviceByID(RouterBEYPPPOE01)
	snap = gen.GeneratePPPoESnapshot(pppoeDev, topo.TenantID, ts)
	if snap.ActiveSessions <= 0 {
		t.Errorf("PPPoE device should have >0 sessions, got %d", snap.ActiveSessions)
	}
	if snap.ActiveSessions > snap.MaxSessions {
		t.Errorf("sessions (%d) exceed max (%d)", snap.ActiveSessions, snap.MaxSessions)
	}
}

// ---------------------------------------------------------------------------
// Scenario engine tests
// ---------------------------------------------------------------------------

func TestScenarioEngineActivate(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)

	if eng.Active() != ScenarioHealthy {
		t.Errorf("expected healthy, got %s", eng.Active())
	}

	if err := eng.Activate(ScenarioRouterDown); err != nil {
		t.Fatalf("Activate failed: %v", err)
	}
	if eng.Active() != ScenarioRouterDown {
		t.Errorf("expected router-down, got %s", eng.Active())
	}
}

func TestScenarioEngineReset(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioLinkSaturation)
	eng.Reset()

	if eng.Active() != ScenarioHealthy {
		t.Errorf("expected healthy after reset, got %s", eng.Active())
	}
}

func TestScenarioInvalidName(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)

	if err := eng.Activate("nonexistent"); err == nil {
		t.Error("expected error for unknown scenario")
	}
}

func TestScenarioOverridesRouterDown(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioRouterDown)

	o := eng.Override()
	if o.RouterStatus == nil || *o.RouterStatus != "offline" {
		t.Error("router-down scenario should set RouterStatus to offline")
	}
	if len(o.Alerts) == 0 {
		t.Error("router-down scenario should produce alerts")
	}
}

func TestScenarioOverridesLinkSaturation(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioLinkSaturation)

	o := eng.Override()
	uplinkID := uuid.MustParse("f0000000-0000-0000-0001-000000000003")
	if util, ok := o.InterfaceUtilOverride[uplinkID]; !ok || util != 93 {
		t.Error("link-saturation should override uplink utilization to 93")
	}
}

func TestScenarioOverridesSessionSpike(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioSessionSpike)

	o := eng.Override()
	if count, ok := o.PPPoESessionOverride[RouterBEYPPPOE01]; !ok || count != 950 {
		t.Error("session-spike should override BEY-PPPOE-01 sessions to 950")
	}
}

func TestScenarioFlapping(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioFlappingInterface)

	triAccessEther1 := uuid.MustParse("f0000000-0000-0000-0007-000000000001")

	// tick 0 → even → interface down
	o := eng.Override()
	if !o.InterfaceDown[triAccessEther1] {
		t.Error("tick 0 should have interface down")
	}

	eng.Tick()
	// tick 1 → odd → interface up
	o = eng.Override()
	if o.InterfaceDown[triAccessEther1] {
		t.Error("tick 1 should have interface up")
	}
}

// ---------------------------------------------------------------------------
// Scenario apply helpers
// ---------------------------------------------------------------------------

func TestApplyRouterMetricsRouterDown(t *testing.T) {
	gen := NewTelemetryGenerator(42)
	topo := DefaultTopology()
	ts := time.Now()

	dev := topo.DeviceByID(RouterTRIACCESS01)
	m := gen.GenerateRouterMetrics(dev, topo.TenantID, ts)

	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioRouterDown)
	o := eng.Override()
	o.ApplyRouterMetrics(&m, dev)

	if m.CPUPercent != 0 || m.MemoryPercent != 0 || m.UptimeSeconds != 0 {
		t.Error("router-down should zero out metrics for the downed router")
	}
}

func TestApplyInterfaceMetricsDown(t *testing.T) {
	gen := NewTelemetryGenerator(42)
	topo := DefaultTopology()
	ts := time.Now()

	dev := topo.DeviceByID(RouterBEYUPSTREAM)
	iface := &dev.Interfaces[1] // ether2-isp-b

	m := gen.GenerateInterfaceMetrics(iface, topo.TenantID, ts)

	eng := NewScenarioEngine(topo)
	_ = eng.Activate(ScenarioUpstreamOutage)
	o := eng.Override()
	o.ApplyInterfaceMetrics(&m, iface)

	if m.InOctets != 0 || m.OutOctets != 0 {
		t.Error("upstream-outage should zero out traffic on downed interface")
	}
}

func TestApplyPPPoESessionOverride(t *testing.T) {
	snap := PPPoESnapshot{
		RouterID:       RouterBEYPPPOE01,
		ActiveSessions: 600,
		MaxSessions:    1000,
	}

	eng := NewScenarioEngine(DefaultTopology())
	_ = eng.Activate(ScenarioSessionSpike)
	o := eng.Override()
	o.ApplyPPPoESnapshot(&snap)

	if snap.ActiveSessions != 950 {
		t.Errorf("expected 950 sessions, got %d", snap.ActiveSessions)
	}
}

// ---------------------------------------------------------------------------
// All scenarios produce valid overrides
// ---------------------------------------------------------------------------

func TestAllScenariosProduceOverrides(t *testing.T) {
	topo := DefaultTopology()
	eng := NewScenarioEngine(topo)

	for _, name := range AllScenarios() {
		if err := eng.Activate(name); err != nil {
			t.Fatalf("Activate(%s) failed: %v", name, err)
		}
		o := eng.Override()
		if o == nil {
			t.Errorf("scenario %s produced nil override", name)
		}
	}
}
