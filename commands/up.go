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
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"
)

func Init_Up_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var noBuild bool

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "Run docker up",
		Long:  `Runs docker compose up command.`,
		Run: func(cmd *cobra.Command, args []string) {
			customUpCommand := strings.TrimSpace(
				app.SettingsFile.GetString("up.command", "", ""),
			)
			if customUpCommand == "" {
				baseArgs := []string{"docker", "compose", "up"}
				if !noBuild {
					baseArgs = append(baseArgs, "--build")
				}

				shellArgs := append(baseArgs, args...)

				app.RunShellCommandByArgs(shellArgs[0], shellArgs[1:]...)
			} else {
				app.RunShellCommand(customUpCommand)
			}
		},
	}

	upCmd.Flags().BoolVarP(&noBuild, "no-build", "", false, "do not use --build flag for docker-compose")

	parentCmd.AddCommand(
		upCmd,
	)
}
