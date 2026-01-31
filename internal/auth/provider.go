package auth

import (
	"context"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

// AuthProvider defines the interface for authentication providers
// Implementations can be local (in-memory/file-based) or external (Keycloak, Auth0, OIDC)
type AuthProvider interface {
	// IssueToken creates a new JWT token for authenticated user
	IssueToken(ctx context.Context, user *models.User, tenant *models.Tenant) (*TokenPair, error)

	// ValidateToken verifies and parses a JWT token
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)

	// RefreshToken exchanges a refresh token for new token pair
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// RevokeToken invalidates a token (for logout)
	RevokeToken(ctx context.Context, tokenString string) error

	// GetProviderName returns the provider identifier
	GetProviderName() string
}
