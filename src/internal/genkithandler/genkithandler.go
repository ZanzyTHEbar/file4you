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

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
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
		// TODO: googleAICfg := googleai.Config{ // Uncomment when googleai package is available
		// APIKey:       appCfg.Genkit.Plugins.GoogleAI.APIKey,
		// DefaultModel: appCfg.Genkit.Plugins.GoogleAI.DefaultModel,
		// }
		// Timeout would typically be part of the client used by the plugin,
		// or a specific option if the plugin supports it directly.
		// googleAIOpt, err := googleai.InitWithOptions(ctx, googleAICfg) // Hypothetical InitWithOptions
		// For Genkit, plugins are often just initialized and they register themselves or provide an option.
		// Let's assume Init registers the plugin and returns an error if it fails.
		// Or, it might return a genkit.Plugin which can be passed as an option.
		// The exact mechanism depends on the Genkit SDK design.
		// Example:
		// if err := googleai.Init(ctx, googleAICfg); err != nil {
		// slog.Error("Failed to initialize Google AI plugin for Genkit", "error", err)
		// }
		// TODO: Another pattern:
		// plugin, err := googleai.NewPlugin(ctx, googleAICfg)
		// if err != nil {
		// slog.Error("Failed to initialize Google AI plugin", "error", err)
		// } else {
		// genkitOpts = append(genkitOpts, genkit.WithPlugin(plugin)) // Hypothetical option
		// }
		slog.Info("Google AI Plugin configured (illustrative)", "apiKeySet", appCfg.Genkit.Plugins.GoogleAI.APIKey != "", "model", appCfg.Genkit.Plugins.GoogleAI.DefaultModel)
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

func (s *Service) registerExampleFlows(ctx context.Context) error {
	examplePromptFlowFn := func(ctx context.Context, input string) (string, error) {
		prefixPrompt, found := s.GetPrompt("example_prefix")
		fullPromptText := input
		if found {
			fullPromptText = prefixPrompt.Content + input
			slog.Debug("Example flow: Used prompt 'example_prefix'", "input", input, "fullPrompt", fullPromptText)
		} else {
			slog.Warn("Example flow: Prompt 'example_prefix' not found, using raw input for prompt.", "input", input)
		}

		// Illustrative call to genkit.Generate.
		// Assumes a model named "defaultModel" will be configured by one of the plugins.
		// The actual model name (e.g., "googleAI/gemini-1.5-pro-latest") should be used once plugins are configured.
		// Or, a default model can be set for the genkit instance.
		slog.Info("Example flow: Attempting to call genkit.Generate", "model", "defaultModel", "prompt", fullPromptText)

		// In a real scenario, ensure "defaultModel" or a specific model alias is correctly configured.
		// For now, this call will likely fail if no model named "defaultModel" is registered.
		// This is for illustrative purposes to show the structure.

		// 1. Look up the model by name.
		// The actual model name (e.g., "gemini-1.5-pro-latest") and provider (e.g., "googleai")
		// should be used once plugins are configured.
		// The error indicates LookupModel wants (g *Genkit, providerName string, modelName string).
		model := genkit.LookupModel(s.g, "defaultProvider", "defaultModel") // Using placeholder "defaultProvider"
		if model == nil {
			slog.Error("Example flow: Model 'defaultModel' from 'defaultProvider' not found. Ensure it's configured and plugins are initialized.")
			return "", errors.New("model 'defaultModel' (provider 'defaultProvider') not found in examplePromptFlow")
		}

		// 2. Construct the ModelRequest.
		// The error indicates model.Generate wants (ctx, *ai.ModelRequest, ai.ModelStreamCallback).
		// We'll create a simple request with the prompt.
		// A common way is to use messages. ai.NewUserMessage creates a *ai.Message.
		// ai.WithPrompt is a PromptingOption, not directly a ModelRequest.
		// We need to build an *ai.ModelRequest.
		// One way is to set the Messages field.
		// ai.NewUserMessage expects *ai.Part arguments.
		request := &ai.ModelRequest{
			Messages: []*ai.Message{ai.NewUserMessage(ai.NewTextPart(fullPromptText))},
			// Other request options like Temperature, MaxOutputTokens, etc., could be set here
			// or via options if ModelRequest supports them.
			// For now, a simple message-based request.
		}
		// Alternatively, if ai.WithPrompt can be applied to a request:
		// request := &ai.ModelRequest{}
		// err := ai.WithPrompt(fullPromptText).ApplyModelRequest(request) // This is speculative
		// if err != nil {
		//     slog.Error("Example flow: Failed to apply prompt to ModelRequest", "error", err)
		//     return "", errors.Wrapf(err, "failed to create ModelRequest")
		// }

		// 3. Call Generate on the model instance with the request.
		// Pass nil for ModelStreamCallback for non-streaming.
		resp, err := model.Generate(ctx, request, nil)
		if err != nil {
			slog.Error("Example flow: model.Generate failed (this is expected if 'defaultModel' is not properly configured/initialized)", "error", err)
			return "", errors.Wrapf(err, "model.Generate failed in examplePromptFlow")
		}

		if resp == nil {
			slog.Error("Example flow: model.Generate returned a nil response")
			return "", errors.New("model.Generate returned nil response in examplePromptFlow")
		}

		// 4. Extract text from the response.
		// Assuming *ai.ModelResponse has a Text() method that returns a single string.
		// This was based on the previous error "assignment mismatch: 2 variables but resp.Text returns 1 value".
		responseText := resp.Text() // This assumes resp.Text() exists and returns string

		if responseText == "" {
			slog.Warn("Example flow: model.Generate returned an empty text response")
		}

		slog.Debug("Example flow: model.Generate successful", "response", responseText)
		return responseText, nil
	}

	_, err := DefineFlow(s.g, "examplePromptFlow", core.Func[string, string](examplePromptFlowFn))
	if err != nil {
		return errors.Wrapf(err, "failed to define 'examplePromptFlow'")
	}
	slog.Info("Successfully defined 'examplePromptFlow'")

	// Add more flow definitions here...

	return nil
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
