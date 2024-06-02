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
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/resources"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_setup_git_command(parentCmd *cobra.Command, app *types.AppContext) {
	var force bool
	var local bool

	var setupUpdaterCmd = &cobra.Command{
		Use:     "git [name] [email]",
		Aliases: []string{"g", "gt"},
		Args:    cobra.MinimumNArgs(2),
		Short:   "Setup git",
		Long:    `Sets up git with minimum and required settings like name and email.`,
		Run: func(cmd *cobra.Command, args []string) {
			name := strings.TrimSpace(args[0])
			email := strings.TrimSpace(strings.ToLower(args[1]))

			if !force {
				const emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
				emailRegex := regexp.MustCompile(emailRegexPattern)

				if name == "" {
					utils.CloseWithError(fmt.Errorf("no name defined"))
				}

				if !emailRegex.MatchString(email) {
					utils.CloseWithError(fmt.Errorf("no valid email address defined"))
				}
			}

			app.Debug(fmt.Sprintf("Setting up user name as '%v' ...", name))
			if local {
				app.RunShellCommandByArgs("git", "config", "user.name", name)
			} else {
				app.RunShellCommandByArgs("git", "config", "--global", "user.name", name)
			}

			app.Debug(fmt.Sprintf("Setting up user email as '%v' ...", email))
			if local {
				app.RunShellCommandByArgs("git", "config", "user.email", email)
			} else {
				app.RunShellCommandByArgs("git", "config", "--global", "user.email", email)
			}
		},
	}

	parentCmd.Flags().BoolVarP(&force, "force", "", false, "no checks")
	parentCmd.Flags().BoolVarP(&local, "local", "", false, "no --global flag")

	parentCmd.AddCommand(
		setupUpdaterCmd,
	)
}

func init_setup_updater_command(parentCmd *cobra.Command, app *types.AppContext) {
	var installPath string

	var setupUpdaterCmd = &cobra.Command{
		Use:     "updater",
		Aliases: []string{"u", "up", "upt"},
		Short:   "Setup updater",
		Long:    `Sets up a an updater script for this tool.`,
		Run: func(cmd *cobra.Command, args []string) {
			binPath, err := app.GetBinFolderPath()
			utils.CheckForError(err)

			consoleFormatter := utils.GetBestChromaFormatterName()
			consoleStyle := utils.GetBestChromaStyleName()

			goos := runtime.GOOS
			goarch := runtime.GOARCH

			var createScript func()

			if utils.IsWindows() {
				// not supported
				createScript = nil
			} else {
				targetFolder := strings.TrimSpace(installPath)
				if targetFolder == "" {
					targetFolder = "/usr/local/bin"
				}

				bashScriptFilePath := path.Join(binPath, "gpm-update")

				var sha256Command string
				if goos == "darwin" {
					sha256Command = "shasum -a 256 gpm.tar.gz.sha256"
				} else {
					sha256Command = "sha256sum -c gpm.tar.gz.sha256"
				}

				createScript = func() {
					templateData, err := resources.Templates.ReadFile("templates/gpm-update.sh")
					utils.CheckForError(err)

					template, err := template.New("gpm-update.sh").Parse(string(templateData))
					utils.CheckForError(err)

					var bashScriptBuffer bytes.Buffer
					template.Execute(&bashScriptBuffer, map[string]string{
						"GOOS":          goos,
						"GOARCH":        goarch,
						"SHA256Command": sha256Command,
						"TargetFolder":  targetFolder,
					})
					utils.CheckForError(err)
					defer bashScriptBuffer.Reset()

					bashScript := bashScriptBuffer.String()

					app.Debug(fmt.Sprintf("Writing bash script to '%v' ...", bashScriptFilePath))
					os.WriteFile(bashScriptFilePath, []byte(bashScript), constants.DefaultFileMode)

					fmt.Printf(
						"Wrote following script to '%v':%v%v",
						color.New(color.FgWhite, color.Bold).Sprint(bashScriptFilePath),
						fmt.Sprintln(), fmt.Sprintln(),
					)

					err = quick.Highlight(os.Stdout, bashScript, "shell", consoleFormatter, consoleStyle)
					if err != nil {
						fmt.Print(bashScript)
					}
				}
			}

			if createScript != nil {
				createScript()
			} else {
				utils.CloseWithError(fmt.Errorf("system %v/%v is not supported yet", goos, goarch))
			}
		},
	}

	setupUpdaterCmd.Flags().StringVarP(&installPath, "install-path", "", "", "custom target folder")

	parentCmd.AddCommand(
		setupUpdaterCmd,
	)
}

func Init_Setup_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var setupCmd = &cobra.Command{
		Use:     "setup [resource]",
		Aliases: []string{"sup", "stup"},
		Short:   "Setup resource",
		Long:    `Sets up a resource like an updater script for this tool.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_setup_git_command(setupCmd, app)
	init_setup_updater_command(setupCmd, app)

	parentCmd.AddCommand(
		setupCmd,
	)
}
