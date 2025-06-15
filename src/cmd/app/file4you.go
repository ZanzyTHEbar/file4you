package main

import (
	"file4you/internal/cli"
	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/filesystem"
	"file4you/internal/terminal"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// TODO: Implement graceful shutdown handling

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
	defer centralDB.Close()

	fs, err := filesystem.New(interactor, centralDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize desktop file system")
	}

	cmdParams := &cli.CmdParams{
		Filesystem:     fs,
		Interactor: interactor,
		CentralDB:  centralDB,
	}
	rootCmd := cli.NewRootCMD(cmdParams)
	if err := rootCmd.Root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
