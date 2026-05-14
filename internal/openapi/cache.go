package openapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Cache handles caching of parsed OpenAPI specs.
type Cache struct {
	dir  string
	ttl  time.Duration
}

// NewCache creates a new cache instance.
func NewCache(dir string, ttl time.Duration) *Cache {
	if dir == "" {
		dir = "/tmp/mcp_admin_cache"
	}
	if ttl == 0 {
		ttl = 1 * time.Hour
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Warn("failed to create cache directory", "error", err)
	}

	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

// Get loads a cached spec if it exists and is not expired.
func (c *Cache) Get(key string) ([]byte, error) {
	filePath := c.getPath(key)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if time.Since(info.ModTime()) > c.ttl {
		slog.Debug("cache entry expired", "key", key)
		return nil, fmt.Errorf("cache expired")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Set saves data to the cache.
func (c *Cache) Set(key string, data []byte) error {
	filePath := c.getPath(key)
	return os.WriteFile(filePath, data, 0644)
}

// Delete removes an entry from the cache.
func (c *Cache) Delete(key string) error {
	filePath := c.getPath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
			slog.Warn("failed to remove cache entry", "file", entry.Name(), "error", err)
		}
	}

	return nil
}

// GetJSON loads and unmarshals cached JSON data.
func (c *Cache) GetJSON(key string, v any) error {
	data, err := c.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// SetJSON marshals and caches JSON data.
func (c *Cache) SetJSON(key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Set(key, data)
}

// getPath returns the file path for a cache key.
func (c *Cache) getPath(key string) string {
	// Simple hash to make safe filename
	hash := fmt.Sprintf("%x", []byte(key))
	return filepath.Join(c.dir, "cache_"+hash+".json")
}
