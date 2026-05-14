// Package proxy provides HTTP client for upstream API calls.
package proxy

import (
	"net/http"
)

// AuthExtractor extracts authentication headers from an incoming HTTP request.
type AuthExtractor struct {
	bearerToken string
	apiKey      string
	sessionID   string
}

// NewAuthExtractor creates a new auth extractor with stdio-mode credentials.
func NewAuthExtractor(bearerToken, apiKey, sessionID string) *AuthExtractor {
	return &AuthExtractor{
		bearerToken: bearerToken,
		apiKey:      apiKey,
		sessionID:   sessionID,
	}
}

// Extract extracts auth headers from an incoming HTTP request.
// It checks in order: Authorization header, X-API-KEY header, sessionid header.
// Falls back to stdio credentials if no headers found.
func (e *AuthExtractor) Extract(req *http.Request) map[string]string {
	headers := make(map[string]string)

	// Check for Bearer token
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		headers["Authorization"] = authHeader
		return headers
	}

	// Check for X-API-KEY
	apiKey := req.Header.Get("X-API-KEY")
	if apiKey != "" {
		headers["X-API-KEY"] = apiKey
		return headers
	}

	// Check for sessionid
	sessionID := req.Header.Get("sessionid")
	if sessionID != "" {
		headers["sessionid"] = sessionID
		return headers
	}

	// Fall back to stdio credentials
	if e.bearerToken != "" {
		headers["Authorization"] = "Bearer " + e.bearerToken
		return headers
	}
	if e.apiKey != "" {
		headers["X-API-KEY"] = e.apiKey
		return headers
	}
	if e.sessionID != "" {
		headers["sessionid"] = e.sessionID
		return headers
	}

	return headers
}

// AddMCPHeaders adds required MCP-specific headers.
func AddMCPHeaders(headers map[string]string, requestID string) {
	// X-API-APP=1 short-circuits api_check_auth_limits() for MCP callers
	headers["X-API-APP"] = "1"

	// X-Request-Id for tracing
	if requestID == "" {
		requestID = generateRequestID()
	}
	headers["X-Request-Id"] = requestID
}

// generateRequestID generates a simple request ID.
func generateRequestID() string {
	// Simple implementation - in production could use UUID
	return "mcp-" + randomString(8)
}

// randomString generates a random string of given length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
