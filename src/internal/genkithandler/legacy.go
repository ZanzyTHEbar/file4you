// Package genkithandler provides integration with the genkit AI platform.
// This file contains legacy functions for backward compatibility.
package genkithandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"file4you/internal/db"
	"file4you/internal/deskfs"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// defaultService is a package-level variable to hold a singleton Genkit Service instance
// for legacy functions that don't explicitly manage a Service.
var defaultService *Service
var defaultGenkitInstance *genkit.Genkit

// Init is a legacy initialization function.
// It initializes a default Genkit service for the package.
func Init() {
	slog.Info("Legacy Init() called. Initializing default genkithandler service.")
	ctx := context.Background()
	// Initialize with default options or allow configuration via environment variables.
	// For now, using no specific options.
	svc, err := NewService(ctx)
	if err != nil {
		slog.Error("Failed to initialize default genkithandler service in legacy Init", "error", err)
		return
	}
	defaultService = svc
	defaultGenkitInstance = svc.g // Store the underlying genkit.Genkit instance too
	// Legacy RegisterFlows and RegisterTools should be called here if they were part of the old Init.
	// For now, assuming they are called separately or implicitly.
	if err := RegisterFlows(ctx); err != nil { // This RegisterFlows is the legacy one
		slog.Error("Failed to register flows during legacy Init", "error", err)
	}
}

// InitializeGenkit is a legacy function.
// It initializes and returns a Genkit instance (now our Service wrapped as interface{}).
func InitializeGenkit(ctx context.Context) (interface{}, error) {
	slog.Warn("Legacy InitializeGenkit called. Consider migrating to NewService.")
	if defaultService == nil || defaultService.g != defaultGenkitInstance {
		// If Init() wasn't called or the instance changed, re-initialize.
		svc, err := NewService(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize genkithandler service in InitializeGenkit: %w", err)
		}
		defaultService = svc
		defaultGenkitInstance = svc.g
	}

	// Ensure legacy flows are registered with this instance
	if err := RegisterFlows(ctx); err != nil { // This RegisterFlows is the legacy one
		return nil, fmt.Errorf("failed to register flows in InitializeGenkit: %w", err)
	}

	slog.Info("Legacy Genkit instance initialized successfully via InitializeGenkit")
	return defaultService.g, nil // Return the raw *genkit.Genkit instance for closer compatibility
}

// --- Legacy Flow Registration and Retrieval ---

// RegisterFlows (legacy) defines and registers all predefined legacy flows.
// This function now uses the new DefineFlow helpers with the defaultGenkitInstance.
func RegisterFlows(ctx context.Context) error {
	slog.Warn("Legacy RegisterFlows called.")
	if defaultGenkitInstance == nil {
		return errors.New("default genkit instance not initialized in RegisterFlows")
	}

	// Example: Re-implementing greetingFlow
	greetingHandler := func(ctx context.Context, name string) (string, error) {
		return "Hello, " + name, nil
	}
	_, err := DefineFlow(defaultGenkitInstance, "greetingFlow", greetingHandler)
	if err != nil {
		return fmt.Errorf("failed to register legacy greetingFlow: %w", err)
	}

	backupHandler := func(ctx context.Context, input BackupToolInput) (BackupToolOutput, error) {
		// TODO: Placeholder for actual backup logic.
		// This would typically involve calling the backup tool or embedding its logic.
		slog.Info("Legacy backupFlow invoked with input", "input", input)
		// TODO: For now, assume it calls the 'performBackup' tool.
		// This creates a dependency: the 'performBackup' tool must be registered.
		output, toolErr := ExecuteTool[BackupToolInput, BackupToolOutput](ctx, defaultGenkitInstance, "performBackup", input)
		if toolErr != nil {
			return BackupToolOutput{}, fmt.Errorf("error executing performBackup tool in backupFlow: %w", toolErr)
		}
		return output, nil
	}
	_, err = DefineFlow(defaultGenkitInstance, "backupFlow", backupHandler)
	if err != nil {
		return fmt.Errorf("failed to register legacy backupFlow: %w", err)
	}

	slog.Info("Legacy flows (greetingFlow, backupFlow) registered.")
	return nil
}

// legacyFlowRunner is a placeholder for the core.Flow runner from the old implementation.
// It's used by GetGreetingFlow and GetBackupFlow to maintain API compatibility.
type legacyFlowRunner struct {
	name string
	g    *genkit.Genkit // Store the genkit instance to use for execution
}

// Run is a placeholder implementation for the core.Flow.Run method.
func (r *legacyFlowRunner) Run(ctx context.Context, input interface{}) (interface{}, error) {
	slog.Warn("Legacy flow runner's Run method called directly, using new ExecuteFlow", "flow", r.name)
	if r.g == nil {
		return nil, errors.New("genkit instance not available in legacyFlowRunner")
	}

	switch r.name {
	case "greetingFlow":
		name, ok := input.(string)
		if !ok {
			return nil, errors.New("invalid input type for greetingFlow (expected string)")
		}
		return ExecuteFlow[string, string](ctx, r.g, "greetingFlow", name)

	case "backupFlow":
		backupInput, ok := input.(BackupToolInput)
		if !ok {
			return nil, errors.New("invalid input type for backupFlow, expected BackupToolInput")
		}
		return ExecuteFlow[BackupToolInput, BackupToolOutput](ctx, r.g, "backupFlow", backupInput)

	default:
		return nil, errors.New("unknown legacy flow: " + r.name)
	}
}

