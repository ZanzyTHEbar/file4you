// Package genkithandler provides a simplified interface for integrating with Genkit.
package genkithandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/filesystem/trees"
	"file4you/internal/genkithandler/errors"
	"file4you/internal/genkithandler/providers"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// Service holds an initialized Genkit instance and provides methods
// for interacting with Genkit functionalities.
type Service struct {
	g               *genkit.Genkit
	cfg             config.File4YouGenkitHandlerConfig
	promptsDir      string
	loadedPrompts   map[string]Prompt
	providerManager *providers.Manager
}

// NewService initializes a new Genkit instance and returns a Service.
// It uses the globally loaded AppConfig.
func NewService(ctx context.Context, dfs *deskfs.DesktopFS, cdb *db.CentralDBProvider) (*Service, error) {

	appCfg := config.AppConfig

	// Load prompts
	prompts, err := LoadPrompts(appCfg.Genkit.Prompts.Directory)
	if err != nil {

		slog.Error("Failed to load prompts for Genkit handler", "directory", appCfg.Genkit.Prompts.Directory, "error", err)
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	// Initialize provider manager
	providerManager := providers.NewManager()

	// Prepare Genkit options based on configuration
	var genkitOpts []genkit.GenkitOption

	// Telemetry configuration has been deprioritized.
	// // Configure Telemetry
	// // This is illustrative; actual Genkit telemetry configuration might differ.
	// // Assuming OTLP (OpenTelemetry Protocol) exporter for traces and metrics.
	// if appCfg.Genkit.Telemetry.LoggingLevel != "" { // This line will cause a compile error as Telemetry field is removed
	// // Assuming Genkit has a way to set log level, perhaps via a telemetry option
	// // or a global setting. For this example, let's imagine an OTLP option.
	// slog.Info("Configuring Genkit Telemetry", "loggingLevel", appCfg.Genkit.Telemetry.LoggingLevel, "traceSampler", appCfg.Genkit.Telemetry.TraceSampler)
	// // TODO: Example: telemetryOpt, err := otlp.Init(ctx, otlp.Config{
	// // LoggingLevel: appCfg.Genkit.Telemetry.LoggingLevel,
	// // TraceSampler: appCfg.Genkit.Telemetry.TraceSampler,
	// // Endpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), // Or from config
	// // })
	// // if err != nil {
	// // slog.Error("Failed to initialize OTLP telemetry for Genkit", "error", err)
	// // } else if telemetryOpt != nil {
	// // genkitOpts = append(genkitOpts, telemetryOpt)
	// // }
	// // TODO: For now, we'll just log, as specific GenkitOption for this is unknown.
	// }

	// Configure Plugins
	// Google AI Plugin
	if appCfg.Genkit.Plugins.GoogleAI.APIKey != "" {
		googleAIPlugin := &googlegenai.GoogleAI{
			APIKey: appCfg.Genkit.Plugins.GoogleAI.APIKey,
		}
		genkitOpts = append(genkitOpts, genkit.WithPlugins(googleAIPlugin))
		
		slog.Info("Google AI Plugin configured", 
			"apiKeySet", appCfg.Genkit.Plugins.GoogleAI.APIKey != "", 
			"model", appCfg.Genkit.Plugins.GoogleAI.DefaultModel)
	}

	// OpenAI Plugin
	if appCfg.Genkit.Plugins.OpenAI.APIKey != "" {
		// openAICfg := openai.Config{ // Uncomment when openai package is available
		// APIKey:         appCfg.Genkit.Plugins.OpenAI.APIKey,
		// DefaultModel:   appCfg.Genkit.Plugins.OpenAI.DefaultModel,
		// RequestTimeout: time.Duration(appCfg.Genkit.Plugins.OpenAI.TimeoutSeconds) * time.Second,
		// }
		// Similar to Google AI, the initialization pattern would depend on the Genkit SDK.
		// TODO: Example:
		// if err := openai.Init(ctx, openAICfg); err != nil {
		// slog.Error("Failed to initialize OpenAI plugin for Genkit", "error", err)
		// }
		// Or:
		// plugin, err := openai.NewPlugin(ctx, openAICfg)
		// if err != nil {
		// slog.Error("Failed to initialize OpenAI plugin", "error", err)
		// } else {
		// genkitOpts = append(genkitOpts, genkit.WithPlugin(plugin)) // Hypothetical option
		// }
		slog.Info("OpenAI Plugin configured (illustrative)", "apiKeySet", appCfg.Genkit.Plugins.OpenAI.APIKey != "", "model", appCfg.Genkit.Plugins.OpenAI.DefaultModel, "timeoutSeconds", appCfg.Genkit.Plugins.OpenAI.TimeoutSeconds)
	}

	// Initialize Genkit with the constructed options
	g, err := genkit.Init(ctx, genkitOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Genkit: %w", err)
	}

	// Initialize AI provider manager
	if err := providerManager.Initialize(ctx, g, appCfg.Genkit); err != nil {
		slog.Warn("Failed to initialize AI provider manager", "error", err)
		// Continue without AI providers for now - allows the service to work without API keys
	}

	s := &Service{
		g:               g,
		cfg:             appCfg.File4You.GenkitHandler,
		promptsDir:      appCfg.Genkit.Prompts.Directory,
		loadedPrompts:   prompts,
		providerManager: providerManager,
	}

	// Register tools before flows so flows can call tools
	if err := RegisterCoreTools(s.g, dfs, cdb); err != nil {
		slog.Error("Failed to register core tools", "error", err)
	}
	if err := RegisterCoreFlows(s.g); err != nil {
		slog.Error("Failed to register core flows", "error", err)
	}

	return s, nil
}

// OrganizeFiles uses AI to analyze and suggest organization for the given files
func (s *Service) OrganizeFiles(ctx context.Context, files []trees.FileMetadata, directoryPath string, userContext string) (*FileOrganizationResult, error) {
	if s.providerManager == nil {
		return nil, errors.New("AI provider manager not initialized")
	}

	// Create prompt context
	promptContext := CreateFileOrganizationContext(files, directoryPath, userContext)

	// Get organization prompt
	prompt, found := GetPrompt("file_organization", s.loadedPrompts)
	if !found {
		return nil, errors.New("file_organization prompt not found")
	}

	// Render prompt with context
	renderedPrompt, err := RenderPrompt(prompt, promptContext)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render organization prompt")
	}

	// Generate organization suggestions using AI
	response, err := s.providerManager.GenerateText(ctx, s.g, renderedPrompt)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate organization suggestions")
	}

	// Parse the AI response (assuming JSON format)
	result, err := ParseOrganizationResult(response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse AI organization response")
	}

	slog.Info("AI file organization completed",
		"files_analyzed", len(files),
		"suggested_actions", len(result.SuggestedActions),
		"confidence", result.Confidence)

	return result, nil
}

