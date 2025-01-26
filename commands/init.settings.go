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

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_init_settings_command(parentCmd *cobra.Command, app *types.AppContext) {
	var force bool

	var initSettingsCmd = &cobra.Command{
		Use:     "settings",
		Aliases: []string{"s"},
		Short:   "Init global settings",
		Long:    `Inits global settings.yaml file.`,
		Run: func(cmd *cobra.Command, args []string) {
			settingsFile, err := app.GetDefaultSettingsFilePath()
			utils.CheckForError(err)

			doesExist, err := utils.IsFileExisting(settingsFile)
			utils.CheckForError(err)

			if doesExist && !force {
				utils.CheckForError(fmt.Errorf("file '%s' already exists", settingsFile))
			}

			initialSettings := map[string]interface{}{}

			yamlData, err := yaml.Marshal(&initialSettings)
			utils.CheckForError(err)

			err = os.WriteFile(settingsFile, yamlData, 0664)
			utils.CheckForError(err)

			if doesExist {
				app.Write([]byte(fmt.Sprintf("Re-Initialized settings in '%v'", settingsFile)))
			} else {
				app.Write([]byte(fmt.Sprintf("Initialized new settings in '%v'", settingsFile)))
			}
			app.Write([]byte(fmt.Sprintln()))
		},
	}

	initSettingsCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "overwrite existing file")

	parentCmd.AddCommand(
		initSettingsCmd,
	)
}
