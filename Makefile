.PHONY: build build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows test clean

# Default build
build:
	go build -ldflags="-s -w" -o bin/mcp-proxy-admin ./cmd/server

# Cross-compilation targets
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/mcp-proxy-admin-linux-amd64 ./cmd/server

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/mcp-proxy-admin-linux-arm64 ./cmd/server

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/mcp-proxy-admin-darwin-amd64 ./cmd/server

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/mcp-proxy-admin-darwin-arm64 ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/mcp-proxy-admin-windows-amd64.exe ./cmd/server

# Build all platforms
build-all: build build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows

# Testing
test:
	go test -v -race ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Development
dev:
	go run ./cmd/server

# Linting
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Build with version info
build-version: export VERSION ?= dev
build-version:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/mcp-proxy-admin ./cmd/server
