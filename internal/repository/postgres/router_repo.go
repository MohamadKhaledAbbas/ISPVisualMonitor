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

// RouterRepo implements repository.RouterRepository
type RouterRepo struct {
	db *sql.DB
}

// NewRouterRepo creates a new router repository
func NewRouterRepo(db *sql.DB) repository.RouterRepository {
	return &RouterRepo{db: db}
}

// Create creates a new router
func (r *RouterRepo) Create(ctx context.Context, router *models.Router) error {
	query := `
		INSERT INTO routers (id, tenant_id, pop_id, name, hostname, management_ip,
			location, router_type, vendor, model, os_version, serial_number, status,
			polling_enabled, polling_interval_seconds, snmp_version, snmp_community,
			snmp_port, description, created_at, updated_at, last_polled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`

	now := time.Now()
	router.CreatedAt = now
	router.UpdatedAt = now

	if router.ID == uuid.Nil {
		router.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		router.ID, router.TenantID, router.POPID, router.Name, router.Hostname,
		router.ManagementIP, router.Location, router.RouterType, router.Vendor,
		router.Model, router.OSVersion, router.SerialNumber, router.Status,
		router.PollingEnabled, router.PollingIntervalSeconds, router.SNMPVersion,
		router.SNMPCommunity, router.SNMPPort, router.Description,
		router.CreatedAt, router.UpdatedAt, router.LastPolledAt,
	)

	return err
}

// GetByID retrieves a router by ID with tenant isolation
func (r *RouterRepo) GetByID(ctx context.Context, tenantID, routerID uuid.UUID) (*models.Router, error) {
	query := `
		SELECT id, tenant_id, pop_id, name, hostname, management_ip, location,
			router_type, vendor, model, os_version, serial_number, status,
			polling_enabled, polling_interval_seconds, snmp_version, snmp_community,
			snmp_port, description, created_at, updated_at, last_polled_at
		FROM routers
		WHERE id = $1 AND tenant_id = $2
	`

	router := &models.Router{}
	err := r.db.QueryRowContext(ctx, query, routerID, tenantID).Scan(
		&router.ID, &router.TenantID, &router.POPID, &router.Name, &router.Hostname,
		&router.ManagementIP, &router.Location, &router.RouterType, &router.Vendor,
		&router.Model, &router.OSVersion, &router.SerialNumber, &router.Status,
		&router.PollingEnabled, &router.PollingIntervalSeconds, &router.SNMPVersion,
		&router.SNMPCommunity, &router.SNMPPort, &router.Description,
		&router.CreatedAt, &router.UpdatedAt, &router.LastPolledAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("router not found")
	}

	return router, err
}

// List retrieves a paginated list of routers for a tenant
func (r *RouterRepo) List(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]*models.Router, int64, error) {
	countQuery := `SELECT COUNT(*) FROM routers WHERE tenant_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, tenant_id, pop_id, name, hostname, management_ip, location,
			router_type, vendor, model, os_version, serial_number, status,
			polling_enabled, polling_interval_seconds, snmp_version, snmp_community,
			snmp_port, description, created_at, updated_at, last_polled_at
		FROM routers
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	routers := make([]*models.Router, 0)
	for rows.Next() {
		router := &models.Router{}
		err := rows.Scan(
			&router.ID, &router.TenantID, &router.POPID, &router.Name, &router.Hostname,
			&router.ManagementIP, &router.Location, &router.RouterType, &router.Vendor,
			&router.Model, &router.OSVersion, &router.SerialNumber, &router.Status,
			&router.PollingEnabled, &router.PollingIntervalSeconds, &router.SNMPVersion,
			&router.SNMPCommunity, &router.SNMPPort, &router.Description,
			&router.CreatedAt, &router.UpdatedAt, &router.LastPolledAt,
		)
		if err != nil {
			return nil, 0, err
		}
		routers = append(routers, router)
	}

	return routers, total, rows.Err()
}

// Update updates a router
func (r *RouterRepo) Update(ctx context.Context, router *models.Router) error {
	query := `
		UPDATE routers
		SET pop_id = $1, name = $2, hostname = $3, management_ip = $4, location = $5,
			router_type = $6, vendor = $7, model = $8, os_version = $9, serial_number = $10,
			status = $11, polling_enabled = $12, polling_interval_seconds = $13,
			snmp_version = $14, snmp_community = $15, snmp_port = $16, description = $17,
			updated_at = $18, last_polled_at = $19
		WHERE id = $20 AND tenant_id = $21
	`

	router.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		router.POPID, router.Name, router.Hostname, router.ManagementIP, router.Location,
		router.RouterType, router.Vendor, router.Model, router.OSVersion, router.SerialNumber,
		router.Status, router.PollingEnabled, router.PollingIntervalSeconds,
		router.SNMPVersion, router.SNMPCommunity, router.SNMPPort, router.Description,
		router.UpdatedAt, router.LastPolledAt,
		router.ID, router.TenantID,
	)

	return err
}

// Delete deletes a router with tenant isolation
func (r *RouterRepo) Delete(ctx context.Context, tenantID, routerID uuid.UUID) error {
	query := `DELETE FROM routers WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.ExecContext(ctx, query, routerID, tenantID)
	return err
}
