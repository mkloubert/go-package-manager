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
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/glamour"
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

func init_generate_project_command(parentCmd *cobra.Command, app *types.AppContext) {
	var alwaysYes bool
	var customModel string
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
You can use any popular module if needed as well if I does want something else.`,
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

			// go mod init
			p := utils.CreateShellCommandByArgs("go", "mod", "init", projectUrl)
			p.Dir = outDir
			p.Stdout = nil
			p.Stderr = nil
			app.Debug(fmt.Sprintf("Initializing project '%s' ...", projectUrl))
			utils.RunCommand(p)

			currentTemperature := temperature

			apiOptions := types.CreateAIChatOptions{
				SystemPrompt: &systemPrompt,
			}

			api, err := app.CreateAIChat(apiOptions)
			utils.CheckForError(err)

			stdin, err := utils.LoadFromSTDINIfAvailable()
			utils.CheckForError(err)

			userMessage := strings.Join(args[1:], " ")
			if stdin != nil {
				userMessage += string(*stdin)
			}

			model := customModel
			if strings.TrimSpace(model) == "" {
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
						"description": "A list of steps to do",
						"items": map[string]interface{}{
							"oneOf": []map[string]interface{}{
								// create_file
								{
									"type": "object",
									"required": []string{
										"content",
										"description",
										"relative_file_path",
										"title",
										"type",
									},
									"description": "Contains information for creating a file",
									"properties": map[string]interface{}{
										"content": map[string]interface{}{
											"type":        "string",
											"description": "The content that is written to the file without any explanation",
										},
										"description": map[string]interface{}{
											"type":        "string",
											"description": "A description of the create file step",
										},
										"relative_file_path": map[string]interface{}{
											"type":        "string",
											"description": "The relative path and name of the file to create",
											"examples":    []string{"foo/bar.txt", "foo/bar/buzz.tsx"},
										},
										"title": map[string]interface{}{
											"type":        "string",
											"description": "A (short) description of the create file step as title",
										},
										"type": map[string]interface{}{
											"type":        "string",
											"description": "The type",
											"enum":        []string{"create_file"},
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
			utils.CheckForError(err)

			askUser := func(question string) bool {
				if !alwaysYes {
					reader := bufio.NewReader(os.Stdin)

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

			for i, step := range response.Steps {
				stepNr := i + 1
				stepDescription := step["description"].(string)
				stepTitle := step["title"].(string)
				stepType := step["type"].(string)

				app.Debug(fmt.Sprintf("Step #%v (%s): %s", stepNr, stepTitle, stepDescription))

				if stepType == "create_file" {
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

					p = utils.CreateShellCommandByArgs("go", "get", moduleUrl)
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
			p = utils.CreateShellCommandByArgs("go", "mod", "tidy")
			p.Dir = outDir
			p.Stdout = nil
			p.Stderr = nil
			app.Debug(fmt.Sprintf("Cleanup project '%s' ...", projectUrl))
			utils.RunCommand(p)

			// output final summary
			out, _ := glamour.Render(response.FinalSummary, "dark")
			fmt.Println(out)
		},
	}

	projectCmd.Flags().BoolVarP(&force, "force", "f", false, "remove existing output directory before start")
	projectCmd.Flags().StringVarP(&customModel, "model", "", "", "custom AI model")
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
	init_generate_project_command(generateCmd, app)

	parentCmd.AddCommand(
		generateCmd,
	)
}
