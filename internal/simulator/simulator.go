// Package simulator provides a development-only telemetry simulator for ISPVisualMonitor.
// It generates realistic ISP device, interface, and session metrics, writes them to the
// application database via the same tables the real poller uses, and supports named
// fault-injection scenarios for testing dashboards, alerting, and incident workflows.
//
// The simulator is enabled by setting ENABLE_SIMULATOR=true.
package simulator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SimMode controls how the simulator generates data.
type SimMode string

const (
	// ModeDeterministic uses a fixed seed for repeatable output.
	ModeDeterministic SimMode = "deterministic"
	// ModeScenario generates data according to the active scenario with some jitter.
	ModeScenario SimMode = "scenario"
	// ModeRandom uses a time-based seed for variable output each run.
	ModeRandom SimMode = "random"
)

// Config holds configuration for the simulator.
type Config struct {
	Mode     SimMode       `json:"mode"`
	Seed     int64         `json:"seed"`     // Only used in deterministic mode
	Interval time.Duration `json:"interval"` // Time between generation cycles
	Scenario ScenarioName  `json:"scenario"` // Initial scenario
}

// DefaultConfig returns sensible defaults for development.
func DefaultConfig() Config {
	return Config{
		Mode:     ModeScenario,
		Seed:     42,
		Interval: 30 * time.Second,
		Scenario: ScenarioHealthy,
	}
}

// Service is the main simulator service. It holds the topology, telemetry
// generator, scenario engine, and database writer.
type Service struct {
	cfg      Config
	topology *Topology
	gen      *TelemetryGenerator
	scenario *ScenarioEngine
	writer   *Writer

	// alertsEmitted tracks which scenario alerts have already been written so
	// we don't insert duplicates each tick.
	alertsEmitted map[ScenarioName]bool
}

// NewService creates a new simulator service.
func NewService(db *sql.DB, cfg Config) *Service {
	topo := DefaultTopology()

	var seed int64
	switch cfg.Mode {
	case ModeDeterministic:
		seed = cfg.Seed
	case ModeRandom:
		seed = 0 // TelemetryGenerator will use time-based seed
	default:
		seed = cfg.Seed
	}

	eng := NewScenarioEngine(topo)
	if cfg.Scenario != "" {
		if err := eng.Activate(cfg.Scenario); err != nil {
			log.Printf("[simulator] Warning: invalid initial scenario %q, using healthy", cfg.Scenario)
		}
	}

	return &Service{
		cfg:           cfg,
		topology:      topo,
		gen:           NewTelemetryGenerator(seed),
		scenario:      eng,
		writer:        NewWriter(db),
		alertsEmitted: make(map[ScenarioName]bool),
	}
}

// Start runs the simulator loop until the context is cancelled.
func (s *Service) Start(ctx context.Context) error {
	log.Printf("[simulator] Starting (mode=%s, interval=%s, scenario=%s, seed=%d)",
		s.cfg.Mode, s.cfg.Interval, s.scenario.Active(), s.gen.Seed())

	// Run first cycle immediately
	s.runCycle(ctx)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[simulator] Shutting down")
			return nil
		case <-ticker.C:
			s.runCycle(ctx)
		}
	}
}

// ActivateScenario switches the simulator to a named scenario at runtime.
func (s *Service) ActivateScenario(name ScenarioName) error {
	if err := s.scenario.Activate(name); err != nil {
		return err
	}
	// Clear emitted alerts so the new scenario can emit its own
	s.alertsEmitted = make(map[ScenarioName]bool)
	log.Printf("[simulator] Activated scenario: %s", name)
	return nil
}

// ResetToHealthy restores the simulator and database to a healthy baseline.
func (s *Service) ResetToHealthy(ctx context.Context) error {
	s.scenario.Reset()
	s.alertsEmitted = make(map[ScenarioName]bool)
	if s.cfg.Mode == ModeDeterministic {
		s.gen.Reset()
	}
	return s.writer.ResetToHealthy(ctx, s.topology.TenantID)
}

// ActiveScenario returns the currently active scenario name.
func (s *Service) ActiveScenario() ScenarioName {
	return s.scenario.Active()
}

