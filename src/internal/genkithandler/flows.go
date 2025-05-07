package genkithandler

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// greetingFlowRunner stores the runner for the greetingFlow.
// It's populated by RegisterFlows.
var greetingFlowRunner *core.Flow[string, string, struct{}]

// RegisterFlows defines and registers all flows with the given Genkit instance.
// It stores the defined flow runners for later retrieval.
func RegisterFlows(g *genkit.Genkit) error {
	// Define a simple flow that takes a name and returns a greeting.
	// The flow function captures 'g' from the RegisterFlows scope to use for Genkit operations.
	definedFlow := genkit.DefineFlow(g, "greetingFlow",
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

	if definedFlow == nil {
		// This case should ideally not happen if DefineFlow succeeds without panic,
		// but good to check. genkit.DefineFlow might panic on error, or log.
		// Depending on genkit's error handling, this might need adjustment.
		// For now, we assume it returns nil on failure to register.
		return fmt.Errorf("failed to define flow 'greetingFlow'")
	}
	greetingFlowRunner = definedFlow // Store the runner

	// Define other flows here in the future and store their runners similarly
	return nil
}

// GetGreetingFlow retrieves the pre-registered greeting flow runner.
// Returns nil if the flow was not successfully registered or RegisterFlows hasn't been called.
func GetGreetingFlow() *core.Flow[string, string, struct{}] {
	return greetingFlowRunner
}