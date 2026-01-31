package handlers

import (
	"net/http"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TopologyHandler struct {
	topologyService *service.TopologyService
	validator       *validator.Validate
}

func NewTopologyHandler(topologyService *service.TopologyService, validator *validator.Validate) *TopologyHandler {
	return &TopologyHandler{
		topologyService: topologyService,
		validator:       validator,
	}
}

func (h *TopologyHandler) HandleGetTopology(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	topology, err := h.topologyService.GetTopology(r.Context(), tenantID)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, topology)
}

func (h *TopologyHandler) HandleGetTopologyGeoJSON(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	geoJSON, err := h.topologyService.GetTopologyGeoJSON(r.Context(), tenantID)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, geoJSON)
}
