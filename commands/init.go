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

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Init_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var force bool

	var initCmd = &cobra.Command{
		Use:   "init [resource]",
		Short: "Init project or resource",
		Long:  `Inits if no argument is defined the gpm.yaml, file otherwise a resource like a workflow.`,
		Run: func(cmd *cobra.Command, args []string) {
			gpmFilePath := path.Join(app.Cwd, "gpm.yaml")
			gpmDirPath := path.Dir(gpmFilePath)
			gpmFileName := path.Base(gpmFilePath)

			app.Debug(fmt.Sprintf("Will initialize '%v' file in '%v' directory ...", gpmFileName, gpmDirPath))

			isGpmFileExisting, err := utils.IsFileExisting(gpmFilePath)
			utils.CheckForError(err)

			if isGpmFileExisting {
				app.Debug(fmt.Sprintf("Found %v file in '%v'", gpmFileName, gpmDirPath))

				if !force {
					utils.CloseWithError(fmt.Errorf("%v already exists in '%v'", gpmFileName, gpmDirPath))
				}
			}

			app.Debug(fmt.Sprintf("Building content for '%v' file ...", gpmFileName))
			initialGpmFile := types.GpmFile{
				Files: []string{},
				Scripts: map[string]string{
					"test": "go test .",
				},
			}

			app.Debug(fmt.Sprintf("Serializing content of '%v' file to YAML ...", gpmFileName))
			yamlData, err := yaml.Marshal(&initialGpmFile)
			utils.CheckForError(err)

			app.Debug(fmt.Sprintf("Writing content to '%v' file of '%v' directory ...", gpmFileName, gpmDirPath))
			err = os.WriteFile(gpmFilePath, yamlData, constants.DefaultFileMode)
			utils.CheckForError(err)

			fmt.Printf("✅ '%v' has been initialized%v", gpmFileName, fmt.Sprintln())
		},
	}

	initCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "overwrite existing resource")

	parentCmd.AddCommand(
		initCmd,
	)
}
