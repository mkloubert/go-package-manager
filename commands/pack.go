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
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Pack_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var all bool
	var name string
	var noArch bool
	var noComment bool
	var noOs bool
	var noPostScript bool
	var noPreScript bool
	var noTag bool
	var version string

	var packCmd = &cobra.Command{
		Use:     "pack",
		Aliases: []string{"p", "pk"},
		Short:   "Pack project",
		Long:    `Packs and zips project files`,
		Run: func(cmd *cobra.Command, args []string) {
			if !noPreScript {
				_, ok := app.GpmFile.Scripts[constants.PrePackScriptName]
				if ok {
					app.RunScript(constants.PrePackScriptName)
				}
			}

			var outputFormats []string

			projectName := path.Base(app.Cwd)
			customVersion := strings.TrimSpace(version)

			var latestVersion *ver.Version
			var err error
			if customVersion == "" {
				latestVersion, err = app.GetLatestVersion()

				if err != nil {
					utils.CloseWithError(err)
				}
			} else {
				latestVersion, err = ver.NewVersion(customVersion)
				if err != nil {
					utils.CloseWithError(err)
				}
			}

			if latestVersion == nil {
				latestVersion, _ = ver.NewVersion("0.0.0")
			}

			app.Debug(fmt.Sprintf("Will use version '%v'", latestVersion.String()))

			if all || len(args) > 0 {
				app.Debug(fmt.Sprintf("Running '%v' ...", "go tool dist list"))
				output, err := exec.Command("go", "tool", "dist", "list").Output()
				if err != nil {
					utils.CloseWithError(err)
				}

				// collect all possible targets from output
				var allSupportedArchitecture []string
				for _, l := range strings.Split(string(output), "\n") {
					supportedArchitecture := strings.TrimSpace(l)
					if supportedArchitecture != "" {
						allSupportedArchitecture = append(allSupportedArchitecture, supportedArchitecture)
					}
				}

				if all {
					outputFormats = append(outputFormats, allSupportedArchitecture...)
				} else {
					// take arguments as regex filter
					// and save only unique ones

					matchingArchitectures := map[string]bool{}
					defer func() {
						matchingArchitectures = nil
					}()

					for _, regexFilter := range args {
						for _, supportedArchitecture := range allSupportedArchitecture {
							_, ok := matchingArchitectures[supportedArchitecture]
							if ok {
								continue // already in list
							}

							r := regexp.MustCompile(regexFilter)
							if r.MatchString(supportedArchitecture) {
								matchingArchitectures[supportedArchitecture] = true

								outputFormats = append(outputFormats, supportedArchitecture)
							}
						}
					}
				}
			} else {
				// no arguments => take current OS and CPU

				outputFormats = append(outputFormats, fmt.Sprintf("%v/%v", runtime.GOOS, runtime.GOARCH))
			}

			app.Debug(fmt.Sprintf("Will handle following output formats: %v", outputFormats))

			for _, format := range outputFormats {
				func() {
					parts := strings.SplitN(format, "/", 2)

					goos := parts[0]
					goarch := parts[1]

					app.Debug(fmt.Sprintf("Will pack for '%v/%v' ...", goos, goarch))

					zipFileName := projectName
					if !noTag {
						if latestVersion != nil {
							zipFileName += "-v" + latestVersion.String()
						}
					}
					if !noOs {
						zipFileName += "-" + goos
					}
					if !noArch {
						zipFileName += "-" + goarch
					}
					zipFileName += ".zip"

					zipFilePath := path.Join(app.Cwd, zipFileName)
					app.Debug(fmt.Sprintf("Will pack to '%v' ...", zipFilePath))

					zipFile, err := os.Create(zipFilePath)
					if err != nil {
						utils.CloseWithError(err)
					}
					defer func() {
						app.Debug(fmt.Sprintf("Finish and close zip file '%v' ...", zipFilePath))
						zipFile.Close()
					}()

					app.Debug(fmt.Sprintf("Start packing file(s) to '%v' ...", zipFilePath))
					zipWriter := zip.NewWriter(zipFile)
					defer func() {
						err := zipWriter.Close()
						if err != nil {
							utils.CloseWithError(err)
						}
					}()

					if !noComment {
						err = zipWriter.SetComment("created with gpm - Go Package Manager (https://gpm.kloubert.dev)")
						if err != nil {
							utils.CloseWithError(err)
						}
					}

					err = zipWriter.Flush()
					if err != nil {
						utils.CloseWithError(err)
					}

					executableFilename := strings.TrimSpace(name)
					if executableFilename == "" {
						executableFilename = projectName
					}
					if goos == "windows" {
						executableFilename += constants.WindowsExecutableExt
					}

					app.Debug(
						fmt.Sprintf(
							"Running to '%v' for '%v/%v' ...",
							fmt.Sprintf("go build -o %v .", executableFilename),
							goos, goarch,
						),
					)
					p := utils.CreateShellCommandByArgs("go", "build", "-o", executableFilename, ".")
					p.Dir = app.Cwd
					p.Env = append(p.Env, "GOOS="+goos, "GOARCH="+goarch)

					utils.RunCommand(p)

					filesToPack, err := app.ListFiles()
					if err != nil {
						utils.CloseWithError(err)
					}

					for _, f := range filesToPack {
						func() {
							fileReader, err := os.Open(f)
							if err != nil {
								utils.CloseWithError(err)
							}
							defer fileReader.Close()

							fileInfo, err := os.Stat(f)
							if err != nil {
								utils.CloseWithError(err)
							}

							relPath, err := filepath.Rel(app.Cwd, f)
							if err != nil {
								utils.CloseWithError(err)
							}

							header, err := zip.FileInfoHeader(fileInfo)
							if err != nil {
								utils.CloseWithError(err)
							}
							header.Name = relPath
							header.Modified = fileInfo.ModTime()

							fileWriter, err := zipWriter.CreateHeader(header)
							if err != nil {
								utils.CloseWithError(err)
							}

							app.Debug(fmt.Sprintf("Packing file '%v' into '%v' ...", relPath, zipFilePath))
							io.Copy(fileWriter, fileReader)
						}()
					}
				}()
			}

			if !noPostScript {
				_, ok := app.GpmFile.Scripts[constants.PostPackScriptName]
				if ok {
					app.RunScript(constants.PostPackScriptName)
				}
			}
		},
	}

	packCmd.Flags().BoolVarP(&all, "all", "", false, "compile for all architectures")
	packCmd.Flags().StringVarP(&name, "name", "", "", "custom name of output executable file")
	packCmd.Flags().BoolVarP(&noArch, "no-arch", "", false, "do not add cpu architecture to output filename")
	packCmd.Flags().BoolVarP(&noArch, "no-comment", "", false, "do not add global comment to zip file")
	packCmd.Flags().BoolVarP(&noOs, "no-os", "", false, "do not add operating system to output filename")
	packCmd.Flags().BoolVarP(&noPostScript, "no-post-script", "", false, "do not handle '"+constants.PostPackScriptName+"' script")
	packCmd.Flags().BoolVarP(&noPreScript, "no-pre-script", "", false, "do not handle '"+constants.PrePackScriptName+"' script")
	packCmd.Flags().BoolVarP(&noTag, "no-tag", "", false, "do not add tag to output file")
	packCmd.Flags().StringVarP(&version, "version", "", "", "custom version number")

	parentCmd.AddCommand(
		packCmd,
	)
}
