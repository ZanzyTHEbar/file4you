# Genkit Handler Package Documentation

The genkithandler package provides integration with the Genkit AI platform, allowing File4You to leverage AI capabilities for various tasks.

## Enterprise Architecture

The architecture follows a layered approach:

```
                +------------------+
                |   Public API     |
                +------------------+
                         |
                         v
                +------------------+
                |     Handler      |
                +------------------+
                         |
                         v
                +------------------+
                |     Client       |
                +------------------+
                         |
                         v
                +------------------+
                |   Genkit SDK     |
                +------------------+
```

### Key Components (and Current Status)

1.  **Client Layer**: Abstractions for communicating with the Genkit API.
    *   *Status: Not yet implemented.* (Currently uses Genkit SDK directly).
    *   Planned: Interface definitions. (Future: HTTP client implementation, Mock client for testing).

2.  **Error Handling**: Comprehensive error framework.
    *   *Status: Partially implemented / In progress.* (Basic error propagation via `fmt.Errorf` is in place. The `errors/` subdirectory is intended for custom error types and utilities).
    *   Planned: Custom error types with context, Error categorization and retriability, Error wrapping utilities.

3.  **Configuration**: Flexible configuration system.
    *   *Status: Not yet implemented.*
    *   Planned: File-based and environment-based configuration, Validation and defaults, Typed configuration objects.

4.  **Telemetry**: Observability tooling.
    *   *Status: Partially implemented.* (`slog` is used in `legacy.go`. Comprehensive integration is pending).
    *   Planned: Structured logging, Metrics collection, Distributed tracing.

5.  **Security**: Enterprise security controls.
    *   *Status: Not yet implemented.*
    *   Planned: Authentication mechanisms, Authorization framework, Audit logging.

6.  **Models**: Domain models.
    *   *Status: Partially implemented.* (`types.go` defines some I/O structs. The `models/` subdirectory is intended for more comprehensive domain models).
    *   Planned: Flow and tool definitions, Request/response structures, Context management.

7.  **Handler**: Main orchestration layer.
    *   *Status: Implemented.* (`genkithandler.go`, `flows.go`, `tools.go` provide core functionality).
    *   Features: Flow execution, Tool registration, Resource management (basic, via `Service` struct).

## Usage

### Basic Usage (New API)

This example demonstrates using the `Service`-based approach.

```go
import (
	"context"
	"fmt"
	"log"

	"file4you/internal/genkithandler" // Assuming this is the correct import path
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/core"
	// Import any necessary Genkit plugins, e.g.,
	// "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	ctx := context.Background()

	// Initialize the Genkit Service
	// Replace with actual GenkitOptions, e.g., plugins
	svc, err := genkithandler.NewService(ctx /* genkit.WithPlugins(&plugins.SomePlugin{}) */)
	if err != nil {
		log.Fatalf("Failed to initialize genkithandler.Service: %v", err)
	}
	defer svc.Close(ctx) // Assuming Close might do cleanup

	// Get the underlying Genkit instance from the Service
	g := svc.Genkit()

	// Define a flow using the handler's wrapper
	greetingFlowHandler := func(ctx context.Context, name string) (string, error) {
		return "Hello, " + name, nil
	}
	_, err = genkithandler.DefineFlow(g, "greetingFlow", greetingFlowHandler)
	if err != nil {
		log.Fatalf("Failed to define flow: %v", err)
	}

	// Execute the flow using the handler's wrapper
	result, err := genkithandler.ExecuteFlow[string, string](ctx, g, "greetingFlow", "World from New API")
	if err != nil {
		log.Fatalf("Failed to execute flow: %v", err)
	}
	fmt.Println(result) // Output: Hello, World from New API
}
```

### Legacy Usage (Backward Compatible)

