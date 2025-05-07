package cli_util

import (
	"file4you/internal/cli"

	"github.com/spf13/cobra"
)

type HelpCMD struct {
	Help *cobra.Command
}

var helpShowAll bool

func NewHelp(params *cli.CmdParams) *cobra.Command {
	helpCmd := &cobra.Command{
		Use:     "detailed_help",
		Aliases: []string{"h"},
		Short:   "Display help for file4you", // Updated name
		Long:    `Display help for file4you. Shows custom help information.`, // Updated name
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			// Call the new ShowCustomHelp method on the Interactor.
			// cmd.CommandPath() provides a unique string for the command.
			params.Interactor.ShowCustomHelp(helpShowAll, cmd.CommandPath())
			return nil
		},
	}

	// add an --all/-a flag
	helpCmd.Flags().BoolVarP(&helpShowAll, "all", "a", false, "Show all commands")

	return helpCmd
}
