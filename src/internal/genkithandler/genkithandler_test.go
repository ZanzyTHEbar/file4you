package genkithandler

import (
	"testing"
)

// Skip these tests for now as they require a proper Genkit implementation
// We'll focus on the more isolated functionality that doesn't rely on external libraries

/*
func TestInitializeGenkit(t *testing.T) {
	t.Skip("Skipping as it requires proper Genkit implementation")
}

func TestGetGreetingFlow(t *testing.T) {
	t.Skip("Skipping as it requires proper Genkit implementation")
}

func TestGetBackupFlow(t *testing.T) {
	t.Skip("Skipping as it requires proper Genkit implementation")
}
*/

// This test doesn't depend on Genkit implementation
func TestSessionID(t *testing.T) {
	id1 := SessionID()
	id2 := SessionID()

	if id1 == "" {
		t.Fatal("SessionID returned empty string")
	}

	if id1 == id2 {
		t.Fatal("Expected different session IDs, got identical values")
	}
}
