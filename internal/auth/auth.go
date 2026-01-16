package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
)

// Claims represents JWT claims
type Claims struct {
	UserID      uuid.UUID   `json:"user_id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	Email       string      `json:"email"`
	Roles       []string    `json:"roles"`
	Permissions []string    `json:"permissions"`
	jwt.RegisteredClaims
}

// Service handles authentication operations
type Service struct {
	jwtSecret string
	tokenExpiry time.Duration
}

// NewService creates a new auth service
func NewService(jwtSecret string, tokenExpiryMinutes int) *Service {
	return &Service{
		jwtSecret: jwtSecret,
		tokenExpiry: time.Duration(tokenExpiryMinutes) * time.Minute,
	}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken generates a new JWT token
func (s *Service) GenerateToken(claims *Claims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(s.tokenExpiry))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	
	if err != nil {
		return nil, ErrInvalidToken
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}
