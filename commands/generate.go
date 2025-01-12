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
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/glamour"
	"github.com/google/uuid"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init_generate_documentation_command(parentCmd *cobra.Command, app *types.AppContext) {
	var man bool
	var markdown bool
	var rest bool
	var yaml bool

	var documentationCmd = &cobra.Command{
		Use:     "documentation [resource]",
		Aliases: []string{"doc", "docs", "dox"},
		Short:   "Generate documentation",
		Long:    `Generate documentation into the current directory.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !man && !markdown && !rest && !yaml {
				app.Debug("Setting 'markdown' as default format ...")

				// default is Markdown
				markdown = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			outDir := app.Cwd
			if len(args) > 0 {
				outDir = strings.TrimSpace(args[0])
			}

			outDir, err := app.EnsureFolder(outDir)
			utils.CheckForError(err)

			// collect generators by flags
			generators := make([]func(), 0)
			if man {
				// man pages
				generators = append(generators, func() {
					app.Debug("Generating man pages in", outDir, "...")

					header := doc.GenManHeader{}

					err := doc.GenManTree(cmd.Root(), &header, outDir)
					utils.CheckForError(err)
				})
			}
			if markdown {
				// Markdown files
				generators = append(generators, func() {
					app.Debug("Generating Markdown files in", outDir, "...")

					err := doc.GenMarkdownTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}
			if rest {
				// ReST files
				generators = append(generators, func() {
					app.Debug("Generating ReST files in", outDir, "...")

					err := doc.GenReSTTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}
			if yaml {
				// YAML files
				generators = append(generators, func() {
					app.Debug("Generating YAML files in", outDir, "...")

					err := doc.GenYamlTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}

			// execute generators
			for _, generate := range generators {
				generate()
			}
		},
	}

	documentationCmd.Flags().BoolVarP(&man, "man", "", false, "generate man pages")
	documentationCmd.Flags().BoolVarP(&markdown, "markdown", "m", false, "generate Markdown files")
	documentationCmd.Flags().BoolVarP(&rest, "rest", "r", false, "generate ReST files")
	documentationCmd.Flags().BoolVarP(&yaml, "yaml", "y", false, "generate YAML files")

	parentCmd.AddCommand(
		documentationCmd,
	)
}

func init_generate_password_command(parentCmd *cobra.Command, app *types.AppContext) {
	var allBytes bool
	var base64Output bool
	var count uint16
	var charset string
	var copyToClipboard bool
	var length uint16
	var minLength uint16
	var noOutput bool
	var waitTime int

	var generatePasswordCmd = &cobra.Command{
		Use:     "password",
		Aliases: []string{"passwd", "passwds", "passwords", "pwd", "pwds"},
		Short:   "Generate password",
		Long:    `Generates one or more passwords.`,
		Run: func(cmd *cobra.Command, args []string) {
			clipboardContent := ""

			var addClipboardContent func(text string)
			if copyToClipboard {
				addClipboardContent = func(text string) {
					clipboardContent += text
				}
			} else {
				addClipboardContent = func(text string) {
					// dummy
				}
			}

			if minLength > 0 {
				if minLength > length {
					utils.CheckForError(fmt.Errorf("min-length %v cannot be greater then length %v", minLength, length))
				}
			}

			var i uint16 = 0
			for {
				if i == count {
					break
				}

				i++
				if i > 1 {
					fmt.Println()
					addClipboardContent(fmt.Sprintln())

					time.Sleep(time.Duration(waitTime) * time.Millisecond)
				}

				app.Debug(fmt.Sprintf("Generating passwords %v ...", i))

				var passwordLength uint16
				if minLength > 0 {
					randVal := utils.GenerateRandomUint16()

					passwordLength = utils.MaxUint16(randVal%length, minLength)
				} else {
					passwordLength = length
				}

				app.Debug(fmt.Sprintf("Password length %v ...", passwordLength))

				password := make([]byte, int(passwordLength))

				if allBytes {
					// use any byte
					app.Debug("Will use no charset ...")

					_, err := cryptoRand.Read(password)
					utils.CheckForError(err)
				} else {
					passwordCharset := charset
					if passwordCharset == "" {
						passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/|"
					}

					app.Debug(fmt.Sprintf("Will use charset: %s", charset))

					for j := range password {
						index, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(passwordCharset))))
						utils.CheckForError(err)

						password[j] = passwordCharset[index.Int64()]
					}
				}

				var passwordToOutput string
				if base64Output {
					app.Debug("Base64 output ...")

					passwordToOutput = base64.URLEncoding.EncodeToString(password)
				} else {
					passwordToOutput = string(password)
				}

				if !noOutput {
					fmt.Print(passwordToOutput)
				}

				addClipboardContent(passwordToOutput)
			}

			if copyToClipboard {
				app.Debug("Copy all to clipboard ...")

				err := clipboard.WriteAll(clipboardContent)
				utils.CheckForError(err)
			}
		},
	}

	generatePasswordCmd.Flags().BoolVarP(&allBytes, "all-bytes", "", false, "use any byte for password")
	generatePasswordCmd.Flags().BoolVarP(&base64Output, "base64", "", false, "output as Base64 string")
	generatePasswordCmd.Flags().StringVarP(&charset, "charset", "", "", "custom charset")
	generatePasswordCmd.Flags().BoolVarP(&copyToClipboard, "copy", "", false, "copy final content to clipboard")
	generatePasswordCmd.Flags().Uint16VarP(&count, "count", "", 1, "custom number password to generate at once")
	generatePasswordCmd.Flags().Uint16VarP(&length, "length", "", 20, "custom length of password")
	generatePasswordCmd.Flags().Uint16VarP(&minLength, "min-length", "", 0, "if defined the length of password will be flexible")
	generatePasswordCmd.Flags().BoolVarP(&noOutput, "no-output", "", false, "do not output to console")
	generatePasswordCmd.Flags().IntVarP(&waitTime, "wait-time", "", 0, "the time in millieconds to wait between two steps")

	parentCmd.AddCommand(
		generatePasswordCmd,
	)
}

func init_generate_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var alwaysYes bool
	var force bool
	var noGitInit bool
	var origin string
	var output string
	var sshUrl bool
	var temperature float32

	var projectCmd = &cobra.Command{
		Use:     "project [module_name]",
		Aliases: []string{"p", "prj", "proj", "project"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Generate project",
		Long:    `Generate project using AI into the current directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			projectUrl := strings.TrimSpace(args[0])
			if projectUrl == "" {
				projectUrl = "example.com/my-go-module"
			}

			systemPrompt := ""
			if !app.NoSystemPrompt {
				systemPrompt = app.GetSystemAIPrompt(
					fmt.Sprintf(`You are an expert in Golang and help me setting up a project that can be opened with Visual Studio Code.
You can assume the following:
- all required tools are installed
- commands will be executed in a common terminal on a "%s" operating system with "%s" architecture
- I start in a directory where are only go.mod and go.sum files
Always create a JSON list of all required steps I have to do so at the end there is a ready-to-use project that I can run with 'go run .' or something similar.
Split code into different files if this makes sense and return all files.
You can use any popular module if needed as well if I does want something else.
Always return the current and complete state based on our current conversation.`,
						runtime.GOOS,
						runtime.GOARCH,
					))
			}

			app.Debug(fmt.Sprintf("System prompt: %s", systemPrompt))

			outDir := app.GetFullPathOrDefault(output, app.Cwd)

			if force {
				app.Debug(fmt.Sprintf("Checking if directory '%s' exists ...", outDir))
				doesOutDirExist, err := utils.IsDirExisting(outDir)
				utils.CheckForError(err)

				if doesOutDirExist {
					app.Debug(fmt.Sprintf("Removing directory '%s' ...", outDir))
					err := os.RemoveAll(outDir)
					utils.CheckForError(err)
				}
			}

			outDir, err := app.EnsureFolder(outDir)
			utils.CheckForError(err)

			app.Debug(fmt.Sprintf("Output directory: %s", outDir))

			currentTemperature := temperature

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

			if model != "" {
				api.UpdateModel(model)
			}
			api.UpdateTemperature(currentTemperature)

			app.Debug(fmt.Sprintf("Provider: %s", api.GetProvider()))
			app.Debug(fmt.Sprintf("Model: %s", api.GetModel()))
			app.Debug(fmt.Sprintf("Temperature: %v", currentTemperature))

			editor := types.NewAIEditor(app, projectUrl)

			var lastResponse *types.GenerateProjectStepsResponse = nil
			updateFileTree := func() {
				files := make([]types.AIEditorFileItem, 0)

				if lastResponse != nil {
					modulesToInstall := map[string]bool{}

					for _, step := range lastResponse.Steps {
						stepType, ok := step["type"].(string)
						if !ok {
							continue
						}

						if stepType == "file" {
							relativeFilePath := step["relative_file_path"].(string)
							utils.CheckForError(err)

							content, _ := step["content"].(string)

							files = append(files, types.AIEditorFileItem{
								Name:    relativeFilePath,
								Content: []byte(content),
							})
						} else if stepType == "install_module" {
							// install module

							moduleUrl, ok := step["module_url"].(string)
							if ok {
								modulesToInstall[moduleUrl] = true
							}
						}
					}

					if len(modulesToInstall) > 0 {
						compilerVersion, err := app.GetCurrentCompilerVersion()

						goCompiler := "0.0.0"
						if err == nil && compilerVersion != nil {
							goCompiler = compilerVersion.String()
						}

						goModContent := fmt.Sprintf(`module %s

go %s

require (
`, projectUrl, goCompiler)

						for modUrl := range modulesToInstall {
							goModContent = goModContent + fmt.Sprintf(`%v%v latest
`, "\t", modUrl)
						}

						goModContent = goModContent + `)`
						files = append(files, types.AIEditorFileItem{
							Name:    "go.mod",
							Content: []byte(goModContent),
						})
					}
				}

				editor.UpdateFileTree(files)
			}
			updateFromLastResponse := func() {
				updateFileTree()
			}

			editor.OnCreateClick = func() error {
				if lastResponse == nil {
					return errors.New("no chat response available")
				}

				getFullOutputPath := func(p string) (string, error) {
					dir := p
					if !filepath.IsAbs(dir) {
						dir = filepath.Join(outDir, dir)
					}

					if dir != outDir && !strings.HasPrefix(dir, fmt.Sprintf("%s%s", outDir, string(filepath.Separator))) {
						return dir, errors.New("invalid directory")
					}

					return dir, nil
				}

				askUser := func(question string) bool {
					if !alwaysYes {
						reader := bufio.NewReader(app.In)

						for {
							fmt.Printf("%s Do you want to do this (Y/n)?: ", question)

							userInput, err := reader.ReadString('\n')
							utils.CheckForError(err)

							userInput = strings.TrimSpace(
								strings.ToLower(userInput),
							)

							switch userInput {
							case "", "y", "yes":
								return true
							case "n", "no":
								return false
							}
						}
					}

					return true
				}

				return editor.StopWith(func() error {
					// git init
					if !noGitInit {
						p := utils.CreateShellCommandByArgs("git", "init")
						p.Dir = outDir
						p.Stdout = nil
						p.Stderr = nil
						app.Debug("Initializing git repository ...")
						utils.RunCommand(p)

						// repo URL for origin
						originUrl := strings.TrimSpace(origin)
						if originUrl == "" {
							originUrl = fmt.Sprintf("https://%s", projectUrl)

							if sshUrl {
								// convert to SSH

								parsedURL, err := url.Parse(originUrl)
								utils.CheckForError(err)

								host := parsedURL.Hostname()
								userRepoPath := parsedURL.Path[1:]

								originUrl = fmt.Sprintf("git@%s:%s.git", host, userRepoPath)
							}
						}

						p = utils.CreateShellCommandByArgs("git", "remote", "add", "origin", originUrl)
						p.Dir = outDir
						p.Stdout = nil
						p.Stderr = nil
						app.Debug(fmt.Sprintf("Adding git remote '%s' as 'origin' ...", originUrl))
						utils.RunCommand(p)
					}

					// cleanup project
					p := utils.CreateShellCommandByArgs("go", "mod", "init", projectUrl)
					p.Dir = outDir
					p.Stdout = nil
					p.Stderr = nil
					app.Debug(fmt.Sprintf("Cleanup project '%s' ...", projectUrl))
					utils.RunCommand(p)

					// run steps
					for i, step := range lastResponse.Steps {
						stepNr := i + 1
						stepDescription := step["description"].(string)
						stepTitle := step["title"].(string)
						stepType := step["type"].(string)

						app.Debug(fmt.Sprintf("Step #%v (%s): %s", stepNr, stepTitle, stepDescription))

						if stepType == "file" {
							// create a file

							relativeFilePath := step["relative_file_path"].(string)
							fullPath, err := getFullOutputPath(relativeFilePath)
							utils.CheckForError(err)
							content := step["content"].(string)

							if !askUser(fmt.Sprintf("Step #%v will create a file '%s'.", stepNr, relativeFilePath)) {
								continue
							}

							app.Debug(fmt.Sprintf("Creating file '%s' ...", fullPath))
							os.WriteFile(fullPath, []byte(content), 0664)
						} else if stepType == "install_module" {
							// install module

							moduleUrl := step["module_url"].(string)

							if !askUser(fmt.Sprintf("Step #%v will install a module from '%s'.", stepNr, moduleUrl)) {
								continue
							}

							p := utils.CreateShellCommandByArgs("go", "get", moduleUrl)
							p.Dir = outDir
							p.Stdout = nil
							p.Stderr = nil
							app.Debug(fmt.Sprintf("Installing module '%s' ...", moduleUrl))
							utils.RunCommand(p)
						} else {
							app.L.Println("[STOP]", fmt.Sprintf("Step of type '%s' is not supported", stepType))
							os.Exit(666)
						}
					}

					// cleanup project
					app.TidyUp()

					// output final summary
					out, _ := glamour.Render(lastResponse.FinalSummary, "dark")
					fmt.Println(out)

					return nil
				})
			}

			editor.OnResetClick = func() error {
				editor.ChatHistory.Clear()

				lastResponse = nil
				updateFromLastResponse()

				editor.ChatEditor.SetText("", true)
				editor.UI.SetFocus(editor.ChatEditor)

				return nil
			}

			var numberOfRequests uint64 = 0
			editor.OnSendClick = func(userMessage string) error {
				now := time.Now()
				formattedNow := now.Format("2006-01-02 15:04:05")

				var schema = map[string]interface{}{
					"type":     "object",
					"required": []string{"final_summary", "steps"},
					"properties": map[string]interface{}{
						"final_summary": map[string]interface{}{
							"type":        "string",
							"description": "This is the Markdown text in pretty human readable format that will be displayed after all steps has been made and where you in details explain what you did and what the user finally has to do (the text must be written as if you had carried out the steps)",
						},
						"steps": map[string]interface{}{
							"type":        "array",
							"description": "The current and aggregated list of steps to do",
							"items": map[string]interface{}{
								"oneOf": []map[string]interface{}{
									// file
									{
										"type": "object",
										"required": []string{
											"content",
											"description",
											"relative_file_path",
											"title",
											"type",
										},
										"description": "Contains information for a specific file of a list that is part of the project",
										"properties": map[string]interface{}{
											"content": map[string]interface{}{
												"type":        "string",
												"description": "The content that is written to the file without any explanation",
											},
											"description": map[string]interface{}{
												"type":        "string",
												"description": "A description of the file step",
											},
											"relative_file_path": map[string]interface{}{
												"type":        "string",
												"description": "The relative path and name of the file",
												"examples":    []string{"foo/bar.txt", "foo/bar/buzz.tsx"},
											},
											"title": map[string]interface{}{
												"type":        "string",
												"description": "A (short) description of the file step as title",
											},
											"type": map[string]interface{}{
												"type":        "string",
												"description": "The type",
												"enum":        []string{"file"},
											},
										},
									},

									// install_module
									{
										"type": "object",
										"required": []string{
											"module_url",
											"description",
											"title",
											"type",
										},
										"description": "Contains information for creating a file",
										"properties": map[string]interface{}{
											"description": map[string]interface{}{
												"type":        "string",
												"description": "A description of the install module step",
											},
											"module_url": map[string]interface{}{
												"type":        "string",
												"description": "The URL to the module which can be used with 'go get <URL>' to install a module",
												"examples":    []string{"github.com/foo/bar", "example.com/project-repo"},
											},
											"title": map[string]interface{}{
												"type":        "string",
												"description": "A (short) description of the install module step as title",
											},
											"type": map[string]interface{}{
												"type":        "string",
												"description": "The type",
												"enum":        []string{"install_module"},
											},
										},
									},
								},
							},
						},
					},
				}

				var jsonAnswer string
				api.WithJsonSchema(userMessage, "GenerateProjectStepsResponseSchema", schema, func(messageChunk string) error {
					jsonAnswer += messageChunk
					return nil
				})

				var response types.GenerateProjectStepsResponse
				err = json.Unmarshal([]byte(jsonAnswer), &response)
				if err != nil {
					return err
				}

				numberOfRequests = numberOfRequests + 1
				nr := numberOfRequests

				updateWithThisResponse := func() {
					lastResponse = &response
					updateFromLastResponse()
				}

				updateWithThisResponse()

				editor.ChatEditor.SetText("", true)
				editor.UI.SetFocus(editor.Tree)

				itemText := fmt.Sprintf("#%s - %s", fmt.Sprint(nr), formattedNow)

				editor.ChatHistory.InsertItem(0, itemText, "", 0, func() {
					editor.ChatEditor.SetText("", true)
				})
				editor.ChatHistory.SetCurrentItem(0)

				return nil
			}

			updateFromLastResponse()

			err = editor.Run()
			utils.CheckForError(err)
		},
	}

	projectCmd.Flags().BoolVarP(&force, "force", "f", false, "remove existing output directory before start")
	projectCmd.Flags().BoolVarP(&noGitInit, "no-git-init", "", false, "do not initialize git directory")
	projectCmd.Flags().StringVarP(&origin, "origin", "", "", "custom git origin url")
	projectCmd.Flags().StringVarP(&output, "output", "o", "", "custom output directory")
	projectCmd.Flags().BoolVarP(&sshUrl, "ssh", "", false, "use SSH url for git repository instead HTTP")
	projectCmd.Flags().Float32VarP(&temperature, "temperature", "", utils.GetAIChatTemperature(0.3), "custom temperature value")
	projectCmd.Flags().BoolVarP(&alwaysYes, "y", "", false, "do not ask user to execute each step")

	parentCmd.AddCommand(
		projectCmd,
	)
}

