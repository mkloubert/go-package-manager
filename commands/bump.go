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

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Bump_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var breaking bool
	var feature bool
	var fix bool
	var force bool
	var major int64
	var minor int64
	var message string
	var patch int64

	var bumpVersionCmd = &cobra.Command{
		Use:     "bump [args]",
		Aliases: []string{"bp", "bmp"},
		Short:   "Bump version",
		Long:    `Bumps a version number.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !app.NoPreScript {
				// prebump defined?
				_, ok := app.GpmFile.Scripts[constants.PreBumpScriptName]
				if ok {
					app.RunScript(constants.PreBumpScriptName)
				}
			}

			// custom bump logic?
			_, ok := app.GpmFile.Scripts[constants.BumpScriptName]
			if !app.NoScript && ok {
				app.RunScript(constants.BumpScriptName, args...)
			} else {
				pvm := app.NewVersionManager()

				bumpOptions := types.BumpProjectVersionOptions{
					Arguments: &args,
					Breaking:  &breaking,
					Feature:   &feature,
					Fix:       &fix,
					Force:     &force,
					Major:     &major,
					Message:   &message,
					Minor:     &minor,
					Patch:     &patch,
				}

				newVersion, err := pvm.Bump(bumpOptions)
				utils.CheckForError(err)

				if newVersion != nil {
					fmt.Printf("v%s%s", newVersion.String(), fmt.Sprintln())
				}
			}

			if !app.NoPostScript {
				// postbump defined?
				_, ok = app.GpmFile.Scripts[constants.PostBumpScriptName]
				if ok {
					app.RunScript(constants.PostBumpScriptName)
				}
			}
		},
	}

	bumpVersionCmd.Flags().BoolVarP(&breaking, "breaking", "", false, "increase major part by 1")
	bumpVersionCmd.Flags().BoolVarP(&feature, "feature", "", false, "increase minor part by 1")
	bumpVersionCmd.Flags().BoolVarP(&fix, "fix", "", false, "increase patch part by 1")
	bumpVersionCmd.Flags().BoolVarP(&force, "force", "", false, "ignore value of previous version")
	bumpVersionCmd.Flags().Int64VarP(&major, "major", "", -1, "set major part")
	bumpVersionCmd.Flags().StringVarP(&message, "message", "", "", "custom git message")
	bumpVersionCmd.Flags().Int64VarP(&minor, "minor", "", -1, "set minor part")
	bumpVersionCmd.Flags().Int64VarP(&patch, "patch", "", -1, "set patch part")

	parentCmd.AddCommand(
		bumpVersionCmd,
	)
}
