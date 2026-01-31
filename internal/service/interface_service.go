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

// InterfaceService handles interface business logic
type InterfaceService struct {
	interfaceRepo repository.InterfaceRepository
	routerRepo    repository.RouterRepository
	logger        *zap.Logger
}

// NewInterfaceService creates a new interface service
func NewInterfaceService(
	interfaceRepo repository.InterfaceRepository,
	routerRepo repository.RouterRepository,
	logger *zap.Logger,
) *InterfaceService {
	return &InterfaceService{
		interfaceRepo: interfaceRepo,
		routerRepo:    routerRepo,
		logger:        logger,
	}
}

// GetInterface retrieves an interface by ID
func (s *InterfaceService) GetInterface(ctx context.Context, tenantID, interfaceID uuid.UUID) (dto.InterfaceDTO, error) {
	iface, err := s.interfaceRepo.GetByID(ctx, tenantID, interfaceID)
	if err != nil {
		s.logger.Error("Failed to get interface", zap.Error(err))
		return dto.InterfaceDTO{}, fmt.Errorf("interface not found")
	}

	return toInterfaceDTO(iface, ""), nil
}

// ListInterfaces retrieves a list of interfaces
func (s *InterfaceService) ListInterfaces(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]dto.InterfaceDTO, int64, error) {
	interfaces, total, err := s.interfaceRepo.List(ctx, tenantID, opts)
	if err != nil {
		s.logger.Error("Failed to list interfaces", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list interfaces")
	}

	interfaceDTOs := make([]dto.InterfaceDTO, len(interfaces))
	for i, iface := range interfaces {
		interfaceDTOs[i] = toInterfaceDTO(iface, "")
	}

	return interfaceDTOs, total, nil
}

// ListRouterInterfaces retrieves a list of interfaces for a specific router
func (s *InterfaceService) ListRouterInterfaces(ctx context.Context, tenantID, routerID uuid.UUID, opts repository.ListOptions) ([]dto.InterfaceDTO, int64, error) {
	// Get router name for the DTO
	router, err := s.routerRepo.GetByID(ctx, tenantID, routerID)
	if err != nil {
		s.logger.Error("Failed to get router", zap.Error(err))
		return nil, 0, fmt.Errorf("router not found")
	}

	interfaces, total, err := s.interfaceRepo.ListByRouter(ctx, tenantID, routerID, opts)
	if err != nil {
		s.logger.Error("Failed to list router interfaces", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list router interfaces")
	}

	interfaceDTOs := make([]dto.InterfaceDTO, len(interfaces))
	for i, iface := range interfaces {
		interfaceDTOs[i] = toInterfaceDTO(iface, router.Name)
	}

	return interfaceDTOs, total, nil
}

// toInterfaceDTO converts an Interface model to InterfaceDTO
func toInterfaceDTO(iface *models.Interface, routerName string) dto.InterfaceDTO {
	return dto.InterfaceDTO{
		ID:          iface.ID,
		TenantID:    iface.TenantID,
		RouterID:    iface.RouterID,
		RouterName:  routerName,
		Name:        iface.Name,
		Description: iface.Description,
		IfIndex:     iface.IfIndex,
		IfType:      iface.IfType,
		SpeedMbps:   iface.SpeedMbps,
		MTU:         iface.MTU,
		MACAddress:  iface.MACAddress,
		IPAddress:   iface.IPAddress,
		SubnetMask:  iface.SubnetMask,
		Status:      iface.Status,
		AdminStatus: iface.AdminStatus,
		CreatedAt:   iface.CreatedAt,
		UpdatedAt:   iface.UpdatedAt,
	}
}
