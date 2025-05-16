package internal

import (
	"os"
	"path/filepath"
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
		panic("Unable to get home directory")
	}
	return homeDir
}
