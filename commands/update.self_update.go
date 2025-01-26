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
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func run_self_update_command(
	app *types.AppContext,
	force bool, noVersionPrint bool, powerShell bool, powerShellBin string, updateScript string, userAgent string,
) {
	app.Debug("Will start self-update ...")

	chromaSettings := app.GetChromaSettings()

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
		customUpdateScript = strings.TrimSpace(
			app.GetEnvValue("GPM_UPDATE_SCRIPT"),
		)
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
			p.Stderr = app.ErrorOut
			p.Stdout = app.Out

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

			chromaSettings.Highlight(string(pwshScript), "powershell")

			fmt.Println()
			fmt.Println()

			reader := bufio.NewReader(app.In)

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
			p.Stderr = app.ErrorOut
			p.Stdout = app.Out

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

			chromaSettings.Highlight(string(bashScript), "shell")

			fmt.Println()
			fmt.Println()

			reader := bufio.NewReader(app.In)

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
}
