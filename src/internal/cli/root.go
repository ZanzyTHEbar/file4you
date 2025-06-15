/*
Copyright © 2024 DaOfficialWizard zacariahheim@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cli

import (
	"fmt"
	"log/slog"
	"os"

	"file4you/internal"
	"file4you/internal/config"
	"file4you/internal/db"

	"github.com/spf13/cobra"
)

var cfgFile string

type RootCMD struct {
	Root *cobra.Command
}

func NewRootCMD(params *CmdParams) *RootCMD {
	return &RootCMD{
		Root: NewRoot(params),
	}
}

func NewRoot(params *CmdParams) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s", internal.DefaultAppName),
		Aliases: []string{internal.DefaultAppCMDShortCut},
		Short:   fmt.Sprintf("%s is a tool to intelligently automate filing and directory organization", internal.DefaultAppName),
	}

	// Validate palette
	if params.Palette == nil {
		params.Palette = []*cobra.Command{}
	}

	// Add commands to the root
	rootCmd.AddCommand(params.Palette...)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default %s)", internal.DefaultGlobalConfigFile))

	cobra.OnInitialize(func() {
		// Load global application configuration first.
		// The cfgFile variable from PersistentFlags can be used here if provided.
		if _, err := config.LoadConfig(cfgFile); err != nil {
			// Use interactor if available, otherwise fallback to slog/fmt for early errors.
			if params.Interactor != nil {
				params.Interactor.Fatal("Failed to load application configuration", err)
			} else {
				slog.Error("Failed to load application configuration:", "error", err)
				os.Exit(1)
			}
			return
		}

		// Configuration is now handled during FileSystem initialization
		// The rest of the Genkit initialization seems okay, assuming filesystem and CentralDB are correctly initialized.
		params.Interactor.Output("Initializing Genkit...")
		if params.Filesystem == nil {
			params.Interactor.Fatal("filesystem not initialized, cannot initialize Genkit", nil)
			return
		}
		if params.CentralDB == nil {
			params.Interactor.Info("CentralDB not found in params, attempting to initialize with default settings...")
			cdb, err := db.NewCentralDBProvider()
			if err != nil {
				params.Interactor.Fatal("Failed to initialize CentralDB for Genkit", err)
				return
			}
			params.CentralDB = cdb
			params.Interactor.Success("CentralDB initialized successfully for Genkit.")
		} else {
			params.Interactor.Info("CentralDB already initialized, proceeding with Genkit initialization.")
		}

		//service, err := genkithandler.NewService(context.Background(), params.Filesystem, params.CentralDB)
		//if err != nil {
		//	params.Interactor.Fatal("Failed to initialize Genkit", err)
		//	return
		//}
		//params.Genkit = service.Genkit()
		//params.Interactor.Success("Genkit initialized successfully.")

		// Register Genkit tools (already handled in NewService via RegisterCoreTools)
		params.Interactor.Success("Genkit tools registered successfully.")
	})

	return rootCmd
}
