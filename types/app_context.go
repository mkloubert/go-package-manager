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
	"runtime"
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
	AliasesFile      AliasesFile  // aliases.yaml file in home folder
	AliasesFilePath  string       // custom file path of the `aliases.yaml` file from CLI flags
	Cwd              string       // current working directory
	EnvFiles         []string     // one or more env files
	Environment      string       // the name of the environment
	ErrorOut         io.Writer    // error output
	GpmFile          GpmFile      // the gpm.y(a)ml file
	GpmRootPath      string       // custom app root path from CLI flags
	In               io.Reader    // the input stream
	IsCI             bool         // indicates if app runs in CI environment like GitHub action or GitLab runner
	L                *log.Logger  // the logger to use
	Model            string       // custom model from CLI flags
	NoSystemPrompt   bool         // do not use system prompt
	Ollama           bool         // use Ollama
	Out              io.Writer    // the output stream
	ProjectsFile     ProjectsFile // projects.yaml file in home folder
	ProjectsFilePath string       // custom file path of the `projects.yaml` file from CLI flags
	Prompt           string       // custom (AI) prompt
	SystemPrompt     string       // custom system prompt
	Verbose          bool         // output verbose information
}

// ChatWithAIOption stores settings for
// `ChatWithAI()` method
type ChatWithAIOption struct {
	Model        *string // custom model
	SystemPrompt *string // custom system prompt
	Temperature  *int    // custom temperature
}

