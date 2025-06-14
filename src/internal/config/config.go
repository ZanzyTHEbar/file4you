package config

import (
	"file4you/internal"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	Genkit   GenkitConfig   `mapstructure:"genkit"`
	File4You File4YouConfig `mapstructure:"file4you"`
}

// GenkitConfig stores Genkit related configurations.
type GenkitConfig struct {
	Plugins GenkitPluginsConfig `mapstructure:"plugins"`
	Prompts GenkitPromptsConfig `mapstructure:"prompts"`
}

// GenkitPluginsConfig stores plugin configurations.
type GenkitPluginsConfig struct {
	GoogleAI GenkitPlugin `mapstructure:"googleAI"`
	OpenAI   GenkitPlugin `mapstructure:"openAI"`
}

// GenkitPlugin stores common configuration for a Genkit plugin.
type GenkitPlugin struct {
	APIKey         string `mapstructure:"apiKey"`
	DefaultModel   string `mapstructure:"defaultModel"`
	TimeoutSeconds int    `mapstructure:"timeoutSeconds"`
}

// GenkitPromptsConfig stores prompts configurations.
type GenkitPromptsConfig struct {
	Directory string `mapstructure:"directory"`
}

// DatabaseConfig stores database connection details.
type DatabaseConfig struct {
	DSN  string `mapstructure:"dsn"`
	Type string `mapstructure:"type"`
}

// File4YouConfig stores file4you specific configurations.
type File4YouConfig struct {
	GenkitHandler          File4YouGenkitHandlerConfig `mapstructure:"genkithandler"`
	TargetDir              string                      `mapstructure:"targetDir"`
	CacheDir               string                      `mapstructure:"cacheDir"`
	Database               DatabaseConfig              `mapstructure:"database"`
	OrganizeTimeoutMinutes int                         `mapstructure:"organizeTimeoutMinutes"`
}

// File4YouGenkitHandlerConfig stores genkithandler specific feature flags.
type File4YouGenkitHandlerConfig struct {
	FeatureFlags map[string]bool `mapstructure:"featureFlags"`
}

var AppConfig Config

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		viper.AddConfigPath(filepath.Join("etc", internal.DefaultAppName))
		viper.AddConfigPath(internal.DefaultConfigPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Set default values
	viper.SetDefault("genkit.prompts.directory", "./prompts")
	viper.SetDefault("genkit.plugins.openai.timeoutSeconds", 60)
	// Add defaults for all known feature flags.
	// viper.SetDefault("file4you.genkithandler.featureFlags.someNewFeature", false)

	// Defaults from the old deskfs.Config / new consolidated fields
	// Ensure internal.DefaultCacheDir, internal.DefaultDatabaseDSN, internal.DefaultDatabaseType are defined.
	viper.SetDefault("file4you.targetDir", ".")
	// Assuming internal.DefaultCacheDir is defined, e.g., in internal/globals.go
	viper.SetDefault("file4you.cacheDir", internal.DefaultCacheDir)
	// Assuming internal.DefaultDatabaseDSN and internal.DefaultDatabaseType are defined
	viper.SetDefault("file4you.database.dsn", internal.DefaultDatabaseDSN)
	viper.SetDefault("file4you.database.type", internal.DefaultDatabaseType)
	viper.SetDefault("file4you.organizeTimeoutMinutes", 10)

	viper.AutomaticEnv()                                   // Read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores in env var names e.g. genkit.plugins.googleAI.apiKey becomes GENKIT_PLUGINS_GOOGLEAI_APIKEY

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; defaults will be used. This is not an error for the application to halt on.
			// It's good practice to log this situation if a logger is available here.
			// fmt.Printf("Warning: Config file not found at expected locations. Using default values. Searched: %s\n", viper.ConfigFileUsed())
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &AppConfig, nil
}
