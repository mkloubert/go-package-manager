// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package commands

import (
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
)

func Init_Install_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var noPostScript bool
	var noPreScript bool
	var noTidyScript bool
	var noUpdate bool
	var tidy bool
	var tidyArgs []string

	var installCmd = &cobra.Command{
		Use:     "install [module name or url]",
		Aliases: []string{"i", "inst"},
		Short:   "Installs one or more modules",
		Long:    `Gets and installs one or more modules by a short name or a valid URL to a git repository.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !noPreScript {
				_, ok := app.GpmFile.Scripts[constants.PreInstallScriptName]
				if ok {
					app.RunScript(constants.PreInstallScriptName)
				}
			}

			for _, moduleName := range args {
				urls := app.GetModuleUrls(moduleName)

				for _, u := range urls {
					if noUpdate {
						app.RunShellCommandByArgs("go", "get", u)
					} else {
						app.RunShellCommandByArgs("go", "get", "-u", u)
					}
				}
			}

			if !noPostScript {
				_, ok := app.GpmFile.Scripts[constants.PostInstallScriptName]
				if ok {
					app.RunScript(constants.PostInstallScriptName)
				}
			}

			if tidy {
				app.TidyUp(types.TidyUpOptions{
					Arguments: &tidyArgs,
					NoScript:  &noTidyScript,
				})
			}
		},
	}

	installCmd.Flags().BoolVarP(&noPostScript, "no-post-script", "", false, "do not handle '"+constants.PostInstallScriptName+"' script")
	installCmd.Flags().BoolVarP(&noPreScript, "no-pre-script", "", false, "do not handle '"+constants.PreInstallScriptName+"' script")
	installCmd.Flags().BoolVarP(&noPreScript, "no-tidy-script", "", false, "do not handle '"+constants.TidyScriptName+"' script")
	installCmd.Flags().BoolVarP(&noUpdate, "no-update", "n", false, "do not update modules")
	installCmd.Flags().BoolVarP(&tidy, "tidy", "", false, "tidy up project after install")
	installCmd.Flags().StringArrayVarP(&tidyArgs, "tidy-arg", "", []string{}, "arguments for tidy command")

	parentCmd.AddCommand(
		installCmd,
	)
}
