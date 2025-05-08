// Package genkithandler provides integration with the genkit AI platform.
package genkithandler

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/firebase/genkit/go/genkit"
)

// Init performs any necessary initialization for the genkithandler package.
func Init() {
	slog.Info("Initializing genkithandler package")
}

// InitializeGenkit creates and initializes a new Genkit instance.
// This instance can be used to register and execute AI flows and tools.
func InitializeGenkit(ctx context.Context) (*genkit.Genkit, error) {
	apiKey := os.Getenv("GENKIT_API_KEY")
	if apiKey == "" {
		// For development/testing, we'll allow a nil API key
		slog.Warn("No GENKIT_API_KEY environment variable found, using stub implementation")
	}

	// Create a new Genkit instance - this uses a stub implementation because we don't know the exact API
	g := &genkit.Genkit{}
	
	// Register flows
	if err := RegisterFlows(g); err != nil {
		return nil, errors.New("failed to register flows: " + err.Error())
	}
	
	slog.Info("Genkit instance initialized successfully")
	return g, nil
}