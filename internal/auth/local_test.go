package auth

import (
	"context"
	"testing"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
)

func getTestConfig() *config.AuthConfig {
	return &config.AuthConfig{
		Provider:         "local",
		JWTSecret:        "test-secret-key-for-unit-tests-only",
		JWTSigningMethod: "HS256",
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		Issuer:           "ispvisualmonitor-test",
		BcryptCost:       10, // Lower cost for faster tests
	}
}

func getTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
		Status:   "active",
	}
}

func getTestTenant() *models.Tenant {
	return &models.Tenant{
		ID:   uuid.New(),
		Name: "Test Tenant",
		Slug: "test-tenant",
	}
}

func TestNewLocalProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.AuthConfig
		expectError bool
	}{
		{
			name:        "valid HS256 config",
			config:      getTestConfig(),
			expectError: false,
		},
		{
			name: "valid RS256 config without keys (falls back to HS256)",
			config: &config.AuthConfig{
				Provider:         "local",
				JWTSecret:        "test-secret",
				JWTSigningMethod: "RS256",
				AccessTokenTTL:   15 * time.Minute,
				RefreshTokenTTL:  7 * 24 * time.Hour,
				Issuer:           "test",
			},
			expectError: false,
		},
		{
			name: "invalid signing method",
			config: &config.AuthConfig{
				Provider:         "local",
				JWTSecret:        "test-secret",
				JWTSigningMethod: "INVALID",
				AccessTokenTTL:   15 * time.Minute,
				RefreshTokenTTL:  7 * 24 * time.Hour,
				Issuer:           "test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewLocalProvider(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("expected non-nil provider")
			}

			if provider.GetProviderName() != "local" {
				t.Errorf("expected provider name 'local', got %s", provider.GetProviderName())
			}
		})
	}
}

func TestLocalProvider_IssueToken(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	tokenPair, err := provider.IssueToken(ctx, user, tenant)

	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	if tokenPair == nil {
		t.Fatal("expected non-nil token pair")
	}

	if tokenPair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	if tokenPair.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %s", tokenPair.TokenType)
	}

	if tokenPair.ExpiresIn <= 0 {
		t.Errorf("expected positive ExpiresIn, got %d", tokenPair.ExpiresIn)
	}
}

