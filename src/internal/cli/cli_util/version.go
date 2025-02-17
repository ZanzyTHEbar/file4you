package cli_util

import (
	"file4you/internal/cli"
	"file4you/version"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

type VersionCMD struct {
	Version *cobra.Command
}

func NewVersion(params *cli.CmdParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of DesktopCleaner",
		Long:  `All software has versions. This is DesktopCleaner's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}

	return versionCmd
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
