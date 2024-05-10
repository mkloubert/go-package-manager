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
		Long:    `Runs a command by name which is defined in packages.ya(m)l file.`,
		Run: func(cmd *cobra.Command, args []string) {
			commandsToExecute := []string{}

			for _, scriptName := range args {
				scriptName = strings.TrimSpace(scriptName)
				if scriptName == "" {
					continue
				}

				cmdToExecute, ok := app.PackagesFile.Scripts[scriptName]
				if !ok {
					utils.CloseWithError(fmt.Errorf("script '%v' not found", scriptName))
					return
				}

				commandsToExecute = append(commandsToExecute, cmdToExecute)
			}

			for _, cmdToExecute := range commandsToExecute {
				app.Debug(fmt.Sprintf("Running '%v' ...", cmdToExecute))

				p := utils.CreateShellCommand(cmdToExecute)

				if err := p.Run(); err != nil {
					utils.CloseWithError(err)
				}
			}
		},
	}

	parentCmd.AddCommand(
		runCmd,
	)
}
