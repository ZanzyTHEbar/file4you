package fs

import (
	"file4you/internal/cli"
	"file4you/internal/filesystem/options"
	"os"

	"github.com/spf13/cobra"
)

type OrganizeCMD struct {
	Organize *cobra.Command
}

var fileParams *options.FilePathParams = options.NewFilePathParams()

func NewOrganize(params *cli.CmdParams) *cobra.Command {
	organizeCmd := &cobra.Command{
		Use:     "organize",
		Aliases: []string{"o"},
		Short:   "Organize files in the specified directory, based on the configuration",
		Long:    `Organize files based on the configuration. Optionally specify a destination directory. If not provided, the current working directory is used.`,
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			if err := organizeFiles(params); err != nil {
				// Error is handled within organizeFiles using params.Interactor
				return err // Return error for Cobra to handle
			}
			return nil
		},
	}

	// Define flags and configuration settings
	organizeCmd.Flags().BoolVar(&fileParams.RemoveAfter, "remove", false, "Remove files after organizing")
	organizeCmd.Flags().BoolVarP(&fileParams.Recursive, "recursive", "r", false, "Recursively organize files")
	organizeCmd.Flags().BoolVarP(&fileParams.DryRun, "dryrun", "n", false, "Dry run to simulate organization")
	organizeCmd.Flags().IntVarP(&fileParams.MaxDepth, "max-depth", "x", -1, "Maximum depth for recursion")
	organizeCmd.Flags().BoolVarP(&fileParams.GitEnabled, "git-enabled", "g", false, "Enable Git operations")
	organizeCmd.Flags().BoolVarP(&fileParams.CopyFiles, "copy", "c", false, "Enable move as Copy operation, required when moving files across partitions. If not enabled, will default to copy when move is not possible.")
	organizeCmd.Flags().StringVarP(&fileParams.SourceDir, "srcDir", "d", "", "Destination directory to organize files from")
	organizeCmd.Flags().StringVarP(&fileParams.TargetDir, "target", "t", "", "Target directory to organize files into")

	return organizeCmd
}

func organizeFiles(params *cli.CmdParams) error {
	// Set default directories if not provided
	if fileParams.SourceDir == "" {
		var err error
		fileParams.SourceDir, err = os.Getwd()
		if err != nil {
			params.Interactor.Error("Error getting current working directory", err)
			return err
		}
	}

	if fileParams.TargetDir == "" {
		fileParams.TargetDir = fileParams.SourceDir
	}

	params.Interactor.StartSpinner("Organizing files...")

	// Initialize Git if Git is enabled and repository is not already initialized
	if fileParams.GitEnabled {
		if !params.Filesystem.IsGitRepo(fileParams.SourceDir) {
			params.Interactor.Info("Git operations enabled, but no Git repository detected. Initializing Git repository.")
			if err := params.Filesystem.InitGitRepo(fileParams.SourceDir); err != nil {
				params.Interactor.StopSpinner(false, "Git initialization failed.")
				params.Interactor.Error("Error initializing Git repository", err)
				return err
			}
			params.Interactor.Info("Git repository initialized successfully.")
		} else {
			params.Interactor.Info("Git repository detected.")
		}
	} else {
		params.Interactor.Warning("Git operations disabled. Proceeding without Git.")
	}

	// Execute the organization logic with EnhancedOrganize
	if err := params.Filesystem.EnhancedOrganize(params.Filesystem.InstanceConfig(), fileParams); err != nil {
		params.Interactor.StopSpinner(false, "Organization failed.")
		params.Interactor.Error("Error organizing files", err)
		return err
	}

	params.Interactor.StopSpinner(true, "Files organized successfully.")
	params.Interactor.Success("Organization complete.") // More concise success message

	return nil
}
