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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

const branchSlugRegex = `[^/a-z0-9\\s-]`

func Init_Checkout_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var suggest bool
	var yes bool

	var checkoutCmd = &cobra.Command{
		Use:     "checkout",
		Aliases: []string{"co"},
		Short:   "Checks out a git branch",
		Long:    `Checks out a git branch while optionally using AI for suggestion of new branches.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			branchNameOrDescription := strings.TrimSpace(args[0])

			branches, err := app.GetGitBranches()
			utils.CheckForError(err)

			if suggest {
				// suggest branch name by AI from description
				branchDescription := strings.Join(args, " ")

				jsonStr, err := json.Marshal(branchDescription)
				utils.CheckForError(err)

				aiPrompts := app.GetAIPromptSettings(
					fmt.Sprintf(`I need the name for a git branch of maximum 48 characters.
For the context I give you the following description: %v
Use only the following format for the full name: prefix/name
Allowed are the following prefixes:
- "feature/" for features (e.g. "feature/audio-chat")
- "bugfix/" for bugfixes (e.g. "bugfix/wrong-score")
- "hotfix/" for hotfixes (e.g. "hotfix/critical-payment-issue")
- "docs/" for documentation (e.g. "docs/assets-optimization")
The name must match the description.
Your full name for the branch without your explanation:`, string(jsonStr)),
					`You are a assistant for git operations. Do exactly what the user wants.`,
				)

				app.Debug(fmt.Sprintf("Chat with AI using following prompt: %v", aiPrompts.Prompt))
				answer, err := app.ChatWithAI(aiPrompts.Prompt, types.ChatWithAIOption{
					SystemPrompt: aiPrompts.SystemPrompt,
				})
				utils.CheckForError(err)

				branchName := utils.Slugify(answer, branchSlugRegex)

				if !yes {
					for {
						fmt.Printf("Do you want to create a branch called '%v'? (Y/n): ", branchName)

						var response string
						fmt.Scanln(&response)
						response = strings.ToLower(strings.TrimSpace(response))

						if response == "y" || response == "" {
							break
						} else if response == "n" {
							os.Exit(3)
							return
						}
					}
				}

				cmdArgs := []string{"git", "checkout", "-b", branchName}

				p := utils.CreateShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)

				app.Debug(fmt.Sprintf("Running '%v' ...", strings.Join(cmdArgs, " ")))
				utils.RunCommand(p)
			} else {
				var branchName string

				if strings.HasPrefix(branchNameOrDescription, ":") {
					// branch names starting with : will handled as aliases
					// which can be setuped via an environment variable e.g.

					branchAlias := strings.TrimSpace(branchNameOrDescription[1:])
					branchAliasUpper := strings.ToUpper(branchAlias)

					branchName = strings.TrimSpace(
						app.GetEnvValue(fmt.Sprintf("GPM_BRANCH_%v", branchAliasUpper)),
					)
					if branchName == "" {
						branchName = branchAlias
					}
				} else {
					// in this case `branchNameOrDescription` will be handled as branch name
					branchName = utils.Slugify(branchNameOrDescription, branchSlugRegex)
				}

				var cmdArgs []string

				isBranchExisting := utils.IndexOfString(branches, branchName) > -1
				if isBranchExisting {
					cmdArgs = []string{"git", "checkout", branchName}
				} else {
					cmdArgs = []string{"git", "checkout", "-b", branchName}
				}

				p := utils.CreateShellCommandByArgs(cmdArgs[0], cmdArgs[1:]...)

				app.Debug(fmt.Sprintf("Running '%v' ...", strings.Join(cmdArgs, " ")))
				utils.RunCommand(p)
			}
		},
	}

	checkoutCmd.Flags().BoolVarP(&suggest, "suggest", "s", false, "suggest name for new branch by AI")
	checkoutCmd.Flags().BoolVarP(&yes, "yes", "y", false, "auto select 'yes'")

	parentCmd.AddCommand(
		checkoutCmd,
	)
}