```go
ctx := context.Background()
// Initialize legacy system (which internally calls NewService)
genkithandler.Init() // Or use InitializeGenkit if preferred by legacy code path

// Get the backup flow (assuming it's registered by legacy Init)
backupFlowRunner := genkithandler.GetBackupFlow()
if backupFlowRunner == nil {
    log.Fatal("Backup flow not registered or default Genkit instance not initialized.")
    return
}

// The legacy GetBackupFlow returns an interface{} that needs type assertion
// or a custom runner that wraps the execution.
// Based on legacy.go, it returns a *legacyFlowRunner.
type LegacyFlowRunner interface {
    Run(ctx context.Context, input interface{}) (interface{}, error)
}

legacyRunner, ok := backupFlowRunner.(LegacyFlowRunner)
if !ok {
    log.Fatal("Retrieved backup flow runner is not of expected type.")
    return
}

// Execute the flow
// Assuming BackupToolInput and BackupToolOutput are defined in genkithandler package
input := genkithandler.BackupToolInput{} // Provide appropriate input
rawResult, err := legacyRunner.Run(context.Background(), input)
if err != nil {
    log.Fatalf("Failed to execute legacy backup flow: %v", err)
}

// Type assert the result
backupOutput, ok := rawResult.(genkithandler.BackupToolOutput)
if !ok {
    log.Fatalf("Legacy backup flow result is not of type BackupToolOutput. Actual: %T", rawResult)
}

// Use the result
fmt.Println(backupOutput.SuccessMessage)
```

## Advanced Features

### Custom Configuration

```go
// Load a custom configuration
// cfg, err := config.LoadConfig("/path/to/config.json") // Placeholder for actual config loading
// if err != nil {
//     log.Fatalf("Failed to load config: %v", err)
// }

// Override specific settings
// cfg.Client.Timeout = config.Duration(30 * time.Second)
// cfg.Features.EnableCaching = true
```
*Note: Configuration system is not yet implemented in `internal/genkithandler`.*

### Telemetry Integration

```go
// Create a structured logger
// logger := telemetry.NewStructuredLogger("myComponent", telemetry.InfoLevel) // Placeholder

// Create a metrics collector
// metrics := telemetry.NewSimpleMetricsCollector("myComponent") // Placeholder

// Use telemetry in a handler
// h, err := handler.New( // Placeholder for a more abstract handler if developed
//     handler.WithLogger(logger),
//     handler.WithMetrics(metrics),
// )
```
*Note: Comprehensive telemetry integration is not yet implemented in `internal/genkithandler`.*

### Authentication and Authorization

```go
// Create an authenticator
// authenticator := security.NewAPIKeyAuthenticator(map[string]string{ // Placeholder
//     "my-api-key": "tenant-1",
// })

// Create an authorizer
// authorizer := security.NewRoleBasedAuthorizer(map[string][]security.Permission{ // Placeholder
//     "admin": {
//         {Resource: "*", Action: "*"},
//     },
// })

// Use security in a handler
// h, err := handler.New( // Placeholder
//     handler.WithAuthenticator(authenticator),
//     handler.WithAuthorizer(authorizer),
// )
```
*Note: Security features are not yet implemented in `internal/genkithandler`.*

## Testing

The package includes comprehensive testing utilities:

```go
// Create a mock client for testing
// mockClient := client.NewMockClient() // Placeholder for when a client layer exists

// Configure the mock
// mockClient.SetFlowResponse("myFlow", "Mock response")

// Use the mock in tests
// h, err := handler.New( // Placeholder
//     handler.WithClient(mockClient),
// )

// Execute a flow using the mock
// result, err := h.ExecuteFlow[string, string](ctx, "myFlow", "input")
// result will be "Mock response"

// Verify the mock was called correctly
// callHistory := mockClient.GetCallHistory()
```
*Note: Mocking utilities for a dedicated client layer are not yet implemented in `internal/genkithandler`.*

## Extending

To add new flows or tools:

1. Define new models in the `models` package if needed (or `types.go` for simple I/O).
2. Create flow/tool handlers using the `genkithandler.DefineFlow` or `genkithandler.DefineTool` functions with an initialized `*genkit.Genkit` instance (typically obtained from `genkithandler.NewService().Genkit()`).
3. Ensure appropriate error handling.
4. Consider telemetry, configuration, and security as these enterprise features are developed.
