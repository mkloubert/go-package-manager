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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-version"
	"github.com/joho/godotenv"
	"github.com/mkloubert/go-package-manager/utils"

	constants "github.com/mkloubert/go-package-manager/constants"
)

// AIPrompts stores prompts for AI chats
type AIPrompts struct {
	Prompt       string  // the prompt
	SystemPrompt *string // the system prompt, if defined
}

// An AppContext contains all information for running this app
type AppContext struct {
	AliasesFile    AliasesFile  // aliases.yaml file in home folder
	Cwd            string       // current working directory
	EnvFiles       []string     // one or more env files
	Environment    string       // the name of the environment
	GpmFile        GpmFile      // the gpm.y(a)ml file
	L              *log.Logger  // the logger to use
	NoSystemPrompt bool         // do not use system prompt
	Ollama         bool         // use Ollama
	ProjectsFile   ProjectsFile // projects.yaml file in home folder
	Prompt         string       // custom (AI) prompt
	SystemPrompt   string       // custom system prompt
	Verbose        bool         // output verbose information
}

// ChatWithAIOption stores settings for
// `ChatWithAI()` function
type ChatWithAIOption struct {
	Model        *string // custom model
	SystemPrompt *string // custom system prompt
	Temperature  *int    // custom temperature
}

// OllamaGenerateResponse is the response of
// a successful Ollama API call
type OllamaGenerateResponse struct {
	Model    string `json:"model"`    // used model
	Response string `json:"response"` // the response
}

// OpenAIChatCompletionResponseV1 stores data of a successful
// OpenAI chat completion response API response (version 1)
type OpenAIChatCompletionResponseV1 struct {
	Choices []OpenAIChatCompletionResponseV1Choice `json:"choices"` // list of choices
	Model   string                                 `json:"model"`   // used model
	Usage   OpenAIChatCompletionResponseV1Usage    `json:"usage"`   // the usage
}

// OpenAIChatCompletionResponseV1Choice is an item inside `choices` property
// of an `OpenAIChatCompletionResponseV1` object
type OpenAIChatCompletionResponseV1Choice struct {
	Index   int32                                       `json:"index"`   // the zero-based index
	Message OpenAIChatCompletionResponseV1ChoiceMessage `json:"message"` // the message information
}

// OpenAIChatCompletionResponseV1ChoiceMessage contains data for `message` property
// of an `OpenAIChatCompletionResponseV1ChoiceMessage` object
type OpenAIChatCompletionResponseV1ChoiceMessage struct {
	Content string `json:"content"` // the message context
	Role    string `json:"role"`    // the role like 'user' or 'assistant'
}

// OpenAIChatCompletionResponseV1Usage contains data for `usage` property
// of an `OpenAIChatCompletionResponseV1` object
type OpenAIChatCompletionResponseV1Usage struct {
	CompletionTokens int32 `json:"completion_tokens"` // number of completion tokens
	PromptTokens     int32 `json:"prompt_tokens"`     // number of prompt tokens
	TotalTokens      int32 `json:"total_tokens"`      // number of total used tokens
}

const aiApiOllama = "ollama"
const aiApiOpenAI = "openai"

// ChatWithAI() - does a simple AI chat based on the current app settings
func (app *AppContext) ChatWithAI(prompt string, options ...ChatWithAIOption) (string, error) {
	OPENAI_API_KEY := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))

	GPM_AI_API := strings.TrimSpace(
		strings.ToLower(os.Getenv("GPM_AI_API")),
	)
	if GPM_AI_API == "" {
		if app.Ollama {
			GPM_AI_API = aiApiOllama
		} else {
			if OPENAI_API_KEY == "" {
				GPM_AI_API = aiApiOllama
			} else {
				GPM_AI_API = aiApiOpenAI
			}
		}
	}

	if GPM_AI_API == aiApiOpenAI {
		app.Debug("Using Open AI API ...")

		if OPENAI_API_KEY == "" {
			return "", fmt.Errorf("no api key found for OpenAI")
		}

		return app.chatWithOpenAI(prompt, options...)
	}

	if GPM_AI_API == aiApiOllama {
		app.Debug("Using Ollama API ...")

		return app.chatWithOllama(prompt, options...)
	}

	return "", fmt.Errorf("ai api '%v' is not supported", GPM_AI_API)
}

