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

func Init_Down_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var downCmd = &cobra.Command{
		Use:   "down",
		Short: "Run docker down",
		Long:  `Runs docker compose down command.`,
		Run: func(cmd *cobra.Command, args []string) {
			customDownCommand := strings.TrimSpace(
				app.SettingsFile.GetString("down.command", "", ""),
			)
			if customDownCommand == "" {
				shellArgs := append([]string{"docker", "compose", "down"}, args...)

				app.RunShellCommandByArgs(shellArgs[0], shellArgs[1:]...)
			} else {
				app.RunShellCommand(customDownCommand)
			}
		},
	}

	parentCmd.AddCommand(
		downCmd,
	)
}
