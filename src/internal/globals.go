package internal

import (
	"os"
	"path/filepath"
)

var (
	// DefaultConfigPath is the default path to the config file
	DefaultAppName             = "file4you"
	DefaultAppCMDShortCut      = "f4u"
	DefaultConfigFolderName    = DefaultAppName
	DefaultConfigPath          = filepath.Join(os.Getenv("HOME"), ".config", DefaultConfigFolderName)
	DefaultCacheDir            = filepath.Join(DefaultConfigPath, ".cache")
	DefaultCentralDBPath       = filepath.Join(DefaultConfigPath, "central.db")
	DefaultWorkspaceDotDir     = "." + DefaultConfigFolderName
	DefaultWorkspaceDBPath     = filepath.Join(DefaultWorkspaceDotDir, "workspace.db")
	DefaultWorkspaceConfigFile = filepath.Join(DefaultWorkspaceDotDir, "config.toml")
	DefaultGlobalConfigFile    = filepath.Join(DefaultConfigPath, "config.toml")
)
