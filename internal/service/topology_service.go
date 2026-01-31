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

// TopologyService handles network topology business logic
type TopologyService struct {
	routerRepo    repository.RouterRepository
	interfaceRepo repository.InterfaceRepository
	linkRepo      repository.LinkRepository
	logger        *zap.Logger
}

// NewTopologyService creates a new topology service
func NewTopologyService(
	routerRepo repository.RouterRepository,
	interfaceRepo repository.InterfaceRepository,
	linkRepo repository.LinkRepository,
	logger *zap.Logger,
) *TopologyService {
	return &TopologyService{
		routerRepo:    routerRepo,
		interfaceRepo: interfaceRepo,
		linkRepo:      linkRepo,
		logger:        logger,
	}
}

// GetTopology retrieves the network topology
func (s *TopologyService) GetTopology(ctx context.Context, tenantID uuid.UUID) (*dto.TopologyResponse, error) {
	// Get all routers
	routers, _, err := s.routerRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list routers", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Get all links
	links, _, err := s.linkRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list links", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Get all interfaces to map interface IDs to router IDs
	interfaces, _, err := s.interfaceRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list interfaces", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Create interface ID to router ID mapping
	interfaceToRouter := make(map[uuid.UUID]uuid.UUID)
	for _, iface := range interfaces {
		interfaceToRouter[iface.ID] = iface.RouterID
	}

	// Convert routers to RouterNode
	routerNodes := make([]dto.RouterNode, len(routers))
	for i, router := range routers {
		routerNodes[i] = dto.RouterNode{
			ID:           router.ID,
			Name:         router.Name,
			ManagementIP: router.ManagementIP,
			Status:       router.Status,
			Vendor:       router.Vendor,
			Model:        router.Model,
		}
		if router.Location != nil {
			routerNodes[i].Location = postGISToGeoPoint(*router.Location)
		}
	}

	// Convert links to TopologyLink
	topologyLinks := make([]dto.TopologyLink, len(links))
	for i, link := range links {
		topologyLinks[i] = dto.TopologyLink{
			ID:                link.ID,
			Name:              link.Name,
			SourceInterfaceID: link.SourceInterfaceID,
			TargetInterfaceID: link.TargetInterfaceID,
			SourceRouterID:    interfaceToRouter[link.SourceInterfaceID],
			TargetRouterID:    interfaceToRouter[link.TargetInterfaceID],
			LinkType:          link.LinkType,
			CapacityMbps:      link.CapacityMbps,
			LatencyMs:         link.LatencyMs,
			Status:            link.Status,
		}
	}

	return &dto.TopologyResponse{
		Routers: routerNodes,
		Links:   topologyLinks,
	}, nil
}

// GetTopologyGeoJSON retrieves the network topology in GeoJSON format
func (s *TopologyService) GetTopologyGeoJSON(ctx context.Context, tenantID uuid.UUID) (*dto.GeoJSONResponse, error) {
	// Get all routers
	routers, _, err := s.routerRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list routers", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Get all links
	links, _, err := s.linkRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list links", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Get all interfaces to map interface IDs to router IDs
	interfaces, _, err := s.interfaceRepo.List(ctx, tenantID, repository.ListOptions{})
	if err != nil {
		s.logger.Error("Failed to list interfaces", zap.Error(err))
		return nil, fmt.Errorf("failed to get topology")
	}

	// Create interface ID to router mapping
	interfaceToRouter := make(map[uuid.UUID]*models.Router)
	for _, iface := range interfaces {
		for _, router := range routers {
			if router.ID == iface.RouterID {
				interfaceToRouter[iface.ID] = router
				break
			}
		}
	}

	features := []dto.GeoFeature{}

	// Add router features as Points
	for _, router := range routers {
		if router.Location != nil {
			geoPoint := postGISToGeoPoint(*router.Location)
			if geoPoint != nil {
				feature := dto.GeoFeature{
					Type: "Feature",
					Geometry: dto.GeoGeometry{
						Type:        "Point",
						Coordinates: []float64{geoPoint.Longitude, geoPoint.Latitude},
					},
					Properties: map[string]interface{}{
						"id":            router.ID,
						"type":          "router",
						"name":          router.Name,
						"management_ip": router.ManagementIP,
						"status":        router.Status,
						"vendor":        router.Vendor,
						"model":         router.Model,
					},
				}
				features = append(features, feature)
			}
		}
	}

	// Add link features as LineStrings (only if both routers have locations)
	for _, link := range links {
		sourceRouter := interfaceToRouter[link.SourceInterfaceID]
		targetRouter := interfaceToRouter[link.TargetInterfaceID]

		if sourceRouter != nil && targetRouter != nil &&
			sourceRouter.Location != nil && targetRouter.Location != nil {
			sourceGeoPoint := postGISToGeoPoint(*sourceRouter.Location)
			targetGeoPoint := postGISToGeoPoint(*targetRouter.Location)

			if sourceGeoPoint != nil && targetGeoPoint != nil {
				feature := dto.GeoFeature{
					Type: "Feature",
					Geometry: dto.GeoGeometry{
						Type: "LineString",
						Coordinates: [][]float64{
							{sourceGeoPoint.Longitude, sourceGeoPoint.Latitude},
							{targetGeoPoint.Longitude, targetGeoPoint.Latitude},
						},
					},
					Properties: map[string]interface{}{
						"id":                  link.ID,
						"type":                "link",
						"name":                link.Name,
						"source_interface_id": link.SourceInterfaceID,
						"target_interface_id": link.TargetInterfaceID,
						"source_router_id":    sourceRouter.ID,
						"target_router_id":    targetRouter.ID,
						"link_type":           link.LinkType,
						"capacity_mbps":       link.CapacityMbps,
						"latency_ms":          link.LatencyMs,
						"status":              link.Status,
					},
				}
				features = append(features, feature)
			}
		}
	}

	return &dto.GeoJSONResponse{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}
