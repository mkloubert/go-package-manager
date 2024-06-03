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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/constants"
	"github.com/mkloubert/go-package-manager/resources"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func init_show_dependencies_command(parentCmd *cobra.Command, app *types.AppContext) {
	var height string
	var scale float32
	var title string
	var width string

	var showDependenciesCmd = &cobra.Command{
		Use:     "dependencies",
		Aliases: []string{"dependency", "dep"},
		Short:   "Show resource",
		Long:    `Shows a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			// these are values we will use in CSS
			// of the output HTML
			graphWidth := strings.TrimSpace(width)
			graphHeight := strings.TrimSpace(height)

			cmdArgs := []string{"go", "mod", "graph"}

			p := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			p.Dir = app.Cwd

			app.Debug(fmt.Sprintf("Running '%v' ...", strings.Join(cmdArgs, " ")))
			dependencyGraph, err := p.Output()
			utils.CheckForError(err)

			// start Mermaid graph
			mermaidGraph := fmt.Sprintf(`---
title: %v
---
flowchart LR%v`,
				title, fmt.Sprintln())
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

			mermaidCodeJSONBuffer := &bytes.Buffer{}

			encoder := json.NewEncoder(mermaidCodeJSONBuffer)
			encoder.SetEscapeHTML(false)

			app.Debug("Encoding Mermaid graph to JSON ...")
			err = encoder.Encode(mermaidGraph)
			utils.CheckForError(err)

			mermaidCodeJSONString := mermaidCodeJSONBuffer.Bytes()

			app.Debug("Loading HTML template ...")
			templateData, err := resources.Templates.ReadFile("templates/go-dependency-graph.html")
			utils.CheckForError(err)

			app.Debug("Loading Mermaid JS ...")
			mermaidJSData, err := resources.JavaScripts.ReadFile("javascripts/mermaid.min.js")
			utils.CheckForError(err)

			app.Debug("Parsing HTML template ...")
			htmlTemplate, err := template.New("go-dependency-graph.html").Parse(string(templateData))
			utils.CheckForError(err)

			app.Debug("Creating HTML output ...")
			var htmlBuffer bytes.Buffer
			htmlTemplate.Execute(&htmlBuffer, map[string]string{
				"GraphHeight":           graphHeight,
				"GraphScale":            fmt.Sprintf("%f", scale),
				"GraphWidth":            graphWidth,
				"MermaidJS":             string(mermaidJSData),
				"MermaidCodeJSONString": string(mermaidCodeJSONString),
			})
			utils.CheckForError(err)
			defer htmlBuffer.Reset()

			// save final HTML to temp file
			htmlFile, err := os.CreateTemp("", "gpm-dependency-graph-*.html")
			utils.CheckForError(err)

			htmlFileName := htmlFile.Name()

			// write final HTML to temp file
			app.Debug(fmt.Sprintf("Writing HTML to '%v' ...", htmlFileName))
			err = os.WriteFile(htmlFileName, htmlBuffer.Bytes(), constants.DefaultFileMode)
			utils.CheckForError(err)

			// open temp file with default file handler
			app.Debug(fmt.Sprintf("Opening '%v' ...", htmlFileName))
			err = utils.OpenUrl(htmlFileName)
			utils.CheckForError(err)
		},
	}

	showDependenciesCmd.Flags().StringVarP(&height, "height", "", "100%", "custom CSS height of the graph")
	showDependenciesCmd.Flags().Float32VarP(&scale, "scale", "", 1.0, "custom scale of the graph")
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
