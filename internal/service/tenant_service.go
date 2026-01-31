package service

import (
	"context"
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantService handles tenant business logic
type TenantService struct {
	tenantRepo repository.TenantRepository
	logger     *zap.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(
	tenantRepo repository.TenantRepository,
	logger *zap.Logger,
) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		logger:     logger,
	}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.TenantDTO, error) {
	// Check if slug already exists
	existingTenant, err := s.tenantRepo.GetBySlug(ctx, req.Slug)
	if err == nil && existingTenant != nil {
		return nil, fmt.Errorf("slug already exists")
	}

	tenant := &models.Tenant{
		ID:               uuid.New(),
		Name:             req.Name,
		Slug:             req.Slug,
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,
		SubscriptionTier: req.SubscriptionTier,
		MaxDevices:       req.MaxDevices,
		MaxUsers:         req.MaxUsers,
		Status:           "active",
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		s.logger.Error("Failed to create tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to create tenant")
	}

	s.logger.Info("Tenant created successfully", zap.String("tenant_id", tenant.ID.String()))

	return toTenantDTO(tenant), nil
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*dto.TenantDTO, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, fmt.Errorf("tenant not found")
	}

	return toTenantDTO(tenant), nil
}

// ListTenants retrieves a list of tenants
func (s *TenantService) ListTenants(ctx context.Context, opts repository.ListOptions) ([]*dto.TenantDTO, int64, error) {
	tenants, total, err := s.tenantRepo.List(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list tenants")
	}

	tenantDTOs := make([]*dto.TenantDTO, len(tenants))
	for i, tenant := range tenants {
		tenantDTOs[i] = toTenantDTO(tenant)
	}

	return tenantDTOs, total, nil
}

// UpdateTenant updates an existing tenant
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantRequest) (*dto.TenantDTO, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, fmt.Errorf("tenant not found")
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.ContactEmail != nil {
		tenant.ContactEmail = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		tenant.ContactPhone = req.ContactPhone
	}
	if req.SubscriptionTier != nil {
		tenant.SubscriptionTier = *req.SubscriptionTier
	}
	if req.MaxDevices != nil {
		tenant.MaxDevices = *req.MaxDevices
	}
	if req.MaxUsers != nil {
		tenant.MaxUsers = *req.MaxUsers
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to update tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to update tenant")
	}

	s.logger.Info("Tenant updated successfully", zap.String("tenant_id", tenant.ID.String()))

	return toTenantDTO(tenant), nil
}

// toTenantDTO converts a Tenant model to TenantDTO
func toTenantDTO(tenant *models.Tenant) *dto.TenantDTO {
	return &dto.TenantDTO{
		ID:               tenant.ID,
		Name:             tenant.Name,
		Slug:             tenant.Slug,
		ContactEmail:     tenant.ContactEmail,
		ContactPhone:     tenant.ContactPhone,
		SubscriptionTier: tenant.SubscriptionTier,
		MaxDevices:       tenant.MaxDevices,
		MaxUsers:         tenant.MaxUsers,
		Status:           tenant.Status,
		TrialEndsAt:      tenant.TrialEndsAt,
		CreatedAt:        tenant.CreatedAt,
		UpdatedAt:        tenant.UpdatedAt,
	}
}
