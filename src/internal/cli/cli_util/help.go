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
		Short:   "Display help for DesktopCleaner",
		Long:    `Display help for DesktopCleaner.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: params.term.PrintCustomHelp(helpShowAll)
		},
	}

	// add an --all/-a flag
	helpCmd.Flags().BoolVarP(&helpShowAll, "all", "a", false, "Show all commands")

	return helpCmd
}
