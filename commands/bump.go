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

	"github.com/hashicorp/go-version"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func init_bump_version_command(parentCmd *cobra.Command, app *types.AppContext) {
	var breaking bool
	var feature bool
	var fix bool
	var major int64
	var minor int64
	var message string
	var patch int64

	var versionCmd = &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "ver"},
		Short:   "Bump version",
		Long:    `Bumps a version number.`,
		Run: func(cmd *cobra.Command, args []string) {
			latestVersion, err := app.GetLatestVersion()
			if err != nil {
				utils.CloseWithError(err)
			}

			if latestVersion == nil {
				latestVersion, _ = version.NewVersion("0.0.0")
			}

			segments := latestVersion.Segments64()
			currentMajor := segments[0]
			currentMinor := segments[1]
			currentPatch := segments[2]

			newMajor := currentMajor
			if major > -1 {
				newMajor = major
			}
			newMinor := currentMinor
			if minor > -1 {
				newMinor = minor
			}
			newPatch := currentPatch
			if patch > -1 {
				newPatch = patch
			}

			if !breaking && !feature && !fix {
				// default is increase minor
				newMinor++
			} else {
				if breaking {
					newMajor++
				}
				if feature {
					newMinor++
				}
				if fix {
					newPatch++
				}
			}

			nextVersion, err := version.NewVersion(
				fmt.Sprintf(
					"%v.%v.%v",
					newMajor, newMinor, newPatch,
				),
			)
			if err != nil {
				utils.CloseWithError(err)
			}

			if nextVersion.LessThanOrEqual(latestVersion) {
				utils.CloseWithError(fmt.Errorf("new version is not greater than latest one"))
			}

			gitMessage := strings.TrimSpace(message)
			if gitMessage == "" {
				gitMessage = fmt.Sprintf("version %v", nextVersion.String())
			}

			tagName := fmt.Sprintf("v%v", nextVersion.String())

			app.RunShellCommandByArgs("git", "tag", "-a", tagName, "-m", gitMessage)
		},
	}

	versionCmd.Flags().BoolVarP(&breaking, "breaking", "", false, "increase major part by 1")
	versionCmd.Flags().BoolVarP(&feature, "feature", "", false, "increase minor part by 1")
	versionCmd.Flags().BoolVarP(&fix, "fix", "", false, "increase patch part by 1")
	versionCmd.Flags().Int64VarP(&major, "major", "", -1, "set major part")
	versionCmd.Flags().StringVarP(&message, "message", "", "", "custom git message")
	versionCmd.Flags().Int64VarP(&minor, "minor", "", -1, "set minor part")
	versionCmd.Flags().Int64VarP(&patch, "patch", "", -1, "set patch part")

	parentCmd.AddCommand(
		versionCmd,
	)
}

func Init_Bump_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var bumpCmd = &cobra.Command{
		Use:     "bump",
		Aliases: []string{"bp", "bmp"},
		Short:   "Bump resource",
		Long:    `Bumps a resource like a version.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_bump_version_command(bumpCmd, app)

	parentCmd.AddCommand(
		bumpCmd,
	)
}
