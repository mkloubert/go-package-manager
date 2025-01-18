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
	"os"
	"os/exec"
	"strings"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
)

func Init_Cron_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var cronCmd = &cobra.Command{
		Use:   "cron [pattern] [command] [args]",
		Args:  cobra.MinimumNArgs(2),
		Short: "Cron job",
		Long:  `Runs scripts or executables periodically using cron syntax.`,
		Run: func(cmd *cobra.Command, args []string) {
			patterns := strings.TrimSpace(args[0])

			command := strings.TrimSpace(args[1])
			commandArgs := args[2:]

			allCommandArgs := args[1:]

			c := cron.New()

			c.AddFunc(patterns, func() {
				p := exec.Command(command, commandArgs...)
				p.Dir = app.Cwd
				p.Env = os.Environ()
				p.Stderr = app.ErrorOut
				p.Stdin = app.In
				p.Stdout = app.Out

				app.Debug(fmt.Sprintf("Try running '%s' ...", strings.Join(allCommandArgs, " ")))

				err := p.Run()
				if err != nil {
					app.Out.Write(
						[]byte(fmt.Sprintf(
							"could not execute %s %s: %s%s",
							command, strings.Join(commandArgs, " "),
							err.Error(),
							fmt.Sprintln(),
						)),
					)
				}
			})

			go func() {
				app.Debug(fmt.Sprintf("Will execute '%s' every '%s' ...", strings.Join(allCommandArgs, " "), patterns))

				c.Start()
			}()

			select {}
		},
	}

	cronCmd.DisableFlagParsing = true

	parentCmd.AddCommand(
		cronCmd,
	)
}
