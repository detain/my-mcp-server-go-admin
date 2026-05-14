package openapi

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser("")
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.cacheDir == "" {
		t.Error("cacheDir should not be empty")
	}
}

func TestNewParserWithDir(t *testing.T) {
	p := NewParser("/tmp/test_cache")
	if p.cacheDir != "/tmp/test_cache" {
		t.Errorf("expected cacheDir '/tmp/test_cache', got '%s'", p.cacheDir)
	}
}

func TestGetCacheFile(t *testing.T) {
	p := NewParser("/tmp/cache")
	file := p.getCacheFile("https://example.com/spec.yaml")
	if file == "" {
		t.Error("getCacheFile returned empty string")
	}
	if !strings.Contains(file, "/tmp/cache/") {
		t.Errorf("cache file should be in /tmp/cache/, got '%s'", file)
	}
}

func TestCacheKey(t *testing.T) {
	p1 := NewParser("/tmp/cache")
	p2 := NewParser("/tmp/cache")

	// Same URL should produce same cache key
	key1 := p1.getCacheFile("https://example.com/spec")
	key2 := p2.getCacheFile("https://example.com/spec")
	if key1 != key2 {
		t.Errorf("same URL should produce same cache key: %s vs %s", key1, key2)
	}

	// Different URLs should produce different cache keys
	key3 := p1.getCacheFile("https://example.com/other")
	if key1 == key3 {
		t.Errorf("different URLs should produce different cache keys: %s vs %s", key1, key3)
	}
}
