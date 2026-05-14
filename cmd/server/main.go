package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/myadmin/go-mcp-proxy-admin/internal/oauth"
	"github.com/myadmin/go-mcp-proxy-admin/internal/server"
)

var (
	version = "1.0.0"
	name    = "myadmin-admin-mcp"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load environment variables from .env if present
	_ = godotenv.Load()

	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		slog.Info("mcp-proxy-admin version", "version", version)
		os.Exit(0)
	}

	slog.Info("starting mcp-proxy-admin server",
		"name", name,
		"version", version,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create MCP server
	mcpServer := server.NewServer(name, version)
	if err := mcpServer.Initialize(ctx); err != nil {
		slog.Error("failed to initialize MCP server", "error", err)
		os.Exit(1)
	}

	// Detect transport mode
	transport := server.DetectTransport()

	if transport == "stdio" {
		slog.Info("running in STDIO mode")
		if err := mcpServer.Run(ctx); err != nil {
			slog.Error("STDIO server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// HTTP mode - set up Gin router
	slog.Info("running in HTTP mode")
	setupHTTPServer(ctx, cancel, mcpServer)
}

// setupHTTPServer configures and starts the HTTP server with Gin.
func setupHTTPServer(ctx context.Context, cancel context.CancelFunc, mcpServer *server.Server) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Create MCP HTTP handler - use gin.WrapH for http.Handler
	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return mcpServer.Impl()
	}, nil)

	// MCP endpoint - use gin.WrapH for http.Handler
	router.Any("/mcp", gin.WrapH(mcpHandler))
	router.POST("/mcp", gin.WrapH(mcpHandler))
	router.GET("/mcp", gin.WrapH(mcpHandler))

	// OAuth protected resource metadata endpoint
	router.GET("/.well-known/oauth-protected-resource", gin.WrapF(oauth.ProtectedResourceMetadata))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"name":    name,
			"version": version,
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("shutting down HTTP server...")
		cancel()

		if err := srv.Shutdown(context.Background()); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("HTTP server listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}

	slog.Info("HTTP server stopped")
}
