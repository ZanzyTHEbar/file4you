package internal

import (
	"log"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var (
	// DefaultConfigPath is the default path to the config file
	DefaultAppName             = "file4you"
	DefaultAppCMDShortCut      = "f4u"
	DefaultConfigPath          = filepath.Join(getHomeDir(), ".config", DefaultAppName)
	DefaultCacheDir            = filepath.Join(DefaultConfigPath, ".cache")
	DefaultCentralDBPath       = filepath.Join(DefaultConfigPath, "central.db")
	DefaultWorkspaceDotDir     = "." + DefaultAppName
	DefaultWorkspaceDBPath     = filepath.Join(DefaultWorkspaceDotDir, "workspace.db")
	DefaultWorkspaceConfigFile = filepath.Join(DefaultWorkspaceDotDir, "config.toml")
	DefaultGlobalConfigFile    = filepath.Join(DefaultConfigPath, "config.toml")

	// Default Database settings
	DefaultDatabaseDSN  = "file::memory:?cache=shared" // Default to in-memory SQLite
	DefaultDatabaseType = "sqlite3"
)

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current working directory if home directory is unavailable
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			// Last resort - use tmp directory
			log.Printf("Unable to get home or working directory, using /tmp: %v", err)
			return "/tmp"
		}
		log.Printf("Unable to get home directory, using current working directory: %v", err)
		return cwd
	}
	return homeDir
}

// GetLogger returns a properly configured zerolog logger instance
func GetLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}
