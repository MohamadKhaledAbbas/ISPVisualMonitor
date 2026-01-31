// Package config provides configuration management for the ISP Visual Monitor application.
package config

import (
	"os"
	"strconv"
)

// DeploymentMode represents the deployment environment
type DeploymentMode string

const (
	// DeploymentModeDevelopment is for local development
	DeploymentModeDevelopment DeploymentMode = "development"
	// DeploymentModeProduction is for cloud-hosted production
	DeploymentModeProduction DeploymentMode = "production"
	// DeploymentModeOnPremise is for on-premise self-hosted deployments
	DeploymentModeOnPremise DeploymentMode = "on-premise"
)

// DeploymentConfig holds deployment-specific configuration
type DeploymentConfig struct {
	Mode      DeploymentMode `env:"DEPLOYMENT_MODE" envDefault:"development"`
	CloudMode bool           `env:"CLOUD_MODE" envDefault:"false"`

	// License validation
	LicenseKey       string `env:"LICENSE_KEY"`
	LicenseServerURL string `env:"LICENSE_SERVER_URL" envDefault:"https://license.ispmonitor.com/v1"`

	// Feature flags
	EnableMetrics   bool `env:"ENABLE_METRICS" envDefault:"true"`
	EnableTracing   bool `env:"ENABLE_TRACING" envDefault:"false"`
	EnableProfiling bool `env:"ENABLE_PROFILING" envDefault:"false"`

	// Logging
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`
}

// LoadDeployment loads deployment configuration from environment variables
func LoadDeployment() *DeploymentConfig {
	return &DeploymentConfig{
		Mode:             DeploymentMode(getEnvWithDefault("DEPLOYMENT_MODE", "development")),
		CloudMode:        getEnvBool("CLOUD_MODE", false),
		LicenseKey:       os.Getenv("LICENSE_KEY"),
		LicenseServerURL: getEnvWithDefault("LICENSE_SERVER_URL", "https://license.ispmonitor.com/v1"),
		EnableMetrics:    getEnvBool("ENABLE_METRICS", true),
		EnableTracing:    getEnvBool("ENABLE_TRACING", false),
		EnableProfiling:  getEnvBool("ENABLE_PROFILING", false),
		LogLevel:         getEnvWithDefault("LOG_LEVEL", "info"),
		LogFormat:        getEnvWithDefault("LOG_FORMAT", "json"),
	}
}

// IsProduction returns true if running in production mode
func (c *DeploymentConfig) IsProduction() bool {
	return c.Mode == DeploymentModeProduction
}

// IsOnPremise returns true if running in on-premise mode
func (c *DeploymentConfig) IsOnPremise() bool {
	return c.Mode == DeploymentModeOnPremise
}

// IsDevelopment returns true if running in development mode
func (c *DeploymentConfig) IsDevelopment() bool {
	return c.Mode == DeploymentModeDevelopment
}

// RequiresLicense returns true if license validation is required
func (c *DeploymentConfig) RequiresLicense() bool {
	return c.IsProduction() || c.IsOnPremise()
}

// Helper functions

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
