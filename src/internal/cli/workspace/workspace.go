package workspace

import (
	"file4you/internal/cli"
	"fmt"
	"os"

	"github.com/google/uuid" // Import for UUID parsing
	"github.com/spf13/cobra"
)

type WorkspaceCMD struct {
	Workspace *cobra.Command
}

func NewWorkspace(params *cli.CmdParams) *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Manage workspaces",
		Long:    `Manage workspaces including creating, updating, and deleting workspaces.`,
		// No RunE needed for a command group, subcommands will have it
	}

	// Subcommand: create
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		Long:  `Create a new workspace with the specified root path and configuration. IF root-path is not provided, the current working directory is used.`,
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			rootPath, _ := cmd.Flags().GetString("root-path")
			config, _ := cmd.Flags().GetString("config")

			if rootPath == "" {
				params.Interactor.Warning("root-path not provided, using current working directory.")
				var err error
				rootPath, err = os.Getwd()
				if err != nil {
					params.Interactor.Error("Error getting current working directory", err)
					return err
				}
				params.Interactor.Info(fmt.Sprintf("Using current directory: %s", rootPath))
			}

			workspaceID, err := params.Filesystem.GetWorkspaceManager().CreateWorkspace(rootPath, config)
			if err != nil {
				params.Interactor.Error("Error creating workspace", err)
				return err
			}
			params.Interactor.Success(fmt.Sprintf("Workspace created successfully with ID: %s", workspaceID))
			return nil
		},
	}
	createCmd.Flags().String("root-path", "", "Root path for the workspace") // Removed (required) as it defaults
	createCmd.Flags().String("config", "", "Configuration data for the workspace")

	// Subcommand: update
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing workspace",
		Long:  `Update the configuration for an existing workspace by ID.`,
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			idStr, _ := cmd.Flags().GetString("id")
			config, _ := cmd.Flags().GetString("config")

			// Manual check for idStr == "" removed as MarkFlagRequired will handle it.

			workspaceUUID, err := uuid.Parse(idStr)
			if err != nil {
				params.Interactor.Error(fmt.Sprintf("Invalid Workspace ID format: '%s'", idStr), err)
				return err
			}

			err = params.Filesystem.GetWorkspaceManager().UpdateWorkspace(workspaceUUID, config)
			if err != nil {
				params.Interactor.Error(fmt.Sprintf("Error updating workspace with ID %s", idStr), err)
				return err
			}
			params.Interactor.Success(fmt.Sprintf("Workspace with ID %s updated successfully", idStr))
			return nil
		},
	}
	updateCmd.Flags().String("id", "", "ID of the workspace to update")
	_ = updateCmd.MarkFlagRequired("id")
	updateCmd.Flags().String("config", "", "New configuration data for the workspace")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		Long:  `List all workspaces with their IDs and root paths.`,
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			workspaces, err := params.Filesystem.GetWorkspaceManager().ListWorkspaces()
			if err != nil {
				params.Interactor.Error("Error listing workspaces", err)
				return err
			}
			if len(workspaces) == 0 {
				params.Interactor.Info("No workspaces found.")
				return nil
			}
			params.Interactor.Output("Workspaces:")
			for _, ws := range workspaces {
				// Assuming ws.ID is a string. If it's an int, adjust formatting.
				params.Interactor.Outputf("  ID: %s, Root Path: %s", ws.ID.String(), ws.RootPath)
			}
			return nil
		},
	}

	// Subcommand: delete
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workspace",
		Long:  `Delete an existing workspace by its ID.`,
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			idStr, _ := cmd.Flags().GetString("id")

			// Manual check for idStr == "" removed as MarkFlagRequired will handle it.

			workspaceUUID, err := uuid.Parse(idStr)
			if err != nil {
				params.Interactor.Error(fmt.Sprintf("Invalid Workspace ID format: '%s'", idStr), err)
				return err
			}

			confirmMsg := fmt.Sprintf("Are you sure you want to delete workspace '%s'? This action cannot be undone.", idStr)
			confirmed, err := params.Interactor.Confirm(confirmMsg, false)
			if err != nil {
				params.Interactor.Error("Confirmation failed", err)
				return err
			}

			if !confirmed {
				params.Interactor.Info("Delete operation cancelled by user.")
				return nil
			}

			params.Interactor.StartSpinner(fmt.Sprintf("Deleting workspace %s...", idStr))
			err = params.Filesystem.GetWorkspaceManager().DeleteWorkspace(workspaceUUID)
			if err != nil {
				params.Interactor.StopSpinner(false, "Deletion failed.")
				params.Interactor.Error(fmt.Sprintf("Error deleting workspace with ID %s", idStr), err)
				return err
			}
			params.Interactor.StopSpinner(true, "Deletion successful.")
			params.Interactor.Success(fmt.Sprintf("Workspace with ID %s deleted successfully", idStr))
			return nil
		},
	}
	deleteCmd.Flags().String("id", "", "ID of the workspace to delete")
	_ = deleteCmd.MarkFlagRequired("id")

	// Add subcommands to the workspace command
	workspaceCmd.AddCommand(createCmd, updateCmd, deleteCmd, listCmd)
	return workspaceCmd
}
