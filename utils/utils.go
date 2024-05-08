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
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

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

// GetBoolFlag() - returns a boolean command line flag value without error
func GetBoolFlag(cmd *cobra.Command, name string, defaultValue bool) bool {
	val, err := cmd.Flags().GetBool(name)
	if err == nil {
		return val
	}

	return defaultValue
}
