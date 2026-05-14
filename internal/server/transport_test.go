package server

import (
	"testing"
)

func TestDetectTransport(t *testing.T) {
	// Test that DetectTransport returns a valid transport type
	transport := DetectTransport()
	if transport != "stdio" && transport != "http" {
		t.Errorf("DetectTransport() returned unexpected value: %s", transport)
	}
}

func TestDetectTransportConsistency(t *testing.T) {
	// Multiple calls should return the same result
	transport1 := DetectTransport()
	transport2 := DetectTransport()
	if transport1 != transport2 {
		t.Errorf("DetectTransport() returned inconsistent results: %s, %s", transport1, transport2)
	}
}
