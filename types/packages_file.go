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
	"os"
	"path"

	"github.com/goccy/go-yaml"
	"github.com/mkloubert/go-package-manager/utils"
)

// A PackagesFile stores all data of a packages.y(a)ml file.
type PackagesFile struct {
	Packages map[string]PackagesFilePackageItem `yaml:"packages"` // the package mappings
}

// A PackagesFilePackageItem is an item inside `PackagesFile.Packages` map.
type PackagesFilePackageItem struct {
	Sources []string `yaml:"sources"` // one or more source repositories
}

// LoadPackagesFileIfExist - Loads a packages.y(a)ml file if it exists
// and return `true` if file has been loaded successfully.
func LoadPackagesFileIfExist(app *AppContext) bool {
	cwd, err := os.Getwd()
	if err == nil {
		packagesFilePath := path.Join(cwd, "packages.yaml")
		info, err := os.Stat(packagesFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return false
			}

			utils.CloseWithError(err)
		}

		if info.IsDir() {
			return false
		}

		yamlData, err := os.ReadFile(packagesFilePath)
		if err != nil {
			utils.CloseWithError(err)
		}

		var pf PackagesFile
		err = yaml.Unmarshal(yamlData, &pf)
		if err != nil {
			utils.CloseWithError(err)
		}

		app.PackagesFile = pf
		return true
	}

	return false
}
