package handlers

import (
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/utils"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type InterfaceHandler struct {
	interfaceService *service.InterfaceService
	validator        *validator.Validate
}

func NewInterfaceHandler(interfaceService *service.InterfaceService, validator *validator.Validate) *InterfaceHandler {
	return &InterfaceHandler{
		interfaceService: interfaceService,
		validator:        validator,
	}
}

func (h *InterfaceHandler) HandleListInterfaces(w http.ResponseWriter, r *http.Request) {
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

	interfaces, total, err := h.interfaceService.ListInterfaces(r.Context(), tenantID, opts)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondPaginated(w, interfaces, page, pageSize, total)
}

func (h *InterfaceHandler) HandleListRouterInterfaces(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	vars := mux.Vars(r)
	routerID, err := uuid.Parse(vars["router_id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid router ID"))
		return
	}

	page, pageSize := parsePagination(r)
	opts := repository.ListOptions{
		Page:     page,
		PageSize: pageSize,
	}

	interfaces, total, err := h.interfaceService.ListRouterInterfaces(r.Context(), tenantID, routerID, opts)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondPaginated(w, interfaces, page, pageSize, total)
}
