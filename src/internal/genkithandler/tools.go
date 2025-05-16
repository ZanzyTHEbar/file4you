// Package genkithandler provides a simplified interface for integrating with Genkit.
package genkithandler

import (
	"context"
	"encoding/json"
	"fmt"

	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/genkithandler/errors" // Import custom errors

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// DefineTool defines a new Genkit tool (action) and registers it with the provided Genkit instance.
// It's a wrapper around genkit.DefineTool.
// The fn parameter is the actual tool implementation.
// Its signature must match what genkit.DefineTool expects: func(ctx *ai.ToolContext, input In) (Out, error)
func DefineTool[In, Out any](
	g *genkit.Genkit,
	name, description string,
	fn func(ctx *ai.ToolContext, input In) (Out, error),
) (ai.Tool, error) {
	if g == nil {
		return nil, errors.New("genkit instance is nil")
	}
	if name == "" {
		return nil, errors.New("tool name cannot be empty")
	}
	if description == "" {
		// Consider if description can be optional or if an error is appropriate.
		// For now, maintaining previous logic.
		return nil, errors.New("tool description cannot be empty")
	}
	if fn == nil {
		return nil, errors.New("tool function (fn) cannot be nil")
	}

	// genkit.DefineTool takes the Genkit instance, name, description, and the function.
	tool := genkit.DefineTool(g, name, description, fn)
	if tool == nil {
		// This condition might be hard to hit if DefineTool panics or always returns non-nil.
		return nil, errors.Errorf("failed to define tool %s", name)
	}
	return tool, nil
}

// LookupTool retrieves a previously defined Genkit tool by its name from the Genkit instance.
// Returns nil if the tool is not found.
// Note: The original audit mentioned genkit.LookupTool returns (tool, error).
// The provided file content shows it returns ai.Tool directly. Adhering to file content.
func LookupTool(g *genkit.Genkit, name string) ai.Tool {
	if g == nil || name == "" {
		return nil
	}
	return genkit.LookupTool(g, name)
}

/*
RegisterCoreTools registers core tools (performBackup, organizeDirectory, manageWorkspace) using the new Genkit API.
Call this during Genkit initialization.
*/
func RegisterCoreTools(g *genkit.Genkit, dfs *deskfs.DesktopFS, cdb *db.CentralDBProvider) error {
	// performBackup tool
	backupToolHandler := func(ctx *ai.ToolContext, input BackupToolInput) (BackupToolOutput, error) {
		deskFSPath := "mock/deskfs/backup/path"
		centralDBPath := "mock/centraldb/backup/path"
		successMsg := "Backup performed successfully (new system)."
		if dfs == nil || cdb == nil {
			return BackupToolOutput{}, errors.New("DeskFS or CentralDB provider is nil in backup tool")
		}
		return BackupToolOutput{
			DeskFSBackupPath:    deskFSPath,
			CentralDBBackupPath: centralDBPath,
			Message:             successMsg,
		}, nil
	}
	// DEBUG: Print tool registration
	fmt.Println("DEBUG: Registering performBackup tool")
	if _, err := DefineTool[BackupToolInput, BackupToolOutput](g, "performBackup", "Performs a backup of application data.", backupToolHandler); err != nil {
		return err
	}

	// organizeDirectory tool
	organizeHandler := func(ctx *ai.ToolContext, input interface{}) (map[string]string, error) {
		return map[string]string{
			"status": "Directory organization not implemented yet (new system)",
		}, nil
	}
	if _, err := DefineTool[interface{}, map[string]string](g, "organizeDirectory", "Organizes a directory using AI.", organizeHandler); err != nil {
		return err
	}

	// manageWorkspace tool
	workspaceHandler := func(ctx *ai.ToolContext, input interface{}) (map[string]string, error) {
		return map[string]string{
			"status": "Workspace management not implemented yet (new system)",
		}, nil
	}
	if _, err := DefineTool[interface{}, map[string]string](g, "manageWorkspace", "Manages and configures workspaces with AI.", workspaceHandler); err != nil {
		return err
	}
	return nil
}

// ExecuteTool looks up a tool by name and executes it with the provided input.
// This is a convenience wrapper.
func ExecuteTool[In, Out any](
	ctx context.Context,
	g *genkit.Genkit,
	toolName string,
	input In,
) (Out, error) {
	var zeroOut Out
	if g == nil {
		return zeroOut, errors.New("genkit instance is nil")
	}
	if toolName == "" {
		return zeroOut, errors.New("tool name cannot be empty for execution")
	}

	tool := LookupTool(g, toolName)
	if tool == nil {
		return zeroOut, errors.NewToolNotFoundError(toolName, nil)
	}

	outputRaw, err := tool.RunRaw(ctx, input)
	if err != nil {
		return zeroOut, errors.Wrapf(err, "tool '%s' execution failed", toolName)
	}

	var output Out
	if m, ok := outputRaw.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(m)
		if err != nil {
			return zeroOut, errors.Wrapf(err, "failed to marshal tool '%s' output", toolName)
		}
		if err := json.Unmarshal(jsonData, &output); err != nil {
			return zeroOut, errors.Wrapf(err, "failed to unmarshal tool '%s' output", toolName)
		}
	} else if typedOutput, ok := outputRaw.(Out); ok {
		output = typedOutput
	} else {
		typeErr := errors.Errorf("unexpected output type for tool '%s': %T", toolName, outputRaw)
		return zeroOut, errors.WithCode(typeErr, "TYPE_ASSERTION_FAILED")
	}

	return output, nil
}
