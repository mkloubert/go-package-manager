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
	"os"
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/briandowns/spinner"
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
			consoleFormatter := utils.GetBestChromaFormatterName()
			consoleStyle := utils.GetBestChromaStyleName()

			systemPrompt := ""
			if !app.NoSystemPrompt {
				systemPrompt = app.GetSystemAIPrompt("")
			}

			apiOptions := types.CreateAIChatOptions{
				SystemPrompt: &systemPrompt,
			}

			api, err := app.CreateAIChat(apiOptions)
			if err != nil {
				utils.CloseWithError(err)
			}

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

			printHelp := func() {
				fmt.Println("Following commands are supported:")
				fmt.Println("\t/cls               clear screen")
				fmt.Println("\t/exit              exit")
				fmt.Println("\t/format <name>     formatter for console output")
				fmt.Println("\t/help              print this help")
				fmt.Println("\t/info              print information about current chat settings")
				fmt.Println("\t/model <name>      switch to another model")
				fmt.Println("\t/reset             reset conversation")
				fmt.Println("\t/style <name>      console style")
				fmt.Println("\t/system <text>     reset conversation and update system prompt")
			}

			printAIInfo := func() {
				fmt.Printf("AI: %v (%v)%v", api.GetProvider(), api.GetModel(), fmt.Sprintln())
				fmt.Printf("System prompt: %v", systemPrompt)
				fmt.Println(api.MoreInfo())
			}

			printInitialScreen := func() {
				printAIInfo()
				fmt.Println()
				printHelp()
			}

			utils.ClearConsole()
			printInitialScreen()

			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print(">>> ")

				scanner.Scan()

				userInput := strings.TrimSpace(scanner.Text())
				if userInput == "" {
					fmt.Printf("[INPUT ERROR] Please submit input%v", fmt.Sprintln())
					continue
				}

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
						consoleFormatter = newFormatter
					}

					continue
				} else if lowerUserInput == "/help" {
					printHelp()
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
				} else if lowerUserInput == "/reset" {
					resetConversation()

					utils.ClearConsole()
					printInitialScreen()

					continue
				} else if strings.HasPrefix(lowerUserInput, "/style ") {
					newStyle := strings.TrimSpace(lowerUserInput[7:])
					if newStyle == "" {
						fmt.Printf("[INPUT ERROR] Please define a style%v", fmt.Sprintln())
					} else {
						consoleStyle = newStyle
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
					err := quick.Highlight(os.Stdout, answer, "markdown", consoleFormatter, consoleStyle)
					if err != nil {
						fmt.Print(answer)
					}
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
