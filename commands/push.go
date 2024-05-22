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

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Push_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var defaultRemoteOnly bool

	var pushCmd = &cobra.Command{
		Use:     "push [remotes]",
		Aliases: []string{"psh"},
		Short:   "Push to remotes",
		Long:    `Push to all git remotes or to specific ones.`,
		Run: func(cmd *cobra.Command, args []string) {
			currentBranchName, _ := app.GetCurrentGitBranch()

			var remotes []string
			if len(args) == 0 {
				listOfRemotes, err := app.GetGitRemotes()
				utils.CheckForError(err)

				remotes = append(remotes, listOfRemotes...)
			} else {
				remotes = append(remotes, args...)
			}

			if len(remotes) == 0 {
				utils.CloseWithError(fmt.Errorf("no remotes found"))
			}

			if defaultRemoteOnly {
				// default only
				remotes = []string{remotes[0]}
			}

			for _, r := range remotes {
				cmdArgs := []string{"git", "push", r, currentBranchName}

				app.RunShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)
			}
		},
	}

	pushCmd.Flags().BoolVarP(&defaultRemoteOnly, "default", "d", false, "default / first remote only")

	parentCmd.AddCommand(
		pushCmd,
	)
}
