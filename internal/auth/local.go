package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
	
	// ErrRevokedToken is returned when token has been revoked
	ErrRevokedToken = errors.New("token has been revoked")
	
	// ErrInvalidTokenType is returned when token type is invalid
	ErrInvalidTokenType = errors.New("invalid token type")
	
	// ErrInvalidSigningMethod is returned when signing method is not supported
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

// LocalProvider implements AuthProvider using local JWT tokens
// Supports both symmetric (HS256) and asymmetric (RS256) signing
type LocalProvider struct {
	config     *config.AuthConfig
	blacklist  TokenBlacklist
	signingKey interface{} // []byte for HS256, *rsa.PrivateKey for RS256
	verifyKey  interface{} // []byte for HS256, *rsa.PublicKey for RS256
	signingMethod jwt.SigningMethod
}

// NewLocalProvider creates a new local JWT provider
func NewLocalProvider(cfg *config.AuthConfig) (*LocalProvider, error) {
	provider := &LocalProvider{
		config:    cfg,
		blacklist: NewInMemoryBlacklist(),
	}
	
	// Setup signing method and keys
	switch cfg.JWTSigningMethod {
	case "HS256":
		provider.signingMethod = jwt.SigningMethodHS256
		provider.signingKey = []byte(cfg.JWTSecret)
		provider.verifyKey = []byte(cfg.JWTSecret)
		
	case "RS256":
		provider.signingMethod = jwt.SigningMethodRS256
		
		// For RS256, we need private and public keys
		// In production, these should be loaded from files or secrets manager
		// For now, we'll use the JWTSecret as a fallback to HS256
		if cfg.JWTPrivateKey != "" && cfg.JWTPublicKey != "" {
			privateKey, err := parseRSAPrivateKey(cfg.JWTPrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
			}
			
			publicKey, err := parseRSAPublicKey(cfg.JWTPublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
			}
			
			provider.signingKey = privateKey
			provider.verifyKey = publicKey
		} else {
			// Fallback to HS256 if keys are not provided
			provider.signingMethod = jwt.SigningMethodHS256
			provider.signingKey = []byte(cfg.JWTSecret)
			provider.verifyKey = []byte(cfg.JWTSecret)
		}
		
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidSigningMethod, cfg.JWTSigningMethod)
	}
	
	return provider, nil
}

// IssueToken creates a new JWT token pair for authenticated user
func (p *LocalProvider) IssueToken(ctx context.Context, user *models.User, tenant *models.Tenant) (*TokenPair, error) {
	now := time.Now()
	
	// Create access token
	accessClaims := &Claims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Roles:     []string{}, // TODO: Load from database
		Permissions: []string{}, // TODO: Load from database
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID.String(),
			Issuer:    p.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.config.AccessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	
	accessToken := jwt.NewWithClaims(p.signingMethod, accessClaims)
	accessTokenString, err := accessToken.SignedString(p.signingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	
	// Create refresh token
	refreshClaims := &Claims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID.String(),
			Issuer:    p.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.config.RefreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	
	refreshToken := jwt.NewWithClaims(p.signingMethod, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(p.signingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(p.config.AccessTokenTTL.Seconds()),
	}, nil
}

// ValidateToken verifies and parses a JWT token
func (p *LocalProvider) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if token.Method != p.signingMethod {
			return nil, fmt.Errorf("%w: expected %s, got %s", 
				ErrInvalidSigningMethod, p.signingMethod.Alg(), token.Method.Alg())
		}
		return p.verifyKey, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	
	// Check if token is blacklisted
	isBlacklisted, err := p.blacklist.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	
	if isBlacklisted {
		return nil, ErrRevokedToken
	}
	
	return claims, nil
}

// RefreshToken exchanges a refresh token for new token pair
func (p *LocalProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate the refresh token
	claims, err := p.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	
	// Verify it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}
	
	// Create a mock user from claims
	// In production, you might want to fetch fresh user data from database
	user := &models.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
	}
	
	tenant := &models.Tenant{
		ID: claims.TenantID,
	}
	
	// Issue new token pair
	return p.IssueToken(ctx, user, tenant)
}

// RevokeToken invalidates a token by adding it to the blacklist
func (p *LocalProvider) RevokeToken(ctx context.Context, tokenString string) error {
	// Parse token to get claims (without validation for revocation)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return ErrInvalidToken
	}
	
	// Add to blacklist until expiration
	return p.blacklist.Add(ctx, claims.ID, claims.ExpiresAt.Time)
}

// GetProviderName returns the provider identifier
func (p *LocalProvider) GetProviderName() string {
	return "local"
}

// Helper functions for RSA key parsing

func parseRSAPrivateKey(keyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	
	return rsaKey, nil
}

func parseRSAPublicKey(keyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}
	
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	
	return rsaKey, nil
}
