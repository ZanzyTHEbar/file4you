package cli_util

import (
	"context"
	"errors"
	"fmt"

	"file4you/internal/cli"
	"file4you/internal/genkithandler"

	"github.com/spf13/cobra"
)

// NewGreetCmd creates a new cobra command for the greet flow.
func NewGreetCmd(params *cli.CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Greets a person using a Genkit flow",
		Long:  `Greets a person using a Genkit flow. This is a test command for Genkit integration.`,
		Args:  cobra.ExactArgs(1), // Expects exactly one argument: the name
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				// For a direct error from validation, we might not have a prior error object.
				// In this case, we create a new one.
				validationErr := errors.New("name argument cannot be empty")
				params.Interactor.Error("Validation failed", validationErr)
				return validationErr
			}

			params.Interactor.Output(fmt.Sprintf("Initializing Genkit to greet %s...", name))

			// Initialize Genkit
			g, err := genkithandler.InitializeGenkit(context.Background())
			if err != nil {
				params.Interactor.Error("Failed to initialize Genkit", err)
				return err // Return the original error
			}

			params.Interactor.Output("Genkit initialized. Registering flows...")

			// Register flows
			if err := genkithandler.RegisterFlows(g); err != nil {
				params.Interactor.Error("Failed to register Genkit flows", err)
				return err // Return the original error
			}

			params.Interactor.Output("Flows registered. Retrieving GreetingFlow...")

			// Get the GreetingFlow runner
			greetingFlow := genkithandler.GetGreetingFlow()
			if greetingFlow == nil {
				flowRetrievalErr := errors.New("GreetingFlow not found after registration")
				params.Interactor.Error("Failed to retrieve flow", flowRetrievalErr)
				return flowRetrievalErr
			}

			params.Interactor.Output("Running GreetingFlow...")

			// Run the GreetingFlow
			// The flow runner's Run method takes context and input.
			greeting, err := greetingFlow.Run(context.Background(), name)
			if err != nil {
				params.Interactor.Error("GreetingFlow execution failed", err)
				return err // Return the original error
			}

			params.Interactor.Output(fmt.Sprintf("Flow response: %s", greeting))
			return nil
		},
	}
	return cmd
}