// CreateAIChatOptions stores settings for
// `CreateAIChat()` method
type CreateAIChatOptions struct {
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

// TidyUpOptions - options for app.TidyUp() method
type TidyUpOptions struct {
	Arguments *[]string // command line argumuments
	NoScript  *bool     // true if not running 'tidy' script from gpm.yaml file
}

// ChatWithAI() - does a simple AI chat based on the current app settings
func (app *AppContext) ChatWithAI(prompt string, options ...ChatWithAIOption) (string, error) {
	settings, err := app.GetAIChatSettings()
	if err != nil {
		return "", err
	}

	if settings.Provider == constants.AIApiOpenAI {
		app.Debug("Using Open AI API ...")

		if settings.ApiKey == nil || *settings.ApiKey == "" {
			return "", fmt.Errorf("no api key found for OpenAI")
		}

		return app.chatWithOpenAI(prompt, settings, options...)
	}

	if settings.Provider == constants.AIApiOllama {
		app.Debug("Using Ollama API ...")

		return app.chatWithOllama(prompt, options...)
	}

	return "", fmt.Errorf("no implementation for ai api '%v'", settings.Provider)
}

func (app *AppContext) chatWithOllama(prompt string, options ...ChatWithAIOption) (string, error) {
	model := strings.TrimSpace(app.Model)
	if model == "" {
		model = utils.GetDefaultAIChatModel() // no explicit => take default
	}
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

func (app *AppContext) chatWithOpenAI(prompt string, settings AIChatSettings, options ...ChatWithAIOption) (string, error) {
	apiKey := *settings.ApiKey
	var systemPrompt *string
	temperature := 0

	model := strings.TrimSpace(app.Model)
	if model == "" {
		model = utils.GetDefaultAIChatModel()
	}

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
		model = "gpt-4o-mini"
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
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

// app.CreateAIChat() - creates a new ChatAI instance based on the current settings
func (app *AppContext) CreateAIChat(options ...CreateAIChatOptions) (ChatAI, error) {
	settings, err := app.GetAIChatSettings()
	if err != nil {
		return nil, err
	}

	initialModel := strings.TrimSpace(app.Model)
	systemPrompt := ""

	for _, o := range options {
		if o.Model != nil {
			initialModel = strings.TrimSpace(*o.Model)
		}
		if o.SystemPrompt != nil {
			systemPrompt = strings.TrimSpace(*o.SystemPrompt)
		}
	}

	if initialModel == "" {
		initialModel = utils.GetDefaultAIChatModel()
	}

	var api ChatAI = &OllamaAIChat{}
	if settings.Provider == constants.AIApiOllama {
		ollama := OllamaAIChat{
			Verbose: app.Verbose,
		}

		if initialModel == "" {
			initialModel = "llama3.3"
		}

		api = &ollama
	} else if settings.Provider == constants.AIApiOpenAI {
		openai := OpenAIChat{
			Verbose: app.Verbose,
		}

		if initialModel == "" {
			initialModel = "gpt-4o-mini"
		}
		if settings.ApiKey != nil {
			openai.ApiKey = *settings.ApiKey
		}

		api = &openai
	}

	if api != nil {
		if systemPrompt == "" {
			api.ClearHistory()
		} else {
			api.UpdateSystem(systemPrompt)
		}

		api.UpdateModel(initialModel)

		return api, nil
	}
	return nil, fmt.Errorf("'%v' ai chat provider not implemented", settings.Provider)
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

	return app.EnsureFolder(binPath)
}

// app.EnsureFolder() - ensures and returns the path of a specific folder
func (app *AppContext) EnsureFolder(dir string) (string, error) {
	folderPath := app.GetFullPathOrDefault(dir, app.Cwd)

	info, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(folderPath, constants.DefaultDirMode)
			if err == nil {
				return folderPath, nil
			}
			return "", nil
		}
		return "", err
	}

	if info.IsDir() {
		return folderPath, nil
	}
	return "", fmt.Errorf("%v is no directory", folderPath)
}

// app.EnsureRootFolder() - ensures the root directory for this app exists, and
// returns its path on success
func (app *AppContext) EnsureRootFolder() (string, error) {
	rootDir, err := app.GetRootPath()
	if err == nil {
		return app.EnsureFolder(rootDir)
	}
	return "", err
}

// app.GetAIChatSettings() - returns AI chat settings based on this app
func (app *AppContext) GetAIChatSettings() (AIChatSettings, error) {
	var settings AIChatSettings

	OPENAI_API_KEY := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))

	GPM_AI_API := strings.TrimSpace(
		strings.ToLower(os.Getenv("GPM_AI_API")),
	)
	if GPM_AI_API == "" {
		if app.Ollama {
			GPM_AI_API = constants.AIApiOllama
		} else {
			if OPENAI_API_KEY == "" {
				GPM_AI_API = constants.AIApiOllama
			} else {
				GPM_AI_API = constants.AIApiOpenAI
			}
		}
	}

	var err error = nil

	switch GPM_AI_API {
	case constants.AIApiOpenAI:
		if OPENAI_API_KEY != "" {
			settings.ApiKey = &OPENAI_API_KEY
		}
		settings.Provider = GPM_AI_API
	case constants.AIApiOllama:
		settings.Provider = GPM_AI_API
	default:
		err = fmt.Errorf("ai api '%v' is not supported", GPM_AI_API)
	}

	return settings, err
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
	// first try from cli flag
	customFile := strings.TrimSpace(
		app.AliasesFilePath,
	)
	if customFile == "" {
		// now from environment variable
		customFile = strings.TrimSpace(
			os.Getenv("GPM_ALIASES_FILE"),
		)
	}
	if customFile != "" && path.IsAbs(customFile) {
		return customFile, nil
	}

	// now try from <GPM-ROOT> ...

	rootDir, err := app.GetRootPath()
	if err == nil {
		if customFile != "" {
			return path.Join(rootDir, customFile), nil
		}
		return path.Join(rootDir, "aliases.yaml"), nil
	}
	return "", err
}

// app.GetBinFolderPath() - returns the possible path of a central bin folder
func (app *AppContext) GetBinFolderPath() (string, error) {
	gpmDirPath, err := app.GetRootPath()
	if err != nil {
		return "", err
	}

	var binPath string

	GPM_BIN_PATH := strings.TrimSpace(os.Getenv("GPM_BIN_PATH"))
	if GPM_BIN_PATH != "" {
		binPath = GPM_BIN_PATH
	} else {
		binPath = path.Join(gpmDirPath, "bin")
	}

	if !path.IsAbs(binPath) {
		binPath = path.Join(gpmDirPath, binPath)
	}

	return binPath, nil
}

