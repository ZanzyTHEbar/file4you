// Package providers defines interfaces and common functionality for AI providers
package providers

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// AIProvider defines the interface that all AI providers must implement
type AIProvider interface {
	// Initialize sets up the provider with Genkit
	Initialize(ctx context.Context, g *genkit.Genkit) error

	// GetModel returns the configured model name
	GetModel() string

	// GenerateText generates text using the provider's model
	GenerateText(ctx context.Context, g *genkit.Genkit, prompt string) (string, error)

	// GenerateWithStructuredOutput generates structured output
	GenerateWithStructuredOutput(ctx context.Context, g *genkit.Genkit, prompt string, outputType interface{}) (*ai.ModelResponse, error)

	// IsAvailable checks if the provider is properly configured and available
	IsAvailable() bool
}

// ExtendedAIProvider provides additional capabilities beyond the basic interface
type ExtendedAIProvider interface {
	AIProvider

	// SupportsStructuredOutput returns whether this provider supports structured output
	SupportsStructuredOutput() bool

	// GetMaxTokens returns the maximum token limit for the current model
	GetMaxTokens() int
}

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeGoogleAI  ProviderType = "googleai"
	ProviderTypeOllamaAI  ProviderType = "ollama"
	ProviderTypeAnthropic  ProviderType = "anthropic"
	ProviderTypeAzureAI    ProviderType = "azureai"
)

// ProviderConfig contains common configuration for all providers
type ProviderConfig struct {
	Type          ProviderType
	Enabled       bool
	Primary       bool
	FallbackOrder int
}
