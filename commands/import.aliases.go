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

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_import_aliases_command(parentCmd *cobra.Command, app *types.AppContext) {
	var noDefaultSource bool
	var reset bool

	var importAliasCmd = &cobra.Command{
		Use:     "aliases [source]",
		Aliases: []string{"a", "al", "alias"},
		Short:   "Import alias",
		Long:    `Downloads one or more alias file from external resources and merge them with local one.`,
		Run: func(cmd *cobra.Command, args []string) {
			importFromYaml := func(yamlData []byte) {
				var aliasFile types.AliasesFile
				err := yaml.Unmarshal(yamlData, &aliasFile)
				utils.CheckForError(err)

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

			// collect sources ...
			aliasSources := make([]string, 0)
			aliasSources = append(aliasSources, args...)
			if !noDefaultSource && len(aliasSources) == 0 {
				// add default

				GPM_DEFAULT_ALIAS_SOURCE := strings.TrimSpace(
					app.GetEnvValue("GPM_DEFAULT_ALIAS_SOURCE"),
				)
				if GPM_DEFAULT_ALIAS_SOURCE == "" {
					GPM_DEFAULT_ALIAS_SOURCE = constants.DefaultAliasSource
				}

				defaultSources := strings.Split(GPM_DEFAULT_ALIAS_SOURCE, "\n")

				aliasSources = append(aliasSources, defaultSources...)
			}

			// collect data ...
			for _, s := range aliasSources {
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
				app.Debug("Updating aliases from STDIN ...")
				importFromYaml(*stdin)
			}

			// ... finally update aliases file
			err = app.UpdateAliasesFile()
			utils.CheckForError(err)
		},
	}

	importAliasCmd.Flags().BoolVarP(&noDefaultSource, "no-default", "", false, "no default source")
	importAliasCmd.Flags().BoolVarP(&reset, "reset", "", false, "reset before import entries")

	parentCmd.AddCommand(
		importAliasCmd,
	)
}
