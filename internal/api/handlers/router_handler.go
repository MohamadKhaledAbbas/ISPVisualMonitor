package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RouterHandler struct {
	routerService *service.RouterService
	validator     *validator.Validate
}

func NewRouterHandler(routerService *service.RouterService, validator *validator.Validate) *RouterHandler {
	return &RouterHandler{
		routerService: routerService,
		validator:     validator,
	}
}

func (h *RouterHandler) HandleListRouters(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	page, pageSize := parsePagination(r)
	opts := repository.ListOptions{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	}

	routers, total, err := h.routerService.ListRouters(r.Context(), tenantID, opts)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondPaginated(w, routers, page, pageSize, total)
}

func (h *RouterHandler) HandleCreateRouter(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal.WithDetails("Tenant context not found"))
		return
	}

	var req dto.CreateRouterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	router, err := h.routerService.CreateRouter(r.Context(), tenantID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			api.RespondError(w, http.StatusConflict, api.ErrConflict.WithDetails("Router already exists"))
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondCreated(w, router)
}

func (h *RouterHandler) HandleGetRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routerID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid router ID"))
		return
	}

	router, err := h.routerService.GetRouter(r.Context(), routerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, router)
}

func (h *RouterHandler) HandleUpdateRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routerID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid router ID"))
		return
	}

	var req dto.UpdateRouterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	router, err := h.routerService.UpdateRouter(r.Context(), routerID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, router)
}

func (h *RouterHandler) HandleDeleteRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routerID, err := uuid.Parse(vars["id"])
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid router ID"))
		return
	}

	if err := h.routerService.DeleteRouter(r.Context(), routerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.RespondError(w, http.StatusNotFound, api.ErrNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondNoContent(w)
}

func parsePagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	return page, pageSize
}
