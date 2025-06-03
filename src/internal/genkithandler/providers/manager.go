// Package providers implements provider management and coordination
package providers

import (
	"context"
	"log/slog"

	"file4you/internal/config"
	"file4you/internal/genkithandler/errors"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// Manager handles multiple AI providers and provides fallback capabilities
type Manager struct {
	providers     map[ProviderType]AIProvider
	primaryType   ProviderType
	fallbackOrder []ProviderType
	initialized   bool
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers:     make(map[ProviderType]AIProvider),
		fallbackOrder: []ProviderType{},
	}
}

// RegisterProvider registers an AI provider with the manager
func (m *Manager) RegisterProvider(providerType ProviderType, provider AIProvider) error {
	if provider == nil {
		return errors.Errorf("provider cannot be nil for type %s", providerType)
	}

	m.providers[providerType] = provider
	slog.Info("Registered AI provider", "type", string(providerType))

	return nil
}

// Initialize initializes all registered providers
func (m *Manager) Initialize(ctx context.Context, g *genkit.Genkit, cfg config.GenkitConfig) error {
	if m.initialized {
		return nil
	}

	// TODO: Add a proper registration mechanism for providers
	// This is a placeholder for future provider registration logic

	// Initialize OpenAI provider if configured
	//if cfg.Plugins.OpenAI.APIKey != "" {
	//	openaiProvider := NewOpenAIProvider(cfg.Plugins.OpenAI)
	//	if err := m.RegisterProvider(ProviderTypeOpenAI, openaiProvider); err != nil {
	//		return errors.Wrapf(err, "failed to register OpenAI provider")
	//	}
	//
	//	if err := openaiProvider.Initialize(ctx, g); err != nil {
	//		slog.Warn("Failed to initialize OpenAI provider", "error", err)
	//	} else {
	//		m.fallbackOrder = append(m.fallbackOrder, ProviderTypeOpenAI)
	//		if m.primaryType == "" {
	//			m.primaryType = ProviderTypeOpenAI
	//		}
	//	}
	//}

	// Initialize Google AI provider if configured
	if cfg.Plugins.GoogleAI.APIKey != "" {
		googleaiProvider := NewGoogleAIProvider(cfg.Plugins.GoogleAI)
		if err := m.RegisterProvider(ProviderTypeGoogleAI, googleaiProvider); err != nil {
			return errors.Wrapf(err, "failed to register Google AI provider")
		}

		if err := googleaiProvider.Initialize(ctx, g); err != nil {
			slog.Warn("Failed to initialize Google AI provider", "error", err)
		} else {
			m.fallbackOrder = append(m.fallbackOrder, ProviderTypeGoogleAI)
			if m.primaryType == "" {
				m.primaryType = ProviderTypeGoogleAI
			}
		}
	}

	if len(m.fallbackOrder) == 0 {
		return errors.New("no AI providers are configured and available")
	}

	slog.Info("Provider manager initialized",
		"primary", string(m.primaryType),
		"fallback_order", m.fallbackOrder,
		"total_providers", len(m.providers))

	m.initialized = true
	return nil
}

// GenerateText generates text using the primary provider with fallback support
func (m *Manager) GenerateText(ctx context.Context, g *genkit.Genkit, prompt string) (string, error) {
	if !m.initialized {
		return "", errors.New("provider manager not initialized")
	}

	// Try primary provider first
	if m.primaryType != "" {
		if provider, exists := m.providers[m.primaryType]; exists && provider.IsAvailable() {
			result, err := provider.GenerateText(ctx, g, prompt)
			if err == nil {
				return result, nil
			}
			slog.Warn("Primary provider failed, trying fallback",
				"primary", string(m.primaryType),
				"error", err.Error())
		}
	}

	// Try fallback providers
	for _, providerType := range m.fallbackOrder {
		if providerType == m.primaryType {
			continue // Already tried
		}

		if provider, exists := m.providers[providerType]; exists && provider.IsAvailable() {
			result, err := provider.GenerateText(ctx, g, prompt)
			if err == nil {
				slog.Info("Fallback provider succeeded", "provider", string(providerType))
				return result, nil
			}
			slog.Warn("Fallback provider failed",
				"provider", string(providerType),
				"error", err.Error())
		}
	}

	return "", errors.New("all AI providers failed to generate text")
}

// GenerateWithStructuredOutput generates structured output with fallback support
func (m *Manager) GenerateWithStructuredOutput(ctx context.Context, g *genkit.Genkit, prompt string, outputType interface{}) (*ai.ModelResponse, error) {
	if !m.initialized {
		return nil, errors.New("provider manager not initialized")
	}

	// Try primary provider first
	if m.primaryType != "" {
		if provider, exists := m.providers[m.primaryType]; exists && provider.IsAvailable() {
			result, err := provider.GenerateWithStructuredOutput(ctx, g, prompt, outputType)
			if err == nil {
				return result, nil
			}
			slog.Warn("Primary provider failed for structured output, trying fallback",
				"primary", string(m.primaryType),
				"error", err.Error())
		}
	}

	// Try fallback providers
	for _, providerType := range m.fallbackOrder {
		if providerType == m.primaryType {
			continue // Already tried
		}

		if provider, exists := m.providers[providerType]; exists && provider.IsAvailable() {
			result, err := provider.GenerateWithStructuredOutput(ctx, g, prompt, outputType)
			if err == nil {
				slog.Info("Fallback provider succeeded for structured output", "provider", string(providerType))
				return result, nil
			}
			slog.Warn("Fallback provider failed for structured output",
				"provider", string(providerType),
				"error", err.Error())
		}
	}

	return nil, errors.New("all AI providers failed to generate structured output")
}

// SetPrimaryProvider sets the primary provider to use
func (m *Manager) SetPrimaryProvider(providerType ProviderType) error {
	if provider, exists := m.providers[providerType]; !exists || !provider.IsAvailable() {
		return errors.Errorf("provider %s is not available", providerType)
	}

	m.primaryType = providerType
	slog.Info("Primary provider changed", "new_primary", string(providerType))
	return nil
}

// GetAvailableProviders returns a list of available provider types
func (m *Manager) GetAvailableProviders() []ProviderType {
	var available []ProviderType
	for providerType, provider := range m.providers {
		if provider.IsAvailable() {
			available = append(available, providerType)
		}
	}
	return available
}

// GetProviderStatus returns the status of all providers
func (m *Manager) GetProviderStatus() map[ProviderType]bool {
	status := make(map[ProviderType]bool)
	for providerType, provider := range m.providers {
		status[providerType] = provider.IsAvailable()
	}
	return status
}

// GetPrimaryProvider returns the current primary provider type
func (m *Manager) GetPrimaryProvider() ProviderType {
	return m.primaryType
}

// GetProvider returns a specific provider by type
func (m *Manager) GetProvider(providerType ProviderType) (AIProvider, bool) {
	provider, exists := m.providers[providerType]
	return provider, exists
}
