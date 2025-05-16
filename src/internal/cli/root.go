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
	"context" // Added for Genkit initialization
	"errors"
	"fmt"

	"file4you/internal"
	"file4you/internal/db"            // Added for CentralDB initialization
	"file4you/internal/genkithandler" // Added for Genkit functions

	"github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/firebase/genkit/go/genkit"
	"github.com/spf13/cobra"
	// "github.com/spf13/viper" // Viper instance was unused
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
		params.DeskFS.InitConfig(cfgFile, params.Interactor)
		// params.DeskFS.InstanceConfig is *deskfs.DeskFSConfig
		// deskfs.DeskFSConfig embeds gobaselogger.Config
		// So, params.DeskFS.InstanceConfig.Config is the embedded gobaselogger.Config
		logger.InitLogger(&params.DeskFS.InstanceConfig.Config)

		// Accessing Logger.Level from the embedded gobaselogger.Config
		params.Interactor.Outputf("Configuration loaded. Log level set to: %s", params.DeskFS.InstanceConfig.Logger.Level)

		// Initialize Genkit
		params.Interactor.Output("Initializing Genkit...")
		genkitInstanceInter, err := genkithandler.InitializeGenkit(context.Background())
		if err != nil {
			params.Interactor.Fatal("Failed to initialize Genkit", err)
			return // Exit if Genkit initialization fails
		}

		// Type assert the returned interface{} to *genkit.Genkit
		genkitActualInstance, ok := genkitInstanceInter.(*genkit.Genkit)
		if !ok {
			// This should ideally not happen if InitializeGenkit behaves as expected
			params.Interactor.Fatal("Failed to assert Genkit instance type", errors.New("genkit instance type assertion failed"))
			return
		}
		params.Genkit = genkitActualInstance // Store the *genkit.Genkit instance
		params.Interactor.Success("Genkit initialized successfully.")

		// Register Genkit tools
		params.Interactor.Output("Registering Genkit tools...")
		if params.DeskFS == nil {
			params.Interactor.Fatal("DeskFS not initialized, cannot register tools", nil)
			return
		}

		// Ensure CentralDB is initialized and available in params
		if params.CentralDB == nil {
			params.Interactor.Info("CentralDB not found in params, attempting to initialize with default settings...")
			cdb, err := db.NewCentralDBProvider()
			if err != nil {
				params.Interactor.Fatal("Failed to initialize CentralDB for tool registration", err)
				return
			}
			params.CentralDB = cdb
			params.Interactor.Success("CentralDB initialized successfully for tool registration.")
		} else {
			params.Interactor.Info("CentralDB already initialized, proceeding with tool registration.")
		}

		// Use LegacyRegisterBackupTool. It takes the genkit instance (params.Genkit), DeskFS, and CentralDB.
		// The first argument to LegacyRegisterBackupTool was gInter interface{}, which is the *genkit.Genkit instance.
		genkithandler.LegacyRegisterBackupTool(params.Genkit, params.DeskFS, params.CentralDB)
		// Similarly for other tools if they are to be registered here:
		// genkithandler.RegisterOrganizeTool(params.Genkit, params.DeskFS)
		// genkithandler.RegisterWorkspaceTool(params.Genkit, params.DeskFS)
		params.Interactor.Success("Genkit tools registered successfully.")
	})

	return rootCmd
}
