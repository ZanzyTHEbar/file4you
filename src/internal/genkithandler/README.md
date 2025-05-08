# Genkit Handler Package Documentation

The genkithandler package provides integration with the Genkit AI platform, allowing File4You to leverage AI capabilities for various tasks.

## Overview

This package contains the following key components:

1. **Initialization**: Functions to initialize and set up a Genkit instance.
2. **Flows**: Pre-defined AI flows for operations like backups and greetings.
3. **Tools**: Tools that can be registered with Genkit to perform specific tasks.
4. **Types**: Type definitions used by the flows and tools.

## Usage

### Initialization

To initialize the Genkit system, use:

```go
ctx := context.Background()
genkitInstance, err := genkithandler.InitializeGenkit(ctx)
if err != nil {
    // Handle error
}
```

### Flows

Flows are pre-registered during initialization. To retrieve and execute a flow:

```go
// Get the backup flow
backupFlow := genkithandler.GetBackupFlow()
if backupFlow == nil {
    // Flow not registered
    return
}

// Execute the flow
result, err := backupFlow.Run(context.Background(), genkithandler.BackupToolInput{})
if err != nil {
    // Handle error
}

// Use the result
fmt.Println(result.SuccessMessage)
```

### Tools

To register tools with a Genkit instance:

```go
// Register the backup tool
genkithandler.RegisterBackupTool(genkitInstance, desktopFS, centralDB)
```

Set the `GENKIT_API_KEY` environment variable to authenticate with the Genkit service. If not set, a stub implementation will be used for development and testing.

## Extending

To add new flows or tools:

1. Define new types in `types.go` if needed
2. Add new flow definitions in `flows.go`
3. Add new tool registrations in `tools.go`
4. Update the initialization logic in `init.go` if necessary
