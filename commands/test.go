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

const testScriptName = "test"

func Init_Test_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var testCmd = &cobra.Command{
		Use:     "test",
		Aliases: []string{"t", "tst"},
		Short:   "Runs tests",
		Long:    `Runs tests or 'test' script, if defined.`,
		Run: func(cmd *cobra.Command, args []string) {
			_, ok := app.GpmFile.Scripts[testScriptName]

			if ok {
				app.RunScript(testScriptName, args...)
			} else {
				app.RunShellCommandByArgs("go", "test", ".")
			}
		},
	}

	parentCmd.AddCommand(
		testCmd,
	)
}
