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

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Publish_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var breaking bool
	var defaultRemoteOnly bool
	var feature bool
	var fix bool
	var force bool
	var major int64
	var minor int64
	var message string
	var noBump bool
	var patch int64

	var publishCmd = &cobra.Command{
		Use:     "publish [remotes]",
		Aliases: []string{"pub"},
		Short:   "Publish version",
		Long:    `Bumps the version of the current project and pushes it to all remote repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			currentBranchName, _ := app.GetCurrentGitBranch()

			if !noBump {
				pvm := app.NewVersionManager()

				bumpOptions := types.BumpProjectVersionOptions{
					Breaking: &breaking,
					Feature:  &feature,
					Fix:      &fix,
					Force:    &force,
					Major:    &major,
					Message:  &message,
					Minor:    &minor,
					Patch:    &patch,
				}

				newVersion, err := pvm.Bump(bumpOptions)
				utils.CheckForError(err)

				if newVersion != nil {
					fmt.Printf("v%s%s", newVersion.String(), fmt.Sprintln())
				}
			}

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
				// first push code
				{
					cmdArgs := []string{"git", "push", r, currentBranchName}

					app.RunShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)
				}

				// then push tags
				{
					cmdArgs := []string{"git", "push", r, "--tags"}

					app.RunShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)
				}
			}
		},
	}

	publishCmd.Flags().BoolVarP(&breaking, "breaking", "", false, "increase major part by 1")
	publishCmd.Flags().BoolVarP(&defaultRemoteOnly, "default", "d", false, "default / first remote only")
	publishCmd.Flags().BoolVarP(&feature, "feature", "", false, "increase minor part by 1")
	publishCmd.Flags().BoolVarP(&fix, "fix", "", false, "increase patch part by 1")
	publishCmd.Flags().BoolVarP(&force, "force", "", false, "ignore value of previous version")
	publishCmd.Flags().Int64VarP(&major, "major", "", -1, "set major part")
	publishCmd.Flags().StringVarP(&message, "message", "", "", "custom git message")
	publishCmd.Flags().Int64VarP(&minor, "minor", "", -1, "set minor part")
	publishCmd.Flags().BoolVarP(&noBump, "no-bump", "", false, "do not bump version")
	publishCmd.Flags().Int64VarP(&patch, "patch", "", -1, "set patch part")

	parentCmd.AddCommand(
		publishCmd,
	)
}