func GetGreetingFlow() interface{} {
	slog.Warn("Legacy GetGreetingFlow called.")
	if defaultGenkitInstance == nil {
		slog.Error("Default genkit instance not initialized for GetGreetingFlow. Call Init() or InitializeGenkit() first.")
		return nil
	}
	// Removed: flow := genkit.LookupAction(defaultGenkitInstance, "greetingFlow") as LookupAction is undefined and flow was unused.
	var found bool
	for _, f := range genkit.ListFlows(defaultGenkitInstance) {
		if f.Name() == "greetingFlow" {
			found = true
			break
		}
	}
	if !found {
		slog.Error("greetingFlow not found in default genkit instance. Ensure RegisterFlows was successful.")
		return nil
	}
	return &legacyFlowRunner{name: "greetingFlow", g: defaultGenkitInstance}
}

// GetBackupFlow retrieves a runner for the pre-registered backup flow.
func GetBackupFlow() interface{} {
	slog.Warn("Legacy GetBackupFlow called.")
	if defaultGenkitInstance == nil {
		slog.Error("Default genkit instance not initialized for GetBackupFlow. Call Init() or InitializeGenkit() first.")
		return nil
	}
	var found bool
	for _, f := range genkit.ListFlows(defaultGenkitInstance) {
		if f.Name() == "backupFlow" {
			found = true
			break
		}
	}
	if !found {
		slog.Error("backupFlow not found in default genkit instance. Ensure RegisterFlows was successful.")
		return nil
	}
	return &legacyFlowRunner{name: "backupFlow", g: defaultGenkitInstance}
}

// --- Legacy Tool Registration ---

// LegacyRegisterBackupTool defines and registers the 'performBackup' tool.
func LegacyRegisterBackupTool(gInter interface{}, dfs *deskfs.DesktopFS, cdb *db.CentralDBProvider) {
	slog.Warn("LegacyRegisterBackupTool called.")
	// The gInter is the old *genkit.Genkit. We should use defaultGenkitInstance.
	if defaultGenkitInstance == nil {
		slog.Error("Default genkit instance not initialized for LegacyRegisterBackupTool. Call Init() or InitializeGenkit() first.")
		return
	}

	backupToolHandler := func(ctx *ai.ToolContext, input BackupToolInput) (BackupToolOutput, error) {
		slog.Info("Legacy performBackup tool invoked", "input", input)
		// Actual backup logic using dfs and cdb
		// This is a simplified placeholder.
		deskFSPath := "mock/deskfs/backup/path"
		centralDBPath := "mock/centraldb/backup/path"
		successMsg := "Backup performed successfully (legacy tool)."

		if dfs == nil || cdb == nil {
			return BackupToolOutput{}, errors.New("DeskFS or CentralDB provider is nil in backup tool")
		}
		// Simulate backup operations
		slog.Info("Simulating DeskFS backup...", "target", deskFSPath)
		slog.Info("Simulating CentralDB backup...", "target", centralDBPath)

		return BackupToolOutput{
			DeskFSBackupPath:    deskFSPath,
			CentralDBBackupPath: centralDBPath,
			SuccessMessage:      successMsg,
		}, nil
	}

	_, err := DefineTool(defaultGenkitInstance, "performBackup", "Performs a backup of application data.", backupToolHandler)
	if err != nil {
		slog.Error("Failed to register legacy performBackup tool", "error", err)
	} else {
		slog.Info("Legacy performBackup tool registered.")
	}
}

// RegisterOrganizeTool (legacy) defines and registers the 'organizeDirectory' tool.
func RegisterOrganizeTool(gInter interface{}, dfs *deskfs.DesktopFS) {
	slog.Warn("Legacy RegisterOrganizeTool called.")
	if defaultGenkitInstance == nil {
		slog.Error("Default genkit instance not initialized for RegisterOrganizeTool.")
		return
	}

	organizeHandler := func(ctx *ai.ToolContext, input interface{}) (interface{}, error) {
		// TODO: Implement this function when directory organization functionality is ready
		slog.Info("Legacy organizeDirectory tool invoked", "input", input)
		return map[string]string{
			"status": "Directory organization not implemented yet (legacy tool)",
		}, nil
	}

	_, err := DefineTool(defaultGenkitInstance, "organizeDirectory", "Organizes a directory using AI.", organizeHandler)
	if err != nil {
		slog.Error("Failed to register legacy organizeDirectory tool", "error", err)
	} else {
		slog.Info("Legacy organizeDirectory tool registered.")
	}
}

// RegisterWorkspaceTool (legacy) defines and registers the 'manageWorkspace' tool.
func RegisterWorkspaceTool(gInter interface{}, dfs *deskfs.DesktopFS) {
	slog.Warn("Legacy RegisterWorkspaceTool called.")
	if defaultGenkitInstance == nil {
		slog.Error("Default genkit instance not initialized for RegisterWorkspaceTool.")
		return
	}

	workspaceHandler := func(ctx *ai.ToolContext, input interface{}) (interface{}, error) {
		// TODO: Implement this function when workspace management functionality is ready
		slog.Info("Legacy manageWorkspace tool invoked", "input", input)
		return map[string]string{
			"status": "Workspace management not implemented yet (legacy tool)",
		}, nil
	}

	_, err := DefineTool(defaultGenkitInstance, "manageWorkspace", "Manages and configures workspaces with AI.", workspaceHandler)
	if err != nil {
		slog.Error("Failed to register legacy manageWorkspace tool", "error", err)
	} else {
		slog.Info("Legacy manageWorkspace tool registered.")
	}
}

// Note: The original `types.go` defined BackupToolInput, BackupToolOutput,
// greetingFlowRunner, backupFlowRunner, and SessionID.
// The flow runners are effectively replaced by legacyFlowRunner.
// Input/Output types are still needed. SessionID is independent.
// These might need to be moved to a shared `types.go` or `datatypes.go` if they are
// used by both legacy and new code, or kept here if only for legacy.
// For now, assuming they are defined elsewhere or will be added to this package.
