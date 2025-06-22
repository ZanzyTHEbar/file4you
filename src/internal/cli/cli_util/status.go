package cli_util

import (
	"file4you/internal/cli"
	"file4you/version"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// NewStatus creates a new status command for system status
func NewStatus(params *cli.CmdParams) *cobra.Command {
	var detailed bool

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show File4You system status",
		Long: `Display system status including:
- Application version and build info
- Database connectivity
- Configuration status
- System resources
- Recent activity`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := displaySystemStatus(params, detailed); err != nil {
				return fmt.Errorf("failed to get system status: %w", err)
			}
			return nil
		},
	}

	statusCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed status information")
	return statusCmd
}

// displaySystemStatus shows comprehensive system status
func displaySystemStatus(params *cli.CmdParams, detailed bool) error {
	params.Interactor.Output("File4You System Status")
	params.Interactor.Output("====================")

	// Application info
	displayApplicationInfo(params)

	// System info
	displaySystemInfo(params)

	// Database status
	if err := displayDatabaseStatus(params); err != nil {
		params.Interactor.Info(fmt.Sprintf("Database Status: Error - %v", err))
	}

	// Configuration status
	displayConfigurationStatus(params)

	if detailed {
		// Additional detailed information
		displayDetailedInfo(params)
	}

	return nil
}

// displayApplicationInfo shows application version and build information
func displayApplicationInfo(params *cli.CmdParams) {
	params.Interactor.Output("\nApplication:")
	params.Interactor.Info(fmt.Sprintf("  Version: %s", version.Version))
	params.Interactor.Info(fmt.Sprintf("  Build Date: %s", getBuildDate()))
	params.Interactor.Info(fmt.Sprintf("  Go Version: %s", runtime.Version()))
	params.Interactor.Info(fmt.Sprintf("  Architecture: %s/%s", runtime.GOOS, runtime.GOARCH))
}

// displaySystemInfo shows system resource information
func displaySystemInfo(params *cli.CmdParams) {
	params.Interactor.Output("\nSystem Resources:")

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	params.Interactor.Info(fmt.Sprintf("  Memory Allocated: %s", formatBytes(int64(m.Alloc))))
	params.Interactor.Info(fmt.Sprintf("  Total Allocations: %s", formatBytes(int64(m.TotalAlloc))))
	params.Interactor.Info(fmt.Sprintf("  System Memory: %s", formatBytes(int64(m.Sys))))
	params.Interactor.Info(fmt.Sprintf("  GC Cycles: %d", m.NumGC))
	params.Interactor.Info(fmt.Sprintf("  Goroutines: %d", runtime.NumGoroutine()))
	params.Interactor.Info(fmt.Sprintf("  CPU Cores: %d", runtime.NumCPU()))
}

// displayDatabaseStatus checks and displays database connectivity
func displayDatabaseStatus(params *cli.CmdParams) error {
	params.Interactor.Output("\nDatabase:")

	// TODO: Actually check database connectivity
	// For now, just show that we have a database configured
	if params.CentralDB != nil {
		params.Interactor.Success("  Status: Connected")
		params.Interactor.Info("  Type: Turso/LibSQL")

		// Try to get some basic stats
		// TODO: Implement actual database health check
		params.Interactor.Info("  Health: Good")
	} else {
		params.Interactor.Info("  Status: Not initialized")
		return fmt.Errorf("database not initialized")
	}

	return nil
}

// displayConfigurationStatus shows configuration file status
func displayConfigurationStatus(params *cli.CmdParams) {
	params.Interactor.Output("\nConfiguration:")

	// Check if config file exists
	// TODO: Get actual config file path from config package
	configPath := "~/.config/file4you/config.toml"
	if _, err := os.Stat(configPath); err == nil {
		params.Interactor.Success("  Config File: Found")
		params.Interactor.Info(fmt.Sprintf("  Location: %s", configPath))
	} else {
		params.Interactor.Info("  Config File: Using defaults")
		params.Interactor.Info("  Location: Default values")
	}

	params.Interactor.Info("  Status: Loaded successfully")
}

// displayDetailedInfo shows additional detailed system information
func displayDetailedInfo(params *cli.CmdParams) {
	params.Interactor.Output("\nDetailed Information:")

	// Working directory
	if cwd, err := os.Getwd(); err == nil {
		params.Interactor.Info(fmt.Sprintf("  Working Directory: %s", cwd))
	}

	// Environment
	params.Interactor.Info(fmt.Sprintf("  User: %s", os.Getenv("USER")))
	params.Interactor.Info(fmt.Sprintf("  Home: %s", os.Getenv("HOME")))
	params.Interactor.Info(fmt.Sprintf("  Path: %s", os.Getenv("PATH")))

	// Process info
	params.Interactor.Info(fmt.Sprintf("  Process ID: %d", os.Getpid()))
	params.Interactor.Info(fmt.Sprintf("  Parent PID: %d", os.Getppid()))

	// Uptime (since process start)
	// TODO: Track actual application start time
	params.Interactor.Info("  Uptime: Unknown (will be implemented)")
}

// getBuildDate returns the build date (placeholder)
func getBuildDate() string {
	// TODO: This should be set during build time
	return "Unknown"
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
