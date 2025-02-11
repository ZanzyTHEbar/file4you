package deskfs

import (
	"context"
	"desktop-cleaner/internal"
	"desktop-cleaner/internal/filesystem/trees"
	"fmt"
	"log/slog"
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
	FileTypeTree *trees.FileTypeTree `toml:"file_type_tree"`
	TargetDir    string              `toml:"target_dir"`
	CacheDir     string              `toml:"cache_dir"`
}

type IntermediateConfig struct {
	gobaselogger.Config
	FileTypes map[string][]string `toml:"file_types"` // Ensure TOML tag matches the file
	CacheDir  string              `toml:"cache_dir"`
}

func CreateDirIfNotExist(path string) {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", filepath.Dir(path), err))
			errMsg := fmt.Sprintf("Error creating directory at %s", filepath.Dir(path))
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}
	}
}

func NewIntermediateConfig(optionalPath string) *IntermediateConfig {
	var configPath string

	// Step 1: Determine the configuration file path
	if optionalPath != "" {
		configPath = optionalPath
	}

	if _, err := os.Stat(configPath); err == nil {
		slog.Warn(fmt.Sprintf("Optional path provided: %s\n", optionalPath))
	} else if os.Stat(internal.DefaultWorkspaceConfigFile); err == nil {
		slog.Warn(fmt.Sprintf("Config file found: %s\n", internal.DefaultWorkspaceConfigFile))
		configPath = optionalPath
	} else {
		slog.Warn(fmt.Sprintf("Config file found: %s\n", internal.DefaultGlobalConfigFile))
		configPath = internal.DefaultGlobalConfigFile
	}

	slog.Info(fmt.Sprintf("Config path: %s\n", configPath))

	var defaultConfig IntermediateConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig = getDefaultConfig()
		slog.Info(fmt.Sprintf("\nPath %s: %v", filepath.Dir(configPath), err))
		CreateDirIfNotExist(filepath.Dir(configPath))
		file, err := os.Create(configPath)
		if err != nil {
			slog.Error(fmt.Sprintf("Error creating default config file: %v", err))
			return nil
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		if err := encoder.Encode(defaultConfig); err != nil {
			slog.Error(fmt.Sprintf("Error writing default config file: %v", err))
			return nil
		}
		slog.Info(fmt.Sprintf("Default config file created at %s", configPath))
	} else {
		// Step 3: Decode the existing config file
		slog.Info(fmt.Sprintf("Loading config file from %s", configPath))
		var tempConfig map[string]interface{}
		if _, err := toml.DecodeFile(configPath, &tempConfig); err != nil {
			slog.Error(fmt.Sprintf("Error decoding config file: %v", err))
			return nil
		}

		slog.Debug(fmt.Sprintf("TempConfig (raw): %+v\n", tempConfig))

		// Decode configuration file into IntermediateConfig
		if _, err := toml.DecodeFile(configPath, &defaultConfig); err != nil {
			slog.Error(fmt.Sprintf("Error decoding config file to struct: %v", err))
			return nil
		}
	}

	// Step 4: Confirm loaded config (case-sensitive)
	slog.Debug(fmt.Sprintf("Loaded file_types (case-sensitive): %+v\n", defaultConfig.FileTypes))

	return &defaultConfig
}

func NewDeskFSConfig() *DeskFSConfig {
	return &DeskFSConfig{
		FileTypeTree: trees.NewFileTypeTree(),
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
		},
		CacheDir: internal.DefaultCacheDir,
	}
}
