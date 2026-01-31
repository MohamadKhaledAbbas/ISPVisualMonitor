package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/utils"
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
		Page:     page,
		PageSize: pageSize,
	}

	tenants, total, err := h.tenantService.ListTenants(r.Context(), opts)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondPaginated(w, tenants, page, pageSize, total)
}

func (h *TenantHandler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.RespondValidationError(w, err)
		return
	}

	tenant, err := h.tenantService.CreateTenant(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.RespondError(w, http.StatusConflict, utils.ErrConflict.WithDetails("Tenant already exists"))
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondCreated(w, tenant)
}

func (h *TenantHandler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid tenant ID"))
		return
	}

	tenant, err := h.tenantService.GetTenant(r.Context(), tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondJSON(w, http.StatusOK, tenant)
}

func (h *TenantHandler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid tenant ID"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.RespondValidationError(w, err)
		return
	}

	tenant, err := h.tenantService.UpdateTenant(r.Context(), tenantID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondJSON(w, http.StatusOK, tenant)
}
