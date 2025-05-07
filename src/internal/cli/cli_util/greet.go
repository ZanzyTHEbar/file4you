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
				return errors.New("name argument cannot be empty")
			}

			params.Interactor.Output(fmt.Sprintf("Initializing Genkit to greet %s...", name))

			// Initialize Genkit
			// Note: InitializeGenkit now returns (*genkit.Genkit, error)
			g, err := genkithandler.InitializeGenkit(context.Background())
			if err != nil {
				params.Interactor.Error(fmt.Errorf("failed to initialize Genkit: %w", err))
				return err
			}

			params.Interactor.Output("Genkit initialized. Running GreetingFlow...")

			// Run the GreetingFlow
			// The flow definition is: func(ctx context.Context, g *genkit.Genkit, name string) (string, error)
			greeting, err := genkithandler.GreetingFlow.Run(context.Background(), g, name)
			if err != nil {
				params.Interactor.Error(fmt.Errorf("greetingFlow failed: %w", err))
				return err
			}

			params.Interactor.Output(fmt.Sprintf("Flow response: %s", greeting))
			return nil
		},
	}
	return cmd
}
