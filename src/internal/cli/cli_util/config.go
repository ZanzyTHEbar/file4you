package cli_util

import (
	"file4you/internal/cli"
	"file4you/internal/config"
	"fmt"

	"github.com/spf13/cobra"
)

// NewConfig creates a new config command for configuration management
func NewConfig(params *cli.CmdParams) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage File4You configuration",
		Long: `Configuration management for File4You including:
- View current configuration
- Set configuration values
- Reset to defaults
- Show configuration file location`,
	}

	// Add subcommands
	configCmd.AddCommand(newConfigShow(params))
	configCmd.AddCommand(newConfigSet(params))
	configCmd.AddCommand(newConfigGet(params))
	configCmd.AddCommand(newConfigReset(params))
	configCmd.AddCommand(newConfigPath(params))

	return configCmd
}

// newConfigShow shows current configuration
func newConfigShow(params *cli.CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			params.Interactor.Output("Current Configuration:")
			params.Interactor.Output("=====================")

			// Display configuration sections
			displayDatabaseConfig(params, &cfg.File4You)
			displayGenkitConfig(params, &cfg.Genkit)
			displayGeneralConfig(params, &cfg.File4You)

			return nil
		},
	}
}

// newConfigGet gets a specific configuration value
func newConfigGet(params *cli.CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			value := getConfigValue(&cfg.File4You, key)
			if value == "" {
				params.Interactor.Output(fmt.Sprintf("Configuration key '%s' not found", key))
				return nil
			}

			params.Interactor.Output(fmt.Sprintf("%s = %s", key, value))
			return nil
		},
	}
}

// newConfigSet sets a configuration value
func newConfigSet(params *cli.CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// TODO: Implement configuration setting
			params.Interactor.Info("Configuration setting will be implemented")
			params.Interactor.Output(fmt.Sprintf("Would set: %s = %s", key, value))

			return nil
		},
	}
}

// newConfigReset resets configuration to defaults
func newConfigReset(params *cli.CmdParams) *cobra.Command {
	var confirm bool

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				params.Interactor.Output("This will reset all configuration to defaults. Use --confirm to proceed.")
				return nil
			}

			// TODO: Implement configuration reset
			params.Interactor.Info("Configuration reset will be implemented")
			params.Interactor.Success("Configuration would be reset to defaults")

			return nil
		},
	}

	resetCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm the reset operation")
	return resetCmd
}

// newConfigPath shows configuration file path
func newConfigPath(params *cli.CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Get actual config file path
			params.Interactor.Output("Configuration file location:")
			params.Interactor.Info("~/.config/file4you/config.toml")

			return nil
		},
	}
}

// displayDatabaseConfig shows database configuration
func displayDatabaseConfig(params *cli.CmdParams, cfg *config.File4YouConfig) {
	params.Interactor.Output("\nDatabase:")
	params.Interactor.Info(fmt.Sprintf("  Type: %s", cfg.Database.Type))
	params.Interactor.Info(fmt.Sprintf("  DSN: %s", cfg.Database.DSN))
}

// displayGenkitConfig shows Genkit configuration
func displayGenkitConfig(params *cli.CmdParams, cfg *config.GenkitConfig) {
	params.Interactor.Output("\nGenkit:")
	params.Interactor.Info(fmt.Sprintf("  Prompts Directory: %s", cfg.Prompts.Directory))
	params.Interactor.Info(fmt.Sprintf("  OpenAI Model: %s", cfg.Plugins.OpenAI.DefaultModel))
	params.Interactor.Info(fmt.Sprintf("  GoogleAI Model: %s", cfg.Plugins.GoogleAI.DefaultModel))
}

// displayGeneralConfig shows general configuration
func displayGeneralConfig(params *cli.CmdParams, cfg *config.File4YouConfig) {
	params.Interactor.Output("\nGeneral:")
	params.Interactor.Info(fmt.Sprintf("  Target Directory: %s", cfg.TargetDir))
	params.Interactor.Info(fmt.Sprintf("  Cache Directory: %s", cfg.CacheDir))
	params.Interactor.Info(fmt.Sprintf("  Organize Timeout: %d minutes", cfg.OrganizeTimeoutMinutes))
}

// getConfigValue retrieves a configuration value by key
func getConfigValue(cfg *config.File4YouConfig, key string) string {
	switch key {
	case "target_dir":
		return cfg.TargetDir
	case "cache_dir":
		return cfg.CacheDir
	case "organize_timeout":
		return fmt.Sprintf("%d", cfg.OrganizeTimeoutMinutes)
	case "database.type":
		return cfg.Database.Type
	case "database.dsn":
		return cfg.Database.DSN
	}
	return ""
}
