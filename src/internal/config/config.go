package config

import (
	"fmt"
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
Plugins   GenkitPluginsConfig   `mapstructure:"plugins"`
Prompts   GenkitPromptsConfig   `mapstructure:"prompts"`
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

// File4YouConfig stores file4you specific configurations.
type File4YouConfig struct {
	GenkitHandler File4YouGenkitHandlerConfig `mapstructure:"genkithandler"`
}

// File4YouGenkitHandlerConfig stores genkithandler specific feature flags.
type File4YouGenkitHandlerConfig struct {
	FeatureFlags map[string]bool `mapstructure:"featureFlags"`
}

var AppConfig Config

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath) // Path to look for the config file in
	} else {
		viper.AddConfigPath(".")               // Look for config in current directory
		viper.AddConfigPath("..")              // Look for config in parent directory (project root if running from src)
		viper.AddConfigPath("/etc/file4you/")  // Path to look for the config file in
		viper.AddConfigPath("$HOME/.file4you") // Call multiple times to add many search paths
		viper.SetConfigName("config")          // Name of config file (without extension)
		viper.SetConfigType("yaml")            // REQUIRED if the config file does not have the extension in the name
}

// Set default values
viper.SetDefault("genkit.prompts.directory", "./prompts")
viper.SetDefault("genkit.plugins.openai.timeoutSeconds", 60)
	// Example for a feature flag, assuming it might exist.
	// Add defaults for all known feature flags.
	viper.SetDefault("file4you.genkithandler.featureFlags.someNewFeature", false)

	viper.AutomaticEnv()                                   // Read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores in env var names e.g. genkit.plugins.googleAI.apiKey becomes GENKIT_PLUGINS_GOOGLEAI_APIKEY

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// return nil, fmt.Errorf("config file not found: %w", err)
			// For now, we'll allow the app to run with defaults or only env vars if config file is not present
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &AppConfig, nil
}
