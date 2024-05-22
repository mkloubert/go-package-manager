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

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_New_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var noInit bool

	var newCmd = &cobra.Command{
		Use:     "new [project name]",
		Aliases: []string{"n", "nw"},
		Short:   "New project",
		Long:    `Initializes one project as defined in projects.yaml file.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := strings.TrimSpace(args[0])

			gitResource, ok := app.ProjectsFile.Projects[projectName]
			if !ok {
				utils.CloseWithError(fmt.Errorf("project '%v' not found", gitResource))
			}

			var gitDir string
			var outDir string
			if len(args) == 1 {
				outDir = strings.TrimSuffix(path.Base(gitResource), ".git")
				gitDir = path.Join(app.Cwd, outDir, ".git")

				app.RunShellCommandByArgs("git", "clone", gitResource)
			} else {
				outDir = strings.TrimSpace(args[1])
				gitDir = path.Join(app.Cwd, outDir, ".git")

				app.RunShellCommandByArgs("git", "clone", gitResource, "-o", outDir)
			}

			app.Debug(fmt.Sprintf("Removing '%v' folder ...", gitDir))
			err := os.RemoveAll(gitDir)
			utils.CheckForError(err)

			if !noInit {
				p := utils.CreateShellCommandByArgs("git", "init")
				p.Dir = outDir

				app.Debug(fmt.Sprintf("Initializing git in '%v' folder ...", outDir))
				utils.RunCommand(p)
			}
		},
	}

	newCmd.Flags().BoolVarP(&noInit, "no-init", "n", false, "do not initialize git project")

	parentCmd.AddCommand(
		newCmd,
	)
}
