package main

import (
	"file4you/internal/cli"
	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/terminal"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load application configuration. This will populate config.AppConfig
	// No need to assign to a new variable, LoadConfig modifies the global AppConfig directly.
	if _, err := config.LoadConfig(""); err != nil { // Pass empty string to use default config paths
		log.Fatal().Err(err).Msg("Failed to load application configuration")
	}

	// Now access the loaded configuration via the global config.AppConfig
	// cfg := config.AppConfig.File4You // This is File4YouConfig // Removed unused variable

	term := terminal.NewTerminal() // Corrected: No arguments
	interactor := cli.NewCobraInteractor(term)

	// Initialize CentralDBProvider. It uses internal.DefaultCentralDBPath by default.
	centralDB, err := db.NewCentralDBProvider() // Corrected: No arguments
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize central database provider")
	}
	// defer centralDB.Close() // Assuming centralDB has a Close method, add if available

	// Pass the interactor and the centralDB provider to NewDesktopFS
	// NewDesktopFS now correctly takes ui.Interactor and db.ICentralDBProvider
	deskFS := deskfs.NewDesktopFS(interactor, centralDB) // Corrected: single return value

	// Assuming CmdParams is the expected type for NewRootCMD
	// This might need further adjustment based on the actual definition of CmdParams
	// and how it's intended to be populated.
	// For now, creating an empty CmdParams.
	cmdParams := &cli.CmdParams{
		DeskFS:     deskFS,
		Interactor: interactor, // Corrected field name to Interactor
		CentralDB:  centralDB,
	}
	rootCmd := cli.NewRootCMD(cmdParams) // Corrected to NewRootCMD and passing params
	if err := rootCmd.Root.Execute(); err != nil { // Execute the Root field of the returned struct
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