func (app *AppContext) chatWithOllama(prompt string, options ...ChatWithAIOption) (string, error) {
	model := utils.GetDefaultAIChatModel()
	if model == "" {
		return "", fmt.Errorf("no ai model defined")
	}
	var systemPrompt *string
	temperature := 0

	for _, o := range options {
		if o.Model != nil {
			model = *o.Model
		}
		if o.SystemPrompt != nil {
			systemPrompt = o.SystemPrompt
		}
		if o.Temperature != nil {
			temperature = *o.Temperature
		}
	}

	url := "http://localhost:11434/api/generate"

	data := map[string]interface{}{
		"model":       model,
		"prompt":      prompt,
		"stream":      false,
		"temperature": temperature,
	}

	if systemPrompt != nil {
		data["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	app.Debug(fmt.Sprintf("Will do POST request to '%v' with body: %v", url, string(jsonData)))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected response: %v", resp.StatusCode)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response OllamaGenerateResponse
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return "", err
	}

	return response.Response, nil
}

func (app *AppContext) chatWithOpenAI(prompt string, options ...ChatWithAIOption) (string, error) {
	model := utils.GetDefaultAIChatModel()
	var systemPrompt *string
	temperature := 0

	OPENAI_API_KEY := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))

	for _, o := range options {
		if o.Model != nil {
			model = *o.Model
		}
		if o.SystemPrompt != nil {
			systemPrompt = o.SystemPrompt
		}
		if o.Temperature != nil {
			temperature = *o.Temperature
		}
	}

	if model == "" {
		model = "gpt-3.5-turbo"
	}

	url := "https://api.openai.com/v1/chat/completions"

	messages := make([]interface{}, 0)
	if systemPrompt != nil {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": prompt,
	})

	data := map[string]interface{}{
		"messages":    messages,
		"model":       model,
		"temperature": temperature,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OPENAI_API_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected response: %v", resp.StatusCode)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return "", err
	}

	answer := ""
	for _, c := range response.Choices {
		if c.Message.Role == "assistant" {
			answer = c.Message.Content
		}
	}

	return answer, nil
}

// app.Debug() - writes debug information with the underlying logger
func (app *AppContext) Debug(v ...any) *AppContext {
	if app.Verbose {
		app.L.Printf("[VERBOSE] %v", fmt.Sprintln(v...))
	}

	return app
}

// app.EnsureBinFolder() - ensures and returns the path of central bin folder
func (app *AppContext) EnsureBinFolder() (string, error) {
	binPath, err := app.GetBinFolderPath()
	if err != nil {
		return "", err
	}

	info, err := os.Stat(binPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(binPath, constants.DefaultDirMode)
			if err == nil {
				return binPath, nil
			}
			return "", nil
		}
		return "", err
	}

	if info.IsDir() {
		return binPath, nil
	}
	return "", fmt.Errorf("%v is no directory", binPath)
}

// app.GetAIPrompt() - returns the AI prompt based on the current app settings
func (app *AppContext) GetAIPrompt(defaultPrompt string) string {
	prompt := app.Prompt // first from command line arguments

	if prompt == "" {
		prompt = os.Getenv("GPM_AI_PROMPT") // no from environment variable
	}

	if prompt == "" {
		prompt = defaultPrompt // take the default
	}

	return prompt
}

// app.GetAIPromptSettings() - returns AI prompt settings
func (app *AppContext) GetAIPromptSettings(defaultPrompt string, defaultSystemPrompt string) AIPrompts {
	var systemPrompt *string
	if !app.NoSystemPrompt {
		systemPromptToUse := app.GetSystemAIPrompt(defaultSystemPrompt)

		systemPrompt = &systemPromptToUse
	}

	return AIPrompts{
		Prompt:       app.GetAIPrompt(defaultPrompt),
		SystemPrompt: systemPrompt,
	}
}

// app.GetAliasesFilePath() - returns the possible path of the aliases.yaml file
func (app *AppContext) GetAliasesFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		return path.Join(homeDir, ".gpm/aliases.yaml"), nil
	}
	return "", err
}

// app.GetBinFolderPath() - returns the possible path of a central bin folder
func (app *AppContext) GetBinFolderPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		var binPath string

		GPM_BIN_PATH := strings.TrimSpace(os.Getenv("GPM_BIN_PATH"))
		if GPM_BIN_PATH != "" {
			binPath = GPM_BIN_PATH
		} else {
			binPath = path.Join(homeDir, ".gpm/bin")
		}

		if !path.IsAbs(binPath) {
			binPath = path.Join(app.Cwd, binPath)
		}

		return binPath, nil
	}
	return "", nil
}

