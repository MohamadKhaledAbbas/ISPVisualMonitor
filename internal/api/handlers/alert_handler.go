package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/utils"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AlertHandler struct {
	alertService *service.AlertService
	validator    *validator.Validate
}

func NewAlertHandler(alertService *service.AlertService, validator *validator.Validate) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		validator:    validator,
	}
}

func (h *AlertHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	page, pageSize := parsePagination(r)
	opts := repository.ListOptions{
		Page:     page,
		PageSize: pageSize,
	}

	alerts, total, err := h.alertService.ListAlerts(r.Context(), tenantID, opts)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondPaginated(w, alerts, page, pageSize, total)
}

func (h *AlertHandler) HandleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid alert ID"))
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal.WithDetails("User context not found"))
		return
	}

	var req dto.AcknowledgeAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.RespondValidationError(w, err)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	err = h.alertService.AcknowledgeAlert(r.Context(), tenantID, alertID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondNoContent(w)
}
