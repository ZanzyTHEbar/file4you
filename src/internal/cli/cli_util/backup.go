package cli_util

import (
	"errors"
	"fmt"

	"file4you/internal/cli"

	"github.com/spf13/cobra"
)

// NewBackupCmd creates a new cobra command for backup operations.
func NewBackupCmd(params *cli.CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup [target]",
		Short: "Backup databases, workspaces, or all data.",
		Long: `Backup operations for file4you.
Targets can be:
  all          - Backup all data including the central database.
  database     - Backup only the central database.
  workspace [id] - Backup a specific workspace by its ID.
  workspaces   - Backup all workspaces.
`,
		Args: cobra.MinimumNArgs(1), // Expects at least one argument: the target
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			var workspaceID string
			if target == "workspace" {
				if len(args) < 2 {
					err := errors.New("workspace ID is required when target is 'workspace'")
					params.Interactor.Error("Invalid arguments for backup 'workspace'", err)
					return err
				}
				workspaceID = args[1]
			}

			confirmed, err := params.Interactor.Confirm(fmt.Sprintf("Are you sure you want to backup '%s'?", target), false)
			if err != nil {
				params.Interactor.Error("Confirmation failed", err)
				return err
			}

			if !confirmed {
				params.Interactor.Info("Backup operation cancelled by user.")
				return nil
			}

			params.Interactor.Output(fmt.Sprintf("Starting backup for target: %s...", target))
			if workspaceID != "" {
				params.Interactor.Output(fmt.Sprintf("Workspace ID: %s", workspaceID))
			}
			params.Interactor.StartSpinner("Performing backup...")

			// Placeholder for actual backup logic
			// This logic will eventually be a Genkit flow/tool.
			// For example: err := backupService.Backup(target, workspaceID, params.Interactor)

			// Simulate work
			// time.Sleep(3 * time.Second) // Example placeholder for actual work
			backupSuccessful := true // Placeholder
			var backupErr error = nil    // Placeholder

			if backupErr != nil {
				params.Interactor.StopSpinner(false, "Backup failed.")
				params.Interactor.Error(fmt.Sprintf("Failed to backup '%s'", target), backupErr)
				return backupErr
			}

			if backupSuccessful {
				params.Interactor.StopSpinner(true, "Backup completed successfully.")
				params.Interactor.Success(fmt.Sprintf("Target '%s' backed up.", target))
			} else {
				// This case might not be reached if backupErr is always set on failure
				params.Interactor.StopSpinner(false, "Backup did not complete as expected.")
				params.Interactor.Warning(fmt.Sprintf("Backup for '%s' may not be complete.", target))
			}
			return nil
		},
	}
	// Add flags if needed, e.g., --output-directory, --compression-level
	return cmd
}