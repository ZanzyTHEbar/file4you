package git

import (
	"file4you/internal/cli"
	"fmt"

	"github.com/spf13/cobra"
)

type RewindCMD struct {
	Rewind *cobra.Command
}

func NewRewind(params *cli.CmdParams) *cobra.Command {
	// rewindCmd represents the rewind command
	rewindCmd := &cobra.Command{
		Use:     "rewind [steps-or-sha]",
		Aliases: []string{"rw"},
		Short:   "Rewind the operations to an earlier state",
		Long: `Git must be installed and on your PATH for this to work. Using the power of git to rewind the operations to an earlier state.
	
	You can pass a "steps" number or a commit sha. If a steps number is passed, we will rewind the operations that many steps. If a commit sha is passed, we will rewind to that commit. If neither a steps number nor a commit sha is passed, the target scope will be rewound by 1 step.
	`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { // Changed to RunE
			return rewind(params, args)
		},
	}

	return rewindCmd
}

func rewind(params *cli.CmdParams, args []string) error {
	var stepsOrSha string
	if len(args) > 0 {
		stepsOrSha = args[0]
	} else {
		stepsOrSha = "1" // Default to 1 step
	}

	confirmed, err := params.Interactor.Confirm(fmt.Sprintf("Are you sure you want to rewind by '%s'? This may alter your file system state.", stepsOrSha), false)
	if err != nil {
		params.Interactor.Error("Confirmation failed", err)
		return err
	}

	if !confirmed {
		params.Interactor.Info("Rewind operation cancelled by user.")
		return nil
	}

	params.Interactor.StartSpinner(fmt.Sprintf("Rewinding to %s ...", stepsOrSha))

	// Rewind to the target sha
	if err := params.Filesystem.GitRewind(params.Filesystem.GetCwd(), stepsOrSha); err != nil {
		params.Interactor.StopSpinner(false, "Rewind failed.")
		params.Interactor.Error(fmt.Sprintf("Error rewinding to %s", stepsOrSha), err)
		return err
	}

	params.Interactor.StopSpinner(true, fmt.Sprintf("Successfully rewound to %s.", stepsOrSha))
	params.Interactor.Success("Rewind operation complete.")
	return nil
}
