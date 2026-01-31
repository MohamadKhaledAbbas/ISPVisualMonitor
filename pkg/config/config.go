package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	API      APIConfig
	Database DatabaseConfig
	Poller   PollerConfig
	Auth     AuthConfig
}

// APIConfig holds API server configuration
type APIConfig struct {
	Port            int
	JWTSecret       string
	AllowedOrigins  []string
	RateLimitPerMin int
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

// PollerConfig holds router poller configuration
type PollerConfig struct {
	WorkerCount      int
	DefaultInterval  int // seconds
	TimeoutSeconds   int
	RetryAttempts    int
	ConcurrentPolls  int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Provider         string        // local, keycloak, auth0, oidc
	JWTSecret        string        // Secret for HS256 signing
	JWTSigningMethod string        // HS256 or RS256
	JWTPrivateKey    string        // For RS256 (PEM format)
	JWTPublicKey     string        // For RS256 (PEM format)
	AccessTokenTTL   time.Duration // Access token expiration time
	RefreshTokenTTL  time.Duration // Refresh token expiration time
	Issuer           string        // JWT issuer identifier
	BcryptCost       int           // Bcrypt cost for password hashing
	
	// For external OIDC providers (Keycloak, Auth0, etc.)
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	
	// Legacy fields (deprecated, but kept for backward compatibility)
	TokenExpiry        int // minutes (deprecated - use AccessTokenTTL)
	RefreshTokenExpiry int // days (deprecated - use RefreshTokenTTL)
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		API: APIConfig{
			Port:            getEnvInt("API_PORT", 8080),
			JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
			AllowedOrigins:  []string{getEnv("ALLOWED_ORIGINS", "*")},
			RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 100),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "ispmonitor"),
			Password: getEnv("DB_PASSWORD", "ispmonitor"),
			DBName:   getEnv("DB_NAME", "ispmonitor"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 25),
			MinConns: getEnvInt("DB_MIN_CONNS", 5),
		},
		Poller: PollerConfig{
			WorkerCount:      getEnvInt("POLLER_WORKERS", 10),
			DefaultInterval:  getEnvInt("POLLER_INTERVAL", 300),
			TimeoutSeconds:   getEnvInt("POLLER_TIMEOUT", 30),
			RetryAttempts:    getEnvInt("POLLER_RETRY", 3),
			ConcurrentPolls:  getEnvInt("POLLER_CONCURRENT", 50),
		},
		Auth: AuthConfig{
			Provider:         getEnv("AUTH_PROVIDER", "local"),
			JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
			JWTSigningMethod: getEnv("JWT_SIGNING_METHOD", "HS256"),
			JWTPrivateKey:    getEnv("JWT_PRIVATE_KEY", ""),
			JWTPublicKey:     getEnv("JWT_PUBLIC_KEY", ""),
			AccessTokenTTL:   getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:  getEnvDuration("REFRESH_TOKEN_TTL", 168*time.Hour), // 7 days
			Issuer:           getEnv("JWT_ISSUER", "ispvisualmonitor"),
			BcryptCost:       getEnvInt("BCRYPT_COST", 12),
			
			// OIDC configuration
			OIDCIssuerURL:    getEnv("OIDC_ISSUER_URL", ""),
			OIDCClientID:     getEnv("OIDC_CLIENT_ID", ""),
			OIDCClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
			
			// Legacy fields for backward compatibility
			TokenExpiry:        getEnvInt("TOKEN_EXPIRY_MIN", 60),
			RefreshTokenExpiry: getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		},
	}

	// Validate required fields
	if cfg.API.JWTSecret == "change-me-in-production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

// ConnectionString returns PostgreSQL connection string
func (db *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
