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
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

type GoModFile struct {
	Module  GoModFileModule        `json:"Module,omitempty"`
	Go      string                 `json:"Go,omitempty"`
	Require []GoModFileRequireItem `json:"Require,omitempty"`
}

type GoModFileModule struct {
	Path string `json:"Path,omitempty"`
}

type GoModFileRequireItem struct {
	Path     string `json:"Path,omitempty"`
	Indirect *bool  `json:"Indirect,omitempty"`
	Version  string `json:"Version,omitempty"`
}

type GoProxyModuleInfo struct {
	Time    string `json:"Time,omitempty"`
	Version string `json:"Version,omitempty"`
}

func Init_Doctor_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var doctorCmd = &cobra.Command{
		Use:   "doctor",
		Short: "Checks preconditions and audits",
		Long:  `Runs precondition checks and audits for the current project.`,
		Run: func(cmd *cobra.Command, args []string) {
			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			tHeadColor := color.New(color.FgWhite, color.Bold).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()

			goModFile := app.GetFullPathOrDefault("go.mod", "")
			if goModFile != "" {
				doesGoModFileExist, err := utils.IsFileExisting(goModFile)
				if err == nil {
					if doesGoModFileExist {
						app.Debug(fmt.Sprintf("Found '%s' file", goModFile))

						fmt.Println("Checking go.mod file ...")

						s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
						s.Prefix = "\t["
						s.Suffix = "] Validating file ..."
						s.Start()

						p := exec.Command("go", "mod", "edit", "-json")
						p.Dir = app.Cwd
						p.Stderr = nil
						p.Stdin = nil
						p.Stdout = nil
						output, err := p.Output()

						s.Stop()

						if err == nil {
							var goMod GoModFile
							err := json.Unmarshal(output, &goMod)
							if err == nil {
								fmt.Printf("\t[%s] Module: %s%s", green("✓"), goMod.Module.Path, fmt.Sprintln())

								goVersion, err := version.NewVersion(strings.TrimSpace(goMod.Go))
								if err == nil {
									fmt.Printf("\t[%s] Go Version: %s%s", green("✓"), goVersion.String(), fmt.Sprintln())
								} else {
									fmt.Printf("\t[%s] Invalid Go version '%s': %s%s", red("!"), goMod.Go, err.Error(), fmt.Sprintln())
								}

								fmt.Println()

								// cleanups and extract items as references
								allItems := make([]*GoModFileRequireItem, 0)
								directItems := make([]*GoModFileRequireItem, 0)
								for _, item := range goMod.Require {
									refItem := &item

									refItem.Path = strings.TrimSpace(strings.ToLower(refItem.Path))
									refItem.Version = strings.TrimSpace(refItem.Version)

									allItems = append(allItems, refItem)
									if refItem.Indirect == nil || !*refItem.Indirect {
										directItems = append(directItems, refItem)
									}
								}

								if len(directItems) > 0 {
									fmt.Println("Checking dependencies for up-to-dateness ...")
									for i, item := range directItems {
										s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
										s.Prefix = "\t["
										s.Suffix = fmt.Sprintf("] Checking '%s' (%v/%v) ...", item.Path, i+1, len(directItems))
										s.Start()

										thisVersion, err := version.NewVersion(strings.TrimSpace(item.Version))
										if err == nil {
											url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", item.Path)
											req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
											if err == nil {
												client := &http.Client{}
												resp, err := client.Do(req)
												if err == nil {
													defer resp.Body.Close()

													if resp.StatusCode == 200 {
														responseData, err := io.ReadAll(resp.Body)
														if err == nil {
															var infoFromProxy GoProxyModuleInfo
															err := json.Unmarshal(responseData, &infoFromProxy)

															if err == nil {
																otherVersion, err := version.NewVersion(strings.TrimSpace(item.Version))
																if err == nil {
																	s.Stop()

																	if otherVersion.LessThanOrEqual(thisVersion) {
																		fmt.Printf("\t[%s] '%s' is up-to-date%s", green("✓"), item.Path, fmt.Sprintln())
																	} else {
																		fmt.Printf("\t[%s] '%s' is outdated: %s < %s%s", yellow("⚠️"), item.Path, thisVersion.String(), otherVersion.String(), fmt.Sprintln())
																	}
																} else {
																	s.Stop()

																	fmt.Printf("\t[%s] Invalid version from '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
																}
															} else {
																s.Stop()

																fmt.Printf("\t[%s] Invalid JSON from '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
															}
														} else {
															s.Stop()

															fmt.Printf("\t[%s] Could not read response from '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
														}

													} else {
														s.Stop()

														fmt.Printf("\t[%s] Unexpected response from '%s': %v%s", red("!"), url, resp.Status, fmt.Sprintln())
													}
												} else {
													s.Stop()

													fmt.Printf("\t[%s] Could not do request to '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
												}
											} else {
												s.Stop()

												fmt.Printf("\t[%s] Could not start request to '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
											}
										} else {
											s.Stop()

											fmt.Printf("\t[%s] Version of '%s' is invalid: %s%s", red("!"), item.Path, err.Error(), fmt.Sprintln())
										}
									}
									fmt.Println()

									fmt.Println("Checking for unsed dependencies ...")
									for i, item := range allItems {
										s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
										s.Prefix = "\t["
										s.Suffix = fmt.Sprintf("] Checking '%s' (%v/%v) ...", item.Path, i+1, len(allItems))
										s.Start()

										p := exec.Command("go", "mod", "why", "-m", item.Path)
										p.Dir = app.Cwd
										p.Stderr = nil
										p.Stdin = nil
										p.Stdout = nil
										output, err := p.Output()

										s.Stop()

										if err == nil {
											strOutput := string(output)
											if strings.Contains(strOutput, fmt.Sprintf("module does not need module %s)", item.Path)) {
												fmt.Printf("\t[%s] Module '%s' is not used, run 'gpm uninstall %s' or a single 'gpm tidy' to fix this%s", red("!"), item.Path, item.Path, fmt.Sprintln())
											} else {
												fmt.Printf("\t[%s] '%s' has no known issues%s", green("✓"), item.Path, fmt.Sprintln())
											}
										} else {
											fmt.Printf("\t[%s] Check failed for '%s':%s%s", red("!"), item.Path, err.Error(), fmt.Sprintln())
										}
									}
									fmt.Println()

									fmt.Println("Checking all dependencies for security issues ...")
									for i, item := range allItems {
										s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
										s.Prefix = "\t["
										s.Suffix = fmt.Sprintf("] Checking '%s' (%v/%v) ...", item.Path, i+1, len(allItems))
										s.Start()

										url := "https://api.osv.dev/v1/query"
										body := map[string]interface{}{
											"version": item.Version,
											"package": map[string]interface{}{
												"name":      item.Path,
												"ecosystem": "Go",
											},
										}

										jsonData, err := json.Marshal(&body)
										if err == nil {
											req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
											if err == nil {
												req.Header.Set("Content-Type", "application/json")
												// ... and finally send the JSON data
												client := &http.Client{}
												resp, err := client.Do(req)
												if err == nil {
													defer resp.Body.Close()

													if resp.StatusCode == 200 {
														responseData, err := io.ReadAll(resp.Body)
														if err == nil {
															var osvResponse types.OsvDevResponse
															err = json.Unmarshal(responseData, &osvResponse)
															if err == nil {
																reportNoIssues := func() {
																	s.Stop()

																	fmt.Printf("\t[%s] '%s' has no known issues%s", green("✓"), item.Path, fmt.Sprintln())
																}

																if osvResponse.Vulnerabilities != nil {
																	vulnerabilities := []types.OsvDevResponseVulnerabilityItem{}
																	vulnerabilities = append(vulnerabilities, *osvResponse.Vulnerabilities...)
																	vulnerabilitiesCount := len(vulnerabilities)

																	if vulnerabilitiesCount > 0 {
																		s.Stop()

																		fmt.Printf("\t[%s] Found %v known security issues in '%s':%s", red("!"), vulnerabilitiesCount, url, fmt.Sprintln())

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
																	} else {
																		reportNoIssues()
																	}
																} else {
																	reportNoIssues()
																}
															} else {
																s.Stop()

																fmt.Printf("\t[%s] Invalid JSON from '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
															}
														} else {
															s.Stop()

															fmt.Printf("\t[%s] Could not read response from '%s': %s%s", red("!"), url, err.Error(), fmt.Sprintln())
														}
													} else {
														s.Stop()

														fmt.Printf("\t[%s] Unexpected response from '%s': %v%s", red("!"), url, resp.Status, fmt.Sprintln())
													}
												} else {
													s.Stop()

													fmt.Printf("\t[%s] Could not do request to '%s':%s%s", red("!"), url, err.Error(), fmt.Sprintln())
												}
											} else {
												s.Stop()

												fmt.Printf("\t[%s] Could not prepare request for '%s':%s%s", red("!"), url, err.Error(), fmt.Sprintln())
											}
										} else {
											s.Stop()

											fmt.Printf("\t[%s] JSON is for '%s' cannot be created:%s%s", red("!"), url, err.Error(), fmt.Sprintln())
										}
									}
									fmt.Println()
								}
							} else {
								fmt.Printf("\t[%s] JSON is invalid, try run 'go mod edit -json':%s%s", red("!"), err.Error(), fmt.Sprintln())
								fmt.Println()
							}
						} else {
							fmt.Printf("\t[%s] File is invalid, try run 'go mod edit -json'%s", red("!"), fmt.Sprintln())
							fmt.Println()
						}
					}
				} else {
					fmt.Printf("[%s] Could not check go.mod file: %s%s", yellow("⚠️"), err.Error(), fmt.Sprintln())
					fmt.Println()
				}
			}

			fmt.Println("Environment variables ...")
			{
				vars := make([]string, 0)
				vars = append(vars, "GOPATH", "GOROOT", "GOPROXY")

				for _, varName := range vars {
					varValue := os.Getenv(varName)
					if varValue != "" {
						fmt.Printf("\t[%s] %s is set: %s%s", green("✓"), varName, varValue, fmt.Sprintln())
					} else {
						fmt.Printf("\t[%s] %s is not set%s", yellow("⚠️"), varName, fmt.Sprintln())
					}
				}
			}
		},
	}

	parentCmd.AddCommand(
		doctorCmd,
	)
}
