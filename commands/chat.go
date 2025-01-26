//go:build !openbsd

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
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Chat_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var chatCmd = &cobra.Command{
		Use:     "chat",
		Aliases: []string{"ct"},
		Short:   "AI chat",
		Long:    `Chats with an AI model.`,
		Run: func(cmd *cobra.Command, args []string) {
			chromaSettings := app.GetChromaSettings()

			systemPrompt := ""
			if !app.NoSystemPrompt {
				systemPrompt = app.GetSystemAIPrompt("")
			}

			currentTemperature := app.GetAITemperature(0.3)

			apiOptions := types.CreateAIChatOptions{
				SystemPrompt: &systemPrompt,
			}

			api, err := app.CreateAIChat(apiOptions)
			utils.CheckForError(err)

			var resetConversation func()
			setupResetConversation := func() {
				if systemPrompt == "" {
					resetConversation = func() {
						api.ClearHistory()
					}
				} else {
					resetConversation = func() {
						api.UpdateSystem(systemPrompt)
					}
				}
			}

			setupResetConversation()

			printAIInfo := func() {
				systemPromptToDisplay := systemPrompt
				if systemPromptToDisplay == "" {
					systemPromptToDisplay = "(none)"
				} else {
					systemPromptToDisplay = color.New(color.FgWhite, color.Bold).Sprint(systemPromptToDisplay)
				}

				fmt.Printf("System prompt: %v%v", systemPromptToDisplay, fmt.Sprintln())
				fmt.Printf("Temperature: %v", currentTemperature)
				fmt.Println(api.GetMoreInfo())
			}

			printInitialScreen := func() {
				printAIInfo()
				fmt.Println()
			}

			utils.ClearConsole()
			printInitialScreen()

			history := []string{}
			addInputToHistory := func(input string) {
				if strings.TrimSpace(input) == "" {
					return
				}

				history = append(history, input)
			}

			completer := func(in prompt.Document) []prompt.Suggest {
				w := strings.TrimSpace(in.GetWordBeforeCursorWithSpace())
				if w != "" {
					return []prompt.Suggest{}
				}

				// convert utils.ChatPromptSuggestion to prompt.Suggest
				s := make([]prompt.Suggest, 0)
				for _, suggestion := range utils.GetChatPromptSugesstions() {
					s = append(s, prompt.Suggest{Text: suggestion.Text, Description: suggestion.Description})
				}

				return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
			}

			reset := func() {
				resetConversation()

				utils.ClearConsole()
				printInitialScreen()
			}

			showCompletionAtStart := true
			for {
				fmt.Printf(
					"%v@%v%v",
					api.GetModel(), api.GetProvider(),
					api.GetPromptSuffix(),
				)

				userInputOptions := []prompt.Option{
					prompt.OptionPrefixTextColor(prompt.Yellow),
					prompt.OptionHistory(history),
					prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
					prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
					prompt.OptionSuggestionBGColor(prompt.DarkGray),
					prompt.OptionCompletionOnDown(),
					prompt.OptionMaxSuggestion(10),
				}
				if showCompletionAtStart {
					userInputOptions = append(userInputOptions, prompt.OptionShowCompletionAtStart())
				}

				userInput := strings.TrimSpace(
					prompt.Input(
						" >>> ",
						completer,
						userInputOptions...,
					),
				)
				if userInput == "" {
					fmt.Printf("[INPUT ERROR] Please submit input%v", fmt.Sprintln())
					continue
				}

				showCompletionAtStart = false

				lowerUserInput := strings.ToLower(userInput)

				if lowerUserInput == "/cls" {
					utils.ClearConsole()
					continue
				} else if lowerUserInput == "/exit" {
					break
				} else if strings.HasPrefix(lowerUserInput, "/format ") {
					newFormatter := strings.TrimSpace(lowerUserInput[8:])
					if newFormatter == "" {
						fmt.Printf("[INPUT ERROR] Please define a formatter%v", fmt.Sprintln())
					} else {
						chromaSettings.Formatter = newFormatter
					}

					continue
				} else if lowerUserInput == "/info" {
					printAIInfo()
					continue
				} else if strings.HasPrefix(lowerUserInput, "/model ") {
					newModel := strings.TrimSpace(lowerUserInput[6:])
					if newModel == "" {
						fmt.Printf("[INPUT ERROR] Please define a model%v", fmt.Sprintln())
					} else {
						api.UpdateModel(newModel)

						printAIInfo()
					}

					continue
				} else if lowerUserInput == "/nosystem" {
					systemPrompt = ""

					reset()
					continue
				} else if lowerUserInput == "/reset" {
					reset()
					continue
				} else if strings.HasPrefix(lowerUserInput, "/style ") {
					newStyle := strings.TrimSpace(lowerUserInput[7:])
					if newStyle == "" {
						fmt.Printf("[INPUT ERROR] Please define a style%v", fmt.Sprintln())
					} else {
						chromaSettings.Style = newStyle
					}

					continue
				} else if strings.HasPrefix(lowerUserInput, "/system ") {
					newSystemPrompt := strings.TrimSpace(userInput[8:])
					if newSystemPrompt == "" {
						fmt.Printf("[INPUT ERROR] Please define a system prompt%v", fmt.Sprintln())
					} else {
						systemPrompt = newSystemPrompt
						setupResetConversation()

						resetConversation()
					}

					continue
				} else if strings.HasPrefix(lowerUserInput, "/temp ") {
					newTempValue := strings.TrimSpace(userInput[6:])
					if newTempValue == "" {
						fmt.Printf("[INPUT ERROR] Please define a temperature value%v", fmt.Sprintln())
					} else {
						value64, err := strconv.ParseFloat(newTempValue, 32)
						if err != nil {
							fmt.Printf("[INPUT ERROR] Could not parse input value to number: %v%v", err, fmt.Sprintln())
						} else {
							currentTemperature = float32(value64)

							api.UpdateTemperature(currentTemperature)
						}
					}

					continue
				} else if strings.HasPrefix(lowerUserInput, "/") {
					fmt.Printf("[INPUT ERROR] Invalid command '%v'%v", userInput, fmt.Sprintln())
					continue
				}

				s := spinner.New(spinner.CharSets[24], 100*time.Millisecond)
				s.Start()
				s.Suffix = " Waiting for assistant ..."

				answer := ""
				err := api.SendMessage(
					userInput,
					func(messageChunk string) error {
						answer += messageChunk
						return nil
					},
				)

				s.Stop()

				if err == nil {
					addInputToHistory(userInput)

					chromaSettings.HighlightMarkdown(answer)
				} else {
					fmt.Printf("[AI ERROR]: %v", err)
				}
				fmt.Println()
			}
		},
	}

	parentCmd.AddCommand(
		chatCmd,
	)
}
