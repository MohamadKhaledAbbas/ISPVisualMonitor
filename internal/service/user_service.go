package service

import (
	"context"
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserService handles user business logic
type UserService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (dto.UserDTO, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return dto.UserDTO{}, fmt.Errorf("user not found")
	}

	userDTO := toUserDTO(user)
	return userDTO, nil
}

// ListUsers retrieves a list of users
func (s *UserService) ListUsers(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]dto.UserDTO, int64, error) {
	users, total, err := s.userRepo.List(ctx, tenantID, opts)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list users")
	}

	userDTOs := make([]dto.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = toUserDTO(user)
	}

	return userDTOs, total, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) (dto.UserDTO, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return dto.UserDTO{}, fmt.Errorf("user not found")
	}

	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return dto.UserDTO{}, fmt.Errorf("failed to update user")
	}

	s.logger.Info("User updated successfully", zap.String("user_id", user.ID.String()))

	userDTO := toUserDTO(user)
	return userDTO, nil
}
