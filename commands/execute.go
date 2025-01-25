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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/atotto/clipboard"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Exec_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var customTemperature float32
	var errorCode int
	var force bool
	var noStdin bool
	var successCode int
	var withExitCode bool

	var execCmd = &cobra.Command{
		Use:     "execute [user message]",
		Aliases: []string{"e", "ex", "exec"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Execute shell command",
		Long:    `Executes a shell command with the environment variables loaded from .env files.`,
		Run: func(cmd *cobra.Command, args []string) {
			userMessage := app.Prompt
			userMessage += strings.TrimSpace(
				strings.Join(args, " "),
			)

			if !noStdin {
				stdin, err := app.LoadFromInputIfAvailable()
				utils.CheckForError(err)

				if stdin != nil {
					userMessage += string(*stdin)
				}
			}

			userMessage = strings.TrimSpace(userMessage)

			shell := utils.GetShell()
			app.Debug(fmt.Sprintf("Shell: %v", shell))
			operatingSystem := runtime.GOOS
			app.Debug(fmt.Sprintf("Operating system: %v", operatingSystem))

			chat, err := app.CreateAIChat()
			utils.CheckForError(err)

			app.Debug(fmt.Sprintf("AI: %v", chat.GetProvider()))
			app.Debug(fmt.Sprintf("Model: %v", chat.GetModel()))

			if !app.NoSystemPrompt {
				systemPrompt := app.SystemPrompt
				if systemPrompt == "" {
					// default

					systemPrompt = fmt.Sprintf(
						`You are helpful assistant and an expert in shell "%v" for "%v" operating systems.
You will help the user to create shell commands without your opinion and without explanation.
You can always expect that the user has installed all required software.`,
						shell, operatingSystem,
					)
				}

				app.Debug(fmt.Sprintf("System prompt: %v", systemPrompt))
				chat.UpdateSystem(systemPrompt)
			}

			if customTemperature != -1 {
				app.Debug(fmt.Sprintf("Temperature: %v", customTemperature))
				chat.UpdateTemperature(customTemperature)
			}

			userMessageJSONData, err := json.Marshal(userMessage)
			utils.CheckForError(err)

			var finalUserMessage string
			tries := 0
			tryAgain := func(reasonWhyRejected string) {
				reasonWhyRejected = strings.TrimSpace(reasonWhyRejected)

				tries = tries + 1

				if tries > 1 {
					var rejectMessage string
					if reasonWhyRejected == "" {
						rejectMessage = "The user did not accept any of your previous commands."
					} else {
						reasonWhyRejectedJSONData, err := json.Marshal(reasonWhyRejected)
						utils.CheckForError(err)

						rejectMessage = fmt.Sprintf(
							`The user did not accept your previous command.
The reason and additional instructions from user for the next command: %v.`,
							string(reasonWhyRejectedJSONData),
						)
					}

					finalUserMessage = fmt.Sprintf(
						`%v
The user needs an alternative shell command for this while being in folder "%v": %v.
Your shell command without Markdown which can directly executed (if multiple steps are required combine them together):`,
						rejectMessage,
						app.Cwd,
						string(userMessageJSONData),
					)
				} else {
					// first try

					finalUserMessage = fmt.Sprintf(
						`The user needs a shell command for this while being in folder "%v": %v.
Your shell command without Markdown which can directly executed (if multiple steps are required combine them together):`,
						app.Cwd,
						string(userMessageJSONData),
					)
				}

				app.Debug(fmt.Sprintf("User message: %v", finalUserMessage))
			}

			var answer string
			generateAnswer := func() error {
				answer = ""

				return chat.SendMessage(finalUserMessage, func(messageChunk string) error {
					answer += messageChunk
					return nil
				})
			}

			tryAgain("")
			utils.CheckForError(generateAnswer())

			executeCommand := func() {
				p := utils.CreateShellCommand(answer)
				p.Dir = app.Cwd
				p.Stdout = app.Out
				p.Stderr = app.ErrorOut
				p.Stdin = app.In

				err := p.Run()

				if withExitCode {
					if status, ok := p.ProcessState.Sys().(syscall.WaitStatus); ok {
						os.Exit(status.ExitStatus())
					} else {
						if err != nil {
							os.Exit(errorCode)
						} else {
							os.Exit(successCode)
						}
					}
				} else {
					utils.CheckForError(err)
				}
			}

			if force {
				executeCommand()
			} else {
				// ask before execute

				showPrompt := func() {
					fmt.Printf("Execute '%v'?%v", answer, fmt.Sprintln())
					fmt.Print("[E]xecute, [c]opy, [t]ry again, [a]bort ")
				}
				showPrompt()

				for {
					fmt.Print("> ")

					reader := bufio.NewReader(app.In)
					input, err := reader.ReadString('\n')

					if err != nil {
						log.Println("[ERROR]", err.Error())
						continue
					}

					input = strings.TrimSpace(strings.ToLower(input))
					if input == "" || input == "e" {
						executeCommand()

						break
					} else if input == "a" {
						break
					} else if input == "c" {
						app.Debug(fmt.Sprintf("Copying '%v' to clipboard ...", answer))
						err := clipboard.WriteAll(answer)
						utils.CheckForError(err)

						break
					} else if input == "t" {
						fmt.Print("Reason (can be blank): ")

						reader := bufio.NewReader(app.In)
						reason, err := reader.ReadString('\n')
						utils.CheckForError(err)

						tryAgain(reason)

						err = generateAnswer()
						if err == nil {
							showPrompt()
						} else {
							log.Println("[ERROR]", err.Error())
						}
					} else {
						log.Printf("%v not supported", input)
					}
				}
			}
		},
	}

	execCmd.Flags().IntVarP(&errorCode, "error-code", "", 1, "custom error code")
	execCmd.Flags().BoolVarP(&force, "force", "", false, "do not ask before execute")
	execCmd.Flags().BoolVarP(&noStdin, "no-stdin", "", false, "do not load from STDIN")
	execCmd.Flags().IntVarP(&successCode, "success-code", "", 0, "custom success code")
	execCmd.Flags().Float32VarP(&customTemperature, "temperature", "", -1, "custom temperature value")
	execCmd.Flags().BoolVarP(&withExitCode, "with-exit-code", "", false, "also exit with code from execution")

	parentCmd.AddCommand(
		execCmd,
	)
}
