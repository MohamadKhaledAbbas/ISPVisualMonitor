package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *service.AuthService, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			api.RespondError(w, http.StatusUnauthorized, api.ErrInvalidCredentials)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			api.RespondError(w, http.StatusConflict, api.ErrConflict.WithDetails("User already exists"))
			return
		}
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondCreated(w, resp)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, api.ErrBadRequest.WithDetails("Invalid request body"))
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		api.RespondValidationError(w, err)
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, api.ErrUnauthorized.WithDetails("Invalid refresh token"))
		return
	}

	api.RespondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		api.RespondError(w, http.StatusUnauthorized, api.ErrUnauthorized.WithDetails("Authorization header required"))
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		api.RespondError(w, http.StatusUnauthorized, api.ErrUnauthorized.WithDetails("Invalid authorization header format"))
		return
	}

	token := parts[1]
	if err := h.authService.Logout(r.Context(), token); err != nil {
		api.RespondError(w, http.StatusInternalServerError, api.ErrInternal)
		return
	}

	api.RespondNoContent(w)
}
