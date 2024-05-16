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

					logAuditError := func(step string, err error) {
						app.L.Printf("ERROR (%v): Could not audit module #%v (%v %v): '%v'%v",
							step,
							i,
							*modulePath, *moduleVersion,
							err,
							fmt.Sprintln(),
						)
					}

					// now we have all information
					fmt.Printf("%v\t%v%v", *modulePath, *moduleVersion, fmt.Sprintln())

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
						logAuditError("create request body", err)
						return
					}

					app.Debug(
						fmt.Sprintf(
							"Doing POST request to '%v' for module #%v (%v %v) ...",
							url,
							i,
							*modulePath, *moduleVersion,
						),
					)

					// start the request
					req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
					if err != nil {
						logAuditError("create POST request", err)
						return
					}

					// setup ...
					req.Header.Set("Content-Type", "application/json")
					// ... and finally send the JSON data
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						logAuditError("do POST request", err)
						return
					}
					defer resp.Body.Close()

					// load the response
					osvResponseData, err := io.ReadAll(resp.Body)
					if err != nil {
						logAuditError("read response data", err)
						return
					}

					// parse the response
					var osvResponse types.OsvDevResponse
					err = json.Unmarshal(osvResponseData, &osvResponse)
					if err != nil {
						logAuditError("unmarshal response data", err)
						return
					}

					if osvResponse.Vulnerabilities == nil {
						// we have no data here

						app.Debug(
							fmt.Sprintf(
								"Found no vulnerabilities for module #%v (%v %v) ...",
								i,
								*modulePath, *moduleVersion,
							),
						)
						return
					}

					vulnerabilities := *osvResponse.Vulnerabilities
					vulnerabilitiesCount := len(vulnerabilities)

					app.Debug(
						fmt.Sprintf(
							"Found %v vulnerabilities for module #%v (%v %v) ...",
							vulnerabilitiesCount,
							i,
							*modulePath, *moduleVersion,
						),
					)

					for _, v := range vulnerabilities {
						summaryJsonData, err := json.Marshal(v.Summary)
						if err != nil {
							summaryJsonData = []byte{}
						}

						// output [<Severity>] <Id> <Summary>
						// with 1 leading tab
						fmt.Printf(
							"\t[%v]\t%v\t%v%v",
							v.Severity,
							v.Id,
							string(summaryJsonData),
							fmt.Sprintln(),
						)

						if v.References == nil {
							continue
						}

						// output references
						for _, r := range *v.References {
							// output [<Type>] <Url>
							// with 2 leading tabs
							fmt.Printf(
								"\t\t[%v]\t%v%v",
								r.Type,
								r.Url,
								fmt.Sprintln(),
							)
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
