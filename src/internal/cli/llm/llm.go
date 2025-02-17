package llm

import (
	"file4you/internal/cli"
	"file4you/internal/llm"
	"fmt"

	"github.com/spf13/cobra"
)

func NewLLMAgent(params *cli.CmdParams) *cobra.Command {
	var (
		model    string
		dryRun   bool
		maxFiles int
	)

	llmagentCmd := &cobra.Command{
		Use:     "llmagent [prompt]",
		Aliases: []string{"llm"},
		Short:   "Run the LLM agent to reorganize files in the specified directory",
		Long: `Run the LLM agent to reorganize files based on the configuration. 
		Optionally specify a prompt. If not provided, the default prompt is used.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runLLMAgent(params, args, model, dryRun, maxFiles)
		},
	}

	llmagentCmd.Flags().StringVar(&model, "model", "gpt-3.5-turbo", "LLM model to use")
	llmagentCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show proposed changes without executing them")
	llmagentCmd.Flags().IntVar(&maxFiles, "max-files", 100, "Maximum number of files to process")

	return llmagentCmd
}

func runLLMAgent(params *cli.CmdParams, args []string, model string, dryRun bool, maxFiles int) {
	var userPrompt string
	if len(args) > 0 {
		userPrompt = args[0]
	}

	params.Term.ToggleSpinner(true, "Initializing LLM agent...")

	// Create LLM client
	client, err := llm.NewClient(llm.WithModel(model))
	if err != nil {
		params.Term.OutputErrorAndExit("Failed to initialize LLM client: %v", err)
	}

	// Create orchestrator with the client
	orchestrator := llm.NewOrchestrator(client, nil)

	// Start processing files
	params.Term.ToggleSpinner(true, "Analyzing files...")

	// Get files from current directory
	files, err := params.DeskFS.ListFiles(params.DeskFS.Cwd, maxFiles)
	if err != nil {
		params.Term.OutputErrorAndExit("Failed to list files: %v", err)
	}

	// Process each file
	var decisions []llm.DestinationDecision
	for _, file := range files {
		decision, err := orchestrator.ProcessFile(cmd.Context(),
			file.Name(),
			file.Type(),
			file.Preview(),
			file.ModTime().Format("2006-01-02"))

		if err != nil {
			params.Term.OutputWarning("Failed to process %s: %v", file.Name(), err)
			continue
		}
		decisions = append(decisions, decision)
	}

	params.Term.ToggleSpinner(false, "")

	// If dry run, just show the proposed changes
	if dryRun {
		displayProposedChanges(params, decisions)
		return
	}

	// Execute the changes
	executeChanges(params, decisions)
}

func displayProposedChanges(params *cli.CmdParams, decisions []llm.DestinationDecision) {
	params.Term.OutputInfo("\nProposed changes:")
	for _, d := range decisions {
		params.Term.OutputInfo("Move %s → %s/%s",
			d.OriginalPath,
			d.DestinationFolder,
			d.NewFileName)
	}
}

func executeChanges(params *cli.CmdParams, decisions []llm.DestinationDecision) {
	params.Term.ToggleSpinner(true, "Applying changes...")

	for _, d := range decisions {
		if err := params.DeskFS.MoveFile(
			d.OriginalPath,
			fmt.Sprintf("%s/%s", d.DestinationFolder, d.NewFileName),
		); err != nil {
			params.Term.OutputWarning("Failed to move %s: %v", d.OriginalPath, err)
			continue
		}
	}

	params.Term.ToggleSpinner(false, "")
	params.Term.OutputSuccess("Successfully reorganized files using LLM agent")
}
