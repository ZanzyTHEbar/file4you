package cli_util

import (
	"errors"
	"fmt"

	"file4you/internal/cli"

	"github.com/spf13/cobra"
)

// NewClearCmd creates a new cobra command for clear operations.
func NewClearCmd(params *cli.CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear [target]",
		Short: "Clear cache, workspace data, or all data.",
		Long: `Clear operations for file4you.
Targets can be:
  all          - Clear all data including the central database and cache.
  cache        - Clear the application cache.
  database     - Clear the central database.
  workspace [id] - Clear a specific workspace by its ID.
  workspaces   - Clear all workspaces.
`,
		Args: cobra.MinimumNArgs(1), // Expects at least one argument: the target
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			var workspaceID string
			if target == "workspace" {
				if len(args) < 2 {
					err := errors.New("workspace ID is required when target is 'workspace'")
					params.Interactor.Error("Invalid arguments for clear 'workspace'", err)
					return err
				}
				workspaceID = args[1]
			}

			confirmMsg := fmt.Sprintf("Are you sure you want to clear '%s'? This action is irreversible.", target)
			if target == "all" || target == "database" {
				confirmMsg = fmt.Sprintf("DANGER ZONE: Are you absolutely sure you want to clear '%s'? This will delete significant data and is irreversible.", target)
			}

			confirmed, err := params.Interactor.Confirm(confirmMsg, false)
			if err != nil {
				params.Interactor.Error("Confirmation failed", err)
				return err
			}

			if !confirmed {
				params.Interactor.Info("Clear operation cancelled by user.")
				return nil
			}

			params.Interactor.Output(fmt.Sprintf("Starting clear operation for target: %s...", target))
			if workspaceID != "" {
				params.Interactor.Output(fmt.Sprintf("Workspace ID: %s", workspaceID))
			}
			params.Interactor.StartSpinner("Performing clear operation...")

			// Placeholder for actual clear logic
			// This logic will eventually be a Genkit flow/tool.
			// For example: err := clearService.Clear(target, workspaceID, params.Interactor)

			// Simulate work
			// time.Sleep(3 * time.Second) // Example placeholder for actual work
			clearSuccessful := true // Placeholder
			var clearErr error = nil   // Placeholder

			if clearErr != nil {
				params.Interactor.StopSpinner(false, "Clear operation failed.")
				params.Interactor.Error(fmt.Sprintf("Failed to clear '%s'", target), clearErr)
				return clearErr
			}

			if clearSuccessful {
				params.Interactor.StopSpinner(true, "Clear operation completed successfully.")
				params.Interactor.Success(fmt.Sprintf("Target '%s' cleared.", target))
			} else {
				params.Interactor.StopSpinner(false, "Clear operation did not complete as expected.")
				params.Interactor.Warning(fmt.Sprintf("Clear operation for '%s' may not be complete.", target))
			}
			return nil
		},
	}
	// Add flags if needed, e.g., --force (though confirmation is already built-in)
	return cmd
}
