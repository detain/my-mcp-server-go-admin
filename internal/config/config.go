// Package config provides configuration management for the MCP server.
package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the MCP server.
type Config struct {
	OpenAPISpecURL  string
	APIBaseURL      string
	SessionDir      string
	CacheDir        string
	ServerName      string
	ServerVersion   string
	BearerToken     string
	APIKey          string
	SessionID       string
	AuthServerURL   string
	ServerURL       string
}

// Load loads configuration from environment variables.
// It first loads from .env file if present, then overrides with actual environment variables.
func Load() (*Config, error) {
	// Load .env file if present (ignore errors)
	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file found, using environment variables")
	}

	config := &Config{
		OpenAPISpecURL:  getEnv("OPENAPI_SPEC_URL", ""),
		APIBaseURL:      getEnv("API_BASE_URL", ""),
		SessionDir:      getEnv("SESSION_DIR", "/tmp/mcp_admin_sessions"),
		CacheDir:        getEnv("CACHE_DIR", "/tmp/mcp_admin_cache"),
		ServerName:      getEnv("SERVER_NAME", "myadmin-admin-mcp"),
		ServerVersion:   getEnv("SERVER_VERSION", "1.0.0"),
		BearerToken:     getEnv("BEARER_TOKEN", ""),
		APIKey:          getEnv("API_KEY", ""),
		SessionID:       getEnv("SESSION_ID", ""),
		AuthServerURL:   getEnv("AUTH_SERVER_URL", ""),
		ServerURL:       getEnv("SERVER_URL", ""),
	}

	// Validate required variables
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks that required configuration values are present.
func (c *Config) Validate() error {
	if c.OpenAPISpecURL == "" {
		return fmt.Errorf("OPENAPI_SPEC_URL is required")
	}
	if c.APIBaseURL == "" {
		return fmt.Errorf("API_BASE_URL is required")
	}
	return nil
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
