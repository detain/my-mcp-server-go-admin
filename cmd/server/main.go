package main

import (
	"log/slog"
	"os"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting mcp-proxy-admin server",
		"name", "myadmin-admin-mcp",
		"version", "1.0.0",
	)
}
