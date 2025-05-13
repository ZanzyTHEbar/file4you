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
		Long:  `Greets a person using a Genkit flow. This command relies on Genkit being initialized centrally.`,
		Args:  cobra.ExactArgs(1), // Expects exactly one argument: the name
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				validationErr := errors.New("name argument cannot be empty")
				params.Interactor.Error("Validation failed", validationErr)
				return validationErr
			}

			params.Interactor.Output(fmt.Sprintf("Attempting to greet %s using Genkit...", name))

			// Check if Genkit was initialized centrally
			if params.Genkit == nil {
				initErr := errors.New("genkit not initialized; this should have happened in root command initialization")
				params.Interactor.Error("Genkit initialization error", initErr)
				return initErr
			}
			params.Interactor.Output("Genkit is initialized. Retrieving GreetingFlow...")

			// Get the GreetingFlow runner
			// Assumes flows were registered during the central Genkit initialization
			flowRunnerInter := genkithandler.GetGreetingFlow()
			if flowRunnerInter == nil {
				flowRetrievalErr := errors.New("GreetingFlow runner not found; ensure it was registered")
				params.Interactor.Error("Failed to retrieve flow runner", flowRetrievalErr)
				return flowRetrievalErr
			}

			// Type assert to the specific runner type defined in legacy.go
			type legacyFlowRunner interface {
				Run(ctx context.Context, input interface{}) (interface{}, error)
			}

			greetingFlowRunner, ok := flowRunnerInter.(legacyFlowRunner)
			if !ok {
				typeErr := errors.New("retrieved flow runner is not of expected type (legacyFlowRunner)")
				params.Interactor.Error("Type assertion failed", typeErr)
				return typeErr
			}

			params.Interactor.Output("Running GreetingFlow...")

			// Run the GreetingFlow
			response, err := greetingFlowRunner.Run(context.Background(), name)
			if err != nil {
				params.Interactor.Error("GreetingFlow execution failed", err)
				return err
			}

			greeting, ok := response.(string)
			if !ok {
				typeErr := errors.New("flow response is not of expected type (string)")
				params.Interactor.Error("Flow response type error", typeErr)
				return typeErr
			}

			params.Interactor.Output(fmt.Sprintf("Flow response: %s", greeting))
			return nil
		},
	}
	return cmd
}