// app.GetEnvFilePaths() - returns possible paths of .env* files
func (app *AppContext) GetEnvFilePaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		envFilename := ".env"
		envFileWithSuffix := ".env"
		envLocalFilename := ".env.local"

		environment := app.GetEnvironment()
		if environment != "" {
			envFileWithSuffix += "." + environment
		}

		envFilePaths := utils.RemoveDuplicatesInStringList(
			[]string{
				path.Join(homeDir, ".gpm/"+envFilename),                  // ${HOME}/.env
				path.Join(app.Cwd, envFilename),                          // <PROJECT-DIR>/.env
				path.Join(app.Cwd, envFileWithSuffix),                    // <PROJECT-DIR>/.env<SUFFIX>
				path.Join(app.Cwd, envLocalFilename),                     // <PROJECT-DIR>/.env.local
				path.Join(app.Cwd, envFilename+"."+environment+".local"), // <PROJECT-DIR>/.env<SUFFIX>.local
			},
		)

		return envFilePaths, nil
	} else {
		return []string{}, err
	}
}

// app.GetCurrentGitBranch() - returns the name of the current branch using git command
func (app *AppContext) GetCurrentGitBranch() (string, error) {
	p := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	p.Dir = app.Cwd

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
	p.Dir = app.Cwd

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

// app.GetEnvironment() - returns the name of the environment
func (app *AppContext) GetEnvironment() string {
	environment := strings.TrimSpace(app.Environment) // first try --environment flag
	if environment == "" {
		environment = os.Getenv("GPM_ENV") // now try GPM_ENV
	}

	return strings.TrimSpace(
		strings.ToLower(environment),
	)
}

// app.GetGitRemotes() - returns the list of remotes using git command
func (app *AppContext) GetGitRemotes() ([]string, error) {
	p := exec.Command("git", "remote")
	p.Dir = app.Cwd

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

// app.GetGitTags() - returns the list of tags using git command
func (app *AppContext) GetGitTags() ([]string, error) {
	p := exec.Command("git", "tag")
	p.Dir = app.Cwd

	var output bytes.Buffer
	p.Stdout = &output

	err := p.Run()
	if err != nil {
		return []string{}, err
	}
	defer output.Reset()

	tags := strings.Split(
		strings.TrimSpace(output.String()), "\n",
	)

	return tags, nil
}

// app.GetAliasesFilePath() - returns the possible path of the gpm.yaml file
func (app *AppContext) GetGpmFilePath() (string, error) {
	return path.Join(app.Cwd, "gpm.yaml"), nil
}

// app.GetLatestVersion() - Returns the latest version based on the Git tags
// of the current repository or nil if not found.
func (app *AppContext) GetLatestVersion() (*version.Version, error) {
	allVersions, err := app.GetVersions()
	if err != nil {
		return nil, err
	}

	var latestVersion *version.Version
	for _, v := range allVersions {
		updateVersion := func() {
			latestVersion = v
		}

		if latestVersion != nil {
			if latestVersion.LessThanOrEqual(v) {
				updateVersion()
			}
		} else {
			updateVersion()
		}
	}

	return latestVersion, nil
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

// app.GetProjectsFilePath() - returns the possible path of the projects.yaml file
func (app *AppContext) GetProjectsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		return path.Join(homeDir, ".gpm/projects.yaml"), nil
	} else {
		return "", err
	}
}

// app.GetSystemAIPrompt() - returns the AI system prompt based on the current app settings
func (app *AppContext) GetSystemAIPrompt(defaultPrompt string) string {
	prompt := app.SystemPrompt // first from command line arguments

	if prompt == "" {
		prompt = os.Getenv("GPM_AI_SYSTEM_PROMPT") // no from environment variable
	}

	if prompt == "" {
		prompt = defaultPrompt // take the default
	}

	return prompt
}

// app.GetVersions() - Returns all versions represented by Git tags
// inside the current working directory.
func (app *AppContext) GetVersions() ([]*version.Version, error) {
	var versions []*version.Version

	tags, err := app.GetGitTags()
	if err != nil {
		return versions, err
	}

	for _, t := range tags {
		v, err := version.NewVersion(t)
		if err == nil {
			versions = append(versions, v)
		}
	}

	return versions, nil
}

// app.ListFiles() - Lists all files inside the current working directory
// based of the patterns from "files" section of gpm.yaml file.
func (app *AppContext) ListFiles() ([]string, error) {
	var patterns []string
	if len(app.GpmFile.Files) == 0 {
		executableFilename := path.Base(app.Cwd)
		if utils.IsWindows() {
			executableFilename += constants.WindowsExecutableExt
		}

		patterns = append(
			patterns,
			"^"+executableFilename+"$",
			"^CHANGELOG.md$", "^CONTRIBUTING.md$", "^CONTRIBUTION.md$", "^LICENSE$", "^README.md$",
		)
	} else {
		patterns = append(patterns, app.GpmFile.Files...)
	}

	var files []string
	matchingFiles := map[string]bool{}

	for _, p := range patterns {
		filesByPattern, err := utils.ListFiles(app.Cwd, p)
		if err != nil {
			return nil, err
		}

		for _, f := range filesByPattern {
			_, ok := matchingFiles[f]
			if !ok {
				matchingFiles[f] = true
				files = append(files, f)
			}
		}
	}

	matchingFiles = nil

	return files, nil
}

// app.LoadAliasesFileIfExist - Loads a gpm.y(a)ml file if it exists
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadAliasesFileIfExist() bool {
	defer func() {
		if app.AliasesFile.Aliases == nil {
			app.AliasesFile.Aliases = map[string][]string{}
		}
	}()

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
	} else {
		utils.CloseWithError(err)
	}

	return false
}

