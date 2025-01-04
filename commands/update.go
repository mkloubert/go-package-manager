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
	"fmt"
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"
)

func Init_Update_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var force bool
	var noCleanup bool
	var noVersionPrint bool
	var powerShell bool
	var powerShellBin string
	var selfUpdate bool
	var updateScript string
	var userAgent string

	var updateCmd = &cobra.Command{
		Use:     "update <modules>",
		Aliases: []string{"upd"},
		Short:   "Update dependencies",
		Long:    `Updates all or only specific dependencies in this project.`,
		Run: func(cmd *cobra.Command, args []string) {
			if selfUpdate {
				run_self_update_command(
					app,
					force, noVersionPrint, powerShell, powerShellBin, updateScript, userAgent,
				)
			} else {
				modulesToUpdate := make([]string, 0)
				for _, moduleNameOrUrl := range args {
					// maybe alias => module url(s)
					moduleUrls := app.GetModuleUrls(moduleNameOrUrl)

					modulesToUpdate = append(modulesToUpdate, moduleUrls...)
				}

				additionalShellArgs := make([]string, 0)
				additionalShellArgs = append(additionalShellArgs, modulesToUpdate...)
				if len(modulesToUpdate) == 0 {
					app.Debug("Will update all modules in project ...")

					// update all in this project instead specific onces
					additionalShellArgs = append(additionalShellArgs, "./...")
				} else {
					// update specific ones
					app.Debug(fmt.Sprintf("Will update following modules in project: %s", strings.Join(modulesToUpdate, ",")))
				}

				allShellArgs := make([]string, 0)
				allShellArgs = append(allShellArgs, "get", "-u")
				allShellArgs = append(allShellArgs, additionalShellArgs...)

				app.RunShellCommandByArgs("go", allShellArgs...)

				if !noCleanup {
					app.TidyUp()
				}
			}
		},
	}

	updateCmd.Flags().BoolVarP(&force, "force", "", false, "force self-update")
	updateCmd.Flags().BoolVarP(&noCleanup, "no-cleanup", "", false, "do not cleanup go.mod and go.sum")
	updateCmd.Flags().BoolVarP(&noVersionPrint, "no-version-print", "", false, "do not print new version after successful update")
	updateCmd.Flags().BoolVarP(&powerShell, "powershell", "", false, "force execution of PowerShell script")
	updateCmd.Flags().StringVarP(&powerShellBin, "powershell-bin", "", "", "custom binary of the PowerShell")
	updateCmd.Flags().BoolVarP(&selfUpdate, "self", "", false, "update this binary instead")
	updateCmd.Flags().StringVarP(&updateScript, "update-script", "", "", "custom URL to update script")
	updateCmd.Flags().StringVarP(&userAgent, "user-agent", "", "", "custom string for user agent")

	parentCmd.AddCommand(
		updateCmd,
	)
}
