package auth

import (
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
)

// NewAuthProvider creates an auth provider based on configuration
// Supports multiple provider types for different deployment scenarios
func NewAuthProvider(cfg *config.AuthConfig) (AuthProvider, error) {
	switch cfg.Provider {
	case "local":
		return NewLocalProvider(cfg)
		
	case "oidc", "keycloak", "auth0":
		// Stub for future OIDC implementation
		return NewOIDCProvider(cfg)
		
	default:
		return nil, fmt.Errorf("unknown auth provider: %s", cfg.Provider)
	}
}

// NewOIDCProvider is a stub for future OIDC provider implementation
// This will support external identity providers like Keycloak, Auth0, etc.
func NewOIDCProvider(cfg *config.AuthConfig) (AuthProvider, error) {
	return nil, fmt.Errorf("OIDC provider not yet implemented")
}