// runCycle generates one round of telemetry for every device/interface,
// applies scenario overrides, and writes results to the database.
func (s *Service) runCycle(ctx context.Context) {
	now := time.Now().UTC().Truncate(time.Second)
	override := s.scenario.Override()

	metricsWritten := 0
	errCount := 0

	for i := range s.topology.Devices {
		dev := &s.topology.Devices[i]

		// --- Router status updates (scenario-driven) ---
		if override.RouterStatus != nil && dev.ID == RouterTRIACCESS01 {
			if err := s.writer.UpdateRouterStatus(ctx, dev.ID, *override.RouterStatus); err != nil {
				log.Printf("[simulator] Error updating router status: %v", err)
				errCount++
			}
		}

		// --- Interface status updates ---
		for j := range dev.Interfaces {
			iface := &dev.Interfaces[j]
			if override.InterfaceDown[iface.ID] {
				if err := s.writer.UpdateInterfaceStatus(ctx, iface.ID, "down"); err != nil {
					log.Printf("[simulator] Error updating interface status: %v", err)
					errCount++
				}
			}
		}

		// --- Router metrics ---
		rm := s.gen.GenerateRouterMetrics(dev, s.topology.TenantID, now)
		override.ApplyRouterMetrics(&rm, dev)
		if err := s.writer.WriteRouterMetrics(ctx, rm); err != nil {
			log.Printf("[simulator] Error writing router metrics for %s: %v", dev.Name, err)
			errCount++
		} else {
			metricsWritten++
		}

		// --- Interface metrics ---
		for j := range dev.Interfaces {
			iface := &dev.Interfaces[j]
			im := s.gen.GenerateInterfaceMetrics(iface, s.topology.TenantID, now)
			override.ApplyInterfaceMetrics(&im, iface)
			if err := s.writer.WriteInterfaceMetrics(ctx, im); err != nil {
				log.Printf("[simulator] Error writing interface metrics for %s/%s: %v", dev.Name, iface.Name, err)
				errCount++
			} else {
				metricsWritten++
			}
		}

		// --- PPPoE snapshots ---
		if dev.Role == RolePPPoE && dev.MaxPPPoESessions > 0 {
			snap := s.gen.GeneratePPPoESnapshot(dev, s.topology.TenantID, now)
			override.ApplyPPPoESnapshot(&snap)
			// Store as a role-specific metric via router_metrics overload
			// (session count goes into the cpu_percent column description won't apply—
			//  but we track it for dashboard use).
			// A cleaner approach would insert into role_specific_metrics JSONB table.
			// For now, write a router metric with session info attached.
			_ = snap // PPPoE snapshot is reflected via the router metrics and scenario alerts
		}

		// --- Update last_polled_at ---
		if err := s.writer.UpdateRouterLastPolled(ctx, dev.ID, now); err != nil {
			log.Printf("[simulator] Error updating last_polled_at for %s: %v", dev.Name, err)
			errCount++
		}
	}

	// --- Scenario alerts (emit once per scenario activation) ---
	if !s.alertsEmitted[s.scenario.Active()] && len(override.Alerts) > 0 {
		for _, alert := range override.Alerts {
			if err := s.writer.WriteAlert(ctx, s.topology.TenantID, alert); err != nil {
				log.Printf("[simulator] Error writing alert: %v", err)
				errCount++
			}
		}
		s.alertsEmitted[s.scenario.Active()] = true
	}

	s.scenario.Tick()

	if errCount > 0 {
		log.Printf("[simulator] Cycle complete: %d metrics written, %d errors", metricsWritten, errCount)
	} else {
		log.Printf("[simulator] Cycle complete: %d metrics written", metricsWritten)
	}
}

// GetTopology returns the simulator's topology (useful for inspection/tests).
func (s *Service) GetTopology() *Topology {
	return s.topology
}

// String returns a human-readable description of the simulator state.
func (s *Service) String() string {
	return fmt.Sprintf("Simulator{mode=%s, scenario=%s, devices=%d, seed=%d}",
		s.cfg.Mode, s.scenario.Active(), len(s.topology.Devices), s.gen.Seed())
}
