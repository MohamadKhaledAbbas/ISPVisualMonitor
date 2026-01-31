package repository

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

// ListOptions contains common listing options
type ListOptions struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
	Filters  map[string]interface{}
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TenantRepository defines the interface for tenant data access
type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	List(ctx context.Context, opts ListOptions) ([]*models.Tenant, int64, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RouterRepository defines the interface for router data access
type RouterRepository interface {
	Create(ctx context.Context, router *models.Router) error
	GetByID(ctx context.Context, tenantID, routerID uuid.UUID) (*models.Router, error)
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*models.Router, int64, error)
	Update(ctx context.Context, router *models.Router) error
	Delete(ctx context.Context, tenantID, routerID uuid.UUID) error
}

// InterfaceRepository defines the interface for interface data access
type InterfaceRepository interface {
	GetByID(ctx context.Context, tenantID, interfaceID uuid.UUID) (*models.Interface, error)
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*models.Interface, int64, error)
	ListByRouter(ctx context.Context, tenantID, routerID uuid.UUID, opts ListOptions) ([]*models.Interface, int64, error)
}

// LinkRepository defines the interface for link data access
type LinkRepository interface {
	GetByID(ctx context.Context, tenantID, linkID uuid.UUID) (*models.Link, error)
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*models.Link, int64, error)
}

// AlertRepository defines the interface for alert data access
type AlertRepository interface {
	GetByID(ctx context.Context, tenantID, alertID uuid.UUID) (*models.Alert, error)
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*models.Alert, int64, error)
	Acknowledge(ctx context.Context, tenantID, alertID, userID uuid.UUID) error
}
