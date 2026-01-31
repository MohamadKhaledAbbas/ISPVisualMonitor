package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/utils"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/middleware"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService *service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService *service.UserService, validator *validator.Validate) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
	}
}

func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
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

	users, total, err := h.userService.ListUsers(r.Context(), tenantID, opts)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondPaginated(w, users, page, pageSize, total)
}

func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid user ID"))
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid user ID"))
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, utils.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.RespondValidationError(w, err)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal.WithDetails("User context not found"))
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondError(w, http.StatusNotFound, utils.ErrNotFound)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, utils.ErrInternal)
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}
