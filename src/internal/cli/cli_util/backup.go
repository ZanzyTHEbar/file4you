package cli_util

import (
	"context" // Added for Genkit flow execution
	"errors"
	"fmt"

	"file4you/internal/cli"
	"file4you/internal/genkithandler" // Added for Genkit flows and types

	"github.com/spf13/cobra"
)

// NewBackupCmd creates a new cobra command for backup operations.
func NewBackupCmd(params *cli.CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup [target]",
		Short: "Backup databases, workspaces, or all data using Genkit flow.",
		Long: `Backup operations for file4you, now orchestrated by Genkit.
Targets can be:
  all          - Backup all data including the central database. (Invokes backupFlow)
  database     - Backup only the central database. (TODO: Specific flow or tool option)
  workspace [id] - Backup a specific workspace by its ID. (TODO: Specific flow or tool option)
  workspaces   - Backup all workspaces. (TODO: Specific flow or tool option)
`,
		Args: cobra.MinimumNArgs(1), // Expects at least one argument: the target
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			// var workspaceID string // Currently, backupFlow doesn't take specific workspaceID yet
			// if target == "workspace" {
			// 	if len(args) < 2 {
			// 		err := errors.New("workspace ID is required when target is 'workspace'")
			// 		params.Interactor.Error("Invalid arguments for backup 'workspace'", err)
			// 		return err
			// 	}
			// 	workspaceID = args[1]
			// }

			// For now, only "all" target is directly mapped to the generic backupFlow
			if target != "all" {
				msg := fmt.Sprintf("Target '%s' is not yet fully implemented with Genkit flows. Only 'all' is currently supported via backupFlow.", target)
				params.Interactor.Warning(msg)
				// return errors.New(msg) // Or proceed with a non-Genkit path if available
				// For now, let's try to run the generic backup flow for any target to test it.
				// Later, specific inputs or flows will be needed.
				params.Interactor.Info(fmt.Sprintf("Proceeding with generic backupFlow for target '%s'. This might not be what you expect.", target))

			}

			confirmed, err := params.Interactor.Confirm(fmt.Sprintf("Are you sure you want to backup using Genkit flow for target '%s'?", target), false)
			if err != nil {
				params.Interactor.Error("Confirmation failed", err)
				return err
			}

			if !confirmed {
				params.Interactor.Info("Backup operation cancelled by user.")
				return nil
			}

			params.Interactor.Output(fmt.Sprintf("Starting Genkit backup flow for target: %s...", target))
			params.Interactor.StartSpinner("Performing backup via Genkit flow...")

			if params.Genkit == nil {
				initErr := errors.New("genkit not initialized; this should have happened in root command initialization")
				params.Interactor.StopSpinner(false, "Genkit initialization error.")
				params.Interactor.Error("Genkit initialization error", initErr)
				return initErr
			}

			flowRunnerInter := genkithandler.GetBackupFlow()
			if flowRunnerInter == nil {
				flowRetrievalErr := errors.New("BackupFlow runner not found; ensure it was registered")
				params.Interactor.StopSpinner(false, "Failed to retrieve backup flow runner.")
				params.Interactor.Error("Failed to retrieve flow runner", flowRetrievalErr)
				return flowRetrievalErr
			}

			// Type assert to the specific runner type defined in legacy.go
			type legacyFlowRunner interface {
				Run(ctx context.Context, input interface{}) (interface{}, error)
			}
			backupFlowRunner, ok := flowRunnerInter.(legacyFlowRunner)
			if !ok {
				typeErr := errors.New("retrieved flow runner is not of expected type (legacyFlowRunner)")
				params.Interactor.StopSpinner(false, "Type assertion failed for flow runner.")
				params.Interactor.Error("Type assertion failed", typeErr)
				return typeErr
			}

			// Prepare input for the backupFlow.
			// This needs to match the expected input type of the BackupFlow
			flowInput := genkithandler.BackupToolInput{
				// Populate fields as necessary, e.g., from command flags or config
				// For a generic "all" backup, these might be determined by the flow/tool itself.
				// SourcePath:      "TBD: SourcePath from config or flags",
				// DestinationPath: "TBD: DestinationPath from config or flags",
				// BackupType:      "full", // Example
			}
			// if target == "database" { flowInput.BackupTarget = "centraldb_only" } // Example for future extension

			response, err := backupFlowRunner.Run(context.Background(), flowInput)
			if err != nil {
				params.Interactor.StopSpinner(false, "Genkit backup flow failed.")
				params.Interactor.Error(fmt.Sprintf("Genkit backupFlow execution failed for target '%s'", target), err)
				return err
			}

			// Process the response
			backupOutput, ok := response.(genkithandler.BackupToolOutput)
			if !ok {
				typeErr := errors.New("BackupFlow response is not of expected type BackupToolOutput")
				params.Interactor.StopSpinner(false, "Flow response type error.")
				params.Interactor.Error("Flow response type error", typeErr)
				return typeErr
			}

			if backupOutput.Error() != "" {
				respErr := errors.New(backupOutput.Error())
				params.Interactor.StopSpinner(false, "BackupFlow reported an error.")
				params.Interactor.Error("BackupFlow reported an error", respErr)
				return respErr
			}

			params.Interactor.StopSpinner(true, "Genkit backup flow completed successfully.")
			params.Interactor.Success(fmt.Sprintf("Genkit backup flow for target '%s' finished.", target))
			params.Interactor.Output(fmt.Sprintf("Flow Result: %s", backupOutput.SuccessMessage))
			if backupOutput.DeskFSBackupPath != "" {
				params.Interactor.Output(fmt.Sprintf("DeskFS Backup Path: %s", backupOutput.DeskFSBackupPath))
			}
			if backupOutput.CentralDBBackupPath != "" {
				params.Interactor.Output(fmt.Sprintf("CentralDB Backup Path: %s", backupOutput.CentralDBBackupPath))
			}

			return nil
		},
	}
	// Add flags if needed, e.g., --output-directory, --compression-level
	return cmd
}
