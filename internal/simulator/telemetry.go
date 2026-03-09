package simulator

import (
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// TelemetryGenerator produces synthetic metrics for simulated devices.
// It uses a seeded random source for deterministic replay when needed.
type TelemetryGenerator struct {
	rng  *rand.Rand
	seed int64
}

// NewTelemetryGenerator creates a new generator. Passing seed=0 uses the
// current timestamp, giving random behaviour; any other seed produces
// deterministic output.
func NewTelemetryGenerator(seed int64) *TelemetryGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &TelemetryGenerator{
		rng:  rand.New(rand.NewSource(seed)),
		seed: seed,
	}
}

// Seed returns the seed used by this generator.
func (g *TelemetryGenerator) Seed() int64 { return g.seed }

// Reset re-seeds the generator so the same sequence of values is produced.
func (g *TelemetryGenerator) Reset() {
	g.rng = rand.New(rand.NewSource(g.seed))
}

// ---------- Router-level metrics ----------

// RouterMetrics holds time-series metrics for a single router at a point in time.
type RouterMetrics struct {
	RouterID           uuid.UUID
	TenantID           uuid.UUID
	Timestamp          time.Time
	CPUPercent         float64
	MemoryPercent      float64
	UptimeSeconds      int64
	TemperatureCelsius float64
}

// GenerateRouterMetrics produces a single RouterMetrics sample.
// The values jitter around the device baseline.
func (g *TelemetryGenerator) GenerateRouterMetrics(dev *SimDevice, tenantID uuid.UUID, ts time.Time) RouterMetrics {
	return RouterMetrics{
		RouterID:           dev.ID,
		TenantID:           tenantID,
		Timestamp:          ts,
		CPUPercent:         g.jitter(dev.BaselineCPU, 8),
		MemoryPercent:      g.jitter(dev.BaselineMemory, 5),
		UptimeSeconds:      dev.UptimeSeconds + int64(time.Since(ts).Seconds()),
		TemperatureCelsius: g.jitter(dev.BaselineTemp, 3),
	}
}

// ---------- Interface-level metrics ----------

// InterfaceMetrics holds time-series metrics for a single interface.
type InterfaceMetrics struct {
	InterfaceID        uuid.UUID
	TenantID           uuid.UUID
	Timestamp          time.Time
	InOctets           int64
	OutOctets          int64
	InPackets          int64
	OutPackets         int64
	InErrors           int64
	OutErrors          int64
	InDiscards         int64
	OutDiscards        int64
	UtilizationPercent float64
}

// GenerateInterfaceMetrics produces a single InterfaceMetrics sample.
func (g *TelemetryGenerator) GenerateInterfaceMetrics(iface *SimInterface, tenantID uuid.UUID, ts time.Time) InterfaceMetrics {
	inMbps := g.jitter(iface.BaselineInMbps, iface.BaselineInMbps*0.15)
	outMbps := g.jitter(iface.BaselineOutMbps, iface.BaselineOutMbps*0.15)

	// Convert Mbps to octets (bytes) for a 5-minute sample interval
	const sampleSeconds = 300
	inOctets := int64(inMbps * 125000 * sampleSeconds) // Mbps → bytes/s → bytes in interval
	outOctets := int64(outMbps * 125000 * sampleSeconds)

	// Rough packet estimation (~750 bytes average packet)
	inPackets := inOctets / 750
	outPackets := outOctets / 750

	// Small random error/discard counts (healthy baseline)
	inErrors := int64(g.rng.Intn(5))
	outErrors := int64(g.rng.Intn(3))
	inDiscards := int64(g.rng.Intn(2))
	outDiscards := int64(g.rng.Intn(2))

	// Utilization based on speed
	utilisation := 0.0
	if iface.SpeedMbps > 0 {
		utilisation = math.Max(inMbps, outMbps) / float64(iface.SpeedMbps) * 100
	}

	return InterfaceMetrics{
		InterfaceID:        iface.ID,
		TenantID:           tenantID,
		Timestamp:          ts,
		InOctets:           inOctets,
		OutOctets:          outOctets,
		InPackets:          inPackets,
		OutPackets:         outPackets,
		InErrors:           inErrors,
		OutErrors:          outErrors,
		InDiscards:         inDiscards,
		OutDiscards:        outDiscards,
		UtilizationPercent: math.Min(utilisation, 100),
	}
}

// ---------- PPPoE session counts ----------

// PPPoESnapshot represents a point-in-time PPPoE session count for a device.
type PPPoESnapshot struct {
	RouterID       uuid.UUID
	TenantID       uuid.UUID
	Timestamp      time.Time
	ActiveSessions int
	MaxSessions    int
}

// GeneratePPPoESnapshot produces a PPPoE session count for a PPPoE device.
func (g *TelemetryGenerator) GeneratePPPoESnapshot(dev *SimDevice, tenantID uuid.UUID, ts time.Time) PPPoESnapshot {
	if dev.MaxPPPoESessions == 0 {
		return PPPoESnapshot{RouterID: dev.ID, TenantID: tenantID, Timestamp: ts}
	}
	// ~60-80 % of capacity by default
	baseline := float64(dev.MaxPPPoESessions) * (0.6 + g.rng.Float64()*0.2)
	active := int(g.jitter(baseline, baseline*0.05))
	if active < 0 {
		active = 0
	}
	if active > dev.MaxPPPoESessions {
		active = dev.MaxPPPoESessions
	}
	return PPPoESnapshot{
		RouterID:       dev.ID,
		TenantID:       tenantID,
		Timestamp:      ts,
		ActiveSessions: active,
		MaxSessions:    dev.MaxPPPoESessions,
	}
}

// ---------- helpers ----------

// jitter adds gaussian-ish noise (uniform in [-spread, +spread]) around a baseline.
// Values are clamped to [0, 100] when they represent percentages.
func (g *TelemetryGenerator) jitter(baseline, spread float64) float64 {
	v := baseline + (g.rng.Float64()*2-1)*spread
	if v < 0 {
		v = 0
	}
	return math.Round(v*100) / 100 // two-decimal precision
}
