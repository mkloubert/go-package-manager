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

package main

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/commands"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

var rootCmd = &cobra.Command{
	Use:     "gpm",
	Short:   "Package manager for Go",
	Long:    `A package manager for Go projects which simplifies the way of installing dependencies and setting up projects.`,
	Version: AppVersion,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	cwd, err := os.Getwd()
	utils.CheckForError(err)

	var app types.AppContext
	app.L = log.Default()
	app.Cwd = cwd
	app.ErrorOut = os.Stderr
	app.In = os.Stdin
	app.IsCI = strings.TrimSpace(strings.ToLower(os.Getenv("CI"))) == "true"
	app.Out = os.Stdout

	// use "aliases-file flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.AliasesFilePath, "aliases-file", "", "", "custom aliases file")
	// use "environment flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.Environment, "environment", "", "", "name of the environment")
	// use "env-file flag" everywhere
	rootCmd.PersistentFlags().StringArrayVarP(&app.EnvFiles, "env-file", "e", []string{}, "one or more environment files")
	// use "gpm-root flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.GpmRootPath, "gpm-root", "", "", "custom root directory for this app")
	// use custom AI model
	rootCmd.PersistentFlags().StringVarP(&app.Model, "model", "", "", "custom AI model")
	// use "no-system-prompt flag" everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.NoSystemPrompt, "no-system-prompt", "", false, "do not use system prompt")
	// use "ollama flag" everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.Ollama, "ollama", "", false, "use Ollama")
	// use no-post-script everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.NoPostScript, "no-post-script", "", false, "do not handle 'post script' in gpm.yaml")
	// use no-pre-script everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.NoPreScript, "no-pre-script", "", false, "do not handle 'pre script' in gpm.yaml")
	// use no-script everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.NoScript, "no-script", "", false, "do not handle script in gpm.yaml")
	// use "prompt flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.Prompt, "prompt", "", "", "custom (AI) prompt")
	// use "projects-file flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.ProjectsFilePath, "projects-file", "", "", "custom projects file")
	// use "settings file" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.SettingsFilePath, "settings", "", "", "custom settings file")
	// use "system-prompt flag" everywhere
	rootCmd.PersistentFlags().StringVarP(&app.SystemPrompt, "system-prompt", "", "", "custom (AI) system prompt")
	// use "temperature flag" everywhere
	rootCmd.PersistentFlags().Float32VarP(&app.Temperature, "temperature", "", -1, "custom (AI) temperature")
	// use "verbose flag" everywhere
	rootCmd.PersistentFlags().BoolVarP(&app.Verbose, "verbose", "v", false, "verbose output")

	app.LoadEnvFilesIfExist()
	app.LoadSettingsFileIfExist()
	app.LoadAliasesFileIfExist()
	app.LoadProjectsFileIfExist()
	app.LoadGpmFileIfExist()

	// initialize commands
	commands.Init_Add_Command(rootCmd, &app)
	commands.Init_Base64_Command(rootCmd, &app)
	commands.Init_Build_Command(rootCmd, &app)
	commands.Init_Bump_Command(rootCmd, &app)
	commands.Init_Cat_Command(rootCmd, &app)
	commands.Init_Chat_Command(rootCmd, &app)
	commands.Init_Checkout_Command(rootCmd, &app)
	commands.Init_Compress_Command(rootCmd, &app)
	commands.Init_Cron_Command(rootCmd, &app)
	commands.Init_Describe_Command(rootCmd, &app)
	commands.Init_Diff_Command(rootCmd, &app)
	commands.Init_Doctor_Command(rootCmd, &app)
	commands.Init_Down_Command(rootCmd, &app)
	commands.Init_Exec_Command(rootCmd, &app)
	commands.Init_Generate_Command(rootCmd, &app)
	commands.Init_GUID_Command(rootCmd, &app)
	commands.Init_Import_Command(rootCmd, &app)
	commands.Init_Init_Command(rootCmd, &app)
	commands.Init_Install_Command(rootCmd, &app)
	commands.Init_List_Command(rootCmd, &app)
	commands.Init_Make_Command(rootCmd, &app)
	commands.Init_Monitor_Command(rootCmd, &app)
	commands.Init_New_Command(rootCmd, &app)
	commands.Init_Now_Command(rootCmd, &app)
	commands.Init_Open_Command(rootCmd, &app)
	commands.Init_Pack_Command(rootCmd, &app)
	commands.Init_Password_Command(rootCmd, &app)
	commands.Init_Prompt_Command(rootCmd, &app)
	commands.Init_Publish_Command(rootCmd, &app)
	commands.Init_Pull_Command(rootCmd, &app)
	commands.Init_Push_Command(rootCmd, &app)
	commands.Init_Remove_Command(rootCmd, &app)
	commands.Init_Run_Command(rootCmd, &app)
	commands.Init_Setup_Command(rootCmd, &app)
	commands.Init_Show_Command(rootCmd, &app)
	commands.Init_Sleep_Command(rootCmd, &app)
	commands.Init_Start_Command(rootCmd, &app)
	commands.Init_Sync_Command(rootCmd, &app)
	commands.Init_Test_Command(rootCmd, &app)
	commands.Init_Tidy_Command(rootCmd, &app)
	commands.Init_Uncompress_Command(rootCmd, &app)
	commands.Init_Uninstall_Command(rootCmd, &app)
	commands.Init_Up_Command(rootCmd, &app)
	commands.Init_Update_Command(rootCmd, &app)

	// execute
	if err := rootCmd.Execute(); err != nil {
		utils.CloseWithError(err)
	}
}
