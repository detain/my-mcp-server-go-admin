// Package openapi provides OpenAPI spec fetching, parsing, and tool generation.
package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser handles fetching and parsing OpenAPI specifications.
type Parser struct {
	cacheDir   string
	httpClient *http.Client
	cacheTTL   time.Duration
}

// NewParser creates a new OpenAPI parser.
func NewParser(cacheDir string) *Parser {
	if cacheDir == "" {
		cacheDir = "/tmp/mcp_admin_cache"
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Warn("failed to create cache directory", "error", err)
	}

	return &Parser{
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheTTL: 1 * time.Hour,
	}
}

// FetchSpec fetches an OpenAPI spec from a URL and returns the parsed specification.
func (p *Parser) FetchSpec(specURL string) (*openapi3.T, error) {
	slog.Info("fetching OpenAPI spec", "url", specURL)

	// Try cache first
	if cachedSpec, err := p.loadFromCache(specURL); err == nil && cachedSpec != nil {
		slog.Info("using cached OpenAPI spec")
		return cachedSpec, nil
	}

	// Fetch from URL
	resp, err := p.httpClient.Get(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse YAML/JSON
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Validate the spec (skip validation as some specs may have minor issues)
	_ = spec

	// Cache the parsed spec
	if err := p.saveToCache(specURL, spec); err != nil {
		slog.Warn("failed to cache OpenAPI spec", "error", err)
	}

	return spec, nil
}

// loadFromCache attempts to load a cached OpenAPI spec.
func (p *Parser) loadFromCache(specURL string) (*openapi3.T, error) {
	cacheFile := p.getCacheFile(specURL)

	info, err := os.Stat(cacheFile)
	if err != nil {
		return nil, err
	}

	// Check if cache is stale
	if time.Since(info.ModTime()) > p.cacheTTL {
		slog.Info("cache expired", "file", cacheFile)
		return nil, fmt.Errorf("cache expired")
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var spec openapi3.T
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// saveToCache saves the OpenAPI spec to cache.
func (p *Parser) saveToCache(specURL string, spec *openapi3.T) error {
	data, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	cacheFile := p.getCacheFile(specURL)
	return os.WriteFile(cacheFile, data, 0644)
}

// getCacheFile returns the path to the cache file for a given spec URL.
func (p *Parser) getCacheFile(specURL string) string {
	cacheKey := fmt.Sprintf("%x", url.QueryEscape(specURL))
	return filepath.Join(p.cacheDir, "openapi_"+cacheKey+".json")
}

// ClearCache removes the cached spec for a given URL.
func (p *Parser) ClearCache(specURL string) error {
	cacheFile := p.getCacheFile(specURL)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}
