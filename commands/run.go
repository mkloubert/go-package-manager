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

func Init_Run_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var runCmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Runs a command by name",
		Long:    `Runs a command by name which is defined in gpm.ya(m)l file.`,
		Run: func(cmd *cobra.Command, args []string) {
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
				for _, scriptName := range scriptsToExecute {
					app.RunScript(scriptName)
				}
			}
		},
	}

	parentCmd.AddCommand(
		runCmd,
	)
}