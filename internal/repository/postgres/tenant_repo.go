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

// TenantRepo implements repository.TenantRepository
type TenantRepo struct {
	db *sql.DB
}

// NewTenantRepo creates a new tenant repository
func NewTenantRepo(db *sql.DB) repository.TenantRepository {
	return &TenantRepo{db: db}
}

// Create creates a new tenant
func (r *TenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, contact_email, contact_phone, 
			subscription_tier, max_devices, max_users, status, trial_ends_at, 
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.ContactEmail, tenant.ContactPhone,
		tenant.SubscriptionTier, tenant.MaxDevices, tenant.MaxUsers, tenant.Status,
		tenant.TrialEndsAt, tenant.CreatedAt, tenant.UpdatedAt,
	)

	return err
}

// GetByID retrieves a tenant by ID
func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, 
			max_devices, max_users, status, trial_ends_at, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	tenant := &models.Tenant{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.ContactEmail, &tenant.ContactPhone,
		&tenant.SubscriptionTier, &tenant.MaxDevices, &tenant.MaxUsers, &tenant.Status,
		&tenant.TrialEndsAt, &tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}

	return tenant, err
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, 
			max_devices, max_users, status, trial_ends_at, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`

	tenant := &models.Tenant{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.ContactEmail, &tenant.ContactPhone,
		&tenant.SubscriptionTier, &tenant.MaxDevices, &tenant.MaxUsers, &tenant.Status,
		&tenant.TrialEndsAt, &tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}

	return tenant, err
}

// List retrieves a paginated list of tenants
func (r *TenantRepo) List(ctx context.Context, opts repository.ListOptions) ([]*models.Tenant, int64, error) {
	countQuery := `SELECT COUNT(*) FROM tenants`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (opts.Page - 1) * opts.PageSize
	query := `
		SELECT id, name, slug, contact_email, contact_phone, subscription_tier, 
			max_devices, max_users, status, trial_ends_at, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, opts.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tenants := make([]*models.Tenant, 0)
	for rows.Next() {
		tenant := &models.Tenant{}
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.ContactEmail, &tenant.ContactPhone,
			&tenant.SubscriptionTier, &tenant.MaxDevices, &tenant.MaxUsers, &tenant.Status,
			&tenant.TrialEndsAt, &tenant.CreatedAt, &tenant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}

	return tenants, total, rows.Err()
}

// Update updates a tenant
func (r *TenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $1, slug = $2, contact_email = $3, contact_phone = $4,
			subscription_tier = $5, max_devices = $6, max_users = $7,
			status = $8, trial_ends_at = $9, updated_at = $10
		WHERE id = $11
	`

	tenant.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		tenant.Name, tenant.Slug, tenant.ContactEmail, tenant.ContactPhone,
		tenant.SubscriptionTier, tenant.MaxDevices, tenant.MaxUsers,
		tenant.Status, tenant.TrialEndsAt, tenant.UpdatedAt,
		tenant.ID,
	)

	return err
}

// Delete deletes a tenant
func (r *TenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
