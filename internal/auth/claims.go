package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims structure
// Includes standard JWT claims plus custom claims for multi-tenant RBAC
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	TokenType   string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}
