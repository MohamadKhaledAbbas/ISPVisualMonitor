package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// InterfaceRepo implements repository.InterfaceRepository
type InterfaceRepo struct {
	db *sql.DB
}

// NewInterfaceRepo creates a new interface repository
func NewInterfaceRepo(db *sql.DB) repository.InterfaceRepository {
	return &InterfaceRepo{db: db}
}

// GetByID retrieves an interface by ID with tenant isolation
func (r *InterfaceRepo) GetByID(ctx context.Context, tenantID, interfaceID uuid.UUID) (*models.Interface, error) {
	query := `
		SELECT id, tenant_id, router_id, name, description, if_index, if_type,
			speed_mbps, mtu, mac_address, ip_address, subnet_mask, status,
			admin_status, created_at, updated_at
		FROM interfaces
		WHERE id = $1 AND tenant_id = $2
	`
	
	iface := &models.Interface{}
	err := r.db.QueryRowContext(ctx, query, interfaceID, tenantID).Scan(
		&iface.ID, &iface.TenantID, &iface.RouterID, &iface.Name, &iface.Description,
		&iface.IfIndex, &iface.IfType, &iface.SpeedMbps, &iface.MTU, &iface.MACAddress,
		&iface.IPAddress, &iface.SubnetMask, &iface.Status, &iface.AdminStatus,
		&iface.CreatedAt, &iface.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("interface not found")
	}
	
	return iface, err
}

// List retrieves a paginated list of interfaces for a tenant
func (r *InterfaceRepo) List(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]*models.Interface, int64, error) {
	countQuery := `SELECT COUNT(*) FROM interfaces WHERE tenant_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, tenant_id, router_id, name, description, if_index, if_type,
			speed_mbps, mtu, mac_address, ip_address, subnet_mask, status,
			admin_status, created_at, updated_at
		FROM interfaces
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	interfaces := make([]*models.Interface, 0)
	for rows.Next() {
		iface := &models.Interface{}
		err := rows.Scan(
			&iface.ID, &iface.TenantID, &iface.RouterID, &iface.Name, &iface.Description,
			&iface.IfIndex, &iface.IfType, &iface.SpeedMbps, &iface.MTU, &iface.MACAddress,
			&iface.IPAddress, &iface.SubnetMask, &iface.Status, &iface.AdminStatus,
			&iface.CreatedAt, &iface.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		interfaces = append(interfaces, iface)
	}
	
	return interfaces, total, rows.Err()
}

// ListByRouter retrieves a paginated list of interfaces for a specific router with tenant isolation
func (r *InterfaceRepo) ListByRouter(ctx context.Context, tenantID, routerID uuid.UUID, opts repository.ListOptions) ([]*models.Interface, int64, error) {
	countQuery := `SELECT COUNT(*) FROM interfaces WHERE tenant_id = $1 AND router_id = $2`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, tenantID, routerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, tenant_id, router_id, name, description, if_index, if_type,
			speed_mbps, mtu, mac_address, ip_address, subnet_mask, status,
			admin_status, created_at, updated_at
		FROM interfaces
		WHERE tenant_id = $1 AND router_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, routerID, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	interfaces := make([]*models.Interface, 0)
	for rows.Next() {
		iface := &models.Interface{}
		err := rows.Scan(
			&iface.ID, &iface.TenantID, &iface.RouterID, &iface.Name, &iface.Description,
			&iface.IfIndex, &iface.IfType, &iface.SpeedMbps, &iface.MTU, &iface.MACAddress,
			&iface.IPAddress, &iface.SubnetMask, &iface.Status, &iface.AdminStatus,
			&iface.CreatedAt, &iface.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		interfaces = append(interfaces, iface)
	}
	
	return interfaces, total, rows.Err()
}
