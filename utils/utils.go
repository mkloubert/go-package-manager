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
	"bufio"
	"bytes"
	"fmt"
	"io"
	mathRand "math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/schollz/progressbar/v3"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

type SpliTextOptions struct {
	MaxChunkSize     *int // default 3000
	MaxOverheadWords *int // default 15
	MaxOverheadChars *int // default 100
}

// Base64FromDataURI() - extracts Base64 part from data URI
func Base64FromDataURI(dataURI string) (string, error) {
	dataURI = strings.TrimSpace(dataURI)
	if !strings.HasPrefix(dataURI, "data:") {
		return "", fmt.Errorf("no data URI")
	}

	parts := strings.Split(dataURI, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid data URI format")
	}

	return strings.TrimSpace(parts[1]), nil
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

// DownloadFromUrl() - downloads data from URL
func DownloadFromUrl(url string) ([]byte, error) {
	buffer := bytes.Buffer{}
	_, err := DownloadFromUrlTo(&buffer, url)

	return buffer.Bytes(), err
}

// DownloadFromUrlTo() - downloads data from URL to an io.Writer
func DownloadFromUrlTo(w io.Writer, url string) (int64, error) {
	if !IsDownloadUrl(url) {
		url = "https://" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return io.Copy(w, resp.Body)
}

// EnsureMaxSliceLength() - ensures that the length of an array is
// not greater than a maximum and returns a truncated copy; otherwise
// the input array
func EnsureMaxSliceLength[T any](slice []T, maxLength int) []T {
	if len(slice) > maxLength {
		return slice[0:maxLength]
	}
	return slice
}

// GenerateRandomUint16() - creates a new random uint16 value
func GenerateRandomUint16() uint16 {
	return uint16(mathRand.Intn(1 << 16)) // 1 << 16 is 65536, the range of uint16
}

// GetBoolFlag() - returns a boolean command line flag value without error
func GetBoolFlag(cmd *cobra.Command, name string, defaultValue bool) bool {
	val, err := cmd.Flags().GetBool(name)
	if err == nil {
		return val
	}

	return defaultValue
}

// GetEnvVar() - returns, if found, the value of an existing environment
// variable by its name ignoring case sensitivity
func GetEnvVar(name string) *string {
	lowerName := strings.TrimSpace(strings.ToUpper(name))

	var value *string = nil

	allVars := os.Environ()
	for _, kv := range allVars {
		sep := strings.Index(kv, "=")

		var n string
		var v string
		if sep > -1 {
			n = kv[0:sep]
			v = kv[:sep+1]
		} else {
			n = kv
		}

		n = strings.TrimSpace(strings.ToLower(n))
		if n == lowerName {
			value = &v
		}
	}

	return value
}

// GetNumberOfOpenFilesByPid() - returns the number of open files by pid
func GetNumberOfOpenFilesByPid(pid int32) (int64, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return -1, err
	}

	openFiles, err := proc.OpenFiles()
	if err == nil {
		return int64(len(openFiles)), nil
	}

	if IsWindows() {
		return -1, err
	}

	if IsMacOS() {
		cmd := exec.Command("lsof", "-p", strconv.FormatInt(int64(pid), 10))
		output, err := cmd.Output()
		if err != nil {
			return -1, err
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		count := 0
		for scanner.Scan() {
			count++
		}

		return int64(count - 1), nil
	}

	fdDir := filepath.Join("/proc", strconv.FormatInt(int64(pid), 10), "fd")

	files, err := os.ReadDir(fdDir)
	if err != nil {
		return -1, err
	}

	return int64(len(files)), nil
}

// GetShell()- returns the name of the current shell
func GetShell() string {
	shellName := ""

	if IsWindows() {
		comspec := GetEnvVar("COMSPEC")
		if comspec != nil {
			shellName = *comspec
		} else {
			powershell := GetEnvVar("PSModulePath")
			if powershell != nil && *powershell != "" {
				shellName = "PowerShell"
			}
		}
	} else {
		shellName = os.Getenv("SHELL")
	}

	shellName = strings.TrimSpace(shellName)

	if shellName == "" {
		shellName = "unknown"
	} else {
		lowerShellName := strings.ToLower(shellName)
		if strings.Contains(lowerShellName, "cmd.exe") {
			shellName = "cmd.exe"
		} else if strings.Contains(lowerShellName, "zsh") {
			shellName = "Z shell"
		} else if strings.Contains(lowerShellName, "bash") {
			shellName = "Bash"
		}
	}

	return shellName
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

// IsDownloadUrl() - checks if url is a valid URL to a resource
// that can be downloaded
func IsDownloadUrl(url string) bool {
	url = strings.TrimSpace(url)

	return strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:")
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

// MaxUint16() - returns the maximum uint16 value
func MaxUint16(a, b uint16, more ...uint16) uint16 {
	var result uint16 = b
	if a > b {
		result = a
	}

	for _, c := range more {
		if c > result {
			result = c
		}
	}

	return result
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

// UpdateUsageSparkline() - updates color of and
// data of an widgets.Sparkline item
// that represents "usage data"
func UpdateUsageSparkline(s *widgets.Sparkline, newData []float64) {
	newColor := ui.ColorWhite

	maxVal := s.MaxVal

	if maxVal != 0.0 && len(newData) > 0 {
		firstItem := newData[0]
		unusageValue := maxVal - firstItem
		percentage := unusageValue / maxVal

		newColor = ui.ColorGreen
		if percentage <= 0.25 { // <= 25%?
			newColor = ui.ColorRed
		} else if percentage <= 0.5 { // <= 50%?
			newColor = ui.ColorYellow
		}
	}

	s.Data = newData
	s.LineColor = newColor
}
