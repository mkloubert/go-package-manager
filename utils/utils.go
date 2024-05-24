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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

type SpliTextOptions struct {
	MaxChunkSize     *int // default 3000
	MaxOverheadWords *int // default 15
	MaxOverheadChars *int // default 100
}

// CheckForError() - exits with code 1 and output an error (if there is an error)
func CheckForError(err error) {
	if err != nil {
		CloseWithError(err)
	}
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

// ClearConsole() - clears the console
func ClearConsole() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	fmt.Print("\033[H\033[2J")
	return nil
}

// CloseWithError() - exits with code 1 and output an error
func CloseWithError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// CreateProgressBar() - creates a simple progress bar with default settings
func CreateProgressBar(totalCount int, description string) *progressbar.ProgressBar {
	newBar := progressbar.NewOptions(
		totalCount,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	return newBar
}

// CreateShellCommand() - creates a new shell command based on the operating system
// without running it
func CreateShellCommand(c string) *exec.Cmd {
	var p *exec.Cmd
	if IsWindows() {
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

// DownloadFromUrl() - downloads data from URL
func DownloadFromUrl(url string) ([]byte, error) {
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "https://" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// GetAIChatTemperature() - returns the value for AI chat conversation temperature
func GetAIChatTemperature(defaultValue float32) float32 {
	GPM_AI_CHAT_TEMPERATURE := strings.TrimSpace(os.Getenv("GPM_AI_CHAT_TEMPERATURE"))
	if GPM_AI_CHAT_TEMPERATURE != "" {
		value64, err := strconv.ParseFloat(GPM_AI_CHAT_TEMPERATURE, 32)
		if err == nil {
			return float32(value64)
		}
	}

	return defaultValue
}

// GetBestChromaFormatterName() - returns the best syntax highlight formatter for the console
func GetBestChromaFormatterName() string {
	GPM_TERMINAL_FORMATTER := strings.TrimSpace(
		strings.ToLower(os.Getenv("GPM_TERMINAL_FORMATTER")),
	)
	if GPM_TERMINAL_FORMATTER != "" {
		return GPM_TERMINAL_FORMATTER
	}

	switch os := runtime.GOOS; os {
	case "darwin":
	case "linux":
		return "terminal16m"
	case "windows":
		return "terminal256"
	}

	return "terminal"
}

// GetBestChromaStyleName() - returns the best syntax highlight style for the console
func GetBestChromaStyleName() string {
	GPM_TERMINAL_STYLE := strings.TrimSpace(
		strings.ToLower(os.Getenv("GPM_TERMINAL_STYLE")),
	)

	if GPM_TERMINAL_STYLE != "" {
		return GPM_TERMINAL_STYLE
	}
	return "dracula"
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

// IsDirExisting() - checks if path is an existing directory
func IsDirExisting(dp string) (bool, error) {
	info, err := os.Stat(dp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return info.IsDir(), nil
}

// IsFileExisting() - checks if path is an existing file
func IsFileExisting(fp string) (bool, error) {
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return !info.IsDir(), nil
}

// IsWindows() - checks if current operating system is Microsoft Windows or not
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// LoadFromSTDINIfAvailable() - loads data from STDIN if available
func LoadFromSTDINIfAvailable() (*[]byte, error) {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			return &data, nil
		}
		return nil, err
	}

	return nil, nil
}

// ListFiles() - checks if current operating system is Microsoft Windows or not
func ListFiles(dir string, pattern string) ([]string, error) {
	var matchingFiles []string

	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}

		r := regexp.MustCompile(pattern)
		if r.MatchString(relPath) {
			matchingFiles = append(matchingFiles, p)
		}

		return nil
	})
	return matchingFiles, err
}

// OpenUrl() - opens a URL by the default application handler
func OpenUrl(url string) error {
	var args []string
	var cmd string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

// RemoveDuplicatesInStringList() - removes duplicates in string list
func RemoveDuplicatesInStringList(arr []string) []string {
	alreadySeen := map[string]bool{}
	result := []string{}

	for _, value := range arr {
		if _, alreadyExists := alreadySeen[value]; !alreadyExists {
			alreadySeen[value] = true
			result = append(result, value)
		}
	}

	alreadySeen = nil

	return result
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

// SplitText() - splits text into chunks of a maximum size
func SplitText(text string, options ...SpliTextOptions) []string {
	maxChunkSize := 3000
	maxOverheadWords := 15
	maxOverheadChars := 100

	// collect options
	for _, o := range options {
		if o.MaxChunkSize != nil {
			maxChunkSize = *o.MaxChunkSize
		}
		if o.MaxOverheadWords != nil {
			maxOverheadWords = *o.MaxOverheadWords
		}
		if o.MaxOverheadChars != nil {
			maxOverheadChars = *o.MaxOverheadChars
		}
	}

	var chunks []string
	var overhead string

	for len(text) > 0 {
		if len(text) <= maxChunkSize {
			// we have no overhead (anymore)
			chunks = append(chunks, overhead+text)
			break
		}

		// determine end position of current chunk
		end := maxChunkSize
		for end < len(text) && (text[end] != ' ' && text[end] != '\n') {
			end++
		}

		// temp chunk before adjusting for overhead
		chunk := text[:end]
		remainingText := text[end:]

		words := strings.Fields(chunk)
		wordCount := len(words)

		if wordCount > maxOverheadWords {
			// overhead of words
			overheadWords := words[wordCount-maxOverheadWords:]

			overhead = strings.Join(overheadWords, " ")
			chunk = strings.Join(words[:wordCount-maxOverheadWords], " ")
		} else if len(chunk) > maxChunkSize-maxOverheadChars {
			// overhead of chunk size
			overhead = chunk[maxChunkSize-maxOverheadChars:]
			chunk = chunk[:maxChunkSize-maxOverheadChars]
		} else {
			overhead = "" // no overhead
		}

		chunks = append(chunks, overhead+chunk)

		// prepare the remaining text for next iteration
		text = overhead + remainingText
		overhead = ""
	}

	return chunks
}

// ToUrlForOpenHandler() - converts an input URL to a URL which
// can be opened by handler of the current operating system
func ToUrlForOpenHandler(originalUrl string) (string, error) {
	urlObj, err := url.Parse(originalUrl)
	if err != nil {
		return "", err
	}

	port := urlObj.Port()
	if port != "" {
		port = ":" + port
	}

	query := urlObj.RawQuery
	if query != "" {
		query = "?" + query
	}

	return fmt.Sprintf(
		"https://%v%v%v%v",
		urlObj.Host, port,
		urlObj.Path, query,
	), nil
}
