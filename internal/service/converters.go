package service

import (
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

// toUserDTO converts a User model to UserDTO
func toUserDTO(user *models.User) dto.UserDTO {
	return dto.UserDTO{
		ID:            user.ID,
		TenantID:      user.TenantID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
