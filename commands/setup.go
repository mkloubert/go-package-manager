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
	"runtime"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_setup_updater_command(parentCmd *cobra.Command, app *types.AppContext) {
	var setupUpdaterCmd = &cobra.Command{
		Use:     "updater",
		Aliases: []string{"u", "up", "upt"},
		Short:   "Setup updater",
		Long:    `Sets up a an updater script for this tool.`,
		Run: func(cmd *cobra.Command, args []string) {
			binPath, err := app.GetBinFolderPath()
			if err != nil {
				utils.CloseWithError(err)
			}

			consoleFormatter := utils.GetBestChromaFormatterName()
			consoleStyle := utils.GetBestChromaStyleName()

			goos := runtime.GOOS
			goarch := runtime.GOARCH

			var createScript func()

			if utils.IsWindows() {
				// not supported
				createScript = nil
			} else {
				bashScriptFilePath := path.Join(binPath, "gpm-update")

				var sha256Command string
				if goos == "darwin" {
					sha256Command = "shasum -a 256 gpm.tar.gz.sha256"
				} else {
					sha256Command = "sha256sum -c gpm.tar.gz.sha256"
				}

				createScript = func() {
					bashScript := fmt.Sprintf(`#!/bin/bash

handle_error() {
    echo "Error: $1"
    exit 1
}

echo "gpm-update"
echo ""

echo "Finding download URL and SHA256 URL ..."
latest_release_info=$(wget -qO- https://api.github.com/repos/mkloubert/go-package-manager/releases/latest) || handle_error "Could not fetch release infos"
download_url=$(echo "$latest_release_info" | jq -r '.assets[].browser_download_url | select(contains("gpm") and contains("%v") and contains("%v") and (. | tostring | contains("sha256") | not))') || handle_error "Could not parse download URL"
sha256_url=$(echo "$latest_release_info" | jq -r '.assets[].browser_download_url | select(contains("gpm") and contains("%v") and contains("%v") and contains("sha256"))') || handle_error "Could not parse SHA256 URL"

if [ -z "$download_url" ]; then
  handle_error "No valid download URL found"
fi

if [ -z "$sha256_url" ]; then
  handle_error "No valid SHA256 URL found"
fi

echo "Downloading tarball from '$download_url'..."
wget -q "$download_url" -O gpm.tar.gz || handle_error "Failed to download tarball"

echo "Downloading SHA256 file from '$sha256_url'..."
wget -q "$sha256_url" -O gpm.tar.gz.sha256 || handle_error "Failed to download SHA256 file"

echo "Verifying tarball ..."
%v || handle_error "SHA256 verification failed"

echo "Extracting binary ..."
tar -xzOf gpm.tar.gz gpm > gpm || handle_error "Could not extract 'gpm' binary"

echo "Installing 'gpm' to /usr/local/bin ..."
sudo mv gpm /usr/local/bin/gpm || handle_error "Could not move 'gpm' to '/usr/local/bin'"
sudo chmod +x /usr/local/bin/gpm || handle_error "Could not update permissions of 'gpm' binary"

echo "Cleaning up ..."
rm gpm.tar.gz gpm.tar.gz.sha256 || handle_error "Cleanups failed"

echo "'gpm' successfully installed or updated 👍"
`,
						goos, goarch,
						goos, goarch,
						sha256Command,
					)

					app.Debug(fmt.Sprintf("Writing bash script to '%v' ...", bashScriptFilePath))
					os.WriteFile(bashScriptFilePath, []byte(bashScript), constants.DefaultFileMode)

					fmt.Printf(
						"Wrote following script to '%v':%v%v",
						color.New(color.FgWhite, color.Bold).Sprint(bashScriptFilePath),
						fmt.Sprintln(), fmt.Sprintln(),
					)

					err := quick.Highlight(os.Stdout, bashScript, "shell", consoleFormatter, consoleStyle)
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

	init_setup_updater_command(setupCmd, app)

	parentCmd.AddCommand(
		setupCmd,
	)
}
