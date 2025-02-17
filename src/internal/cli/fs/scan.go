package fs

import (
	"encoding/json"
	"file4you/internal/cli"
	"file4you/internal/filesystem/trees"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type ScanCMD struct {
	Organize *cobra.Command
}

func NewScanCmd(params *cli.CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan [directory]",
		Short: "Scan directory tree and output JSON",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var root string
			if len(args) > 0 {
				root = args[0]
			} else {
				var err error
				root, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current working directory: %w", err)
				}
			}

			dt := trees.NewDirectoryTree(trees.WithRoot(root))

			// Populate the directory tree recursively
			if err := dt.Walk(); err != nil {
				return fmt.Errorf("error walking directory tree: %w", err)
			}

			jsonBytes, err := json.MarshalIndent(dt, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal directory tree to JSON: %w", err)
			}

			fmt.Println(string(jsonBytes))
			return nil
		},
	}

	return cmd
}
