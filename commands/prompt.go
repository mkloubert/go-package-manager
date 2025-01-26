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
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Prompt_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var assistantMessages []string
	var isChat bool
	var userMessages []string

	var promptCmd = &cobra.Command{
		Use:   "prompt",
		Short: "AI prompt",
		Long:  `Executes a prompt or chat conversation using an AI API.`,
		Run: func(cmd *cobra.Command, args []string) {
			userMessageCount := len(userMessages)
			assistantMessageCount := len(assistantMessages)

			if userMessageCount != assistantMessageCount {
				utils.CloseWithError(
					fmt.Errorf(
						"number of user and assistant messages are different (%v / %v)",
						userMessageCount, assistantMessageCount,
					),
				)
			}

			systemPrompt := ""
			if !app.NoSystemPrompt {
				systemPrompt = app.GetSystemAIPrompt("")
			}

			model := strings.TrimSpace(app.Model)
			if model == "" {
				model = app.GetDefaultAIChatModel()
			}

			stdin, err := app.LoadFromInputIfAvailable()
			utils.CheckForError(err)

			newUserMessage := strings.Join(args, " ")
			if stdin != nil {
				newUserMessage += string(*stdin)
			}

			aiChat, err := app.CreateAIChat()
			utils.CheckForError(err)

			isChatConversation := isChat || assistantMessageCount > 0

			aiChat.UpdateModel(model)
			aiChat.UpdateSystem(systemPrompt)

			var temperature float32
			if isChatConversation {
				temperature = app.GetAITemperature(0.3)
			} else {
				temperature = app.GetAITemperature(0)
			}
			aiChat.UpdateTemperature(temperature)

			app.Debug("Setup AI chat with following settings:")
			app.Debug(fmt.Sprintf("Provider: %v", aiChat.GetProvider()))
			app.Debug(fmt.Sprintf("Model: %v", aiChat.GetModel()))
			app.Debug(fmt.Sprintf("Temperature: %v", temperature))
			app.Debug(fmt.Sprintf("System prompt: %v", systemPrompt))

			answer := ""
			onMessageUpdate := func(messageChunk string) error {
				answer += messageChunk
				return nil
			}

			if isChatConversation {
				// chat conversation

				app.Debug(fmt.Sprintf("Type: %v", "chat conversation"))
				app.Debug(fmt.Sprintf("Prompt: %v", newUserMessage))

				err := aiChat.SendMessage(newUserMessage, onMessageUpdate)
				utils.CheckForError(err)
			} else {
				// completion operation

				app.Debug(fmt.Sprintf("Type: %v", "completion"))
				app.Debug(fmt.Sprintf("Prompt: %v", newUserMessage))

				err := aiChat.SendPrompt(newUserMessage, onMessageUpdate)
				utils.CheckForError(err)
			}

			fmt.Print(answer)
		},
	}

	promptCmd.Flags().StringArrayVarP(&assistantMessages, "assistant", "", []string{}, "assistant messages")
	promptCmd.Flags().BoolVarP(&isChat, "chat", "", false, "is chat conversation and no completion operation")
	promptCmd.Flags().StringArrayVarP(&userMessages, "user", "", []string{}, "user messages")

	parentCmd.AddCommand(
		promptCmd,
	)
}
