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

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_import_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var reset bool

	var importAliasCmd = &cobra.Command{
		Use:     "aliases [source]",
		Aliases: []string{"a", "al", "alias"},
		Short:   "Import alias",
		Long:    `Downloads alias files from external resources and merge them with local one.`,
		Run: func(cmd *cobra.Command, args []string) {
			importFromYaml := func(yamlData []byte) {
				var aliasFile types.AliasesFile
				err := yaml.Unmarshal(yamlData, &aliasFile)
				if err != nil {
					utils.CloseWithError(err)
				}

				if aliasFile.Aliases == nil {
					return
				}

				for alias, urls := range aliasFile.Aliases {
					app.Debug(fmt.Sprintf("Updating alias '%v' with '%v' ...", alias, urls))
					app.AliasesFile.Aliases[alias] = urls
				}
			}

			if reset {
				app.AliasesFile.Aliases = map[string][]string{}
			}

			// collect data ...
			for _, a := range args {
				alias := strings.TrimSpace(a)
				if alias == "" {
					continue
				}

				yamlData, err := app.LoadDataFrom(alias)
				if err != nil {
					utils.CloseWithError(err)
				}

				importFromYaml(yamlData)
			}

			stdin, err := utils.LoadFromSTDINIfAvailable()
			if err != nil {
				utils.CloseWithError(err)
			}
			if stdin != nil {
				app.Debug("Updating projects from STDIN ...")
				importFromYaml(*stdin)
			}

			// ... finally update aliases file
			err = app.UpdateAliasesFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	importAliasCmd.Flags().BoolVarP(&reset, "reset", "", false, "reset before import entries")

	parentCmd.AddCommand(
		importAliasCmd,
	)
}

func init_import_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var reset bool

	var importProjectCmd = &cobra.Command{
		Use:     "projects [source]",
		Aliases: []string{"p", "pr", "prj", "prjs", "project"},
		Short:   "Import project",
		Long:    `Downloads project files from external resources and merge them with local one.`,
		Run: func(cmd *cobra.Command, args []string) {
			importFromYaml := func(yamlData []byte) {
				var projectFile types.ProjectsFile
				err := yaml.Unmarshal(yamlData, &projectFile)
				if err != nil {
					utils.CloseWithError(err)
				}

				if projectFile.Projects == nil {
					return
				}

				for alias, url := range projectFile.Projects {
					app.Debug(fmt.Sprintf("Updating project '%v' with '%v' ...", alias, url))
					app.ProjectsFile.Projects[alias] = url
				}
			}

			if reset {
				app.ProjectsFile.Projects = map[string]string{}
			}

			// collect data ...
			for _, a := range args {
				source := strings.TrimSpace(a)
				if source == "" {
					continue
				}

				yamlData, err := app.LoadDataFrom(source)
				if err != nil {
					utils.CloseWithError(err)
				}

				importFromYaml(yamlData)
			}

			stdin, err := utils.LoadFromSTDINIfAvailable()
			if err != nil {
				utils.CloseWithError(err)
			}
			if stdin != nil {
				app.Debug("Updating projects from STDIN ...")
				importFromYaml(*stdin)
			}

			// ... finally update projects file
			err = app.UpdateProjectsFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	importProjectCmd.Flags().BoolVarP(&reset, "reset", "", false, "reset before import entries")

	parentCmd.AddCommand(
		importProjectCmd,
	)
}

func Init_Import_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var importCmd = &cobra.Command{
		Use:     "import [resource]",
		Aliases: []string{"im", "imp"},
		Short:   "Import resource",
		Long:    `Imports a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_import_alias_command(importCmd, app)
	init_import_project_command(importCmd, app)

	parentCmd.AddCommand(
		importCmd,
	)
}
