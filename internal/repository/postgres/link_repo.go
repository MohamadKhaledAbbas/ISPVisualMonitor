package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// LinkRepo implements repository.LinkRepository
type LinkRepo struct {
	db *sql.DB
}

// NewLinkRepo creates a new link repository
func NewLinkRepo(db *sql.DB) repository.LinkRepository {
	return &LinkRepo{db: db}
}

// GetByID retrieves a link by ID with tenant isolation
func (r *LinkRepo) GetByID(ctx context.Context, tenantID, linkID uuid.UUID) (*models.Link, error) {
	query := `
		SELECT id, tenant_id, name, source_interface_id, target_interface_id,
			link_type, capacity_mbps, latency_ms, status, path_geometry,
			description, created_at, updated_at
		FROM links
		WHERE id = $1 AND tenant_id = $2
	`

	link := &models.Link{}
	err := r.db.QueryRowContext(ctx, query, linkID, tenantID).Scan(
		&link.ID, &link.TenantID, &link.Name, &link.SourceInterfaceID, &link.TargetInterfaceID,
		&link.LinkType, &link.CapacityMbps, &link.LatencyMs, &link.Status, &link.PathGeometry,
		&link.Description, &link.CreatedAt, &link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("link not found")
	}

	return link, err
}

// List retrieves a paginated list of links for a tenant
func (r *LinkRepo) List(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]*models.Link, int64, error) {
	countQuery := `SELECT COUNT(*) FROM links WHERE tenant_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, tenant_id, name, source_interface_id, target_interface_id,
			link_type, capacity_mbps, latency_ms, status, path_geometry,
			description, created_at, updated_at
		FROM links
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	links := make([]*models.Link, 0)
	for rows.Next() {
		link := &models.Link{}
		err := rows.Scan(
			&link.ID, &link.TenantID, &link.Name, &link.SourceInterfaceID, &link.TargetInterfaceID,
			&link.LinkType, &link.CapacityMbps, &link.LatencyMs, &link.Status, &link.PathGeometry,
			&link.Description, &link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		links = append(links, link)
	}

	return links, total, rows.Err()
}
