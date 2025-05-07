package genkithandler

import (
	"context"
	"fmt"

	"file4you/internal/db"
	"file4you/internal/deskfs"

	"github.com/firebase/genkit/go/genkit"
	// We can use genkit.Logger(ctx) for logging within tools if needed.
)

// BackupToolInput defines the input for the backup tool.
// For now, it's empty, implying a full backup. It can be extended with options.
type BackupToolInput struct {
	// Example: BackupTarget string (e.g., "deskfs_only", "centraldb_only", "all")
}

// BackupToolOutput defines the output structure for the backup tool.
// This is what an LLM would receive as a structured response if it invoked this tool.
type BackupToolOutput struct {
	DeskFSBackupPath    string `json:"deskFSBackupPath,omitempty"`
	CentralDBBackupPath string `json:"centralDBBackupPath,omitempty"`
	SuccessMessage      string `json:"successMessage"`
}

// RegisterBackupTool defines and registers the 'performBackup' tool with the provided Genkit instance.
// This tool encapsulates the core logic for backing up application data from DeskFS and CentralDB.
// It is designed to be callable by an LLM as part of a Genkit flow.
func RegisterBackupTool(g *genkit.Genkit, dfs *deskfs.DeskFS, cdb *db.CentralDBProvider) {
	genkit.DefineToolWith(
		g,               // The specific Genkit instance to associate this tool with.
		"performBackup", // The unique name identifying this tool.
		"Performs a full backup of application data, including DeskFS and CentralDB. "+ // Description for LLM.
			"Returns paths to the created backup files and a success message.",
		// The function implementing the tool's logic.
		func(ctx context.Context, input BackupToolInput) (BackupToolOutput, error) {
			// genkit.Logger(ctx).Info("performBackup tool execution started", "input", input)
			var output BackupToolOutput
			var err error

			// Perform DeskFS backup.
			output.DeskFSBackupPath, err = dfs.Backup()
			if err != nil {
				// genkit.Logger(ctx).Error("performBackup tool: DeskFS backup failed", "error", err)
				return BackupToolOutput{}, fmt.Errorf("DeskFS backup failed: %w", err)
			}

			// Perform CentralDB backup.
			output.CentralDBBackupPath, err = cdb.Backup()
			if err != nil {
				// genkit.Logger(ctx).Error("performBackup tool: CentralDB backup failed", "error", err, "deskFSBackupPath", output.DeskFSBackupPath)
				// Decide on error handling: e.g., should we attempt to roll back DeskFS backup?
				// For now, report that DeskFS was backed up but CentralDB failed.
				return BackupToolOutput{}, fmt.Errorf("CentralDB backup failed (DeskFS backup completed at %s): %w", output.DeskFSBackupPath, err)
			}

			output.SuccessMessage = fmt.Sprintf("Backup tool executed successfully. DeskFS backup at '%s', CentralDB backup at '%s'.", output.DeskFSBackupPath, output.CentralDBBackupPath)
			// genkit.Logger(ctx).Info("performBackup tool execution finished successfully", "output", output)
			return output, nil
		},
	)
}

// TODO: Define and register other tools here as functionalities are migrated/exposed to Genkit.
// Examples:
// - func RegisterFileOrganizationTool(g *genkit.Genkit, dfs *deskfs.DeskFS, interactor ui.Interactor)
// - func RegisterGitRewindTool(g *genkit.Genkit, projectPath string, interactor ui.Interactor)
// - func RegisterWorkspaceManagementTool(g *genkit.Genkit, ...)
