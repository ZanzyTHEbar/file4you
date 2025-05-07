package genkithandler

import (
	"context"
	"fmt"

	// Added for action.Run
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// greetingFlowRunner stores the runner for the greetingFlow.
// It's populated by RegisterFlows.
var greetingFlowRunner *core.Flow[string, string, struct{}]
var backupFlowRunner *core.Flow[BackupToolInput, BackupToolOutput, struct{}] // Added for backup flow

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
		// This case should ideally not happen if DefineFlow succeeds without panic,
		// but good to check. genkit.DefineFlow might panic on error, or log.
		// Depending on genkit's error handling, this might need adjustment.
		// For now, we assume it returns nil on failure to register.
		return fmt.Errorf("failed to define flow 'greetingFlow'")
	}
	greetingFlowRunner = definedGreetingFlow // Store the runner

	// Define the backup flow
	// This flow takes BackupToolInput and uses the performBackup tool (action).
	definedBackupFlow := genkit.DefineFlow(g, "backupFlow",
		func(ctx context.Context, input BackupToolInput) (BackupToolOutput, error) {
			// Look up the registered action (tool) using the Genkit instance 'g',
			// which acts as an ActionRegistry.
			actionDef := g.LookupAction("performBackup")
			if actionDef == nil {
				return BackupToolOutput{}, fmt.Errorf("backupFlow: action 'performBackup' not found in registry")
			}

			// Run the action.
			// The input to actionDef.Run is 'any'; 'input' is already BackupToolInput.
			// The output from actionDef.Run is 'any'; we need to assert it to BackupToolOutput.
			// Passing nil for core.RunnerOptions for now.
			actionOutputAny, err := actionDef.Run(ctx, input, nil)
			if err != nil {
				return BackupToolOutput{}, fmt.Errorf("backupFlow: failed to run performBackup action: %w", err)
			}

			// Type assert the output to the expected BackupToolOutput type.
			output, ok := actionOutputAny.(BackupToolOutput)
			if !ok {
				return BackupToolOutput{}, fmt.Errorf("backupFlow: action 'performBackup' returned unexpected type: got %T, want BackupToolOutput", actionOutputAny)
			}

			return output, nil
		},
	)

	if definedBackupFlow == nil {
		return fmt.Errorf("failed to define flow 'backupFlow'")
	}
	backupFlowRunner = definedBackupFlow // Store the runner

	// Define other flows here in the future and store their runners similarly
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