func init_generate_uuid_command(parentCmd *cobra.Command, app *types.AppContext) {
	var base64Output bool
	var count uint16
	var copyToClipboard bool
	var length uint16
	var noOutput bool
	var waitTime int

	var generatePasswordCmd = &cobra.Command{
		Use:     "guid",
		Aliases: []string{"guids", "uuid", "uuids"},
		Short:   "Generate UUID",
		Long:    `Generates one or more UUIDs/GUIDs.`,
		Run: func(cmd *cobra.Command, args []string) {
			clipboardContent := ""

			var addClipboardContent func(text string)
			if copyToClipboard {
				addClipboardContent = func(text string) {
					clipboardContent += text
				}
			} else {
				addClipboardContent = func(text string) {
					// dummy
				}
			}

			var i uint16 = 0
			for {
				if i == count {
					break
				}

				i++
				if i > 1 {
					fmt.Println()
					addClipboardContent(fmt.Sprintln())

					time.Sleep(time.Duration(waitTime) * time.Millisecond)
				}

				app.Debug(fmt.Sprintf("Generating passwords %v ...", i))

				guid, err := uuid.NewRandom()
				utils.CheckForError(err)

				var passwordToOutput string
				if base64Output {
					app.Debug("Base64 output ...")

					passwordToOutput = base64.URLEncoding.EncodeToString(guid[:])
				} else {
					passwordToOutput = guid.String()
				}

				if !noOutput {
					fmt.Print(passwordToOutput)
				}

				addClipboardContent(passwordToOutput)
			}

			if copyToClipboard {
				app.Debug("Copy all to clipboard ...")

				err := clipboard.WriteAll(clipboardContent)
				utils.CheckForError(err)
			}
		},
	}

	generatePasswordCmd.Flags().BoolVarP(&base64Output, "base64", "", false, "output as Base64 string")
	generatePasswordCmd.Flags().BoolVarP(&copyToClipboard, "copy", "", false, "copy final content to clipboard")
	generatePasswordCmd.Flags().Uint16VarP(&count, "count", "", 1, "custom number password to generate at once")
	generatePasswordCmd.Flags().Uint16VarP(&length, "length", "", 20, "custom length of password")
	generatePasswordCmd.Flags().BoolVarP(&noOutput, "no-output", "", false, "do not output to console")
	generatePasswordCmd.Flags().IntVarP(&waitTime, "wait-time", "", 0, "the time in millieconds to wait between two steps")

	parentCmd.AddCommand(
		generatePasswordCmd,
	)
}

func Init_Generate_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var generateCmd = &cobra.Command{
		Use:     "generate [resource]",
		Aliases: []string{"g", "gen"},
		Short:   "Generate resource",
		Long:    `Generates resources like documentation.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_generate_documentation_command(generateCmd, app)
	init_generate_password_command(generateCmd, app)
	init_generate_project_command(generateCmd, app)
	init_generate_uuid_command(generateCmd, app)

	parentCmd.AddCommand(
		generateCmd,
	)
}
