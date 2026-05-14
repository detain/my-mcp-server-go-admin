// Package server provides MCP server implementation with transport detection.
package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the MCP server instance.
type Server struct {
	impl    *mcp.Server
	name    string
	version string
}

// NewServer creates a new MCP server instance.
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
	}
}

// Initialize sets up the MCP server with the implementation details.
func (s *Server) Initialize(ctx context.Context) error {
	slog.Info("initializing MCP server",
		"name", s.name,
		"version", s.version,
	)

	s.impl = mcp.NewServer(&mcp.Implementation{
		Name:    s.name,
		Version: s.version,
	}, nil)

	return nil
}

// Impl returns the underlying MCP server implementation.
func (s *Server) Impl() *mcp.Server {
	return s.impl
}

// DetectTransport detects whether to use STDIO or HTTP transport
// based on whether stdin is a character device (TTY) or not.
func DetectTransport() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		slog.Warn("failed to stat stdin, defaulting to HTTP", "error", err)
		return "http"
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		slog.Info("TTY detected, using HTTP transport")
		return "http"
	}

	slog.Info("stdin is pipe/redirect, using STDIO transport")
	return "stdio"
}

// Run starts the MCP server using the appropriate transport.
func (s *Server) Run(ctx context.Context) error {
	if s.impl == nil {
		return ErrServerNotInitialized
	}

	transport := DetectTransport()

	switch transport {
	case "stdio":
		slog.Info("starting STDIO transport")
		return s.impl.Run(ctx, &mcp.StdioTransport{})
	default:
		slog.Info("starting HTTP transport")
		// HTTP transport is handled separately via Gin handler
		return nil
	}
}
