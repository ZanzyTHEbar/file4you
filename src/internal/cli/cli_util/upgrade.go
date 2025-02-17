package cli_util

import (
	"file4you/internal/cli"
	"file4you/internal/terminal"
	"fmt"

	"github.com/spf13/cobra"
)

type UpgradeCMD struct {
	Upgrade *cobra.Command
}

var UpgradeShowAll bool

func NewUpgrade(params *cli.CmdParams) *cobra.Command {
	UpgradeCmd := &cobra.Command{
		Use:     "upgrade ",
		Aliases: []string{"u"},
		Short:   "Upgrade DesktopCleaner to the latest version",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			upgrade(params)
		},
	}

	return UpgradeCmd
}

func upgrade(params *cli.CmdParams) {

	params.Term.ToggleSpinner(true, "Checking for updates ...")

	// Trigger Upgrade logic from the Upgrade.go file

	upgrade := terminal.NewUpgrade(params.Term)

	upgrade.CheckForUpgrade()

	fmt.Println("✅ Context is up to date")
}
