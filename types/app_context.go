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

package types

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/mkloubert/go-package-manager/utils"
)

// An AppContext contains all information for running this app
type AppContext struct {
	AliasesFile AliasesFile // aliases.yaml file in home folder
	Cwd         string      // current working directory
	GpmFile     GpmFile     // the gpm.y(a)ml file
	L           *log.Logger // the logger to use
	Verbose     bool        // output verbose information
}

// app.Debug() - writes debug information with the underlying logger
func (app *AppContext) Debug(v ...any) *AppContext {
	if app.Verbose {
		app.L.Printf("[VERBOSE] %v", fmt.Sprintln(v...))
	}

	return app
}

// app.GetAliasesFilePath() - returns the possible path of the aliases.yaml file
func (app *AppContext) GetAliasesFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		return path.Join(homeDir, ".gpm/aliases.yaml"), nil
	} else {
		return "", err
	}
}

// app.GetCurrentGitBranch() - returns the name of the current branch using git command
func (app *AppContext) GetCurrentGitBranch() (string, error) {
	p := exec.Command("git", "symbolic-ref", "--short", "HEAD")

	var output bytes.Buffer
	p.Stdout = &output

	err := p.Run()
	if err != nil {
		return "", err
	}
	defer output.Reset()

	return strings.TrimSpace(output.String()), nil
}

// app.GetGitBranches() - returns the list of branches using git command
func (app *AppContext) GetGitBranches() ([]string, error) {
	p := exec.Command("git", "branch", "-a")

	var output bytes.Buffer
	p.Stdout = &output

	err := p.Run()
	if err != nil {
		return []string{}, err
	}
	defer output.Reset()

	lines := strings.Split(output.String(), "\n")

	var branchNames []string
	for _, l := range lines {
		name := strings.TrimSpace(l)
		if name == "" {
			continue
		}

		name = strings.TrimSpace(
			strings.TrimPrefix(name, "* "),
		)
		if name != "" {
			branchNames = append(branchNames, name)
		}
	}

	return branchNames, nil
}

// app.GetGitRemotes() - returns the list of remotes using git command
func (app *AppContext) GetGitRemotes() ([]string, error) {
	p := exec.Command("git", "remote")

	var output bytes.Buffer
	p.Stdout = &output

	err := p.Run()
	if err != nil {
		return []string{}, err
	}
	defer output.Reset()

	lines := strings.Split(output.String(), "\n")

	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}

	remotes := make([]string, 0)
	for _, l := range lines {
		r := strings.TrimSpace(l)
		if r != "" {
			remotes = append(remotes, r)
		}
	}

	return remotes, nil
}

// app.GetAliasesFilePath() - returns the possible path of the gpm.yaml file
func (app *AppContext) GetGpmFilePath() (string, error) {
	return path.Join(app.Cwd, "gpm.yaml"), nil
}

// app.GetModuleUrls() - returns the list of module urls based on the
// information from gpm.y(a)ml file
func (app *AppContext) GetModuleUrls(moduleNameOrUrl string) []string {
	moduleNameOrUrl = utils.CleanupModuleName(moduleNameOrUrl)

	urls := make([]string, 0)

	for alias, sources := range app.AliasesFile.Aliases {
		if alias == moduleNameOrUrl {
			for _, s := range sources {
				urls = append(urls, utils.CleanupModuleName(s))
			}

			break
		}
	}

	if len(urls) == 0 {
		// take input as fallback
		urls = append(urls, moduleNameOrUrl)
	}

	return urls
}

// app.LoadAliasesFileIfExist - Loads a gpm.y(a)ml file if it exists
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadAliasesFileIfExist() bool {
	aliasesFilePath, err := app.GetAliasesFilePath()
	if err == nil {
		isExisting, err := utils.IsFileExisting(aliasesFilePath)
		if err != nil {
			utils.CloseWithError(err)
		}

		if !isExisting {
			return false
		}

		app.Debug(fmt.Sprintf("Loading '%v' file ...", aliasesFilePath))

		yamlData, err := os.ReadFile(aliasesFilePath)
		if err != nil {
			utils.CloseWithError(err)
		}

		var aliases AliasesFile
		err = yaml.Unmarshal(yamlData, &aliases)
		if err != nil {
			utils.CloseWithError(err)
		}

		app.AliasesFile = aliases
		return true
	}

	return false
}

// app.LoadGpmFileIfExist - Loads a gpm.y(a)ml file if it exists
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadGpmFileIfExist() bool {
	gpmFilePath, err := app.GetGpmFilePath()
	if err != nil {
		utils.CloseWithError(err)
	}

	isExisting, err := utils.IsFileExisting(gpmFilePath)
	if err != nil {
		utils.CloseWithError(err)
	}

	if !isExisting {
		return false
	}

	app.Debug(fmt.Sprintf("Loading '%v' file ...", gpmFilePath))

	yamlData, err := os.ReadFile(gpmFilePath)
	if err != nil {
		utils.CloseWithError(err)
	}

	var gpm GpmFile
	err = yaml.Unmarshal(yamlData, &gpm)
	if err != nil {
		utils.CloseWithError(err)
	}

	app.GpmFile = gpm
	return true
}

// app.RunCurrentProject() - runs the current go project
func (app *AppContext) RunCurrentProject(additionalArgs ...string) {
	p := utils.CreateShellCommandByArgs("go", "run", ".")

	app.Debug(fmt.Sprintf("Running '%v' ...", "go run ."))
	utils.RunCommand(p, additionalArgs...)
}

// app.RunScript() - runs a script defined in gpm.y(a)ml file
func (app *AppContext) RunScript(scriptName string, additionalArgs ...string) {
	cmdToExecute := app.GpmFile.Scripts[scriptName]

	p := utils.CreateShellCommand(cmdToExecute)

	app.Debug(fmt.Sprintf("Running script '%v' ...", scriptName))
	utils.RunCommand(p, additionalArgs...)
}

// app.RunShellCommandByArgs() - runs a shell command by arguments
func (app *AppContext) RunShellCommandByArgs(c string, a ...string) {
	app.Debug(fmt.Sprintf("Running '%v %v' ...", c, strings.Join(a, " ")))

	p := utils.CreateShellCommandByArgs(c, a...)

	utils.RunCommand(p)
}

func (app *AppContext) UpdateAliasesFile() error {
	aliasesFilePath, err := app.GetAliasesFilePath()
	if err != nil {
		return err
	}

	aliasesFileDirectoryPath := path.Dir(aliasesFilePath)

	isExisting, err := utils.IsDirExisting(aliasesFileDirectoryPath)
	if err != nil {
		return err
	}

	if !isExisting {
		app.Debug(fmt.Sprintf("Creating directory '%v' ...", aliasesFileDirectoryPath))

		err = os.MkdirAll(aliasesFileDirectoryPath, 0750)
		if err != nil {
			return err
		}
	}

	yamlData, err := yaml.Marshal(&app.AliasesFile)
	if err != nil {
		utils.CloseWithError(err)
	}

	app.Debug(fmt.Sprintf("Updating file '%v' ...", aliasesFilePath))
	return os.WriteFile(aliasesFilePath, yamlData, 0750)
}
