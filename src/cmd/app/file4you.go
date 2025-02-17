package main

import (
	"file4you/internal/cli"
	"file4you/internal/cli/cli_util"
	"file4you/internal/cli/fs"
	"file4you/internal/cli/git"
	"file4you/internal/cli/workspace"
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/terminal"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// Setup the Dependancy Injection

	term := terminal.NewTerminal()

	// Initialize the Central Database
	centralDB, err := db.NewCentralDBProvider()
	if err != nil {
		slog.Error("Failed to initialize central database:", "msg", err)
		os.Exit(1)
	}

	deskFS := deskfs.NewDesktopFS(term, centralDB)
	defer centralDB.Close()

	// Setup the Root Command
	rootParams := &cli.CmdParams{
		Term:      term,
		DeskFS:    deskFS,
		CentralDB: centralDB,
	}

	palette := generatePalette(rootParams)
	rootParams.Palette = palette

	rootCmd := cli.NewRootCMD(rootParams)

	if err := rootCmd.Root.Execute(); err != nil {
		term.OutputErrorAndExit("Error executing root command: %v", err)
		slog.Error(fmt.Sprintf("Error executing root command: %v", err.Error()))
	}
}

func generatePalette(params *cli.CmdParams) []*cobra.Command {

	rewindCmd := git.NewRewind(params)
	rewind := cli.NewFile4YouCMD(rewindCmd).Root
	helpUtil := cli.NewFile4YouCMD(cli_util.NewHelp(params)).Root
	versionUtil := cli.NewFile4YouCMD(cli_util.NewVersion(params)).Root
	upgradeUtil := cli.NewFile4YouCMD(cli_util.NewUpgrade(params)).Root
	organize := cli.NewFile4YouCMD(fs.NewOrganize(params)).Root
	workspace := cli.NewFile4YouCMD(workspace.NewWorkspace(params)).Root

	// Add commands here
	return []*cobra.Command{
		rewind,
		helpUtil,
		versionUtil,
		upgradeUtil,
		organize,
		workspace,
	}
}
