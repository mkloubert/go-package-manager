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
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func run_scripts(app *types.AppContext, args []string) {
	scriptsToExecute := []string{}

	for _, scriptName := range args {
		scriptName = strings.TrimSpace(scriptName)
		if scriptName == "" {
			continue
		}

		_, ok := app.GpmFile.Scripts[scriptName]
		if !ok {
			utils.CloseWithError(fmt.Errorf("script '%v' not found", scriptName))
		}

		scriptsToExecute = append(scriptsToExecute, scriptName)
	}

	if len(scriptsToExecute) == 0 {
		app.RunCurrentProject()
	} else {
		// run scripts

		for _, scriptName := range scriptsToExecute {
			app.RunScript(scriptName)
		}
	}
}

func Init_Run_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var mode string

	var runCmd = &cobra.Command{
		Use:     "run [resource]",
		Aliases: []string{"r"},
		Short:   "Run resource",
		Long:    `Runs resources like scripts by name.`,
		Run: func(cmd *cobra.Command, args []string) {
			m := strings.TrimSpace(strings.ToLower(mode))

			switch m {
			case "":
			case "s":
			case "script":
			case "scripts":
				run_scripts(app, args)
			default:
				utils.CloseWithError(fmt.Errorf("invalid value '%v' for mode", m))
			}
		},
	}

	runCmd.Flags().StringVarP(&mode, "mode", "m", "", "the mode like scripts or workflows")

	parentCmd.AddCommand(
		runCmd,
	)
}
