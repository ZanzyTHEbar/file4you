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
		Run: func(cmd *cobra.Command, args []string) {
			rewind(params, args)
		},
	}

	return rewindCmd
}

func rewind(params *cli.CmdParams, args []string) {
	var stepsOrSha string
	if len(args) > 0 {
		stepsOrSha = args[0]
	} else {
		stepsOrSha = "1"
	}

	params.Term.ToggleSpinner(true, fmt.Sprintf("Rewinding to %s ...", stepsOrSha))

	// Rewind to the target sha
	if err := params.DeskFS.GitRewind(params.DeskFS.Cwd, stepsOrSha); err != nil {
		params.Term.OutputErrorAndExit("Error rewinding to %s: %v", stepsOrSha, err)
	}

	params.Term.ToggleSpinner(false, "")
}
