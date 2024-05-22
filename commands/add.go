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

func init_add_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var reset bool

	var addAliasCmd = &cobra.Command{
		Use:     "alias",
		Aliases: []string{"a"},
		Short:   "Add package alias",
		Long:    `Adds an alias for one or more packages.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			alias := strings.TrimSpace(args[0])

			if reset {
				app.Debug(fmt.Sprintf("Resetting list of package alias '%v' ...", alias))
				app.AliasesFile.Aliases[alias] = []string{}
			}

			sources := app.AliasesFile.Aliases[alias]

			for _, s := range args[1:] {
				s = strings.TrimSpace(s)
				if s != "" {
					app.Debug(fmt.Sprintf("Adding source '%v' for package alias '%v' ...", s, alias))
					sources = append(sources, s)
				}
			}

			app.AliasesFile.Aliases[alias] = sources

			err := app.UpdateAliasesFile()
			utils.CheckForError(err)
		},
	}

	addAliasCmd.Flags().BoolVarP(&reset, "reset", "r", false, "reset list before add")

	parentCmd.AddCommand(
		addAliasCmd,
	)
}

func init_add_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var addProjectCmd = &cobra.Command{
		Use:     "project [alias] [git resource]",
		Aliases: []string{"p", "prj"},
		Short:   "Add project",
		Long:    `Adds project with a specific alias and Git resource.`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			alias := strings.TrimSpace(args[0])
			gitResource := strings.TrimSpace(args[1])

			app.ProjectsFile.Projects[alias] = gitResource

			err := app.UpdateProjectsFile()
			utils.CheckForError(err)
		},
	}

	parentCmd.AddCommand(
		addProjectCmd,
	)
}

func Init_Add_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var addCmd = &cobra.Command{
		Use:     "add [resource]",
		Aliases: []string{"ad"},
		Short:   "Add command",
		Long:    `Adds a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_add_alias_command(addCmd, app)
	init_add_project_command(addCmd, app)

	parentCmd.AddCommand(
		addCmd,
	)
}