func (app *AppContext) loadEnvFile(envFilePath string) {
	app.Debug(fmt.Sprintf("Loading env file '%v' ...", envFilePath))
	err := godotenv.Overload(envFilePath)
	if err != nil {
		utils.CloseWithError(err)
	}
}

// app.LoadEnvFilesIfExist() - Loads .env* files if they exist
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadEnvFilesIfExist() {
	envFilePaths, err := app.GetEnvFilePaths()
	if err == nil {
		for _, envFilePath := range envFilePaths {
			isExisting, err := utils.IsFileExisting(envFilePath)
			if err != nil {
				utils.CloseWithError(err)
			}

			if isExisting {
				app.loadEnvFile(envFilePath)
			}
		}

		// now from `--env-file` flags
		for _, envFilePath := range app.EnvFiles {
			app.loadEnvFile(envFilePath)
		}
	} else {
		utils.CloseWithError(err)
	}
}

// app.LoadGpmFileIfExist() - Loads a gpm.y(a)ml file if it exists
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

	gpm, err := LoadGpmFile(gpmFilePath)
	if err != nil {
		utils.CloseWithError(err)
	}

	app.GpmFile = gpm

	return true
}

// app.LoadProjectsFileIfExist() - Loads an aliases.yaml file if it exists
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadProjectsFileIfExist() bool {
	defer func() {
		if app.ProjectsFile.Projects == nil {
			app.ProjectsFile.Projects = map[string]string{}
		}
	}()

	projectsFilePath, err := app.GetProjectsFilePath()
	if err == nil {
		isExisting, err := utils.IsFileExisting(projectsFilePath)
		if err != nil {
			utils.CloseWithError(err)
		}

		if !isExisting {
			return false
		}

		app.Debug(fmt.Sprintf("Loading '%v' file ...", projectsFilePath))

		yamlData, err := os.ReadFile(projectsFilePath)
		if err != nil {
			utils.CloseWithError(err)
		}

		var projects ProjectsFile
		err = yaml.Unmarshal(yamlData, &projects)
		if err != nil {
			utils.CloseWithError(err)
		}

		app.ProjectsFile = projects
		return true
	} else {
		utils.CloseWithError(err)
	}

	return false
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

// app.UpdateAliasesFile() - Updates the aliases.yaml file in home folder.
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

		err = os.MkdirAll(aliasesFileDirectoryPath, constants.DefaultDirMode)
		if err != nil {
			return err
		}
	}

	yamlData, err := yaml.Marshal(&app.AliasesFile)
	if err != nil {
		utils.CloseWithError(err)
	}

	app.Debug(fmt.Sprintf("Updating file '%v' ...", aliasesFilePath))
	return os.WriteFile(aliasesFilePath, yamlData, constants.DefaultFileMode)
}

// app.UpdateProjectsFile() - Updates the projects.yaml file in home folder.
func (app *AppContext) UpdateProjectsFile() error {
	projectsFilePath, err := app.GetProjectsFilePath()
	if err != nil {
		return err
	}

	projectsFileDirectoryPath := path.Dir(projectsFilePath)

	isExisting, err := utils.IsDirExisting(projectsFileDirectoryPath)
	if err != nil {
		return err
	}

	if !isExisting {
		app.Debug(fmt.Sprintf("Creating directory '%v' ...", projectsFileDirectoryPath))

		err = os.MkdirAll(projectsFileDirectoryPath, constants.DefaultDirMode)
		if err != nil {
			return err
		}
	}

	yamlData, err := yaml.Marshal(&app.ProjectsFile)
	if err != nil {
		utils.CloseWithError(err)
	}

	app.Debug(fmt.Sprintf("Updating file '%v' ...", projectsFilePath))
	return os.WriteFile(projectsFilePath, yamlData, constants.DefaultFileMode)
}
