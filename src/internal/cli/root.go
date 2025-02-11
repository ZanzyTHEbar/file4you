/*
Copyright © 2024 DaOfficialWizard pyr0ndet0s97@gmail.com

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
	"github.com/ZanzyTHEbar/go-basetools/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		Use:     "desktop-cleaner [command] [flags]",
		Aliases: []string{"dcx"},
		Short:   "DesktopCleaner is a tool to automate the clean up of a specified directory",
	}

	// Validate palette
	if params.Palette == nil {
		params.Palette = []*cobra.Command{}
	}

	// Add commands to the root
	rootCmd.AddCommand(params.Palette...)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/.desktop_cleaner/.desktop_cleaner.toml)")

	viper.AutomaticEnv() // read in environment variables that match

	params.DeskFS.InitConfig(cfgFile)

	logger.InitLogger(&params.DeskFS.InstanceConfig.Config)

	return rootCmd
}
