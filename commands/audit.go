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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Audit_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var auditCmd = &cobra.Command{
		Use:   "audit",
		Short: "Audit modules",
		Long:  `Audits modules of the current project using API of osv.dev`,
		Run: func(cmd *cobra.Command, args []string) {
			tHeadColor := color.New(color.FgWhite, color.Bold).SprintFunc()

			modules, err := app.GetGoModules()
			if err != nil {
				utils.CloseWithError(err)
			}

			for i, m := range modules {
				func() {
					modulePath := m.Path
					if modulePath == nil {
						app.Debug(fmt.Sprintf("Skipping module #%v which has no path defined", i))
						return
					}

					moduleVersion := m.Version
					if moduleVersion == nil {
						app.Debug(fmt.Sprintf("Skipping module #%v (%v) which has no version defined", i, *modulePath))
						return
					}

					coloredModuleName := color.New(color.FgWhite, color.Bold).Sprint(*modulePath)
					coloredModuleVersion := color.New(color.FgWhite, color.Bold).Sprint(*moduleVersion)

					s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
					s.Start()
					s.Suffix = fmt.Sprintf(
						" %v (%v)",
						coloredModuleName, coloredModuleVersion,
					)
					s.Color("white")

					stopByError := func(err error) {
						s.Stop()

						fmt.Printf(
							"❌ %v (%v): %v%v",
							coloredModuleName, coloredModuleVersion,
							color.New(color.FgYellow, color.BgRed, color.Bold).Sprint(err),
							fmt.Sprintln(),
						)
					}

					// prepare request to osv.dev API
					url := "https://api.osv.dev/v1/query"
					body := map[string]interface{}{
						"version": *moduleVersion,
						"package": map[string]interface{}{
							"name":      *modulePath,
							"ecosystem": "Go",
						},
					}

					// serialize body
					jsonData, err := json.Marshal(&body)
					if err != nil {
						stopByError(fmt.Errorf("could not serialize request body: %v", err))
						return
					}

					// start the request
					req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
					if err != nil {
						stopByError(fmt.Errorf("could not prepare POST request to '%v': %v", url, err))
						return
					}

					// setup ...
					req.Header.Set("Content-Type", "application/json")
					// ... and finally send the JSON data
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						stopByError(fmt.Errorf("could not do POST request to '%v': %v", url, err))
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != 200 {
						stopByError(fmt.Errorf("unexpected response from '%v': %v", url, resp.StatusCode))
						return
					}

					// load the response
					osvResponseData, err := io.ReadAll(resp.Body)
					if err != nil {
						stopByError(fmt.Errorf("could not do load response from '%v': %v", url, err))
						return
					}

					// parse the response
					var osvResponse types.OsvDevResponse
					err = json.Unmarshal(osvResponseData, &osvResponse)
					if err != nil {
						stopByError(fmt.Errorf("could not parse response from '%v': %v", url, err))
						return
					}

					s.Stop()

					printNoIssueInfo := func() {
						fmt.Printf(
							"✅ %v (%v)%v",
							coloredModuleName, coloredModuleVersion,
							fmt.Sprintln(),
						)
					}

					if osvResponse.Vulnerabilities == nil {
						printNoIssueInfo()
						return
					}

					// create copy of array in osvResponse.Vulnerabilities ...
					vulnerabilities := []types.OsvDevResponseVulnerabilityItem{}
					vulnerabilities = append(vulnerabilities, *osvResponse.Vulnerabilities...)
					vulnerabilitiesCount := len(vulnerabilities)

					if vulnerabilitiesCount == 0 {
						printNoIssueInfo()
						return
					}

					// sort by severity (desc)
					sort.Slice(vulnerabilities, func(x int, y int) bool {
						vulnX := vulnerabilities[x]
						vulnY := vulnerabilities[y]

						_, compX := vulnX.GetSeverityDisplayValues()
						_, compY := vulnY.GetSeverityDisplayValues()
						if compX != compY {
							return compX > compY
						}

						return false
					})

					fmt.Printf(
						"⚠️ %v (%v):%v",
						coloredModuleName, coloredModuleVersion,
						fmt.Sprintln(),
					)

					var tBuffer bytes.Buffer

					// output in buffer first
					t := table.NewWriter()
					t.SetOutputMirror(&tBuffer)

					// header
					t.AppendHeader(table.Row{tHeadColor("#"), tHeadColor("Severity"), tHeadColor("ID"), tHeadColor("Summary")})
					for vi, v := range vulnerabilities {
						if vi > 0 {
							// add separator at top
							t.AppendSeparator()
						}

						severity, _ := v.GetSeverityDisplayValues()

						// output basic issue info
						t.AppendRow(table.Row{vi + 1, severity, v.Id, v.Summary})

						if v.References != nil {
							// add references

							references := []types.OsvDevResponseVulnerabilityItemReference{}
							references = append(references, *v.References...)

							// sort references by type, then by URL
							sort.Slice(references, func(x int, y int) bool {
								refX := references[x]
								refY := references[y]

								typeX := strings.TrimSpace(strings.ToLower(refX.Type))
								typeY := strings.TrimSpace(strings.ToLower(refY.Type))
								if typeX != typeY {
									return typeX < typeY
								}

								urlX := strings.TrimSpace(strings.ToLower(refX.Url))
								urlY := strings.TrimSpace(strings.ToLower(refY.Url))

								return urlX < urlY
							})

							if len(references) > 0 {
								// build reference list

								t.AppendSeparator()

								for ri, r := range references {
									refCol := ""
									if ri == 0 {
										refCol = tHeadColor("References:")
									}

									t.AppendRow(table.Row{"", refCol, r.Type, r.Url})
								}

								t.AppendSeparator()
							}
						}
					}

					// render final table
					t.Render()

					// output final table with prefix
					prefix := "  "
					output := tBuffer.String()
					for _, line := range strings.Split(output, fmt.Sprintln()) {
						if len(line) > 0 {
							fmt.Printf("%v%s%v", prefix, line, fmt.Sprintln())
						}
					}
				}()
			}
		},
	}

	parentCmd.AddCommand(
		auditCmd,
	)
}
