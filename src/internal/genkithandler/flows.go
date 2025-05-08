// Package genkithandler provides integration with the genkit AI platform.
package genkithandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// RegisterFlows defines and registers all flows with the given Genkit instance.
// It stores the defined flow runners for later retrieval.
func RegisterFlows(g *genkit.Genkit) error {
	// Define a simple flow that takes a name and returns a greeting.
	// The flow function captures 'g' from the RegisterFlows scope to use for Genkit operations.
	definedGreetingFlow := genkit.DefineFlow(g, "greetingFlow",
		func(ctx context.Context, name string) (string, error) {
			model := googlegenai.GoogleAIModel(g, "gemini-1.5-flash")
			if model == nil {
				return "", fmt.Errorf("model '%s' not found in genkit instance", "gemini-1.5-flash")
			}

			prompt := "Write a friendly greeting to " + name + "."

			resp, err := genkit.GenerateText(ctx, g, ai.WithModel(model), ai.WithPrompt(prompt))
			if err != nil {
				return "", fmt.Errorf("GenerateText failed: %w", err)
			}

			return resp, nil
		},
	)

	if definedGreetingFlow == nil {
		return fmt.Errorf("failed to define flow 'greetingFlow'")
	}
	greetingFlowRunner = definedGreetingFlow // Store the runner

	// Define the backup flow with a stub implementation
	definedBackupFlow := genkit.DefineFlow(g, "backupFlow",
		func(ctx context.Context, input BackupToolInput) (BackupToolOutput, error) {
			slog.Info("backupFlow called - using stub implementation")

			// TODO: In a real implementation, we would call the proper tool
			// For now, we'll implement the backup logic directly
			var output BackupToolOutput
			output.SuccessMessage = "Backup flow executed successfully (stub implementation)"

			// Here we would actually call the backup service or function
			// output.DeskFSBackupPath = "path/to/deskfs/backup"
			// output.CentralDBBackupPath = "path/to/centraldb/backup"

			return output, nil
		},
	)

	if definedBackupFlow == nil {
		return fmt.Errorf("failed to define flow 'backupFlow'")
	}
	backupFlowRunner = definedBackupFlow // Store the runner

	return nil
}

// GetGreetingFlow retrieves the pre-registered greeting flow runner.
// Returns nil if the flow was not successfully registered or RegisterFlows hasn't been called.
func GetGreetingFlow() *core.Flow[string, string, struct{}] {
	return greetingFlowRunner
}

// GetBackupFlow retrieves the pre-registered backup flow runner.
// Returns nil if the flow was not successfully registered or RegisterFlows hasn't been called.
func GetBackupFlow() *core.Flow[BackupToolInput, BackupToolOutput, struct{}] {
	return backupFlowRunner
}