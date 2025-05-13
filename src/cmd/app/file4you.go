package main

import (
	"file4you/internal/cli"
	"file4you/internal/cli/cli_util"
	"file4you/internal/cli/fs"
	"file4you/internal/cli/git"
	"file4you/internal/cli/workspace"
	"file4you/internal/config" // Added for configuration loading
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/terminal"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// Load application configuration
	if _, err := config.LoadConfig(""); err != nil {
		slog.Error("Failed to load application configuration:", "msg", err)
		os.Exit(1)
	}

	// Setup the Dependancy Injection

	term := terminal.NewTerminal()
	// Create the CobraInteractor, which implements the ui.Interactor interface
	interactor := cli.NewCobraInteractor(term)

	// Initialize the Central Database
	centralDB, err := db.NewCentralDBProvider()
	if err != nil {
		// Use interactor for fatal errors if it's already initialized and makes sense
		// For early init errors, slog or direct stderr might still be appropriate
		slog.Error("Failed to initialize central database:", "msg", err)
		os.Exit(1)
	}
	// DeskFS might also need the interactor if it performs UI operations directly
	// Or, its methods should return errors/data to be handled by the caller (command/flow)
	deskFS := deskfs.NewDesktopFS(term, centralDB) // Consider if deskFS needs interactor
	defer centralDB.Close()

	// Setup the Root Command
	// CmdParams now takes ui.Interactor
	rootParams := &cli.CmdParams{
		Interactor: interactor,
		DeskFS:     deskFS,
		CentralDB:  centralDB,
		// Genkit will be initialized in root.go's OnInitialize
	}

	// It's important that the palette is assigned to rootParams.Palette
	// *before* NewRootCMD is called, because NewRootCMD (specifically NewRoot)
	// might iterate over this palette (e.g., to add commands).
	// The Genkit initialization in OnInitialize will happen *after* NewRoot has run
	// but *before* any command's RunE is executed.
	rootParams.Palette = generatePalette(rootParams)

	rootCmd := cli.NewRootCMD(rootParams)

	if err := rootCmd.Root.Execute(); err != nil {
		// Use the interactor for displaying the final execution error
		interactor.Error("Error executing root command", err)
		slog.Error(fmt.Sprintf("Error executing root command: %v", err.Error()))
		os.Exit(1) // Interactor.Fatal could also be used if it calls os.Exit
	}
}

func generatePalette(params *cli.CmdParams) []*cobra.Command {

	rewindCmd := git.NewRewind(params)
	rewind := cli.NewFile4YouCMD(rewindCmd).Root

	helpUtil := cli.NewFile4YouCMD(cli_util.NewHelp(params)).Root
	versionUtil := cli.NewFile4YouCMD(cli_util.NewVersion(params)).Root
	upgradeUtil := cli.NewFile4YouCMD(cli_util.NewUpgrade(params)).Root
	backupUtil := cli.NewFile4YouCMD(cli_util.NewBackupCmd(params)).Root // Added Backup command
	clearUtil := cli.NewFile4YouCMD(cli_util.NewClearCmd(params)).Root   // Added Clear command
	organize := cli.NewFile4YouCMD(fs.NewOrganize(params)).Root
	workspaceCmd := workspace.NewWorkspace(params)
	ws := cli.NewFile4YouCMD(workspaceCmd).Root
	greetCmd := cli_util.NewGreetCmd(params)

	// Add commands here
	return []*cobra.Command{
		rewind,
		helpUtil,
		versionUtil,
		upgradeUtil,
		backupUtil, 
		clearUtil,
		organize,
		ws,
		greetCmd,
	}
}
