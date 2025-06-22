package main

import (
	"file4you/internal/cli"
	"file4you/internal/cli/cli_util"
	fscli "file4you/internal/cli/fs"
	"file4you/internal/cli/workspace"
	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/filesystem"
	"file4you/internal/terminal"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	if _, err := config.LoadConfig(""); err != nil {
		log.Fatal().Err(err).Msg("Failed to load application configuration")
	}

	term := terminal.NewTerminal()
	interactor := cli.NewCobraInteractor(term)

	centralDB, err := db.NewCentralDBProvider()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize central database provider")
	}
	defer func() {
		log.Info().Msg("Closing database connections...")
		if closeErr := centralDB.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing central database")
		}
	}()

	fs, err := filesystem.New(interactor, centralDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize desktop file system")
	}

	cmdParams := &cli.CmdParams{
		Filesystem: fs,
		Interactor: interactor,
		CentralDB:  centralDB,
	}

	// Initialize command palette
	cmdParams.Palette = []*cobra.Command{
		// File system commands
		fscli.NewOrganize(cmdParams),
		fscli.NewAnalyze(cmdParams),
		fscli.NewSearch(cmdParams),

		// Workspace management commands
		workspace.NewWorkspace(cmdParams),

		// Utility commands
		cli_util.NewVersion(cmdParams),
		cli_util.NewConfig(cmdParams),
		cli_util.NewStatus(cmdParams),
	}

	rootCmd := cli.NewRootCMD(cmdParams)

	// Check if this is a simple command that doesn't need graceful shutdown
	args := os.Args[1:]
	isSimpleCommand := len(args) == 0 ||
		contains(args, "--help") || contains(args, "-h") ||
		contains(args, "--version") || contains(args, "-v") ||
		contains(args, "help") || contains(args, "completion")

	if isSimpleCommand {
		// For simple commands, just execute directly
		if err := rootCmd.Root.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	// For long-running commands, use graceful shutdown
	setupGracefulShutdown(rootCmd)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func setupGracefulShutdown(rootCmd *cli.RootCMD) {
	// Channel to listen for interrupt/terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Channel to signal command completion
	cmdDone := make(chan error, 1)

	// Run the command in a goroutine
	go func() {
		cmdDone <- rootCmd.Root.Execute()
	}()

	// Wait for either command completion or shutdown signal
	select {
	case err := <-cmdDone:
		// Command completed normally
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return
	case <-sigChan:
		log.Info().Msg("Received interrupt signal, shutting down gracefully...")
		// Give the command a moment to finish gracefully
		select {
		case err := <-cmdDone:
			log.Info().Msg("Application shutdown complete")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				os.Exit(1)
			}
		case <-time.After(5 * time.Second):
			log.Warn().Msg("Forced shutdown after timeout")
			os.Exit(1)
		}
	}
}
