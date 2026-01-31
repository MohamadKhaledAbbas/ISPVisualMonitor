package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// APIError represents a structured API error response
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Predefined error codes
var (
	ErrNotFound           = &APIError{Code: "NOT_FOUND", Message: "Resource not found"}
	ErrUnauthorized       = &APIError{Code: "UNAUTHORIZED", Message: "Authentication required"}
	ErrForbidden          = &APIError{Code: "FORBIDDEN", Message: "Permission denied"}
	ErrValidation         = &APIError{Code: "VALIDATION_ERROR", Message: "Validation failed"}
	ErrInternal           = &APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrConflict           = &APIError{Code: "CONFLICT", Message: "Resource already exists"}
	ErrBadRequest         = &APIError{Code: "BAD_REQUEST", Message: "Invalid request"}
	ErrInvalidCredentials = &APIError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password"}
)

// NewAPIError creates a new API error with custom message
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// WithDetails adds details to an API error
func (e *APIError) WithDetails(details interface{}) *APIError {
	return &APIError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator errors to ValidationError slice
func FormatValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fe.Field(),
				Message: formatFieldError(fe),
			})
		}
	}

	return validationErrors
}

// formatFieldError formats a single field error
func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, statusCode int, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

// RespondValidationError sends a validation error response
func RespondValidationError(w http.ResponseWriter, err error) {
	validationErrors := FormatValidationErrors(err)
	RespondError(w, http.StatusBadRequest, ErrValidation.WithDetails(validationErrors))
}
