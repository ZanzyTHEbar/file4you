package deskfs

import (
	// "context" // No longer used
	"file4you/internal"
	"file4you/internal/filesystem/trees"
	"file4you/internal/ui" // Added for Interactor
	"fmt"
	"log/slog" // Keep for internal/debug logging not directly for user UI
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ZanzyTHEbar/assert-lib"
	gobaselogger "github.com/ZanzyTHEbar/go-basetools/logger"
)

var (
	ConfigAssertHandler = assert.NewAssertHandler()
)

// Config holds the mapping of file types to extensions
type DeskFSConfig struct {
	gobaselogger.Config
	FileTypeTree           *trees.FileTypeTree `toml:"file_type_tree"`
	TargetDir              string              `toml:"target_dir"`
	CacheDir               string              `toml:"cache_dir"`
	Database               DatabaseConfig      `toml:"database"`                 // Added Database config
	OrganizeTimeoutMinutes int                 `toml:"organize_timeout_minutes"` // Timeout for the EnhancedOrganize operation
}

// Added DatabaseConfig struct
type DatabaseConfig struct {
	DSN  string `toml:"dsn"`
	Type string `toml:"type"`
}

type IntermediateConfig struct {
	gobaselogger.Config
	FileTypes map[string][]string `toml:"file_types"` // Ensure TOML tag matches the file
	CacheDir  string              `toml:"cache_dir"`
}

func CreateDirIfNotExist(path string, interactor ui.Interactor) { // Added Interactor
	// Create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			// slog.Info(fmt.Sprintf("Path %s: %v", filepath.Dir(path), err)) // Internal log
			errMsg := fmt.Sprintf("Error creating directory at %s", path)
			// ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error) // Internal assert
			if interactor != nil {
				interactor.Error(errMsg, err)
			} else {
				slog.Error(errMsg, "error", err) // Fallback if no interactor
			}
		}
	}
}

func NewIntermediateConfig(optionalPath string, interactor ui.Interactor) *IntermediateConfig { // Added Interactor
	var configPath string

	// Step 1: Determine the configuration file path
	if optionalPath != "" {
		if _, err := os.Stat(optionalPath); err != nil {
			if interactor != nil {
				interactor.Warning(fmt.Sprintf("Invalid optional config path provided: %s. Error: %v", optionalPath, err))
			} else {
				slog.Warn(fmt.Sprintf("Invalid optional config path provided: %s. Error: %v", optionalPath, err))
			}
			// Decide on fallback behavior: return nil, use default, or attempt to create?
			// For now, let's try to proceed to default global config if optional is bad.
			optionalPath = "" // Clear it so it falls through to default logic
		}
		configPath = optionalPath
	}

	// Fallback logic if optionalPath was not provided or was invalid
	if configPath == "" {
		if _, err := os.Stat(internal.DefaultWorkspaceConfigFile); err == nil {
			configPath = internal.DefaultWorkspaceConfigFile
			if interactor != nil {
				interactor.Info(fmt.Sprintf("Using workspace config file: %s", configPath))
			} else {
				slog.Info(fmt.Sprintf("Using workspace config file: %s", configPath))
			}
		} else {
			configPath = internal.DefaultGlobalConfigFile
			if interactor != nil {
				interactor.Info(fmt.Sprintf("Using global config file: %s", configPath))
			} else {
				slog.Info(fmt.Sprintf("Using global config file: %s", configPath))
			}
		}
	} else {
		if interactor != nil {
			interactor.Info(fmt.Sprintf("Using specified config file: %s", configPath))
		} else {
			slog.Info(fmt.Sprintf("Using specified config file: %s", configPath))
		}
	}

	var defaultConfig IntermediateConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig = getDefaultConfig()
		if interactor != nil {
			interactor.Info(fmt.Sprintf("Config file not found at %s. Creating default config.", configPath))
		} else {
			slog.Info(fmt.Sprintf("Config file not found at %s. Creating default config.", configPath))
		}
		CreateDirIfNotExist(filepath.Dir(configPath), interactor)
		file, err_create := os.Create(configPath)
		if err_create != nil {
			if interactor != nil {
				interactor.Error(fmt.Sprintf("Error creating default config file at %s", configPath), err_create)
			} else {
				slog.Error(fmt.Sprintf("Error creating default config file at %s: %v", configPath, err_create))
			}
			return nil // Or a default in-memory config
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		if err_encode := encoder.Encode(defaultConfig); err_encode != nil {
			if interactor != nil {
				interactor.Error(fmt.Sprintf("Error writing default config to %s", configPath), err_encode)
			} else {
				slog.Error(fmt.Sprintf("Error writing default config to %s: %v", configPath, err_encode))
			}
			return nil // Or a default in-memory config
		}
		if interactor != nil {
			interactor.Success(fmt.Sprintf("Default config file created at %s", configPath))
		} else {
			slog.Info(fmt.Sprintf("Default config file created at %s", configPath))
		}
	} else {
		if interactor != nil {
			interactor.Info(fmt.Sprintf("Loading config file from %s", configPath))
		} else {
			slog.Info(fmt.Sprintf("Loading config file from %s", configPath))
		}
		// var tempConfig map[string]interface{} // Keep for debug if needed
		// if _, err_decode_map := toml.DecodeFile(configPath, &tempConfig); err_decode_map != nil {
		// 	slog.Error(fmt.Sprintf("Error decoding config file to map: %v", err_decode_map)) // Keep as slog for debug
		// 	// return nil // Don't fail here, try to decode to struct
		// }
		// slog.Debug(fmt.Sprintf("TempConfig (raw): %+v\n", tempConfig)) // Keep as slog for debug

		if _, err_decode_struct := toml.DecodeFile(configPath, &defaultConfig); err_decode_struct != nil {
			if interactor != nil {
				interactor.Error(fmt.Sprintf("Error decoding config file %s into struct", configPath), err_decode_struct)
			} else {
				slog.Error(fmt.Sprintf("Error decoding config file %s into struct: %v", configPath, err_decode_struct))
			}
			// Return a default config or handle error appropriately
			// For now, returning an empty struct to avoid nil pointer, but signaling failure is important.
			interactor.Warning("Returning empty default config due to decoding error.")
			return &IntermediateConfig{}
		}
	}

	slog.Debug(fmt.Sprintf("Loaded file_types (case-sensitive): %+v\n", defaultConfig.FileTypes)) // Keep as slog for debug

	return &defaultConfig
}

