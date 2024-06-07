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
	"fmt"
	"html"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/resources"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_show_dependencies_command(parentCmd *cobra.Command, app *types.AppContext) {
	var height string
	var infoboxWidth string
	var output string
	var scale float32
	var shouldNotOpen bool
	var sidebarWidth string
	var title string
	var width string

	var showDependenciesCmd = &cobra.Command{
		Use:     "dependencies",
		Aliases: []string{"dependency", "dep", "deps"},
		Short:   "Show resource",
		Long:    `Shows a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			// these are values we will use in CSS
			// of the output HTML
			appName := app.GetName()
			graphWidth := strings.TrimSpace(width)
			graphHeight := strings.TrimSpace(height)
			graphInfoboxWidth := strings.TrimSpace(infoboxWidth)
			graphSidebarWidth := strings.TrimSpace(sidebarWidth)

			cmdArgs := []string{"go", "mod", "graph"}

			p := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			p.Dir = app.Cwd

			app.Debug(fmt.Sprintf("Running '%v' ...", strings.Join(cmdArgs, " ")))
			dependencyGraph, err := p.Output()
			utils.CheckForError(err)

			installedModulesAndVersions := map[string]bool{}

			// start Mermaid graph
			mermaidGraph := fmt.Sprintln("flowchart <<<GraphDirection>>>")
			blockStyles := map[string]string{}
			addBlockStyle := func(h string) {
				bg, fg := utils.GenerateColorsFromString(h)

				blockStyles[h] = fmt.Sprintf(
					"%v fill:#%02x%02x%02x,color:#%02x%02x%02x",
					h,
					bg.R, bg.G, bg.B,
					fg.R, fg.G, fg.B,
				)
			}

			scanner := bufio.NewScanner(strings.NewReader(string(dependencyGraph)))
			for scanner.Scan() {
				// read line and split into
				// parts from space as separator
				line := scanner.Text()
				parts := strings.Fields(line)

				if len(parts) != 2 {
					continue
				}

				// get left and right part
				left := strings.TrimSpace(parts[0])
				right := strings.TrimSpace(parts[1])

				installedModulesAndVersions[left] = true
				installedModulesAndVersions[right] = true

				// setup IDs
				leftBlockId := utils.HashSHA256([]byte(left))
				rightBlockId := utils.HashSHA256([]byte(right))
				app.Debug(fmt.Sprintf("Setup dependency between blocks '%v' and '%v' ...", leftBlockId, rightBlockId))

				// text of left box as JSON string
				leftBlockText, err := utils.SerializeStringToJSON(left)
				utils.CheckForError(err)
				// text of right box as JSON string
				rightBlockText, err := utils.SerializeStringToJSON(right)
				utils.CheckForError(err)

				mermaidGraph += fmt.Sprintf(
					"    %s[%s] --> %s[%s]%s",
					leftBlockId, leftBlockText,
					rightBlockId, rightBlockText,
					"\n",
				)

				addBlockStyle(leftBlockId)
				addBlockStyle(rightBlockId)
			}

			for blockId, style := range blockStyles {
				app.Debug(fmt.Sprintf("Setup style for block '%v' with '%v' ...", blockId, style))
				mermaidGraph += fmt.Sprintf(
					"    style %s%s",
					style,
					"\n",
				)
			}

			err = scanner.Err()
			utils.CheckForError(err)

			// first collect
			installedModuleHtmlList := []interface{}{}
			for k := range installedModulesAndVersions {
				installedModuleHtmlList = append(installedModuleHtmlList, k)
			}
			sort.Slice(installedModuleHtmlList, func(x, y int) bool {
				strX := installedModuleHtmlList[x].(string)
				strY := installedModuleHtmlList[y].(string)

				return strings.ToLower(strX) < strings.ToLower(strY)
			})
			for i := range installedModuleHtmlList {
				nameAndVersion := strings.TrimSpace(
					installedModuleHtmlList[i].(string),
				)

				name := nameAndVersion
				version := ""

				sepIndex := strings.Index(nameAndVersion, "@")
				if sepIndex > -1 {
					version = strings.TrimSpace(name[sepIndex+1:])
					name = strings.TrimSpace(name[0:sepIndex])
				}

				moduleLink := ""
				if name != "" {
					moduleLink = fmt.Sprintf("https://%v", name)
				}

				installedModuleHtmlList[i] = map[string]interface{}{
					"EscapedName":    html.EscapeString(name),
					"EscapedVersion": html.EscapeString(version),
					"EscapedVersionAndName": html.EscapeString(
						fmt.Sprintf("%v@%v", name, version),
					),
					"Id":      nameAndVersion,
					"Link":    moduleLink,
					"Name":    name,
					"Version": version,
				}
			}

			mermaidJSData, err := resources.JavaScripts.ReadFile("javascripts/mermaid@10.9.1.min.js")
			utils.CheckForError(err)

			htmlFileName := strings.TrimSpace(app.GetFullPathOrDefault(output, ""))
			if htmlFileName == "" {
				// save final HTML to temp file instead
				htmlFile, err := os.CreateTemp("", "gpm-dependency-graph-*.html")
				utils.CheckForError(err)

				htmlFileName = htmlFile.Name()
			}

			// write final HTML to file
			app.Debug(fmt.Sprintf("Writing HTML to '%v' ...", htmlFileName))
			r := &types.ReactRenderer{
				ContentClass: "h-screen flex",
				ExternalModules: map[string]types.ReactRendererExternalModule{
					"mermaid": {
						Type: "module",
						Url:  utils.ToDataUri(mermaidJSData, "text/javascript"),
					},
				},
				Vars: map[string]interface{}{
					"appName":              appName,
					"graphDirection":       "LR",
					"graphHeight":          graphHeight,
					"graphScalePercentage": scale * 100.0,
					"graphWidth":           graphWidth,
					"infoboxWidth":         graphInfoboxWidth,
					"mermaidGraph":         mermaidGraph,
					"moduleList":           installedModuleHtmlList,
					"sidebarWidth":         graphSidebarWidth,
				},
			}
			// JSX template
			err = r.AddJsxTemplate("go-dependency-graph")
			utils.CheckForError(err)
			// Mermaid
			err = r.AddJavascriptTemplate("mermaid@10.9.1.min")
			utils.CheckForError(err)
			// Tailwind
			err = r.AddJavascriptTemplate("tailwindcss@3.4.3")
			utils.CheckForError(err)
			htmlData, err := r.Render(path.Base(htmlFileName))
			utils.CheckForError(err)
			os.WriteFile(htmlFileName, htmlData, constants.DefaultFileMode)

			if !shouldNotOpen {
				// open file with default file handler
				app.Debug(fmt.Sprintf("Opening '%v' ...", htmlFileName))
				err = utils.OpenUrl(htmlFileName)
				utils.CheckForError(err)
			}
		},
	}

	showDependenciesCmd.Flags().BoolVarP(&shouldNotOpen, "do-not-open", "", false, "do not open file after created")
	showDependenciesCmd.Flags().StringVarP(&height, "height", "", "100%", "custom CSS height of the graph")
	showDependenciesCmd.Flags().StringVarP(&infoboxWidth, "infobox-width", "", "320px", "custom width of the infobox")
	showDependenciesCmd.Flags().StringVarP(&output, "output", "o", "", "custom output file")
	showDependenciesCmd.Flags().Float32VarP(&scale, "scale", "", 3.0, "custom scale of the graph")
	showDependenciesCmd.Flags().StringVarP(&sidebarWidth, "sidebar-width", "", "420px", "custom width of the sidebar")
	showDependenciesCmd.Flags().StringVarP(&title, "title", "", "GPM Dependency Graph", "custom title of the graph")
	showDependenciesCmd.Flags().StringVarP(&width, "width", "", "100%", "custom CSS width of the graph")

	parentCmd.AddCommand(
		showDependenciesCmd,
	)
}

func Init_Show_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var showCmd = &cobra.Command{
		Use:     "show [resource]",
		Aliases: []string{"shw", "sh"},
		Short:   "Show resource",
		Long:    `Shows a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_show_dependencies_command(showCmd, app)

	parentCmd.AddCommand(
		showCmd,
	)
}