// CategorizeFile uses AI to categorize a single file
func (s *Service) CategorizeFile(ctx context.Context, file trees.FileMetadata, userContext string) (*FileCategorization, error) {
	if s.providerManager == nil {
		return nil, errors.New("AI provider manager not initialized")
	}

	// Create single-file context
	fileInfo := FileInfo{
		Name:      filepath.Base(file.FilePath),
		Path:      file.FilePath,
		Size:      file.Size,
		Extension: strings.TrimPrefix(filepath.Ext(file.FilePath), "."),
		IsDir:     file.IsDir,
		ModTime:   file.ModTime.Format("2006-01-02 15:04:05"),
		Checksum:  file.Checksum,
		Metadata:  make(map[string]interface{}),
	}

	// Create categorization context
	context := PromptContext{
		Files:       []FileInfo{fileInfo},
		FileCount:   1,
		UserContext: userContext,
	}

	// Get categorization prompt
	prompt, found := GetPrompt("file_categorization", s.loadedPrompts)
	if !found {
		return nil, errors.New("file_categorization prompt not found")
	}

	// Render prompt with context
	renderedPrompt, err := RenderPrompt(prompt, context)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render categorization prompt")
	}

	// Generate categorization using AI
	response, err := s.providerManager.GenerateText(ctx, s.g, renderedPrompt)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate file categorization")
	}

	// Parse the AI response
	result, err := ParseCategorizationResult(response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse AI categorization response")
	}

	slog.Debug("AI file categorization completed",
		"file", fileInfo.Name,
		"category", result.Category,
		"confidence", result.Confidence)

	return result, nil
}

