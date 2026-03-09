package simulator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Writer persists simulator-generated telemetry into the application database.
// It writes to the same tables used by the real poller, exercising the real
// data path.
type Writer struct {
	db *sql.DB
}

// NewWriter creates a Writer that uses the given database connection.
func NewWriter(db *sql.DB) *Writer {
	return &Writer{db: db}
}

// WriteRouterMetrics inserts a router_metrics row.
func (w *Writer) WriteRouterMetrics(ctx context.Context, m RouterMetrics) error {
	const q = `
		INSERT INTO router_metrics
			(tenant_id, router_id, timestamp, cpu_percent, memory_percent, uptime_seconds, temperature_celsius)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := w.db.ExecContext(ctx, q,
		m.TenantID, m.RouterID, m.Timestamp,
		m.CPUPercent, m.MemoryPercent, m.UptimeSeconds, m.TemperatureCelsius,
	)
	return err
}

// WriteInterfaceMetrics inserts an interface_metrics row.
func (w *Writer) WriteInterfaceMetrics(ctx context.Context, m InterfaceMetrics) error {
	const q = `
		INSERT INTO interface_metrics
			(tenant_id, interface_id, timestamp, in_octets, out_octets,
			 in_packets, out_packets, in_errors, out_errors,
			 in_discards, out_discards, utilization_percent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := w.db.ExecContext(ctx, q,
		m.TenantID, m.InterfaceID, m.Timestamp,
		m.InOctets, m.OutOctets,
		m.InPackets, m.OutPackets,
		m.InErrors, m.OutErrors,
		m.InDiscards, m.OutDiscards,
		m.UtilizationPercent,
	)
	return err
}

// WriteAlert inserts an alert row.
func (w *Writer) WriteAlert(ctx context.Context, tenantID uuid.UUID, a ScenarioAlert) error {
	const q = `
		INSERT INTO alerts
			(id, tenant_id, name, description, severity, status, target_type, target_id, triggered_at, metadata)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, $8, $9)
	`
	_, err := w.db.ExecContext(ctx, q,
		uuid.New(), tenantID, a.Name, a.Description,
		a.Severity, a.TargetType, a.TargetID,
		time.Now(), a.Metadata,
	)
	return err
}

// UpdateRouterStatus sets the status field for a given router.
func (w *Writer) UpdateRouterStatus(ctx context.Context, routerID uuid.UUID, status string) error {
	_, err := w.db.ExecContext(ctx,
		"UPDATE routers SET status = $1, updated_at = NOW() WHERE id = $2",
		status, routerID,
	)
	return err
}

// UpdateInterfaceStatus sets the status field for a given interface.
func (w *Writer) UpdateInterfaceStatus(ctx context.Context, ifaceID uuid.UUID, status string) error {
	_, err := w.db.ExecContext(ctx,
		"UPDATE interfaces SET status = $1, updated_at = NOW() WHERE id = $2",
		status, ifaceID,
	)
	return err
}

// UpdateRouterLastPolled sets last_polled_at for a given router.
func (w *Writer) UpdateRouterLastPolled(ctx context.Context, routerID uuid.UUID, ts time.Time) error {
	_, err := w.db.ExecContext(ctx,
		"UPDATE routers SET last_polled_at = $1 WHERE id = $2",
		ts, routerID,
	)
	return err
}

// ResetToHealthy restores all demo routers and interfaces to active/up,
// and resolves all active alerts. This is the programmatic equivalent of
// the `healthy` demo scenario script.
func (w *Writer) ResetToHealthy(ctx context.Context, tenantID uuid.UUID) error {
	queries := []string{
		"UPDATE routers SET status = 'active' WHERE tenant_id = $1",
		"UPDATE interfaces SET status = 'up' WHERE tenant_id = $1",
		"UPDATE links SET status = 'up' WHERE tenant_id = $1",
		"UPDATE alerts SET status = 'resolved', resolved_at = NOW() WHERE tenant_id = $1 AND status IN ('active', 'acknowledged')",
	}
	for _, q := range queries {
		if _, err := w.db.ExecContext(ctx, q, tenantID); err != nil {
			return fmt.Errorf("reset healthy: %w", err)
		}
	}
	log.Println("[simulator] Reset to healthy baseline")
	return nil
}