// app.GetCurrentCompilerVersion() - tries to detect the current Go compiler
// version that should be used
func (app *AppContext) GetCurrentCompilerVersion() (*version.Version, error) {
	app.Debug("Checking for version in 'go.mod' file ...")

	// first try detect in `go.mod` of current directory
	goModFile := app.GetFullPathOrDefault("go.mod", "")
	if goModFile != "" {
		doesGoModExist, err := utils.IsFileExisting(goModFile)
		if err == nil && doesGoModExist {
			goModContent, err := os.ReadFile(goModFile)
			if err == nil {
				var versionInGoMod *version.Version = nil

				// search for line with `go <version>`
				for _, line := range strings.Split(string(goModContent), "\n") {
					trimmedLine := strings.TrimSpace(line)
					if !strings.HasPrefix(trimmedLine, "go ") {
						continue
					}

					maybeVersion := strings.TrimSpace(trimmedLine[3:])
					ver, err := version.NewVersion(maybeVersion)
					if err == nil {
						versionInGoMod = ver
					}

					// found => stop here, even if failed
					break
				}

				if versionInGoMod != nil {
					// take from go.mod
					return versionInGoMod, nil
				}
			}
		}
	}

	// now try via `go version`
	app.Debug("Running 'go version' ...")

	p := exec.Command("go", "version")
	p.Env = os.Environ()
	p.Dir = app.Cwd

	output, err := p.Output()
	if err == nil {
		versionOutput := strings.TrimSpace(string(output))
		fields := strings.Fields(versionOutput)
		if len(fields) > 2 {
			ver, err := version.NewVersion(fields[2][2:])
			if err == nil {
				return ver, nil // from `go version`
			}
		}
	}

	app.Debug("Try get version from 'runtime.Version()' ...")
	runtimeVersion := runtime.Version()
	ver, err := version.NewVersion(runtimeVersion[2:])
	if err == nil {
		return ver, nil // from `runtime.Version()`
	}

	return nil, fmt.Errorf("could not detect Go compiler version")
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

// app.GetEnvFilePaths() - returns possible paths of .env* files
func (app *AppContext) GetEnvFilePaths() ([]string, error) {
	rootDir, err := app.GetRootPath()
	if err == nil {
		envFilename := ".env"
		envLocalFilename := ".env.local"

		envFilePaths := utils.RemoveDuplicatesInStringList(
			[]string{
				path.Join(rootDir, envFilename),      // <GPM-ROOT>/.env
				path.Join(app.Cwd, envFilename),      // <PROJECT-DIR>/.env
				path.Join(app.Cwd, envLocalFilename), // <PROJECT-DIR>/.env.local
			},
		)

		return envFilePaths, nil
	}
	return []string{}, err
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

// app.GetFullPathOrDefault() - returns full version of a path or a default if
// input is empty
func (app *AppContext) GetFullPathOrDefault(p string, d string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return d
	}

	if path.IsAbs(p) {
		return p
	}
	return path.Join(app.Cwd, p)
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

// app.GetGoModules() - returns the list of installed Go modules of current project
func (app *AppContext) GetGoModules() ([]GoModule, error) {
	modules := []GoModule{}

	p := exec.Command("go", "list", "-m", "-json", "all")
	p.Dir = app.Cwd

	app.Debug("Running 'go list -m -json all' ...")
	output, err := p.Output()
	if err != nil {
		return modules, err
	}

	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var module GoModule
		err := decoder.Decode(&module)
		if err != nil {
			return modules, err
		}

		modules = append(modules, module)
	}

	return modules, nil
}

// app.GetAliasesFilePath() - returns the possible path of the gpm.yaml file
func (app *AppContext) GetGpmFilePath() (string, error) {
	return path.Join(app.Cwd, "gpm.yaml"), nil
}

// app.GetGpmFilesSection() - returns `Files` section in `gpm.yaml` files based on the current environment
func (app *AppContext) GetGpmFilesSection() []string {
	return app.GpmFile.GetFilesSectionByEnvSafe(app.GetEnvironment())
}

// app.GetModuleUrls() - returns the list of module urls based on the
// information from aliases.y(a)ml file if possible
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

// app.GetName() - returns the of the current app
func (app *AppContext) GetName() string {
	name := strings.TrimSpace(app.GpmFile.Name)
	if name == "" {
		name = strings.TrimSpace(
			path.Base(app.Cwd),
		)
	}

	return name
}

// app.GetProjectsFilePath() - returns the possible path of the projects.yaml file
func (app *AppContext) GetProjectsFilePath() (string, error) {
	// first try from cli flag
	customFile := strings.TrimSpace(
		app.ProjectsFilePath,
	)
	if customFile == "" {
		// now from environment variable
		customFile = strings.TrimSpace(
			os.Getenv("GPM_PROJECTS_FILE"),
		)
	}
	if customFile != "" && path.IsAbs(customFile) {
		return customFile, nil
	}

	// now try from <GPM-ROOT> ...

	rootDir, err := app.GetRootPath()
	if err == nil {
		if customFile != "" {
			return path.Join(rootDir, customFile), nil
		}
		return path.Join(rootDir, "projects.yaml"), nil
	}
	return "", err
}

// app.GetRootPath() - returns the root directory for this app, usually inside the user's
// home directory
func (app *AppContext) GetRootPath() (string, error) {
	// first try from cli flag
	customDir := strings.TrimSpace(
		app.GpmRootPath,
	)
	if customDir == "" {
		// now from environment variable
		customDir = strings.TrimSpace(
			os.Getenv("GPM_ROOT_BASE_PATH"),
		)
	}
	if customDir != "" && path.IsAbs(customDir) {
		return customDir, nil
	}

	// subfolder inside ${HOME}
	homeDir, err := os.UserHomeDir()
	if err == nil {
		var subDir string
		if customDir == "" {
			subDir = ".gpm" // default
		} else {
			subDir = customDir
		}

		// add environment as suffix if defined
		safeEnvName := utils.SanitizeFilename(
			app.GetEnvironment(),
		)
		if safeEnvName != "" {
			subDir = fmt.Sprintf("%s%s%s", subDir, string(os.PathSeparator), safeEnvName)
		}

		return path.Join(homeDir, subDir), nil
	}
	return "", err
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

// app.ListFiles() - Lists all files inside the current working directory
// based of the patterns from "files" section of gpm.yaml file.
func (app *AppContext) ListFiles() ([]string, error) {
	gpmFiles := app.GetGpmFilesSection()

	var patterns []string
	if len(gpmFiles) == 0 {
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
		patterns = append(patterns, gpmFiles...)
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
	utils.CheckForError(err)

	isExisting, err := utils.IsFileExisting(aliasesFilePath)
	utils.CheckForError(err)

	if !isExisting {
		return false
	}

	app.Debug(fmt.Sprintf("Loading '%v' file ...", aliasesFilePath))

	yamlData, err := os.ReadFile(aliasesFilePath)
	utils.CheckForError(err)

	var aliases AliasesFile
	err = yaml.Unmarshal(yamlData, &aliases)
	utils.CheckForError(err)

	app.AliasesFile = aliases
	return true
}

// app.LoadDataFrom() - loads binary data from a source like
// local file system or web URL
func (app *AppContext) LoadDataFrom(source string) ([]byte, error) {
	source = strings.TrimSpace(source)

	if strings.HasPrefix(source, "https:") || strings.HasPrefix(source, "http:") {
		// from web
		app.Debug(fmt.Sprintf("Loading data from web resource '%v' ...", source))
		return utils.DownloadFromUrl(source)
	} else {
		// local file system

		filePath := source
		if !path.IsAbs(filePath) {
			filePath = path.Join(app.Cwd, filePath)
		}

		app.Debug(fmt.Sprintf("Loading data from local resource '%v' ...", filePath))
		return os.ReadFile(filePath)
	}
}

func (app *AppContext) loadEnvFile(envFilePath string) {
	app.Debug(fmt.Sprintf("Loading env file '%v' ...", envFilePath))

	err := godotenv.Overload(envFilePath)
	utils.CheckForError(err)
}

// app.LoadEnvFilesIfExist() - Loads .env* files if they exist
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadEnvFilesIfExist() {
	envFilePaths, err := app.GetEnvFilePaths()
	utils.CheckForError(err)

	for _, envFilePath := range envFilePaths {
		isExisting, err := utils.IsFileExisting(envFilePath)
		utils.CheckForError(err)

		if isExisting {
			app.loadEnvFile(envFilePath)
		}
	}

	// now from `--env-file` flags
	for _, envFilePath := range app.EnvFiles {
		app.loadEnvFile(envFilePath)
	}
}

// app.LoadGpmFileIfExist() - Loads a gpm.y(a)ml file if it exists
// and return `true` if file has been loaded successfully.
func (app *AppContext) LoadGpmFileIfExist() bool {
	gpmFilePath, err := app.GetGpmFilePath()
	utils.CheckForError(err)

	isExisting, err := utils.IsFileExisting(gpmFilePath)
	utils.CheckForError(err)

	if !isExisting {
		return false
	}

	app.Debug(fmt.Sprintf("Loading '%v' file ...", gpmFilePath))

	gpm, err := LoadGpmFile(gpmFilePath)
	utils.CheckForError(err)

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
	utils.CheckForError(err)

	isExisting, err := utils.IsFileExisting(projectsFilePath)
	utils.CheckForError(err)

	if !isExisting {
		return false
	}

	app.Debug(fmt.Sprintf("Loading '%v' file ...", projectsFilePath))

	yamlData, err := os.ReadFile(projectsFilePath)
	utils.CheckForError(err)

	var projects ProjectsFile
	err = yaml.Unmarshal(yamlData, &projects)
	utils.CheckForError(err)

	app.ProjectsFile = projects
	return true
}

// app.NewVersionManager() - creates a new `ProjectVersionManager` instance based on
// this application context
func (app *AppContext) NewVersionManager() *ProjectVersionManager {
	pvm := &ProjectVersionManager{
		app: app,
	}

	return pvm
}

// app.Read() - implementation for an io.Reader
func (app *AppContext) Read(p []byte) (int, error) {
	if app.In == nil {
		return 0, nil // deactivated
	}

	return app.In.Read(p)
}

// app.ReadAllInputs() - reads from all inputs (STDIN and files, in that order)
// and returns binary data
func (app *AppContext) ReadAllInputs(files ...string) ([]byte, error) {
	buffer := bytes.Buffer{}
	_, err := app.WriteAllInputsTo(&buffer, files...)

	return buffer.Bytes(), err
}

// app.RunCurrentProject() - runs the current go project
func (app *AppContext) RunCurrentProject(additionalArgs ...string) {
	p := utils.CreateShellCommandByArgs("go", "run", ".")

	app.Debug(fmt.Sprintf("Running '%v' ...", "go run ."))
	utils.RunCommand(p, additionalArgs...)
}

// app.RunScript() - runs a script defined in gpm.y(a)ml file
func (app *AppContext) RunScript(scriptName string, additionalArgs ...string) {
	finalScriptName := scriptName

	// try to check if there is a script name with environment prefix
	// like `dev:foo` if script is called `foo` and environment `dev` e.g.
	envName := app.GetEnvironment()
	if envName != "" {
		scriptNameWithEnv := fmt.Sprintf("%s:%s", envName, scriptName)

		_, ok := app.GpmFile.Scripts[scriptNameWithEnv]
		if ok {
			finalScriptName = scriptNameWithEnv
		}
	}

	cmdToExecute := app.GpmFile.Scripts[finalScriptName]

	p := utils.CreateShellCommand(cmdToExecute)

	app.Debug(fmt.Sprintf("Running script '%v' ...", scriptName))
	utils.RunCommand(p, additionalArgs...)
}

// app.RunShellCommand() - runs a shell command in app's context
func (app *AppContext) RunShellCommand(cmd string) {
	app.Debug(fmt.Sprintf("Running '%v' ...", cmd))

	p := utils.CreateShellCommand(cmd)
	p.Dir = app.Cwd

	utils.RunCommand(p)
}

// app.RunShellCommandByArgs() - runs a shell command by arguments in app's context
func (app *AppContext) RunShellCommandByArgs(c string, a ...string) {
	app.Debug(fmt.Sprintf("Running '%v %v' ...", c, strings.Join(a, " ")))

	p := utils.CreateShellCommandByArgs(c, a...)
	p.Dir = app.Cwd

	utils.RunCommand(p)
}

// app.TidyUp() - runs 'go mod tidy' for the current project (folder)
func (app *AppContext) TidyUp(options ...TidyUpOptions) {
	args := []string{}
	noScript := false

	// collect and overwrite options if needed
	for _, o := range options {
		if o.Arguments != nil {
			args = *o.Arguments
		}
		if o.NoScript != nil {
			noScript = *o.NoScript
		}
	}

	_, ok := app.GpmFile.Scripts[constants.TidyScriptName]
	if !noScript && ok {
		app.RunScript(constants.TidyScriptName, args...)
	} else {
		app.RunShellCommandByArgs("go", "mod", "tidy")
	}
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
	utils.CheckForError(err)

	app.Debug(fmt.Sprintf("Updating alias file '%v' ...", aliasesFilePath))
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
	utils.CheckForError(err)

	app.Debug(fmt.Sprintf("Updating project file '%v' ...", projectsFilePath))
	return os.WriteFile(projectsFilePath, yamlData, constants.DefaultFileMode)
}

// app.Write() - implementation for an io.Writer
func (app *AppContext) Write(p []byte) (int, error) {
	if app.Out == nil {
		return len(p), nil // deactivated
	}

	return app.Out.Write(p)
}

// app.Write() - implementation for an io.Writer
func (app *AppContext) WriteError(p []byte) (int, error) {
	if app.ErrorOut == nil {
		return len(p), nil // deactivated
	}

	return app.ErrorOut.Write(p)
}

// app.WriteAllInputsTo() - reads from all inputs (STDIN and files, in that order) to an io.Writer
func (app *AppContext) WriteAllInputsTo(w io.Writer, files ...string) (int64, error) {
	var totalWritten int64 = 0

	// first check STDIN
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		written, err := io.Copy(w, os.Stdin)
		if err != nil {
			return written, err
		}

		totalWritten += written
	}

	// now from files
	for _, f := range files {
		filePathOrUrl := strings.TrimSpace(f)
		if filePathOrUrl == "" {
			continue
		}

		var readData func() (int64, error)
		if utils.IsDownloadUrl(filePathOrUrl) {
			// in this case `filePath` is a downloadable URL

			readData = func() (int64, error) {
				return utils.DownloadFromUrlTo(app.Out, filePathOrUrl)
			}
		} else {
			filePath := app.GetFullPathOrDefault(filePathOrUrl, "")

			readData = func() (int64, error) {
				file, err := os.Open(filePath)
				utils.CheckForError(err)
				defer file.Close()

				return io.Copy(app.Out, file)
			}
		}

		written, err := readData()

		if err != nil {
			return totalWritten, err
		}
		totalWritten += written
	}

	return totalWritten, nil
}
