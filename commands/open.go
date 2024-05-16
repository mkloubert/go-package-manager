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

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_open_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var openAliasCmd = &cobra.Command{
		Use:     "alias [name]",
		Aliases: []string{"a", "al", "aliases"},
		Short:   "Open alias",
		Long:    `Opens the URL of an alias in the operating system.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				alias := strings.TrimSpace(a)
				if alias == "" {
					continue
				}

				urls, ok := app.AliasesFile.Aliases[alias]
				if ok {
					for _, u := range urls {
						urlToOpen, err := utils.ToUrlForOpenHandler(u)
						if err == nil {
							app.Debug(fmt.Sprintf("Opening alias '%v' with URL '%v' ...", urlToOpen, alias))
							err = utils.OpenUrl(u)
							if err != nil {
								app.Debug(fmt.Sprintf("Warning: Could not open URL '%v' of alias '%v': '%v'", urlToOpen, alias, err))
							}
						} else {
							app.Debug(fmt.Sprintf("Warning: Could not parse URL '%v' of alias '%v': '%v'", u, alias, err))
						}
					}
				} else {
					app.Debug(fmt.Sprintf("Warning: Alias '%v' not found!", alias))
				}
			}
		},
	}

	parentCmd.AddCommand(
		openAliasCmd,
	)
}

func init_open_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var openProjectCmd = &cobra.Command{
		Use:     "project [name]",
		Aliases: []string{"p", "pr", "prj", "prjs", "projects"},
		Short:   "Open project",
		Long:    `Opens the URL of a project in the operating system.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, p := range args {
				projectAlias := strings.TrimSpace(p)
				if projectAlias == "" {
					continue
				}

				url, ok := app.ProjectsFile.Projects[projectAlias]
				if ok {
					urlToOpen, err := utils.ToUrlForOpenHandler(url)
					if err == nil {
						app.Debug(fmt.Sprintf("Opening project '%v' with URL '%v' ...", urlToOpen, projectAlias))
						err = utils.OpenUrl(urlToOpen)
						if err != nil {
							app.Debug(fmt.Sprintf("Warning: Could not open URL '%v' of project '%v': '%v'", urlToOpen, projectAlias, err))
						}
					} else {
						app.Debug(fmt.Sprintf("Warning: Could not parse URL '%v' of project '%v': '%v'", url, projectAlias, err))
					}
				} else {
					app.Debug(fmt.Sprintf("Warning: Project '%v' not found!", projectAlias))
				}
			}
		},
	}

	parentCmd.AddCommand(
		openProjectCmd,
	)
}

func Init_Open_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var openCmd = &cobra.Command{
		Use:     "open [resource]",
		Aliases: []string{"o", "opn"},
		Short:   "Open resource",
		Long:    `Opens a resource in the operating system.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_open_alias_command(openCmd, app)
	init_open_project_command(openCmd, app)

	parentCmd.AddCommand(
		openCmd,
	)
}
