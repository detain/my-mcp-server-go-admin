package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("OPENAPI_SPEC_URL")
	os.Unsetenv("API_BASE_URL")

	_, err := Load()
	if err == nil {
		t.Error("expected error when required vars are missing")
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set required env vars
	os.Setenv("OPENAPI_SPEC_URL", "https://example.com/spec.yaml")
	os.Setenv("API_BASE_URL", "https://example.com/api")

	defer func() {
		os.Unsetenv("OPENAPI_SPEC_URL")
		os.Unsetenv("API_BASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.OpenAPISpecURL != "https://example.com/spec.yaml" {
		t.Errorf("expected OpenAPISpecURL 'https://example.com/spec.yaml', got '%s'", cfg.OpenAPISpecURL)
	}

	if cfg.APIBaseURL != "https://example.com/api" {
		t.Errorf("expected APIBaseURL 'https://example.com/api', got '%s'", cfg.APIBaseURL)
	}
}

func TestDefaultValues(t *testing.T) {
	os.Setenv("OPENAPI_SPEC_URL", "https://example.com/spec.yaml")
	os.Setenv("API_BASE_URL", "https://example.com/api")

	defer func() {
		os.Unsetenv("OPENAPI_SPEC_URL")
		os.Unsetenv("API_BASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check default values
	if cfg.SessionDir != "/tmp/mcp_admin_sessions" {
		t.Errorf("expected default SessionDir '/tmp/mcp_admin_sessions', got '%s'", cfg.SessionDir)
	}

	if cfg.CacheDir != "/tmp/mcp_admin_cache" {
		t.Errorf("expected default CacheDir '/tmp/mcp_admin_cache', got '%s'", cfg.CacheDir)
	}

	if cfg.ServerName != "myadmin-admin-mcp" {
		t.Errorf("expected default ServerName 'myadmin-admin-mcp', got '%s'", cfg.ServerName)
	}

	if cfg.ServerVersion != "1.0.0" {
		t.Errorf("expected default ServerVersion '1.0.0', got '%s'", cfg.ServerVersion)
	}
}

func TestOptionalVars(t *testing.T) {
	os.Setenv("OPENAPI_SPEC_URL", "https://example.com/spec.yaml")
	os.Setenv("API_BASE_URL", "https://example.com/api")
	os.Setenv("BEARER_TOKEN", "test_token")
	os.Setenv("API_KEY", "test_key")

	defer func() {
		os.Unsetenv("OPENAPI_SPEC_URL")
		os.Unsetenv("API_BASE_URL")
		os.Unsetenv("BEARER_TOKEN")
		os.Unsetenv("API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BearerToken != "test_token" {
		t.Errorf("expected BearerToken 'test_token', got '%s'", cfg.BearerToken)
	}

	if cfg.APIKey != "test_key" {
		t.Errorf("expected APIKey 'test_key', got '%s'", cfg.APIKey)
	}
}
