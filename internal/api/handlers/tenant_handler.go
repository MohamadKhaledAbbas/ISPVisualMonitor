package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TenantHandler struct {
	tenantService *service.TenantService
	validator     *validator.Validate
}

func NewTenantHandler(tenantService *service.TenantService, validator *validator.Validate) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		validator:     validator,
	}
}

func (h *TenantHandler) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	opts := repository.ListOptions{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	}

	tenants, total, err := h.tenantService.ListTenants(r.Context(), opts)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondPaginated(w, tenants, page, pageSize, total)
}

func (h *TenantHandler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	tenant, err := h.tenantService.CreateTenant(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			api.RespondError(w, http.StatusConflict, api.ErrConflict.WithDetails("Tenant already exists"))
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondCreated(w, tenant)
}

func (h *TenantHandler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid tenant ID"))
		return
	}

	tenant, err := h.tenantService.GetTenant(r.Context(), tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, tenant)
}

func (h *TenantHandler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid tenant ID"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	tenant, err := h.tenantService.UpdateTenant(r.Context(), tenantID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, tenant)
}
