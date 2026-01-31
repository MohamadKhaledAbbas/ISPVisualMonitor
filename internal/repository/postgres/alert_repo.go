package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// AlertRepo implements repository.AlertRepository
type AlertRepo struct {
	db *sql.DB
}

// NewAlertRepo creates a new alert repository
func NewAlertRepo(db *sql.DB) repository.AlertRepository {
	return &AlertRepo{db: db}
}

// GetByID retrieves an alert by ID with tenant isolation
func (r *AlertRepo) GetByID(ctx context.Context, tenantID, alertID uuid.UUID) (*models.Alert, error) {
	query := `
		SELECT id, tenant_id, rule_id, name, description, severity, status,
			target_type, target_id, triggered_at, acknowledged_at, acknowledged_by,
			resolved_at, metadata
		FROM alerts
		WHERE id = $1 AND tenant_id = $2
	`

	alert := &models.Alert{}
	err := r.db.QueryRowContext(ctx, query, alertID, tenantID).Scan(
		&alert.ID, &alert.TenantID, &alert.RuleID, &alert.Name, &alert.Description,
		&alert.Severity, &alert.Status, &alert.TargetType, &alert.TargetID,
		&alert.TriggeredAt, &alert.AcknowledgedAt, &alert.AcknowledgedBy,
		&alert.ResolvedAt, &alert.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found")
	}

	return alert, err
}

// List retrieves a paginated list of alerts for a tenant
func (r *AlertRepo) List(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]*models.Alert, int64, error) {
	countQuery := `SELECT COUNT(*) FROM alerts WHERE tenant_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, tenant_id, rule_id, name, description, severity, status,
			target_type, target_id, triggered_at, acknowledged_at, acknowledged_by,
			resolved_at, metadata
		FROM alerts
		WHERE tenant_id = $1
		ORDER BY triggered_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	alerts := make([]*models.Alert, 0)
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.TenantID, &alert.RuleID, &alert.Name, &alert.Description,
			&alert.Severity, &alert.Status, &alert.TargetType, &alert.TargetID,
			&alert.TriggeredAt, &alert.AcknowledgedAt, &alert.AcknowledgedBy,
			&alert.ResolvedAt, &alert.Metadata,
		)
		if err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, total, rows.Err()
}

// Acknowledge acknowledges an alert with tenant isolation
func (r *AlertRepo) Acknowledge(ctx context.Context, tenantID, alertID, userID uuid.UUID) error {
	query := `
		UPDATE alerts
		SET acknowledged_at = $1, acknowledged_by = $2
		WHERE id = $3 AND tenant_id = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, userID, alertID, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}
