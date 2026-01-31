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

// RouterService handles router business logic
type RouterService struct {
	routerRepo repository.RouterRepository
	logger     *zap.Logger
}

// NewRouterService creates a new router service
func NewRouterService(
	routerRepo repository.RouterRepository,
	logger *zap.Logger,
) *RouterService {
	return &RouterService{
		routerRepo: routerRepo,
		logger:     logger,
	}
}

// CreateRouter creates a new router
func (s *RouterService) CreateRouter(ctx context.Context, tenantID uuid.UUID, req *dto.CreateRouterRequest) (*dto.RouterDTO, error) {
	router := &models.Router{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		Name:                   req.Name,
		Hostname:               req.Hostname,
		ManagementIP:           req.ManagementIP,
		RouterType:             "core",
		Status:                 "active",
		PollingEnabled:         true,
		PollingIntervalSeconds: 60,
		SNMPPort:               161,
	}

	if req.Vendor != "" {
		router.Vendor = &req.Vendor
	}
	if req.Model != nil {
		router.Model = req.Model
	}
	if req.POPID != nil {
		popID, err := uuid.Parse(*req.POPID)
		if err == nil {
			router.POPID = &popID
		}
	}
	if req.Location != nil {
		location := geoPointToPostGIS(req.Location)
		router.Location = &location
	}
	if req.SNMPVersion != nil {
		router.SNMPVersion = *req.SNMPVersion
	} else {
		router.SNMPVersion = "v2c"
	}
	if req.SNMPCommunity != nil {
		router.SNMPCommunity = req.SNMPCommunity
	}

	if err := s.routerRepo.Create(ctx, router); err != nil {
		s.logger.Error("Failed to create router", zap.Error(err))
		return nil, fmt.Errorf("failed to create router")
	}

	s.logger.Info("Router created successfully", zap.String("router_id", router.ID.String()))

	return toRouterDTO(router), nil
}

// GetRouter retrieves a router by ID
func (s *RouterService) GetRouter(ctx context.Context, tenantID, routerID uuid.UUID) (*dto.RouterDTO, error) {
	router, err := s.routerRepo.GetByID(ctx, tenantID, routerID)
	if err != nil {
		s.logger.Error("Failed to get router", zap.Error(err))
		return nil, fmt.Errorf("router not found")
	}

	return toRouterDTO(router), nil
}

// ListRouters retrieves a list of routers
func (s *RouterService) ListRouters(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]*dto.RouterDTO, int64, error) {
	routers, total, err := s.routerRepo.List(ctx, tenantID, opts)
	if err != nil {
		s.logger.Error("Failed to list routers", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list routers")
	}

	routerDTOs := make([]*dto.RouterDTO, len(routers))
	for i, router := range routers {
		routerDTOs[i] = toRouterDTO(router)
	}

	return routerDTOs, total, nil
}

// UpdateRouter updates an existing router
func (s *RouterService) UpdateRouter(ctx context.Context, tenantID, routerID uuid.UUID, req *dto.UpdateRouterRequest) (*dto.RouterDTO, error) {
	router, err := s.routerRepo.GetByID(ctx, tenantID, routerID)
	if err != nil {
		s.logger.Error("Failed to get router", zap.Error(err))
		return nil, fmt.Errorf("router not found")
	}

	if req.Name != nil {
		router.Name = *req.Name
	}
	if req.Hostname != nil {
		router.Hostname = req.Hostname
	}
	if req.ManagementIP != nil {
		router.ManagementIP = *req.ManagementIP
	}
	if req.Vendor != nil {
		router.Vendor = req.Vendor
	}
	if req.Model != nil {
		router.Model = req.Model
	}
	if req.Status != nil {
		router.Status = *req.Status
	}
	if req.Location != nil {
		location := geoPointToPostGIS(req.Location)
		router.Location = &location
	}
	if req.POPID != nil {
		popID, err := uuid.Parse(*req.POPID)
		if err == nil {
			router.POPID = &popID
		}
	}

	if err := s.routerRepo.Update(ctx, router); err != nil {
		s.logger.Error("Failed to update router", zap.Error(err))
		return nil, fmt.Errorf("failed to update router")
	}

	s.logger.Info("Router updated successfully", zap.String("router_id", router.ID.String()))

	return toRouterDTO(router), nil
}

// DeleteRouter deletes a router
func (s *RouterService) DeleteRouter(ctx context.Context, tenantID, routerID uuid.UUID) error {
	if err := s.routerRepo.Delete(ctx, tenantID, routerID); err != nil {
		s.logger.Error("Failed to delete router", zap.Error(err))
		return fmt.Errorf("failed to delete router")
	}

	s.logger.Info("Router deleted successfully", zap.String("router_id", routerID.String()))

	return nil
}

// toRouterDTO converts a Router model to RouterDTO
func toRouterDTO(router *models.Router) *dto.RouterDTO {
	routerDTO := &dto.RouterDTO{
		ID:           router.ID,
		TenantID:     router.TenantID,
		Name:         router.Name,
		Hostname:     router.Hostname,
		ManagementIP: router.ManagementIP,
		Vendor:       router.Vendor,
		Model:        router.Model,
		OSVersion:    router.OSVersion,
		Status:       router.Status,
		POPID:        router.POPID,
		CreatedAt:    router.CreatedAt,
		UpdatedAt:    router.UpdatedAt,
	}

	if router.Location != nil {
		routerDTO.Location = postGISToGeoPoint(*router.Location)
	}

	return routerDTO
}
