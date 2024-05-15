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

func Init_Make_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var executable string
	var name string
	var noAutoExt bool

	var makeCmd = &cobra.Command{
		Use:     "make [git resource]",
		Aliases: []string{"m", "mk"},
		Short:   "Make project",
		Long:    `Downloads a Git repository and build it.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, gitResource := range args {
				func() {
					// get `$HOME/.gpm/bin` folder
					binPath, err := app.EnsureBinFolder()
					if err != nil {
						utils.CloseWithError(err)
					}

					// current executable path
					selfPath, err := os.Executable()
					if err != nil {
						utils.CloseWithError(err)
					}

					// get project name from git resource
					projectName := strings.TrimSuffix(
						path.Base(gitResource), ".git",
					)

					// create temp folder where to clone
					// git repo to
					tempDir, err := os.MkdirTemp("", "*-"+projectName)
					if err != nil {
						utils.CloseWithError(err)
					}
					defer func() {
						app.Debug(fmt.Sprintf("Removing folder '%v' ...", tempDir))
						os.RemoveAll(tempDir)
					}()

					tempDirName := path.Base(tempDir)

					// clone repo
					app.Debug(fmt.Sprintf("Cloning '%v' to '%v' ...", gitResource, tempDir))
					app.RunShellCommandByArgs("git", "clone", "--depth", "1", gitResource, tempDir)

					buildArgs := []string{selfPath, "build"}
					buildArgs = append(buildArgs, args[1:]...)

					// run `gpm build` in cloned repository
					p := utils.CreateShellCommandByArgs(buildArgs[0], buildArgs[1:]...)
					p.Dir = tempDir

					app.Debug(fmt.Sprintf("Running separate build in '%v' ...", tempDir))
					utils.RunCommand(p)

					// define possible executable file names
					outExecutableFilenameByProject := strings.TrimSpace(name)
					outExecutableFilenameByTempDir := tempDirName
					if outExecutableFilenameByProject == "" {
						outExecutableFilenameByProject = projectName
					}
					if !noAutoExt && utils.IsWindows() {
						// Windows uses .exe

						outExecutableFilenameByProject += ".exe"
						outExecutableFilenameByTempDir += ".exe"
					}

					outExecutableFilePathByProject := path.Join(tempDir, outExecutableFilenameByProject)
					outExecutableFilePathByTempDir := path.Join(tempDir, outExecutableFilenameByTempDir)

					isOutExecutableFileByProjectExisting, err := utils.IsFileExisting(outExecutableFilePathByProject)
					if err != nil {
						utils.CloseWithError(err)
					}

					var buildExecutableFilePath string

					if isOutExecutableFileByProjectExisting {
						// found executable file in repo
						buildExecutableFilePath = outExecutableFilePathByProject
					} else {
						// try to find executable by name of temp directory instead

						isOutExecutableFileByTempDirNameExisting, err := utils.IsFileExisting(outExecutableFilenameByTempDir)
						if err != nil {
							utils.CloseWithError(err)
						}

						if isOutExecutableFileByTempDirNameExisting {
							buildExecutableFilePath = outExecutableFilePathByTempDir
						} else {
							utils.CloseWithError(fmt.Errorf("no matching executable file found. use --executable flag to specify"))
						}
					}

					executableNameInBinFolder := strings.TrimSpace(executable)
					if executableNameInBinFolder == "" {
						// use project name as default for the
						// name of the final executable file in
						// ${HOME}/.gpm/bin folder
						executableNameInBinFolder = projectName
					}

					executableFileInBinFolder := path.Join(binPath, executableNameInBinFolder)

					isExecutableFileInBinFolderExisting, err := utils.IsFileExisting(executableFileInBinFolder)
					if err == nil {
						if isExecutableFileInBinFolderExisting {
							app.Debug(fmt.Sprintf("Removing executable '%v' ...", executableFileInBinFolder))
							os.Remove(executableFileInBinFolder)
						}
					} else {
						utils.CloseWithError(err)
					}

					// move build executable to ${HOME}/.gpm/bin folder
					app.Debug(fmt.Sprintf("Moving build executable '%v' to '%v' ...", buildExecutableFilePath, executableFileInBinFolder))
					err = os.Rename(buildExecutableFilePath, executableFileInBinFolder)
					if err != nil {
						utils.CloseWithError(err)
					}

					// make file in ${HOME}/.gpm/bin folder executable
					app.Debug(fmt.Sprintf("Setting up permissions for '%v' executable ...", executableFileInBinFolder))
					err = os.Chmod(executableFileInBinFolder, constants.DefaultDirMode)
					if err != nil {
						utils.CloseWithError(err)
					}
				}()
			}
		},
	}

	makeCmd.Flags().StringVarP(&name, "name", "", "", "custom name of output executable file")
	makeCmd.Flags().BoolVarP(&noAutoExt, "no-auto-extension", "", false, "do not add file extension automatically")
	makeCmd.Flags().StringVarP(&name, "executable", "", "", "custom name of executable file in bin folder")

	parentCmd.AddCommand(
		makeCmd,
	)
}
