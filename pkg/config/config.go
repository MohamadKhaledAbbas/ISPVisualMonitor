package config

import (
	"fmt"
	"os"
	"strconv"
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
	TokenExpiry        int // minutes
	RefreshTokenExpiry int // days
	BcryptCost         int
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
			TokenExpiry:        getEnvInt("TOKEN_EXPIRY_MIN", 60),
			RefreshTokenExpiry: getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
			BcryptCost:         getEnvInt("BCRYPT_COST", 12),
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
