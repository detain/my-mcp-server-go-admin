package openapi

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir, time.Hour)

	if c.dir != tmpDir {
		t.Errorf("expected dir '%s', got '%s'", tmpDir, c.dir)
	}
	if c.ttl != time.Hour {
		t.Errorf("expected ttl 1h, got %v", c.ttl)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	c := NewCache(t.TempDir(), time.Hour)
	_, err := c.Get("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir, time.Hour)

	key := "test_key"
	data := []byte(`{"test": "data"}`)

	if err := c.Set(key, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := c.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("expected '%s', got '%s'", string(data), string(got))
	}
}

func TestCacheDelete(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir, time.Hour)

	key := "test_key"
	data := []byte(`{"test": "data"}`)

	if err := c.Set(key, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := c.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := c.Get(key)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestCacheExpired(t *testing.T) {
	tmpDir := t.TempDir()
	// Very short TTL
	c := NewCache(tmpDir, 10*time.Millisecond)

	key := "test_key"
	data := []byte(`{"test": "data"}`)

	if err := c.Set(key, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)

	_, err := c.Get(key)
	if err == nil {
		t.Error("expected error for expired cache")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir, time.Hour)

	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := "test_key_" + string(rune('a'+i))
		data := []byte(`{"test": "data"}`)
		if err := c.Set(key, data); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Clear all
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Check all entries are gone
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestCacheGetPath(t *testing.T) {
	c := NewCache("/tmp/cache", time.Hour)
	path := c.getPath("test_key")

	expected := filepath.Join("/tmp/cache", "cache_746573745f6b6579.json")
	if path != expected {
		t.Errorf("expected path '%s', got '%s'", expected, path)
	}
}
