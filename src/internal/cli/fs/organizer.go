package fs

import (
	"file4you/internal/cli"
	deskfs "file4you/internal/deskfs"
	"os"

	"github.com/spf13/cobra"
)

type OrganizeCMD struct {
	Organize *cobra.Command
}

var fileParams *deskfs.FilePathParams = deskfs.NewFilePathParams()

func NewOrganize(params *cli.CmdParams) *cobra.Command {
	organizeCmd := &cobra.Command{
		Use:     "organize",
		Aliases: []string{"o"},
		Short:   "Organize files in the specified directory, based on the configuration",
		Long:    `Organize files based on the configuration. Optionally specify a destination directory. If not provided, the current working directory is used.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := organizeFiles(params); err != nil {
				params.Term.OutputErrorAndExit("Error organizing files: %v", err)
			}
		},
	}

	// Define flags and configuration settings
	organizeCmd.Flags().BoolVar(&fileParams.RemoveAfter, "remove", false, "Remove files after organizing")
	organizeCmd.Flags().BoolVar(&fileParams.NamesOnly, "names-only", false, "Organize by names only")
	organizeCmd.Flags().BoolVar(&fileParams.ForceSkipIgnore, "force-skip-ignore", false, "Force skip ignored files")
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
			params.Term.OutputErrorAndExit("Error getting current working directory: %v", err)
		}
	}

	if fileParams.TargetDir == "" {
		fileParams.TargetDir = fileParams.SourceDir
	}

	params.Term.ToggleSpinner(true, "Organizing files...")

	// Initialize Git if Git is enabled and repository is not already initialized
	if fileParams.GitEnabled {
		if !params.DeskFS.IsGitRepo(fileParams.SourceDir) {
			params.Term.OutputInfo("Git operations enabled, but no Git repository detected. Initializing Git repository.")
			if err := params.DeskFS.InitGitRepo(fileParams.SourceDir); err != nil {
				params.Term.OutputErrorAndExit("Error initializing Git repository: %v", err)
			}
		} else {
			params.Term.OutputInfo("Git repository detected.")
		}
	} else {
		params.Term.OutputWarning("Git operations disabled. Proceeding without Git.")
	}

	// Execute the organization logic with EnhancedOrganize
	if err := params.DeskFS.EnhancedOrganize(params.DeskFS.InstanceConfig, fileParams); err != nil {
		params.Term.OutputErrorAndExit("Error organizing files: %v", err)
	}

	params.Term.ToggleSpinner(false, "")
	params.Term.OutputSuccess("Files organized successfully.")

	return nil
}