// DetectDuplicates uses AI to identify potential duplicate files
func (s *Service) DetectDuplicates(ctx context.Context, files []trees.FileMetadata, directoryPath string, userContext string) (*DuplicateDetectionResult, error) {
	if s.providerManager == nil {
		return nil, errors.New("AI provider manager not initialized")
	}

	// Create prompt context for duplicate detection
	promptContext := CreateFileOrganizationContext(files, directoryPath, userContext)

	// Get duplicate detection prompt
	prompt, found := GetPrompt("duplicate_detection", s.loadedPrompts)
	if !found {
		return nil, errors.New("duplicate_detection prompt not found")
	}

	// Render prompt with context
	renderedPrompt, err := RenderPrompt(prompt, promptContext)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render duplicate detection prompt")
	}

	// Generate duplicate analysis using AI
	response, err := s.providerManager.GenerateText(ctx, s.g, renderedPrompt)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate duplicate analysis")
	}

	// Parse the AI response
	result, err := ParseDuplicateDetectionResult(response)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse AI duplicate detection response")
	}

	slog.Info("AI duplicate detection completed",
		"files_analyzed", len(files),
		"duplicate_groups", len(result.DuplicateGroups),
		"confidence", result.Confidence)

	return result, nil
}

// GetProviderStatus returns the status of all AI providers
func (s *Service) GetProviderStatus() map[providers.ProviderType]bool {
	if s.providerManager == nil {
		return make(map[providers.ProviderType]bool)
	}
	return s.providerManager.GetProviderStatus()
}

// SetPrimaryProvider changes the primary AI provider
func (s *Service) SetPrimaryProvider(providerType providers.ProviderType) error {
	if s.providerManager == nil {
		return errors.New("AI provider manager not initialized")
	}
	return s.providerManager.SetPrimaryProvider(providerType)
}

// Genkit returns the underlying Genkit instance.
// This can be used for advanced scenarios where direct access to Genkit is needed.
func (s *Service) Genkit() *genkit.Genkit {
	return s.g
}

// GetPrompt retrieves a loaded prompt by its name.
// It returns the Prompt struct and a boolean indicating if the prompt was found.
func (s *Service) GetPrompt(name string) (Prompt, bool) {
	if s.loadedPrompts == nil {
		return Prompt{}, false
	}
	prompt, ok := s.loadedPrompts[name]
	return prompt, ok
}

// Close performs any cleanup required by the Service.
// For now, it's a placeholder.
func (s *Service) Close(ctx context.Context) error {
	// TODO: Add cleanup logic if necessary, e.g., for plugins or resources
	// that require explicit shutdown.
	return nil
}

// AI Response Parsing Functions

// ParseOrganizationResult parses the AI response for file organization
func ParseOrganizationResult(response string) (*FileOrganizationResult, error) {
	var result FileOrganizationResult
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse organization result: %w", err)
	}
	return &result, nil
}

// ParseCategorizationResult parses the AI response for file categorization
func ParseCategorizationResult(response string) (*FileCategorization, error) {
	var result FileCategorization
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse categorization result: %w", err)
	}
	return &result, nil
}

// ParseDuplicateDetectionResult parses the AI response for duplicate detection
func ParseDuplicateDetectionResult(response string) (*DuplicateDetectionResult, error) {
	var result DuplicateDetectionResult
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duplicate detection result: %w", err)
	}
	return &result, nil
}
