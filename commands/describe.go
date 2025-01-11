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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Describe_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var customLanguage string
	var customMessage string
	var prettyOutput bool
	var simple bool
	var temperature float32
	var yamlOutput bool

	var describeCmd = &cobra.Command{
		Use:     "describe [files]",
		Aliases: []string{"desc"},
		Short:   "Describe data",
		Long:    `Describes the data, like images, with AI.`,
		Run: func(cmd *cobra.Command, args []string) {
			allInputs, err := app.ReadAllInputs(args...)
			utils.CheckForError(err)

			consoleFormatter := utils.GetBestChromaFormatterName()
			consoleStyle := utils.GetBestChromaStyleName()

			contentType := strings.ToLower(http.DetectContentType(allInputs))
			if !strings.HasPrefix(contentType, "image/") {
				// current only images are supported
				utils.CheckForError(fmt.Errorf("content type %s is not supported", contentType))
			}

			systemPrompt := ""
			if !app.NoSystemPrompt {
				systemPrompt = app.GetSystemAIPrompt("You are a helpful assistant who helps me to generate accessible content.")
			}

			apiOptions := types.CreateAIChatOptions{
				SystemPrompt: &systemPrompt,
			}

			api, err := app.CreateAIChat(apiOptions)
			utils.CheckForError(err)

			model := strings.TrimSpace(app.Model)
			if model == "" {
				app.Debug("Setting up default model ...")

				if api.GetProvider() == "openai" {
					model = "gpt-4o-mini"
				} else if api.GetProvider() == "ollama" {
					model = "llama3.3"
				}
			}

			currentTemperature := temperature

			if model != "" {
				api.UpdateModel(model)
			}
			api.UpdateTemperature(currentTemperature)

			language := strings.TrimSpace(customLanguage)
			if language == "" {
				language = "english"
			}

			if simple {
				language = fmt.Sprintf("%s (only in simple language)", language)
			}

			message := strings.TrimSpace(customMessage)
			if message == "" {
				message = fmt.Sprintf("Describe what is in the image and answer in %v", language)
			}

			app.Debug(fmt.Sprintf("Provider: %s", api.GetProvider()))
			app.Debug(fmt.Sprintf("Model: %s", api.GetModel()))
			app.Debug(fmt.Sprintf("Temperature: %v", currentTemperature))
			app.Debug(fmt.Sprintf("Message: %v", message))
			app.Debug(fmt.Sprintf("Content type: %v", contentType))

			var base64InputData strings.Builder
			encoder := base64.NewEncoder(base64.StdEncoding, &base64InputData)
			encoder.Write(allInputs)
			utils.CheckForError(err)

			dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64InputData.String())

			imageDescription, err := api.DescribeImage(message, dataURI)
			utils.CheckForError(err)

			outputData := func(data []byte, syntax string) {
				if prettyOutput {
					err = quick.Highlight(app.Out, string(data), syntax, consoleFormatter, consoleStyle)
					if err != nil {
						fmt.Print(string(data))
					}
				} else {
					fmt.Print(string(data))
				}
			}

			if yamlOutput {
				yamlData, err := yaml.Marshal(&imageDescription)
				utils.CheckForError(err)

				outputData(yamlData, "yaml")
			} else {
				if prettyOutput {
					jsonData, err := json.MarshalIndent(&imageDescription, "", "  ")
					utils.CheckForError(err)

					outputData(jsonData, "json")
				} else {
					jsonData, err := json.Marshal(&imageDescription)
					utils.CheckForError(err)

					outputData(jsonData, "json")
				}
			}
		},
	}

	describeCmd.Flags().StringVarP(&customLanguage, "language", "", "", "custom response language")
	describeCmd.Flags().StringVarP(&customMessage, "message", "", "", "custom AI model")
	describeCmd.Flags().BoolVarP(&prettyOutput, "pretty", "", false, "pretty output")
	describeCmd.Flags().BoolVarP(&simple, "simple", "", simple, "use simple language")
	describeCmd.Flags().Float32VarP(&temperature, "temperature", "", utils.GetAIChatTemperature(0.3), "custom temperature value")
	describeCmd.Flags().BoolVarP(&yamlOutput, "yaml", "", false, "use YAML instead of JSON")

	parentCmd.AddCommand(
		describeCmd,
	)
}
