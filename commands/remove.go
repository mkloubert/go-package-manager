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
	"path"
	"strings"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func init_remove_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var removeAliasCmd = &cobra.Command{
		Use:     "alias [alias name]",
		Aliases: []string{"a", "aliases"},
		Short:   "Remove package alias",
		Long:    `Removes one or more aliases.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				alias := strings.TrimSpace(a)

				app.Debug(fmt.Sprintf("Removing package alias '%v' ...", alias))
				delete(app.AliasesFile.Aliases, alias)
			}

			err := app.UpdateAliasesFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	parentCmd.AddCommand(
		removeAliasCmd,
	)
}

func init_remove_binary_command(parentCmd *cobra.Command, app *types.AppContext) {
	var noAutoExt bool

	var removeBinaryCmd = &cobra.Command{
		Use:     "binary [executable name]",
		Aliases: []string{"b", "bin", "bins", "binaries"},
		Short:   "Remove package alias",
		Long:    `Removes one or more aliases.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				binName := strings.TrimSpace(a)

				binFilename := binName
				if !noAutoExt {
					if utils.IsWindows() && !strings.HasSuffix(binFilename, constants.WindowsExecutableExt) {
						binFilename += constants.WindowsExecutableExt
					}
				}

				binPath, err := app.EnsureBinFolder()
				if err != nil {
					utils.CloseWithError(err)
				}

				executableFilePath := path.Join(binPath, binFilename)

				isExecutableFileExisting, err := utils.IsFileExisting(executableFilePath)
				if err != nil {
					utils.CloseWithError(err)
				}
				if !isExecutableFileExisting {
					app.Debug(fmt.Sprintf("Executable file '%v' not found", executableFilePath))
					return
				}

				app.Debug(fmt.Sprintf("Removing executable file '%v' ...", executableFilePath))
				err = os.Remove(executableFilePath)
				if err != nil {
					utils.CloseWithError(err)
				}
			}
		},
	}

	removeBinaryCmd.Flags().BoolVarP(&noAutoExt, "no-auto-extension", "", false, "do not add file extension automatically")

	parentCmd.AddCommand(
		removeBinaryCmd,
	)
}

func init_remove_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var removeAliasCmd = &cobra.Command{
		Use:     "project [alias name]",
		Aliases: []string{"p", "projects", "prj", "prjs"},
		Short:   "Remove project",
		Long:    `Removes one or more projects with their git resources.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				alias := strings.TrimSpace(a)

				app.Debug(fmt.Sprintf("Removing project '%v' ...", alias))
				delete(app.ProjectsFile.Projects, alias)
			}

			err := app.UpdateProjectsFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	parentCmd.AddCommand(
		removeAliasCmd,
	)
}

func Init_Remove_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var removeCmd = &cobra.Command{
		Use:     "remove [resource]",
		Aliases: []string{"rm"},
		Short:   "Remove command",
		Long:    `Removes a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_remove_alias_command(removeCmd, app)
	init_remove_binary_command(removeCmd, app)
	init_remove_project_command(removeCmd, app)

	parentCmd.AddCommand(
		removeCmd,
	)
}
