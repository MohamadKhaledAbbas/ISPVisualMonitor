package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestClaims_Structure(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	email := "test@example.com"
	roles := []string{"admin", "user"}
	permissions := []string{"read", "write", "delete"}

	claims := &Claims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID.String(),
			Issuer:    "test-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	// Verify all fields are set correctly
	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.TenantID != tenantID {
		t.Errorf("expected TenantID %s, got %s", tenantID, claims.TenantID)
	}

	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}

	if len(claims.Roles) != len(roles) {
		t.Errorf("expected %d roles, got %d", len(roles), len(claims.Roles))
	}

	if len(claims.Permissions) != len(permissions) {
		t.Errorf("expected %d permissions, got %d", len(permissions), len(claims.Permissions))
	}

	if claims.TokenType != "access" {
		t.Errorf("expected TokenType 'access', got %s", claims.TokenType)
	}
}

func TestClaims_TokenTypes(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
	}{
		{"access token", "access"},
		{"refresh token", "refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				UserID:    uuid.New(),
				TenantID:  uuid.New(),
				Email:     "test@example.com",
				TokenType: tt.tokenType,
			}

			if claims.TokenType != tt.tokenType {
				t.Errorf("expected TokenType %s, got %s", tt.tokenType, claims.TokenType)
			}
		})
	}
}

func TestClaims_EmptyRolesAndPermissions(t *testing.T) {
	claims := &Claims{
		UserID:      uuid.New(),
		TenantID:    uuid.New(),
		Email:       "test@example.com",
		Roles:       []string{},
		Permissions: []string{},
		TokenType:   "access",
	}

	if claims.Roles == nil {
		t.Error("Roles should not be nil")
	}

	if claims.Permissions == nil {
		t.Error("Permissions should not be nil")
	}

	if len(claims.Roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(claims.Roles))
	}

	if len(claims.Permissions) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(claims.Permissions))
	}
}

func TestClaims_MultipleRolesAndPermissions(t *testing.T) {
	roles := []string{"admin", "user", "moderator", "viewer"}
	permissions := []string{"read", "write", "delete", "update", "create"}

	claims := &Claims{
		UserID:      uuid.New(),
		TenantID:    uuid.New(),
		Email:       "test@example.com",
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
	}

	if len(claims.Roles) != len(roles) {
		t.Errorf("expected %d roles, got %d", len(roles), len(claims.Roles))
	}

	if len(claims.Permissions) != len(permissions) {
		t.Errorf("expected %d permissions, got %d", len(permissions), len(claims.Permissions))
	}

	// Verify all roles are present
	for i, role := range roles {
		if claims.Roles[i] != role {
			t.Errorf("expected role %s at position %d, got %s", role, i, claims.Roles[i])
		}
	}

	// Verify all permissions are present
	for i, perm := range permissions {
		if claims.Permissions[i] != perm {
			t.Errorf("expected permission %s at position %d, got %s", perm, i, claims.Permissions[i])
		}
	}
}

func TestClaims_RegisteredClaims(t *testing.T) {
	now := time.Now()
	jti := uuid.New().String()
	userID := uuid.New()
	issuer := "test-issuer"

	claims := &Claims{
		UserID:    userID,
		TenantID:  uuid.New(),
		Email:     "test@example.com",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID.String(),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	if claims.ID != jti {
		t.Errorf("expected ID %s, got %s", jti, claims.ID)
	}

	if claims.Subject != userID.String() {
		t.Errorf("expected Subject %s, got %s", userID.String(), claims.Subject)
	}

	if claims.Issuer != issuer {
		t.Errorf("expected Issuer %s, got %s", issuer, claims.Issuer)
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt should not be nil")
	}

	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}

	if claims.NotBefore == nil {
		t.Error("NotBefore should not be nil")
	}
}