func TestLocalProvider_ValidateToken(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	tokenPair, err := provider.IssueToken(ctx, user, tenant)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Test valid token
	claims, err := provider.ValidateToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.TenantID != user.TenantID {
		t.Errorf("expected tenant ID %s, got %s", user.TenantID, claims.TenantID)
	}

	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.TokenType != "access" {
		t.Errorf("expected token type 'access', got %s", claims.TokenType)
	}

	// Test invalid token
	_, err = provider.ValidateToken(ctx, "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}

	// Test expired token
	expiredCfg := getTestConfig()
	expiredCfg.AccessTokenTTL = -1 * time.Hour // Already expired
	expiredProvider, _ := NewLocalProvider(expiredCfg)
	expiredTokenPair, _ := expiredProvider.IssueToken(ctx, user, tenant)
	
	_, err = expiredProvider.ValidateToken(ctx, expiredTokenPair.AccessToken)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestLocalProvider_RefreshToken(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	tokenPair, err := provider.IssueToken(ctx, user, tenant)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Test valid refresh
	newTokenPair, err := provider.RefreshToken(ctx, tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("failed to refresh token: %v", err)
	}

	if newTokenPair == nil {
		t.Fatal("expected non-nil token pair")
	}

	if newTokenPair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	if newTokenPair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	// New tokens should be different
	if newTokenPair.AccessToken == tokenPair.AccessToken {
		t.Error("expected new access token to be different")
	}

	// Test refresh with access token (should fail)
	_, err = provider.RefreshToken(ctx, tokenPair.AccessToken)
	if err != ErrInvalidTokenType {
		t.Errorf("expected ErrInvalidTokenType, got %v", err)
	}

	// Test refresh with invalid token
	_, err = provider.RefreshToken(ctx, "invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestLocalProvider_RevokeToken(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	tokenPair, err := provider.IssueToken(ctx, user, tenant)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Token should be valid before revocation
	_, err = provider.ValidateToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("token should be valid before revocation: %v", err)
	}

	// Revoke token
	err = provider.RevokeToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("failed to revoke token: %v", err)
	}

	// Token should be invalid after revocation
	_, err = provider.ValidateToken(ctx, tokenPair.AccessToken)
	if err != ErrRevokedToken {
		t.Errorf("expected ErrRevokedToken after revocation, got %v", err)
	}

	// Test revoking invalid token
	err = provider.RevokeToken(ctx, "invalid-token")
	if err == nil {
		t.Error("expected error when revoking invalid token")
	}
}

func TestLocalProvider_MultiTenantIsolation(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Create users from different tenants
	tenant1 := getTestTenant()
	tenant2 := getTestTenant()

	user1 := getTestUser()
	user1.TenantID = tenant1.ID
	user1.Email = "user1@tenant1.com"

	user2 := getTestUser()
	user2.TenantID = tenant2.ID
	user2.Email = "user2@tenant2.com"

	ctx := context.Background()

	// Issue tokens for both users
	token1, err := provider.IssueToken(ctx, user1, tenant1)
	if err != nil {
		t.Fatalf("failed to issue token for user1: %v", err)
	}

	token2, err := provider.IssueToken(ctx, user2, tenant2)
	if err != nil {
		t.Fatalf("failed to issue token for user2: %v", err)
	}

	// Validate tokens
	claims1, err := provider.ValidateToken(ctx, token1.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate token1: %v", err)
	}

	claims2, err := provider.ValidateToken(ctx, token2.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate token2: %v", err)
	}

	// Verify tenant isolation
	if claims1.TenantID == claims2.TenantID {
		t.Error("expected different tenant IDs")
	}

	if claims1.TenantID != tenant1.ID {
		t.Errorf("token1 should have tenant1 ID")
	}

	if claims2.TenantID != tenant2.ID {
		t.Errorf("token2 should have tenant2 ID")
	}

	// Verify user isolation
	if claims1.UserID == claims2.UserID {
		t.Error("expected different user IDs")
	}

	if claims1.Email == claims2.Email {
		t.Error("expected different emails")
	}
}

func TestLocalProvider_TokenExpiration(t *testing.T) {
	// Create provider with very short token lifetime
	cfg := getTestConfig()
	cfg.AccessTokenTTL = 1 * time.Second

	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	tokenPair, err := provider.IssueToken(ctx, user, tenant)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Token should be valid immediately
	_, err = provider.ValidateToken(ctx, tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("token should be valid immediately: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should be expired now
	_, err = provider.ValidateToken(ctx, tokenPair.AccessToken)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestLocalProvider_ConcurrentAccess(t *testing.T) {
	cfg := getTestConfig()
	provider, err := NewLocalProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID

	ctx := context.Background()
	
	// Issue token
	tokenPair, err := provider.IssueToken(ctx, user, tenant)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Concurrent validation
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := provider.ValidateToken(ctx, tokenPair.AccessToken)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent validation error: %v", err)
	}
}

// Benchmark token operations
func BenchmarkLocalProvider_IssueToken(b *testing.B) {
	cfg := getTestConfig()
	provider, _ := NewLocalProvider(cfg)
	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.IssueToken(ctx, user, tenant)
	}
}

func BenchmarkLocalProvider_ValidateToken(b *testing.B) {
	cfg := getTestConfig()
	provider, _ := NewLocalProvider(cfg)
	user := getTestUser()
	tenant := getTestTenant()
	user.TenantID = tenant.ID
	ctx := context.Background()
	tokenPair, _ := provider.IssueToken(ctx, user, tenant)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.ValidateToken(ctx, tokenPair.AccessToken)
	}
}
