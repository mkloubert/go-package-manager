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
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Install_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var noUpdate bool

	var installCmd = &cobra.Command{
		Use:     "install [module name or url]",
		Aliases: []string{"i"},
		Short:   "Installs one or more modules",
		Long:    `Gets and installs one or more modules by a short name or a valid URL to a git repository.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, moduleName := range args {
				urls := app.GetModuleUrls(moduleName)

				for _, u := range urls {
					var p *exec.Cmd
					if noUpdate {
						app.Debug("Running 'go get " + u + "' ...")
						p = utils.CreateShellCommandByArgs("go", "get", u)
					} else {
						app.Debug("Running 'go get -u " + u + "' ...")
						p = utils.CreateShellCommandByArgs("go", "get", "-u", u)
					}

					utils.RunCommand(p)
				}
			}
		},
	}

	installCmd.Flags().BoolVarP(&noUpdate, "no-update", "n", false, "do not update modules")

	parentCmd.AddCommand(
		installCmd,
	)
}
