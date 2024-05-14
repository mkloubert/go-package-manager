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
	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"
)

const buildScriptName = "build"

func Init_Build_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var noScript bool

	var buildCmd = &cobra.Command{
		Use:     "build",
		Aliases: []string{"b"},
		Short:   "Runs build command",
		Long:    `Runs the 'build' script or the official 'go build .'.`,
		Run: func(cmd *cobra.Command, args []string) {
			_, ok := app.GpmFile.Scripts[buildScriptName]

			if !noScript && ok {
				app.RunScript(buildScriptName, args...)
			} else {
				cmdArgs := []string{"go", "build", "."}
				cmdArgs = append(cmdArgs, args...)

				app.RunShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)
			}
		},
	}

	buildCmd.Flags().BoolVarP(&noScript, "no-script", "n", false, "do not handle 'build' script")

	parentCmd.AddCommand(
		buildCmd,
	)
}
