package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/auth"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo     repository.UserRepository
	tenantRepo   repository.TenantRepository
	authProvider auth.AuthProvider
	logger       *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	authProvider auth.AuthProvider,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tenantRepo:   tenantRepo,
		authProvider: authProvider,
		logger:       logger,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login failed: user not found", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Warn("Login failed: invalid password", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check user status
	if user.Status != "active" {
		s.logger.Warn("Login failed: user not active", zap.String("email", req.Email), zap.String("status", user.Status))
		return nil, fmt.Errorf("user account is not active")
	}

	// Get tenant information
	tenant, err := s.tenantRepo.GetByID(ctx, user.TenantID)
	if err != nil {
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant information")
	}

	// Check tenant status
	if tenant.Status != "active" {
		s.logger.Warn("Login failed: tenant not active", zap.String("tenant_id", tenant.ID.String()))
		return nil, fmt.Errorf("tenant account is not active")
	}

	// Issue tokens
	tokenPair, err := s.authProvider.IssueToken(ctx, user, tenant)
	if err != nil {
		s.logger.Error("Failed to issue token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate authentication token")
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update last login", zap.Error(err))
		// Non-critical error, continue
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         toUserDTO(user),
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (dto.UserDTO, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return dto.UserDTO{}, fmt.Errorf("email already registered")
	}

	// Determine tenant
	var tenantID uuid.UUID
	if req.TenantID != "" {
		// User is invited to existing tenant
		tid, err := uuid.Parse(req.TenantID)
		if err != nil {
			return dto.UserDTO{}, fmt.Errorf("invalid tenant ID")
		}

		// Verify tenant exists and is active
		tenant, err := s.tenantRepo.GetByID(ctx, tid)
		if err != nil {
			return dto.UserDTO{}, fmt.Errorf("invalid tenant")
		}
		if tenant.Status != "active" {
			return dto.UserDTO{}, fmt.Errorf("tenant is not active")
		}

		tenantID = tid
	} else {
		// Create new tenant for this user
		tenant := &models.Tenant{
			ID:               uuid.New(),
			Name:             req.FirstName + "'s Organization",
			Slug:             fmt.Sprintf("tenant-%s", uuid.New().String()[:8]),
			ContactEmail:     req.Email,
			SubscriptionTier: "free",
			MaxDevices:       10,
			MaxUsers:         5,
			Status:           "active",
		}

		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			s.logger.Error("Failed to create tenant", zap.Error(err))
			return dto.UserDTO{}, fmt.Errorf("failed to create tenant")
		}

		tenantID = tenant.ID
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password, auth.DefaultBcryptCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return dto.UserDTO{}, fmt.Errorf("failed to process password")
	}

	// Create user
	user := &models.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		FirstName:     &req.FirstName,
		LastName:      &req.LastName,
		Status:        "active",
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return dto.UserDTO{}, fmt.Errorf("failed to create user")
	}

	s.logger.Info("User registered successfully", zap.String("email", user.Email))

	return toUserDTO(user), nil
}

// RefreshToken exchanges a refresh token for a new token pair
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	tokenPair, err := s.authProvider.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Warn("Token refresh failed", zap.Error(err))
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	return &dto.RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout revokes the user's token
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.authProvider.RevokeToken(ctx, token); err != nil {
		s.logger.Error("Failed to revoke token", zap.Error(err))
		return fmt.Errorf("failed to logout")
	}

	return nil
}
