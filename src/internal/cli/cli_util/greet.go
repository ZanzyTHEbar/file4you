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
			params.Interactor.Output("Genkit is initialized. Running GreetingFlow...")

			greeting, err := genkithandler.ExecuteFlow[string, string](
				context.Background(),
				params.Genkit,
				"greetingFlow",
				name,
			)
			if err != nil {
				params.Interactor.Error("GreetingFlow execution failed", err)
				return err
			}

			params.Interactor.Output(fmt.Sprintf("Flow response: %s", greeting))
			return nil
		},
	}
	return cmd
}
