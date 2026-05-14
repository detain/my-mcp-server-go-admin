// Package oauth provides OAuth 2.1 protected resource metadata implementation.
package oauth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// ProtectedResourceMetadata returns OAuth 2.1 protected resource metadata.
func ProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	serverURL := r.Host
	if r.TLS != nil {
		serverURL = "https://" + serverURL
	} else {
		serverURL = "http://" + serverURL
	}

	metadata := &oauthex.ProtectedResourceMetadata{
		Resource:             serverURL,
		AuthorizationServers: []string{serverURL},
		ScopesSupported:      []string{"read", "write", "admin"},
		BearerMethodsSupported: []string{"header"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcp"`)

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		slog.Error("failed to encode protected resource metadata", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Debug("served protected resource metadata")
}
