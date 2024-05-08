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
	"fmt"
	"log"
	"strings"

	"github.com/mkloubert/go-package-manager/utils"
)

// An AppContext contains all information for running this app
type AppContext struct {
	L            *log.Logger  // the logger to use
	PackagesFile PackagesFile // the packages.y(a)ml file
	Verbose      bool         // output verbose information
}

// app.Debug() - writes debug information with the underlying logger
func (app *AppContext) Debug(v ...any) *AppContext {
	if app.Verbose {
		app.L.Printf("[VERBOSE] %v", fmt.Sprintln(v...))
	}

	return app
}

// app.GetModuleUrls() - returns the list of module urls based on the
// information from packages.y(a)ml file
func (app *AppContext) GetModuleUrls(moduleNameOrUrl string) []string {
	moduleNameOrUrl = utils.CleanupModuleName(moduleNameOrUrl)

	urls := make([]string, 0)

	for k, v := range app.PackagesFile.Packages {
		// collect all module aliases
		allModuleAliases := []string{strings.TrimSpace(k)}        // main alias
		allModuleAliases = append(allModuleAliases, v.Aliases...) // sub aliases

		// checkout if matching
		for _, ma := range allModuleAliases {
			if ma == moduleNameOrUrl {
				for _, s := range v.Sources {
					urls = append(urls, utils.CleanupModuleName(s))
				}

				break
			}
		}
	}

	if len(urls) == 0 {
		// take input as fallback
		urls = append(urls, moduleNameOrUrl)
	}

	return urls
}