func NewDeskFSConfig() *DeskFSConfig {
	return &DeskFSConfig{
		FileTypeTree:           trees.NewFileTypeTree(),
		OrganizeTimeoutMinutes: 10, // Default to 10 minutes
	}
}

func (dfc *DeskFSConfig) BuildFileTypeTree(config *IntermediateConfig) *DeskFSConfig {
	// Populate FileTypeTree using the intermediate config data
	dfc.FileTypeTree.PopulateFileTypes(config.FileTypes)
	return dfc
}

func (dfc *IntermediateConfig) SaveConfig(config *IntermediateConfig, filePath string) error {
	dfc.Config.Cfg.Set("file_types", config.FileTypes)
	dfc.Config.Cfg.Set("logger.style", config.Logger.Style)
	dfc.Config.Cfg.Set("logger.level", config.Logger.Level)
	dfc.Config.Cfg.Set("cache_dir", config.CacheDir)

	if err := dfc.Config.Cfg.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// Returns the default configuration
// Baseline configuration is used within the cwd of the user, and is just an example map of file names to extensions.
// We can support more metrics other than file types.
func getDefaultConfig() IntermediateConfig {
	return IntermediateConfig{
		FileTypes: map[string][]string{
			"Notes":      {".md", ".rtf", ".txt"},
			"Docs":       {".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"},
			"EXE":        {".exe", ".appimage", ".msi"},
			"Vids":       {".mp4", ".mov", ".avi", ".mkv"},
			"Compressed": {".zip", ".rar", ".tar", ".gz", ".7z"},
			"Scripts":    {".sh", ".bat"},
			"Installers": {".deb", ".rpm"},
			"Books":      {".epub", ".mobi"},
			"Music":      {".mp3", ".wav", ".ogg", ".flac"},
			"PDFS":       {".pdf"},
			"Pics":       {".bmp", ".gif", ".jpg", ".jpeg", ".svg", ".png"},
			"Torrents":   {".torrent"},
			"CODE": {
				".c", ".h", ".py", ".rs", ".go", ".js", ".ts", ".jsx", ".tsx", ".html",
				".css", ".php", ".java", ".cpp", ".cs", ".vb", ".sql", ".pl", ".swift",
				".kt", ".r", ".m", ".asm",
			},
			"Markup": {
				".json", ".xml", ".yml", ".yaml", ".ini", ".toml", ".cfg", ".conf", ".log",
			},
		},
		Config: gobaselogger.Config{
			Logger: gobaselogger.Logger{
				Style: "json",
				Level: gobaselogger.LoggerLevels["debug"].String(),
			},
			// Add default database config here if IntermediateConfig is used to write full DeskFSConfig
			// For now, assuming DeskFSConfig is populated from IntermediateConfig and then Database part is set
		},
		CacheDir: internal.DefaultCacheDir,
		// Note: The getDefaultConfig currently returns IntermediateConfig.
		// The DeskFSConfig which includes DatabaseConfig will be populated from this
		// and other sources (like viper for specific DB settings if not in the primary TOML).
		// If the TOML is expected to contain DB settings directly under a [database] table,
		// then IntermediateConfig might also need to reflect that, or the parsing logic
		// in NewIntermediateConfig needs to handle it.
		// For now, we assume the main config file loaded into DeskFSConfig will have the [database] section.
	}
}

// Helper function to get default DeskFSConfig including database defaults
// This can be used when a config file is missing entirely or the database section is absent.
func GetDefaultDeskFSConfig() *DeskFSConfig {
	return &DeskFSConfig{
		Config: gobaselogger.Config{
			Logger: gobaselogger.Logger{
				Style: "json",
				Level: gobaselogger.LoggerLevels["debug"].String(),
			},
		},
		FileTypeTree: trees.NewFileTypeTree(), // Initialize with an empty tree
		TargetDir:    ".",                     // Default to current directory
		CacheDir:     internal.DefaultCacheDir,
		Database: DatabaseConfig{
			DSN:  internal.DefaultDatabaseDSN,  // Assuming this constant exists
			Type: internal.DefaultDatabaseType, // Assuming this constant exists
		},
		OrganizeTimeoutMinutes: 10, // Default to 10 minutes
	}
}
