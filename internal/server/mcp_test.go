package server

import (
	"context"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer("test-server", "1.0.0")
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.name != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", s.name)
	}
	if s.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", s.version)
	}
}

func TestServerInitialize(t *testing.T) {
	s := NewServer("test-server", "1.0.0")
	ctx := context.Background()

	if err := s.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if s.Impl() == nil {
		t.Error("Impl() returned nil after Initialize")
	}
}

func TestServerInitializeTwice(t *testing.T) {
	s := NewServer("test-server", "1.0.0")
	ctx := context.Background()

	if err := s.Initialize(ctx); err != nil {
		t.Fatalf("first Initialize failed: %v", err)
	}

	// Second Initialize should reset the server
	if err := s.Initialize(ctx); err != nil {
		t.Fatalf("second Initialize failed: %v", err)
	}
}

func TestServerImplBeforeInitialize(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	// Impl() before Initialize should return nil
	if s.Impl() != nil {
		t.Error("Impl() should return nil before Initialize")
	}
}
