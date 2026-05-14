package proxy

import (
	"net/http"
	"testing"
)

func TestAuthExtractorFromHeaders(t *testing.T) {
	e := NewAuthExtractor("", "", "")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Authorization", "Bearer test_token")

	headers := e.Extract(req)

	if headers["Authorization"] != "Bearer test_token" {
		t.Errorf("expected Authorization header 'Bearer test_token', got '%s'", headers["Authorization"])
	}
}

func TestAuthExtractorAPIKey(t *testing.T) {
	e := NewAuthExtractor("", "", "")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-API-KEY", "test_api_key")

	headers := e.Extract(req)

	if headers["X-API-KEY"] != "test_api_key" {
		t.Errorf("expected X-API-KEY header 'test_api_key', got '%s'", headers["X-API-KEY"])
	}
}

func TestAuthExtractorSessionID(t *testing.T) {
	e := NewAuthExtractor("", "", "")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("sessionid", "test_session_id")

	headers := e.Extract(req)

	if headers["sessionid"] != "test_session_id" {
		t.Errorf("expected sessionid header 'test_session_id', got '%s'", headers["sessionid"])
	}
}

func TestAuthExtractorFallbackToStdio(t *testing.T) {
	e := NewAuthExtractor("bearer_from_env", "api_key_from_env", "session_from_env")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	// No auth headers set

	headers := e.Extract(req)

	if headers["Authorization"] != "Bearer bearer_from_env" {
		t.Errorf("expected fallback Authorization header, got '%s'", headers["Authorization"])
	}
}

func TestAuthExtractorPriority(t *testing.T) {
	e := NewAuthExtractor("bearer_fallback", "api_key_fallback", "")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Authorization", "Bearer from_request")

	headers := e.Extract(req)

	// Request header should take priority over fallback
	if headers["Authorization"] != "Bearer from_request" {
		t.Errorf("expected Authorization from request, got '%s'", headers["Authorization"])
	}
}
