package providers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"file4you/internal/config"
	"file4you/internal/genkithandler/errors"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// RetryConfig defines retry behavior for API calls
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// GoogleAIProvider represents the Google AI provider configuration using Genkit's built-in plugin
type GoogleAIProvider struct {
	config      config.GenkitPlugin
	initialized bool
	retryConfig RetryConfig
}

// NewGoogleAIProvider creates a new Google AI provider instance
func NewGoogleAIProvider(cfg config.GenkitPlugin) *GoogleAIProvider {
	// Default retry configuration
	retryConfig := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}

	return &GoogleAIProvider{
		config:      cfg,
		retryConfig: retryConfig,
	}
}

// Initialize sets up the Google AI provider with Genkit
// Note: The GoogleAI plugin should actually be registered during Genkit initialization
// This method is kept for compatibility with the existing provider interface
func (p *GoogleAIProvider) Initialize(ctx context.Context, g *genkit.Genkit) error {
	if p.initialized {
		return nil
	}

	if p.config.APIKey == "" {
		return errors.New("Google AI API key is required")
	}

	// Log the initialization - the actual plugin initialization happens in the main Genkit setup
	slog.Info("Google AI provider ready",
		"model", p.GetModel(),
		"has_api_key", p.config.APIKey != "")

	p.initialized = true
	return nil
}

// GetModel returns the configured model for Google AI
func (p *GoogleAIProvider) GetModel() string {
	if p.config.DefaultModel == "" {
		return "gemini-1.5-pro" // Default to stable Gemini model
	}
	return p.config.DefaultModel
}

// GenerateText generates text using the Google AI provider
func (p *GoogleAIProvider) GenerateText(ctx context.Context, g *genkit.Genkit, prompt string) (string, error) {
	if !p.initialized {
		return "", errors.New("provider not initialized")
	}

	// Use the built-in Genkit generate function with the Google AI model
	response, err := p.withRetry(ctx, func() (string, error) {
		// Get the model reference from the plugin
		model := googlegenai.GoogleAIModel(g, p.GetModel())
		if model == nil {
			return "", fmt.Errorf("model %s not found or not registered", p.GetModel())
		}

		result, err := genkit.Generate(ctx, g,
			ai.WithModel(model),
			ai.WithPrompt(prompt),
		)

		if err != nil {
			return "", err
		}

		return result.Text(), nil
	})

	return response, err
}

// GenerateWithStructuredOutput generates text with structured output using the Google AI provider
func (p *GoogleAIProvider) GenerateWithStructuredOutput(ctx context.Context, g *genkit.Genkit, prompt string, outputType interface{}) (*ai.ModelResponse, error) {
	if !p.initialized {
		return nil, errors.New("provider not initialized")
	}

	// Use the built-in Genkit generate function with structured output
	return p.withRetryStructured(ctx, func() (*ai.ModelResponse, error) {
		// Get the model reference from the plugin
		model := googlegenai.GoogleAIModel(g, p.GetModel())
		if model == nil {
			return nil, fmt.Errorf("model %s not found or not registered", p.GetModel())
		}

		result, err := genkit.Generate(ctx, g,
			ai.WithModel(model),
			ai.WithPrompt(prompt),
			ai.WithOutputType(outputType),
		)

		return result, err
	})
}

// withRetry implements retry logic for text generation
func (p *GoogleAIProvider) withRetry(ctx context.Context, fn func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= p.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := p.retryConfig.BaseDelay * time.Duration(1<<(attempt-1))
			if delay > p.retryConfig.MaxDelay {
				delay = p.retryConfig.MaxDelay
			}

			slog.Debug("Retrying Google AI request",
				"attempt", attempt,
				"delay", delay.String(),
				"last_error", lastErr.Error())

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			break
		}
	}

	return "", fmt.Errorf("google AI request failed after %d attempts: %w", p.retryConfig.MaxRetries+1, lastErr)
}

// withRetryStructured implements retry logic for structured generation
func (p *GoogleAIProvider) withRetryStructured(ctx context.Context, fn func() (*ai.ModelResponse, error)) (*ai.ModelResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= p.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := p.retryConfig.BaseDelay * time.Duration(1<<(attempt-1))
			if delay > p.retryConfig.MaxDelay {
				delay = p.retryConfig.MaxDelay
			}

			slog.Debug("Retrying Google AI structured request",
				"attempt", attempt,
				"delay", delay.String(),
				"last_error", lastErr.Error())

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			break
		}
	}

	return nil, fmt.Errorf("google AI structured request failed after %d attempts: %w", p.retryConfig.MaxRetries+1, lastErr)
}

// isRetryable determines if an error should trigger a retry
func (p *GoogleAIProvider) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retryable conditions for Google AI API
	retryablePatterns := []string{
		"rate limit",
		"too many requests",
		"quota exceeded",
		"service unavailable",
		"internal error",
		"timeout",
		"connection reset",
		"temporary failure",
		"server error",
		"resource exhausted",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsAvailable checks if the Google AI provider is available and configured
func (p *GoogleAIProvider) IsAvailable() bool {
	return p.config.APIKey != "" && p.initialized
}

// SupportsStructuredOutput indicates whether this provider supports structured output
func (p *GoogleAIProvider) SupportsStructuredOutput() bool {
	return true // Google AI/Gemini supports structured output
}

// GetMaxTokens returns the maximum token limit for the configured model
func (p *GoogleAIProvider) GetMaxTokens() int {
	// Return conservative limits for different Gemini models
	model := p.GetModel()
	switch model {
	case "gemini-1.5-pro", "gemini-1.5-pro-latest":
		return 2097152 // 2M tokens for Gemini 1.5 Pro
	case "gemini-1.5-flash", "gemini-1.5-flash-latest":
		return 1048576 // 1M tokens for Gemini 1.5 Flash
	case "gemini-2.0-flash":
		return 1048576 // 1M tokens for Gemini 2.0 Flash
	default:
		return 32768 // Conservative default
	}
}
