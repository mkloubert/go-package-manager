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
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_import_projects_command(parentCmd *cobra.Command, app *types.AppContext) {
	var noDefaultSource bool
	var reset bool

	var importProjectsCmd = &cobra.Command{
		Use:     "projects [source]",
		Aliases: []string{"p", "pr", "prj", "prjs", "project"},
		Short:   "Import project",
		Long:    `Downloads project files from external resources and merge them with local one.`,
		Run: func(cmd *cobra.Command, args []string) {
			importFromYaml := func(yamlData []byte) {
				var projectFile types.ProjectsFile
				err := yaml.Unmarshal(yamlData, &projectFile)
				utils.CheckForError(err)

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

			// collect sources ...
			projectSources := make([]string, 0)
			projectSources = append(projectSources, args...)
			if !noDefaultSource && len(projectSources) == 0 {
				// add default(s)

				GPM_DEFAULT_PROJECT_SOURCE := strings.TrimSpace(
					os.Getenv("GPM_DEFAULT_PROJECT_SOURCE"),
				)
				if GPM_DEFAULT_PROJECT_SOURCE == "" {
					GPM_DEFAULT_PROJECT_SOURCE = "https://raw.githubusercontent.com/mkloubert/go-package-manager/refs/heads/main/projects.yaml"
				}

				defaultSources := strings.Split(GPM_DEFAULT_PROJECT_SOURCE, "\n")

				projectSources = append(projectSources, defaultSources...)
			}

			// collect data ...
			for _, s := range projectSources {
				source := strings.TrimSpace(s)
				if source == "" {
					continue
				}

				yamlData, err := app.LoadDataFrom(source)
				utils.CheckForError(err)

				importFromYaml(yamlData)
			}

			stdin, err := app.LoadFromInputIfAvailable()
			utils.CheckForError(err)
			if stdin != nil {
				app.Debug("Updating projects from STDIN ...")
				importFromYaml(*stdin)
			}

			// ... finally update projects file
			err = app.UpdateProjectsFile()
			utils.CheckForError(err)
		},
	}

	importProjectsCmd.Flags().BoolVarP(&noDefaultSource, "no-default", "", false, "no default source")
	importProjectsCmd.Flags().BoolVarP(&reset, "reset", "", false, "reset before import entries")

	parentCmd.AddCommand(
		importProjectsCmd,
	)
}
