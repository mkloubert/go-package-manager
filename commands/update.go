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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/alecthomas/chroma/quick"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Update_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var force bool
	var noCleanup bool
	var noVersionPrint bool
	var powerShell bool
	var powerShellBin string
	var selfUpdate bool
	var updateScript string
	var userAgent string

	var updateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"upd"},
		Short:   "Update dependencies",
		Long:    `Updates all dependencies in this project.`,
		Run: func(cmd *cobra.Command, args []string) {
			if selfUpdate {
				app.Debug("Will start self-update ...")

				consoleFormatter := utils.GetBestChromaFormatterName()
				consoleStyle := utils.GetBestChromaStyleName()

				customUserAgent := strings.TrimSpace(userAgent)
				if customUserAgent == "" {
					customUserAgent = browser.Chrome()
				}

				customPowerShellBin := strings.TrimSpace(powerShellBin)
				if customPowerShellBin == "" {
					customPowerShellBin = "powershell"
				}

				customUpdateScript := strings.TrimSpace(updateScript)
				if customUpdateScript == "" {
					customUpdateScript = strings.TrimSpace(os.Getenv("GPM_UPDATE_SCRIPT"))
				}

				downloadScript := func(url string) ([]byte, error) {
					app.Debug(fmt.Sprintf("Download from '%s' ...", url))
					app.Debug(fmt.Sprintf("User agent: %s", customUserAgent))

					req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
					if err != nil {
						return []byte{}, err
					}

					req.Header.Set("User-Agent", customUserAgent)

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						return []byte{}, err
					}
					defer resp.Body.Close()

					if resp.StatusCode != 200 {
						return []byte{}, fmt.Errorf("unexpected response: %v", resp.StatusCode)
					}

					responseData, err := io.ReadAll(resp.Body)

					return responseData, err
				}

				showNewVersion := func() {
					if noVersionPrint {
						return
					}

					app.RunShellCommandByArgs("gpm", "--version")
				}

				if powerShell || utils.IsWindows() {
					// PowerShell
					app.Debug(fmt.Sprintf("Will use PowerShell '%s' ...", powerShellBin))

					scriptUrl := customUpdateScript
					if scriptUrl == "" {
						scriptUrl = "https://raw.githubusercontent.com/mkloubert/go-package-manager/main/sh.kloubert.dev/gpm.ps1"
					} else {
						su, err := utils.ToUrlForOpenHandler(scriptUrl)
						utils.CheckForError(err)

						scriptUrl = su
					}

					pwshScript, err := downloadScript(scriptUrl)
					utils.CheckForError(err)

					executeScript := func() {
						p := exec.Command(customPowerShellBin, "-NoProfile", "-Command", "-")
						p.Dir = app.Cwd
						p.Stderr = os.Stderr
						p.Stdout = os.Stdout

						stdinPipe, err := p.StdinPipe()
						utils.CheckForError(err)

						err = p.Start()
						utils.CheckForError(err)

						go func() {
							defer stdinPipe.Close()
							stdinPipe.Write([]byte(pwshScript))
						}()

						err = p.Wait()
						utils.CheckForError(err)

						showNewVersion()
						os.Exit(0)
					}

					if force {
						executeScript()
					} else {
						// ask the user first

						err = quick.Highlight(os.Stdout, string(pwshScript), "powershell", consoleFormatter, consoleStyle)
						if err != nil {
							fmt.Print(string(pwshScript))
						}

						fmt.Println()
						fmt.Println()

						reader := bufio.NewReader(os.Stdin)

						for {
							fmt.Print("Do you really want to run this PowerShell script (Y/n)? ")
							userInput, _ := reader.ReadString('\n')
							userInput = strings.TrimSpace(strings.ToLower(userInput))

							switch userInput {
							case "", "y", "yes":
								executeScript()
							case "n", "no":
								os.Exit(0)
							}
						}
					}
				} else if utils.IsPOSIXLikeOS() {
					// if POSIX-like => sh
					app.Debug("Will use UNIX shell ...")

					scriptUrl := customUpdateScript
					if scriptUrl == "" {
						scriptUrl = "https://raw.githubusercontent.com/mkloubert/go-package-manager/main/sh.kloubert.dev/gpm.sh"
					} else {
						su, err := utils.ToUrlForOpenHandler(scriptUrl)
						utils.CheckForError(err)

						scriptUrl = su
					}

					bashScript, err := downloadScript(scriptUrl)
					utils.CheckForError(err)

					executeScript := func() {
						p := exec.Command("sh")
						p.Dir = app.Cwd
						p.Stderr = os.Stderr
						p.Stdout = os.Stdout

						stdinPipe, err := p.StdinPipe()
						utils.CheckForError(err)

						err = p.Start()
						utils.CheckForError(err)

						go func() {
							defer stdinPipe.Close()
							stdinPipe.Write([]byte(bashScript))
						}()

						err = p.Wait()
						utils.CheckForError(err)

						showNewVersion()
						os.Exit(0)
					}

					if force {
						executeScript()
					} else {
						// ask the user first

						err = quick.Highlight(os.Stdout, string(bashScript), "shell", consoleFormatter, consoleStyle)
						if err != nil {
							fmt.Print(string(bashScript))
						}

						fmt.Println()
						fmt.Println()

						reader := bufio.NewReader(os.Stdin)

						for {
							fmt.Print("Do you really want to run this bash script (Y/n)? ")
							userInput, _ := reader.ReadString('\n')
							userInput = strings.TrimSpace(strings.ToLower(userInput))

							switch userInput {
							case "", "y", "yes":
								executeScript()
							case "n", "no":
								os.Exit(0)
							}
						}
					}
				} else {
					utils.CheckForError(fmt.Errorf("self-update for %s/%s is not supported yet", runtime.GOOS, runtime.GOARCH))
				}
			} else {
				app.Debug("Will start project dependencies ...")

				app.RunShellCommandByArgs("go", "get", "-u", "./...")

				if !noCleanup {
					app.RunShellCommandByArgs("go", "mod", "tidy")
				}
			}
		},
	}

	updateCmd.Flags().BoolVarP(&force, "force", "", false, "force self-update")
	updateCmd.Flags().BoolVarP(&noCleanup, "no-cleanup", "", false, "do not cleanup go.mod and go.sum")
	updateCmd.Flags().BoolVarP(&noVersionPrint, "no-version-print", "", false, "do not print new version after successful update")
	updateCmd.Flags().BoolVarP(&powerShell, "powershell", "", false, "force execution of PowerShell script")
	updateCmd.Flags().StringVarP(&powerShellBin, "powershell-bin", "", "", "custom binary of the PowerShell")
	updateCmd.Flags().BoolVarP(&selfUpdate, "self", "", false, "update this binary instead")
	updateCmd.Flags().StringVarP(&updateScript, "update-script", "", "", "custom URL to update script")
	updateCmd.Flags().StringVarP(&userAgent, "user-agent", "", "", "custom string for user agent")

	parentCmd.AddCommand(
		updateCmd,
	)
}
