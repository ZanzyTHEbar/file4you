package workspace

import (
	"desktop-cleaner/internal/cli"
	"fmt"
	"os"

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
	}

	// Subcommand: create
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		Long:  `Create a new workspace with the specified root path and configuration. IF root-path is not provided, the current working directory is used.`,
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, _ := cmd.Flags().GetString("root-path")
			config, _ := cmd.Flags().GetString("config")

			if rootPath == "" {
				params.Term.OutputWarning("Warn: root-path is required, using $(pwd)")
				var err error
				rootPath, err = os.Getwd()
				if err != nil {
					params.Term.OutputErrorAndExit("Error getting current working directory: %v", err)
				}
			}

			workspaceID, err := params.DeskFS.WorkspaceManager.CreateWorkspace(rootPath, config)
			if err != nil {
				params.Term.OutputErrorAndExit("Error creating workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace created successfully with ID: %s", workspaceID))
		},
	}
	createCmd.Flags().String("root-path", "", "Root path for the workspace (required)")
	createCmd.Flags().String("config", "", "Configuration data for the workspace")

	// Subcommand: update
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing workspace",
		Long:  `Update the configuration for an existing workspace by ID.`,
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt("id")
			config, _ := cmd.Flags().GetString("config")

			// List the workspaces, and then update the workspace with the given ID
			workspaces, err := params.DeskFS.WorkspaceManager.ListWorkspaces()
			if err != nil {
				params.Term.OutputErrorAndExit("Error listing workspaces: %v", err)
			}

			// workspaces are already ordered by timestamp, so we can use the index as the ID
			if id <= 0 {
				params.Term.OutputErrorAndExit("Error: valid workspace ID is required")
			}
			// Subtract 1 from the ID to get the index
			id--

			err = params.DeskFS.WorkspaceManager.UpdateWorkspace(workspaces[id].ID, config)
			if err != nil {
				params.Term.OutputErrorAndExit("Error updating workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace with ID %d updated successfully", id))
		},
	}
	updateCmd.Flags().Int("id", 0, "ID of the workspace to update (required)")
	updateCmd.Flags().String("config", "", "New configuration data for the workspace")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		Long:  `List all workspaces with their IDs and root paths.`,
		Run: func(cmd *cobra.Command, args []string) {
			workspaces, err := params.DeskFS.WorkspaceManager.ListWorkspaces()
			if err != nil {
				params.Term.OutputErrorAndExit("Error listing workspaces: %v", err)
			}
			params.Term.OutputSuccess("Workspaces:")
			for _, ws := range workspaces {
				params.Term.OutputInfo(fmt.Sprintf("ID: %s, Root Path: %s", ws.ID, ws.RootPath))
			}
		},
	}

	// Subcommand: delete
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workspace",
		Long:  `Delete an existing workspace by its ID.`,
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetInt("id")

			// List the workspaces, and then update the workspace with the given ID
			workspaces, err := params.DeskFS.WorkspaceManager.ListWorkspaces()
			if err != nil {
				params.Term.OutputErrorAndExit("Error listing workspaces: %v", err)
			}

			// workspaces are already ordered by timestamp, so we can use the index as the ID
			if id <= 0 {
				params.Term.OutputErrorAndExit("Error: valid workspace ID is required")
			}
			// Subtract 1 from the ID to get the index
			id--
			err = params.DeskFS.WorkspaceManager.DeleteWorkspace(workspaces[id].ID)
			if err != nil {
				params.Term.OutputErrorAndExit("Error deleting workspace: %v", err)
			}
			params.Term.OutputSuccess(fmt.Sprintf("Workspace with ID %d deleted successfully", id))
		},
	}
	deleteCmd.Flags().Int("id", 0, "ID of the workspace to delete (required)")

	// Add subcommands to the workspace command
	workspaceCmd.AddCommand(createCmd, updateCmd, deleteCmd, listCmd)
	return workspaceCmd
}
