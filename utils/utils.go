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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// ChatWithAIOption stores settings for
// `ChatWithAI()` function
type ChatWithAIOption struct {
	Model        *string // custom model
	SystemPrompt *string // custom system prompt
	Temperature  *int    // custom temperature
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

// CleanupModuleName() - cleans up a module name
func CleanupModuleName(moduleName string) string {
	moduleName = strings.TrimSpace(moduleName)

	parsedURL, err := url.Parse(moduleName)
	if err == nil {
		moduleName = fmt.Sprintf(
			"%v%v%v",
			parsedURL.Host, parsedURL.Port(),
			parsedURL.Path,
		)
	}

	return strings.TrimSpace(moduleName)
}

// CloseWithError() - exits with code 1 and output an error
func CloseWithError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// ChatWithAI() - does a simple AI chat based on the current app settings
func ChatWithAI(prompt string, options ...ChatWithAIOption) (string, error) {
	OPENAI_API_KEY := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if OPENAI_API_KEY == "" {
		return "", fmt.Errorf("no key found in OPENAI_API_KEY")
	}

	return chatWithOpenAI(prompt, options...)
}

func chatWithOpenAI(prompt string, options ...ChatWithAIOption) (string, error) {
	model := GetDefaultAIChatModel()
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

// CreateShellCommand() - creates a new shell command based on the operating system
// without running it
func CreateShellCommand(c string) *exec.Cmd {
	var p *exec.Cmd
	if runtime.GOOS == "windows" {
		p = CreateShellCommandByArgs("cmd", "/C", c)
	} else {
		// UNIX / Linux
		p = CreateShellCommandByArgs("sh", "-c", c)
	}

	return p
}

// CreateShellCommand() - creates a new shell command without running it
func CreateShellCommandByArgs(c string, args ...string) *exec.Cmd {
	p := exec.Command(c, args...)

	p.Env = os.Environ()
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	p.Stdin = os.Stdin

	return p
}

// GetBoolFlag() - returns a boolean command line flag value without error
func GetBoolFlag(cmd *cobra.Command, name string, defaultValue bool) bool {
	val, err := cmd.Flags().GetBool(name)
	if err == nil {
		return val
	}

	return defaultValue
}

// GetDefaultAIChatModel() - returns the name of the default AI chat model
func GetDefaultAIChatModel() string {
	return strings.TrimSpace(os.Getenv("GPM_AI_CHAT_MODEL"))
}

// IndexOfString() - returns the zero-based index of a string in a string array
// or -1 if not found
func IndexOfString(arr []string, value string) int {
	for index, str := range arr {
		if str == value {
			return index
		}
	}

	return -1
}

// RunCommand() - runs a command and exists on error
func RunCommand(p *exec.Cmd, additionalArgs ...string) {
	p.Args = append(p.Args, additionalArgs...)

	if err := p.Run(); err != nil {
		CloseWithError(err)
	}
}

// Slugify() - slugifies a string
func Slugify(str string, rx ...string) string {
	str = strings.ToLower(str)

	if len(rx) == 0 {
		// only english letter, digits, whitespaces and -
		rx = []string{`[^a-z0-9\\s-]`}
	}

	for _, r := range rx {
		str = regexp.MustCompile(r).ReplaceAllString(str, "")
	}

	// only digits, english letters, whitespaces and -
	// whitespace to -
	str = regexp.MustCompile(`\\s+`).ReplaceAllString(str, "-")
	// remove beginning and ending - characters
	str = strings.Trim(str, "-")

	return str
}
