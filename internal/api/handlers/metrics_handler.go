package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type MetricsHandler struct {
	metricsService *service.MetricsService
	validator      *validator.Validate
}

func NewMetricsHandler(metricsService *service.MetricsService, validator *validator.Validate) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
		validator:      validator,
	}
}

func (h *MetricsHandler) HandleGetInterfaceMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	interfaceID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid interface ID"))
		return
	}

	from, to := parseTimeRange(r)

	metrics, err := h.metricsService.GetInterfaceMetrics(r.Context(), interfaceID, from, to)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, metrics)
}

func (h *MetricsHandler) HandleGetRouterMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routerID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid router ID"))
		return
	}

	from, to := parseTimeRange(r)

	metrics, err := h.metricsService.GetRouterMetrics(r.Context(), routerID, from, to)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, metrics)
}

func parseTimeRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	to := now
	from := now.Add(-1 * time.Hour)

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	return from, to
}
