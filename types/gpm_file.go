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
	"os"

	"github.com/goccy/go-yaml"
)

// GpmFile stores all data of a gpm.y(a)ml file.
type GpmFile struct {
	Contributors []GpmFileContributor   `yaml:"contributors,omitempty"` // list of contributors
	Description  string                 `yaml:"description,omitempty"`  // the description
	DisplayName  string                 `yaml:"display_name,omitempty"` // the display name
	Donations    map[string]string      `yaml:"donations,omitempty"`    // one or more donation links
	Files        []string               `yaml:"files,omitempty"`        // whitelist of file patterns which are used by pack command for example
	Homepage     string                 `yaml:"homepage,omitempty"`     // the homepage
	License      string                 `yaml:"license,omitempty"`      // the license
	Name         string                 `yaml:"name,omitempty"`         // the name
	Repositories []GpmFileRepository    `yaml:"repositories,omitempty"` // source code repository information
	Scripts      map[string]string      `yaml:"scripts,omitempty"`      // one or more scripts
	Settings     map[string]interface{} `yaml:"settings,omitempty"`     // custom settings
	yamlData     []byte
}

// GpmFileContributor is an item inside `Contributors` of a
// `GpmFile` instance
type GpmFileContributor struct {
	Homepage string `yaml:"homepage,omitempty"` // the homepage url
	Name     string `yaml:"name,omitempty"`     // the full name
	Role     string `yaml:"role,omitempty"`     // the role
}

// GpmFileRepository is an item inside `Repositories` of a
// `GpmFile` instance
type GpmFileRepository struct {
	Name string `yaml:"name,omitempty"` // the full name
	Type string `yaml:"type,omitempty"` // the type
	Url  string `yaml:"url,omitempty"`  // the url
}

// GetFilesSectionByEnvSafe() - will return environment specific `files` section in `gpm.yaml`
// file, if exists, otherwise the default one
func (g *GpmFile) GetFilesSectionByEnvSafe(envName string) []string {
	if envName != "" {
		var gpmFileAsMap map[string]interface{}
		err := yaml.Unmarshal(g.yamlData, &gpmFileAsMap)

		if err == nil && gpmFileAsMap != nil {
			key := fmt.Sprintf("files:%s", envName)

			maybeArray, ok := gpmFileAsMap[key]
			if ok && maybeArray != nil {
				files, ok := maybeArray.([]string)
				if ok && files != nil {
					return files // found existing, valid string array
				}
			}
		}
	}
	return g.Files
}

// GetSettingsSectionByEnvSafe() - will return environment specific `settings` section in `gpm.yaml`
// file, if exists, otherwise the default one
func (g *GpmFile) GetSettingsSectionByEnvSafe(envName string) map[string]interface{} {
	if envName != "" {
		var gpmFileAsMap map[string]interface{}
		err := yaml.Unmarshal(g.yamlData, &gpmFileAsMap)

		if err == nil && gpmFileAsMap != nil {
			key := fmt.Sprintf("settings:%s", envName)

			maybeMap, ok := gpmFileAsMap[key]
			if ok && maybeMap != nil {
				settings, ok := maybeMap.(map[string]interface{})
				if ok && settings != nil {
					return settings // found existing, valid map
				}
			}
		}
	}
	return g.Settings
}

// LoadGpmFile() - Loads a gpm.yaml file via a file path
func LoadGpmFile(gpmFilePath string) (GpmFile, error) {
	var gpm GpmFile
	defer func() {
		if gpm.Contributors == nil {
			gpm.Contributors = []GpmFileContributor{}
		}
		if gpm.Donations == nil {
			gpm.Donations = map[string]string{}
		}
		if gpm.Files == nil {
			gpm.Files = []string{}
		}
		if gpm.Repositories == nil {
			gpm.Repositories = []GpmFileRepository{}
		}
		if gpm.Scripts == nil {
			gpm.Scripts = map[string]string{}
		}
		if gpm.Settings == nil {
			gpm.Settings = map[string]interface{}{}
		}
	}()

	yamlData, err := os.ReadFile(gpmFilePath)
	if err != nil {
		return gpm, err
	}

	err = yaml.Unmarshal(yamlData, &gpm)
	gpm.yamlData = yamlData

	return gpm, err
}
